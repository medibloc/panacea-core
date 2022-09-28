package crypto_test

import (
	"encoding/hex"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/did/client/crypto"
)

var (
	baseDir = "my_keystore"
	address = "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm#key1"
	passwd  = "nein-danke"
	priv    = secp256k1.GenPrivKey()
)

type keyStoreTestSuite struct {
	testsuite.TestSuite
}

func TestKeyStoreTestSuite(t *testing.T) {
	suite.Run(t, new(keyStoreTestSuite))
}

func (suite keyStoreTestSuite) AfterTest(_, _ string) {
	err := os.RemoveAll(baseDir)
	suite.Require().NoError(err)
}

// Check if the keystore can crypto a JSON provided by Web3 Secret Storage Definition
// https://github.com/ethereum/wiki/wiki/Web3-Secret-Storage-Definition#test-vectors
func (suite keyStoreTestSuite) TestKeyStore_DecryptWeb3() {
	ks := newKeyStore(suite)
	secret, err := ks.Load("testdata/web3.json", "testpassword")
	suite.Require().NoError(err)
	suite.Require().Equal("7a28b5ba57c53603b0b07b56bba752f7784bf506fa95edc395f5cf6c7514fe9d", hex.EncodeToString(secret))
}

func (suite keyStoreTestSuite) TestKeyStore_SaveAndLoad() {
	ks := newKeyStore(suite)

	path, err := ks.Save(address, priv[:], passwd)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(path)

	loadedPriv, err := ks.Load(path, passwd)
	suite.Require().NoError(err)
	suite.Require().Equal(priv[:], secp256k1.PrivKey(loadedPriv))
}

func (suite keyStoreTestSuite) TestKeyStore_Load_WithInvalidPath() {
	ks := newKeyStore(suite)
	path, _ := ks.Save(address, priv[:], passwd)
	_, err := ks.Load(path+path, passwd)
	suite.Require().Error(err)
}

func (suite keyStoreTestSuite) TestKeyStore_Load_WithInvalidPassword() {
	ks := newKeyStore(suite)
	path, _ := ks.Save(address, priv[:], passwd)
	_, err := ks.Load(path, passwd+passwd)
	suite.Require().Error(err)
}

func (suite keyStoreTestSuite) TestKeyStore_LoadByAddress_RecentFile() {
	ks := newKeyStore(suite)
	_, err := ks.Save(address, priv[:], passwd)
	suite.Require().NoError(err)

	newPriv := secp256k1.GenPrivKey()
	_, err = ks.Save(address, newPriv[:], passwd)
	suite.Require().NoError(err)

	privBytes, err := ks.LoadByAddress(address, passwd)
	suite.Require().NoError(err)
	suite.Require().Equal(newPriv[:], secp256k1.PrivKey(privBytes))
}

func (suite keyStoreTestSuite) TestKeyStore_LoadByAddress_NotExist() {
	ks := newKeyStore(suite)
	privBytes, err := ks.LoadByAddress(address, passwd)
	suite.Require().Error(err)
	suite.Require().Nil(privBytes)
}

func newKeyStore(suite keyStoreTestSuite) *crypto.KeyStore {
	ks, err := crypto.NewKeyStore(baseDir)
	suite.Require().NoError(err)
	suite.Require().NotNil(ks)
	return ks
}
