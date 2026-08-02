#!/usr/bin/env bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

REHEARSAL_ROOT="${PANACEA_REHEARSAL_ROOT:-$REPO_ROOT/.local/upgrade/cosmos-sdk-v0.50/rehearsal}"
OLD_TAG="${PANACEA_REHEARSAL_OLD_TAG:-v2.2.1}"
UPGRADE_NAME="${PANACEA_REHEARSAL_UPGRADE_NAME:-v2.3.0}"
CHAIN_ID="${PANACEA_REHEARSAL_CHAIN_ID:-panacea-local-upgrade-rehearsal}"
MONIKER="${PANACEA_REHEARSAL_MONIKER:-panacea-rehearsal-validator}"
VALIDATOR_KEY="${PANACEA_REHEARSAL_VALIDATOR_KEY:-validator}"
RECIPIENT_KEY="${PANACEA_REHEARSAL_RECIPIENT_KEY:-recipient}"
DENOM="${PANACEA_REHEARSAL_DENOM:-umed}"
NODE_RPC="${PANACEA_REHEARSAL_NODE_RPC:-tcp://127.0.0.1:26657}"
UPGRADE_HEIGHT_OFFSET="${PANACEA_REHEARSAL_UPGRADE_HEIGHT_OFFSET:-40}"
GOV_VOTING_PERIOD="${PANACEA_REHEARSAL_GOV_VOTING_PERIOD:-20s}"
GOV_MAX_DEPOSIT_PERIOD="${PANACEA_REHEARSAL_GOV_MAX_DEPOSIT_PERIOD:-20s}"
PROPOSAL_DEPOSIT_AMOUNT="${PANACEA_REHEARSAL_PROPOSAL_DEPOSIT_AMOUNT:-1000000}"
PROPOSAL_DEPOSIT="${PANACEA_REHEARSAL_PROPOSAL_DEPOSIT:-${PROPOSAL_DEPOSIT_AMOUNT}${DENOM}}"
TX_FEES="${PANACEA_REHEARSAL_TX_FEES:-1000${DENOM}}"
TX_GAS="${PANACEA_REHEARSAL_TX_GAS:-300000}"
SMOKE_AMOUNT="${PANACEA_REHEARSAL_SMOKE_AMOUNT:-1${DENOM}}"
WAIT_TIMEOUT_SECONDS="${PANACEA_REHEARSAL_WAIT_TIMEOUT_SECONDS:-180}"
KEEP_RUNNING="${PANACEA_REHEARSAL_KEEP_RUNNING:-0}"

BIN_ROOT="$REHEARSAL_ROOT/bin"
SRC_ROOT="$REHEARSAL_ROOT/src"
OLD_SRC="$SRC_ROOT/$OLD_TAG"
OLD_BIN="$BIN_ROOT/panacead-$OLD_TAG"
NEW_BIN="$BIN_ROOT/panacead-$UPGRADE_NAME"
DAEMON_HOME="$REHEARSAL_ROOT/home"
LOG_DIR="$REHEARSAL_ROOT/logs"
LOG_FILE="$LOG_DIR/cosmovisor.log"
TX_LOG="$LOG_DIR/tx.jsonl"
PID_FILE="$REHEARSAL_ROOT/cosmovisor.pid"
BACKUP_DIR="$REHEARSAL_ROOT/backups"
TARGET_HEIGHT_FILE="$REHEARSAL_ROOT/upgrade-height"
PROPOSAL_ID_FILE="$REHEARSAL_ROOT/proposal-id"
GO_BUILD_CACHE="${PANACEA_REHEARSAL_GOCACHE:-$REHEARSAL_ROOT/go-build-cache}"

log() {
  printf '[upgrade-rehearsal] %s\n' "$*"
}

fail() {
  printf '[upgrade-rehearsal] ERROR: %s\n' "$*" >&2
  if [ -f "$LOG_FILE" ]; then
    printf '\n[upgrade-rehearsal] Last cosmovisor log lines:\n' >&2
    tail -n 80 "$LOG_FILE" >&2 || true
  fi
  exit 1
}

usage() {
  cat <<EOF
Usage: $0 [command]

Commands:
  run             Clean, build old/new binaries, init a local chain, run the full $UPGRADE_NAME Cosmovisor rehearsal.
  build           Build the $OLD_TAG old binary and the current $UPGRADE_NAME binary into .local/.
  init            Reset local chain state and initialize Cosmovisor home from existing binaries.
  start           Start Cosmovisor in the background.
  submit-upgrade  Submit the $UPGRADE_NAME software-upgrade proposal.
  vote            Vote yes on the latest submitted proposal.
  wait-upgrade    Wait until the upgrade height is reached and Cosmovisor switches binaries.
  smoke           Run one bank-send smoke test against the running chain.
  status          Print local height, pid, and Cosmovisor current symlink.
  stop            Stop the background Cosmovisor process.
  reset-state     Stop and remove local chain home, logs, backups, and proposal state. Keeps built binaries.
  clean           Stop and remove the whole rehearsal root, including built binaries and archived source.

Environment overrides use the PANACEA_REHEARSAL_* prefix. Common examples:
  PANACEA_REHEARSAL_ROOT=$REHEARSAL_ROOT
  PANACEA_REHEARSAL_KEEP_RUNNING=1
  PANACEA_REHEARSAL_UPGRADE_HEIGHT_OFFSET=$UPGRADE_HEIGHT_OFFSET
EOF
}

require_tool() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required tool: $1"
}

assert_safe_root() {
  case "$REHEARSAL_ROOT" in
    "$REPO_ROOT/.local/"* | /private/tmp/* | /tmp/*)
      return
      ;;
  esac

  if [ "${PANACEA_REHEARSAL_ALLOW_EXTERNAL_ROOT:-0}" != "1" ]; then
    fail "refusing to remove or initialize outside repo .local/: $REHEARSAL_ROOT"
  fi
}

require_common_tools() {
  require_tool git
  require_tool go
  require_tool jq
  require_tool make
  require_tool perl
  require_tool tar
}

old_panacead() {
  printf '%s\n' "$OLD_BIN"
}

current_panacead() {
  local current="$DAEMON_HOME/cosmovisor/current/bin/panacead"
  if [ -x "$current" ]; then
    printf '%s\n' "$current"
  else
    old_panacead
  fi
}

is_running() {
  if [ ! -f "$PID_FILE" ]; then
    return 1
  fi

  local pid
  pid="$(cat "$PID_FILE")"
  [ -n "$pid" ] && kill -0 "$pid" >/dev/null 2>&1
}

stop_node() {
  if ! is_running; then
    rm -f "$PID_FILE"
    return
  fi

  local pid
  pid="$(cat "$PID_FILE")"
  log "Stopping Cosmovisor pid $pid"
  kill "$pid" >/dev/null 2>&1 || true

  local deadline=$((SECONDS + 30))
  while kill -0 "$pid" >/dev/null 2>&1; do
    if [ "$SECONDS" -ge "$deadline" ]; then
      kill -9 "$pid" >/dev/null 2>&1 || true
      break
    fi
    sleep 1
  done

  rm -f "$PID_FILE"
}

clean_all() {
  assert_safe_root
  stop_node
  rm -rf "$REHEARSAL_ROOT"
  log "Removed $REHEARSAL_ROOT"
}

reset_state() {
  assert_safe_root
  stop_node
  rm -rf "$DAEMON_HOME" "$LOG_DIR" "$BACKUP_DIR" "$PID_FILE" "$TARGET_HEIGHT_FILE" "$PROPOSAL_ID_FILE"
  mkdir -p "$LOG_DIR"
  : >"$TX_LOG"
  log "Reset local chain state under $REHEARSAL_ROOT"
}

build_binaries() {
  require_common_tools
  assert_safe_root

  mkdir -p "$BIN_ROOT" "$SRC_ROOT" "$GO_BUILD_CACHE"
  export GOCACHE="$GO_BUILD_CACHE"

  log "Building old binary from tag $OLD_TAG"
  rm -rf "$OLD_SRC"
  mkdir -p "$OLD_SRC"
  git -C "$REPO_ROOT" archive "$OLD_TAG" | tar -x -C "$OLD_SRC"

  local old_commit
  old_commit="$(git -C "$REPO_ROOT" rev-parse "$OLD_TAG^{commit}")"
  make -C "$OLD_SRC" build LEDGER_ENABLED=false BUILDDIR="$BIN_ROOT/old" VERSION="${OLD_TAG#v}" COMMIT="$old_commit"
  cp "$BIN_ROOT/old/panacead" "$OLD_BIN"

  log "Building current new binary for $UPGRADE_NAME"
  make -C "$REPO_ROOT" build LEDGER_ENABLED=false BUILDDIR="$BIN_ROOT/new"
  cp "$BIN_ROOT/new/panacead" "$NEW_BIN"

  log "Old binary: $OLD_BIN"
  "$OLD_BIN" version --long | sed -n '1,8p'
  log "New binary: $NEW_BIN"
  "$NEW_BIN" version --long | sed -n '1,8p'
}

require_binaries() {
  [ -x "$OLD_BIN" ] || fail "old binary is missing. Run: $0 build"
  [ -x "$NEW_BIN" ] || fail "new binary is missing. Run: $0 build"
}

install_cosmovisor_layout() {
  mkdir -p \
    "$DAEMON_HOME/cosmovisor/genesis/bin" \
    "$DAEMON_HOME/cosmovisor/upgrades/$UPGRADE_NAME/bin"

  cp "$OLD_BIN" "$DAEMON_HOME/cosmovisor/genesis/bin/panacead"
  cp "$NEW_BIN" "$DAEMON_HOME/cosmovisor/upgrades/$UPGRADE_NAME/bin/panacead"
  chmod +x \
    "$DAEMON_HOME/cosmovisor/genesis/bin/panacead" \
    "$DAEMON_HOME/cosmovisor/upgrades/$UPGRADE_NAME/bin/panacead"
  ln -sfn "$DAEMON_HOME/cosmovisor/genesis" "$DAEMON_HOME/cosmovisor/current"
}

validator_addr() {
  "$(old_panacead)" keys show "$VALIDATOR_KEY" -a --home "$DAEMON_HOME" --keyring-backend test
}

recipient_addr() {
  "$(old_panacead)" keys show "$RECIPIENT_KEY" -a --home "$DAEMON_HOME" --keyring-backend test
}

patch_local_genesis_and_config() {
  local genesis="$DAEMON_HOME/config/genesis.json"
  local tmp
  tmp="$(mktemp "$REHEARSAL_ROOT/genesis.XXXXXX")"

  jq \
    --arg denom "$DENOM" \
    --arg min_deposit "$PROPOSAL_DEPOSIT_AMOUNT" \
    --arg voting_period "$GOV_VOTING_PERIOD" \
    --arg max_deposit_period "$GOV_MAX_DEPOSIT_PERIOD" \
    '.app_state.gov.params.min_deposit = [{"denom": $denom, "amount": $min_deposit}]
     | .app_state.gov.params.voting_period = $voting_period
     | .app_state.gov.params.max_deposit_period = $max_deposit_period' \
    "$genesis" >"$tmp"
  mv "$tmp" "$genesis"

  perl -0pi -e 's/timeout_commit = ".*?"/timeout_commit = "1s"/' "$DAEMON_HOME/config/config.toml"
  perl -0pi -e 's/create_empty_blocks_interval = ".*?"/create_empty_blocks_interval = "0s"/' "$DAEMON_HOME/config/config.toml"
  perl -0pi -e "s/minimum-gas-prices = \".*?\"/minimum-gas-prices = \"0$DENOM\"/" "$DAEMON_HOME/config/app.toml"
}

init_chain() {
  require_common_tools
  require_binaries
  reset_state
  install_cosmovisor_layout

  local old
  old="$(old_panacead)"

  log "Initializing $CHAIN_ID in $DAEMON_HOME"
  "$old" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$DAEMON_HOME"
  "$old" keys add "$VALIDATOR_KEY" --home "$DAEMON_HOME" --keyring-backend test --output json >"$REHEARSAL_ROOT/$VALIDATOR_KEY-key.json"
  "$old" keys add "$RECIPIENT_KEY" --home "$DAEMON_HOME" --keyring-backend test --output json >"$REHEARSAL_ROOT/$RECIPIENT_KEY-key.json"

  "$old" genesis add-genesis-account "$VALIDATOR_KEY" "1000000000000$DENOM" --home "$DAEMON_HOME" --keyring-backend test
  "$old" genesis add-genesis-account "$RECIPIENT_KEY" "1000000000$DENOM" --home "$DAEMON_HOME" --keyring-backend test
  "$old" genesis gentx "$VALIDATOR_KEY" "500000000000$DENOM" \
    --chain-id "$CHAIN_ID" \
    --home "$DAEMON_HOME" \
    --keyring-backend test \
    --moniker "$MONIKER" \
    --commission-rate "0.05" \
    --commission-max-rate "0.20" \
    --commission-max-change-rate "0.01"
  "$old" genesis collect-gentxs --home "$DAEMON_HOME"

  patch_local_genesis_and_config
  "$old" genesis validate-genesis --home "$DAEMON_HOME"
  log "Initialized validator $(validator_addr) and recipient $(recipient_addr)"
}

start_node() {
  require_tool cosmovisor
  require_binaries

  if is_running; then
    fail "Cosmovisor is already running with pid $(cat "$PID_FILE")"
  fi

  mkdir -p "$LOG_DIR" "$BACKUP_DIR"
  : >"$LOG_FILE"

  log "Starting Cosmovisor. Log: $LOG_FILE"
  (
    exec env \
      DAEMON_NAME=panacead \
      DAEMON_HOME="$DAEMON_HOME" \
      DAEMON_ALLOW_DOWNLOAD_BINARIES=false \
      DAEMON_RESTART_AFTER_UPGRADE=true \
      DAEMON_DATA_BACKUP_DIR="$BACKUP_DIR" \
      DAEMON_POLL_INTERVAL=300ms \
      UNSAFE_SKIP_BACKUP=true \
      cosmovisor run start --home "$DAEMON_HOME"
  ) >>"$LOG_FILE" 2>&1 &

  printf '%s\n' "$!" >"$PID_FILE"
  wait_for_height 2
  log "Cosmovisor started at height $(latest_height)"
}

latest_height() {
  local bin status
  bin="$(current_panacead)"
  status="$("$bin" status --node "$NODE_RPC" 2>/dev/null || true)"
  if [ -z "$status" ]; then
    return 1
  fi
  jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height // empty' <<<"$status"
}

wait_for_height() {
  local target="$1"
  local deadline=$((SECONDS + WAIT_TIMEOUT_SECONDS))
  local height

  while [ "$SECONDS" -lt "$deadline" ]; do
    height="$(latest_height || true)"
    if [[ "$height" =~ ^[0-9]+$ ]] && [ "$height" -ge "$target" ]; then
      return
    fi

    if [ -f "$PID_FILE" ] && ! is_running; then
      fail "Cosmovisor exited before reaching height $target"
    fi
    sleep 1
  done

  fail "timed out waiting for height $target; latest height: ${height:-unknown}"
}

wait_for_next_block() {
  local height
  height="$(latest_height)"
  wait_for_height "$((height + 1))"
}

wait_for_tx() {
  local txhash="$1"
  local deadline=$((SECONDS + WAIT_TIMEOUT_SECONDS))
  local bin result code

  while [ "$SECONDS" -lt "$deadline" ]; do
    bin="$(current_panacead)"
    result="$("$bin" query tx "$txhash" --node "$NODE_RPC" --output json 2>/dev/null || true)"
    if jq -e . >/dev/null 2>&1 <<<"$result"; then
      code="$(jq -r '.code // 0' <<<"$result")"
      if [ "$code" = "0" ]; then
        return
      fi
      fail "tx $txhash failed with code $code: $(jq -r '.raw_log // .codespace // empty' <<<"$result")"
    fi
    sleep 1
  done

  fail "timed out waiting for tx $txhash"
}

broadcast_tx() {
  local tx_json code txhash
  if ! tx_json="$("$@" 2>&1)"; then
    fail "tx command failed: $tx_json"
  fi

  printf '%s\n' "$tx_json" >>"$TX_LOG"

  if ! jq -e . >/dev/null 2>&1 <<<"$tx_json"; then
    fail "tx command did not return JSON: $tx_json"
  fi

  code="$(jq -r '.code // 0' <<<"$tx_json")"
  if [ "$code" != "0" ]; then
    fail "tx check failed with code $code: $(jq -r '.raw_log // empty' <<<"$tx_json")"
  fi

  txhash="$(jq -r '.txhash // empty' <<<"$tx_json")"
  [ -n "$txhash" ] || fail "tx response did not include txhash: $tx_json"
  wait_for_tx "$txhash"
}

set_tx_common_flags() {
  TX_COMMON_FLAGS=(
    --home "$DAEMON_HOME" \
    --chain-id "$CHAIN_ID" \
    --node "$NODE_RPC" \
    --keyring-backend test \
    --fees "$TX_FEES" \
    --gas "$TX_GAS" \
    --broadcast-mode sync \
    --output json \
    -y
  )
}

balance_of() {
  local addr="$1"
  local bin result
  bin="$(current_panacead)"
  result="$("$bin" query bank balances "$addr" --denom "$DENOM" --node "$NODE_RPC" --output json)"
  jq -r --arg denom "$DENOM" '.amount // .balance.amount // ([.balances[]? | select(.denom == $denom) | .amount][0]) // "0"' <<<"$result"
}

smoke_bank_send() {
  local label="${1:-manual}"
  local bin recipient before after
  set_tx_common_flags

  bin="$(current_panacead)"
  recipient="$(recipient_addr)"
  before="$(balance_of "$recipient")"

  log "Running $label bank send smoke test: $SMOKE_AMOUNT to $recipient"
  broadcast_tx "$bin" tx bank send "$VALIDATOR_KEY" "$recipient" "$SMOKE_AMOUNT" --from "$VALIDATOR_KEY" "${TX_COMMON_FLAGS[@]}"
  wait_for_next_block

  after="$(balance_of "$recipient")"
  if [ "$after" -le "$before" ]; then
    fail "$label bank send did not increase recipient balance: before=$before after=$after"
  fi
  log "$label bank send passed: recipient balance $before -> $after $DENOM"
}

latest_proposal_id() {
  local bin proposals
  bin="$(current_panacead)"
  proposals="$("$bin" query gov proposals --node "$NODE_RPC" --output json --reverse --limit 20)"
  jq -r '[.proposals[]? | (.id // .proposal_id)] | map(tonumber) | max // empty' <<<"$proposals"
}

submit_upgrade() {
  local bin height target proposal_id
  set_tx_common_flags

  height="$(latest_height)"
  target="$((height + UPGRADE_HEIGHT_OFFSET))"
  printf '%s\n' "$target" >"$TARGET_HEIGHT_FILE"

  bin="$(current_panacead)"
  log "Submitting software-upgrade proposal $UPGRADE_NAME at height $target"
  broadcast_tx "$bin" tx gov submit-legacy-proposal software-upgrade "$UPGRADE_NAME" \
    --from "$VALIDATOR_KEY" \
    --title "$UPGRADE_NAME local upgrade rehearsal" \
    --description "Local Cosmovisor rehearsal of the $UPGRADE_NAME upgrade handler." \
    --deposit "$PROPOSAL_DEPOSIT" \
    --upgrade-height "$target" \
    --upgrade-info "{}" \
    --no-validate \
    "${TX_COMMON_FLAGS[@]}"

  wait_for_next_block
  proposal_id="$(latest_proposal_id)"
  [ -n "$proposal_id" ] || fail "could not find submitted proposal id"
  printf '%s\n' "$proposal_id" >"$PROPOSAL_ID_FILE"
  log "Submitted proposal id $proposal_id for height $target"
}

proposal_status() {
  local proposal_id="$1"
  local bin result
  bin="$(current_panacead)"
  result="$("$bin" query gov proposal "$proposal_id" --node "$NODE_RPC" --output json)"
  jq -r '.proposal.status // .status // empty' <<<"$result"
}

vote_upgrade() {
  [ -f "$PROPOSAL_ID_FILE" ] || fail "proposal id file is missing. Run: $0 submit-upgrade"

  local bin proposal_id
  set_tx_common_flags

  proposal_id="$(cat "$PROPOSAL_ID_FILE")"
  bin="$(current_panacead)"
  log "Voting yes on proposal $proposal_id"
  broadcast_tx "$bin" tx gov vote "$proposal_id" yes --from "$VALIDATOR_KEY" "${TX_COMMON_FLAGS[@]}"
}

wait_for_proposal_passed() {
  [ -f "$PROPOSAL_ID_FILE" ] || fail "proposal id file is missing"

  local proposal_id status deadline
  proposal_id="$(cat "$PROPOSAL_ID_FILE")"
  deadline=$((SECONDS + WAIT_TIMEOUT_SECONDS))

  while [ "$SECONDS" -lt "$deadline" ]; do
    status="$(proposal_status "$proposal_id")"
    case "$status" in
      PROPOSAL_STATUS_PASSED | Passed | passed)
        log "Proposal $proposal_id passed"
        return
        ;;
      PROPOSAL_STATUS_REJECTED | PROPOSAL_STATUS_FAILED | Rejected | Failed | rejected | failed)
        fail "proposal $proposal_id ended with status $status"
        ;;
    esac
    sleep 2
  done

  fail "timed out waiting for proposal $proposal_id to pass; latest status: ${status:-unknown}"
}

wait_for_upgrade_switch() {
  [ -f "$TARGET_HEIGHT_FILE" ] || fail "upgrade height file is missing. Run: $0 submit-upgrade"

  local target current_link
  target="$(cat "$TARGET_HEIGHT_FILE")"
  log "Waiting for upgrade height $target"
  wait_for_height "$((target + 2))"

  current_link="$(readlink "$DAEMON_HOME/cosmovisor/current" || true)"
  case "$current_link" in
    *"$UPGRADE_NAME"*)
      log "Cosmovisor current symlink switched to $current_link"
      ;;
    *)
      fail "Cosmovisor current symlink did not switch to $UPGRADE_NAME; current=$current_link"
      ;;
  esac
}

status_summary() {
  local height current_link pid
  height="$(latest_height || true)"
  current_link="$(readlink "$DAEMON_HOME/cosmovisor/current" 2>/dev/null || true)"
  pid="stopped"
  if is_running; then
    pid="$(cat "$PID_FILE")"
  fi

  cat <<EOF
root: $REHEARSAL_ROOT
chain-id: $CHAIN_ID
upgrade-name: $UPGRADE_NAME
height: ${height:-unavailable}
cosmovisor-pid: $pid
current: ${current_link:-unavailable}
log: $LOG_FILE
EOF
}

run_all() {
  clean_all
  build_binaries
  init_chain
  start_node

  if [ "$KEEP_RUNNING" != "1" ]; then
    trap stop_node EXIT
  fi

  smoke_bank_send pre-upgrade
  submit_upgrade
  vote_upgrade
  wait_for_proposal_passed
  wait_for_upgrade_switch
  smoke_bank_send post-upgrade

  log "$UPGRADE_NAME upgrade rehearsal completed"
  status_summary

  if [ "$KEEP_RUNNING" != "1" ]; then
    stop_node
    trap - EXIT
  else
    log "Cosmovisor left running because PANACEA_REHEARSAL_KEEP_RUNNING=1"
  fi
}

case "${1:-run}" in
  run)
    run_all
    ;;
  build)
    build_binaries
    ;;
  init)
    init_chain
    ;;
  start)
    start_node
    ;;
  submit-upgrade)
    submit_upgrade
    ;;
  vote)
    vote_upgrade
    ;;
  wait-upgrade)
    wait_for_upgrade_switch
    ;;
  smoke)
    smoke_bank_send manual
    ;;
  status)
    status_summary
    ;;
  stop)
    stop_node
    ;;
  reset-state)
    reset_state
    ;;
  clean)
    clean_all
    ;;
  -h | --help | help)
    usage
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
