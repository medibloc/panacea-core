package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/spf13/cobra"
)

func CmdGetDataValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-data-validator [data validator address]",
		Short: "Query a data validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DataValidator(cmd.Context(), &types.QueryDataValidatorRequest{
				Address: args[0],
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

func CmdGetPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-pool [poolID]",
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

func CmdGetDataValidationCertificates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-validation-certs [poolID] [round]",
		Short: "Query data validation certificates by pool and round",
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

			res, err := queryClient.DataValidationCertificates(cmd.Context(), &types.QueryDataValidationCertificatesRequest{
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

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetDataPassRedeemReceipt() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-data-pass-receipt [poolID] [round] [nftID]",
		Short: "Query a data pass redeem receipt",
		Args:  cobra.ExactArgs(4),
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

			nftID, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.DataPassRedeemReceipt(cmd.Context(), &types.QueryDataPassRedeemReceiptRequest{
				PoolId: poolID,
				Round:  round,
				NftId:  nftID,
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
		Use:   "data-pass-redeem-receipts [poolID]",
		Short: "Query data pass redeem receipts by pool",
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
