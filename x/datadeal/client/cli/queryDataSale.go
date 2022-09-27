package cli

import (
	"encoding/base64"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func GetCmdDataSale() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datasale [dealID] [verifiableCID]",
		Short: "Query a datasale",
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

			verifiableCID, err := base64.StdEncoding.DecodeString(args[1])
			if err != nil {
				return err
			}

			req := &types.QueryDataSaleRequest{
				DealId:        dealID,
				VerifiableCid: verifiableCID,
			}
			res, err := queryClient.DataSale(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
