from agent_sync.args import parse_args
from agent_sync.settings import DEFAULT_TARGETS


def test_parse_args_defaults(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-sync", "/repo"])

    args = parse_args()

    assert args.repo_root == "/repo"
    assert args.targets == DEFAULT_TARGETS
    assert args.dry_run is False
    assert args.sync_report is False
    assert args.verbose is False


def test_parse_args_targets(monkeypatch):
    monkeypatch.setattr(
        "sys.argv",
        ["agent-sync", "/repo", "--targets", "AGENTS.md", "skills", "agents"],
    )

    args = parse_args()

    assert args.repo_root == "/repo"
    assert args.targets == ["AGENTS.md", "skills", "agents"]


def test_parse_args_dry_run(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-sync", "/repo", "--dry-run"])

    args = parse_args()

    assert args.dry_run is True


def test_parse_args_sync_report(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-sync", "/repo", "--sync-report"])

    args = parse_args()

    assert args.sync_report is True


def test_parse_args_verbose(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-sync", "/repo", "--verbose"])

    args = parse_args()

    assert args.verbose is True


def test_parse_args_all_flags_together(monkeypatch):
    monkeypatch.setattr(
        "sys.argv",
        [
            "agent-sync",
            "/repo",
            "--targets",
            "AGENTS.md",
            "CLAUDE.md",
            "--dry-run",
            "--sync-report",
            "--verbose",
        ],
    )

    args = parse_args()

    assert args.repo_root == "/repo"
    assert args.targets == ["AGENTS.md", "CLAUDE.md"]
    assert args.dry_run is True
    assert args.sync_report is True
    assert args.verbose is True
