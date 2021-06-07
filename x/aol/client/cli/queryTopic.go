package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/x/aol/types"
	"github.com/spf13/cobra"
)

func CmdGetTopic() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-topic [ownerAddress] [topicName]",
		Short: "get a topic",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetTopicRequest{
				OwnerAddress: args[0],
				TopicName:    args[1],
			}

			res, err := queryClient.Topic(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdListTopics() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-topic [ownerAddress]",
		Short: "list all topics of the owner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryListTopicsRequest{
				OwnerAddress: args[0],
				Pagination:   pageReq,
			}

			res, err := queryClient.Topics(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
