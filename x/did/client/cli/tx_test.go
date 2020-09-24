package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client/context"

	"github.com/medibloc/panacea-core/x/did/client/crypto"

	"github.com/medibloc/panacea-core/x/did/types"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	sdk.GetConfig().SetBech32PrefixForAccount("panacea", "panaceapub")
	os.RemoveAll(keystoreBaseDir())

	os.Exit(m.Run())
}

// Check if MsgCreateDID is created with the proper signature.
func TestNewMsgCreateDID(t *testing.T) {
	privKey, _ := crypto.GenSecp256k1PrivKey("", "")

	// create a message
	msg, err := newMsgCreateDID(getCliContext(t), privKey)
	require.NoError(t, err)

	// check if veriMethod is correct
	veriMethod, _ := msg.Document.VeriMethodByID(msg.VeriMethodID)
	pubKey, _ := types.NewPubKeyFromBase58(veriMethod.PubKeyBase58)
	require.Equal(t, privKey.PubKey(), pubKey)

	// check if the signature can be verifiable with the initial sequence
	_, ok := types.Verify(msg.Signature, msg.Document, types.InitialSequence, pubKey)
	require.True(t, ok)
}

// Check if empty strings are returned when the interactive mode is disabled.
func TestReadBIP39ParamsFrom_NotInteractive(t *testing.T) {
	mnemonic, passphrase, err := readBIP39ParamsFrom(false, nil)
	require.NoError(t, err)
	require.Empty(t, mnemonic)
	require.Empty(t, passphrase)
}

// Check if all input values are read correctly.
func TestReadBIP39ParamsFrom(t *testing.T) {
	inputMnemonic := "travel broken word scare punch suggest air behind process gather sick void potato double furnace"
	inputPassphrase := "mypasswd"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\n%s\n%s\n", inputMnemonic, inputPassphrase, inputPassphrase,
	)))

	mnemonic, passphrase, err := readBIP39ParamsFrom(true, reader)
	require.NoError(t, err)
	require.Equal(t, inputMnemonic, mnemonic)
	require.Equal(t, inputPassphrase, passphrase)
}

// Check if an empty passphrase are accepted.
func TestReadBIP39ParamsFrom_EmptyPassphrase(t *testing.T) {
	inputMnemonic := "travel broken word scare punch suggest air behind process gather sick void potato double furnace"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\n\n", inputMnemonic,
	)))

	mnemonic, passphrase, err := readBIP39ParamsFrom(true, reader)
	require.NoError(t, err)
	require.Equal(t, inputMnemonic, mnemonic)
	require.Equal(t, "", passphrase)
}

// Check if an error occurs when passphrases don't match.
func TestReadBIP39ParamsFrom_PassphraseNotMatched(t *testing.T) {
	inputMnemonic := "travel broken word scare punch suggest air behind process gather sick void potato double furnace"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\npasswd1\npasswd2\n", inputMnemonic,
	)))

	_, _, err := readBIP39ParamsFrom(true, reader)
	require.Error(t, err, "passphrases don't match")
}

// Check if an error occurs when mnemonic is invalid.
func TestReadBIP39ParamsFrom_InvalidMnemonic(t *testing.T) {
	inputMnemonic := "travel broken"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\npasswd1\npasswd1\n", inputMnemonic,
	)))

	_, _, err := readBIP39ParamsFrom(true, reader)
	require.Error(t, err, "invalid mnemonic")
}

// Check if the private key is stored and loaded correctly by the password specified.
func TestSaveAndGetPrivKeyFromKeyStore(t *testing.T) {
	veriMethodID := types.VeriMethodID("key1")
	privKey, _ := crypto.GenSecp256k1PrivKey("", "")

	reader := bufio.NewReader(strings.NewReader("mypassword1\nmypassword1\n"))
	require.NoError(t, savePrivKeyToKeyStore(veriMethodID, privKey, reader))

	reader = bufio.NewReader(strings.NewReader("mypassword1\n"))
	privKeyLoaded, err := getPrivKeyFromKeyStore(veriMethodID, reader)
	require.NoError(t, err)
	require.Equal(t, privKey, privKeyLoaded)
}

func getCliContext(t *testing.T) context.CLIContext {
	viper.Set(client.FlagTrustNode, true)
	fromAddr, err := sdk.AccAddressFromBech32("panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq")
	require.NoError(t, err)
	return context.NewCLIContext().WithFromAddress(fromAddr)
}
