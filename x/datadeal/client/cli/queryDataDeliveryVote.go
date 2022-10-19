package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func CmdGetDataDeliveryVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-delivery-vote [deal_id] [data_hash] [voter_address]",
		Short: "Query a dataDelivery vote info",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			dealID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryDataDeliveryVoteRequest{
				DealId:       dealID,
				DataHash:     args[1],
				VoterAddress: args[2],
			}

			res, err := queryClient.DataDeliveryVote(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
