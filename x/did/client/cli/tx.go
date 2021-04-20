package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/go-bip39"
	didcrypto "github.com/medibloc/panacea-core/x/did/client/crypto"
	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	flagInteractive = "interactive"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	didTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "did transaction subcommands",
	}

	didTxCmd.AddCommand(client.PostCommands(
		GetCmdCreateDID(cdc),
		GetCmdUpdateDID(cdc),
		GetCmdDeactivateDID(cdc),
	)...)

	return didTxCmd
}

func GetCmdCreateDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-did",
		Short: "Create a DID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			inBuf := bufio.NewReader(cmd.InOrStdin())

			mnemonic, bip39Passphrase, err := readBIP39ParamsFrom(viper.GetBool(flagInteractive), inBuf)
			if err != nil {
				return err
			}
			privKey, err := didcrypto.GenSecp256k1PrivKey(mnemonic, bip39Passphrase)
			if err != nil {
				return err
			}

			msg, err := newMsgCreateDID(cliCtx, privKey)
			if err != nil {
				return err
			}

			if err := savePrivKeyToKeyStore(msg.VerificationMethodID, privKey, inBuf); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Bool(flagInteractive, false, "Interactively prompt user for BIP39 mnemonic and passphrase")
	return cmd
}

// newMsgCreateDID creates a MsgCreateDID by generating a DID and a DID document from the networkID and privKey.
// It generates the minimal DID document which contains only one public key information,
// so that it can be extended by MsgUpdateDID later.
func newMsgCreateDID(cliCtx context.CLIContext, privKey secp256k1.PrivKeySecp256k1) (types.MsgCreateDID, error) {
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(privKey))
	did := types.NewDID(pubKey)
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethods := []types.VerificationMethod{
		types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey),
	}
	authentications := []types.VerificationRelationship{
		types.NewVerificationRelationship(verificationMethods[0].ID),
	}
	doc := types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods), types.WithAuthentications(authentications))

	sig, err := types.Sign(doc, types.InitialSequence, privKey)
	if err != nil {
		return types.MsgCreateDID{}, err
	}

	msg := types.NewMsgCreateDID(did, doc, verificationMethodID, sig, cliCtx.GetFromAddress())
	if err := msg.ValidateBasic(); err != nil {
		return types.MsgCreateDID{}, err
	}
	return msg, nil
}

func GetCmdUpdateDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-did [did] [key-id] [did-doc-path]",
		Short: "Update a DID Document",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			inBuf := bufio.NewReader(cmd.InOrStdin())

			did, err := types.ParseDID(args[0])
			if err != nil {
				return err
			}
			verificationMethodID, err := types.ParseVerificationMethodID(args[1], did)
			if err != nil {
				return err
			}
			// read an input file of DID document
			doc, err := readDIDDocFrom(args[2])
			if err != nil {
				return err
			}
			privKey, err := getPrivKeyFromKeyStore(verificationMethodID, inBuf)
			if err != nil {
				return err
			}

			// For proving that I know the private key. It signs on the DIDDocument.
			sig, err := signUsingCurrentSeq(cliCtx, did, privKey, doc)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateDID(did, doc, verificationMethodID, sig, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

func GetCmdDeactivateDID(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-did [did] [key-id]",
		Short: "Deactivate a DID Document",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			inBuf := bufio.NewReader(cmd.InOrStdin())

			did, err := types.ParseDID(args[0])
			if err != nil {
				return err
			}
			verificationMethodID, err := types.ParseVerificationMethodID(args[1], did)
			if err != nil {
				return err
			}
			privKey, err := getPrivKeyFromKeyStore(verificationMethodID, inBuf)
			if err != nil {
				return err
			}

			// For proving that I know the private key. It signs on the DID, not DIDDocument.
			sig, err := signUsingCurrentSeq(cliCtx, did, privKey, did)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeactivateDID(did, verificationMethodID, sig, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}

// readBIP39ParamsFrom reads a mnemonic and a bip39 passphrase from the reader in the interactive mode.
// It returns empty strings in the non-interactive mode, so that they can be auto-generated by crypto.GenSecp256k1PrivKey.
func readBIP39ParamsFrom(interactive bool, reader *bufio.Reader) (string, string, error) {
	if !interactive {
		return "", "", nil
	}

	// mnemonic can be an empty string
	mnemonic, err := client.GetString("Enter your BIP39 mnemonic, or hit enter to generate one:", reader)
	if err != nil {
		return "", "", err
	}
	if mnemonic != "" && !bip39.IsMnemonicValid(mnemonic) {
		return "", "", fmt.Errorf("invalid mnemonic")
	}

	// passphrase can be an empty string
	passphrase, err := client.GetString("Enter your BIP39 passphrase, or hit enter:", reader)
	if err != nil {
		return "", "", err
	}
	if passphrase != "" {
		repeat, err := client.GetString("Repeat the passphrase:", reader)
		if err != nil {
			return "", "", err
		}
		if passphrase != repeat {
			return "", "", fmt.Errorf("passphrases don't match")
		}
	}

	return mnemonic, passphrase, nil
}

// readDIDDocFrom reads a DID document from a JSON file.
// It returns an error if the JSON file is invalid or the DID document loaded is invalid.
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

func keystoreBaseDir() string {
	return filepath.Join(viper.GetString(cli.HomeFlag), "did_keystore")
}

// savePrivKeyToKeyStore saves a privKey using a password which is read from the reader.
func savePrivKeyToKeyStore(verificationMethodID types.VerificationMethodID, privKey secp256k1.PrivKeySecp256k1, reader *bufio.Reader) error {
	passwd, err := client.GetCheckPassword(
		"Enter a password to encrypt your key for DID to disk:",
		"Repeat the password:",
		reader,
	)
	if err != nil {
		return err
	}
	ks, err := didcrypto.NewKeyStore(keystoreBaseDir())
	if err != nil {
		return err
	}
	_, err = ks.Save(string(verificationMethodID), privKey[:], passwd)
	return err
}

// getPrivKeyFromKeyStore loads a privKey using a password which is read from the reader.
func getPrivKeyFromKeyStore(verificationMethodID types.VerificationMethodID, reader *bufio.Reader) (secp256k1.PrivKeySecp256k1, error) {
	passwd, err := client.GetPassword("Enter a password to decrypt your key for DID on disk:", reader)
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	ks, err := didcrypto.NewKeyStore(keystoreBaseDir())
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	privKeyBytes, err := ks.LoadByAddress(string(verificationMethodID), passwd)
	if err != nil {
		return secp256k1.PrivKeySecp256k1{}, err
	}

	return secp256k1util.PrivKeyFromBytes(privKeyBytes)
}

// signUsingCurrentSeq generates a signature using the current sequence stored in the blockchain.
func signUsingCurrentSeq(cliCtx context.CLIContext, did types.DID, privKey crypto.PrivKey, data types.Signable) ([]byte, error) {
	docWithSeq, err := queryDIDDocumentWithSeq(cliCtx, did)
	if err != nil {
		return nil, err
	}
	return types.Sign(data, docWithSeq.Seq, privKey)
}
