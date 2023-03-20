package dep

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/spf13/cobra"
)

func HashDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash-data [data-file-path]",
		Short: "Hash a data by sha256 algorithm",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			origData, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			jsonData, err := jsoncanonicalizer.Transform(origData)
			if err != nil {
				return err
			}

			jsonDataHashBz := sha256.Sum256(jsonData)
			_, err = fmt.Fprintln(cmd.OutOrStdout(), hex.EncodeToString(jsonDataHashBz[:]))
			if err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
