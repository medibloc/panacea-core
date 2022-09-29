package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func CmdGetDeals() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deals",
		Short: "Query deals",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			req := &types.QueryDealsRequest{
				Pagination: pageReq,
			}
			res, err := queryClient.Deals(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all deals")
	return cmd
}

func CmdGetDeal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deal [deal-id]",
		Short: "Query a deal info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			dealID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			req := &types.QueryDealRequest{
				DealId: dealID,
			}
			res, err := queryClient.Deal(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
