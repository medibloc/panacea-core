package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/types/testsuite"
	"github.com/medibloc/panacea-core/x/did/types"
)

type queryDIDTestSuite struct {
	testsuite.TestSuite
}

func TestQueryDIDTestSuite(t *testing.T) {
	suite.Run(t, new(queryDIDTestSuite))
}

func (suite queryDIDTestSuite) TestDIDDocumentWithSeq() {
	didKeeper := suite.DIDKeeper
	did := "did1:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	docWithSeq, _ := makeTestDIDDocumentWithSeq(did)

	didKeeper.SetDIDDocument(suite.Ctx, did, docWithSeq)

	req := types.QueryDIDRequest{Did: did}
	res, err := didKeeper.DIDDocumentWithSeq(sdk.WrapSDKContext(suite.Ctx), &req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Equal(docWithSeq, *res.DidDocumentWithSeq)
}
