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

			did, err := types.ValidateDID(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryDIDRequest{
				DidBase64: base64.StdEncoding.EncodeToString([]byte(did)),
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
