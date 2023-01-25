package aries_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/aries-framework-go/component/storageutil/mem"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/jsonld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/signer"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ed25519signature2018"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/verifier"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	ldstore "github.com/hyperledger/aries-framework-go/pkg/store/ld"
	jsld "github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"
)

func TestJsonDoc(t *testing.T) {

	docJson := validDoc

	doc, err := did.ParseDocument([]byte(docJson))
	require.NoError(t, err)
	require.NotNil(t, doc)

	require.Equal(t, doc.ID, "did:example:21tDAKCERh95uGgKbJNHYp")

	hexDecodeValue, err := hex.DecodeString("02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71")
	eAuthentication := []did.Verification{
		{VerificationMethod: *did.NewVerificationMethodFromBytes("did:example:123456789abcdefghi#keys-1",
			"Secp256k1VerificationKey2018",
			"did:example:123456789abcdefghi",
			base58.Decode("H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV")), Relationship: did.Authentication},
		{VerificationMethod: did.VerificationMethod{
			ID:         "did:example:123456789abcdefghs#key3",
			Controller: "did:example:123456789abcdefghs",
			Type:       "RsaVerificationKey2018",
			Value:      hexDecodeValue,
		}, Relationship: did.Authentication, Embedded: true},
	}

	require.Equal(t, eAuthentication, doc.Authentication)

	docBz, err := doc.JSONBytes()
	require.NoError(t, err)

	doc2, err := did.ParseDocument(docBz)
	require.NoError(t, err)
	require.NotNil(t, doc2)
	require.Equal(t, doc, doc2)

}

const validDoc = `{
  "@context": ["https://w3id.org/did/v1"],
  "id": "did:example:21tDAKCERh95uGgKbJNHYp",
  "verificationMethod": [
    {
      "id": "did:example:123456789abcdefghi#keys-1",
      "type": "Secp256k1VerificationKey2018",
      "controller": "did:example:123456789abcdefghi",
      "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
    }
  ],
  "authentication": [
    "did:example:123456789abcdefghi#keys-1",
    {
      "id": "did:example:123456789abcdefghs#key3",
      "type": "RsaVerificationKey2018",
      "controller": "did:example:123456789abcdefghs",
      "publicKeyHex": "02b97c30de767f084ce3080168ee293053ba33b235d7116a3263d29f1450936b71"
    }
  ],
  "created": "2002-10-10T17:00:00Z"
}`

type httpRemoteDocumentLoader struct{}

func (l *httpRemoteDocumentLoader) LoadDocument(url string) (*jsld.RemoteDocument, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := jsld.DocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return &jsld.RemoteDocument{
		DocumentURL: url,
		Document:    doc,
	}, nil
}

func TestSign(t *testing.T) {
	// make document loader
	storeProvider := mem.NewProvider()
	contextStore, err := ldstore.NewContextStore(storeProvider)
	require.NoError(t, err)
	remoteProviderStore, err := ldstore.NewRemoteProviderStore(storeProvider)
	require.NoError(t, err)

	ctx, err := context.New(
		context.WithJSONLDContextStore(contextStore),
		context.WithJSONLDRemoteProviderStore(remoteProviderStore),
	)
	require.NoError(t, err)

	loader, err := ld.NewDocumentLoader(
		ctx,
		ld.WithRemoteDocumentLoader(&httpRemoteDocumentLoader{}),
	)
	require.NoError(t, err)

	// make doc sample
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)

	const (
		didContext      = "https://w3id.org/did/v1"
		securityContext = "https://w3id.org/security/v1"
	)
	vm := did.VerificationMethod{
		ID:         "did:method:abc" + "#key-1",
		Type:       "Ed25519VerificationKey2018",
		Controller: "did:method:abc",
		Value:      pubKey,
	}
	createdTime := time.Now()

	signerContext := &signer.Context{
		Creator:       "did:method:abc" + "#key-1",
		SignatureType: "Ed25519Signature2018",
	}

	didDoc := &did.Doc{
		Context:            []string{didContext, securityContext},
		ID:                 "did:method:abc",
		VerificationMethod: []did.VerificationMethod{vm},
		Created:            &createdTime,
	}

	jsonDoc, err := didDoc.JSONBytes()
	require.NoError(t, err)

	// sign doc
	docSigner := signer.New(ed25519signature2018.New(
		suite.WithSigner(getSigner(privKey))))

	signedDoc, err := docSigner.Sign(signerContext, jsonDoc, jsonld.WithDocumentLoader(loader))
	require.NoError(t, err)

	sigVerifier := ed25519signature2018.New(suite.WithVerifier(ed25519signature2018.NewPublicKeyVerifier()))

	doc, err := did.ParseDocument(signedDoc)
	require.Nil(t, err)
	require.NotNil(t, doc)
	err = doc.VerifyProof([]verifier.SignatureSuite{sigVerifier}, jsonld.WithDocumentLoader(loader))
	require.NoError(t, err)

	doc.Proof[0].ProofValue = []byte("invalid")
	err = doc.VerifyProof([]verifier.SignatureSuite{sigVerifier}, jsonld.WithDocumentLoader(loader))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "ed25519: invalid signature")
}

func getSigner(privKey []byte) *testSigner {
	return &testSigner{privateKey: privKey}
}

type testSigner struct {
	privateKey []byte
}

func (s *testSigner) Sign(doc []byte) ([]byte, error) {
	if l := len(s.privateKey); l != ed25519.PrivateKeySize {
		return nil, errors.New("ed25519: bad private key length")
	}

	return ed25519.Sign(s.privateKey, doc), nil
}
