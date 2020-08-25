package keystore_test

import (
	"os"
	"testing"

	"github.com/medibloc/panacea-core/x/did/client/keystore"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

var (
	baseDir = "my_keystore"
	address = "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh#key1"
	passwd  = "nein-danke"
	priv    = secp256k1.GenPrivKey()
)

func TestKeyStore_SaveAndLoad(t *testing.T) {
	ks := newKeyStore(t)

	path, err := ks.Save(address, priv[:], passwd)
	require.NoError(t, err)
	require.NotEmpty(t, path)

	loadedPriv, err := ks.Load(path, passwd)
	require.NoError(t, err)
	require.Equal(t, priv[:], loadedPriv)
}

func TestKeyStore_Load_WithInvalidPath(t *testing.T) {
	ks := newKeyStore(t)
	path, _ := ks.Save(address, priv[:], passwd)
	_, err := ks.Load(path+path, passwd)
	require.Error(t, err)
}

func TestKeyStore_Load_WithInvalidPassword(t *testing.T) {
	ks := newKeyStore(t)
	path, _ := ks.Save(address, priv[:], passwd)
	_, err := ks.Load(path, passwd+passwd)
	require.Error(t, err)
}

func TestKeyStore_LoadByAddress_RecentFile(t *testing.T) {
	ks := newKeyStore(t)
	_, err := ks.Save(address, priv[:], passwd)
	require.NoError(t, err)

	newPriv := secp256k1.GenPrivKey()
	_, err = ks.Save(address, newPriv[:], passwd)
	require.NoError(t, err)

	privBytes, err := ks.LoadByAddress(address, passwd)
	require.NoError(t, err)
	require.Equal(t, newPriv[:], privBytes)
}

func TestKeyStore_LoadByAddress_NotExist(t *testing.T) {
	ks := newKeyStore(t)
	privBytes, err := ks.LoadByAddress(address, passwd)
	require.Error(t, err)
	require.Nil(t, privBytes)
}

func newKeyStore(t *testing.T) *keystore.KeyStore {
	ks, err := keystore.NewKeyStore(baseDir)
	require.NoError(t, err)
	require.NotNil(t, ks)

	t.Cleanup(func() {
		os.RemoveAll(baseDir)
	})

	return ks
}
