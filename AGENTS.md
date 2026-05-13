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
- Available flags:
  - `--targets`
  - `--dry-run`
  - `--sync-report`
  - `--verbose`
  - `--config`
  - `--save-config`

## Config Usage

- Config file format: YAML with a top-level `providers` key containing provider objects (`name`, `path`, `files`).
- Loading config: `--config <path>` loads providers from that YAML and merges them with `supported_providers` by provider `name`.
- Missing config file: loading a non-existent `--config` path is supported and treated as an empty provider config.
- Empty config file/object: treated as no configured providers.
- Invalid config shape: if a config file exists but omits the `providers` key, raise an error.
- Saving config: `--save-config` writes merged providers (defaults + loaded config) to `--config`; loaded config providers win on name conflicts.
- `--save-config` requires `--config <path>` so there is an explicit destination file.

## Testing

- Test framework: `pytest`.
- Run all tests: `uv run pytest`.
- Run a single file:
  - `uv run pytest tests/test_args.py`
  - `uv run pytest tests/test_config.py`
  - `uv run pytest tests/test_logging.py`
  - `uv run pytest tests/test_sync_engine.py`
- Keep tests focused and minimal; prefer unit tests for argument parsing, config loading/merging, logging behavior, and provider filesystem flow.

## Development Rules

- Respect `pyproject.toml` and `uv.lock` as source-of-truth for dependencies.
- Keep `requires-python` compatibility intact.
- If dependencies change, update lockfile with `uv lock`.
- Prefer `pathlib` and explicit path handling over string path manipulation.
- Log important file operations; avoid silent destructive behavior.
- To add a provider, create a new `Provider` instance in `src/agent_sync/providers.py` with the provider `name`, `path`, and `files`, then append that instance to the `supported_providers` list. `Provider` itself is defined in `src/agent_sync/types.py`.
- Keep sync flow explicit: use `enumerate_targets` (in `src/agent_sync/sync_engine.py`) to build expected `(provider, source, dest, specific)` entries, then `_move_files` and `_symlink_files` (in `src/agent_sync/__main__.py`) to move existing sources into the repo and link provider paths back to repo destinations. Use `generate_sync_report` to render the `--sync-report` output from those expected entries.

## Safety for File Operations

- Treat move/symlink actions as high-risk operations.
- Prefer dry-run validation before real execution when changing behavior.
- Do not overwrite existing real files/directories unless explicitly intended.
- Preserve existing user data and local provider configuration.

## Change Discipline

- Make the smallest change that solves the task.
- Do not refactor unrelated code.
- Keep naming consistent with current codebase (`providers`, `supported_providers`, `enumerate_targets`, `move`, `symlink`, `generate_sync_report`, `_move_files`, `_symlink_files`, `merge_providers`, `load_config`, `save_config`).
- Update documentation when behavior or commands change.

## Quick Task Checklist

- Confirm intent and scope.
- Run/verify with `uv run ...`.
- Validate sync changes with both `--dry-run` and real `--sync-report` runs when behavior changes.
- Keep edits minimal and targeted.
- Re-check filesystem edge cases.
- Summarize what changed and any risks.
