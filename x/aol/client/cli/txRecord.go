package cli

import (
	"bufio"
	"fmt"
	"os"

	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func CmdAddRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-record [ownerAddress] [topicName] [key] [value]",
		Short: "Add a new record",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ownerAddress := args[0]
			topicName := args[1]
			key := []byte(args[2])
			value := []byte(args[3])
			writerAddress := clientCtx.GetFromAddress().String()
			feePayerAddress := viper.GetString(flagFeePayer)

			msg := types.NewMsgAddRecord(topicName, key, value, writerAddress, ownerAddress, feePayerAddress)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return generateOrBroadcastTxWithMultiSigners(clientCtx, cmd.Flags(), msg.GetSigners(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(flagFeePayer, "", "optional address to pay for the fee")

	return cmd
}

// generateOrBroadcastTxWithMultiSigners is a fork of tx.GenerateOrBroadcastTxCLI, but with multi signers.
func generateOrBroadcastTxWithMultiSigners(clientCtx client.Context, flagSet *pflag.FlagSet, signerAddrs []sdk.AccAddress, msgs ...sdk.Msg) error {
	txf := tx.NewFactoryCLI(clientCtx, flagSet)
	if txf.SignMode() == signing.SignMode_SIGN_MODE_UNSPECIFIED {
		txf = txf.WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)
	}

	if clientCtx.GenerateOnly {
		return tx.GenerateTx(clientCtx, txf, msgs...)
	}
	return broadcastTxWithMultiSigners(clientCtx, txf, signerAddrs, msgs...)
}

// broadcastTxWithMultiSigners is a fork of tx.BroadcastTx, but with multi signers.
func broadcastTxWithMultiSigners(clientCtx client.Context, txf tx.Factory, signerAddrs []sdk.AccAddress, msgs ...sdk.Msg) error {
	if txf.SimulateAndExecute() || clientCtx.Simulate {
		_, adjusted, err := tx.CalculateGas(clientCtx.QueryWithData, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", tx.GasEstimateResponse{GasEstimate: txf.Gas()})
	}

	if clientCtx.Simulate {
		return nil
	}

	txBuilder, err := tx.BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return err
	}

	if !clientCtx.SkipConfirm {
		out, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(os.Stderr, "%s\n\n", out)

		buf := bufio.NewReader(os.Stdin)
		ok, err := input.GetConfirmation("confirm transaction before signing and broadcasting", buf, os.Stderr)

		if err != nil || !ok {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", "cancelled transaction")
			return err
		}
	}

	accounts, err := retrieveAccounts(clientCtx, txf, signerAddrs)
	if err != nil {
		return err
	}

	if err := gatherAllSignerInfos(txBuilder, txf.SignMode(), accounts); err != nil {
		return err
	}

	if err := signWithMultiSigners(clientCtx, txBuilder, txf.SignMode(), accounts); err != nil {
		return err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}

	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}

type accountInKeyring struct {
	keyInfo       keyring.Info
	accountNumber uint64
	sequence      uint64
}

func retrieveAccounts(clientCtx client.Context, txf tx.Factory, addrs []sdk.AccAddress) ([]accountInKeyring, error) {
	var accounts []accountInKeyring

	for _, addr := range addrs {
		keyInfo, err := clientCtx.Keyring.KeyByAddress(addr)
		if err != nil {
			return nil, err
		}

		// If offline, use accNum and accSeq which were specified in flags
		accNum := txf.AccountNumber()
		accSeq := txf.Sequence()
		if !clientCtx.Offline {
			accNum, accSeq, err = clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, addr)
			if err != nil {
				return nil, err
			}
		}

		accounts = append(accounts, accountInKeyring{
			keyInfo:       keyInfo,
			accountNumber: accNum,
			sequence:      accSeq,
		})
	}

	return accounts, nil
}

func gatherAllSignerInfos(txBuilder client.TxBuilder, signMode signing.SignMode, accounts []accountInKeyring) error {
	var sigsV2 []signing.SignatureV2

	for _, account := range accounts {
		sigV2 := signing.SignatureV2{
			PubKey: account.keyInfo.GetPubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: nil,
			},
			Sequence: account.sequence,
		}
		sigsV2 = append(sigsV2, sigV2)
	}

	return txBuilder.SetSignatures(sigsV2...)
}

func signWithMultiSigners(clientCtx client.Context, txBuilder client.TxBuilder, signMode signing.SignMode, accounts []accountInKeyring) error {
	var sigsV2 []signing.SignatureV2

	for _, account := range accounts {
		signerData := xauthsigning.SignerData{
			ChainID:       clientCtx.ChainID,
			AccountNumber: account.accountNumber,
			Sequence:      account.sequence,
		}
		signBytes, err := clientCtx.TxConfig.SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
		if err != nil {
			return err
		}
		signature, pubKey, err := clientCtx.Keyring.Sign(account.keyInfo.GetName(), signBytes)
		if err != nil {
			return err
		}
		sigV2 := signing.SignatureV2{
			PubKey: pubKey,
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: signature,
			},
			Sequence: account.sequence,
		}

		sigsV2 = append(sigsV2, sigV2)
	}

	return txBuilder.SetSignatures(sigsV2...)
}
