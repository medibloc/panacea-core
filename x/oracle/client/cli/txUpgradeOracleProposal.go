package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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

			content, err := makeProposalContent(cmd)
			if err != nil {
				return err
			}

			// TODO: govv1 has been introduced in v0.46 but its spec will be updated in v0.47 as below.
			//       Until then, it is safer to use govv1beta1.
			//
			// https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.0-rc2
			// > In v0.46 with x/gov v1, these fields were not present (while present in v1beta1). After community feedback, they have been added in x/gov v1.
			msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
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

	if err := cmd.MarkFlagRequired(FlagUpgradeHeight); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired(FlagUpgradeUniqueID); err != nil {
		panic(err)
	}
	return cmd
}

func makeProposalContent(cmd *cobra.Command) (govv1beta1.Content, error) {
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
	if err := plan.ValidateBasic(); err != nil {
		return nil, err
	}
	content := types.NewOracleUpgradeProposal(title, description, plan)
	return content, nil
}
