package harness

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/phedayat/agent_synchronizer/internal/syncengine"
)

// Seams for tests to substitute stubs without touching the filesystem.
var (
	syncTarget                 = syncengine.SyncTarget
	syncFlattenedSkills        = syncengine.SyncFlattenedSkills
	syncPartiallyGroupedSkills = syncengine.SyncPartiallyGroupedSkills
)

// Harness mirrors sync_engine.py's Harness ABC as a single config-driven
// struct: every subclass in the Python source shares the same shape of
// overrides, so a struct with empty-string-means-no-op fields (matching
// Cursor's literal `pass` for sync_config and every non-Claude harness's
// no-op sync_hooks) captures the behavior without an interface hierarchy.
type Harness struct {
	Name     string
	Home     string
	RepoRoot string

	ConfigSrc, ConfigDest string
	RulesSrc, RulesDest   string
	HooksSrc, HooksDest   string

	PreserveSkillGroups bool
}

func (h *Harness) RepoDir() string {
	return filepath.Join(h.RepoRoot, h.Name)
}

func (h *Harness) CommonDir() string {
	return filepath.Join(h.RepoRoot, "common")
}

func (h *Harness) SkillsDir() string {
	return filepath.Join(h.RepoDir(), "skills")
}

func (h *Harness) SyncSkills() error {
	if h.PreserveSkillGroups {
		return syncPartiallyGroupedSkills(
			filepath.Join(h.Home, "skills"),
			filepath.Join(h.CommonDir(), "skills"),
			h.SkillsDir(),
		)
	}
	return syncFlattenedSkills(
		filepath.Join(h.Home, "skills"),
		filepath.Join(h.CommonDir(), "skills"),
		h.SkillsDir(),
	)
}

func (h *Harness) SyncSubagents() error {
	return syncTarget(filepath.Join(h.CommonDir(), "agents"), filepath.Join(h.Home, "agents"))
}

func (h *Harness) SyncConfig() error {
	if h.ConfigSrc == "" {
		return nil
	}
	return syncTarget(h.ConfigSrc, h.ConfigDest)
}

func (h *Harness) SyncRules() error {
	if h.RulesSrc == "" {
		return nil
	}
	return syncTarget(h.RulesSrc, h.RulesDest)
}

func (h *Harness) SyncHooks() error {
	if h.HooksSrc == "" {
		return nil
	}
	return syncTarget(h.HooksSrc, h.HooksDest)
}

// extrasKnownNames returns the top-level basenames already handled by an
// explicit Sync* method, so SyncExtras skips them.
func (h *Harness) extrasKnownNames() map[string]struct{} {
	known := map[string]struct{}{"skills": {}, "agents": {}}
	for _, src := range []string{h.ConfigSrc, h.RulesSrc, h.HooksSrc} {
		if src != "" {
			known[filepath.Base(src)] = struct{}{}
		}
	}
	return known
}

// syncExtrasDir symlinks every top-level entry of dir into h.Home except
// those in known, via the same SyncTarget primitive used for config/rules/
// hooks: a file or directory becomes one symlink, real pre-existing content
// at dest is absorbed first.
func (h *Harness) syncExtrasDir(dir string, known map[string]struct{}) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		if _, ok := known[name]; ok {
			continue
		}
		if err := syncTarget(filepath.Join(dir, name), filepath.Join(h.Home, name)); err != nil {
			return err
		}
	}
	return nil
}

// SyncExtras symlinks every file/dir in the harness's own repo dir and in
// common/ that isn't already covered by SyncSkills/SyncSubagents/
// SyncConfig/SyncRules/SyncHooks. The repo dir is synced first so a
// harness-specific entry claims its dest name before common/ is considered
// for it: SyncTarget treats an already-symlinked dest as done, so whichever
// source runs first wins on a name collision, matching skills'
// harness-overrides-common precedent.
func (h *Harness) SyncExtras() error {
	known := h.extrasKnownNames()
	if err := h.syncExtrasDir(h.RepoDir(), known); err != nil {
		return err
	}
	return h.syncExtrasDir(h.CommonDir(), known)
}

func (h *Harness) Sync() error {
	if err := h.SyncSkills(); err != nil {
		return err
	}
	if err := h.SyncSubagents(); err != nil {
		return err
	}
	if err := h.SyncConfig(); err != nil {
		return err
	}
	if err := h.SyncRules(); err != nil {
		return err
	}
	if err := h.SyncHooks(); err != nil {
		return err
	}
	return h.SyncExtras()
}

// userHomeDir wraps os.UserHomeDir; a failure here mirrors Python's
// Path.home() raising, an unrecoverable environment condition, so
// constructors panic rather than threading an error through a signature
// that AllHarnesses's callers expect to return []Harness directly.
func userHomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return dir
}

func NewClaude(repo string) *Harness {
	home := filepath.Join(userHomeDir(), ".claude")
	repoDir := filepath.Join(repo, "claude")
	return &Harness{
		Name:       "claude",
		Home:       home,
		RepoRoot:   repo,
		ConfigSrc:  filepath.Join(repoDir, "settings.json"),
		ConfigDest: filepath.Join(home, "settings.json"),
		RulesSrc:   filepath.Join(repoDir, "CLAUDE.md"),
		RulesDest:  filepath.Join(home, "CLAUDE.md"),
		HooksSrc:   filepath.Join(repoDir, "hooks"),
		HooksDest:  filepath.Join(home, "hooks"),
	}
}

func NewCodex(repo string) *Harness {
	home := filepath.Join(userHomeDir(), ".codex")
	repoDir := filepath.Join(repo, "codex")
	return &Harness{
		Name:       "codex",
		Home:       home,
		RepoRoot:   repo,
		ConfigSrc:  filepath.Join(repoDir, "config.toml"),
		ConfigDest: filepath.Join(home, "config.toml"),
		RulesSrc:   filepath.Join(repo, "common", "AGENTS.md"),
		RulesDest:  filepath.Join(home, "AGENTS.md"),
	}
}

func NewCursor(repo string) *Harness {
	home := filepath.Join(userHomeDir(), ".cursor")
	repoDir := filepath.Join(repo, "cursor")
	return &Harness{
		Name:      "cursor",
		Home:      home,
		RepoRoot:  repo,
		RulesSrc:  filepath.Join(repoDir, ".cursorrules"),
		RulesDest: filepath.Join(home, ".cursorrules"),
	}
}

func NewOpenCode(repo string) *Harness {
	home := filepath.Join(userHomeDir(), ".config", "opencode")
	repoDir := filepath.Join(repo, "opencode")
	return &Harness{
		Name:       "opencode",
		Home:       home,
		RepoRoot:   repo,
		ConfigSrc:  filepath.Join(repoDir, "opencode.jsonc"),
		ConfigDest: filepath.Join(home, "opencode.jsonc"),
		RulesSrc:   filepath.Join(repo, "common", "AGENTS.md"),
		RulesDest:  filepath.Join(home, "AGENTS.md"),
	}
}

func NewHermes(repo string) *Harness {
	home := filepath.Join(userHomeDir(), ".hermes")
	repoDir := filepath.Join(repo, "hermes")
	return &Harness{
		Name:                "hermes",
		Home:                home,
		RepoRoot:            repo,
		ConfigSrc:           filepath.Join(repoDir, "config.json"),
		ConfigDest:          filepath.Join(home, "config.json"),
		RulesSrc:            filepath.Join(repo, "common", "AGENTS.md"),
		RulesDest:           filepath.Join(home, "AGENTS.md"),
		PreserveSkillGroups: true,
	}
}

// AllHarnesses returns the fixed harness list, matching Python's
// ALL_HARNESSES order: Claude, Codex, Cursor, OpenCode, Hermes.
func AllHarnesses(repo string) []*Harness {
	return []*Harness{
		NewClaude(repo),
		NewCodex(repo),
		NewCursor(repo),
		NewOpenCode(repo),
		NewHermes(repo),
	}
}
