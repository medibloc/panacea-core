package cli

import (
	"errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/aol/types"
	"github.com/spf13/cobra"
)

const (
	RouteListTopic  = "custom/aol/listTopic"
	RouteTopic      = "custom/aol/topic"
	RouteListWriter = "custom/aol/listWriter"
	RouteWriter     = "custom/aol/writer"
	RouteRecord     = "custom/aol/record"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	aolQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the aol module",
	}

	aolQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryListTopic(cdc),
		GetCmdQueryListWriter(cdc),
		GetCmdQueryRecord(cdc),
		GetCmdQueryTopic(cdc),
		GetCmdQueryWriter(cdc),
	)...)

	return aolQueryCmd
}

func GetCmdQueryListTopic(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "list-topics [owner_address]",
		Short: "List topics",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			ownerAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.QueryListTopicParams{Owner: ownerAddr}
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(RouteListTopic, bz)
			if err != nil {
				return err
			}

			var topics types.Topics
			cdc.MustUnmarshalJSON(res, &topics)
			return cliCtx.PrintOutput(topics)
		},
	}
}

func GetCmdQueryTopic(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-topic [owner_address] [topic_name]",
		Short: "Get a topic",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			ownerAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			topicName := args[1]

			params := types.QueryTopicParams{
				Owner:     ownerAddr,
				TopicName: topicName,
			}
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(RouteTopic, bz)
			if err != nil {
				return err
			}

			var topic types.Topic
			cdc.MustUnmarshalJSON(res, &topic)
			return cliCtx.PrintOutput(topic)
		},
	}
}

func GetCmdQueryListWriter(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "list-writers [owner_address] [topic_name]",
		Short: "List writers",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			ownerAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			topicName := args[1]

			params := types.QueryListWriterParams{
				Owner:     ownerAddr,
				TopicName: topicName,
			}
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(RouteListWriter, bz)
			if err != nil {
				return err
			}

			var writers types.Writers
			cdc.MustUnmarshalJSON(res, &writers)
			return cliCtx.PrintOutput(writers)
		},
	}
}

func GetCmdQueryWriter(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-writer [owner_address] [topic_name] [writer_address]",
		Short: "Get a writer",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			ownerAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			topicName := args[1]
			writerAddr, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			params := types.QueryWriterParams{
				Owner:     ownerAddr,
				TopicName: topicName,
				Writer:    writerAddr,
			}
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(RouteWriter, bz)
			if err != nil {
				return err
			}

			var writer types.Writer
			cdc.MustUnmarshalJSON(res, &writer)
			return cliCtx.PrintOutput(writer)
		},
	}
}

func GetCmdQueryRecord(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get-record [owner_address] [topic_name] [offset]",
		Short: "Get a record",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			ownerAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			topicName := args[1]
			offset, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			params := types.QueryRecordParams{
				Owner:     ownerAddr,
				TopicName: topicName,
				Offset:    offset,
			}
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(RouteRecord, bz)
			if err != nil {
				return err
			}

			var rec types.Record
			cdc.MustUnmarshalJSON(res, &rec)
			if rec.IsEmpty() {
				return errors.New("record not found")
			}
			return cliCtx.PrintOutput(rec)
		},
	}
}
