package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdGetOracleRegistrationVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle [unique_id] [voting_target_address] [voter_address]",
		Short: "Query a oracleRegistration vote info",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			_, err = sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return errors.Wrap(err, "voting_target_address is invalid")
			}

			_, err = sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return errors.Wrap(err, "voter_address is invalid")
			}

			queryClient := types.NewQueryClient(clientCtx)
			req := &types.QueryOracleRegistrationVoteRequest{
				UniqueId:            args[0],
				VotingTargetAddress: args[1],
				VoterAddress:        args[2],
			}
			res, err := queryClient.OracleRegistrationVote(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
