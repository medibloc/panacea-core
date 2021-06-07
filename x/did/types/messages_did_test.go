package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/types/testsuite"
	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/x/did/types"
)

type messageTestSuite struct {
	testsuite.TestSuite
}

func TestMessageTestSuite(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForAccount("panacea", "panaceapub")
	suite.Run(t, new(messageTestSuite))
}

func (suite messageTestSuite) TestMsgCreateDID() {
	doc := newDIDDocument()
	sig := []byte("my-sig")
	fromAddr := getFromAddress(suite)

	msg := types.NewMsgCreateDID(doc.ID, doc, doc.VerificationMethods[0].ID, sig, fromAddr.String())
	suite.Require().Equal(doc.ID, msg.DID)
	suite.Require().Equal(doc, *msg.Document)
	suite.Require().Equal(doc.VerificationMethods[0].ID, msg.VerificationMethodID)
	suite.Require().Equal(sig, msg.Signature)
	suite.Require().Equal(fromAddr.String(), msg.FromAddress)

	suite.Require().Equal(types.RouterKey, msg.Route())
	suite.Require().Equal("create_did", msg.Type())
	suite.Require().Nil(msg.ValidateBasic())
	suite.Require().Equal(1, len(msg.GetSigners()))
	suite.Require().Equal(fromAddr, msg.GetSigners()[0])

	suite.Require().Equal(`{"did":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","document":{"@context":"https://www.w3.org/ns/did/v1","assertionMethod":[{"controller":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key2","publicKeyBase58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"EcdsaSecp256k1VerificationKey2019"}],"authentication":["did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"],"id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","service":[{"id":"service1","serviceEndpoint":"https://example.org","type":"LinkedDomains"}],"verificationMethod":[{"controller":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm","id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1","publicKeyBase58":"qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b","type":"EcdsaSecp256k1VerificationKey2019"}]},"from_address":"panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq","signature":"bXktc2ln","verification_method_id":"did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"}`,
		string(msg.GetSignBytes()),
	)
}

func getFromAddress(suite messageTestSuite) sdk.AccAddress {
	fromAddr, err := sdk.AccAddressFromBech32("panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq")
	suite.Require().NoError(err)
	return fromAddr
}

func newDIDDocument() types.DIDDocument {
	did, _ := types.ParseDID("did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm")
	verificationMethodID := types.NewVerificationMethodID(did, "key1")
	pubKey, _ := secp256k1util.PubKeyFromBase58("qoRmLNBEXoaKDE8dKffMq2DBNxacTEfvbKRuFrccYW1b")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, did, secp256k1util.PubKeyBytes(pubKey))
	verificationMethods := []*types.VerificationMethod{&verificationMethod}
	verificationRelationship := types.NewVerificationRelationship(verificationMethods[0].ID)
	authentications := []*types.VerificationRelationship{&verificationRelationship}
	verificationRelationshipDedicated := types.NewVerificationRelationshipDedicated(
		types.NewVerificationMethod(
			types.NewVerificationMethodID(did, "key2"),
			types.ES256K_2019, did, secp256k1util.PubKeyBytes(pubKey),
		),
	)
	assertionMethods := []*types.VerificationRelationship{&verificationRelationshipDedicated}
	service := types.NewService("service1", "LinkedDomains", "https://example.org")
	services := []*types.Service{&service}

	return types.NewDIDDocument(
		did,
		types.WithVerificationMethods(verificationMethods),
		types.WithAuthentications(authentications),
		types.WithAssertionMethods(assertionMethods),
		types.WithServices(services),
	)
}
