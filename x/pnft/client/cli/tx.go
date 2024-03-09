package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/spf13/cobra"
)

func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(NewCmdCreateDenom())
	cmd.AddCommand(NewCmdUpdateDenom())
	cmd.AddCommand(NewCmdDeleteDenom())
	cmd.AddCommand(NewCmdTransferDenom())

	return cmd
}
