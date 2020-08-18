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

type DID string

const (
	DIDMethod     = "panacea"
	Base58Charset = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
)

func NewDID(pubKey crypto.PubKey, keyType PubKeyType) DID {
	networkID := "testnet" // TODO: get this from somewhere
	idStr := mustPubKeyBase58(pubKey, keyType, 16)
	return DID(fmt.Sprintf("did:%s:%s:%s", DIDMethod, networkID, idStr))
}

func (did DID) IsValid() bool {
	pattern := fmt.Sprintf("^did:panacea:(mainnet|testnet):[%s]{21,22}$", Base58Charset)
	match, _ := regexp.MatchString(pattern, string(did))
	return match
}

func (did DID) IsEmpty() bool {
	return did == ""
}

type DIDDocument struct {
	ID              DID              `json:"id"`
	PubKeys         []PubKey         `json:"publicKey"`
	Authentications []Authentication `json:"authentication"`
}

func NewDIDDocument(id DID, pubKey PubKey) DIDDocument {
	return DIDDocument{
		ID:              id,
		PubKeys:         []PubKey{pubKey},
		Authentications: []Authentication{Authentication(pubKey.ID)},
	}
}

func (doc DIDDocument) IsValid() bool {
	if !doc.ID.IsValid() || doc.PubKeys == nil || doc.Authentications == nil {
		return false
	}

	for _, pubKey := range doc.PubKeys {
		if !pubKey.IsValid() {
			return false
		}
	}

	for _, auth := range doc.Authentications {
		if !auth.IsValid() {
			return false
		}
		if _, ok := doc.PubKeyByID(PubKeyID(auth)); !ok {
			return false
		}
	}

	return true
}

func (doc DIDDocument) IsEmpty() bool {
	return doc.ID.IsEmpty()
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
func (doc DIDDocument) PubKeyByID(id PubKeyID) (*PubKey, bool) {
	for i := 0; i < len(doc.PubKeys); i++ {
		pubKey := &doc.PubKeys[i]
		if pubKey.ID == id {
			return pubKey, true
		}
	}
	return nil, false
}

type PubKey struct {
	ID        PubKeyID   `json:"id"`
	Type      PubKeyType `json:"type"`
	KeyBase58 string     `json:"publicKeyBase58"`
}

type PubKeyID string

func MustNewPubKey(id PubKeyID, key crypto.PubKey, keyType PubKeyType) PubKey {
	if id == "" {
		panic("id shouldn't be empty")
	}
	if key == nil {
		panic("key shouldn't be empty")
	}
	if !keyType.IsValid() {
		panic("keyType is invalid")
	}

	return PubKey{
		ID:        id,
		Type:      keyType,
		KeyBase58: mustPubKeyBase58(key, keyType, 0),
	}
}

func mustPubKeyBase58(key crypto.PubKey, keyType PubKeyType, truncateLen int) string {
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

func (pk PubKey) IsValid() bool {
	if pk.ID == "" || pk.Type.IsValid() {
		return false
	}

	pattern := fmt.Sprintf("^[%s]$", Base58Charset)
	matched, _ := regexp.MatchString(pattern, pk.KeyBase58)
	return matched
}

type PubKeyType string

const (
	ES256K PubKeyType = "Secp256k1VerificationKey2018"
)

func (t PubKeyType) IsValid() bool {
	switch t {
	case ES256K:
		return true
	}
	return false
}

// TODO: to be extended
type Authentication PubKeyID

func (a Authentication) IsValid() bool {
	return a != ""
}

// NewPrivKeyFromBase58 decodes a base58-encoded Secp256k1 private key.
// It returns an error when the length of the input is invalid.
func NewPrivKeyFromBase58(b58 string) (secp256k1.PrivKeySecp256k1, error) {
	var key secp256k1.PrivKeySecp256k1
	decoded := base58.Decode(b58)
	if len(decoded) != len(key) {
		return key, fmt.Errorf("invalid Secp256k1 private key. len:%d, expected:%d", len(decoded), len(key))
	}
	copy(key[:], decoded)
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
