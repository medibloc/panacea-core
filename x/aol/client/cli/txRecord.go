package cli

import (
	"github.com/spf13/cobra"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/x/aol/types"
)

func CmdCreateRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-record [key] [value] [nanoTimestamp] [writerAddress]",
		Short: "Creates a new record",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsKey := string(args[0])
			argsValue := string(args[1])
			argsNanoTimestamp, _ := strconv.ParseInt(args[2], 10, 64)
			argsWriterAddress := string(args[3])

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateRecord(clientCtx.GetFromAddress().String(), string(argsKey), string(argsValue), int32(argsNanoTimestamp), string(argsWriterAddress))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdUpdateRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-record [id] [key] [value] [nanoTimestamp] [writerAddress]",
		Short: "Update a record",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			argsKey := string(args[1])
			argsValue := string(args[2])
			argsNanoTimestamp, _ := strconv.ParseInt(args[3], 10, 64)
			argsWriterAddress := string(args[4])

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateRecord(clientCtx.GetFromAddress().String(), id, string(argsKey), string(argsValue), int32(argsNanoTimestamp), string(argsWriterAddress))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDeleteRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-record [id] [key] [value] [nanoTimestamp] [writerAddress]",
		Short: "Delete a record by id",
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

			msg := types.NewMsgDeleteRecord(clientCtx.GetFromAddress().String(), id)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
