package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

const (
	flagOracleUniqueID          = "oracle-unique-id"
	flagNodePublicKey           = "node-public-key"
	flagNodePubKeyRemoteReport  = "node-public-key-remote-report"
	flagTrustedBlockHeight      = "trusted-block-height"
	flagTrustedBlockHash        = "trusted-block-hash"
	flagOracleEndpoint          = "oracle-endpoint"
	flagOracleCommRate          = "oracle-commission-rate"
	flagOracleCommMaxRate       = "oracle-commission-max-rate"
	flagOracleCommMaxChangeRate = "oracle-commission-max-change-rate"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdRegisterOracle())
	cmd.AddCommand(CmdUpdateOracleInfo())
	cmd.AddCommand(CmdUpgradeOracle())

	return cmd
}
