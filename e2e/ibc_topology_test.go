package e2e_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestIBCTopologySmoke(t *testing.T) {
	if os.Getenv("PANACEA_E2E_IBC_TOPOLOGY") != "1" {
		t.Skip("set PANACEA_E2E_IBC_TOPOLOGY=1 to run the pinned Osmosis/Hermes topology smoke")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	topology, err := harness.StartIBCTopology(ctx, t, harness.IBCTopologyConfig{
		PanaceaImage: harness.V221Image(),
	})
	require.NoError(t, err)
	defer topology.RecordTestPanic()
	require.Equal(t, "panacea-osmosis", topology.Path)
	require.Equal(t, harness.PinnedOsmosisImage(), topology.Osmosis.Config().Images[0])
	require.Equal(t, "31.0.2@sha256:8de930072fef03ea034b5a38f3cf93e5f47b6ccb8b1776a34e402aa47c819e0e", topology.Osmosis.Config().Images[0].Version)
	require.Equal(t, "hermes 1.8.2", topology.HermesVersion())
	require.Equal(t, "1.8.2+06dfbaf", topology.HermesIdentity())
	require.NotEmpty(t, topology.ArtifactDir())

	panaceaHeight, err := topology.Panacea.Height(ctx)
	require.NoError(t, err)
	osmosisHeight, err := topology.Osmosis.Height(ctx)
	require.NoError(t, err)
	target := panaceaHeight
	if osmosisHeight > target {
		target = osmosisHeight
	}
	require.NoError(t, topology.WaitForHeight(ctx, target+2))
}
