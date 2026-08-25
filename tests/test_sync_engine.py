from pathlib import Path
from unittest.mock import patch

from agent_synchronizer import sync_engine


def test_approve_true_when_input_is_y():
    with patch("builtins.input", return_value="y"):
        assert sync_engine._approve("Proceed") is True


def test_approve_false_for_non_y_response():
    with patch("builtins.input", return_value="n"):
        assert sync_engine._approve("Proceed") is False


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
    dest = tmp_path / "nested" / "dest"

    with (
        patch.object(sync_engine, "_approve", return_value=True),
        patch.object(sync_engine.logger, "info"),
        patch.object(Path, "move", autospec=True) as move_mock,
        patch.object(Path, "copy", autospec=True) as copy_mock,
    ):
        sync_engine.move(src, dest)

    move_mock.assert_called_once_with(src, dest)
    copy_mock.assert_not_called()
    assert dest.parent.exists()


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


def test_symlink_creates_link_when_valid(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"

    sync_engine.symlink(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == src.resolve()


def test_symlink_replaces_real_dir_when_approved(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"
    dest.mkdir()
    (dest / "existing").touch()

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine.symlink(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == src.resolve()


def test_symlink_replaces_real_file_when_approved(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"
    dest.write_text("real file")

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine.symlink(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == src.resolve()


def test_symlink_leaves_real_dest_when_declined(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"
    dest.mkdir()

    with patch.object(sync_engine, "_approve", return_value=False):
        sync_engine.symlink(src, dest)

    assert not dest.is_symlink()
    assert dest.is_dir()


def test_absorb_moves_missing_children_only(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    (src / "shared.txt").write_text("in src")

    dest = tmp_path / "dest"
    dest.mkdir()
    (dest / "shared.txt").write_text("in dest")
    (dest / "unique.txt").write_text("only in dest")

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine._absorb(src, dest)

    assert (src / "unique.txt").read_text() == "only in dest"
    assert (src / "shared.txt").read_text() == "in src"
    assert not (dest / "unique.txt").exists()
    assert (dest / "shared.txt").exists()


def test_absorb_is_one_level_only(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()

    dest = tmp_path / "dest"
    nested = dest / "child" / "grandchild"
    nested.mkdir(parents=True)
    (nested / "leaf.txt").write_text("leaf")

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine._absorb(src, dest)

    assert (src / "child" / "grandchild" / "leaf.txt").read_text() == "leaf"
    assert not (dest / "child").exists()


def test_sync_target_noop_when_dest_is_symlink(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"
    dest.symlink_to(src, target_is_directory=True)

    with patch.object(sync_engine, "_absorb") as absorb_mock:
        sync_engine.sync_target(src, dest)

    absorb_mock.assert_not_called()
    assert dest.resolve() == src.resolve()


def test_sync_target_migrates_when_src_missing(tmp_path: Path):
    src = tmp_path / "src"
    dest = tmp_path / "dest"
    dest.mkdir()
    (dest / "file.txt").write_text("content")

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine.sync_target(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == src.resolve()
    assert (src / "file.txt").read_text() == "content"


def test_sync_target_absorbs_then_symlinks_when_both_exist(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    (src / "shared.txt").write_text("in src")

    dest = tmp_path / "dest"
    dest.mkdir()
    (dest / "unique.txt").write_text("only in dest")

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine.sync_target(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == src.resolve()
    assert (src / "unique.txt").read_text() == "only in dest"


def test_sync_target_symlinks_directly_when_dest_missing(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"

    sync_engine.sync_target(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == src.resolve()


def test_sync_target_noop_when_neither_exists(tmp_path: Path):
    src = tmp_path / "src"
    dest = tmp_path / "dest"

    sync_engine.sync_target(src, dest)

    assert not dest.exists()
    assert not src.exists()
