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
	RouteResolveDID = "custom/did/resolveDid"
)

func GetCmdResolveDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-did [did]",
		Short: "Resolve a DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			id := types.DID(args[0])
			if !id.IsValid() {
				return types.ErrInvalidDID(id)
			}

			params := did.ResolveDIDParams{id}
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryWithData(RouteResolveDID, bz)
			if err != nil {
				return err
			}

			var doc types.DIDDocument
			cdc.MustUnmarshalJSON(res, &doc)
			if doc.IsEmpty() {
				return errors.New("DID not found")
			}
			return cliCtx.PrintOutput(doc)
		},
	}
	return cmd
}
