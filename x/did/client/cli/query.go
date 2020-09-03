package cli

import (
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
			id, err := types.NewDIDFrom(args[0])
			if err != nil {
				return err
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

// queryDIDDocumentWithSeq gets a DIDDocumentWithSeq from the blockchain.
// It returns an error if the DID doesn't exist or the DID has been deactivated.
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
		return types.DIDDocumentWithSeq{}, types.ErrDIDNotFound(id)
	}
	if doc.Deactivated() {
		return types.DIDDocumentWithSeq{}, types.ErrDIDDeactivated(id)
	}

	return doc, nil
}
