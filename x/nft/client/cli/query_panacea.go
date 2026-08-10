package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

const (
	flagClassID = "class-id"
	flagOwner   = "owner"
)

// NewClassRecordQueryCmd queries a standard class and its Panacea policy.
func NewClassRecordQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "class-record [class-id]",
		Short: "Query an NFT class and its Panacea policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := queryClientContext(cmd)
			if err != nil {
				return err
			}
			response, err := types.NewQueryClient(clientCtx).ClassRecord(
				cmd.Context(),
				&types.QueryClassRecordRequest{ClassId: args[0]},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewNFTRecordQueryCmd queries a live NFT or permanent burn tombstone.
func NewNFTRecordQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nft-record [class-id] [nft-id]",
		Short: "Query a live NFT or burn tombstone with its lifecycle",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := queryClientContext(cmd)
			if err != nil {
				return err
			}
			response, err := types.NewQueryClient(clientCtx).NFTRecord(
				cmd.Context(),
				&types.QueryNFTRecordRequest{
					ClassId: args[0],
					NftId:   args[1],
				},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(response)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// NewNFTRecordsQueryCmd lists live NFT records using optional class and owner
// filters. The server requires at least one filter.
func NewNFTRecordsQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nft-records",
		Short: "Query live NFT records by class, owner, or both",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := queryClientContext(cmd)
			if err != nil {
				return err
			}
			pageFlags, err := client.FlagSetWithPageKeyDecoded(cmd.Flags())
			if err != nil {
				return err
			}
			pageRequest, err := client.ReadPageRequest(pageFlags)
			if err != nil {
				return err
			}
			classID, err := cmd.Flags().GetString(flagClassID)
			if err != nil {
				return err
			}
			owner, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}

			response, err := types.NewQueryClient(clientCtx).NFTRecords(
				cmd.Context(),
				&types.QueryNFTRecordsRequest{
					ClassId:    classID,
					Owner:      owner,
					Pagination: pageRequest,
				},
			)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(response)
		},
	}

	cmd.Flags().String(flagClassID, "", "filter by canonical class ID")
	cmd.Flags().String(flagOwner, "", "filter by canonical owner address")
	cmd.Flags().String(flags.FlagPageKey, "", "pagination cursor returned by the previous query")
	cmd.Flags().Uint64(flags.FlagLimit, 100, "maximum number of records to return (up to 100)")
	cmd.Flags().Bool(flags.FlagReverse, false, "return records in reverse order")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func queryClientContext(cmd *cobra.Command) (client.Context, error) {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return client.Context{}, err
	}
	return clientCtx.
		WithCmdContext(cmd.Context()).
		WithOutput(cmd.OutOrStdout()), nil
}
