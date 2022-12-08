package cli

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdRegisterOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-oracle [unique-ID] [node-public-key] [node-public-key-remote-report] [trusted-block-height] [trusted-block-hash] [endpoint] [oracle-commission-rate] [oracle-commission-max-rate] [oracle-commission-max-change-rate]",
		Short: "Register a new oracle",
		Args:  cobra.ExactArgs(9),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			oracleAddress := clientCtx.GetFromAddress().String()
			nodePubKey, err := base64.StdEncoding.DecodeString(args[1])
			if err != nil {
				return err
			}

			nodePubKeyRemoteReport, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return err
			}

			trustedBlockHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			trustedBlockHash, err := hex.DecodeString(args[4])
			if err != nil {
				return err
			}

			endpoint := args[5]

			oracleCommissionRateStr := args[6]

			if len(oracleCommissionRateStr) == 0 {
				return fmt.Errorf("oracleCommissionRate is empty")
			}

			oracleCommissionRate, err := sdk.NewDecFromStr(oracleCommissionRateStr)
			if err != nil {
				return err
			}

			oracleCommissionMaxRateStr := args[7]

			if len(oracleCommissionMaxRateStr) == 0 {
				return fmt.Errorf("oracleCommissionMaxRate is empty")
			}

			oracleCommissionMaxRate, err := sdk.NewDecFromStr(oracleCommissionMaxRateStr)
			if err != nil {
				return err
			}

			oracleCommissionMaxChangeRateStr := args[8]

			if len(oracleCommissionMaxChangeRateStr) == 0 {
				return fmt.Errorf("oracleCommissionMaxChangeRate is empty")
			}

			oracleCommissionMaxChangeRate, err := sdk.NewDecFromStr(oracleCommissionMaxChangeRateStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterOracle(
				args[0],
				oracleAddress,
				nodePubKey,
				nodePubKeyRemoteReport,
				trustedBlockHeight,
				trustedBlockHash,
				endpoint,
				oracleCommissionRate,
				oracleCommissionMaxRate,
				oracleCommissionMaxChangeRate,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
