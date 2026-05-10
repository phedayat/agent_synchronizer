from argparse import ArgumentParser

from .settings import DEFAULT_TARGETS

def parse_args():
    parser = ArgumentParser(
        description="Agent Sync - Synchronize and manage agent files across providers."
    )
    parser.add_argument(
        "repo_root",
        help="Root directory of the agent repository."
    )
    parser.add_argument(
        "--targets",
        nargs="+",
        default=DEFAULT_TARGETS,
        help="List of target files and directories to synchronize."
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Perform a trial run with no changes made."
    )
    parser.add_argument(
        "--sync-report",
        action="store_true",
        help="Print a report of the synchronization process."
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Enable verbose logging output."
    )
    parser.add_argument(
        "--config",
        default=None,
        help="Path to a YAML config file describing providers.",
    )
    parser.add_argument(
        "--save-config",
        action="store_true",
        help="Save merged providers (defaults + loaded config) to --config.",
    )
    return parser.parse_args()