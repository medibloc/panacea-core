package cli

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/google/uuid"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/spf13/cobra"
)

func NewCmdCreateDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create-denom",
		Long: "Create a new denom.",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := cmd.Flags().GetString(FlagDenomId)

			if err != nil {
				return err
			}

			if id == "" {
				id = uuid.New().String()
			}

			symbol, err := cmd.Flags().GetString(FlagDenomSymbol)
			if err != nil {
				return err
			}

			denomName, err := cmd.Flags().GetString(FlagDenomName)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(FlagDenomDescription)
			if err != nil {
				return err
			}

			uri, err := cmd.Flags().GetString(FlagDenomURI)
			if err != nil {
				return err
			}

			uriHash, err := cmd.Flags().GetString(FlagDenomURIHash)
			if err != nil {
				return err
			}

			data, err := cmd.Flags().GetString(FlagDenomData)
			if err != nil {
				return err
			}

			creator := clientCtx.GetFromAddress().String()

			msg := types.NewMsgCreateDenomRequest(
				id,
				symbol,
				denomName,
				description,
				uri,
				uriHash,
				creator,
				data,
			)

			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(types.ErrCreateDenom, err.Error())
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagDenomName)
	cmd.Flags().String(FlagDenomId, "", "Set the id for a Denom. If this value is empty, a random value will be generated")
	cmd.Flags().String(FlagDenomSymbol, "", "Set the symbol for a Denom")
	cmd.Flags().String(FlagDenomName, "", "Set the name for a Denom")
	cmd.Flags().String(FlagDenomDescription, "", "Set the description for a Denom")
	cmd.Flags().String(FlagDenomURI, "", "Set the URI for a Denom")
	cmd.Flags().String(FlagDenomURIHash, "", "Set the URI hash for a Denom")
	cmd.Flags().String(FlagDenomData, "", "Set the data for a Denom")
	return cmd
}

func NewCmdUpdateDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "update-denom [denom-id]",
		Long: "update a exist denom.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denomId := args[0]

			symbol, err := cmd.Flags().GetString(FlagDenomSymbol)
			if err != nil {
				return err
			}

			denomName, err := cmd.Flags().GetString(FlagDenomName)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(FlagDenomDescription)
			if err != nil {
				return err
			}

			uri, err := cmd.Flags().GetString(FlagDenomURI)
			if err != nil {
				return err
			}

			uriHash, err := cmd.Flags().GetString(FlagDenomURIHash)
			if err != nil {
				return err
			}

			data, err := cmd.Flags().GetString(FlagDenomData)
			if err != nil {
				return err
			}

			updater := clientCtx.GetFromAddress().String()

			msg := types.NewMsgUpdateDenomRequest(
				denomId,
				symbol,
				denomName,
				description,
				uri,
				uriHash,
				data,
				updater,
			)

			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(types.ErrUpdateDenom, err.Error())
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(FlagDenomName)
	cmd.Flags().String(FlagDenomSymbol, "", "Set the symbol for a Denom")
	cmd.Flags().String(FlagDenomName, "", "Set the name for a Denom")
	cmd.Flags().String(FlagDenomDescription, "", "Set the description for a Denom")
	cmd.Flags().String(FlagDenomURI, "", "Set the URI for a Denom")
	cmd.Flags().String(FlagDenomURIHash, "", "Set the URI hash for a Denom")
	cmd.Flags().String(FlagDenomData, "", "Set the data for a Denom")
	return cmd
}

func NewCmdDeleteDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "delete-denom [denom-id]",
		Long: "Delete a exist denom.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denomId := args[0]

			remover := clientCtx.GetFromAddress().String()

			msg := types.NewMsgDeleteDenomRequest(denomId, remover)

			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(types.ErrDeleteDenom, err.Error())
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewCmdTransferDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "transfer-denom [denom-id] [receiver]",
		Long: "Transfer denom owner to receiver.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denomId := args[0]
			receiver := args[1]

			sender := clientCtx.GetFromAddress().String()

			msg := types.NewMsgTransferRequest(denomId, sender, receiver)

			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(types.ErrTransferDenom, err.Error())
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
