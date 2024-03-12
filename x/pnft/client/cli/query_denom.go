package cli

import (
	"context"
	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/spf13/cobra"
)

func NewCmdGetDenoms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-denom",
		Short: "List all denoms",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			pagination, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return errors.Wrap(types.ErrGetDenom, err.Error())
			}

			queryClient := types.NewQueryClient(clientCtx)

			msg := types.NewQueryDenomsRequest(pagination)
			if err := msg.ValidateBasic(); err != nil {
				return errors.Wrap(types.ErrGetDenom, err.Error())
			}

			res, err := queryClient.Denoms(context.Background(), msg)

			if err != nil {
				return errors.Wrap(types.ErrGetDenom, err.Error())
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func NewCmdGetDenomsByOwner() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-denom-by-owner [owner]",
		Short: "List denoms by owner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			owner := args[0]

			queryClient := types.NewQueryClient(clientCtx)

			msg := types.NewQueryDenomsByOwnerRequest(owner)

			if err := msg.ValidateBasic(); err != nil {
				return errors.Wrap(types.ErrGetDenom, err.Error())
			}

			res, err := queryClient.DenomsByOwner(context.Background(), msg)

			if err != nil {
				return errors.Wrap(types.ErrGetDenom, err.Error())
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func NewCmdGetDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-denom [denom-id]",
		Short: "Get denom",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			denomId := args[0]

			queryClient := types.NewQueryClient(clientCtx)

			msg := types.NewQueryDenomRequest(denomId)

			if err := msg.ValidateBasic(); err != nil {
				return errors.Wrap(types.ErrGetDenom, err.Error())
			}

			res, err := queryClient.Denom(context.Background(), msg)

			if err != nil {
				return errors.Wrap(types.ErrGetDenom, err.Error())
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
