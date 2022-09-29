package crypto

import (
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	oraclekeeper "github.com/medibloc/panacea-core/v2/x/oracle/keeper"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	FlagCipherText     = "cipher-text"
	FlagCipherTextPath = "path"
)

func Command(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt-data [file-path] [key-name] [dealID]",
		Short: "Decrypt and output the data file encrypted with shared key which consists of oracle public key and buyer private key.",
		Long: `This command can decrypt encrypted data with shared key which consists of oracle public key and buyer private key.
				And your key should be stored in the localStore.
				(According to the following command (panacead keys add ...)`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			//TODO: args[1] will be used for getting a deal's nonce
			decryptedData, err := Decrypt(clientCtx, cmd, args[0], args[1])
			if err != nil {
				return err
			}

			cmd.Println(string(decryptedData))

			return nil
		},
	}

	cmd.PersistentFlags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.PersistentFlags().String(flags.FlagKeyringDir, "", "The client Keyring directory; if omitted, the default 'home' directory will be used")
	cmd.PersistentFlags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.PersistentFlags().String(cli.OutputFlag, "text", "Output format (text|json)")
	cmd.Flags().String(FlagCipherText, "", "Cipher text in the form of a byte array")
	cmd.Flags().String(FlagCipherTextPath, "", "Path to the file where the cipher text in the form of a byte array is stored")

	return cmd
}

func Decrypt(clientCtx client.Context, cmd *cobra.Command, filePath, keyName string) ([]byte, error) {
	privKeyHex, err := keyring.NewUnsafe(clientCtx.Keyring).UnsafeExportPrivKeyHex(keyName)
	if err != nil {
		return nil, err
	}

	privKeyBz, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, err
	}

	buyerPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz)

	ctx := sdk.UnwrapSDKContext(cmd.Context())
	params := oraclekeeper.Keeper{}.GetParams(ctx)

	pubKeybz, err := base64.StdEncoding.DecodeString(params.OraclePublicKey)
	if err != nil {
		return nil, err
	}

	oraclePubkey, err := btcec.ParsePubKey(pubKeybz, btcec.S256())
	if err != nil {
		return nil, err
	}

	sharedKey := crypto.DeriveSharedKey(buyerPrivKey, oraclePubkey, crypto.KDFSHA256)

	encryptedData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	//TODO: Nonce will be implemented after CreateDeal merged
	decryptedData, err := crypto.DecryptWithAES256(sharedKey, nil, encryptedData)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
