from pathlib import Path
from dataclasses import dataclass


@dataclass
class Provider:
    name: str
    path: Path
    files: list[str]


@dataclass
class Config:
    common: list[str]
    providers: list[Provider]


Expected = list[tuple[Provider, Path, Path, bool]]
