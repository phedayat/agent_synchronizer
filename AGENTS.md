# AGENTS.md

## Purpose

- This project synchronizes shared agent files/directories across provider folders.
- Keep implementation minimal, explicit, and safe for filesystem operations.

## Core Principles

- Prefer small, readable functions and clear control flow.
- Apply DRY and KISS; avoid speculative abstractions.
- Make no unrelated code changes.
- Favor idempotent behavior for filesystem operations.

## Tooling

- Use `uv` for all Python workflows in this repository.
- Do not use `pip`, `pipenv`, `poetry`, or bare `python`/`pytest` when `uv` equivalents exist.

## Standard Commands

- Run CLI: `uv run agent-synchronizer <repo_root>`
- No flags; `repo_root` is the only argument.
- Before considering any change complete, run `make prepare` (lint, format, typecheck, test) and ensure it passes.

## Harnesses

- Each harness (`claude`, `codex`, `cursor`, `opencode`) is a fixed class in
  `src/agent_synchronizer/harnesses.py` implementing the `Harness` ABC
  (`sync_skills`, `sync_subagents`, `sync_config`, `sync_rules`, `sync_hooks`).
- `Harness.sync()` calls all five methods in sequence; there is no per-method CLI flag.
- Only Claude syncs hooks today (`<repo_root>/claude/hooks` → `~/.claude/hooks`);
  `sync_hooks()` is a documented no-op for Codex, Cursor, and OpenCode.
- Subagents sync from `<repo_root>/common/agents` for every harness. Config
  and rules are per-harness (`<repo_root>/<harness>/...`), except
  Claude/Codex/OpenCode's rules, which also come from `common/`
  (`AGENTS.md`/`CLAUDE.md` per the mapping in `harnesses.py`).
- Skills are flattened: `<repo_root>/common/skills` may group skills into
  subfolders any number of levels deep (a directory is a skill once it
  directly contains `SKILL.md`), and `<repo_root>/<harness>/skills` holds
  optional harness-specific skills, overriding a same-named common skill
  outright. `Harness.sync_skills()` calls
  `sync_engine.sync_flattened_skills(dest, *sources)` to place every skill
  directly under `<home>/skills`, regardless of how it's grouped in the
  repo. Unlike every other sync target, `<home>/skills` itself is a real
  directory (not a symlink) — each skill inside it is its own symlink,
  since a single symlink can't point at a flattened, merged view of
  multiple source directories.
- `Cursor.sync_config()` is a documented no-op — Cursor has no config file to sync today.

## Testing

- Test framework: `pytest`.
- Run all tests: `uv run pytest`.
- Run a single file:
  - `uv run pytest tests/test_args.py`
  - `uv run pytest tests/test_harnesses.py`
  - `uv run pytest tests/test_logging.py`
  - `uv run pytest tests/test_main.py`
  - `uv run pytest tests/test_sync_engine.py`
- Keep tests focused and minimal; prefer unit tests for argument parsing, harness sync-target mapping, logging behavior, and filesystem sync flow.
- Every line and branch of `src/agent_synchronizer/` must be covered by a
  unit test — including edge cases and failure modes, not just the happy
  path. Check with `uv run --with coverage coverage run --branch -m pytest
  -q && uv run --with coverage coverage report -m --include="src/*"`
  before considering a change complete.

## Development Rules

- Respect `pyproject.toml` and `uv.lock` as source-of-truth for dependencies.
- Keep `requires-python` compatibility intact.
- If dependencies change, update lockfile with `uv lock`.
- Prefer `pathlib` and explicit path handling over string path manipulation.
- Log important file operations; avoid silent destructive behavior.
- To add a harness, define a new `Harness` subclass in `src/agent_synchronizer/harnesses.py` implementing the five `sync_*` methods, then add it to `ALL_HARNESSES`.
- Every new feature (method, function, class, or CLI behavior) must ship with an associated unit test in the same change, without exception.
- Keep sync flow explicit: each harness's `sync_*` method calls `sync_engine.sync_target(src, dest)` for its specific path pair. `sync_target` uses `_absorb` to reconcile dest-only content into the repo before calling `symlink`; `move` relocates whole subtrees. `sync_skills` is the one exception, calling `sync_engine.sync_flattened_skills(dest, *sources)` instead, since skills may be grouped in the repo but must land flat at the destination.

## Safety for File Operations

- Treat move/symlink actions as high-risk operations.
- There is no dry-run mode; validate behavior changes against a scratch
  `repo_root`/`$HOME` before running against a real one.
- Do not overwrite existing real files/directories unless explicitly intended.
- Preserve existing user data and local provider configuration.

## Change Discipline

- Make the smallest change that solves the task.
- Do not refactor unrelated code.
- Keep naming consistent with current codebase (`Harness`, `ALL_HARNESSES`, `sync_target`, `_absorb`, `symlink`, `move`).
- Update documentation when behavior or commands change.

## Quick Task Checklist

- Confirm intent and scope.
- Run/verify with `uv run ...`.
- The CLI takes no flags (there is no dry-run mode); validate sync changes
  against a scratch `repo_root` and scratch `$HOME` (see
  `scripts/smoke_test.sh` and `tests/smoke_test_skills_flatten.sh` for the
  pattern) before running against a real repo/home.
- Keep edits minimal and targeted.
- Re-check filesystem edge cases.
- Summarize what changed and any risks.
