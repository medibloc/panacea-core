package decrypt

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tendermint/tendermint/libs/cli"
	"io/ioutil"
)

const (
	FlagCipherText     = "cipher-text"
	FlagCipherTextPath = "path"
)

func Command(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt [name] (--cipher-text [ciphertext] | --path [ciphertext-file])",
		Short: "Decrypt and output the ciphertext file encrypted with ECDSA PublicKey.",
		Long: `This command can decrypt ciphertext encrypted with ECDSA PublicKey. 
And your key should be stored in the localStore.
(According to the following command (panacead keys add ...)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			privKeyHex, err := keyring.NewUnsafe(clientCtx.Keyring).UnsafeExportPrivKeyHex(args[0])
			if err != nil {
				return err
			}

			privKey, err := hex.DecodeString(privKeyHex)
			if err != nil {
				return err
			}

			content, err := getContent(cmd.Flags())

			if err != nil {
				return fmt.Errorf("failed read cipherText. %w", err)
			}

			ecdsaPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKey)
			if err != nil {
				return err
			}
			plainText, err := btcec.Decrypt(ecdsaPrivKey, content)
			if err != nil {
				return err
			}

			cmd.Println(string(plainText))

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

func getContent(flag *pflag.FlagSet) ([]byte, error) {
	cipherText, err := flag.GetString(FlagCipherText)
	if err != nil {
		return nil, err
	}
	cipherTextPath, err := flag.GetString(FlagCipherTextPath)
	if err != nil {
		return nil, err
	}

	if len(cipherText) == 0 && len(cipherTextPath) == 0 {
		return nil, fmt.Errorf("either --text or --path should be required")
	} else if len(cipherText) != 0 && len(cipherTextPath) != 0 {
		return nil, fmt.Errorf("only one of --text or --path should be specified")
	}

	if len(cipherText) != 0 {
		return []byte(cipherText), nil
	}

	content, err := ioutil.ReadFile(cipherTextPath)
	if err != nil {
		return nil, err
	}
	return content, nil
}
