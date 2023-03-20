package dep

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/medibloc/panacea-core/v2/crypto"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

// DecryptDataCmd decrypts data encrypted by the shared key which is generated by oracle private key and buyer's public key
func DecryptDataCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt-data [input-file-path] [key-name] [encrypted-secret-key]",
		Short: "Decrypt data with encrypted secret key",
		Long: `
This command decrypts the encrypted data with the encrypted secret key.
The encrypted combinedKey can be obtained from Oracle.
The key to be used for encryption should be stored in the localStore.
If not stored, please add the key first via the following command.
panacead keys add ...
		`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			oracleQueryClient := oracletypes.NewQueryClient(clientCtx)
			oracleResp, err := oracleQueryClient.Params(cmd.Context(), &oracletypes.QueryOracleParamsRequest{})
			if err != nil {
				return err
			}
			oraclePubKey := oracleResp.GetParams().MustDecodeOraclePublicKey()

			// TODO: I think we need to change it to get Secret from Oracle node.(Maybe...)
			encryptedSecretKey, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return err
			}

			encryptedData, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			decryptedData, err := decrypt(clientCtx, args[1], encryptedSecretKey, oraclePubKey, encryptedData)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(decryptedData))
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.PersistentFlags().String(flags.FlagKeyringDir, "", "The client Keyring directory; if omitted, the default 'home' directory will be used")
	cmd.PersistentFlags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.PersistentFlags().String(cli.OutputFlag, "text", "Output format (text|json)")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func decrypt(clientCtx client.Context, keyName string, encryptedCombinedKey, oraclePubKeyBz, encryptedData []byte) ([]byte, error) {
	privKeyHex, err := keyring.NewUnsafe(clientCtx.Keyring).UnsafeExportPrivKeyHex(keyName)
	if err != nil {
		return nil, err
	}

	privKeyBz, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, err
	}

	consumerPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz)

	oraclePubkey, err := btcec.ParsePubKey(oraclePubKeyBz, btcec.S256())
	if err != nil {
		return nil, err
	}

	sharedKey := crypto.DeriveSharedKey(consumerPrivKey, oraclePubkey, crypto.KDFSHA256)

	combinedKey, err := crypto.Decrypt(sharedKey, nil, encryptedCombinedKey)
	if err != nil {
		return nil, err
	}

	return crypto.Decrypt(combinedKey, nil, encryptedData)
}
