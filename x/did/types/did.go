package types

import (
	"encoding/json"
	"fmt"
	"regexp"

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
	return sdk.MustSortJSON(didCodec.MustMarshalJSON(did))
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
	PubKeys         []PubKey         `json:"publicKey"`
	Authentications []Authentication `json:"authentication"`
}

func NewDIDDocument(id DID, pubKey PubKey) DIDDocument {
	return DIDDocument{
		Contexts:        Contexts{ContextDIDV1},
		ID:              id,
		PubKeys:         []PubKey{pubKey},
		Authentications: []Authentication{newAuthentication(pubKey.ID)},
	}
}

func (doc DIDDocument) Valid() bool {
	if doc.Empty() { // deactivated
		return true
	}

	if !doc.ID.Valid() || doc.PubKeys == nil || doc.Authentications == nil {
		return false
	}

	if doc.Contexts == nil || !doc.Contexts.Valid() {
		return false
	}

	for _, pubKey := range doc.PubKeys {
		if !pubKey.Valid(doc.ID) {
			return false
		}
	}

	for _, auth := range doc.Authentications {
		if !auth.Valid(doc.ID) {
			return false
		}
		if !auth.hasDedicatedPubKey() {
			if _, ok := doc.PubKeyByID(auth.KeyID); !ok {
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
	return sdk.MustSortJSON(didCodec.MustMarshalJSON(doc))
}

// PubKeyByID finds a PubKey by ID.
// If the corresponding PubKey doesn't exist, it returns a false.
func (doc DIDDocument) PubKeyByID(id KeyID) (PubKey, bool) {
	//TODO: Sadly, Amino codec doesn't accept maps. Find the way to make this efficient.
	for _, auth := range doc.Authentications {
		if auth.KeyID == id {
			for _, pubKey := range doc.PubKeys {
				if pubKey.ID == id {
					return pubKey, true
				}
			}
			return PubKey{}, false
		}
	}
	return PubKey{}, false
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

type KeyID string

func NewKeyID(did DID, name string) KeyID {
	// https://www.w3.org/TR/did-core/#fragment
	return KeyID(fmt.Sprintf("%v#%s", did, name))
}

func ParseKeyID(id string, did DID) (KeyID, error) {
	keyID := KeyID(id)
	if !keyID.Valid(did) {
		return "", ErrInvalidKeyID(id)
	}
	return keyID, nil
}

func (id KeyID) Valid(did DID) bool {
	pattern := fmt.Sprintf(`^%v#\S+$`, did)
	matched, _ := regexp.MatchString(pattern, string(id))
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

type PubKey struct {
	ID        KeyID   `json:"id"`
	Type      KeyType `json:"type"`
	KeyBase58 string  `json:"publicKeyBase58"`
}

func NewPubKey(id KeyID, keyType KeyType, key crypto.PubKey) PubKey {
	return PubKey{
		ID:        id,
		Type:      keyType,
		KeyBase58: newPubKeyBase58(key, keyType, 0),
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

func (pk PubKey) Valid(did DID) bool {
	if !pk.ID.Valid(did) || !pk.Type.Valid() {
		return false
	}

	pattern := fmt.Sprintf("^[%s]+$", Base58Charset)
	matched, _ := regexp.MatchString(pattern, pk.KeyBase58)
	return matched
}

type Authentication struct {
	KeyID KeyID
	// DedicatedPubKey is not nil if it is only authorized for authentication
	// https://www.w3.org/TR/did-core/#example-18-authentication-property-containing-three-verification-methods
	DedicatedPubKey *PubKey
}

func newAuthentication(keyID KeyID) Authentication {
	return Authentication{KeyID: keyID, DedicatedPubKey: nil}
}

func newAuthenticationDedicated(pubKey PubKey) Authentication {
	return Authentication{KeyID: pubKey.ID, DedicatedPubKey: &pubKey}
}

func (a Authentication) hasDedicatedPubKey() bool {
	return a.DedicatedPubKey != nil
}

func (a Authentication) Valid(did DID) bool {
	if !a.KeyID.Valid(did) {
		return false
	}
	if a.DedicatedPubKey != nil {
		if !a.DedicatedPubKey.Valid(did) || a.DedicatedPubKey.ID != a.KeyID {
			return false
		}
	}
	return true
}

func (a Authentication) MarshalJSON() ([]byte, error) {
	// if dedicated
	if a.DedicatedPubKey != nil {
		return json.Marshal(a.DedicatedPubKey)
	}
	// if not dedicated
	return json.Marshal(a.KeyID)
}

func (a *Authentication) UnmarshalJSON(bz []byte) error {
	// if not dedicated
	var keyID KeyID
	err := json.Unmarshal(bz, &keyID)
	if err == nil {
		*a = newAuthentication(keyID)
		return nil
	}

	// if dedicated
	var pubKey PubKey
	if err := json.Unmarshal(bz, &pubKey); err != nil {
		return err
	}
	*a = newAuthenticationDedicated(pubKey)
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
