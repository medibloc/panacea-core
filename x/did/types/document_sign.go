package types

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"net/http"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/hyperledger/aries-framework-go/component/storageutil/mem"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/jsonld"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/signer"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ecdsasecp256k1signature2019"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/verifier"
	"github.com/hyperledger/aries-framework-go/pkg/framework/context"
	ldstore "github.com/hyperledger/aries-framework-go/pkg/store/ld"
	jsld "github.com/piprate/json-gold/ld"
)

func SignDocument(doc []byte, verificationID string, sequence uint64, privKey *btcec.PrivateKey) ([]byte, error) {

	signerContext := NewSignerContext(verificationID, ES256K_2019_SIG, sequence)
	documentSigner := GetECDSASigner(privKey)
	loader, err := GetDocumentLoader()
	if err != nil {
		return nil, fmt.Errorf("error to get document loader: %v", err)
	}

	signedDocument, err := documentSigner.Sign(signerContext, doc, jsonld.WithDocumentLoader(loader))
	if err != nil {
		return nil, fmt.Errorf("error to sign document: %v", err)
	}

	return signedDocument, nil
}

func VerifyProof(doc ariesdid.Doc) error {

	sigVerifier := ecdsasecp256k1signature2019.New(suite.WithVerifier(ecdsasecp256k1signature2019.NewPublicKeyVerifier()))
	loader, err := GetDocumentLoader()
	if err != nil {
		return err
	}
	if err := doc.VerifyProof([]verifier.SignatureSuite{sigVerifier}, jsonld.WithDocumentLoader(loader)); err != nil {
		return err
	}

	return nil
}

func GetDocumentLoader() (*ld.DocumentLoader, error) {
	storeProvider := mem.NewProvider()
	contextStore, err := ldstore.NewContextStore(storeProvider)
	if err != nil {
		return nil, err
	}

	remoteProviderStore, err := ldstore.NewRemoteProviderStore(storeProvider)
	if err != nil {
		return nil, err
	}
	ctx, err := context.New(
		context.WithJSONLDContextStore(contextStore),
		context.WithJSONLDRemoteProviderStore(remoteProviderStore),
	)
	if err != nil {
		return nil, err
	}

	loader, err := ld.NewDocumentLoader(
		ctx,
		ld.WithRemoteDocumentLoader(&httpRemoteDocumentLoader{}),
	)
	if err != nil {
		return nil, err
	}

	return loader, nil
}

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

func NewSignerContext(vmID string, signatureType string, sequence uint64) *signer.Context {
	return &signer.Context{
		SignatureType: signatureType,
		Creator:       vmID,
		Domain:        strconv.FormatUint(sequence, 10),
	}
}

func GetECDSASigner(privKey *btcec.PrivateKey) *signer.DocumentSigner {
	return signer.New(ecdsasecp256k1signature2019.New(
		suite.WithSigner(getSigner(privKey.ToECDSA()))))
}

type ecdsaSigner struct {
	privateKey *ecdsa.PrivateKey
}

func getSigner(privKey *ecdsa.PrivateKey) *ecdsaSigner {
	return &ecdsaSigner{privateKey: privKey}
}

func (signer *ecdsaSigner) Sign(doc []byte) ([]byte, error) {
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
