package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"

	didCmds "github.com/medibloc/panacea-core/x/did/client/cli"
)

// ModuleClient exports all client functionalities from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	didQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the did module",
	}

	didQueryCmd.AddCommand(client.GetCommands(
		didCmds.GetCmdResolveDID(mc.cdc),
	)...)

	return didQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	didTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "did transaction subcommands",
	}

	didTxCmd.AddCommand(client.PostCommands(
		didCmds.GetCmdCreateDID(mc.cdc),
		didCmds.GetCmdUpdateDID(mc.cdc),
		didCmds.GetCmdDeleteDID(mc.cdc),
	)...)

	return didTxCmd
}
