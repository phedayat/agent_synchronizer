package synclog

import (
	"bytes"
	"context"
	"log/slog"
	"regexp"
	"strings"
	"testing"
)

var lineRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3} - (\S+) - (\S+) - (.*)\n$`)

func parseLine(t *testing.T, out string) (name, level, msg string) {
	t.Helper()
	m := lineRe.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("output %q does not match expected format", out)
	}
	return m[1], m[2], m[3]
}

func TestNewDefaultsToWarning(t *testing.T) {
	var buf bytes.Buffer
	l := newWithWriter("main", false, &buf)

	l.Info("hidden")
	if buf.Len() != 0 {
		t.Fatalf("expected no output at WARNING level, got %q", buf.String())
	}

	l.Warning("shown")
	name, level, msg := parseLine(t, buf.String())
	if name != "main" || level != "WARNING" || msg != "shown" {
		t.Fatalf("unexpected fields: name=%q level=%q msg=%q", name, level, msg)
	}
}

func TestNewVerboseSetsInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	l := newWithWriter("sync_engine", true, &buf)

	l.Info("hello")
	name, level, msg := parseLine(t, buf.String())
	if name != "sync_engine" || level != "INFO" || msg != "hello" {
		t.Fatalf("unexpected fields: name=%q level=%q msg=%q", name, level, msg)
	}
}

func TestInfoDelegates(t *testing.T) {
	var buf bytes.Buffer
	l := newWithWriter("n", true, &buf)

	l.Info("hello")

	_, level, msg := parseLine(t, buf.String())
	if level != "INFO" || msg != "hello" {
		t.Fatalf("Info did not delegate correctly: level=%q msg=%q", level, msg)
	}
}

func TestWarningDelegates(t *testing.T) {
	var buf bytes.Buffer
	l := newWithWriter("n", false, &buf)

	l.Warning("hello")

	_, level, msg := parseLine(t, buf.String())
	if level != "WARNING" || msg != "hello" {
		t.Fatalf("Warning did not delegate correctly: level=%q msg=%q", level, msg)
	}
}

func TestErrorDelegates(t *testing.T) {
	var buf bytes.Buffer
	l := newWithWriter("n", false, &buf)

	l.Error("hello")

	_, level, msg := parseLine(t, buf.String())
	if level != "ERROR" || msg != "hello" {
		t.Fatalf("Error did not delegate correctly: level=%q msg=%q", level, msg)
	}
}

func TestCriticalDelegates(t *testing.T) {
	var buf bytes.Buffer
	l := newWithWriter("n", false, &buf)

	l.Critical("hello")

	_, level, msg := parseLine(t, buf.String())
	if level != "CRITICAL" || msg != "hello" {
		t.Fatalf("Critical did not delegate correctly: level=%q msg=%q", level, msg)
	}
}

func TestSetVerbosityTogglesLevel(t *testing.T) {
	var buf bytes.Buffer
	l := newWithWriter("n", true, &buf)

	l.SetVerbosity(false)
	l.Info("hidden")
	if buf.Len() != 0 {
		t.Fatalf("expected no output after SetVerbosity(false), got %q", buf.String())
	}

	l.SetVerbosity(true)
	l.Info("shown")
	if buf.Len() == 0 {
		t.Fatalf("expected output after SetVerbosity(true), got none")
	}
}

func TestNewWritesToStderr(t *testing.T) {
	l := New("n", false)
	if l.logger == nil {
		t.Fatal("expected non-nil logger from New")
	}
}

func TestHandlerFallsBackToLevelStringForUnknownLevel(t *testing.T) {
	var buf bytes.Buffer
	l := newWithWriter("n", true, &buf)

	// A level with no entry in levelNames must fall back to Level.String().
	unknown := slog.LevelError + 1
	l.logger.Log(context.Background(), unknown, "custom")

	_, level, msg := parseLine(t, buf.String())
	if msg != "custom" {
		t.Fatalf("unexpected message: %q", msg)
	}
	if !strings.Contains(level, "ERROR+1") {
		t.Fatalf("expected fallback level string containing ERROR+1, got %q", level)
	}
}

func TestHandlerWithAttrsAndWithGroupReturnSelf(t *testing.T) {
	h := &handler{w: &bytes.Buffer{}, name: "n", level: new(slog.LevelVar)}

	if got := h.WithAttrs([]slog.Attr{slog.String("k", "v")}); got != h {
		t.Fatalf("WithAttrs should return the same handler, got %v", got)
	}
	if got := h.WithGroup("g"); got != h {
		t.Fatalf("WithGroup should return the same handler, got %v", got)
	}
}
