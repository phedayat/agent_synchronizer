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


def test_move_noop_when_destination_is_existing_file(tmp_path: Path):
    src = tmp_path / "src.txt"
    src.write_text("source")
    dest = tmp_path / "dest.txt"
    dest.write_text("dest")

    with patch.object(sync_engine.logger, "info") as info:
        sync_engine.move(src, dest)

    info.assert_called_once()
    assert dest.read_text() == "dest"
    assert src.read_text() == "source"


def test_move_file_calls_move_when_approved(tmp_path: Path):
    src = tmp_path / "src.txt"
    src.write_text("content")
    dest = tmp_path / "nested" / "dest.txt"

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine.move(src, dest)

    assert not src.exists()
    assert dest.read_text() == "content"


def test_move_file_calls_copy_when_not_approved(tmp_path: Path):
    src = tmp_path / "src.txt"
    src.write_text("content")
    dest = tmp_path / "dest.txt"

    with patch.object(sync_engine, "_approve", return_value=False):
        sync_engine.move(src, dest)

    assert src.exists()
    assert src.read_text() == "content"
    assert dest.read_text() == "content"


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


def test_symlink_replaces_stale_symlink_to_dir_when_approved(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    old_target = tmp_path / "old_target"
    old_target.mkdir()
    dest = tmp_path / "dest"
    dest.symlink_to(old_target, target_is_directory=True)

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine.symlink(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == src.resolve()
    assert old_target.is_dir()


def test_symlink_leaves_real_dest_when_declined(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"
    dest.mkdir()

    with patch.object(sync_engine, "_approve", return_value=False):
        sync_engine.symlink(src, dest)

    assert not dest.is_symlink()
    assert dest.is_dir()


def test_symlink_leaves_stale_symlink_when_declined(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    old_target = tmp_path / "old_target"
    old_target.mkdir()
    dest = tmp_path / "dest"
    dest.symlink_to(old_target, target_is_directory=True)

    with patch.object(sync_engine, "_approve", return_value=False):
        sync_engine.symlink(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == old_target.resolve()


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


def test_absorb_declined_child_stays_in_dest_and_reports_incomplete(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()

    dest = tmp_path / "dest"
    dest.mkdir()
    (dest / "unique.txt").write_text("only in dest")

    with patch.object(sync_engine, "_approve", return_value=False):
        result = sync_engine._absorb(src, dest)

    assert result is False
    assert (dest / "unique.txt").read_text() == "only in dest"
    assert (src / "unique.txt").read_text() == "only in dest"


def test_absorb_both_files_is_noop_and_reports_fully_absorbed(tmp_path: Path):
    src = tmp_path / "src.txt"
    src.write_text("src content")
    dest = tmp_path / "dest.txt"
    dest.write_text("dest content")

    result = sync_engine._absorb(src, dest)

    assert result is True
    assert src.read_text() == "src content"
    assert dest.read_text() == "dest content"


def test_absorb_noop_when_dest_missing(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    dest = tmp_path / "dest"

    sync_engine._absorb(src, dest)

    assert not dest.exists()
    assert list(src.iterdir()) == []


def test_absorb_noop_when_dest_is_symlink(tmp_path: Path):
    src = tmp_path / "src"
    src.mkdir()
    real = tmp_path / "real"
    real.mkdir()
    (real / "leaf.txt").write_text("leaf")
    dest = tmp_path / "dest"
    dest.symlink_to(real, target_is_directory=True)

    sync_engine._absorb(src, dest)

    assert list(src.iterdir()) == []
    assert (real / "leaf.txt").exists()


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


def test_sync_target_both_plain_files_does_not_raise(tmp_path: Path):
    src = tmp_path / "config.toml"
    src.write_text("in repo")
    dest = tmp_path / "config_home.toml"
    dest.write_text("in home")

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine.sync_target(src, dest)

    assert dest.is_symlink()
    assert dest.resolve() == src.resolve()
    assert src.read_text() == "in repo"


def test_sync_target_declined_absorb_child_is_not_destroyed_by_symlink(
    tmp_path: Path,
):
    """Regression test: declining a child's absorb must stick even if the
    subsequent symlink-replace prompt would have been approved."""
    src = tmp_path / "src"
    src.mkdir()

    dest = tmp_path / "dest"
    dest.mkdir()
    (dest / "unique.txt").write_text("only in dest")

    # First prompt (absorbing "unique.txt") is declined; any later prompt
    # (e.g. symlink's replace-dest confirmation) would be approved.
    with (
        patch.object(sync_engine, "_approve", side_effect=[False, True, True]),
        patch.object(sync_engine, "symlink") as symlink_mock,
    ):
        sync_engine.sync_target(src, dest)

    symlink_mock.assert_not_called()
    assert dest.is_dir()
    assert (dest / "unique.txt").read_text() == "only in dest"


def test_sync_target_noop_when_neither_exists(tmp_path: Path):
    src = tmp_path / "src"
    dest = tmp_path / "dest"

    sync_engine.sync_target(src, dest)

    assert not dest.exists()
    assert not src.exists()


def _make_skill(root: Path, *parts: str):
    skill_dir = root.joinpath(*parts)
    skill_dir.mkdir(parents=True)
    (skill_dir / "SKILL.md").write_text(f"# {parts[-1]}")
    return skill_dir


def test_iter_skill_dirs_finds_top_level_skill(tmp_path: Path):
    root = tmp_path / "skills"
    _make_skill(root, "skill_a")

    found = list(sync_engine._iter_skill_dirs(root))

    assert found == [root / "skill_a"]


def test_iter_skill_dirs_finds_grouped_skill(tmp_path: Path):
    root = tmp_path / "skills"
    _make_skill(root, "group_1", "skill_a")

    found = list(sync_engine._iter_skill_dirs(root))

    assert found == [root / "group_1" / "skill_a"]


def test_iter_skill_dirs_finds_arbitrarily_nested_skill(tmp_path: Path):
    root = tmp_path / "skills"
    _make_skill(root, "group_1", "group_2", "skill_a")

    found = list(sync_engine._iter_skill_dirs(root))

    assert found == [root / "group_1" / "group_2" / "skill_a"]


def test_iter_skill_dirs_stops_descending_once_matched(tmp_path: Path):
    root = tmp_path / "skills"
    skill = _make_skill(root, "skill_a")
    (skill / "nested_dir").mkdir()
    (skill / "nested_dir" / "SKILL.md").write_text("should not be found")

    found = list(sync_engine._iter_skill_dirs(root))

    assert found == [skill]


def test_iter_skill_dirs_missing_root_yields_nothing(tmp_path: Path):
    found = list(sync_engine._iter_skill_dirs(tmp_path / "missing"))

    assert found == []


def test_collect_skills_merges_sources_last_wins(tmp_path: Path):
    common = tmp_path / "common"
    override = tmp_path / "override"
    common_skill = _make_skill(common, "shared")
    override_skill = _make_skill(override, "shared")
    only_common = _make_skill(common, "only_common")

    skills = sync_engine._collect_skills(common, override)

    assert skills == {"shared": override_skill, "only_common": only_common}
    assert common_skill != override_skill


def test_sync_flattened_skills_symlinks_each_skill_directly_under_dest(
    tmp_path: Path,
):
    common = tmp_path / "common"
    _make_skill(common, "group_1", "skill_a")
    _make_skill(common, "skill_b")
    dest = tmp_path / "dest"

    sync_engine.sync_flattened_skills(dest, common)

    assert (dest / "skill_a").resolve() == (common / "group_1" / "skill_a").resolve()
    assert (dest / "skill_b").resolve() == (common / "skill_b").resolve()
    assert not (dest / "group_1").exists()


def test_sync_flattened_skills_harness_specific_overrides_common(tmp_path: Path):
    common = tmp_path / "common"
    override = tmp_path / "override"
    _make_skill(common, "skill_a")
    override_skill = _make_skill(override, "skill_a")
    dest = tmp_path / "dest"

    sync_engine.sync_flattened_skills(dest, common, override)

    assert (dest / "skill_a").resolve() == override_skill.resolve()


def test_sync_flattened_skills_absorbs_unmanaged_real_dir(tmp_path: Path):
    common = tmp_path / "common"
    override = tmp_path / "override"
    dest = tmp_path / "dest"
    dest.mkdir()
    (dest / "manual_skill").mkdir()
    (dest / "manual_skill" / "SKILL.md").write_text("manual")

    with patch.object(sync_engine, "_approve", return_value=True):
        sync_engine.sync_flattened_skills(dest, common, override)

    assert (dest / "manual_skill").is_symlink()
    assert (dest / "manual_skill").resolve() == (override / "manual_skill").resolve()
    assert (override / "manual_skill" / "SKILL.md").exists()


def test_sync_flattened_skills_leaves_correct_symlink_alone(tmp_path: Path):
    common = tmp_path / "common"
    skill = _make_skill(common, "skill_a")
    dest = tmp_path / "dest"
    dest.mkdir()
    (dest / "skill_a").symlink_to(skill, target_is_directory=True)

    with patch.object(sync_engine.logger, "info") as info:
        sync_engine.sync_flattened_skills(dest, common)

    assert (dest / "skill_a").resolve() == skill.resolve()
    assert any("already correct" in call.args[0] for call in info.call_args_list)


def test_sync_flattened_skills_rebuilds_stale_whole_dir_symlink(tmp_path: Path):
    common = tmp_path / "common"
    _make_skill(common, "skill_a")
    old_target = tmp_path / "old_common_skills"
    old_target.mkdir()
    dest = tmp_path / "dest"
    dest.symlink_to(old_target, target_is_directory=True)

    sync_engine.sync_flattened_skills(dest, common)

    assert not dest.is_symlink()
    assert dest.is_dir()
    assert (dest / "skill_a").resolve() == (common / "skill_a").resolve()
