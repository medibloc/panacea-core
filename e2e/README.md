# Panacea real-node E2E tests

This module runs real `panacead` processes in Docker through Interchaintest.
It does not call or source `scripts/upgrade-local/run.sh`; that script remains
the independent single-validator Cosmovisor rehearsal.

## Prerequisites

- Docker with BuildKit and the current daemon reachable from `docker context inspect`
- Go 1.23.12 selected locally (`GOTOOLCHAIN=local` is enforced by default)
- the full Git history containing commit `a1b342939ba6ac3092aeebbee6a2fa741a34d47f`

Verify the complete selected toolchain before running a suite:

```sh
mise exec go@1.23.12 -- ./scripts/e2e/run.sh check
```

Prefer selecting Go 1.23.12 with the local version manager so the command above
prints `go1.23.12`. The runner enforces `GOTOOLCHAIN=local` and `GOWORK=off` for
the nested E2E module, so it cannot silently download a different toolchain or
join an ambient workspace. The preflight checks the selected `go` command,
`GOVERSION`, `GOROOT`, `GOTOOLDIR` and compiler identity before compilation. It
prints all toolchain paths and a remediation hint if a stale `GOROOT` or `GOBIN`
mixes installations.

## Commands

From the repository root:

```sh
mise exec go@1.23.12 -- ./scripts/e2e/run.sh check
mise exec go@1.23.12 -- ./scripts/e2e/run.sh check-clean
mise exec go@1.23.12 -- ./scripts/e2e/run.sh build-current
mise exec go@1.23.12 -- ./scripts/e2e/run.sh build-v2.2.1
mise exec go@1.23.12 -- ./scripts/e2e/run.sh build-images
mise exec go@1.23.12 -- ./scripts/e2e/run.sh build-test-binary
mise exec go@1.23.12 -- ./scripts/e2e/run.sh build
mise exec go@1.23.12 -- ./scripts/e2e/run.sh unit
mise exec go@1.23.12 -- ./scripts/e2e/run.sh smoke
mise exec go@1.23.12 -- ./scripts/e2e/run.sh v2.2.1
mise exec go@1.23.12 -- ./scripts/e2e/run.sh compatibility
mise exec go@1.23.12 -- ./scripts/e2e/run.sh negative
mise exec go@1.23.12 -- ./scripts/e2e/run.sh restart
mise exec go@1.23.12 -- ./scripts/e2e/run.sh consensus
mise exec go@1.23.12 -- ./scripts/e2e/run.sh upgrade
mise exec go@1.23.12 -- ./scripts/e2e/run.sh upgrade-deep
mise exec go@1.23.12 -- ./scripts/e2e/run.sh upgrade-chaos
mise exec go@1.23.12 -- ./scripts/e2e/run.sh state-sync
mise exec go@1.23.12 -- ./scripts/e2e/run.sh config-compat
mise exec go@1.23.12 -- ./scripts/e2e/run.sh ibc-upgrade
mise exec go@1.23.12 -- ./scripts/e2e/run.sh network-faults
mise exec go@1.23.12 -- ./scripts/e2e/run.sh release-builds
mise exec go@1.23.12 -- ./scripts/e2e/run.sh release-hardening
mise exec go@1.23.12 -- ./scripts/e2e/run.sh release-hardening-inner
mise exec go@1.23.12 -- ./scripts/e2e/run.sh load
mise exec go@1.23.12 -- ./scripts/e2e/run.sh all
mise exec go@1.23.12 -- ./scripts/e2e/run.sh help
```

`build` is the clean-build contract: it builds the current worktree node
image deliberately stamped `2.3.0`, the immutable v2.2.1 node image, and the
E2E test binary. `build-images` builds only the two node images.

Live Docker suites compile `panacea-e2e.test` once under `E2E_ROOT` and execute
that binary directly. Keep using the standalone runner instead of replacing a
live command with `go test`: on macOS/Colima the Go test driver can leave
Interchaintest's Docker wildcard host-port RPC dial blocked even while the
nodes themselves are healthy. `unit` still uses `go test` for both isolated
nested modules.

`restart` runs its two recovery scenarios in separate test-binary processes.
If Docker reports the exact transient `failed to bind host port ... address
already in use` setup race, that scenario is retried once; every other failure
is returned immediately.

Every Docker suite command selects an explicit canonical test regex and has a Go test
deadline in addition to the bounded waits inside the harness:

| Command | Topology | Default deadline | Contract |
|---|---:|---:|---|
| `smoke` | 1 validator + 1 full node | 12m per invocation | Current `v2.3.0` image NFT lifecycle, endpoint parity, pagination, the isolated failure-artifact probe, and unsupported DB-backend startup rejection |
| `v2.2.1` | 1 validator + 1 full node | 12m | Fresh-network compatibility at immutable v2.2.1 source |
| `negative` | fresh 1 validator + 1 full node networks | 40m | Authorization/state integrity, protocol boundaries, and raw signed-wire rejection paths |
| `restart` | fresh 1 validator + 1 full node networks | 35m | Graceful and abrupt restart, export/bootstrap, portable application snapshot restore, and fresh full-node block sync |
| `consensus` | 4 validators + 1 full node | 18m | One-validator tolerance, quorum loss, committed transaction recovery, and common-history catch-up with same-height block ID/app hash, peer, and catching-up proof |
| `upgrade` | 4 validators + 1 full node | 35m | Coordinated exact `v2.2.1` to `v2.3.0` image switch, delayed-validator catch-up, the connected P0 transaction/state matrix, migration/state preservation, and post-upgrade lifecycle |
| `upgrade-deep` | 4 validators + 1 full node | 50m per scenario | The normal connected P0 matrix plus the adversarial legacy-PNFT upgrade path |
| `upgrade-chaos` | 4 validators + 1 full node | 40m | Three deterministic switch orders, bounded quorum stalls, forced restart, delayed-node catch-up, and single-history recovery |
| `ibc-upgrade` | Panacea + pinned Osmosis + Hermes | 45m | Existing-channel ICS-20 transfer, in-flight acknowledgement, timeout/refund, relayer restart, and post-upgrade bidirectional continuity |
| `state-sync` | current-version validator sources + new full node | 20m | Actual CometBFT state sync, corrupt/unavailable/bad-trust failures, restart, and history parity; the connected `upgrade-deep` lane separately proves joining upgraded sources |
| `config-compat` | v2.2.1 node home on current binary | 25m | v0.47 config preservation/migration and current endpoint/config-command contracts |
| `network-faults` | run-owned validator/full-node network | 25m | Container-scoped partition, proxy delay/jitter/loss, DNS/container recreation, endpoint isolation, slow-client/churn, and WebSocket recovery |
| `release-builds` | functional host platform + linux/amd64 + linux/arm64 | 6h total; 35m per architecture upgrade | Clean-HEAD provenance, functional-image identity capture, cold dependency provenance, warm-offline builds, pinned no-network Docker compilation, host binary checksum equivalence, version/smoke, and a real upgrade on each architecture |
| `load` | 1 validator + 1 full node | 25m | Boundary datasets, concurrent REST/gRPC load, mixed transactions, query-gas rejection, and required peer/catching-up/mempool/resource observations |

The load suite is an observational baseline, not a production SLA. Its pass
conditions are functional: bounded queries succeed, over-limit queries are
rejected without crashing a node or halting consensus, and validators continue
to produce blocks. Every runtime sample must include catching-up state and
nonnegative peer count, mempool transaction count, and mempool byte count; the
final samples must report that every node is caught up. Raw samples,
percentiles, resource observations, and the execution environment are retained
so a later run can be compared explicitly.
It is the existing short local baseline, not the separate P2 quick/release soak;
running it does not satisfy any P2 operational-validation completion condition.

`compatibility` combines smoke and v2.2.1 compatibility.
`all` is the complete functional aggregate. It runs harness unit tests, builds
the shared current and v2.2.1 images once, and then runs smoke, v2.2.1,
negative, restart, consensus, both normal and adversarial upgrades, upgrade
chaos, IBC upgrade continuity, state sync, config compatibility, local network
faults, and load. It does not run `release-builds`. The aggregate is
deliberately long-running. Override an individual budget only when the
execution environment requires it, for example
`E2E_LOAD_TIMEOUT=35m mise exec go@1.23.12 -- ./scripts/e2e/run.sh load`.

`release-hardening` is the serial artifact-first release gate. Its exact inner
sequence is clean-source check, `all`, `release-builds`, a second clean-source
check, and coverage merge/gate creation. It intentionally excludes P2 soak,
production snapshot, and real ingress/firewall checks; those are tracked by the
separate operational-validation goal. The aggregate has a bounded 12-hour total
deadline, configurable with
`E2E_RELEASE_AGGREGATE_TOTAL_TIMEOUT_SECONDS`. `release-builds` can be run alone
when only the cold/offline and multi-architecture release evidence is required;
its separate child deadline is six hours, configurable with
`E2E_RELEASE_TOTAL_TIMEOUT_SECONDS`. The release commands record a failed
release-run artifact before image construction when tracked, staged, or
untracked source changes exist; commit the exact source to be certified first.
`release-hardening-inner` is the aggregate runner's bounded child command; run
the public `release-hardening` wrapper in normal operation so timeout and
artifact-first failure handling cover the full sequence.
The standalone release-build command also builds and captures the functional
old/current images before independently rebuilding both release architectures.
The aggregate refuses to reuse its fresh `p0p1-*` artifact root and finishes only after merging
the connected-upgrade and isolated-IBC matrices into
`upgrade/coverage-matrix.json`; every claimed passed artifact must exist and
all 13 required rows must have passed all five phases. It then writes
`release/gate-manifest.json` only after finding exactly one successful, cleaned
run for every required P0/P1 lane, proving that all live lanes used the same
content-addressed old/current Panacea image IDs, requiring exactly one recorded
old-to-current switch for every declared node in upgrade lanes, and tying those
functional image IDs to the native-platform release image inspections through
equal `panacead` SHA-256 checksums. It also validates the two-architecture
release-build manifest against the same clean source commit.

The restart suite proves ordinary fresh-node block sync and application DB
snapshot restore; it does not claim CometBFT state-sync verification. The
consensus suite waits for every selected node to reach the target height, stop
catching up, and report at least one connected peer before comparing the exact
block ID and post-FinalizeBlock app hash at that height. The upgrade suite
switches Interchaintest-managed node images from `v2.2.1` to `v2.3.0` and proves
consensus and state migration. It does not test Cosmovisor and does not call
`scripts/upgrade-local/run.sh`.

The first run downloads Go modules on the host into `.local/e2e/go-mod`.
Panacea's Docker compile itself receives a generated vendor tree, uses pinned
multi-architecture base-image digests, and runs with networking disabled. A
machine without the base images or Go module inputs still needs network access
for that initial materialization.

The IBC upgrade lane performs a read-only Osmosis mainnet preflight immediately
before creating its local Docker topology. It queries exactly one explicit RPC
endpoint and one explicit REST endpoint to confirm `osmosis-1`, the pinned
v31.0.2 binary identity plus observable SDK/IBC-Go/CometBFT build metadata, and the active
`transfer/channel-82 -> transfer/channel-1` `ics20-1` fixture. Defaults are
`https://rpc.osmosis.zone` and `https://lcd.osmosis.zone`; operators may replace
them with `PANACEA_E2E_OSMOSIS_MAINNET_RPC_ENDPOINT` and
`PANACEA_E2E_OSMOSIS_MAINNET_REST_ENDPOINT`. There is no skip, retry endpoint,
or fallback: a firewall failure, unavailable response, stale status, or fixture
mismatch writes `ibc/mainnet-preflight.json` and fails before Docker starts.
The preflight never broadcasts a public-network transaction.

Osmosis REST node-info exposes the original Cosmos SDK module path/version and
the effective SDK version, but not the Go replacement path. The local pinned
binary's `version --long` output must therefore match the complete
`github.com/cosmos/cosmos-sdk@v0.50.14 => github.com/osmosis-labs/cosmos-sdk@v0.50.14-v30-osmo`
contract. `ibc/osmosis-source-contract.json` separately records the exact
v31.0.2 commit, SHA-256-pinned `go.mod` and transfer-wiring source, and static
receive/send stack. Live channel and node-info queries do not expose middleware
wiring or per-channel middleware activation, so the evidence reports
`pinned-source-live-limited`, never live-confirmed middleware.

## Isolation and diagnostics

Every network receives a random chain/run ID. Generated module/build caches and
successful or failed run artifacts stay below the repository's resolved
`.local/e2e/` directory. The only alternate root accepted by the path preflight
is a resolved path under the literal `/tmp`; arbitrary `TMPDIR` values are not
trusted.
`scripts/e2e/validate-paths.sh` runs before any command creates directories and
rejects traversal, symlink escapes, repository-root overrides, and caches that
resolve outside `E2E_ROOT`. Artifacts include only allowlisted public
diagnostics:

```text
<E2E_ROOT>/<run-id>/
  manifest.json
  cleanup.json
  genesis.json
  versions.txt
  nodes/<node>/config/{app,client,config}.toml
  nodes/<node>/status.json
  nodes/<node>/container-state.json
  nodes/<node>/logs/container.log
  nodes/history.jsonl
  tx/
    requests.jsonl              # submitted CLI and signed transaction requests
    broadcast-results.jsonl     # CLI/RPC broadcast and CheckTx results
    query-attempts.jsonl        # transaction commit-query attempts
    committed-results.jsonl     # committed DeliverTx results
    raw-requests.jsonl          # raw-wire requests from negative tests
    raw-broadcast-results.jsonl # raw-wire RPC broadcast results
  queries/results.jsonl
  queries/failures.jsonl
  metrics/...                 # load suite
  recovery/
    wal-replay.jsonl          # forced-restart WAL catch-up/replay markers
    ...                       # restart, export, snapshot, and sync evidence
  upgrade/...                 # upgrade suite
  ibc/mainnet-preflight.json  # fail-closed live Osmosis status/channel contract
  ibc/osmosis-source-contract.json # commit/hash-pinned dependency and wiring source facts
  ibc/chains/...              # version --long, binary and genesis checksums
  ibc-compatibility-matrix.json # old/new dependencies, channel, middleware, Hermes
  failure-summary.txt         # failed run or diagnostic/cleanup failure only
```

The default `E2E_ROOT` is `.local/e2e`. `manifest.json` and `cleanup.json`
describe both successful and failed runs; a successful run does not need a
`failure-summary.txt`. Artifacts remain local unless the operator explicitly
copies or uploads them.

Private validator keys, node keys, keyrings, and test mnemonics are never
collected. Caller-supplied deterministic test mnemonics are staged through a
mode-0600 file instead of a logged command, consumed through stdin, and removed
both by the recovery shell trap and an independent bounded cleanup call. Any
cleanup failure fails the wallet import. Cleanup targets only Docker resources
labeled with the run ID and then verifies that no matching container, volume,
or network remains.

## Standalone execution

The E2E module is intentionally absent from the root `Makefile` and GitHub CI
workflows. Root builds and `go test ./...` also do not enter this nested module.
An operator must invoke `scripts/e2e/run.sh` explicitly, retain the selected
artifact directory, and review its manifest and cleanup result. The ordinary
Docker publish workflow builds and pushes independently; it neither runs E2E nor
consumes `release/gate-manifest.json`. Consequently, a passing GitHub workflow
is not evidence that any standalone E2E command passed.

For a release candidate, run `release-hardening` from a clean committed HEAD on
a host that can actually execute both `linux/amd64` and `linux/arm64`, then
archive the resulting `E2E_ROOT` before publishing. Use distinct roots for
concurrent or repeated runs; no command reads another run's node homes or
artifacts.

The independent Cosmovisor operator rehearsal remains in CI for release branches
and manual dispatch. It runs `scripts/upgrade-local/run.sh run` and stays
separate from the standalone `upgrade` E2E command; neither implementation
calls or sources the other.
