package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/x/aol/types"
)

var (
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())
)

const (
	flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
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

	// this line is used by starport scaffolding # 1

	cmd.AddCommand(CmdCreateOwner())
	cmd.AddCommand(CmdUpdateOwner())
	cmd.AddCommand(CmdDeleteOwner())

	cmd.AddCommand(CmdCreateRecord())
	cmd.AddCommand(CmdUpdateRecord())
	cmd.AddCommand(CmdDeleteRecord())

	cmd.AddCommand(CmdCreateWriter())
	cmd.AddCommand(CmdUpdateWriter())
	cmd.AddCommand(CmdDeleteWriter())

	cmd.AddCommand(CmdCreateTopic())
	cmd.AddCommand(CmdUpdateTopic())
	cmd.AddCommand(CmdDeleteTopic())

	return cmd
}
