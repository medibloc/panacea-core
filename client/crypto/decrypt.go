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
		Use:   "crypto-data [file-path] [dealID] [key-name]",
		Short: "Decrypt and output the data file encrypted with shared key which consists of oracle public key and buyer private key.",
		Long: `This command can decrypt encrypted data with shared key which consists of oracle public key and buyer private key.
				And your key should be stored in the localStore.
				(According to the following command (panacead keys add ...)`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			privKeyHex, err := keyring.NewUnsafe(clientCtx.Keyring).UnsafeExportPrivKeyHex(args[3])
			if err != nil {
				return err
			}

			privKeyBz, err := hex.DecodeString(privKeyHex)
			if err != nil {
				return err
			}

			buyerPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz)

			ctx := sdk.UnwrapSDKContext(cmd.Context())
			params := oraclekeeper.Keeper{}.GetParams(ctx)

			pubKeybz, err := base64.StdEncoding.DecodeString(params.OraclePublicKey)
			if err != nil {
				return err
			}

			oraclePubkey, err := btcec.ParsePubKey(pubKeybz, btcec.S256())
			if err != nil {
				return nil
			}

			sharedKey := crypto.DeriveSharedKey(buyerPrivKey, oraclePubkey, crypto.KDFSHA256)

			encryptedData, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}

			//TODO: Nonce will be implemented after CreateDeal merged
			decryptedData, err := crypto.DecryptWithAES256(sharedKey, nil, encryptedData)
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
