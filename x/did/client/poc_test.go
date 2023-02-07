package aries_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/aries-framework-go/component/storageutil/mem"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/jsonld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/signer"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ecdsasecp256k1signature2019"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/verifier"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	ldstore "github.com/hyperledger/aries-framework-go/pkg/store/ld"
	jsld "github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
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

	//privKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	//require.NoError(t, err)
	//pubKey := &privKey.PublicKey
	//pubKeyBytes := elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y)

	privKey := secp256k1.GenPrivKey()
	btcecPrivKey, btcecPubKey := btcec.PrivKeyFromBytes(btcec.S256(), privKey.Bytes())

	const (
		didContext      = "https://w3id.org/did/v1"
		securityContext = "https://w3id.org/security/v1"
	)
	vm := did.VerificationMethod{
		ID:         "did:method:abc" + "#key1",
		Type:       "EcdsaSecp256k1VerificationKey2019",
		Controller: "did:method:abc",
		Value:      btcecPubKey.SerializeUncompressed(),
	}
	createdTime := time.Now()

	signerContext := &signer.Context{
		Creator:       "did:method:abc" + "#key1",
		SignatureType: "EcdsaSecp256k1Signature2019",
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
	docSigner := signer.New(ecdsasecp256k1signature2019.New(
		suite.WithSigner(getSigner(btcecPrivKey.ToECDSA()))))

	signedDoc, err := docSigner.Sign(signerContext, jsonDoc, jsonld.WithDocumentLoader(loader))
	require.NoError(t, err)
	sigVerifier := ecdsasecp256k1signature2019.New(suite.WithVerifier(ecdsasecp256k1signature2019.NewPublicKeyVerifier()))

	doc, err := did.ParseDocument(signedDoc)
	require.Nil(t, err)
	require.NotNil(t, doc)
	err = doc.VerifyProof([]verifier.SignatureSuite{sigVerifier}, jsonld.WithDocumentLoader(loader))
	require.NoError(t, err)

	doc.Proof[0].ProofValue = []byte("invalid")
	err = doc.VerifyProof([]verifier.SignatureSuite{sigVerifier}, jsonld.WithDocumentLoader(loader))
	require.NotNil(t, err)
}

func getSigner(privKey *ecdsa.PrivateKey) *testSigner {
	return &testSigner{privateKey: privKey}
}

//type testSigner struct {
//	privateKey secp256k1.PrivKey
//}

//func (s *testSigner) Sign(doc []byte) ([]byte, error) {
//	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), s.privateKey)
//	sig, err := priv.Sign(getSHA256(doc))
//	if err != nil {
//		return nil, fmt.Errorf("failed to sign: %w", err)
//	}
//	return serializeSig(sig), nil
//}

type testSigner struct {
	privateKey *ecdsa.PrivateKey
}

func (signer *testSigner) Sign(doc []byte) ([]byte, error) {
	hasher := crypto.SHA256.New()
	_, _ = hasher.Write(doc) //nolint:errcheck
	hashed := hasher.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, signer.privateKey, hashed)
	if err != nil {
		return nil, err
	}

	curveBits := signer.privateKey.Curve.Params().BitSize

	keyBytes := curveBits / 8
	if curveBits%8 > 0 {
		keyBytes++
	}

	copyPadded := func(source []byte, size int) []byte {
		dest := make([]byte, size)
		copy(dest[size-len(source):], source)

		return dest
	}
	return append(copyPadded(r.Bytes(), keyBytes), copyPadded(s.Bytes(), keyBytes)...), nil

}
