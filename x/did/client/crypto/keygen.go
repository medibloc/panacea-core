package crypto

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

const (
	mnemonicEntropySize = 256
	defaultAccountForHD = 0
	defaultIndexForHD   = 0
)

// GenSecp256k1PrivKey generates a Secp256k1 private key from mnemonic anad bip39Passphrase.
// If mnemonic is an empty string, this function generates a random mnemonic.
// The bip39Passphrase can be an empty string.
func GenSecp256k1PrivKey(mnemonic, bip39Passphrase string) (secp256k1.PrivKey, error) {
	if mnemonic == "" { // generate a random mnemonic
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			return secp256k1.PrivKey{}, err
		}
		mnemonic, err = bip39.NewMnemonic(entropySeed[:])
		if err != nil {
			return secp256k1.PrivKey{}, err
		}
		fmt.Fprintf(os.Stderr, "A random mnemonic was generated: %s\n", mnemonic)
	} else if !bip39.IsMnemonicValid(mnemonic) {
		return secp256k1.PrivKey{}, fmt.Errorf("invalid mnemonic: %s", mnemonic)
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return secp256k1.PrivKey{}, err
	}

	hdPath := hd.NewFundraiserParams(defaultAccountForHD, sdk.GetConfig().GetCoinType(), defaultIndexForHD).String()
	masterPriv, chainCode := hd.ComputeMastersFromSeed(seed)
	return hd.DerivePrivateKeyForPath(masterPriv, chainCode, hdPath)
}
