package cli

import (
	"fmt"
	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/medibloc/panacea-core/x/aol/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group aol queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// this line is used by starport scaffolding # 1

	cmd.AddCommand(CmdListOwner())
	cmd.AddCommand(CmdShowOwner())

	cmd.AddCommand(CmdListRecord())
	cmd.AddCommand(CmdShowRecord())

	cmd.AddCommand(CmdListWriter())
	cmd.AddCommand(CmdShowWriter())

	cmd.AddCommand(CmdListTopic())
	cmd.AddCommand(CmdShowTopic())

	return cmd
}
