package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/spf13/cobra"
)

func CmdRegisterDataValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-data-validator [endpoint URL]",
		Short: "register data validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromDataValidatorAddress := clientCtx.GetFromAddress()
			dataValidator := types.DataValidator{
				Address:  fromDataValidatorAddress.String(),
				Endpoint: args[0],
			}
			msg := types.NewMsgRegisterDataValidator(&dataValidator)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
