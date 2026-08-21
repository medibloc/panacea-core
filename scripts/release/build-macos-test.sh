#!/bin/sh

set -eu

test_script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P)
builder_script="$test_script_dir/build-macos.sh"
fixture_root=$(mktemp -d "${TMPDIR:-/tmp}/panacea-macos-builder.XXXXXX")
trap 'rm -rf "$fixture_root"' EXIT HUP INT TERM

fixture_repo="$fixture_root/repo"
fixture_fake_bin="$fixture_root/fake-bin"
mkdir -p "$fixture_repo/scripts/release" "$fixture_fake_bin"
cp "$builder_script" "$fixture_repo/scripts/release/build-macos.sh"
chmod +x "$fixture_repo/scripts/release/build-macos.sh"

cat >"$fixture_repo/go.mod" <<'EOF'
module example.com/panacea-macos-builder-test

go 1.26.5
EOF
cat >"$fixture_repo/go.sum" <<'EOF'
EOF
cat >"$fixture_repo/Makefile" <<'EOF'
release-build-darwin:
	@false
EOF

cat >"$fixture_fake_bin/go" <<'EOF'
#!/bin/sh
case "$*" in
version) printf 'go version go1.26.5 darwin/%s\n' "${FAKE_GO_ARCH:-arm64}" ;;
'tool compile -V=full') printf 'compile version go1.26.5\n' ;;
'mod verify') printf 'all modules verified\n' ;;
'mod vendor') mkdir -p vendor; : >vendor/modules.txt ;;
*) printf 'unexpected fake go arguments: %s\n' "$*" >&2; exit 2 ;;
esac
EOF

cat >"$fixture_fake_bin/uname" <<'EOF'
#!/bin/sh
case "${1:-}" in
-s) printf '%s\n' "${FAKE_UNAME_SYSTEM:-Darwin}" ;;
-m) printf '%s\n' "${FAKE_UNAME_MACHINE:-arm64}" ;;
*) printf '%s fixture %s\n' "${FAKE_UNAME_SYSTEM:-Darwin}" "${FAKE_UNAME_MACHINE:-arm64}" ;;
esac
EOF

cat >"$fixture_fake_bin/file" <<'EOF'
#!/bin/sh
case "${FAKE_GO_ARCH:-arm64}" in
amd64) fixture_file_arch=x86_64 ;;
arm64) fixture_file_arch=arm64 ;;
*) exit 2 ;;
esac
printf '%s: Mach-O 64-bit executable %s\n' "$1" "$fixture_file_arch"
EOF

cat >"$fixture_fake_bin/make" <<'EOF'
#!/bin/sh
fixture_version=
fixture_commit=
fixture_release_goarch=
fixture_release_output=
fixture_target=
for fixture_arg in "$@"; do
	case "$fixture_arg" in
	VERSION=*) fixture_version=${fixture_arg#VERSION=} ;;
	COMMIT=*) fixture_commit=${fixture_arg#COMMIT=} ;;
	RELEASE_GOARCH=*) fixture_release_goarch=${fixture_arg#RELEASE_GOARCH=} ;;
	RELEASE_OUTPUT=*) fixture_release_output=${fixture_arg#RELEASE_OUTPUT=} ;;
	*=*) ;;
	*) fixture_target=$fixture_arg ;;
	esac
done
[ "$fixture_target" = release-build-darwin ]
[ "$fixture_release_goarch" = "${FAKE_GO_ARCH:-arm64}" ]
[ "${GOOS:-}" = darwin ]
[ "${GOARCH:-}" = "${FAKE_GO_ARCH:-arm64}" ]
[ "${CGO_ENABLED:-}" = 0 ]
[ "${LEDGER_ENABLED:-}" = false ]
[ -f vendor/modules.txt ]
: "${fixture_release_output:?}"
mkdir -p "$(dirname -- "$fixture_release_output")"
cat >"$fixture_release_output" <<PANACEAD
#!/bin/sh
if [ "\${1:-}" = version ] && [ "\${2:-}" = --long ]; then
	printf 'name: panacea-core\\nversion: %s\\ncommit: %s\\nbuild_tags: netgo\\ngo: go version go1.26.5 darwin/%s\\n' '$fixture_version' '$fixture_commit' '${FAKE_GO_ARCH:-arm64}'
	exit 0
fi
exit 2
PANACEAD
chmod +x "$fixture_release_output"
EOF

chmod +x "$fixture_fake_bin/go" "$fixture_fake_bin/uname" \
	"$fixture_fake_bin/file" "$fixture_fake_bin/make"

(
	cd "$fixture_repo"
	git init -q
	git config user.email fixture@example.com
	git config user.name fixture
	git add go.mod go.sum Makefile scripts/release/build-macos.sh
	git commit -q -m 'fixture release'
	git tag v9.9.9
)
fixture_commit=$(git -C "$fixture_repo" rev-parse HEAD)

fixture_output="$fixture_root/build.out"
PATH="$fixture_fake_bin:$PATH" \
	PANACEA_MACOS_BUILD_ROOT="$fixture_root/build-root" \
	PANACEA_MACOS_ARTIFACT_ROOT="$fixture_root/artifacts" \
	PANACEA_MACOS_EXPECTED_TAG=v9.9.9 \
	PANACEA_MACOS_EXPECTED_COMMIT="$fixture_commit" \
	sh "$fixture_repo/scripts/release/build-macos.sh" >"$fixture_output" 2>&1

fixture_artifacts="$fixture_root/artifacts/v9.9.9/darwin-arm64"
test -x "$fixture_artifacts/panacead-darwin-arm64"
test -s "$fixture_artifacts/panacead-darwin-arm64.sha256"
grep -q '^release_tag=v9.9.9$' "$fixture_artifacts/build-info.txt"
grep -q "^source_commit=$fixture_commit$" "$fixture_artifacts/build-info.txt"
grep -q '^platform=darwin/arm64$' "$fixture_artifacts/build-info.txt"
grep -q '^microarchitecture=GOARM64=v8.0$' "$fixture_artifacts/build-info.txt"
grep -q '^cgo_enabled=0$' "$fixture_artifacts/build-info.txt"
grep -q '^ledger_enabled=false$' "$fixture_artifacts/build-info.txt"
grep -q '^build_tags=netgo$' "$fixture_artifacts/build-info.txt"
grep -q '^build_contract=panacea-darwin-purego-v1$' "$fixture_artifacts/build-info.txt"
grep -q '^dependency_mode=vendor$' "$fixture_artifacts/build-info.txt"
grep -q '^packaging=raw-binary$' "$fixture_artifacts/build-info.txt"
grep -q 'panacead-darwin-arm64$' "$fixture_artifacts/panacead-darwin-arm64.sha256"
grep -q 'Signing, packaging, upload, installation, and node execution were not performed' "$fixture_output"

FAKE_GO_ARCH=amd64 \
	FAKE_UNAME_MACHINE=x86_64 \
	PATH="$fixture_fake_bin:$PATH" \
	PANACEA_MACOS_BUILD_ROOT="$fixture_root/intel-build-root" \
	PANACEA_MACOS_ARTIFACT_ROOT="$fixture_root/intel-artifacts" \
	PANACEA_MACOS_EXPECTED_TAG=v9.9.9 \
	PANACEA_MACOS_EXPECTED_COMMIT="$fixture_commit" \
	sh "$fixture_repo/scripts/release/build-macos.sh" >"$fixture_root/intel.out" 2>&1

fixture_intel_artifacts="$fixture_root/intel-artifacts/v9.9.9/darwin-amd64"
test -x "$fixture_intel_artifacts/panacead-darwin-amd64"
test -s "$fixture_intel_artifacts/panacead-darwin-amd64.sha256"
grep -q '^platform=darwin/amd64$' "$fixture_intel_artifacts/build-info.txt"
grep -q '^microarchitecture=GOAMD64=v1$' "$fixture_intel_artifacts/build-info.txt"
grep -q 'panacead-darwin-amd64$' "$fixture_intel_artifacts/panacead-darwin-amd64.sha256"

touch "$fixture_repo/untracked.go"
if PATH="$fixture_fake_bin:$PATH" \
	PANACEA_MACOS_BUILD_ROOT="$fixture_root/dirty-build-root" \
	PANACEA_MACOS_ARTIFACT_ROOT="$fixture_root/dirty-artifacts" \
	PANACEA_MACOS_EXPECTED_TAG=v9.9.9 \
	sh "$fixture_repo/scripts/release/build-macos.sh" >"$fixture_root/dirty.out" 2>&1; then
	printf 'dirty source fixture unexpectedly passed\n' >&2
	exit 1
fi
grep -q 'requires a clean source tree' "$fixture_root/dirty.out"
rm "$fixture_repo/untracked.go"

if FAKE_UNAME_SYSTEM=Linux \
	PATH="$fixture_fake_bin:$PATH" \
	PANACEA_MACOS_BUILD_ROOT="$fixture_root/linux-build-root" \
	PANACEA_MACOS_ARTIFACT_ROOT="$fixture_root/linux-artifacts" \
	PANACEA_MACOS_EXPECTED_TAG=v9.9.9 \
	sh "$fixture_repo/scripts/release/build-macos.sh" >"$fixture_root/linux.out" 2>&1; then
	printf 'Linux host fixture unexpectedly passed\n' >&2
	exit 1
fi
grep -q 'unsupported host OS: Linux' "$fixture_root/linux.out"

sh "$builder_script" --help >"$fixture_root/help.out"
grep -q 'The binary is not signed, packaged, uploaded, or installed' "$fixture_root/help.out"

printf 'macOS release builder fixtures passed\n'
