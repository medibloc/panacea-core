package harness

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ibcInFlightTimeout               = 30 * time.Minute
	ibcPostUpgradeTimeout            = 15 * time.Second
	ibcPacketTerminalWaitMax         = int64(120)
	ibcOsmosisProgressSampleInterval = time.Second
	ibcOsmosisNoProgressBound        = 15 * time.Second
)

// IBCPanaceaUpgradeCallback upgrades only the Panacea chain and returns the
// committed software-upgrade height. The supplied Network view shares the
// topology's chain and artifact store, so callers can use the normal upgrade
// transaction and image-switch APIs without gaining access to Osmosis.
type IBCPanaceaUpgradeCallback func(context.Context, *Network) (int64, error)

// RunUpgradeContinuity executes the staged packet, caller-owned Panacea
// upgrade, relay/timeout checks, post-upgrade transfers, and restart proof.
func (n *IBCTopology) RunUpgradeContinuity(
	ctx context.Context,
	req IBCPreUpgradeTransferRequest,
	callback IBCPanaceaUpgradeCallback,
) (IBCUpgradeContinuityEvidence, error) {
	var zero IBCUpgradeContinuityEvidence
	if _, err := n.StageUpgradeInFlightPacket(ctx, req); err != nil {
		return zero, err
	}
	if _, err := n.RunPanaceaUpgradeStep(ctx, callback); err != nil {
		return zero, err
	}
	return n.ResumeAndVerifyUpgradeContinuity(ctx, req)
}

// ResumeAndVerifyUpgradeContinuity is implemented below in the same runtime
// module; keeping the public staged API separate lets a test own the upgrade
// transaction while the harness owns all IBC assertions.
func (n *IBCTopology) ResumeAndVerifyUpgradeContinuity(
	ctx context.Context,
	req IBCPreUpgradeTransferRequest,
) (IBCUpgradeContinuityEvidence, error) {
	return n.resumeAndVerifyUpgradeContinuity(ctx, req)
}

func (n *IBCTopology) resumeAndVerifyUpgradeContinuity(
	ctx context.Context,
	req IBCPreUpgradeTransferRequest,
) (IBCUpgradeContinuityEvidence, error) {
	var zero IBCUpgradeContinuityEvidence
	if ctx == nil {
		return zero, errors.New("IBC upgrade continuity context is required")
	}
	if n == nil {
		return zero, errors.New("IBC topology is required")
	}
	n.lifecycleMu.Lock()
	defer n.lifecycleMu.Unlock()
	if err := n.validateTransferRuntime(); err != nil {
		return zero, err
	}
	if err := validateIBCTransferRequest(req); err != nil {
		return zero, err
	}
	if n.channel == nil || n.inFlightCheckpoint == nil || n.inFlightTx == nil ||
		n.panaceaUpgradeStep == nil || n.postUpgradeBeforeRelay == nil {
		return zero, errors.New("stage the packet and complete the Panacea upgrade before resuming IBC continuity")
	}
	if n.upgradeContinuityComplete {
		return zero, errors.New("IBC upgrade continuity is already complete")
	}
	checkpoint := *n.inFlightCheckpoint
	if req.PanaceaUser.FormattedAddress() != checkpoint.SourceNativeBalance.Address ||
		req.OsmosisUser.FormattedAddress() != checkpoint.DestinationVoucherBalance.Address ||
		req.Amount.String() != checkpoint.Packet.Amount {
		return zero, errors.New("IBC continuity request does not match the staged packet users and amount")
	}

	preRelayBalances, err := n.captureInFlightPreRelayBalances(ctx, req)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-pre-relay-balances", err)
	}
	firstRestart, err := n.startHermes(ctx, "post-upgrade-relay", "health-check-post-upgrade-relay")
	if err != nil {
		return zero, err
	}
	relay, afterRelay, err := n.observeInFlightRelay(ctx, req, preRelayBalances)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-relay", err)
	}
	if err := n.artifacts.base.writeJSON("ibc/upgrade/in-flight-relay.json", relay); err != nil {
		return zero, err
	}
	timeoutEvidence, err := n.runPostUpgradeTimeout(ctx, req, afterRelay)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-timeout-refund", err)
	}
	postTransfers, err := n.runPostUpgradeBidirectionalTransfers(ctx, req)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-post-bidirectional", err)
	}

	if err := n.stopHermes(ctx, "post-transfer-restart"); err != nil {
		return zero, err
	}
	secondRestart, err := n.startHermes(ctx, "post-transfer-restart", "health-check-post-transfer-restart")
	if err != nil {
		return zero, err
	}
	panaceaBeforeRestart, panaceaRestarts, osmosisBeforeRestart, osmosisRestarts, err := n.restartEveryIBCNode(ctx)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-all-node-restarts", err)
	}
	finalState, err := n.queryIBCLinkState(ctx, "post-upgrade-after-node-restart")
	if err != nil {
		return zero, n.ibcOperationError("upgrade-final-link-state", err)
	}
	finalPackets, err := n.queryFinalPacketStates(ctx, relay, timeoutEvidence)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-final-packet-states", err)
	}
	finalTraces, err := n.queryFinalDenomTraces(ctx)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-final-denom-traces", err)
	}
	finalBalances, err := n.queryFinalRestartBalances(ctx, postTransfers.FinalBalances)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-final-balances", err)
	}
	finalEscrows, err := n.queryFinalRestartEscrows(ctx, postTransfers.EscrowBalances)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-final-escrows", err)
	}
	nodeRestartSemantics, err := n.queryPostRestartNodeSemantics(
		ctx,
		panaceaBeforeRestart, panaceaRestarts,
		osmosisBeforeRestart, osmosisRestarts,
		finalState, relay, timeoutEvidence,
	)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-post-restart-node-semantics", err)
	}
	if n.preUpgradeTransferEvidence == nil {
		return zero, n.ibcOperationError("upgrade-denom-trace-continuity", errors.New("pre-upgrade transfer evidence is missing"))
	}
	traceContinuity := []IBCDenomTraceSnapshot{
		n.preUpgradeTransferEvidence.DenomTraces,
		postTransfers.DenomTraces,
		{Phase: "post-restart", Traces: finalTraces},
	}

	panaceaHeight, osmosisHeight, err := n.chainHeights(ctx)
	if err != nil {
		return zero, err
	}
	panaceaHeight = stableIBCScanEndHeight(panaceaHeight)
	osmosisHeight = stableIBCScanEndHeight(osmosisHeight)
	relay.ReceiveCount, err = countSuccessfulRecvPackets(
		ctx, n.Osmosis, checkpoint.DestinationScanStartHeight, osmosisHeight, n.inFlightTx.Packet,
	)
	if err != nil {
		return zero, err
	}
	relay.AcknowledgementCount, err = countSuccessfulAcknowledgements(
		ctx, n.Panacea, n.inFlightTx.Height, panaceaHeight, n.inFlightTx.Packet,
	)
	if err != nil {
		return zero, err
	}
	timeoutPacket, err := packetFromSendEvidence(timeoutEvidence.Packet)
	if err != nil {
		return zero, err
	}
	timeoutEvidence.TimeoutCount, err = countSuccessfulTimeouts(ctx, n.Panacea, timeoutEvidence.Packet.TxHeight, panaceaHeight, timeoutPacket)
	if err != nil {
		return zero, err
	}
	timeoutEvidence.ReceiveCount, err = countSuccessfulRecvPackets(ctx, n.Osmosis, 1, osmosisHeight, timeoutPacket)
	if err != nil {
		return zero, err
	}

	evidence := IBCUpgradeContinuityEvidence{
		Phase:                   ibcUpgradeContinuityPhase,
		OriginalChannel:         *n.channel,
		InFlight:                checkpoint,
		PanaceaUpgrade:          *n.panaceaUpgradeStep,
		PostUpgradeBeforeRelay:  *n.postUpgradeBeforeRelay,
		InFlightRelay:           relay,
		AfterInFlightRelay:      afterRelay,
		Timeout:                 timeoutEvidence,
		PostUpgradeTransfers:    postTransfers,
		FinalAfterHermesRestart: finalState,
		HermesRestarts:          []IBCHermesRestartEvidence{firstRestart, secondRestart},
		PanaceaNodeRestarts:     panaceaRestarts,
		OsmosisNodeRestarts:     osmosisRestarts,
		NodeRestartSemantics:    nodeRestartSemantics,
		FinalPacketStates:       finalPackets,
		FinalDenomTraces:        finalTraces,
		DenomTraceContinuity:    traceContinuity,
		FinalBalances:           finalBalances,
		FinalEscrowBalances:     finalEscrows,
	}
	if err := evidence.Validate(); err != nil {
		return zero, n.ibcOperationError("upgrade-continuity-evidence-validate", err)
	}
	if err := n.recordIBCUpgradeGRPCQueryRecords(evidence); err != nil {
		return zero, n.ibcOperationError("upgrade-continuity-grpc-query-evidence", err)
	}
	if err := n.artifacts.base.writeJSON("ibc/upgrade/continuity.json", evidence); err != nil {
		return zero, n.ibcOperationError("upgrade-continuity-evidence-artifact", err)
	}
	n.upgradeContinuityComplete = true
	return evidence, nil
}

// StageUpgradeInFlightPacket stops Hermes and commits one long-timeout
// Panacea->Osmosis packet, proving its source commitment exists while receipt
// and acknowledgement are absent.
func (n *IBCTopology) StageUpgradeInFlightPacket(
	ctx context.Context,
	req IBCPreUpgradeTransferRequest,
) (IBCInFlightPacketCheckpoint, error) {
	var zero IBCInFlightPacketCheckpoint
	if ctx == nil {
		return zero, errors.New("IBC upgrade continuity context is required")
	}
	if n == nil {
		return zero, errors.New("IBC topology is required")
	}

	n.lifecycleMu.Lock()
	defer n.lifecycleMu.Unlock()
	if err := n.validateTransferRuntime(); err != nil {
		return zero, err
	}
	if n.channel == nil || !n.preUpgradeTransferComplete {
		return zero, errors.New("complete the pre-upgrade bidirectional transfers before staging an upgrade packet")
	}
	if n.upgradeContinuityAttempted {
		return zero, errors.New("IBC upgrade continuity was already attempted on this topology")
	}
	if !n.relayerStarted {
		return zero, errors.New("Hermes must be running before the in-flight staging boundary")
	}
	if err := validateIBCTransferRequest(req); err != nil {
		return zero, err
	}
	n.upgradeContinuityAttempted = true

	handshake := *n.channel
	beforeState, err := n.queryIBCLinkState(ctx, "before-in-flight")
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-state-before", err)
	}
	panaceaAddress := req.PanaceaUser.FormattedAddress()
	osmosisAddress := req.OsmosisUser.FormattedAddress()
	panaceaDenom := n.Panacea.Config().Denom
	medOnOsmosis := ibcVoucherDenom(handshake.Osmosis, panaceaDenom)
	sourceBefore, err := n.Panacea.GetBalance(ctx, panaceaAddress, panaceaDenom)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-source-balance-before", err)
	}
	destinationBefore, err := n.Osmosis.GetBalance(ctx, osmosisAddress, medOnOsmosis)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-destination-balance-before", err)
	}
	escrowAddress, escrowBefore, err := queryIBCEscrowBalance(ctx, n.Panacea, handshake.Panacea, panaceaDenom)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-escrow-before", err)
	}
	destinationStartHeight, err := n.Osmosis.Height(ctx)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-destination-height", err)
	}

	if err := n.stopHermes(ctx, "stage-in-flight"); err != nil {
		return zero, err
	}
	timeoutOptions, err := relativeIBCTimestampTransferOptions(ibcInFlightTimeout)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-timeout-options", err)
	}
	tx, err := n.Panacea.SendIBCTransfer(
		ctx,
		handshake.Panacea.ChannelID,
		req.PanaceaUser.KeyName(),
		ibc.WalletAmount{
			Address: osmosisAddress,
			Denom:   panaceaDenom,
			Amount:  req.Amount,
		},
		timeoutOptions,
	)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-send", err)
	}
	if err := validateIBCTransferTx(tx, handshake.Panacea, handshake.Osmosis); err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-send-validate", err)
	}
	if _, err := committedPacketTimeoutTimestamp(tx); err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-send-deadline", err)
	}
	fee := n.Panacea.GetGasFeesInNativeDenom(tx.GasSpent)
	send := packetSendEvidenceFromTx(tx, ibcPanaceaToOsmosis, handshake.Panacea, handshake.Osmosis, panaceaDenom, req.Amount, fee)
	if err := n.recordIBCPacketStage("in-flight-staged", "send", packetLifecycleFromSend(send)); err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-send-artifact", err)
	}

	if err := testutil.WaitForBlocks(ctx, 2, n.Panacea, n.Osmosis); err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-no-relay-wait", err)
	}
	afterState, err := n.queryIBCLinkState(ctx, "in-flight-staged")
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-state-after", err)
	}
	commitment, exists, err := queryPacketCommitment(ctx, n.Panacea, handshake.Panacea, tx.Packet.Sequence)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-commitment", err)
	}
	if !exists || len(commitment) == 0 {
		return zero, n.ibcOperationError("upgrade-in-flight-commitment", errors.New("staged packet has no source commitment"))
	}
	receipt, err := queryPacketReceipt(ctx, n.Osmosis, handshake.Osmosis, tx.Packet.Sequence)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-receipt", err)
	}
	acknowledgement, ackExists, err := queryPacketAcknowledgement(ctx, n.Osmosis, handshake.Osmosis, tx.Packet.Sequence)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-acknowledgement", err)
	}
	if receipt || ackExists {
		return zero, n.ibcOperationError("upgrade-in-flight-terminal-state", errors.New("staged packet was relayed while Hermes was stopped"))
	}

	sourceAfter, err := n.Panacea.GetBalance(ctx, panaceaAddress, panaceaDenom)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-source-balance-after", err)
	}
	destinationAfter, err := n.Osmosis.GetBalance(ctx, osmosisAddress, medOnOsmosis)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-destination-balance-after", err)
	}
	_, escrowAfter, err := queryIBCEscrowBalance(ctx, n.Panacea, handshake.Panacea, panaceaDenom)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-escrow-after", err)
	}
	expectedSource := sourceBefore.Sub(req.Amount).SubRaw(fee)
	checkpoint := IBCInFlightPacketCheckpoint{
		Phase:                      "in-flight-staged",
		Channel:                    handshake,
		BeforeSendState:            beforeState,
		AfterSendState:             afterState,
		Packet:                     send,
		Commitment:                 base64.StdEncoding.EncodeToString(commitment),
		DestinationReceipt:         receipt,
		DestinationAcknowledgement: base64IfNotEmpty(acknowledgement),
		DestinationScanStartHeight: destinationStartHeight,
		SourceNativeBalance: newIBCBalanceEvidence(
			handshake.Panacea.ChainID, panaceaAddress, panaceaDenom, sourceBefore, sourceAfter, expectedSource,
		),
		DestinationVoucherBalance: newIBCBalanceEvidence(
			handshake.Osmosis.ChainID, osmosisAddress, medOnOsmosis, destinationBefore, destinationAfter, destinationBefore,
		),
		SourceEscrowLock: newIBCEscrowBalanceEvidence(
			"in-flight-staged", handshake.Panacea, escrowAddress, panaceaDenom, escrowBefore, escrowAfter, req.Amount,
		),
	}
	if err := checkpoint.Validate(); err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-evidence-validate", err)
	}
	if err := n.artifacts.base.writeJSON("ibc/upgrade/in-flight-staged.json", checkpoint); err != nil {
		return zero, n.ibcOperationError("upgrade-in-flight-evidence-artifact", err)
	}
	n.inFlightTx = &tx
	n.inFlightCheckpoint = &checkpoint
	return checkpoint, nil
}

// RunPanaceaUpgradeStep brackets a caller-supplied v2.2.1->current upgrade,
// proves every Panacea container changed while Osmosis did not, explicitly
// updates both retained clients, and re-queries the original link before relay.
func (n *IBCTopology) RunPanaceaUpgradeStep(
	ctx context.Context,
	callback IBCPanaceaUpgradeCallback,
) (IBCPanaceaUpgradeStepEvidence, error) {
	var zero IBCPanaceaUpgradeStepEvidence
	if ctx == nil {
		return zero, errors.New("IBC Panacea upgrade context is required")
	}
	if n == nil {
		return zero, errors.New("IBC topology is required")
	}
	if callback == nil {
		return zero, errors.New("Panacea upgrade callback is required")
	}

	n.lifecycleMu.Lock()
	defer n.lifecycleMu.Unlock()
	if err := n.validateTransferRuntime(); err != nil {
		return zero, err
	}
	if n.inFlightCheckpoint == nil || n.inFlightTx == nil {
		return zero, errors.New("stage the in-flight packet before upgrading Panacea")
	}
	if n.panaceaUpgradeStep != nil {
		return zero, errors.New("Panacea upgrade callback was already attempted")
	}
	if n.relayerStarted {
		return zero, errors.New("Hermes must remain stopped throughout the Panacea upgrade")
	}
	currentBinaryContract, err := currentPanaceaBinaryContract()
	if err != nil {
		return zero, n.ibcOperationError("upgrade-current-binary-contract", err)
	}

	panaceaBefore, err := captureIBCNodeRuntimeIdentities(n.Panacea)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-panacea-identities-before", err)
	}
	osmosisBefore, err := captureIBCNodeRuntimeIdentities(n.Osmosis)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-osmosis-identities-before", err)
	}
	beforeHeight, err := n.Panacea.Height(ctx)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-panacea-height-before", err)
	}
	osmosisProgressMonitor, err := startIBCHeightProgressMonitor(ctx, n.Osmosis)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-osmosis-progress-start", err)
	}

	upgradeHeight, callbackErr := callback(ctx, n.panaceaNetworkView())
	osmosisProgress, osmosisProgressErr := osmosisProgressMonitor.stop(ctx)
	afterHeight, heightErr := n.Panacea.Height(ctx)
	panaceaAfter, panaceaIdentityErr := captureIBCNodeRuntimeIdentities(n.Panacea)
	osmosisAfter, osmosisIdentityErr := captureIBCNodeRuntimeIdentities(n.Osmosis)
	panaceaPostBinary, panaceaPostBinaryErr := captureIBCChainBinaryEvidence(
		ctx,
		n.Panacea,
		"post-upgrade",
		currentBinaryContract,
		n.artifacts.base,
		"ibc/chains/panacea/post-upgrade",
	)
	osmosisPostBinary, osmosisPostBinaryErr := captureIBCChainBinaryEvidence(
		ctx,
		n.Osmosis,
		"post-upgrade",
		pinnedOsmosisBinaryContract(),
		n.artifacts.base,
		"ibc/chains/osmosis/post-upgrade",
	)
	osmosisGenesisPost, osmosisGenesisPostErr := captureIBCGenesisChecksumSnapshot(
		ctx,
		n.Osmosis,
		"post-upgrade",
		n.artifacts.base,
		"ibc/chains/osmosis",
		n.osmosisGenesisInitial.Common,
	)
	osmosisGenesisEvidence := IBCGenesisImmutabilityEvidence{
		ChainID:     n.Osmosis.Config().ChainID,
		File:        "config/genesis.json",
		Initial:     n.osmosisGenesisInitial,
		PostUpgrade: osmosisGenesisPost,
		Immutable:   osmosisGenesisPostErr == nil && n.osmosisGenesisInitial.Common == osmosisGenesisPost.Common,
	}
	osmosisGenesisValidationErr := osmosisGenesisEvidence.Validate()
	osmosisGenesisArtifactErr := n.artifacts.base.writeJSON(
		"ibc/chains/osmosis/genesis-checksums.json",
		osmosisGenesisEvidence,
	)
	n.panaceaPostUpgradeBinary = panaceaPostBinary
	n.osmosisPostUpgradeBinary = osmosisPostBinary
	n.osmosisGenesisImmutability = osmosisGenesisEvidence
	evidence := IBCPanaceaUpgradeStepEvidence{
		CallbackCompleted: callbackErr == nil,
		UpgradeHeight:     upgradeHeight,
		BeforeHeight:      beforeHeight,
		AfterHeight:       afterHeight,
		From:              V221Image(),
		To:                CurrentImage(),
		PanaceaBefore:     panaceaBefore,
		PanaceaAfter:      panaceaAfter,
		OsmosisBefore:     osmosisBefore,
		OsmosisAfter:      osmosisAfter,
		OsmosisProgress:   osmosisProgress,
	}
	osmosisProgressArtifactErr := n.artifacts.base.writeJSON("ibc/upgrade/osmosis-height-progress.json", osmosisProgress)
	joined := errors.Join(
		callbackErr,
		osmosisProgressErr,
		heightErr,
		panaceaIdentityErr,
		osmosisIdentityErr,
		panaceaPostBinaryErr,
		osmosisPostBinaryErr,
		osmosisGenesisPostErr,
		osmosisGenesisValidationErr,
		osmosisGenesisArtifactErr,
		osmosisProgressArtifactErr,
	)
	if joined != nil {
		evidence.Error = joined.Error()
	}
	n.panaceaUpgradeStep = &evidence
	if err := n.artifacts.base.writeJSON("ibc/upgrade/panacea-step.json", evidence); err != nil {
		joined = errors.Join(joined, err)
	}
	if joined != nil {
		return evidence, n.ibcOperationError("upgrade-panacea-callback", joined)
	}
	if err := evidence.Validate(); err != nil {
		return evidence, n.ibcOperationError("upgrade-panacea-evidence-validate", err)
	}
	if err := n.Relayer.UpdateClients(ctx, n.execReporter, n.Path); err != nil {
		return evidence, n.ibcOperationError("upgrade-hermes-client-update", err)
	}
	if _, err := n.artifacts.execHermesEvidence(
		ctx, n.hermes, n.execReporter, "health-check-post-upgrade-before-relay", []string{"hermes", "health-check"},
	); err != nil {
		return evidence, n.ibcOperationError("upgrade-hermes-health-check-before-relay", err)
	}

	state, err := n.queryIBCLinkState(ctx, "post-upgrade-before-relay")
	if err != nil {
		return evidence, n.ibcOperationError("upgrade-link-state-before-relay", err)
	}
	commitment, exists, err := queryPacketCommitment(ctx, n.Panacea, n.channel.Panacea, n.inFlightTx.Packet.Sequence)
	if err != nil || !exists || len(commitment) == 0 {
		return evidence, n.ibcOperationError("upgrade-in-flight-commitment-after-upgrade", errors.Join(err, errors.New("in-flight commitment did not survive the Panacea upgrade")))
	}
	receipt, err := queryPacketReceipt(ctx, n.Osmosis, n.channel.Osmosis, n.inFlightTx.Packet.Sequence)
	if err != nil {
		return evidence, n.ibcOperationError("upgrade-in-flight-receipt-after-upgrade", err)
	}
	_, ackExists, err := queryPacketAcknowledgement(ctx, n.Osmosis, n.channel.Osmosis, n.inFlightTx.Packet.Sequence)
	if err != nil {
		return evidence, n.ibcOperationError("upgrade-in-flight-ack-after-upgrade", err)
	}
	if receipt || ackExists {
		return evidence, n.ibcOperationError("upgrade-in-flight-terminal-after-upgrade", errors.New("in-flight packet was relayed before Hermes restarted"))
	}
	if err := n.artifacts.base.writeJSON("ibc/upgrade/post-upgrade-before-relay.json", state); err != nil {
		return evidence, n.ibcOperationError("upgrade-link-state-before-relay-artifact", err)
	}
	if err := n.recordIBCCompatibilityMatrix(ctx); err != nil {
		return evidence, n.ibcOperationError("upgrade-compatibility-matrix", err)
	}
	n.postUpgradeBeforeRelay = &state
	return evidence, nil
}

func validateIBCTransferRequest(req IBCPreUpgradeTransferRequest) error {
	if req.PanaceaUser == nil || req.OsmosisUser == nil {
		return errors.New("one funded IBC user per chain is required")
	}
	if req.Amount.IsNil() || !req.Amount.IsPositive() {
		return errors.New("IBC transfer amount must be positive")
	}
	if strings.TrimSpace(req.PanaceaUser.FormattedAddress()) == "" || strings.TrimSpace(req.OsmosisUser.FormattedAddress()) == "" {
		return errors.New("IBC wallet formatted addresses are required")
	}
	return nil
}

func validateIBCTransferTx(tx ibc.Tx, source, destination IBCChannelEndpoint) error {
	if err := tx.Validate(); err != nil {
		return err
	}
	if tx.Packet.SourcePort != source.PortID || tx.Packet.SourceChannel != source.ChannelID ||
		tx.Packet.DestPort != destination.PortID || tx.Packet.DestChannel != destination.ChannelID {
		return fmt.Errorf(
			"packet endpoint = %s/%s -> %s/%s, want %s/%s -> %s/%s",
			tx.Packet.SourcePort, tx.Packet.SourceChannel, tx.Packet.DestPort, tx.Packet.DestChannel,
			source.PortID, source.ChannelID, destination.PortID, destination.ChannelID,
		)
	}
	return nil
}

func packetSendEvidenceFromTx(
	tx ibc.Tx,
	direction string,
	source IBCChannelEndpoint,
	destination IBCChannelEndpoint,
	denom string,
	amount sdkmath.Int,
	fee int64,
) IBCPacketSendEvidence {
	return IBCPacketSendEvidence{
		Direction:          direction,
		SourceChainID:      source.ChainID,
		DestinationChainID: destination.ChainID,
		TxHash:             tx.TxHash,
		TxHeight:           tx.Height,
		Sequence:           tx.Packet.Sequence,
		SourcePort:         tx.Packet.SourcePort,
		SourceChannel:      tx.Packet.SourceChannel,
		DestinationPort:    tx.Packet.DestPort,
		DestinationChannel: tx.Packet.DestChannel,
		Denom:              denom,
		Amount:             amount.String(),
		GasFee:             fmt.Sprintf("%d", fee),
		PacketData:         base64.StdEncoding.EncodeToString(tx.Packet.Data),
		TimeoutHeight:      tx.Packet.TimeoutHeight,
		TimeoutTimestamp:   uint64(tx.Packet.TimeoutTimestamp),
	}
}

func packetLifecycleFromSend(send IBCPacketSendEvidence) IBCPacketLifecycleEvidence {
	return IBCPacketLifecycleEvidence{
		Direction:          send.Direction,
		SourceChainID:      send.SourceChainID,
		DestinationChainID: send.DestinationChainID,
		TxHash:             send.TxHash,
		TxHeight:           send.TxHeight,
		Sequence:           send.Sequence,
		SourcePort:         send.SourcePort,
		SourceChannel:      send.SourceChannel,
		DestinationPort:    send.DestinationPort,
		DestinationChannel: send.DestinationChannel,
		Denom:              send.Denom,
		Amount:             send.Amount,
		PacketData:         send.PacketData,
		TimeoutHeight:      send.TimeoutHeight,
		TimeoutTimestamp:   send.TimeoutTimestamp,
	}
}

func newIBCBalanceEvidence(chainID, address, denom string, before, after, expected sdkmath.Int) IBCBalanceEvidence {
	return IBCBalanceEvidence{
		ChainID:       chainID,
		Address:       address,
		Denom:         denom,
		Before:        before.String(),
		After:         after.String(),
		ExpectedAfter: expected.String(),
	}
}

func newIBCEscrowBalanceEvidence(
	phase string,
	endpoint IBCChannelEndpoint,
	address string,
	denom string,
	before sdkmath.Int,
	after sdkmath.Int,
	delta sdkmath.Int,
) IBCEscrowBalanceEvidence {
	return IBCEscrowBalanceEvidence{
		Phase:         phase,
		ChainID:       endpoint.ChainID,
		PortID:        endpoint.PortID,
		ChannelID:     endpoint.ChannelID,
		Address:       address,
		Denom:         denom,
		Before:        before.String(),
		After:         after.String(),
		ExpectedDelta: delta.String(),
		ExpectedAfter: before.Add(delta).String(),
	}
}

func queryIBCEscrowBalance(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	endpoint IBCChannelEndpoint,
	denom string,
) (string, sdkmath.Int, error) {
	if chain == nil || endpoint.ChainID != chain.Config().ChainID || strings.TrimSpace(denom) == "" {
		return "", sdkmath.Int{}, errors.New("IBC escrow query chain, endpoint, or denomination is invalid")
	}
	address, err := sdk.Bech32ifyAddressBytes(chain.Config().Bech32Prefix, transfertypes.GetEscrowAddress(endpoint.PortID, endpoint.ChannelID))
	if err != nil {
		return "", sdkmath.Int{}, fmt.Errorf("encode IBC escrow address: %w", err)
	}
	balance, err := chain.GetBalance(ctx, address, denom)
	if err != nil {
		return "", sdkmath.Int{}, fmt.Errorf("query IBC escrow %s balance: %w", address, err)
	}
	return address, balance, nil
}

type ibcHeightProgressMonitor struct {
	chainID string
	stopCh  chan chan []IBCHeightSample
}

func startIBCHeightProgressMonitor(ctx context.Context, chain *cosmos.CosmosChain) (*ibcHeightProgressMonitor, error) {
	if chain == nil {
		return nil, errors.New("IBC height progress chain is required")
	}
	startedAt := time.Now().UTC()
	height, err := chain.Height(ctx)
	if err != nil {
		return nil, fmt.Errorf("query initial IBC height progress: %w", err)
	}
	monitor := &ibcHeightProgressMonitor{
		chainID: chain.Config().ChainID,
		stopCh:  make(chan chan []IBCHeightSample),
	}
	go func() {
		samples := []IBCHeightSample{{ObservedAt: startedAt, Height: height}}
		ticker := time.NewTicker(ibcOsmosisProgressSampleInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sampleHeight, sampleErr := chain.Height(ctx)
				sample := IBCHeightSample{ObservedAt: time.Now().UTC(), Height: sampleHeight}
				if sampleErr != nil {
					sample.Error = sampleErr.Error()
				}
				samples = append(samples, sample)
			case response := <-monitor.stopCh:
				sampleHeight, sampleErr := chain.Height(ctx)
				sample := IBCHeightSample{ObservedAt: time.Now().UTC(), Height: sampleHeight}
				if sampleErr != nil {
					sample.Error = sampleErr.Error()
				}
				samples = append(samples, sample)
				response <- samples
				return
			}
		}
	}()
	return monitor, nil
}

func (m *ibcHeightProgressMonitor) stop(ctx context.Context) (IBCHeightProgressEvidence, error) {
	var zero IBCHeightProgressEvidence
	if m == nil {
		return zero, errors.New("IBC height progress monitor is required")
	}
	response := make(chan []IBCHeightSample, 1)
	select {
	case m.stopCh <- response:
	case <-ctx.Done():
		return zero, ctx.Err()
	}
	var samples []IBCHeightSample
	select {
	case samples = <-response:
	case <-ctx.Done():
		return zero, ctx.Err()
	}
	if len(samples) < 2 {
		return zero, errors.New("IBC height progress monitor returned too few samples")
	}
	maxNoProgress := time.Duration(0)
	lastAdvanceAt := samples[0].ObservedAt
	lastHeight := samples[0].Height
	for _, sample := range samples[1:] {
		stalled := sample.ObservedAt.Sub(lastAdvanceAt)
		if stalled > maxNoProgress {
			maxNoProgress = stalled
		}
		if sample.Height > lastHeight {
			lastHeight = sample.Height
			lastAdvanceAt = sample.ObservedAt
			continue
		}
	}
	evidence := IBCHeightProgressEvidence{
		ChainID:             m.chainID,
		StartedAt:           samples[0].ObservedAt,
		CompletedAt:         samples[len(samples)-1].ObservedAt,
		StartHeight:         samples[0].Height,
		EndHeight:           samples[len(samples)-1].Height,
		MaxNoProgressMillis: maxNoProgress.Milliseconds(),
		BoundMillis:         ibcOsmosisNoProgressBound.Milliseconds(),
		Samples:             samples,
	}
	if err := evidence.validate(); err != nil {
		return evidence, err
	}
	return evidence, nil
}

func base64IfNotEmpty(value []byte) string {
	if len(value) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(value)
}

func captureIBCNodeRuntimeIdentities(chain *cosmos.CosmosChain) ([]IBCNodeRuntimeIdentity, error) {
	if chain == nil {
		return nil, errors.New("IBC chain is required")
	}
	nodes := chain.Nodes()
	if len(nodes) == 0 {
		return nil, errors.New("IBC chain has no nodes")
	}
	identities := make([]IBCNodeRuntimeIdentity, 0, len(nodes))
	for index, node := range nodes {
		if node == nil {
			return nil, fmt.Errorf("IBC chain node %d is nil", index)
		}
		identity := IBCNodeRuntimeIdentity{
			Name:        node.Name(),
			ContainerID: node.ContainerID(),
			Image:       imageRefFromDocker(node.Image),
		}
		if _, err := nodeIdentityMap([]IBCNodeRuntimeIdentity{identity}); err != nil {
			return nil, err
		}
		identities = append(identities, identity)
	}
	return identities, nil
}

func (n *IBCTopology) panaceaNetworkView() *Network {
	return &Network{Chain: n.Panacea, artifacts: n.artifacts.base}
}

func (n *IBCTopology) stopHermes(ctx context.Context, phase string) error {
	beforePanacea, beforeOsmosis, err := n.chainHeights(ctx)
	if err != nil {
		return n.ibcOperationError(phase+"-hermes-stop-heights", err)
	}
	if err := n.Relayer.StopRelayer(ctx, n.execReporter); err != nil {
		return n.ibcOperationError(phase+"-hermes-stop", err)
	}
	n.relayerStarted = false
	if err := n.artifacts.base.appendJSONLine("ibc/hermes/restarts.jsonl", map[string]any{
		"recorded_at":    time.Now().UTC(),
		"phase":          phase,
		"operation":      "stop",
		"panacea_height": beforePanacea,
		"osmosis_height": beforeOsmosis,
	}); err != nil {
		return n.ibcOperationError(phase+"-hermes-stop-artifact", err)
	}
	return nil
}

func (n *IBCTopology) startHermes(
	ctx context.Context,
	phase string,
	healthArtifact string,
) (IBCHermesRestartEvidence, error) {
	var zero IBCHermesRestartEvidence
	if n.relayerStarted {
		return zero, n.ibcOperationError(phase+"-hermes-start", errors.New("Hermes is already running"))
	}
	beforePanacea, beforeOsmosis, err := n.chainHeights(ctx)
	if err != nil {
		return zero, n.ibcOperationError(phase+"-hermes-start-heights-before", err)
	}
	if err := n.Relayer.StartRelayer(ctx, n.execReporter, n.Path); err != nil {
		return zero, n.ibcOperationError(phase+"-hermes-start", err)
	}
	n.relayerStarted = true
	if err := testutil.WaitForBlocks(ctx, 2, n.Panacea, n.Osmosis); err != nil {
		return zero, n.ibcOperationError(phase+"-hermes-start-blocks", err)
	}
	if _, err := n.artifacts.execHermesEvidence(
		ctx, n.hermes, n.execReporter, healthArtifact, []string{"hermes", "health-check"},
	); err != nil {
		return zero, n.ibcOperationError(phase+"-hermes-health-check", err)
	}
	afterPanacea, afterOsmosis, err := n.chainHeights(ctx)
	if err != nil {
		return zero, n.ibcOperationError(phase+"-hermes-start-heights-after", err)
	}
	evidence := IBCHermesRestartEvidence{
		Phase:                phase,
		PanaceaBeforeHeight:  beforePanacea,
		OsmosisBeforeHeight:  beforeOsmosis,
		PanaceaAfterHeight:   afterPanacea,
		OsmosisAfterHeight:   afterOsmosis,
		HealthCheckCompleted: true,
	}
	if err := evidence.validate(); err != nil {
		return zero, n.ibcOperationError(phase+"-hermes-restart-evidence", err)
	}
	if err := n.artifacts.base.appendJSONLine("ibc/hermes/restarts.jsonl", map[string]any{
		"recorded_at": time.Now().UTC(),
		"operation":   "start",
		"evidence":    evidence,
	}); err != nil {
		return zero, n.ibcOperationError(phase+"-hermes-start-artifact", err)
	}
	return evidence, nil
}

func countSuccessfulRecvPackets(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	startHeight int64,
	endHeight int64,
	want ibc.Packet,
) (int, error) {
	return countSuccessfulIBCPackets(ctx, chain, startHeight, endHeight, "/ibc.core.channel.v1.MsgRecvPacket", func(value []byte) (bool, error) {
		var message channeltypes.MsgRecvPacket
		if err := proto.Unmarshal(value, &message); err != nil {
			return false, err
		}
		return recvPacketMatches(message.Packet, want), nil
	})
}

func countSuccessfulAcknowledgements(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	startHeight int64,
	endHeight int64,
	want ibc.Packet,
) (int, error) {
	return countSuccessfulIBCPackets(ctx, chain, startHeight, endHeight, "/ibc.core.channel.v1.MsgAcknowledgement", func(value []byte) (bool, error) {
		var message channeltypes.MsgAcknowledgement
		if err := proto.Unmarshal(value, &message); err != nil {
			return false, err
		}
		return toInterchaintestPacket(message.Packet).Equal(want), nil
	})
}

func countSuccessfulTimeouts(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	startHeight int64,
	endHeight int64,
	want ibc.Packet,
) (int, error) {
	return countSuccessfulIBCPackets(ctx, chain, startHeight, endHeight, "/ibc.core.channel.v1.MsgTimeout", func(value []byte) (bool, error) {
		var message channeltypes.MsgTimeout
		if err := proto.Unmarshal(value, &message); err != nil {
			return false, err
		}
		return toInterchaintestPacket(message.Packet).Equal(want), nil
	})
}

func countSuccessfulIBCPackets(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	startHeight int64,
	endHeight int64,
	typeURL string,
	matches func([]byte) (bool, error),
) (int, error) {
	if chain == nil || startHeight < 1 || endHeight < startHeight || matches == nil {
		return 0, errors.New("IBC packet count range is invalid")
	}
	count := 0
	for height := startHeight; height <= endHeight; height++ {
		values, err := successfulIBCMessageValues(ctx, chain, height, typeURL)
		if err != nil {
			return 0, err
		}
		for _, value := range values {
			matched, err := matches(value)
			if err != nil {
				return 0, fmt.Errorf("decode %s at height %d: %w", typeURL, height, err)
			}
			if matched {
				count++
			}
		}
	}
	return count, nil
}

type ibcInFlightPreRelayBalances struct {
	sourceNative       sdkmath.Int
	destinationVoucher sdkmath.Int
	destinationDenom   string
}

func validateInFlightPreRelayBalances(
	checkpoint IBCInFlightPacketCheckpoint,
	sourceNative sdkmath.Int,
	destinationVoucher sdkmath.Int,
) error {
	if sourceNative.String() != checkpoint.SourceNativeBalance.After {
		return fmt.Errorf(
			"in-flight source balance changed during the Panacea upgrade: got %s, want %s",
			sourceNative.String(), checkpoint.SourceNativeBalance.After,
		)
	}
	if destinationVoucher.String() != checkpoint.DestinationVoucherBalance.After {
		return fmt.Errorf(
			"in-flight destination balance changed before Hermes restart: got %s, want %s",
			destinationVoucher.String(), checkpoint.DestinationVoucherBalance.After,
		)
	}
	return nil
}

// captureInFlightPreRelayBalances runs while Hermes is still stopped. These
// values are the race-free baseline for the relay assertions: once Hermes is
// started, it is allowed to deliver the packet immediately.
func (n *IBCTopology) captureInFlightPreRelayBalances(
	ctx context.Context,
	req IBCPreUpgradeTransferRequest,
) (ibcInFlightPreRelayBalances, error) {
	var zero ibcInFlightPreRelayBalances
	if n.relayerStarted {
		return zero, errors.New("Hermes must remain stopped while capturing pre-relay balances")
	}
	checkpoint := *n.inFlightCheckpoint
	handshake := *n.channel
	sourceAddress := req.PanaceaUser.FormattedAddress()
	destinationAddress := req.OsmosisUser.FormattedAddress()
	sourceDenom := n.Panacea.Config().Denom
	destinationDenom := ibcVoucherDenom(handshake.Osmosis, sourceDenom)

	sourceNative, err := n.Panacea.GetBalance(ctx, sourceAddress, sourceDenom)
	if err != nil {
		return zero, err
	}
	destinationVoucher, err := n.Osmosis.GetBalance(ctx, destinationAddress, destinationDenom)
	if err != nil {
		return zero, err
	}
	if err := validateInFlightPreRelayBalances(checkpoint, sourceNative, destinationVoucher); err != nil {
		return zero, err
	}

	evidence := struct {
		RecordedAt                time.Time          `json:"recorded_at"`
		Phase                     string             `json:"phase"`
		HermesRunning             bool               `json:"hermes_running"`
		SourceNativeBalance       IBCBalanceEvidence `json:"source_native_balance"`
		DestinationVoucherBalance IBCBalanceEvidence `json:"destination_voucher_balance"`
	}{
		RecordedAt:    time.Now().UTC(),
		Phase:         "post-upgrade-before-relay",
		HermesRunning: n.relayerStarted,
		SourceNativeBalance: newIBCBalanceEvidence(
			handshake.Panacea.ChainID, sourceAddress, sourceDenom, sourceNative, sourceNative, sourceNative,
		),
		DestinationVoucherBalance: newIBCBalanceEvidence(
			handshake.Osmosis.ChainID, destinationAddress, destinationDenom,
			destinationVoucher, destinationVoucher, destinationVoucher,
		),
	}
	if err := n.artifacts.base.writeJSON("ibc/upgrade/pre-relay-balances.json", evidence); err != nil {
		return zero, err
	}
	return ibcInFlightPreRelayBalances{
		sourceNative:       sourceNative,
		destinationVoucher: destinationVoucher,
		destinationDenom:   destinationDenom,
	}, nil
}

func (n *IBCTopology) observeInFlightRelay(
	ctx context.Context,
	req IBCPreUpgradeTransferRequest,
	preRelay ibcInFlightPreRelayBalances,
) (IBCInFlightRelayEvidence, IBCLinkStateSnapshot, error) {
	var zeroRelay IBCInFlightRelayEvidence
	var zeroState IBCLinkStateSnapshot
	checkpoint := *n.inFlightCheckpoint
	tx := *n.inFlightTx
	handshake := *n.channel
	sourceBefore := preRelay.sourceNative
	destinationBefore := preRelay.destinationVoucher
	destinationDenom := preRelay.destinationDenom

	destinationCurrentHeight, err := n.Osmosis.Height(ctx)
	if err != nil {
		return zeroRelay, zeroState, fmt.Errorf("query Osmosis height before receive polling: %w", err)
	}
	recvHeight, err := pollForRecvPacket(
		ctx,
		n.Osmosis,
		checkpoint.DestinationScanStartHeight,
		ibcPacketTerminalPollMaxHeight(checkpoint.DestinationScanStartHeight, destinationCurrentHeight),
		tx.Packet,
	)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	sourceCurrentHeight, err := n.Panacea.Height(ctx)
	if err != nil {
		return zeroRelay, zeroState, fmt.Errorf("query Panacea height before acknowledgement polling: %w", err)
	}
	ackAtHeight, err := pollForPacketAcknowledgement(
		ctx,
		n.Panacea,
		tx.Height,
		ibcPacketTerminalPollMaxHeight(tx.Height, sourceCurrentHeight),
		tx.Packet,
	)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	if err := ackAtHeight.Ack.Validate(); err != nil {
		return zeroRelay, zeroState, err
	}
	rawAck := base64.StdEncoding.EncodeToString(ackAtHeight.Ack.Acknowledgement)
	if err := validateSuccessfulAcknowledgement(rawAck); err != nil {
		return zeroRelay, zeroState, err
	}

	panaceaHeight, osmosisHeight, err := n.chainHeights(ctx)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	panaceaHeight = stableIBCScanEndHeight(panaceaHeight)
	osmosisHeight = stableIBCScanEndHeight(osmosisHeight)
	recvCount, err := countSuccessfulRecvPackets(ctx, n.Osmosis, checkpoint.DestinationScanStartHeight, osmosisHeight, tx.Packet)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	ackCount, err := countSuccessfulAcknowledgements(ctx, n.Panacea, tx.Height, panaceaHeight, tx.Packet)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	commitment, commitmentExists, err := queryPacketCommitment(ctx, n.Panacea, handshake.Panacea, tx.Packet.Sequence)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	receipt, err := queryPacketReceipt(ctx, n.Osmosis, handshake.Osmosis, tx.Packet.Sequence)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	ackCommitment, ackExists, err := queryPacketAcknowledgement(ctx, n.Osmosis, handshake.Osmosis, tx.Packet.Sequence)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	if commitmentExists || len(commitment) != 0 || !ackExists {
		return zeroRelay, zeroState, errors.New("in-flight packet did not reach its commitment/acknowledgement terminal state")
	}

	destinationExpected := destinationBefore.Add(req.Amount)
	if err := cosmos.PollForBalance(ctx, n.Osmosis, 50, ibc.WalletAmount{
		Address: req.OsmosisUser.FormattedAddress(), Denom: destinationDenom, Amount: destinationExpected,
	}); err != nil {
		return zeroRelay, zeroState, err
	}
	sourceAfter, err := n.Panacea.GetBalance(ctx, req.PanaceaUser.FormattedAddress(), n.Panacea.Config().Denom)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	destinationAfter, err := n.Osmosis.GetBalance(ctx, req.OsmosisUser.FormattedAddress(), destinationDenom)
	if err != nil {
		return zeroRelay, zeroState, err
	}
	lifecycle := packetLifecycleFromSend(checkpoint.Packet)
	lifecycle.Recv = IBCPacketObservation{Observed: true, ChainID: handshake.Osmosis.ChainID, Height: recvHeight}
	lifecycle.Ack = IBCAcknowledgementObservation{
		Observed: true, ChainID: handshake.Panacea.ChainID, Height: ackAtHeight.Height, Acknowledgement: rawAck,
	}
	if err := n.recordIBCPacketStage(ibcUpgradeContinuityPhase, "recv", lifecycle); err != nil {
		return zeroRelay, zeroState, err
	}
	if err := n.recordIBCPacketStage(ibcUpgradeContinuityPhase, "ack", lifecycle); err != nil {
		return zeroRelay, zeroState, err
	}
	relay := IBCInFlightRelayEvidence{
		Packet:                     lifecycle,
		ReceiveCount:               recvCount,
		AcknowledgementCount:       ackCount,
		CommitmentCleared:          !commitmentExists,
		DestinationReceipt:         receipt,
		DestinationAcknowledgement: base64IfNotEmpty(ackCommitment),
		SourceNativeBalance: newIBCBalanceEvidence(
			handshake.Panacea.ChainID, req.PanaceaUser.FormattedAddress(), n.Panacea.Config().Denom, sourceBefore, sourceAfter, sourceBefore,
		),
		DestinationVoucherBalance: newIBCBalanceEvidence(
			handshake.Osmosis.ChainID, req.OsmosisUser.FormattedAddress(), destinationDenom, destinationBefore, destinationAfter, destinationExpected,
		),
	}
	if err := relay.validate(checkpoint); err != nil {
		return zeroRelay, zeroState, err
	}
	state, err := n.queryIBCLinkState(ctx, "post-upgrade-after-in-flight-relay")
	if err != nil {
		return zeroRelay, zeroState, err
	}
	return relay, state, nil
}

func ibcPacketTerminalPollMaxHeight(startHeight, currentHeight int64) int64 {
	if currentHeight < startHeight {
		currentHeight = startHeight
	}
	return currentHeight + ibcPacketTerminalWaitMax
}

func (n *IBCTopology) runPostUpgradeTimeout(
	ctx context.Context,
	req IBCPreUpgradeTransferRequest,
	beforeState IBCLinkStateSnapshot,
) (IBCPacketTimeoutEvidence, error) {
	var zero IBCPacketTimeoutEvidence
	if !n.relayerStarted {
		return zero, errors.New("Hermes must be running before the timeout pause")
	}
	handshake := *n.channel
	panaceaAddress := req.PanaceaUser.FormattedAddress()
	osmosisAddress := req.OsmosisUser.FormattedAddress()
	panaceaDenom := n.Panacea.Config().Denom
	destinationDenom := ibcVoucherDenom(handshake.Osmosis, panaceaDenom)
	sourceBefore, err := n.Panacea.GetBalance(ctx, panaceaAddress, panaceaDenom)
	if err != nil {
		return zero, err
	}
	destinationBefore, err := n.Osmosis.GetBalance(ctx, osmosisAddress, destinationDenom)
	if err != nil {
		return zero, err
	}
	escrowAddress, escrowBeforeSend, err := queryIBCEscrowBalance(ctx, n.Panacea, handshake.Panacea, panaceaDenom)
	if err != nil {
		return zero, err
	}
	destinationStartHeight, err := n.Osmosis.Height(ctx)
	if err != nil {
		return zero, err
	}

	if err := n.Relayer.PauseRelayer(ctx); err != nil {
		return zero, n.ibcOperationError("upgrade-timeout-hermes-pause", err)
	}
	paused := true
	defer func() {
		if paused {
			_ = n.Relayer.ResumeRelayer(context.Background())
		}
	}()
	if err := n.artifacts.base.appendJSONLine("ibc/hermes/restarts.jsonl", map[string]any{
		"recorded_at": time.Now().UTC(), "phase": ibcPhasePostUpgradeTimeout, "operation": "pause",
	}); err != nil {
		return zero, err
	}

	timeoutOptions, err := relativeIBCTimestampTransferOptions(ibcPostUpgradeTimeout)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-timeout-options", err)
	}
	tx, err := n.Panacea.SendIBCTransfer(
		ctx,
		handshake.Panacea.ChannelID,
		req.PanaceaUser.KeyName(),
		ibc.WalletAmount{Address: osmosisAddress, Denom: panaceaDenom, Amount: req.Amount},
		timeoutOptions,
	)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-timeout-send", err)
	}
	if err := validateIBCTransferTx(tx, handshake.Panacea, handshake.Osmosis); err != nil {
		return zero, n.ibcOperationError("upgrade-timeout-send-validate", err)
	}
	timeoutTimestamp, err := committedPacketTimeoutTimestamp(tx)
	if err != nil {
		return zero, n.ibcOperationError("upgrade-timeout-send-deadline", err)
	}
	fee := n.Panacea.GetGasFeesInNativeDenom(tx.GasSpent)
	send := packetSendEvidenceFromTx(tx, ibcPanaceaToOsmosis, handshake.Panacea, handshake.Osmosis, panaceaDenom, req.Amount, fee)
	if err := n.recordIBCPacketStage(ibcPhasePostUpgradeTimeout, "send", packetLifecycleFromSend(send)); err != nil {
		return zero, err
	}
	commitment, exists, err := queryPacketCommitment(ctx, n.Panacea, handshake.Panacea, tx.Packet.Sequence)
	if err != nil || !exists || len(commitment) == 0 {
		return zero, errors.Join(err, errors.New("timeout packet source commitment was not created"))
	}
	_, escrowLocked, err := queryIBCEscrowBalance(ctx, n.Panacea, handshake.Panacea, panaceaDenom)
	if err != nil {
		return zero, err
	}
	receipt, err := queryPacketReceipt(ctx, n.Osmosis, handshake.Osmosis, tx.Packet.Sequence)
	if err != nil {
		return zero, err
	}
	if receipt {
		return zero, errors.New("timeout packet was received while Hermes was paused")
	}
	if err := waitForChainTimestamp(ctx, n.Osmosis, timeoutTimestamp); err != nil {
		return zero, n.ibcOperationError("upgrade-timeout-expiry-wait", err)
	}
	if err := n.Relayer.ResumeRelayer(ctx); err != nil {
		return zero, n.ibcOperationError("upgrade-timeout-hermes-resume", err)
	}
	paused = false
	if err := n.artifacts.base.appendJSONLine("ibc/hermes/restarts.jsonl", map[string]any{
		"recorded_at": time.Now().UTC(), "phase": ibcPhasePostUpgradeTimeout, "operation": "resume",
	}); err != nil {
		return zero, err
	}
	if _, err := n.artifacts.execHermesEvidence(
		ctx, n.hermes, n.execReporter, "health-check-post-upgrade-timeout-resume", []string{"hermes", "health-check"},
	); err != nil {
		return zero, err
	}

	timeoutResumePanaceaHeight, err := n.Panacea.Height(ctx)
	if err != nil {
		return zero, err
	}
	timeoutHeight, err := pollForTimeoutPacket(
		ctx,
		n.Panacea,
		tx.Height,
		ibcPacketTerminalPollMaxHeight(tx.Height, timeoutResumePanaceaHeight),
		tx.Packet,
	)
	if err != nil {
		return zero, err
	}
	expectedSource := sourceBefore.SubRaw(fee)
	if err := cosmos.PollForBalance(ctx, n.Panacea, 50, ibc.WalletAmount{
		Address: panaceaAddress, Denom: panaceaDenom, Amount: expectedSource,
	}); err != nil {
		return zero, err
	}
	panaceaHeight, osmosisHeight, err := n.chainHeights(ctx)
	if err != nil {
		return zero, err
	}
	panaceaHeight = stableIBCScanEndHeight(panaceaHeight)
	osmosisHeight = stableIBCScanEndHeight(osmosisHeight)
	timeoutCount, err := countSuccessfulTimeouts(ctx, n.Panacea, tx.Height, panaceaHeight, tx.Packet)
	if err != nil {
		return zero, err
	}
	recvCount, err := countSuccessfulRecvPackets(ctx, n.Osmosis, destinationStartHeight, osmosisHeight, tx.Packet)
	if err != nil {
		return zero, err
	}
	commitment, commitmentExists, err := queryPacketCommitment(ctx, n.Panacea, handshake.Panacea, tx.Packet.Sequence)
	if err != nil {
		return zero, err
	}
	receipt, err = queryPacketReceipt(ctx, n.Osmosis, handshake.Osmosis, tx.Packet.Sequence)
	if err != nil {
		return zero, err
	}
	ack, ackExists, err := queryPacketAcknowledgement(ctx, n.Osmosis, handshake.Osmosis, tx.Packet.Sequence)
	if err != nil {
		return zero, err
	}
	if commitmentExists || len(commitment) != 0 || ackExists {
		return zero, errors.New("timeout packet did not clear its commitment or unexpectedly wrote an acknowledgement")
	}
	sourceAfter, err := n.Panacea.GetBalance(ctx, panaceaAddress, panaceaDenom)
	if err != nil {
		return zero, err
	}
	destinationAfter, err := n.Osmosis.GetBalance(ctx, osmosisAddress, destinationDenom)
	if err != nil {
		return zero, err
	}
	_, escrowAfterRefund, err := queryIBCEscrowBalance(ctx, n.Panacea, handshake.Panacea, panaceaDenom)
	if err != nil {
		return zero, err
	}
	afterState, err := n.queryIBCLinkState(ctx, "post-upgrade-after-timeout")
	if err != nil {
		return zero, err
	}
	evidence := IBCPacketTimeoutEvidence{
		Phase:                      ibcPhasePostUpgradeTimeout,
		BeforeSendState:            beforeState,
		AfterTimeoutState:          afterState,
		Packet:                     send,
		TimeoutCount:               timeoutCount,
		ReceiveCount:               recvCount,
		CommitmentCleared:          !commitmentExists,
		DestinationReceipt:         receipt,
		DestinationAcknowledgement: base64IfNotEmpty(ack),
		SourceNativeBalance: newIBCBalanceEvidence(
			handshake.Panacea.ChainID, panaceaAddress, panaceaDenom, sourceBefore, sourceAfter, expectedSource,
		),
		DestinationVoucherBalance: newIBCBalanceEvidence(
			handshake.Osmosis.ChainID, osmosisAddress, destinationDenom, destinationBefore, destinationAfter, destinationBefore,
		),
		SourceEscrowLock: newIBCEscrowBalanceEvidence(
			ibcPhasePostUpgradeTimeout+"-lock", handshake.Panacea, escrowAddress, panaceaDenom,
			escrowBeforeSend, escrowLocked, req.Amount,
		),
		SourceEscrowRefund: newIBCEscrowBalanceEvidence(
			ibcPhasePostUpgradeTimeout+"-refund", handshake.Panacea, escrowAddress, panaceaDenom,
			escrowLocked, escrowAfterRefund, req.Amount.Neg(),
		),
	}
	if timeoutHeight < tx.Height {
		return zero, errors.New("timeout message predates the packet send")
	}
	if err := evidence.validate(handshake); err != nil {
		return zero, err
	}
	if err := n.recordIBCPacketStage(ibcPhasePostUpgradeTimeout, "timeout", packetLifecycleFromSend(send)); err != nil {
		return zero, err
	}
	if err := n.artifacts.base.writeJSON("ibc/upgrade/timeout-refund.json", evidence); err != nil {
		return zero, err
	}
	return evidence, nil
}

func pollForTimeoutPacket(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	startHeight int64,
	maxHeight int64,
	want ibc.Packet,
) (int64, error) {
	poller := testutil.BlockPoller[int64]{
		CurrentHeight: stableIBCQueryHeight(chain.Height),
		PollFunc: func(ctx context.Context, height int64) (int64, error) {
			values, err := successfulIBCMessageValues(ctx, chain, height, "/ibc.core.channel.v1.MsgTimeout")
			if err != nil {
				return 0, err
			}
			for _, value := range values {
				var message channeltypes.MsgTimeout
				if err := proto.Unmarshal(value, &message); err != nil {
					return 0, err
				}
				if toInterchaintestPacket(message.Packet).Equal(want) {
					return height, nil
				}
			}
			return 0, testutil.ErrNotFound
		},
	}
	height, err := poller.DoPoll(ctx, startHeight, maxHeight)
	if err != nil {
		return 0, fmt.Errorf("find timeout for packet sequence %d: %w", want.Sequence, err)
	}
	return height, nil
}

func relativeIBCTimestampTransferOptions(timeout time.Duration) (ibc.TransferOptions, error) {
	timeoutNanos := timeout.Nanoseconds()
	if timeoutNanos <= 0 {
		return ibc.TransferOptions{}, errors.New("relative IBC timeout duration must be positive")
	}
	return ibc.TransferOptions{
		Timeout: &ibc.IBCTimeout{NanoSeconds: uint64(timeoutNanos)},
	}, nil
}

func committedPacketTimeoutTimestamp(tx ibc.Tx) (uint64, error) {
	timestamp := uint64(tx.Packet.TimeoutTimestamp)
	if timestamp == 0 {
		return 0, errors.New("committed IBC packet timeout timestamp is missing")
	}
	return timestamp, nil
}

func waitForChainTimestamp(ctx context.Context, chain *cosmos.CosmosChain, target uint64) error {
	if target == 0 {
		return errors.New("IBC timeout timestamp must be positive")
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		height, err := chain.Height(ctx)
		if err == nil {
			block, blockErr := chain.GetFullNode().Client.Block(ctx, &height)
			if blockErr == nil && block != nil && block.Block != nil && uint64(block.Block.Time.UnixNano()) > target {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for chain time after %d: %w", target, ctx.Err())
		case <-ticker.C:
		}
	}
}

func (n *IBCTopology) runPostUpgradeBidirectionalTransfers(
	ctx context.Context,
	req IBCPreUpgradeTransferRequest,
) (IBCPostUpgradeTransferEvidence, error) {
	var zero IBCPostUpgradeTransferEvidence
	handshake := *n.channel
	panaceaAddress := req.PanaceaUser.FormattedAddress()
	osmosisAddress := req.OsmosisUser.FormattedAddress()
	panaceaDenom := n.Panacea.Config().Denom
	osmosisDenom := n.Osmosis.Config().Denom
	medOnOsmosis := ibcVoucherDenom(handshake.Osmosis, panaceaDenom)
	osmoOnPanacea := ibcVoucherDenom(handshake.Panacea, osmosisDenom)

	panaceaNativeBefore, err := n.Panacea.GetBalance(ctx, panaceaAddress, panaceaDenom)
	if err != nil {
		return zero, err
	}
	panaceaVoucherBefore, err := n.Panacea.GetBalance(ctx, panaceaAddress, osmoOnPanacea)
	if err != nil {
		return zero, err
	}
	osmosisNativeBefore, err := n.Osmosis.GetBalance(ctx, osmosisAddress, osmosisDenom)
	if err != nil {
		return zero, err
	}
	osmosisVoucherBefore, err := n.Osmosis.GetBalance(ctx, osmosisAddress, medOnOsmosis)
	if err != nil {
		return zero, err
	}
	panaceaEscrowAddress, panaceaEscrowBefore, err := queryIBCEscrowBalance(ctx, n.Panacea, handshake.Panacea, panaceaDenom)
	if err != nil {
		return zero, err
	}
	osmosisEscrowAddress, osmosisEscrowBefore, err := queryIBCEscrowBalance(ctx, n.Osmosis, handshake.Osmosis, osmosisDenom)
	if err != nil {
		return zero, err
	}

	panaceaTransfer, panaceaFee, err := n.executeIBCTransfer(
		ctx, ibcPhasePostUpgrade, ibcPanaceaToOsmosis,
		n.Panacea, n.Osmosis, req.PanaceaUser, req.OsmosisUser,
		handshake.Panacea, handshake.Osmosis, panaceaDenom, req.Amount,
	)
	if err != nil {
		return zero, err
	}
	osmosisTransfer, osmosisFee, err := n.executeIBCTransfer(
		ctx, ibcPhasePostUpgrade, ibcOsmosisToPanacea,
		n.Osmosis, n.Panacea, req.OsmosisUser, req.PanaceaUser,
		handshake.Osmosis, handshake.Panacea, osmosisDenom, req.Amount,
	)
	if err != nil {
		return zero, err
	}

	panaceaNativeExpected := panaceaNativeBefore.Sub(req.Amount).SubRaw(panaceaFee)
	panaceaVoucherExpected := panaceaVoucherBefore.Add(req.Amount)
	osmosisNativeExpected := osmosisNativeBefore.Sub(req.Amount).SubRaw(osmosisFee)
	osmosisVoucherExpected := osmosisVoucherBefore.Add(req.Amount)
	if panaceaNativeExpected.IsNegative() || osmosisNativeExpected.IsNegative() {
		return zero, errors.New("post-upgrade IBC source balance is insufficient for amount and fee")
	}
	for _, expected := range []struct {
		chain  *cosmos.CosmosChain
		wallet ibc.WalletAmount
	}{
		{n.Panacea, ibc.WalletAmount{Address: panaceaAddress, Denom: panaceaDenom, Amount: panaceaNativeExpected}},
		{n.Panacea, ibc.WalletAmount{Address: panaceaAddress, Denom: osmoOnPanacea, Amount: panaceaVoucherExpected}},
		{n.Osmosis, ibc.WalletAmount{Address: osmosisAddress, Denom: osmosisDenom, Amount: osmosisNativeExpected}},
		{n.Osmosis, ibc.WalletAmount{Address: osmosisAddress, Denom: medOnOsmosis, Amount: osmosisVoucherExpected}},
	} {
		if err := cosmos.PollForBalance(ctx, expected.chain, 50, expected.wallet); err != nil {
			return zero, err
		}
	}

	panaceaNativeAfter, err := n.Panacea.GetBalance(ctx, panaceaAddress, panaceaDenom)
	if err != nil {
		return zero, err
	}
	panaceaVoucherAfter, err := n.Panacea.GetBalance(ctx, panaceaAddress, osmoOnPanacea)
	if err != nil {
		return zero, err
	}
	osmosisNativeAfter, err := n.Osmosis.GetBalance(ctx, osmosisAddress, osmosisDenom)
	if err != nil {
		return zero, err
	}
	osmosisVoucherAfter, err := n.Osmosis.GetBalance(ctx, osmosisAddress, medOnOsmosis)
	if err != nil {
		return zero, err
	}
	_, panaceaEscrowAfter, err := queryIBCEscrowBalance(ctx, n.Panacea, handshake.Panacea, panaceaDenom)
	if err != nil {
		return zero, err
	}
	_, osmosisEscrowAfter, err := queryIBCEscrowBalance(ctx, n.Osmosis, handshake.Osmosis, osmosisDenom)
	if err != nil {
		return zero, err
	}
	postUpgradeTraces, err := n.queryDenomTraceSnapshot(ctx, ibcPhasePostUpgrade)
	if err != nil {
		return zero, err
	}
	evidence := IBCPostUpgradeTransferEvidence{
		Phase:     ibcPhasePostUpgrade,
		Channel:   handshake,
		Transfers: []IBCPacketLifecycleEvidence{panaceaTransfer, osmosisTransfer},
		FinalBalances: []IBCBalanceEvidence{
			newIBCBalanceEvidence(handshake.Panacea.ChainID, panaceaAddress, panaceaDenom, panaceaNativeBefore, panaceaNativeAfter, panaceaNativeExpected),
			newIBCBalanceEvidence(handshake.Panacea.ChainID, panaceaAddress, osmoOnPanacea, panaceaVoucherBefore, panaceaVoucherAfter, panaceaVoucherExpected),
			newIBCBalanceEvidence(handshake.Osmosis.ChainID, osmosisAddress, osmosisDenom, osmosisNativeBefore, osmosisNativeAfter, osmosisNativeExpected),
			newIBCBalanceEvidence(handshake.Osmosis.ChainID, osmosisAddress, medOnOsmosis, osmosisVoucherBefore, osmosisVoucherAfter, osmosisVoucherExpected),
		},
		EscrowBalances: []IBCEscrowBalanceEvidence{
			newIBCEscrowBalanceEvidence(ibcPhasePostUpgrade, handshake.Panacea, panaceaEscrowAddress, panaceaDenom, panaceaEscrowBefore, panaceaEscrowAfter, req.Amount),
			newIBCEscrowBalanceEvidence(ibcPhasePostUpgrade, handshake.Osmosis, osmosisEscrowAddress, osmosisDenom, osmosisEscrowBefore, osmosisEscrowAfter, req.Amount),
		},
		DenomTraces: postUpgradeTraces,
	}
	if err := evidence.Validate(); err != nil {
		return zero, err
	}
	if err := n.artifacts.base.writeJSON("ibc/state/post-upgrade-bidirectional.json", evidence); err != nil {
		return zero, err
	}
	return evidence, nil
}

func (n *IBCTopology) restartEveryIBCNode(
	ctx context.Context,
) ([]IBCNodeRuntimeIdentity, []NodeRestartEvidence, []IBCNodeRuntimeIdentity, []NodeRestartEvidence, error) {
	panaceaBefore, err := captureIBCNodeRuntimeIdentities(n.Panacea)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	osmosisBefore, err := captureIBCNodeRuntimeIdentities(n.Osmosis)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	panaceaRestarts, err := n.restartEveryChainNode(ctx, "Panacea", n.Panacea)
	if err != nil {
		return panaceaBefore, panaceaRestarts, osmosisBefore, nil, err
	}
	osmosisRestarts, err := n.restartEveryChainNode(ctx, "Osmosis", n.Osmosis)
	if err != nil {
		return panaceaBefore, panaceaRestarts, osmosisBefore, osmosisRestarts, err
	}
	artifact := map[string]any{
		"panacea": panaceaRestarts,
		"osmosis": osmosisRestarts,
	}
	if err := n.artifacts.base.writeJSON("ibc/upgrade/all-node-restarts.json", artifact); err != nil {
		return panaceaBefore, panaceaRestarts, osmosisBefore, osmosisRestarts, err
	}
	return panaceaBefore, panaceaRestarts, osmosisBefore, osmosisRestarts, nil
}

func (n *IBCTopology) restartEveryChainNode(
	ctx context.Context,
	chainLabel string,
	chain *cosmos.CosmosChain,
) ([]NodeRestartEvidence, error) {
	view := &Network{Chain: chain, artifacts: n.artifacts.base}
	nodes := make([]*cosmos.ChainNode, 0, len(chain.Nodes()))
	nodes = append(nodes, chain.FullNodes...)
	nodes = append(nodes, chain.Validators...)
	if len(nodes) == 0 {
		return nil, fmt.Errorf("%s has no nodes to restart", chainLabel)
	}
	restarts := make([]NodeRestartEvidence, 0, len(nodes))
	for _, node := range nodes {
		restart, err := view.GracefulRestartNode(ctx, node)
		if err != nil {
			return restarts, err
		}
		restarts = append(restarts, restart)
	}
	if err := validateChainNodeRestarts(chainLabel, restarts, len(nodes)); err != nil {
		return restarts, err
	}
	return restarts, nil
}

func (n *IBCTopology) queryFinalPacketStates(
	ctx context.Context,
	relay IBCInFlightRelayEvidence,
	timeoutEvidence IBCPacketTimeoutEvidence,
) ([]IBCPacketTerminalStateEvidence, error) {
	handshake := *n.channel
	_, successCommitmentExists, err := queryPacketCommitment(ctx, n.Panacea, handshake.Panacea, relay.Packet.Sequence)
	if err != nil {
		return nil, err
	}
	successReceipt, err := queryPacketReceipt(ctx, n.Osmosis, handshake.Osmosis, relay.Packet.Sequence)
	if err != nil {
		return nil, err
	}
	successAck, successAckExists, err := queryPacketAcknowledgement(ctx, n.Osmosis, handshake.Osmosis, relay.Packet.Sequence)
	if err != nil {
		return nil, err
	}
	_, timeoutCommitmentExists, err := queryPacketCommitment(ctx, n.Panacea, handshake.Panacea, timeoutEvidence.Packet.Sequence)
	if err != nil {
		return nil, err
	}
	timeoutReceipt, err := queryPacketReceipt(ctx, n.Osmosis, handshake.Osmosis, timeoutEvidence.Packet.Sequence)
	if err != nil {
		return nil, err
	}
	timeoutAck, timeoutAckExists, err := queryPacketAcknowledgement(ctx, n.Osmosis, handshake.Osmosis, timeoutEvidence.Packet.Sequence)
	if err != nil {
		return nil, err
	}
	if successCommitmentExists || !successAckExists || timeoutCommitmentExists || timeoutAckExists {
		return nil, errors.New("post-restart packet commitments or acknowledgement commitments are invalid")
	}
	states := []IBCPacketTerminalStateEvidence{
		{
			Kind:                       ibcPacketTerminalSuccess,
			Sequence:                   relay.Packet.Sequence,
			DestinationReceipt:         successReceipt,
			DestinationAcknowledgement: base64IfNotEmpty(successAck),
		},
		{
			Kind:                       ibcPacketTerminalTimeout,
			Sequence:                   timeoutEvidence.Packet.Sequence,
			DestinationReceipt:         timeoutReceipt,
			DestinationAcknowledgement: base64IfNotEmpty(timeoutAck),
		},
	}
	for _, terminal := range states {
		if err := terminal.validate(relay.Packet.Sequence, timeoutEvidence.Packet.Sequence, relay.DestinationAcknowledgement); err != nil {
			return nil, err
		}
	}
	return states, nil
}

func (n *IBCTopology) queryFinalDenomTraces(ctx context.Context) ([]IBCDenomTraceEvidence, error) {
	snapshot, err := n.queryDenomTraceSnapshot(ctx, "post-restart")
	if err != nil {
		return nil, err
	}
	if err := n.artifacts.base.writeJSON("ibc/state/post-restart-denom-traces.json", snapshot); err != nil {
		return nil, err
	}
	return snapshot.Traces, nil
}

func (n *IBCTopology) queryDenomTraceSnapshot(ctx context.Context, phase string) (IBCDenomTraceSnapshot, error) {
	var zero IBCDenomTraceSnapshot
	handshake := *n.channel
	panaceaTrace, err := queryIBCDenomTrace(
		ctx,
		n.Panacea,
		handshake.Panacea,
		n.Osmosis.Config().Denom,
	)
	if err != nil {
		return zero, err
	}
	osmosisTrace, err := queryIBCDenomTrace(
		ctx,
		n.Osmosis,
		handshake.Osmosis,
		n.Panacea.Config().Denom,
	)
	if err != nil {
		return zero, err
	}
	snapshot := IBCDenomTraceSnapshot{Phase: phase, Traces: []IBCDenomTraceEvidence{panaceaTrace, osmosisTrace}}
	if err := snapshot.validate(handshake, phase); err != nil {
		return zero, err
	}
	return snapshot, nil
}

func queryIBCDenomTrace(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	endpoint IBCChannelEndpoint,
	baseDenom string,
) (IBCDenomTraceEvidence, error) {
	voucherDenom := ibcVoucherDenom(endpoint, baseDenom)
	hash := strings.TrimPrefix(voucherDenom, "ibc/")
	response, err := transfertypes.NewQueryClient(chain.GetFullNode().GrpcConn).DenomTrace(ctx, &transfertypes.QueryDenomTraceRequest{Hash: hash})
	if err != nil {
		return IBCDenomTraceEvidence{}, err
	}
	if response == nil || response.DenomTrace == nil {
		return IBCDenomTraceEvidence{}, errors.New("IBC denom trace query returned no trace")
	}
	return IBCDenomTraceEvidence{
		ChainID:      chain.Config().ChainID,
		VoucherDenom: voucherDenom,
		Hash:         hash,
		Path:         response.DenomTrace.Path,
		BaseDenom:    response.DenomTrace.BaseDenom,
	}, nil
}

func (n *IBCTopology) queryFinalRestartBalances(
	ctx context.Context,
	prior []IBCBalanceEvidence,
) ([]IBCBalanceEvidence, error) {
	if len(prior) != 4 {
		return nil, fmt.Errorf("post-transfer balance evidence has %d entries, want 4", len(prior))
	}
	final := make([]IBCBalanceEvidence, 0, len(prior))
	for _, balance := range prior {
		var chain *cosmos.CosmosChain
		switch balance.ChainID {
		case n.Panacea.Config().ChainID:
			chain = n.Panacea
		case n.Osmosis.Config().ChainID:
			chain = n.Osmosis
		default:
			return nil, fmt.Errorf("post-transfer balance uses unknown chain %q", balance.ChainID)
		}
		current, err := chain.GetBalance(ctx, balance.Address, balance.Denom)
		if err != nil {
			return nil, err
		}
		final = append(final, IBCBalanceEvidence{
			ChainID:       balance.ChainID,
			Address:       balance.Address,
			Denom:         balance.Denom,
			Before:        balance.After,
			After:         current.String(),
			ExpectedAfter: balance.After,
		})
	}
	if err := validateFinalRestartBalances(prior, final); err != nil {
		return nil, err
	}
	if err := n.artifacts.base.writeJSON("ibc/state/post-restart-balances.json", final); err != nil {
		return nil, err
	}
	return final, nil
}

func (n *IBCTopology) queryFinalRestartEscrows(
	ctx context.Context,
	prior []IBCEscrowBalanceEvidence,
) ([]IBCEscrowBalanceEvidence, error) {
	if len(prior) != 2 {
		return nil, fmt.Errorf("post-transfer escrow evidence has %d entries, want 2", len(prior))
	}
	final := make([]IBCEscrowBalanceEvidence, 0, len(prior))
	for _, escrow := range prior {
		var chain *cosmos.CosmosChain
		var endpoint IBCChannelEndpoint
		switch escrow.ChainID {
		case n.Panacea.Config().ChainID:
			chain, endpoint = n.Panacea, n.channel.Panacea
		case n.Osmosis.Config().ChainID:
			chain, endpoint = n.Osmosis, n.channel.Osmosis
		default:
			return nil, fmt.Errorf("post-transfer escrow uses unknown chain %q", escrow.ChainID)
		}
		address, balance, err := queryIBCEscrowBalance(ctx, chain, endpoint, escrow.Denom)
		if err != nil {
			return nil, err
		}
		final = append(final, IBCEscrowBalanceEvidence{
			Phase:         "post-restart",
			ChainID:       escrow.ChainID,
			PortID:        escrow.PortID,
			ChannelID:     escrow.ChannelID,
			Address:       address,
			Denom:         escrow.Denom,
			Before:        escrow.After,
			After:         balance.String(),
			ExpectedDelta: "0",
			ExpectedAfter: escrow.After,
		})
	}
	if err := validateFinalEscrowBalances(prior, final, *n.channel); err != nil {
		return nil, err
	}
	if err := n.artifacts.base.writeJSON("ibc/state/post-restart-escrow-balances.json", final); err != nil {
		return nil, err
	}
	return final, nil
}

func (n *IBCTopology) queryPostRestartNodeSemantics(
	ctx context.Context,
	panaceaBefore []IBCNodeRuntimeIdentity,
	panaceaRestarts []NodeRestartEvidence,
	osmosisBefore []IBCNodeRuntimeIdentity,
	osmosisRestarts []NodeRestartEvidence,
	finalState IBCLinkStateSnapshot,
	relay IBCInFlightRelayEvidence,
	timeoutEvidence IBCPacketTimeoutEvidence,
) ([]IBCChainRestartEvidence, error) {
	panacea, err := n.queryChainPostRestartNodeSemantics(
		ctx, n.Panacea, n.channel.Panacea, "source", panaceaBefore, panaceaRestarts,
		finalState.PanaceaNextSequenceSend, relay, timeoutEvidence,
	)
	if err != nil {
		return nil, err
	}
	osmosis, err := n.queryChainPostRestartNodeSemantics(
		ctx, n.Osmosis, n.channel.Osmosis, "destination", osmosisBefore, osmosisRestarts,
		finalState.OsmosisNextSequenceSend, relay, timeoutEvidence,
	)
	if err != nil {
		return nil, err
	}
	groups := []IBCChainRestartEvidence{panacea, osmosis}
	if err := n.artifacts.base.writeJSON("ibc/state/post-restart-node-semantics.json", groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (n *IBCTopology) queryChainPostRestartNodeSemantics(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	endpoint IBCChannelEndpoint,
	role string,
	before []IBCNodeRuntimeIdentity,
	restarts []NodeRestartEvidence,
	wantSequence uint64,
	relay IBCInFlightRelayEvidence,
	timeoutEvidence IBCPacketTimeoutEvidence,
) (IBCChainRestartEvidence, error) {
	var zero IBCChainRestartEvidence
	after, err := captureIBCNodeRuntimeIdentities(chain)
	if err != nil {
		return zero, err
	}
	beforeByName, err := nodeIdentityMap(before)
	if err != nil {
		return zero, err
	}
	afterByName, err := nodeIdentityMap(after)
	if err != nil {
		return zero, err
	}
	restartByName := make(map[string]NodeRestartEvidence, len(restarts))
	for _, restart := range restarts {
		restartByName[restart.Node] = restart
	}
	evidence := IBCChainRestartEvidence{ChainID: chain.Config().ChainID, Nodes: make([]IBCNodePostRestartEvidence, 0, len(chain.Nodes()))}
	for _, node := range chain.Nodes() {
		if node == nil {
			return zero, errors.New("IBC post-restart chain contains a nil node")
		}
		identityBefore, beforeOK := beforeByName[node.Name()]
		identityAfter, afterOK := afterByName[node.Name()]
		restart, restartOK := restartByName[node.Name()]
		if !beforeOK || !afterOK || !restartOK || identityBefore != identityAfter {
			return zero, fmt.Errorf("IBC node %q restart identity is missing or changed", node.Name())
		}
		height, err := node.Height(ctx)
		if err != nil {
			return zero, err
		}
		clientStatus, connectionState, channelState, nextSequence, err := queryIBCNodeLinkSemantics(ctx, node, endpoint)
		if err != nil {
			return zero, err
		}
		if nextSequence != wantSequence {
			return zero, fmt.Errorf("IBC node %q next sequence = %d, want %d", node.Name(), nextSequence, wantSequence)
		}
		packets, err := queryIBCNodePacketSemantics(ctx, node, endpoint, role, relay, timeoutEvidence)
		if err != nil {
			return zero, err
		}
		trace, err := queryIBCDenomTraceOnNode(ctx, chain.Config().ChainID, node, endpoint, counterpartyBaseDenom(*n.channel, chain.Config().ChainID))
		if err != nil {
			return zero, err
		}
		evidence.Nodes = append(evidence.Nodes, IBCNodePostRestartEvidence{
			ChainID:          chain.Config().ChainID,
			IdentityBefore:   identityBefore,
			IdentityAfter:    identityAfter,
			Restart:          restart,
			ObservedHeight:   height,
			ClientStatus:     clientStatus,
			ConnectionState:  connectionState,
			ChannelState:     channelState,
			NextSequenceSend: nextSequence,
			PacketSemantics:  packets,
			DenomTrace:       trace,
		})
	}
	return evidence, nil
}

func queryIBCNodeLinkSemantics(
	ctx context.Context,
	node *cosmos.ChainNode,
	endpoint IBCChannelEndpoint,
) (string, string, string, uint64, error) {
	clientResponse, err := clienttypes.NewQueryClient(node.GrpcConn).ClientStatus(
		ctx, &clienttypes.QueryClientStatusRequest{ClientId: endpoint.ClientID},
	)
	if err != nil || clientResponse == nil || strings.TrimSpace(clientResponse.Status) == "" {
		return "", "", "", 0, errors.Join(err, errors.New("IBC node client status is empty"))
	}
	connectionResponse, err := connectiontypes.NewQueryClient(node.GrpcConn).Connection(
		ctx, &connectiontypes.QueryConnectionRequest{ConnectionId: endpoint.ConnectionID},
	)
	if err != nil || connectionResponse == nil || connectionResponse.Connection == nil {
		return "", "", "", 0, errors.Join(err, errors.New("IBC node connection response is empty"))
	}
	connection := connectionResponse.Connection
	if connection.ClientId != endpoint.ClientID ||
		connection.Counterparty.ConnectionId != endpoint.CounterpartyConnectionID {
		return "", "", "", 0, errors.New("IBC node connection identity changed after restart")
	}
	channelResponse, err := channeltypes.NewQueryClient(node.GrpcConn).Channel(
		ctx, &channeltypes.QueryChannelRequest{PortId: endpoint.PortID, ChannelId: endpoint.ChannelID},
	)
	if err != nil || channelResponse == nil || channelResponse.Channel == nil {
		return "", "", "", 0, errors.Join(err, errors.New("IBC node channel response is empty"))
	}
	channel := channelResponse.Channel
	if channel.Counterparty.PortId != endpoint.PortID || channel.Counterparty.ChannelId != endpoint.CounterpartyChannelID ||
		channel.Version != endpoint.Version || !ibcStateEquals(channel.Ordering.String(), strings.ToUpper(strings.TrimPrefix(endpoint.Ordering, "ORDER_"))) {
		return "", "", "", 0, errors.New("IBC node channel semantics changed after restart")
	}
	nextSequence, err := queryIBCNextSequenceSendOnNode(ctx, node, endpoint)
	if err != nil {
		return "", "", "", 0, err
	}
	return clientResponse.Status, connection.State.String(), channel.State.String(), nextSequence, nil
}

func queryIBCNextSequenceSendOnNode(ctx context.Context, node *cosmos.ChainNode, endpoint IBCChannelEndpoint) (uint64, error) {
	result, err := node.Client.ABCIQuery(ctx, "store/ibc/key", host.NextSequenceSendKey(endpoint.PortID, endpoint.ChannelID))
	if err != nil {
		return 0, err
	}
	if result == nil || result.Response.IsErr() {
		return 0, errors.New("IBC node next send sequence query failed")
	}
	return decodeIBCNextSequenceSend(result.Response.Value)
}

func queryIBCNodePacketSemantics(
	ctx context.Context,
	node *cosmos.ChainNode,
	endpoint IBCChannelEndpoint,
	role string,
	relay IBCInFlightRelayEvidence,
	timeoutEvidence IBCPacketTimeoutEvidence,
) (IBCNodePacketSemanticsEvidence, error) {
	evidence := IBCNodePacketSemanticsEvidence{
		Role: role, SuccessSequence: relay.Packet.Sequence, TimeoutSequence: timeoutEvidence.Packet.Sequence,
	}
	switch role {
	case "source":
		var err error
		_, evidence.SuccessCommitmentPresent, err = queryPacketCommitmentOnNode(ctx, node, endpoint, relay.Packet.Sequence)
		if err != nil {
			return evidence, err
		}
		_, evidence.TimeoutCommitmentPresent, err = queryPacketCommitmentOnNode(ctx, node, endpoint, timeoutEvidence.Packet.Sequence)
		if err != nil {
			return evidence, err
		}
		if evidence.SuccessCommitmentPresent || evidence.TimeoutCommitmentPresent {
			return evidence, errors.New("IBC source node retained a terminal packet commitment")
		}
	case "destination":
		var err error
		evidence.SuccessReceipt, err = queryPacketReceiptOnNode(ctx, node, endpoint, relay.Packet.Sequence)
		if err != nil {
			return evidence, err
		}
		evidence.TimeoutReceipt, err = queryPacketReceiptOnNode(ctx, node, endpoint, timeoutEvidence.Packet.Sequence)
		if err != nil {
			return evidence, err
		}
		successAck, successAckExists, err := queryPacketAcknowledgementOnNode(ctx, node, endpoint, relay.Packet.Sequence)
		if err != nil {
			return evidence, err
		}
		timeoutAck, timeoutAckExists, err := queryPacketAcknowledgementOnNode(ctx, node, endpoint, timeoutEvidence.Packet.Sequence)
		if err != nil {
			return evidence, err
		}
		if !successAckExists || timeoutAckExists {
			return evidence, errors.New("IBC destination node acknowledgement terminal semantics are invalid")
		}
		evidence.SuccessAcknowledgement = base64IfNotEmpty(successAck)
		evidence.TimeoutAcknowledgement = base64IfNotEmpty(timeoutAck)
	default:
		return evidence, fmt.Errorf("unknown IBC node packet role %q", role)
	}
	return evidence, nil
}

func queryPacketCommitmentOnNode(ctx context.Context, node *cosmos.ChainNode, endpoint IBCChannelEndpoint, sequence uint64) ([]byte, bool, error) {
	response, err := channeltypes.NewQueryClient(node.GrpcConn).PacketCommitment(ctx, &channeltypes.QueryPacketCommitmentRequest{
		PortId: endpoint.PortID, ChannelId: endpoint.ChannelID, Sequence: sequence,
	})
	if status.Code(err) == codes.NotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if response == nil || len(response.Commitment) == 0 {
		return nil, false, errors.New("IBC node packet commitment response is empty")
	}
	return response.Commitment, true, nil
}

func queryPacketReceiptOnNode(ctx context.Context, node *cosmos.ChainNode, endpoint IBCChannelEndpoint, sequence uint64) (bool, error) {
	response, err := channeltypes.NewQueryClient(node.GrpcConn).PacketReceipt(ctx, &channeltypes.QueryPacketReceiptRequest{
		PortId: endpoint.PortID, ChannelId: endpoint.ChannelID, Sequence: sequence,
	})
	if err != nil || response == nil {
		return false, errors.Join(err, errors.New("IBC node packet receipt response is empty"))
	}
	return response.Received, nil
}

func queryPacketAcknowledgementOnNode(ctx context.Context, node *cosmos.ChainNode, endpoint IBCChannelEndpoint, sequence uint64) ([]byte, bool, error) {
	response, err := channeltypes.NewQueryClient(node.GrpcConn).PacketAcknowledgement(ctx, &channeltypes.QueryPacketAcknowledgementRequest{
		PortId: endpoint.PortID, ChannelId: endpoint.ChannelID, Sequence: sequence,
	})
	if status.Code(err) == codes.NotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if response == nil || len(response.Acknowledgement) == 0 {
		return nil, false, errors.New("IBC node packet acknowledgement response is empty")
	}
	return response.Acknowledgement, true, nil
}

func counterpartyBaseDenom(channel IBCChannelHandshake, chainID string) string {
	if chainID == channel.Panacea.ChainID {
		return "uosmo"
	}
	if chainID == channel.Osmosis.ChainID {
		return "umed"
	}
	return ""
}

func queryIBCDenomTraceOnNode(
	ctx context.Context,
	chainID string,
	node *cosmos.ChainNode,
	endpoint IBCChannelEndpoint,
	baseDenom string,
) (IBCDenomTraceEvidence, error) {
	voucherDenom := ibcVoucherDenom(endpoint, baseDenom)
	hash := strings.TrimPrefix(voucherDenom, "ibc/")
	response, err := transfertypes.NewQueryClient(node.GrpcConn).DenomTrace(
		ctx, &transfertypes.QueryDenomTraceRequest{Hash: hash},
	)
	if err != nil || response == nil || response.DenomTrace == nil {
		return IBCDenomTraceEvidence{}, errors.Join(err, errors.New("IBC node denom trace response is empty"))
	}
	return IBCDenomTraceEvidence{
		ChainID: chainID, VoucherDenom: voucherDenom, Hash: hash,
		Path: response.DenomTrace.Path, BaseDenom: response.DenomTrace.BaseDenom,
	}, nil
}

func packetFromSendEvidence(send IBCPacketSendEvidence) (ibc.Packet, error) {
	data, err := base64.StdEncoding.DecodeString(send.PacketData)
	if err != nil {
		return ibc.Packet{}, err
	}
	return ibc.Packet{
		Sequence:         send.Sequence,
		SourcePort:       send.SourcePort,
		SourceChannel:    send.SourceChannel,
		DestPort:         send.DestinationPort,
		DestChannel:      send.DestinationChannel,
		Data:             data,
		TimeoutHeight:    send.TimeoutHeight,
		TimeoutTimestamp: ibc.Nanoseconds(send.TimeoutTimestamp),
	}, nil
}

func (n *IBCTopology) queryIBCLinkState(ctx context.Context, phase string) (IBCLinkStateSnapshot, error) {
	var zero IBCLinkStateSnapshot
	if n.channel == nil {
		return zero, errors.New("IBC transfer channel is not open")
	}
	handshake := *n.channel
	panaceaEndpoint, err := snapshotIBCChannelEndpoint(
		ctx, n.Relayer, n.execReporter, handshake.Panacea.ChainID, handshake.Osmosis.ChainID,
	)
	if err != nil {
		return zero, fmt.Errorf("query Panacea channel endpoint: %w", err)
	}
	osmosisEndpoint, err := snapshotIBCChannelEndpoint(
		ctx, n.Relayer, n.execReporter, handshake.Osmosis.ChainID, handshake.Panacea.ChainID,
	)
	if err != nil {
		return zero, fmt.Errorf("query Osmosis channel endpoint: %w", err)
	}
	observed := IBCChannelHandshake{Path: n.Path, Panacea: panaceaEndpoint, Osmosis: osmosisEndpoint}
	panaceaStatus, err := queryIBCClientStatus(ctx, n.Panacea, observed.Panacea.ClientID)
	if err != nil {
		return zero, fmt.Errorf("query Panacea client status: %w", err)
	}
	osmosisStatus, err := queryIBCClientStatus(ctx, n.Osmosis, observed.Osmosis.ClientID)
	if err != nil {
		return zero, fmt.Errorf("query Osmosis client status: %w", err)
	}
	panaceaSequence, err := queryIBCNextSequenceSend(ctx, n.Panacea, observed.Panacea)
	if err != nil {
		return zero, fmt.Errorf("query Panacea next send sequence: %w", err)
	}
	osmosisSequence, err := queryIBCNextSequenceSend(ctx, n.Osmosis, observed.Osmosis)
	if err != nil {
		return zero, fmt.Errorf("query Osmosis next send sequence: %w", err)
	}
	panaceaHeight, osmosisHeight, err := n.chainHeights(ctx)
	if err != nil {
		return zero, err
	}
	snapshot := IBCLinkStateSnapshot{
		Phase:                   phase,
		Channel:                 observed,
		PanaceaClientStatus:     panaceaStatus,
		OsmosisClientStatus:     osmosisStatus,
		PanaceaHeight:           panaceaHeight,
		OsmosisHeight:           osmosisHeight,
		PanaceaNextSequenceSend: panaceaSequence,
		OsmosisNextSequenceSend: osmosisSequence,
	}
	if err := snapshot.ValidateAgainst(handshake); err != nil {
		return zero, err
	}
	return snapshot, nil
}

func (n *IBCTopology) chainHeights(ctx context.Context) (int64, int64, error) {
	panaceaHeight, err := n.Panacea.Height(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("query Panacea height: %w", err)
	}
	osmosisHeight, err := n.Osmosis.Height(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("query Osmosis height: %w", err)
	}
	return panaceaHeight, osmosisHeight, nil
}

func queryIBCClientStatus(ctx context.Context, chain *cosmos.CosmosChain, clientID string) (string, error) {
	response, err := clienttypes.NewQueryClient(chain.GetFullNode().GrpcConn).ClientStatus(ctx, &clienttypes.QueryClientStatusRequest{ClientId: clientID})
	if err != nil {
		return "", err
	}
	if response == nil || strings.TrimSpace(response.Status) == "" {
		return "", errors.New("IBC client status response is empty")
	}
	return response.Status, nil
}

func queryIBCNextSequenceSend(ctx context.Context, chain *cosmos.CosmosChain, endpoint IBCChannelEndpoint) (uint64, error) {
	// NextSequenceSend was added to the channel gRPC service in ibc-go v8.
	// Query the version-stable ICS-04 store key so the same observation works
	// against both the v2.2.1 (ibc-go v7) node and the upgraded v8 node.
	result, err := chain.GetFullNode().Client.ABCIQuery(
		ctx,
		"store/ibc/key",
		host.NextSequenceSendKey(endpoint.PortID, endpoint.ChannelID),
	)
	if err != nil {
		return 0, err
	}
	if result == nil {
		return 0, errors.New("IBC next send sequence ABCI response is empty")
	}
	if result.Response.IsErr() {
		return 0, fmt.Errorf(
			"IBC next send sequence ABCI query failed with code %d: %s",
			result.Response.Code,
			result.Response.Log,
		)
	}
	return decodeIBCNextSequenceSend(result.Response.Value)
}

func decodeIBCNextSequenceSend(value []byte) (uint64, error) {
	if len(value) != 8 {
		return 0, fmt.Errorf("IBC next send sequence store value has %d bytes, want 8", len(value))
	}
	sequence := binary.BigEndian.Uint64(value)
	if sequence == 0 {
		return 0, errors.New("IBC next send sequence store value is zero")
	}
	return sequence, nil
}

func queryPacketCommitment(ctx context.Context, chain *cosmos.CosmosChain, endpoint IBCChannelEndpoint, sequence uint64) ([]byte, bool, error) {
	response, err := channeltypes.NewQueryClient(chain.GetFullNode().GrpcConn).PacketCommitment(ctx, &channeltypes.QueryPacketCommitmentRequest{
		PortId: endpoint.PortID, ChannelId: endpoint.ChannelID, Sequence: sequence,
	})
	if status.Code(err) == codes.NotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if response == nil || len(response.Commitment) == 0 {
		return nil, false, errors.New("packet commitment query returned an empty commitment")
	}
	return append([]byte(nil), response.Commitment...), true, nil
}

func queryPacketReceipt(ctx context.Context, chain *cosmos.CosmosChain, endpoint IBCChannelEndpoint, sequence uint64) (bool, error) {
	response, err := channeltypes.NewQueryClient(chain.GetFullNode().GrpcConn).PacketReceipt(ctx, &channeltypes.QueryPacketReceiptRequest{
		PortId: endpoint.PortID, ChannelId: endpoint.ChannelID, Sequence: sequence,
	})
	if err != nil {
		return false, err
	}
	if response == nil {
		return false, errors.New("packet receipt query returned no response")
	}
	return response.Received, nil
}

func queryPacketAcknowledgement(ctx context.Context, chain *cosmos.CosmosChain, endpoint IBCChannelEndpoint, sequence uint64) ([]byte, bool, error) {
	response, err := channeltypes.NewQueryClient(chain.GetFullNode().GrpcConn).PacketAcknowledgement(ctx, &channeltypes.QueryPacketAcknowledgementRequest{
		PortId: endpoint.PortID, ChannelId: endpoint.ChannelID, Sequence: sequence,
	})
	if status.Code(err) == codes.NotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if response == nil || len(response.Acknowledgement) == 0 {
		return nil, false, errors.New("packet acknowledgement query returned an empty acknowledgement")
	}
	return append([]byte(nil), response.Acknowledgement...), true, nil
}
