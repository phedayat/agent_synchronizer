from pathlib import Path

from .types import Provider

codex = Provider(name="codex", path=Path.home() / ".codex", files=["config.toml"])
claude = Provider(
    name="claude", path=Path.home() / ".claude", files=["settings.json", "CLAUDE.md"]
)
cursor = Provider(name="cursor", path=Path.home() / ".cursor", files=[".cursorrules"])
opencode = Provider(
    name="opencode", path=Path.home() / ".config" / "opencode", files=["opencode.jsonc"]
)

supported_providers = [
    codex,
    claude,
    cursor,
    opencode,
]
