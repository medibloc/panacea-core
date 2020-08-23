package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestNewDID(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	did := NewDID(Mainnet, pubKey, ES256K)
	regex := fmt.Sprintf("^did:panacea:mainnet:[%s]{21,22}$", Base58Charset)
	require.Regexp(t, regex, did)
}

func TestNewDIDFrom(t *testing.T) {
	str := "did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"
	did, err := NewDIDFrom(str)
	assert.NoError(t, err)
	assert.EqualValues(t, str, did)

	str = "did:panacea:t1estnet:KS5zGZt66Me8MCctZBYrP"
	_, err = NewDIDFrom(str)
	require.EqualError(t, err, ErrInvalidDID(str).Error())
}

func TestDID_Empty(t *testing.T) {
	assert.True(t, DID("").Empty())
	assert.False(t, DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP").Empty())
}

func TestNewNetworkID(t *testing.T) {
	id, err := NewNetworkID("mainnet")
	assert.NoError(t, err)
	assert.Equal(t, Mainnet, id)

	id, err = NewNetworkID("testnet")
	assert.NoError(t, err)
	assert.Equal(t, Testnet, id)

	_, err = NewNetworkID("testn124et")
	assert.EqualError(t, err, ErrInvalidNetworkID("testn124et").Error())
}

func TestNewDIDDocument(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	keyID := NewKeyID(did, "key1")
	pubKey := NewPubKey(keyID, ES256K, secp256k1.GenPrivKey().PubKey())

	doc := NewDIDDocument(did, pubKey)
	assert.True(t, doc.Valid())
	assert.Equal(t, did, doc.ID)
	assert.Equal(t, 1, len(doc.PubKeys))
	assert.Equal(t, pubKey, doc.PubKeys[0])
	assert.Equal(t, 1, len(doc.Authentications))
	assert.EqualValues(t, keyID, doc.Authentications[0])

	pubKeyFound, ok := doc.PubKeyByID(keyID)
	assert.True(t, ok)
	assert.Equal(t, pubKey, *pubKeyFound)

	_, ok = doc.PubKeyByID("invalid_key_id")
	assert.False(t, ok)
}

func TestDIDDocument_Empty(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	keyID := NewKeyID(did, "key1")
	pubKey := NewPubKey(keyID, ES256K, secp256k1.GenPrivKey().PubKey())
	doc := NewDIDDocument(did, pubKey)
	assert.False(t, doc.Empty())

	assert.True(t, DIDDocument{}.Empty())
}

func TestNewKeyID(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	expectedID := fmt.Sprintf("%s#key1", did)
	id := NewKeyID(did, "key1")
	assert.True(t, id.Valid())
	assert.EqualValues(t, expectedID, id)

	id, err := NewKeyIDFrom(expectedID)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedID, id)

	_, err = NewKeyIDFrom("invalid_id")
	assert.Error(t, err, ErrInvalidKeyID("invalid_id"))
}

func TestKeyType_Valid(t *testing.T) {
	assert.True(t, ES256K.Valid())
	assert.False(t, KeyType("invalid").Valid())
}

func TestPrivKey_Valid(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	keyID := NewKeyID(did, "key1")

	k := NewPrivKey(keyID, ES256K, secp256k1.GenPrivKey(), "passwd")
	assert.True(t, k.Valid())

	k = NewPrivKey("invalid_id", ES256K, secp256k1.GenPrivKey(), "passwd")
	assert.False(t, k.Valid())

	k = NewPrivKey(keyID, "invalid_type", secp256k1.GenPrivKey(), "passwd")
	assert.False(t, k.Valid())
}

func TestPrivKey_Decrypt(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	keyID := NewKeyID(did, "key1")
	privKey := secp256k1.GenPrivKey()
	k := NewPrivKey(keyID, ES256K, privKey, "passwd")

	decrypted, err := k.Decrypt("passwd")
	assert.NoError(t, err)
	assert.Equal(t, privKey, decrypted)

	_, err = k.Decrypt("wrong_passwd")
	assert.EqualError(t, err, "invalid account password")
}

func TestNewPubKey(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	keyID := NewKeyID(did, "key1")
	k := NewPubKey(did, ES256K, secp256k1.GenPrivKey().PubKey())
}
