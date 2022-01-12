package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/market/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(NewCreateDealCmd())
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
		deal.WantDataCount,
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
