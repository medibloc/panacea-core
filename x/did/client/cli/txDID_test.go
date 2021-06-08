package cli

import (
	"bufio"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/types/testsuite"
	"github.com/medibloc/panacea-core/x/did/client/crypto"
	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/x/did/types"
)

type txTestSuite struct {
	testsuite.TestSuite
}

func TestTxTestSuite(t *testing.T) {
	suite.Run(t, new(txTestSuite))
}

func (suite txTestSuite) AfterTest(_, _ string) {
	err := os.RemoveAll(baseDir)
	suite.Require().NoError(err)
}

func (suite txTestSuite) TestNewMsgCreateDID() {
	privKey, _ := crypto.GenSecp256k1PrivKey("", "")
	fromAddr, err := sdk.AccAddressFromBech32("panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq")
	suite.Require().NoError(err)

	// create a message
	msg, err := newMsgCreateDID(fromAddr, privKey)
	suite.Require().NoError(err)

	// check if verificationMethod is correct
	verificationMethod, _ := msg.Document.VerificationMethodByID(msg.VerificationMethodID)
	pubKey, _ := secp256k1util.PubKeyFromBase58(verificationMethod.PubKeyBase58)
	suite.Require().Equal(privKey.PubKey(), pubKey)

	// check if the signature can be verifiable with the initial sequence
	_, ok := types.Verify(msg.Signature, msg.Document, types.InitialSequence, pubKey)
	suite.Require().True(ok)
}

func (suite txTestSuite) TestReadBIP39ParamsFrom_NotInteractive() {
	mnemonic, passphrase, err := readBIP39ParamsFrom(false, nil)
	suite.Require().NoError(err)
	suite.Require().Empty(mnemonic)
	suite.Require().Empty(passphrase)
}

func (suite txTestSuite) TestReadBIP39ParamsFrom() {
	inputMnemonic := "travel broken word scare punch suggest air behind process gather sick void potato double furnace"
	inputPassphrase := "mypasswd"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\n%s\n%s\n", inputMnemonic, inputPassphrase, inputPassphrase,
	)))

	mnemonic, passphrase, err := readBIP39ParamsFrom(true, reader)
	suite.Require().NoError(err)
	suite.Require().Equal(inputMnemonic, mnemonic)
	suite.Require().Equal(inputPassphrase, passphrase)
}

func (suite txTestSuite) TestReadBIP39ParamsFrom_EmptyPassphrase() {
	inputMnemonic := "travel broken word scare punch suggest air behind process gather sick void potato double furnace"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\n\n", inputMnemonic,
	)))

	mnemonic, passphrase, err := readBIP39ParamsFrom(true, reader)
	suite.Require().NoError(err)
	suite.Require().Equal(inputMnemonic, mnemonic)
	suite.Require().Equal("", passphrase)
}

func (suite txTestSuite) TestReadBIP39ParamsFrom_PassphraseNotMatched() {
	inputMnemonic := "travel broken word scare punch suggest air behind process gather sick void potato double furnace"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\npasswd1\npasswd2\n", inputMnemonic,
	)))

	_, _, err := readBIP39ParamsFrom(true, reader)
	suite.Require().Error(err, "passphrases don't match")
}

func (suite txTestSuite) TestReadBIP39ParamsFrom_InvalidMnemonic() {
	inputMnemonic := "travel broken"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\npasswd1\npasswd1\n", inputMnemonic,
	)))

	_, _, err := readBIP39ParamsFrom(true, reader)
	suite.Require().Error(err, "invalid mnemonic")
}

func (suite txTestSuite) TestSaveAndGetPrivKeyFromKeyStore() {
	verificationMethodID := "key1"
	privKey, _ := crypto.GenSecp256k1PrivKey("", "")

	reader := bufio.NewReader(strings.NewReader("mypassword1\nmypassword1\n"))
	suite.Require().NoError(savePrivKeyToKeyStore(verificationMethodID, privKey, reader))

	reader = bufio.NewReader(strings.NewReader("mypassword1\n"))
	privKeyLoaded, err := getPrivKeyFromKeyStore(verificationMethodID, reader)
	suite.Require().NoError(err)
	suite.Require().Equal(privKey, privKeyLoaded)
}

func (suite txTestSuite) TestReadDIDDocOneContext() {
	path := "./testdata/did_one_context.json"
	doc, err := readDIDDocFrom(path)

	suite.Require().NoError(err)
	contexts := *doc.Contexts
	suite.Require().Equal(1, len(contexts))
	suite.Require().Equal(types.ContextDIDV1, contexts[0])
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", doc.ID)
	suite.Require().Equal(1, len(doc.VerificationMethods))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", doc.VerificationMethods[0].Controller)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", doc.VerificationMethods[0].ID)
	suite.Require().Equal("hfiFwEqzHPx3RbQBmkgg4UEMtejfbL27CspYNKiVuURN", doc.VerificationMethods[0].PubKeyBase58)
	suite.Require().Equal("Secp256k1VerificationKey2018", doc.VerificationMethods[0].Type)
	suite.Require().Equal(1, len(doc.Authentications))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", doc.Authentications[0].VerificationMethodID)
}

func (suite txTestSuite) TestReadDIDDocTwoContexts() {
	path := "./testdata/did_multi_context.json"
	doc, err := readDIDDocFrom(path)

	suite.Require().NoError(err)
	contexts := *doc.Contexts
	suite.Require().Equal(2, len(contexts))
	suite.Require().Equal(types.ContextDIDV1, contexts[0])
	suite.Require().Equal("https://medibloc.org/ko", contexts[1])
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", doc.ID)
	suite.Require().Equal(1, len(doc.VerificationMethods))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", doc.VerificationMethods[0].Controller)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", doc.VerificationMethods[0].ID)
	suite.Require().Equal("hfiFwEqzHPx3RbQBmkgg4UEMtejfbL27CspYNKiVuURN", doc.VerificationMethods[0].PubKeyBase58)
	suite.Require().Equal("Secp256k1VerificationKey2018", doc.VerificationMethods[0].Type)
	suite.Require().Equal(1, len(doc.Authentications))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", doc.Authentications[0].VerificationMethodID)
}