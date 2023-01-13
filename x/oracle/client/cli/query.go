package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group oracle queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdGetOracles())
	cmd.AddCommand(CmdGetOracle())
	cmd.AddCommand(CmdGetOracleRegistrations())
	cmd.AddCommand(CmdGetOracleRegistration())
	cmd.AddCommand(CmdGetOracleUpgrade())
	cmd.AddCommand(CmdGetOracleUpgrades())
	cmd.AddCommand(CmdGetOracleUpgradeInfo())
	cmd.AddCommand(CmdGetParams())

	return cmd
}
