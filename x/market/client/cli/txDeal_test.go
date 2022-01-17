package cli_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

var acc1 = secp256k1.GenPrivKey().PubKey().Address()

type txTestSuite struct {
	testsuite.TestSuite

	cfg     network.Config
	network *network.Network
}

func TestTxTestSuite(t *testing.T) {
	suite.Run(t, new(txTestSuite))
}

// TestNewMsgCreateDeal
// TODO: Test Client Command-Line MsgCreateDeal
func (suite *txTestSuite) TestNewMsgCreateDeal() {

}
