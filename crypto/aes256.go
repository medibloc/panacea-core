package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/btcec"
)

// DeriveSharedKey derives a shared key (which can be used for asymmetric encryption)
// using a specified KDF (Key Derivation Function)
// from a shared secret generated by Diffie-Hellman key exchange (ECDH).
func DeriveSharedKey(priv *btcec.PrivateKey, pub *btcec.PublicKey, kdf func([]byte) []byte) []byte {
	sharedSecret := btcec.GenerateSharedSecret(priv, pub)
	return kdf(sharedSecret)
}

// KDFSHA256 is a key derivation function which uses SHA256.
func KDFSHA256(in []byte) []byte {
	out := sha256.Sum256(in)
	return out[:]
}

// Encrypt combines secretKey and secondKey to encrypt with AES256-GCM method.
func Encrypt(secretKey, additional, data []byte) ([]byte, error) {
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("secret key is not for AES-256: total %d bits", 8*len(secretKey))
	}

	// prepare AES-256-GSM cipher
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// make random nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// encrypt data with second key
	ciphertext := aesGCM.Seal(nonce, nonce, data, additional)
	return ciphertext, nil
}

// Decrypt combines secretKey and secondKey to decrypt AES256-GCM.
func Decrypt(secretKey []byte, additional []byte, ciphertext []byte) ([]byte, error) {
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("secret key is not for AES-256: total %d bits", 8*len(secretKey))
	}

	// prepare AES-256-GCM cipher
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	nonce, pureCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// decrypt ciphertext with second key
	plaintext, err := aesgcm.Open(nil, nonce, pureCiphertext, additional)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}