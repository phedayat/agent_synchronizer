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

- Run CLI: `uv run agent-sync <repo_root>`
- Available flags:
  - `--targets`
  - `--dry-run`
  - `--sync-report`
  - `--verbose`

## Testing

- Test framework: `pytest`.
- Run all tests: `uv run pytest`.
- Run a single file:
  - `uv run pytest tests/test_args.py`
  - `uv run pytest tests/test_logging.py`
  - `uv run pytest tests/test_providers.py`
- Keep tests focused and minimal; prefer unit tests for argument parsing, logging behavior, and provider filesystem flow.

## Development Rules

- Respect `pyproject.toml` and `uv.lock` as source-of-truth for dependencies.
- Keep `requires-python` compatibility intact.
- If dependencies change, update lockfile with `uv lock`.
- Prefer `pathlib` and explicit path handling over string path manipulation.
- Log important file operations; avoid silent destructive behavior.

## Safety for File Operations

- Treat move/symlink actions as high-risk operations.
- Prefer dry-run validation before real execution when changing behavior.
- Do not overwrite existing real files/directories unless explicitly intended.
- Preserve existing user data and local provider configuration.

## Change Discipline

- Make the smallest change that solves the task.
- Do not refactor unrelated code.
- Keep naming consistent with current codebase (`providers`, `discover_files`, `move`, `symlink`).
- Update documentation when behavior or commands change.

## Quick Task Checklist

- Confirm intent and scope.
- Run/verify with `uv run ...`.
- Keep edits minimal and targeted.
- Re-check filesystem edge cases.
- Summarize what changed and any risks.
