package main

import (
	"encoding/json"
	"io"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/medibloc/panacea-core/app"
	panaceaInit "github.com/medibloc/panacea-core/init"
	panaceaServer "github.com/medibloc/panacea-core/server"
	"github.com/medibloc/panacea-core/types/util"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
)

// panacead custom flags
const flagInvCheckPeriod = "inv-check-period"

var invCheckPeriod uint

func main() {
	cdc := app.MakeCodec()

	config := sdk.GetConfig()
	config.SetCoinType(371)
	config.SetFullFundraiserPath("44'/371'/0'/0/0")
	config.SetBech32PrefixForAccount(util.Bech32PrefixAccAddr, util.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(util.Bech32PrefixValAddr, util.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(util.Bech32PrefixConsAddr, util.Bech32PrefixConsPub)
	config.Seal()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "panacead",
		Short:             "Panacea Daemon (server)",
		PersistentPreRunE: panaceaServer.PersistentPreRunEFn(ctx),
	}

	rootCmd.AddCommand(panaceaInit.InitCmd(ctx, cdc))
	rootCmd.AddCommand(panaceaInit.CollectGenTxsCmd(ctx, cdc))
	rootCmd.AddCommand(panaceaInit.TestnetFilesCmd(ctx, cdc))
	rootCmd.AddCommand(panaceaInit.GenTxCmd(ctx, cdc))
	rootCmd.AddCommand(panaceaInit.AddGenesisAccountCmd(ctx, cdc))
	rootCmd.AddCommand(panaceaInit.ValidateGenesisCmd(ctx, cdc))
	rootCmd.AddCommand(client.NewCompletionCmd(rootCmd, true))

	panaceaServer.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "PA", app.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewPanaceaApp(
		logger, db, traceStore, true, invCheckPeriod,
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
	)
}

func exportAppStateAndTMValidators(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		pApp := app.NewPanaceaApp(logger, db, traceStore, false, uint(1))
		err := pApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return pApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}
	pApp := app.NewPanaceaApp(logger, db, traceStore, true, uint(1))
	return pApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}
