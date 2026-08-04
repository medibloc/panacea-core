package harness

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	dockerclient "github.com/docker/docker/client"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/relayer"
	"github.com/strangelove-ventures/interchaintest/v8/relayer/hermes"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"go.uber.org/zap/zaptest"
)

const (
	ibcTopologyPath = "panacea-osmosis"

	// Interchaintest v8.8.1 asks Docker for a 30-second graceful stop when it
	// stops the relayer container. The caller context must outlive that Docker
	// deadline so Docker can force-stop the process, return the response, and
	// leave enough time for Interchaintest to capture logs and remove it.
	interchaintestRelayerStopGrace = 30 * time.Second
	ibcRelayerCleanupTimeout       = interchaintestRelayerStopGrace + 15*time.Second
)

// IBCTopologyConfig controls the independent two-chain topology smoke. The
// counterparty and relayer releases are intentionally not configurable: this
// slice exists to prove the exact Osmosis/Hermes compatibility contract.
type IBCTopologyConfig struct {
	PanaceaImage ImageRef
	RunID        string
	ArtifactRoot string
}

type ibcTopologyPlan struct {
	runID          string
	path           string
	panaceaSpec    *interchaintest.ChainSpec
	osmosisSpec    *interchaintest.ChainSpec
	hermesImage    ibc.DockerImage
	descriptor     IBCTopologyDescriptor
	artifactConfig Config
	preflight      OsmosisMainnetPreflightConfig
	panaceaBinary  IBCBinaryVersionContract
}

// IBCTopology is a running Panacea/Osmosis pair whose relayer is configured
// but whose IBC path has not yet been created. Keeping path creation explicit
// lets the upgrade continuity suite prove that it never replaces the original
// clients, connection, or channel after the upgrade.
type IBCTopology struct {
	Panacea *cosmos.CosmosChain
	Osmosis *cosmos.CosmosChain
	Relayer ibc.Relayer
	Path    string

	hermesVersion  string
	hermesIdentity string
	hermesRuntime  HermesRuntimeIdentityEvidence
	artifacts      *ibcTopologyArtifacts
	hermes         *hermes.Relayer
	execReporter   ibc.RelayerExecReporter

	lifecycleMu                 sync.Mutex
	channel                     *IBCChannelHandshake
	channelAttempted            bool
	relayerStarted              bool
	preUpgradeTransferAttempted bool
	preUpgradeTransferComplete  bool
	preUpgradeTransferEvidence  *IBCPreUpgradeTransferEvidence
	upgradeContinuityAttempted  bool
	inFlightCheckpoint          *IBCInFlightPacketCheckpoint
	inFlightTx                  *ibc.Tx
	panaceaUpgradeStep          *IBCPanaceaUpgradeStepEvidence
	postUpgradeBeforeRelay      *IBCLinkStateSnapshot
	upgradeContinuityComplete   bool
	mainnetPreflight            OsmosisMainnetPreflightEvidence
	panaceaPreUpgradeBinary     IBCChainBinaryEvidence
	panaceaPostUpgradeBinary    IBCChainBinaryEvidence
	osmosisPreUpgradeBinary     IBCChainBinaryEvidence
	osmosisPostUpgradeBinary    IBCChainBinaryEvidence
	osmosisGenesisInitial       IBCGenesisChecksumSnapshot
	osmosisGenesisImmutability  IBCGenesisImmutabilityEvidence
}

// RecordTestPanic delegates the shared panic artifact contract for IBC live
// tests and re-panics with the original value.
func (n *IBCTopology) RecordTestPanic() {
	recovered := recover()
	if recovered == nil {
		return
	}
	if n != nil && n.artifacts != nil && n.artifacts.base != nil {
		n.artifacts.base.recordTestPanic(recovered)
	}
	panic(recovered)
}

type runOwnedTestName string

func (n runOwnedTestName) Name() string { return string(n) }

func newIBCTopologyPlan(cfg IBCTopologyConfig) (*ibcTopologyPlan, error) {
	runID := cfg.RunID
	if runID == "" {
		var err error
		runID, err = newRunID()
		if err != nil {
			return nil, err
		}
	}

	panaceaImage := cfg.PanaceaImage
	if panaceaImage.Repository == "" && panaceaImage.Version == "" {
		panaceaImage = CurrentImage()
	}
	panaceaSpec, err := NewPanaceaChainSpec(runID, panaceaImage, Topology{
		Validators:    1,
		FullNodes:     1,
		TimeoutCommit: "1s",
	})
	if err != nil {
		return nil, err
	}
	osmosisSpec, err := NewOsmosisChainSpec(runID)
	if err != nil {
		return nil, err
	}
	panaceaBinary, err := panaceaBinaryContractForImage(panaceaImage)
	if err != nil {
		return nil, err
	}
	preflight, err := resolveOsmosisMainnetPreflightConfig()
	if err != nil {
		return nil, err
	}

	descriptor := IBCTopologyDescriptor{
		Path:              ibcTopologyPath,
		PanaceaChainID:    panaceaSpec.ChainConfig.ChainID,
		OsmosisChainID:    osmosisSpec.ChainConfig.ChainID,
		PanaceaValidators: *panaceaSpec.NumValidators,
		PanaceaFullNodes:  *panaceaSpec.NumFullNodes,
		OsmosisValidators: *osmosisSpec.NumValidators,
		OsmosisFullNodes:  *osmosisSpec.NumFullNodes,
		SkipPathCreation:  true,
	}
	artifactConfig := Config{
		Image:         panaceaImage,
		NumValidators: descriptor.PanaceaValidators,
		NumFullNodes:  descriptor.PanaceaFullNodes,
		RunID:         runID,
		ArtifactRoot:  cfg.ArtifactRoot,
		TimeoutCommit: "1s",
	}

	return &ibcTopologyPlan{
		runID:          runID,
		path:           ibcTopologyPath,
		panaceaSpec:    panaceaSpec,
		osmosisSpec:    osmosisSpec,
		hermesImage:    PinnedHermesImage(),
		descriptor:     descriptor,
		artifactConfig: artifactConfig,
		preflight:      preflight,
		panaceaBinary:  planPanaceaBinaryCopy(panaceaBinary),
	}, nil
}

func planPanaceaBinaryCopy(contract IBCBinaryVersionContract) IBCBinaryVersionContract {
	contract.Dependencies = append([]IBCDependencyContract(nil), contract.Dependencies...)
	return contract
}

// StartIBCTopology builds the two real chains and configures the exact pinned
// Hermes image against both Docker-network endpoints. It intentionally stops
// before client/connection/channel creation; this is the independent topology
// and provenance smoke used as the foundation for the full upgrade scenario.
func StartIBCTopology(
	ctx context.Context,
	t *testing.T,
	cfg IBCTopologyConfig,
) (topology *IBCTopology, retErr error) {
	t.Helper()
	if ctx == nil {
		return nil, errors.New("IBC topology context is required")
	}

	plan, err := newIBCTopologyPlan(cfg)
	if err != nil {
		return nil, err
	}
	store, err := newArtifactStore(t.Name(), plan.runID, plan.artifactConfig)
	if err != nil {
		return nil, err
	}
	t.Logf("IBC topology artifacts: %s", store.dir)

	var artifacts *ibcTopologyArtifacts
	defer func() {
		if recovered := recover(); recovered != nil {
			retErr = fmt.Errorf("IBC topology setup panic: %v", recovered)
			store.setBuildError(retErr)
		}
		if retErr == nil || store.cleanupIsRegistered() {
			return
		}
		store.setBuildError(retErr)
		if artifacts != nil {
			if closeErr := artifacts.closeReporter(); closeErr != nil {
				t.Logf("close setup-failure Hermes reporter: %v", closeErr)
			}
		}
		if collectErr := store.collect(true); collectErr != nil {
			t.Logf("collect IBC setup-failure artifacts: %v", collectErr)
		}
	}()

	artifacts, err = newIBCTopologyArtifacts(store, plan.descriptor)
	if err != nil {
		return nil, err
	}
	preflightClient := &http.Client{
		Timeout: plan.preflight.Timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	preflight, preflightErr := runOsmosisMainnetPreflight(ctx, preflightClient, plan.preflight, time.Now().UTC())
	if writeErr := store.writeJSON(osmosisMainnetPreflightArtifactPath, preflight); writeErr != nil {
		return nil, errors.Join(preflightErr, fmt.Errorf("write Osmosis mainnet preflight artifact: %w", writeErr))
	}
	if preflightErr != nil {
		store.recordFailure("ibc-osmosis-mainnet-preflight", preflightErr)
		return nil, fmt.Errorf("Osmosis mainnet preflight: %w", preflightErr)
	}

	cosmos.SetSDKConfig("panacea")
	logger := zaptest.NewLogger(t)
	chainFactory := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
		plan.panaceaSpec,
		plan.osmosisSpec,
	})
	chains, err := chainFactory.Chains(plan.runID)
	if err != nil {
		return nil, fmt.Errorf("create IBC topology chains: %w", err)
	}
	if len(chains) != 2 {
		return nil, fmt.Errorf("create IBC topology chains: got %d chains, want 2", len(chains))
	}
	panacea, ok := chains[0].(*cosmos.CosmosChain)
	if !ok {
		return nil, fmt.Errorf("unexpected Panacea Interchaintest chain type %T", chains[0])
	}
	osmosis, ok := chains[1].(*cosmos.CosmosChain)
	if !ok {
		return nil, fmt.Errorf("unexpected Osmosis Interchaintest chain type %T", chains[1])
	}

	dockerClient, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("create Docker client: %w", err)
	}
	setupOwned := true
	defer func() {
		if !setupOwned {
			return
		}
		if cleanupErr := cleanupDockerResources(dockerClient, plan.runID); cleanupErr != nil {
			t.Logf("rollback IBC Docker setup for %s: %v", plan.runID, cleanupErr)
		}
	}()

	networkID, err := setupDockerNetwork(ctx, dockerClient, plan.runID)
	if err != nil {
		setupErr := fmt.Errorf("set up bounded IBC Docker network: %w", err)
		rollbackErr := cleanupDockerResources(dockerClient, plan.runID)
		setupOwned = false
		combined := errors.Join(setupErr, rollbackErr)
		store.setBuildError(combined)
		return nil, combined
	}
	store.attach(panacea, dockerClient, networkID)

	execReporter := artifacts.reporter.RelayerExecReporter(t)
	var (
		interchain      *interchaintest.Interchain
		hermesRelayer   *hermes.Relayer
		interchainBuilt bool
	)
	sequence := cleanupSequence{
		closeInterchain: func() error {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), ibcRelayerCleanupTimeout)
			defer cancel()
			stopErr := runCleanupStep("Hermes stop", func() error {
				if hermesRelayer == nil {
					return nil
				}
				return hermesRelayer.StopRelayer(cleanupCtx, execReporter)
			})
			closeErr := runCleanupStep("IBC interchain close", func() error {
				if interchain == nil || !interchainBuilt {
					return nil
				}
				return interchain.Close()
			})
			reporterErr := runCleanupStep("Hermes reporter close", artifacts.closeReporter)
			return errors.Join(stopErr, closeErr, reporterErr)
		},
		collectArtifacts: func(failed bool) error {
			return artifacts.collect(failed, osmosis, dockerClient)
		},
		cleanupDocker: func() error {
			return cleanupDockerResources(dockerClient, plan.runID)
		},
		finalizeArtifacts: store.recordCleanup,
	}
	t.Cleanup(func() {
		if err := sequence.run(t.Failed()); err != nil {
			t.Errorf("finalize IBC topology run %s: %v", plan.runID, err)
		}
	})
	setupOwned = false
	store.markCleanupRegistered()

	relayerFactory := interchaintest.NewBuiltinRelayerFactory(
		ibc.Hermes,
		logger,
		relayer.CustomDockerImage(
			plan.hermesImage.Repository,
			plan.hermesImage.Version,
			plan.hermesImage.UIDGID,
		),
	)
	builtRelayer := relayerFactory.Build(runOwnedTestName(plan.runID), dockerClient, networkID)
	hermesRelayer, ok = builtRelayer.(*hermes.Relayer)
	if !ok {
		err := fmt.Errorf("unexpected Hermes relayer type %T", builtRelayer)
		store.setBuildError(err)
		return nil, err
	}

	interchain = interchaintest.NewInterchain().
		WithLog(logger).
		AddChain(panacea).
		AddChain(osmosis).
		AddRelayer(hermesRelayer, "hermes").
		AddLink(interchaintest.InterchainLink{
			Chain1:  panacea,
			Chain2:  osmosis,
			Relayer: hermesRelayer,
			Path:    plan.path,
		})
	topology = &IBCTopology{
		Panacea:          panacea,
		Osmosis:          osmosis,
		Relayer:          hermesRelayer,
		Path:             plan.path,
		artifacts:        artifacts,
		hermes:           hermesRelayer,
		execReporter:     execReporter,
		mainnetPreflight: preflight,
	}
	interchainBuilt = true
	if err := interchain.Build(ctx, execReporter, interchaintest.InterchainBuildOptions{
		TestName:         plan.runID,
		Client:           dockerClient,
		NetworkID:        networkID,
		SkipPathCreation: true,
	}); err != nil {
		store.setBuildError(err)
		return topology, fmt.Errorf("build Panacea/Osmosis topology: %w", err)
	}
	panaceaBinary, panaceaBinaryErr := captureIBCChainBinaryEvidence(
		ctx,
		panacea,
		"pre-upgrade",
		plan.panaceaBinary,
		store,
		"ibc/chains/panacea/pre-upgrade",
	)
	osmosisBinary, osmosisBinaryErr := captureIBCChainBinaryEvidence(
		ctx,
		osmosis,
		"pre-upgrade",
		pinnedOsmosisBinaryContract(),
		store,
		"ibc/chains/osmosis/pre-upgrade",
	)
	osmosisGenesis, osmosisGenesisErr := captureIBCGenesisChecksumSnapshot(
		ctx,
		osmosis,
		"pre-upgrade",
		store,
		"ibc/chains/osmosis",
		"",
	)
	topology.panaceaPreUpgradeBinary = panaceaBinary
	topology.osmosisPreUpgradeBinary = osmosisBinary
	topology.osmosisGenesisInitial = osmosisGenesis
	osmosisIdentityArtifactErr := store.writeJSON("ibc/chains/osmosis/identity.json", osmosisBinary)
	if identityErr := errors.Join(panaceaBinaryErr, osmosisBinaryErr, osmosisGenesisErr, osmosisIdentityArtifactErr); identityErr != nil {
		store.setBuildError(identityErr)
		return topology, fmt.Errorf("validate initial IBC chain identities: %w", identityErr)
	}
	if err := topology.fundRuntimeHermesWallets(ctx); err != nil {
		store.setBuildError(err)
		return topology, fmt.Errorf("fund runtime Hermes wallets: %w", err)
	}

	identity, err := artifacts.recordHermesRuntimeEvidence(ctx, hermesRelayer, execReporter, dockerClient)
	if err != nil {
		store.setBuildError(err)
		return topology, fmt.Errorf("record Hermes topology evidence: %w", err)
	}
	topology.hermesVersion = identity.VersionOutput
	topology.hermesIdentity = identity.ReleaseIdentifier
	topology.hermesRuntime = identity
	store.markRunning()
	return topology, nil
}

// Hermes derives Cosmos keys with its own keyring implementation. Panacea's
// coin type differs from the Cosmos default, so the runtime address returned
// by Hermes can differ from the address Interchaintest placed in genesis from
// the same mnemonic. Fund and verify the actual restored runtime addresses
// before any client transaction; never persist the mnemonic.
func (n *IBCTopology) fundRuntimeHermesWallets(ctx context.Context) error {
	if n == nil || n.Relayer == nil || n.Panacea == nil || n.Osmosis == nil || n.artifacts == nil || n.artifacts.base == nil {
		return errors.New("IBC topology is not initialized for Hermes wallet funding")
	}
	amount := sdkmath.NewInt(10_000_000_000)
	type fundedWallet struct {
		ChainID string `json:"chain_id"`
		Address string `json:"address"`
		Denom   string `json:"denom"`
		Before  string `json:"before"`
		Funded  string `json:"funded"`
		After   string `json:"after"`
	}
	evidence := make([]fundedWallet, 0, 2)
	fundedAny := false
	for _, chain := range []*cosmos.CosmosChain{n.Panacea, n.Osmosis} {
		chainID := chain.Config().ChainID
		wallet, ok := n.Relayer.GetWallet(chainID)
		if !ok || wallet == nil || strings.TrimSpace(wallet.FormattedAddress()) == "" {
			return fmt.Errorf("Hermes runtime wallet for %s is unavailable", chainID)
		}
		before, err := chain.GetBalance(ctx, wallet.FormattedAddress(), chain.Config().Denom)
		if err != nil {
			return fmt.Errorf("query Hermes runtime wallet before funding on %s: %w", chainID, err)
		}
		funded := sdkmath.ZeroInt()
		if before.LT(amount) {
			funded = amount.Sub(before)
			if err := chain.SendFunds(ctx, interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
				Address: wallet.FormattedAddress(),
				Denom:   chain.Config().Denom,
				Amount:  funded,
			}); err != nil {
				return fmt.Errorf("fund Hermes runtime wallet on %s: %w", chainID, err)
			}
			fundedAny = true
		}
		evidence = append(evidence, fundedWallet{
			ChainID: chainID,
			Address: wallet.FormattedAddress(),
			Denom:   chain.Config().Denom,
			Before:  before.String(),
			Funded:  funded.String(),
		})
	}
	if fundedAny {
		if err := testutil.WaitForBlocks(ctx, 2, n.Panacea, n.Osmosis); err != nil {
			return fmt.Errorf("wait for Hermes runtime wallet funding: %w", err)
		}
	}
	for index, funded := range evidence {
		chain := n.Panacea
		if funded.ChainID == n.Osmosis.Config().ChainID {
			chain = n.Osmosis
		}
		balance, err := chain.GetBalance(ctx, funded.Address, funded.Denom)
		if err != nil {
			return fmt.Errorf("query Hermes runtime wallet on %s: %w", funded.ChainID, err)
		}
		if balance.LT(amount) {
			return fmt.Errorf("Hermes runtime wallet on %s has %s%s, want at least %s%s", funded.ChainID, balance, funded.Denom, amount, funded.Denom)
		}
		evidence[index].After = balance.String()
	}
	return n.artifacts.base.writeJSON("ibc/hermes/runtime-wallet-funding.json", evidence)
}

// WaitForHeight proves that both independent chains are committing blocks.
func (n *IBCTopology) WaitForHeight(ctx context.Context, target int64) error {
	if n == nil || n.Panacea == nil || n.Osmosis == nil {
		return errors.New("IBC topology is not initialized")
	}
	type chainHeight struct {
		name   string
		height func(context.Context) (int64, error)
	}
	result := make(chan error, 2)
	for _, chain := range []chainHeight{
		{name: "panacea", height: n.Panacea.Height},
		{name: "osmosis", height: n.Osmosis.Height},
	} {
		chain := chain
		go func() {
			if err := waitForHeight(ctx, target, chain.height); err != nil {
				result <- fmt.Errorf("wait for %s height %d: %w", chain.name, target, err)
				return
			}
			result <- nil
		}()
	}
	var waitErrors []error
	for range 2 {
		if err := <-result; err != nil {
			waitErrors = append(waitErrors, err)
		}
	}
	joined := errors.Join(waitErrors...)
	if joined != nil && n.artifacts != nil {
		n.artifacts.base.recordFailure("ibc-topology-height", joined)
	}
	return joined
}

// HermesVersion returns the runtime output captured from the pinned container.
func (n *IBCTopology) HermesVersion() string {
	if n == nil {
		return ""
	}
	return strings.TrimSpace(n.hermesVersion)
}

// HermesIdentity returns the release identity formed from the exact runtime
// version and the pinned upstream source commit.
func (n *IBCTopology) HermesIdentity() string {
	if n == nil {
		return ""
	}
	return strings.TrimSpace(n.hermesIdentity)
}

// ArtifactDir returns the run-owned diagnostic directory.
func (n *IBCTopology) ArtifactDir() string {
	if n == nil || n.artifacts == nil || n.artifacts.base == nil {
		return ""
	}
	return n.artifacts.base.dir
}

// RecordUpgradeCoverageMatrix writes the same coverage contract used by the
// connected single-chain upgrade lane into this topology's Panacea run root.
func (n *IBCTopology) RecordUpgradeCoverageMatrix(matrix UpgradeCoverageMatrix) error {
	if n == nil {
		return errors.New("IBC topology is unavailable")
	}
	return n.panaceaNetworkView().RecordUpgradeCoverageMatrix(matrix)
}
