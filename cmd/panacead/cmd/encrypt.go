package cmd

import (
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/medibloc/panacea-core/v2/crypto"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

const (
	FlagCipherText     = "cipher-text"
	FlagCipherTextPath = "path"
)

func EncryptData(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt-data [input-file-path] [output-file-path] [key-name] [dealID]",
		Short: "Encrypt the data file with shared key which consists of oracle public key and seller private key.",
		Long: `This command can encrypt data with shared key which consists of oracle public key and seller private key.
				And your key should be stored in the localStore.
				(According to the following command (panacead keys add ...)`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			dealQueryClient := datadealtypes.NewQueryClient(clientCtx)
			oracleQueryClient := oracletypes.NewQueryClient(clientCtx)

			dealID, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			originData, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}

			oracleResp, err := oracleQueryClient.Params(cmd.Context(), &oracletypes.QueryOracleParamsRequest{})
			if err != nil {
				return err
			}
			oraclePubKey := oracleResp.GetParams().GetOraclePublicKey()

			dealResp, err := dealQueryClient.Deal(cmd.Context(), &datadealtypes.QueryDealRequest{DealId: dealID})
			if err != nil {
				return err
			}
			nonce := dealResp.GetDeal().GetNonce()

			encryptedData, err := encrypt(clientCtx, args[2], originData, oraclePubKey, nonce)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(args[1], encryptedData, 0644)
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
	cmd.Flags().String(FlagCipherText, "", "Cipher text in the form of a byte array")
	cmd.Flags().String(FlagCipherTextPath, "", "Path to the file where the cipher text in the form of a byte array is stored")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func encrypt(clientCtx client.Context, keyName string, originData []byte, oraclePubKey string, nonce []byte) ([]byte, error) {
	privKeyHex, err := keyring.NewUnsafe(clientCtx.Keyring).UnsafeExportPrivKeyHex(keyName)
	if err != nil {
		return nil, err
	}

	privKeyBz, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, err
	}

	sellerPrivKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz)

	oraclePubKeyBz, err := base64.StdEncoding.DecodeString(oraclePubKey)
	if err != nil {
		return nil, err
	}

	oraclePubkey, err := btcec.ParsePubKey(oraclePubKeyBz, btcec.S256())
	if err != nil {
		return nil, err
	}

	sharedKey := crypto.DeriveSharedKey(sellerPrivKey, oraclePubkey, crypto.KDFSHA256)

	encryptedData, err := crypto.Encrypt(sharedKey, nonce, originData)
	if err != nil {
		return nil, err
	}

	return encryptedData, nil
}
