package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdGetOracleRegistrations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-registrations [unique-id]",
		Short: "Query oracle-registrations info",
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

			req := &types.QueryOracleRegistrationsRequest{
				UniqueId:   args[0],
				Pagination: pageReq,
			}
			res, err := queryClient.OracleRegistrations(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "oracle-registrations")

	return cmd
}

func CmdGetOracleRegistration() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-registration [unique-id] [oracle-address]",
		Short: "Query a oracle registration info",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.OracleRegistration(cmd.Context(), &types.QueryOracleRegistrationRequest{
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
