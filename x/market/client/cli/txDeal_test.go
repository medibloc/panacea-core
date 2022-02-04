package cli

import (
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/stretchr/testify/suite"
	"testing"
)

type txTestSuite struct {
	testsuite.TestSuite
}

func TestTxTestSuite(t *testing.T) {
	suite.Run(t, new(txTestSuite))
}

// TestNewMsgCreateDeal
// TODO: Test Client Command-Line MsgCreateDeal
func (suite *txTestSuite) TestCmdCreateDeal() {

}

//func (suite *txTestSuite) TestParseCreateDealFlag() {
//	var createDeal createDealInputs
//
//	testInputs := makeTestCreateDealInputs()
//
//	file, err := ioutil.ReadFile("./testdata/create_deal_file.json")
//	suite.Require().NoError(err)
//
//	getString, err := suite.fs.GetString(FlagDealFile)
//}

// TODO: Test Client Command-Line MsgSellData
func (suite *txTestSuite) TestCmdSellData() {

}

//func makeTestCreateDealInputs() createDealInputs {
//	return createDealInputs{
//		DataSchema:            []string{"https://www.json.ld"},
//		Budget:                "10000000umed",
//		MaxNumData:            10000,
//		TrustedDataValidators: []string{"panacea153rk89lyqnahmzfygy6ca6gs0n69g9sa8kdght"},
//	}
//}
