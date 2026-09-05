package syncengine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mkSkill(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("skill"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func withApprovedStdin(t *testing.T) {
	t.Helper()
	old := Stdin
	Stdin = strings.NewReader(strings.Repeat("y\n", 50))
	t.Cleanup(func() { Stdin = old })
}

func TestIterSkillDirsMissingRootYieldsNothing(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	if got := iterSkillDirs(root); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestIterSkillDirsFindsTopLevelSkill(t *testing.T) {
	root := t.TempDir()
	mkSkill(t, root)
	got := iterSkillDirs(root)
	if len(got) != 1 || got[0] != root {
		t.Fatalf("expected [%s], got %v", root, got)
	}
}

func TestIterSkillDirsFindsGroupedSkill(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "group-a", "skill-a")
	b := filepath.Join(root, "group-b", "skill-b")
	mkSkill(t, a)
	mkSkill(t, b)
	got := iterSkillDirs(root)
	if len(got) != 2 || got[0] != a || got[1] != b {
		t.Fatalf("expected [%s %s], got %v", a, b, got)
	}
}

func TestIterSkillDirsFindsArbitrarilyNestedSkill(t *testing.T) {
	root := t.TempDir()
	deep := filepath.Join(root, "l1", "l2", "l3", "skill-deep")
	mkSkill(t, deep)
	got := iterSkillDirs(root)
	if len(got) != 1 || got[0] != deep {
		t.Fatalf("expected [%s], got %v", deep, got)
	}
}

func TestIterSkillDirsStopsDescendingOnceMatched(t *testing.T) {
	root := t.TempDir()
	mkSkill(t, root)
	// A subdirectory that would itself look like a skill if descended into.
	nested := filepath.Join(root, "nested-skill")
	mkSkill(t, nested)
	got := iterSkillDirs(root)
	if len(got) != 1 || got[0] != root {
		t.Fatalf("expected only [%s], got %v", root, got)
	}
}

func TestCollectSkillsMergesSourcesLastWins(t *testing.T) {
	src1 := t.TempDir()
	src2 := t.TempDir()
	shared1 := filepath.Join(src1, "shared")
	shared2 := filepath.Join(src2, "shared")
	only1 := filepath.Join(src1, "only1")
	mkSkill(t, shared1)
	mkSkill(t, shared2)
	mkSkill(t, only1)

	got := collectSkills(src1, src2)
	if len(got) != 2 {
		t.Fatalf("expected 2 skills, got %v", got)
	}
	if got["shared"] != shared2 {
		t.Fatalf("expected last source to win: got %s want %s", got["shared"], shared2)
	}
	if got["only1"] != only1 {
		t.Fatalf("expected %s, got %s", only1, got["only1"])
	}
}

func TestSyncFlattenedSkillsSymlinksEachSkillDirectlyUnderDest(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "skills")
	skillA := filepath.Join(src, "skill-a")
	mkSkill(t, skillA)

	if err := SyncFlattenedSkills(dest, src); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dest, "skill-a")
	if !isSymlink(link) {
		t.Fatalf("expected %s to be a symlink", link)
	}
	if resolve(link) != resolve(skillA) {
		t.Fatalf("expected symlink to point at %s", skillA)
	}
}

func TestSyncFlattenedSkillsHarnessSpecificOverridesCommon(t *testing.T) {
	common := t.TempDir()
	harness := t.TempDir()
	dest := filepath.Join(t.TempDir(), "skills")
	commonSkill := filepath.Join(common, "skill-x")
	harnessSkill := filepath.Join(harness, "skill-x")
	mkSkill(t, commonSkill)
	mkSkill(t, harnessSkill)

	if err := SyncFlattenedSkills(dest, common, harness); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dest, "skill-x")
	if resolve(link) != resolve(harnessSkill) {
		t.Fatalf("expected harness-specific skill to win, got target resolving to %s", link)
	}
}

func TestSyncFlattenedSkillsAbsorbsUnmanagedRealDir(t *testing.T) {
	withApprovedStdin(t)
	src := t.TempDir()
	dest := t.TempDir()
	mkSkill(t, filepath.Join(src, "skill-a"))

	unmanaged := filepath.Join(dest, "unmanaged-skill")
	if err := os.MkdirAll(unmanaged, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unmanaged, "SKILL.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SyncFlattenedSkills(dest, src); err != nil {
		t.Fatal(err)
	}

	movedTarget := filepath.Join(src, "unmanaged-skill")
	if !isDir(movedTarget) {
		t.Fatalf("expected unmanaged dir moved to %s", movedTarget)
	}
	link := filepath.Join(dest, "unmanaged-skill")
	if !isSymlink(link) {
		t.Fatalf("expected %s to become a symlink", link)
	}
	if resolve(link) != resolve(movedTarget) {
		t.Fatalf("expected symlink to point at %s", movedTarget)
	}
}

func TestSyncFlattenedSkillsLeavesCorrectSymlinkAlone(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()
	skillA := filepath.Join(src, "skill-a")
	mkSkill(t, skillA)

	if err := SyncFlattenedSkills(dest, src); err != nil {
		t.Fatal(err)
	}
	// Second call: dest/skill-a is already a correct symlink -> skipped in
	// the absorb loop, then symlink() itself no-ops since it's already correct.
	if err := SyncFlattenedSkills(dest, src); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dest, "skill-a")
	if !isSymlink(link) {
		t.Fatalf("expected %s to remain a symlink", link)
	}
	if resolve(link) != resolve(skillA) {
		t.Fatalf("expected symlink to still point at %s", skillA)
	}
}

func TestIterSkillDirsErrorsWhenReadDirFails(t *testing.T) {
	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	if err := os.MkdirAll(blocked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(blocked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(blocked, 0o755) })

	if got := iterSkillDirs(blocked); got != nil {
		t.Fatalf("expected nil on ReadDir error, got %v", got)
	}
}

func TestSyncFlattenedSkillsErrorsWhenRemovingStaleSymlinkFails(t *testing.T) {
	src := t.TempDir()
	mkSkill(t, filepath.Join(src, "skill-a"))

	parent := t.TempDir()
	dest := filepath.Join(parent, "skills-link")
	if err := os.Symlink(t.TempDir(), dest); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(parent, 0o755) })

	if err := SyncFlattenedSkills(dest, src); err == nil {
		t.Fatal("expected error when removing stale symlink fails")
	}
}

func TestSyncFlattenedSkillsErrorsWhenMkdirDestFails(t *testing.T) {
	src := t.TempDir()
	mkSkill(t, filepath.Join(src, "skill-a"))

	parent := t.TempDir()
	blockingFile := filepath.Join(parent, "blocked")
	if err := os.WriteFile(blockingFile, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(blockingFile, "skills")

	if err := SyncFlattenedSkills(dest, src); err == nil {
		t.Fatal("expected error when dest cannot be created")
	}
}

func TestSyncFlattenedSkillsErrorsWhenReadDirDestFails(t *testing.T) {
	src := t.TempDir()
	mkSkill(t, filepath.Join(src, "skill-a"))

	dest := t.TempDir()
	if err := os.Chmod(dest, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dest, 0o755) })

	if err := SyncFlattenedSkills(dest, src); err == nil {
		t.Fatal("expected error when dest cannot be read")
	}
}

func TestSyncFlattenedSkillsErrorsWhenMoveFails(t *testing.T) {
	withApprovedStdin(t)
	dest := t.TempDir()
	unmanaged := filepath.Join(dest, "unmanaged-skill")
	if err := os.MkdirAll(unmanaged, 0o755); err != nil {
		t.Fatal(err)
	}

	// A plain file where the last source would need to be a directory
	// makes move()'s parent MkdirAll fail.
	lastSource := filepath.Join(t.TempDir(), "blocked-source")
	if err := os.WriteFile(lastSource, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SyncFlattenedSkills(dest, lastSource); err == nil {
		t.Fatal("expected error when move fails")
	}
}

func TestSyncFlattenedSkillsErrorsWhenSymlinkFails(t *testing.T) {
	src := t.TempDir()
	mkSkill(t, filepath.Join(src, "skill-a"))

	dest := t.TempDir()
	if err := os.Chmod(dest, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dest, 0o755) })

	if err := SyncFlattenedSkills(dest, src); err == nil {
		t.Fatal("expected error when symlink creation fails")
	}
}

func TestSyncFlattenedSkillsSkipsUnmanagedSymlinkNotInDesired(t *testing.T) {
	src := t.TempDir()
	mkSkill(t, filepath.Join(src, "skill-a"))

	dest := t.TempDir()
	orphanTarget := t.TempDir()
	orphan := filepath.Join(dest, "orphan")
	if err := os.Symlink(orphanTarget, orphan); err != nil {
		t.Fatal(err)
	}

	if err := SyncFlattenedSkills(dest, src); err != nil {
		t.Fatal(err)
	}

	if !isSymlink(orphan) {
		t.Fatalf("expected orphan symlink to remain untouched")
	}
	if resolve(orphan) != resolve(orphanTarget) {
		t.Fatalf("expected orphan symlink target unchanged")
	}
}

func TestSyncFlattenedSkillsRebuildsStaleWholeDirSymlink(t *testing.T) {
	src := t.TempDir()
	staleTarget := t.TempDir()
	dest := filepath.Join(t.TempDir(), "skills-link")
	skillA := filepath.Join(src, "skill-a")
	mkSkill(t, skillA)

	if err := os.Symlink(staleTarget, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) {
		t.Fatalf("setup: expected %s to be a symlink", dest)
	}

	if err := SyncFlattenedSkills(dest, src); err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(dest)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("expected %s to become a real directory, still a symlink", dest)
	}
	link := filepath.Join(dest, "skill-a")
	if !isSymlink(link) || resolve(link) != resolve(skillA) {
		t.Fatalf("expected %s to symlink to %s", link, skillA)
	}
}
