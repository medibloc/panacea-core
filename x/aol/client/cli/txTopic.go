package cli

import (
	"github.com/spf13/cobra"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func CmdCreateTopic() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-topic [description] [totalRecords] [totalWriter]",
		Short: "Creates a new topic",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsDescription := string(args[0])
			argsTotalRecords, _ := strconv.ParseInt(args[1], 10, 64)
			argsTotalWriter, _ := strconv.ParseInt(args[2], 10, 64)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateTopic(clientCtx.GetFromAddress().String(), string(argsDescription), int32(argsTotalRecords), int32(argsTotalWriter))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdUpdateTopic() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-topic [id] [description] [totalRecords] [totalWriter]",
		Short: "Update a topic",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			argsDescription := string(args[1])
			argsTotalRecords, _ := strconv.ParseInt(args[2], 10, 64)
			argsTotalWriter, _ := strconv.ParseInt(args[3], 10, 64)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateTopic(clientCtx.GetFromAddress().String(), id, string(argsDescription), int32(argsTotalRecords), int32(argsTotalWriter))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDeleteTopic() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-topic [id] [description] [totalRecords] [totalWriter]",
		Short: "Delete a topic by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeleteTopic(clientCtx.GetFromAddress().String(), id)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
