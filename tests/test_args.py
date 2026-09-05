from agent_synchronizer.args import parse_args


def test_parse_args_defaults(monkeypatch):
    monkeypatch.setattr("sys.argv", ["agent-synchronizer", "/repo"])

    args = parse_args()

    assert args.repo_root == "/repo"
