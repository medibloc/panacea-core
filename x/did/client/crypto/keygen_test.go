package crypto_test

import (
	"testing"

	"github.com/medibloc/panacea-core/x/did/client/crypto"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestGenSecp256k1PrivKey(t *testing.T) {
	privKey, err := crypto.GenSecp256k1PrivKey("", "")
	require.NoError(t, err)
	require.NotEqual(t, secp256k1.PrivKeySecp256k1{}, privKey)
}

func TestGenSecp256k1PrivKey_InvalidMnemonic(t *testing.T) {
	privKey, err := crypto.GenSecp256k1PrivKey("dummy", "")
	require.Error(t, err, "invalid mnemonic: dummy")
	require.Equal(t, secp256k1.PrivKeySecp256k1{}, privKey)
}
