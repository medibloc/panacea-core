package e2e_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	recoveryLocalClassID = "recovery.class"
	recoveryNFTID        = "restart.1"
	recoveryNFTData      = `{"@type":"/panacea.nft.v1.BasicNFTData","name":"Recovery NFT","description":"restart and snapshot recovery fixture"}`
)

var recoveryNFTURIHash = "sha256:" + strings.Repeat("c", 64)

type recoveryBoundaryState struct {
	BankBalance string               `json:"bank_balance"`
	Class       harness.SemanticJSON `json:"class"`
	NFT         harness.SemanticJSON `json:"nft"`
	Owner       string               `json:"owner"`
	Supply      uint64               `json:"supply"`
}

func TestRestartRecoveryNodeBoundaries(t *testing.T) {
	if os.Getenv("PANACEA_E2E_RESTART") != "1" {
		t.Skip("set PANACEA_E2E_RESTART=1 or use ./scripts/e2e/run.sh restart")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.CurrentImage(),
		NumValidators: 1,
		NumFullNodes:  1,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.NoError(t, network.WaitForHeight(ctx, 3))
	require.NoError(t, network.WaitForFullNode(ctx, 3))

	validator := network.Chain.Validators[0]
	fullNode := network.Chain.FullNodes[0]
	creator := interchaintest.GetAndFundTestUsers(
		t,
		ctx,
		"recovery-creator",
		sdkmath.NewInt(40_000_000),
		network.Chain,
	)[0]
	owner, err := network.BuildWallet(ctx, "recovery-owner", "")
	require.NoError(t, err)
	classID := creator.FormattedAddress() + ":" + recoveryLocalClassID

	_, err = network.BroadcastAndWaitTx(
		ctx,
		"recovery-bank-state",
		validator,
		creator.KeyName(),
		"bank", "send",
		creator.KeyName(),
		owner.FormattedAddress(),
		"10000000umed",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	_, err = network.BroadcastAndWaitTx(
		ctx,
		"recovery-create-class",
		validator,
		creator.KeyName(),
		"nft", "create-class",
		recoveryLocalClassID,
		"Recovery Class",
		"RCV",
		"owner-transferable",
		"true",
		"2",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	mintTx, err := network.BroadcastAndWaitTx(
		ctx,
		"recovery-mint",
		validator,
		creator.KeyName(),
		"nft", "mint",
		classID,
		recoveryNFTID,
		owner.FormattedAddress(),
		"--uri", "https://example.test/nfts/restart.1.json",
		"--uri-hash", recoveryNFTURIHash,
		"--data", recoveryNFTData,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)

	midState := captureRecoveryBoundaryState(t, ctx, network, "mid-before-fullnode-restart", classID, recoveryNFTID, owner.FormattedAddress())
	require.Equal(t, "10000000", midState.BankBalance)
	require.Equal(t, owner.FormattedAddress(), midState.Owner)
	require.Equal(t, uint64(1), midState.Supply)
	assertRecoveryNFTPhase(t, midState.NFT, false)
	assertRecoveryMintedCount(t, midState.Class, "1")

	midHistoryHeight := mintTx.HeightInt64() + 1
	require.NoError(t, network.WaitForHeight(ctx, midHistoryHeight))
	require.NoError(t, network.WaitForFullNode(ctx, midHistoryHeight))
	beforeFullNodeRestart, err := network.CaptureRecoveryCheckpoint(ctx, "before-fullnode-graceful", fullNode, midHistoryHeight)
	require.NoError(t, err)
	_, err = network.RequireSameHistoryAtHeight(ctx, midHistoryHeight, validator, fullNode)
	require.NoError(t, err)

	require.NoError(t, network.RestartFullNodeGracefully(ctx, "fullnode-graceful-restart", 0))
	afterFullNodeSameHeight, err := network.CaptureRecoveryCheckpoint(ctx, "after-fullnode-graceful-history", fullNode, midHistoryHeight)
	require.NoError(t, err)
	afterFullNodeProgress, err := network.CaptureRecoveryCheckpoint(ctx, "after-fullnode-graceful-head", fullNode, 0)
	require.NoError(t, err)
	require.NoError(t, harness.ValidateRecoveryContinuity(beforeFullNodeRestart, afterFullNodeSameHeight, afterFullNodeProgress))
	require.Equal(
		t,
		midState,
		captureRecoveryBoundaryState(t, ctx, network, "mid-after-fullnode-restart", classID, recoveryNFTID, owner.FormattedAddress()),
	)

	burnTx, err := network.BroadcastAndWaitTx(
		ctx,
		"recovery-burn",
		validator,
		owner.KeyName(),
		"nft", "burn",
		classID,
		recoveryNFTID,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	burnedState := captureRecoveryBoundaryState(t, ctx, network, "burned-before-validator-sigkill", classID, recoveryNFTID, owner.FormattedAddress())
	require.Empty(t, burnedState.Owner)
	require.Zero(t, burnedState.Supply)
	assertRecoveryNFTPhase(t, burnedState.NFT, true)
	assertRecoveryMintedCount(t, burnedState.Class, "1")

	burnHistoryHeight := burnTx.HeightInt64() + 1
	require.NoError(t, network.WaitForHeight(ctx, burnHistoryHeight))
	require.NoError(t, network.WaitForFullNode(ctx, burnHistoryHeight))
	beforeValidatorKill, err := network.CaptureRecoveryCheckpoint(ctx, "before-validator-sigkill", validator, burnHistoryHeight)
	require.NoError(t, err)
	require.NoError(t, network.RestartValidatorAbruptly(ctx, "validator-wal-replay", 0))
	require.NoError(t, network.WaitForHeight(ctx, burnHistoryHeight+1))
	require.NoError(t, network.WaitForFullNode(ctx, burnHistoryHeight+1))
	afterValidatorSameHeight, err := network.CaptureRecoveryCheckpoint(ctx, "after-validator-sigkill-history", validator, burnHistoryHeight)
	require.NoError(t, err)
	afterValidatorProgress, err := network.CaptureRecoveryCheckpoint(ctx, "after-validator-sigkill-head", validator, 0)
	require.NoError(t, err)
	require.NoError(t, harness.ValidateRecoveryContinuity(beforeValidatorKill, afterValidatorSameHeight, afterValidatorProgress))
	require.Equal(
		t,
		burnedState,
		captureRecoveryBoundaryState(t, ctx, network, "burned-after-validator-sigkill", classID, recoveryNFTID, owner.FormattedAddress()),
	)
	require.NoError(t, network.WaitForFullNode(ctx, afterValidatorProgress.Height))
	_, err = network.RequireSameHistoryAtHeight(ctx, afterValidatorProgress.Height, validator, fullNode)
	require.NoError(t, err)

	snapshotHeight, err := fullNode.Height(ctx)
	require.NoError(t, err)
	beforeSnapshotRestore, err := network.CaptureRecoveryCheckpoint(ctx, "before-fullnode-snapshot-restore", fullNode, snapshotHeight)
	require.NoError(t, err)
	snapshot, err := network.RestoreFullNodeFromLocalSnapshot(ctx, "fullnode-local-snapshot", 0, snapshotHeight)
	require.NoError(t, err)
	require.Equal(t, uint64(snapshotHeight), snapshot.Height)
	require.Positive(t, snapshot.Format)
	require.Positive(t, snapshot.Chunks)
	require.NoError(t, network.WaitForFullNode(ctx, snapshotHeight+1))
	afterSnapshotSameHeight, err := network.CaptureRecoveryCheckpoint(ctx, "after-fullnode-snapshot-history", fullNode, snapshotHeight)
	require.NoError(t, err)
	afterSnapshotProgress, err := network.CaptureRecoveryCheckpoint(ctx, "after-fullnode-snapshot-head", fullNode, 0)
	require.NoError(t, err)
	require.NoError(t, harness.ValidateRecoveryContinuity(beforeSnapshotRestore, afterSnapshotSameHeight, afterSnapshotProgress))
	require.Equal(
		t,
		burnedState,
		captureRecoveryBoundaryState(t, ctx, network, "burned-after-snapshot-restore", classID, recoveryNFTID, owner.FormattedAddress()),
	)

	exportHeight, err := validator.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForFullNode(ctx, exportHeight))
	beforeExport, err := network.CaptureRecoveryCheckpoint(ctx, "before-deterministic-export", validator, exportHeight)
	require.NoError(t, err)
	exportEvidence, err := network.ExportValidatorGenesisDeterministically(ctx, "deterministic-genesis-export", 0, exportHeight)
	require.NoError(t, err)
	require.Equal(t, exportHeight, exportEvidence.Height)
	require.NotEmpty(t, exportEvidence.Digest)
	require.Positive(t, exportEvidence.Bytes)
	require.NoError(t, network.WaitForHeight(ctx, exportHeight+1))
	require.NoError(t, network.WaitForFullNode(ctx, exportHeight+1))
	afterExportSameHeight, err := network.CaptureRecoveryCheckpoint(ctx, "after-deterministic-export-history", validator, exportHeight)
	require.NoError(t, err)
	afterExportProgress, err := network.CaptureRecoveryCheckpoint(ctx, "after-deterministic-export-head", validator, 0)
	require.NoError(t, err)
	require.NoError(t, harness.ValidateRecoveryContinuity(beforeExport, afterExportSameHeight, afterExportProgress))
	require.Equal(
		t,
		burnedState,
		captureRecoveryBoundaryState(t, ctx, network, "burned-after-deterministic-export", classID, recoveryNFTID, owner.FormattedAddress()),
	)

	freshSync, err := network.AddAndSyncFreshFullNode(ctx, "fresh-fullnode-block-sync-nft-state")
	require.NoError(t, err)
	require.Len(t, freshSync.History, 2)
	addedFullNode := network.Chain.FullNodes[freshSync.PreviousNodeCount]
	blockSyncHeight := freshSync.TargetHeight
	require.NoError(t, network.WaitForNodeHeight(ctx, fullNode, blockSyncHeight))
	referenceBlock, err := network.CaptureRecoveryCheckpoint(ctx, "block-sync-reference", validator, blockSyncHeight)
	require.NoError(t, err)
	addedBlock, err := network.CaptureRecoveryCheckpoint(ctx, "block-sync-added-fullnode", addedFullNode, blockSyncHeight)
	require.NoError(t, err)
	require.NoError(t, harness.ValidateBlockSync(referenceBlock, addedBlock))
	_, err = network.RequireSameHistoryAtHeight(ctx, blockSyncHeight, validator, fullNode, addedFullNode)
	require.NoError(t, err)
	require.Equal(
		t,
		burnedState,
		captureRecoveryBoundaryStateOnNode(
			t,
			ctx,
			network,
			addedFullNode,
			"burned-from-added-block-sync-fullnode",
			classID,
			recoveryNFTID,
			owner.FormattedAddress(),
		),
	)

	require.NotEmpty(t, exportEvidence.Contents)
	validatorKey, err := validator.PrivValFileContent(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validatorKey)

	// Take one final export only after all source processes are stopped, and do
	// not restart them. The copied consensus key therefore has not signed at
	// the imported genesis's initial height.
	importExport, err := network.StopAndExportValidatorGenesisDeterministically(
		ctx,
		"terminal-export-for-import",
		0,
	)
	require.NoError(t, err)
	require.GreaterOrEqual(t, importExport.Height, exportEvidence.Height)
	require.NotEmpty(t, importExport.Contents)
	exportedNetwork, err := harness.StartFromExport(
		ctx,
		t,
		harness.Config{
			Image:         harness.CurrentImage(),
			NumValidators: 1,
			NumFullNodes:  1,
		},
		importExport.Contents,
		validatorKey,
	)
	require.NoError(t, err)
	defer exportedNetwork.RecordTestPanic()
	for i := range validatorKey {
		validatorKey[i] = 0
	}
	validatorKey = nil
	require.Equal(t, network.Chain.Config().ChainID, exportedNetwork.Chain.Config().ChainID)
	exportedHeight, err := exportedNetwork.Chain.Height(ctx)
	require.NoError(t, err)
	require.Greater(t, exportedHeight, importExport.Height)
	require.NoError(t, exportedNetwork.WaitForFullNode(ctx, exportedHeight))
	_, err = exportedNetwork.RequireSameHistoryAtHeight(
		ctx,
		exportedHeight,
		exportedNetwork.Chain.Validators[0],
		exportedNetwork.Chain.FullNodes[0],
	)
	require.NoError(t, err)
	exportedState := captureRecoveryBoundaryState(
		t,
		ctx,
		exportedNetwork,
		"burned-from-exported-genesis",
		classID,
		recoveryNFTID,
		owner.FormattedAddress(),
	)
	require.Equal(t, burnedState, exportedState)
	require.NoError(t, exportedNetwork.WriteArtifactJSON("recovery/export-bootstrap.json", map[string]any{
		"source_chain_id":      network.Chain.Config().ChainID,
		"destination_chain_id": exportedNetwork.Chain.Config().ChainID,
		"export_height":        importExport.Height,
		"export_digest":        importExport.Digest,
		"bootstrapped_height":  exportedHeight,
		"semantic_state_equal": true,
	}))
}

func captureRecoveryBoundaryState(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	step string,
	classID string,
	nftID string,
	bankAddress string,
) recoveryBoundaryState {
	t.Helper()
	classJSON, err := network.FullNodeCLIQuery(ctx, step+"-class", "nft", "class-record", classID)
	require.NoError(t, err)
	canonicalClass, err := harness.NewSemanticJSON(classJSON)
	require.NoError(t, err)
	nftJSON, err := network.FullNodeCLIQuery(ctx, step+"-nft", "nft", "nft-record", classID, nftID)
	require.NoError(t, err)
	canonicalNFT, err := harness.NewSemanticJSON(nftJSON)
	require.NoError(t, err)
	owner, err := network.QueryNFTOwnerGRPC(ctx, step+"-owner", classID, nftID)
	require.NoError(t, err)
	supply, err := network.QueryNFTSupplyGRPC(ctx, step+"-supply", classID)
	require.NoError(t, err)
	balance, err := network.QueryFullNodeBalance(ctx, bankAddress, "umed")
	require.NoError(t, err)
	return recoveryBoundaryState{
		BankBalance: balance.String(),
		Class:       canonicalClass,
		NFT:         canonicalNFT,
		Owner:       owner.Owner,
		Supply:      supply.Amount,
	}
}

func assertRecoveryNFTPhase(t *testing.T, contents harness.SemanticJSON, burned bool) {
	t.Helper()
	var response struct {
		NFTRecord struct {
			Live          json.RawMessage `json:"live"`
			BurnTombstone json.RawMessage `json:"burn_tombstone"`
		} `json:"nft_record"`
	}
	require.NoError(t, json.Unmarshal(contents, &response))
	if burned {
		require.True(t, isJSONNull(response.NFTRecord.Live))
		require.False(t, isJSONNull(response.NFTRecord.BurnTombstone))
		return
	}
	require.False(t, isJSONNull(response.NFTRecord.Live))
	require.True(t, isJSONNull(response.NFTRecord.BurnTombstone))
}

func assertRecoveryMintedCount(t *testing.T, contents harness.SemanticJSON, want string) {
	t.Helper()
	var response struct {
		ClassRecord struct {
			MintedCount string `json:"minted_count"`
		} `json:"class_record"`
	}
	require.NoError(t, json.Unmarshal(contents, &response))
	require.Equal(t, want, response.ClassRecord.MintedCount)
}

func isJSONNull(value json.RawMessage) bool {
	return len(value) == 0 || string(value) == "null"
}
