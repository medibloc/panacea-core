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

func NewDID(networkID NetworkID, pubKey crypto.PubKey, keyType PubKeyType) DID {
	idStr := newPubKeyBase58(pubKey, keyType, 16)
	return DID(fmt.Sprintf("did:%s:%s:%s", DIDMethod, networkID, idStr))
}

func (did DID) Valid() bool {
	pattern := fmt.Sprintf("^did:panacea:%s:[%s]{21,22}$", networkIDRegex(), Base58Charset)
	matched, _ := regexp.MatchString(pattern, string(did))
	return matched
}

func (did DID) Empty() bool {
	return did == ""
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

func (doc DIDDocument) Valid() bool {
	if !doc.ID.Valid() || doc.PubKeys == nil || doc.Authentications == nil {
		return false
	}

	for _, pubKey := range doc.PubKeys {
		if !pubKey.Valid() {
			return false
		}
	}

	for _, auth := range doc.Authentications {
		if !auth.Valid() {
			return false
		}
		if _, ok := doc.PubKeyByID(PubKeyID(auth)); !ok {
			return false
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

func NewPubKey(id PubKeyID, key crypto.PubKey, keyType PubKeyType) PubKey {
	return PubKey{
		ID:        id,
		Type:      keyType,
		KeyBase58: newPubKeyBase58(key, keyType, 0),
	}
}

func newPubKeyBase58(key crypto.PubKey, keyType PubKeyType, truncateLen int) string {
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

func (pk PubKey) Valid() bool {
	if pk.ID == "" || !pk.Type.Valid() {
		return false
	}

	pattern := fmt.Sprintf("^[%s]+$", Base58Charset)
	matched, _ := regexp.MatchString(pattern, pk.KeyBase58)
	return matched
}

type PubKeyType string

const (
	ES256K PubKeyType = "Secp256k1VerificationKey2018"
)

func (t PubKeyType) Valid() bool {
	switch t {
	case ES256K:
		return true
	}
	return false
}

// TODO: to be extended
type Authentication PubKeyID

func (a Authentication) Valid() bool {
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
