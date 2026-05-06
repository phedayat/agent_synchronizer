from pathlib import Path

from .args import parse_args
from .utils.logging import Logger
from .settings import SUPPORTED_PROVIDERS
from .providers import (
    discover_files, 
    move,
    generate_sync_report,
    symlink,
)

logger = Logger("main")

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
        logger.error(f"Repository root {repo} does not exist")
        raise FileNotFoundError(f"Repository root {repo} does not exist")
    if not repo.is_dir():
        logger.error(f"Repository root {repo} must be a directory")
        raise NotADirectoryError(f"Repository root {repo} must be a directory")

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
        logger.info("Generating sync report...")
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
