package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/medibloc/panacea-core/v2/x/market/types"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"io/ioutil"
	"strconv"
)

func CmdCreateDeal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-deal [flags]",
		Short: "create a new deal",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewBuildCreateDealMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String(FlagDealFile, "", "Deal json file path")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdSellData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell-data [flags]",
		Short: "sell data",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return nil
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			txf, msg, err := NewSellDataMsg(clientCtx, txf, cmd.Flags())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String(DataVerificationCertificateFile, "", "Data Verification Certificate file path")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdDeactivateDeal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-deal [dealId]",
		Short: "deactivate-deal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dealId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			requesterAddress := clientCtx.GetFromAddress()
			msg := types.NewMsgDeactivateDeal(dealId, requesterAddress.String())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewBuildCreateDealMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	deal, err := parseCreateDealFlags(fs)
	if err != nil {
		return txf, nil, fmt.Errorf("faild to parse deal: %w", err)
	}

	budget, err := sdk.ParseCoinNormalized(deal.Budget)
	if err != nil {
		return txf, nil, err
	}

	msg := types.NewMsgCreateDeal(
		deal.DataSchema,
		&budget,
		deal.MaxNumData,
		deal.TrustedDataValidators,
		clientCtx.GetFromAddress().String(),
	)

	return txf, msg, nil
}

func parseCreateDealFlags(fs *flag.FlagSet) (*createDealInputs, error) {
	var createDeal createDealInputs
	dealFile, _ := fs.GetString(FlagDealFile)

	if dealFile == "" {
		return nil, fmt.Errorf("need deal json file using --%s flag", FlagDealFile)
	}

	contents, err := ioutil.ReadFile(dealFile)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(contents))

	if err := dec.Decode(&createDeal); err != nil {
		return nil, err
	}

	return &createDeal, nil
}

func NewSellDataMsg(clientCtx client.Context, txf tx.Factory, fs *flag.FlagSet) (tx.Factory, sdk.Msg, error) {
	sellData, err := parseSellDataFlags(fs)
	if err != nil {
		return txf, nil, fmt.Errorf("failed to parse data certificate file: %w", err)
	}

	msg := types.NewMsgSellData(sellData, clientCtx.GetFromAddress().String())

	return txf, msg, nil
}

func parseSellDataFlags(fs *flag.FlagSet) (types.DataValidationCertificate, error) {
	sellData := types.DataValidationCertificate{}
	receiptFile, _ := fs.GetString(DataVerificationCertificateFile)

	if receiptFile == "" {
		return types.DataValidationCertificate{}, fmt.Errorf("need receipt json file using --%s flag", DataVerificationCertificateFile)
	}

	contents, err := ioutil.ReadFile(receiptFile)
	if err != nil {
		return types.DataValidationCertificate{}, err
	}

	err = jsonpb.Unmarshal(bytes.NewReader(contents), &sellData)
	if err != nil {
		return types.DataValidationCertificate{}, err
	}

	return sellData, nil
}
