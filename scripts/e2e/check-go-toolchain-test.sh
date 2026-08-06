#!/bin/sh

set -eu

script_root=$(CDPATH='' && cd -P "$(dirname "$0")" && pwd -P)
check_script="$script_root/check-go-toolchain.sh"
fixture_root=$(mktemp -d "${TMPDIR:-/tmp}/panacea-e2e-go-toolchain.XXXXXX")
trap 'rm -rf "$fixture_root"' EXIT

make_fixture() {
	name=$1
	compiler_version=$2
	root="$fixture_root/$name/goroot"
	tool_dir="$root/pkg/tool/test_arch"
	mkdir -p "$root/bin" "$tool_dir"
	cat >"$root/bin/go" <<'EOF'
#!/bin/sh
case "${1:-}:${2:-}" in
env:GOVERSION) printf '%s\n' "${FAKE_GOVERSION:?}" ;;
env:GOROOT) printf '%s\n' "${FAKE_GOROOT:?}" ;;
env:GOTOOLDIR) printf '%s\n' "${FAKE_GOTOOLDIR:?}" ;;
env:GOBIN) printf '%s\n' "${FAKE_GOBIN:-}" ;;
version:) printf 'go version %s test/arch\n' "${FAKE_GO_VERSION:?}" ;;
*) printf 'unexpected fake go arguments: %s\n' "$*" >&2; exit 2 ;;
esac
EOF
	cat >"$tool_dir/compile" <<EOF
#!/bin/sh
printf '%s\n' '$compiler_version'
EOF
	chmod +x "$root/bin/go" "$tool_dir/compile"
	printf '%s\n' "$root"
}

run_check() {
	root=$1
	output=$2
	fake_go_version=${3:-go1.26.5}
	fake_goversion=${4:-go1.26.5}
	fake_goroot=${5:-$root}
	fake_tooldir=${6:-$root/pkg/tool/test_arch}
	FAKE_GO_VERSION="$fake_go_version" \
		FAKE_GOVERSION="$fake_goversion" \
		FAKE_GOROOT="$fake_goroot" \
		FAKE_GOTOOLDIR="$fake_tooldir" \
		FAKE_GOBIN="$root/bin" \
		E2E_GO_VERSION=1.26.5 \
		E2E_GOTOOLCHAIN=local \
		E2E_GO_BINARY="$root/bin/go" \
		sh "$check_script" >"$output" 2>&1
}

matching_root=$(make_fixture matching 'compile version go1.26.5')
run_check "$matching_root" "$fixture_root/matching.out"
grep -q 'E2E Go toolchain: go1.26.5' "$fixture_root/matching.out"

mixed_root=$(make_fixture mixed 'compile version go1.25.7')
if run_check "$mixed_root" "$fixture_root/mixed.out"; then
	printf 'mixed compiler fixture unexpectedly passed\n' >&2
	exit 1
fi
grep -q 'compiler does not match' "$fixture_root/mixed.out"
grep -q 'compile version go1.25.7' "$fixture_root/mixed.out"
grep -q 'unset GOROOT GOBIN' "$fixture_root/mixed.out"

executable_mismatch_root=$(make_fixture executable-mismatch 'compile version go1.26.5')
if run_check "$executable_mismatch_root" "$fixture_root/executable-mismatch.out" go1.25.7 go1.26.5; then
	printf 'mismatched go executable fixture unexpectedly passed\n' >&2
	exit 1
fi
grep -q 'selected Go command is inconsistent' "$fixture_root/executable-mismatch.out"
grep -q 'go version:    go1.25.7' "$fixture_root/executable-mismatch.out"

goversion_mismatch_root=$(make_fixture goversion-mismatch 'compile version go1.26.5')
if run_check "$goversion_mismatch_root" "$fixture_root/goversion-mismatch.out" go1.26.5 go1.25.7; then
	printf 'mismatched GOVERSION fixture unexpectedly passed\n' >&2
	exit 1
fi
grep -q 'selected Go command is inconsistent' "$fixture_root/goversion-mismatch.out"
grep -q 'GOVERSION:     go1.25.7' "$fixture_root/goversion-mismatch.out"

split_root=$(make_fixture split-root 'compile version go1.26.5')
split_toolchain_root=$(make_fixture split-toolchain 'compile version go1.26.5')
if run_check \
	"$split_root" \
	"$fixture_root/split-toolchain.out" \
	go1.26.5 \
	go1.26.5 \
	"$split_root" \
	"$split_toolchain_root/pkg/tool/test_arch"; then
	printf 'GOTOOLDIR outside GOROOT fixture unexpectedly passed\n' >&2
	exit 1
fi
grep -q 'GOTOOLDIR .* is outside GOROOT' "$fixture_root/split-toolchain.out"

printf 'check-go-toolchain fixtures passed\n'
