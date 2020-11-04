package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/medibloc/panacea-core/x/currency/types"
	"github.com/spf13/cobra"
)

const (
	flagInteractive = "interactive"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	currencyTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Currency transaction subcommands",
	}

	currencyTxCmd.AddCommand(client.PostCommands(
		GetCmdIssueCurrency(cdc),
	)...)

	return currencyTxCmd
}

func GetCmdIssueCurrency(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue [amount]",
		Short: "Issue a new currency and deposit coins the amount to the issuer's account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			issuerAddr := cliCtx.GetFromAddress()
			amount, err := sdk.ParseCoin(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgIssueCurrency(amount, issuerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	return cmd
}
