#!/bin/sh

# Artifact-first, process-bounded wrapper for the complete P0/P1 release gate.
# The nested release-build runner has its own watchdog, but the functional
# prerequisites execute before it. This wrapper therefore owns the standalone
# deadline and preserves a gate-level failure even when a child is interrupted.

set -eu

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P)
repo_root=$(CDPATH='' cd -- "$script_dir/../.." && pwd -P)
cd "$repo_root"

aggregate_started_epoch=$(date -u +%s)
E2E_RELEASE_HARDENING_RUN_ID=${E2E_RELEASE_HARDENING_RUN_ID:-"p0p1-$(date -u +%Y%m%d%H%M%S)-$$"}
if [ -n "${E2E_RELEASE_AGGREGATE_BASE_ROOT:-}" ]; then
  aggregate_base_root=$E2E_RELEASE_AGGREGATE_BASE_ROOT
  while [ "$aggregate_base_root" != "/" ] &&
    [ "${aggregate_base_root%/}" != "$aggregate_base_root" ]; do
    aggregate_base_root=${aggregate_base_root%/}
  done
  E2E_ROOT="$aggregate_base_root/$E2E_RELEASE_HARDENING_RUN_ID"
  E2E_GOCACHE="$E2E_ROOT/go-build"
  E2E_GOMODCACHE="$E2E_ROOT/go-mod"
else
  E2E_ROOT=${E2E_ROOT:-"$repo_root/.local/e2e/$E2E_RELEASE_HARDENING_RUN_ID"}
  E2E_GOCACHE=${E2E_GOCACHE:-"$E2E_ROOT/go-build"}
  E2E_GOMODCACHE=${E2E_GOMODCACHE:-"$E2E_ROOT/go-mod"}
fi
E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS=${E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS:-43200}
E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS=${E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS:-120}
E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS=${E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS:-10}
E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS=${E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS:-5}
E2E_RUNNER=${E2E_RUNNER:-"$repo_root/scripts/e2e/run.sh"}
export E2E_RUNNER
DOCKER=${DOCKER:-docker}

bootstrap_parent="$repo_root/.local/e2e"
bootstrap_root="$bootstrap_parent/p0p1-bootstrap-$(date -u +%Y%m%d%H%M%S)-$$"
control_dir="$bootstrap_parent/.aggregate-control-$(date -u +%Y%m%d%H%M%S)-$$"
artifact_root=$bootstrap_root
release_dir="$artifact_root/release"
mkdir -p "$release_dir" "$control_dir"
chmod 700 "$artifact_root" "$release_dir" "$control_dir"

stage=initializing
source_commit=unknown
watchdog_pid=
watchdog_pid_file="$control_dir/watchdog-pid.txt"
watchdog_timer_pid_file="$control_dir/watchdog-timer-pid.txt"
artifact_root_file="$control_dir/artifact-root.txt"
artifact_root_lock_dir="$control_dir/artifact-root.lock"
cleanup_requested_file="$control_dir/cleanup-requested.txt"
watchdog_trace="$release_dir/aggregate-watchdog-trace.txt"
cleanup_failed=0
aggregate_work_timeout_seconds=unvalidated
aggregate_work_deadline_epoch=unvalidated
printf '%s\n' "$artifact_root" >"$artifact_root_file"

json_escape() {
  awk 'BEGIN { ORS="" }
    {
      if (NR > 1) printf "\\n"
      for (i = 1; i <= length($0); i++) {
        c = substr($0, i, 1)
        if (c == "\\") printf "\\\\"
        else if (c == "\"") printf "\\\""
        else if (c == "\t") printf "\\t"
        else if (c == "\r") printf "\\r"
        else printf "%s", c
      }
    }'
}

write_gate_failure_at() {
  failure_release_dir=$1
  failure_stage=$2
  failure_error=$3
  mkdir -p "$failure_release_dir"
  escaped_stage=$(printf '%s' "$failure_stage" | json_escape)
  escaped_error=$(printf '%s' "$failure_error" | json_escape)
  recorded_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
  {
    printf '{\n'
    printf '  "schema_version": "2",\n'
    printf '  "recorded_at": "%s",\n' "$recorded_at"
    printf '  "stage": "%s",\n' "$escaped_stage"
    printf '  "error": "%s"\n' "$escaped_error"
    printf '}\n'
  } >"$failure_release_dir/gate-failure.json.tmp"
  mv "$failure_release_dir/gate-failure.json.tmp" "$failure_release_dir/gate-failure.json"
}

write_gate_failure() {
  write_gate_failure_at "$release_dir" "$1" "$2"
}

invalidate_gate_manifest_at() {
  invalidation_release_dir=$1
  rm -f "$invalidation_release_dir/gate-manifest.json" \
    "$invalidation_release_dir/gate-manifest.json.tmp"
}

write_status() {
  status_result=$1
  status_code=$2
  {
    printf 'result=%s\n' "$status_result"
    printf 'exit_code=%s\n' "$status_code"
    printf 'stage=%s\n' "$stage"
    printf 'run_id=%s\n' "$E2E_RELEASE_HARDENING_RUN_ID"
    printf 'source_commit=%s\n' "$source_commit"
    printf 'artifact_root=%s\n' "$artifact_root"
    printf 'total_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS"
    printf 'cleanup_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS"
    printf 'force_exit_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS"
    printf 'work_timeout_seconds=%s\n' "$aggregate_work_timeout_seconds"
    printf 'work_deadline_epoch=%s\n' "$aggregate_work_deadline_epoch"
    printf 'child_exit_margin_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS"
  } >"$release_dir/aggregate-status.txt.tmp"
  mv "$release_dir/aggregate-status.txt.tmp" "$release_dir/aggregate-status.txt"
}

write_forced_status_at() {
  forced_release_dir=$1
  forced_stage=$2
  forced_code=$3
  forced_source_commit=unknown
  if [ -s "$forced_release_dir/source-commit.txt" ]; then
    IFS= read -r forced_source_commit <"$forced_release_dir/source-commit.txt" || forced_source_commit=unknown
  fi
  {
    printf 'result=failed\n'
    printf 'exit_code=%s\n' "$forced_code"
    printf 'stage=%s\n' "$forced_stage"
    printf 'run_id=%s\n' "$E2E_RELEASE_HARDENING_RUN_ID"
    printf 'source_commit=%s\n' "$forced_source_commit"
    printf 'artifact_root=%s\n' "${forced_release_dir%/release}"
    printf 'total_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS"
    printf 'cleanup_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS"
    printf 'force_exit_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS"
    printf 'work_timeout_seconds=%s\n' "$aggregate_work_timeout_seconds"
    printf 'work_deadline_epoch=%s\n' "$aggregate_work_deadline_epoch"
    printf 'cleanup_result=timed-out\n'
  } >"$forced_release_dir/aggregate-status.txt.tmp"
  mv "$forced_release_dir/aggregate-status.txt.tmp" "$forced_release_dir/aggregate-status.txt"
}

publish_artifact_root() {
  published_artifact_root=$1
  printf '%s\n' "$published_artifact_root" >"$artifact_root_file.tmp"
  mv "$artifact_root_file.tmp" "$artifact_root_file"
}

set_artifact_root() {
  new_artifact_root=$1
  artifact_root=$new_artifact_root
  release_dir="$artifact_root/release"
  watchdog_trace="$release_dir/aggregate-watchdog-trace.txt"
  publish_artifact_root "$artifact_root"
}

lock_artifact_root() {
  while ! mkdir "$artifact_root_lock_dir" 2>/dev/null; do
    sleep 1
  done
}

unlock_artifact_root() {
  rmdir "$artifact_root_lock_dir" 2>/dev/null || true
}

aggregate_timeout_is_active() (
  active_artifact_root=$(awk 'NR == 1 { print; exit }' "$artifact_root_file" 2>/dev/null || true)
  if [ -z "$active_artifact_root" ]; then
    active_artifact_root=$artifact_root
  fi
  [ -s "$active_artifact_root/release/overall-timeout.txt" ] ||
    [ -s "$bootstrap_root/release/overall-timeout.txt" ]
)

aggregate_parent_pid() {
  inspected_pid=$1
  case "$inspected_pid" in
    '' | *[!0-9]*) return 1 ;;
  esac
  ps -o ppid= -p "$inspected_pid" 2>/dev/null | awk 'NR == 1 { print $1 }'
}

aggregate_is_descendant() {
  candidate_pid=$1
  ancestor_pid=$2
  ancestry_hops=0
  case "$candidate_pid:$ancestor_pid" in
    *[!0-9:]*) return 1 ;;
  esac
  while [ "$ancestry_hops" -lt 1024 ]; do
    candidate_parent=$(aggregate_parent_pid "$candidate_pid") || return 1
    if [ "$candidate_parent" = "$ancestor_pid" ]; then
      return 0
    fi
    case "$candidate_parent" in
      '' | *[!0-9]* | 0 | 1) return 1 ;;
    esac
    [ "$candidate_parent" != "$candidate_pid" ] || return 1
    candidate_pid=$candidate_parent
    ancestry_hops=$((ancestry_hops + 1))
  done
  return 1
}

aggregate_child_pids() (
  inspected_parent=$1
  aggregate_process_table=$(ps -eo pid=,ppid= 2>/dev/null) || exit $?
  printf '%s\n' "$aggregate_process_table" |
    awk -v inspected_parent="$inspected_parent" '$2 == inspected_parent { print $1 }'
)

aggregate_signal_tree() (
  tree_pid=$1
  runner_pid=$2
  excluded_pid=$3
  signal_name=$4
  case "$signal_name" in TERM | KILL) ;; *) return 1 ;; esac
  [ "$tree_pid" != "$excluded_pid" ] || return 0
  aggregate_child_pids "$tree_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    aggregate_signal_tree "$child_pid" "$runner_pid" "$excluded_pid" "$signal_name"
  done
  if [ "$signal_name" = TERM ] &&
    [ "$(aggregate_parent_pid "$tree_pid" 2>/dev/null || true)" = "$runner_pid" ] &&
    [ -n "$(aggregate_child_pids "$tree_pid" | awk 'NR == 1 { print $1 }')" ]; then
    # Keep the direct wrapper owned by the aggregate runner until its branch
    # has drained. Deeper workers still receive TERM so cooperative cleanup
    # traps run, while a TERM-resistant worker cannot be orphaned before the
    # bounded hard-kill phase.
    return 0
  fi
  if aggregate_is_descendant "$tree_pid" "$runner_pid"; then
    kill "-$signal_name" "$tree_pid" 2>/dev/null || true
  fi
)

aggregate_signal_runner_children() {
  runner_pid=$1
  excluded_pid=$2
  signal_name=$3
  aggregate_child_pids "$runner_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    aggregate_signal_tree "$child_pid" "$runner_pid" "$excluded_pid" "$signal_name"
  done
}

aggregate_process_is_live_non_zombie() {
  inspected_pid=$1
  process_state=$(ps -o stat= -p "$inspected_pid" 2>/dev/null | awk 'NR == 1 { print $1 }')
  case "$process_state" in
    '' | Z*) return 1 ;;
  esac
  return 0
}

aggregate_tree_has_live_children() {
  inspected_parent=$1
  for child_pid in $(aggregate_child_pids "$inspected_parent" || true); do
    if aggregate_process_is_live_non_zombie "$child_pid"; then
      return 0
    fi
  done
  return 1
}

# Freeze each parent before terminating its children. A cooperative shell can
# therefore neither respawn a short-lived child nor orphan a TERM-resistant
# descendant. Once all children are gone (zombies are already dead), deliver
# TERM and resume the parent so its cleanup trap can run.
aggregate_graceful_terminate_tree() (
  tree_pid=$1
  runner_pid=$2
  excluded_pid=$3
  [ "$tree_pid" != "$excluded_pid" ] || return 0
  aggregate_is_descendant "$tree_pid" "$runner_pid" || return 0
  aggregate_process_is_live_non_zombie "$tree_pid" || return 0
  kill -STOP "$tree_pid" 2>/dev/null || return 0
  aggregate_child_pids "$tree_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    aggregate_graceful_terminate_tree "$child_pid" "$runner_pid" "$excluded_pid"
  done
  if ! aggregate_tree_has_live_children "$tree_pid" &&
    aggregate_is_descendant "$tree_pid" "$runner_pid"; then
    kill -TERM "$tree_pid" 2>/dev/null || true
    kill -CONT "$tree_pid" 2>/dev/null || true
  fi
)

aggregate_graceful_terminate_runner_children() {
  runner_pid=$1
  excluded_pid=$2
  aggregate_child_pids "$runner_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    aggregate_graceful_terminate_tree "$child_pid" "$runner_pid" "$excluded_pid"
  done
}

aggregate_stop_tree() (
  tree_pid=$1
  runner_pid=$2
  excluded_pid=$3
  [ "$tree_pid" != "$excluded_pid" ] || return 0
  if aggregate_is_descendant "$tree_pid" "$runner_pid"; then
    kill -STOP "$tree_pid" 2>/dev/null || true
  else
    return 0
  fi
  aggregate_child_pids "$tree_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    aggregate_stop_tree "$child_pid" "$runner_pid" "$excluded_pid"
  done
)

aggregate_stop_runner_children() {
  runner_pid=$1
  excluded_pid=$2
  aggregate_child_pids "$runner_pid" | while IFS= read -r child_pid; do
    [ -n "$child_pid" ] || continue
    aggregate_stop_tree "$child_pid" "$runner_pid" "$excluded_pid"
  done
}

aggregate_runner_has_children() (
  inspected_runner=$1
  excluded_pid=$2
  for child_pid in $(aggregate_child_pids "$inspected_runner" || true); do
    if [ "$child_pid" != "$excluded_pid" ] &&
      aggregate_is_descendant "$child_pid" "$inspected_runner" &&
      aggregate_process_is_live_non_zombie "$child_pid"; then
      return 0
    fi
  done
  return 1
)

aggregate_watchdog_is_owned() {
  owned_watchdog_pid=$1
  owned_runner_pid=$2
  [ "$(aggregate_parent_pid "$owned_watchdog_pid" 2>/dev/null || true)" = "$owned_runner_pid" ]
}

aggregate_watchdog_sleep() {
  sleep_seconds=$1
  sleep "$sleep_seconds" &
  timer_pid=$!
  printf '%s\n' "$timer_pid" >"$watchdog_timer_pid_file"
  if wait "$timer_pid"; then sleep_result=0; else sleep_result=$?; fi
  rm -f "$watchdog_timer_pid_file"
  return "$sleep_result"
}

start_aggregate_watchdog() {
  runner_pid=$$
  watchdog_now_epoch=$(date -u +%s)
  watchdog_work_seconds=$((aggregate_work_deadline_epoch - watchdog_now_epoch))
  if [ "$watchdog_work_seconds" -lt 0 ]; then
    watchdog_work_seconds=0
  fi
  (
    trap - EXIT HUP INT TERM
    if ! aggregate_watchdog_sleep "$watchdog_work_seconds"; then
      [ -s "$cleanup_requested_file" ] || exit 0
      lock_artifact_root
      trap 'unlock_artifact_root' EXIT
      cleanup_artifact_root=$(awk 'NR == 1 { print; exit }' "$artifact_root_file" 2>/dev/null || true)
      cleanup_release_dir="$cleanup_artifact_root/release"
      mkdir -p "$cleanup_release_dir"
      cleanup_trace="$cleanup_release_dir/aggregate-watchdog-trace.txt"
      printf 'event=early-cleanup-request runner_pid=%s cleanup_timeout_seconds=%s\n' \
        "$runner_pid" "$E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS" >>"$cleanup_trace"
      if ! aggregate_watchdog_sleep "$E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS"; then
        exit 0
      fi
      aggregate_watchdog_is_owned "$(awk 'NR == 1 { print $1 }' "$watchdog_pid_file" 2>/dev/null || true)" "$runner_pid" || exit 0
      {
        printf 'result=timed-out\n'
        printf 'stage=aggregate-cleanup-timeout\n'
        printf 'cleanup_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS"
      } >"$cleanup_release_dir/cleanup-timeout.txt"
      invalidate_gate_manifest_at "$cleanup_release_dir"
      write_gate_failure_at "$cleanup_release_dir" aggregate-cleanup-timeout \
        'P0/P1 aggregate cleanup exceeded its bounded deadline'
      write_forced_status_at "$cleanup_release_dir" aggregate-cleanup-timeout 1
      printf 'event=early-cleanup-timeout\n' >>"$cleanup_trace"
      cleanup_watchdog_identity=$(awk 'NR == 1 { print $1 }' "$watchdog_pid_file" 2>/dev/null || true)
      aggregate_watchdog_is_owned "$cleanup_watchdog_identity" "$runner_pid" || exit 0
      kill -STOP "$runner_pid" 2>/dev/null || true
      aggregate_stop_runner_children "$runner_pid" "$cleanup_watchdog_identity"
      aggregate_signal_runner_children "$runner_pid" "$cleanup_watchdog_identity" KILL
      invalidate_gate_manifest_at "$cleanup_release_dir"
      if aggregate_watchdog_is_owned "$cleanup_watchdog_identity" "$runner_pid"; then
        printf 'event=runner-cont-after-child-kill\n' >>"$cleanup_trace"
        kill -CONT "$runner_pid" 2>/dev/null || true
      fi
      if ! aggregate_watchdog_sleep "$E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS"; then
        exit 0
      fi
      if aggregate_watchdog_is_owned "$cleanup_watchdog_identity" "$runner_pid"; then
        kill -STOP "$runner_pid" 2>/dev/null || true
        aggregate_stop_runner_children "$runner_pid" "$cleanup_watchdog_identity"
        aggregate_signal_runner_children "$runner_pid" "$cleanup_watchdog_identity" KILL
        kill -KILL "$runner_pid" 2>/dev/null || true
      fi
      rm -f "$watchdog_timer_pid_file" "$watchdog_pid_file" "$artifact_root_file" \
        "$cleanup_requested_file"
      unlock_artifact_root
      rmdir "$control_dir" 2>/dev/null || true
      exit 0
    fi
    lock_artifact_root
    trap 'unlock_artifact_root' EXIT
    timeout_artifact_root=$(awk 'NR == 1 { print; exit }' "$artifact_root_file" 2>/dev/null || true)
    timeout_release_dir="$timeout_artifact_root/release"
    mkdir -p "$timeout_release_dir"
    watchdog_timeout_marker="$timeout_release_dir/overall-timeout.txt"
    watchdog_trace="$timeout_release_dir/aggregate-watchdog-trace.txt"
    {
      printf 'total_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS"
      printf 'work_timeout_seconds=%s\n' "$aggregate_work_timeout_seconds"
    } >"$watchdog_timeout_marker"
    watchdog_identity=$(awk 'NR == 1 { print $1 }' "$watchdog_pid_file" 2>/dev/null || true)
    if ! aggregate_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
      invalidate_gate_manifest_at "$timeout_release_dir"
      write_gate_failure_at "$timeout_release_dir" watchdog-identity-invalid \
        'aggregate watchdog ownership could not be proven'
      exit 1
    fi
    remaining=$E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS
    runner_signalled=0
    printf 'event=deadline watchdog_pid=%s runner_pid=%s work_timeout_seconds=%s\n' \
      "$watchdog_identity" "$runner_pid" "$aggregate_work_timeout_seconds" >"$watchdog_trace"
    while [ "$remaining" -gt 0 ]; do
      aggregate_graceful_terminate_runner_children "$runner_pid" "$watchdog_identity"
      printf 'event=term-sweep remaining=%s\n' "$remaining" >>"$watchdog_trace"
      if ! aggregate_runner_has_children "$runner_pid" "$watchdog_identity"; then
        if aggregate_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
          kill -TERM "$runner_pid" 2>/dev/null || true
        fi
        runner_signalled=1
        printf 'event=runner-term\n' >>"$watchdog_trace"
        break
      fi
      if ! aggregate_watchdog_sleep 1; then exit 0; fi
      remaining=$((remaining - 1))
    done
    if [ "$runner_signalled" -eq 1 ] && [ "$remaining" -gt 0 ]; then
      if ! aggregate_watchdog_sleep "$remaining"; then exit 0; fi
    fi
    aggregate_watchdog_is_owned "$watchdog_identity" "$runner_pid" || exit 0
    {
      printf 'result=timed-out\n'
      printf 'timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS"
      printf 'cleanup_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS"
    } >"$timeout_release_dir/cleanup-timeout.txt"
    invalidate_gate_manifest_at "$timeout_release_dir"
    write_gate_failure_at "$timeout_release_dir" overall-timeout \
      'P0/P1 aggregate cleanup exceeded its bounded deadline'
    write_forced_status_at "$timeout_release_dir" overall-timeout 124
    printf 'event=cleanup-timeout\n' >>"$watchdog_trace"
    if aggregate_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
      kill -STOP "$runner_pid" 2>/dev/null || true
    fi
    aggregate_stop_runner_children "$runner_pid" "$watchdog_identity"
    aggregate_signal_runner_children "$runner_pid" "$watchdog_identity" KILL
    invalidate_gate_manifest_at "$timeout_release_dir"
    if aggregate_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
      # The overall-timeout trap must be allowed to convert the forced stop
      # into the stable aggregate exit code 124. A second frozen tree sweep
      # below still prevents cleanup descendants from escaping the hard cap.
      kill -CONT "$runner_pid" 2>/dev/null || true
    fi
    if ! aggregate_watchdog_sleep "$E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS"; then exit 0; fi
    if aggregate_watchdog_is_owned "$watchdog_identity" "$runner_pid"; then
      kill -STOP "$runner_pid" 2>/dev/null || true
      aggregate_stop_runner_children "$runner_pid" "$watchdog_identity"
      aggregate_signal_runner_children "$runner_pid" "$watchdog_identity" KILL
      kill -KILL "$runner_pid" 2>/dev/null || true
    fi
    rm -f "$watchdog_timer_pid_file" "$watchdog_pid_file" "$artifact_root_file" \
      "$cleanup_requested_file"
    unlock_artifact_root
    rmdir "$control_dir" 2>/dev/null || true
  ) &
  watchdog_pid=$!
  printf '%s\n' "$watchdog_pid" >"$watchdog_pid_file"
  while [ ! -s "$watchdog_timer_pid_file" ]; do
    kill -0 "$watchdog_pid" 2>/dev/null || break
    :
  done
}

cleanup_run_labels() {
  aggregate_cleanup_log="$release_dir/aggregate-cleanup.txt"
  : >"$aggregate_cleanup_log"
  found_run=0
  for run_dir in "$artifact_root"/run-* "$artifact_root"/release-*/upgrade-*/run-*; do
    [ -d "$run_dir" ] || continue
    [ ! -L "$run_dir" ] || continue
    run_id=${run_dir##*/}
    case "$run_id" in run-????????????) ;; *) continue ;; esac
    if printf '%s\n' "$run_id" | grep -Eqv '^run-[0-9a-f]{12}$'; then
      continue
    fi
    found_run=1
    printf 'run_path=%s run_id=%s\n' "$run_dir" "$run_id" >>"$aggregate_cleanup_log"
    if ! command -v "$DOCKER" >/dev/null 2>&1; then
      printf 'docker executable unavailable while cleaning %s\n' "$run_id" >>"$aggregate_cleanup_log"
      cleanup_failed=1
      continue
    fi
    for resource in container volume network; do
      case "$resource" in
        container) ids=$("$DOCKER" ps -aq --filter "label=ibc-test=$run_id" 2>>"$aggregate_cleanup_log") || ids=__list_failed__ ;;
        volume) ids=$("$DOCKER" volume ls -q --filter "label=ibc-test=$run_id" 2>>"$aggregate_cleanup_log") || ids=__list_failed__ ;;
        network) ids=$("$DOCKER" network ls -q --filter "label=ibc-test=$run_id" 2>>"$aggregate_cleanup_log") || ids=__list_failed__ ;;
      esac
      if [ "$ids" = __list_failed__ ]; then cleanup_failed=1; continue; fi
      for id in $ids; do
        case "$resource" in
          container) "$DOCKER" rm -f "$id" >>"$aggregate_cleanup_log" 2>&1 || cleanup_failed=1 ;;
          volume) "$DOCKER" volume rm -f "$id" >>"$aggregate_cleanup_log" 2>&1 || cleanup_failed=1 ;;
          network) "$DOCKER" network rm "$id" >>"$aggregate_cleanup_log" 2>&1 || cleanup_failed=1 ;;
        esac
      done
      case "$resource" in
        container) remaining_ids=$("$DOCKER" ps -aq --filter "label=ibc-test=$run_id" 2>>"$aggregate_cleanup_log") || remaining_ids=__list_failed__ ;;
        volume) remaining_ids=$("$DOCKER" volume ls -q --filter "label=ibc-test=$run_id" 2>>"$aggregate_cleanup_log") || remaining_ids=__list_failed__ ;;
        network) remaining_ids=$("$DOCKER" network ls -q --filter "label=ibc-test=$run_id" 2>>"$aggregate_cleanup_log") || remaining_ids=__list_failed__ ;;
      esac
      if [ "$remaining_ids" = __list_failed__ ]; then
        cleanup_failed=1
      elif [ -n "$remaining_ids" ]; then
        printf '%s resources remain for label ibc-test=%s: %s\n' \
          "$resource" "$run_id" "$remaining_ids" >>"$aggregate_cleanup_log"
        cleanup_failed=1
      fi
    done
  done
  if [ "$found_run" -eq 0 ]; then
    printf 'no aggregate run labels discovered\n' >>"$aggregate_cleanup_log"
  fi
}

cleanup_aggregate() {
  saved_code=$?
  trap - EXIT HUP INT TERM
  if aggregate_timeout_is_active; then
    saved_code=124
    stage=overall-timeout
  fi

  if [ -n "$watchdog_pid" ] && ! aggregate_timeout_is_active; then
    printf '%s\n' "$stage" >"$cleanup_requested_file.tmp"
    mv "$cleanup_requested_file.tmp" "$cleanup_requested_file"
    cleanup_timer_pid=
    if [ -s "$watchdog_timer_pid_file" ]; then
      IFS= read -r cleanup_timer_pid <"$watchdog_timer_pid_file" || cleanup_timer_pid=
    fi
    case "$cleanup_timer_pid" in
      '' | *[!0-9]*) ;;
      *) kill -TERM "$cleanup_timer_pid" 2>/dev/null || true ;;
    esac
  fi

  cleanup_run_labels

  if [ -s "$release_dir/cleanup-timeout.txt" ] &&
    grep -Eq '^stage=aggregate-cleanup-timeout$' "$release_dir/cleanup-timeout.txt"; then
    saved_code=1
    stage=aggregate-cleanup-timeout
  fi

  if [ -n "$watchdog_pid" ]; then
    kill -STOP "$watchdog_pid" 2>/dev/null || true
    timer_pid=
    if [ -s "$watchdog_timer_pid_file" ]; then
      IFS= read -r timer_pid <"$watchdog_timer_pid_file" || timer_pid=
    fi
    case "$timer_pid" in '' | *[!0-9]*) ;; *) kill -KILL "$timer_pid" 2>/dev/null || true ;; esac
    kill -KILL "$watchdog_pid" 2>/dev/null || true
    wait "$watchdog_pid" 2>/dev/null || true
    rm -f "$watchdog_pid_file" "$watchdog_timer_pid_file"
  fi

  rm -f "$artifact_root_file" "$cleanup_requested_file"
  rmdir "$artifact_root_lock_dir" 2>/dev/null || true
  rmdir "$control_dir" 2>/dev/null || true

  if [ "$saved_code" -eq 0 ] && [ "$cleanup_failed" -eq 0 ]; then
    stage=complete
    write_status passed 0
    exit 0
  fi
  if [ "$saved_code" -eq 0 ]; then
    saved_code=1
    stage=aggregate-cleanup
  fi
  invalidate_gate_manifest_at "$release_dir"
  if [ ! -s "$release_dir/gate-failure.json" ]; then
    write_gate_failure "$stage" "P0/P1 aggregate failed with exit code $saved_code"
  fi
  write_status failed "$saved_code"
  exit "$saved_code"
}
trap cleanup_aggregate EXIT

handle_aggregate_signal() {
  if aggregate_timeout_is_active; then
    stage=overall-timeout
    exit 124
  fi
  stage=interrupted
  exit 130
}
trap handle_aggregate_signal HUP INT TERM

write_status running 0

stage=validate-opt-in
if [ "${PANACEA_E2E_RELEASE_AGGREGATE:-}" != 1 ]; then
  echo 'PANACEA_E2E_RELEASE_AGGREGATE=1 is required' >&2
  exit 2
fi

stage=validate-run-id
if printf '%s\n' "$E2E_RELEASE_HARDENING_RUN_ID" | grep -Eqv '^p0p1-[0-9A-Za-z][0-9A-Za-z._-]*$'; then
  echo "invalid P0/P1 aggregate run ID: $E2E_RELEASE_HARDENING_RUN_ID" >&2
  exit 2
fi

for timeout_value in \
  "$E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS" \
  "$E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS" \
  "$E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS" \
  "$E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS"; do
  case "$timeout_value" in '' | *[!0-9]* | 0) echo 'aggregate timeouts must be positive integers' >&2; exit 2 ;; esac
done
reserved_timeout=$((E2E_RELEASE_AGGREGATE_CLEANUP_TIMEOUT_SECONDS + E2E_RELEASE_AGGREGATE_FORCE_EXIT_TIMEOUT_SECONDS))
if [ "$E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS" -le "$reserved_timeout" ]; then
  echo 'aggregate total timeout must exceed cleanup and force-exit timeouts' >&2
  exit 2
fi
aggregate_work_timeout_seconds=$((E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS - reserved_timeout))
aggregate_work_deadline_epoch=$((aggregate_started_epoch + aggregate_work_timeout_seconds))
write_status running 0

stage=validate-process-control
aggregate_child_pids "$$" >/dev/null
start_aggregate_watchdog

stage=validate-paths
export E2E_ROOT E2E_GOCACHE E2E_GOMODCACHE
sh "$script_dir/validate-paths.sh" >"$release_dir/path-validation.txt" 2>&1

stage=relocate-artifacts
aggregate_parent=${E2E_ROOT%/*}
mkdir -p "$aggregate_parent"
lock_artifact_root
if [ -e "$E2E_ROOT" ] || [ -L "$E2E_ROOT" ]; then
  unlock_artifact_root
  echo "refusing to reuse P0/P1 aggregate artifact root: $E2E_ROOT" >&2
  exit 2
fi
mv "$artifact_root" "$E2E_ROOT"
set_artifact_root "$E2E_ROOT"
unlock_artifact_root
sh "$script_dir/validate-paths.sh" >>"$release_dir/path-validation.txt" 2>&1
write_status running 0

stage=validate-clean-source
source_commit=$(git rev-parse HEAD)
printf '%s\n' "$source_commit" >"$release_dir/source-commit.txt"
git status --porcelain=v1 --untracked-files=all >"$release_dir/source-status.txt"
git diff --binary HEAD -- >"$release_dir/source-diff.patch"
if [ -s "$release_dir/source-status.txt" ] || [ -s "$release_dir/source-diff.patch" ]; then
  echo 'P0/P1 aggregate requires a clean HEAD worktree' >&2
  exit 2
fi

{
  printf 'aggregate_total_timeout_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS"
  printf 'aggregate_work_timeout_seconds=%s\n' "$aggregate_work_timeout_seconds"
  printf 'aggregate_work_deadline_epoch=%s\n' "$aggregate_work_deadline_epoch"
  printf 'configured_release_total_timeout_seconds=%s\n' "${E2E_RELEASE_TOTAL_TIMEOUT_SECONDS:-21600}"
  printf 'child_exit_margin_seconds=%s\n' "$E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS"
  printf 'contract=the aggregate deadline is authoritative; the nested release runner is capped to the remaining work window minus the child-exit margin\n'
} >"$release_dir/budget-contract.txt"

stage=run-p0p1-suites
{
	printf 'command=release-hardening-inner\n'
	printf 'source_commit=%s\n' "$source_commit"
	printf 'work_deadline_epoch=%s\n' "$aggregate_work_deadline_epoch"
} >"$release_dir/aggregate-command.txt"
set +e
E2E_ROOT="$E2E_ROOT" \
	E2E_GOCACHE="$E2E_GOCACHE" \
	E2E_GOMODCACHE="$E2E_GOMODCACHE" \
	E2E_RELEASE_HARDENING_RUN_ID="$E2E_RELEASE_HARDENING_RUN_ID" \
	E2E_RELEASE_AGGREGATE_WORK_DEADLINE_EPOCH="$aggregate_work_deadline_epoch" \
	E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS="$E2E_RELEASE_AGGREGATE_CHILD_EXIT_MARGIN_SECONDS" \
	E2E_CURRENT_SOURCE_COMMIT="$source_commit" \
	COMMIT="$source_commit" \
	"$E2E_RUNNER" release-hardening-inner >"$release_dir/aggregate.log" 2>&1
runner_status=$?
set -e
if aggregate_timeout_is_active; then
	runner_status=124
fi
if [ "$runner_status" -ne 0 ]; then
	exit "$runner_status"
fi

stage=validate-final-gate
if [ ! -s "$release_dir/gate-manifest.json" ]; then
  echo 'P0/P1 inner command completed without release/gate-manifest.json' >&2
  exit 1
fi
final_commit=$(git rev-parse HEAD)
git status --porcelain=v1 --untracked-files=all >"$release_dir/source-status-final.txt"
git diff --binary HEAD -- >"$release_dir/source-diff-final.patch"
if [ "$final_commit" != "$source_commit" ] || [ -s "$release_dir/source-status-final.txt" ] || [ -s "$release_dir/source-diff-final.patch" ]; then
  echo 'source changed during the P0/P1 aggregate' >&2
  exit 2
fi

stage=complete
