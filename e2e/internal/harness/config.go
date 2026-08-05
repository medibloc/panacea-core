package harness

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

const (
	currentImageRepository = "panacea-e2e-current"
	v221ImageRepository    = "panacea-e2e-v2.2.1"
	localImageVersion      = "local"
)

// CometBFT caps chain IDs at 50 bytes. "panacea-" consumes eight bytes, so
// an ASCII run ID may contain at most 42 bytes.
var runIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,41}$`)

// ImageRef identifies an already-built panacead image.
type ImageRef struct {
	Repository string
	Version    string
}

// CurrentImage returns the local image built from the current worktree.
func CurrentImage() ImageRef {
	repository := os.Getenv("PANACEA_E2E_IMAGE_REPOSITORY")
	if repository == "" {
		repository = currentImageRepository
	}
	version := os.Getenv("PANACEA_E2E_IMAGE_VERSION")
	if version == "" {
		version = localImageVersion
	}
	return ImageRef{Repository: repository, Version: version}
}

// V221Image returns the separately built v2.2.1 compatibility image.
func V221Image() ImageRef {
	repository := os.Getenv("PANACEA_E2E_V221_IMAGE_REPOSITORY")
	if repository == "" {
		repository = v221ImageRepository
	}
	version := os.Getenv("PANACEA_E2E_V221_IMAGE_VERSION")
	if version == "" {
		version = localImageVersion
	}
	return ImageRef{Repository: repository, Version: version}
}

// Topology describes how many independent validator and full-node processes
// Interchaintest must create.
type Topology struct {
	Validators         int
	FullNodes          int
	DBBackend          string
	TimeoutCommit      string
	QueryGasLimit      uint64
	SnapshotInterval   uint64
	SnapshotKeepRecent uint32
	EnableTelemetry    bool

	// Test-only genesis overrides. Their zero values leave source genesis intact.
	StakingUnbondingTime          string
	SlashingSignedBlocksWindow    int64
	SlashingMinSignedPerWindow    string
	SlashingDowntimeJailDuration  string
	SlashingSlashFractionDowntime string
}

// NewPanaceaChainSpec returns the thin Panacea adapter used by every suite.
func NewPanaceaChainSpec(runID string, image ImageRef, topology Topology) (*interchaintest.ChainSpec, error) {
	if !runIDPattern.MatchString(runID) {
		return nil, fmt.Errorf("run ID %q must match %s", runID, runIDPattern)
	}
	if strings.TrimSpace(image.Repository) == "" || strings.TrimSpace(image.Version) == "" {
		return nil, errors.New("image repository and version are required")
	}
	if topology.Validators < 1 {
		return nil, errors.New("at least one validator is required")
	}
	if topology.FullNodes < 0 {
		return nil, errors.New("full-node count cannot be negative")
	}

	chainName := "panacea-" + runID
	coinDecimals := int64(6)
	queryGasLimit := topology.QueryGasLimit
	if queryGasLimit == 0 {
		queryGasLimit = 10_000_000
	}
	appOverrides := testutil.Toml{
		"query-gas-limit": queryGasLimit,
	}
	if topology.SnapshotInterval > 0 {
		appOverrides["state-sync"] = testutil.Toml{
			"snapshot-interval":    topology.SnapshotInterval,
			"snapshot-keep-recent": topology.SnapshotKeepRecent,
		}
	}
	if topology.EnableTelemetry {
		appOverrides["telemetry"] = testutil.Toml{
			"enabled":                   true,
			"prometheus-retention-time": int64(60),
		}
	}
	dbBackend := strings.TrimSpace(topology.DBBackend)
	if dbBackend == "" {
		dbBackend = "goleveldb"
	}
	cometOverrides := testutil.Toml{
		"db_backend": dbBackend,
	}
	if strings.TrimSpace(topology.TimeoutCommit) != "" {
		cometOverrides["consensus"] = testutil.Toml{
			"timeout_commit": topology.TimeoutCommit,
		}
	}
	genesisOverrides := []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.gov.params.voting_period", "20s"),
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", "20s"),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.denom", "umed"),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.amount", "1"),
	}
	if value := strings.TrimSpace(topology.StakingUnbondingTime); value != "" {
		genesisOverrides = append(genesisOverrides,
			cosmos.NewGenesisKV("app_state.staking.params.unbonding_time", value),
		)
	}
	if topology.SlashingSignedBlocksWindow != 0 {
		genesisOverrides = append(genesisOverrides,
			cosmos.NewGenesisKV(
				"app_state.slashing.params.signed_blocks_window",
				strconv.FormatInt(topology.SlashingSignedBlocksWindow, 10),
			),
		)
	}
	for _, override := range []struct {
		path  string
		value string
	}{
		{path: "app_state.slashing.params.min_signed_per_window", value: topology.SlashingMinSignedPerWindow},
		{path: "app_state.slashing.params.downtime_jail_duration", value: topology.SlashingDowntimeJailDuration},
		{path: "app_state.slashing.params.slash_fraction_downtime", value: topology.SlashingSlashFractionDowntime},
	} {
		if value := strings.TrimSpace(override.value); value != "" {
			genesisOverrides = append(genesisOverrides, cosmos.NewGenesisKV(override.path, value))
		}
	}

	return &interchaintest.ChainSpec{
		Name:          chainName,
		ChainName:     chainName,
		Version:       image.Version,
		NumValidators: &topology.Validators,
		NumFullNodes:  &topology.FullNodes,
		ChainConfig: ibc.ChainConfig{
			Type:           ibc.Cosmos,
			Name:           chainName,
			ChainID:        chainName,
			Bin:            "panacead",
			Bech32Prefix:   "panacea",
			Denom:          "umed",
			CoinType:       "371",
			CoinDecimals:   &coinDecimals,
			GasPrices:      "5umed",
			GasAdjustment:  1.3,
			Gas:            "auto",
			TrustingPeriod: "336h",
			Images: []ibc.DockerImage{{
				Repository: image.Repository,
				Version:    image.Version,
				UIDGID:     "0:0",
			}},
			UsingChainIDFlagCLI: false,
			ModifyGenesis:       cosmos.ModifyGenesis(genesisOverrides),
			ConfigFileOverrides: map[string]any{
				"config/app.toml":    appOverrides,
				"config/config.toml": cometOverrides,
			},
		},
	}, nil
}
