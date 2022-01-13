package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/market/types"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"io/ioutil"
)

func NewCreateDealCmd() *cobra.Command {
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
		deal.TargetNumData,
		deal.TrustedDataValidators,
		clientCtx.GetFromAddress().String(),
	)

	return txf, msg, nil
}

func parseCreateDealFlags(fs *flag.FlagSet) (*createDealInputs, error) {
	deal := &createDealInputs{}
	dealFile, _ := fs.GetString(FlagDealFile)

	if dealFile == "" {
		return nil, fmt.Errorf("need deal json file using --%s flag", FlagDealFile)
	}

	contents, err := ioutil.ReadFile(dealFile)
	if err != nil {
		return nil, err
	}

	err = deal.UnmarshalJSON(contents)
	if err != nil {
		return nil, err
	}

	return deal, nil
}
