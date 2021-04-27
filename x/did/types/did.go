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

type DID string

func NewDID(pubKey []byte) DID {
	hash := sha256.New()
	_, err := hash.Write(pubKey)
	if err != nil {
		panic("failed to calculate SHA256 for DID")
	}
	idStr := base58.Encode(hash.Sum(nil))
	return DID(fmt.Sprintf("did:%s:%s", DIDMethod, idStr))
}

func ParseDID(str string) (DID, error) {
	did := DID(str)
	if !did.Valid() {
		return "", ErrInvalidDID(str)
	}
	return did, nil
}

func (did DID) Valid() bool {
	pattern := fmt.Sprintf("^%s$", didRegex())
	matched, _ := regexp.MatchString(pattern, string(did))
	return matched
}

func didRegex() string {
	// https://www.w3.org/TR/did-core/#did-syntax
	return fmt.Sprintf("did:%s:[%s]{32,44}", DIDMethod, Base58Charset)
}

func (did DID) Empty() bool {
	return did == ""
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (did DID) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(did))
}

type DIDDocument struct {
	Contexts              Contexts                   `json:"@context"`
	ID                    DID                        `json:"id"`
	Controller            DID                        `json:"controller,omitempty"`
	VerificationMethods   []VerificationMethod       `json:"verificationMethod,omitempty"`
	Authentications       []VerificationRelationship `json:"authentication,omitempty"`
	AssertionMethods      []VerificationRelationship `json:"assertionMethod,omitempty"`
	KeyAgreements         []VerificationRelationship `json:"keyAgreement,omitempty"`
	CapabilityInvocations []VerificationRelationship `json:"capabilityInvocation,omitempty"`
	CapabilityDelegations []VerificationRelationship `json:"capabilityDelegation,omitempty"`
	Services              []Service                  `json:"service,omitempty"`
}

func NewDIDDocument(id DID, opts ...DIDDocumentOption) DIDDocument {
	doc := DIDDocument{
		Contexts: Contexts{ContextDIDV1},
		ID:       id,
	}

	for _, opt := range opts {
		opt(&doc)
	}

	return doc
}

// DIDDocumentOption is for optional properties of DID Document
type DIDDocumentOption func(opts *DIDDocument)

func WithController(controller DID) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.Controller = controller
	}
}

func WithVerificationMethods(verificationMethods []VerificationMethod) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.VerificationMethods = verificationMethods
	}
}

func WithAuthentications(authentications []VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.Authentications = authentications
	}
}

func WithAssertionMethods(assertionMethods []VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.AssertionMethods = assertionMethods
	}
}

func WithKeyAgreements(keyAgreements []VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.KeyAgreements = keyAgreements
	}
}

func WithCapabilityInvocations(capabilityInvocations []VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.CapabilityInvocations = capabilityInvocations
	}
}

func WithCapabilityDelegations(capabilityDelegations []VerificationRelationship) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.CapabilityDelegations = capabilityDelegations
	}
}

func WithServices(services []Service) DIDDocumentOption {
	return func(opts *DIDDocument) {
		opts.Services = services
	}
}

func (doc DIDDocument) Valid() bool {
	if doc.Empty() { // deactivated
		return true
	}

	if !doc.ID.Valid() || doc.VerificationMethods == nil || doc.Authentications == nil {
		return false
	}

	if !doc.Controller.Empty() && !doc.Controller.Valid() {
		return false
	}

	if doc.Contexts == nil || !doc.Contexts.Valid() {
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

func (doc DIDDocument) validVerificationRelationships(relationships []VerificationRelationship) bool {
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
	return doc.ID.Empty()
}

func (doc DIDDocument) String() string {
	bz, _ := json.Marshal(doc)
	return string(bz)
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (doc DIDDocument) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(doc))
}

// VerificationMethodByID finds a VerificationMethod by ID.
// If the corresponding VerificationMethod doesn't exist, it returns a false.
func (doc DIDDocument) VerificationMethodByID(id VerificationMethodID) (VerificationMethod, bool) {
	for _, verificationMethod := range doc.VerificationMethods {
		if verificationMethod.ID == id {
			return verificationMethod, true
		}
	}
	return VerificationMethod{}, false
}

// VerificationMethodFrom finds a VerificationMethod from the slice of VerificationRelationship by its ID.
// There are two types of VerificationRelationship. If it has a dedicated VerificationMethod, it is returned as it is.
// If the relationship has only a ID of VerificationMethod, this function tries to find a corresponding VerificationMethod in the DIDDocument.
func (doc DIDDocument) VerificationMethodFrom(relationships []VerificationRelationship, id VerificationMethodID) (VerificationMethod, bool) {
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

type Contexts []Context

func (ctxs Contexts) Valid() bool {
	if ctxs == nil || len(ctxs) == 0 || ctxs[0] != ContextDIDV1 { // the 1st one must be ContextDIDV1
		return false
	}

	set := make(map[Context]struct{}, len(ctxs))
	for _, ctx := range ctxs {
		_, dup := set[ctx] // check the duplication
		if dup || !ctx.Valid() {
			return false
		}
		set[ctx] = struct{}{}
	}
	return true
}

func (ctxs Contexts) MarshalJSON() ([]byte, error) {
	if len(ctxs) == 1 { // if only one, treat it as a single string
		return json.Marshal(ctxs[0])
	}
	return json.Marshal([]Context(ctxs)) // if not, as a list
}

func (ctxs *Contexts) UnmarshalJSON(bz []byte) error {
	var single Context
	err := json.Unmarshal(bz, &single)
	if err == nil {
		*ctxs = Contexts{single}
		return nil
	}

	var multiple []Context
	if err := json.Unmarshal(bz, &multiple); err != nil {
		return err
	}
	*ctxs = multiple
	return nil
}

type Context string

// https://w3c.github.io/did-spec-registries/#context
const (
	ContextDIDV1 Context = "https://www.w3.org/ns/did/v1"
)

func (ctx Context) Valid() bool {
	// TODO: The context can be any URI string. But, don't validate it strictly yet until W3C finalizes the spec.
	return ctx != ""
}

type VerificationMethodID string

func NewVerificationMethodID(did DID, name string) VerificationMethodID {
	// https://www.w3.org/TR/did-core/#fragment
	return VerificationMethodID(fmt.Sprintf("%v#%s", did, name))
}

func ParseVerificationMethodID(id string, did DID) (VerificationMethodID, error) {
	methodID := VerificationMethodID(id)
	if !methodID.Valid(did) {
		return "", ErrInvalidVerificationMethodID(id)
	}
	return methodID, nil
}

const (
	maxVerificationMethodIDLen = 128
)

func (id VerificationMethodID) Valid(did DID) bool {
	prefix := fmt.Sprintf("%v#", did)
	if !strings.HasPrefix(string(id), prefix) {
		return false
	}

	// Limit the length because it can be used for keystore filenames.
	// Max filename length on Linux is usually 256 bytes.
	if len(string(id))-len(prefix) > maxVerificationMethodIDLen {
		return false
	}

	suffix := string(id)[len(prefix):]
	matched, _ := regexp.MatchString(`^\S+$`, suffix) // no whitespace
	return matched
}

type KeyType string

// https://w3c.github.io/did-spec-registries/#verification-method-types
const (
	JSONWEBKEY_2020 KeyType = "JsonWebKey2020"
	ES256K_2019     KeyType = "EcdsaSecp256k1VerificationKey2019"
	ES256K_2018     KeyType = "Secp256k1VerificationKey2018" // deprecated
	ED25519_2018    KeyType = "Ed25519VerificationKey2018"
	BLS1281G1_2020  KeyType = "Bls12381G1Key2020"
	BLS1281G2_2020  KeyType = "Bls12381G2Key2020"
	GPG_2020        KeyType = "GpgVerificationKey2020"
	RSA_2018        KeyType = "RsaVerificationKey2018"
	X25519_2019     KeyType = "X25519KeyAgreementKey2019"
	SS256K_2019     KeyType = "SchnorrSecp256k1VerificationKey2019"
	ES256K_R_2020   KeyType = "EcdsaSecp256k1RecoveryMethod2020"
)

func (t KeyType) Valid() bool {
	switch t {
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

	// TODO: Return false, after W3C finalizes the spec.
	if t == "" {
		return false
	}
	log.Printf("[warn] unknown key type: %s\n", t) // TODO: Use tendermint logger
	return true
}

type VerificationMethod struct {
	ID         VerificationMethodID `json:"id"`
	Type       KeyType              `json:"type"`
	Controller DID                  `json:"controller"`
	//TODO: support various pubkey representation (not fully-defined yet by W3C): https://w3c.github.io/did-spec-registries/#verification-method-types
	PubKeyBase58 string `json:"publicKeyBase58"`
}

func NewVerificationMethod(id VerificationMethodID, keyType KeyType, controller DID, pubKey []byte) VerificationMethod {
	return VerificationMethod{
		ID:           id,
		Type:         keyType,
		Controller:   controller,
		PubKeyBase58: base58.Encode(pubKey),
	}
}

func (pk VerificationMethod) Valid(did DID) bool {
	if !pk.ID.Valid(did) || !pk.Type.Valid() {
		return false
	}

	pattern := fmt.Sprintf("^[%s]+$", Base58Charset)
	matched, _ := regexp.MatchString(pattern, pk.PubKeyBase58)
	return matched
}

type VerificationRelationship struct {
	VerificationMethodID VerificationMethodID
	// DedicatedVerificationMethod is not nil if it is only authorized for this verification relationship
	// https://www.w3.org/TR/did-core/#authentication
	DedicatedVerificationMethod *VerificationMethod
}

func NewVerificationRelationship(verificationMethodID VerificationMethodID) VerificationRelationship {
	return VerificationRelationship{VerificationMethodID: verificationMethodID, DedicatedVerificationMethod: nil}
}

func NewVerificationRelationshipDedicated(verificationMethod VerificationMethod) VerificationRelationship {
	return VerificationRelationship{VerificationMethodID: verificationMethod.ID, DedicatedVerificationMethod: &verificationMethod}
}

func (v VerificationRelationship) hasDedicatedMethod() bool {
	return v.DedicatedVerificationMethod != nil
}

func (v VerificationRelationship) Valid(did DID) bool {
	if !v.VerificationMethodID.Valid(did) {
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
	var verificationMethodID VerificationMethodID
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

type Service struct {
	ID string `json:"id"`
	//TODO: check strictly after the spec is finalized: https://w3c.github.io/did-spec-registries/#service-types
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

func NewService(id string, type_ string, serviceEndpoint string) Service {
	return Service{ID: id, Type: type_, ServiceEndpoint: serviceEndpoint}
}

func (s Service) Valid() bool {
	return s.ID != "" && s.Type != "" && s.ServiceEndpoint != ""
}

// DIDDocumentWithSeq is for storing a Sequence along with a DIDDocument.
// The Sequence is used to make DID operations not replay-able. It's used to generate signatures of DID operations.
type DIDDocumentWithSeq struct {
	Document DIDDocument `json:"document"`
	Seq      Sequence    `json:"sequence"`
}

func NewDIDDocumentWithSeq(doc DIDDocument, seq Sequence) DIDDocumentWithSeq {
	return DIDDocumentWithSeq{
		Document: doc,
		Seq:      seq,
	}
}

// Empty returns true if all members in DIDDocumentWithSeq are empty.
// The empty struct means that the entity doesn't exist.
func (d DIDDocumentWithSeq) Empty() bool {
	return d.Document.Empty() && d.Seq == InitialSequence
}

func (d DIDDocumentWithSeq) Valid() bool {
	return d.Document.Valid()
}

// Deactivate creates a new DIDDocumentWithSeq with an empty DIDDocument (tombstone).
// Note that it requires a new sequence.
func (d DIDDocumentWithSeq) Deactivate(newSeq Sequence) DIDDocumentWithSeq {
	return NewDIDDocumentWithSeq(DIDDocument{}, newSeq)
}

// Deactivated returns true if the DIDDocument has been activated.
func (d DIDDocumentWithSeq) Deactivated() bool {
	return d.Document.Empty() && d.Seq != InitialSequence
}
