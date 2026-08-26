package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestIBCUpgradeContinuity(t *testing.T) {
	if os.Getenv("PANACEA_E2E_IBC_UPGRADE") != "1" {
		t.Skip("use ./scripts/e2e/run.sh ibc-upgrade to run the pinned Hermes/Osmosis upgrade continuity scenario")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
	defer cancel()
	topology, err := harness.StartIBCTopology(ctx, t, harness.IBCTopologyConfig{PanaceaImage: harness.V221Image()})
	require.NoError(t, err)
	defer topology.RecordTestPanic()
	require.Equal(t, "hermes 1.8.2", topology.HermesVersion())
	require.Equal(t, "1.8.2+06dfbaf", topology.HermesIdentity())

	panaceaHeight, err := topology.Panacea.Height(ctx)
	require.NoError(t, err)
	osmosisHeight, err := topology.Osmosis.Height(ctx)
	require.NoError(t, err)
	targetHeight := panaceaHeight
	if osmosisHeight > targetHeight {
		targetHeight = osmosisHeight
	}
	require.NoError(t, topology.WaitForHeight(ctx, targetHeight+2))

	handshake, err := topology.OpenTransferChannel(ctx)
	require.NoError(t, err)
	users := interchaintest.GetAndFundTestUsers(
		t, ctx, "ibc-upgrade-continuity", sdkmath.NewInt(250_000_000), topology.Panacea, topology.Osmosis,
	)
	require.Len(t, users, 2)
	require.NoError(t, testutil.WaitForBlocks(ctx, 2, topology.Panacea, topology.Osmosis))
	req := harness.IBCPreUpgradeTransferRequest{
		PanaceaUser: users[0],
		OsmosisUser: users[1],
		Amount:      sdkmath.NewInt(1_000_000),
	}
	preUpgrade, err := topology.RunPreUpgradeBidirectionalTransfers(ctx, req)
	require.NoError(t, err)
	require.Equal(t, handshake, preUpgrade.Channel)

	evidence, err := topology.RunUpgradeContinuity(ctx, req, runIBCPanaceaV221Upgrade)
	require.NoError(t, err)
	require.NoError(t, evidence.Validate())
	require.Equal(t, handshake, evidence.OriginalChannel)
	require.Equal(t, 1, evidence.InFlightRelay.ReceiveCount)
	require.Equal(t, 1, evidence.InFlightRelay.AcknowledgementCount)
	require.Equal(t, 1, evidence.Timeout.TimeoutCount)
	require.Zero(t, evidence.Timeout.ReceiveCount)
	require.True(t, evidence.Timeout.CommitmentCleared)
	require.Len(t, evidence.PostUpgradeTransfers.Transfers, 2)
	require.Len(t, evidence.HermesRestarts, 2)
	require.Len(t, evidence.PanaceaNodeRestarts, len(topology.Panacea.Nodes()))
	require.Len(t, evidence.OsmosisNodeRestarts, len(topology.Osmosis.Nodes()))
	require.Len(t, evidence.NodeRestartSemantics, 2)
	require.Len(t, evidence.FinalDenomTraces, 2)
	require.Len(t, evidence.DenomTraceContinuity, 3)
	require.Len(t, evidence.FinalBalances, 4)
	require.Len(t, evidence.FinalEscrowBalances, 2)
	require.NoError(t, topology.RecordUpgradeCoverageMatrix(buildIBCUpgradeCoverageMatrix(
		evidence,
		os.Getenv("PANACEA_E2E_CURRENT_BINARY_VERSION"),
	)))

	_, err = topology.OpenTransferChannel(ctx)
	require.Error(t, err, "the continuity suite must not bypass a failure by creating a replacement handshake")
}

func runIBCPanaceaV221Upgrade(ctx context.Context, network *harness.Network) (int64, error) {
	proposer, err := interchaintest.GetAndFundTestUserWithMnemonic(
		ctx, "ibc-upgrade-proposer", "", sdkmath.NewInt(100_000_000), network.Chain,
	)
	if err != nil {
		return 0, err
	}
	if err := testutil.WaitForBlocks(ctx, 2, network.Chain); err != nil {
		return 0, err
	}
	baseHeight, err := network.Chain.Height(ctx)
	if err != nil {
		return 0, err
	}
	upgradeHeight := baseHeight + 60
	proposalTx, err := network.BroadcastAndWaitTx(
		ctx,
		"ibc-upgrade-submit-proposal",
		network.Chain.Validators[0],
		proposer.KeyName(),
		"gov", "submit-legacy-proposal", "software-upgrade", upgradeName,
		"--title", "Panacea v2.3.0 IBC continuity upgrade",
		"--description", "Retain the live Panacea/Osmosis IBC channel through the binary upgrade",
		"--deposit", "1umed",
		"--upgrade-height", strconv.FormatInt(upgradeHeight, 10),
		"--upgrade-info", "{}",
		"--no-validate",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return 0, err
	}
	proposalID, err := proposalIDFromCommittedTx(proposalTx)
	if err != nil {
		return 0, err
	}
	if _, err := network.BroadcastAndWaitTx(
		ctx,
		"ibc-upgrade-vote",
		network.Chain.Validators[0],
		"validator",
		"gov", "vote", strconv.FormatUint(proposalID, 10), "yes",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	); err != nil {
		return 0, err
	}
	if err := waitForProposalPassed(ctx, network, proposalID); err != nil {
		return 0, err
	}
	halt, err := waitForOldBinaryUpgradeHalt(ctx, network, upgradeHeight)
	if writeErr := network.WriteArtifactJSON("ibc/upgrade/old-binary-halt.json", halt); writeErr != nil {
		return 0, errors.Join(err, writeErr)
	}
	if err != nil {
		return 0, err
	}
	switchCtx, switchCancel := context.WithTimeout(ctx, 3*time.Minute)
	_, err = network.SwitchNodeImagesTogether(
		switchCtx, "ibc-upgrade-all-panacea-nodes", network.Chain.Nodes(), harness.CurrentImage(),
	)
	switchCancel()
	if err != nil {
		return 0, err
	}
	postUpgradeTarget := upgradeHeight + 3
	if err := network.WaitForNodeHeight(ctx, network.Chain.Validators[0], postUpgradeTarget); err != nil {
		return 0, err
	}
	if err := network.WaitForFullNode(ctx, postUpgradeTarget); err != nil {
		return 0, err
	}
	if err := network.WriteArtifactJSON("ibc/upgrade/callback.json", map[string]any{
		"proposal_id": proposalID, "upgrade_height": upgradeHeight, "post_upgrade_target": postUpgradeTarget,
	}); err != nil {
		return 0, fmt.Errorf("record IBC upgrade callback: %w", err)
	}
	return upgradeHeight, nil
}
