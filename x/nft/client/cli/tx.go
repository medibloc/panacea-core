package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

// GetTxCmd returns the custom NFT transaction root. AutoCLI adds the Panacea
// commands that do not require custom argument handling to this root.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "NFT transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewMintCmd(),
		NewSendCmd(),
	)
	return cmd
}
