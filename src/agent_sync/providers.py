from pathlib import Path


class Provider:
    def __init__(self, name: str, path: str, files: list[str]):
        self.name = name
        self.path = Path(path)
        self.files = files
