package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/medibloc/panacea-core/x/token/client/internal"
	"github.com/medibloc/panacea-core/x/token/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	tokenQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the token module",
	}

	tokenQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryToken(cdc),
		GetCmdListTokens(cdc),
	)...)

	return tokenQueryCmd
}

func GetCmdQueryToken(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-token [symbol]",
		Short: "Get an token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			symbol := types.Symbol(args[0])

			bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryTokenParams(symbol))
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(internal.RouteToken, bz)
			if err != nil {
				return err
			}

			var token types.Token
			cliCtx.Codec.MustUnmarshalJSON(res, &token)
			if token.Empty() {
				return types.ErrTokenNotExists(symbol)
			}

			return cliCtx.PrintOutput(token)
		},
	}
	return cmd
}

func GetCmdListTokens(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-tokens",
		Short: "List all token symbols issued",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.Query(internal.RouteListTokens)
			if err != nil {
				return err
			}

			var symbols types.Tokens
			cliCtx.Codec.MustUnmarshalJSON(res, &symbols)
			return cliCtx.PrintOutput(symbols)
		},
	}
	return cmd
}
