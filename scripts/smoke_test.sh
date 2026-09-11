#!/usr/bin/env bash
# Manual smoke test for the sync CLI against a scratch repo + scratch $HOME.
set -euo pipefail

WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT

REPO="$WORKDIR/repo"
export HOME="$WORKDIR/home"

mkdir -p "$REPO"
# Resolve symlinks (e.g. macOS /var -> /private/var) so string comparisons
# against readlink output match the resolved paths the CLI stores.
REPO="$(cd "$REPO" && pwd -P)"
mkdir -p "$HOME/.claude/skills/bar"
echo "unique-claude-md-content" >"$HOME/.claude/CLAUDE.md"

cd "$(dirname "$0")/.."
go build -o "$WORKDIR/agent-synchronizer" ./cmd/agent-synchronizer

set +o pipefail
yes y | "$WORKDIR/agent-synchronizer" "$REPO"
status=${PIPESTATUS[1]}
set -o pipefail
[ "$status" -eq 0 ] || {
    echo "FAIL: agent-synchronizer exited $status" >&2
    exit 1
}

fail() {
    echo "FAIL: $1" >&2
    exit 1
}

[ -d "$REPO/claude/skills/bar" ] || fail "bar/ was not absorbed into claude/skills/"

[ -L "$HOME/.claude/CLAUDE.md" ] || fail "claude/CLAUDE.md is not a symlink"
[ "$(readlink "$HOME/.claude/CLAUDE.md")" = "$REPO/claude/CLAUDE.md" ] \
    || fail "claude/CLAUDE.md does not point at the repo"

[ ! -L "$HOME/.claude/skills" ] || fail "skills/ should be a real directory of per-skill symlinks, not a whole-directory symlink"
[ -L "$HOME/.claude/skills/bar" ] || fail "bar is not symlinked under skills/"
[ "$(readlink "$HOME/.claude/skills/bar")" = "$REPO/claude/skills/bar" ] \
    || fail "skills/bar does not point at claude/skills/bar"

echo "OK"
