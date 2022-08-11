package cli

import (
	"encoding/base64"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdRegisterOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-oracle [unique ID] [node public key] [node public key remote report] [trusted block height] [trusted block hash]",
		Short: "Register a new oracle",
		Args:  cobra.ExactArgs(5),
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

			trustedBlockHash, err := base64.StdEncoding.DecodeString(args[4])
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterOracle(args[0], oracleAddress, nodePubKey, nodePubKeyRemoteReport, trustedBlockHeight, trustedBlockHash)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
