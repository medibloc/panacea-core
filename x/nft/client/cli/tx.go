package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

// GetTxCmd returns the custom NFT transaction root. Panacea transaction
// commands are added to this root by AutoCLI.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "NFT transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(NewSendCmd())
	return cmd
}
