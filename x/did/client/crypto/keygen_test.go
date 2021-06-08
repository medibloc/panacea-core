package crypto_test

import (
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/medibloc/panacea-core/types/testsuite"
	"github.com/medibloc/panacea-core/x/did/client/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type keygenTestSuite struct {
	testsuite.TestSuite
}

func TestKeygenTestSuite(t *testing.T) {
	suite.Run(t, new(keygenTestSuite))
}

func (suite keygenTestSuite) TestGenSecp256k1PrivKey() {
	privKey, err := crypto.GenSecp256k1PrivKey("", "")
	suite.Require().NoError(err)
	suite.Require().NotEqual(secp256k1.PrivKey{}, privKey)
}

func (suite keygenTestSuite) TestGenSecp256k1PrivKey_InvalidMnemonic() {
	privKey, err := crypto.GenSecp256k1PrivKey("dummy", "")
	suite.Require().Error(err, "invalid mnemonic: dummy")
	suite.Require().Equal(secp256k1.PrivKey{}, privKey)
}
