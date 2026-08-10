#!/bin/sh

set -eu

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <git-commit-ish> <empty-destination-directory>" >&2
  exit 2
fi

source_ref=$1
destination=$2

if [ -L "$destination" ]; then
  echo "refusing symlink source destination: $destination" >&2
  exit 2
fi
if [ -e "$destination" ]; then
  if [ ! -d "$destination" ]; then
    echo "source destination is not a directory: $destination" >&2
    exit 2
  fi
  if [ -n "$(find "$destination" -mindepth 1 -print -quit)" ]; then
    echo "source destination must be empty: $destination" >&2
    exit 2
  fi
else
  mkdir -p "$destination"
fi

source_commit=$(git rev-parse --verify "$source_ref^{commit}")
archive_dir=$(mktemp -d "${TMPDIR:-/tmp}/panacea-git-source.XXXXXX")
archive_path="$archive_dir/source.tar"
cleanup_stage_git_source() {
  rm -f "$archive_path"
  rmdir "$archive_dir" 2>/dev/null || true
}
trap cleanup_stage_git_source EXIT HUP INT TERM

git archive --format=tar --output="$archive_path" "$source_commit"
tar -xf "$archive_path" -C "$destination"
printf '%s\n' "$source_commit"
