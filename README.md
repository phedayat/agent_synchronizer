# Agent Synchronizer

CLI tool for syncing skills, subagents, AGENTS/CLAUDE.md and provider-specific files.

## Install 

```shell
uv add agent-synchronizer
```

## Usage

```shell
uv run agent-synchronizer <repo_root>
```

## Repository Layout

```
<repo_root>/
├── common/
│   ├── skills/
│   ├── agents/
│   └── AGENTS.md
├── claude/
│   ├── CLAUDE.md
│   └── settings.json
├── codex/
│   └── config.toml
├── cursor/
│   └── .cursorrules
└── opencode/
    └── opencode.jsonc
```

`common/` holds skills, subagents, and rules shared across harnesses. Each
harness's own directory holds files specific to it (config, and for Claude
and Cursor, rules).

## Supported Harnesses

We use the `$HOME`-based config directories for each harness.

- Claude
- Codex
- OpenCode
- Cursor