package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.RemoveAll(keystoreBaseDir())
	os.Exit(m.Run())
}

func TestReadBIP39ParamsFrom_NotInteractive(t *testing.T) {
	mnemonic, passphrase, err := readBIP39ParamsFrom(false, nil)
	require.NoError(t, err)
	require.Empty(t, mnemonic)
	require.Empty(t, passphrase)
}

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

func TestReadBIP39ParamsFrom_PassphraseNotMatched(t *testing.T) {
	inputMnemonic := "travel broken word scare punch suggest air behind process gather sick void potato double furnace"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\npasswd1\npasswd2\n", inputMnemonic,
	)))

	_, _, err := readBIP39ParamsFrom(true, reader)
	require.Error(t, err, "passphrases don't match")
}

func TestReadBIP39ParamsFrom_InvalidMnemonic(t *testing.T) {
	inputMnemonic := "travel broken"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\npasswd1\npasswd1\n", inputMnemonic,
	)))

	_, _, err := readBIP39ParamsFrom(true, reader)
	require.Error(t, err, "invalid mnemonic")
}

func TestSaveAndGetPrivKeyFromKeyStore(t *testing.T) {
	keyID := types.KeyID("key1")
	privKey := secp256k1.GenPrivKey()

	reader := bufio.NewReader(strings.NewReader("mypassword1\nmypassword1\n"))
	require.NoError(t, savePrivKeyToKeyStore(keyID, privKey, reader))

	reader = bufio.NewReader(strings.NewReader("mypassword1\n"))
	privKeyLoaded, err := getPrivKeyFromKeyStore(keyID, reader)
	require.NoError(t, err)
	require.Equal(t, privKey, privKeyLoaded)
}
