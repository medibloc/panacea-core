package types_test

import (
	"fmt"

	"strings"
	"testing"

	"github.com/btcsuite/btcutil/base58"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/types/testsuite"
	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/x/did/types"
)

type didTestSuite struct {
	testsuite.TestSuite
}

func TestDidTestSuite(t *testing.T) {
	suite.Run(t, new(didTestSuite))
}

func (suite didTestSuite) TestNewDID() {
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))

	did := types.NewDID(pubKey)
	regex := fmt.Sprintf("^did:panacea:[%s]{32,44}$", types.Base58Charset)
	suite.Require().Regexp(regex, did)
}

func (suite didTestSuite) TestParseDID() {
	str := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	did, err := types.ParseDID(str)
	suite.Require().NoError(err)
	suite.Require().EqualValues(str, did)

	str = "did:panacea:7Prd74ry1Uct87nZqL3n"
	_, err = types.ParseDID(str)

	suite.Require().ErrorIs(types.ErrInvalidDID, err)
}

func (suite didTestSuite) TestDID_Empty() {
	suite.Require().True(types.EmptyDID(""))
	suite.Require().False(types.EmptyDID("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
}

func (suite didTestSuite) TestDID_GetSignBytes() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	signableDID := types.SignableDID(did)
	var signableDID2 types.SignableDID
	err := types.ModuleCdc.Amino.UnmarshalJSON(signableDID.GetSignBytes(), &signableDID2)
	suite.Require().NoError(err)
	suite.Require().Equal(signableDID, signableDID2)
}

func (suite didTestSuite) TestNewDIDDocument() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].ID)
	authentications := []*types.VerificationRelationship{&verificationRelationship}
	service := types.NewService("service1", "LinkedDomains", "https://service.org")
	services := []*types.Service{&service}

	doc := types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods), types.WithAuthentications(authentications), types.WithServices(services))
	suite.Require().True(doc.Valid())
	suite.Require().Equal(did, doc.ID)
	suite.Require().Empty(doc.Controller)
	suite.Require().EqualValues(verificationMethods, doc.VerificationMethods)
	suite.Require().EqualValues(authentications, doc.Authentications)
	suite.Require().Empty(doc.AssertionMethods)
	suite.Require().Empty(doc.KeyAgreements)
	suite.Require().Empty(doc.CapabilityInvocations)
	suite.Require().Empty(doc.CapabilityDelegations)
	suite.Require().EqualValues(services, doc.Services)
}

func (suite didTestSuite) TestDIDDocument_Empty() {
	suite.Require().False(getValidDIDDocument().Empty())
	suite.Require().True(types.DIDDocument{}.Empty())
}

func (suite didTestSuite) TestDIDDocument_Invalid() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	invalidVerificationRelationship := types.NewVerificationRelationship(types.NewVerificationMethodID("invalid did", "key1"))
	invalidVerificationRelationships := []*types.VerificationRelationship{
		&invalidVerificationRelationship,
	}
	service := types.NewService("", "", "")
	invalidServices := []*types.Service{
		&service,
	}

	suite.Require().False(types.NewDIDDocument("invalid did").Valid())
	suite.Require().False(types.NewDIDDocument(did, types.WithController("invalid did")).Valid())
	suite.Require().False(types.NewDIDDocument(did, types.WithAuthentications(invalidVerificationRelationships)).Valid())
	suite.Require().False(types.NewDIDDocument(did, types.WithAssertionMethods(invalidVerificationRelationships)).Valid())
	suite.Require().False(types.NewDIDDocument(did, types.WithKeyAgreements(invalidVerificationRelationships)).Valid())
	suite.Require().False(types.NewDIDDocument(did, types.WithCapabilityInvocations(invalidVerificationRelationships)).Valid())
	suite.Require().False(types.NewDIDDocument(did, types.WithCapabilityDelegations(invalidVerificationRelationships)).Valid())
	suite.Require().False(types.NewDIDDocument(did, types.WithCapabilityDelegations(invalidVerificationRelationships)).Valid())
	suite.Require().False(types.NewDIDDocument(did, types.WithCapabilityDelegations(invalidVerificationRelationships)).Valid())
	suite.Require().False(types.NewDIDDocument(did, types.WithServices(invalidServices)).Valid())

}

func (suite didTestSuite) TestDIDDocument_VerificationMethodByID() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	doc := types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods))

	found, ok := doc.VerificationMethodByID(verificationMethodID)
	suite.Require().True(ok)
	suite.Require().Equal(*verificationMethods[0], found)

	_, ok = doc.VerificationMethodByID(types.NewVerificationMethodID(did, "key2"))
	suite.Require().False(ok)
}

func (suite didTestSuite) TestDIDDocument_VerificationMethodFrom() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].ID)
	authentications := []*types.VerificationRelationship{&verificationRelationship}
	doc := types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods), types.WithAuthentications(authentications))

	found, ok := doc.VerificationMethodFrom(doc.Authentications, verificationMethodID)
	suite.Require().True(ok)
	suite.Require().Equal(*verificationMethods[0], found)

	_, ok = doc.VerificationMethodFrom(doc.Authentications, types.NewVerificationMethodID(did, "key2"))
	suite.Require().False(ok)

	doc.Authentications = []*types.VerificationRelationship{} // clear authentications
	_, ok = doc.VerificationMethodFrom(doc.Authentications, verificationMethodID)
	suite.Require().False(ok)
}

func (suite didTestSuite) TestContexts_Valid() {
	suite.Require().False(types.ValidateContexts(types.JSONStringOrStrings{}))
	suite.Require().True(types.ValidateContexts(types.JSONStringOrStrings{types.ContextDIDV1}))
	suite.Require().True(types.ValidateContexts(types.JSONStringOrStrings{types.ContextDIDV1, "https://example.com"}))
	suite.Require().False(types.ValidateContexts(types.JSONStringOrStrings{"https://example.com", types.ContextDIDV1}))
	suite.Require().False(types.ValidateContexts(types.JSONStringOrStrings{types.ContextDIDV1, types.ContextDIDV1}))

	var ctxs types.JSONStringOrStrings = nil
	suite.Require().False(types.ValidateContexts(ctxs))
}

func (suite didTestSuite) TestContexts_MarshalJSON() {
	bz, err := types.ModuleCdc.Amino.MarshalJSON(types.JSONStringOrStrings{types.ContextDIDV1})
	suite.Require().NoError(err)
	suite.Require().Equal(fmt.Sprintf(`"%v"`, types.ContextDIDV1), string(bz))

	bz, err = types.ModuleCdc.Amino.MarshalJSON(types.JSONStringOrStrings{types.ContextDIDV1, "https://example.com"})
	suite.Require().NoError(err)
	suite.Require().Equal(fmt.Sprintf(`["%v","%v"]`, types.ContextDIDV1, "https://example.com"), string(bz))
}

func (suite didTestSuite) TestContexts_UnmarshalJSON() {
	var ctxs types.JSONStringOrStrings

	bz := []byte(fmt.Sprintf(`["%v","%v"]`, types.ContextDIDV1, "https://example.com"))
	suite.Require().NoError(types.ModuleCdc.Amino.UnmarshalJSON(bz, &ctxs))
	suite.Require().Equal(types.JSONStringOrStrings{types.ContextDIDV1, "https://example.com"}, ctxs)

	bz = []byte(fmt.Sprintf(`"%v"`, types.ContextDIDV1))
	suite.Require().NoError(types.ModuleCdc.Amino.UnmarshalJSON(bz, &ctxs))
	suite.Require().Equal(types.JSONStringOrStrings{types.ContextDIDV1}, ctxs)
}

func (suite didTestSuite) TestNewVerificationMethodID() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	expectedID := fmt.Sprintf("%s#key1", did)
	id := types.NewVerificationMethodID(did, "key1")
	suite.Require().True(types.ValidateVerificationMethodID(id, did))
	suite.Require().EqualValues(expectedID, id)

	id, err := types.ParseVerificationMethodID(expectedID, did)
	suite.Require().NoError(err)
	suite.Require().EqualValues(expectedID, id)
}

func (suite didTestSuite) TestVerificationMethodID_Valid() {
	validate := types.ValidateVerificationMethodID

	// normal
	suite.Require().True(validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))

	// if suffix has whitespaces
	suite.Require().False(validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm# key1", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
	suite.Require().False(validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1 ", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))

	// if suffix is empty
	suite.Require().False(validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
	suite.Require().False(validate("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))

	// if prefix (DID) is invalid
	suite.Require().False(validate("invalid#key1", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
	suite.Require().False(validate("did:panacea:87nZqL3ny7aR7C7Prd74ry1Uctg46JamVbJgk8azVgUm#key1", "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))

	// if suffix is too long
	var builder strings.Builder
	builder.WriteString("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#")
	for i := 0; i < types.MaxVerificationMethodIDLen+1; i++ {
		builder.WriteByte('k')
	}
	suite.Require().False(validate(builder.String(), "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"))
}

func (suite didTestSuite) TestKeyType_Valid() {
	suite.Require().True(types.ValidateKeyType(types.ES256K_2019))
	suite.Require().True(types.ValidateKeyType("NewKeyType2021"))
	suite.Require().False(types.ValidateKeyType(""))
}

func (suite didTestSuite) TestNewVerificationMethod() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	pub := types.NewVerificationMethod(types.NewVerificationMethodID(did, "key1"), types.ES256K_2019, did, pubKey)
	suite.Require().True(pub.Valid(did))

	suite.Require().Equal(pubKey[:], base58.Decode(pub.PubKeyBase58))
}

func (suite didTestSuite) TestVerificationRelationship_Valid() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)

	auth := types.VerificationRelationship{VerificationMethodID: verificationMethodID, DedicatedVerificationMethod: nil}
	suite.Require().True(auth.Valid(did))
	auth = types.VerificationRelationship{VerificationMethodID: verificationMethodID, DedicatedVerificationMethod: &verificationMethod}
	suite.Require().True(auth.Valid(did))

	auth = types.VerificationRelationship{VerificationMethodID: "invalid", DedicatedVerificationMethod: nil}
	suite.Require().False(auth.Valid(did))
	auth = types.VerificationRelationship{VerificationMethodID: verificationMethodID, DedicatedVerificationMethod: &types.VerificationMethod{ID: "invalid"}}
	suite.Require().False(auth.Valid(did))
	auth = types.VerificationRelationship{VerificationMethodID: types.NewVerificationMethodID(did, "key2"), DedicatedVerificationMethod: &verificationMethod}
	suite.Require().False(auth.Valid(did))
}

func (suite didTestSuite) TestVerificationRelationship_MarshalJSON() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)

	auth := types.NewVerificationRelationship(verificationMethodID)
	bz, err := auth.MarshalJSON()
	suite.Require().NoError(err)
	suite.Require().Equal(fmt.Sprintf(`"%v"`, verificationMethodID), string(bz))

	auth = types.NewVerificationRelationshipDedicated(verificationMethod)
	bz, err = auth.MarshalJSON()
	suite.Require().NoError(err)
	regex := fmt.Sprintf(`{"id":"%v","type":"%v","controller":"%v","publicKeyBase58":"%v"}`, verificationMethodID, types.ES256K_2019, did, verificationMethod.PubKeyBase58)
	suite.Require().Regexp(regex, string(bz))
}

func (suite didTestSuite) TestVerificationRelationship_UnmarshalJSON() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)

	var auth types.VerificationRelationship
	bz := []byte(fmt.Sprintf(`"%v"`, verificationMethodID))
	suite.Require().NoError(auth.UnmarshalJSON(bz))
	suite.Require().Equal(types.NewVerificationRelationship(verificationMethodID), auth)
	suite.Require().True(auth.Valid(did))

	bz = []byte(fmt.Sprintf(`{"id":"%v","type":"%v","controller":"%v","publicKeyBase58":"%v"}`, verificationMethodID, types.ES256K_2019, did, verificationMethod.PubKeyBase58))
	suite.Require().NoError(auth.UnmarshalJSON(bz))
	suite.Require().Equal(types.NewVerificationRelationshipDedicated(verificationMethod), auth)
	suite.Require().True(auth.Valid(did))
}

func (suite didTestSuite) TestService_Valid() {
	suite.Require().True(types.NewService("service1", "LinkedDomains", "https://domain.com").Valid())
	suite.Require().False(types.NewService("", "LinkedDomains", "https://domain.com").Valid())
	suite.Require().False(types.NewService("service1", "", "https://domain.com").Valid())
	suite.Require().False(types.NewService("service1", "LinkedDomains", "").Valid())
}

func (suite didTestSuite) TestDIDDocumentWithSeq_Empty() {
	document := getValidDIDDocument()
	suite.Require().False(types.NewDIDDocumentWithSeq(&document, types.InitialSequence).Empty())
	suite.Require().True(types.DIDDocumentWithSeq{}.Empty())
}

func (suite didTestSuite) TestDIDDocumentWithSeq_Valid() {
	doc := getValidDIDDocument()
	suite.Require().True(types.NewDIDDocumentWithSeq(&doc, types.InitialSequence).Valid())
	suite.Require().False(types.DIDDocumentWithSeq{
		Document: &types.DIDDocument{ID: "invalid_did"},
	}.Valid())
}

func (suite didTestSuite) TestDIDDocumentWithSeq_Deactivate() {
	document := getValidDIDDocument()
	docWithSeq := types.NewDIDDocumentWithSeq(&document, types.InitialSequence)
	deactivated := docWithSeq.Deactivate(types.InitialSequence + 1)
	suite.Require().True(deactivated.Deactivated())
	suite.Require().False(deactivated.Empty())
	suite.Require().True(deactivated.Valid())
}

func getValidDIDDocument() types.DIDDocument {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(secp256k1.GenPrivKey()))
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, pubKey)
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].ID)
	authentications := []*types.VerificationRelationship{&verificationRelationship}
	return types.NewDIDDocument(did, types.WithVerificationMethods(verificationMethods), types.WithAuthentications(authentications))
}
