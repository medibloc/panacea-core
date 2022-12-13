package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdUpdateOracleInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-oracle-info",
		Short: "update an oracle information",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			oracleAddress := clientCtx.GetFromAddress().String()

			endpoint, err := cmd.Flags().GetString(flagOracleEndpoint)
			if err != nil {
				return fmt.Errorf("failed to get oracle end point")
			}

			oracleCommissionRateStr, err := cmd.Flags().GetString(flagOracleCommRate)
			if err != nil {
				return fmt.Errorf("failed to get oralce commission rate")
			}

			if len(oracleCommissionRateStr) == 0 {
				return fmt.Errorf("oracleCommissionRate is empty")
			}

			oracleCommissionRate, err := sdk.NewDecFromStr(oracleCommissionRateStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateOracleInfo(oracleAddress, endpoint, oracleCommissionRate)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagOracleEndpoint, "", "oracle's endpoint")
	cmd.Flags().String(flagOracleCommRate, "", "oracle's commission rate")

	if err := cmd.MarkFlagRequired(flagOracleCommRate); err != nil {
		panic(err)
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
