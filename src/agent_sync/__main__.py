from pathlib import Path

from . import config as cfg
from .args import parse_args
from .providers import supported_providers
from .sync_engine import (
    enumerate_targets,
    generate_sync_report,
    move,
    symlink,
)
from .types import Expected, Provider
from .utils.logging import Logger

logger = Logger("main")


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
    config: str | None,
    save_config: bool,
):
    logger.set_verbosity(verbose)

    loaded: list[Provider] = []
    if config is not None:
        loaded = cfg.load_config(Path(config))
    if save_config:
        if config is None:
            raise ValueError("--save-config requires --config <path>")
        cfg.save_config(Path(config), loaded)

    providers = cfg.merge_providers(supported_providers, loaded)

    repo = Path(repo_root)
    if not repo.exists():
        logger.error(f"Repository root {repo} does not exist")
        raise FileNotFoundError(f"Repository root {repo} does not exist")
    if not repo.is_dir():
        logger.error(f"Repository root {repo} must be a directory")
        raise NotADirectoryError(f"Repository root {repo} must be a directory")

    logger.info(
        f"Discovering files in {[p.path for p in providers]} for targets {targets}"
    )
    expected = enumerate_targets(providers, targets, repo)

    logger.info("Moving files to repo")
    _move_files(expected, dry_run)

    logger.info("Linking files from repo back to providers")
    _symlink_files(expected, dry_run)

    if sync_report:
        logger.info("Generating sync report")
        print(generate_sync_report(expected))


def cli():
    args = parse_args()
    main(
        args.targets,
        args.dry_run,
        args.sync_report,
        args.verbose,
        args.repo_root,
        args.config,
        args.save_config,
    )


if __name__ == "__main__":
    cli()
