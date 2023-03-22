package dep

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/spf13/cobra"
)

func HashJSONCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash-json [json-file-path]",
		Short: "Hash a JSON data by sha256 algorithm",
		Long: `
This command can hash a JSON data by sha256 algorithm. The data will be encoded to canonical JSON format and printed as a hash value.`,
		Args: cobra.ExactArgs(1),
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
