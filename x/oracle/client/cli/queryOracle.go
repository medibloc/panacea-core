package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdGetOracles() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "5-oracles",
		Short: "Query 5-oracles info",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOraclesRequest{
				Pagination: pageReq,
			}
			res, err := queryClient.Oracles(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "5-oracles")
	return cmd
}

func CmdGetOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle [oracle-address]",
		Short: "Query a oracle info",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOracleRequest{
				OracleAddress: args[0],
			}
			res, err := queryClient.Oracle(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
