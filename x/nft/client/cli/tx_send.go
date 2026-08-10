package cli

import (
	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

// NewSendCmd transfers an NFT through the standard MsgSend route implemented
// by Panacea's policy-aware wrapper.
func NewSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [class-id] [nft-id] [receiver]",
		Short: "Transfer ownership of an NFT",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientCtx = clientCtx.
				WithCmdContext(cmd.Context()).
				WithOutput(cmd.OutOrStdout())

			msg := &upstreamnft.MsgSend{
				ClassId:  args[0],
				Id:       args[1],
				Sender:   clientCtx.GetFromAddress().String(),
				Receiver: args[2],
			}
			return clienttx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
		SilenceUsage: true,
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
