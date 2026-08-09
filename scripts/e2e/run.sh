#!/bin/sh

# Standalone entrypoint for the opt-in real-node E2E harness.
#
# Keep this runner independent from repository build and automation entrypoints.
# Every Go command is pinned to the selected local toolchain and runs with the
# nested e2e module isolated from any ambient go.work file.

set -eu

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd -P)
runner="$script_dir/run.sh"

usage() {
	cat <<'EOF'
Usage: ./scripts/e2e/run.sh COMMAND

Commands:
  check                    Validate E2E paths and the exact Go toolchain
  check-clean              Require a clean, unchanged HEAD worktree
  build-current            Build the current Panacea E2E image
  build-v2.2.1             Build the pinned Panacea v2.2.1 E2E image
  build-images             Build both E2E images
  build-test-binary        Compile the E2E test binary
  build                    Build both images and the E2E test binary
  unit                     Run E2E harness unit tests
  smoke                    Run current-node smoke and failure-probe tests
  v2.2.1                   Run the pinned v2.2.1 compatibility test
  compatibility            Run current and v2.2.1 compatibility tests
  negative                 Run negative NFT boundary tests
  restart                  Run restart, snapshot, and sync recovery tests
  consensus                Run four-validator quorum recovery tests
  upgrade                  Run the v2.2.1-to-current upgrade test
  cosmovisor               Run an explicit old-to-current Cosmovisor rehearsal
  upgrade-deep             Run normal and legacy-PNFT upgrade matrices
  upgrade-chaos            Run upgrade-boundary chaos tests
  state-sync               Run state-sync success and rejection tests
  config-compat            Run legacy node-home compatibility tests
  ibc-upgrade              Run IBC continuity across the upgrade
  network-faults           Run local Docker network fault tests
  release-builds           Build and verify multi-architecture release images
  release-hardening        Run the complete artifact-first release gate
  release-hardening-inner  Run the internal release-gate sequence
  load                     Run the short load/resource baseline
  all                      Run every functional live suite (no release build)
  help                     Show this help
EOF
}

if [ "$#" -eq 0 ]; then
	usage >&2
	exit 2
fi
if [ "$#" -ne 1 ]; then
	printf 'expected exactly one E2E command\n' >&2
	usage >&2
	exit 2
fi

command_name=$1
case "$command_name" in
	help | --help | -h)
		usage
		exit 0
		;;
	check | check-clean | build-current | build-v2.2.1 | build-images | \
		build-test-binary | build | unit | smoke | v2.2.1 | compatibility | \
		negative | restart | consensus | upgrade | cosmovisor | upgrade-deep | \
		upgrade-chaos | state-sync | config-compat | ibc-upgrade | \
		network-faults | release-builds | release-hardening | \
		release-hardening-inner | load | all) ;;
	*)
		printf 'unknown E2E command: %s\n' "$command_name" >&2
		usage >&2
		exit 2
		;;
esac

cd "$repo_root"

E2E_GO_VERSION=${E2E_GO_VERSION:-1.26.5}
if [ "$E2E_GO_VERSION" != 1.26.5 ]; then
	printf 'E2E_GO_VERSION must be 1.26.5, got: %s\n' "$E2E_GO_VERSION" >&2
	exit 2
fi
E2E_GOTOOLCHAIN=${E2E_GOTOOLCHAIN:-local}
if [ "$E2E_GOTOOLCHAIN" != local ]; then
	printf 'E2E_GOTOOLCHAIN must be local, got: %s\n' "$E2E_GOTOOLCHAIN" >&2
	exit 2
fi
E2E_GO_BINARY=${E2E_GO_BINARY:-go}
E2E_ROOT=${E2E_ROOT:-"$repo_root/.local/e2e"}
E2E_GOCACHE=${E2E_GOCACHE:-"$E2E_ROOT/go-build"}
E2E_GOMODCACHE=${E2E_GOMODCACHE:-"$E2E_ROOT/go-mod"}

E2E_TEST_TIMEOUT=${E2E_TEST_TIMEOUT:-12m}
E2E_NEGATIVE_TIMEOUT=${E2E_NEGATIVE_TIMEOUT:-40m}
E2E_RESTART_TIMEOUT=${E2E_RESTART_TIMEOUT:-35m}
E2E_CONSENSUS_TIMEOUT=${E2E_CONSENSUS_TIMEOUT:-18m}
E2E_UPGRADE_TIMEOUT=${E2E_UPGRADE_TIMEOUT:-35m}
E2E_UPGRADE_DEEP_TIMEOUT=${E2E_UPGRADE_DEEP_TIMEOUT:-50m}
E2E_UPGRADE_CHAOS_TIMEOUT=${E2E_UPGRADE_CHAOS_TIMEOUT:-40m}
E2E_STATE_SYNC_TIMEOUT=${E2E_STATE_SYNC_TIMEOUT:-20m}
E2E_CONFIG_COMPAT_TIMEOUT=${E2E_CONFIG_COMPAT_TIMEOUT:-25m}
E2E_IBC_UPGRADE_TIMEOUT=${E2E_IBC_UPGRADE_TIMEOUT:-45m}
E2E_NETWORK_FAULT_TIMEOUT=${E2E_NETWORK_FAULT_TIMEOUT:-25m}
E2E_RELEASE_UPGRADE_TIMEOUT=${E2E_RELEASE_UPGRADE_TIMEOUT:-35m}
E2E_RELEASE_TOTAL_TIMEOUT_SECONDS=${E2E_RELEASE_TOTAL_TIMEOUT_SECONDS:-21600}
E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS=${E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS:-60}
E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS=${E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS:-5}
E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=${E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS:-43200}
E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=${E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS:-120}
E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=${E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS:-10}
E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS=${E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS:-5}
E2E_LOAD_TIMEOUT=${E2E_LOAD_TIMEOUT:-25m}

E2E_CURRENT_IMAGE_REPOSITORY=${E2E_CURRENT_IMAGE_REPOSITORY:-panacea-e2e-current}
E2E_V221_IMAGE_REPOSITORY=${E2E_V221_IMAGE_REPOSITORY:-panacea-e2e-v2.2.1}
E2E_IMAGE_VERSION=${E2E_IMAGE_VERSION:-local}
E2E_CURRENT_BINARY_VERSION=${E2E_CURRENT_BINARY_VERSION:-2.3.0}
E2E_CURRENT_SOURCE_COMMIT=${E2E_CURRENT_SOURCE_COMMIT:-}
E2E_RELEASE_HARDENING_RUN_ID=${E2E_RELEASE_HARDENING_RUN_ID:-}
E2E_CURRENT_IMAGE="$E2E_CURRENT_IMAGE_REPOSITORY:$E2E_IMAGE_VERSION"
E2E_V221_IMAGE="$E2E_V221_IMAGE_REPOSITORY:$E2E_IMAGE_VERSION"
E2E_DOCKERFILE=${E2E_DOCKERFILE:-"$repo_root/e2e/docker/Dockerfile"}
E2E_V221_COMMIT=${E2E_V221_COMMIT:-a1b342939ba6ac3092aeebbee6a2fa741a34d47f}
E2E_V221_TM_VERSION=${E2E_V221_TM_VERSION:-v0.37.18}
E2E_DOCKER_BUILD_ARGS=${E2E_DOCKER_BUILD_ARGS:-}
DOCKER=${DOCKER:-docker}

if [ -n "$E2E_CURRENT_SOURCE_COMMIT" ]; then
	source_commit=$(git rev-parse --verify "$E2E_CURRENT_SOURCE_COMMIT^{commit}")
elif [ -n "${COMMIT:-}" ]; then
	source_commit=$COMMIT
else
	source_commit=$(git log -1 --format='%H')
fi

paths_checked=0
go_checked=0
current_image_built=0
v221_image_built=0
test_binary_built=0
if [ "${E2E_DOCKER_HOST+x}" = x ]; then
	docker_host_resolved=1
else
	E2E_DOCKER_HOST=
	docker_host_resolved=0
fi

check_paths() {
	if [ "$paths_checked" -eq 1 ]; then
		return 0
	fi
	E2E_ROOT="$E2E_ROOT" \
		E2E_GOCACHE="$E2E_GOCACHE" \
		E2E_GOMODCACHE="$E2E_GOMODCACHE" \
		sh "$script_dir/validate-paths.sh"
	paths_checked=1
}

check_go() {
	check_paths
	if [ "$go_checked" -eq 1 ]; then
		return 0
	fi
	GOWORK=off \
		E2E_GO_VERSION="$E2E_GO_VERSION" \
		E2E_GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
		E2E_GO_BINARY="$E2E_GO_BINARY" \
		sh "$script_dir/check-go-toolchain.sh"
	go_checked=1
}

check_clean() {
	check_paths
	clean_source_status=$(git status --porcelain=v1 --untracked-files=all)
	if [ -n "$clean_source_status" ]; then
		printf 'release-hardening requires a clean HEAD worktree:\n' >&2
		git status --short --untracked-files=all >&2
		exit 2
	fi
	clean_runtime_commit=$(git rev-parse HEAD)
	if [ "$clean_runtime_commit" != "$source_commit" ]; then
		printf 'release-hardening HEAD changed from %s to %s\n' \
			"$source_commit" "$clean_runtime_commit" >&2
		exit 2
	fi
}

resolve_docker_host() {
	if [ "$docker_host_resolved" -eq 1 ]; then
		return 0
	fi
	E2E_DOCKER_HOST=$("$DOCKER" context inspect --format '{{.Endpoints.docker.Host}}' 2>/dev/null || true)
	docker_host_resolved=1
}

prepare_go_dirs() {
	mkdir -p "$E2E_ROOT" "$E2E_GOCACHE" "$E2E_GOMODCACHE"
}

build_current_image_body() (
	set -eu
	stage_dir=$(mktemp -d "$E2E_ROOT/current-build.XXXXXX")
	current_source_dir=$repo_root
	current_tools_dir=$script_dir
	current_dockerfile=$E2E_DOCKERFILE
	build_commit=$source_commit
	trap 'rm -rf "$stage_dir"' EXIT HUP INT TERM
	if [ -n "$E2E_CURRENT_SOURCE_COMMIT" ]; then
		current_source_dir="$stage_dir/source"
		mkdir -p "$current_source_dir"
		build_commit=$(sh "$script_dir/stage-git-source.sh" "$E2E_CURRENT_SOURCE_COMMIT" "$current_source_dir")
		expected_commit=$(git rev-parse --verify "$E2E_CURRENT_SOURCE_COMMIT^{commit}")
		if [ "$build_commit" != "$expected_commit" ]; then
			printf 'staged current source commit %s does not match %s\n' \
				"$build_commit" "$expected_commit" >&2
			exit 2
		fi
		current_tools_dir="$current_source_dir/scripts/e2e"
		current_dockerfile="$current_source_dir/e2e/docker/Dockerfile"
	fi
	(
		cd "$current_source_dir"
		GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
			GOWORK=off \
			GOCACHE="$E2E_GOCACHE" \
			GOMODCACHE="$E2E_GOMODCACHE" \
			"$E2E_GO_BINARY" mod vendor -o "$stage_dir/vendor"
	)
	cmt_version=$(awk '$1 == "github.com/cometbft/cometbft" { if ($2 == "=>") version = $4; else version = $2 } END { print version }' "$current_source_dir/go.mod")
	test -n "$cmt_version"
	# Callers may provide additional, trusted
	# Docker CLI words through E2E_DOCKER_BUILD_ARGS.
	# shellcheck disable=SC2086
	"$DOCKER" build \
		$E2E_DOCKER_BUILD_ARGS \
		--network=none \
		--build-context panacea_vendor="$stage_dir/vendor" \
		--build-context panacea_e2e_tools="$current_tools_dir" \
		--file "$current_dockerfile" \
		--build-arg PANACEA_VERSION="$E2E_CURRENT_BINARY_VERSION" \
		--build-arg PANACEA_COMMIT="$build_commit" \
		--build-arg PANACEA_CMT_VERSION="$cmt_version" \
		--tag "$E2E_CURRENT_IMAGE" \
		"$current_source_dir"
)

build_current_image() {
	if [ "$current_image_built" -eq 1 ]; then
		return 0
	fi
	check_go
	prepare_go_dirs
	build_current_image_body
	current_image_built=1
}

build_v221_image_body() (
	set -eu
	git cat-file -e "$E2E_V221_COMMIT^{commit}"
	stage_dir=$(mktemp -d "$E2E_ROOT/v2.2.1-build.XXXXXX")
	v221_source_dir="$stage_dir/source"
	v221_tools_dir=$script_dir
	v221_dockerfile=$E2E_DOCKERFILE
	mkdir -p "$v221_source_dir"
	trap 'rm -rf "$stage_dir"' EXIT HUP INT TERM
	git archive --format=tar --output="$stage_dir/source.tar" "$E2E_V221_COMMIT"
	tar -xf "$stage_dir/source.tar" -C "$v221_source_dir"
	if [ -n "$E2E_CURRENT_SOURCE_COMMIT" ]; then
		tools_source_dir="$stage_dir/current-source"
		mkdir -p "$tools_source_dir"
		staged_tools_commit=$(sh "$script_dir/stage-git-source.sh" "$E2E_CURRENT_SOURCE_COMMIT" "$tools_source_dir")
		expected_tools_commit=$(git rev-parse --verify "$E2E_CURRENT_SOURCE_COMMIT^{commit}")
		if [ "$staged_tools_commit" != "$expected_tools_commit" ]; then
			printf 'staged current tools commit %s does not match %s\n' \
				"$staged_tools_commit" "$expected_tools_commit" >&2
			exit 2
		fi
		v221_tools_dir="$tools_source_dir/scripts/e2e"
		v221_dockerfile="$tools_source_dir/e2e/docker/Dockerfile"
	fi
	(
		cd "$v221_source_dir"
		GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
			GOWORK=off \
			GOCACHE="$E2E_GOCACHE" \
			GOMODCACHE="$E2E_GOMODCACHE" \
			"$E2E_GO_BINARY" mod vendor -o "$stage_dir/vendor"
	)
	# shellcheck disable=SC2086
	"$DOCKER" build \
		$E2E_DOCKER_BUILD_ARGS \
		--network=none \
		--build-context panacea_vendor="$stage_dir/vendor" \
		--build-context panacea_e2e_tools="$v221_tools_dir" \
		--file "$v221_dockerfile" \
		--build-arg PANACEA_VERSION=2.2.1 \
		--build-arg PANACEA_COMMIT="$E2E_V221_COMMIT" \
		--build-arg PANACEA_TM_VERSION="$E2E_V221_TM_VERSION" \
		--tag "$E2E_V221_IMAGE" \
		"$v221_source_dir"
)

build_v221_image() {
	if [ "$v221_image_built" -eq 1 ]; then
		return 0
	fi
	check_go
	prepare_go_dirs
	build_v221_image_body
	v221_image_built=1
}

build_images() {
	build_current_image
	build_v221_image
}

build_test_binary() {
	if [ "$test_binary_built" -eq 1 ]; then
		return 0
	fi
	check_go
	prepare_go_dirs
	(
		cd "$repo_root/e2e"
		GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
			GOWORK=off \
			GOCACHE="$E2E_GOCACHE" \
			GOMODCACHE="$E2E_GOMODCACHE" \
			"$E2E_GO_BINARY" test -c -o "$E2E_ROOT/panacea-e2e.test" .
	)
	test_binary_built=1
}

unit_body() {
	check_go
	prepare_go_dirs
	GOWORK=off sh "$script_dir/check-go-toolchain-test.sh"
	(
		cd "$repo_root/e2e"
		GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
			GOWORK=off \
			GOCACHE="$E2E_GOCACHE" \
			GOMODCACHE="$E2E_GOMODCACHE" \
			"$E2E_GO_BINARY" test -count=1 ./...
	)
	(
		cd "$repo_root/scripts/e2e/faultproxy"
		GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
			GOWORK=off \
			GOCACHE="$E2E_GOCACHE" \
			GOMODCACHE="$E2E_GOMODCACHE" \
			"$E2E_GO_BINARY" test -count=1 ./...
	)
}

run_current_test() {
	current_suite_flag=$1
	current_suite_timeout=$2
	current_suite_pattern=$3
	build_test_binary
	resolve_docker_host
	mkdir -p "$E2E_ROOT" "$E2E_GOCACHE"
	(
		cd "$repo_root/e2e"
		env \
			"$current_suite_flag=1" \
			PANACEA_E2E_ROOT="$E2E_ROOT" \
			PANACEA_E2E_IMAGE_REPOSITORY="$E2E_CURRENT_IMAGE_REPOSITORY" \
			PANACEA_E2E_IMAGE_VERSION="$E2E_IMAGE_VERSION" \
			DOCKER_HOST="$E2E_DOCKER_HOST" \
			GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
			GOWORK=off \
			GOCACHE="$E2E_GOCACHE" \
			GOMODCACHE="$E2E_GOMODCACHE" \
			"$E2E_ROOT/panacea-e2e.test" \
			-test.timeout "$current_suite_timeout" -test.count=1 -test.v \
			-test.run "$current_suite_pattern"
	)
}

run_v221_test() {
	build_test_binary
	resolve_docker_host
	mkdir -p "$E2E_ROOT" "$E2E_GOCACHE"
	(
		cd "$repo_root/e2e"
		PANACEA_E2E_V221=1 \
			PANACEA_E2E_ROOT="$E2E_ROOT" \
			PANACEA_E2E_V221_IMAGE_REPOSITORY="$E2E_V221_IMAGE_REPOSITORY" \
			PANACEA_E2E_V221_IMAGE_VERSION="$E2E_IMAGE_VERSION" \
			DOCKER_HOST="$E2E_DOCKER_HOST" \
			GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
			GOWORK=off \
			GOCACHE="$E2E_GOCACHE" \
			GOMODCACHE="$E2E_GOMODCACHE" \
			"$E2E_ROOT/panacea-e2e.test" \
			-test.timeout "$E2E_TEST_TIMEOUT" -test.count=1 -test.v \
			-test.run '^TestV221Compatibility$'
	)
}

run_upgrade_test() {
	upgrade_suite_flag=$1
	upgrade_suite_timeout=$2
	upgrade_suite_pattern=$3
	build_test_binary
	resolve_docker_host
	mkdir -p "$E2E_ROOT" "$E2E_GOCACHE"
	(
		cd "$repo_root/e2e"
		env \
			"$upgrade_suite_flag=1" \
			PANACEA_E2E_ROOT="$E2E_ROOT" \
			PANACEA_E2E_IMAGE_REPOSITORY="$E2E_CURRENT_IMAGE_REPOSITORY" \
			PANACEA_E2E_V221_IMAGE_REPOSITORY="$E2E_V221_IMAGE_REPOSITORY" \
			PANACEA_E2E_IMAGE_VERSION="$E2E_IMAGE_VERSION" \
			PANACEA_E2E_V221_IMAGE_VERSION="$E2E_IMAGE_VERSION" \
			PANACEA_E2E_CURRENT_BINARY_VERSION="$E2E_CURRENT_BINARY_VERSION" \
			PANACEA_E2E_CURRENT_COMMIT="$source_commit" \
			DOCKER_HOST="$E2E_DOCKER_HOST" \
			GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
			GOWORK=off \
			GOCACHE="$E2E_GOCACHE" \
			GOMODCACHE="$E2E_GOMODCACHE" \
			"$E2E_ROOT/panacea-e2e.test" \
			-test.timeout "$upgrade_suite_timeout" -test.count=1 -test.v \
			-test.run "$upgrade_suite_pattern"
	)
}

smoke_body() {
	run_current_test PANACEA_E2E "$E2E_TEST_TIMEOUT" '^TestSmokeNodeBoundary$'
	run_current_test PANACEA_E2E_FAILURE_PROBE "$E2E_TEST_TIMEOUT" \
		'^(TestFailureArtifactsAndCleanup|TestUnsupportedDBBackendFailsStartup)$'
}

v221_body() {
	run_v221_test
}

negative_body() {
	run_current_test PANACEA_E2E "$E2E_NEGATIVE_TIMEOUT" \
		'^(TestNFTNegativeStateIntegrity|TestNFTNegativeProtocolBoundaries)$'
}

run_current_test_with_host_port_retry() {
	retry_suite_flag=$1
	retry_suite_timeout=$2
	retry_suite_pattern=$3
	build_test_binary
	resolve_docker_host
	prepare_go_dirs

	attempt=1
	while :; do
		attempt_log=$(mktemp "$E2E_ROOT/host-port-retry.XXXXXX")
		attempt_status_file="$attempt_log.status"
		(
			set +e
			run_current_test "$retry_suite_flag" "$retry_suite_timeout" "$retry_suite_pattern"
			attempt_status=$?
			printf '%s\n' "$attempt_status" >"$attempt_status_file"
			exit 0
		) 2>&1 | tee "$attempt_log"

		if [ ! -s "$attempt_status_file" ]; then
			rm -f "$attempt_log" "$attempt_status_file"
			printf 'live E2E attempt did not record an exit status\n' >&2
			return 125
		fi
		attempt_status=$(sed -n '1p' "$attempt_status_file")
		case "$attempt_status" in
			'' | *[!0-9]*)
				rm -f "$attempt_log" "$attempt_status_file"
				printf 'live E2E attempt recorded invalid exit status: %s\n' "$attempt_status" >&2
				return 125
				;;
		esac
		retryable=0
		if grep -Eq 'failed to bind host port .*address already in use' "$attempt_log"; then
			retryable=1
		fi
		rm -f "$attempt_log" "$attempt_status_file"

		if [ "$attempt_status" -eq 0 ]; then
			return 0
		fi
		if [ "$attempt" -ge 2 ] || [ "$retryable" -ne 1 ]; then
			return "$attempt_status"
		fi
		printf 'retrying live E2E once after Docker host-port allocation race\n' >&2
		attempt=$((attempt + 1))
	done
}

restart_body() {
	run_current_test_with_host_port_retry PANACEA_E2E_RESTART "$E2E_RESTART_TIMEOUT" \
		'^TestRestartRecoveryNodeBoundaries$'
	run_current_test_with_host_port_retry PANACEA_E2E_RESTART "$E2E_RESTART_TIMEOUT" \
		'^TestPortableApplicationSnapshotRestoreAndFreshFullNodeSync$'
}

consensus_body() {
	run_current_test_with_host_port_retry PANACEA_E2E_CONSENSUS "$E2E_CONSENSUS_TIMEOUT" \
		'^TestFourValidatorQuorumFaultAndRecovery$'
}

run_upgrade_normal_test() {
	upgrade_normal_timeout=$1
	run_upgrade_test PANACEA_E2E_UPGRADE "$upgrade_normal_timeout" \
		'^TestV221ToCurrentMultiValidatorUpgrade$'
}

upgrade_body() {
	run_upgrade_normal_test "$E2E_UPGRADE_TIMEOUT"
}

upgrade_legacy_pnft_body() {
	run_upgrade_test PANACEA_E2E_UPGRADE "$E2E_UPGRADE_DEEP_TIMEOUT" \
		'^TestV221ToCurrentLegacyPNFTAdversarialUpgrade$'
}

upgrade_deep_body() {
	run_upgrade_normal_test "$E2E_UPGRADE_DEEP_TIMEOUT"
	upgrade_legacy_pnft_body
}

cosmovisor_body() {
	if [ -z "${PANACEA_REHEARSAL_OLD_TAG:-}" ]; then
		printf 'cosmovisor requires PANACEA_REHEARSAL_OLD_TAG (for example, v2.2.1)\n' >&2
		exit 2
	fi
	if [ -z "${PANACEA_REHEARSAL_UPGRADE_NAME:-}" ]; then
		printf 'cosmovisor requires PANACEA_REHEARSAL_UPGRADE_NAME (for example, v2.3.0)\n' >&2
		exit 2
	fi
	if [ "$PANACEA_REHEARSAL_OLD_TAG" = "$PANACEA_REHEARSAL_UPGRADE_NAME" ]; then
		printf 'cosmovisor old tag and upgrade name must differ: %s\n' \
			"$PANACEA_REHEARSAL_OLD_TAG" >&2
		exit 2
	fi

	check_go
	rehearsal_root=${PANACEA_REHEARSAL_ROOT:-"$E2E_ROOT/cosmovisor"}
	GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
		GOWORK=off \
		PANACEA_REHEARSAL_ROOT="$rehearsal_root" \
		PANACEA_REHEARSAL_OLD_TAG="$PANACEA_REHEARSAL_OLD_TAG" \
		PANACEA_REHEARSAL_UPGRADE_NAME="$PANACEA_REHEARSAL_UPGRADE_NAME" \
		"$repo_root/scripts/upgrade-local/run.sh" run
}

upgrade_chaos_body() {
	run_upgrade_test PANACEA_E2E_UPGRADE_CHAOS "$E2E_UPGRADE_CHAOS_TIMEOUT" \
		'^TestV221UpgradeBoundaryChaos$'
}

state_sync_body() {
	run_current_test PANACEA_E2E_STATE_SYNC "$E2E_STATE_SYNC_TIMEOUT" \
		'^TestActualCometStateSyncAndBadTrustHash$'
}

config_compat_body() {
	run_upgrade_test PANACEA_E2E_CONFIG_COMPAT "$E2E_CONFIG_COMPAT_TIMEOUT" \
		'^TestV047NodeHomeConfigCompatibility$'
}

ibc_upgrade_body() {
	run_upgrade_test PANACEA_E2E_IBC_UPGRADE "$E2E_IBC_UPGRADE_TIMEOUT" \
		'^TestIBCUpgradeContinuity$'
}

network_faults_body() {
	run_current_test PANACEA_E2E_NETWORK_FAULTS "$E2E_NETWORK_FAULT_TIMEOUT" \
		'^TestLocalDockerNetworkAndEndpointFaults$'
}

load_body() {
	run_current_test PANACEA_E2E_LOAD "$E2E_LOAD_TIMEOUT" \
		'^TestFullNodeLoadAndResourceBaseline$'
}

release_builds_body() {
	resolve_docker_host
	functional_images_prebuilt=0
	if [ "$current_image_built" -eq 1 ] && [ "$v221_image_built" -eq 1 ]; then
		functional_images_prebuilt=1
	fi
	release_total=$E2E_RELEASE_TOTAL_TIMEOUT_SECONDS
	if [ -n "${E2E_RELEASE_AGGREGATE_WORK_DEADLINE_EPOCH:-}" ]; then
		now_epoch=$(date -u +%s)
		remaining_seconds=$((E2E_RELEASE_AGGREGATE_WORK_DEADLINE_EPOCH - now_epoch))
		child_margin=${E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS:-5}
		child_total=$((remaining_seconds - child_margin))
		child_reserved=$((E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS + E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS))
		if [ "$child_total" -le "$child_reserved" ]; then
			printf 'insufficient aggregate budget remains for bounded release-build cleanup: remaining=%s margin=%s\n' \
				"$remaining_seconds" "$child_margin" >&2
			exit 2
		fi
		if [ "$release_total" -gt "$child_total" ]; then
			release_total=$child_total
		fi
	fi
	PANACEA_E2E_RELEASE_HARDENING=1 \
		PANACEA_E2E_RELEASE_MULTIARCH_UPGRADE=1 \
		E2E_ROOT="$E2E_ROOT" \
		E2E_GOCACHE="$E2E_GOCACHE" \
		E2E_GOMODCACHE="$E2E_GOMODCACHE" \
		E2E_GO_VERSION="$E2E_GO_VERSION" \
		E2E_GO_BINARY="$E2E_GO_BINARY" \
		E2E_GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
		E2E_CURRENT_BINARY_VERSION="$E2E_CURRENT_BINARY_VERSION" \
		E2E_V221_COMMIT="$E2E_V221_COMMIT" \
		E2E_V221_TM_VERSION="$E2E_V221_TM_VERSION" \
		E2E_RELEASE_UPGRADE_TIMEOUT="$E2E_RELEASE_UPGRADE_TIMEOUT" \
		E2E_RELEASE_TOTAL_TIMEOUT_SECONDS="$release_total" \
		E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS="$E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS" \
		E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS="$E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS" \
		E2E_FUNCTIONAL_CURRENT_IMAGE="$E2E_CURRENT_IMAGE" \
		E2E_FUNCTIONAL_OLD_IMAGE="$E2E_V221_IMAGE" \
		E2E_FUNCTIONAL_IMAGES_PREBUILT="$functional_images_prebuilt" \
		E2E_RUNNER="$runner" \
		DOCKER="$DOCKER" \
		DOCKER_HOST="$E2E_DOCKER_HOST" \
		sh "$script_dir/release-hardening.sh"
}

release_hardening_body() {
	PANACEA_E2E_RELEASE_AGGREGATE=1 \
		E2E_RELEASE_AGGREGATE_BASE_ROOT="$E2E_ROOT" \
		E2E_RELEASE_HARDENING_RUN_ID="$E2E_RELEASE_HARDENING_RUN_ID" \
		E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS="$E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS" \
		E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS="$E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS" \
		E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS="$E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS" \
		E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS="$E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS" \
		E2E_RUNNER="$runner" \
		DOCKER="$DOCKER" \
		sh "$script_dir/release-hardening-aggregate.sh"
}

coverage_merge() {
	(
		cd "$repo_root/e2e"
		GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
			GOWORK=off \
			GOCACHE="$E2E_GOCACHE" \
			GOMODCACHE="$E2E_GOMODCACHE" \
			"$E2E_GO_BINARY" run ./cmd/coverage-merge \
			-root "$E2E_ROOT" \
			-output 'upgrade/coverage-matrix.json' \
			-source-commit "$source_commit"
	)
}

release_hardening_inner_body() {
	check_clean
	all_body
	release_builds_body
	check_clean
	coverage_merge
}

all_body() {
	unit_body
	# Every functional lane shares these exact two images. Build each once.
	build_images
	smoke_body
	v221_body
	negative_body
	restart_body
	consensus_body
	upgrade_deep_body
	upgrade_chaos_body
	ibc_upgrade_body
	state_sync_body
	config_compat_body
	network_faults_body
	load_body
}

case "$command_name" in
	check) check_go ;;
	check-clean) check_clean ;;
	build-current) build_current_image ;;
	build-v2.2.1) build_v221_image ;;
	build-images) build_images ;;
	build-test-binary) build_test_binary ;;
	build)
		build_images
		build_test_binary
		;;
	unit) unit_body ;;
	smoke)
		build_current_image
		smoke_body
		;;
	v2.2.1)
		build_v221_image
		v221_body
		;;
	compatibility)
		build_current_image
		smoke_body
		build_v221_image
		v221_body
		;;
	negative)
		build_current_image
		negative_body
		;;
	restart)
		build_current_image
		restart_body
		;;
	consensus)
		build_current_image
		consensus_body
		;;
	upgrade)
		build_images
		upgrade_body
		;;
	cosmovisor) cosmovisor_body ;;
	upgrade-deep)
		build_images
		upgrade_deep_body
		;;
	upgrade-chaos)
		build_images
		upgrade_chaos_body
		;;
	state-sync)
		build_current_image
		state_sync_body
		;;
	config-compat)
		build_images
		config_compat_body
		;;
	ibc-upgrade)
		build_images
		ibc_upgrade_body
		;;
	network-faults)
		build_current_image
		network_faults_body
		;;
	release-builds) release_builds_body ;;
	release-hardening) release_hardening_body ;;
	release-hardening-inner) release_hardening_inner_body ;;
	load)
		build_current_image
		load_body
		;;
	all) all_body ;;
esac
