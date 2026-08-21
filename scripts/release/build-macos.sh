#!/bin/sh

# Build a release-grade native macOS Panacea binary from an exact tag.
#
# The builder requires a clean source tree and the Go version declared by
# go.mod. It stages the tagged commit, vendors dependencies, invokes the
# canonical Darwin build target, verifies the native Mach-O binary, and emits
# the binary, checksum, and provenance under artifacts/. It does not sign,
# package, upload, install, or run a node with the binary.

set -eu
umask 077

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd -P)

macos_build_root=${PANACEA_MACOS_BUILD_ROOT:-"$repo_root/.local/macos-build"}
macos_artifact_root=${PANACEA_MACOS_ARTIFACT_ROOT:-"$repo_root/artifacts/macos-build"}
macos_expected_tag=${PANACEA_MACOS_EXPECTED_TAG:-}
macos_expected_commit=${PANACEA_MACOS_EXPECTED_COMMIT:-}
macos_go_command=${PANACEA_MACOS_GO_BINARY:-go}
macos_goproxy=${PANACEA_MACOS_GOPROXY:-https://proxy.golang.org,direct}
macos_gosumdb=${PANACEA_MACOS_GOSUMDB:-sum.golang.org}
macos_work_dir=
macos_artifact_stage=

usage() {
	cat <<'EOF'
Usage: ./scripts/release/build-macos.sh

Build and verify a native macOS Panacea release binary from a clean, exactly
tagged checkout. The binary is not signed, packaged, uploaded, or installed.

Optional environment variables:
  PANACEA_MACOS_EXPECTED_TAG       Require this exact tag at HEAD
  PANACEA_MACOS_EXPECTED_COMMIT    Require this exact commit at HEAD
  PANACEA_MACOS_GO_BINARY          Go command to use (default: go)
  PANACEA_MACOS_BUILD_ROOT         Isolated build/cache directory
  PANACEA_MACOS_ARTIFACT_ROOT      Output directory root
  PANACEA_MACOS_GOPROXY            Go module proxy setting
  PANACEA_MACOS_GOSUMDB            Go checksum database setting

Supported native build hosts:
  darwin/amd64
  darwin/arm64
EOF
}

die() {
	printf 'macOS release build error: %s\n' "$*" >&2
	exit 1
}

require_command() {
	command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

sha256_file() {
	macos_hash_path=$1
	if command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$macos_hash_path" | awk 'NR == 1 { print $1 }'
	elif command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$macos_hash_path" | awk 'NR == 1 { print $1 }'
	else
		die 'shasum or sha256sum is required'
	fi
}

cleanup() {
	if [ -n "$macos_work_dir" ] && [ -d "$macos_work_dir" ]; then
		rm -rf "$macos_work_dir"
	fi
	if [ -n "$macos_artifact_stage" ] && [ -d "$macos_artifact_stage" ]; then
		rm -rf "$macos_artifact_stage"
	fi
}

trap cleanup EXIT HUP INT TERM

if [ "$#" -gt 1 ]; then
	usage >&2
	exit 2
fi
if [ "$#" -eq 1 ]; then
	case "$1" in
	help | --help | -h)
		usage
		exit 0
		;;
	*)
		printf 'unknown argument: %s\n' "$1" >&2
		usage >&2
		exit 2
		;;
	esac
fi

for macos_command in git make tar awk grep file cp chmod mkdir mktemp mv rm uname env; do
	require_command "$macos_command"
done
require_command "$macos_go_command"
macos_go_binary=$(command -v "$macos_go_command")
macos_go_bin_dir=$(dirname -- "$macos_go_binary")

cd "$repo_root"
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die 'source is not a Git worktree'

source_commit=$(git rev-parse --verify 'HEAD^{commit}')
if [ -n "$macos_expected_commit" ]; then
	expected_commit=$(git rev-parse --verify "$macos_expected_commit^{commit}" 2>/dev/null) ||
		die "expected commit is unavailable: $macos_expected_commit"
	[ "$source_commit" = "$expected_commit" ] ||
		die "HEAD $source_commit does not match expected commit $expected_commit"
fi

if [ -n "$macos_expected_tag" ]; then
	release_tag=$macos_expected_tag
	tag_commit=$(git rev-parse --verify "$release_tag^{commit}" 2>/dev/null) ||
		die "expected tag is unavailable: $release_tag"
	[ "$tag_commit" = "$source_commit" ] ||
		die "tag $release_tag resolves to $tag_commit, not HEAD $source_commit"
else
	release_tag=$(git describe --tags --exact-match HEAD 2>/dev/null) ||
		die 'HEAD must have an exact release tag'
fi
case "$release_tag" in
v[0-9]*) ;;
*) die "release tag must start with v followed by a digit: $release_tag" ;;
esac
case "$release_tag" in
*[!A-Za-z0-9._-]*) die "release tag contains unsafe path characters: $release_tag" ;;
esac

source_status=$(git status --porcelain=v1 --untracked-files=all)
if [ -n "$source_status" ]; then
	printf 'macOS release build requires a clean source tree:\n%s\n' "$source_status" >&2
	exit 1
fi

macos_host_os=$(uname -s)
[ "$macos_host_os" = Darwin ] || die "unsupported host OS: $macos_host_os (macOS required)"
macos_host_machine=$(uname -m)
case "$macos_host_machine" in
x86_64 | amd64)
	macos_go_arch=amd64
	macos_file_arch=x86_64
	macos_microarch=GOAMD64=v1
	;;
arm64 | aarch64)
	macos_go_arch=arm64
	macos_file_arch=arm64
	macos_microarch=GOARM64=v8.0
	;;
*) die "unsupported macOS architecture: $macos_host_machine" ;;
esac

module_go_version=$(awk '$1 == "go" { print $2; exit }' go.mod)
[ -n "$module_go_version" ] || die 'go.mod does not declare a Go version'
macos_go_version_output=$(GOTOOLCHAIN=local GOENV=off "$macos_go_binary" version)
[ "$macos_go_version_output" = "go version go$module_go_version darwin/$macos_go_arch" ] ||
	die "Go executable does not match go.mod and the native host: $macos_go_version_output"
macos_compiler_version=$(GOTOOLCHAIN=local GOENV=off "$macos_go_binary" tool compile -V=full)
case "$macos_compiler_version" in
*"go$module_go_version"*) ;;
*) die "Go compiler does not match go.mod: $macos_compiler_version" ;;
esac

release_version=${release_tag#v}
macos_platform="darwin-$macos_go_arch"
macos_release_root="$macos_artifact_root/$release_tag"
macos_artifact_dir="$macos_release_root/$macos_platform"
[ ! -e "$macos_artifact_dir" ] ||
	die "artifact directory already exists; refusing to overwrite: $macos_artifact_dir"

mkdir -p "$macos_build_root" "$macos_release_root"
chmod 700 "$macos_build_root" "$macos_release_root"
macos_work_dir=$(mktemp -d "$macos_build_root/work.XXXXXX")
macos_source_dir="$macos_work_dir/source"
macos_tmp_dir="$macos_work_dir/tmp"
macos_source_tar="$macos_work_dir/source.tar"
macos_staged_binary="$macos_work_dir/panacead"
macos_gopath="$macos_build_root/gopath"
macos_gocache="$macos_build_root/cache/go-build-$macos_go_arch"
macos_gomodcache="$macos_build_root/cache/go-mod"
mkdir -p "$macos_source_dir" "$macos_tmp_dir" "$macos_gopath" "$macos_gocache" "$macos_gomodcache"

git archive --format=tar --output="$macos_source_tar" "$source_commit"
tar -xf "$macos_source_tar" -C "$macos_source_dir"

cd "$macos_source_dir"
env \
	PATH="$macos_go_bin_dir:$PATH" \
	LC_ALL=C \
	TZ=UTC \
	TMPDIR="$macos_tmp_dir" \
	GOENV=off \
	GOTOOLCHAIN=local \
	GOWORK=off \
	GOFLAGS=-buildvcs=false \
	GOEXPERIMENT= \
	GOFIPS140=off \
	GODEBUG= \
	GOPATH="$macos_gopath" \
	GOCACHE="$macos_gocache" \
	GOMODCACHE="$macos_gomodcache" \
	GOPROXY="$macos_goproxy" \
	GOSUMDB="$macos_gosumdb" \
	GOPRIVATE= \
	GONOPROXY= \
	GONOSUMDB= \
	"$macos_go_binary" mod verify
env \
	PATH="$macos_go_bin_dir:$PATH" \
	LC_ALL=C \
	TZ=UTC \
	TMPDIR="$macos_tmp_dir" \
	GOENV=off \
	GOTOOLCHAIN=local \
	GOWORK=off \
	GOFLAGS=-buildvcs=false \
	GOEXPERIMENT= \
	GOFIPS140=off \
	GODEBUG= \
	GOPATH="$macos_gopath" \
	GOCACHE="$macos_gocache" \
	GOMODCACHE="$macos_gomodcache" \
	GOPROXY="$macos_goproxy" \
	GOSUMDB="$macos_gosumdb" \
	GOPRIVATE= \
	GONOPROXY= \
	GONOSUMDB= \
	"$macos_go_binary" mod vendor

printf 'Building Panacea %s (%s) with Go %s\n' \
	"$release_version" "$macos_platform" "$module_go_version"
env \
	PATH="$macos_go_bin_dir:$PATH" \
	LC_ALL=C \
	TZ=UTC \
	TMPDIR="$macos_tmp_dir" \
	GOENV=off \
	GOTOOLCHAIN=local \
	GOWORK=off \
	GOFLAGS=-buildvcs=false \
	GOEXPERIMENT= \
	GOFIPS140=off \
	GODEBUG= \
	GOOS=darwin \
	GOARCH="$macos_go_arch" \
	CGO_ENABLED=0 \
	LEDGER_ENABLED=false \
	GOPATH="$macos_gopath" \
	GOCACHE="$macos_gocache" \
	GOMODCACHE="$macos_gomodcache" \
	GOPROXY="$macos_goproxy" \
	GOSUMDB="$macos_gosumdb" \
	GOPRIVATE= \
	GONOPROXY= \
	GONOSUMDB= \
	COSMOS_BUILD_OPTIONS= \
	BUILD_TAGS= \
	make \
	VERSION="$release_version" \
	COMMIT="$source_commit" \
	LEDGER_ENABLED=false \
	COSMOS_BUILD_OPTIONS= \
	BUILD_TAGS= \
	build_tags=netgo \
	build_tags_comma_sep=netgo \
	RELEASE_GOARCH="$macos_go_arch" \
	RELEASE_OUTPUT="$macos_staged_binary" \
	release-build-darwin
cd "$repo_root"

[ -x "$macos_staged_binary" ] || die "build did not produce $macos_staged_binary"
macos_version_output=$("$macos_staged_binary" version --long)
printf '%s\n' "$macos_version_output" | grep -Fq "version: $release_version" ||
	die "binary version does not contain expected version $release_version"
printf '%s\n' "$macos_version_output" | grep -Fq "commit: $source_commit" ||
	die "binary version does not contain expected commit $source_commit"
printf '%s\n' "$macos_version_output" | grep -Fq 'build_tags: netgo' ||
	die 'binary version does not contain expected netgo build tag'

macos_file_output=$(file "$macos_staged_binary")
case "$macos_file_output" in
*'Mach-O 64-bit executable'*"$macos_file_arch"*) ;;
*) die "binary is not a native $macos_file_arch Mach-O executable: $macos_file_output" ;;
esac

final_commit=$(git rev-parse --verify 'HEAD^{commit}')
final_status=$(git status --porcelain=v1 --untracked-files=all)
[ "$final_commit" = "$source_commit" ] || die 'source commit changed during the build'
[ -z "$final_status" ] || {
	printf 'source tree changed during the build:\n%s\n' "$final_status" >&2
	exit 1
}

macos_artifact_stage=$(mktemp -d "$macos_release_root/.$macos_platform.XXXXXX")
macos_binary_name="panacead-$macos_platform"
macos_binary_path="$macos_artifact_stage/$macos_binary_name"
cp "$macos_staged_binary" "$macos_binary_path"
chmod 0755 "$macos_binary_path"
macos_binary_sha256=$(sha256_file "$macos_binary_path")
printf '%s  %s\n' "$macos_binary_sha256" "$macos_binary_name" \
	>"$macos_artifact_stage/$macos_binary_name.sha256"
printf '%s\n' "$macos_version_output" >"$macos_artifact_stage/panacead-version-long.txt"
printf '%s\n' "$macos_file_output" >"$macos_artifact_stage/file.txt"
printf '%s\n' "$macos_go_version_output" >"$macos_artifact_stage/go-version.txt"
printf '%s\n' "$macos_compiler_version" >"$macos_artifact_stage/compiler-version.txt"
{
	printf 'schema_version=1\n'
	printf 'release_tag=%s\n' "$release_tag"
	printf 'release_version=%s\n' "$release_version"
	printf 'source_commit=%s\n' "$source_commit"
	printf 'platform=darwin/%s\n' "$macos_go_arch"
	printf 'microarchitecture=%s\n' "$macos_microarch"
	printf 'go_version=%s\n' "$module_go_version"
	printf 'go_version_output=%s\n' "$macos_go_version_output"
	printf 'compiler_version=%s\n' "$macos_compiler_version"
	printf 'gotoolchain=local\n'
	printf 'gowork=off\n'
	printf 'buildvcs=false\n'
	printf 'gofips140=off\n'
	printf 'cgo_enabled=0\n'
	printf 'ledger_enabled=false\n'
	printf 'build_tags=netgo\n'
	printf 'cosmos_build_options=\n'
	printf 'extra_build_tags=\n'
	printf 'build_contract=panacea-darwin-purego-v1\n'
	printf 'dependency_mode=vendor\n'
	printf 'packaging=raw-binary\n'
	printf 'binary=%s\n' "$macos_binary_name"
	printf 'binary_sha256=%s\n' "$macos_binary_sha256"
} >"$macos_artifact_stage/build-info.txt"

mv "$macos_artifact_stage" "$macos_artifact_dir"
macos_artifact_stage=

printf '\nmacOS release binary created successfully.\n'
printf '  Binary:    %s/%s\n' "$macos_artifact_dir" "$macos_binary_name"
printf '  SHA256:    %s\n' "$macos_binary_sha256"
printf '  Build info: %s/build-info.txt\n' "$macos_artifact_dir"
printf '  Signing, packaging, upload, installation, and node execution were not performed.\n'
