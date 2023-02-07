package cli

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkcodec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/cli"

	didcrypto "github.com/medibloc/panacea-core/v2/x/did/client/crypto"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

const (
	flagInteractive     = "interactive"
	baseDir             = "did_keystore"
	didDocumentDataType = "github.com/hyperledger/aries-framework-go/pkg/doc/did.Doc@v0.1.8"
)

func CmdCreateDID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-did",
		Short: "Create a DID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())

			mnemonic, bip39Passphrase, err := readBIP39ParamsFrom(viper.GetBool(flagInteractive), inBuf)
			if err != nil {
				return err
			}
			privKey, err := didcrypto.GenSecp256k1PrivKey(mnemonic, bip39Passphrase)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress()

			msg, err := newMsgCreateDID(fromAddress, privKey)
			if err != nil {
				return err
			}
			if err := savePrivKeyToKeyStore(msg.VerificationMethodId, privKey, inBuf); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Bool(flagInteractive, false, "Interactively prompt user for BIP39 mnemonic and passphrase")
	return cmd
}

func CmdUpdateDID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-did [did] [key-id] [did-doc-path]",
		Short: "Update a DID Document",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())

			did := args[0]
			if err := types.ValidateDID(did); err != nil {
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
			sig, err := signUsingCurrentSeq(clientCtx, did, privKey, &doc)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress()
			err = cmd.Flags().Set(flags.FlagFrom, fromAddress.String())
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateDID(did, doc, verificationMethodID, sig, fromAddress.String())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdDeactivateDID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-did [did] [key-id]",
		Short: "Deactivate a DID Document",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			inBuf := bufio.NewReader(cmd.InOrStdin())

			did := args[0]
			if err := types.ValidateDID(did); err != nil {
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

			// TODO: As the document signing part improves, this part may need to be changed.
			// For proving that I know the private key. It signs on the DIDDocument.
			document := types.DIDDocument{
				Document:         nil,
				DocumentDataType: "",
			}
			sig, err := signUsingCurrentSeq(clientCtx, did, privKey, &document)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress()
			msg := types.NewMsgDeactivateDID(did, verificationMethodID, sig, fromAddress.String())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// readBIP39ParamsFrom reads a mnemonic and a bip39 passphrase from the reader in the interactive mode.
// It returns empty strings in the non-interactive mode, so that they can be auto-generated by crypto.GenSecp256k1PrivKey.
func readBIP39ParamsFrom(interactive bool, reader *bufio.Reader) (string, string, error) {
	if !interactive {
		return "", "", nil
	}

	// mnemonic can be an empty string
	mnemonic, err := input.GetString("Enter your BIP39 mnemonic, or hit enter to generate one:", reader)
	if err != nil {
		return "", "", err
	}
	if mnemonic != "" && !bip39.IsMnemonicValid(mnemonic) {
		return "", "", fmt.Errorf("invalid mnemonic")
	}

	// passphrase can be an empty string
	passphrase, err := input.GetString("Enter your BIP39 passphrase, or hit enter:", reader)
	if err != nil {
		return "", "", err
	}
	if passphrase != "" {
		repeat, err := input.GetString("Repeat the passphrase:", reader)
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
	return filepath.Join(viper.GetString(cli.HomeFlag), baseDir)
}

// savePrivKeyToKeyStore saves a privKey using a password which is read from the reader.
func savePrivKeyToKeyStore(verificationMethodID string, privKey secp256k1.PrivKey, reader *bufio.Reader) error {
	passwd, err := getCheckPassword(reader)
	if err != nil {
		return err
	}
	ks, err := didcrypto.NewKeyStore(keystoreBaseDir())
	if err != nil {
		return err
	}
	_, err = ks.Save(verificationMethodID, privKey[:], passwd)
	return err
}

// Deprecated to https://github.com/cosmos/cosmos-sdk/pull/5904/commits/c16d93e90d6a698cead7c19b55fcede44587aa7f
func getCheckPassword(reader *bufio.Reader) (string, error) {
	pass, err := input.GetPassword(
		"Enter a password to encrypt your key for DID to disk:",
		reader,
	)
	if err != nil {
		return "", err
	}

	pass2, err := input.GetPassword(
		"Repeat the password:",
		reader,
	)
	if err != nil {
		return "", err
	}
	if pass != pass2 {
		return "", errors.New("passphrases don't match")
	}
	return pass, nil
}

// newMsgCreateDID creates a MsgCreateDID by generating a DID and a DID document from the networkID and privKey.
// It generates the minimal DID document which contains only one public key information,
// so that it can be extended by MsgUpdateDID later.
func newMsgCreateDID(fromAddress sdk.AccAddress, privKey secp256k1.PrivKey) (types.MsgCreateDID, error) {
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(privKey))
	newDid := types.NewDID(pubKey)
	verificationMethodID := types.NewVerificationMethodID(newDid, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, newDid, pubKey)
	authentication := types.NewVerification(verificationMethod, ariesdid.Authentication)
	createdTime := time.Now()

	document := types.NewDocument(newDid,
		ariesdid.WithVerificationMethod([]ariesdid.VerificationMethod{verificationMethod}),
		ariesdid.WithAuthentication([]ariesdid.Verification{authentication}),
		ariesdid.WithCreatedTime(createdTime))

	didDocument, err := types.NewDIDDocument(document, didDocumentDataType)
	if err != nil {
		return types.MsgCreateDID{}, err
	}

	sig, err := types.Sign(&didDocument, types.InitialSequence, privKey)
	if err != nil {
		return types.MsgCreateDID{}, err
	}

	msg := types.NewMsgCreateDID(newDid, didDocument, verificationMethodID, sig, fromAddress.String())
	if err := msg.ValidateBasic(); err != nil {
		return types.MsgCreateDID{}, err
	}
	return msg, nil
}

// readDIDDocFrom reads a DID document from a JSON file.
// It returns an error if the JSON file is invalid or the DID document is invalid.
func readDIDDocFrom(path string) (types.DIDDocument, error) {
	var doc types.DIDDocument

	file, err := os.Open(path)
	if err != nil {
		return doc, err
	}
	defer file.Close()

	fileData, err := os.ReadFile(path)
	if err != nil {
		return doc, err
	}

	document, err := ariesdid.ParseDocument(fileData)
	if err != nil {
		return doc, err
	}
	ti := time.Now()
	document.Updated = &ti

	didDocument, err := types.NewDIDDocument(*document, didDocumentDataType)
	if err != nil {
		return doc, err
	}

	return didDocument, nil
}

// getPrivKeyFromKeyStore loads a privKey using a password which is read from the reader.
func getPrivKeyFromKeyStore(verificationMethodID string, reader *bufio.Reader) (secp256k1.PrivKey, error) {
	passwd, err := input.GetPassword("Enter a password to decrypt your key for DID on disk:", reader)
	if err != nil {
		return secp256k1.PrivKey{}, err
	}

	ks, err := didcrypto.NewKeyStore(keystoreBaseDir())
	if err != nil {
		return secp256k1.PrivKey{}, err
	}

	privKeyBytes, err := ks.LoadByAddress(verificationMethodID, passwd)
	if err != nil {
		return secp256k1.PrivKey{}, err
	}

	return secp256k1util.PrivKeyFromBytes(privKeyBytes)
}

// signUsingCurrentSeq generates a signature using the current sequence stored in the blockchain.
func signUsingCurrentSeq(clientCtx client.Context, did string, privKey crypto.PrivKey, data sdkcodec.ProtoMarshaler) ([]byte, error) {
	queryClient := types.NewQueryClient(clientCtx)

	params := &types.QueryDIDRequest{
		DidBase64: base64.StdEncoding.EncodeToString([]byte(did)),
	}

	res, err := queryClient.DID(context.Background(), params)
	if err != nil {
		return []byte{}, err
	}

	return types.Sign(data, res.DidDocumentWithSeq.Sequence, privKey)
}
