package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DIDMethod     = "panacea"
	Base58Charset = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

type SignableDID string

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (did SignableDID) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.Amino.MustMarshalJSON(did))
}

func NewDID(pubKey []byte) string {
	hash := sha256.New()
	_, err := hash.Write(pubKey)
	if err != nil {
		panic("failed to calculate SHA256 for DID")
	}
	idStr := base58.Encode(hash.Sum(nil))
	return fmt.Sprintf("did:%s:%s", DIDMethod, idStr)
}

func ParseDID(str string) (string, error) {
	did := str
	if !ValidDID(did) {
		return "", Error(ErrInvalidDID, str)
	}
	return did, nil
}

func ValidDID(did string) bool {
	pattern := fmt.Sprintf("^%s$", didRegex())
	matched, _ := regexp.MatchString(pattern, did)
	return matched
}

func didRegex() string {
	// https://www.w3.org/TR/did-core/#did-syntax
	return fmt.Sprintf("did:%s:[%s]{32,44}", DIDMethod, Base58Charset)
}

func EmptyDID(did string) bool {
	return did == ""
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func GetSignBytesDID(did string) []byte {
	return sdk.MustSortJSON(ModuleCdc.Amino.MustMarshalJSON(did))
}

func NewDIDDocument(id string, opts ...DIDDocumentOption) DIDDocument {
	doc := DIDDocument{
		Contexts: []string{ContextDIDV1},
		ID:       id,
	}

	for _, opt := range opts {
		opt(&doc)
	}

	return doc
}


type DIDDocumentOption func(opts *DIDDocument)

func WithController(controller string) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.Controller = controller
	}
}

func WithVerificationMethods(verificationMethods []*VerificationMethod) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.VerificationMethods = verificationMethods
	}
}

func WithAuthentications(authentications []*VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.Authentications = authentications
	}
}

func WithAssertionMethods(assertionMethods []*VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.AssertionMethods = assertionMethods
	}
}

func WithKeyAgreements(keyAgreements []*VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.KeyAgreements = keyAgreements
	}
}

func WithCapabilityInvocations(capabilityInvocations []*VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.CapabilityInvocations = capabilityInvocations
	}
}

func WithCapabilityDelegations(capabilityDelegations []*VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.CapabilityDelegations = capabilityDelegations
	}
}

func WithServices(services []*Service) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.Services = services
	}
}

func (doc DIDDocument) Valid() bool {
	if doc.Empty() { // deactivated
		return true
	}

	if !ValidDID(doc.ID) || doc.VerificationMethods == nil || doc.Authentications == nil {
		return false
	}

	if !EmptyDID(doc.Controller) && !ValidDID(doc.Controller) {
		return false
	}

	if doc.Contexts == nil || !ValidContexts(doc.Contexts) {
		return false
	}

	for _, verificationMethod := range doc.VerificationMethods {
		if !verificationMethod.Valid(doc.ID) {
			return false
		}
	}

	if !doc.validVerificationRelationships(doc.Authentications) {
		return false
	}
	if !doc.validVerificationRelationships(doc.AssertionMethods) {
		return false
	}
	if !doc.validVerificationRelationships(doc.KeyAgreements) {
		return false
	}
	if !doc.validVerificationRelationships(doc.CapabilityInvocations) {
		return false
	}
	if !doc.validVerificationRelationships(doc.CapabilityDelegations) {
		return false
	}

	for _, service := range doc.Services {
		if !service.Valid() {
			return false
		}
	}

	return true
}

func (doc DIDDocument) validVerificationRelationships(relationships []*VerificationRelationship) bool {
	for _, relationship := range relationships {
		if !relationship.Valid(doc.ID) {
			return false
		}
		if !relationship.hasDedicatedMethod() {
			// if the relationship isn't a dedicated verification method,
			// the referenced verification method must be presented in the 'verificationMethod' property.
			if _, ok := doc.VerificationMethodByID(relationship.VerificationMethodID); !ok {
				return false
			}
		}
	}
	return true
}


func (doc DIDDocument) Empty() bool {
	return EmptyDID(doc.ID)
}

// TODO Duplicate with protobuf
/*func (doc DIDDocument) String() string {
	bz, _ := json.Marshal(doc)
	return string(bz)
}*/

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (doc DIDDocument) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&doc))
}

// VerificationMethodByID finds a VerificationMethod by ID.
// If the corresponding VerificationMethod doesn't exist, it returns a false.
func (doc DIDDocument) VerificationMethodByID(id string) (VerificationMethod, bool) {
	for _, verificationMethod := range doc.VerificationMethods {
		if verificationMethod.ID == id {
			return *verificationMethod, true
		}
	}
	return VerificationMethod{}, false
}

// VerificationMethodFrom finds a VerificationMethod from the slice of VerificationRelationship by its ID.
// There are two types of VerificationRelationship. If it has a dedicated VerificationMethod, it is returned as it is.
// If the relationship has only a ID of VerificationMethod, this function tries to find a corresponding VerificationMethod in the DIDDocument.
func (doc DIDDocument) VerificationMethodFrom(relationships []*VerificationRelationship, id string) (VerificationMethod, bool) {
	for _, relationship := range relationships {
		if relationship.VerificationMethodID == id {
			if relationship.hasDedicatedMethod() {
				return *relationship.DedicatedVerificationMethod, true
			} else {
				return doc.VerificationMethodByID(id)
			}
		}
	}

	return VerificationMethod{}, false
}

func ValidContexts(ctxs []string) bool {
	if ctxs == nil || len(ctxs) == 0 || ctxs[0] != ContextDIDV1 { // the 1st one must be ContextDIDV1
		return false
	}

	set := make(map[string]struct{}, len(ctxs))
	for _, ctx := range ctxs {
		_, dup := set[ctx] // check the duplication
		if dup || !ValidContext(ctx) {
			return false
		}
		set[ctx] = struct{}{}
	}
	return true
}

func MarshalJSONContexts(ctxs []string) ([]byte, error) {
	if len(ctxs) == 1 { // if only one, treat it as a single string
		return json.Marshal(ctxs[0])
	}
	return json.Marshal(ctxs) // if not, as a list
}

func UnmarshalJSONContexts(bz []byte, ctxs *[]string) error {
	var single string
	err := json.Unmarshal(bz, &single)
	if err == nil {
		ctxs = &[]string{}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(bz, &multiple); err != nil {
		return err
	}
	*ctxs = multiple
	return nil
}

const (
	ContextDIDV1  = "https://www.w3.org/ns/did/v1"
)

func ValidContext(ctx string) bool {
	return ctx != ""
}

func NewVerificationMethodID(did string, name string) string {
	// https://www.w3.org/TR/did-core/#fragment
	return fmt.Sprintf("%v#%s", did, name)
}

func ParseVerificationMethodID(id string, did string) (string, error) {
	methodID := id
	if !ValidVerificationMethodID(id, did) {
		return "", Error(ErrInvalidVerificationMethodID, id)
	}
	return methodID, nil
}

const (
	maxVerificationMethodIDLen = 128
)

func ValidVerificationMethodID(verificationMethodID string, did string) bool {
	prefix := fmt.Sprintf("%v#", did)
	if !strings.HasPrefix(verificationMethodID, prefix) {
		return false
	}

	// Limit the length because it can be used for keystore filenames.
	// Max filename length on Linux is usually 256 bytes.
	if len(verificationMethodID)-len(prefix) > maxVerificationMethodIDLen {
		return false
	}

	suffix := verificationMethodID[len(prefix):]
	matched, _ := regexp.MatchString(`^\S+$`, suffix) // no whitespace
	return matched
}

const (
	JSONWEBKEY_2020  = "JsonWebKey2020"
	ES256K_2019      = "EcdsaSecp256k1VerificationKey2019"
	ES256K_2018      = "Secp256k1VerificationKey2018" // deprecated
	ED25519_2018     = "Ed25519VerificationKey2018"
	BLS1281G1_2020   = "Bls12381G1Key2020"
	BLS1281G2_2020   = "Bls12381G2Key2020"
	GPG_2020         = "GpgVerificationKey2020"
	RSA_2018         = "RsaVerificationKey2018"
	X25519_2019      = "X25519KeyAgreementKey2019"
	SS256K_2019      = "SchnorrSecp256k1VerificationKey2019"
	ES256K_R_2020    = "EcdsaSecp256k1RecoveryMethod2020"
)

func ValidKeyType(keyType string) bool {
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

func NewVerificationMethod(id string, keyType string, controller string, pubKey []byte) VerificationMethod {
	return VerificationMethod{
		ID:           id,
		Type:         keyType,
		Controller:   controller,
		PubKeyBase58: base58.Encode(pubKey),
	}
}

func (pk VerificationMethod) Valid(did string) bool {
	if !ValidVerificationMethodID(pk.ID, did) || !ValidKeyType(pk.Type) {
		return false
	}

	pattern := fmt.Sprintf("^[%s]+$", Base58Charset)
	matched, _ := regexp.MatchString(pattern, pk.PubKeyBase58)
	return matched
}

func NewVerificationRelationship(verificationMethodID string) VerificationRelationship {
	return VerificationRelationship{VerificationMethodID: verificationMethodID, DedicatedVerificationMethod: nil}
}

func NewVerificationRelationshipDedicated(verificationMethod VerificationMethod) VerificationRelationship {
	return VerificationRelationship{VerificationMethodID: verificationMethod.ID, DedicatedVerificationMethod: &verificationMethod}
}

func (v VerificationRelationship) hasDedicatedMethod() bool {
	return v.DedicatedVerificationMethod != nil
}

func (v VerificationRelationship) Valid(did string) bool {
	if !ValidVerificationMethodID(v.VerificationMethodID, did) {
		return false
	}
	if v.DedicatedVerificationMethod != nil {
		if !v.DedicatedVerificationMethod.Valid(did) || v.DedicatedVerificationMethod.ID != v.VerificationMethodID {
			return false
		}
	}
	return true
}

func (v VerificationRelationship) MarshalJSON() ([]byte, error) {
	// if dedicated
	if v.DedicatedVerificationMethod != nil {
		return json.Marshal(v.DedicatedVerificationMethod)
	}
	// if not dedicated
	return json.Marshal(v.VerificationMethodID)
}

func (v *VerificationRelationship) UnmarshalJSON(bz []byte) error {
	// if not dedicated
	var verificationMethodID string
	err := json.Unmarshal(bz, &verificationMethodID)
	if err == nil {
		*v = NewVerificationRelationship(verificationMethodID)
		return nil
	}

	// if dedicated
	var verificationMethod VerificationMethod
	if err := json.Unmarshal(bz, &verificationMethod); err != nil {
		return err
	}
	*v = NewVerificationRelationshipDedicated(verificationMethod)
	return nil
}

func NewService(id string, type_ string, serviceEndpoint string) Service {
	return Service{ID: id, Type: type_, ServiceEndpoint: serviceEndpoint}
}

func (s Service) Valid() bool {
	return s.ID != "" && s.Type != "" && s.ServiceEndpoint != ""
}

func NewDIDDocumentWithSeq(doc *DIDDocument, seq uint64) DIDDocumentWithSeq {
	return DIDDocumentWithSeq{
		Document: doc,
		Seq:      seq,
	}
}

// Empty returns true if all members in DIDDocumentWithSeq are empty.
// The empty struct means that the entity doesn't exist.
func (d DIDDocumentWithSeq) Empty() bool {
	return d.Document == nil || d.Document.Empty() && d.Seq == InitialSequence
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
	return d.Document.Empty() && d.Seq != InitialSequence
}