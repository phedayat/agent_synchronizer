import pytest

from agent_synchronizer.args import parse_args


def test_parse_args_defaults(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-synchronizer", "/repo"])

    args = parse_args()

    assert args.repo_root == "/repo"
    assert args.targets is None
    assert args.dry_run is False
    assert args.sync_report is False
    assert args.verbose is False
    assert args.config is None
    assert args.save_config is False


def test_parse_args_targets(monkeypatch):
    monkeypatch.setattr(
        "sys.argv",
        ["agent-synchronizer", "/repo", "--targets", "AGENTS.md", "skills", "agents"],
    )

    args = parse_args()

    assert args.repo_root == "/repo"
    assert args.targets == ["AGENTS.md", "skills", "agents"]


def test_parse_args_dry_run(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-synchronizer", "/repo", "--dry-run"])

    args = parse_args()

    assert args.dry_run is True


def test_parse_args_sync_report(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-synchronizer", "/repo", "--sync-report"])

    args = parse_args()

    assert args.sync_report is True


def test_parse_args_verbose(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-synchronizer", "/repo", "--verbose"])

    args = parse_args()

    assert args.verbose is True


def test_parse_args_config(monkeypatch):
    monkeypatch.setattr(
        "sys.argv", ["agent-synchronizer", "/repo", "--config", "/path/to/config.yaml"]
    )

    args = parse_args()

    assert args.config == "/path/to/config.yaml"
    assert args.save_config is False


def test_parse_args_save_config(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-synchronizer", "/repo", "--save-config"])

    args = parse_args()

    assert args.save_config is True
    assert args.config is None


def test_parse_args_config_missing_value_raises(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-synchronizer", "/repo", "--config"])

    with pytest.raises(SystemExit):
        parse_args()


def test_parse_args_all_flags_together(monkeypatch):
    monkeypatch.setattr(
        "sys.argv",
        [
            "agent-synchronizer",
            "/repo",
            "--targets",
            "AGENTS.md",
            "CLAUDE.md",
            "--dry-run",
            "--sync-report",
            "--verbose",
            "--config",
            "/path/to/config.yaml",
            "--save-config",
        ],
    )

    args = parse_args()

    assert args.repo_root == "/repo"
    assert args.targets == ["AGENTS.md", "CLAUDE.md"]
    assert args.dry_run is True
    assert args.sync_report is True
    assert args.verbose is True
    assert args.config == "/path/to/config.yaml"
    assert args.save_config is True
