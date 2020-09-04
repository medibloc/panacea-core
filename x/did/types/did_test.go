package types

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/btcsuite/btcutil/base58"

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

func TestParseDID(t *testing.T) {
	str := "did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"
	did, err := ParseDID(str)
	require.NoError(t, err)
	require.EqualValues(t, str, did)

	str = "did:panacea:t1estnet:KS5zGZt66Me8MCctZBYrP"
	_, err = ParseDID(str)
	require.EqualError(t, err, ErrInvalidDID(str).Error())
}

func TestDID_Empty(t *testing.T) {
	require.True(t, DID("").Empty())
	require.False(t, DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP").Empty())
}

func TestDID_GetSignBytes(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	var did2 DID
	require.NoError(t, codec.New().UnmarshalJSON(did.GetSignBytes(), &did2))
	require.Equal(t, did, did2)
}

func TestNewNetworkID(t *testing.T) {
	id, err := NewNetworkID("mainnet")
	require.NoError(t, err)
	require.Equal(t, Mainnet, id)

	id, err = NewNetworkID("testnet")
	require.NoError(t, err)
	require.Equal(t, Testnet, id)

	_, err = NewNetworkID("testn124et")
	require.EqualError(t, err, ErrInvalidNetworkID("testn124et").Error())
}

func TestNewDIDDocument(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	keyID := NewKeyID(did, "key1")
	pubKey := NewPubKey(keyID, ES256K, secp256k1.GenPrivKey().PubKey())

	doc := NewDIDDocument(did, pubKey)
	require.True(t, doc.Valid())
	require.Equal(t, did, doc.ID)
	require.Equal(t, 1, len(doc.PubKeys))
	require.Equal(t, pubKey, doc.PubKeys[0])
	require.Equal(t, 1, len(doc.Authentications))
	require.EqualValues(t, keyID, doc.Authentications[0])

	pubKeyFound, ok := doc.PubKeyByID(keyID)
	require.True(t, ok)
	require.Equal(t, pubKey, *pubKeyFound)

	_, ok = doc.PubKeyByID("invalid_key_id")
	require.False(t, ok)
}

func TestDIDDocument_Empty(t *testing.T) {
	require.False(t, getValidDIDDocument().Empty())
	require.True(t, DIDDocument{}.Empty())
}

func TestNewKeyID(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	expectedID := fmt.Sprintf("%s#key1", did)
	id := NewKeyID(did, "key1")
	require.True(t, id.Valid(did))
	require.EqualValues(t, expectedID, id)

	id, err := ParseKeyID(expectedID, did)
	require.NoError(t, err)
	require.EqualValues(t, expectedID, id)
}

func TestKeyID_Valid(t *testing.T) {
	// normal
	require.True(t, KeyID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))

	// if suffix has whitespaces
	require.False(t, KeyID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP# key1").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))
	require.False(t, KeyID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1 ").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))

	// if suffix is empty
	require.False(t, KeyID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))
	require.False(t, KeyID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))

	// if prefix (DID) is invalid
	require.False(t, KeyID("invalid#key1").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))
	require.False(t, KeyID("did:panacea:mainnet:KS5zGZt66Me8MCctZBYrP#key1").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))
}

func TestKeyType_Valid(t *testing.T) {
	require.True(t, ES256K.Valid())
	require.False(t, KeyType("invalid").Valid())
}

func TestNewPubKey(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	pubKey := secp256k1.GenPrivKey().PubKey()
	pub := NewPubKey(NewKeyID(did, "key1"), ES256K, pubKey)
	require.True(t, pub.Valid(did))

	expected := pubKey.(secp256k1.PubKeySecp256k1)
	require.Equal(t, expected[:], base58.Decode(pub.KeyBase58))
}

func TestDIDDocumentWithSeq_Empty(t *testing.T) {
	require.False(t, NewDIDDocumentWithSeq(getValidDIDDocument(), InitialSequence).Empty())
	require.True(t, DIDDocumentWithSeq{}.Empty())
}

func TestDIDDocumentWithSeq_Valid(t *testing.T) {
	doc := getValidDIDDocument()
	require.True(t, NewDIDDocumentWithSeq(doc, InitialSequence).Valid())
	require.False(t, DIDDocumentWithSeq{
		Document: DIDDocument{ID: "invalid_did"},
	}.Valid())
}

func TestDIDDocumentWithSeq_Deactivate(t *testing.T) {
	docWithSeq := NewDIDDocumentWithSeq(getValidDIDDocument(), InitialSequence)
	deactivated := docWithSeq.Deactivate(InitialSequence + 1)
	require.True(t, deactivated.Deactivated())
	require.False(t, deactivated.Empty())
	require.True(t, deactivated.Valid())
}

func getValidDIDDocument() DIDDocument {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	keyID := NewKeyID(did, "key1")
	pubKey := NewPubKey(keyID, ES256K, secp256k1.GenPrivKey().PubKey())
	return NewDIDDocument(did, pubKey)
}
