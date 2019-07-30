package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/medibloc/panacea-core/x/aol/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagDescription = "description"
	flagMoniker     = "moniker"
	flagFeePayer    = "payer"
)

func GetCmdCreateTopic(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-topic [topic]",
		Short: "Create a topic",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

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
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
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
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

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
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
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
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

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
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func GetCmdAddRecord(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-record [owner_address] [topic] [key] [value]",
		Short: "Add new record",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

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
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
	cmd.Flags().String(flagFeePayer, "", "optional address to pay for the fee")
	return cmd
}
