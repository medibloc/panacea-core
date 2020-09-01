package keys_test

import (
	"testing"

	"github.com/medibloc/panacea-core/x/did/client/keys"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestGeneratePrivKey(t *testing.T) {
	privKey, err := keys.GeneratePrivKey("", "")
	require.NoError(t, err)
	require.NotEqual(t, secp256k1.PrivKeySecp256k1{}, privKey)
}

func TestGeneratePrivKey_InvalidMnemonic(t *testing.T) {
	privKey, err := keys.GeneratePrivKey("dummy", "")
	require.Error(t, err, "invalid mnemonic: dummy")
	require.Equal(t, secp256k1.PrivKeySecp256k1{}, privKey)
}
