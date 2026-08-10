package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

// GetQueryCmd returns the custom NFT query root. Standard NFT query commands
// are added to this root by AutoCLI.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "NFT query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewClassRecordQueryCmd(),
		NewNFTRecordQueryCmd(),
		NewNFTRecordsQueryCmd(),
	)
	return cmd
}
