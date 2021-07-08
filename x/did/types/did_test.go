package types_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/btcsuite/btcutil/base58"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/x/did/types"
)

func TestMain(m *testing.M) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("panacea", "panaceapub")
	config.Seal()

	os.Exit(m.Run())
}

func TestNewDID(t *testing.T) {
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))

	did := types.NewDID(pubKey)
	regex := fmt.Sprintf("^did:panacea:[%s]{32,44}$", types.Base58Charset)
	require.Regexp(t, regex, did)
}

func TestParseDID(t *testing.T) {
	str := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	did, err := types.ParseDID(str)
	require.NoError(t, err)
	require.EqualValues(t, str, did)

	str = "did:panacea:7Prd74ry1Uct87nZqL3n"
	_, err = types.ParseDID(str)

	require.ErrorIs(t, types.ErrInvalidDID, err)
}

func TestDID_Empty(t *testing.T) {
	require.True(t, types.EmptyDID(""))
	require.False(t, types.EmptyDID("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
}

func TestNewDIDDocument(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].Id)
	authentications := []types.VerificationRelationship{verificationRelationship}
	service := types.NewService("service1", "LinkedDomains", "https://service.org")
	services := []*types.Service{&service}

	doc := types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods), types.WithAuthentications(authentications), types.WithServices(services))
	require.True(t, doc.Valid())
	require.Equal(t, did, doc.Id)
	require.Empty(t, doc.Controller)
	require.EqualValues(t, verificationMethods, doc.VerificationMethods)
	require.EqualValues(t, authentications, doc.Authentications)
	require.Empty(t, doc.AssertionMethods)
	require.Empty(t, doc.KeyAgreements)
	require.Empty(t, doc.CapabilityInvocations)
	require.Empty(t, doc.CapabilityDelegations)
	require.EqualValues(t, services, doc.Services)
}

func TestDIDDocument_Empty(t *testing.T) {
	require.False(t, getValidDIDDocument().Empty())
	require.True(t, types.DIDDocument{}.Empty())
}

func TestDIDDocument_Invalid(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	invalidVerificationRelationship := types.NewVerificationRelationship(types.NewVerificationMethodID("invalid did", "key1"))
	invalidVerificationRelationships := []types.VerificationRelationship{
		invalidVerificationRelationship,
	}
	service := types.NewService("", "", "")
	invalidServices := []*types.Service{
		&service,
	}

	require.False(t, types.NewDIDDocument("invalid did").Valid())
	require.False(t, types.NewDIDDocument(did, types.WithController("invalid did")).Valid())
	require.False(t, types.NewDIDDocument(did, types.WithAuthentications(invalidVerificationRelationships)).Valid())
	require.False(t, types.NewDIDDocument(did, types.WithAssertionMethods(invalidVerificationRelationships)).Valid())
	require.False(t, types.NewDIDDocument(did, types.WithKeyAgreements(invalidVerificationRelationships)).Valid())
	require.False(t, types.NewDIDDocument(did, types.WithCapabilityInvocations(invalidVerificationRelationships)).Valid())
	require.False(t, types.NewDIDDocument(did, types.WithCapabilityDelegations(invalidVerificationRelationships)).Valid())
	require.False(t, types.NewDIDDocument(did, types.WithCapabilityDelegations(invalidVerificationRelationships)).Valid())
	require.False(t, types.NewDIDDocument(did, types.WithCapabilityDelegations(invalidVerificationRelationships)).Valid())
	require.False(t, types.NewDIDDocument(did, types.WithServices(invalidServices)).Valid())

}

func TestDIDDocument_VerificationMethodByID(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	doc := types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods))

	found, ok := doc.VerificationMethodByID(verificationMethodID)
	require.True(t, ok)
	require.Equal(t, *verificationMethods[0], found)

	_, ok = doc.VerificationMethodByID(types.NewVerificationMethodID(did, "key2"))
	require.False(t, ok)
}

func TestDIDDocument_VerificationMethodFrom(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].Id)
	authentications := []types.VerificationRelationship{verificationRelationship}
	doc := types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods), types.WithAuthentications(authentications))

	found, ok := doc.VerificationMethodFrom(doc.Authentications, verificationMethodID)
	require.True(t, ok)
	require.Equal(t, *verificationMethods[0], found)

	_, ok = doc.VerificationMethodFrom(doc.Authentications, types.NewVerificationMethodID(did, "key2"))
	require.False(t, ok)

	doc.Authentications = []types.VerificationRelationship{} // clear authentications
	_, ok = doc.VerificationMethodFrom(doc.Authentications, verificationMethodID)
	require.False(t, ok)
}

func TestContexts_Valid(t *testing.T) {
	require.False(t, types.ValidateContexts(types.JSONStringOrStrings{}))
	require.True(t, types.ValidateContexts(types.JSONStringOrStrings{types.ContextDIDV1}))
	require.True(t, types.ValidateContexts(types.JSONStringOrStrings{types.ContextDIDV1, "https://example.com"}))
	require.False(t, types.ValidateContexts(types.JSONStringOrStrings{"https://example.com", types.ContextDIDV1}))
	require.False(t, types.ValidateContexts(types.JSONStringOrStrings{types.ContextDIDV1, types.ContextDIDV1}))

	var ctxs types.JSONStringOrStrings = nil
	require.False(t, types.ValidateContexts(ctxs))
}

func TestContexts_MarshalJSON(t *testing.T) {
	bz, err := types.ModuleCdc.Amino.MarshalJSON(types.JSONStringOrStrings{types.ContextDIDV1})
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`"%v"`, types.ContextDIDV1), string(bz))

	bz, err = types.ModuleCdc.Amino.MarshalJSON(types.JSONStringOrStrings{types.ContextDIDV1, "https://example.com"})
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`["%v","%v"]`, types.ContextDIDV1, "https://example.com"), string(bz))
}

func TestContexts_UnmarshalJSON(t *testing.T) {
	var ctxs types.JSONStringOrStrings

	bz := []byte(fmt.Sprintf(`["%v","%v"]`, types.ContextDIDV1, "https://example.com"))
	require.NoError(t, types.ModuleCdc.Amino.UnmarshalJSON(bz, &ctxs))
	require.Equal(t, types.JSONStringOrStrings{types.ContextDIDV1, "https://example.com"}, ctxs)

	bz = []byte(fmt.Sprintf(`"%v"`, types.ContextDIDV1))
	require.NoError(t, types.ModuleCdc.Amino.UnmarshalJSON(bz, &ctxs))
	require.Equal(t, types.JSONStringOrStrings{types.ContextDIDV1}, ctxs)
}

func TestNewVerificationMethodID(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	expectedID := fmt.Sprintf("%s#key1", did)
	id := types.NewVerificationMethodID(did, "key1")
	require.True(t, types.ValidateVerificationMethodID(id, did))
	require.EqualValues(t, expectedID, id)

	id, err := types.ParseVerificationMethodID(expectedID, did)
	require.NoError(t, err)
	require.EqualValues(t, expectedID, id)
}

func TestVerificationMethodID_Valid(t *testing.T) {
	validate := types.ValidateVerificationMethodID

	// normal
	require.True(t, validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))

	// if suffix has whitespaces
	require.False(t, validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm# key1", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
	require.False(t, validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1 ", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))

	// if suffix is empty
	require.False(t, validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
	require.False(t, validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))

	// if prefix (DID) is invalid
	require.False(t, validate("invalid#key1", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
	require.False(t, validate("did:panacea:87nZqL3ny7aR7C7Prd74ry1Uctg46JamVbJgk8azVgUm#key1", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))

	// if suffix is too long
	var builder strings.Builder
	builder.WriteString("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#")
	for i := 0; i < types.MaxVerificationMethodIDLen+1; i++ {
		builder.WriteByte('k')
	}
	require.False(t, validate(builder.String(), "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
}

func TestKeyType_Valid(t *testing.T) {
	require.True(t, types.ValidateKeyType(types.ES256K_2019))
	require.True(t, types.ValidateKeyType("NewKeyType2021"))
	require.False(t, types.ValidateKeyType(""))
}

func TestNewVerificationMethod(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	pub := types.NewVerificationMethod(types.NewVerificationMethodID(did, "key1"), types.ES256K_2019, did, pubKey)
	require.True(t, pub.Valid(did))

	require.Equal(t, pubKey[:], base58.Decode(pub.PubKeyBase58))
}

func TestVerificationRelationship_Valid(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)

	auth := types.NewVerificationRelationship(verificationMethodID)
	require.True(t, auth.Valid(did))
	auth = types.NewVerificationRelationshipDedicated(verificationMethod)
	require.True(t, auth.Valid(did))

	auth = types.NewVerificationRelationship("invalid")
	require.False(t, auth.Valid(did))
	auth = types.NewVerificationRelationshipDedicated(types.VerificationMethod{Id: "invalid"})
	require.False(t, auth.Valid(did))
}

func TestVerificationRelationship_MarshalJSON(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)

	auth := types.NewVerificationRelationship(verificationMethodID)
	bz, err := auth.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`"%v"`, verificationMethodID), string(bz))

	auth = types.NewVerificationRelationshipDedicated(verificationMethod)
	bz, err = auth.MarshalJSON()
	require.NoError(t, err)
	regex := fmt.Sprintf(`{"id":"%v","type":"%v","controller":"%v","publicKeyBase58":"%v"}`, verificationMethodID, types.ES256K_2019, did, verificationMethod.PubKeyBase58)
	require.Regexp(t, regex, string(bz))
}

func TestVerificationRelationship_UnmarshalJSON(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)

	var auth types.VerificationRelationship
	bz := []byte(fmt.Sprintf(`"%v"`, verificationMethodID))
	require.NoError(t, auth.UnmarshalJSON(bz))
	require.Equal(t, types.NewVerificationRelationship(verificationMethodID), auth)
	require.True(t, auth.Valid(did))

	bz = []byte(fmt.Sprintf(`{"id":"%v","type":"%v","controller":"%v","publicKeyBase58":"%v"}`, verificationMethodID, types.ES256K_2019, did, verificationMethod.PubKeyBase58))
	require.NoError(t, auth.UnmarshalJSON(bz))
	require.Equal(t, types.NewVerificationRelationshipDedicated(verificationMethod), auth)
	require.True(t, auth.Valid(did))
}

func TestService_Valid(t *testing.T) {
	require.True(t, types.NewService("service1", "LinkedDomains", "https://domain.com").Valid())
	require.False(t, types.NewService("", "LinkedDomains", "https://domain.com").Valid())
	require.False(t, types.NewService("service1", "", "https://domain.com").Valid())
	require.False(t, types.NewService("service1", "LinkedDomains", "").Valid())
}

func TestDIDDocumentWithSeq_Empty(t *testing.T) {
	document := getValidDIDDocument()
	require.False(t, types.NewDIDDocumentWithSeq(&document, types.InitialSequence).Empty())
	require.True(t, types.DIDDocumentWithSeq{}.Empty())
}

func TestDIDDocumentWithSeq_Valid(t *testing.T) {
	doc := getValidDIDDocument()
	require.True(t, types.NewDIDDocumentWithSeq(&doc, types.InitialSequence).Valid())
	require.False(t, types.DIDDocumentWithSeq{
		Document: &types.DIDDocument{Id: "invalid_did"},
	}.Valid())
}

func TestDIDDocumentWithSeq_Deactivate(t *testing.T) {
	document := getValidDIDDocument()
	docWithSeq := types.NewDIDDocumentWithSeq(&document, types.InitialSequence)
	deactivated := docWithSeq.Deactivate(types.InitialSequence + 1)
	require.True(t, deactivated.Deactivated())
	require.False(t, deactivated.Empty())
	require.True(t, deactivated.Valid())
}

func getValidDIDDocument() types.DIDDocument {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].Id)
	authentications := []types.VerificationRelationship{verificationRelationship}
	return types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods), types.WithAuthentications(authentications))
}
