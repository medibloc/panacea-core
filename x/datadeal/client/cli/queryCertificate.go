package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func CmdGetCertificates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificates [deal-id]",
		Short: "Query certificates",
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

			dealID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			req := &types.QueryCertificates{
				DealId:     dealID,
				Pagination: pageReq,
			}
			res, err := queryClient.Certificates(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetCertificate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificate [deal-id] [data-hash]",
		Short: "Query a certificate info",
		Args:  cobra.ExactArgs(2),
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

			req := &types.QueryCertificate{
				DealId:   dealID,
				DataHash: args[1],
			}
			res, err := queryClient.Certificate(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}