import logging
from unittest.mock import patch

from agent_sync.utils.logging import Logger


def test_logger_initializes_with_warning_level_by_default():
    logger_name = "test_logging_default_level"
    logging.getLogger(logger_name).handlers.clear()

    wrapped = Logger(name=logger_name)

    assert wrapped.logger.level == logging.WARNING


def test_logger_initializes_with_info_level_when_verbose_true():
    logger_name = "test_logging_verbose_level"
    logging.getLogger(logger_name).handlers.clear()

    wrapped = Logger(name=logger_name, verbose=True)

    assert wrapped.logger.level == logging.INFO


def test_get_logger_adds_single_handler_when_missing():
    logger_name = "test_logging_single_handler"
    base = logging.getLogger(logger_name)
    base.handlers.clear()

    wrapped = Logger(name=logger_name)

    assert len(wrapped.logger.handlers) == 1


def test_get_logger_does_not_add_duplicate_handlers():
    logger_name = "test_logging_no_duplicate_handlers"
    base = logging.getLogger(logger_name)
    base.handlers.clear()

    first = Logger(name=logger_name)
    first_count = len(first.logger.handlers)

    second = Logger(name=logger_name)

    assert first_count == 1
    assert len(second.logger.handlers) == 1


def test_set_verbosity_sets_info_level_for_true():
    logger_name = "test_logging_set_verbosity_true"
    logging.getLogger(logger_name).handlers.clear()

    wrapped = Logger(name=logger_name)
    wrapped.set_verbosity(True)

    assert wrapped.logger.level == logging.INFO


def test_set_verbosity_sets_warning_level_for_false():
    logger_name = "test_logging_set_verbosity_false"
    logging.getLogger(logger_name).handlers.clear()

    wrapped = Logger(name=logger_name, verbose=True)
    wrapped.set_verbosity(False)

    assert wrapped.logger.level == logging.WARNING


def test_info_delegates_to_underlying_logger():
    logger_name = "test_logging_info_delegate"
    logging.getLogger(logger_name).handlers.clear()
    wrapped = Logger(name=logger_name)

    with patch.object(wrapped.logger, "info") as info_mock:
        wrapped.info("hello")

    info_mock.assert_called_once_with("hello")


def test_warning_delegates_to_underlying_logger():
    logger_name = "test_logging_warning_delegate"
    logging.getLogger(logger_name).handlers.clear()
    wrapped = Logger(name=logger_name)

    with patch.object(wrapped.logger, "warning") as warning_mock:
        wrapped.warning("hello")

    warning_mock.assert_called_once_with("hello")


def test_error_delegates_to_underlying_logger():
    logger_name = "test_logging_error_delegate"
    logging.getLogger(logger_name).handlers.clear()
    wrapped = Logger(name=logger_name)

    with patch.object(wrapped.logger, "error") as error_mock:
        wrapped.error("hello")

    error_mock.assert_called_once_with("hello")


def test_critical_delegates_to_underlying_logger():
    logger_name = "test_logging_critical_delegate"
    logging.getLogger(logger_name).handlers.clear()
    wrapped = Logger(name=logger_name)

    with patch.object(wrapped.logger, "critical") as critical_mock:
        wrapped.critical("hello")

    critical_mock.assert_called_once_with("hello")
