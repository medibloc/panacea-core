package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"io"
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
	if !ValidateDID(did) {
		return "", sdkerrors.Wrapf(ErrInvalidDID, "did: %v", str)
	}
	return did, nil
}

func ValidateDID(did string) bool {
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
		Contexts: &Contexts{ContextDIDV1},
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

	if !ValidateDID(doc.ID) || doc.VerificationMethods == nil || doc.Authentications == nil {
		return false
	}

	if !EmptyDID(doc.Controller) && !ValidateDID(doc.Controller) {
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

type Contexts []Context

func (contexts *Contexts) appendToItself(str string) {
	*contexts = append(*contexts, Context(str))
}

func (contexts Contexts) Size() (n int) {
	var l int
	_ = l
	if len(contexts) > 0 {
		for _, s := range contexts {
			l = len(s)
			n += 1 + l + sovDid(uint64(l))
		}
	}
	return n
}

func (contexts Contexts) MarshalTo(dAtA []byte) (int, error) {
	size := contexts.Size()
	return contexts.MarshalToSizedBuffer(dAtA[:size])
}

func (contexts Contexts) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	for iNdEx := len(contexts) - 1; iNdEx >= 0; iNdEx-- {
		i -= len(contexts[iNdEx])
		copy(dAtA[i:], contexts[iNdEx])
		i = encodeVarintDid(dAtA, i, uint64(len(contexts[iNdEx])))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (contexts *Contexts) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDid
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: DIDDocument: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DIDDocument: illegal tag %d (wire type %d)", fieldNum, wire)
		}

		if wireType != 2 {
			return fmt.Errorf("proto: wrong wireType = %d for field Contexts", wireType)
		}
		var stringLen uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDid
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			stringLen |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		intStringLen := int(stringLen)
		if intStringLen < 0 {
			return ErrInvalidLengthDid
		}
		postIndex := iNdEx + intStringLen
		if postIndex < 0 {
			return ErrInvalidLengthDid
		}
		if postIndex > l {
			return io.ErrUnexpectedEOF
		}

		switch fieldNum {
		case 1:
			contexts.appendToItself(string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDid(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthDid
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (contexts Contexts) Valid() bool {
	if contexts == nil || len(contexts) == 0 || contexts[0] != ContextDIDV1 { // the 1st one must be ContextDIDV1
		return false
	}

	set := make(map[Context]struct{}, len(contexts))
	for _, ctx := range contexts {
		_, dup := set[ctx] // check the duplication
		if dup || !ctx.Valid() {
			return false
		}
		set[ctx] = struct{}{}
	}
	return true
}

func (contexts Contexts) MarshalJSON() ([]byte, error) {
	if len(contexts) == 1 { // if only one, treat it as a single string
		return json.Marshal(contexts[0])
	}
	return json.Marshal([]Context(contexts)) // if not, as a list
}

func (contexts *Contexts) UnmarshalJSON(bz []byte) error {
	var single Context
	err := json.Unmarshal(bz, &single)
	if err == nil {
		*contexts = Contexts{single}
		return nil
	}

	var multiple []Context
	if err := json.Unmarshal(bz, &multiple); err != nil {
		return err
	}
	*contexts = multiple
	return nil
}

type Context string

func (ctx Context) Valid() bool {
	// TODO: The context can be any URI string. But, don't validate it strictly yet until W3C finalizes the spec.
	return ctx != ""
}

const (
	ContextDIDV1 = "https://www.w3.org/ns/did/v1"
)

func ValidateContext(ctx string) bool {
	return ctx != ""
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

const (
	maxVerificationMethodIDLen = 128
)

func ValidateVerificationMethodID(verificationMethodID string, did string) bool {
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

func NewVerificationMethod(id string, keyType string, controller string, pubKey []byte) VerificationMethod {
	return VerificationMethod{
		ID:           id,
		Type:         keyType,
		Controller:   controller,
		PubKeyBase58: base58.Encode(pubKey),
	}
}

func (pk VerificationMethod) Valid(did string) bool {
	if !ValidateVerificationMethodID(pk.ID, did) || !ValidateKeyType(pk.Type) {
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
	if !ValidateVerificationMethodID(v.VerificationMethodID, did) {
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
