package e2e_test

import (
	"context"
	"os"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestIBCPreUpgradeBidirectionalTransfers(t *testing.T) {
	if os.Getenv("PANACEA_E2E_IBC_PRE_UPGRADE") != "1" {
		t.Skip("set PANACEA_E2E_IBC_PRE_UPGRADE=1 to run the pinned pre-upgrade bidirectional ICS-20 test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	topology, err := harness.StartIBCTopology(ctx, t, harness.IBCTopologyConfig{
		PanaceaImage: harness.V221Image(),
	})
	require.NoError(t, err)
	defer topology.RecordTestPanic()

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
	require.NoError(t, handshake.Validate())

	users := interchaintest.GetAndFundTestUsers(
		t,
		ctx,
		"ibc-pre-upgrade",
		sdkmath.NewInt(100_000_000),
		topology.Panacea,
		topology.Osmosis,
	)
	require.Len(t, users, 2)
	require.NoError(t, testutil.WaitForBlocks(ctx, 2, topology.Panacea, topology.Osmosis))

	evidence, err := topology.RunPreUpgradeBidirectionalTransfers(ctx, harness.IBCPreUpgradeTransferRequest{
		PanaceaUser: users[0],
		OsmosisUser: users[1],
		Amount:      sdkmath.NewInt(1_000_000),
	})
	require.NoError(t, err)
	require.NoError(t, evidence.Validate())
	require.Equal(t, handshake, evidence.Channel)
	require.Len(t, evidence.Transfers, 2)
	require.Len(t, evidence.FinalBalances, 4)
}
