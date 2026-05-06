from argparse import ArgumentParser
from pathlib import Path

from .utils.logging import Logger
from .settings import DEFAULT_TARGETS, SUPPORTED_PROVIDERS
from .providers import (
    discover_files, 
    move,
    generate_sync_report,
    symlink,
)

logger = Logger("main")

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
    return parser.parse_args()


def main(
    targets: list[str],
    dry_run: bool,
    sync_report: bool,
    verbose: bool,
    repo_root: str,
):
    logger.set_verbosity(verbose)

    repo = Path(repo_root)
    if not repo.exists():
        logger.error(f"Repository root {repo} does not exist.")
        raise FileNotFoundError(f"Repository root {repo} does not exist.")

    logger.info(f"Discovering files in {SUPPORTED_PROVIDERS.values()} for targets {targets}")

    files_found = list(discover_files(SUPPORTED_PROVIDERS, targets))
    
    if not dry_run:
        for provider, file in files_found:
            logger.info(f"Moving {file} to {repo / file.name}")
            move(file, repo / file.name)

    repo_files = list(filter(lambda x: x.name in targets, repo.iterdir()))
    if not dry_run:
        for provider in SUPPORTED_PROVIDERS.values():
            for file in repo_files:
                logger.info(f"Symlinking {file} to {provider / file.name}")
                symlink(file, provider / file.name)

    if sync_report:
        print(generate_sync_report(files_found, repo_files))


def cli():
    args = parse_args()
    main(
        args.targets,
        args.dry_run,
        args.sync_report,
        args.verbose,
        args.repo_root,
    )


if __name__ == "__main__":
    cli()
