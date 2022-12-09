package cli

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdRegisterOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-oracle",
		Short: "Register a new oracle",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			oracleAddress := clientCtx.GetFromAddress().String()

			uniqueID, err := cmd.Flags().GetString(flagOracleUniqueID)
			if err != nil {
				return fmt.Errorf("failed to get oracle unique ID")
			}

			nodePubkeyStr, err := cmd.Flags().GetString(flagNodePublicKey)
			if err != nil {
				return fmt.Errorf("failed to get node public key")
			}

			nodePubKey, err := base64.StdEncoding.DecodeString(nodePubkeyStr)
			if err != nil {
				return err
			}

			nodePubkeyRemoteReportStr, err := cmd.Flags().GetString(flagNodePubKeyRemoteReport)
			if err != nil {
				return fmt.Errorf("failed to get node public key remote report")
			}

			nodePubKeyRemoteReport, err := base64.StdEncoding.DecodeString(nodePubkeyRemoteReportStr)
			if err != nil {
				return err
			}

			trustedBlockHeightStr, err := cmd.Flags().GetString(flagTrustedBlockHeight)
			if err != nil {
				return fmt.Errorf("failed to get trsuted block height")
			}

			trustedBlockHeight, err := strconv.ParseInt(trustedBlockHeightStr, 10, 64)
			if err != nil {
				return err
			}

			trustedBlockHashStr, err := cmd.Flags().GetString(flagTrustedBlockHash)
			if err != nil {
				return fmt.Errorf("failed to get trsuted block hash")
			}

			trustedBlockHash, err := hex.DecodeString(trustedBlockHashStr)
			if err != nil {
				return err
			}

			endpoint, err := cmd.Flags().GetString(flagOracleEndpoint)
			if err != nil {
				return fmt.Errorf("failed to get oralce end point")
			}

			oracleCommissionRateStr, err := cmd.Flags().GetString(flagOracleCommRate)
			if err != nil {
				return fmt.Errorf("failed to get oralce commission rate")
			}

			if len(oracleCommissionRateStr) == 0 {
				return fmt.Errorf("oracleCommissionRate is empty")
			}

			oracleCommissionRate, err := sdk.NewDecFromStr(oracleCommissionRateStr)
			if err != nil {
				return err
			}

			oracleCommissionMaxRateStr, err := cmd.Flags().GetString(flagOracleCommMaxRate)
			if err != nil {
				return fmt.Errorf("failed to get oralce commission max rate")
			}

			if len(oracleCommissionMaxRateStr) == 0 {
				return fmt.Errorf("oracleCommissionMaxRate is empty")
			}

			oracleCommissionMaxRate, err := sdk.NewDecFromStr(oracleCommissionMaxRateStr)
			if err != nil {
				return err
			}

			oracleCommissionMaxChangeRateStr, err := cmd.Flags().GetString(flagOracleCommMaxChangeRate)
			if err != nil {
				return fmt.Errorf("failed to get oralce commission max change rate")
			}

			if len(oracleCommissionMaxChangeRateStr) == 0 {
				return fmt.Errorf("oracleCommissionMaxChangeRate is empty")
			}

			oracleCommissionMaxChangeRate, err := sdk.NewDecFromStr(oracleCommissionMaxChangeRateStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterOracle(
				uniqueID,
				oracleAddress,
				nodePubKey,
				nodePubKeyRemoteReport,
				trustedBlockHeight,
				trustedBlockHash,
				endpoint,
				oracleCommissionRate,
				oracleCommissionMaxRate,
				oracleCommissionMaxChangeRate,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagOracleUniqueID, "", "unique ID for oracle")
	cmd.Flags().String(flagNodePublicKey, "", "node public key for oracle")
	cmd.Flags().String(flagNodePubKeyRemoteReport, "", "node public key remote report for oracle")
	cmd.Flags().String(flagTrustedBlockHeight, "", "trusted block height of panacea trusted block")
	cmd.Flags().String(flagTrustedBlockHash, "", "trusted block hash of panacea trusted block")
	cmd.Flags().String(flagOracleEndpoint, "", "oracle's endpoint")
	cmd.Flags().String(flagOracleCommRate, "", "oracle's commission rate")
	cmd.Flags().String(flagOracleCommMaxRate, "", "oracle's commission max rate")
	cmd.Flags().String(flagOracleCommMaxChangeRate, "", "oracle's commission max change rate")

	if err := cmd.MarkFlagRequired(flagOracleUniqueID); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flagNodePublicKey); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flagNodePubKeyRemoteReport); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flagTrustedBlockHeight); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flagTrustedBlockHash); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flagOracleCommRate); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flagOracleCommMaxRate); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(flagOracleCommMaxChangeRate); err != nil {
		panic(err)
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
