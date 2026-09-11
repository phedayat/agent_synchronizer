package synclog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// LevelCritical is a custom slog level above LevelError, mirroring Python's
// logging.CRITICAL as used by utils/logging.py.
const LevelCritical = slog.LevelError + 4

var levelNames = map[slog.Level]string{
	slog.LevelInfo:  "INFO",
	slog.LevelWarn:  "WARNING",
	slog.LevelError: "ERROR",
	LevelCritical:   "CRITICAL",
}

// handler renders records as "asctime - name - levelname - message", matching
// the literal formatter string in utils/logging.py:
// "%(asctime)s - %(name)s - %(levelname)s - %(message)s".
type handler struct {
	w     io.Writer
	name  string
	level *slog.LevelVar
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	ts := r.Time.Format("2006-01-02 15:04:05") + fmt.Sprintf(",%03d", r.Time.Nanosecond()/1e6)
	levelName, ok := levelNames[r.Level]
	if !ok {
		levelName = r.Level.String()
	}
	_, err := fmt.Fprintf(h.w, "%s - %s - %s - %s\n", ts, h.name, levelName, r.Message)
	return err
}

func (h *handler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *handler) WithGroup(_ string) slog.Handler      { return h }

// Logger mirrors utils/logging.py's Logger class: a named, per-instance
// wrapper with independently toggleable verbosity.
type Logger struct {
	logger *slog.Logger
	level  *slog.LevelVar
}

// New constructs a Logger writing to stderr, matching Python's
// logging.StreamHandler() default stream.
func New(name string, verbose bool) *Logger {
	return newWithWriter(name, verbose, os.Stderr)
}

func newWithWriter(name string, verbose bool, w io.Writer) *Logger {
	level := new(slog.LevelVar)
	l := &Logger{
		logger: slog.New(&handler{w: w, name: name, level: level}),
		level:  level,
	}
	l.SetVerbosity(verbose)
	return l
}

// SetVerbosity maps verbose=true to INFO and verbose=false to WARNING, per
// utils/logging.py's set_verbosity.
func (l *Logger) SetVerbosity(verbose bool) {
	if verbose {
		l.level.Set(slog.LevelInfo)
	} else {
		l.level.Set(slog.LevelWarn)
	}
}

func (l *Logger) Info(msg string) {
	l.logger.Log(context.Background(), slog.LevelInfo, msg)
}

func (l *Logger) Warning(msg string) {
	l.logger.Log(context.Background(), slog.LevelWarn, msg)
}

func (l *Logger) Error(msg string) {
	l.logger.Log(context.Background(), slog.LevelError, msg)
}

func (l *Logger) Critical(msg string) {
	l.logger.Log(context.Background(), LevelCritical, msg)
}
