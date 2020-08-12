package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func GetCmdCreateDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-did",
		Short: "Create a DID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			privKey := secp256k1.GenPrivKey() //TODO: store this somewhere safely
			pubKey := privKey.PubKey()

			did := types.NewDID(pubKey)
			doc := types.NewDIDDocument(
				did,
				types.MustNewPubKey("key1", pubKey, types.ES256K),
			)

			msg := types.NewMsgCreateDID(did, doc, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	return cmd
}
