from pathlib import Path

import yaml

from .providers import Provider, supported_providers
from .utils.logging import Logger

logger = Logger("config")


def merge_providers(base: list[Provider], override: list[Provider]) -> list[Provider]:
    by_name: dict[str, Provider] = {provider.name: provider for provider in base}
    for provider in override:
        by_name[provider.name] = provider
    return list(by_name.values())


def load_config(path: Path) -> list[Provider]:
    try:
        with path.open("r") as file:
            data = yaml.safe_load(file) or {}
    except FileNotFoundError:
        logger.info(f"Config file {path} not found; using empty provider config")
        return []

    if not data:
        return []
    if "providers" not in data:
        raise KeyError(f"Config file {path} is missing required 'providers' key")

    return [
        Provider(
            name=entry["name"],
            path=Path(entry["path"]),
            files=list(entry["files"]),
        )
        for entry in data["providers"]
    ]


def save_config(path: Path, providers: list[Provider]) -> None:
    merged = merge_providers(supported_providers, providers)
    payload = {
        "providers": [
            {
                "name": provider.name,
                "path": str(provider.path),
                "files": list(provider.files),
            }
            for provider in merged
        ]
    }

    path.parent.mkdir(parents=True, exist_ok=True)
    logger.info(f"Saving config to {path}")
    with path.open("w") as file:
        yaml.safe_dump(payload, file, sort_keys=False)
