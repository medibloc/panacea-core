package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func CmdCreateDeal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-deal [deal-file-path]",
		Short: "create a new deal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := newCreateDealMsg(args[0])
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func newCreateDealMsg(file string) (sdk.Msg, error) {
	var msg *types.MsgCreateDeal

	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(contents, &msg); err != nil {
		return nil, err
	}

	return msg, nil
}
