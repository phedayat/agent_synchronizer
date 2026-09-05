import shutil
from pathlib import Path

from .utils.logging import Logger

logger = Logger("sync_engine", verbose=True)


def _approve(message: str) -> bool:
    return input(message + " (y/n): ").strip().lower() == "y"


def move(src: Path, dest: Path):
    """
    Moves the file or directory 'src' to 'dest'.
    If 'dest' already exists, logs the fact and returns without moving.
    """
    if src.is_symlink():
        logger.info(f"Source path {src} is a symlink.")
        return
    if dest.exists():
        logger.info(f"Destination path {dest} already exists.")
        return

    dest.parent.mkdir(parents=True, exist_ok=True)
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
        if not _approve(f"Replace existing {dest} with a symlink to {src}?"):
            logger.info(f"Leaving {dest} in place.")
            return
        if dest.is_symlink():
            dest.unlink()
        elif dest.is_dir():
            shutil.rmtree(dest)
        else:
            dest.unlink()

    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.symlink_to(src, target_is_directory=src.is_dir())


def _absorb(src: Path, dest: Path) -> bool:
    """Move each direct child of dest missing from src into src, one level deep.

    Returns False if any child could not be fully absorbed (e.g. its move was
    declined and it remains in place at dest), meaning dest still holds real
    content and must not be destroyed by a subsequent symlink() call.
    """
    if not dest.exists() or dest.is_symlink() or not dest.is_dir():
        return True
    existing = (
        {child.name for child in src.iterdir()}
        if src.exists() and src.is_dir()
        else set()
    )
    fully_absorbed = True
    for child in dest.iterdir():
        if child.name not in existing:
            move(child, src / child.name)
            if child.exists():
                fully_absorbed = False
    return fully_absorbed


def sync_target(src: Path, dest: Path):
    if dest.is_symlink():
        logger.info(f"Destination path {dest} is already a symlink.")
        return
    if dest.exists() and not src.exists():
        move(dest, src)
        symlink(src, dest)
        return
    if dest.exists() and src.exists():
        if _absorb(src, dest):
            symlink(src, dest)
        else:
            logger.info(f"Leaving {dest} in place: not fully absorbed into {src}.")
        return
    symlink(src, dest)


def _iter_skill_dirs(root: Path):
    """Yield every directory under root that directly contains SKILL.md,
    at any depth, without descending into a directory once matched."""
    if not root.exists():
        return
    if (root / "SKILL.md").exists():
        yield root
        return
    for child in sorted(p for p in root.iterdir() if p.is_dir()):
        yield from _iter_skill_dirs(child)


def _collect_skills(*sources: Path) -> dict[str, Path]:
    """Map skill name -> directory across sources, later sources win."""
    skills: dict[str, Path] = {}
    for source in sources:
        for skill_dir in _iter_skill_dirs(source):
            skills[skill_dir.name] = skill_dir
    return skills


def sync_flattened_skills(dest: Path, *sources: Path) -> None:
    """
    Flattens one or more (possibly grouped) skill source roots into dest,
    so every skill directory sits directly under dest. Sources are merged
    in order; a same-named skill in a later source overrides an earlier one
    entirely. Real, unmanaged content already at dest is absorbed into the
    last source before being symlinked back (mirrors sync_target's
    onboarding behavior, applied per-skill instead of per-directory).
    """
    desired = _collect_skills(*sources)

    if dest.is_symlink():
        dest.unlink()
    dest.mkdir(parents=True, exist_ok=True)

    for child in list(dest.iterdir()):
        if child.name in desired or child.is_symlink():
            continue
        target = sources[-1] / child.name
        move(child, target)
        desired[child.name] = target

    for name, src in desired.items():
        symlink(src, dest / name)
