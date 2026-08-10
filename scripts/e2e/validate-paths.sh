#!/bin/sh

set -eu

fail() {
	printf '%s\n' "$*" >&2
	exit 1
}

require_absolute_clean_path() {
	path_name=$1
	path_value=$2
	case "$path_value" in
	"")
		fail "$path_name must not be empty"
		;;
	/*) ;;
	*)
		fail "$path_name must be absolute: $path_value"
		;;
	esac
	case "/$path_value/" in
	*'/../'* | *'/./'*)
		fail "$path_name must not contain . or .. components: $path_value"
		;;
	esac
}

# Resolve all existing symlinks without creating the missing suffix. This is
# the shell-side equivalent of the harness's resolveThroughExistingAncestor.
resolve_through_existing_ancestor() {
	path_value=$1
	while [ "$path_value" != "/" ] && [ "${path_value%/}" != "$path_value" ]; do
		path_value=${path_value%/}
	done

	probe=$path_value
	suffix=
	while [ ! -e "$probe" ] && [ ! -L "$probe" ]; do
		[ "$probe" != "/" ] || fail "no existing ancestor for $path_value"
		component=${probe##*/}
		[ -n "$component" ] || fail "cannot resolve path component in $path_value"
		suffix="/$component$suffix"
		probe=${probe%/*}
		[ -n "$probe" ] || probe=/
	done
	[ -d "$probe" ] || fail "existing path ancestor is not a directory: $probe"
	resolved_ancestor=$(CDPATH='' && cd -P "$probe" && pwd -P) || fail "cannot resolve path ancestor: $probe"
	printf '%s%s\n' "$resolved_ancestor" "$suffix"
}

path_within() {
	base=$1
	target=$2
	[ "$base" = "/" ] && return 0
	[ "$target" = "$base" ] && return 0
	case "$target" in
	"$base"/*) return 0 ;;
	*) return 1 ;;
	esac
}

: "${E2E_ROOT:?E2E_ROOT is required}"
: "${E2E_GOCACHE:?E2E_GOCACHE is required}"
: "${E2E_GOMODCACHE:?E2E_GOMODCACHE is required}"

require_absolute_clean_path E2E_ROOT "$E2E_ROOT"
require_absolute_clean_path E2E_GOCACHE "$E2E_GOCACHE"
require_absolute_clean_path E2E_GOMODCACHE "$E2E_GOMODCACHE"

script_root=$(CDPATH='' && cd -P "$(dirname "$0")/../.." && pwd -P) || fail "cannot resolve repository root from $0"
repository_root=$(resolve_through_existing_ancestor "$script_root")
repository_e2e_root=$(resolve_through_existing_ancestor "$repository_root/.local/e2e")
slash_tmp_root=$(resolve_through_existing_ancestor /tmp)
e2e_root=$(resolve_through_existing_ancestor "$E2E_ROOT")
go_cache=$(resolve_through_existing_ancestor "$E2E_GOCACHE")
go_mod_cache=$(resolve_through_existing_ancestor "$E2E_GOMODCACHE")

if ! path_within "$repository_e2e_root" "$e2e_root" &&
	! path_within "$slash_tmp_root" "$e2e_root"; then
	fail "E2E_ROOT must resolve under $repository_e2e_root or $slash_tmp_root: $e2e_root"
fi
if ! path_within "$e2e_root" "$go_cache"; then
	fail "E2E_GOCACHE must resolve under E2E_ROOT $e2e_root: $go_cache"
fi
if ! path_within "$e2e_root" "$go_mod_cache"; then
	fail "E2E_GOMODCACHE must resolve under E2E_ROOT $e2e_root: $go_mod_cache"
fi
