package cli

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/medibloc/panacea-core/x/did"
	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/spf13/cobra"
)

const (
	RouteDID = "custom/did/did"
)

func GetCmdQueryDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-did [did]",
		Short: "Get a DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			id := types.DID(args[0])
			if !id.Valid() {
				return types.ErrInvalidDID(args[0])
			}

			params := did.QueryDIDParams{DID: id}
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryWithData(RouteDID, bz)
			if err != nil {
				return err
			}

			var doc types.DIDDocument
			cdc.MustUnmarshalJSON(res, &doc)
			if doc.Empty() {
				return errors.New("DID not found")
			}
			return cliCtx.PrintOutput(doc)
		},
	}
	return cmd
}
