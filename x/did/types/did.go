package types

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

const (
	DIDMethod     = "panacea"
	Base58Charset = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

type DID string

func NewDID(networkID NetworkID, pubKey crypto.PubKey, keyType KeyType) DID {
	idStr := newPubKeyBase58(pubKey, keyType, 16)
	return DID(fmt.Sprintf("did:%s:%s:%s", DIDMethod, networkID, idStr))
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
	return fmt.Sprintf("did:panacea:%s:[%s]{21,22}", networkIDRegex(), Base58Charset)
}

func (did DID) Empty() bool {
	return did == ""
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (did DID) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(did))
}

type NetworkID string

const (
	Mainnet NetworkID = "mainnet"
	Testnet NetworkID = "testnet"
)

func NewNetworkID(str string) (NetworkID, error) {
	switch NetworkID(str) {
	case Mainnet:
		return Mainnet, nil
	case Testnet:
		return Testnet, nil
	}
	return "", ErrInvalidNetworkID(str)
}

func networkIDRegex() string {
	return fmt.Sprintf("(%s|%s)", Mainnet, Testnet)
}

type DIDDocument struct {
	Contexts        Contexts         `json:"@context"`
	ID              DID              `json:"id"`
	VeriMethods     []VeriMethod     `json:"verificationMethod"`
	Authentications []Authentication `json:"authentication"`
}

func NewDIDDocument(id DID, veriMethod VeriMethod) DIDDocument {
	return DIDDocument{
		Contexts:        Contexts{ContextDIDV1},
		ID:              id,
		VeriMethods:     []VeriMethod{veriMethod},
		Authentications: []Authentication{newAuthentication(veriMethod.ID)},
	}
}

func (doc DIDDocument) Valid() bool {
	if doc.Empty() { // deactivated
		return true
	}

	if !doc.ID.Valid() || doc.VeriMethods == nil || doc.Authentications == nil {
		return false
	}

	if doc.Contexts == nil || !doc.Contexts.Valid() {
		return false
	}

	for _, veriMethod := range doc.VeriMethods {
		if !veriMethod.Valid(doc.ID) {
			return false
		}
	}

	for _, auth := range doc.Authentications {
		if !auth.Valid(doc.ID) {
			return false
		}
		if !auth.hasDedicatedMethod() {
			if _, ok := doc.VeriMethodByID(auth.VeriMethodID); !ok {
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

// VeriMethodByID finds a VeriMethod by ID.
// If the corresponding VeriMethod doesn't exist, it returns a false.
func (doc DIDDocument) VeriMethodByID(id VeriMethodID) (VeriMethod, bool) {
	//TODO: Sadly, Amino codec doesn't accept maps. Find the way to make this efficient.
	for _, auth := range doc.Authentications {
		if auth.VeriMethodID == id {
			for _, veriMethod := range doc.VeriMethods {
				if veriMethod.ID == id {
					return veriMethod, true
				}
			}
			return VeriMethod{}, false
		}
	}
	return VeriMethod{}, false
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

const (
	ContextDIDV1      Context = "https://www.w3.org/ns/did/v1"
	ContextSecurityV1 Context = "https://w3id.org/security/v1"
)

func (ctx Context) Valid() bool {
	switch ctx {
	case ContextDIDV1, ContextSecurityV1:
		return true
	default:
		return false
	}
}

type VeriMethodID string

func NewVeriMethodID(did DID, name string) VeriMethodID {
	// https://www.w3.org/TR/did-core/#fragment
	return VeriMethodID(fmt.Sprintf("%v#%s", did, name))
}

func ParseVeriMethodID(id string, did DID) (VeriMethodID, error) {
	methodID := VeriMethodID(id)
	if !methodID.Valid(did) {
		return "", ErrInvalidVeriMethodID(id)
	}
	return methodID, nil
}

const (
	maxVeriMethodIDLen = 128
)

func (id VeriMethodID) Valid(did DID) bool {
	prefix := fmt.Sprintf("%v#", did)
	if !strings.HasPrefix(string(id), prefix) {
		return false
	}

	// Limit the length because it can be used for keystore filenames.
	// Max filename length on Linux is usually 256 bytes.
	if len(string(id))-len(prefix) > maxVeriMethodIDLen {
		return false
	}

	suffix := string(id)[len(prefix):]
	matched, _ := regexp.MatchString(`^\S+$`, suffix) // no whitespace
	return matched
}

type KeyType string

const (
	ES256K KeyType = "Secp256k1VerificationKey2018"
)

func (t KeyType) Valid() bool {
	switch t {
	case ES256K:
		return true
	}
	return false
}

// TODO: support various types: https://www.w3.org/TR/did-core/#key-types-and-formats
type VeriMethod struct {
	ID           VeriMethodID `json:"id"`
	Type         KeyType      `json:"type"`
	Controller   DID          `json:"controller"`
	PubKeyBase58 string       `json:"publicKeyBase58"`
}

func NewVeriMethod(id VeriMethodID, keyType KeyType, controller DID, key crypto.PubKey) VeriMethod {
	return VeriMethod{
		ID:           id,
		Type:         keyType,
		Controller:   controller,
		PubKeyBase58: newPubKeyBase58(key, keyType, 0),
	}
}

func newPubKeyBase58(key crypto.PubKey, keyType KeyType, truncateLen int) string {
	switch keyType {
	case ES256K:
		return encodePubKeyES256K(key, truncateLen)
	}
	panic(fmt.Sprintf("unsupported pubkey type: %v", keyType))
}

func encodePubKeyES256K(key crypto.PubKey, truncateLen int) string {
	keyES256K := key.(secp256k1.PubKeySecp256k1)

	var k []byte
	if truncateLen > 0 {
		k = keyES256K[:truncateLen]
	} else {
		k = keyES256K[:]
	}

	return base58.Encode(k)
}

func (pk VeriMethod) Valid(did DID) bool {
	if !pk.ID.Valid(did) || !pk.Type.Valid() {
		return false
	}

	// TODO: support the other DID as a controller
	if pk.Controller != did {
		return false
	}

	pattern := fmt.Sprintf("^[%s]+$", Base58Charset)
	matched, _ := regexp.MatchString(pattern, pk.PubKeyBase58)
	return matched
}

type Authentication struct {
	VeriMethodID VeriMethodID
	// DedicatedMethod is not nil if it is only authorized for authentication
	// https://www.w3.org/TR/did-core/#example-18-authentication-property-containing-three-verification-methods
	DedicatedMethod *VeriMethod
}

func newAuthentication(veriMethodID VeriMethodID) Authentication {
	return Authentication{VeriMethodID: veriMethodID, DedicatedMethod: nil}
}

func newAuthenticationDedicated(veriMethod VeriMethod) Authentication {
	return Authentication{VeriMethodID: veriMethod.ID, DedicatedMethod: &veriMethod}
}

func (a Authentication) hasDedicatedMethod() bool {
	return a.DedicatedMethod != nil
}

func (a Authentication) Valid(did DID) bool {
	if !a.VeriMethodID.Valid(did) {
		return false
	}
	if a.DedicatedMethod != nil {
		if !a.DedicatedMethod.Valid(did) || a.DedicatedMethod.ID != a.VeriMethodID {
			return false
		}
	}
	return true
}

func (a Authentication) MarshalJSON() ([]byte, error) {
	// if dedicated
	if a.DedicatedMethod != nil {
		return json.Marshal(a.DedicatedMethod)
	}
	// if not dedicated
	return json.Marshal(a.VeriMethodID)
}

func (a *Authentication) UnmarshalJSON(bz []byte) error {
	// if not dedicated
	var veriMethodID VeriMethodID
	err := json.Unmarshal(bz, &veriMethodID)
	if err == nil {
		*a = newAuthentication(veriMethodID)
		return nil
	}

	// if dedicated
	var veriMethod VeriMethod
	if err := json.Unmarshal(bz, &veriMethod); err != nil {
		return err
	}
	*a = newAuthenticationDedicated(veriMethod)
	return nil
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

// NewPrivKeyFromBytes converts a byte slice into a Secp256k1 private key.
// It returns an error when the length of the input is invalid.
func NewPrivKeyFromBytes(bz []byte) (secp256k1.PrivKeySecp256k1, error) {
	var key secp256k1.PrivKeySecp256k1
	if len(bz) != len(key) {
		return key, fmt.Errorf("invalid Secp256k1 private key. len:%d, expected:%d", len(bz), len(key))
	}
	copy(key[:], bz)
	return key, nil
}

// NewPubKeyFromBase58 decodes a base58-encoded Secp256k1 public key.
// It returns an error when the length of the input is invalid.
func NewPubKeyFromBase58(b58 string) (secp256k1.PubKeySecp256k1, error) {
	var key secp256k1.PubKeySecp256k1
	decoded := base58.Decode(b58)
	if len(decoded) != len(key) {
		return key, fmt.Errorf("invalid Secp256k1 public key. len:%d, expected:%d", len(decoded), len(key))
	}
	copy(key[:], decoded)
	return key, nil
}
