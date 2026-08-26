from abc import ABC, abstractmethod
from pathlib import Path

from . import sync_engine


class Harness(ABC):
    name: str
    home: Path

    def __init__(self, repo: Path):
        self.repo = repo

    @property
    def repo_dir(self) -> Path:
        return self.repo / self.name

    @property
    def common_dir(self) -> Path:
        return self.repo / "common"

    @abstractmethod
    def sync_skills(self) -> None: ...

    @abstractmethod
    def sync_subagents(self) -> None: ...

    @abstractmethod
    def sync_config(self) -> None: ...

    @abstractmethod
    def sync_rules(self) -> None: ...

    def sync(self) -> None:
        self.sync_skills()
        self.sync_subagents()
        self.sync_config()
        self.sync_rules()


class Claude(Harness):
    name = "claude"

    def __init__(self, repo: Path):
        super().__init__(repo)
        self.home = Path.home() / ".claude"

    def sync_skills(self) -> None:
        sync_engine.sync_target(self.common_dir / "skills", self.home / "skills")

    def sync_subagents(self) -> None:
        sync_engine.sync_target(self.common_dir / "agents", self.home / "agents")

    def sync_config(self) -> None:
        sync_engine.sync_target(
            self.repo_dir / "settings.json", self.home / "settings.json"
        )

    def sync_rules(self) -> None:
        sync_engine.sync_target(self.repo_dir / "CLAUDE.md", self.home / "CLAUDE.md")


class Codex(Harness):
    name = "codex"

    def __init__(self, repo: Path):
        super().__init__(repo)
        self.home = Path.home() / ".codex"

    def sync_skills(self) -> None:
        sync_engine.sync_target(self.common_dir / "skills", self.home / "skills")

    def sync_subagents(self) -> None:
        sync_engine.sync_target(self.common_dir / "agents", self.home / "agents")

    def sync_config(self) -> None:
        sync_engine.sync_target(
            self.repo_dir / "config.toml", self.home / "config.toml"
        )

    def sync_rules(self) -> None:
        sync_engine.sync_target(self.common_dir / "AGENTS.md", self.home / "AGENTS.md")


class Cursor(Harness):
    name = "cursor"

    def __init__(self, repo: Path):
        super().__init__(repo)
        self.home = Path.home() / ".cursor"

    def sync_skills(self) -> None:
        sync_engine.sync_target(self.common_dir / "skills", self.home / "skills")

    def sync_subagents(self) -> None:
        sync_engine.sync_target(self.common_dir / "agents", self.home / "agents")

    def sync_config(self) -> None:
        pass

    def sync_rules(self) -> None:
        sync_engine.sync_target(
            self.repo_dir / ".cursorrules", self.home / ".cursorrules"
        )


class OpenCode(Harness):
    name = "opencode"

    def __init__(self, repo: Path):
        super().__init__(repo)
        self.home = Path.home() / ".config" / "opencode"

    def sync_skills(self) -> None:
        sync_engine.sync_target(self.common_dir / "skills", self.home / "skills")

    def sync_subagents(self) -> None:
        sync_engine.sync_target(self.common_dir / "agents", self.home / "agents")

    def sync_config(self) -> None:
        sync_engine.sync_target(
            self.repo_dir / "opencode.jsonc", self.home / "opencode.jsonc"
        )

    def sync_rules(self) -> None:
        sync_engine.sync_target(self.common_dir / "AGENTS.md", self.home / "AGENTS.md")


ALL_HARNESSES: list[type[Harness]] = [Claude, Codex, Cursor, OpenCode]
