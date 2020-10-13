package init

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/medibloc/panacea-core/types/assets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/types"
	"os"
	"time"
)

// InitCmd wraps the genutil.InitCmd to inject specific parameters for Panacea.
// It reads the default genesis.json, modifies and exports it.
func InitCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager, defaultNodeHome string) *cobra.Command {
	initCmd := genutilcli.InitCmd(ctx, cdc, mbm, defaultNodeHome)
	initCmd.PostRunE = func(cmd *cobra.Command, args []string) error {
		config := ctx.Config
		config.SetRoot(viper.GetString(cli.HomeFlag))

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

		genDoc.AppState, err = cdc.MarshalJSON(appState)
		if err != nil {
			return fmt.Errorf("failed to marshal the app state JSON: %w", err)
		}

		if err := genutil.ExportGenesisFile(genDoc, genFile); err != nil {
			return fmt.Errorf("failed to export genesis file: %w", err)
		}


		return nil
	}

	return initCmd
}

// overrideGenesis overrides some parameters in the genesis doc to the panacea-specific values.
func overrideGenesis(cdc *codec.Codec, genDoc *types.GenesisDoc) (genutil.AppMap, error) {
	var appState genutil.AppMap
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the app state JSON: %w", err)
	}

	var stakingGenState staking.GenesisState
	if err := cdc.UnmarshalJSON(appState[staking.ModuleName], &stakingGenState); err != nil {
		return nil, err
	}
	stakingGenState.Params.UnbondingTime = time.Second * 60 * 60 * 24 * 7 * 3 // three weeks
	stakingGenState.Params.MaxValidators = 21
	stakingGenState.Params.BondDenom = assets.MicroMedDenom
	appState[staking.ModuleName] = cdc.MustMarshalJSON(stakingGenState)

	var mintGenState mint.GenesisState
	if err := cdc.UnmarshalJSON(appState[mint.ModuleName], &mintGenState); err != nil {
		return nil, err
	}
	mintGenState.Minter = mint.InitialMinter(sdk.NewDecWithPrec(5, 2))
	mintGenState.Params.MintDenom = assets.MicroMedDenom
	mintGenState.Params.InflationRateChange = sdk.NewDecWithPrec(2, 2)
	mintGenState.Params.InflationMin = sdk.NewDecWithPrec(4, 2)
	mintGenState.Params.InflationMax = sdk.NewDecWithPrec(6, 2)
	mintGenState.Params.BlocksPerYear = uint64(60 * 60 * 24 * 365.25)  // assuming 1 second block time
	appState[mint.ModuleName] = cdc.MustMarshalJSON(mintGenState)

	var distributionGenState distribution.GenesisState
	if err := cdc.UnmarshalJSON(appState[distribution.ModuleName], &distributionGenState); err != nil {
		return nil, err
	}
	distributionGenState.CommunityTax = sdk.ZeroDec()
	appState[distribution.ModuleName] = cdc.MustMarshalJSON(distributionGenState)

	var govGenState gov.GenesisState
	if err := cdc.UnmarshalJSON(appState[gov.ModuleName], &govGenState); err != nil {
		return nil, err
	}
	minDepositTokens := sdk.TokensFromConsensusPower(100000) // 100,000 MED
	govGenState.DepositParams.MinDeposit = sdk.Coins{sdk.NewCoin(assets.MicroMedDenom, minDepositTokens)}
	govGenState.DepositParams.MaxDepositPeriod = 60 * 60 * 24 * 14 * time.Second // 14 days
	govGenState.VotingParams.VotingPeriod = 60 * 60 * 24 * 14 * time.Second // 14 days
	appState[gov.ModuleName] = cdc.MustMarshalJSON(govGenState)

	var crisisGenState crisis.GenesisState
	if err := cdc.UnmarshalJSON(appState[crisis.ModuleName], &crisisGenState); err != nil {
		return nil, err
	}
	crisisGenState.ConstantFee = sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(1000000000000)) // Spend 1,000,000 MED for invariants check
	appState[crisis.ModuleName] = cdc.MustMarshalJSON(crisisGenState)

	var slashingGenState slashing.GenesisState
	if err := cdc.UnmarshalJSON(appState[slashing.ModuleName], &slashingGenState); err != nil {
		return nil, err
	}
	slashingGenState.Params.MaxEvidenceAge = 60 * 60 * 24 * 7 * 3 * time.Second // 3 weeks as same as unbonding period
	slashingGenState.Params.SignedBlocksWindow = 10000
	slashingGenState.Params.MinSignedPerWindow = sdk.NewDecWithPrec(5, 2)
	appState[slashing.ModuleName] = cdc.MustMarshalJSON(slashingGenState)

	return appState, nil
}

