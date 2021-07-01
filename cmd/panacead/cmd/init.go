package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/medibloc/panacea-core/app"
	"github.com/medibloc/panacea-core/types/assets"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
)

// InitCmd wraps the genutilcli.InitCmd to inject specific parameters for Panacea.
// It reads the default genesis.json, modifies and exports it.
// Reference: https://github.com/public-awesome/stargaze/blob/b92bf9847559b9c7f4ac08576d056d3d00efe12c/cmd/starsd/cmd/init.go
func InitCmd(mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	initCmd := genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome)
	initCmd.PostRunE = func(cmd *cobra.Command, args []string) error {
		clientCtx := client.GetClientContextFromCmd(cmd)
		cdc := clientCtx.JSONMarshaler

		serverCtx := server.GetServerContextFromCmd(cmd)
		config := serverCtx.Config

		config.SetRoot(clientCtx.HomeDir)

		genFile := config.GenesisFile()
		genDoc := &types.GenesisDoc{}

		if _, err := os.Stat(genFile); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		} else {
			genDoc, err = types.GenesisDocFromFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to read genesis file: %w", err)
			}
		}

		appState, err := overrideGenesis(cdc, genDoc)
		if err != nil {
			return fmt.Errorf("failed to override genesis: %w", err)
		}

		genDoc.AppState = appState
		if err := genutil.ExportGenesisFile(genDoc, genFile); err != nil {
			return fmt.Errorf("failed to export genesis file: %w", err)
		}

		return nil
	}

	return initCmd
}

const (
	blockTimeSec    = 1
	unbondingPeriod = 60 * 60 * 24 * 7 * 3 * time.Second // three weeks
)

// overrideGenesis overrides some parameters in the genesis doc to the panacea-specific values.
func overrideGenesis(cdc codec.JSONMarshaler, genDoc *types.GenesisDoc) (json.RawMessage, error) {
	appState := make(map[string]json.RawMessage)
	if err := tmjson.Unmarshal(genDoc.AppState, &appState); err != nil {
		return nil, fmt.Errorf("failed to JSON unmarshal initial genesis state %w", err)
	}

	var stakingGenState stakingtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[stakingtypes.ModuleName], &stakingGenState); err != nil {
		return nil, err
	}
	stakingGenState.Params.UnbondingTime = unbondingPeriod
	stakingGenState.Params.MaxValidators = 50
	stakingGenState.Params.BondDenom = assets.MicroMedDenom
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
	mintGenState.Params.BlocksPerYear = uint64(60*60*24*365.25) * uint64(blockTimeSec)
	appState[minttypes.ModuleName] = cdc.MustMarshalJSON(&mintGenState)

	var distrGenState distrtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[distrtypes.ModuleName], &distrGenState); err != nil {
		return nil, err
	}
	distrGenState.Params.CommunityTax = sdk.ZeroDec()
	appState[distrtypes.ModuleName] = cdc.MustMarshalJSON(&distrGenState)

	var govGenState govtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[govtypes.ModuleName], &govGenState); err != nil {
		return nil, err
	}
	minDepositTokens := sdk.TokensFromConsensusPower(100000) // 100,000 MED
	govGenState.DepositParams.MinDeposit = sdk.Coins{sdk.NewCoin(assets.MicroMedDenom, minDepositTokens)}
	govGenState.DepositParams.MaxDepositPeriod = 60 * 60 * 24 * 14 * time.Second // 14 days
	govGenState.VotingParams.VotingPeriod = 60 * 60 * 24 * 14 * time.Second      // 14 days
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
