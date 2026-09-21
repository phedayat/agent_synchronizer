# AGENTS.md

## Purpose

- This project synchronizes shared agent files/directories across provider folders.
- Keep implementation minimal, explicit, and safe for filesystem operations.
- The repo is the source of truth for both its own files and the agent
  config (destination) directories; this tool's job is to keep the repo's
  files and the destination symlinks in sync bidirectionally, including
  removing destination symlinks (and directories) whose repo source has
  been removed.

## Core Principles

- Prefer small, readable functions and clear control flow.
- Apply DRY and KISS; avoid speculative abstractions.
- Make no unrelated code changes.
- Favor idempotent behavior for filesystem operations.

## Tooling

- Use the Go toolchain (`go build`, `go test`, `go vet`, `gofmt`) and the
  `Makefile` targets for all workflows in this repository.
- No third-party build tool is required; do not introduce one.

## Standard Commands

- Run CLI (development): `go run ./cmd/agent-synchronizer <repo_root>`
- Run CLI (built binary): `go build -o bin/agent-synchronizer ./cmd/agent-synchronizer && ./bin/agent-synchronizer <repo_root>`
- No flags; `repo_root` is the only argument.
- Before considering any change complete, run `make prepare` (lint, format, typecheck, test) and ensure it passes.

## Harnesses

- Every harness (`claude`, `codex`, `cursor`, `opencode`, `hermes`) is a
  value of the single config-driven `Harness` struct in
  `internal/harness/harness.go`, not a separate class per harness.
  Per-harness constructors (`NewClaude`, `NewCodex`, `NewCursor`,
  `NewOpenCode`, `NewHermes`) populate its fields (`ConfigSrc`/`ConfigDest`,
  `RulesSrc`/`RulesDest`, `HooksSrc`/`HooksDest`); an empty source field
  means that step is a no-op for that harness. `Harness.Sync()` calls
  `SyncSkills`, `SyncSubagents`, `SyncConfig`, `SyncRules`, `SyncHooks` in
  sequence; there is no per-method CLI flag.
- Only Claude syncs hooks today (`<repo_root>/claude/hooks` → `~/.claude/hooks`);
  `SyncHooks()` is a no-op for Codex, Cursor, and OpenCode (empty `HooksSrc`).
- Subagents sync from `<repo_root>/common/agents` for every harness. Config
  and rules are per-harness (`<repo_root>/<harness>/...`), except
  Claude/Codex/OpenCode's rules, which also come from `common/`
  (`AGENTS.md`/`CLAUDE.md` per the mapping in the `New*` constructors in
  `internal/harness/harness.go`).
- Skills are flattened: `<repo_root>/common/skills` may group skills into
  subfolders any number of levels deep (a directory is a skill once it
  directly contains `SKILL.md`), and `<repo_root>/<harness>/skills` holds
  optional harness-specific skills, overriding a same-named common skill
  outright. `Harness.SyncSkills()` calls
  `syncengine.SyncFlattenedSkills(dest, *sources)` to place every skill
  directly under `<home>/skills`, regardless of how it's grouped in the
  repo. Unlike every other sync target, `<home>/skills` itself is a real
  directory (not a symlink) — each skill inside it is its own symlink,
  since a single symlink can't point at a flattened, merged view of
  multiple source directories.
- Cursor's `ConfigSrc`/`ConfigDest` are left empty, so `SyncConfig()` is a
  no-op — Cursor has no config file to sync today.
- `Harness` has a `PreserveSkillGroups bool` field (default `false`).
  Hermes is the one harness that sets it to `true`: for Hermes,
  `<repo_root>/hermes/skills` keeps its group structure at the destination
  (`~/.hermes/skills/<group>/<skill>`), while `common/skills` still
  flattens to `~/.hermes/skills/<skill>` regardless of grouping in the
  repo. Every other harness leaves `PreserveSkillGroups` at `false` and is
  unaffected. A new skill later added directly at `~/.hermes/skills/<name>`
  or inside an existing group (`~/.hermes/skills/<group>/<name>`) is
  absorbed back into the repo by `syncengine.SyncPartiallyGroupedSkills`:
  one inside a known group goes to `<repo_root>/hermes/skills/<group>/<name>`,
  one outside every known group goes to `<repo_root>/common/skills/<name>`
  instead, since it has no group of its own.

## Testing

- Test framework: Go's standard `testing` package.
- Run all tests: `go test ./...` (or `make test` for `-v -cover`).
- Run a single package's tests:
  - `go test ./internal/synclog/...`
  - `go test ./internal/syncengine/...`
  - `go test ./internal/harness/...`
  - `go test ./cmd/agent-synchronizer/...`
- Keep tests focused and minimal; prefer unit tests for argument parsing, harness sync-target mapping, logging behavior, and filesystem sync flow.
- Every line and branch of the codebase must be covered by a unit test —
  including edge cases and failure modes, not just the happy path. CI
  enforces 100% total coverage; check locally with `go test ./...
  -coverprofile=coverage.out && go tool cover -func=coverage.out` before
  considering a change complete.

## Development Rules

- Respect `go.mod` and `go.sum` as source-of-truth for dependencies.
- Keep the `go.mod` Go-version requirement intact.
- If dependencies change, update `go.sum` with `go mod tidy`.
- Prefer `path/filepath` and explicit path handling over string path manipulation.
- Log important file operations; avoid silent destructive behavior.
- To add a harness, add a new `New*` constructor in
  `internal/harness/harness.go` that builds a `Harness` value with the
  right fields set, then add it to `AllHarnesses`.
- Every new feature (method, function, type, or CLI behavior) must ship with an associated unit test in the same change, without exception.
- Keep sync flow explicit: each harness's `Sync*` method calls
  `syncengine.SyncTarget(src, dest)` for its specific path pair.
  `SyncTarget` uses `absorb` to reconcile dest-only content into the repo
  before calling `symlink`; `move` relocates whole subtrees. `SyncSkills`
  is the one exception, calling `syncengine.SyncFlattenedSkills(dest,
  *sources)` instead, since skills may be grouped in the repo but must
  land flat at the destination.

## Safety for File Operations

- Treat move/symlink actions as high-risk operations.
- There is no dry-run mode; validate behavior changes against a scratch
  `repo_root`/`$HOME` before running against a real one.
- Do not overwrite existing real files/directories unless explicitly intended.
- Preserve existing user data and local provider configuration.
- Deleting a destination symlink or directory because its repo source is
  gone is intentional cleanup, not data loss — the repo source of truth
  never changes as a result; before removing a directory, confirm it holds
  no absorbable real (non-symlink) content, and absorb that content back
  into the repo first.
- Removal recursion always walks strictly downward through real (non-
  symlink) directories on the destination side, so it terminates and never
  follows a symlink into a cycle.

## Change Discipline

- Make the smallest change that solves the task.
- Do not refactor unrelated code.
- Keep naming consistent with current codebase (`Harness`, `AllHarnesses`, `SyncTarget`, `absorb`, `symlink`, `move`).
- Update documentation when behavior or commands change.

## Quick Task Checklist

- Confirm intent and scope.
- Run/verify with `go run ./cmd/agent-synchronizer <repo_root>`.
- The CLI takes no flags (there is no dry-run mode); validate sync changes
  against a scratch `repo_root` and scratch `$HOME` (see
  `scripts/smoke_test.sh` and `tests/smoke_test_skills_flatten.sh` for the
  pattern) before running against a real repo/home.
- Keep edits minimal and targeted.
- Re-check filesystem edge cases.
- Summarize what changed and any risks.
