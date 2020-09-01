package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/go-bip39"
	"github.com/medibloc/panacea-core/x/did/client/keystore"
	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	flagInteractive = "interactive"

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

			privKey, err := genPrivKey(viper.GetBool(flagInteractive))
			if err != nil {
				return err
			}

			networkID, err := types.NewNetworkID(args[0])
			if err != nil {
				return err
			}

			pubKey := privKey.PubKey()
			did := types.NewDID(networkID, pubKey, types.ES256K)
			keyID := types.NewKeyID(did, "key1")
			doc := types.NewDIDDocument(did, types.NewPubKey(keyID, types.ES256K, pubKey))

			passwd, err := client.GetCheckPassword(
				"Enter a password to encrypt your key for DID to disk:",
				"Repeat the password:",
				client.BufferStdin(),
			)
			if err != nil {
				return err
			}

			ks, err := keystore.NewKeyStore(keystoreBaseDir())
			if err != nil {
				return err
			}
			_, err = ks.Save(string(keyID), privKey[:], passwd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateDID(did, doc, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}

	cmd.Flags().Bool(flagInteractive, false, "Interactively prompt user for BIP39 mnemonic and passphrase")
	return cmd
}

func GetCmdUpdateDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-did [did] [key-id] [did-doc-path]",
		Short: "Update a DID Document",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			did, err := types.NewDIDFrom(args[0])
			if err != nil {
				return err
			}
			keyID, err := types.NewKeyIDFrom(args[1])
			if err != nil {
				return err
			}
			// read an input file of DID document
			doc, err := readDIDDocFrom(args[2])
			if err != nil {
				return err
			}

			privKey, err := getPrivKeyFromKeyStore(keyID)
			if err != nil {
				return err
			}

			// For proving that I know the private key. It signs on the DIDDocument.
			sig, err := sign(cliCtx, did, privKey, doc)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateDID(did, doc, keyID, sig, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	return cmd
}

func GetCmdDeleteDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-did [did] [key-id]",
		Short: "Delete a DID Document",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			did := types.DID(args[0])
			if !did.Valid() {
				return types.ErrInvalidDID(string(did))
			}
			keyID, err := types.NewKeyIDFrom(args[1])
			if err != nil {
				return err
			}

			privKey, err := getPrivKeyFromKeyStore(keyID)
			if err != nil {
				return err
			}

			// For proving that I know the private key. It signs on the DID, not DIDDocument.
			sig, err := sign(cliCtx, did, privKey, did)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteDID(did, keyID, sig, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	return cmd
}

func genPrivKey(interactive bool) (secp256k1.PrivKeySecp256k1, error) {
	var err error
	var mnemonic string
	var bip39Passphrase string

	if interactive {
		mnemonic, bip39Passphrase, err = readBIP39ParamsFrom(client.BufferStdin())
		if err != nil {
			return secp256k1.PrivKeySecp256k1{}, err
		}
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
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	hdPath := hd.NewFundraiserParams(defaultAccountForHD, sdk.GetConfig().GetCoinType(), defaultIndexForHD).String()
	masterPriv, chainCode := hd.ComputeMastersFromSeed(seed)
	return hd.DerivePrivateKeyForPath(masterPriv, chainCode, hdPath)
}

func readBIP39ParamsFrom(buf *bufio.Reader) (string, string, error) {
	// mnemonic can be an empty string
	mnemonic, err := client.GetString("Enter your BIP39 mnemonic, or hit enter to generate one:", buf)
	if err != nil {
		return "", "", err
	}
	if mnemonic != "" && !bip39.IsMnemonicValid(mnemonic) {
		return "", "", fmt.Errorf("invalid mnemonic")
	}

	// passphrase can be an empty string
	passphrase, err := client.GetString("Enter your BIP39 passphrase, or hit enter:", buf)
	if err != nil {
		return "", "", err
	}
	if passphrase != "" {
		repeat, err := client.GetString("Repeat the passphrase:", buf)
		if err != nil {
			return "", "", err
		}
		if passphrase != repeat {
			return "", "", fmt.Errorf("passphrases don't match")
		}
	}

	return mnemonic, passphrase, nil
}

func keystoreBaseDir() string {
	return filepath.Join(viper.GetString(cli.HomeFlag), "did_keystore")
}

func readDIDDocFrom(path string) (types.DIDDocument, error) {
	var doc types.DIDDocument

	file, err := os.Open(path)
	if err != nil {
		return doc, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&doc)
	if err != nil {
		return doc, fmt.Errorf("fail to decode DIDDocument JSON: %w", err)
	}
	if !doc.Valid() {
		return doc, types.ErrInvalidDIDDocument(doc)
	}

	return doc, nil
}

func getPrivKeyFromKeyStore(keyID types.KeyID) (secp256k1.PrivKeySecp256k1, error) {
	passwd, err := client.GetPassword(
		"Enter a password to decrypt your key for DID on disk:",
		client.BufferStdin(),
	)
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	ks, err := keystore.NewKeyStore(keystoreBaseDir())
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	privKeyBytes, err := ks.LoadByAddress(string(keyID), passwd)
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	return types.NewPrivKeyFromBytes(privKeyBytes)
}

func sign(cliCtx context.CLIContext, did types.DID, privKey crypto.PrivKey, data types.Signable) ([]byte, error) {
	// get a DIDDocumentWithSeq to use its Seq for signing
	doc, err := queryDID(cliCtx, did)
	if err != nil {
		return nil, err
	}
	return types.Sign(data, doc.Seq, privKey)
}
