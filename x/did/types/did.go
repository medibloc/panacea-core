package types

import (
	"fmt"
	"regexp"

	"github.com/btcsuite/btcutil/base58"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type DID string

const (
	DIDMethod = "panacea"
)

func NewDID(pubKey crypto.PubKey) DID {
	networkID := "testnet"    // TODO: get this from somewhere
	idStr := pubKey.Address() // TODO: correct?
	return DID(fmt.Sprintf("did:%s:%s:%s", DIDMethod, networkID, idStr))
}

func (did DID) IsValid() bool {
	// TODO: check if the pattern is correct, according to the encoding of id-string
	match, _ := regexp.MatchString("^did:panacea:(mainnet|testnet):[1-9A-Za-z]{21,22}$", string(did))
	return match
}

type DIDDocument struct {
	ID              DID              `json:"id"`
	PubKeys         []PubKey         `json:"public_key"`
	Authentications []Authentication `json:"authentication"`
}

func NewDIDDocument(id DID, pubKey PubKey) DIDDocument {
	return DIDDocument{
		ID:              id,
		PubKeys:         []PubKey{pubKey},
		Authentications: []Authentication{NewAuthentication(pubKey)},
	}
}

type PubKey struct {
	ID         string
	Type       PubKeyType
	KeyEncoded string
	Encoding   PubKeyEncoding
}

func MustNewPubKey(id string, key crypto.PubKey, keyType PubKeyType) PubKey {
	if id == "" {
		panic("id shouldn't be empty")
	}
	if key == nil {
		panic("key shouldn't be empty")
	}
	if !keyType.IsValid() {
		panic("keyType is invalid")
	}

	keyEncoded, encoding := mustEncodePubKey(key, keyType)
	return PubKey{
		ID:         id,
		Type:       keyType,
		KeyEncoded: keyEncoded,
		Encoding:   encoding,
	}
}

func mustEncodePubKey(key crypto.PubKey, keyType PubKeyType) (string, PubKeyEncoding) {
	switch keyType {
	case ES256K:
		return encodePubKeyES256K(key)
	}
	panic(fmt.Sprintf("unsupported pubkey type: %v", keyType))
}

func encodePubKeyES256K(key crypto.PubKey) (string, PubKeyEncoding) {
	k := key.(secp256k1.PubKeySecp256k1)
	return base58.Encode(k[:]), BASE58
}

type PubKeyType string

const (
	ES256K PubKeyType = "Secp256k1VerificationKey"
)

func (t PubKeyType) IsValid() bool {
	switch t {
	case ES256K:
		return true
	}
	return false
}

type PubKeyEncoding string

const (
	BASE58 PubKeyEncoding = "publicKeyBase58"
)

func (t PubKeyEncoding) IsValid() bool {
	switch t {
	case BASE58:
		return true
	}
	return false
}

type Authentication struct {
	PubKeyID string
}

func NewAuthentication(pubKey PubKey) Authentication {
	return Authentication{
		PubKeyID: pubKey.ID,
	}
}
