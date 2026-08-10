#!/bin/sh

set -eu

test_script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P)
builder_script="$test_script_dir/build-validator.sh"
fixture_root=$(mktemp -d "${TMPDIR:-/tmp}/panacea-validator-builder.XXXXXX")
trap 'rm -rf "$fixture_root"' EXIT HUP INT TERM

fixture_repo="$fixture_root/repo"
fixture_fake_bin="$fixture_root/fake-bin"
fixture_go_tree="$fixture_root/go-tree"
fixture_go_tarball="$fixture_root/fake-go1.26.5.linux-amd64.tar.gz"
fixture_go_tarball_arm64="$fixture_root/fake-go1.26.5.linux-arm64.tar.gz"
mkdir -p "$fixture_repo/scripts/release" "$fixture_fake_bin" "$fixture_go_tree/go/bin"
cp "$builder_script" "$fixture_repo/scripts/release/build-validator.sh"
chmod +x "$fixture_repo/scripts/release/build-validator.sh"

cat >"$fixture_repo/go.mod" <<'EOF'
module example.com/panacea-validator-builder-test

go 1.26.5
EOF
cat >"$fixture_repo/Makefile" <<'EOF'
release-build:
	@false
EOF
cat >"$fixture_go_tree/go/bin/go" <<'EOF'
#!/bin/sh
case "$*" in
version) printf 'go version go1.26.5 linux/%s\n' "${FAKE_GO_ARCH:-amd64}" ;;
'tool compile -V=full') printf 'compile version go1.26.5\n' ;;
'mod vendor') mkdir -p vendor; : >vendor/modules.txt ;;
*) printf 'unexpected fake go arguments: %s\n' "$*" >&2; exit 2 ;;
esac
EOF
chmod +x "$fixture_go_tree/go/bin/go"
tar -C "$fixture_go_tree" -czf "$fixture_go_tarball" go
cp "$fixture_go_tarball" "$fixture_go_tarball_arm64"

cat >"$fixture_fake_bin/uname" <<'EOF'
#!/bin/sh
case "${1:-}" in
-s) printf 'Linux\n' ;;
-m) printf '%s\n' "${FAKE_UNAME_MACHINE:-x86_64}" ;;
*) printf 'Linux validator-fixture 6.8.0 x86_64 GNU/Linux\n' ;;
esac
EOF
cat >"$fixture_fake_bin/sha256sum" <<'EOF'
#!/bin/sh
case "${1:-}" in
*fake-go1.26.5.linux-amd64.tar.gz)
	printf '5c2c3b16caefa1d968a94c1daca04a7ca301a496d9b086e17ad77bb81393f053  %s\n' "$1"
	;;
*fake-go1.26.5.linux-arm64.tar.gz)
	printf 'fe4789e92b1f33358680864bbe8704289e7bb5fc207d80623c308935bd696d49  %s\n' "$1"
	;;
*)
	printf 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  %s\n' "$1"
	;;
esac
EOF
cat >"$fixture_fake_bin/make" <<'EOF'
#!/bin/sh
fixture_version=
fixture_commit=
fixture_ledger_enabled=
fixture_build_tags=
fixture_build_tags_comma_sep=
fixture_cosmos_build_options=not-cleared
fixture_extra_build_tags=not-cleared
fixture_release_goarch=
fixture_release_output=
fixture_go_build_mod=
fixture_target=
for fixture_arg in "$@"; do
	case "$fixture_arg" in
	VERSION=*) fixture_version=${fixture_arg#VERSION=} ;;
	COMMIT=*) fixture_commit=${fixture_arg#COMMIT=} ;;
	LEDGER_ENABLED=*) fixture_ledger_enabled=${fixture_arg#LEDGER_ENABLED=} ;;
	build_tags=*) fixture_build_tags=${fixture_arg#build_tags=} ;;
	build_tags_comma_sep=*) fixture_build_tags_comma_sep=${fixture_arg#build_tags_comma_sep=} ;;
	COSMOS_BUILD_OPTIONS=*) fixture_cosmos_build_options=${fixture_arg#COSMOS_BUILD_OPTIONS=} ;;
	BUILD_TAGS=*) fixture_extra_build_tags=${fixture_arg#BUILD_TAGS=} ;;
	RELEASE_GOARCH=*) fixture_release_goarch=${fixture_arg#RELEASE_GOARCH=} ;;
	RELEASE_OUTPUT=*) fixture_release_output=${fixture_arg#RELEASE_OUTPUT=} ;;
	GO_BUILD_MOD=*) fixture_go_build_mod=${fixture_arg#GO_BUILD_MOD=} ;;
	*=*) ;;
	*) fixture_target=$fixture_arg ;;
	esac
done
[ "$fixture_ledger_enabled" = false ]
[ "$fixture_build_tags" = netgo ]
[ "$fixture_build_tags_comma_sep" = netgo ]
[ -z "$fixture_cosmos_build_options" ]
[ -z "$fixture_extra_build_tags" ]
[ -z "${COSMOS_BUILD_OPTIONS:-}" ]
[ -z "${BUILD_TAGS:-}" ]
[ "$fixture_target" = release-build ]
[ "$fixture_release_goarch" = "$GOARCH" ]
[ -z "$fixture_go_build_mod" ]
[ -f vendor/modules.txt ]
: "${fixture_release_output:?}"
mkdir -p "$(dirname "$fixture_release_output")"
cat >"$fixture_release_output" <<PANACEAD
#!/bin/sh
if [ "\${1:-}" = version ] && [ "\${2:-}" = --long ]; then
	printf 'name: panacea-core\\nversion: %s\\ncommit: %s\\nbuild_tags: netgo\\ngo: go version go1.26.5 linux/%s\\n' '$fixture_version' '$fixture_commit' '$GOARCH'
	exit 0
fi
exit 2
PANACEAD
chmod +x "$fixture_release_output"
EOF
cat >"$fixture_fake_bin/file" <<'EOF'
#!/bin/sh
printf '%s: ELF 64-bit LSB executable, x86-64, statically linked, stripped\n' "$1"
EOF
chmod +x "$fixture_fake_bin/uname" "$fixture_fake_bin/sha256sum" \
	"$fixture_fake_bin/make" "$fixture_fake_bin/file"

(
	cd "$fixture_repo"
	git init -q
	git config user.name 'Validator Builder Test'
	git config user.email validator-builder@example.invalid
	git config commit.gpgsign false
	git add go.mod Makefile scripts/release/build-validator.sh
	git commit -q -m 'fixture release'
	git tag v9.9.9
)
fixture_commit=$(git -C "$fixture_repo" rev-parse HEAD)

fixture_output="$fixture_root/build.out"
PATH="$fixture_fake_bin:$PATH" \
	COSMOS_BUILD_OPTIONS=secp \
	BUILD_TAGS=host-specific-tag \
	PANACEA_VALIDATOR_BUILD_ROOT="$fixture_root/build-root" \
	PANACEA_VALIDATOR_ARTIFACT_ROOT="$fixture_root/artifacts" \
	PANACEA_VALIDATOR_GO_TARBALL="$fixture_go_tarball" \
	PANACEA_VALIDATOR_EXPECTED_TAG=v9.9.9 \
	PANACEA_VALIDATOR_EXPECTED_COMMIT="$fixture_commit" \
	sh "$fixture_repo/scripts/release/build-validator.sh" >"$fixture_output" 2>&1

fixture_artifacts="$fixture_root/artifacts/v9.9.9/linux-amd64"
test -x "$fixture_artifacts/panacead-linux-amd64"
grep -q '^release_tag=v9.9.9$' "$fixture_artifacts/build-info.txt"
grep -q "^source_commit=$fixture_commit$" "$fixture_artifacts/build-info.txt"
grep -q '^cgo_enabled=0$' "$fixture_artifacts/build-info.txt"
grep -q '^ledger_enabled=false$' "$fixture_artifacts/build-info.txt"
grep -q '^buildvcs=false$' "$fixture_artifacts/build-info.txt"
grep -q '^gofips140=off$' "$fixture_artifacts/build-info.txt"
grep -q '^build_tags=netgo$' "$fixture_artifacts/build-info.txt"
grep -q '^cosmos_build_options=$' "$fixture_artifacts/build-info.txt"
grep -q '^extra_build_tags=$' "$fixture_artifacts/build-info.txt"
grep -q '^build_contract=panacea-linux-static-v1$' "$fixture_artifacts/build-info.txt"
grep -q '^dependency_mode=vendor$' "$fixture_artifacts/build-info.txt"
grep -q '^microarchitecture=GOAMD64=v1$' "$fixture_artifacts/build-info.txt"
grep -q '^aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  panacead-linux-amd64$' \
	"$fixture_artifacts/panacead-linux-amd64.sha256"
grep -q 'Deployment and node restart were not performed' "$fixture_output"

FAKE_GO_ARCH=arm64 \
	FAKE_UNAME_MACHINE=aarch64 \
	PATH="$fixture_fake_bin:$PATH" \
	PANACEA_VALIDATOR_BUILD_ROOT="$fixture_root/build-root" \
	PANACEA_VALIDATOR_ARTIFACT_ROOT="$fixture_root/artifacts" \
	PANACEA_VALIDATOR_GO_TARBALL="$fixture_go_tarball_arm64" \
	PANACEA_VALIDATOR_EXPECTED_TAG=v9.9.9 \
	PANACEA_VALIDATOR_EXPECTED_COMMIT="$fixture_commit" \
	sh "$fixture_repo/scripts/release/build-validator.sh" >"$fixture_root/build-arm64.out" 2>&1

fixture_arm64_artifacts="$fixture_root/artifacts/v9.9.9/linux-arm64"
test -x "$fixture_arm64_artifacts/panacead-linux-arm64"
grep -q '^platform=linux/arm64$' "$fixture_arm64_artifacts/build-info.txt"
grep -q '^microarchitecture=GOARM64=v8.0$' "$fixture_arm64_artifacts/build-info.txt"

fixture_invalid_tarball="$fixture_root/invalid-go.tar.gz"
cp "$fixture_go_tarball" "$fixture_invalid_tarball"
if PATH="$fixture_fake_bin:$PATH" \
	PANACEA_VALIDATOR_BUILD_ROOT="$fixture_root/checksum-build-root" \
	PANACEA_VALIDATOR_ARTIFACT_ROOT="$fixture_root/checksum-artifacts" \
	PANACEA_VALIDATOR_GO_TARBALL="$fixture_invalid_tarball" \
	PANACEA_VALIDATOR_EXPECTED_TAG=v9.9.9 \
	sh "$fixture_repo/scripts/release/build-validator.sh" >"$fixture_root/checksum.out" 2>&1; then
	printf 'invalid Go checksum fixture unexpectedly passed\n' >&2
	exit 1
fi
grep -q 'Go tarball checksum mismatch' "$fixture_root/checksum.out"

touch "$fixture_repo/untracked.go"
if PATH="$fixture_fake_bin:$PATH" \
	PANACEA_VALIDATOR_BUILD_ROOT="$fixture_root/dirty-build-root" \
	PANACEA_VALIDATOR_ARTIFACT_ROOT="$fixture_root/dirty-artifacts" \
	PANACEA_VALIDATOR_GO_TARBALL="$fixture_go_tarball" \
	PANACEA_VALIDATOR_EXPECTED_TAG=v9.9.9 \
	sh "$fixture_repo/scripts/release/build-validator.sh" >"$fixture_root/dirty.out" 2>&1; then
	printf 'dirty source fixture unexpectedly passed\n' >&2
	exit 1
fi
grep -q 'requires a clean source tree' "$fixture_root/dirty.out"
rm "$fixture_repo/untracked.go"

if PATH="$fixture_fake_bin:$PATH" \
	PANACEA_VALIDATOR_BUILD_ROOT="$fixture_root/mismatch-build-root" \
	PANACEA_VALIDATOR_ARTIFACT_ROOT="$fixture_root/mismatch-artifacts" \
	PANACEA_VALIDATOR_GO_TARBALL="$fixture_go_tarball" \
	PANACEA_VALIDATOR_EXPECTED_TAG=v9.9.8 \
	sh "$fixture_repo/scripts/release/build-validator.sh" >"$fixture_root/tag-mismatch.out" 2>&1; then
	printf 'missing expected tag fixture unexpectedly passed\n' >&2
	exit 1
fi
grep -q 'expected tag is unavailable' "$fixture_root/tag-mismatch.out"

sh "$builder_script" --help >"$fixture_root/help.out"
grep -q 'The host Go installation and the running node are never modified' "$fixture_root/help.out"

printf 'validator builder fixtures passed\n'
