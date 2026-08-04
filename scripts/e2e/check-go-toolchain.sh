#!/bin/sh

set -eu

fail() {
	printf '%s\n' "$*" >&2
	exit 1
}

: "${E2E_GO_VERSION:?E2E_GO_VERSION is required}"

go_command=${E2E_GO_BINARY:-go}
go_toolchain=${E2E_GOTOOLCHAIN:-local}

go_path=$(command -v "$go_command" 2>/dev/null) || fail "E2E Go executable was not found: $go_command"

go_env() {
	GOTOOLCHAIN=$go_toolchain "$go_path" env "$1"
}

go_version=$(GOTOOLCHAIN=$go_toolchain "$go_path" version 2>&1) || fail "cannot run E2E Go executable $go_path: $go_version"
go_version=${go_version#go version }
go_version=${go_version%% *}
goversion=$(go_env GOVERSION) || fail "cannot read GOVERSION from $go_path"
goroot=$(go_env GOROOT) || fail "cannot read GOROOT from $go_path"
gotooldir=$(go_env GOTOOLDIR) || fail "cannot read GOTOOLDIR from $go_path"
gobin=$(go_env GOBIN) || fail "cannot read GOBIN from $go_path"

[ -n "$goroot" ] || fail "E2E Go toolchain reported an empty GOROOT"
[ -n "$gotooldir" ] || fail "E2E Go toolchain reported an empty GOTOOLDIR"
[ -x "$gotooldir/compile" ] || fail "E2E Go compiler is missing or not executable: $gotooldir/compile"

compiler_version=$("$gotooldir/compile" -V=full 2>&1) || fail "cannot execute E2E Go compiler $gotooldir/compile: $compiler_version"
expected="go$E2E_GO_VERSION"

case "$gotooldir" in
"$goroot"/pkg/tool/*) ;;
*)
	fail "E2E Go toolchain is mixed: GOTOOLDIR $gotooldir is outside GOROOT $goroot"
	;;
esac

if [ "$go_version" != "$expected" ] || [ "$goversion" != "$expected" ]; then
	cat >&2 <<EOF
E2E requires $expected, but the selected Go command is inconsistent.
  go executable: $go_path
  go version:    $go_version
  GOVERSION:     $goversion
  GOROOT:        $goroot
  GOTOOLDIR:     $gotooldir
  GOBIN:         ${gobin:-<empty>}
  compiler:      $compiler_version
Select Go $E2E_GO_VERSION and clear stale overrides (for example: unset GOROOT GOBIN).
EOF
	exit 1
fi

case "$compiler_version" in
*"$expected"*) ;;
*)
	cat >&2 <<EOF
E2E Go compiler does not match the selected Go command.
  expected:      $expected
  go executable: $go_path
  go version:    $go_version
  GOVERSION:     $goversion
  GOROOT:        $goroot
  GOTOOLDIR:     $gotooldir
  GOBIN:         ${gobin:-<empty>}
  compiler:      $compiler_version
Select one complete Go $E2E_GO_VERSION installation and clear stale overrides (for example: unset GOROOT GOBIN).
EOF
	exit 1
	;;
esac

printf 'E2E Go toolchain: %s (%s, %s)\n' "$goversion" "$goroot" "$compiler_version"
