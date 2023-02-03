package types

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	log "github.com/sirupsen/logrus"
)

const (
	DIDMethod                  = "panacea"
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

func ValidateDID(did string) (string, error) {
	_, err := ariesdid.Parse(did)
	if err != nil {
		return "", sdkerrors.Wrapf(ErrInvalidDID, "did: %v", did)
	}
	return did, nil
}

func (doc DIDDocument) Valid() bool {
	if doc.Empty() { // deactivated
		return true
	}

	_, err := ariesdid.ParseDocument(doc.Document)

	return err == nil
}

func (doc DIDDocument) Empty() bool {
	return doc.DocumentDataType == ""
}

func ValidateDocument(documentBz []byte) error {
	_, err := ariesdid.ParseDocument(documentBz)
	if err != nil {
		return err
	}
	return nil
}

func NewVerificationMethodID(did string, name string) string {
	// https://www.w3.org/TR/did-core/#fragment
	return fmt.Sprintf("%v#%s", did, name)
}

func ParseVerificationMethodID(id string, did string) (string, error) {
	methodID := id
	if !ValidateVerificationMethodID(id, did) {
		return "", sdkerrors.Wrapf(ErrInvalidVerificationMethodID, "verificationMethodID: %v, did: %v", id, did)
	}
	return methodID, nil
}

func ValidateVerificationMethodID(verificationMethodID string, did string) bool {
	prefix := fmt.Sprintf("%v#", did)
	if !strings.HasPrefix(verificationMethodID, prefix) {
		return false
	}

	// Limit the length because it can be used for keystore filenames.
	// Max filename length on Linux is usually 256 bytes.
	if len(verificationMethodID)-len(prefix) > MaxVerificationMethodIDLen {
		return false
	}

	suffix := verificationMethodID[len(prefix):]
	matched, _ := regexp.MatchString(`^\S+$`, suffix) // no whitespace
	return matched
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

	return *doc
}

func NewDIDDocument(document ariesdid.Doc, documentDataType string) (DIDDocument, error) {

	documentBz, err := document.JSONBytes()
	if err != nil {
		return DIDDocument{}, err
	}
	if err := ValidateDocument(documentBz); err != nil {
		return DIDDocument{}, err
	}

	didDocument := DIDDocument{
		Document:         documentBz,
		DocumentDataType: documentDataType,
	}

	return didDocument, nil
}

const (
	JSONWEBKEY_2020 = "JsonWebKey2020"
	ES256K_2019     = "EcdsaSecp256k1VerificationKey2019"
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

func ValidateKeyType(keyType string) bool {
	switch keyType {
	case JSONWEBKEY_2020,
		ES256K_2019,
		ES256K_2018,
		ED25519_2018,
		BLS1281G1_2020,
		BLS1281G2_2020,
		GPG_2020,
		RSA_2018,
		X25519_2019,
		SS256K_2019,
		ES256K_R_2020:
		return true
	}

	if keyType == "" {
		return false
	}
	log.Printf("[warn] unknown key type: %s\n", keyType) // TODO: Use tendermint logger
	return true
}

func NewDIDDocumentWithSeq(doc *DIDDocument, seq uint64) DIDDocumentWithSeq {
	return DIDDocumentWithSeq{
		Document: doc,
		Sequence: seq,
	}
}

// Empty returns true if all members in DIDDocumentWithSeq are empty.
// The empty struct means that the entity doesn't exist.
func (d DIDDocumentWithSeq) Empty() bool {
	return d.Document == nil || d.Document.Empty() && d.Sequence == InitialSequence
}

func (d DIDDocumentWithSeq) Valid() bool {
	return d.Document.Valid()
}

// Deactivate creates a new DIDDocumentWithSeq with an empty DIDDocument (tombstone).
// Note that it requires a new sequence.
func (d DIDDocumentWithSeq) Deactivate(newSeq uint64) DIDDocumentWithSeq {
	return NewDIDDocumentWithSeq(&DIDDocument{}, newSeq)
}

// Deactivated returns true if the DIDDocument has been activated.
func (d DIDDocumentWithSeq) Deactivated() bool {
	return d.Document.Empty() && d.Sequence != InitialSequence
}
