from argparse import ArgumentParser


def parse_args():
    parser = ArgumentParser(
        description="Agent Sync - Synchronize and manage agent files across providers."
    )
    parser.add_argument("repo_root", help="Root directory of the agent repository.")
    return parser.parse_args()
