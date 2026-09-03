import os
from unittest.mock import MagicMock, patch

import pytest

from agent_synchronizer.__main__ import main


def test_main_resolves_relative_repo_root(tmp_path, monkeypatch):
    repo = tmp_path / "repo"
    repo.mkdir()
    monkeypatch.chdir(tmp_path)

    harness_instance = MagicMock()
    harness_cls = MagicMock(return_value=harness_instance)
    harness_instance.name = "fake"

    with patch("agent_synchronizer.__main__.ALL_HARNESSES", [harness_cls]):
        main(os.path.join(".", "repo"))

    harness_cls.assert_called_once_with(repo.resolve())


def test_main_calls_sync_once_per_harness(tmp_path):
    repo = tmp_path / "repo"
    repo.mkdir()

    harness_instance = MagicMock()
    harness_cls = MagicMock(return_value=harness_instance)
    harness_instance.name = "fake"

    with patch("agent_synchronizer.__main__.ALL_HARNESSES", [harness_cls]):
        main(str(repo))

    harness_cls.assert_called_once_with(repo)
    harness_instance.sync.assert_called_once()


def test_main_raises_when_repo_root_missing(tmp_path):
    missing = tmp_path / "missing"

    with pytest.raises(FileNotFoundError):
        main(str(missing))


def test_main_raises_when_repo_root_not_a_dir(tmp_path):
    file_path = tmp_path / "not_a_dir"
    file_path.write_text("content")

    with pytest.raises(NotADirectoryError):
        main(str(file_path))
