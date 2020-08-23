package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/btcsuite/btcutil/base58"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagUseMnemonic = "use-mnemonic"

	mnemonicEntropySize = 256
	defaultAccountForHD = 0
	defaultIndexForHD   = 0
)

func GetCmdCreateDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-did [network-id]",
		Short: "Create a DID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			privKey, err := genPrivKey(viper.GetBool(flagUseMnemonic))
			if err != nil {
				return err
			}
			fmt.Println(base58.Encode(privKey[:])) //TODO: don't print it. store it securely.

			networkID, err := types.NewNetworkID(args[0])
			if err != nil {
				return err
			}

			pubKey := privKey.PubKey()
			did := types.NewDID(networkID, pubKey, types.ES256K)
			doc := types.NewDIDDocument(did, types.NewPubKey("key1", pubKey, types.ES256K))

			msg := types.NewMsgCreateDID(did, doc, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	cmd.Flags().Bool(flagUseMnemonic, false, "Use BIP39 Mnemonic to generate a private key")
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
			if !did.Valid() {
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
			if err != nil || !doc.Valid() {
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

func genPrivKey(useMnemonic bool) (secp256k1.PrivKeySecp256k1, error) {
	if !useMnemonic {
		return secp256k1.GenPrivKey(), nil // use OS randomness
	}

	mnemonic, err := client.GetString("Enter your BIP39 mnemonic, or hit enter to generate one:", client.BufferStdin())
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	if mnemonic == "" { // generate a random mnemonic
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return secp256k1.PrivKeySecp256k1{}, err
		}
		mnemonic, err = bip39.NewMnemonic(entropySeed[:])
		if err != nil {
			return secp256k1.PrivKeySecp256k1{}, err
		}
		fmt.Fprintf(os.Stderr, "A random mnemonic was generated: %s\n", mnemonic)
	} else if !bip39.IsMnemonicValid(mnemonic) {
		return secp256k1.PrivKeySecp256k1{}, fmt.Errorf("invalid mnemonic")
	}

	bip39Passphrase, err := client.GetCheckPassword(
		"Enter your BIP39 passphrase:",
		"Repeat the password:",
		client.BufferStdin(),
	)
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	//TODO: can I use this?
	hdPath := hd.NewFundraiserParams(defaultAccountForHD, sdk.GetConfig().GetCoinType(), defaultIndexForHD).String()
	masterPriv, chainCode := hd.ComputeMastersFromSeed(seed)
	return hd.DerivePrivateKeyForPath(masterPriv, chainCode, hdPath)
}
