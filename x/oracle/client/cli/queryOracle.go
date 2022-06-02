package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdGetOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle [oracle-address]",
		Short: "Query a oracle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Oracle(cmd.Context(), &types.QueryOracleRequest{
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
