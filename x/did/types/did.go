package types

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/btcsuite/btcutil/base58"
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

type PubKey struct {
	ID        PubKeyID   `json:"id"`
	Type      PubKeyType `json:"type"`
	KeyBase58 string     `json:"publicKeyBase58"`
}

type PubKeyID string
type PubKeyType string
type Authentication PubKeyID

const (
	ES256K PubKeyType = "Secp256k1VerificationKey2018"
)

func NewDIDDocument(id DID, pubKey PubKey) DIDDocument {
	return DIDDocument{
		ID:              id,
		PubKeys:         []PubKey{pubKey},
		Authentications: []Authentication{Authentication(pubKey.ID)},
	}
}

func (doc DIDDocument) IsEmpty() bool {
	return doc.ID.IsEmpty()
}

func (doc DIDDocument) String() string {
	bz, _ := json.Marshal(doc)
	return string(bz)
}

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

func (t PubKeyType) IsValid() bool {
	switch t {
	case ES256K:
		return true
	}
	return false
}
