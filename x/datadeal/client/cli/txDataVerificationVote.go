package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func CmdVoteDataVerification() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data-verification-vote [path]",
		Short: "Vote for data verification",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := readMsgDataVerificationVoteFrom(args[0])
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

func readMsgDataVerificationVoteFrom(path string) (*types.MsgVoteDataVerification, error) {
	var msg types.MsgVoteDataVerification

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := jsonpb.Unmarshal(file, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
