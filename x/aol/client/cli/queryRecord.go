package cli

import (
	"context"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/x/aol/types"
	"github.com/spf13/cobra"
)

func CmdGetRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-record [ownerAddress] [topicName] [offset]",
		Short: "get a record",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			offset, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			params := &types.QueryGetRecordRequest{
				OwnerAddress: args[0],
				TopicName:    args[1],
				Offset:       offset,
			}

			res, err := queryClient.Record(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
