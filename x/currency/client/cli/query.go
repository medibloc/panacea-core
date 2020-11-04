package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/medibloc/panacea-core/x/currency/types"
	"github.com/spf13/cobra"
)

const (
	RouteIssuance = "custom/currency/issuance"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	currencyQueryCmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Querying commands for the currency module",
	}

	currencyQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryIssuance(cdc),
	)...)

	return currencyQueryCmd
}

func GetCmdQueryIssuance(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-issuance [denom]",
		Short: "Get an issuance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			denom := args[0]

			issuance, err := queryIssuance(cliCtx, denom)
			if err != nil {
				return err
			}
			return cliCtx.PrintOutput(issuance)
		},
	}
	return cmd
}

// queryIssuance gets a Issuance from the blockchain.
// It returns an error if the denom has never been issued.
func queryIssuance(cliCtx context.CLIContext, denom string) (types.Issuance, error) {
	bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryIssuanceParams(denom))
	if err != nil {
		return types.Issuance{}, err
	}

	res, _, err := cliCtx.QueryWithData(RouteIssuance, bz)
	if err != nil {
		return types.Issuance{}, err
	}

	var issuance types.Issuance
	cliCtx.Codec.MustUnmarshalJSON(res, &issuance)
	if issuance.Empty() {
		return types.Issuance{}, types.ErrDenomNotExists(denom)
	}

	return issuance, nil
}
