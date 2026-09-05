package syncengine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withStdin(t *testing.T, input string) {
	t.Helper()
	old := Stdin
	Stdin = strings.NewReader(input)
	t.Cleanup(func() { Stdin = old })
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// --- approve ---

func TestApproveTrueWhenInputIsY(t *testing.T) {
	withStdin(t, "y\n")
	if !approve("Proceed") {
		t.Fatal("expected true for y")
	}
}

func TestApproveFalseForNonYResponse(t *testing.T) {
	withStdin(t, "n\n")
	if approve("Proceed") {
		t.Fatal("expected false for n")
	}
}

func TestApproveFalseWhenNoInput(t *testing.T) {
	withStdin(t, "")
	if approve("Proceed") {
		t.Fatal("expected false when scanner has nothing to read")
	}
}

func TestApproveTrimsAndLowercases(t *testing.T) {
	withStdin(t, "  Y  \n")
	if !approve("Proceed") {
		t.Fatal("expected true for padded/uppercase Y")
	}
}

// --- move ---

func TestMoveNoopWhenSourceIsSymlink(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	link := filepath.Join(tmp, "link")
	if err := os.Symlink(src, link); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(tmp, "dest")

	if err := move(link, dest); err != nil {
		t.Fatal(err)
	}
	if exists(dest) {
		t.Fatal("dest should not exist")
	}
}

func TestMoveNoopWhenDestinationExists(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)

	if err := move(src, dest); err != nil {
		t.Fatal(err)
	}
}

func TestMoveNoopWhenDestinationIsExistingFile(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	mustWriteFile(t, src, "source")
	dest := filepath.Join(tmp, "dest.txt")
	mustWriteFile(t, dest, "dest")

	if err := move(src, dest); err != nil {
		t.Fatal(err)
	}
	if mustReadFile(t, dest) != "dest" {
		t.Fatal("dest should be untouched")
	}
	if mustReadFile(t, src) != "source" {
		t.Fatal("src should be untouched")
	}
}

func TestMoveCallsMoveWhenApproved(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	mustWriteFile(t, filepath.Join(src, "f.txt"), "content")
	dest := filepath.Join(tmp, "nested", "dest")

	withStdin(t, "y\n")
	if err := move(src, dest); err != nil {
		t.Fatal(err)
	}
	if exists(src) {
		t.Fatal("src should be gone after move")
	}
	if mustReadFile(t, filepath.Join(dest, "f.txt")) != "content" {
		t.Fatal("dest should contain moved content")
	}
}

func TestMoveFileCallsMoveWhenApproved(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	mustWriteFile(t, src, "content")
	dest := filepath.Join(tmp, "nested", "dest.txt")

	withStdin(t, "y\n")
	if err := move(src, dest); err != nil {
		t.Fatal(err)
	}
	if exists(src) {
		t.Fatal("src should be gone")
	}
	if mustReadFile(t, dest) != "content" {
		t.Fatal("dest should have content")
	}
}

func TestMoveCallsCopyWhenNotApproved(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	mustWriteFile(t, filepath.Join(src, "f.txt"), "content")
	dest := filepath.Join(tmp, "dest")

	withStdin(t, "n\n")
	if err := move(src, dest); err != nil {
		t.Fatal(err)
	}
	if !exists(src) {
		t.Fatal("src should remain after decline")
	}
	if mustReadFile(t, filepath.Join(dest, "f.txt")) != "content" {
		t.Fatal("dest should be a copy")
	}
}

func TestMoveFileCallsCopyWhenNotApproved(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	mustWriteFile(t, src, "content")
	dest := filepath.Join(tmp, "dest.txt")

	withStdin(t, "n\n")
	if err := move(src, dest); err != nil {
		t.Fatal(err)
	}
	if !exists(src) {
		t.Fatal("src should remain")
	}
	if mustReadFile(t, src) != "content" || mustReadFile(t, dest) != "content" {
		t.Fatal("both src and dest should have content")
	}
}

// --- symlink ---

func TestSymlinkNoopWhenSourceMissing(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "missing")
	dest := filepath.Join(tmp, "dest")

	if err := symlink(src, dest); err != nil {
		t.Fatal(err)
	}
	if exists(dest) {
		t.Fatal("dest should not exist")
	}
}

func TestSymlinkNoopWhenDestinationAlreadyCorrect(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")
	if err := os.Symlink(src, dest); err != nil {
		t.Fatal(err)
	}

	if err := symlink(src, dest); err != nil {
		t.Fatal(err)
	}
}

func TestSymlinkCreatesLinkWhenValid(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")

	if err := symlink(src, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) {
		t.Fatal("dest should be a symlink")
	}
	if resolve(dest) != resolve(src) {
		t.Fatal("dest should resolve to src")
	}
}

func TestSymlinkReplacesRealDirWhenApproved(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	mustWriteFile(t, filepath.Join(dest, "existing"), "")

	withStdin(t, "y\n")
	if err := symlink(src, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) || resolve(dest) != resolve(src) {
		t.Fatal("dest should be a symlink to src")
	}
}

func TestSymlinkReplacesRealFileWhenApproved(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")
	mustWriteFile(t, dest, "real file")

	withStdin(t, "y\n")
	if err := symlink(src, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) || resolve(dest) != resolve(src) {
		t.Fatal("dest should be a symlink to src")
	}
}

func TestSymlinkReplacesStaleSymlinkToDirWhenApproved(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	oldTarget := filepath.Join(tmp, "old_target")
	mustMkdir(t, oldTarget)
	dest := filepath.Join(tmp, "dest")
	if err := os.Symlink(oldTarget, dest); err != nil {
		t.Fatal(err)
	}

	withStdin(t, "y\n")
	if err := symlink(src, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) || resolve(dest) != resolve(src) {
		t.Fatal("dest should now point at src")
	}
	if !isDir(oldTarget) {
		t.Fatal("old target should be untouched")
	}
}

func TestSymlinkLeavesRealDestWhenDeclined(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)

	withStdin(t, "n\n")
	if err := symlink(src, dest); err != nil {
		t.Fatal(err)
	}
	if isSymlink(dest) || !isDir(dest) {
		t.Fatal("dest should remain a real dir")
	}
}

func TestSymlinkLeavesStaleSymlinkWhenDeclined(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	oldTarget := filepath.Join(tmp, "old_target")
	mustMkdir(t, oldTarget)
	dest := filepath.Join(tmp, "dest")
	if err := os.Symlink(oldTarget, dest); err != nil {
		t.Fatal(err)
	}

	withStdin(t, "n\n")
	if err := symlink(src, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) || resolve(dest) != resolve(oldTarget) {
		t.Fatal("dest should still point at old target")
	}
}

// --- absorb ---

func TestAbsorbNoopWhenDestMissing(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")

	ok, err := absorb(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected true when dest missing")
	}
}

func TestAbsorbNoopWhenDestIsSymlink(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	real := filepath.Join(tmp, "real")
	mustMkdir(t, real)
	mustWriteFile(t, filepath.Join(real, "leaf.txt"), "leaf")
	dest := filepath.Join(tmp, "dest")
	if err := os.Symlink(real, dest); err != nil {
		t.Fatal(err)
	}

	ok, err := absorb(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected true when dest is symlink")
	}
	if !exists(filepath.Join(real, "leaf.txt")) {
		t.Fatal("real target should be untouched")
	}
}

func TestAbsorbBothFilesIsNoopAndReportsFullyAbsorbed(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	mustWriteFile(t, src, "src content")
	dest := filepath.Join(tmp, "dest.txt")
	mustWriteFile(t, dest, "dest content")

	ok, err := absorb(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected true when dest is not a dir")
	}
}

func TestAbsorbMovesMissingChildrenOnly(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	mustWriteFile(t, filepath.Join(src, "shared.txt"), "in src")

	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	mustWriteFile(t, filepath.Join(dest, "shared.txt"), "in dest")
	mustWriteFile(t, filepath.Join(dest, "unique.txt"), "only in dest")

	withStdin(t, "y\n")
	ok, err := absorb(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected fully absorbed")
	}
	if mustReadFile(t, filepath.Join(src, "unique.txt")) != "only in dest" {
		t.Fatal("unique.txt should have moved into src")
	}
	if mustReadFile(t, filepath.Join(src, "shared.txt")) != "in src" {
		t.Fatal("shared.txt in src should be untouched")
	}
	if exists(filepath.Join(dest, "unique.txt")) {
		t.Fatal("unique.txt should be gone from dest")
	}
	if !exists(filepath.Join(dest, "shared.txt")) {
		t.Fatal("shared.txt should remain in dest (skipped)")
	}
}

func TestAbsorbIsOneLevelOnly(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)

	dest := filepath.Join(tmp, "dest")
	nested := filepath.Join(dest, "child", "grandchild")
	mustMkdir(t, nested)
	mustWriteFile(t, filepath.Join(nested, "leaf.txt"), "leaf")

	withStdin(t, "y\n")
	ok, err := absorb(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected fully absorbed")
	}
	if mustReadFile(t, filepath.Join(src, "child", "grandchild", "leaf.txt")) != "leaf" {
		t.Fatal("nested subtree should move as a unit")
	}
	if exists(filepath.Join(dest, "child")) {
		t.Fatal("child should be gone from dest")
	}
}

func TestAbsorbDeclinedChildStaysInDestAndReportsIncomplete(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)

	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	mustWriteFile(t, filepath.Join(dest, "unique.txt"), "only in dest")

	withStdin(t, "n\n")
	ok, err := absorb(src, dest)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected false when child move declined")
	}
	if mustReadFile(t, filepath.Join(dest, "unique.txt")) != "only in dest" {
		t.Fatal("dest content should remain")
	}
	if mustReadFile(t, filepath.Join(src, "unique.txt")) != "only in dest" {
		t.Fatal("src should have received a copy")
	}
}

// --- syncTarget ---

func TestSyncTargetNoopWhenDestIsSymlink(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")
	if err := os.Symlink(src, dest); err != nil {
		t.Fatal(err)
	}

	if err := syncTarget(src, dest); err != nil {
		t.Fatal(err)
	}
	if resolve(dest) != resolve(src) {
		t.Fatal("dest should still point at src")
	}
}

func TestSyncTargetMigratesWhenSrcMissing(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	mustWriteFile(t, filepath.Join(dest, "file.txt"), "content")

	withStdin(t, "y\n")
	if err := syncTarget(src, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) || resolve(dest) != resolve(src) {
		t.Fatal("dest should be a symlink to src")
	}
	if mustReadFile(t, filepath.Join(src, "file.txt")) != "content" {
		t.Fatal("content should have migrated into src")
	}
}

func TestSyncTargetAbsorbsThenSymlinksWhenBothExist(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	mustWriteFile(t, filepath.Join(src, "shared.txt"), "in src")

	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	mustWriteFile(t, filepath.Join(dest, "unique.txt"), "only in dest")

	withStdin(t, "y\ny\n")
	if err := syncTarget(src, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) || resolve(dest) != resolve(src) {
		t.Fatal("dest should be a symlink to src")
	}
	if mustReadFile(t, filepath.Join(src, "unique.txt")) != "only in dest" {
		t.Fatal("unique.txt should be absorbed into src")
	}
}

func TestSyncTargetBothPlainFilesDoesNotRaise(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "config.toml")
	mustWriteFile(t, src, "in repo")
	dest := filepath.Join(tmp, "config_home.toml")
	mustWriteFile(t, dest, "in home")

	withStdin(t, "y\n")
	if err := syncTarget(src, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) || resolve(dest) != resolve(src) {
		t.Fatal("dest should be a symlink to src")
	}
	if mustReadFile(t, src) != "in repo" {
		t.Fatal("src content should be untouched")
	}
}

func TestSyncTargetDeclinedAbsorbChildIsNotDestroyedBySymlink(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)

	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	mustWriteFile(t, filepath.Join(dest, "unique.txt"), "only in dest")

	// Decline the absorb prompt; any later prompt would be approved.
	withStdin(t, "n\ny\ny\n")
	if err := syncTarget(src, dest); err != nil {
		t.Fatal(err)
	}
	if isSymlink(dest) {
		t.Fatal("dest must not become a symlink when absorb is incomplete")
	}
	if !isDir(dest) {
		t.Fatal("dest should remain a real dir")
	}
	if mustReadFile(t, filepath.Join(dest, "unique.txt")) != "only in dest" {
		t.Fatal("unique.txt should remain in dest")
	}
}

func TestSyncTargetSymlinksDirectlyWhenDestMissing(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")

	if err := syncTarget(src, dest); err != nil {
		t.Fatal(err)
	}
	if !isSymlink(dest) || resolve(dest) != resolve(src) {
		t.Fatal("dest should be a symlink to src")
	}
}

func TestSyncTargetNoopWhenNeitherExists(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	dest := filepath.Join(tmp, "dest")

	if err := syncTarget(src, dest); err != nil {
		t.Fatal(err)
	}
	if exists(dest) || exists(src) {
		t.Fatal("neither path should exist")
	}
}

// --- ResolveRepoRoot ---

func TestResolveRepoRootResolvesExistingPath(t *testing.T) {
	tmp := t.TempDir()
	got, err := ResolveRepoRoot(tmp)
	if err != nil {
		t.Fatal(err)
	}
	want := resolve(tmp)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveRepoRootFallsBackWhenPathMissing(t *testing.T) {
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "does-not-exist")
	got, err := ResolveRepoRoot(missing)
	if err != nil {
		t.Fatal(err)
	}
	if got != missing {
		t.Fatalf("got %q want %q", got, missing)
	}
}

// --- small primitive helpers ---

func TestIsDirFalseWhenPathMissing(t *testing.T) {
	tmp := t.TempDir()
	if isDir(filepath.Join(tmp, "missing")) {
		t.Fatal("expected false for a missing path")
	}
}

func TestExistsFalseWhenPathMissing(t *testing.T) {
	tmp := t.TempDir()
	if exists(filepath.Join(tmp, "missing")) {
		t.Fatal("expected false for a missing path")
	}
}

func TestIsSymlinkFalseWhenPathMissing(t *testing.T) {
	tmp := t.TempDir()
	if isSymlink(filepath.Join(tmp, "missing")) {
		t.Fatal("expected false for a missing path")
	}
}

func TestResolveFallsBackWhenPathMissing(t *testing.T) {
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "missing")
	if got := resolve(missing); got != missing {
		t.Fatalf("got %q want %q", got, missing)
	}
}

func withFailingGetwd(t *testing.T) {
	t.Helper()
	old := getwd
	getwd = func() (string, error) {
		return "", errors.New("injected getwd failure")
	}
	t.Cleanup(func() { getwd = old })
}

func TestResolveFallsBackWhenGetwdFails(t *testing.T) {
	withFailingGetwd(t)
	if got := resolve("relative"); got != "relative" {
		t.Fatalf("got %q, want the original relative path unchanged", got)
	}
}

func TestResolveRepoRootErrorsWhenGetwdFails(t *testing.T) {
	withFailingGetwd(t)
	if _, err := ResolveRepoRoot("relative"); err == nil {
		t.Fatal("expected error when getwd fails")
	}
}

func TestReadLineReturnsFalseOnImmediateEOF(t *testing.T) {
	_, ok := readLine(strings.NewReader(""))
	if ok {
		t.Fatal("expected false on immediate EOF")
	}
}

func TestReadLineReturnsTrueOnUnterminatedFinalLine(t *testing.T) {
	line, ok := readLine(strings.NewReader("y"))
	if !ok || line != "y" {
		t.Fatalf("got line=%q ok=%v, want y/true", line, ok)
	}
}

// --- copyPath / copyDir / copyFile / movePath ---

func TestCopyPathErrorsWhenSourceMissing(t *testing.T) {
	tmp := t.TempDir()
	err := copyPath(filepath.Join(tmp, "missing"), filepath.Join(tmp, "dest"))
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestCopyPathCopiesDirectoryRecursively(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, filepath.Join(src, "nested"))
	mustWriteFile(t, filepath.Join(src, "nested", "leaf.txt"), "leaf")
	dest := filepath.Join(tmp, "dest")

	if err := copyPath(src, dest); err != nil {
		t.Fatal(err)
	}
	if mustReadFile(t, filepath.Join(dest, "nested", "leaf.txt")) != "leaf" {
		t.Fatal("expected recursive copy")
	}
	if !exists(filepath.Join(src, "nested", "leaf.txt")) {
		t.Fatal("source should remain in place")
	}
}

func TestCopyDirErrorsWhenReadDirFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	mustWriteFile(t, src, "not a dir")
	info, err := os.Lstat(src)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(tmp, "dest")

	if err := copyDir(src, dest, info); err == nil {
		t.Fatal("expected error when reading a non-directory as a directory")
	}
}

func TestCopyFileErrorsWhenSourceMissing(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "missing.txt")
	info := fakeFileInfo{}
	if err := copyFile(src, filepath.Join(tmp, "dest.txt"), info); err == nil {
		t.Fatal("expected error opening a missing source file")
	}
}

func TestCopyFileErrorsWhenDestinationParentMissing(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	mustWriteFile(t, src, "content")
	info, err := os.Lstat(src)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(tmp, "no-such-dir", "dest.txt")

	if err := copyFile(src, dest, info); err == nil {
		t.Fatal("expected error creating destination under a missing parent")
	}
}

func TestMovePathFallsBackToCopyWhenRenameFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	mustWriteFile(t, filepath.Join(src, "f.txt"), "content")
	// Rename fails because the destination's parent does not exist; the
	// fallback copyDir creates it via MkdirAll, then the source is removed.
	dest := filepath.Join(tmp, "no-such-parent", "dest")

	if err := movePath(src, dest); err != nil {
		t.Fatal(err)
	}
	if exists(src) {
		t.Fatal("expected source removed after fallback move")
	}
	if mustReadFile(t, filepath.Join(dest, "f.txt")) != "content" {
		t.Fatal("expected content copied to destination")
	}
}

func TestMovePathErrorsWhenFallbackCopyFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "missing")
	dest := filepath.Join(tmp, "dest")

	if err := movePath(src, dest); err == nil {
		t.Fatal("expected error when both rename and fallback copy fail")
	}
}

func withReadOnlyDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

func withUnreadableDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
}

func TestCopyDirErrorsWhenMkdirFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	blocked := filepath.Join(tmp, "blocked")
	mustMkdir(t, blocked)
	withReadOnlyDir(t, blocked)
	dest := filepath.Join(blocked, "dest")

	info, err := os.Lstat(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := copyDir(src, dest, info); err == nil {
		t.Fatal("expected mkdir error under a read-only parent")
	}
}

func TestCopyDirErrorsWhenChildCopyFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	child := filepath.Join(src, "child")
	mustMkdir(t, child)
	withUnreadableDir(t, child)
	dest := filepath.Join(tmp, "dest")

	info, err := os.Lstat(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := copyDir(src, dest, info); err == nil {
		t.Fatal("expected error copying an unreadable child directory")
	}
}

func TestAbsorbErrorsWhenChildMoveFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "blocked", "src")
	blocked := filepath.Join(tmp, "blocked")
	mustMkdir(t, blocked)
	withReadOnlyDir(t, blocked)

	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	mustWriteFile(t, filepath.Join(dest, "unique.txt"), "content")

	if _, err := absorb(src, dest); err == nil {
		t.Fatal("expected error propagated from a failing child move")
	}
}

func TestMoveErrorsWhenMkdirParentFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	blocked := filepath.Join(tmp, "blocked")
	mustMkdir(t, blocked)
	withReadOnlyDir(t, blocked)
	dest := filepath.Join(blocked, "nested", "dest")

	if err := move(src, dest); err == nil {
		t.Fatal("expected mkdir error under a read-only parent")
	}
}

func TestSymlinkErrorsWhenMkdirParentFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	blocked := filepath.Join(tmp, "blocked")
	mustMkdir(t, blocked)
	withReadOnlyDir(t, blocked)
	dest := filepath.Join(blocked, "nested", "dest")

	if err := symlink(src, dest); err == nil {
		t.Fatal("expected mkdir error under a read-only parent")
	}
}

func TestSymlinkErrorsWhenRemovingSymlinkFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	oldTarget := filepath.Join(tmp, "old_target")
	mustMkdir(t, oldTarget)
	parent := filepath.Join(tmp, "parent")
	mustMkdir(t, parent)
	dest := filepath.Join(parent, "dest")
	if err := os.Symlink(oldTarget, dest); err != nil {
		t.Fatal(err)
	}
	withReadOnlyDir(t, parent)

	withStdin(t, "y\n")
	if err := symlink(src, dest); err == nil {
		t.Fatal("expected error removing symlink under a read-only parent")
	}
}

func TestSymlinkErrorsWhenRemovingRealDirFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	parent := filepath.Join(tmp, "parent")
	mustMkdir(t, parent)
	dest := filepath.Join(parent, "dest")
	mustMkdir(t, dest)
	withReadOnlyDir(t, parent)

	withStdin(t, "y\n")
	if err := symlink(src, dest); err == nil {
		t.Fatal("expected error removing real dir under a read-only parent")
	}
}

func TestSymlinkErrorsWhenRemovingRealFileFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	parent := filepath.Join(tmp, "parent")
	mustMkdir(t, parent)
	dest := filepath.Join(parent, "dest")
	mustWriteFile(t, dest, "real file")
	withReadOnlyDir(t, parent)

	withStdin(t, "y\n")
	if err := symlink(src, dest); err == nil {
		t.Fatal("expected error removing real file under a read-only parent")
	}
}

func TestSymlinkErrorsWhenSymlinkCallFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	blocked := filepath.Join(tmp, "blocked")
	mustMkdir(t, blocked)
	dest := filepath.Join(blocked, "dest")
	withReadOnlyDir(t, blocked)

	if err := symlink(src, dest); err == nil {
		t.Fatal("expected error creating symlink under a read-only parent")
	}
}

func TestAbsorbErrorsWhenSourceReadDirFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	withUnreadableDir(t, src)
	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	mustWriteFile(t, filepath.Join(dest, "unique.txt"), "content")

	if _, err := absorb(src, dest); err == nil {
		t.Fatal("expected error reading an unreadable source directory")
	}
}

func TestAbsorbErrorsWhenDestReadDirFails(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	withUnreadableDir(t, dest)

	if _, err := absorb(src, dest); err == nil {
		t.Fatal("expected error reading an unreadable dest directory")
	}
}

func TestSyncTargetPropagatesAbsorbError(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	mustMkdir(t, src)
	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	withUnreadableDir(t, dest)

	if err := syncTarget(src, dest); err == nil {
		t.Fatal("expected error propagated from absorb")
	}
}

func TestSyncTargetPropagatesMoveError(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "blocked", "src")
	blocked := filepath.Join(tmp, "blocked")
	mustMkdir(t, blocked)
	withReadOnlyDir(t, blocked)
	dest := filepath.Join(tmp, "dest")
	mustMkdir(t, dest)
	mustWriteFile(t, filepath.Join(dest, "file.txt"), "content")

	withStdin(t, "y\n")
	if err := syncTarget(src, dest); err == nil {
		t.Fatal("expected error propagated from move when migrating dest into a read-only src parent")
	}
}

type fakeFileInfo struct{ os.FileInfo }

func (fakeFileInfo) Mode() os.FileMode { return 0o644 }

func TestResolveRepoRootResolvesRelativePath(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	got, err := ResolveRepoRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	if got != resolve(tmp) {
		t.Fatalf("got %q want %q", got, resolve(tmp))
	}
}
