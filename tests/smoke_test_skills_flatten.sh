#!/usr/bin/env bash
# Manual smoke test for skill flattening: grouped + harness-specific skills
# in the central repo must land flat, one directory per skill, under the
# harness's skills/ folder, with harness-specific skills winning collisions.
set -euo pipefail

WORKDIR=$(mktemp -d)
trap 'rm -rf "$WORKDIR"' EXIT

REPO="$WORKDIR/repo"
HOME_DIR="$WORKDIR/home"

mkdir -p "$REPO/common/skills/group_1/skill_11"
echo "# skill_11 (common, grouped)" >"$REPO/common/skills/group_1/skill_11/SKILL.md"

mkdir -p "$REPO/common/skills/skill_21"
echo "# skill_21 (common)" >"$REPO/common/skills/skill_21/SKILL.md"

mkdir -p "$REPO/claude/skills/skill_11"
echo "# skill_11 (claude override)" >"$REPO/claude/skills/skill_11/SKILL.md"

cd "$(dirname "$0")/.."

uv run python - "$REPO" "$HOME_DIR" <<'PY'
import sys
from pathlib import Path

from agent_synchronizer.harnesses import Claude

repo = Path(sys.argv[1])
home = Path(sys.argv[2])

harness = Claude(repo)
harness.home = home / ".claude"
harness.sync_skills()
PY

fail() {
    echo "FAIL: $1" >&2
    exit 1
}

SKILLS="$HOME_DIR/.claude/skills"

[ -d "$SKILLS" ] || fail "skills/ was not created"
[ ! -e "$SKILLS/group_1" ] || fail "group_1/ should not survive flattening"

[ -L "$SKILLS/skill_11" ] || fail "skill_11 is not a symlink"
resolved_11=$(cd "$SKILLS/skill_11" && pwd -P)
expected_11=$(cd "$REPO/claude/skills/skill_11" && pwd -P)
[ "$resolved_11" = "$expected_11" ] \
    || fail "skill_11 should resolve to the claude-specific skill, not common"

[ -L "$SKILLS/skill_21" ] || fail "skill_21 is not a symlink"
resolved_21=$(cd "$SKILLS/skill_21" && pwd -P)
expected_21=$(cd "$REPO/common/skills/skill_21" && pwd -P)
[ "$resolved_21" = "$expected_21" ] \
    || fail "skill_21 should resolve to the common skill"

echo "OK"
