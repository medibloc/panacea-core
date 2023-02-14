package cli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
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

			did := args[0]
			if err := types.ValidateDID(did); err != nil {
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

			documentBz := res.DidDocument.Document
			document, err := ariesdid.ParseDocument(documentBz)
			if err != nil {
				return err
			}

			jsonDocument, err := json.Marshal(document)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), string(jsonDocument))
			if err != nil {
				return err
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
