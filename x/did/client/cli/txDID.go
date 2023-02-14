package cli

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/go-bip39"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
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
	flagInteractive = "interactive"
	baseDir         = "did_keystore"
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

			msg, verificationMethodID, err := newMsgCreateDID(fromAddress, privKey)
			if err != nil {
				return err
			}
			if err := savePrivKeyToKeyStore(verificationMethodID, privKey, inBuf); err != nil {
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

			verificationMethodID := args[1]
			if err := types.ValidateVerificationMethodID(verificationMethodID, did); err != nil {
				return err
			}
			if err != nil {
				return err
			}

			// read an input file of DID document
			doc, err := readDIDDocFrom(args[2])
			if err != nil {
				return err
			}

			// get private key and sign document with sequence
			privKey, err := getPrivKeyFromKeyStore(verificationMethodID, inBuf)
			if err != nil {
				return err
			}

			signedDocument, err := signUsingNextSequence(clientCtx, did, verificationMethodID, privKey, doc)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress()
			err = cmd.Flags().Set(flags.FlagFrom, fromAddress.String())
			if err != nil {
				return err
			}
			didDocument := types.NewDIDDocument(signedDocument, types.DidDocumentDataType)

			msg := types.NewMsgUpdateDID(did, didDocument, fromAddress.String())
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
		Use:   "deactivate-did [did] [key-id] [did-doc-path]",
		Short: "Deactivate a DID Document",
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

			verificationMethodID := args[1]
			if err := types.ValidateVerificationMethodID(verificationMethodID, did); err != nil {
				return err
			}
			if err != nil {
				return err
			}

			// read an input file of DID document
			doc, err := readDIDDocFrom(args[2])
			if err != nil {
				return err
			}

			// get private key and sign document with sequence
			privKey, err := getPrivKeyFromKeyStore(verificationMethodID, inBuf)
			if err != nil {
				return err
			}

			signedDocument, err := signUsingNextSequence(clientCtx, did, verificationMethodID, privKey, doc)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress()

			didDocument := types.NewDIDDocument(signedDocument, types.DidDocumentDataType)

			msg := types.NewMsgDeactivateDID(did, didDocument, fromAddress.String())
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
// It generates the minimal DID document which contains only one public key information for did ownership,
// so that it can be extended by MsgUpdateDID later.
func newMsgCreateDID(fromAddress sdk.AccAddress, privKey secp256k1.PrivKey) (types.MsgCreateDID, string, error) {
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(privKey))

	btcecPrivKey, btcecPubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKey.Bytes())

	newDid := types.NewDID(pubKey)
	verificationMethodID := types.NewVerificationMethodID(newDid, "ownership")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, newDid, btcecPubKey.SerializeUncompressed())
	authentication := types.NewVerification(verificationMethod, ariesdid.Authentication)
	createdTime := time.Now()

	document := types.NewDocument(newDid,
		ariesdid.WithVerificationMethod([]ariesdid.VerificationMethod{verificationMethod}),
		ariesdid.WithAuthentication([]ariesdid.Verification{authentication}),
		ariesdid.WithCreatedTime(createdTime),
	)

	documentBz, err := document.JSONBytes()
	if err != nil {
		return types.MsgCreateDID{}, "", err
	}

	signedDocument, err := types.SignDocument(documentBz, verificationMethodID, types.InitialSequence, btcecPrivKey)
	if err != nil {
		return types.MsgCreateDID{}, "", err
	}

	didDocument := types.NewDIDDocument(signedDocument, types.DidDocumentDataType)

	msg := types.NewMsgCreateDID(newDid, didDocument, fromAddress.String())
	if err := msg.ValidateBasic(); err != nil {
		return types.MsgCreateDID{}, "", err
	}
	return msg, verificationMethodID, nil
}

// readDIDDocFrom reads a DID document from a JSON file.
// It returns an error if the JSON file is invalid or the DID document is invalid.
func readDIDDocFrom(path string) ([]byte, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	document, err := ariesdid.ParseDocument(fileData)
	if err != nil {
		return nil, err
	}
	ti := time.Now()
	document.Updated = &ti

	documentBz, err := document.JSONBytes()
	if err != nil {
		return nil, err
	}

	return documentBz, nil
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

// signUsingNextSequence generates a signature using the current sequence stored in the blockchain.
func signUsingNextSequence(clientCtx client.Context, did, vmID string, privKey crypto.PrivKey, newDoc []byte) ([]byte, error) {

	// get stored did document sequence
	queryClient := types.NewQueryClient(clientCtx)
	params := &types.QueryDIDRequest{
		DidBase64: base64.StdEncoding.EncodeToString([]byte(did)),
	}
	res, err := queryClient.DID(context.Background(), params)
	if err != nil {
		return nil, err
	}

	storedDIDDocument := res.DidDocument
	document, err := ariesdid.ParseDocument(storedDIDDocument.Document)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrParseDocument, "failed to parse stored did document. error: %v", err)
	}
	sequence, err := strconv.ParseUint(document.Proof[0].Domain, 10, 64)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidSequence, "error: %v", err)
	}

	// make new document proof value to nil
	newDocument, err := ariesdid.ParseDocument(newDoc)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrParseDocument, "failed to parse new did document. error: %v", err)
	}
	newDocument.Proof = nil
	newDocumentBz, err := newDocument.JSONBytes()
	if err != nil {
		return nil, err
	}

	// get ecdsa private key from secp256k1 key
	btcecPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKey.Bytes())

	return types.SignDocument(newDocumentBz, vmID, sequence+1, btcecPrivKey)
}
