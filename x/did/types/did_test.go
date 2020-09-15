package types

import (
	"fmt"
	"strings"
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
	veriMethodID := NewVeriMethodID(did, "key1")
	veriMethod := NewVeriMethod(veriMethodID, ES256K, did, secp256k1.GenPrivKey().PubKey())

	doc := NewDIDDocument(did, veriMethod)
	require.True(t, doc.Valid())
	require.Equal(t, did, doc.ID)
	require.Equal(t, 1, len(doc.VeriMethods))
	require.Equal(t, veriMethod, doc.VeriMethods[0])
	require.Equal(t, 1, len(doc.Authentications))
	require.EqualValues(t, veriMethodID, doc.Authentications[0].VeriMethodID)
}

func TestDIDDocument_Empty(t *testing.T) {
	require.False(t, getValidDIDDocument().Empty())
	require.True(t, DIDDocument{}.Empty())
}

func TestDIDDocument_VeriMethodByID(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	veriMethodID := NewVeriMethodID(did, "key1")
	veriMethod := NewVeriMethod(veriMethodID, ES256K, did, secp256k1.GenPrivKey().PubKey())
	doc := NewDIDDocument(did, veriMethod)

	found, ok := doc.VeriMethodByID(veriMethodID)
	require.True(t, ok)
	require.Equal(t, veriMethod, found)

	_, ok = doc.VeriMethodByID(NewVeriMethodID(did, "key2"))
	require.False(t, ok)

	doc.Authentications = []Authentication{} // clear authentications
	_, ok = doc.VeriMethodByID(veriMethodID)
	require.False(t, ok)
}

func TestContexts_Valid(t *testing.T) {
	require.False(t, Contexts{}.Valid())
	require.True(t, Contexts{ContextDIDV1}.Valid())
	require.True(t, Contexts{ContextDIDV1, ContextSecurityV1}.Valid())
	require.False(t, Contexts{ContextSecurityV1, ContextDIDV1}.Valid())
	require.False(t, Contexts{ContextDIDV1, "https://something.com"}.Valid())
	require.False(t, Contexts{ContextDIDV1, ContextDIDV1}.Valid())

	var ctxs Contexts = nil
	require.False(t, ctxs.Valid())
}

func TestContexts_MarshalJSON(t *testing.T) {
	bz, err := didCodec.MarshalJSON(Contexts{ContextDIDV1})
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`"%v"`, ContextDIDV1), string(bz))

	bz, err = didCodec.MarshalJSON(Contexts{ContextDIDV1, ContextSecurityV1})
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`["%v","%v"]`, ContextDIDV1, ContextSecurityV1), string(bz))
}

func TestContexts_UnmarshalJSON(t *testing.T) {
	var ctxs Contexts

	bz := []byte(fmt.Sprintf(`["%v","%v"]`, ContextDIDV1, ContextSecurityV1))
	require.NoError(t, didCodec.UnmarshalJSON(bz, &ctxs))
	require.Equal(t, Contexts{ContextDIDV1, ContextSecurityV1}, ctxs)

	bz = []byte(fmt.Sprintf(`"%v"`, ContextDIDV1))
	require.NoError(t, didCodec.UnmarshalJSON(bz, &ctxs))
	require.Equal(t, Contexts{ContextDIDV1}, ctxs)
}

func TestNewVeriMethodID(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	expectedID := fmt.Sprintf("%s#key1", did)
	id := NewVeriMethodID(did, "key1")
	require.True(t, id.Valid(did))
	require.EqualValues(t, expectedID, id)

	id, err := ParseVeriMethodID(expectedID, did)
	require.NoError(t, err)
	require.EqualValues(t, expectedID, id)
}

func TestVeriMethodID_Valid(t *testing.T) {
	// normal
	require.True(t, VeriMethodID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))

	// if suffix has whitespaces
	require.False(t, VeriMethodID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP# key1").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))
	require.False(t, VeriMethodID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#key1 ").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))

	// if suffix is empty
	require.False(t, VeriMethodID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))
	require.False(t, VeriMethodID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))

	// if prefix (DID) is invalid
	require.False(t, VeriMethodID("invalid#key1").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))
	require.False(t, VeriMethodID("did:panacea:mainnet:KS5zGZt66Me8MCctZBYrP#key1").Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))

	// if suffix is too long
	var builder strings.Builder
	builder.WriteString("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP#")
	for i := 0; i < maxVeriMethodIDLen+1; i++ {
		builder.WriteByte('k')
	}
	require.False(t, VeriMethodID(builder.String()).Valid("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP"))
}

func TestKeyType_Valid(t *testing.T) {
	require.True(t, ES256K.Valid())
	require.False(t, KeyType("invalid").Valid())
}

func TestNewVeriMethod(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	pubKey := secp256k1.GenPrivKey().PubKey()
	pub := NewVeriMethod(NewVeriMethodID(did, "key1"), ES256K, did, pubKey)
	require.True(t, pub.Valid(did))

	expected := pubKey.(secp256k1.PubKeySecp256k1)
	require.Equal(t, expected[:], base58.Decode(pub.PubKeyBase58))
}

func TestAuthentication_Valid(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	veriMethodID := NewVeriMethodID(did, "key1")
	veriMethod := NewVeriMethod(veriMethodID, ES256K, did, secp256k1.GenPrivKey().PubKey())

	auth := Authentication{VeriMethodID: veriMethodID, DedicatedMethod: nil}
	require.True(t, auth.Valid(did))
	auth = Authentication{VeriMethodID: veriMethodID, DedicatedMethod: &veriMethod}
	require.True(t, auth.Valid(did))

	auth = Authentication{VeriMethodID: "invalid", DedicatedMethod: nil}
	require.False(t, auth.Valid(did))
	auth = Authentication{VeriMethodID: veriMethodID, DedicatedMethod: &VeriMethod{ID: "invalid"}}
	require.False(t, auth.Valid(did))
	auth = Authentication{VeriMethodID: NewVeriMethodID(did, "key2"), DedicatedMethod: &veriMethod}
	require.False(t, auth.Valid(did))
}

func TestAuthentication_MarshalJSON(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	veriMethodID := NewVeriMethodID(did, "key1")
	veriMethod := NewVeriMethod(veriMethodID, ES256K, did, secp256k1.GenPrivKey().PubKey())

	auth := newAuthentication(veriMethodID)
	bz, err := auth.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`"%v"`, veriMethodID), string(bz))

	auth = newAuthenticationDedicated(veriMethod)
	bz, err = auth.MarshalJSON()
	require.NoError(t, err)
	regex := fmt.Sprintf(`{"id":"%v","type":"%v","controller":"%v","publicKeyBase58":"%v"}`, veriMethodID, ES256K, did, veriMethod.PubKeyBase58)
	require.Regexp(t, regex, string(bz))
}

func TestAuthentication_UnmarshalJSON(t *testing.T) {
	did := DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")
	veriMethodID := NewVeriMethodID(did, "key1")
	veriMethod := NewVeriMethod(veriMethodID, ES256K, did, secp256k1.GenPrivKey().PubKey())

	var auth Authentication
	bz := []byte(fmt.Sprintf(`"%v"`, veriMethodID))
	require.NoError(t, auth.UnmarshalJSON(bz))
	require.Equal(t, newAuthentication(veriMethodID), auth)
	require.True(t, auth.Valid(did))

	bz = []byte(fmt.Sprintf(`{"id":"%v","type":"%v","controller":"%v","publicKeyBase58":"%v"}`, veriMethodID, ES256K, did, veriMethod.PubKeyBase58))
	require.NoError(t, auth.UnmarshalJSON(bz))
	require.Equal(t, newAuthenticationDedicated(veriMethod), auth)
	require.True(t, auth.Valid(did))
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
	veriMethodID := NewVeriMethodID(did, "key1")
	veriMethod := NewVeriMethod(veriMethodID, ES256K, did, secp256k1.GenPrivKey().PubKey())
	return NewDIDDocument(did, veriMethod)
}
