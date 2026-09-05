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
│   ├── skills/
│   ├── CLAUDE.md
│   ├── settings.json
│   └── hooks/
├── codex/
│   ├── skills/
│   └── config.toml
├── cursor/
│   ├── skills/
│   └── .cursorrules
└── opencode/
    ├── skills/
    └── opencode.jsonc
```

`common/` holds skills, subagents, and rules shared across harnesses. Each
harness's own directory holds files specific to it (config, and for Claude
and Cursor, rules), plus an optional `skills/` folder for skills specific to
that harness.

Skills may be grouped into subfolders anywhere under a `skills/` directory,
at any depth — every synced harness sees them flattened, one directory per
skill. A skill in `<harness>/skills/` overrides a same-named skill in
`common/skills/`.

## Supported Harnesses

We use the `$HOME`-based config directories for each harness.

- Claude
- Codex
- OpenCode
- Cursor