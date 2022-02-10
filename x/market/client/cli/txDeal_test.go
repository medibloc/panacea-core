package cli

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseSellDataFlags(t *testing.T) {
	flags := CmdSellData().Flags()
	err := flags.Set(DataVerificationCertificateFile, "./testdata/data_certificate_file.json")

	require.NoError(t, err)

	_, err = parseSellDataFlags(flags)
	require.NoError(t, err)
}
