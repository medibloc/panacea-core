package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/spf13/cobra"
)

const (
	FlagUpgradeUniqueID = "upgrade-unique-id"
	FlagUpgradeHeight   = "upgrade-height"
)

func CmdUpgradeOracleProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle-upgrade (--upgrade-unique-id [uniqueID]) (--upgrade-height [height]) [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit an oracle upgrade proposal",
		Long: "Submit an oracle upgrade proposal along with an initial deposit.\n + " +
			"You must enter the uniqueID of a new version of oracle and target block height for upgrade.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(govcli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			content, err := parseArgsToContent(cmd)
			if err != nil {
				return err
			}

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(govcli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(govcli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Int64(FlagUpgradeHeight, 0, "The height at which the upgrade must happen")
	cmd.Flags().String(FlagUpgradeUniqueID, "", "Oracle's uniqueID to be upgraded")

	return cmd
}

func parseArgsToContent(cmd *cobra.Command) (govtypes.Content, error) {
	title, err := cmd.Flags().GetString(govcli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(govcli.FlagDescription)
	if err != nil {
		return nil, err
	}

	height, err := cmd.Flags().GetInt64(FlagUpgradeHeight)
	if err != nil {
		return nil, err
	}

	uniqueID, err := cmd.Flags().GetString(FlagUpgradeUniqueID)
	if err != nil {
		return nil, err
	}

	plan := types.Plan{UniqueId: uniqueID, Height: height}
	content := types.NewOracleUpgradeProposal(title, description, plan)
	return content, nil
}
