package cli

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/x/token/types"
)

const (
	flagMintable = "mintable"
)

func CmdIssueToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-token [name] [short-symbol] [total-supply-in-micro]",
		Short: "Issue a new token and deposit the total supply to the owner's account",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			shortSymbol := args[1]
			totalSupplyMicro, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return types.ErrInvalidTotalSupply
			}
			mintable := viper.GetBool(flagMintable)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			ownerAddr := clientCtx.GetFromAddress().String()

			msg := types.NewMsgIssueToken(name, shortSymbol, totalSupplyMicro, mintable, ownerAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Bool(flagMintable, false, "Make the new token mintable")

	return cmd
}
