package cli

import (
	"context"
	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/spf13/cobra"
)

func NewCmdGetPNFTs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-pnft [denom-id]",
		Short: "List pnft by denomId",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			denomId := args[0]
			queryClient := types.NewQueryServiceClient(clientCtx)

			msg := types.NewQueryServicePNFTsRequest(denomId)
			if err := msg.ValidateBasic(); err != nil {
				return errors.Wrap(types.ErrGetPNFT, err.Error())
			}

			res, err := queryClient.PNFTs(context.Background(), msg)

			if err != nil {
				return errors.Wrap(types.ErrGetPNFT, err.Error())
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func NewCmdGetPNFTsByOwner() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-pnft-by-owner [denom-id] [owner]",
		Short: "List pnft by denomId and owner",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			denomId := args[0]
			owner := args[1]

			queryClient := types.NewQueryServiceClient(clientCtx)

			msg := types.NewQueryServicePNFTsByOwnerRequest(denomId, owner)
			if err := msg.ValidateBasic(); err != nil {
				return errors.Wrap(types.ErrGetPNFT, err.Error())
			}

			res, err := queryClient.PNFTsByDenomOwner(context.Background(), msg)

			if err != nil {
				return errors.Wrap(types.ErrGetPNFT, err.Error())
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func NewCmdGetPNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-pnf [denom-id] [id]",
		Short: "Get pnft by denomId and id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			denomId := args[0]
			id := args[1]

			queryClient := types.NewQueryServiceClient(clientCtx)

			msg := types.NewQueryServicePNFTRequest(denomId, id)
			if err := msg.ValidateBasic(); err != nil {
				return errors.Wrap(types.ErrGetPNFT, err.Error())
			}

			res, err := queryClient.PNFT(context.Background(), msg)

			if err != nil {
				return errors.Wrap(types.ErrGetPNFT, err.Error())
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
