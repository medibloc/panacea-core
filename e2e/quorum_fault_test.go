package e2e_test

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	quorumLocalClassID = "quorum.fault"
	quorumNFTID        = "quorum.1"
	quorumCreatorFunds = "20000000umed"
)

type quorumClassRecordResponse struct {
	ClassRecord struct {
		Class struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"class"`
		Policy struct {
			ClassID    string `json:"class_id"`
			Creator    string `json:"creator"`
			Controller string `json:"controller"`
		} `json:"policy"`
		MintedCount string `json:"minted_count"`
	} `json:"class_record"`
}

func TestFourValidatorQuorumFaultAndRecovery(t *testing.T) {
	if os.Getenv("PANACEA_E2E_CONSENSUS") != "1" {
		t.Skip("set PANACEA_E2E_CONSENSUS=1 to run the Docker consensus suite")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.CurrentImage(),
		NumValidators: 4,
		NumFullNodes:  1,
		TimeoutCommit: "1s",
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.Len(t, network.Chain.Validators, 4)
	require.Len(t, network.Chain.FullNodes, 1)
	fullNode := network.Chain.FullNodes[0]

	startupHeight := quorumNodeHeight(t, ctx, fullNode)
	startupWindow := quorumProgress(t, ctx, network, "startup-progress", fullNode, startupHeight, 3)
	allNodes := append([]*cosmos.ChainNode(nil), network.Chain.Validators...)
	allNodes = append(allNodes, fullNode)
	quorumAgreement(t, ctx, network, "startup-all-nodes", startupWindow.TargetHeight, allNodes...)
	validatorSet, err := network.ValidatorSet(ctx, fullNode, startupWindow.TargetHeight)
	require.NoError(t, err)
	require.Len(t, validatorSet, 4)
	require.Positive(t, validatorSet[0].Power)
	for _, validator := range validatorSet[1:] {
		require.Equal(t, validatorSet[0].Power, validator.Power, "validators must have equal voting power")
	}

	creator, err := network.BuildWallet(ctx, "quorum-nft-creator", "")
	require.NoError(t, err)
	receiver, err := network.BuildWallet(ctx, "quorum-nft-receiver", "")
	require.NoError(t, err)
	fundCtx, fundCancel := context.WithTimeout(ctx, time.Minute)
	_, err = network.BroadcastAndWaitTx(
		fundCtx,
		"quorum-fund-nft-creator",
		network.Chain.Validators[0],
		interchaintest.FaucetAccountKeyName,
		"bank", "send",
		interchaintest.FaucetAccountKeyName,
		creator.FormattedAddress(),
		quorumCreatorFunds,
		"--broadcast-mode", "sync",
	)
	fundCancel()
	require.NoError(t, err)

	classID := creator.FormattedAddress() + ":" + quorumLocalClassID
	created, err := network.BroadcastAndWaitTx(
		ctx,
		"quorum-nft-create-before-fault",
		network.Chain.Validators[0],
		creator.KeyName(),
		"nft", "create-class",
		quorumLocalClassID,
		"Quorum Fault Class",
		"QUORUM",
		"owner-transferable",
		"true",
		"10",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, created, "panacea.nft.v1.EventClassCreated", map[string]string{
		"class_id": classID,
		"creator":  creator.FormattedAddress(),
	})
	minted, err := network.BroadcastAndWaitTx(
		ctx,
		"quorum-nft-mint-before-fault",
		network.Chain.Validators[0],
		creator.KeyName(),
		"nft", "mint", classID, quorumNFTID, creator.FormattedAddress(),
		"--data", lifecycleDataJSON,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	assertTxEvent(t, minted, "cosmos.nft.v1beta1.EventMint", map[string]string{
		"class_id": classID,
		"id":       quorumNFTID,
		"owner":    creator.FormattedAddress(),
	})
	beforeFault := queryNFTRecord(t, ctx, network, "quorum-nft-before-fault", classID, quorumNFTID)
	require.NotNil(t, beforeFault.NFTRecord.Live)
	require.Equal(t, creator.FormattedAddress(), beforeFault.NFTRecord.Live.Owner)

	stopOneCtx, stopOneCancel := context.WithTimeout(ctx, 30*time.Second)
	err = network.StopQuorumValidator(stopOneCtx, "stop-validator-3", 3)
	stopOneCancel()
	require.NoError(t, err)
	oneFaultStart := quorumNodeHeight(t, ctx, fullNode)
	oneFaultWindow := quorumProgress(t, ctx, network, "one-validator-down-progress", fullNode, oneFaultStart, 3)
	quorumAgreement(
		t,
		ctx,
		network,
		"one-validator-down-agreement",
		oneFaultWindow.TargetHeight,
		network.Chain.Validators[0],
		network.Chain.Validators[1],
		network.Chain.Validators[2],
		fullNode,
	)

	stopTwoCtx, stopTwoCancel := context.WithTimeout(ctx, 30*time.Second)
	err = network.StopQuorumValidator(stopTwoCtx, "stop-validator-2", 2)
	stopTwoCancel()
	require.NoError(t, err)
	stallCtx, stallCancel := context.WithTimeout(ctx, 30*time.Second)
	stallWindow, err := network.ObserveQuorumStall(
		stallCtx,
		"two-validators-down-stall",
		fullNode,
		5*time.Second,
		10*time.Second,
	)
	stallCancel()
	require.NoError(t, err)
	require.Equal(t, stallWindow.StartHeight, stallWindow.EndHeight)

	noCommitCtx, noCommitCancel := context.WithTimeout(ctx, 20*time.Second)
	pending, err := network.BroadcastQuorumTxAndObserveNoCommit(
		noCommitCtx,
		8*time.Second,
		"quorum-nft-send-while-stalled",
		network.Chain.Validators[0],
		creator.KeyName(),
		"nft", "send", classID, quorumNFTID, receiver.FormattedAddress(),
	)
	noCommitCancel()
	require.NoError(t, err)
	require.NotNil(t, pending)
	require.NotEmpty(t, pending.TxHash)
	require.Zero(t, pending.CheckTx.Code)

	restartOneCtx, restartOneCancel := context.WithTimeout(ctx, 2*time.Minute)
	err = network.StartQuorumValidator(restartOneCtx, "restart-validator-2", 2)
	restartOneCancel()
	require.NoError(t, err)
	commitCtx, commitCancel := context.WithTimeout(ctx, time.Minute)
	committed, err := network.WaitForQuorumTxCommit(commitCtx, pending)
	commitCancel()
	require.NoError(t, err)
	require.Equal(t, pending.TxHash, committed.TxHash)
	require.Positive(t, committed.HeightInt64())
	sendEvent, found := committed.FindEvent("cosmos.nft.v1beta1.EventSend")
	require.True(t, found)
	require.Equal(t, classID, sendEvent.Attribute("class_id"))
	require.Equal(t, quorumNFTID, sendEvent.Attribute("id"))
	require.Equal(t, creator.FormattedAddress(), sendEvent.Attribute("sender"))
	require.Equal(t, receiver.FormattedAddress(), sendEvent.Attribute("receiver"))

	recoveredWindow := quorumProgress(
		t,
		ctx,
		network,
		"quorum-restored-progress",
		fullNode,
		stallWindow.EndHeight,
		3,
	)
	recoveredAgreementHeight := recoveredWindow.TargetHeight
	if committed.HeightInt64() > recoveredAgreementHeight {
		recoveredAgreementHeight = committed.HeightInt64()
	}
	quorumAgreement(
		t,
		ctx,
		network,
		"quorum-restored-agreement",
		recoveredAgreementHeight,
		network.Chain.Validators[0],
		network.Chain.Validators[1],
		network.Chain.Validators[2],
		fullNode,
	)
	afterRecovery := quorumQueryClass(t, ctx, network, "quorum-class-after-recovery", classID)
	require.Equal(t, classID, afterRecovery.ClassRecord.Class.ID)
	require.Equal(t, creator.FormattedAddress(), afterRecovery.ClassRecord.Policy.Creator)
	require.Equal(t, creator.FormattedAddress(), afterRecovery.ClassRecord.Policy.Controller)
	require.Equal(t, "1", afterRecovery.ClassRecord.MintedCount)
	afterRecoveryNFT := queryNFTRecord(t, ctx, network, "quorum-nft-after-recovery", classID, quorumNFTID)
	require.NotNil(t, afterRecoveryNFT.NFTRecord.Live)
	require.Equal(t, receiver.FormattedAddress(), afterRecoveryNFT.NFTRecord.Live.Owner)

	restartAllCtx, restartAllCancel := context.WithTimeout(ctx, 2*time.Minute)
	err = network.StartQuorumValidator(restartAllCtx, "restart-all-stopped-validators", 3)
	restartAllCancel()
	require.NoError(t, err)
	finalStart := quorumNodeHeight(t, ctx, fullNode)
	finalWindow := quorumProgress(t, ctx, network, "all-validators-progress", fullNode, finalStart, 3)
	quorumAgreement(t, ctx, network, "all-nodes-final-agreement", finalWindow.TargetHeight, allNodes...)
	finalClass := quorumQueryClass(t, ctx, network, "quorum-class-after-all-catch-up", classID)
	require.Equal(t, afterRecovery, finalClass)
	finalNFT := queryNFTRecord(t, ctx, network, "quorum-nft-after-all-catch-up", classID, quorumNFTID)
	require.Equal(t, afterRecoveryNFT, finalNFT)

	for attempt := 1; attempt <= 2; attempt++ {
		restart, restartErr := network.GracefulRestartNode(ctx, fullNode)
		require.NoError(t, restartErr, "full-node restart attempt %d", attempt)
		require.Greater(t, restart.After.Height, restart.Before.Height)
		for _, validator := range network.Chain.Validators {
			require.NoError(t, network.WaitForNodeHeight(ctx, validator, restart.After.Height))
		}
		quorumAgreement(
			t,
			ctx,
			network,
			"full-node-restart-agreement-"+strconv.Itoa(attempt),
			restart.After.Height,
			allNodes...,
		)
	}
}

func quorumNodeHeight(t *testing.T, parent context.Context, node *cosmos.ChainNode) int64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(parent, 15*time.Second)
	defer cancel()
	height, err := node.Height(ctx)
	require.NoError(t, err)
	require.Positive(t, height)
	return height
}

func quorumProgress(
	t *testing.T,
	parent context.Context,
	network *harness.Network,
	phase string,
	node *cosmos.ChainNode,
	startHeight int64,
	minimumBlocks int64,
) harness.QuorumHeightWindow {
	t.Helper()
	ctx, cancel := context.WithTimeout(parent, time.Minute)
	defer cancel()
	window, err := network.WaitForQuorumProgress(ctx, phase, node, startHeight, minimumBlocks)
	require.NoError(t, err)
	require.GreaterOrEqual(t, window.EndHeight, window.TargetHeight)
	return window
}

func quorumAgreement(
	t *testing.T,
	parent context.Context,
	network *harness.Network,
	phase string,
	height int64,
	nodes ...*cosmos.ChainNode,
) harness.QuorumAgreement {
	t.Helper()
	ctx, cancel := context.WithTimeout(parent, 90*time.Second)
	defer cancel()
	agreement, err := network.WaitForQuorumAgreement(ctx, phase, height, nodes...)
	require.NoError(t, err)
	require.NotEmpty(t, agreement.BlockHash)
	require.NotEmpty(t, agreement.AppHash)
	require.Len(t, agreement.Nodes, len(nodes))
	return agreement
}

func quorumQueryClass(
	t *testing.T,
	parent context.Context,
	network *harness.Network,
	step string,
	classID string,
) quorumClassRecordResponse {
	t.Helper()
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	raw, err := network.FullNodeCLIQuery(ctx, step, "nft", "class-record", classID)
	require.NoError(t, err)
	var response quorumClassRecordResponse
	require.NoError(t, json.Unmarshal(raw, &response))
	return response
}
