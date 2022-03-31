package cli

import (
	"fmt"
	"io/ioutil"

	wasmutils "github.com/CosmWasm/wasmd/x/wasm/client/utils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/spf13/cobra"
)

func CmdRegisterNFTContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-nft-contract [wasm code]",
		Short: "register NFT contract to x/datapool module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			wasm, err := ioutil.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file")
			}

			wasm, err = gzipWasm(wasm)
			if err != nil {
				return err
			}

			msg := types.NewMsgRegisterNFTContract(wasm, clientCtx.GetFromAddress().String())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUpgradeNFTContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade-nft-contract [new wasm code]",
		Short: "upgrade NFT contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			newWasmCode, err := ioutil.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file")
			}

			newWasmCode, err = gzipWasm(newWasmCode)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpgradeNFTContract(newWasmCode, clientCtx.GetFromAddress().String())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func gzipWasm(wasm []byte) ([]byte, error) {
	if wasmutils.IsWasm(wasm) {
		wasm, err := wasmutils.GzipIt(wasm)
		if err != nil {
			return nil, err
		}
		return wasm, nil
	} else if !wasmutils.IsGzip(wasm) {
		return nil, fmt.Errorf("invalid input file. Use wasm binary or gzip")
	}

	return wasm, nil
}
