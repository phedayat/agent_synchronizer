from pathlib import Path

from .args import parse_args
from .harnesses import ALL_HARNESSES
from .utils.logging import Logger

logger = Logger("main", verbose=True)


def main(repo_root: str) -> None:
    repo = Path(repo_root).resolve()
    if not repo.exists():
        logger.error(f"Repository root {repo} does not exist")
        raise FileNotFoundError(f"Repository root {repo} does not exist")
    if not repo.is_dir():
        logger.error(f"Repository root {repo} must be a directory")
        raise NotADirectoryError(f"Repository root {repo} must be a directory")

    for harness_cls in ALL_HARNESSES:
        harness = harness_cls(repo)
        logger.info(f"Syncing {harness.name}")
        harness.sync()


def cli():
    args = parse_args()
    main(args.repo_root)


if __name__ == "__main__":
    cli()
