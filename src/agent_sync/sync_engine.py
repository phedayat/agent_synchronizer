from pathlib import Path

from .providers import Provider

from .utils.logging import Logger

logger = Logger("sync_engine", verbose=True)


def _approve(message: str) -> bool:
    return input(message + " (y/n): ").strip().lower() == "y"


def enumerate_targets(
    providers: list[Provider], targets: list[str], repo: Path
) -> list[tuple[Provider, Path, Path, bool]]:
    expected = []
    for provider in providers:
        for target in targets + provider.files:
            specific = target in provider.files
            source = provider.path / target
            dest = repo / provider.name / target if specific else repo / target
            expected.append((provider, source, dest, specific))
    return expected


def move(src: Path, dest: Path):
    """
    Moves the directory 'src' to 'dest'.
    If 'dest' exists, raises an exception.
    """
    if src.is_symlink():
        logger.info(f"Source path {src} is a symlink.")
        return
    if dest.exists():
        logger.info(FileExistsError(f"Destination path {dest} already exists."))
        return

    if _approve(f"Do you want to move {src}?"):
        logger.info(f"Moving {src} to {dest}")
        src.move(dest)
    else:
        logger.info(f"Copying {src} to {dest}")
        src.copy(dest)


def symlink(src: Path, dest: Path):
    if not src.exists():
        logger.info(f"Source path {src} does not exist.")
        return
    if dest.is_symlink() and dest.resolve() == src.resolve():
        logger.info(f"Symlink already correct: {dest} -> {src}")
        return
    elif dest.exists():
        logger.info(FileExistsError(f"Destination path {dest} already exists and is not a symlink."))
        return

    dest.symlink_to(src, target_is_directory=src.is_dir())


def generate_sync_report(files_found: list[tuple[str, Path]], repo_files: list[Path]):
    report = ""
    if not files_found:
        return "No files found"
    if not repo_files:
        return "No repo files found"
    for provider, file, specific in files_found:
        report += f"Provider: {provider.name}, File: {file}, Specific: {specific}\n"
    for repo_file in repo_files:
        report += f"Repo file: {repo_file}\n"
    report += f"Total files: {len(files_found)}\n"
    report += f"Total repo files: {len(repo_files)}\n"
    return report
