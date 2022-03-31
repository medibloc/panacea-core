package cli

import (
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
