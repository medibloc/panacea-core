#!/bin/sh

# Reproducible, multi-architecture Panacea release evidence.
#
# This runner is intentionally opt-in: it materializes dependencies from the
# network, creates a fresh isolated BuildKit builder, builds linux/amd64 and
# linux/arm64 images, and runs a real v2.2.1 -> current upgrade on both images.
# It never treats an unavailable cross-architecture runtime as a passing skip.

set -eu

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd -P)
cd "$repo_root"

E2E_ROOT=${E2E_ROOT:-"$repo_root/.local/e2e"}
E2E_GOCACHE=${E2E_GOCACHE:-"$E2E_ROOT/go-build"}
E2E_GOMODCACHE=${E2E_GOMODCACHE:-"$E2E_ROOT/go-mod"}
E2E_GO_BINARY=${E2E_GO_BINARY:-go}
E2E_GOTOOLCHAIN=${E2E_GOTOOLCHAIN:-local}
E2E_GO_VERSION=${E2E_GO_VERSION:-1.26.5}
E2E_CURRENT_BINARY_VERSION=${E2E_CURRENT_BINARY_VERSION:-2.3.0}
E2E_V221_COMMIT=${E2E_V221_COMMIT:-a1b342939ba6ac3092aeebbee6a2fa741a34d47f}
E2E_V221_TM_VERSION=${E2E_V221_TM_VERSION:-v0.37.18}
E2E_RELEASE_UPGRADE_TIMEOUT=${E2E_RELEASE_UPGRADE_TIMEOUT:-35m}
E2E_RELEASE_TOTAL_TIMEOUT_SECONDS=${E2E_RELEASE_TOTAL_TIMEOUT_SECONDS:-21600}
E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS=${E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS:-60}
E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS=${E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS:-5}
E2E_RELEASE_WORK_TIMEOUT_SECONDS=unvalidated
E2E_FUNCTIONAL_CURRENT_IMAGE=${E2E_FUNCTIONAL_CURRENT_IMAGE:-panacea-e2e-current:local}
E2E_FUNCTIONAL_OLD_IMAGE=${E2E_FUNCTIONAL_OLD_IMAGE:-panacea-e2e-v2.2.1:local}
E2E_RUNNER=${E2E_RUNNER:-"$repo_root/scripts/e2e/run.sh"}
DOCKER=${DOCKER:-docker}

run_id="release-$(date -u +%Y%m%d%H%M%S)-$$"
# Bootstrap evidence always starts below the repository-owned E2E root. This
# remains safe even when an invalid caller-provided E2E_ROOT is the failure.
# After path validation, the run is moved under E2E_ROOT so aggregate evidence
# remains scoped to the caller's run root.
bootstrap_artifact_root="$repo_root/.local/e2e"
release_artifact_root="$bootstrap_artifact_root"
artifact_dir="$release_artifact_root/$run_id"
work_dir="$release_artifact_root/$run_id-work"
mkdir -p "$artifact_dir" "$work_dir"
chmod 700 "$artifact_dir" "$work_dir"

stage=initializing
builder_name="panacea-$run_id"
builder_created=0
created_images=""
created_containers=""
cleanup_failed=0
source_commit=unknown
watchdog_pid=
timeout_marker="$artifact_dir/overall-timeout.txt"
watchdog_pid_file="$artifact_dir/watchdog-pid.txt"
watchdog_timer_pid_file="$artifact_dir/watchdog-timer-pid.txt"
watchdog_trace="$artifact_dir/watchdog-trace.txt"

write_status() {
  result=$1
  exit_code=$2
  {
    printf 'result=%s\n' "$result"
    printf 'exit_code=%s\n' "$exit_code"
    printf 'stage=%s\n' "$stage"
    printf 'run_id=%s\n' "$run_id"
    printf 'source_commit=%s\n' "$source_commit"
    printf 'artifact_dir=%s\n' "$artifact_dir"
    printf 'total_timeout_seconds=%s\n' "$E2E_RELEASE_TOTAL_TIMEOUT_SECONDS"
    printf 'cleanup_timeout_seconds=%s\n' "$E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS"
    printf 'force_exit_timeout_seconds=%s\n' "$E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS"
    printf 'work_timeout_seconds=%s\n' "$E2E_RELEASE_WORK_TIMEOUT_SECONDS"
  } >"$artifact_dir/status.txt"
}

release_parent_pid() (
  inspected_pid=$1
  case "$inspected_pid" in
    '' | *[!0-9]*) return 1 ;;
  esac
  release_parent_output=$(ps -o ppid= -p "$inspected_pid" 2>/dev/null) || exit $?
  printf '%s\n' "$release_parent_output" | awk 'NR == 1 { print $1 }'
)

# Revalidate ancestry immediately before every signal. Besides avoiding stale
# process snapshots, this prevents a recycled PID from targeting a process
# outside this runner's tree.
release_is_descendant() {
  candidate_pid=$1
  ancestor_pid=$2
  ancestry_hops=0
  case "$candidate_pid:$ancestor_pid" in
    *[!0-9:]*) return 1 ;;
  esac
  while [ "$ancestry_hops" -lt 1024 ]; do
    candidate_parent=$(release_parent_pid "$candidate_pid") || return 1
    if [ "$candidate_parent" = "$ancestor_pid" ]; then
      return 0
    fi
    case "$candidate_parent" in
      '' | *[!0-9]* | 0 | 1) return 1 ;;
    esac
    if [ "$candidate_parent" = "$candidate_pid" ]; then
      return 1
    fi
    candidate_pid=$candidate_parent
    ancestry_hops=$((ancestry_hops + 1))
  done
  return 1
}

release_child_pids() (
  inspected_parent=$1
  release_process_table=$(ps -eo pid=,ppid= 2>/dev/null) || exit $?
  printf '%s\n' "$release_process_table" |
    awk -v inspected_parent="$inspected_parent" '$2 == inspected_parent { print $1 }'
)

release_signal_process_tree() (
  tree_pid=$1
  runner_pid=$2
  excluded_pid=$3
  signal_name=$4
  case "$signal_name" in
    TERM | KILL) ;;
    *) return 1 ;;
  esac
  if [ "$tree_pid" = "$excluded_pid" ]; then
    return 0
  fi
  release_child_pids "$tree_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    release_signal_process_tree "$child_pid" "$runner_pid" "$excluded_pid" "$signal_name"
  done
  # During the graceful phase keep an ancestor alive while any direct child
  # remains. Killing the ancestor first could reparent a TERM-resistant child
  # and make ownership impossible to prove safely. The hard KILL traversal is
  # leaf-first and may terminate every still-owned node in one pass.
  if [ "$signal_name" = TERM ] &&
    [ -n "$(release_child_pids "$tree_pid" | awk 'NR == 1 { print $1 }')" ]; then
    return 0
  fi
  if release_is_descendant "$tree_pid" "$runner_pid"; then
    kill "-$signal_name" "$tree_pid" 2>/dev/null || true
  fi
)

release_signal_runner_children() {
  runner_pid=$1
  excluded_pid=$2
  signal_name=$3
  release_child_pids "$runner_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    release_signal_process_tree "$child_pid" "$runner_pid" "$excluded_pid" "$signal_name"
  done
}

# Freeze ancestors before inspecting their children. SIGSTOP cannot be ignored,
# so a TERM-resistant process cannot fork a new child between the final
# descendant snapshot and KILL, then leave that child orphaned.
release_stop_process_tree() (
  tree_pid=$1
  runner_pid=$2
  excluded_pid=$3
  if [ "$tree_pid" = "$excluded_pid" ]; then
    return 0
  fi
  if release_is_descendant "$tree_pid" "$runner_pid"; then
    kill -STOP "$tree_pid" 2>/dev/null || true
  else
    return 0
  fi
  release_child_pids "$tree_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    release_stop_process_tree "$child_pid" "$runner_pid" "$excluded_pid"
  done
)

release_stop_runner_children() {
  runner_pid=$1
  excluded_pid=$2
  release_child_pids "$runner_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    release_stop_process_tree "$child_pid" "$runner_pid" "$excluded_pid"
  done
}

release_runner_has_children() (
  inspected_runner=$1
  excluded_pid=$2
  child_pid_list=$(release_child_pids "$inspected_runner" || true)
  for child_pid in $child_pid_list; do
    if [ "$child_pid" != "$excluded_pid" ] && release_is_descendant "$child_pid" "$inspected_runner"; then
      return 0
    fi
  done
  return 1
)

release_watchdog_is_owned() {
  owned_watchdog_pid=$1
  owned_runner_pid=$2
  [ "$(release_parent_pid "$owned_watchdog_pid" 2>/dev/null || true)" = "$owned_runner_pid" ]
}

write_forced_timeout_status() {
  forced_stage=$1
  timeout_source_commit=unknown
  if [ -s "$artifact_dir/source-commit.txt" ]; then
    IFS= read -r timeout_source_commit <"$artifact_dir/source-commit.txt" || timeout_source_commit=unknown
  fi
  {
    printf 'result=failed\n'
    printf 'exit_code=124\n'
    printf 'stage=%s\n' "$forced_stage"
    printf 'cleanup_result=timed-out\n'
  } >"$artifact_dir/failure.txt.tmp"
  mv "$artifact_dir/failure.txt.tmp" "$artifact_dir/failure.txt"
  {
    printf 'result=failed\n'
    printf 'exit_code=124\n'
    printf 'stage=%s\n' "$forced_stage"
    printf 'run_id=%s\n' "$run_id"
    printf 'source_commit=%s\n' "$timeout_source_commit"
    printf 'artifact_dir=%s\n' "$artifact_dir"
    printf 'total_timeout_seconds=%s\n' "$E2E_RELEASE_TOTAL_TIMEOUT_SECONDS"
    printf 'cleanup_timeout_seconds=%s\n' "$E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS"
    printf 'force_exit_timeout_seconds=%s\n' "$E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS"
    printf 'work_timeout_seconds=%s\n' "$E2E_RELEASE_WORK_TIMEOUT_SECONDS"
  } >"$artifact_dir/status.txt.tmp"
  mv "$artifact_dir/status.txt.tmp" "$artifact_dir/status.txt"
}

release_watchdog_sleep() {
  watchdog_sleep_seconds=$1
  sleep "$watchdog_sleep_seconds" &
  watchdog_timer_pid=$!
  printf '%s\n' "$watchdog_timer_pid" >"$watchdog_timer_pid_file"
  if wait "$watchdog_timer_pid"; then
    watchdog_sleep_result=0
  else
    watchdog_sleep_result=$?
  fi
  rm -f "$watchdog_timer_pid_file"
  return "$watchdog_sleep_result"
}

track_release_container() {
  created_containers="$created_containers $1"
}

untrack_release_container() {
  tracked_container=$1
  created_containers=$(printf '%s\n' "$created_containers" | sed "s/[[:space:]]$tracked_container//")
}

cleanup_harness_run_id() {
  cleanup_run_id=$1
  cleanup_log="$artifact_dir/cleanup-harness-$cleanup_run_id.txt"
  case "$cleanup_run_id" in
    run-????????????) ;;
    *)
      printf 'refused invalid harness cleanup run ID: %s\n' "$cleanup_run_id" >"$cleanup_log"
      return 1
      ;;
  esac
  if printf '%s\n' "$cleanup_run_id" | grep -Eqv '^run-[0-9a-f]{12}$'; then
    printf 'refused invalid harness cleanup run ID: %s\n' "$cleanup_run_id" >"$cleanup_log"
    return 1
  fi

  harness_cleanup_failed=0
  : >"$cleanup_log"
  if ! harness_container_ids=$("$DOCKER" ps -aq --filter "label=ibc-test=$cleanup_run_id" 2>>"$cleanup_log"); then
    printf 'failed to list harness containers\n' >>"$cleanup_log"
    return 1
  fi
  for harness_container_id in $harness_container_ids; do
    if ! "$DOCKER" rm -f "$harness_container_id" >>"$cleanup_log" 2>&1; then
      harness_cleanup_failed=1
    fi
  done
  if ! harness_volume_names=$("$DOCKER" volume ls -q --filter "label=ibc-test=$cleanup_run_id" 2>>"$cleanup_log"); then
    printf 'failed to list harness volumes\n' >>"$cleanup_log"
    return 1
  fi
  for harness_volume_name in $harness_volume_names; do
    if ! "$DOCKER" volume rm -f "$harness_volume_name" >>"$cleanup_log" 2>&1; then
      harness_cleanup_failed=1
    fi
  done
  if ! harness_network_ids=$("$DOCKER" network ls -q --filter "label=ibc-test=$cleanup_run_id" 2>>"$cleanup_log"); then
    printf 'failed to list harness networks\n' >>"$cleanup_log"
    return 1
  fi
  for harness_network_id in $harness_network_ids; do
    if ! "$DOCKER" network rm "$harness_network_id" >>"$cleanup_log" 2>&1; then
      harness_cleanup_failed=1
    fi
  done

  harness_remaining_containers=$("$DOCKER" ps -aq --filter "label=ibc-test=$cleanup_run_id" 2>>"$cleanup_log") || return 1
  harness_remaining_volumes=$("$DOCKER" volume ls -q --filter "label=ibc-test=$cleanup_run_id" 2>>"$cleanup_log") || return 1
  harness_remaining_networks=$("$DOCKER" network ls -q --filter "label=ibc-test=$cleanup_run_id" 2>>"$cleanup_log") || return 1
  if [ -n "$harness_remaining_containers$harness_remaining_volumes$harness_remaining_networks" ]; then
    printf 'labeled harness resources remain after cleanup\n' >>"$cleanup_log"
    harness_cleanup_failed=1
  fi
  [ "$harness_cleanup_failed" -eq 0 ]
}

cleanup_release_hardening() {
  saved_code=$?
  trap - EXIT HUP INT TERM

  if [ -s "$timeout_marker" ]; then
    saved_code=124
    stage=overall-timeout
  fi

  for upgrade_run_dir in "$artifact_dir"/upgrade-*/run-*; do
    [ -d "$upgrade_run_dir" ] || continue
    upgrade_run_id=${upgrade_run_dir##*/}
    if ! cleanup_harness_run_id "$upgrade_run_id"; then
      cleanup_failed=1
    fi
  done

  for container_name in $created_containers; do
    container_cleanup_log="$artifact_dir/cleanup-container-$container_name.txt"
    if ! "$DOCKER" rm -f "$container_name" >"$container_cleanup_log" 2>&1; then
      if ! grep -Fq 'No such container' "$container_cleanup_log"; then
        cleanup_failed=1
      fi
    fi
  done
  if [ "$builder_created" -eq 1 ]; then
    if ! "$DOCKER" buildx rm "$builder_name" >"$artifact_dir/cleanup-builder.txt" 2>&1; then
      cleanup_failed=1
    fi
  fi
  if [ "${PANACEA_E2E_RELEASE_KEEP_IMAGES:-0}" != "1" ]; then
    for image_ref in $created_images; do
      image_cleanup_name=$(printf '%s' "$image_ref" | tr '/:' '__')
      image_cleanup_log="$artifact_dir/cleanup-image-$image_cleanup_name.txt"
      if ! "$DOCKER" image rm "$image_ref" >"$image_cleanup_log" 2>&1; then
        if ! grep -Fq 'No such image' "$image_cleanup_log"; then
          cleanup_failed=1
        fi
      fi
    done
  fi

  # The work directory may still be at the trusted bootstrap location if path
  # validation or relocation failed. Accept only either exact run-owned shape.
  case "$work_dir" in
    "$bootstrap_artifact_root"/release-*-work | "$release_artifact_root"/release-*-work)
      # Go's module cache deliberately makes module directories read-only.
      # Restore owner write permission on directories only so rm can unlink
      # their contents without following or mutating symlink targets.
      if ! find "$work_dir" -type d -exec chmod u+w {} +; then
        cleanup_failed=1
      fi
      if ! rm -rf -- "$work_dir"; then
        cleanup_failed=1
      fi
      ;;
    *)
      printf 'refused unsafe work directory cleanup: %s\n' "$work_dir" >"$artifact_dir/cleanup-work-dir-error.txt"
      cleanup_failed=1
      ;;
  esac

  # Keep the watchdog alive through resource cleanup. Otherwise a wedged
  # Docker daemon could make the supposedly bounded runner wait forever.
  if [ -n "$watchdog_pid" ]; then
    # $watchdog_pid is this shell's unreaped $! child, so it cannot be recycled
    # before wait. Freeze it, kill the exact timer PID it published, then reap
    # both without depending on ps (which may be denied in a CI sandbox).
    kill -STOP "$watchdog_pid" 2>/dev/null || true
    watchdog_timer_pid=
    if [ -s "$watchdog_timer_pid_file" ]; then
      IFS= read -r watchdog_timer_pid <"$watchdog_timer_pid_file" || watchdog_timer_pid=
    fi
    case "$watchdog_timer_pid" in
      '' | *[!0-9]*) ;;
      *) kill -KILL "$watchdog_timer_pid" 2>/dev/null || true ;;
    esac
    kill -KILL "$watchdog_pid" 2>/dev/null || true
    wait "$watchdog_pid" 2>/dev/null || true
    rm -f "$watchdog_timer_pid_file"
  fi

  if [ "$saved_code" -eq 0 ] && [ "$cleanup_failed" -eq 0 ]; then
    write_status passed 0
    exit 0
  fi
  if [ "$saved_code" -eq 0 ]; then
    saved_code=1
    stage=cleanup
  fi
  {
    printf 'result=failed\n'
    printf 'exit_code=%s\n' "$saved_code"
    printf 'stage=%s\n' "$stage"
    if [ "$cleanup_failed" -eq 0 ]; then
      printf 'cleanup_result=passed\n'
    else
      printf 'cleanup_result=failed\n'
    fi
  } >"$artifact_dir/failure.txt"
  write_status failed "$saved_code"
  exit "$saved_code"
}
trap cleanup_release_hardening EXIT
handle_release_signal() {
  if [ -s "$timeout_marker" ]; then
    stage=overall-timeout
    exit 124
  fi
  stage=interrupted
  exit 130
}
trap handle_release_signal HUP INT TERM

write_status running 0

stage=validate-opt-in
if [ "${PANACEA_E2E_RELEASE_HARDENING:-}" != "1" ]; then
  echo "PANACEA_E2E_RELEASE_HARDENING=1 is required; failure evidence was written to $artifact_dir" >&2
  exit 2
fi

case "$E2E_RELEASE_TOTAL_TIMEOUT_SECONDS" in
  '' | *[!0-9]* | 0)
    echo "E2E_RELEASE_TOTAL_TIMEOUT_SECONDS must be a positive integer" >&2
    exit 2
    ;;
esac
case "$E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS" in
  '' | *[!0-9]* | 0)
    echo "E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS must be a positive integer" >&2
    exit 2
    ;;
esac
case "$E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS" in
  '' | *[!0-9]* | 0)
    echo "E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS must be a positive integer" >&2
    exit 2
    ;;
esac
release_reserved_timeout_seconds=$((E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS + E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS))
if [ "$E2E_RELEASE_TOTAL_TIMEOUT_SECONDS" -le "$release_reserved_timeout_seconds" ]; then
  echo "E2E_RELEASE_TOTAL_TIMEOUT_SECONDS must exceed cleanup + force-exit timeouts" >&2
  exit 2
fi
E2E_RELEASE_WORK_TIMEOUT_SECONDS=$((E2E_RELEASE_TOTAL_TIMEOUT_SECONDS - release_reserved_timeout_seconds))
start_release_watchdog() {
  runner_pid=$$
  (
    trap - EXIT HUP INT TERM
    if ! release_watchdog_sleep "$E2E_RELEASE_WORK_TIMEOUT_SECONDS"; then
      exit 0
    fi
    {
      printf 'total_timeout_seconds=%s\n' "$E2E_RELEASE_TOTAL_TIMEOUT_SECONDS"
      printf 'work_timeout_seconds=%s\n' "$E2E_RELEASE_WORK_TIMEOUT_SECONDS"
    } >"$timeout_marker"
    watchdog_identity=$(awk 'NR == 1 { print $1 }' "$watchdog_pid_file" 2>/dev/null || true)
    if ! release_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
      write_forced_timeout_status watchdog-identity-invalid
      exit 1
    fi
    cleanup_seconds_remaining=$E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS
    runner_signalled=0
    printf 'event=deadline watchdog_pid=%s runner_pid=%s work_timeout_seconds=%s\n' \
      "$watchdog_identity" "$runner_pid" "$E2E_RELEASE_WORK_TIMEOUT_SECONDS" >"$watchdog_trace"
    while [ "$cleanup_seconds_remaining" -gt 0 ]; do
      release_signal_runner_children "$runner_pid" "$watchdog_identity" TERM
      watchdog_children=$(release_child_pids "$runner_pid" | awk -v excluded="$watchdog_identity" '$1 != excluded { printf "%s ", $1 }')
      printf 'event=term-sweep remaining=%s children=%s\n' \
        "$cleanup_seconds_remaining" "$watchdog_children" >>"$watchdog_trace"
      if ! release_runner_has_children "$runner_pid" "$watchdog_identity"; then
        if release_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
          kill -TERM "$runner_pid" 2>/dev/null || true
        fi
        runner_signalled=1
        printf 'event=runner-term\n' >>"$watchdog_trace"
        break
      fi
      if ! release_watchdog_sleep 1; then
        exit 0
      fi
      cleanup_seconds_remaining=$((cleanup_seconds_remaining - 1))
    done
    if [ "$runner_signalled" -eq 1 ] && [ "$cleanup_seconds_remaining" -gt 0 ]; then
      if ! release_watchdog_sleep "$cleanup_seconds_remaining"; then
        exit 0
      fi
    fi
    if ! release_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
      exit 0
    fi
    {
      printf 'result=timed-out\n'
      printf 'timeout_seconds=%s\n' "$E2E_RELEASE_TOTAL_TIMEOUT_SECONDS"
      printf 'cleanup_timeout_seconds=%s\n' "$E2E_RELEASE_CLEANUP_TIMEOUT_SECONDS"
      printf 'force_exit_timeout_seconds=%s\n' "$E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS"
      printf 'work_timeout_seconds=%s\n' "$E2E_RELEASE_WORK_TIMEOUT_SECONDS"
    } >"$artifact_dir/cleanup-timeout.txt"
    write_forced_timeout_status cleanup-timeout
    printf 'event=cleanup-timeout\n' >>"$watchdog_trace"
    if release_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
      kill -STOP "$runner_pid" 2>/dev/null || true
    fi
    release_stop_runner_children "$runner_pid" "$watchdog_identity"
    release_signal_runner_children "$runner_pid" "$watchdog_identity" KILL
    if release_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
      printf 'event=runner-cont-after-child-kill\n' >>"$watchdog_trace"
      kill -CONT "$runner_pid" 2>/dev/null || true
    fi
    # A forced child KILL can finally unblock the shell. Give its EXIT trap a
    # separately bounded cleanup window before the last fail-closed sweep.
    if ! release_watchdog_sleep "$E2E_RELEASE_FORCE_EXIT_TIMEOUT_SECONDS"; then
      exit 0
    fi
    if release_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
      printf 'event=runner-hard-stop\n' >>"$watchdog_trace"
      kill -STOP "$runner_pid" 2>/dev/null || true
      release_stop_runner_children "$runner_pid" "$watchdog_identity"
      release_signal_runner_children "$runner_pid" "$watchdog_identity" KILL
      printf 'event=runner-kill\n' >>"$watchdog_trace"
      kill -KILL "$runner_pid" 2>/dev/null || true
    fi
  ) &
  watchdog_pid=$!
  printf '%s\n' "$watchdog_pid" >"$watchdog_pid_file"
}

stage=validate-paths
export E2E_ROOT E2E_GOCACHE E2E_GOMODCACHE E2E_GO_BINARY E2E_GOTOOLCHAIN E2E_GO_VERSION
sh "$script_dir/validate-paths.sh" >"$artifact_dir/path-validation.txt" 2>&1

stage=relocate-artifacts
mkdir -p "$E2E_ROOT"
# Revalidate after creation so a changed existing ancestor or symlink cannot
# silently redirect writes between the preflight and relocation.
sh "$script_dir/validate-paths.sh" >>"$artifact_dir/path-validation.txt" 2>&1
validated_artifact_root=$E2E_ROOT
while [ "$validated_artifact_root" != "/" ] &&
  [ "${validated_artifact_root%/}" != "$validated_artifact_root" ]; do
  validated_artifact_root=${validated_artifact_root%/}
done
if [ "$validated_artifact_root" != "$release_artifact_root" ]; then
  relocated_artifact_dir="$validated_artifact_root/$run_id"
  relocated_work_dir="$validated_artifact_root/$run_id-work"
  if [ -e "$relocated_artifact_dir" ] || [ -e "$relocated_work_dir" ]; then
    echo "refusing to reuse release artifact paths under $validated_artifact_root" >&2
    exit 2
  fi
  mv "$work_dir" "$relocated_work_dir"
  work_dir=$relocated_work_dir
  release_artifact_root=$validated_artifact_root
  mv "$artifact_dir" "$relocated_artifact_dir"
  artifact_dir=$relocated_artifact_dir
  timeout_marker="$artifact_dir/overall-timeout.txt"
  watchdog_pid_file="$artifact_dir/watchdog-pid.txt"
  watchdog_timer_pid_file="$artifact_dir/watchdog-timer-pid.txt"
  watchdog_trace="$artifact_dir/watchdog-trace.txt"
  write_status running 0
fi

stage=validate-process-control
release_child_pids "$$" >/dev/null
start_release_watchdog

stage=validate-clean-source
source_commit=$(git rev-parse HEAD)
printf '%s\n' "$source_commit" >"$artifact_dir/source-commit.txt"
git status --porcelain=v1 --untracked-files=all >"$artifact_dir/source-status.txt"
git diff --binary HEAD -- >"$artifact_dir/source-diff.patch"
if [ -s "$artifact_dir/source-status.txt" ] || [ -s "$artifact_dir/source-diff.patch" ]; then
  echo "release hardening requires a clean HEAD worktree (tracked and untracked files)" >&2
  git status --short --untracked-files=all >&2
  exit 2
fi

stage=prepare-current-source
current_source="$work_dir/current-source"
mkdir -p "$current_source"
staged_source_commit=$(sh "$script_dir/stage-git-source.sh" "$source_commit" "$current_source")
if [ "$staged_source_commit" != "$source_commit" ]; then
  echo "staged current source commit $staged_source_commit does not match $source_commit" >&2
  exit 2
fi

stage=resolve-docker-context
if [ -z "${DOCKER_HOST:-}" ]; then
  DOCKER_HOST=$("$DOCKER" context inspect --format '{{.Endpoints.docker.Host}}')
fi
if [ -z "$DOCKER_HOST" ]; then
  echo "Docker context did not provide a daemon endpoint" >&2
  exit 2
fi
export DOCKER_HOST

stage=validate-go-toolchain
sh "$script_dir/check-go-toolchain.sh" >"$artifact_dir/go-toolchain-validation.txt" 2>&1

stage=build-functional-images
E2E_ROOT="$E2E_ROOT" \
	E2E_GOCACHE="$E2E_GOCACHE" \
	E2E_GOMODCACHE="$E2E_GOMODCACHE" \
	E2E_GO_VERSION="$E2E_GO_VERSION" \
	E2E_GO_BINARY="$E2E_GO_BINARY" \
	E2E_GOTOOLCHAIN="$E2E_GOTOOLCHAIN" \
	E2E_CURRENT_BINARY_VERSION="$E2E_CURRENT_BINARY_VERSION" \
	E2E_CURRENT_SOURCE_COMMIT="$source_commit" \
	E2E_V221_COMMIT="$E2E_V221_COMMIT" \
	E2E_V221_TM_VERSION="$E2E_V221_TM_VERSION" \
	E2E_DOCKER_HOST="$DOCKER_HOST" \
	DOCKER="$DOCKER" \
	"$E2E_RUNNER" build-images \
	>"$artifact_dir/functional-image-build.log" 2>&1

stage=validate-clean-source
git status --porcelain=v1 --untracked-files=all >"$artifact_dir/source-status-after-functional-build.txt"
git diff --binary HEAD -- >"$artifact_dir/source-diff-after-functional-build.patch"
if [ -s "$artifact_dir/source-status-after-functional-build.txt" ] || [ -s "$artifact_dir/source-diff-after-functional-build.patch" ]; then
  echo "release source changed after the clean-source preflight" >&2
  exit 2
fi
git ls-tree -r --name-only "$source_commit" >"$artifact_dir/source-files.txt"
: >"$artifact_dir/source-files-sha256.txt"
while IFS= read -r source_file; do
  if command -v sha256sum >/dev/null 2>&1; then
    source_sha256=$(sha256sum "$current_source/$source_file" | awk 'NR == 1 { print $1 }')
  else
    source_sha256=$(shasum -a 256 "$current_source/$source_file" | awk 'NR == 1 { print $1 }')
  fi
  printf '%s  %s\n' "$source_sha256" "$source_file" >>"$artifact_dir/source-files-sha256.txt"
done <"$artifact_dir/source-files.txt"

stage=validate-pins
grep -h '^FROM ' \
  "$current_source/e2e/docker/Dockerfile.release" \
  "$current_source/e2e/docker/Dockerfile" >"$artifact_dir/base-images.txt"
if grep -Ev '^FROM [^[:space:]@]+@sha256:[0-9a-f]{64}( AS [[:alnum:]_.-]+)?$' "$artifact_dir/base-images.txt" >/dev/null; then
  echo "every E2E Dockerfile base image must be pinned by sha256 digest" >&2
  exit 1
fi
awk -v version="$E2E_GO_VERSION" '
  $1 == "FROM" && index($2, "golang:" version "-") == 1 &&
    $3 == "AS" && $4 == "build-env" { found = 1 }
  END { exit found ? 0 : 1 }
' "$artifact_dir/base-images.txt"
grep -Fqx "go $E2E_GO_VERSION" "$current_source/e2e/go.mod"
grep -Eq '^[[:space:]]+github\.com/strangelove-ventures/interchaintest/v8 v8\.8\.1$' \
  "$current_source/e2e/go.mod"

"$E2E_GO_BINARY" env -json GOVERSION GOROOT GOTOOLDIR GOBIN GOOS GOARCH >"$artifact_dir/go-env.json"
"$E2E_GO_BINARY" version >"$artifact_dir/go-version.txt"
"$E2E_GO_BINARY" tool compile -V=full >"$artifact_dir/compiler-version.txt"
"$DOCKER" version >"$artifact_dir/docker-version.txt"
"$DOCKER" buildx version >"$artifact_dir/buildx-version.txt"
docker_host_scheme=${DOCKER_HOST%%:*}
{
  printf 'scheme=%s\n' "$docker_host_scheme"
  printf 'endpoint_recorded=false\n'
} >"$artifact_dir/docker-host.txt"

stage=capture-functional-image-identity
functional_host_platform=$(
  "$DOCKER" image inspect "$E2E_FUNCTIONAL_CURRENT_IMAGE" --format '{{.Os}}/{{.Architecture}}'
)
old_functional_platform=$(
  "$DOCKER" image inspect "$E2E_FUNCTIONAL_OLD_IMAGE" --format '{{.Os}}/{{.Architecture}}'
)
case "$functional_host_platform" in
  linux/amd64 | linux/arm64) ;;
  *)
    echo "functional current image platform $functional_host_platform is not a release platform" >&2
    exit 2
    ;;
esac
if [ "$old_functional_platform" != "$functional_host_platform" ]; then
  echo "functional image platforms differ: current=$functional_host_platform old=$old_functional_platform" >&2
  exit 2
fi
functional_host_suffix=${functional_host_platform#linux/}

write_binary_sha256() {
  binary_path=$1
  checksum_path=$2
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$binary_path" >"$checksum_path"
  else
    shasum -a 256 "$binary_path" >"$checksum_path"
  fi
}

capture_functional_image() {
  functional_kind=$1
  functional_ref=$2
  functional_prefix="functional-$functional_kind-$functional_host_suffix"
  functional_copy_container="panacea-functional-copy-$functional_kind-$functional_host_suffix-$$"

  "$DOCKER" image inspect "$functional_ref" >"$artifact_dir/$functional_prefix-image-inspect.json"
  functional_image_id=$("$DOCKER" image inspect "$functional_ref" --format '{{.Id}}')
  if ! printf '%s\n' "$functional_image_id" | grep -Eq '^sha256:[0-9a-f]{64}$'; then
    echo "$functional_kind functional image ID is not sha256: $functional_image_id" >&2
    exit 2
  fi
  track_release_container "$functional_copy_container"
  "$DOCKER" create --platform "$functional_host_platform" --name "$functional_copy_container" "$functional_ref" \
    >"$artifact_dir/$functional_prefix-copy-container.txt"
  "$DOCKER" cp "$functional_copy_container:/usr/bin/panacead" \
    "$work_dir/bin-$functional_kind-$functional_host_suffix"
  "$DOCKER" rm "$functional_copy_container" >"$artifact_dir/$functional_prefix-copy-container-cleanup.txt"
  untrack_release_container "$functional_copy_container"
  write_binary_sha256 \
    "$work_dir/bin-$functional_kind-$functional_host_suffix" \
    "$artifact_dir/$functional_prefix-binary-sha256.txt"
}

capture_functional_image current "$E2E_FUNCTIONAL_CURRENT_IMAGE"
capture_functional_image v2.2.1 "$E2E_FUNCTIONAL_OLD_IMAGE"

stage=prepare-old-source
old_source="$work_dir/v2.2.1-source"
mkdir -p "$old_source"
git cat-file -e "$E2E_V221_COMMIT^{commit}"
git archive --format=tar --output="$work_dir/v2.2.1-source.tar" "$E2E_V221_COMMIT"
tar -xf "$work_dir/v2.2.1-source.tar" -C "$old_source"

stage=materialize-cold-caches
cold_mod_cache="$work_dir/go-mod"
cold_build_cache="$work_dir/go-build"
mkdir -p "$cold_mod_cache" "$cold_build_cache"
test -z "$(find "$cold_mod_cache" -mindepth 1 -print -quit)"
test -z "$(find "$cold_build_cache" -mindepth 1 -print -quit)"
{
  printf 'module_cache_initially_empty=true\n'
  printf 'build_cache_initially_empty=true\n'
  printf 'fresh_buildkit_builder=true\n'
} >"$artifact_dir/cold-cache-contract.txt"

(
  cd "$current_source"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" mod download -json all >"$artifact_dir/dependencies-current-download.jsonl"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" list -m -json all >"$artifact_dir/dependencies-current.jsonl"
)
(
  cd "$old_source"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" mod download -json all >"$artifact_dir/dependencies-v2.2.1-download.jsonl"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" list -m -json all >"$artifact_dir/dependencies-v2.2.1.jsonl"
)
(
  cd "$current_source/e2e"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" mod download -json all >"$artifact_dir/dependencies-e2e-download.jsonl"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" list -m -json all >"$artifact_dir/dependencies-e2e.jsonl"
)

stage=warm-offline-host-build
mkdir -p "$work_dir/bin"
(
  cd "$current_source"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOPROXY=off GOSUMDB=off \
    GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" CGO_ENABLED=0 \
    "$E2E_GO_BINARY" build -mod=mod -tags netgo -trimpath -o "$work_dir/bin/panacead-current" ./cmd/panacead
)
(
  cd "$old_source"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOPROXY=off GOSUMDB=off \
    GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" CGO_ENABLED=0 \
    "$E2E_GO_BINARY" build -mod=mod -tags netgo -trimpath -o "$work_dir/bin/panacead-v2.2.1" ./cmd/panacead
)
{
  printf 'GOPROXY=off\n'
  printf 'GOSUMDB=off\n'
  printf 'current_host_build=passed\n'
  printf 'v2.2.1_host_build=passed\n'
} >"$artifact_dir/warm-offline-build.txt"

stage=vendor-offline-contexts
current_vendor="$work_dir/current-vendor"
old_vendor="$work_dir/v2.2.1-vendor"
(
  cd "$current_source"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOPROXY=off GOSUMDB=off \
    GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" mod vendor -o "$current_vendor"
)
(
  cd "$old_source"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOPROXY=off GOSUMDB=off \
    GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" mod vendor -o "$old_vendor"
)

stage=create-fresh-builder
builder_created=1
"$DOCKER" buildx create --name "$builder_name" --driver docker-container >"$artifact_dir/builder-create.txt"
"$DOCKER" buildx inspect --builder "$builder_name" --bootstrap >"$artifact_dir/builder-initial.txt"
"$DOCKER" buildx du --builder "$builder_name" >"$artifact_dir/builder-cache-before-build.txt"
"$DOCKER" buildx du --builder "$builder_name" --format '{{.ID}}' >"$artifact_dir/builder-cache-record-ids-before-build.txt"
if [ -s "$artifact_dir/builder-cache-record-ids-before-build.txt" ]; then
  echo "fresh BuildKit builder unexpectedly contains build/base-image cache records" >&2
  exit 1
fi

cmt_version=$(awk '$1 == "github.com/cometbft/cometbft" { if ($2 == "=>") version = $4; else version = $2 } END { print version }' "$current_source/go.mod")
test -n "$cmt_version"

build_and_verify_image() {
  platform=$1
  suffix=$2
  kind=$3
  context_dir=$4
  vendor_dir=$5

  repository="panacea-e2e-release-$kind-$suffix"
  image_ref="$repository:$run_id"
  metadata="$artifact_dir/$kind-$suffix-build-metadata.json"
  build_log="$artifact_dir/$kind-$suffix-build.log"

  if [ "$kind" = current ]; then
    version=$E2E_CURRENT_BINARY_VERSION
    commit=$source_commit
    extra_arg="--build-arg=PANACEA_CMT_VERSION=$cmt_version"
    dockerfile="$current_source/e2e/docker/Dockerfile.release"
  else
    version=2.2.1
    commit=$E2E_V221_COMMIT
    extra_arg="--build-arg=PANACEA_TM_VERSION=$E2E_V221_TM_VERSION"
    dockerfile="$current_source/e2e/docker/Dockerfile"
  fi
  {
    printf 'platform=%s\n' "$platform"
    printf 'PANACEA_VERSION=%s\n' "$version"
    printf 'PANACEA_COMMIT=%s\n' "$commit"
    printf '%s\n' "${extra_arg#--build-arg=}"
    printf 'docker_network=none\n'
    printf 'no_cache=true\n'
  } >"$artifact_dir/$kind-$suffix-build-args.txt"

  created_images="$created_images $image_ref"
  "$DOCKER" buildx build \
    --builder "$builder_name" \
    --platform "$platform" \
    --network=none \
    --no-cache \
    --provenance=false \
    --build-context "panacea_vendor=$vendor_dir" \
    --build-context "panacea_e2e_tools=$current_source/scripts/e2e" \
    --file "$dockerfile" \
    --build-arg "PANACEA_VERSION=$version" \
    --build-arg "PANACEA_COMMIT=$commit" \
    "$extra_arg" \
    --tag "$image_ref" \
    --metadata-file "$metadata" \
    --load \
    "$context_dir" >"$build_log" 2>&1
  grep -Eq '"containerimage\.digest"[[:space:]]*:[[:space:]]*"sha256:[0-9a-f]{64}"' "$metadata"
  "$DOCKER" image inspect "$image_ref" >"$artifact_dir/$kind-$suffix-image-inspect.json"
  actual_platform=$("$DOCKER" image inspect "$image_ref" --format '{{.Os}}/{{.Architecture}}')
  test "$actual_platform" = "$platform"
  version_container="panacea-version-$kind-$suffix-$$"
  track_release_container "$version_container"
  "$DOCKER" run --rm --name "$version_container" --platform "$platform" \
    --entrypoint /usr/bin/panacead "$image_ref" version --long \
    >"$artifact_dir/$kind-$suffix-version.txt"
  untrack_release_container "$version_container"
  grep -Fq "version: $version" "$artifact_dir/$kind-$suffix-version.txt"
  grep -Fq "commit: $commit" "$artifact_dir/$kind-$suffix-version.txt"

  copy_container="panacea-copy-$kind-$suffix-$$"
  track_release_container "$copy_container"
  "$DOCKER" create --platform "$platform" --name "$copy_container" "$image_ref" >"$artifact_dir/$kind-$suffix-copy-container.txt"
  "$DOCKER" cp "$copy_container:/usr/bin/panacead" "$work_dir/bin/panacead-$kind-$suffix"
  "$DOCKER" rm "$copy_container" >"$artifact_dir/$kind-$suffix-copy-container-cleanup.txt"
  untrack_release_container "$copy_container"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$work_dir/bin/panacead-$kind-$suffix" >"$artifact_dir/$kind-$suffix-binary-sha256.txt"
  else
    shasum -a 256 "$work_dir/bin/panacead-$kind-$suffix" >"$artifact_dir/$kind-$suffix-binary-sha256.txt"
  fi

  # $1 and $node_home are intentionally expanded by the shell in the image.
  # shellcheck disable=SC2016
  smoke_command='node_home=$(mktemp -d); panacead init "$1" --chain-id "$1" --home "$node_home" >/dev/null; panacead genesis validate-genesis "$node_home/config/genesis.json" --home "$node_home"'
  if [ "$kind" = current ]; then
    # shellcheck disable=SC2016
    smoke_command="$smoke_command"'; panacead config view app --home "$node_home" >/dev/null'
  else
    # shellcheck disable=SC2016
    smoke_command="$smoke_command"'; panacead config chain-id --home "$node_home" >/dev/null'
  fi
  smoke_container="panacea-smoke-$kind-$suffix-$$"
  track_release_container "$smoke_container"
  "$DOCKER" run --rm --name "$smoke_container" --platform "$platform" --entrypoint /bin/sh "$image_ref" -ec \
    "$smoke_command" sh "release-$suffix" \
    >"$artifact_dir/$kind-$suffix-smoke.txt" 2>&1
  untrack_release_container "$smoke_container"

  image_digest=$(sed -n 's/.*"containerimage\.digest"[[:space:]]*:[[:space:]]*"\(sha256:[0-9a-f]*\)".*/\1/p' "$metadata" | awk 'NR == 1 { print $1 }')
  image_id=$("$DOCKER" image inspect "$image_ref" --format '{{.Id}}')
  binary_sha256=$(awk 'NR == 1 { print $1 }' "$artifact_dir/$kind-$suffix-binary-sha256.txt")
  printf '%s|%s|%s|%s|%s|%s|%s|%s\n' \
    "$platform" "$kind" "$image_ref" "$image_digest" "$image_id" "$binary_sha256" "$version" "$commit" \
    >>"$artifact_dir/image-index.txt"
}

stage=build-multiarch-images
: >"$artifact_dir/image-index.txt"
for platform in linux/amd64 linux/arm64; do
  suffix=${platform#linux/}
  build_and_verify_image "$platform" "$suffix" current "$current_source" "$current_vendor"
  build_and_verify_image "$platform" "$suffix" v2.2.1 "$old_source" "$old_vendor"
done
"$DOCKER" buildx du --builder "$builder_name" >"$artifact_dir/builder-cache-after-build.txt"

stage=write-host-image-identity
release_current_ref="panacea-e2e-release-current-$functional_host_suffix:$run_id"
release_old_ref="panacea-e2e-release-v2.2.1-$functional_host_suffix:$run_id"
functional_current_id=$("$DOCKER" image inspect "$E2E_FUNCTIONAL_CURRENT_IMAGE" --format '{{.Id}}')
functional_old_id=$("$DOCKER" image inspect "$E2E_FUNCTIONAL_OLD_IMAGE" --format '{{.Id}}')
release_current_id=$("$DOCKER" image inspect "$release_current_ref" --format '{{.Id}}')
release_old_id=$("$DOCKER" image inspect "$release_old_ref" --format '{{.Id}}')
functional_current_sha256=$(awk 'NR == 1 { print $1 }' \
  "$artifact_dir/functional-current-$functional_host_suffix-binary-sha256.txt")
functional_old_sha256=$(awk 'NR == 1 { print $1 }' \
  "$artifact_dir/functional-v2.2.1-$functional_host_suffix-binary-sha256.txt")
release_current_sha256=$(awk 'NR == 1 { print $1 }' \
  "$artifact_dir/current-$functional_host_suffix-binary-sha256.txt")
release_old_sha256=$(awk 'NR == 1 { print $1 }' \
  "$artifact_dir/v2.2.1-$functional_host_suffix-binary-sha256.txt")
if [ "$functional_current_sha256" != "$release_current_sha256" ]; then
  echo "functional current panacead checksum differs from the host-platform release build" >&2
  exit 1
fi
if [ "$functional_old_sha256" != "$release_old_sha256" ]; then
  echo "functional v2.2.1 panacead checksum differs from the host-platform release build" >&2
  exit 1
fi
{
  printf '{\n'
  printf '  "schema_version": "1",\n'
  printf '  "host_platform": "%s",\n' "$functional_host_platform"
  printf '  "images": [\n'
  printf '    {"kind":"current","functional_image_ref":"%s","functional_image_id":"%s","functional_binary_sha256":"%s","release_image_ref":"%s","release_image_id":"%s","release_binary_sha256":"%s"},\n' \
    "$E2E_FUNCTIONAL_CURRENT_IMAGE" "$functional_current_id" "$functional_current_sha256" \
    "$release_current_ref" "$release_current_id" "$release_current_sha256"
  printf '    {"kind":"v2.2.1","functional_image_ref":"%s","functional_image_id":"%s","functional_binary_sha256":"%s","release_image_ref":"%s","release_image_id":"%s","release_binary_sha256":"%s"}\n' \
    "$E2E_FUNCTIONAL_OLD_IMAGE" "$functional_old_id" "$functional_old_sha256" \
    "$release_old_ref" "$release_old_id" "$release_old_sha256"
  printf '  ]\n'
  printf '}\n'
} >"$artifact_dir/host-image-identity.json"

stage=warm-offline-buildkit-build
builder_container=$(
  "$DOCKER" ps \
    --filter "label=com.docker.buildx.builder=$builder_name" \
    --format '{{.ID}}'
)
# Intentional field splitting verifies that the label selected one container.
# shellcheck disable=SC2086
set -- $builder_container
if [ "$#" -ne 1 ]; then
  echo "expected exactly one BuildKit container for $builder_name, got: $builder_container" >&2
  exit 1
fi
builder_container=$1
# Docker's Go template is intentionally single-quoted for the shell.
# shellcheck disable=SC2016
"$DOCKER" inspect --format '{{range $name, $_ := .NetworkSettings.Networks}}{{$name}}{{println}}{{end}}' \
  "$builder_container" >"$artifact_dir/builder-networks-before-offline.txt"
if [ ! -s "$artifact_dir/builder-networks-before-offline.txt" ]; then
  echo "BuildKit container has no network to disconnect for the offline proof" >&2
  exit 1
fi
while IFS= read -r builder_network; do
  [ -n "$builder_network" ] || continue
  "$DOCKER" network disconnect "$builder_network" "$builder_container"
done <"$artifact_dir/builder-networks-before-offline.txt"
# Docker's Go template is intentionally single-quoted for the shell.
# shellcheck disable=SC2016
"$DOCKER" inspect --format '{{range $name, $_ := .NetworkSettings.Networks}}{{$name}}{{println}}{{end}}' \
  "$builder_container" >"$artifact_dir/builder-networks-after-offline.txt"
if [ -s "$artifact_dir/builder-networks-after-offline.txt" ]; then
  echo "BuildKit container still has a network after offline isolation" >&2
  exit 1
fi

warm_offline_buildkit_image() {
  platform=$1
  suffix=$2
  kind=$3
  context_dir=$4
  vendor_dir=$5

  repository="panacea-e2e-release-warm-$kind-$suffix"
  image_ref="$repository:$run_id"
  metadata="$artifact_dir/warm-offline-$kind-$suffix-build-metadata.json"
  build_log="$artifact_dir/warm-offline-$kind-$suffix-build.log"
  if [ "$kind" = current ]; then
    version=$E2E_CURRENT_BINARY_VERSION
    commit=$source_commit
    extra_arg="--build-arg=PANACEA_CMT_VERSION=$cmt_version"
    dockerfile="$current_source/e2e/docker/Dockerfile.release"
  else
    version=2.2.1
    commit=$E2E_V221_COMMIT
    extra_arg="--build-arg=PANACEA_TM_VERSION=$E2E_V221_TM_VERSION"
    dockerfile="$current_source/e2e/docker/Dockerfile"
  fi

  created_images="$created_images $image_ref"
  "$DOCKER" buildx build \
    --builder "$builder_name" \
    --platform "$platform" \
    --network=none \
    --provenance=false \
    --build-context "panacea_vendor=$vendor_dir" \
    --build-context "panacea_e2e_tools=$current_source/scripts/e2e" \
    --file "$dockerfile" \
    --build-arg "PANACEA_VERSION=$version" \
    --build-arg "PANACEA_COMMIT=$commit" \
    "$extra_arg" \
    --tag "$image_ref" \
    --metadata-file "$metadata" \
    --load \
    "$context_dir" >"$build_log" 2>&1
  grep -Eq '"containerimage\.digest"[[:space:]]*:[[:space:]]*"sha256:[0-9a-f]{64}"' "$metadata"
  "$DOCKER" image inspect "$image_ref" >"$artifact_dir/warm-offline-$kind-$suffix-image-inspect.json"
  warm_version_container="panacea-warm-version-$kind-$suffix-$$"
  track_release_container "$warm_version_container"
  "$DOCKER" run --rm --name "$warm_version_container" --platform "$platform" \
    --entrypoint /usr/bin/panacead "$image_ref" version --long \
    >"$artifact_dir/warm-offline-$kind-$suffix-version.txt"
  untrack_release_container "$warm_version_container"
  grep -Fq "version: $version" "$artifact_dir/warm-offline-$kind-$suffix-version.txt"
  grep -Fq "commit: $commit" "$artifact_dir/warm-offline-$kind-$suffix-version.txt"
}

for platform in linux/amd64 linux/arm64; do
  suffix=${platform#linux/}
  warm_offline_buildkit_image "$platform" "$suffix" current "$current_source" "$current_vendor"
  warm_offline_buildkit_image "$platform" "$suffix" v2.2.1 "$old_source" "$old_vendor"
done
printf '%s\n' \
  'builder_networks=none' \
  'registry_access=unavailable' \
  'cache_mode=warm' \
  'platforms=linux/amd64,linux/arm64' \
  'images=current,v2.2.1' \
  >"$artifact_dir/warm-offline-buildkit-contract.txt"

stage=require-multiarch-upgrade-gate
if [ "${PANACEA_E2E_RELEASE_MULTIARCH_UPGRADE:-}" != "1" ]; then
  echo "PANACEA_E2E_RELEASE_MULTIARCH_UPGRADE=1 is required; image-only evidence is not an upgrade compatibility pass" \
    >"$artifact_dir/multiarch-upgrade-not-run.txt"
  exit 2
fi

stage=compile-upgrade-test-binary
release_test_binary="$work_dir/panacea-e2e.test"
(
  cd "$current_source/e2e"
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOPROXY=off GOSUMDB=off \
  GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" test -c -o "$release_test_binary" .
)

stage=multiarch-upgrade
while IFS='|' read -r platform image_kind current_image_ref image_digest image_id binary_sha256 image_version image_commit; do
  [ "$image_kind" = current ] || continue
  suffix=${platform#linux/}
  current_repository=${current_image_ref%%:*}
  old_repository="panacea-e2e-release-v2.2.1-$suffix"
  upgrade_root="$artifact_dir/upgrade-$suffix"
  mkdir -p "$upgrade_root"
  (
    cd "$current_source/e2e"
    PANACEA_E2E_UPGRADE=1 \
    PANACEA_E2E_ROOT="$upgrade_root" \
    PANACEA_E2E_IMAGE_REPOSITORY="$current_repository" \
    PANACEA_E2E_V221_IMAGE_REPOSITORY="$old_repository" \
    PANACEA_E2E_IMAGE_VERSION="$run_id" \
    PANACEA_E2E_V221_IMAGE_VERSION="$run_id" \
    PANACEA_E2E_CURRENT_BINARY_VERSION="$E2E_CURRENT_BINARY_VERSION" \
    PANACEA_E2E_CURRENT_COMMIT="$source_commit" \
    GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOPROXY=off GOSUMDB=off \
    GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
      "$release_test_binary" -test.timeout "$E2E_RELEASE_UPGRADE_TIMEOUT" \
        -test.count=1 -test.v -test.run '^TestV221ToCurrentMultiValidatorUpgrade$' \
        >"$artifact_dir/upgrade-$suffix.log" 2>&1
  )
  printf 'platform=%s\nresult=passed\n' "$platform" >"$artifact_dir/upgrade-$suffix-result.txt"
done <"$artifact_dir/image-index.txt"

stage=revalidate-clean-source
git rev-parse HEAD >"$artifact_dir/source-commit-final.txt"
git status --porcelain=v1 --untracked-files=all >"$artifact_dir/source-status-final.txt"
git diff --binary HEAD -- >"$artifact_dir/source-diff-final.patch"
if [ "$(tr -d '\n' <"$artifact_dir/source-commit-final.txt")" != "$source_commit" ] || \
  [ -s "$artifact_dir/source-status-final.txt" ] || [ -s "$artifact_dir/source-diff-final.patch" ]; then
  echo "release source changed while release-hardening was running" >&2
  exit 2
fi

stage=write-manifest
{
  printf '{\n'
  printf '  "schema_version": "4",\n'
  printf '  "run_id": "%s",\n' "$run_id"
  printf '  "source_commit": "%s",\n' "$source_commit"
  printf '  "source_clean": true,\n'
  printf '  "cold_go_caches": true,\n'
  printf '  "fresh_buildkit_builder": true,\n'
  printf '  "warm_offline_host_build": true,\n'
  printf '  "warm_offline_buildkit_build": true,\n'
  printf '  "docker_build_network": "none",\n'
  printf '  "platforms": ["linux/amd64", "linux/arm64"],\n'
  printf '  "version_and_smoke": true,\n'
  printf '  "multiarch_upgrade_compatibility": true,\n'
  printf '  "image_index": "image-index.txt",\n'
  printf '  "host_platform": "%s",\n' "$functional_host_platform"
  printf '  "host_image_identity": "host-image-identity.json",\n'
  printf '  "images": [\n'
  image_number=0
  while IFS='|' read -r image_platform image_kind image_ref image_digest image_id binary_sha256 image_version image_commit; do
    if [ "$image_number" -gt 0 ]; then
      printf ',\n'
    fi
    printf '    {"kind":"%s","platform":"%s","image_ref":"%s","image_digest":"%s","image_id":"%s","binary_sha256":"%s","version":"%s","source_commit":"%s"}' \
      "$image_kind" "$image_platform" "$image_ref" "$image_digest" "$image_id" "$binary_sha256" "$image_version" "$image_commit"
    image_number=$((image_number + 1))
  done <"$artifact_dir/image-index.txt"
  printf '\n  ]\n'
  printf '}\n'
} >"$artifact_dir/release-hardening-manifest.json"

stage=validate-manifest
(
  cd "$current_source/e2e"
  PANACEA_E2E_RELEASE_MANIFEST="$artifact_dir/release-hardening-manifest.json" \
  GOTOOLCHAIN="$E2E_GOTOOLCHAIN" GOWORK=off GOPROXY=off GOSUMDB=off \
  GOMODCACHE="$cold_mod_cache" GOCACHE="$cold_build_cache" \
    "$E2E_GO_BINARY" test -count=1 -run '^TestValidateReleaseHardeningArtifact$' ./internal/harness \
    >"$artifact_dir/manifest-validation.log" 2>&1
)
stage=complete
