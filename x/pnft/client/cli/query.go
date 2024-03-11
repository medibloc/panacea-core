package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/spf13/cobra"
)

func NewGetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(NewCmdGetDenoms())
	cmd.AddCommand(NewCmdGetDenomsByOwner())
	cmd.AddCommand(NewCmdGetDenom())

	cmd.AddCommand(NewCmdGetPNFTs())
	cmd.AddCommand(NewCmdGetPNFTsByOwner())
	cmd.AddCommand(NewCmdGetPNFT())

	return cmd
}
