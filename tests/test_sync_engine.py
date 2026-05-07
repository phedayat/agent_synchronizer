from pathlib import Path
from unittest.mock import patch

from agent_sync import sync_engine


def test_approve_true_when_input_is_y():
    with patch("builtins.input", return_value="y"):
        assert sync_engine._approve("Proceed") is True


def test_approve_false_for_non_y_response():
    with patch("builtins.input", return_value="n"):
        assert sync_engine._approve("Proceed") is False


def test_discover_files_yields_existing_targets_only(tmp_path: Path):
    provider_path = tmp_path / "provider"
    provider_path.mkdir()
    existing = provider_path / "AGENTS.md"
    existing.write_text("x")

    providers_map = {"cursor": provider_path}
    targets = ["AGENTS.md", "MISSING.md"]

    found = list(sync_engine.discover_files(providers_map, targets))

    assert found == [("cursor", existing)]


def test_discover_files_logs_warning_for_missing_provider(tmp_path: Path):
    missing_provider = tmp_path / "does-not-exist"
    providers_map = {"codex": missing_provider}

    with patch.object(sync_engine.logger, "warning") as warning:
        found = list(sync_engine.discover_files(providers_map, ["AGENTS.md"]))

    assert found == []
    warning.assert_called_once()


def test_move_noop_when_source_is_symlink(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    link = tmp_path / "link"
    link.symlink_to(src, target_is_directory=True)
    dest = tmp_path / "dest"

    with patch.object(sync_engine.logger, "info") as info:
        sync_engine.move(link, dest)

    info.assert_called_once()
    assert not dest.exists()


def test_move_noop_when_destination_exists(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"
    dest.mkdir()

    with patch.object(sync_engine.logger, "info") as info:
        sync_engine.move(src, dest)

    info.assert_called_once()


def test_move_calls_move_when_approved(tmp_path: Path):
    src = tmp_path / "src"
    dest = tmp_path / "dest"

    with patch.object(sync_engine, "_approve", return_value=True), patch.object(
        sync_engine.logger, "info"
    ), patch.object(Path, "move", autospec=True) as move_mock, patch.object(
        Path, "copy", autospec=True
    ) as copy_mock:
        sync_engine.move(src, dest)

    move_mock.assert_called_once_with(src, dest)
    copy_mock.assert_not_called()


def test_move_calls_copy_when_not_approved(tmp_path: Path):
    src = tmp_path / "src"
    dest = tmp_path / "dest"

    with patch.object(sync_engine, "_approve", return_value=False), patch.object(
        sync_engine.logger, "info"
    ), patch.object(Path, "move", autospec=True) as move_mock, patch.object(
        Path, "copy", autospec=True
    ) as copy_mock:
        sync_engine.move(src, dest)

    copy_mock.assert_called_once_with(src, dest)
    move_mock.assert_not_called()


def test_symlink_noop_when_source_missing(tmp_path: Path):
    src = tmp_path / "missing"
    dest = tmp_path / "dest"

    with patch.object(sync_engine.logger, "info") as info:
        sync_engine.symlink(src, dest)

    info.assert_called_once()
    assert not dest.exists()


def test_symlink_noop_when_destination_already_correct(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"
    dest.symlink_to(src, target_is_directory=True)

    with patch.object(sync_engine.logger, "info") as info:
        sync_engine.symlink(src, dest)

    info.assert_called_once()


def test_symlink_noop_when_destination_exists_and_is_not_symlink(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"
    dest.mkdir()

    with patch.object(sync_engine.logger, "info") as info:
        sync_engine.symlink(src, dest)

    info.assert_called_once()


def test_symlink_creates_link_when_valid(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"

    sync_engine.symlink(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == src.resolve()


def test_generate_sync_report_contains_counts_and_entries(tmp_path: Path):
    files_found = [
        ("cursor", tmp_path / "AGENTS.md"),
        ("codex", tmp_path / "skills"),
    ]
    repo_files = [tmp_path / "AGENTS.md"]

    report = sync_engine.generate_sync_report(files_found, repo_files)

    assert "Provider: cursor" in report
    assert "Provider: codex" in report
    assert f"Repo file: {tmp_path / 'AGENTS.md'}" in report
    assert "Total files: 2" in report
    assert "Total repo files: 1" in report
