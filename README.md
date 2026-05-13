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

## Config

```yaml
common:
  - AGENTS.md
  - skills
  - agents
providers:
  - name: codex
    path: /Users/me/.codex
    files:
      - config.toml
```

`common` lists files and directories synced for every provider. Provider `files` are
synced under that provider's directory in the central repo.

## OOTB Providers

We use the `$HOME`-based config directories for each provider.

- Claude
- Codex
- OpenCode
- Cursor