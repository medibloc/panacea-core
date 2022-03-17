package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
	"strconv"
)

func CmdGetDeal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-deal [dealId]",
		Short: "Query a deal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			dealId, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.Deal(cmd.Context(), &types.QueryDealRequest{
				DealId: uint64(dealId),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
