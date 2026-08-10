package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/spf13/cobra"

	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

const (
	flagURI     = "uri"
	flagURIHash = "uri-hash"
	flagData    = "data"
)

// NewMintCmd mints an NFT and resolves the closed NFTData interface with the
// application codec. AutoCLI's dynamic Any resolver cannot resolve metadata
// types that are deliberately not registered as executable sdk.Msg values.
func NewMintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [class-id] [nft-id] [recipient]",
		Short: "Mint an NFT to a recipient",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientCtx = clientCtx.
				WithCmdContext(cmd.Context()).
				WithOutput(cmd.OutOrStdout())

			uri, err := cmd.Flags().GetString(flagURI)
			if err != nil {
				return err
			}
			uriHash, err := cmd.Flags().GetString(flagURIHash)
			if err != nil {
				return err
			}
			dataJSON, err := cmd.Flags().GetString(flagData)
			if err != nil {
				return err
			}
			data, err := parseNFTData(clientCtx, dataJSON)
			if err != nil {
				return err
			}

			msg := &types.MsgMintRequest{
				ClassId:    args[0],
				NftId:      args[1],
				Controller: clientCtx.GetFromAddress().String(),
				Recipient:  args[2],
				Uri:        uri,
				UriHash:    uriHash,
				Data:       data,
			}
			return clienttx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().String(flagURI, "", "URI for the NFT metadata")
	cmd.Flags().String(flagURIHash, "", "Content hash for --uri in sha256:<64 lowercase hex> form")
	cmd.Flags().String(
		flagData,
		"",
		`JSON-encoded NFT data, for example {"@type":"/panacea.nft.v1.BasicNFTData","name":"Certificate"}`,
	)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseNFTData(clientCtx client.Context, dataJSON string) (*cdctypes.Any, error) {
	if dataJSON == "" {
		return nil, nil
	}

	var data types.NFTData
	if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(dataJSON), &data); err != nil {
		return nil, fmt.Errorf("invalid NFT data: %w", err)
	}

	packed, err := cdctypes.NewAnyWithValue(data)
	if err != nil {
		return nil, fmt.Errorf("pack NFT data: %w", err)
	}
	return packed, nil
}
