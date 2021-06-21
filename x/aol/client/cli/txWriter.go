package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func CmdAddWriter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-writer [topicName] [writerAddress]",
		Short: "Add write permission for this topic",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			topicName := args[0]
			writerAddress := args[1]
			moniker := viper.GetString(flagMoniker)
			description := viper.GetString(flagDescription)
			ownerAddress := clientCtx.GetFromAddress().String()

			msg := types.NewMsgAddWriter(topicName, moniker, description, writerAddress, ownerAddress)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(flagMoniker, "", "name of writer")
	cmd.Flags().String(flagDescription, "", "description of writer")

	return cmd
}

func CmdDeleteWriter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-writer [topicName] [writerAddress]",
		Short: "Delete write permission for this topic",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			topicName := args[0]
			writerAddress := args[1]
			ownerAddress := clientCtx.GetFromAddress().String()

			msg := types.NewMsgDeleteWriter(topicName, writerAddress, ownerAddress)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
