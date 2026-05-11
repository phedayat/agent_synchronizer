import logging


class Logger:
    def __init__(self, name: str = __name__, verbose: bool = False):
        self.logger = self._get_logger(name)
        self.set_verbosity(verbose)

    def _get_logger(self, name: str = __name__):
        logger = logging.getLogger(name)

        if not logger.handlers:
            formatter = logging.Formatter(
                "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
            )
            handler = logging.StreamHandler()
            handler.setFormatter(formatter)
            logger.addHandler(handler)
        return logger

    def set_verbosity(self, verbose: bool):
        self.logger.setLevel(logging.INFO if verbose else logging.WARNING)

    def info(self, message: str):
        self.logger.info(message)

    def warning(self, message: str):
        self.logger.warning(message)

    def error(self, message: str):
        self.logger.error(message)

    def critical(self, message: str):
        self.logger.critical(message)
