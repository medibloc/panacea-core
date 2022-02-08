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

// TODO: Test Client Command-Line MsgSellData
func (suite *txTestSuite) TestCmdSellData() {

}
