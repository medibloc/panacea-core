package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	aolCmds "github.com/medibloc/panacea-core/x/aol/client/cli"
	"github.com/medibloc/panacea-core/x/aol/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	aolQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the aol module",
	}

	aolQueryCmd.AddCommand(client.GetCommands(
		aolCmds.GetCmdQueryListTopic(mc.cdc),
		aolCmds.GetCmdQueryListWriter(mc.cdc),
		aolCmds.GetCmdQueryRecord(mc.cdc),
		aolCmds.GetCmdQueryTopic(mc.cdc),
		aolCmds.GetCmdQueryWriter(mc.cdc),
	)...)

	return aolQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	aolTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "aol transaction subcommands",
	}

	aolTxCmd.AddCommand(client.PostCommands(
		aolCmds.GetCmdAddRecord(mc.cdc),
		aolCmds.GetCmdAddWriter(mc.cdc),
		aolCmds.GetCmdCreateTopic(mc.cdc),
		aolCmds.GetCmdDeleteWriter(mc.cdc),
	)...)

	return aolTxCmd
}
