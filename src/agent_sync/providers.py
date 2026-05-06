import shutil

from pathlib import Path

from .utils.logging import Logger

logger = Logger("providers")

home = Path.home()

providers = {
    "cursor": home / ".cursor",
    "codex": home / ".codex",
    "claude": home / ".claude",
}

def discover_files(providers: dict[str, Path], targets: list[str]):
    for provider, path in providers.items():
        if path.exists():
            for target in targets:
                if (path / target).exists():
                    yield provider, path / target
        else:
            logger.warning(f"Provider {provider} not found")


def move(src: Path, dest: Path):
    """
    Moves the directory 'src' to 'dest' and replaces 'src' with a symlink to 'dest'.
    If 'dest' exists, raises an exception.
    """
    if src.is_symlink():
        logger.info(f"Source path {src} is a symlink.")
        return
    if dest.exists() and dest.is_dir():
        for item in src.iterdir():
            dest_item = dest / item.name
            if dest_item.exists():
                logger.info(f"Destination item {dest_item} already exists. Skipping.")
                continue
            item.replace(dest_item)
        shutil.rmtree(src)
        return
    if dest.exists():
        logger.info(FileExistsError(f"Destination path {dest} already exists."))
        return

    src.move(dest)


def symlink(src: Path, dest: Path):
    if not src.exists():
        logger.info(f"Source path {src} does not exist.")
        return
    if dest.is_symlink():
        if dest.resolve() == src.resolve():
            logger.info(f"Symlink already correct: {dest} -> {src}")
            return
        dest.unlink()
    elif dest.exists():
        logger.info(FileExistsError(f"Destination path {dest} already exists and is not a symlink."))
        return

    dest.symlink_to(src, target_is_directory=src.is_dir())


def generate_sync_report(files_found: list[tuple[str, Path]], repo_files: list[Path]):
    report = ""
    for provider, file in files_found:
        report += f"Provider: {provider}\n"
        report += f"File: {file}\n"
    for repo_file in repo_files:
        report += f"Repo file: {repo_file}\n"
    report += f"Total files: {len(files_found)}\n"
    report += f"Total repo files: {len(repo_files)}\n"
    return report