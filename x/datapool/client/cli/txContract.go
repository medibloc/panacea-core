package cli

import (
	"fmt"
	"io/ioutil"

	wasmUtils "github.com/CosmWasm/wasmd/x/wasm/client/utils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	"github.com/spf13/cobra"
)

func CmdRegisterNFTContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-contract [wasm code]",
		Short: "register contract to x/datapool module",
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

			if wasmUtils.IsWasm(wasm) {
				wasm, err = wasmUtils.GzipIt(wasm)

				if err != nil {
					return err
				}
			} else if !wasmUtils.IsGzip(wasm) {
				return fmt.Errorf("invalid input file. Use wasm binary or gzip")
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
		Use:   "upgrade-contract [new wasm code]",
		Short: "upgrade contract",
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

			if wasmUtils.IsWasm(newWasmCode) {
				newWasmCode, err = wasmUtils.GzipIt(newWasmCode)

				if err != nil {
					return err
				}
			} else if !wasmUtils.IsGzip(newWasmCode) {
				return fmt.Errorf("invalid input file. Use wasm binary or gzip")
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
