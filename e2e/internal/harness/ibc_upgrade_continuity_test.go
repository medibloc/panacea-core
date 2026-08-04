package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

func TestIBCPacketTerminalPollMaxHeightWaitsBeyondCurrentTip(t *testing.T) {
	const (
		stagedScanStart          = int64(117)
		restartBeforeOsmosis     = int64(238)
		committedOsmosisRecv     = int64(240)
		stagedPanaceaSend        = int64(149)
		restartBeforePanacea     = int64(232)
		committedPanaceaAck      = int64(236)
		staleOsmosisReceiveLimit = stagedScanStart + ibcPacketTerminalWaitMax
	)

	if committedOsmosisRecv <= staleOsmosisReceiveLimit {
		t.Fatalf("regression fixture recv height %d must exceed stale max %d", committedOsmosisRecv, staleOsmosisReceiveLimit)
	}
	recvMax := ibcPacketTerminalPollMaxHeight(stagedScanStart, restartBeforeOsmosis)
	if got, want := recvMax, int64(358); got != want {
		t.Fatalf("post-upgrade terminal poll max height = %d, want %d", got, want)
	}
	if committedOsmosisRecv > recvMax {
		t.Fatalf("committed recv height %d is outside repaired max %d", committedOsmosisRecv, recvMax)
	}
	ackMax := ibcPacketTerminalPollMaxHeight(stagedPanaceaSend, restartBeforePanacea)
	if committedPanaceaAck > ackMax {
		t.Fatalf("committed ack height %d is outside repaired max %d", committedPanaceaAck, ackMax)
	}
	if got, want := ibcPacketTerminalPollMaxHeight(stagedScanStart, 100), int64(237); got != want {
		t.Fatalf("pre-start terminal poll max height = %d, want %d", got, want)
	}
}

func TestRelativeIBCTimestampTransferOptionsUseDurationWithoutAbsoluteHeightSideEffect(t *testing.T) {
	const timeout = 15 * time.Second

	options, err := relativeIBCTimestampTransferOptions(timeout)
	if err != nil {
		t.Fatalf("build relative timeout options: %v", err)
	}
	if options.Timeout == nil {
		t.Fatal("relative IBC timestamp options omitted the timeout")
	}
	if got, want := options.Timeout.NanoSeconds, uint64(timeout.Nanoseconds()); got != want {
		t.Fatalf("relative timeout nanoseconds = %d, want %d", got, want)
	}
	if options.AbsoluteTimeouts {
		t.Fatal("relative timestamp options unexpectedly made the CLI default height 0-1000 absolute")
	}
	if _, err := relativeIBCTimestampTransferOptions(0); err == nil {
		t.Fatal("zero relative IBC timeout unexpectedly accepted")
	}
}

func TestCommittedPacketTimeoutTimestampUsesObservedSendEvent(t *testing.T) {
	const observedTimestamp = uint64(1_785_860_689_017_387_000)
	tx := ibc.Tx{Packet: ibc.Packet{TimeoutTimestamp: ibc.Nanoseconds(observedTimestamp)}}

	got, err := committedPacketTimeoutTimestamp(tx)
	if err != nil {
		t.Fatalf("read committed packet timeout: %v", err)
	}
	if got != observedTimestamp {
		t.Fatalf("committed timeout timestamp = %d, want %d", got, observedTimestamp)
	}
	if _, err := committedPacketTimeoutTimestamp(ibc.Tx{}); err == nil {
		t.Fatal("missing committed packet timeout unexpectedly accepted")
	}
}

func TestValidateInFlightPreRelayBalancesUsesSnapshotBeforeHermesStarts(t *testing.T) {
	checkpoint := IBCInFlightPacketCheckpoint{
		SourceNativeBalance:       IBCBalanceEvidence{After: "8900"},
		DestinationVoucherBalance: IBCBalanceEvidence{After: "5"},
	}

	if err := validateInFlightPreRelayBalances(checkpoint, sdkmath.NewInt(8900), sdkmath.NewInt(5)); err != nil {
		t.Fatalf("unchanged pre-relay balances rejected: %v", err)
	}
	if err := validateInFlightPreRelayBalances(checkpoint, sdkmath.NewInt(8900), sdkmath.NewInt(1005)); err == nil {
		t.Fatal("already-relayed destination balance unexpectedly accepted as a pre-relay snapshot")
	}
	if err := validateInFlightPreRelayBalances(checkpoint, sdkmath.NewInt(8899), sdkmath.NewInt(5)); err == nil {
		t.Fatal("changed source balance unexpectedly accepted as a pre-relay snapshot")
	}
}

func TestDecodeIBCNextSequenceSendSupportsVersionNeutralStoreValue(t *testing.T) {
	sequence, err := decodeIBCNextSequenceSend([]byte{0, 0, 0, 0, 0, 0, 0, 3})
	if err != nil {
		t.Fatalf("valid sequence rejected: %v", err)
	}
	if sequence != 3 {
		t.Fatalf("sequence = %d, want 3", sequence)
	}

	for _, value := range [][]byte{nil, {0, 0, 0, 1}, {0, 0, 0, 0, 0, 0, 0, 0}} {
		if _, err := decodeIBCNextSequenceSend(value); err == nil {
			t.Fatalf("invalid sequence value %x unexpectedly decoded", value)
		}
	}
}

func TestIBCUpgradeGRPCQueryRecordsCoverLiveLinkPacketAndAcknowledgementBoundaries(t *testing.T) {
	t.Parallel()

	handshake := validIBCChannelHandshake()
	evidence := IBCUpgradeContinuityEvidence{
		OriginalChannel: handshake,
		NodeRestartSemantics: []IBCChainRestartEvidence{{
			ChainID: handshake.Osmosis.ChainID,
			Nodes: []IBCNodePostRestartEvidence{{
				ChainID:         handshake.Osmosis.ChainID,
				IdentityAfter:   IBCNodeRuntimeIdentity{Name: "osmosis-full-0"},
				ObservedHeight:  80,
				ClientStatus:    "Active",
				ConnectionState: "Open",
				ChannelState:    "Open",
				PacketSemantics: IBCNodePacketSemanticsEvidence{
					Role: "destination", SuccessSequence: 2, TimeoutSequence: 3,
					SuccessReceipt: true, TimeoutReceipt: false,
					SuccessAcknowledgement: "AQ==",
				},
			}},
		}},
	}

	records, err := buildIBCUpgradeGRPCQueryRecords(evidence)
	if err != nil {
		t.Fatalf("build IBC gRPC records: %v", err)
	}
	if len(records) != 5 {
		t.Fatalf("gRPC record count = %d, want 5", len(records))
	}
	steps := make(map[string]struct{}, len(records))
	for _, record := range records {
		if record.Boundary != "grpc" || record.Step == "" || record.Height != 0 || record.HistoricalHeight {
			t.Fatalf("invalid gRPC record: %#v", record)
		}
		if record.Request == nil || record.Response == nil || record.Metadata == nil {
			t.Fatalf("gRPC record lacks structured transport evidence: %#v", record)
		}
		steps[record.Step] = struct{}{}
	}
	coverage := IBCUpgradeGRPCQueryCoverage()
	if !coverage.Supported || !coverage.Exercised || len(coverage.Evidence) != len(records) {
		t.Fatalf("gRPC coverage = %#v", coverage)
	}
	runDir := t.TempDir()
	queryPath := filepath.Join(runDir, filepath.FromSlash(ibcUpgradeGRPCQueryArtifactPath))
	if err := os.MkdirAll(filepath.Dir(queryPath), 0o700); err != nil {
		t.Fatalf("create query artifact directory: %v", err)
	}
	queryFile, err := os.OpenFile(queryPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("create query artifact: %v", err)
	}
	encoder := json.NewEncoder(queryFile)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			t.Fatalf("encode query artifact: %v", err)
		}
	}
	if err := queryFile.Close(); err != nil {
		t.Fatalf("close query artifact: %v", err)
	}
	for _, reference := range coverage.Evidence {
		if _, ok := steps[reference.Step]; !ok {
			t.Fatalf("coverage references unrecorded step %q", reference.Step)
		}
		if err := validateUpgradeCoverageQueryEvidence(runDir, reference, false); err != nil {
			t.Fatalf("coverage merger rejected %q: %v", reference.Step, err)
		}
	}
}

func TestIBCUpgradeContinuityRequiresSameChannelSequenceAndExactlyOnceRelay(t *testing.T) {
	now := time.Now().UTC()
	handshake := validIBCChannelHandshake()
	before := validIBCLinkStateSnapshot("before-in-flight", handshake, 10, 20, 2, 2)
	afterSend := validIBCLinkStateSnapshot("in-flight-staged", handshake, 12, 22, 3, 2)
	postUpgrade := validIBCLinkStateSnapshot("post-upgrade-before-relay", handshake, 50, 60, 3, 2)
	afterRelay := validIBCLinkStateSnapshot("post-upgrade-after-in-flight-relay", handshake, 54, 64, 3, 2)
	afterTimeout := validIBCLinkStateSnapshot("post-upgrade-after-timeout", handshake, 58, 68, 4, 2)
	final := validIBCLinkStateSnapshot("post-upgrade-after-node-restart", handshake, 70, 80, 5, 3)

	inFlightSend := IBCPacketSendEvidence{
		Direction:          ibcPanaceaToOsmosis,
		SourceChainID:      handshake.Panacea.ChainID,
		DestinationChainID: handshake.Osmosis.ChainID,
		TxHash:             "AABB",
		TxHeight:           11,
		Sequence:           2,
		SourcePort:         handshake.Panacea.PortID,
		SourceChannel:      handshake.Panacea.ChannelID,
		DestinationPort:    handshake.Osmosis.PortID,
		DestinationChannel: handshake.Osmosis.ChannelID,
		Denom:              "umed",
		Amount:             "1000",
		GasFee:             "100",
		PacketData:         "AQID",
		TimeoutHeight:      "0-0",
		TimeoutTimestamp:   999999,
	}
	inFlightLifecycle := IBCPacketLifecycleEvidence{
		Direction:          inFlightSend.Direction,
		SourceChainID:      inFlightSend.SourceChainID,
		DestinationChainID: inFlightSend.DestinationChainID,
		TxHash:             inFlightSend.TxHash,
		TxHeight:           inFlightSend.TxHeight,
		Sequence:           inFlightSend.Sequence,
		SourcePort:         inFlightSend.SourcePort,
		SourceChannel:      inFlightSend.SourceChannel,
		DestinationPort:    inFlightSend.DestinationPort,
		DestinationChannel: inFlightSend.DestinationChannel,
		Denom:              inFlightSend.Denom,
		Amount:             inFlightSend.Amount,
		Recv:               IBCPacketObservation{Observed: true, ChainID: handshake.Osmosis.ChainID, Height: 61},
		Ack:                IBCAcknowledgementObservation{Observed: true, ChainID: handshake.Panacea.ChainID, Height: 52, Acknowledgement: "eyJyZXN1bHQiOiJBUT09In0="},
	}

	evidence := IBCUpgradeContinuityEvidence{
		Phase:           "v2.2.1-to-current",
		OriginalChannel: handshake,
		InFlight: IBCInFlightPacketCheckpoint{
			Phase:                      "in-flight-staged",
			Channel:                    handshake,
			BeforeSendState:            before,
			AfterSendState:             afterSend,
			Packet:                     inFlightSend,
			Commitment:                 "qrvM3Q==",
			DestinationReceipt:         false,
			DestinationAcknowledgement: "",
			DestinationScanStartHeight: 20,
			SourceNativeBalance:        IBCBalanceEvidence{ChainID: handshake.Panacea.ChainID, Address: "panacea1user", Denom: "umed", Before: "10000", After: "8900", ExpectedAfter: "8900"},
			DestinationVoucherBalance:  IBCBalanceEvidence{ChainID: handshake.Osmosis.ChainID, Address: "osmo1user", Denom: "ibc/MED", Before: "5", After: "5", ExpectedAfter: "5"},
			SourceEscrowLock: IBCEscrowBalanceEvidence{
				Phase: "in-flight-staged", ChainID: handshake.Panacea.ChainID, PortID: handshake.Panacea.PortID, ChannelID: handshake.Panacea.ChannelID,
				Address: "panacea1escrow", Denom: "umed", Before: "100", After: "1100", ExpectedDelta: "1000", ExpectedAfter: "1100",
			},
		},
		PanaceaUpgrade: IBCPanaceaUpgradeStepEvidence{
			CallbackCompleted: true,
			UpgradeHeight:     40,
			BeforeHeight:      39,
			AfterHeight:       44,
			From:              V221Image(),
			To:                CurrentImage(),
			PanaceaBefore: []IBCNodeRuntimeIdentity{
				{Name: "panacea-val-0", ContainerID: "old-val", Image: V221Image()},
				{Name: "panacea-full-0", ContainerID: "old-full", Image: V221Image()},
			},
			PanaceaAfter: []IBCNodeRuntimeIdentity{
				{Name: "panacea-val-0", ContainerID: "new-val", Image: CurrentImage()},
				{Name: "panacea-full-0", ContainerID: "new-full", Image: CurrentImage()},
			},
			OsmosisBefore: []IBCNodeRuntimeIdentity{
				{Name: "osmosis-val-0", ContainerID: "same-osmo", Image: imageRefFromDocker(PinnedOsmosisImage())},
			},
			OsmosisAfter: []IBCNodeRuntimeIdentity{
				{Name: "osmosis-val-0", ContainerID: "same-osmo", Image: imageRefFromDocker(PinnedOsmosisImage())},
			},
			OsmosisProgress: IBCHeightProgressEvidence{
				ChainID: handshake.Osmosis.ChainID, StartedAt: now, CompletedAt: now.Add(2 * time.Second),
				StartHeight: 20, EndHeight: 22, MaxNoProgressMillis: 0, BoundMillis: 15_000,
				Samples: []IBCHeightSample{{ObservedAt: now, Height: 20}, {ObservedAt: now.Add(2 * time.Second), Height: 22}},
			},
		},
		PostUpgradeBeforeRelay: postUpgrade,
		InFlightRelay: IBCInFlightRelayEvidence{
			Packet:                     inFlightLifecycle,
			ReceiveCount:               1,
			AcknowledgementCount:       1,
			CommitmentCleared:          true,
			DestinationReceipt:         true,
			DestinationAcknowledgement: "CPdVftUYJv4Y2EUSvyTsdQAe268hI6R333KgqfNkCnw=",
			SourceNativeBalance:        IBCBalanceEvidence{ChainID: handshake.Panacea.ChainID, Address: "panacea1user", Denom: "umed", Before: "8900", After: "8900", ExpectedAfter: "8900"},
			DestinationVoucherBalance:  IBCBalanceEvidence{ChainID: handshake.Osmosis.ChainID, Address: "osmo1user", Denom: "ibc/MED", Before: "5", After: "1005", ExpectedAfter: "1005"},
		},
		AfterInFlightRelay: afterRelay,
		Timeout: IBCPacketTimeoutEvidence{
			Phase:             "post-upgrade-timeout",
			BeforeSendState:   afterRelay,
			AfterTimeoutState: afterTimeout,
			Packet: IBCPacketSendEvidence{
				Direction: ibcPanaceaToOsmosis, SourceChainID: handshake.Panacea.ChainID, DestinationChainID: handshake.Osmosis.ChainID,
				TxHash: "1122", TxHeight: 55, Sequence: 3, SourcePort: handshake.Panacea.PortID, SourceChannel: handshake.Panacea.ChannelID,
				DestinationPort: handshake.Osmosis.PortID, DestinationChannel: handshake.Osmosis.ChannelID, Denom: "umed", Amount: "1000", GasFee: "100",
				PacketData: "AQID", TimeoutHeight: "0-0", TimeoutTimestamp: 999999,
			},
			TimeoutCount:               1,
			ReceiveCount:               0,
			CommitmentCleared:          true,
			DestinationReceipt:         false,
			DestinationAcknowledgement: "",
			SourceNativeBalance:        IBCBalanceEvidence{ChainID: handshake.Panacea.ChainID, Address: "panacea1user", Denom: "umed", Before: "8900", After: "8800", ExpectedAfter: "8800"},
			DestinationVoucherBalance:  IBCBalanceEvidence{ChainID: handshake.Osmosis.ChainID, Address: "osmo1user", Denom: "ibc/MED", Before: "1005", After: "1005", ExpectedAfter: "1005"},
			SourceEscrowLock: IBCEscrowBalanceEvidence{
				Phase: "post-upgrade-timeout-lock", ChainID: handshake.Panacea.ChainID, PortID: handshake.Panacea.PortID, ChannelID: handshake.Panacea.ChannelID,
				Address: "panacea1escrow", Denom: "umed", Before: "1100", After: "2100", ExpectedDelta: "1000", ExpectedAfter: "2100",
			},
			SourceEscrowRefund: IBCEscrowBalanceEvidence{
				Phase: "post-upgrade-timeout-refund", ChainID: handshake.Panacea.ChainID, PortID: handshake.Panacea.PortID, ChannelID: handshake.Panacea.ChannelID,
				Address: "panacea1escrow", Denom: "umed", Before: "2100", After: "1100", ExpectedDelta: "-1000", ExpectedAfter: "1100",
			},
		},
		PostUpgradeTransfers:    validPostUpgradeBidirectionalEvidence(handshake),
		FinalAfterHermesRestart: final,
		HermesRestarts: []IBCHermesRestartEvidence{
			{Phase: "post-upgrade-relay", PanaceaBeforeHeight: 50, OsmosisBeforeHeight: 60, PanaceaAfterHeight: 54, OsmosisAfterHeight: 64, HealthCheckCompleted: true},
			{Phase: "post-transfer-restart", PanaceaBeforeHeight: 66, OsmosisBeforeHeight: 76, PanaceaAfterHeight: 68, OsmosisAfterHeight: 78, HealthCheckCompleted: true},
		},
		PanaceaNodeRestarts: []NodeRestartEvidence{
			{
				RecordedAt: time.Now().UTC(), Mode: "graceful", Node: "panacea-full-0",
				Before: BlockEvidence{Node: "panacea-full-0", Height: 68, StateHeight: 67, BlockID: "AA", AppHash: "BB"},
				After:  BlockEvidence{Node: "panacea-full-0", Height: 69, StateHeight: 68, BlockID: "CC", AppHash: "DD"},
			},
			{
				RecordedAt: time.Now().UTC(), Mode: "graceful", Node: "panacea-val-0",
				Before: BlockEvidence{Node: "panacea-val-0", Height: 69, StateHeight: 68, BlockID: "CC", AppHash: "DD"},
				After:  BlockEvidence{Node: "panacea-val-0", Height: 70, StateHeight: 69, BlockID: "EE", AppHash: "FF"},
			},
		},
		OsmosisNodeRestarts: []NodeRestartEvidence{
			{
				RecordedAt: now, Mode: "graceful", Node: "osmosis-val-0",
				Before: BlockEvidence{Node: "osmosis-val-0", Height: 79, StateHeight: 78, BlockID: "OA", AppHash: "OB"},
				After:  BlockEvidence{Node: "osmosis-val-0", Height: 80, StateHeight: 79, BlockID: "OC", AppHash: "OD"},
			},
		},
		FinalPacketStates: []IBCPacketTerminalStateEvidence{
			{Kind: "success", Sequence: 2, Commitment: "", DestinationReceipt: true, DestinationAcknowledgement: "CPdVftUYJv4Y2EUSvyTsdQAe268hI6R333KgqfNkCnw="},
			{Kind: "timeout", Sequence: 3, Commitment: "", DestinationReceipt: false, DestinationAcknowledgement: ""},
		},
		FinalDenomTraces: []IBCDenomTraceEvidence{
			{ChainID: handshake.Panacea.ChainID, VoucherDenom: "ibc/OSMO", Hash: "OSMO", Path: handshake.Panacea.PortID + "/" + handshake.Panacea.ChannelID, BaseDenom: "uosmo"},
			{ChainID: handshake.Osmosis.ChainID, VoucherDenom: "ibc/MED", Hash: "MED", Path: handshake.Osmosis.PortID + "/" + handshake.Osmosis.ChannelID, BaseDenom: "umed"},
		},
		FinalBalances: []IBCBalanceEvidence{
			{ChainID: handshake.Panacea.ChainID, Address: "panacea1user", Denom: "umed", Before: "7700", After: "7700", ExpectedAfter: "7700"},
			{ChainID: handshake.Panacea.ChainID, Address: "panacea1user", Denom: "ibc/OSMO", Before: "1005", After: "1005", ExpectedAfter: "1005"},
			{ChainID: handshake.Osmosis.ChainID, Address: "osmo1user", Denom: "uosmo", Before: "8990", After: "8990", ExpectedAfter: "8990"},
			{ChainID: handshake.Osmosis.ChainID, Address: "osmo1user", Denom: "ibc/MED", Before: "2005", After: "2005", ExpectedAfter: "2005"},
		},
		FinalEscrowBalances: []IBCEscrowBalanceEvidence{
			{Phase: "post-restart", ChainID: handshake.Panacea.ChainID, PortID: handshake.Panacea.PortID, ChannelID: handshake.Panacea.ChannelID, Address: "panacea1escrow", Denom: "umed", Before: "1100", After: "1100", ExpectedDelta: "0", ExpectedAfter: "1100"},
			{Phase: "post-restart", ChainID: handshake.Osmosis.ChainID, PortID: handshake.Osmosis.PortID, ChannelID: handshake.Osmosis.ChannelID, Address: "osmo1escrow", Denom: "uosmo", Before: "1200", After: "1200", ExpectedDelta: "0", ExpectedAfter: "1200"},
		},
	}
	evidence.DenomTraceContinuity = []IBCDenomTraceSnapshot{
		validIBCDenomTraceSnapshot(handshake, ibcPhasePreUpgrade),
		validIBCDenomTraceSnapshot(handshake, ibcPhasePostUpgrade),
		validIBCDenomTraceSnapshot(handshake, "post-restart"),
	}
	evidence.NodeRestartSemantics = validIBCNodeRestartSemantics(evidence)

	if err := evidence.Validate(); err != nil {
		t.Fatalf("valid upgrade continuity evidence rejected: %v", err)
	}

	replacement := evidence
	replacement.FinalAfterHermesRestart.Channel.Panacea.ChannelID = "channel-9"
	if err := replacement.Validate(); err == nil {
		t.Fatal("evidence that replaces the original channel unexpectedly validated")
	}

	duplicateRelay := evidence
	duplicateRelay.InFlightRelay.ReceiveCount = 2
	if err := duplicateRelay.Validate(); err == nil {
		t.Fatal("evidence with duplicate in-flight receive unexpectedly validated")
	}

	brokenSequence := evidence
	brokenSequence.PostUpgradeTransfers.Transfers[0].Sequence = 9
	if err := brokenSequence.Validate(); err == nil {
		t.Fatal("evidence with a discontinuous post-upgrade sequence unexpectedly validated")
	}

	missingTimeout := evidence
	missingTimeout.Timeout.TimeoutCount = 0
	if err := missingTimeout.Validate(); err == nil {
		t.Fatal("evidence without exactly one timeout unexpectedly validated")
	}

	receivedTimeout := evidence
	receivedTimeout.Timeout.ReceiveCount = 1
	if err := receivedTimeout.Validate(); err == nil {
		t.Fatal("evidence whose timeout packet was also received unexpectedly validated")
	}

	changedAfterRestart := evidence
	changedAfterRestart.FinalBalances[0].After = "7699"
	if err := changedAfterRestart.Validate(); err == nil {
		t.Fatal("evidence whose balance changed after node restart unexpectedly validated")
	}

	brokenEscrowRefund := evidence
	brokenEscrowRefund.Timeout.SourceEscrowRefund.After = "1101"
	if err := brokenEscrowRefund.Validate(); err == nil {
		t.Fatal("evidence with an incomplete escrow refund unexpectedly validated")
	}

	stalledOsmosis := evidence
	stalledOsmosis.PanaceaUpgrade.OsmosisProgress.EndHeight = stalledOsmosis.PanaceaUpgrade.OsmosisProgress.StartHeight
	stalledOsmosis.PanaceaUpgrade.OsmosisProgress.Samples[1].Height = stalledOsmosis.PanaceaUpgrade.OsmosisProgress.StartHeight
	if err := stalledOsmosis.Validate(); err == nil {
		t.Fatal("evidence whose Osmosis chain did not progress during upgrade unexpectedly validated")
	}

	changedTrace := evidence
	changedTrace.DenomTraceContinuity = append([]IBCDenomTraceSnapshot(nil), evidence.DenomTraceContinuity...)
	changedTrace.DenomTraceContinuity[1].Traces = append([]IBCDenomTraceEvidence(nil), evidence.DenomTraceContinuity[1].Traces...)
	changedTrace.DenomTraceContinuity[1].Traces[0].Hash = "CHANGED"
	if err := changedTrace.Validate(); err == nil {
		t.Fatal("evidence whose denom trace changed across upgrade unexpectedly validated")
	}

	missingOsmosisRestart := evidence
	missingOsmosisRestart.OsmosisNodeRestarts = nil
	if err := missingOsmosisRestart.Validate(); err == nil {
		t.Fatal("evidence without every Osmosis node restart unexpectedly validated")
	}
}

func validIBCLinkStateSnapshot(
	phase string,
	handshake IBCChannelHandshake,
	panaceaHeight int64,
	osmosisHeight int64,
	panaceaNext uint64,
	osmosisNext uint64,
) IBCLinkStateSnapshot {
	return IBCLinkStateSnapshot{
		Phase:                   phase,
		Channel:                 handshake,
		PanaceaClientStatus:     "Active",
		OsmosisClientStatus:     "Active",
		PanaceaHeight:           panaceaHeight,
		OsmosisHeight:           osmosisHeight,
		PanaceaNextSequenceSend: panaceaNext,
		OsmosisNextSequenceSend: osmosisNext,
	}
}

func validPostUpgradeBidirectionalEvidence(handshake IBCChannelHandshake) IBCPostUpgradeTransferEvidence {
	return IBCPostUpgradeTransferEvidence{
		Phase:   "post-upgrade",
		Channel: handshake,
		Transfers: []IBCPacketLifecycleEvidence{
			{
				Direction: ibcPanaceaToOsmosis, SourceChainID: handshake.Panacea.ChainID, DestinationChainID: handshake.Osmosis.ChainID,
				TxHash: "CCDD", TxHeight: 60, Sequence: 4, SourcePort: "transfer", SourceChannel: handshake.Panacea.ChannelID,
				DestinationPort: "transfer", DestinationChannel: handshake.Osmosis.ChannelID, Denom: "umed", Amount: "1000",
				Recv: IBCPacketObservation{Observed: true, ChainID: handshake.Osmosis.ChainID, Height: 65},
				Ack:  IBCAcknowledgementObservation{Observed: true, ChainID: handshake.Panacea.ChainID, Height: 66, Acknowledgement: "eyJyZXN1bHQiOiJBUT09In0="},
			},
			{
				Direction: ibcOsmosisToPanacea, SourceChainID: handshake.Osmosis.ChainID, DestinationChainID: handshake.Panacea.ChainID,
				TxHash: "EEFF", TxHeight: 62, Sequence: 2, SourcePort: "transfer", SourceChannel: handshake.Osmosis.ChannelID,
				DestinationPort: "transfer", DestinationChannel: handshake.Panacea.ChannelID, Denom: "uosmo", Amount: "1000",
				Recv: IBCPacketObservation{Observed: true, ChainID: handshake.Panacea.ChainID, Height: 67},
				Ack:  IBCAcknowledgementObservation{Observed: true, ChainID: handshake.Osmosis.ChainID, Height: 68, Acknowledgement: "eyJyZXN1bHQiOiJBUT09In0="},
			},
		},
		FinalBalances: []IBCBalanceEvidence{
			{ChainID: handshake.Panacea.ChainID, Address: "panacea1user", Denom: "umed", Before: "8800", After: "7700", ExpectedAfter: "7700"},
			{ChainID: handshake.Panacea.ChainID, Address: "panacea1user", Denom: "ibc/OSMO", Before: "5", After: "1005", ExpectedAfter: "1005"},
			{ChainID: handshake.Osmosis.ChainID, Address: "osmo1user", Denom: "uosmo", Before: "10000", After: "8990", ExpectedAfter: "8990"},
			{ChainID: handshake.Osmosis.ChainID, Address: "osmo1user", Denom: "ibc/MED", Before: "1005", After: "2005", ExpectedAfter: "2005"},
		},
		EscrowBalances: validIBCEscrowBalances(handshake, ibcPhasePostUpgrade, "1000"),
		DenomTraces:    validIBCDenomTraceSnapshot(handshake, ibcPhasePostUpgrade),
	}
}

func validIBCNodeRestartSemantics(evidence IBCUpgradeContinuityEvidence) []IBCChainRestartEvidence {
	handshake := evidence.OriginalChannel
	panaceaTrace := validIBCDenomTraceSnapshot(handshake, "post-restart").Traces[0]
	osmosisTrace := validIBCDenomTraceSnapshot(handshake, "post-restart").Traces[1]
	panaceaNodes := make([]IBCNodePostRestartEvidence, 0, len(evidence.PanaceaUpgrade.PanaceaAfter))
	panaceaRestarts := make(map[string]NodeRestartEvidence, len(evidence.PanaceaNodeRestarts))
	for _, restart := range evidence.PanaceaNodeRestarts {
		panaceaRestarts[restart.Node] = restart
	}
	for _, identity := range evidence.PanaceaUpgrade.PanaceaAfter {
		restart := panaceaRestarts[identity.Name]
		panaceaNodes = append(panaceaNodes, IBCNodePostRestartEvidence{
			ChainID: handshake.Panacea.ChainID, IdentityBefore: identity, IdentityAfter: identity, Restart: restart,
			ObservedHeight: restart.After.Height, ClientStatus: "Active", ConnectionState: "Open", ChannelState: "Open",
			NextSequenceSend: evidence.FinalAfterHermesRestart.PanaceaNextSequenceSend,
			PacketSemantics: IBCNodePacketSemanticsEvidence{
				Role: "source", SuccessSequence: evidence.InFlight.Packet.Sequence, TimeoutSequence: evidence.Timeout.Packet.Sequence,
			},
			DenomTrace: panaceaTrace,
		})
	}
	osmosisNodes := make([]IBCNodePostRestartEvidence, 0, len(evidence.PanaceaUpgrade.OsmosisAfter))
	osmosisRestarts := make(map[string]NodeRestartEvidence, len(evidence.OsmosisNodeRestarts))
	for _, restart := range evidence.OsmosisNodeRestarts {
		osmosisRestarts[restart.Node] = restart
	}
	for _, identity := range evidence.PanaceaUpgrade.OsmosisAfter {
		restart := osmosisRestarts[identity.Name]
		osmosisNodes = append(osmosisNodes, IBCNodePostRestartEvidence{
			ChainID: handshake.Osmosis.ChainID, IdentityBefore: identity, IdentityAfter: identity, Restart: restart,
			ObservedHeight: restart.After.Height, ClientStatus: "Active", ConnectionState: "Open", ChannelState: "Open",
			NextSequenceSend: evidence.FinalAfterHermesRestart.OsmosisNextSequenceSend,
			PacketSemantics: IBCNodePacketSemanticsEvidence{
				Role: "destination", SuccessSequence: evidence.InFlight.Packet.Sequence, TimeoutSequence: evidence.Timeout.Packet.Sequence,
				SuccessReceipt: true, SuccessAcknowledgement: evidence.InFlightRelay.DestinationAcknowledgement,
			},
			DenomTrace: osmosisTrace,
		})
	}
	return []IBCChainRestartEvidence{
		{ChainID: handshake.Panacea.ChainID, Nodes: panaceaNodes},
		{ChainID: handshake.Osmosis.ChainID, Nodes: osmosisNodes},
	}
}
