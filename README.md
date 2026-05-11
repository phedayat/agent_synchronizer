# Agent Sync

CLI tool for syncing skills, subagents, AGENTS/CLAUDE.md and provider-specific files.

## Usage

```shell
uv run agent-sync <central_repo> \
    --config <config_path> \
    --save-config \
    --sync-report \
    --verbose \
```

## OOTB Providers

We use the `$HOME`-based config directories for each provider.

- Claude
- Codex
- OpenCode
- Cursor