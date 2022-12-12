package crypto

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/require"
)

func TestDecryptWithAES256(t *testing.T) {
	privKey1, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)
	privKey2, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)

	data := []byte("hello, Panacea")

	shareKey1 := DeriveSharedKey(privKey1, privKey2.PubKey(), KDFSHA256)
	shareKey2 := DeriveSharedKey(privKey2, privKey1.PubKey(), KDFSHA256)

	nonce := make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)

	encryptedData, err := Encrypt(shareKey1, nonce, data)
	require.NoError(t, err)

	decryptedData, err := Decrypt(shareKey2, nonce, encryptedData)
	require.NoError(t, err)

	require.Equal(t, decryptedData, data)
}
