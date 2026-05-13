from pathlib import Path

import pytest
import yaml

from agent_synchronizer.config import load_config, merge_providers, save_config
from agent_synchronizer.providers import Provider, supported_providers
from agent_synchronizer.settings import DEFAULT_COMMON
from agent_synchronizer.types import Config


def test_load_config_returns_common_and_providers(tmp_path):
    config_path = tmp_path / "config.yaml"
    config_path.write_text(
        yaml.safe_dump(
            {
                "common": ["AGENTS.md", "skills"],
                "providers": [
                    {"name": "foo", "path": "/tmp/foo", "files": ["a.txt"]},
                ],
            }
        )
    )

    config = load_config(config_path)

    assert config.common == ["AGENTS.md", "skills"]
    assert len(config.providers) == 1
    assert config.providers[0].name == "foo"
    assert config.providers[0].path == Path("/tmp/foo")
    assert config.providers[0].files == ["a.txt"]


def test_load_config_uses_default_common_when_key_is_missing(tmp_path):
    config_path = tmp_path / "config.yaml"
    config_path.write_text(
        yaml.safe_dump(
            {
                "providers": [
                    {"name": "foo", "path": "/tmp/foo", "files": ["a.txt"]},
                ]
            }
        )
    )

    config = load_config(config_path)

    assert config.common == DEFAULT_COMMON


def test_load_config_missing_providers_key_raises(tmp_path):
    config_path = tmp_path / "config.yaml"
    config_path.write_text(yaml.safe_dump({"other": []}))

    with pytest.raises(KeyError):
        load_config(config_path)


def test_load_config_empty_file_returns_default_config(tmp_path):
    config_path = tmp_path / "config.yaml"
    config_path.write_text("")

    config = load_config(config_path)

    assert config == Config(common=DEFAULT_COMMON, providers=[])


def test_load_config_missing_file_returns_default_config(tmp_path):
    config_path = tmp_path / "missing.yaml"

    config = load_config(config_path)

    assert config == Config(common=DEFAULT_COMMON, providers=[])


def test_save_config_writes_defaults_and_extras(tmp_path):
    config_path = tmp_path / "out.yaml"
    extra = Provider(name="custom", path=Path("/tmp/custom"), files=["x"])

    save_config(config_path, Config(common=["AGENTS.md"], providers=[extra]))

    data = yaml.safe_load(config_path.read_text())
    assert data["common"] == ["AGENTS.md"]
    names = [provider["name"] for provider in data["providers"]]
    for default in supported_providers:
        assert default.name in names
    assert "custom" in names


def test_save_config_existing_overrides_default(tmp_path):
    config_path = tmp_path / "out.yaml"
    override = Provider(
        name="codex", path=Path("/custom/codex"), files=["override.toml"]
    )

    save_config(config_path, Config(common=DEFAULT_COMMON, providers=[override]))

    data = yaml.safe_load(config_path.read_text())
    codex_entry = next(
        provider for provider in data["providers"] if provider["name"] == "codex"
    )
    assert codex_entry["path"] == "/custom/codex"
    assert codex_entry["files"] == ["override.toml"]


def test_save_then_load_roundtrip(tmp_path):
    config_path = tmp_path / "out.yaml"

    save_config(config_path, Config(common=DEFAULT_COMMON, providers=[]))
    config = load_config(config_path)

    assert config.common == DEFAULT_COMMON
    assert sorted(provider.name for provider in config.providers) == sorted(
        provider.name for provider in supported_providers
    )


def test_merge_providers_override_wins():
    base = [Provider(name="a", path=Path("/a"), files=["x"])]
    override = [Provider(name="a", path=Path("/A"), files=["y"])]

    merged = merge_providers(base, override)

    assert len(merged) == 1
    assert merged[0].path == Path("/A")
    assert merged[0].files == ["y"]
