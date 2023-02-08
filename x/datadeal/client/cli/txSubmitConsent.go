package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/spf13/cobra"
)

func CmdSubmitConsent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-consent [consent-file]",
		Short: "submit a consent to provide the data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			consent, err := newConsent(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitConsent(consent)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func newConsent(path string) (*types.Consent, error) {
	contents, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer contents.Close()

	consent := &types.Consent{}
	if err := jsonpb.Unmarshal(contents, consent); err != nil {
		return nil, err
	}

	return consent, nil
}
