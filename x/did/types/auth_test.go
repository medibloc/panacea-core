package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type authTestSuite struct {
	suite.Suite
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(authTestSuite))
}

func (suite *authTestSuite) TestMustGetSignBytesWithSeq() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	signableDID := SignableDID(did)
	signBytes := mustGetSignBytesWithSeq(signableDID, 100)
	suite.Require().NotNil(signBytes)

	var obj dataWithSeq
	suite.Require().NoError(ModuleCdc.Amino.UnmarshalJSON(signBytes, &obj))
	suite.Require().Equal(ModuleCdc.Amino.MustMarshalJSON(signableDID), []byte(obj.Data))
	suite.Require().Equal(uint64(100), obj.Seq)
}

func (suite *authTestSuite) TestSequence() {
	seq := InitialSequence
	suite.Require().Equal(uint64(0), seq)

	nextSeq := nextSequence(seq)
	suite.Require().Equal(uint64(0), seq)
	suite.Require().Equal(uint64(1), nextSeq)
}

func (suite *authTestSuite) TestSignVerify() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	signableDID := SignableDID(did)
	privKey := secp256k1.GenPrivKey()
	seq := uint64(100)

	sig, err := Sign(signableDID, seq, privKey)
	suite.Require().NoError(err)
	suite.Require().NotNil(sig)

	newSeq, ok := Verify(sig, signableDID, seq, privKey.PubKey())
	suite.Require().True(ok)
	suite.Require().Equal(seq+1, newSeq)
}

func (suite *authTestSuite) TestSignVerify_doInvalid() {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	signableDID := SignableDID(did)
	privKey := secp256k1.GenPrivKey()
	anotherPrivKey := secp256k1.GenPrivKey()
	seq := uint64(100)

	sig, err := Sign(signableDID, seq, privKey)
	suite.Require().NoError(err)
	suite.Require().NotNil(sig)

	newSeq, ok := Verify(sig, signableDID, seq, anotherPrivKey.PubKey())
	suite.Require().Equal(false, ok)
	suite.Require().Equal(uint64(0), newSeq)
}
