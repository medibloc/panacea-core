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

func TestV221Compatibility(t *testing.T) {
	if os.Getenv("PANACEA_E2E_V221") != "1" {
		t.Skip("set PANACEA_E2E_V221=1 or use ./scripts/e2e/run.sh v2.2.1")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.V221Image(),
		NumValidators: 1,
		NumFullNodes:  1,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()

	startHeight, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForHeight(ctx, startHeight+3))
	require.NoError(t, network.WaitForFullNode(ctx, startHeight+3))

	users := interchaintest.GetAndFundTestUsers(t, ctx, "v221", sdkmath.NewInt(1_000_000), network.Chain)
	recipient, err := network.BuildWallet(ctx, "v221-recipient", "")
	require.NoError(t, err)

	committed, err := network.BroadcastAndWaitTx(
		ctx,
		"v221-bank-send",
		network.Chain.GetNode(),
		users[0].KeyName(),
		"bank", "send",
		users[0].KeyName(),
		recipient.FormattedAddress(),
		"2210umed",
	)
	require.NoError(t, err)
	require.NotEmpty(t, committed.TxHash)
	require.Positive(t, committed.HeightInt64())

	cliJSON, _, err := network.Chain.FullNodes[0].ExecQuery(
		ctx,
		"bank", "balances", recipient.FormattedAddress(),
	)
	require.NoError(t, err)
	var cliBalances struct {
		Balances []struct {
			Amount string `json:"amount"`
			Denom  string `json:"denom"`
		} `json:"balances"`
	}
	require.NoError(t, json.Unmarshal(cliJSON, &cliBalances))
	require.Len(t, cliBalances.Balances, 1)
	require.Equal(t, "2210", cliBalances.Balances[0].Amount)
	require.Equal(t, "umed", cliBalances.Balances[0].Denom)

	grpcBalance, err := network.QueryFullNodeBalance(ctx, recipient.FormattedAddress(), "umed")
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(2210), grpcBalance)

	rpcAddress, err := network.FullNodeHostAddress(ctx, "26657/tcp")
	require.NoError(t, err)
	restAddress, err := network.FullNodeHostAddress(ctx, "1317/tcp")
	require.NoError(t, err)
	require.NoError(t, harness.RequireRPCStatus(ctx, rpcAddress, startHeight+3))
	require.NoError(t, harness.RequireRESTNodeInfo(ctx, &http.Client{Timeout: 5 * time.Second}, restAddress))

	legacyCreator := buildAndFundNFTWallet(t, ctx, network, "v221-legacy-pnft-creator")
	legacyFixture, err := prepareV221LegacyPNFTFixture(
		ctx,
		network,
		legacyCreator,
		recipient.FormattedAddress(),
	)
	require.NoError(t, err)
	require.Equal(t, legacyCreator.FormattedAddress()+":"+legacyPNFTLocalClassID, legacyFixture.Fixture.DenomID)
	require.Equal(t, legacyPNFTSignedFilePath, legacyFixture.SignedTxPath)
}
