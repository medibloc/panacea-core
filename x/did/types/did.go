package types

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
)

const (
	DIDMethod                  = "panacea"
	DidContext                 = "https://w3id.org/did/v1"
	SecurityContext            = "https://w3id.org/security/v1"
	DidDocumentDataType        = "github.com/hyperledger/aries-framework-go/pkg/doc/did.Doc@v0.1.8"
	InitialSequence            = 0
	Base58Charset              = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	MaxVerificationMethodIDLen = 128
)

func NewDID(pubKey []byte) string {
	hash := sha256.New()
	_, err := hash.Write(pubKey)
	if err != nil {
		panic("failed to calculate SHA256 for DID")
	}
	idStr := base58.Encode(hash.Sum(nil))
	return fmt.Sprintf("did:%s:%s", DIDMethod, idStr)
}

func ValidateDID(did string) error {
	if _, err := ariesdid.Parse(did); err != nil {
		return sdkerrors.Wrapf(ErrInvalidDID, "did: %v, error: %v", did, err)
	}
	return nil
}

func (doc DIDDocument) Valid() bool {
	if doc.Empty() {
		return true
	}
	document, err := ariesdid.ParseDocument(doc.Document)
	if err != nil {
		return false
	}

	// validation for document signing
	if document.Proof == nil {
		return false
	}
	if err := VerifyProof(*document); err != nil {
		return false
	}

	return true
}

func (doc DIDDocument) Empty() bool {
	return doc.DocumentDataType == "" || doc.Document == nil
}
func ValidateDocument(documentBz []byte) error {
	if _, err := ariesdid.ParseDocument(documentBz); err != nil {
		return err
	}
	return nil
}

func NewVerificationMethodID(did string, name string) string {
	// https://www.w3.org/TR/did-core/#fragment
	return fmt.Sprintf("%s#%s", did, name)
}

func ValidateVerificationMethodID(verificationMethodID string, did string) error {
	prefix := fmt.Sprintf("%v#", did)
	if !strings.HasPrefix(verificationMethodID, prefix) {
		return sdkerrors.Wrapf(ErrInvalidVerificationMethodID, "verificationMethodID: %v, did: %v", verificationMethodID, did)
	}

	// Limit the length because it can be used for keystore filenames.
	// Max filename length on Linux is usually 256 bytes.
	if len(verificationMethodID)-len(prefix) > MaxVerificationMethodIDLen {
		return sdkerrors.Wrapf(ErrInvalidVerificationMethodID, "verificationMethodID: %v, did: %v", verificationMethodID, did)
	}

	suffix := verificationMethodID[len(prefix):]
	// no whitespace
	if _, err := regexp.MatchString(`^\S+$`, suffix); err != nil {
		return sdkerrors.Wrapf(ErrInvalidVerificationMethodID, "verificationMethodID: %v, did: %v", verificationMethodID, did)
	}

	return nil
}

func NewVerificationMethod(verificationMethodID string, keyType string, did string, pubKey []byte) ariesdid.VerificationMethod {
	vm := ariesdid.VerificationMethod{
		ID:         verificationMethodID,
		Type:       keyType,
		Controller: did,
		Value:      pubKey,
	}

	return vm
}

func NewVerification(vm ariesdid.VerificationMethod, r ariesdid.VerificationRelationship) ariesdid.Verification {
	return ariesdid.Verification{
		VerificationMethod: vm,
		Relationship:       r,
	}
}

func NewService(id string, serviceType string, endpoint string) ariesdid.Service {
	return ariesdid.Service{
		ID:              id,
		Type:            serviceType,
		ServiceEndpoint: endpoint,
	}
}

func NewDocument(did string, opts ...ariesdid.DocOption) ariesdid.Doc {

	doc := ariesdid.BuildDoc(opts...)
	doc.ID = did
	doc.Context = []string{DidContext, SecurityContext}

	return *doc
}

func NewDIDDocument(documentBz []byte, documentDataType string) (DIDDocument, error) {

	if err := ValidateDocument(documentBz); err != nil {
		return DIDDocument{}, err
	}

	didDocument := DIDDocument{
		Document:         documentBz,
		DocumentDataType: documentDataType,
		Deactivated:      false,
	}

	return didDocument, nil
}

// verification key & signature type
const (
	JSONWEBKEY_2020 = "JsonWebKey2020"
	ES256K_2019     = "EcdsaSecp256k1VerificationKey2019"
	ES256K_2019_SIG = "EcdsaSecp256k1Signature2019"
	ES256K_2018     = "Secp256k1VerificationKey2018" // deprecated
	ED25519_2018    = "Ed25519VerificationKey2018"
	BLS1281G1_2020  = "Bls12381G1Key2020"
	BLS1281G2_2020  = "Bls12381G2Key2020"
	GPG_2020        = "GpgVerificationKey2020"
	RSA_2018        = "RsaVerificationKey2018"
	X25519_2019     = "X25519KeyAgreementKey2019"
	SS256K_2019     = "SchnorrSecp256k1VerificationKey2019"
	ES256K_R_2020   = "EcdsaSecp256k1RecoveryMethod2020"
)
