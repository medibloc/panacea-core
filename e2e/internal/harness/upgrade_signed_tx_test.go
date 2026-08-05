package harness

import (
	"context"
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

func TestClassifyAcceptedCheckTx(t *testing.T) {
	t.Parallel()

	accepted := TxResult{Height: "0", TxHash: "ABC123", Code: 0}
	require.NoError(t, classifyAcceptedCheckTx(accepted))

	rejected := accepted
	rejected.Code = 32
	rejected.Codespace = "sdk"
	require.ErrorContains(t, classifyAcceptedCheckTx(rejected), "codespace=sdk code=32")

	committed := accepted
	committed.Height = "17"
	require.ErrorContains(t, classifyAcceptedCheckTx(committed), "committed height")

	malformedHeight := accepted
	malformedHeight.Height = "not-a-height"
	require.ErrorContains(t, classifyAcceptedCheckTx(malformedHeight), "broadcast height")
}

func TestBroadcastSignedTxFileCheckTxValidatesRequestBeforeMutation(t *testing.T) {
	t.Parallel()

	network := &Network{}
	_, err := network.BroadcastSignedTxFileCheckTx(
		context.Background(),
		"",
		nil,
		"upgrade/signed.json",
	)
	require.ErrorContains(t, err, "transaction step")

	_, err = network.BroadcastSignedTxFileCheckTx(
		context.Background(),
		"signed-checktx",
		nil,
		"upgrade/signed.json",
	)
	require.ErrorContains(t, err, "transaction node")
}

func TestBroadcastSignedTxFileExpectCheckTxFailureValidatesExpectationBeforeMutation(t *testing.T) {
	t.Parallel()

	network := &Network{}
	_, err := network.BroadcastSignedTxFileExpectCheckTxFailure(
		context.Background(),
		"signed-checktx-rejection",
		nil,
		"upgrade/signed.json",
		"",
		32,
	)
	require.ErrorContains(t, err, "codespace")

	_, err = network.BroadcastSignedTxFileExpectCheckTxFailure(
		context.Background(),
		"signed-checktx-rejection",
		nil,
		"upgrade/signed.json",
		"sdk",
		0,
	)
	require.ErrorContains(t, err, "nonzero")
}
