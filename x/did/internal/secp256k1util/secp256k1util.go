package secp256k1util

import (
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// PrivKeyFromBytes converts a byte slice into a Secp256k1 private key.
// It returns an error when the length of the input is invalid.
func PrivKeyFromBytes(bz []byte) (secp256k1.PrivKeySecp256k1, error) {
	var key secp256k1.PrivKeySecp256k1
	if len(bz) != len(key) {
		return key, fmt.Errorf("invalid Secp256k1 private key. len:%d, expected:%d", len(bz), len(key))
	}
	copy(key[:], bz)
	return key, nil
}

// PubKeyFromBase58 decodes a base58-encoded Secp256k1 public key.
// It returns an error when the length of the input is invalid.
func PubKeyFromBase58(b58 string) (secp256k1.PubKeySecp256k1, error) {
	var key secp256k1.PubKeySecp256k1
	decoded := base58.Decode(b58)
	if len(decoded) != len(key) {
		return key, fmt.Errorf("invalid Secp256k1 public key. len:%d, expected:%d", len(decoded), len(key))
	}
	copy(key[:], decoded)
	return key, nil
}

func DerivePubKey(privKey secp256k1.PrivKeySecp256k1) secp256k1.PubKeySecp256k1 {
	return privKey.PubKey().(secp256k1.PubKeySecp256k1)
}

func PubKeyBytes(pubKey secp256k1.PubKeySecp256k1) []byte {
	// Do not use pubKey.Bytes() which does Amino encoding
	return pubKey[:]
}
