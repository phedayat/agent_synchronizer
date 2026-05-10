from pathlib import Path

from .args import parse_args
from .utils.logging import Logger
from .providers import supported_providers, Provider
from .sync_engine import (
    enumerate_targets,
    move,
    generate_sync_report,
    symlink,
)

logger = Logger("main")


Expected = list[tuple[Provider, Path, Path, bool]]


def _move_files(expected: Expected, dry_run: bool) -> None:
    if dry_run:
        return

    for _, source, dest, _ in expected:
        if not source.exists():
            continue
        dest.parent.mkdir(parents=True, exist_ok=True)
        logger.info(f"Moving {source} to {dest}")
        move(source, dest)


def _symlink_files(expected: Expected, dry_run: bool) -> None:
    if dry_run:
        return

    for _, source, dest, _ in expected:
        logger.info(f"Symlinking {source} -> {dest}")
        symlink(dest, source)


def main(
    targets: list[str],
    dry_run: bool,
    sync_report: bool,
    verbose: bool,
    repo_root: str,
):
    logger.set_verbosity(verbose)

    print(f"Dry run: {dry_run}")
    print(f"Sync report: {sync_report}")
    print(f"Verbose: {verbose}")
    print(f"Repo root: {repo_root}")
    print(f"Targets: {targets}")

    repo = Path(repo_root)
    if not repo.exists():
        logger.error(f"Repository root {repo} does not exist")
        raise FileNotFoundError(f"Repository root {repo} does not exist")
    if not repo.is_dir():
        logger.error(f"Repository root {repo} must be a directory")
        raise NotADirectoryError(f"Repository root {repo} must be a directory")

    logger.info(f"Discovering files in {[p.path for p in supported_providers]} for targets {targets}")
    expected = enumerate_targets(supported_providers, targets, repo)

    logger.info("Moving files to repo")
    _move_files(expected, dry_run)

    logger.info("Linking files from repo back to providers")
    _symlink_files(expected, dry_run)

    if sync_report:
        logger.info("Generating sync report")
        files_found = [(provider, source, specific) for provider, source, _, specific in expected if source.exists()]
        repo_files = [dest for _, _, dest, _ in expected if dest.exists()]
        
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
