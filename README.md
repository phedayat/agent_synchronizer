# Agent Synchronizer

CLI tool for syncing skills, subagents, AGENTS/CLAUDE.md and provider-specific files.

## Install

Download a prebuilt binary from
[GitHub Releases](https://github.com/phedayat/agent_synchronizer/releases), or
build from source with Go:

```shell
go install github.com/phedayat/agent_synchronizer/cmd/agent-synchronizer@latest
```

## Usage

```shell
agent-synchronizer <repo_root>
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
├── opencode/
│   ├── skills/
│   └── opencode.jsonc
└── hermes/
    ├── skills/
    └── config.json
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
- Hermes