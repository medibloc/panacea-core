package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func CmdRequestDataDeliveryVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-delivery-vote [dealID] [dataHash]",
		Short: "request again data delivery vote if the data delivery vote failed",
		Args:  cobra.ExactArgs(2),
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

			msgRequestDataDeliveryVote := types.NewMsgRequestDataDeliveryVote(dealID, args[1], requesterAddr.String())

			if err := msgRequestDataDeliveryVote.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgRequestDataDeliveryVote)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
