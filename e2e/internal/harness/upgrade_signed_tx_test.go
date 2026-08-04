package harness

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpgradeSignedTxContainerPath(t *testing.T) {
	t.Parallel()

	got, err := upgradeSignedTxContainerPath("/var/cosmos-chain/panacea", "upgrade/legacy-pnft-signed.json")
	require.NoError(t, err)
	require.Equal(t, "/var/cosmos-chain/panacea/upgrade/legacy-pnft-signed.json", got)

	for _, invalid := range []string{"", ".", "/tmp/tx.json", "../tx.json", "upgrade/../../tx.json"} {
		_, err := upgradeSignedTxContainerPath("/var/cosmos-chain/panacea", invalid)
		require.Error(t, err, invalid)
	}
}
