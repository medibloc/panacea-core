package cli

import (
	"context"
	"encoding/base64"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/spf13/cobra"
)

func CmdGetDID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-did [did]",
		Short: "Get a DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			id, err := types.ParseDID(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			params := &types.QueryServiceDIDRequest{
				DidBase64: base64.StdEncoding.EncodeToString([]byte(id)),
			}

			res, err := queryClient.DID(context.Background(), params)

			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
