package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func CmdDeactivateDeal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-deal [dealID]",
		Short: "Deactivate a deal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dealID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			requesterAddr := clientCtx.GetFromAddress()

			msgDeactivateDeal := types.NewMsgDeactivateDeal(dealID, requesterAddr.String())

			if err := msgDeactivateDeal.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgDeactivateDeal)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
