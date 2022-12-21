package cmd

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

func EncryptDataCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt-data [input-file-path] [key-name]",
		Short: "Encrypt data with shared key which consists of oracle public key and provider's private key",
		Long: `
			This command can encrypt data with shared key which consists of oracle public key and provider's private key.
			The key to be used for encryption should be stored in the localStore.
			If not stored, please add the key first via the following command.
			panacead keys add ...
		`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := oracletypes.NewQueryClient(clientCtx)

			origData, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			params, err := queryClient.Params(cmd.Context(), &oracletypes.QueryOracleParamsRequest{})
			if err != nil {
				return err
			}

			oraclePubKey := params.GetParams().GetOraclePublicKey()

			encryptedData, err := encrypt(clientCtx, args[1], origData, oraclePubKey)
			if err != nil {
				return err
			}

			encryptedDataBase64 := base64.StdEncoding.EncodeToString(encryptedData)

			//cmd.Println(encryptedDataBase64)

			_, err = fmt.Fprintln(cmd.OutOrStdout(), encryptedDataBase64)
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
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func encrypt(clientCtx client.Context, keyName string, origData []byte, oraclePubKeyStr string) ([]byte, error) {
	// get unsafe export private key from keystore
	privKeyHex, err := keyring.NewUnsafe(clientCtx.Keyring).UnsafeExportPrivKeyHex(keyName)
	if err != nil {
		return nil, err
	}

	privKeyBz, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, err
	}

	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz)

	// oracle public key
	oraclePubKeyBz, err := base64.StdEncoding.DecodeString(oraclePubKeyStr)
	if err != nil {
		return nil, err
	}

	oraclePubKey, err := btcec.ParsePubKey(oraclePubKeyBz, btcec.S256())
	if err != nil {
		return nil, err
	}

	// shared key
	sharedKey := crypto.DeriveSharedKey(privKey, oraclePubKey, crypto.KDFSHA256)

	// encrypt data
	encryptedData, err := crypto.Encrypt(sharedKey, nil, origData)
	if err != nil {
		return nil, err
	}

	return encryptedData, nil
}
