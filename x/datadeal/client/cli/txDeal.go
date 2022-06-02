package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func CmdCreateDeal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-deal [deal-file]",
		Short: "create a new deal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := newCreateDealMsg(clientCtx, args[0])
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdSellData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell-data [sell-data-file]",
		Short: "sell data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return nil
			}

			cert, err := readDataCertFile(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgSellData{
				Cert:   cert,
				Seller: clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdDeactivateDeal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-deal [deal-id]",
		Short: "deactivate deal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dealID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			requesterAddress := clientCtx.GetFromAddress()
			msg := types.NewMsgDeactivateDeal(dealID, requesterAddress.String())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newCreateDealMsg(clientCtx client.Context, file string) (sdk.Msg, error) {
	var createDeal createDealInputs

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(contents, &createDeal); err != nil {
		return nil, err
	}

	budget, err := sdk.ParseCoinNormalized(createDeal.Budget)
	if err != nil {
		return nil, err
	}

	msg := types.NewMsgCreateDeal(
		createDeal.DataSchema,
		&budget,
		createDeal.MaxNumData,
		createDeal.TrustedDataValidators,
		clientCtx.GetFromAddress().String(),
	)

	return msg, nil
}

func readDataCertFile(file string) (*types.DataValidationCertificate, error) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var cert types.DataValidationCertificate

	if err := json.Unmarshal(contents, &cert); err != nil {
		return nil, err
	}

	return &cert, nil
}
