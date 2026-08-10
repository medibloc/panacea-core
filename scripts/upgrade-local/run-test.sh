#!/bin/sh

set -eu

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd -P)
runner="$script_dir/run.sh"
fixture_root=$(mktemp -d "/tmp/panacea-upgrade-root-test.XXXXXX")
mkdir -p "$repo_root/.local"
repo_fixture=$(mktemp -d "$repo_root/.local/upgrade-root-test.XXXXXX")
trap 'rm -rf -- "$fixture_root" "$repo_fixture"' EXIT HUP INT TERM

expect_refused() {
  unsafe_root=$1
  output_file=$2
  if PANACEA_REHEARSAL_ROOT="$unsafe_root" "$runner" clean >"$output_file" 2>&1; then
    printf 'unsafe rehearsal root unexpectedly passed: %s\n' "$unsafe_root" >&2
    exit 1
  fi
  grep -Eq 'cannot resolve rehearsal root safely|refusing to remove or initialize' "$output_file"
}

# A lexical prefix must not permit traversal back to the repository root.
expect_refused "$repo_root/.local/upgrade/../.." "$fixture_root/traversal.out"
[ -d "$repo_root" ]

# A symlink below .local must not permit deletion of its repository target.
ln -s "$repo_root" "$repo_fixture/repository-link"
expect_refused "$repo_fixture/repository-link" "$fixture_root/symlink.out"
[ -d "$repo_root/.git" ]

# An actual descendant of the temporary directory remains a valid root.
allowed_root="$fixture_root/allowed"
mkdir -p "$allowed_root"
: >"$allowed_root/marker"
PANACEA_REHEARSAL_ROOT="$allowed_root" "$runner" clean >"$fixture_root/allowed.out" 2>&1
[ ! -e "$allowed_root" ]

if grep -Eq 'keys add .*key\.json' "$runner"; then
  printf 'upgrade rehearsal still persists key-generation JSON\n' >&2
  exit 1
fi

printf 'upgrade rehearsal root safety fixtures passed\n'
