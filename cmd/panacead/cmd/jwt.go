package cmd

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
)

func JwtCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-jwt [key-name] [expiration]",
		Short: "Issue a JWT(Json Web Token) from account",
		Args: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			fromAddress := clientCtx.GetFromAddress()

			expirationDuration, err := time.ParseDuration(args[1])
			if err != nil {
				return err
			}

			issuedJWT, err := issueJWT(clientCtx, fromAddress.String(), args[0], expirationDuration)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(issuedJWT))
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

	return cmd
}

func issueJWT(clientCtx client.Context, issuer, keyName string, expiration time.Duration) ([]byte, error) {
	privKeyHex, err := keyring.NewUnsafe(clientCtx.Keyring).UnsafeExportPrivKeyHex(keyName)
	if err != nil {
		return nil, err
	}

	privKeyBz, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, err
	}

	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKeyBz)

	now := time.Now().Truncate(time.Second)
	token, err := jwt.NewBuilder().
		Issuer(issuer).
		IssuedAt(now).
		NotBefore(now).
		Expiration(now.Add(expiration)).
		Build()
	if err != nil {
		return nil, err
	}

	signedJWT, err := jwt.Sign(token, jwt.WithKey(jwa.ES256K, privKey.ToECDSA()))
	if err != nil {
		return nil, err
	}

	return signedJWT, nil
}
