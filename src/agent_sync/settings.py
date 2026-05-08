from pathlib import Path

DEFAULT_TARGETS = ["AGENTS.md", "CLAUDE.md", ".cursorrules", "skills", "agents"]

_home = Path.home() # this is just for building the provider paths for my system

SUPPORTED_PROVIDERS = {
    "cursor": _home / ".cursor",
    "codex": _home / ".codex",
    "claude": _home / ".claude",
    "opencode": _home / ".config" / "opencode",
}