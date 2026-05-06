from pathlib import Path

DEFAULT_TARGETS = ["AGENTS.md", "CLAUDE.md", ".cursorrules", "skills", "agents"]

_home = Path.home()

SUPPORTED_PROVIDERS = {
    "cursor": _home / ".cursor",
    "codex": _home / ".codex",
    "claude": _home / ".claude",
}