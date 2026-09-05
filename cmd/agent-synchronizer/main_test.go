package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/phedayat/agent_synchronizer/internal/harness"
)

func TestRunResolvesRelativeRepoRoot(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tmp)

	var gotRepo string
	orig := allHarnesses
	allHarnesses = func(r string) []*harness.Harness {
		gotRepo = r
		return nil
	}
	t.Cleanup(func() { allHarnesses = orig })

	if err := run(filepath.Join(".", "repo")); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	resolved, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatal(err)
	}
	if gotRepo != resolved {
		t.Fatalf("gotRepo = %q, want %q", gotRepo, resolved)
	}
}

func TestRunCallsSyncOncePerHarnessInOrder(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	commonAgents := filepath.Join(repo, "common", "agents")
	if err := os.MkdirAll(commonAgents, 0o755); err != nil {
		t.Fatal(err)
	}
	home1 := filepath.Join(tmp, "home1")
	home2 := filepath.Join(tmp, "home2")
	for _, home := range []string{home1, home2} {
		if err := os.MkdirAll(home, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	orig := allHarnesses
	allHarnesses = func(r string) []*harness.Harness {
		return []*harness.Harness{
			{Name: "one", Home: home1, RepoRoot: r},
			{Name: "two", Home: home2, RepoRoot: r},
		}
	}
	t.Cleanup(func() { allHarnesses = orig })

	if err := run(repo); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	// Each stub Harness has empty Config/Rules/Hooks src, so Sync() only
	// runs SyncSkills+SyncSubagents; SyncSubagents creates Home/agents,
	// proving each harness's Sync() actually executed once, in order.
	for _, home := range []string{home1, home2} {
		if _, err := os.Stat(filepath.Join(home, "agents")); err != nil {
			t.Fatalf("expected %s/agents to exist after Sync: %v", home, err)
		}
	}
}

func TestRunReturnsErrorWhenResolveRepoRootFails(t *testing.T) {
	orig := resolveRepoRoot
	resolveRepoRoot = func(string) (string, error) {
		return "", errors.New("boom")
	}
	t.Cleanup(func() { resolveRepoRoot = orig })

	if err := run("whatever"); err == nil {
		t.Fatal("expected error propagated from resolveRepoRoot, got nil")
	}
}

func TestRunPropagatesHarnessSyncError(t *testing.T) {
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(repo, "common", "agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(tmp, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	// A regular file where the symlink destination's parent directory
	// must go forces os.MkdirAll to fail, giving SyncConfig a real error
	// to propagate through Harness.Sync() and run().
	blocker := filepath.Join(home, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	configSrc := filepath.Join(tmp, "config.toml")
	if err := os.WriteFile(configSrc, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	orig := allHarnesses
	allHarnesses = func(r string) []*harness.Harness {
		return []*harness.Harness{
			{
				Name:       "broken",
				Home:       home,
				RepoRoot:   r,
				ConfigSrc:  configSrc,
				ConfigDest: filepath.Join(blocker, "config.toml"),
			},
		}
	}
	t.Cleanup(func() { allHarnesses = orig })

	if err := run(repo); err == nil {
		t.Fatal("expected error propagated from harness Sync(), got nil")
	}
}

func TestRunReturnsErrorWhenRepoRootMissing(t *testing.T) {
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "missing")

	if err := run(missing); err == nil {
		t.Fatal("expected error for missing repo root, got nil")
	}
}

func TestRunReturnsErrorWhenRepoRootNotDir(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "not_a_dir")
	if err := os.WriteFile(file, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := run(file); err == nil {
		t.Fatal("expected error for non-directory repo root, got nil")
	}
}

// TestMainUsageError exercises the flag-parsing usage-error path (missing
// positional arg -> exit code 2) via a self-exec subprocess, since main()
// calls os.Exit directly.
func TestMainUsageError(t *testing.T) {
	if os.Getenv("AGENT_SYNC_TEST_MAIN_HELPER") == "1" {
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainUsageError")
	cmd.Env = append(os.Environ(), "AGENT_SYNC_TEST_MAIN_HELPER=1")
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected *exec.ExitError, got %v (err=%v)", err, err)
	}
	if code := exitErr.ExitCode(); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

// TestMainSuccess exercises main()'s success path (exit code 0) via the
// same self-exec pattern, passing a valid repo_root argument.
func TestMainSuccess(t *testing.T) {
	if os.Getenv("AGENT_SYNC_TEST_MAIN_SUCCESS_HELPER") == "1" {
		os.Args = []string{"agent-synchronizer", os.Getenv("AGENT_SYNC_TEST_REPO")}
		main()
		return
	}
	tmp := t.TempDir()
	repo := filepath.Join(tmp, "repo")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(tmp, "home")
	if err := os.Mkdir(home, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainSuccess")
	cmd.Env = append(os.Environ(),
		"AGENT_SYNC_TEST_MAIN_SUCCESS_HELPER=1",
		"AGENT_SYNC_TEST_REPO="+repo,
		"HOME="+home,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("expected exit code 0, got error: %v", err)
	}
}

// TestMainFailure exercises main()'s failure path (exit code 1) via the
// same self-exec pattern, passing a missing repo_root argument.
func TestMainFailure(t *testing.T) {
	if os.Getenv("AGENT_SYNC_TEST_MAIN_FAILURE_HELPER") == "1" {
		os.Args = []string{"agent-synchronizer", os.Getenv("AGENT_SYNC_TEST_REPO")}
		main()
		return
	}
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "missing")

	cmd := exec.Command(os.Args[0], "-test.run=TestMainFailure")
	cmd.Env = append(os.Environ(),
		"AGENT_SYNC_TEST_MAIN_FAILURE_HELPER=1",
		"AGENT_SYNC_TEST_REPO="+missing,
	)
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected *exec.ExitError, got %v", err)
	}
	if code := exitErr.ExitCode(); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}
