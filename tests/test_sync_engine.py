from pathlib import Path
from unittest.mock import patch

from agent_sync import sync_engine
from agent_sync.providers import Provider


def test_approve_true_when_input_is_y():
    with patch("builtins.input", return_value="y"):
        assert sync_engine._approve("Proceed") is True


def test_approve_false_for_non_y_response():
    with patch("builtins.input", return_value="n"):
        assert sync_engine._approve("Proceed") is False


def test_enumerate_targets_builds_shared_targets_for_each_provider(tmp_path: Path):
    repo = tmp_path / "repo"
    cursor_provider = Provider(name="cursor", path=tmp_path / "cursor", files=[])
    claude_provider = Provider(name="claude", path=tmp_path / "claude", files=[])

    targets = ["AGENTS.md", "skills"]

    expected = sync_engine.enumerate_targets(
        [cursor_provider, claude_provider], targets, repo
    )

    assert expected == [
        (
            cursor_provider,
            cursor_provider.path / "AGENTS.md",
            repo / "AGENTS.md",
            False,
        ),
        (cursor_provider, cursor_provider.path / "skills", repo / "skills", False),
        (
            claude_provider,
            claude_provider.path / "AGENTS.md",
            repo / "AGENTS.md",
            False,
        ),
        (claude_provider, claude_provider.path / "skills", repo / "skills", False),
    ]


def test_enumerate_targets_puts_provider_specific_targets_under_provider_dir(
    tmp_path: Path,
):
    repo = tmp_path / "repo"
    provider = Provider(name="codex", path=tmp_path / "codex", files=["config.toml"])

    expected = sync_engine.enumerate_targets([provider], ["AGENTS.md"], repo)

    assert expected == [
        (provider, provider.path / "AGENTS.md", repo / "AGENTS.md", False),
        (provider, provider.path / "config.toml", repo / "codex" / "config.toml", True),
    ]


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

    with (
        patch.object(sync_engine, "_approve", return_value=True),
        patch.object(sync_engine.logger, "info"),
        patch.object(Path, "move", autospec=True) as move_mock,
        patch.object(Path, "copy", autospec=True) as copy_mock,
    ):
        sync_engine.move(src, dest)

    move_mock.assert_called_once_with(src, dest)
    copy_mock.assert_not_called()


def test_move_calls_copy_when_not_approved(tmp_path: Path):
    src = tmp_path / "src"
    dest = tmp_path / "dest"

    with (
        patch.object(sync_engine, "_approve", return_value=False),
        patch.object(sync_engine.logger, "info"),
        patch.object(Path, "move", autospec=True) as move_mock,
        patch.object(Path, "copy", autospec=True) as copy_mock,
    ):
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
    cursor_provider = Provider(name="cursor", path=tmp_path / "cursor", files=[])
    codex_provider = Provider(
        name="codex", path=tmp_path / "codex", files=["config.toml"]
    )
    repo = tmp_path / "repo"
    expected = [
        (
            cursor_provider,
            cursor_provider.path / "AGENTS.md",
            repo / "AGENTS.md",
            False,
        ),
        (
            codex_provider,
            codex_provider.path / "config.toml",
            repo / "codex" / "config.toml",
            True,
        ),
    ]

    report = sync_engine.generate_sync_report(expected)

    assert "Provider: cursor" in report
    assert "Provider: codex" in report
    assert "Specific: False" in report
    assert "Specific: True" in report
    assert f"Source: {cursor_provider.path / 'AGENTS.md'}" in report
    assert f"Dest: {repo / 'AGENTS.md'}" in report
    assert f"Source: {codex_provider.path / 'config.toml'}" in report
    assert f"Dest: {repo / 'codex' / 'config.toml'}" in report
    assert "Total files: 2" in report


def test_generate_sync_report_returns_no_files_found_when_empty():
    assert sync_engine.generate_sync_report([]) == "No files found"
