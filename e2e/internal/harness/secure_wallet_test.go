package harness

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSecureWalletRecoveryPlanKeepsMnemonicOutOfCommand(t *testing.T) {
	t.Parallel()

	const mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	plan, err := newSecureWalletRecoveryPlan(
		"/var/cosmos-chain/panacea-test",
		"panacead",
		"371",
		"test-wallet",
		mnemonic,
		1234,
	)
	require.NoError(t, err)
	if len(plan.Secret) != len(mnemonic)+1 || string(plan.Secret) != mnemonic+"\n" {
		t.Fatal("wallet recovery plan did not preserve the supplied secret payload")
	}
	if strings.Contains(strings.Join(plan.Command, " "), mnemonic) {
		t.Fatal("wallet recovery command contains the supplied secret")
	}
	if strings.Contains(strings.Join(plan.CleanupCommand, " "), mnemonic) {
		t.Fatal("wallet recovery cleanup command contains the supplied secret")
	}
	require.Contains(t, strings.Join(plan.Command, " "), "test-wallet")
	require.NotContains(t, plan.RelativeSecretPath, "test-wallet")
}

func TestExecuteSecureWalletRecoveryCleansUpAfterStageFailureWithFreshContext(t *testing.T) {
	t.Parallel()

	plan, err := newSecureWalletRecoveryPlan(
		"/var/cosmos-chain/panacea-test",
		"panacead",
		"371",
		"test-wallet",
		"secret test mnemonic",
		1234,
	)
	require.NoError(t, err)

	stageErr := errors.New("stage failed after copying the file")
	cleanupCalled := false
	callerCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err = executeSecureWalletRecovery(
		callerCtx,
		"test-wallet",
		plan,
		nil,
		func(context.Context, []byte, string) error { return stageErr },
		func(execCtx context.Context, command, _ []string) ([]byte, []byte, error) {
			cleanupCalled = true
			if execCtx.Err() != nil {
				t.Fatal("wallet secret cleanup inherited the canceled caller context")
			}
			deadline, ok := execCtx.Deadline()
			if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > walletSecretCleanupTimeout {
				t.Fatal("wallet secret cleanup does not have a fresh bounded deadline")
			}
			if strings.Join(command, "\x00") != strings.Join(plan.CleanupCommand, "\x00") {
				t.Fatal("wallet secret cleanup used an unexpected command")
			}
			return nil, nil, nil
		},
	)
	require.ErrorIs(t, err, stageErr)
	require.True(t, cleanupCalled)
}

func TestExecuteSecureWalletRecoveryJoinsCleanupFailure(t *testing.T) {
	t.Parallel()

	plan, err := newSecureWalletRecoveryPlan(
		"/var/cosmos-chain/panacea-test",
		"panacead",
		"371",
		"test-wallet",
		"secret test mnemonic",
		1234,
	)
	require.NoError(t, err)

	recoveryErr := errors.New("recovery exec failed before the shell started")
	cleanupErr := errors.New("cleanup exec failed")
	execCalls := 0
	err = executeSecureWalletRecovery(
		context.Background(),
		"test-wallet",
		plan,
		nil,
		func(context.Context, []byte, string) error { return nil },
		func(_ context.Context, command, _ []string) ([]byte, []byte, error) {
			execCalls++
			if strings.Join(command, "\x00") == strings.Join(plan.CleanupCommand, "\x00") {
				return nil, []byte("cleanup stderr"), cleanupErr
			}
			return nil, []byte("recovery stderr"), recoveryErr
		},
	)
	require.ErrorIs(t, err, recoveryErr)
	require.ErrorIs(t, err, cleanupErr)
	require.Equal(t, 2, execCalls)
}

func TestSecureWalletRecoveryPlanRejectsUnsafeInputs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		home     string
		binary   string
		coinType string
		keyName  string
		mnemonic string
	}{
		{name: "empty home", binary: "panacead", coinType: "371", keyName: "wallet", mnemonic: "secret"},
		{name: "unsafe key name", home: "/home/node", binary: "panacead", coinType: "371", keyName: "../wallet", mnemonic: "secret"},
		{name: "empty mnemonic", home: "/home/node", binary: "panacead", coinType: "371", keyName: "wallet"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newSecureWalletRecoveryPlan(tc.home, tc.binary, tc.coinType, tc.keyName, tc.mnemonic, 1)
			require.Error(t, err)
		})
	}
}
