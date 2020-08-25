package keystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pborman/uuid"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/sha3"
)

const (
	version         = 3
	kdf             = "scrypt"
	scryptN         = 1 << 18
	scryptP         = 1
	scryptR         = 8
	scryptDKLen     = 32
	saltBytes       = 32
	derivedKeyLen   = 32
	cipherAlgorithm = "aes-128-ctr"
)

// KeyStore stores an encrypted private key on disk.
// It implements the Web3 Secret Storage Definition: https://github.com/ethereum/wiki/wiki/Web3-Secret-Storage-Definition.
type KeyStore struct {
	mtx     sync.RWMutex
	baseDir string
}

// NewKeyStore creates a KeyStore using baseDir. If baseDir doesn't exists, it is created automatically.
func NewKeyStore(baseDir string) (*KeyStore, error) {
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return nil, err
	}

	return &KeyStore{
		baseDir: baseDir,
	}, nil
}

// Save stores a key by encrypting it using passwd.
// The address is the name of the key which can be anything such as a blockchain address or a DID key ID.
// The address is used for generating a file name of the stored key.
func (ks *KeyStore) Save(address string, key []byte, passwd string) (string, error) {
	encryptedKey, err := encryptKey(address, key, passwd)
	if err != nil {
		return "", fmt.Errorf("fail to encrypt the key: %v", err)
	}
	return ks.save(address, encryptedKey)
}

func (ks *KeyStore) save(address string, key encryptedKey) (string, error) {
	ks.mtx.Lock()
	defer ks.mtx.Unlock()

	path := ks.newPath(address)
	if fileExists(path) {
		return "", fmt.Errorf("file is already exists: %s", path)
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(key); err != nil {
		return "", fmt.Errorf("fail to encode encryptedKey: %w", err)
	}

	return path, nil
}

// Load loads a key from path by decrypting it using passwd.
func (ks *KeyStore) Load(path string, passwd string) ([]byte, error) {
	encryptedKey, err := ks.load(path)
	if err != nil {
		return nil, err
	}
	return decryptKey(encryptedKey, passwd)
}

// LoadByAddress loads a key by decrypting it using passwd.
// If there are multiple files related to the address, it uses the recent one.
func (ks *KeyStore) LoadByAddress(address string, passwd string) ([]byte, error) {
	ks.mtx.RLock()
	defer ks.mtx.RUnlock()

	path, err := ks.recentPath(address)
	if err != nil {
		return nil, err
	}

	return ks.Load(path, passwd)
}

func (ks *KeyStore) load(path string) (encryptedKey, error) {
	var key encryptedKey

	ks.mtx.RLock()
	defer ks.mtx.RUnlock()

	file, err := os.Open(path)
	if err != nil {
		return key, err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&key); err != nil {
		return key, fmt.Errorf("fail to decode encryptedKey: %w", err)
	}
	return key, nil
}

func (ks *KeyStore) newPath(address string) string {
	return filepath.Join(
		ks.baseDir,
		fmt.Sprintf(
			"UTC--%s--%s.json",
			time.Now().UTC().Format("2006-01-02T15-04-05.000000000Z"),
			address,
		),
	)
}

func (ks *KeyStore) recentPath(address string) (string, error) {
	matches, err := filepath.Glob(fmt.Sprintf("%s/UTC--*--%s.json", ks.baseDir, address))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("file not found for address: %s", address)
	}

	recentPath := ""
	for _, match := range matches {
		if recentPath < match {
			recentPath = match
		}
	}
	return recentPath, nil
}

// encryptKey encrypts a private key into a JSON using the Scrypt KDF .
func encryptKey(address string, key []byte, passwd string) (encryptedKey, error) {
	// generate a random salt
	salt := make([]byte, saltBytes)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return encryptedKey{}, fmt.Errorf("fail to get random for salt: %w", err)
	}

	// derivedKey from the Scrypt KDF
	derivedKey, err := scrypt.Key([]byte(passwd), salt, scryptN, scryptR, scryptP, scryptDKLen)
	if err != nil {
		return encryptedKey{}, err
	}

	// 128-bit initialisation vector for the cipher
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return encryptedKey{}, fmt.Errorf("fail to get random for iv: %w", err)
	}

	// encrypt the key using AES-128-CTR
	cipherText, err := aesCTRXOR(derivedKey[:derivedKeyLen/2], iv, key[:])
	if err != nil {
		return encryptedKey{}, err
	}

	// MAC to check whether the decryption password was correct or not
	mac, err := newSHA3Keccak256(derivedKey[derivedKeyLen/2:derivedKeyLen], cipherText)
	if err != nil {
		return encryptedKey{}, err
	}

	// return a struct which can be marshalled as JSON
	return encryptedKey{
		Version: version,
		ID:      uuid.NewRandom().String(),
		Address: address,
		Crypto: cryptoParams{
			Cipher:     cipherAlgorithm,
			CipherText: hex.EncodeToString(cipherText),
			CipherParams: cipherParams{
				IV: hex.EncodeToString(iv),
			},
			KDF: kdf,
			KDFParams: kdfParams{
				N:     scryptN,
				R:     scryptR,
				P:     scryptP,
				DKLen: scryptDKLen,
				Salt:  hex.EncodeToString(salt),
			},
			MAC: hex.EncodeToString(mac),
		},
	}, nil
}

func decryptKey(key encryptedKey, passwd string) ([]byte, error) {
	// validate params
	if key.Version != version {
		return nil, fmt.Errorf("unsupported encryption version: %d", key.Version)
	}
	if key.Crypto.Cipher != cipherAlgorithm {
		return nil, fmt.Errorf("unsupported cipher algorithm: %s", key.Crypto.Cipher)
	}
	if key.Crypto.KDF != kdf {
		return nil, fmt.Errorf("unsupported kdf: %s", key.Crypto.Cipher)
	}

	mac, err := hex.DecodeString(key.Crypto.MAC)
	if err != nil {
		return nil, fmt.Errorf("fail to decode mac: %w", err)
	}

	iv, err := hex.DecodeString(key.Crypto.CipherParams.IV)
	if err != nil {
		return nil, fmt.Errorf("fail to decode iv: %w", err)
	}

	cipherText, err := hex.DecodeString(key.Crypto.CipherText)
	if err != nil {
		return nil, fmt.Errorf("fail to decode cipherText: %w", err)
	}

	salt, err := hex.DecodeString(key.Crypto.KDFParams.Salt)
	if err != nil {
		return nil, fmt.Errorf("fail to decode salt: %w", err)
	}

	derivedKey, err := scrypt.Key([]byte(passwd), salt, key.Crypto.KDFParams.N, key.Crypto.KDFParams.R, key.Crypto.KDFParams.P, key.Crypto.KDFParams.DKLen)
	if err != nil {
		return nil, fmt.Errorf("fail to derive a key: %w", err)
	}

	expectedMac, err := newSHA3Keccak256(derivedKey[derivedKeyLen/2:derivedKeyLen], cipherText)
	if err != nil {
		return nil, fmt.Errorf("fail to get an expected MAC: %w", err)
	}
	if !bytes.Equal(expectedMac, mac) {
		return nil, fmt.Errorf("mac verification was failed. the password might be wrong.")
	}

	return aesCTRXOR(derivedKey[:derivedKeyLen/2], iv, cipherText)
}

type encryptedKey struct {
	Version int          `json:"version"`
	ID      string       `json:"id"`
	Address string       `json:"address"`
	Crypto  cryptoParams `json:"crypto"`
}

type cryptoParams struct {
	Cipher       string       `json:"cipher"`
	CipherText   string       `json:"ciphertext"`
	CipherParams cipherParams `json:"cipherparams"`
	KDF          string       `json:"kdf"`
	KDFParams    kdfParams    `json:"kdfparams"`
	MAC          string       `json:"mac"`
}

type cipherParams struct {
	IV string `json:"iv"`
}

type kdfParams struct {
	N     int    `json:"n"`
	R     int    `json:"r"`
	P     int    `json:"p"`
	DKLen int    `json:"dklen"`
	Salt  string `json:"salt"`
}

// aesCTRXOR encrypts or decrypts a data using key and iv (bi-directional function).
func aesCTRXOR(key, iv, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("fail to get a new cipher: %w", err)
	}

	buf := make([]byte, len(data))
	cipher.NewCTR(block, iv).XORKeyStream(buf, data)
	return buf, nil
}

// newSHA3Keccak256 calculates a SHA3-256 (Keccak256).
func newSHA3Keccak256(data ...[]byte) ([]byte, error) {
	hash := sha3.New256()
	for _, b := range data {
		if _, err := hash.Write(b); err != nil {
			return nil, err
		}
	}
	return hash.Sum(nil), nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
