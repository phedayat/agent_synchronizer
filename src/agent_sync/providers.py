from pathlib import Path
from dataclasses import dataclass

@dataclass
class Provider:
    name: str
    path: Path
    files: list[str]

codex = Provider(name="codex", path=Path.home() / ".codex", files=["config.toml"])
claude = Provider(name="claude", path=Path.home() / ".claude", files=["settings.json"])
cursor = Provider(name="cursor", path=Path.home() / ".cursor", files=[".cursorrules"])
opencode = Provider(name="opencode", path=Path.home() / ".config" / "opencode", files=["opencode.jsonc"])

supported_providers = [
    codex,
    claude,
    cursor,
    opencode,
]