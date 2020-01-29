package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/medibloc/panacea-core/x/aol/internal/types"
)

const (
	flagDescription = "description"
	flagMoniker     = "moniker"
	flagFeePayer    = "payer"
)


func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	aolTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Aol transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	aolTxCmd.AddCommand(client.PostCommands(
		GetCmdCreateTopic(cdc),
		GetCmdAddWriter(cdc),
		GetCmdDeleteWriter(cdc),
		GetCmdAddRecord(cdc),
	)...)

	return aolTxCmd
}

func GetCmdCreateTopic(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-topic [topic]",
		Short: "Create a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			ownerAddr := cliCtx.GetFromAddress()
			topic := args[0]
			description := viper.GetString(flagDescription)

			msg := types.MsgCreateTopic{
				TopicName:    topic,
				Description:  description,
				OwnerAddress: ownerAddr,
			}
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagDescription, "", "description of topic")
	return cmd
}

func GetCmdAddWriter(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-writer [topic] [writer_address]",
		Short: "Add write permission for this topic",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			ownerAddr := cliCtx.GetFromAddress()
			topic := args[0]
			description := viper.GetString(flagDescription)
			moniker := viper.GetString(flagMoniker)
			writerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgAddWriter{
				TopicName:     topic,
				Moniker:       moniker,
				Description:   description,
				WriterAddress: writerAddr,
				OwnerAddress:  ownerAddr,
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagMoniker, "", "name of writer")
	cmd.Flags().String(flagDescription, "", "description of writer")
	return cmd
}

func GetCmdDeleteWriter(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delete-writer [topic] [writer_address]",
		Short: "Delete write permission for this topic",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			ownerAddr := cliCtx.GetFromAddress()
			topic := args[0]
			writerAddr, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg := types.MsgDeleteWriter{
				TopicName:     topic,
				WriterAddress: writerAddr,
				OwnerAddress:  ownerAddr,
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdAddRecord(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-record [owner_address] [topic] [key] [value]",
		Short: "Add new record",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			ownerAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			topic := args[1]
			key := args[2]
			value := args[3]
			writerAddr := cliCtx.GetFromAddress()
			feePayerAddr, err := sdk.AccAddressFromBech32(viper.GetString(flagFeePayer))
			if err != nil {
				return err
			}

			msg := types.MsgAddRecord{
				TopicName:       topic,
				Key:             []byte(key),
				Value:           []byte(value),
				WriterAddress:   writerAddr,
				OwnerAddress:    ownerAddr,
				FeePayerAddress: feePayerAddr,
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagFeePayer, "", "optional address to pay for the fee")
	return cmd
}
