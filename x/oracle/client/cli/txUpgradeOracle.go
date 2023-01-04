package cli

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

func CmdUpgradeOracle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade-oracle",
		Short: "Upgrade an oracle",
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

			msg := types.NewMsgUpgradeOracle(uniqueID, oracleAddress, nodePubKey, nodePubKeyRemoteReport, trustedBlockHeight, trustedBlockHash)
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

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
