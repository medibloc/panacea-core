package e2e_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpgradeHaltMempoolSignArgumentsPreserveFutureSequenceOffline(t *testing.T) {
	t.Parallel()

	args := upgradeHaltMempoolSignArguments("/tx.json", "signer", "chain", 20, 1)
	require.True(t, slices.Contains(args, "--offline"))
	require.Equal(t, "1", args[slices.Index(args, "--sequence")+1])
	require.Equal(t, "20", args[slices.Index(args, "--account-number")+1])
}

func TestPlanUpgradeHaltMempoolReconciliation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		observedSequence uint64
		committedPrefix  int
		missingSuffix    []int
	}{
		{observedSequence: 40, committedPrefix: 0, missingSuffix: []int{0, 1}},
		{observedSequence: 41, committedPrefix: 1, missingSuffix: []int{1}},
		{observedSequence: 42, committedPrefix: 2, missingSuffix: nil},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("sequence-%d", test.observedSequence), func(t *testing.T) {
			t.Parallel()
			plan, err := planUpgradeHaltMempoolReconciliation(40, test.observedSequence, 2)
			require.NoError(t, err)
			require.Equal(t, test.committedPrefix, plan.CommittedPrefix)
			require.Equal(t, test.missingSuffix, plan.MissingSuffix)
		})
	}

	_, err := planUpgradeHaltMempoolReconciliation(40, 39, 2)
	require.ErrorContains(t, err, "outside")
	_, err = planUpgradeHaltMempoolReconciliation(40, 43, 2)
	require.ErrorContains(t, err, "outside")
	_, err = planUpgradeHaltMempoolReconciliation(40, 40, 0)
	require.ErrorContains(t, err, "positive")
}

func TestDecodeUpgradeHaltMempoolSignedBankPairRequiresSequentialSameSigner(t *testing.T) {
	t.Parallel()

	first := []byte(upgradeHaltMempoolSignedBankJSON("panacea1signer", "panacea1recipient1", 7, "AQID"))
	second := []byte(upgradeHaltMempoolSignedBankJSON("panacea1signer", "panacea1recipient2", 8, "BAUG"))
	decoded, err := decodeUpgradeHaltMempoolSignedBankPair(first, second)
	require.NoError(t, err)
	require.Equal(t, uint64(7), decoded[0].Sequence)
	require.Equal(t, uint64(8), decoded[1].Sequence)
	require.NotEqual(t, decoded[0].RecipientAddress, decoded[1].RecipientAddress)

	nonSequential := []byte(upgradeHaltMempoolSignedBankJSON("panacea1signer", "panacea1recipient2", 9, "BAUG"))
	_, err = decodeUpgradeHaltMempoolSignedBankPair(first, nonSequential)
	require.ErrorContains(t, err, "consecutive")

	differentSigner := []byte(upgradeHaltMempoolSignedBankJSON("panacea1other", "panacea1recipient2", 8, "BAUG"))
	_, err = decodeUpgradeHaltMempoolSignedBankPair(first, differentSigner)
	require.ErrorContains(t, err, "same signer")
}

func upgradeHaltMempoolSignedBankJSON(signer, recipient string, sequence uint64, signature string) string {
	return fmt.Sprintf(`{
		"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":%q,"to_address":%q,"amount":[{"denom":"umed","amount":"1000000"}]}]},
		"auth_info":{"signer_infos":[{"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":%q}],"fee":{"amount":[{"denom":"umed","amount":"2500000"}],"gas_limit":"500000"}},
		"signatures":[%q]
	}`, signer, recipient, fmt.Sprintf("%d", sequence), signature)
}
