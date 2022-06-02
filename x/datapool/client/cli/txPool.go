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
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/spf13/cobra"
)

func CmdCreatePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [deposit] [pool-params-file]",
		Short: "create a new data pool",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := newCreatePoolMsg(clientCtx, args[0], args[1])
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

func newCreatePoolMsg(clientCtx client.Context, depositCoin, file string) (sdk.Msg, error) {
	var poolParamsInput CreatePoolInput

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

	poolParams := &types.PoolParams{
		DataSchema:         poolParamsInput.DataSchema,
		TargetNumData:      poolParamsInput.TargetNumData,
		MaxNftSupply:       poolParamsInput.MaxNFTSupply,
		NftPrice:           &nftPrice,
		TrustedOracles:     poolParamsInput.TrustedOracles,
		TrustedDataIssuers: poolParamsInput.TrustedDataIssuers,
	}

	deposit, err := sdk.ParseCoinNormalized(depositCoin)
	if err != nil {
		return nil, err
	}

	msg := types.NewMsgCreatePool(poolParams, deposit, clientCtx.GetFromAddress().String())
	return msg, nil
}

func CmdSellData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell-data [sell-data-file]",
		Short: "sell data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			cert, err := readCertificateFromFile(args[0])
			if err != nil {
				return err
			}

			seller := clientCtx.GetFromAddress().String()
			msg := &types.MsgSellData{
				Cert:   cert,
				Seller: seller,
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

func readCertificateFromFile(file string) (*types.DataValidationCertificate, error) {
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

func CmdBuyDataPass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy-data-pass [pool-id] [round] [payment]",
		Short: "buy data pass",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			buyer := clientCtx.GetFromAddress()

			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			round, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			payment, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgBuyDataPass{
				PoolId:  poolID,
				Round:   round,
				Payment: &payment,
				Buyer:   buyer.String(),
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

func CmdRedeemDataPass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem-data-pass [pool-id] [round] [data-pass-id]",
		Short: "redeem data pass",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			redeemer := clientCtx.GetFromAddress()

			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			round, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			dataPassID, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgRedeemDataPass{
				PoolId:     poolID,
				Round:      round,
				DataPassId: dataPassID,
				Redeemer:   redeemer.String(),
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
