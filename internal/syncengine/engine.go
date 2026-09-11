package syncengine

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/phedayat/agent_synchronizer/internal/synclog"
)

// Stdin is the source read by approve; tests may swap it to simulate input.
var Stdin io.Reader = os.Stdin

var logger = synclog.New("sync_engine", true)

// readLine reads a single line from r one byte at a time, stopping at '\n'
// (exclusive) or EOF, mirroring Python's input() which reads exactly one
// line and no more from stdin. A buffered reader would over-consume bytes
// belonging to the next line, breaking successive approve() calls sharing
// one Stdin (as tests do to simulate multiple prompts).
func readLine(r io.Reader) (string, bool) {
	var buf strings.Builder
	b := make([]byte, 1)
	read := false
	for {
		n, err := r.Read(b)
		if n > 0 {
			read = true
			if b[0] == '\n' {
				return buf.String(), true
			}
			buf.WriteByte(b[0])
		}
		if err != nil {
			return buf.String(), read
		}
	}
}

// approve mirrors Python's _approve: prompts on stdout, reads one line from
// Stdin, and reports whether the trimmed, lowercased response is "y".
func approve(message string) bool {
	fmt.Print(message + " (y/n): ")
	line, ok := readLine(Stdin)
	if !ok {
		return false
	}
	response := strings.ToLower(strings.TrimSpace(line))
	return response == "y"
}

// exists reports whether a path exists, following symlinks.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isSymlink reports whether path exists and is itself a symlink.
func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// isDir reports whether path exists and is a directory, following symlinks.
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// getwd is a package-level seam over os.Getwd so tests can inject failures
// that filepath.Abs's internal (uninjectable) call to os.Getwd cannot.
var getwd = os.Getwd

// abs mirrors filepath.Abs, routing its relative-path case through the
// injectable getwd seam above.
func abs(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	wd, err := getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, path), nil
}

// resolve mirrors Python's Path.resolve(): an absolute, symlink-resolved
// path, falling back to the unresolved absolute path if resolution fails
// (e.g. the path does not exist).
func resolve(path string) string {
	absPath, err := abs(path)
	if err != nil {
		return path
	}
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return absPath
	}
	return resolved
}

// ResolveRepoRoot resolves path to an absolute, symlink-resolved form,
// matching Python's Path(repo_root).resolve() best-effort semantics.
func ResolveRepoRoot(path string) (string, error) {
	absPath, err := abs(path)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		return resolved, nil
	}
	return absPath, nil
}

// copyPath recursively copies src to dest, leaving src in place, mirroring
// Python's Path.copy().
func copyPath(src, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dest, info)
	}
	return copyFile(src, dest, info)
}

func copyDir(src, dest string, info fs.FileInfo) error {
	if err := os.MkdirAll(dest, info.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		childSrc := filepath.Join(src, entry.Name())
		childDest := filepath.Join(dest, entry.Name())
		if err := copyPath(childSrc, childDest); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dest string, info fs.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	_, err = io.Copy(out, in)
	return err
}

// movePath relocates src to dest via rename, falling back to copy-then-
// remove across filesystems, mirroring Python's Path.move()/shutil.move().
func movePath(src, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	}
	if err := copyPath(src, dest); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

// move mirrors sync_engine.py's move(src, dest).
func move(src, dest string) error {
	if isSymlink(src) {
		logger.Info(fmt.Sprintf("Source path %s is a symlink.", src))
		return nil
	}
	if exists(dest) {
		logger.Info(fmt.Sprintf("Destination path %s already exists.", dest))
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if approve(fmt.Sprintf("Do you want to move %s?", src)) {
		logger.Info(fmt.Sprintf("Moving %s to %s", src, dest))
		return movePath(src, dest)
	}
	logger.Info(fmt.Sprintf("Copying %s to %s", src, dest))
	return copyPath(src, dest)
}

// symlink mirrors sync_engine.py's symlink(src, dest).
func symlink(src, dest string) error {
	if !exists(src) {
		logger.Info(fmt.Sprintf("Source path %s does not exist.", src))
		return nil
	}
	if isSymlink(dest) && resolve(dest) == resolve(src) {
		logger.Info(fmt.Sprintf("Symlink already correct: %s -> %s", dest, src))
		return nil
	}
	if exists(dest) {
		if !approve(fmt.Sprintf("Replace existing %s with a symlink to %s?", dest, src)) {
			logger.Info(fmt.Sprintf("Leaving %s in place.", dest))
			return nil
		}
		switch {
		case isSymlink(dest):
			if err := os.Remove(dest); err != nil {
				return err
			}
		case isDir(dest):
			if err := os.RemoveAll(dest); err != nil {
				return err
			}
		default:
			if err := os.Remove(dest); err != nil {
				return err
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.Symlink(src, dest)
}

// absorb mirrors sync_engine.py's _absorb(src, dest).
func absorb(src, dest string) (bool, error) {
	if !exists(dest) || isSymlink(dest) || !isDir(dest) {
		return true, nil
	}

	existing := map[string]struct{}{}
	if exists(src) && isDir(src) {
		srcEntries, err := os.ReadDir(src)
		if err != nil {
			return false, err
		}
		for _, entry := range srcEntries {
			existing[entry.Name()] = struct{}{}
		}
	}

	destEntries, err := os.ReadDir(dest)
	if err != nil {
		return false, err
	}

	fullyAbsorbed := true
	for _, entry := range destEntries {
		if _, ok := existing[entry.Name()]; ok {
			continue
		}
		childPath := filepath.Join(dest, entry.Name())
		if err := move(childPath, filepath.Join(src, entry.Name())); err != nil {
			return false, err
		}
		if exists(childPath) {
			fullyAbsorbed = false
		}
	}
	return fullyAbsorbed, nil
}

// SyncTarget mirrors sync_engine.py's sync_target(src, dest).
func SyncTarget(src, dest string) error {
	if isSymlink(dest) {
		logger.Info(fmt.Sprintf("Destination path %s is already a symlink.", dest))
		return nil
	}
	if exists(dest) && !exists(src) {
		if err := move(dest, src); err != nil {
			return err
		}
		return symlink(src, dest)
	}
	if exists(dest) && exists(src) {
		ok, err := absorb(src, dest)
		if err != nil {
			return err
		}
		if ok {
			return symlink(src, dest)
		}
		logger.Info(fmt.Sprintf("Leaving %s in place: not fully absorbed into %s.", dest, src))
		return nil
	}
	return symlink(src, dest)
}
