from unittest.mock import patch

import yaml

from agent_sync.__main__ import main


def test_main_uses_config_common_when_targets_are_omitted(tmp_path):
    repo = tmp_path / "repo"
    repo.mkdir()
    config_path = tmp_path / "config.yaml"
    config_path.write_text(
        yaml.safe_dump(
            {
                "common": ["shared"],
                "providers": [
                    {"name": "custom", "path": str(tmp_path / "custom"), "files": []}
                ],
            }
        )
    )

    with patch(
        "agent_sync.__main__.enumerate_targets", return_value=[]
    ) as enumerate_mock:
        main(None, True, False, False, str(repo), str(config_path), False)

    assert enumerate_mock.call_args.args[1] == ["shared"]


def test_main_targets_override_config_common(tmp_path):
    repo = tmp_path / "repo"
    repo.mkdir()
    config_path = tmp_path / "config.yaml"
    config_path.write_text(
        yaml.safe_dump(
            {
                "common": ["shared"],
                "providers": [
                    {"name": "custom", "path": str(tmp_path / "custom"), "files": []}
                ],
            }
        )
    )

    with patch(
        "agent_sync.__main__.enumerate_targets", return_value=[]
    ) as enumerate_mock:
        main(["override"], True, False, False, str(repo), str(config_path), False)

    assert enumerate_mock.call_args.args[1] == ["override"]
