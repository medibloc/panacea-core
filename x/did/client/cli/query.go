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

			docWithSeq, err := queryDIDDocumentWithSeq(cliCtx, id)
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(docWithSeq.Document)
		},
	}
	return cmd
}

func queryDIDDocumentWithSeq(cliCtx context.CLIContext, id types.DID) (types.DIDDocumentWithSeq, error) {
	bz, err := cliCtx.Codec.MarshalJSON(did.QueryDIDParams{DID: id})
	if err != nil {
		return types.DIDDocumentWithSeq{}, err
	}

	res, err := cliCtx.QueryWithData(RouteDID, bz)
	if err != nil {
		return types.DIDDocumentWithSeq{}, err
	}

	var doc types.DIDDocumentWithSeq
	cliCtx.Codec.MustUnmarshalJSON(res, &doc)
	if doc.Empty() {
		return types.DIDDocumentWithSeq{}, errors.New("DID not found")
	}

	return doc, nil
}
