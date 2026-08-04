package e2e_test

import (
	"context"
	"encoding/json"
	"testing"

	upstreamnft "cosmossdk.io/x/nft"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

// captureRecoveryBoundaryStateOnNode deliberately bypasses the harness's
// first-full-node convenience methods. It proves state through the exact node
// that was dynamically added and block-synced from an empty volume.
func captureRecoveryBoundaryStateOnNode(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	node *cosmos.ChainNode,
	step string,
	classID string,
	nftID string,
	bankAddress string,
) recoveryBoundaryState {
	t.Helper()
	require.NotNil(t, node)

	classState := recoveryNodeSemanticQuery(t, ctx, node, "nft", "class-record", classID)
	nftState := recoveryNodeSemanticQuery(t, ctx, node, "nft", "nft-record", classID, nftID)
	nftClient := upstreamnft.NewQueryClient(node.GrpcConn)
	owner, err := nftClient.Owner(ctx, &upstreamnft.QueryOwnerRequest{ClassId: classID, Id: nftID})
	require.NoError(t, err)
	require.NotNil(t, owner)
	supply, err := nftClient.Supply(ctx, &upstreamnft.QuerySupplyRequest{ClassId: classID})
	require.NoError(t, err)
	require.NotNil(t, supply)
	bank, err := banktypes.NewQueryClient(node.GrpcConn).Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: bankAddress,
		Denom:   "umed",
	})
	require.NoError(t, err)
	require.NotNil(t, bank)
	require.NotNil(t, bank.Balance)

	state := recoveryBoundaryState{
		BankBalance: bank.Balance.Amount.String(),
		Class:       classState,
		NFT:         nftState,
		Owner:       owner.Owner,
		Supply:      supply.Amount,
	}
	require.NoError(t, network.WriteArtifactJSON("recovery/"+step+".json", map[string]any{
		"node":         node.Name(),
		"bank_balance": state.BankBalance,
		"class":        json.RawMessage(state.Class),
		"nft":          json.RawMessage(state.NFT),
		"owner":        state.Owner,
		"supply":       state.Supply,
	}))
	return state
}

func recoveryNodeSemanticQuery(
	t *testing.T,
	ctx context.Context,
	node *cosmos.ChainNode,
	command ...string,
) harness.SemanticJSON {
	t.Helper()
	stdout, stderr, err := node.ExecQuery(ctx, command...)
	require.NoErrorf(t, err, "query %v through %s: %s", command, node.Name(), string(stderr))
	semantic, err := harness.NewSemanticJSON(stdout)
	require.NoError(t, err)
	return semantic
}
