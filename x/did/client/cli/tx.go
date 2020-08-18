package cli

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"os"
)

func GetCmdCreateDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-did",
		Short: "Create a DID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			privKey := secp256k1.GenPrivKey()      //TODO: implement wallet
			fmt.Println(base58.Encode(privKey[:])) //TODO: don't print it. store it securely.
			pubKey := privKey.PubKey()

			did := types.NewDID(pubKey, types.ES256K)
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

func GetCmdUpdateDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-did [did] [priv-key-base58] [pub-key-id] [did-doc-path]",
		Short: "Update a DID Document",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			did := types.DID(args[0])
			if !did.IsValid() {
				return types.ErrInvalidDID(did)
			}

			// private key which is corresponding to the public key registered in the DID document
			// TODO: Don't get this via CLI arg. Implement Wallet.
			privKey, err := types.NewPrivKeyFromBase58(args[1])
			if err != nil {
				return types.ErrInvalidSecp256k1PrivateKey(err)
			}
			pubKeyID := types.PubKeyID(args[2])

			// read an input file of DID document
			file, err := os.Open(args[3])
			if err != nil {
				return err
			}
			defer file.Close()

			var doc types.DIDDocument
			err = json.NewDecoder(file).Decode(&doc)
			if err != nil || !doc.IsValid() {
				return types.ErrInvalidDIDDocument()
			}

			// For proving that I know the private key
			sig, err := privKey.Sign(doc.GetSignBytes())
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateDID(did, doc, pubKeyID, sig, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	return cmd
}
