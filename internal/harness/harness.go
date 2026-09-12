package harness

import (
	"os"
	"path/filepath"

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
	return h.SyncHooks()
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

// AllHarnesses returns the fixed harness list, matching Python's
// ALL_HARNESSES order: Claude, Codex, Cursor, OpenCode.
func AllHarnesses(repo string) []*Harness {
	return []*Harness{
		NewClaude(repo),
		NewCodex(repo),
		NewCursor(repo),
		NewOpenCode(repo),
	}
}
