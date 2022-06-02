package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/spf13/cobra"
)

func CmdGetPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [pool-id]",
		Short: "Query a pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.Pool(cmd.Context(), &types.QueryPoolRequest{
				PoolId: poolID,
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

func CmdGetDataCerts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-certs [pool-id] [round]",
		Short: "Query data certificates by pool and round",
		Args:  cobra.ExactArgs(2),
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

			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			round, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.DataCerts(cmd.Context(), &types.QueryDataCertsRequest{
				PoolId:     poolID,
				Round:      round,
				Pagination: pageReq,
			})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "all data certificates by pool ID and round")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetDataPassRedeemReceipt() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-pass-redeem-receipt [pool-id] [round] [data-pass-id]",
		Short: "Query a data pass redeem receipt",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			round, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			dataPassID, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.DataPassRedeemReceipt(cmd.Context(), &types.QueryDataPassRedeemReceiptRequest{
				PoolId:     poolID,
				Round:      round,
				DataPassId: dataPassID,
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

func CmdGetDataPassRedeemReceipts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-pass-redeem-receipts [pool-id]",
		Short: "Query data pass redeem receipts by pool",
		Args:  cobra.ExactArgs(1),
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

			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.DataPassRedeemReceipts(cmd.Context(), &types.QueryDataPassRedeemReceiptsRequest{
				PoolId:     poolID,
				Pagination: pageReq,
			})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "all data pass by pool ID")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetDataPassRedeemHistory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-pass-redeem-history [redeemer-addr] [pool-id]",
		Short: "Query data pass redeem history",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.DataPassRedeemHistory(cmd.Context(), &types.QueryDataPassRedeemHistoryRequest{
				Redeemer: args[0],
				PoolId:   poolID,
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
