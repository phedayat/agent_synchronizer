from pathlib import Path
from dataclasses import dataclass

@dataclass
class Provider:
    name: str
    path: Path
    files: list[str]

Expected = list[tuple[Provider, Path, Path, bool]]