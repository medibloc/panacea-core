package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
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

			msg, err := newCreatePoolMsg(clientCtx, args[0])
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

func newCreatePoolMsg(clientCtx client.Context, file string) (sdk.Msg, error) {
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

func CmdSellData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell-data [sell data file]",
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
func CmdBuyDataAccessNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy-data-access-nft [pool ID] [round] [payment]",
		Short: "buy data access NFT",
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
