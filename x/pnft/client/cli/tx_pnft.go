package cli

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/spf13/cobra"
)

func NewCmdMintPNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "mint-pnft [denom-id] [id]",
		Long: "Mint a new pnft.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denomId := args[0]
			id := args[1]

			name, err := cmd.Flags().GetString(flagPNFTName)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(flagPNFTDescription)
			if err != nil {
				return err
			}

			uri, err := cmd.Flags().GetString(flagPNFTUri)
			if err != nil {
				return err
			}

			uriHash, err := cmd.Flags().GetString(flagPNFTUriHash)
			if err != nil {
				return err
			}

			data, err := cmd.Flags().GetString(flagPNFTData)
			if err != nil {
				return err
			}

			creator := clientCtx.GetFromAddress().String()

			msg := types.NewMsgMintPNFTRequest(
				denomId,
				id,
				name,
				description,
				uri,
				uriHash,
				creator,
				data,
			)

			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(types.ErrMintPNFT, err.Error())
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flagPNFTName)
	cmd.Flags().String(flagPNFTName, "", "Set the name for a PNFT.")
	cmd.Flags().String(flagPNFTDescription, "", "Set the name for a PNFT.")
	cmd.Flags().String(flagPNFTUri, "", "Set the name for a PNFT.")
	cmd.Flags().String(flagPNFTUriHash, "", "Set the name for a PNFT.")
	cmd.Flags().String(flagPNFTData, "", "Set the name for a PNFT.")

	return cmd
}

func NewCmdTransferPNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "transfer-pnft [denom-id] [id] [receiver]",
		Long: "Mint a new pnft.",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return sdkerrors.Wrap(types.ErrTransferPNFT, err.Error())
			}

			denomId := args[0]
			id := args[1]
			receiver := args[2]
			sender := clientCtx.GetFromAddress().String()

			msg := types.NewMsgTransferPNFTRequest(denomId, id, sender, receiver)

			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(types.ErrTransferPNFT, err.Error())
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCmdBurnPNFT() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "burn-pnft [denom-id] [id]",
		Long: "Mint a new pnft.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return sdkerrors.Wrap(types.ErrTransferPNFT, err.Error())
			}

			denomId := args[0]
			id := args[1]
			burner := clientCtx.GetFromAddress().String()

			msg := types.NewMsgBurnPNFTRequest(denomId, id, burner)

			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(types.ErrTransferPNFT, err.Error())
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
