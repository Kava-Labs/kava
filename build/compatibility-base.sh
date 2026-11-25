#!/bin/sh
set -eu

base=$(git rev-parse --verify "$1^{commit}")
kind=ref
# A comparison across the master restoration must use the released chain API.
# On subsequent changes, use the caller's normal parent or merge base.
old_master=2d37fb9d74bb526e30bb75b8b3b32006e6e8d9ae
if git merge-base --is-ancestor "$base" "$old_master" && git merge-base --is-ancestor "$old_master" HEAD; then
  base=v0.28.2
  kind=tag
fi

if [ "${2:-}" = --buf ]; then
  # Buf clones a single branch. An independent release must be fetched by tag.
  printf '%s=%s\n' "$kind" "$base"
else
  echo "$base"
fi
