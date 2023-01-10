package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdGetOracleUpgrades() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-upgrades [unique-id]",
		Short: "Query oracle-upgrade list",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOracleUpgradesRequest{
				UniqueId:   args[0],
				Pagination: pageReq,
			}
			res, err := queryClient.OracleUpgrades(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "oracle-upgrades")

	return cmd
}

func CmdGetOracleUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-upgrade [unique-id] [oracle-address]",
		Short: "Query an oracle upgrade",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.OracleUpgrade(cmd.Context(), &types.QueryOracleUpgradeRequest{
				UniqueId:      args[0],
				OracleAddress: args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
