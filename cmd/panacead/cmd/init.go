package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/cli"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"time"

	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/spf13/cobra"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagSeed defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"

	// FlagDefaultBondDenom defines the default denom to use in the genesis file.
	FlagDefaultBondDenom = "default-denom"

	// FlagStakingBondDenom defines a flag to specify the staking token in the genesis file.
	FlagStakingBondDenom = "staking-bond-denom"
	blockTimeSec         = 6                                  // 5s of timeout_commit + 1s
	unbondingPeriod      = 60 * 60 * 24 * 7 * 3 * time.Second // three weeks
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

// InitCmd wraps the genutilcli.InitCmd to inject specific parameters for Panacea.
// It reads the default genesis.json, modifies and exports it.
// Reference: https://github.com/public-awesome/stargaze/blob/b92bf9847559b9c7f4ac08576d056d3d00efe12c/cmd/starsd/cmd/init.go
func InitCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			switch {
			case chainID != "":
			case clientCtx.ChainID != "":
				chainID = clientCtx.ChainID
			default:
				chainID = fmt.Sprintf("test-chain-%v", tmrand.Str(6))
			}

			// Get bip39 mnemonic
			var mnemonic string
			recover, _ := cmd.Flags().GetBool(FlagRecover)
			if recover {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				value, err := input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}

				mnemonic = value
				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid mnemonic")
				}
			}

			// Get initial height
			initHeight, _ := cmd.Flags().GetInt64(flags.FlagInitHeight)
			if initHeight < 1 {
				initHeight = 1
			}

			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			genFile := config.GenesisFile()
			overwrite, _ := cmd.Flags().GetBool(FlagOverwrite)
			defaultDenom, _ := cmd.Flags().GetString(FlagDefaultBondDenom)

			// use os.Stat to check if the file exists
			_, err = os.Stat(genFile)
			if !overwrite && !os.IsNotExist(err) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			// Overwrites the SDK default denom for side-effects
			if defaultDenom != "" {
				sdk.DefaultBondDenom = defaultDenom
			}
			appGenState := mbm.DefaultGenesis(cdc)
			genDoc := &types.GenesisDoc{}
			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				genDoc, err = types.GenesisDocFromFile(genFile)
				if err != nil {
					return errors.Wrap(err, "Failed to read genesis doc from file")
				}
			}

			genDoc.ChainID = chainID
			genDoc.Validators = nil
			genDoc.InitialHeight = initHeight
			genDoc.ConsensusParams = &types.ConsensusParams{
				Block:     types.DefaultBlockParams(),
				Evidence:  types.DefaultEvidenceParams(),
				Validator: types.DefaultValidatorParams(),
				Version:   types.DefaultVersionParams(),
			}

			appState, err := overrideGenesis(cdc, genDoc, appGenState)
			if err != nil {
				return errors.Wrap(err, "Failed to marshal default genesis state")
			}

			genDoc.AppState = appState

			if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
				return errors.Wrap(err, "Failed to export genesis file")
			}

			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)
			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			return displayInfo(toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(FlagDefaultBondDenom, "", "genesis file default denomination, if left blank default value is 'stake'")
	cmd.Flags().Int64(flags.FlagInitHeight, 1, "specify the initial block height at genesis")

	return cmd
}

// overrideGenesis overrides some parameters in the genesis doc to the panacea-specific values.
func overrideGenesis(cdc codec.JSONCodec, genDoc *types.GenesisDoc, appState map[string]json.RawMessage) (json.RawMessage, error) {
	var stakingGenState stakingtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[stakingtypes.ModuleName], &stakingGenState); err != nil {
		return nil, err
	}
	stakingGenState.Params.UnbondingTime = unbondingPeriod
	stakingGenState.Params.MaxValidators = 50
	stakingGenState.Params.BondDenom = assets.MicroMedDenom
	stakingGenState.Params.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
	appState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(&stakingGenState)

	var mintGenState minttypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[minttypes.ModuleName], &mintGenState); err != nil {
		return nil, err
	}
	mintGenState.Minter = minttypes.InitialMinter(sdk.NewDecWithPrec(7, 2)) // 7% inflation
	mintGenState.Params.MintDenom = assets.MicroMedDenom
	mintGenState.Params.InflationRateChange = sdk.NewDecWithPrec(3, 2) // 3%
	mintGenState.Params.InflationMin = sdk.NewDecWithPrec(7, 2)        // 7%
	mintGenState.Params.InflationMax = sdk.NewDecWithPrec(10, 2)       // 10%
	mintGenState.Params.BlocksPerYear = uint64(60*60*24*365) / uint64(blockTimeSec)
	appState[minttypes.ModuleName] = cdc.MustMarshalJSON(&mintGenState)

	var distrGenState distrtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[distrtypes.ModuleName], &distrGenState); err != nil {
		return nil, err
	}
	distrGenState.Params.CommunityTax = sdk.ZeroDec()
	appState[distrtypes.ModuleName] = cdc.MustMarshalJSON(&distrGenState)

	var govGenState govv1types.GenesisState
	if err := cdc.UnmarshalJSON(appState[govtypes.ModuleName], &govGenState); err != nil {
		return nil, err
	}
	minDepositTokens := sdk.TokensFromConsensusPower(100000, sdk.DefaultPowerReduction) // 100,000 MED
	govGenState.Params.MinDeposit = sdk.Coins{sdk.NewCoin(assets.MicroMedDenom, minDepositTokens)}
	maxDepositPeriod := 60 * 60 * 24 * 14 * time.Second // 14 days
	govGenState.Params.MaxDepositPeriod = &maxDepositPeriod
	votingPeriod := 60 * 60 * 24 * 3 * time.Second // 3 days (shortened voting period: https://www.mintscan.io/medibloc/proposals/5)
	govGenState.Params.VotingPeriod = &votingPeriod
	appState[govtypes.ModuleName] = cdc.MustMarshalJSON(&govGenState)

	var crisisGenState crisistypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[crisistypes.ModuleName], &crisisGenState); err != nil {
		return nil, err
	}
	crisisGenState.ConstantFee = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000000000000)) // Spend 1,000,000 MED for invariants check
	appState[crisistypes.ModuleName] = cdc.MustMarshalJSON(&crisisGenState)

	var slashingGenState slashingtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[slashingtypes.ModuleName], &slashingGenState); err != nil {
		return nil, err
	}
	slashingGenState.Params.SignedBlocksWindow = 10000
	slashingGenState.Params.MinSignedPerWindow = sdk.NewDecWithPrec(5, 2)
	slashingGenState.Params.SlashFractionDoubleSign = sdk.NewDecWithPrec(5, 2) // 5%
	slashingGenState.Params.SlashFractionDowntime = sdk.NewDecWithPrec(1, 4)   // 0.01%
	appState[slashingtypes.ModuleName] = cdc.MustMarshalJSON(&slashingGenState)

	// Override Tendermint consensus params: https://docs.tendermint.com/master/tendermint-core/using-tendermint.html#fields
	genDoc.ConsensusParams.Evidence.MaxAgeDuration = unbondingPeriod // should correspond with unbondingPeriod for handling Nothing-At-Stake attacks
	genDoc.ConsensusParams.Evidence.MaxAgeNumBlocks = int64(unbondingPeriod.Seconds()) / blockTimeSec

	return tmjson.Marshal(appState)
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", sdk.MustSortJSON(out))

	return err
}
