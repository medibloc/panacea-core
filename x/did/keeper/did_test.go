package keeper_test

import (
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

type didTestSuite struct {
	testsuite.TestSuite
}

func TestDIDTestSuite(t *testing.T) {
	suite.Run(t, new(didTestSuite))
}

func (suite *didTestSuite) TestSetGetDIDDocument() {
	didKeeper := suite.DIDKeeper

	// Input two DIDDocument
	did1 := "did1:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	didDocument1, _, _ := makeTestDIDDocument(did1)
	didKeeper.SetDIDDocument(suite.Ctx, did1, didDocument1)
	did2 := "did1:panacea:1Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgap"
	didDocument2, _, _ := makeTestDIDDocument(did2)
	didKeeper.SetDIDDocument(suite.Ctx, did2, didDocument2)

	// Test one DIDDocument
	resDocument := didKeeper.GetDIDDocument(suite.Ctx, did1)
	suite.Require().NotNil(resDocument)
	suite.Require().Equal(didDocument1, &resDocument)

	// Test all DIDs
	resDIDs := didKeeper.ListDIDs(suite.Ctx)
	suite.Require().Equal(2, len(resDIDs))
	suite.Require().Equal(did2, resDIDs[0])
	suite.Require().Equal(did1, resDIDs[1])
}

func makeTestDIDDocument(id string) (*types.DIDDocument, *btcec.PrivateKey, string) {
	privKey := secp256k1.GenPrivKey()
	btcecPrivKey, btcecPubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKey.Bytes())

	verificationMethodID := types.NewVerificationMethodID(id, "key1")
	verificationMethod := types.NewVerificationMethod(verificationMethodID, types.ES256K_2019, id, btcecPubKey.SerializeUncompressed())

	authentication := types.NewVerification(verificationMethod, ariesdid.Authentication)

	createdTime := time.Now()

	doc := types.NewDocument(id,
		ariesdid.WithVerificationMethod([]ariesdid.VerificationMethod{verificationMethod}),
		ariesdid.WithAuthentication([]ariesdid.Verification{authentication}),
		ariesdid.WithCreatedTime(createdTime),
	)

	docBz, _ := doc.JSONBytes()

	document := &types.DIDDocument{
		Document:         docBz,
		DocumentDataType: types.DidDocumentDataType,
		Deactivated:      false,
	}

	return document, btcecPrivKey, verificationMethodID
}
