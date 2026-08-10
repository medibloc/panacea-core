package harness

import (
	"context"
	"testing"
)

func TestStableIBCQueryHeightLeavesReportedTipForBlockResults(t *testing.T) {
	current := func(context.Context) (int64, error) { return 102, nil }
	height, err := stableIBCQueryHeight(current)(context.Background())
	if err != nil {
		t.Fatalf("stable height returned an error: %v", err)
	}
	if height != 101 {
		t.Fatalf("stable height = %d, want 101", height)
	}

	genesis := func(context.Context) (int64, error) { return 1, nil }
	height, err = stableIBCQueryHeight(genesis)(context.Background())
	if err != nil {
		t.Fatalf("genesis height returned an error: %v", err)
	}
	if height != 1 {
		t.Fatalf("genesis stable height = %d, want 1", height)
	}
}

func TestStableIBCScanEndHeightLeavesReportedTipForBlockResults(t *testing.T) {
	if height := stableIBCScanEndHeight(102); height != 101 {
		t.Fatalf("stable scan end height = %d, want 101", height)
	}
	if height := stableIBCScanEndHeight(1); height != 1 {
		t.Fatalf("genesis stable scan end height = %d, want 1", height)
	}
}

func TestIBCChannelHandshakeRequiresOneMutuallyBoundOpenTransferChannel(t *testing.T) {
	valid := validIBCChannelHandshake()

	if err := valid.Validate(); err != nil {
		t.Fatalf("valid handshake rejected: %v", err)
	}

	mismatched := valid
	mismatched.Osmosis.ChannelID = "channel-7"
	if err := mismatched.Validate(); err == nil {
		t.Fatal("handshake with mismatched counterparty channel unexpectedly validated")
	}
}

func TestPreUpgradeEvidenceRequiresBidirectionalRecvAckAndExactBalances(t *testing.T) {
	handshake := validIBCChannelHandshake()
	evidence := IBCPreUpgradeTransferEvidence{
		Phase:   "pre-upgrade",
		Channel: handshake,
		Transfers: []IBCPacketLifecycleEvidence{
			{
				Direction:          "panacea-to-osmosis",
				SourceChainID:      handshake.Panacea.ChainID,
				DestinationChainID: handshake.Osmosis.ChainID,
				TxHash:             "AABB",
				TxHeight:           12,
				Sequence:           1,
				SourcePort:         "transfer",
				SourceChannel:      handshake.Panacea.ChannelID,
				DestinationPort:    "transfer",
				DestinationChannel: handshake.Osmosis.ChannelID,
				Denom:              "umed",
				Amount:             "1000",
				Recv:               IBCPacketObservation{Observed: true, ChainID: handshake.Osmosis.ChainID, Height: 15},
				Ack:                IBCAcknowledgementObservation{Observed: true, ChainID: handshake.Panacea.ChainID, Height: 16, Acknowledgement: "eyJyZXN1bHQiOiJBUT09In0="},
			},
			{
				Direction:          "osmosis-to-panacea",
				SourceChainID:      handshake.Osmosis.ChainID,
				DestinationChainID: handshake.Panacea.ChainID,
				TxHash:             "CCDD",
				TxHeight:           20,
				Sequence:           1,
				SourcePort:         "transfer",
				SourceChannel:      handshake.Osmosis.ChannelID,
				DestinationPort:    "transfer",
				DestinationChannel: handshake.Panacea.ChannelID,
				Denom:              "uosmo",
				Amount:             "1000",
				Recv:               IBCPacketObservation{Observed: true, ChainID: handshake.Panacea.ChainID, Height: 23},
				Ack:                IBCAcknowledgementObservation{Observed: true, ChainID: handshake.Osmosis.ChainID, Height: 24, Acknowledgement: "eyJyZXN1bHQiOiJBUT09In0="},
			},
		},
		FinalBalances: []IBCBalanceEvidence{
			{ChainID: handshake.Panacea.ChainID, Address: "panacea1sender", Denom: "umed", Before: "100000", After: "98000", ExpectedAfter: "98000"},
			{ChainID: handshake.Panacea.ChainID, Address: "panacea1sender", Denom: "ibc/osmo", Before: "0", After: "1000", ExpectedAfter: "1000"},
			{ChainID: handshake.Osmosis.ChainID, Address: "osmo1sender", Denom: "uosmo", Before: "100000", After: "98900", ExpectedAfter: "98900"},
			{ChainID: handshake.Osmosis.ChainID, Address: "osmo1sender", Denom: "ibc/med", Before: "0", After: "1000", ExpectedAfter: "1000"},
		},
		EscrowBalances: validIBCEscrowBalances(handshake, ibcPhasePreUpgrade, "1000"),
		DenomTraces:    validIBCDenomTraceSnapshot(handshake, ibcPhasePreUpgrade),
	}

	if err := evidence.Validate(); err != nil {
		t.Fatalf("valid bidirectional evidence rejected: %v", err)
	}

	missingAck := evidence
	missingAck.Transfers = append([]IBCPacketLifecycleEvidence(nil), evidence.Transfers...)
	missingAck.Transfers[1].Ack.Observed = false
	if err := missingAck.Validate(); err == nil {
		t.Fatal("evidence missing an acknowledgement unexpectedly validated")
	}

	errorAck := evidence
	errorAck.Transfers = append([]IBCPacketLifecycleEvidence(nil), evidence.Transfers...)
	errorAck.Transfers[0].Ack.Acknowledgement = "eyJlcnJvciI6InRyYW5zZmVyIGZhaWxlZCJ9"
	if err := errorAck.Validate(); err == nil {
		t.Fatal("evidence containing an ICS-20 error acknowledgement unexpectedly validated")
	}

	wrongBalance := evidence
	wrongBalance.FinalBalances = append([]IBCBalanceEvidence(nil), evidence.FinalBalances...)
	wrongBalance.FinalBalances[3].After = "999"
	if err := wrongBalance.Validate(); err == nil {
		t.Fatal("evidence with a final balance mismatch unexpectedly validated")
	}

	wrongEscrow := evidence
	wrongEscrow.EscrowBalances = append([]IBCEscrowBalanceEvidence(nil), evidence.EscrowBalances...)
	wrongEscrow.EscrowBalances[0].After = "999"
	if err := wrongEscrow.Validate(); err == nil {
		t.Fatal("evidence without the exact escrow delta unexpectedly validated")
	}

	wrongTrace := evidence
	wrongTrace.DenomTraces.Traces = append([]IBCDenomTraceEvidence(nil), evidence.DenomTraces.Traces...)
	wrongTrace.DenomTraces.Traces[0].Path = "transfer/channel-99"
	if err := wrongTrace.Validate(); err == nil {
		t.Fatal("evidence with a mismatched denom trace unexpectedly validated")
	}
}

func validIBCEscrowBalances(handshake IBCChannelHandshake, phase, amount string) []IBCEscrowBalanceEvidence {
	return []IBCEscrowBalanceEvidence{
		{
			Phase: phase, ChainID: handshake.Panacea.ChainID, PortID: handshake.Panacea.PortID, ChannelID: handshake.Panacea.ChannelID,
			Address: "panacea1escrow", Denom: "umed", Before: "100", After: "1100", ExpectedDelta: amount, ExpectedAfter: "1100",
		},
		{
			Phase: phase, ChainID: handshake.Osmosis.ChainID, PortID: handshake.Osmosis.PortID, ChannelID: handshake.Osmosis.ChannelID,
			Address: "osmo1escrow", Denom: "uosmo", Before: "200", After: "1200", ExpectedDelta: amount, ExpectedAfter: "1200",
		},
	}
}

func validIBCDenomTraceSnapshot(handshake IBCChannelHandshake, phase string) IBCDenomTraceSnapshot {
	return IBCDenomTraceSnapshot{Phase: phase, Traces: []IBCDenomTraceEvidence{
		{ChainID: handshake.Panacea.ChainID, VoucherDenom: "ibc/OSMO", Hash: "OSMO", Path: handshake.Panacea.PortID + "/" + handshake.Panacea.ChannelID, BaseDenom: "uosmo"},
		{ChainID: handshake.Osmosis.ChainID, VoucherDenom: "ibc/MED", Hash: "MED", Path: handshake.Osmosis.PortID + "/" + handshake.Osmosis.ChannelID, BaseDenom: "umed"},
	}}
}

func validIBCChannelHandshake() IBCChannelHandshake {
	return IBCChannelHandshake{
		Path: "panacea-osmosis",
		Panacea: IBCChannelEndpoint{
			ChainID:                  "panacea-run-a1",
			CounterpartyChainID:      "osmosis-run-a1",
			ClientID:                 "07-tendermint-0",
			CounterpartyClientID:     "07-tendermint-1",
			ConnectionID:             "connection-0",
			CounterpartyConnectionID: "connection-0",
			ConnectionState:          "Open",
			PortID:                   "transfer",
			ChannelID:                "channel-0",
			CounterpartyChannelID:    "channel-0",
			ChannelState:             "Open",
			Ordering:                 "Unordered",
			Version:                  "ics20-1",
		},
		Osmosis: IBCChannelEndpoint{
			ChainID:                  "osmosis-run-a1",
			CounterpartyChainID:      "panacea-run-a1",
			ClientID:                 "07-tendermint-1",
			CounterpartyClientID:     "07-tendermint-0",
			ConnectionID:             "connection-0",
			CounterpartyConnectionID: "connection-0",
			ConnectionState:          "Open",
			PortID:                   "transfer",
			ChannelID:                "channel-0",
			CounterpartyChannelID:    "channel-0",
			ChannelState:             "Open",
			Ordering:                 "Unordered",
			Version:                  "ics20-1",
		},
	}
}
