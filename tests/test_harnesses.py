from unittest.mock import patch

import pytest

from agent_synchronizer.harnesses import (
    ALL_HARNESSES,
    Claude,
    Codex,
    Cursor,
    Harness,
    OpenCode,
)


def test_harness_cannot_be_instantiated_directly(tmp_path):
    with pytest.raises(TypeError):
        Harness(tmp_path)


def test_all_harnesses_contents():
    assert ALL_HARNESSES == [Claude, Codex, Cursor, OpenCode]


def test_claude_sync_targets(tmp_path):
    home = tmp_path / "home" / ".claude"
    harness = Claude(tmp_path)
    harness.home = home

    with (
        patch("agent_synchronizer.harnesses.sync_engine.sync_target") as mock,
        patch(
            "agent_synchronizer.harnesses.sync_engine.sync_flattened_skills"
        ) as skills_mock,
    ):
        harness.sync_skills()
        harness.sync_subagents()
        harness.sync_config()
        harness.sync_rules()
        harness.sync_hooks()

    skills_mock.assert_called_once_with(
        home / "skills", tmp_path / "common" / "skills", tmp_path / "claude" / "skills"
    )
    mock.assert_any_call(tmp_path / "common" / "agents", home / "agents")
    mock.assert_any_call(tmp_path / "claude" / "settings.json", home / "settings.json")
    mock.assert_any_call(tmp_path / "claude" / "CLAUDE.md", home / "CLAUDE.md")
    mock.assert_any_call(tmp_path / "claude" / "hooks", home / "hooks")
    assert mock.call_count == 4


def test_codex_sync_targets(tmp_path):
    home = tmp_path / "home" / ".codex"
    harness = Codex(tmp_path)
    harness.home = home

    with (
        patch("agent_synchronizer.harnesses.sync_engine.sync_target") as mock,
        patch(
            "agent_synchronizer.harnesses.sync_engine.sync_flattened_skills"
        ) as skills_mock,
    ):
        harness.sync_skills()
        harness.sync_subagents()
        harness.sync_config()
        harness.sync_rules()

    skills_mock.assert_called_once_with(
        home / "skills", tmp_path / "common" / "skills", tmp_path / "codex" / "skills"
    )
    mock.assert_any_call(tmp_path / "common" / "agents", home / "agents")
    mock.assert_any_call(tmp_path / "codex" / "config.toml", home / "config.toml")
    mock.assert_any_call(tmp_path / "common" / "AGENTS.md", home / "AGENTS.md")
    assert mock.call_count == 3


def test_cursor_sync_targets(tmp_path):
    home = tmp_path / "home" / ".cursor"
    harness = Cursor(tmp_path)
    harness.home = home

    with (
        patch("agent_synchronizer.harnesses.sync_engine.sync_target") as mock,
        patch(
            "agent_synchronizer.harnesses.sync_engine.sync_flattened_skills"
        ) as skills_mock,
    ):
        harness.sync_skills()
        harness.sync_subagents()
        harness.sync_rules()

    skills_mock.assert_called_once_with(
        home / "skills", tmp_path / "common" / "skills", tmp_path / "cursor" / "skills"
    )
    mock.assert_any_call(tmp_path / "common" / "agents", home / "agents")
    mock.assert_any_call(tmp_path / "cursor" / ".cursorrules", home / ".cursorrules")
    assert mock.call_count == 2


def test_cursor_sync_config_is_noop(tmp_path):
    harness = Cursor(tmp_path)

    with patch("agent_synchronizer.harnesses.sync_engine.sync_target") as mock:
        harness.sync_config()

    mock.assert_not_called()


def test_codex_sync_hooks_is_noop(tmp_path):
    harness = Codex(tmp_path)

    with patch("agent_synchronizer.harnesses.sync_engine.sync_target") as mock:
        harness.sync_hooks()

    mock.assert_not_called()


def test_cursor_sync_hooks_is_noop(tmp_path):
    harness = Cursor(tmp_path)

    with patch("agent_synchronizer.harnesses.sync_engine.sync_target") as mock:
        harness.sync_hooks()

    mock.assert_not_called()


def test_opencode_sync_hooks_is_noop(tmp_path):
    harness = OpenCode(tmp_path)

    with patch("agent_synchronizer.harnesses.sync_engine.sync_target") as mock:
        harness.sync_hooks()

    mock.assert_not_called()


def test_opencode_sync_targets(tmp_path):
    home = tmp_path / "home" / "opencode"
    harness = OpenCode(tmp_path)
    harness.home = home

    with (
        patch("agent_synchronizer.harnesses.sync_engine.sync_target") as mock,
        patch(
            "agent_synchronizer.harnesses.sync_engine.sync_flattened_skills"
        ) as skills_mock,
    ):
        harness.sync_skills()
        harness.sync_subagents()
        harness.sync_config()
        harness.sync_rules()

    skills_mock.assert_called_once_with(
        home / "skills",
        tmp_path / "common" / "skills",
        tmp_path / "opencode" / "skills",
    )
    mock.assert_any_call(tmp_path / "common" / "agents", home / "agents")
    mock.assert_any_call(
        tmp_path / "opencode" / "opencode.jsonc", home / "opencode.jsonc"
    )
    mock.assert_any_call(tmp_path / "common" / "AGENTS.md", home / "AGENTS.md")
    assert mock.call_count == 3


def test_sync_calls_all_five_methods(tmp_path):
    harness = Claude(tmp_path)

    with (
        patch.object(harness, "sync_skills") as skills,
        patch.object(harness, "sync_subagents") as subagents,
        patch.object(harness, "sync_config") as config,
        patch.object(harness, "sync_rules") as rules,
        patch.object(harness, "sync_hooks") as hooks,
    ):
        harness.sync()

    skills.assert_called_once()
    subagents.assert_called_once()
    config.assert_called_once()
    rules.assert_called_once()
    hooks.assert_called_once()
