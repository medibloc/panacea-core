package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/spf13/cobra"
)

func CmdRegisterDataValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-data-validator [endpoint URL]",
		Short: "register data validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromDataValidatorAddress := clientCtx.GetFromAddress()
			dataValidator := types.DataValidator{
				Address:  fromDataValidatorAddress.String(),
				Endpoint: args[0],
			}
			msg := types.NewMsgRegisterDataValidator(&dataValidator)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUpdateDataValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-data-validator [endpoint URL]",
		Short: "update data validator endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return nil
			}

			fromAddress := clientCtx.GetFromAddress()

			msg := types.NewMsgUpdateDataValidator(fromAddress.String(), args[0])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdCreatePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [pool params file]",
		Short: "create a new data pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := NewCreatePoolMsg(clientCtx, args[0])
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

func NewCreatePoolMsg(clientCtx client.Context, file string) (sdk.Msg, error) {
	var poolParamsInput createPoolInput

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file")
	}

	if err := json.Unmarshal(contents, &poolParamsInput); err != nil {
		return nil, err
	}

	nftPrice, err := sdk.ParseCoinNormalized(poolParamsInput.NFTPrice)
	if err != nil {
		return nil, err
	}

	downloadPeriod, err := time.ParseDuration(poolParamsInput.DownloadPeriod)
	if err != nil {
		return nil, err
	}

	poolParams := &types.PoolParams{
		DataSchema:            poolParamsInput.DataSchema,
		TargetNumData:         poolParamsInput.TargetNumData,
		MaxNftSupply:          poolParamsInput.MaxNFTSupply,
		NftPrice:              &nftPrice,
		TrustedDataValidators: poolParamsInput.TrustedDataValidators,
		TrustedDataIssuers:    poolParamsInput.TrustedDataIssuers,
		DownloadPeriod:        &downloadPeriod,
	}

	msg := types.NewMsgCreatePool(poolParams, clientCtx.GetFromAddress().String())
	return msg, nil
}
