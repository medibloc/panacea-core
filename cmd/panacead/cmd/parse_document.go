package cmd

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/spf13/cobra"
)

func ParseDocumentCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parse-document [document-file-path]",
		Short: "parse did document and print",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			origData, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			doc, err := ariesdid.ParseDocument(origData)
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", doc)

			return nil
		},
	}
	cmd.PersistentFlags().String(flags.FlagHome, defaultNodeHome, "The application home directory")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
