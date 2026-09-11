// Package main is the agent-synchronizer CLI entrypoint.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/phedayat/agent_synchronizer/internal/harness"
	"github.com/phedayat/agent_synchronizer/internal/syncengine"
	"github.com/phedayat/agent_synchronizer/internal/synclog"
)

// Version, Commit, and Date are injected at build time via -ldflags -X.
var (
	Version string
	Commit  string
	Date    string
)

var logger = synclog.New("main", true)

// allHarnesses is a seam so tests can substitute stub harnesses.
var allHarnesses = harness.AllHarnesses

// resolveRepoRoot is a seam so tests can force ResolveRepoRoot's error path.
var resolveRepoRoot = syncengine.ResolveRepoRoot

func run(repoRootArg string) error {
	repo, err := resolveRepoRoot(repoRootArg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to resolve repository root %s: %v", repoRootArg, err))
		return err
	}

	info, err := os.Stat(repo)
	if err != nil {
		logger.Error(fmt.Sprintf("Repository root %s does not exist", repo))
		return fmt.Errorf("repository root %s does not exist: %w", repo, os.ErrNotExist)
	}
	if !info.IsDir() {
		logger.Error(fmt.Sprintf("Repository root %s must be a directory", repo))
		return fmt.Errorf("repository root %s must be a directory", repo)
	}

	for _, h := range allHarnesses(repo) {
		logger.Info(fmt.Sprintf("Syncing %s", h.Name))
		if err := h.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Usage: agent-synchronizer <repo_root>")
		os.Exit(2)
	}
	if err := run(flag.Arg(0)); err != nil {
		os.Exit(1)
	}
}
