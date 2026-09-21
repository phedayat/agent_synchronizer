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
			logger.Error(fmt.Sprintf("Failed to sync %s: %v", h.Name, err))
			return err
		}
	}
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: %[1]s <repo_root>

agent-synchronizer keeps a repo's shared agent config (skills, subagents,
config, rules, hooks) bidirectionally in sync with each supported
harness's config directory under your home directory: claude, codex,
cursor, opencode, hermes.

Arguments:
  repo_root   Path to the repo holding the harness config folders
              (common/, claude/, codex/, cursor/, opencode/, hermes/).

There are no flags; there is no dry-run mode.

Example:
  %[1]s ~/Desktop/agent-configs
`, os.Args[0])
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(flag.Arg(0)); err != nil {
		os.Exit(1)
	}
}
