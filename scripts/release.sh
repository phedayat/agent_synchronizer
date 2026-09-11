#!/usr/bin/env bash
# Tags the current main commit and pushes it, triggering release.yml/GoReleaser.
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 vX.Y.Z" >&2
  exit 1
fi

tag="$1"
if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "error: '$tag' is not a valid vX.Y.Z tag" >&2
  exit 1
fi

branch=$(git rev-parse --abbrev-ref HEAD)
if [[ "$branch" != "main" ]]; then
  echo "error: must be run from main (currently on '$branch')" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "error: working tree is not clean" >&2
  exit 1
fi

git fetch origin main
if [[ "$(git rev-parse HEAD)" != "$(git rev-parse origin/main)" ]]; then
  echo "error: local main is not up to date with origin/main" >&2
  exit 1
fi

git tag "$tag"
git push origin "$tag"
echo "pushed $tag; release.yml will build and publish the GitHub Release"
