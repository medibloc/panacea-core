package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/x/aol/types"
	"github.com/spf13/cobra"
)

func CmdGetWriter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-writer [ownerAddress] [topicName] [writerAddress]",
		Short: "get a writer",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetWriterRequest{
				OwnerAddress:  args[0],
				TopicName:     args[1],
				WriterAddress: args[2],
			}

			res, err := queryClient.Writer(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdListWriters() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-writer [ownerAddress] [topicName]",
		Short: "list all writers of <ownerAddress, topicName>",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryListWritersRequest{
				OwnerAddress: args[0],
				TopicName:    args[1],
				Pagination:   pageReq,
			}

			res, err := queryClient.Writers(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
