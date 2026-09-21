package harness

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type targetCall struct {
	src, dest string
}

type skillsCall struct {
	dest    string
	sources []string
}

func stubSync(t *testing.T) (*[]targetCall, *[]skillsCall) {
	t.Helper()
	var targetCalls []targetCall
	var skillsCalls []skillsCall

	origTarget := syncTarget
	origSkills := syncFlattenedSkills
	syncTarget = func(src, dest string) error {
		targetCalls = append(targetCalls, targetCall{src, dest})
		return nil
	}
	syncFlattenedSkills = func(dest string, sources ...string) error {
		skillsCalls = append(skillsCalls, skillsCall{dest, sources})
		return nil
	}
	t.Cleanup(func() {
		syncTarget = origTarget
		syncFlattenedSkills = origSkills
	})
	return &targetCalls, &skillsCalls
}

func stubPartiallyGroupedSkills(t *testing.T) *[]skillsCall {
	t.Helper()
	var calls []skillsCall

	orig := syncPartiallyGroupedSkills
	syncPartiallyGroupedSkills = func(dest, flattenSrc, groupedSrc string) error {
		calls = append(calls, skillsCall{dest, []string{flattenSrc, groupedSrc}})
		return nil
	}
	t.Cleanup(func() {
		syncPartiallyGroupedSkills = orig
	})
	return &calls
}

func containsTarget(calls []targetCall, src, dest string) bool {
	for _, c := range calls {
		if c.src == src && c.dest == dest {
			return true
		}
	}
	return false
}

func TestClaudeSyncTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := "/repo"
	h := NewClaude(repo)
	home := h.Home

	targetCalls, skillsCalls := stubSync(t)

	if err := h.SyncSkills(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncSubagents(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncConfig(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncRules(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncHooks(); err != nil {
		t.Fatal(err)
	}

	wantSkills := skillsCall{
		dest:    filepath.Join(home, "skills"),
		sources: []string{filepath.Join(repo, "common", "skills"), filepath.Join(repo, "claude", "skills")},
	}
	if len(*skillsCalls) != 1 || !reflect.DeepEqual((*skillsCalls)[0], wantSkills) {
		t.Fatalf("skills call = %+v, want %+v", *skillsCalls, wantSkills)
	}

	if !containsTarget(*targetCalls, filepath.Join(repo, "common", "agents"), filepath.Join(home, "agents")) {
		t.Fatal("missing subagents call")
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "claude", "settings.json"), filepath.Join(home, "settings.json")) {
		t.Fatal("missing config call")
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "claude", "CLAUDE.md"), filepath.Join(home, "CLAUDE.md")) {
		t.Fatal("missing rules call")
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "claude", "hooks"), filepath.Join(home, "hooks")) {
		t.Fatal("missing hooks call")
	}
	if len(*targetCalls) != 4 {
		t.Fatalf("targetCalls count = %d, want 4", len(*targetCalls))
	}
}

func TestCodexSyncTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := "/repo"
	h := NewCodex(repo)
	home := h.Home

	targetCalls, skillsCalls := stubSync(t)

	if err := h.SyncSkills(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncSubagents(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncConfig(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncRules(); err != nil {
		t.Fatal(err)
	}

	wantSkills := skillsCall{
		dest:    filepath.Join(home, "skills"),
		sources: []string{filepath.Join(repo, "common", "skills"), filepath.Join(repo, "codex", "skills")},
	}
	if len(*skillsCalls) != 1 || !reflect.DeepEqual((*skillsCalls)[0], wantSkills) {
		t.Fatalf("skills call = %+v, want %+v", *skillsCalls, wantSkills)
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "common", "agents"), filepath.Join(home, "agents")) {
		t.Fatal("missing subagents call")
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "codex", "config.toml"), filepath.Join(home, "config.toml")) {
		t.Fatal("missing config call")
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "common", "AGENTS.md"), filepath.Join(home, "AGENTS.md")) {
		t.Fatal("missing rules call")
	}
	if len(*targetCalls) != 3 {
		t.Fatalf("targetCalls count = %d, want 3", len(*targetCalls))
	}
}

func TestCursorSyncTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := "/repo"
	h := NewCursor(repo)
	home := h.Home

	targetCalls, skillsCalls := stubSync(t)

	if err := h.SyncSkills(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncSubagents(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncRules(); err != nil {
		t.Fatal(err)
	}

	wantSkills := skillsCall{
		dest:    filepath.Join(home, "skills"),
		sources: []string{filepath.Join(repo, "common", "skills"), filepath.Join(repo, "cursor", "skills")},
	}
	if len(*skillsCalls) != 1 || !reflect.DeepEqual((*skillsCalls)[0], wantSkills) {
		t.Fatalf("skills call = %+v, want %+v", *skillsCalls, wantSkills)
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "common", "agents"), filepath.Join(home, "agents")) {
		t.Fatal("missing subagents call")
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "cursor", ".cursorrules"), filepath.Join(home, ".cursorrules")) {
		t.Fatal("missing rules call")
	}
	if len(*targetCalls) != 2 {
		t.Fatalf("targetCalls count = %d, want 2", len(*targetCalls))
	}
}

func TestCursorSyncConfigIsNoop(t *testing.T) {
	h := NewCursor("/repo")
	targetCalls, _ := stubSync(t)

	if err := h.SyncConfig(); err != nil {
		t.Fatal(err)
	}
	if len(*targetCalls) != 0 {
		t.Fatalf("expected no calls, got %+v", *targetCalls)
	}
}

func TestCodexSyncHooksIsNoop(t *testing.T) {
	h := NewCodex("/repo")
	targetCalls, _ := stubSync(t)

	if err := h.SyncHooks(); err != nil {
		t.Fatal(err)
	}
	if len(*targetCalls) != 0 {
		t.Fatalf("expected no calls, got %+v", *targetCalls)
	}
}

func TestCursorSyncHooksIsNoop(t *testing.T) {
	h := NewCursor("/repo")
	targetCalls, _ := stubSync(t)

	if err := h.SyncHooks(); err != nil {
		t.Fatal(err)
	}
	if len(*targetCalls) != 0 {
		t.Fatalf("expected no calls, got %+v", *targetCalls)
	}
}

func TestOpenCodeSyncHooksIsNoop(t *testing.T) {
	h := NewOpenCode("/repo")
	targetCalls, _ := stubSync(t)

	if err := h.SyncHooks(); err != nil {
		t.Fatal(err)
	}
	if len(*targetCalls) != 0 {
		t.Fatalf("expected no calls, got %+v", *targetCalls)
	}
}

func TestOpenCodeSyncTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := "/repo"
	h := NewOpenCode(repo)
	home := h.Home

	targetCalls, skillsCalls := stubSync(t)

	if err := h.SyncSkills(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncSubagents(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncConfig(); err != nil {
		t.Fatal(err)
	}
	if err := h.SyncRules(); err != nil {
		t.Fatal(err)
	}

	wantSkills := skillsCall{
		dest:    filepath.Join(home, "skills"),
		sources: []string{filepath.Join(repo, "common", "skills"), filepath.Join(repo, "opencode", "skills")},
	}
	if len(*skillsCalls) != 1 || !reflect.DeepEqual((*skillsCalls)[0], wantSkills) {
		t.Fatalf("skills call = %+v, want %+v", *skillsCalls, wantSkills)
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "common", "agents"), filepath.Join(home, "agents")) {
		t.Fatal("missing subagents call")
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "opencode", "opencode.jsonc"), filepath.Join(home, "opencode.jsonc")) {
		t.Fatal("missing config call")
	}
	if !containsTarget(*targetCalls, filepath.Join(repo, "common", "AGENTS.md"), filepath.Join(home, "AGENTS.md")) {
		t.Fatal("missing rules call")
	}
	if len(*targetCalls) != 3 {
		t.Fatalf("targetCalls count = %d, want 3", len(*targetCalls))
	}
}

func TestSyncCallsAllSixMethodsInOrder(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	mustMkdirAll(t, filepath.Join(repoRoot, "common", "extra-dir"))
	h := NewClaude(repoRoot)

	var order []string
	origTarget := syncTarget
	origSkills := syncFlattenedSkills
	syncFlattenedSkills = func(dest string, sources ...string) error {
		order = append(order, "skills")
		return nil
	}
	subagentsDest := filepath.Join(h.Home, "agents")
	extrasDest := filepath.Join(h.Home, "extra-dir")
	syncTarget = func(src, dest string) error {
		switch dest {
		case subagentsDest:
			order = append(order, "subagents")
		case h.ConfigDest:
			order = append(order, "config")
		case h.RulesDest:
			order = append(order, "rules")
		case h.HooksDest:
			order = append(order, "hooks")
		case extrasDest:
			order = append(order, "extras")
		}
		return nil
	}
	t.Cleanup(func() {
		syncTarget = origTarget
		syncFlattenedSkills = origSkills
	})

	if err := h.Sync(); err != nil {
		t.Fatal(err)
	}

	want := []string{"skills", "subagents", "config", "rules", "hooks", "extras"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("call order = %v, want %v", order, want)
	}
}

func TestSyncRulesIsNoopWhenEmpty(t *testing.T) {
	h := &Harness{Name: "stub", RepoRoot: "/repo", Home: "/home"}
	targetCalls, _ := stubSync(t)

	if err := h.SyncRules(); err != nil {
		t.Fatal(err)
	}
	if len(*targetCalls) != 0 {
		t.Fatalf("expected no calls, got %+v", *targetCalls)
	}
}

func TestSyncFailFastAtEachStep(t *testing.T) {
	wantErr := os.ErrInvalid

	newFailingSkills := func() { syncFlattenedSkills = func(dest string, sources ...string) error { return wantErr } }
	newOKSkills := func() { syncFlattenedSkills = func(dest string, sources ...string) error { return nil } }

	t.Run("skills", func(t *testing.T) {
		h := NewClaude("/repo")
		origTarget, origSkills := syncTarget, syncFlattenedSkills
		newFailingSkills()
		syncTarget = func(src, dest string) error { t.Fatal("syncTarget should not be called"); return nil }
		t.Cleanup(func() { syncTarget, syncFlattenedSkills = origTarget, origSkills })

		if err := h.Sync(); err != wantErr {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})

	t.Run("config", func(t *testing.T) {
		h := NewClaude("/repo")
		origTarget, origSkills := syncTarget, syncFlattenedSkills
		newOKSkills()
		syncTarget = func(src, dest string) error {
			if dest == h.ConfigDest {
				return wantErr
			}
			return nil
		}
		t.Cleanup(func() { syncTarget, syncFlattenedSkills = origTarget, origSkills })

		if err := h.Sync(); err != wantErr {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})

	t.Run("rules", func(t *testing.T) {
		h := NewClaude("/repo")
		origTarget, origSkills := syncTarget, syncFlattenedSkills
		newOKSkills()
		syncTarget = func(src, dest string) error {
			if dest == h.RulesDest {
				return wantErr
			}
			return nil
		}
		t.Cleanup(func() { syncTarget, syncFlattenedSkills = origTarget, origSkills })

		if err := h.Sync(); err != wantErr {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})

	t.Run("hooks", func(t *testing.T) {
		h := NewClaude("/repo")
		origTarget, origSkills := syncTarget, syncFlattenedSkills
		newOKSkills()
		syncTarget = func(src, dest string) error {
			if dest == h.HooksDest {
				return wantErr
			}
			return nil
		}
		t.Cleanup(func() { syncTarget, syncFlattenedSkills = origTarget, origSkills })

		if err := h.Sync(); err != wantErr {
			t.Fatalf("err = %v, want %v", err, wantErr)
		}
	})
}

func TestUserHomeDirPanicsWhenUnresolvable(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when home directory cannot be resolved")
		}
	}()
	userHomeDir()
}

func TestSyncFailFastOnError(t *testing.T) {
	h := NewClaude("/repo")

	wantErr := os.ErrInvalid
	origTarget := syncTarget
	origSkills := syncFlattenedSkills
	syncFlattenedSkills = func(dest string, sources ...string) error { return nil }
	callCount := 0
	syncTarget = func(src, dest string) error {
		callCount++
		return wantErr
	}
	t.Cleanup(func() {
		syncTarget = origTarget
		syncFlattenedSkills = origSkills
	})

	if err := h.Sync(); err != wantErr {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if callCount != 1 {
		t.Fatalf("expected fail-fast after first error, got %d calls", callCount)
	}
}

func TestPathHelpers(t *testing.T) {
	h := &Harness{Name: "claude", RepoRoot: "/repo"}
	if got, want := h.RepoDir(), filepath.Join("/repo", "claude"); got != want {
		t.Errorf("RepoDir() = %q, want %q", got, want)
	}
	if got, want := h.CommonDir(), filepath.Join("/repo", "common"); got != want {
		t.Errorf("CommonDir() = %q, want %q", got, want)
	}
	if got, want := h.SkillsDir(), filepath.Join("/repo", "claude", "skills"); got != want {
		t.Errorf("SkillsDir() = %q, want %q", got, want)
	}
}

func TestConstructorsUseHomeDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	repo := "/repo"
	cases := []struct {
		name string
		h    *Harness
		want string
	}{
		{"claude", NewClaude(repo), filepath.Join(tmpHome, ".claude")},
		{"codex", NewCodex(repo), filepath.Join(tmpHome, ".codex")},
		{"cursor", NewCursor(repo), filepath.Join(tmpHome, ".cursor")},
		{"opencode", NewOpenCode(repo), filepath.Join(tmpHome, ".config", "opencode")},
		{"hermes", NewHermes(repo), filepath.Join(tmpHome, ".hermes")},
	}
	for _, c := range cases {
		if c.h.Home != c.want {
			t.Errorf("%s Home = %q, want %q", c.name, c.h.Home, c.want)
		}
	}
}

func TestAllHarnesses(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := "/repo"
	got := AllHarnesses(repo)
	if len(got) != 5 {
		t.Fatalf("len = %d, want 5", len(got))
	}
	wantNames := []string{"claude", "codex", "cursor", "opencode", "hermes"}
	for i, name := range wantNames {
		if got[i].Name != name {
			t.Errorf("AllHarnesses()[%d].Name = %q, want %q", i, got[i].Name, name)
		}
	}
}

func TestSyncSkillsFlattenedWhenGroupsNotPreserved(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := "/repo"
	h := NewClaude(repo)
	home := h.Home

	_, flatCalls := stubSync(t)
	groupedCalls := stubPartiallyGroupedSkills(t)

	if err := h.SyncSkills(); err != nil {
		t.Fatal(err)
	}

	wantSkills := skillsCall{
		dest:    filepath.Join(home, "skills"),
		sources: []string{filepath.Join(repo, "common", "skills"), filepath.Join(repo, "claude", "skills")},
	}
	if len(*flatCalls) != 1 || !reflect.DeepEqual((*flatCalls)[0], wantSkills) {
		t.Fatalf("flattened skills call = %+v, want %+v", *flatCalls, wantSkills)
	}
	if len(*groupedCalls) != 0 {
		t.Fatalf("expected no grouped skills calls, got %+v", *groupedCalls)
	}
}

func TestSyncSkillsGroupedWhenPreserved(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := "/repo"
	h := NewHermes(repo)
	home := h.Home

	_, flatCalls := stubSync(t)
	groupedCalls := stubPartiallyGroupedSkills(t)

	if err := h.SyncSkills(); err != nil {
		t.Fatal(err)
	}

	wantSkills := skillsCall{
		dest:    filepath.Join(home, "skills"),
		sources: []string{filepath.Join(repo, "common", "skills"), filepath.Join(repo, "hermes", "skills")},
	}
	if len(*groupedCalls) != 1 || !reflect.DeepEqual((*groupedCalls)[0], wantSkills) {
		t.Fatalf("grouped skills call = %+v, want %+v", *groupedCalls, wantSkills)
	}
	if len(*flatCalls) != 0 {
		t.Fatalf("expected no flattened skills calls, got %+v", *flatCalls)
	}
}

func TestNewHermesConstructor(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	repo := "/repo"

	h := NewHermes(repo)
	home := filepath.Join(tmpHome, ".hermes")

	if h.Home != home {
		t.Errorf("Home = %q, want %q", h.Home, home)
	}
	if want := filepath.Join(repo, "hermes", "config.json"); h.ConfigSrc != want {
		t.Errorf("ConfigSrc = %q, want %q", h.ConfigSrc, want)
	}
	if want := filepath.Join(home, "config.json"); h.ConfigDest != want {
		t.Errorf("ConfigDest = %q, want %q", h.ConfigDest, want)
	}
	if want := filepath.Join(repo, "common", "AGENTS.md"); h.RulesSrc != want {
		t.Errorf("RulesSrc = %q, want %q", h.RulesSrc, want)
	}
	if want := filepath.Join(home, "AGENTS.md"); h.RulesDest != want {
		t.Errorf("RulesDest = %q, want %q", h.RulesDest, want)
	}
	if h.HooksSrc != "" {
		t.Errorf("HooksSrc = %q, want empty", h.HooksSrc)
	}
	if h.HooksDest != "" {
		t.Errorf("HooksDest = %q, want empty", h.HooksDest)
	}
	if !h.PreserveSkillGroups {
		t.Error("PreserveSkillGroups = false, want true")
	}
}

// TestSyncEndToEndFilesystem exercises the real syncengine implementation
// (no stubs) against a temp repo/home, matching this repo's preference for
// real filesystem fixtures over heavy mocking.
func TestSyncEndToEndFilesystem(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("HOME", t.TempDir())

	mustWriteFile(t, filepath.Join(repoRoot, "claude", "settings.json"), "{}")
	mustWriteFile(t, filepath.Join(repoRoot, "claude", "CLAUDE.md"), "# claude")
	mustMkdirAll(t, filepath.Join(repoRoot, "claude", "hooks"))
	mustMkdirAll(t, filepath.Join(repoRoot, "common", "agents"))
	mustMkdirAll(t, filepath.Join(repoRoot, "common", "skills", "my-skill"))
	mustWriteFile(t, filepath.Join(repoRoot, "common", "skills", "my-skill", "SKILL.md"), "# skill")

	h := NewClaude(repoRoot)
	home := h.Home

	if err := h.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if _, err := os.Lstat(filepath.Join(home, "settings.json")); err != nil {
		t.Errorf("settings.json not synced: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(home, "CLAUDE.md")); err != nil {
		t.Errorf("CLAUDE.md not synced: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(home, "hooks")); err != nil {
		t.Errorf("hooks not synced: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(home, "agents")); err != nil {
		t.Errorf("agents not synced: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(home, "skills", "my-skill")); err != nil {
		t.Errorf("skills not flattened/synced: %v", err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestSyncExtrasSymlinksUnknownTopLevelEntries(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	mustWriteFile(t, filepath.Join(repoRoot, "common", "notes.txt"), "shared notes")
	mustMkdirAll(t, filepath.Join(repoRoot, "claude", "extra-dir"))
	mustWriteFile(t, filepath.Join(repoRoot, "claude", "extra-dir", "file.txt"), "content")

	h := NewClaude(repoRoot)
	if err := h.SyncExtras(); err != nil {
		t.Fatal(err)
	}

	notesLink := filepath.Join(h.Home, "notes.txt")
	if info, err := os.Lstat(notesLink); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected %s to be a symlink, err=%v", notesLink, err)
	}
	dirLink := filepath.Join(h.Home, "extra-dir")
	if info, err := os.Lstat(dirLink); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected %s to be a symlink, err=%v", dirLink, err)
	}
	if _, err := os.Stat(filepath.Join(dirLink, "file.txt")); err != nil {
		t.Fatalf("expected extra-dir symlink to resolve to real content: %v", err)
	}
}

func TestSyncExtrasSkipsKnownNames(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	mustMkdirAll(t, filepath.Join(repoRoot, "common", "skills"))
	mustMkdirAll(t, filepath.Join(repoRoot, "common", "agents"))
	mustWriteFile(t, filepath.Join(repoRoot, "claude", "settings.json"), "{}")
	mustWriteFile(t, filepath.Join(repoRoot, "claude", "CLAUDE.md"), "# claude")
	mustMkdirAll(t, filepath.Join(repoRoot, "claude", "hooks"))

	h := NewClaude(repoRoot)
	targetCalls, _ := stubSync(t)

	if err := h.SyncExtras(); err != nil {
		t.Fatal(err)
	}
	if len(*targetCalls) != 0 {
		t.Fatalf("expected no extras calls for known names, got %+v", *targetCalls)
	}
}

func TestSyncExtrasHarnessSpecificOverridesCommon(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	mustWriteFile(t, filepath.Join(repoRoot, "common", "shared.txt"), "common version")
	mustWriteFile(t, filepath.Join(repoRoot, "claude", "shared.txt"), "claude version")

	h := NewClaude(repoRoot)
	if err := h.SyncExtras(); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(h.Home, "shared.txt")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(repoRoot, "claude", "shared.txt")
	if target != want {
		t.Fatalf("symlink target = %q, want %q (harness-specific should win)", target, want)
	}
}

func TestSyncExtrasNoopWhenNeitherDirExists(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	h := NewClaude(filepath.Join(t.TempDir(), "missing-repo"))
	targetCalls, _ := stubSync(t)

	if err := h.SyncExtras(); err != nil {
		t.Fatal(err)
	}
	if len(*targetCalls) != 0 {
		t.Fatalf("expected no calls, got %+v", *targetCalls)
	}
}

func TestSyncExtrasErrorsWhenCommonDirUnreadable(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	commonDir := filepath.Join(repoRoot, "common")
	mustMkdirAll(t, commonDir)
	if err := os.Chmod(commonDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(commonDir, 0o755) })

	h := NewClaude(repoRoot)
	if err := h.SyncExtras(); err == nil {
		t.Fatal("expected error when common dir cannot be read")
	}
}

func TestSyncExtrasErrorsWhenRepoDirUnreadable(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	repoDir := filepath.Join(repoRoot, "claude")
	mustMkdirAll(t, repoDir)
	if err := os.Chmod(repoDir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(repoDir, 0o755) })

	h := NewClaude(repoRoot)
	if err := h.SyncExtras(); err == nil {
		t.Fatal("expected error when repo dir cannot be read")
	}
}

func TestSyncExtrasErrorsWhenSyncTargetFails(t *testing.T) {
	repoRoot := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	mustWriteFile(t, filepath.Join(repoRoot, "common", "notes.txt"), "content")

	h := NewClaude(repoRoot)
	origTarget := syncTarget
	syncTarget = func(src, dest string) error { return os.ErrInvalid }
	t.Cleanup(func() { syncTarget = origTarget })

	if err := h.SyncExtras(); err != os.ErrInvalid {
		t.Fatalf("err = %v, want %v", err, os.ErrInvalid)
	}
}
