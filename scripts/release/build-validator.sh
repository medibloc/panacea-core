#!/bin/sh

# Build a validator-grade Panacea binary without changing the host Go install.
#
# The builder accepts only a clean, exactly tagged source tree. It downloads
# the pinned official Go toolchain into a repository-local ignored directory,
# verifies the upstream checksum, stages the tagged commit in an isolated source
# directory, and emits a static native Linux binary plus provenance under
# artifacts/. It never installs or restarts panacead.

set -eu
umask 077

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd -P)

validator_go_version=1.26.5
validator_build_root=${PANACEA_VALIDATOR_BUILD_ROOT:-"$repo_root/.local/validator-build"}
validator_artifact_root=${PANACEA_VALIDATOR_ARTIFACT_ROOT:-"$repo_root/artifacts/validator-build"}
validator_go_download_base=${PANACEA_VALIDATOR_GO_DOWNLOAD_BASE:-https://go.dev/dl}
validator_go_tarball=${PANACEA_VALIDATOR_GO_TARBALL:-}
validator_expected_tag=${PANACEA_VALIDATOR_EXPECTED_TAG:-}
validator_expected_commit=${PANACEA_VALIDATOR_EXPECTED_COMMIT:-}
validator_goproxy=${PANACEA_VALIDATOR_GOPROXY:-https://proxy.golang.org,direct}
validator_gosumdb=${PANACEA_VALIDATOR_GOSUMDB:-sum.golang.org}

usage() {
	cat <<'EOF'
Usage: ./scripts/release/build-validator.sh

Build a static Panacea validator binary from a clean, exactly tagged checkout.
The host Go installation and the running node are never modified.

Optional environment variables:
  PANACEA_VALIDATOR_EXPECTED_TAG       Require this exact tag at HEAD
  PANACEA_VALIDATOR_EXPECTED_COMMIT    Require this exact commit at HEAD
  PANACEA_VALIDATOR_GO_TARBALL         Use a pre-downloaded official tarball
  PANACEA_VALIDATOR_BUILD_ROOT         Isolated toolchain/cache directory
  PANACEA_VALIDATOR_ARTIFACT_ROOT      Output directory root
  PANACEA_VALIDATOR_GO_DOWNLOAD_BASE   Go download base URL
  PANACEA_VALIDATOR_GOPROXY            Go module proxy setting
  PANACEA_VALIDATOR_GOSUMDB            Go checksum database setting

Supported build hosts:
  linux/amd64
  linux/arm64
EOF
}

die() {
	printf 'validator build error: %s\n' "$*" >&2
	exit 1
}

require_command() {
	command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

sha256_file() {
	validator_hash_path=$1
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$validator_hash_path" | awk 'NR == 1 { print $1 }'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$validator_hash_path" | awk 'NR == 1 { print $1 }'
	else
		die 'sha256sum or shasum is required'
	fi
}

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

for validator_command in git make tar awk sed grep file cp chmod mkdir mktemp mv rmdir uname env; do
	require_command "$validator_command"
done

cd "$repo_root"
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die 'source is not a Git worktree'

source_commit=$(git rev-parse --verify 'HEAD^{commit}')
if [ -n "$validator_expected_commit" ]; then
	expected_commit=$(git rev-parse --verify "$validator_expected_commit^{commit}" 2>/dev/null) ||
		die "expected commit is unavailable: $validator_expected_commit"
	[ "$source_commit" = "$expected_commit" ] ||
		die "HEAD $source_commit does not match expected commit $expected_commit"
fi

if [ -n "$validator_expected_tag" ]; then
	release_tag=$validator_expected_tag
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
	printf 'validator build requires a clean source tree:\n%s\n' "$source_status" >&2
	exit 1
fi

module_go_version=$(awk '$1 == "go" { print $2; exit }' go.mod)
[ "$module_go_version" = "$validator_go_version" ] ||
	die "go.mod requires Go $module_go_version, builder is pinned to $validator_go_version"

validator_host_os=$(uname -s)
[ "$validator_host_os" = Linux ] || die "unsupported host OS: $validator_host_os (Linux required)"
validator_host_arch=$(uname -m)
case "$validator_host_arch" in
	x86_64 | amd64)
		validator_go_arch=amd64
		validator_go_tarball_name="go$validator_go_version.linux-amd64.tar.gz"
		validator_go_tarball_sha256=5c2c3b16caefa1d968a94c1daca04a7ca301a496d9b086e17ad77bb81393f053
		validator_microarch=GOAMD64=v1
		;;
	aarch64 | arm64)
		validator_go_arch=arm64
		validator_go_tarball_name="go$validator_go_version.linux-arm64.tar.gz"
		validator_go_tarball_sha256=fe4789e92b1f33358680864bbe8704289e7bb5fc207d80623c308935bd696d49
		validator_microarch=GOARM64=v8.0
		;;
	*) die "unsupported Linux architecture: $validator_host_arch" ;;
esac

mkdir -p "$validator_build_root/downloads" "$validator_build_root/toolchains"
chmod 700 "$validator_build_root"

if [ -n "$validator_go_tarball" ]; then
	[ -f "$validator_go_tarball" ] || die "Go tarball not found: $validator_go_tarball"
else
	validator_go_tarball="$validator_build_root/downloads/$validator_go_tarball_name"
	if [ ! -f "$validator_go_tarball" ]; then
		require_command curl
		validator_download_tmp="$validator_go_tarball.part.$$"
		trap 'rm -f "$validator_download_tmp"' EXIT HUP INT TERM
		printf 'Downloading %s\n' "$validator_go_tarball_name"
		curl --fail --location --proto '=https' --tlsv1.2 \
			--output "$validator_download_tmp" \
			"$validator_go_download_base/$validator_go_tarball_name"
		mv "$validator_download_tmp" "$validator_go_tarball"
		trap - EXIT HUP INT TERM
	fi
fi

validator_download_sha256=$(sha256_file "$validator_go_tarball")
[ "$validator_download_sha256" = "$validator_go_tarball_sha256" ] ||
	die "Go tarball checksum mismatch: expected $validator_go_tarball_sha256, got $validator_download_sha256"

validator_toolchain_dir="$validator_build_root/toolchains/go$validator_go_version-linux-$validator_go_arch"
if [ -e "$validator_toolchain_dir" ] && [ ! -x "$validator_toolchain_dir/bin/go" ]; then
	die "incomplete toolchain directory already exists: $validator_toolchain_dir"
fi
if [ ! -e "$validator_toolchain_dir" ]; then
	validator_extract_dir=$(mktemp -d "$validator_build_root/extract.XXXXXX")
	tar -C "$validator_extract_dir" -xzf "$validator_go_tarball"
	[ -x "$validator_extract_dir/go/bin/go" ] ||
		die 'official Go archive did not contain go/bin/go'
	mv "$validator_extract_dir/go" "$validator_toolchain_dir"
	rmdir "$validator_extract_dir"
fi

validator_go_binary="$validator_toolchain_dir/bin/go"
PATH="$validator_toolchain_dir/bin:$PATH"
export PATH
unset GOROOT GOBIN GOFLAGS GOEXPERIMENT

validator_go_version_output=$(GOTOOLCHAIN=local GOENV=off "$validator_go_binary" version)
[ "$validator_go_version_output" = "go version go$validator_go_version linux/$validator_go_arch" ] ||
	die "isolated Go executable is inconsistent: $validator_go_version_output"
validator_compiler_version=$(GOTOOLCHAIN=local GOENV=off "$validator_go_binary" tool compile -V=full)
case "$validator_compiler_version" in
	*"go$validator_go_version"*) ;;
	*) die "isolated compiler is inconsistent: $validator_compiler_version" ;;
esac

release_version=${release_tag#v}
validator_platform="linux-$validator_go_arch"
validator_release_root="$validator_artifact_root/$release_tag"
validator_artifact_dir="$validator_release_root/$validator_platform"
mkdir -p "$validator_release_root"
if ! mkdir "$validator_artifact_dir"; then
	die "artifact directory already exists; refusing to overwrite: $validator_artifact_dir"
fi
chmod 700 "$validator_artifact_dir"

validator_gopath="$validator_build_root/gopath"
validator_gocache="$validator_build_root/cache/go-build-$validator_go_arch"
validator_gomodcache="$validator_build_root/cache/go-mod"
validator_stage_bin=$(mktemp -d "$validator_build_root/bin.XXXXXX")
validator_tmp_dir=$(mktemp -d "$validator_build_root/tmp.XXXXXX")
validator_staged_binary="$validator_stage_bin/panacead"
validator_source_dir=$(mktemp -d "$validator_build_root/source.XXXXXX")
validator_source_tar="$validator_tmp_dir/source.tar"
mkdir -p "$validator_gopath" "$validator_gocache" "$validator_gomodcache"

git archive --format=tar --output="$validator_source_tar" "$source_commit"
tar -xf "$validator_source_tar" -C "$validator_source_dir"

cd "$validator_source_dir"
env \
	LC_ALL=C \
	TZ=UTC \
	TMPDIR="$validator_tmp_dir" \
	GOENV=off \
	GOTOOLCHAIN=local \
	GOWORK=off \
	GOFLAGS=-buildvcs=false \
	GOEXPERIMENT= \
	GOFIPS140=off \
	GODEBUG= \
	GOPATH="$validator_gopath" \
	GOCACHE="$validator_gocache" \
	GOMODCACHE="$validator_gomodcache" \
	GOPROXY="$validator_goproxy" \
	GOSUMDB="$validator_gosumdb" \
	GOPRIVATE= \
	GONOPROXY= \
	GONOSUMDB= \
	"$validator_go_binary" mod vendor

printf 'Building Panacea %s (%s) with Go %s\n' \
	"$release_version" "$validator_platform" "$validator_go_version"
env \
	LC_ALL=C \
	TZ=UTC \
	TMPDIR="$validator_tmp_dir" \
	GOENV=off \
	GOTOOLCHAIN=local \
	GOWORK=off \
	GOFLAGS=-buildvcs=false \
	GOEXPERIMENT= \
	GOFIPS140=off \
	GODEBUG= \
	GOOS=linux \
	GOARCH="$validator_go_arch" \
	GOAMD64=v1 \
	GOARM64=v8.0 \
	CGO_ENABLED=0 \
	LEDGER_ENABLED=false \
	GOPATH="$validator_gopath" \
	GOBIN="$validator_stage_bin" \
	GOCACHE="$validator_gocache" \
	GOMODCACHE="$validator_gomodcache" \
	GOPROXY="$validator_goproxy" \
	GOSUMDB="$validator_gosumdb" \
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
	RELEASE_GOARCH="$validator_go_arch" \
	RELEASE_OUTPUT="$validator_staged_binary" \
	release-build
cd "$repo_root"

[ -x "$validator_staged_binary" ] || die "build did not produce $validator_staged_binary"

validator_version_output=$("$validator_staged_binary" version --long)
printf '%s\n' "$validator_version_output" >"$validator_artifact_dir/panacead-version-long.txt"
printf '%s\n' "$validator_version_output" | grep -Fq "version: $release_version" ||
	die "binary version does not contain expected version $release_version"
printf '%s\n' "$validator_version_output" | grep -Fq "commit: $source_commit" ||
	die "binary version does not contain expected commit $source_commit"

validator_file_output=$(file "$validator_staged_binary")
printf '%s\n' "$validator_file_output" >"$validator_artifact_dir/file.txt"
case "$validator_file_output" in
	*'statically linked'*) ;;
	*) die "binary is not reported as statically linked: $validator_file_output" ;;
esac

validator_binary_name="panacead-$validator_platform"
validator_binary_path="$validator_artifact_dir/$validator_binary_name"
cp "$validator_staged_binary" "$validator_binary_path"
chmod 0755 "$validator_binary_path"
validator_binary_sha256=$(sha256_file "$validator_binary_path")
printf '%s  %s\n' "$validator_binary_sha256" "$validator_binary_name" \
	>"$validator_artifact_dir/$validator_binary_name.sha256"

printf '%s\n' "$validator_go_version_output" >"$validator_artifact_dir/go-version.txt"
printf '%s\n' "$validator_compiler_version" >"$validator_artifact_dir/compiler-version.txt"
{
	printf 'schema_version=1\n'
	printf 'release_tag=%s\n' "$release_tag"
	printf 'release_version=%s\n' "$release_version"
	printf 'source_commit=%s\n' "$source_commit"
	printf 'platform=linux/%s\n' "$validator_go_arch"
	printf 'microarchitecture=%s\n' "$validator_microarch"
	printf 'go_version=%s\n' "$validator_go_version"
	printf 'go_tarball=%s\n' "$validator_go_tarball_name"
	printf 'go_tarball_sha256=%s\n' "$validator_go_tarball_sha256"
	printf 'gotoolchain=local\n'
	printf 'gowork=off\n'
	printf 'buildvcs=false\n'
	printf 'gofips140=off\n'
	printf 'cgo_enabled=0\n'
	printf 'ledger_enabled=false\n'
	printf 'build_tags=netgo\n'
	printf 'cosmos_build_options=\n'
	printf 'extra_build_tags=\n'
	printf 'build_contract=panacea-linux-static-v1\n'
	printf 'dependency_mode=vendor\n'
	printf 'binary=%s\n' "$validator_binary_name"
	printf 'binary_sha256=%s\n' "$validator_binary_sha256"
} >"$validator_artifact_dir/build-info.txt"

final_commit=$(git rev-parse --verify 'HEAD^{commit}')
final_status=$(git status --porcelain=v1 --untracked-files=all)
[ "$final_commit" = "$source_commit" ] || die 'source commit changed during the build'
[ -z "$final_status" ] || {
	printf 'source tree changed during the build:\n%s\n' "$final_status" >&2
	exit 1
}

printf '\nValidator binary created successfully.\n'
printf '  Binary:    %s\n' "$validator_binary_path"
printf '  SHA256:    %s\n' "$validator_binary_sha256"
printf '  Build info: %s\n' "$validator_artifact_dir/build-info.txt"
printf '  Deployment and node restart were not performed.\n'
