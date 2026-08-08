package harness

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	dockerclient "github.com/docker/docker/client"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"go.uber.org/zap/zaptest"
)

// Config controls one isolated Interchaintest network.
type Config struct {
	Image              ImageRef
	NumValidators      int
	NumFullNodes       int
	RunID              string
	ArtifactRoot       string
	DBBackend          string
	TimeoutCommit      string
	QueryGasLimit      uint64
	SnapshotInterval   uint64
	SnapshotKeepRecent uint32
	EnableTelemetry    bool

	// DisableOptionalCPUFeatures is for a release image running under a
	// non-native CPU emulator. It must remain false for native E2E runs.
	DisableOptionalCPUFeatures bool

	// Test-only genesis overrides. Their zero values leave source genesis intact.
	StakingUnbondingTime          string
	SlashingSignedBlocksWindow    int64
	SlashingMinSignedPerWindow    string
	SlashingDowntimeJailDuration  string
	SlashingSlashFractionDowntime string
	SetupFailureCategory          NetworkFaultCategory

	// Export bootstrap material is deliberately private and is cleared before
	// Config is retained by the artifact store. Use StartFromExport instead of
	// setting these fields directly.
	exportedGenesis []byte
	validatorKey    []byte
}

// Network is a running Panacea network and its diagnostic artifact sink.
type Network struct {
	Chain *cosmos.CosmosChain

	interchain *interchaintest.Interchain
	artifacts  *artifactStore
	t          *testing.T
	txMu       sync.Mutex
}

// RecordTestPanic must be deferred immediately after Start succeeds. It makes
// the artifact outcome truthful before testing runs cleanup, then re-panics so
// the Go test still fails with the original panic value.
func (n *Network) RecordTestPanic() {
	recovered := recover()
	if recovered == nil {
		return
	}
	if n != nil && n.artifacts != nil {
		n.artifacts.recordTestPanic(recovered)
	}
	panic(recovered)
}

// Start creates and starts a real Panacea network through Interchaintest.
func Start(ctx context.Context, t *testing.T, cfg Config) (network *Network, retErr error) {
	t.Helper()
	if cfg.Image.Repository == "" && cfg.Image.Version == "" {
		cfg.Image = CurrentImage()
	}

	runID := cfg.RunID
	if runID == "" {
		var err error
		runID, err = newRunID()
		if err != nil {
			return nil, err
		}
	}

	spec, err := NewPanaceaChainSpec(runID, cfg.Image, topologyFromConfig(cfg))
	if err != nil {
		return nil, err
	}
	if err := configureExportBootstrap(ctx, spec, cfg); err != nil {
		return nil, err
	}

	artifactConfig := cfg
	artifactConfig.exportedGenesis = nil
	artifactConfig.validatorKey = nil
	store, err := newArtifactStore(t.Name(), runID, artifactConfig)
	if err != nil {
		return nil, err
	}
	t.Logf("e2e artifacts: %s", store.dir)

	defer func() {
		if recovered := recover(); recovered != nil {
			retErr = fmt.Errorf("interchaintest setup panic: %v", recovered)
			store.setBuildError(retErr)
		}
		if retErr != nil && cfg.SetupFailureCategory != "" {
			if categoryErr := store.recordNetworkFaultSetupFailure(cfg.SetupFailureCategory, "harness-start", retErr); categoryErr != nil {
				retErr = errors.Join(retErr, fmt.Errorf("record network fault setup category: %w", categoryErr))
			}
		}
		if retErr != nil && !store.cleanupIsRegistered() {
			store.setBuildError(retErr)
			if err := store.collect(true); err != nil {
				t.Logf("collect setup-failure artifacts: %v", err)
			}
		}
	}()

	cosmos.SetSDKConfig("panacea")
	factory := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{spec})
	chains, err := factory.Chains(runID)
	if err != nil {
		store.setBuildError(err)
		return nil, fmt.Errorf("create Panacea chain: %w", err)
	}
	chain, ok := chains[0].(*cosmos.CosmosChain)
	if !ok {
		return nil, fmt.Errorf("unexpected Interchaintest chain type %T", chains[0])
	}

	return startDockerNetwork(ctx, t, runID, chain, store)
}

func topologyFromConfig(cfg Config) Topology {
	return Topology{
		Validators:                    cfg.NumValidators,
		FullNodes:                     cfg.NumFullNodes,
		DBBackend:                     cfg.DBBackend,
		TimeoutCommit:                 cfg.TimeoutCommit,
		QueryGasLimit:                 cfg.QueryGasLimit,
		SnapshotInterval:              cfg.SnapshotInterval,
		SnapshotKeepRecent:            cfg.SnapshotKeepRecent,
		EnableTelemetry:               cfg.EnableTelemetry,
		DisableOptionalCPUFeatures:    cfg.DisableOptionalCPUFeatures,
		StakingUnbondingTime:          cfg.StakingUnbondingTime,
		SlashingSignedBlocksWindow:    cfg.SlashingSignedBlocksWindow,
		SlashingMinSignedPerWindow:    cfg.SlashingMinSignedPerWindow,
		SlashingDowntimeJailDuration:  cfg.SlashingDowntimeJailDuration,
		SlashingSlashFractionDowntime: cfg.SlashingSlashFractionDowntime,
	}
}

func startDockerNetwork(
	ctx context.Context,
	t *testing.T,
	runID string,
	chain *cosmos.CosmosChain,
	store *artifactStore,
) (*Network, error) {
	dockerClient, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		setupErr := fmt.Errorf("create Docker client: %w", err)
		store.setBuildError(setupErr)
		return nil, setupErr
	}
	setupOwned := true
	defer func() {
		if !setupOwned {
			return
		}
		if cleanupErr := cleanupDockerResources(dockerClient, runID); cleanupErr != nil {
			t.Logf("rollback Docker setup for %s: %v", runID, cleanupErr)
		}
	}()

	networkID, err := setupDockerNetwork(ctx, dockerClient, runID)
	if err != nil {
		setupErr := fmt.Errorf("set up bounded Docker network: %w", err)
		rollbackErr := cleanupDockerResources(dockerClient, runID)
		setupOwned = false
		combined := errors.Join(setupErr, rollbackErr)
		store.setBuildError(combined)
		return nil, combined
	}
	store.attach(chain, dockerClient, networkID)

	// Register the panic-safe owner immediately after network setup returns. This
	// closes the leak window before AddChain, Build, or any later setup step.
	var interchain *interchaintest.Interchain
	sequence := cleanupSequence{
		closeInterchain: func() error {
			if interchain == nil {
				return nil
			}
			return interchain.Close()
		},
		collectArtifacts: store.collect,
		cleanupDocker: func() error {
			return cleanupDockerResources(dockerClient, runID)
		},
		finalizeArtifacts: store.recordCleanup,
	}
	t.Cleanup(func() {
		if err := sequence.run(t.Failed()); err != nil {
			t.Errorf("finalize e2e run %s: %v", runID, err)
		}
	})
	// t.Cleanup now owns the Docker client and labeled resources. Transfer
	// ownership before bookkeeping so a bookkeeping panic cannot double-close
	// the client through both cleanup paths.
	setupOwned = false
	store.markCleanupRegistered()

	interchain = interchaintest.NewInterchain().AddChain(chain)
	network := &Network{
		Chain:      chain,
		interchain: interchain,
		artifacts:  store,
		t:          t,
	}
	err = interchain.Build(ctx, nil, interchaintest.InterchainBuildOptions{
		TestName:         runID,
		Client:           dockerClient,
		NetworkID:        networkID,
		SkipPathCreation: true,
	})
	if err != nil {
		store.setBuildError(err)
		return network, fmt.Errorf("build Panacea network: %w", err)
	}
	store.markRunning()
	return network, nil
}

func newRunID() (string, error) {
	random := make([]byte, 6)
	if _, err := cryptorand.Read(random); err != nil {
		return "", fmt.Errorf("generate run ID: %w", err)
	}
	return "run-" + hex.EncodeToString(random), nil
}

// WaitForHeight waits for committed blocks and reports the last observed
// height when the caller's bounded context expires.
func (n *Network) WaitForHeight(ctx context.Context, target int64) error {
	err := waitForHeight(ctx, target, n.Chain.Height)
	if err != nil {
		n.artifacts.recordFailure("wait-validator-height", err)
	}
	return err
}

// WaitForFullNode waits until the first full node catches up to target.
func (n *Network) WaitForFullNode(ctx context.Context, target int64) error {
	if len(n.Chain.FullNodes) == 0 {
		return errors.New("network has no full node")
	}
	err := waitForHeight(ctx, target, n.Chain.FullNodes[0].Height)
	if err != nil {
		n.artifacts.recordFailure("wait-full-node-height", err)
	}
	return err
}

// QueryFullNodeBalance verifies state through the full node's host gRPC
// connection. Interchaintest's chain-level GetBalance targets validator-0.
func (n *Network) QueryFullNodeBalance(ctx context.Context, address, denom string) (sdkmath.Int, error) {
	if len(n.Chain.FullNodes) == 0 {
		return sdkmath.Int{}, errors.New("network has no full node")
	}
	response, err := banktypes.NewQueryClient(n.Chain.FullNodes[0].GrpcConn).Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: address,
		Denom:   denom,
	})
	if err != nil {
		err = fmt.Errorf("full-node gRPC balance: %w", err)
		n.artifacts.recordFailure("full-node-grpc-balance", err)
		return sdkmath.Int{}, err
	}
	if response.Balance == nil {
		err = fmt.Errorf("full-node gRPC balance returned no coin for %s %s", address, denom)
		n.artifacts.recordFailure("full-node-grpc-balance", err)
		return sdkmath.Int{}, err
	}
	return response.Balance.Amount, nil
}

// FullNodeHostAddress returns an endpoint for the explicitly selected full
// node instead of Interchaintest's validator-backed chain-level endpoint.
func (n *Network) FullNodeHostAddress(ctx context.Context, portID string) (string, error) {
	if len(n.Chain.FullNodes) == 0 {
		return "", errors.New("network has no full node")
	}
	address, err := n.Chain.FullNodes[0].GetHostAddress(ctx, portID)
	if err != nil {
		err = fmt.Errorf("full-node host port %s: %w", portID, err)
		n.artifacts.recordFailure("full-node-host-address", err)
		return "", err
	}
	normalized, err := normalizeHostAddress(address)
	if err != nil {
		err = fmt.Errorf("normalize full-node host port %s: %w", portID, err)
		n.artifacts.recordFailure("full-node-host-address", err)
		return "", err
	}
	return normalized, nil
}

func normalizeHostAddress(address string) (string, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("endpoint must include scheme and host: %q", address)
	}
	switch parsed.Hostname() {
	case "0.0.0.0", "::", "":
		port := parsed.Port()
		if port == "" {
			return "", fmt.Errorf("wildcard endpoint has no port: %q", address)
		}
		parsed.Host = net.JoinHostPort("127.0.0.1", port)
	}
	return parsed.String(), nil
}

func waitForHeight(ctx context.Context, target int64, height func(context.Context) (int64, error)) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var (
		lastHeight int64
		lastErr    error
	)
	for {
		observed, err := height(ctx)
		if err == nil {
			lastHeight = observed
			if observed >= target {
				return nil
			}
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for height %d: last height=%d last error=%v: %w", target, lastHeight, lastErr, ctx.Err())
		case <-ticker.C:
		}
	}
}

// RecordCommittedTx polls the explicit full-node CLI boundary until the
// transaction is visible, validates its hash/height/DeliverTx code, and records
// every attempt. A successful return is also a state-synchronization barrier
// for subsequent full-node queries.
func (n *Network) RecordCommittedTx(ctx context.Context, txHash, chainType string) error {
	if len(n.Chain.FullNodes) == 0 {
		return errors.New("network has no full node")
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var lastQueryErr error
	for attempt := 1; ; attempt++ {
		result, stderr, queryErr := n.Chain.FullNodes[0].ExecQuery(ctx, "tx", txHash)
		attemptRecord := map[string]any{
			"recorded_at": time.Now().UTC(),
			"attempt":     attempt,
			"tx_hash":     txHash,
			"stdout":      boundedArtifactText(result),
			"stderr":      boundedArtifactText(stderr),
			"error":       errorString(queryErr),
		}
		if err := n.artifacts.appendJSONLine("tx/query-attempts.jsonl", attemptRecord); err != nil {
			n.artifacts.recordFailure("record-tx-query-attempt", err)
			return err
		}

		if queryErr == nil {
			height, err := validateCommittedTx(result, txHash)
			if err != nil {
				err = fmt.Errorf("verify committed tx %s: %w", txHash, err)
				n.artifacts.recordFailure("verify-committed-tx", err)
				return err
			}
			record := map[string]any{
				"recorded_at": time.Now().UTC(),
				"chain_type":  chainType,
				"tx_hash":     txHash,
				"height":      height,
				"result":      json.RawMessage(result),
			}
			if err := n.artifacts.appendJSONLine("tx/committed-results.jsonl", record); err != nil {
				n.artifacts.recordFailure("record-committed-tx", err)
				return err
			}
			return nil
		}
		lastQueryErr = fmt.Errorf("query committed tx %s: %w: %s", txHash, queryErr, strings.TrimSpace(string(stderr)))

		select {
		case <-ctx.Done():
			err := fmt.Errorf("wait for committed tx %s: last error=%v: %w", txHash, lastQueryErr, ctx.Err())
			n.artifacts.recordFailure("query-committed-tx", err)
			return err
		case <-ticker.C:
		}
	}
}

func validateCommittedTx(result []byte, expectedHash string) (int64, error) {
	var response struct {
		Height json.RawMessage `json:"height"`
		Hash   string          `json:"txhash"`
		Code   json.RawMessage `json:"code"`
	}
	if err := json.Unmarshal(result, &response); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}
	if len(response.Code) == 0 || string(response.Code) == "null" {
		return 0, errors.New("response is missing code")
	}
	code, err := parseJSONInt(response.Code)
	if err != nil {
		return 0, fmt.Errorf("decode code: %w", err)
	}
	if code != 0 {
		return 0, fmt.Errorf("transaction has non-zero code %d", code)
	}
	if response.Hash == "" || !strings.EqualFold(response.Hash, expectedHash) {
		return 0, fmt.Errorf("response hash %q does not match %q", response.Hash, expectedHash)
	}
	height, err := parseJSONInt(response.Height)
	if err != nil {
		return 0, fmt.Errorf("decode committed height %q: %w", response.Height, err)
	}
	if height < 1 {
		return 0, fmt.Errorf("committed height must be positive, got %d", height)
	}
	return height, nil
}

func parseJSONInt(raw json.RawMessage) (int64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, errors.New("missing integer")
	}
	value := string(raw)
	if raw[0] == '"' {
		if err := json.Unmarshal(raw, &value); err != nil {
			return 0, err
		}
	}
	return strconv.ParseInt(value, 10, 64)
}

func boundedArtifactText(contents []byte) string {
	const maximum = 1 << 20
	if len(contents) <= maximum {
		return string(contents)
	}
	return string(contents[:maximum]) + "\n[truncated at 1 MiB]"
}
