package e2e_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestSmokeNodeBoundary(t *testing.T) {
	if os.Getenv("PANACEA_E2E") != "1" {
		t.Skip("set PANACEA_E2E=1 or use ./scripts/e2e/run.sh smoke")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.CurrentImage(),
		NumValidators: 1,
		NumFullNodes:  1,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()

	startHeight, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForHeight(ctx, startHeight+3))
	require.NoError(t, network.WaitForFullNode(ctx, startHeight+3))

	users := interchaintest.GetAndFundTestUsers(t, ctx, "smoke", sdkmath.NewInt(1_000_000), network.Chain)
	recipient, err := network.BuildWallet(ctx, "recipient", "")
	require.NoError(t, err)

	committed, err := network.BroadcastAndWaitTx(
		ctx,
		"bank-send",
		network.Chain.GetNode(),
		users[0].KeyName(),
		"bank", "send",
		users[0].KeyName(),
		recipient.FormattedAddress(),
		"1234umed",
	)
	require.NoError(t, err)
	require.NotEmpty(t, committed.TxHash)
	require.Positive(t, committed.HeightInt64())

	cliJSON, _, err := network.Chain.FullNodes[0].ExecQuery(
		ctx,
		"bank", "balance", recipient.FormattedAddress(), "umed",
	)
	require.NoError(t, err)
	var cliBalance struct {
		Balance struct {
			Amount string `json:"amount"`
			Denom  string `json:"denom"`
		} `json:"balance"`
	}
	require.NoError(t, json.Unmarshal(cliJSON, &cliBalance))
	require.Equal(t, "1234", cliBalance.Balance.Amount)
	require.Equal(t, "umed", cliBalance.Balance.Denom)

	grpcBalance, err := network.QueryFullNodeBalance(ctx, recipient.FormattedAddress(), "umed")
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1234), grpcBalance)

	rpcAddress, err := network.FullNodeHostAddress(ctx, "26657/tcp")
	require.NoError(t, err)
	restAddress, err := network.FullNodeHostAddress(ctx, "1317/tcp")
	require.NoError(t, err)
	require.NoError(t, harness.RequireRPCStatus(ctx, rpcAddress, startHeight+3))
	require.NoError(t, harness.RequireRESTNodeInfo(ctx, &http.Client{Timeout: 5 * time.Second}, restAddress))

	_ = runNFTLifecycle(t, ctx, network)
}
