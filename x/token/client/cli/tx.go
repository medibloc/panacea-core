package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/medibloc/panacea-core/x/token/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	tokenTxCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Token transaction subcommands",
	}

	tokenTxCmd.AddCommand(client.PostCommands(
		GetCmdIssueToken(cdc),
	)...)

	return tokenTxCmd
}

const (
	flagMintable = "mintable"
)

func GetCmdIssueToken(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue [name] [symbol] [total-supply-in-micro]",
		Short: "Issue a new token and deposit the total supply to the owner's account",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			name := types.Name(args[0])
			symbol := types.ShortSymbol(args[1])
			totalSupplyMicro, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return types.ErrInvalidTotalSupply("invalid total-supply-in-micro: " + args[2])
			}
			mintable := viper.GetBool(flagMintable)
			ownerAddr := cliCtx.GetFromAddress()

			msg := types.NewMsgIssueToken(name, symbol, totalSupplyMicro, mintable, ownerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().Bool(flagMintable, false, "Make the new token mintable")
	return cmd
}
