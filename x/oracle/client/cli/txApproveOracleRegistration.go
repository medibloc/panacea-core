package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdApproveOracleRegistration() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve-oracle-registration [unique ID] [target oracle address] [encrypted oracle private key] [signature]",
		Short: "approve a new oracle registration",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			oracleAddress := clientCtx.GetFromAddress().String()

			approveOracleRegistration := &types.ApproveOracleRegistration{
				UniqueId:               args[0],
				ApproverOracleAddress:  oracleAddress,
				TargetOracleAddress:    args[1],
				EncryptedOraclePrivKey: []byte(args[2]),
			}

			msg := types.NewMsgApproveOracleRegistration(approveOracleRegistration, []byte(args[3]))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
