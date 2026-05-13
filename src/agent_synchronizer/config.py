from pathlib import Path

import yaml

from .providers import Provider, supported_providers
from .settings import DEFAULT_COMMON
from .types import Config
from .utils.logging import Logger

logger = Logger("config")


def merge_providers(base: list[Provider], override: list[Provider]) -> list[Provider]:
    by_name: dict[str, Provider] = {provider.name: provider for provider in base}
    for provider in override:
        by_name[provider.name] = provider
    return list(by_name.values())


def default_config() -> Config:
    return Config(common=list(DEFAULT_COMMON), providers=[])


def load_config(path: Path) -> Config:
    try:
        with path.open("r") as file:
            data = yaml.safe_load(file) or {}
    except FileNotFoundError:
        logger.info(f"Config file {path} not found; using empty provider config")
        return default_config()

    if not data:
        return default_config()
    if "providers" not in data:
        raise KeyError(f"Config file {path} is missing required 'providers' key")

    return Config(
        common=list(data.get("common", DEFAULT_COMMON)),
        providers=[
            Provider(
                name=entry["name"],
                path=Path(entry["path"]),
                files=list(entry["files"]),
            )
            for entry in data["providers"]
        ],
    )


def save_config(path: Path, config: Config) -> None:
    merged = merge_providers(supported_providers, config.providers)
    payload = {
        "common": list(config.common),
        "providers": [
            {
                "name": provider.name,
                "path": str(provider.path),
                "files": list(provider.files),
            }
            for provider in merged
        ],
    }

    path.parent.mkdir(parents=True, exist_ok=True)
    logger.info(f"Saving config to {path}")
    with path.open("w") as file:
        yaml.safe_dump(payload, file, sort_keys=False)
