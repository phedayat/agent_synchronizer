#!/usr/bin/env bash
# Manual smoke test for the sync CLI against a scratch repo + scratch $HOME.
set -euo pipefail

WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT

REPO="$WORKDIR/repo"
export HOME="$WORKDIR/home"

mkdir -p "$REPO"
mkdir -p "$HOME/.claude/skills/bar"
echo "unique-claude-md-content" >"$HOME/.claude/CLAUDE.md"

cd "$(dirname "$0")/.."
yes y | uv run agent-synchronizer "$REPO"

fail() {
    echo "FAIL: $1" >&2
    exit 1
}

[ -d "$REPO/common/skills/bar" ] || fail "bar/ was not moved into common/skills/"

[ -L "$HOME/.claude/CLAUDE.md" ] || fail "claude/CLAUDE.md is not a symlink"
[ "$(readlink "$HOME/.claude/CLAUDE.md")" = "$REPO/claude/CLAUDE.md" ] \
    || fail "claude/CLAUDE.md does not point at the repo"

[ -L "$HOME/.claude/skills" ] || fail "common/skills/ is not symlinked from the harness side"
[ "$(readlink "$HOME/.claude/skills")" = "$REPO/common/skills" ] \
    || fail "claude/skills does not point at common/skills/"

echo "OK"
