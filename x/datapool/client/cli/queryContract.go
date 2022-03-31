package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/spf13/cobra"
)

func CmdGetNFTContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-nft-contract",
		Short: "Query registered NFT contract address",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.NFTContract(cmd.Context(), &types.QueryNFTContractRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
