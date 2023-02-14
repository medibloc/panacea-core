package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/stretchr/testify/suite"

	"github.com/medibloc/panacea-core/v2/types/testsuite"
	"github.com/medibloc/panacea-core/v2/x/did/client/crypto"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

type txTestSuite struct {
	testsuite.TestSuite
}

func TestTxTestSuite(t *testing.T) {
	suite.Run(t, new(txTestSuite))
}

func (suite *txTestSuite) AfterTest(_, _ string) {
	err := os.RemoveAll(baseDir)
	suite.Require().NoError(err)
}

func (suite *txTestSuite) TestNewMsgCreateDID() {
	privKey, _ := crypto.GenSecp256k1PrivKey("", "")
	fromAddr, err := sdk.AccAddressFromBech32("panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq")
	suite.Require().NoError(err)

	// create a message
	msg, _, err := newMsgCreateDID(fromAddr, privKey)
	suite.Require().NoError(err)

	// check if verificationMethod is correct
	doc, err := ariesdid.ParseDocument(msg.Document.Document)
	suite.Require().NoError(err)
	verificationMethod := doc.VerificationMethod[0]
	value := verificationMethod.Value

	_, btcecPubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKey.Bytes())

	suite.Require().Equal(btcecPubKey.SerializeUncompressed(), value)

	// check if the signature can be verifiable with the initial sequence
	err = types.VerifyProof(*doc)
	suite.Require().NoError(err)
}

func (suite *txTestSuite) TestMsgUpdateDID() {
	privKey, _ := crypto.GenSecp256k1PrivKey("", "")
	fromAddr, err := sdk.AccAddressFromBech32("panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq")
	suite.Require().NoError(err)

	// create DID document
	msg, vmID, err := newMsgCreateDID(fromAddr, privKey)
	suite.Require().NoError(err)
	did := msg.Did

	_, err = suite.DIDMsgServer.CreateDID(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)

	// get DID document
	storedDIDDocument := suite.DIDKeeper.GetDIDDocument(suite.Ctx, did)

	document, err := ariesdid.ParseDocument(storedDIDDocument.Document)
	suite.Require().NoError(err)

	sequence, err := strconv.ParseUint(document.Proof[0].Domain, 10, 64)
	suite.Require().NoError(err)
	btcecPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKey.Bytes())

	// sign document with next sequence
	newDocument := document
	newDocument.CapabilityDelegation = append(newDocument.CapabilityDelegation, types.NewVerification(newDocument.VerificationMethod[0], ariesdid.CapabilityDelegation))
	newDocument.Proof = nil
	newDocumentBz, err := newDocument.JSONBytes()
	suite.Require().NoError(err)

	signedDocument, err := types.SignDocument(newDocumentBz, vmID, sequence+1, btcecPrivKey)
	suite.Require().NoError(err)

	didDocument := types.NewDIDDocument(signedDocument, types.DidDocumentDataType)

	updateMsg := types.NewMsgUpdateDID(did, didDocument, fromAddr.String())

	err = updateMsg.ValidateBasic()
	suite.Require().NoError(err)

	_, err = suite.DIDMsgServer.UpdateDID(sdk.WrapSDKContext(suite.Ctx), &updateMsg)
	suite.Require().NoError(err)
	updatedDIDDocument := suite.DIDKeeper.GetDIDDocument(suite.Ctx, did)

	resultDoc, err := ariesdid.ParseDocument(updatedDIDDocument.Document)
	suite.Require().NoError(err)
	suite.Equal(resultDoc.ID, newDocument.ID)
	suite.Equal(1, len(resultDoc.VerificationMethod))
	suite.Equal(resultDoc.Authentication, newDocument.Authentication)
	suite.Equal(resultDoc.CapabilityDelegation, newDocument.CapabilityDelegation)
	suite.Equal("1", resultDoc.Proof[0].Domain)

}

func (suite *txTestSuite) TestMsgDeactivateDID() {
	privKey, _ := crypto.GenSecp256k1PrivKey("", "")
	fromAddr, err := sdk.AccAddressFromBech32("panacea154p6kyu9kqgvcmq63w3vpn893ssy6anpu8ykfq")
	suite.Require().NoError(err)

	// create DID document
	msg, vmID, err := newMsgCreateDID(fromAddr, privKey)
	suite.Require().NoError(err)
	did := msg.Did

	_, err = suite.DIDMsgServer.CreateDID(sdk.WrapSDKContext(suite.Ctx), &msg)
	suite.Require().NoError(err)

	// get DID document
	storedDIDDocument := suite.DIDKeeper.GetDIDDocument(suite.Ctx, did)

	document, err := ariesdid.ParseDocument(storedDIDDocument.Document)
	suite.Require().NoError(err)

	sequence, err := strconv.ParseUint(document.Proof[0].Domain, 10, 64)
	suite.Require().NoError(err)
	btcecPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKey.Bytes())

	// sign document with next sequence
	newDocument := document
	newDocument.Proof = nil
	newDocumentBz, err := newDocument.JSONBytes()
	suite.Require().NoError(err)

	signedDocument, err := types.SignDocument(newDocumentBz, vmID, sequence+1, btcecPrivKey)
	suite.Require().NoError(err)

	didDocument := types.NewDIDDocument(signedDocument, types.DidDocumentDataType)

	deactivateMsg := types.NewMsgDeactivateDID(did, didDocument, fromAddr.String())

	err = deactivateMsg.ValidateBasic()
	suite.Require().NoError(err)

	_, err = suite.DIDMsgServer.DeactivateDID(sdk.WrapSDKContext(suite.Ctx), &deactivateMsg)
	suite.Require().NoError(err)
	deactivatedDIDDocument := suite.DIDKeeper.GetDIDDocument(suite.Ctx, did)
	suite.Require().Equal(deactivatedDIDDocument.Deactivated, true)
	suite.Require().Nil(deactivatedDIDDocument.Document)
	suite.Require().Equal(deactivatedDIDDocument.DocumentDataType, "")
}

func (suite *txTestSuite) TestReadBIP39ParamsFrom_NotInteractive() {
	mnemonic, passphrase, err := readBIP39ParamsFrom(false, nil)
	suite.Require().NoError(err)
	suite.Require().Empty(mnemonic)
	suite.Require().Empty(passphrase)
}

func (suite *txTestSuite) TestReadBIP39ParamsFrom() {
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

func (suite *txTestSuite) TestReadBIP39ParamsFrom_EmptyPassphrase() {
	inputMnemonic := "travel broken word scare punch suggest air behind process gather sick void potato double furnace"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\n\n", inputMnemonic,
	)))

	mnemonic, passphrase, err := readBIP39ParamsFrom(true, reader)
	suite.Require().NoError(err)
	suite.Require().Equal(inputMnemonic, mnemonic)
	suite.Require().Equal("", passphrase)
}

func (suite *txTestSuite) TestReadBIP39ParamsFrom_PassphraseNotMatched() {
	inputMnemonic := "travel broken word scare punch suggest air behind process gather sick void potato double furnace"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\npasswd1\npasswd2\n", inputMnemonic,
	)))

	_, _, err := readBIP39ParamsFrom(true, reader)
	suite.Require().Error(err, "passphrases don't match")
}

func (suite *txTestSuite) TestReadBIP39ParamsFrom_InvalidMnemonic() {
	inputMnemonic := "travel broken"
	reader := bufio.NewReader(strings.NewReader(fmt.Sprintf(
		"%s\npasswd1\npasswd1\n", inputMnemonic,
	)))

	_, _, err := readBIP39ParamsFrom(true, reader)
	suite.Require().Error(err, "invalid mnemonic")
}

func (suite *txTestSuite) TestSaveAndGetPrivKeyFromKeyStore() {
	verificationMethodID := "key1"
	privKey, _ := crypto.GenSecp256k1PrivKey("", "")

	reader := bufio.NewReader(strings.NewReader("mypassword1\nmypassword1\n"))
	suite.Require().NoError(savePrivKeyToKeyStore(verificationMethodID, privKey, reader))

	reader = bufio.NewReader(strings.NewReader("mypassword1\n"))
	privKeyLoaded, err := getPrivKeyFromKeyStore(verificationMethodID, reader)
	suite.Require().NoError(err)
	suite.Require().Equal(privKey, privKeyLoaded)
}

//func (suite *txTestSuite) TestReadDIDDocOneContext() {
//	suite.testReadDIDDocOneContext("./testdata/did_one_context.json")
//}

func (suite *txTestSuite) TestReadDIDDocOneContext_W3C() {
	suite.testReadDIDDocOneContext("./testdata/did_one_context_w3c.json")
}

func (suite *txTestSuite) testReadDIDDocOneContext(path string) {
	doc, err := readDIDDocFrom(path)
	suite.Require().NoError(err)
	document, err := ariesdid.ParseDocument(doc)
	suite.Require().NoError(err)
	contexts := document.Context
	suite.Require().Equal(1, len(contexts))
	suite.Require().Equal(ariesdid.ContextV1, contexts[0])
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.ID)
	suite.Require().Equal(1, len(document.VerificationMethod))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.VerificationMethod[0].Controller)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.VerificationMethod[0].ID)
	suite.Require().Equal(base58.Decode("hfiFwEqzHPx3RbQBmkgg4UEMtejfbL27CspYNKiVuURN"), document.VerificationMethod[0].Value)
	suite.Require().Equal("Secp256k1VerificationKey2018", document.VerificationMethod[0].Type)
	suite.Require().Equal(1, len(document.Authentication))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.Authentication[0].VerificationMethod.ID)
}

//func (suite *txTestSuite) TestReadDIDDocTwoContexts() {
//	suite.testReadDIDDocTwoContexts("./testdata/did_multi_context.json")
//}

func (suite *txTestSuite) TestReadDIDDocTwoContexts_W3C() {
	suite.testReadDIDDocTwoContexts("./testdata/did_multi_context_w3c.json")
}

func (suite *txTestSuite) testReadDIDDocTwoContexts(path string) {
	doc, err := readDIDDocFrom(path)
	suite.Require().NoError(err)
	document, err := ariesdid.ParseDocument(doc)
	suite.Require().NoError(err)
	contexts := document.Context
	suite.Require().Equal(2, len(contexts))
	suite.Require().Equal(ariesdid.ContextV1, contexts[0])
	suite.Require().Equal("https://medibloc.org/ko", contexts[1])
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.ID)
	suite.Require().Equal(1, len(document.VerificationMethod))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.VerificationMethod[0].Controller)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.VerificationMethod[0].ID)
	suite.Require().Equal(base58.Decode("hfiFwEqzHPx3RbQBmkgg4UEMtejfbL27CspYNKiVuURN"), document.VerificationMethod[0].Value)
	suite.Require().Equal("Secp256k1VerificationKey2018", document.VerificationMethod[0].Type)
	suite.Require().Equal(1, len(document.Authentication))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.Authentication[0].VerificationMethod.ID)
}

//func (suite *txTestSuite) TestReadDIDDocMultiRelationship() {
//	suite.testReadDIDDocMultiRelationship("./testdata/did_multi_authentication.json")
//}

func (suite *txTestSuite) TestReadDIDDocMultiRelationship_W3C() {
	suite.testReadDIDDocMultiRelationship("./testdata/did_multi_authentication_w3c.json")
}

func (suite *txTestSuite) TestValidateDocumentInvalidVerificationMethod() {
	_, err := readDIDDocFrom("./testdata/did_invalid_verification_method.json")
	suite.Require().Error(err)
	fmt.Println(err)
}

func (suite *txTestSuite) TestValidateDocumentInvalidAuthentication() {
	_, err := readDIDDocFrom("./testdata/did_invalid_authentication.json")
	suite.Require().Error(err)
	fmt.Println(err)
}

func (suite *txTestSuite) testReadDIDDocMultiRelationship(path string) {
	doc, err := readDIDDocFrom(path)

	suite.Require().NoError(err)
	document, err := ariesdid.ParseDocument(doc)
	suite.Require().NoError(err)
	contexts := document.Context
	suite.Require().Equal(2, len(contexts))
	suite.Require().Equal(ariesdid.ContextV1, contexts[0])
	suite.Require().Equal("https://medibloc.org/ko", contexts[1])
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.ID)
	suite.Require().Equal(2, len(document.VerificationMethod))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.VerificationMethod[0].Controller)
	suite.Require().Equal(base58.Decode("hfiFwEqzHPx3RbQBmkgg4UEMtejfbL27CspYNKiVuURN"), document.VerificationMethod[0].Value)
	suite.Require().Equal("Secp256k1VerificationKey2018", document.VerificationMethod[0].Type)
	suite.Require().Equal(2, len(document.Authentication))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.VerificationMethod[0].ID)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.Authentication[0].VerificationMethod.ID)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.Authentication[1].VerificationMethod.Controller)
	suite.Require().Equal("Secp256k1VerificationKey2018", document.Authentication[1].VerificationMethod.Type)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key2", document.Authentication[1].VerificationMethod.ID)
	suite.Require().Equal(base58.Decode("zH3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"), document.Authentication[1].VerificationMethod.Value)

	suite.Require().Equal(3, len(document.AssertionMethod))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.AssertionMethod[0].VerificationMethod.ID)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.AssertionMethod[1].VerificationMethod.Controller)
	suite.Require().Equal("Secp256k1VerificationKey2018", document.AssertionMethod[1].VerificationMethod.Type)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key2", document.AssertionMethod[1].VerificationMethod.ID)
	suite.Require().Equal(base58.Decode("aH3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPs"), document.AssertionMethod[1].VerificationMethod.Value)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.AssertionMethod[2].VerificationMethod.Controller)
	suite.Require().Equal("Secp256k1VerificationKey2018", document.AssertionMethod[2].VerificationMethod.Type)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key3", document.AssertionMethod[2].VerificationMethod.ID)
	suite.Require().Equal(base58.Decode("bH3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPo"), document.AssertionMethod[2].VerificationMethod.Value)

	suite.Require().Equal(3, len(document.KeyAgreement))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.KeyAgreement[0].VerificationMethod.ID)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.KeyAgreement[1].VerificationMethod.Controller)
	suite.Require().Equal("Secp256k1VerificationKey2018", document.KeyAgreement[1].VerificationMethod.Type)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key2", document.KeyAgreement[1].VerificationMethod.ID)
	suite.Require().Equal(base58.Decode("oH3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPP"), document.KeyAgreement[1].VerificationMethod.Value)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key3", document.KeyAgreement[2].VerificationMethod.ID)

	suite.Require().Equal(2, len(document.CapabilityInvocation))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.CapabilityInvocation[0].VerificationMethod.ID)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.CapabilityInvocation[1].VerificationMethod.Controller)
	suite.Require().Equal("Secp256k1VerificationKey2018", document.CapabilityInvocation[1].VerificationMethod.Type)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key2", document.CapabilityInvocation[1].VerificationMethod.ID)
	suite.Require().Equal(base58.Decode("PH3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPp"), document.CapabilityInvocation[1].VerificationMethod.Value)

	suite.Require().Equal(2, len(document.CapabilityDelegation))
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key1", document.CapabilityDelegation[0].VerificationMethod.ID)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ", document.CapabilityDelegation[1].VerificationMethod.Controller)
	suite.Require().Equal("Secp256k1VerificationKey2018", document.CapabilityDelegation[1].VerificationMethod.Type)
	suite.Require().Equal("did:panacea:27FnaDeQZApXhsRZZDARhWYs2nKFaw3p7evGd9zUSrBZ#key2", document.CapabilityDelegation[1].VerificationMethod.ID)
	suite.Require().Equal(base58.Decode("qH3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPQ"), document.CapabilityDelegation[1].VerificationMethod.Value)
}
