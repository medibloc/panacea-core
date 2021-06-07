package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/x/aol/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagDescription = "description"
	flagMoniker     = "moniker"
	flagFeePayer    = "payer"
)

func CmdCreateTopic() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-topic [topicName]",
		Short: "Creates a new topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			topicName := args[0]
			description := viper.GetString(flagDescription)
			ownerAddress := clientCtx.GetFromAddress().String()

			msg := types.NewMsgCreateTopic(topicName, description, ownerAddress)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(flagDescription, "", "description of topic")

	return cmd
}
