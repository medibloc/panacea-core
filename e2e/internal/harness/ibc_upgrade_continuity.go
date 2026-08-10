package harness

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

const (
	ibcPhasePostUpgrade        = "post-upgrade"
	ibcPhasePostUpgradeTimeout = "post-upgrade-timeout"
	ibcUpgradeContinuityPhase  = "v2.2.1-to-current"
	ibcPacketTerminalSuccess   = "success"
	ibcPacketTerminalTimeout   = "timeout"
)

// IBCLinkStateSnapshot proves the retained path is open and both clients are
// active while preserving each local channel's next send sequence.
type IBCLinkStateSnapshot struct {
	Phase                   string              `json:"phase"`
	Channel                 IBCChannelHandshake `json:"channel"`
	PanaceaClientStatus     string              `json:"panacea_client_status"`
	OsmosisClientStatus     string              `json:"osmosis_client_status"`
	PanaceaHeight           int64               `json:"panacea_height"`
	OsmosisHeight           int64               `json:"osmosis_height"`
	PanaceaNextSequenceSend uint64              `json:"panacea_next_sequence_send"`
	OsmosisNextSequenceSend uint64              `json:"osmosis_next_sequence_send"`
}

// IBCPacketSendEvidence is the committed source-side half of an in-flight
// packet. PacketData and the packet commitment are retained separately so the
// later receive and acknowledgement can be tied to the exact send.
type IBCPacketSendEvidence struct {
	Direction          string `json:"direction"`
	SourceChainID      string `json:"source_chain_id"`
	DestinationChainID string `json:"destination_chain_id"`
	TxHash             string `json:"tx_hash"`
	TxHeight           int64  `json:"tx_height"`
	Sequence           uint64 `json:"sequence"`
	SourcePort         string `json:"source_port"`
	SourceChannel      string `json:"source_channel"`
	DestinationPort    string `json:"destination_port"`
	DestinationChannel string `json:"destination_channel"`
	Denom              string `json:"denom"`
	Amount             string `json:"amount"`
	GasFee             string `json:"gas_fee"`
	PacketData         string `json:"packet_data_base64"`
	TimeoutHeight      string `json:"timeout_height"`
	TimeoutTimestamp   uint64 `json:"timeout_timestamp"`
}

// IBCInFlightPacketCheckpoint proves Hermes was stopped before a single
// Panacea packet advanced the send sequence and left a commitment without a
// destination receipt or acknowledgement.
type IBCInFlightPacketCheckpoint struct {
	Phase                      string                   `json:"phase"`
	Channel                    IBCChannelHandshake      `json:"channel"`
	BeforeSendState            IBCLinkStateSnapshot     `json:"before_send_state"`
	AfterSendState             IBCLinkStateSnapshot     `json:"after_send_state"`
	Packet                     IBCPacketSendEvidence    `json:"packet"`
	Commitment                 string                   `json:"commitment_base64"`
	DestinationReceipt         bool                     `json:"destination_receipt"`
	DestinationAcknowledgement string                   `json:"destination_acknowledgement_base64,omitempty"`
	DestinationScanStartHeight int64                    `json:"destination_scan_start_height"`
	SourceNativeBalance        IBCBalanceEvidence       `json:"source_native_balance"`
	DestinationVoucherBalance  IBCBalanceEvidence       `json:"destination_voucher_balance"`
	SourceEscrowLock           IBCEscrowBalanceEvidence `json:"source_escrow_lock"`
}

// IBCNodeRuntimeIdentity proves which containers and image references were
// retained or replaced by the Panacea-only upgrade callback.
type IBCNodeRuntimeIdentity struct {
	Name        string   `json:"name"`
	ContainerID string   `json:"container_id"`
	Image       ImageRef `json:"image"`
}

// IBCPanaceaUpgradeStepEvidence brackets the caller-supplied upgrade callback.
// Panacea containers must change from v2.2.1 to current while Osmosis remains
// byte-for-byte the same runtime identity.
type IBCPanaceaUpgradeStepEvidence struct {
	CallbackCompleted bool                      `json:"callback_completed"`
	UpgradeHeight     int64                     `json:"upgrade_height"`
	BeforeHeight      int64                     `json:"before_height"`
	AfterHeight       int64                     `json:"after_height"`
	From              ImageRef                  `json:"from"`
	To                ImageRef                  `json:"to"`
	PanaceaBefore     []IBCNodeRuntimeIdentity  `json:"panacea_before"`
	PanaceaAfter      []IBCNodeRuntimeIdentity  `json:"panacea_after"`
	OsmosisBefore     []IBCNodeRuntimeIdentity  `json:"osmosis_before"`
	OsmosisAfter      []IBCNodeRuntimeIdentity  `json:"osmosis_after"`
	OsmosisProgress   IBCHeightProgressEvidence `json:"osmosis_progress"`
	Error             string                    `json:"error,omitempty"`
}

type IBCHeightSample struct {
	ObservedAt time.Time `json:"observed_at"`
	Height     int64     `json:"height"`
	Error      string    `json:"error,omitempty"`
}

// IBCHeightProgressEvidence proves the fixed counterparty kept committing
// blocks throughout the Panacea-only upgrade window, rather than merely being
// reachable before and after it.
type IBCHeightProgressEvidence struct {
	ChainID             string            `json:"chain_id"`
	StartedAt           time.Time         `json:"started_at"`
	CompletedAt         time.Time         `json:"completed_at"`
	StartHeight         int64             `json:"start_height"`
	EndHeight           int64             `json:"end_height"`
	MaxNoProgressMillis int64             `json:"max_no_progress_millis"`
	BoundMillis         int64             `json:"bound_millis"`
	Samples             []IBCHeightSample `json:"samples"`
}

// IBCInFlightRelayEvidence proves exactly one successful receive and one
// successful acknowledgement cleared the original commitment after Hermes
// restarted.
type IBCInFlightRelayEvidence struct {
	Packet                     IBCPacketLifecycleEvidence `json:"packet"`
	ReceiveCount               int                        `json:"receive_count"`
	AcknowledgementCount       int                        `json:"acknowledgement_count"`
	CommitmentCleared          bool                       `json:"commitment_cleared"`
	DestinationReceipt         bool                       `json:"destination_receipt"`
	DestinationAcknowledgement string                     `json:"destination_acknowledgement_base64"`
	SourceNativeBalance        IBCBalanceEvidence         `json:"source_native_balance"`
	DestinationVoucherBalance  IBCBalanceEvidence         `json:"destination_voucher_balance"`
}

// IBCPacketTimeoutEvidence proves that a distinct post-upgrade packet expired
// without a receive, cleared its commitment exactly once, and refunded the
// transfer amount to the source sender (leaving only the send gas fee).
type IBCPacketTimeoutEvidence struct {
	Phase                      string                   `json:"phase"`
	BeforeSendState            IBCLinkStateSnapshot     `json:"before_send_state"`
	AfterTimeoutState          IBCLinkStateSnapshot     `json:"after_timeout_state"`
	Packet                     IBCPacketSendEvidence    `json:"packet"`
	TimeoutCount               int                      `json:"timeout_count"`
	ReceiveCount               int                      `json:"receive_count"`
	CommitmentCleared          bool                     `json:"commitment_cleared"`
	DestinationReceipt         bool                     `json:"destination_receipt"`
	DestinationAcknowledgement string                   `json:"destination_acknowledgement_base64,omitempty"`
	SourceNativeBalance        IBCBalanceEvidence       `json:"source_native_balance"`
	DestinationVoucherBalance  IBCBalanceEvidence       `json:"destination_voucher_balance"`
	SourceEscrowLock           IBCEscrowBalanceEvidence `json:"source_escrow_lock"`
	SourceEscrowRefund         IBCEscrowBalanceEvidence `json:"source_escrow_refund"`
}

// IBCHermesRestartEvidence records the two required daemon restarts and the
// successful health check that follows each one.
type IBCHermesRestartEvidence struct {
	Phase                string `json:"phase"`
	PanaceaBeforeHeight  int64  `json:"panacea_before_height"`
	OsmosisBeforeHeight  int64  `json:"osmosis_before_height"`
	PanaceaAfterHeight   int64  `json:"panacea_after_height"`
	OsmosisAfterHeight   int64  `json:"osmosis_after_height"`
	HealthCheckCompleted bool   `json:"health_check_completed"`
}

// IBCPacketTerminalStateEvidence re-queries the successful and timed-out
// packets after every Panacea node has restarted.
type IBCPacketTerminalStateEvidence struct {
	Kind                       string `json:"kind"`
	Sequence                   uint64 `json:"sequence"`
	Commitment                 string `json:"commitment_base64,omitempty"`
	DestinationReceipt         bool   `json:"destination_receipt"`
	DestinationAcknowledgement string `json:"destination_acknowledgement_base64,omitempty"`
}

// IBCDenomTraceEvidence binds the observed voucher hash to its exact local
// transfer path and counterparty base denomination.
type IBCDenomTraceEvidence struct {
	ChainID      string `json:"chain_id"`
	VoucherDenom string `json:"voucher_denom"`
	Hash         string `json:"hash"`
	Path         string `json:"path"`
	BaseDenom    string `json:"base_denom"`
}

type IBCDenomTraceSnapshot struct {
	Phase  string                  `json:"phase"`
	Traces []IBCDenomTraceEvidence `json:"traces"`
}

// IBCPostUpgradeTransferEvidence is the same bidirectional recv/ack and exact
// balance contract as the pre-upgrade phase, but on the retained channel.
type IBCPostUpgradeTransferEvidence struct {
	Phase          string                       `json:"phase"`
	Channel        IBCChannelHandshake          `json:"channel"`
	Transfers      []IBCPacketLifecycleEvidence `json:"transfers"`
	FinalBalances  []IBCBalanceEvidence         `json:"final_balances"`
	EscrowBalances []IBCEscrowBalanceEvidence   `json:"escrow_balances"`
	DenomTraces    IBCDenomTraceSnapshot        `json:"denom_traces"`
}

type IBCNodePacketSemanticsEvidence struct {
	Role                     string `json:"role"`
	SuccessSequence          uint64 `json:"success_sequence"`
	TimeoutSequence          uint64 `json:"timeout_sequence"`
	SuccessCommitmentPresent bool   `json:"success_commitment_present"`
	TimeoutCommitmentPresent bool   `json:"timeout_commitment_present"`
	SuccessReceipt           bool   `json:"success_receipt"`
	TimeoutReceipt           bool   `json:"timeout_receipt"`
	SuccessAcknowledgement   string `json:"success_acknowledgement_base64,omitempty"`
	TimeoutAcknowledgement   string `json:"timeout_acknowledgement_base64,omitempty"`
}

type IBCNodePostRestartEvidence struct {
	ChainID          string                         `json:"chain_id"`
	IdentityBefore   IBCNodeRuntimeIdentity         `json:"identity_before"`
	IdentityAfter    IBCNodeRuntimeIdentity         `json:"identity_after"`
	Restart          NodeRestartEvidence            `json:"restart"`
	ObservedHeight   int64                          `json:"observed_height"`
	ClientStatus     string                         `json:"client_status"`
	ConnectionState  string                         `json:"connection_state"`
	ChannelState     string                         `json:"channel_state"`
	NextSequenceSend uint64                         `json:"next_sequence_send"`
	PacketSemantics  IBCNodePacketSemanticsEvidence `json:"packet_semantics"`
	DenomTrace       IBCDenomTraceEvidence          `json:"denom_trace"`
}

type IBCChainRestartEvidence struct {
	ChainID string                       `json:"chain_id"`
	Nodes   []IBCNodePostRestartEvidence `json:"nodes"`
}

// IBCUpgradeContinuityEvidence is the complete same-channel v2.2.1-to-current
// proof, including a distinct timeout/refund lifecycle and final re-query
// after Hermes and every Panacea node restart.
type IBCUpgradeContinuityEvidence struct {
	Phase                   string                           `json:"phase"`
	OriginalChannel         IBCChannelHandshake              `json:"original_channel"`
	InFlight                IBCInFlightPacketCheckpoint      `json:"in_flight"`
	PanaceaUpgrade          IBCPanaceaUpgradeStepEvidence    `json:"panacea_upgrade"`
	PostUpgradeBeforeRelay  IBCLinkStateSnapshot             `json:"post_upgrade_before_relay"`
	InFlightRelay           IBCInFlightRelayEvidence         `json:"in_flight_relay"`
	AfterInFlightRelay      IBCLinkStateSnapshot             `json:"after_in_flight_relay"`
	Timeout                 IBCPacketTimeoutEvidence         `json:"timeout"`
	PostUpgradeTransfers    IBCPostUpgradeTransferEvidence   `json:"post_upgrade_transfers"`
	FinalAfterHermesRestart IBCLinkStateSnapshot             `json:"final_after_hermes_restart"`
	HermesRestarts          []IBCHermesRestartEvidence       `json:"hermes_restarts"`
	PanaceaNodeRestarts     []NodeRestartEvidence            `json:"panacea_node_restarts"`
	OsmosisNodeRestarts     []NodeRestartEvidence            `json:"osmosis_node_restarts"`
	NodeRestartSemantics    []IBCChainRestartEvidence        `json:"node_restart_semantics"`
	FinalPacketStates       []IBCPacketTerminalStateEvidence `json:"final_packet_states"`
	FinalDenomTraces        []IBCDenomTraceEvidence          `json:"final_denom_traces"`
	DenomTraceContinuity    []IBCDenomTraceSnapshot          `json:"denom_trace_continuity"`
	FinalBalances           []IBCBalanceEvidence             `json:"final_balances"`
	FinalEscrowBalances     []IBCEscrowBalanceEvidence       `json:"final_escrow_balances"`
}

func (s IBCLinkStateSnapshot) ValidateAgainst(channel IBCChannelHandshake) error {
	var validationErrors []error
	if strings.TrimSpace(s.Phase) == "" {
		validationErrors = append(validationErrors, errors.New("IBC link state phase is required"))
	}
	if err := s.Channel.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("validate IBC link state channel: %w", err))
	}
	if s.Channel != channel {
		validationErrors = append(validationErrors, errors.New("IBC link state does not use the original client, connection, and channel identifiers"))
	}
	if !ibcStateEquals(s.PanaceaClientStatus, "ACTIVE") {
		validationErrors = append(validationErrors, fmt.Errorf("Panacea IBC client status = %q, want active", s.PanaceaClientStatus))
	}
	if !ibcStateEquals(s.OsmosisClientStatus, "ACTIVE") {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis IBC client status = %q, want active", s.OsmosisClientStatus))
	}
	if s.PanaceaHeight < 1 || s.OsmosisHeight < 1 {
		validationErrors = append(validationErrors, errors.New("IBC link state heights must be positive"))
	}
	if s.PanaceaNextSequenceSend == 0 || s.OsmosisNextSequenceSend == 0 {
		validationErrors = append(validationErrors, errors.New("IBC next send sequences must be positive"))
	}
	return errors.Join(validationErrors...)
}

func (e IBCPacketSendEvidence) validate(source, destination IBCChannelEndpoint) error {
	var validationErrors []error
	if e.Direction != ibcPanaceaToOsmosis {
		validationErrors = append(validationErrors, fmt.Errorf("in-flight direction = %q, want %q", e.Direction, ibcPanaceaToOsmosis))
	}
	if e.SourceChainID != source.ChainID || e.DestinationChainID != destination.ChainID {
		validationErrors = append(validationErrors, errors.New("in-flight packet chain IDs do not match the original endpoints"))
	}
	if strings.TrimSpace(e.TxHash) == "" || e.TxHeight < 1 || e.Sequence == 0 {
		validationErrors = append(validationErrors, errors.New("in-flight send transaction identity is incomplete"))
	}
	if e.SourcePort != source.PortID || e.SourceChannel != source.ChannelID ||
		e.DestinationPort != destination.PortID || e.DestinationChannel != destination.ChannelID {
		validationErrors = append(validationErrors, errors.New("in-flight packet does not use the original channel endpoints"))
	}
	if strings.TrimSpace(e.Denom) == "" || !isPositiveInteger(e.Amount) || !isNonNegativeInteger(e.GasFee) {
		validationErrors = append(validationErrors, errors.New("in-flight denomination, amount, or gas fee is invalid"))
	}
	packetData, err := base64.StdEncoding.DecodeString(e.PacketData)
	if err != nil || len(packetData) == 0 {
		validationErrors = append(validationErrors, errors.New("in-flight packet data must be non-empty base64"))
	}
	if strings.TrimSpace(e.TimeoutHeight) == "" && e.TimeoutTimestamp == 0 {
		validationErrors = append(validationErrors, errors.New("in-flight packet requires a timeout"))
	}
	return errors.Join(validationErrors...)
}

func (e IBCInFlightPacketCheckpoint) Validate() error {
	var validationErrors []error
	if e.Phase != "in-flight-staged" {
		validationErrors = append(validationErrors, fmt.Errorf("in-flight checkpoint phase = %q, want in-flight-staged", e.Phase))
	}
	if err := e.Channel.Validate(); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.BeforeSendState.ValidateAgainst(e.Channel); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.AfterSendState.ValidateAgainst(e.Channel); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.Packet.validate(e.Channel.Panacea, e.Channel.Osmosis); err != nil {
		validationErrors = append(validationErrors, err)
	}
	commitment, err := base64.StdEncoding.DecodeString(e.Commitment)
	if err != nil || len(commitment) == 0 {
		validationErrors = append(validationErrors, errors.New("in-flight packet commitment must be non-empty base64"))
	}
	if e.DestinationReceipt || e.DestinationAcknowledgement != "" {
		validationErrors = append(validationErrors, errors.New("in-flight packet was already relayed before the upgrade"))
	}
	if e.DestinationScanStartHeight < 1 {
		validationErrors = append(validationErrors, errors.New("in-flight destination scan start height must be positive"))
	}
	if e.Packet.Sequence != e.BeforeSendState.PanaceaNextSequenceSend ||
		e.AfterSendState.PanaceaNextSequenceSend != e.Packet.Sequence+1 ||
		e.AfterSendState.OsmosisNextSequenceSend != e.BeforeSendState.OsmosisNextSequenceSend {
		validationErrors = append(validationErrors, errors.New("in-flight send sequence did not advance exactly once"))
	}
	if err := e.SourceNativeBalance.validate(); err != nil {
		validationErrors = append(validationErrors, err)
	} else if err := validateSourceSendBalance(e.SourceNativeBalance, e.Packet.Amount, e.Packet.GasFee); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.DestinationVoucherBalance.validate(); err != nil {
		validationErrors = append(validationErrors, err)
	} else if e.DestinationVoucherBalance.Before != e.DestinationVoucherBalance.After {
		validationErrors = append(validationErrors, errors.New("destination voucher changed while Hermes was stopped"))
	}
	if err := e.SourceEscrowLock.validate(e.Channel.Panacea, "in-flight-staged"); err != nil {
		validationErrors = append(validationErrors, err)
	} else if e.SourceEscrowLock.ExpectedDelta != e.Packet.Amount {
		validationErrors = append(validationErrors, errors.New("in-flight escrow lock delta does not equal the packet amount"))
	}
	return errors.Join(validationErrors...)
}

func (e IBCPanaceaUpgradeStepEvidence) Validate() error {
	var validationErrors []error
	if !e.CallbackCompleted || e.Error != "" {
		validationErrors = append(validationErrors, errors.New("Panacea upgrade callback did not complete successfully"))
	}
	if e.UpgradeHeight < 1 || e.BeforeHeight < 1 || e.AfterHeight < e.UpgradeHeight {
		validationErrors = append(validationErrors, errors.New("Panacea upgrade heights are invalid"))
	}
	if strings.TrimSpace(e.From.Repository) == "" || strings.TrimSpace(e.From.Version) == "" ||
		strings.TrimSpace(e.To.Repository) == "" || strings.TrimSpace(e.To.Version) == "" || e.From == e.To {
		validationErrors = append(validationErrors, errors.New("Panacea upgrade image transition is invalid"))
	}
	if err := validateChangedNodeIdentities(e.PanaceaBefore, e.PanaceaAfter, e.From, e.To); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := validateUnchangedNodeIdentities(e.OsmosisBefore, e.OsmosisAfter); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.OsmosisProgress.validate(); err != nil {
		validationErrors = append(validationErrors, err)
	}
	return errors.Join(validationErrors...)
}

func (e IBCInFlightRelayEvidence) validate(checkpoint IBCInFlightPacketCheckpoint) error {
	var validationErrors []error
	if err := e.Packet.validate(checkpoint.Channel.Panacea, checkpoint.Channel.Osmosis); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if !lifecycleMatchesSend(e.Packet, checkpoint.Packet) {
		validationErrors = append(validationErrors, errors.New("relayed packet does not match the staged in-flight send"))
	}
	if e.ReceiveCount != 1 || e.AcknowledgementCount != 1 {
		validationErrors = append(validationErrors, fmt.Errorf("in-flight relay counts = recv:%d ack:%d, want exactly 1 each", e.ReceiveCount, e.AcknowledgementCount))
	}
	if !e.CommitmentCleared || !e.DestinationReceipt {
		validationErrors = append(validationErrors, errors.New("in-flight commitment/receipt terminal state is incomplete"))
	}
	acknowledgementCommitment, err := acknowledgementCommitmentBase64(e.Packet.Ack.Acknowledgement)
	if err != nil {
		validationErrors = append(validationErrors, err)
	} else if e.DestinationAcknowledgement != acknowledgementCommitment {
		validationErrors = append(validationErrors, errors.New("destination acknowledgement commitment does not match the source acknowledgement"))
	}
	if err := e.SourceNativeBalance.validate(); err != nil {
		validationErrors = append(validationErrors, err)
	} else if e.SourceNativeBalance.Before != e.SourceNativeBalance.After {
		validationErrors = append(validationErrors, errors.New("source native balance changed while relaying the staged packet"))
	}
	if err := e.DestinationVoucherBalance.validate(); err != nil {
		validationErrors = append(validationErrors, err)
	} else if err := validateDestinationReceiveBalance(e.DestinationVoucherBalance, checkpoint.Packet.Amount); err != nil {
		validationErrors = append(validationErrors, err)
	}
	return errors.Join(validationErrors...)
}

func (e IBCHeightProgressEvidence) validate() error {
	var validationErrors []error
	if strings.TrimSpace(e.ChainID) == "" || e.StartedAt.IsZero() || e.CompletedAt.IsZero() || !e.CompletedAt.After(e.StartedAt) {
		validationErrors = append(validationErrors, errors.New("IBC counterparty progress window is incomplete"))
	}
	if e.StartHeight < 1 || e.EndHeight <= e.StartHeight {
		validationErrors = append(validationErrors, errors.New("IBC counterparty did not advance during the Panacea upgrade"))
	}
	if e.BoundMillis < 1 || e.MaxNoProgressMillis < 0 || e.MaxNoProgressMillis > e.BoundMillis {
		validationErrors = append(validationErrors, errors.New("IBC counterparty no-progress bound was exceeded"))
	}
	if len(e.Samples) < 2 {
		validationErrors = append(validationErrors, errors.New("IBC counterparty progress requires at least two height samples"))
		return errors.Join(validationErrors...)
	}
	if e.Samples[0].Height != e.StartHeight || e.Samples[len(e.Samples)-1].Height != e.EndHeight {
		validationErrors = append(validationErrors, errors.New("IBC counterparty progress endpoints do not match its samples"))
	}
	for index, sample := range e.Samples {
		if sample.Height < 1 || sample.ObservedAt.IsZero() || sample.Error != "" {
			validationErrors = append(validationErrors, fmt.Errorf("IBC counterparty progress sample %d is invalid", index))
		}
		if index > 0 && (sample.Height < e.Samples[index-1].Height || sample.ObservedAt.Before(e.Samples[index-1].ObservedAt)) {
			validationErrors = append(validationErrors, errors.New("IBC counterparty height samples are not monotonic"))
		}
	}
	return errors.Join(validationErrors...)
}

func (e IBCPacketTimeoutEvidence) validate(channel IBCChannelHandshake) error {
	var validationErrors []error
	if e.Phase != ibcPhasePostUpgradeTimeout {
		validationErrors = append(validationErrors, fmt.Errorf("IBC timeout phase = %q, want %q", e.Phase, ibcPhasePostUpgradeTimeout))
	}
	if err := e.BeforeSendState.ValidateAgainst(channel); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.AfterTimeoutState.ValidateAgainst(channel); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.Packet.validate(channel.Panacea, channel.Osmosis); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if e.Packet.Sequence != e.BeforeSendState.PanaceaNextSequenceSend ||
		e.AfterTimeoutState.PanaceaNextSequenceSend != e.Packet.Sequence+1 ||
		e.AfterTimeoutState.OsmosisNextSequenceSend != e.BeforeSendState.OsmosisNextSequenceSend {
		validationErrors = append(validationErrors, errors.New("timeout packet send sequence did not advance exactly once"))
	}
	if e.TimeoutCount != 1 {
		validationErrors = append(validationErrors, fmt.Errorf("timeout message count = %d, want exactly 1", e.TimeoutCount))
	}
	if e.ReceiveCount != 0 {
		validationErrors = append(validationErrors, fmt.Errorf("timed-out packet receive count = %d, want 0", e.ReceiveCount))
	}
	if !e.CommitmentCleared || e.DestinationReceipt || e.DestinationAcknowledgement != "" {
		validationErrors = append(validationErrors, errors.New("timeout packet terminal commitment, receipt, or acknowledgement state is invalid"))
	}
	if err := e.SourceNativeBalance.validate(); err != nil {
		validationErrors = append(validationErrors, err)
	} else if err := validateSourceRefundBalance(e.SourceNativeBalance, e.Packet.GasFee); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.DestinationVoucherBalance.validate(); err != nil {
		validationErrors = append(validationErrors, err)
	} else if e.DestinationVoucherBalance.Before != e.DestinationVoucherBalance.After {
		validationErrors = append(validationErrors, errors.New("destination voucher changed for a timed-out packet"))
	}
	if err := e.SourceEscrowLock.validate(channel.Panacea, ibcPhasePostUpgradeTimeout+"-lock"); err != nil {
		validationErrors = append(validationErrors, err)
	} else if e.SourceEscrowLock.ExpectedDelta != e.Packet.Amount {
		validationErrors = append(validationErrors, errors.New("timeout escrow lock delta does not equal the packet amount"))
	}
	if err := e.SourceEscrowRefund.validate(channel.Panacea, ibcPhasePostUpgradeTimeout+"-refund"); err != nil {
		validationErrors = append(validationErrors, err)
	} else if e.SourceEscrowRefund.ExpectedDelta != "-"+e.Packet.Amount || e.SourceEscrowRefund.Before != e.SourceEscrowLock.After {
		validationErrors = append(validationErrors, errors.New("timeout escrow refund did not release the exact locked amount"))
	}
	return errors.Join(validationErrors...)
}

func (e IBCHermesRestartEvidence) validate() error {
	if strings.TrimSpace(e.Phase) == "" {
		return errors.New("Hermes restart phase is required")
	}
	if e.PanaceaBeforeHeight < 1 || e.OsmosisBeforeHeight < 1 ||
		e.PanaceaAfterHeight <= e.PanaceaBeforeHeight || e.OsmosisAfterHeight <= e.OsmosisBeforeHeight {
		return errors.New("Hermes restart must be bracketed by advancing chain heights")
	}
	if !e.HealthCheckCompleted {
		return errors.New("Hermes restart health check did not complete")
	}
	return nil
}

func (e IBCPacketTerminalStateEvidence) validate(successSequence, timeoutSequence uint64, successAck string) error {
	switch e.Kind {
	case ibcPacketTerminalSuccess:
		if e.Sequence != successSequence || e.Commitment != "" || !e.DestinationReceipt || e.DestinationAcknowledgement != successAck {
			return errors.New("post-restart successful packet terminal state is invalid")
		}
		return nil
	case ibcPacketTerminalTimeout:
		if e.Sequence != timeoutSequence || e.Commitment != "" || e.DestinationReceipt || e.DestinationAcknowledgement != "" {
			return errors.New("post-restart timeout packet terminal state is invalid")
		}
		return nil
	default:
		return fmt.Errorf("unknown post-restart packet terminal kind %q", e.Kind)
	}
}

func (e IBCDenomTraceEvidence) validate(channel IBCChannelHandshake) error {
	var endpoint IBCChannelEndpoint
	var baseDenom string
	switch e.ChainID {
	case channel.Panacea.ChainID:
		endpoint = channel.Panacea
		baseDenom = "uosmo"
	case channel.Osmosis.ChainID:
		endpoint = channel.Osmosis
		baseDenom = "umed"
	default:
		return fmt.Errorf("denom trace uses unknown chain %q", e.ChainID)
	}
	expectedPath := endpoint.PortID + "/" + endpoint.ChannelID
	if strings.TrimSpace(e.Hash) == "" || e.VoucherDenom != "ibc/"+e.Hash || e.Path != expectedPath || e.BaseDenom != baseDenom {
		return fmt.Errorf("denom trace %s does not match %s/%s/%s", e.VoucherDenom, e.ChainID, expectedPath, baseDenom)
	}
	return nil
}

func (e IBCDenomTraceSnapshot) validate(channel IBCChannelHandshake, phase string) error {
	if e.Phase != phase || len(e.Traces) != 2 {
		return fmt.Errorf("IBC denom trace snapshot %q has %d traces, want phase %q and 2 traces", e.Phase, len(e.Traces), phase)
	}
	seen := make(map[string]struct{}, 2)
	for _, trace := range e.Traces {
		if _, duplicate := seen[trace.ChainID]; duplicate {
			return fmt.Errorf("IBC denom trace snapshot duplicates chain %q", trace.ChainID)
		}
		seen[trace.ChainID] = struct{}{}
		if err := trace.validate(channel); err != nil {
			return err
		}
	}
	return nil
}

func validateChainNodeRestarts(chainLabel string, restarts []NodeRestartEvidence, want int) error {
	if want < 1 || len(restarts) != want {
		return fmt.Errorf("%s node restart evidence has %d nodes, want %d", chainLabel, len(restarts), want)
	}
	seen := make(map[string]struct{}, len(restarts))
	for _, restart := range restarts {
		if restart.Mode != "graceful" || strings.TrimSpace(restart.Node) == "" ||
			restart.Before.Node != restart.Node || restart.After.Node != restart.Node ||
			restart.Before.Height < 1 || restart.After.Height <= restart.Before.Height ||
			strings.TrimSpace(restart.Before.BlockID) == "" || strings.TrimSpace(restart.Before.AppHash) == "" ||
			strings.TrimSpace(restart.After.BlockID) == "" || strings.TrimSpace(restart.After.AppHash) == "" {
			return fmt.Errorf("%s node %q restart evidence is incomplete", chainLabel, restart.Node)
		}
		if _, duplicate := seen[restart.Node]; duplicate {
			return fmt.Errorf("%s node restart %q is duplicated", chainLabel, restart.Node)
		}
		seen[restart.Node] = struct{}{}
	}
	return nil
}

func (e IBCPostUpgradeTransferEvidence) Validate() error {
	var validationErrors []error
	if e.Phase != ibcPhasePostUpgrade {
		validationErrors = append(validationErrors, fmt.Errorf("post-upgrade evidence phase = %q, want %q", e.Phase, ibcPhasePostUpgrade))
	}
	proxyEscrows := append([]IBCEscrowBalanceEvidence(nil), e.EscrowBalances...)
	for index := range proxyEscrows {
		proxyEscrows[index].Phase = ibcPhasePreUpgrade
	}
	proxy := IBCPreUpgradeTransferEvidence{
		Phase:          ibcPhasePreUpgrade,
		Channel:        e.Channel,
		Transfers:      e.Transfers,
		FinalBalances:  e.FinalBalances,
		EscrowBalances: proxyEscrows,
		DenomTraces:    IBCDenomTraceSnapshot{Phase: ibcPhasePreUpgrade, Traces: e.DenomTraces.Traces},
	}
	if err := proxy.Validate(); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := validateEscrowBalanceSet(e.EscrowBalances, e.Channel, e.Transfers, ibcPhasePostUpgrade); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.DenomTraces.validate(e.Channel, ibcPhasePostUpgrade); err != nil {
		validationErrors = append(validationErrors, err)
	}
	return errors.Join(validationErrors...)
}

func (e IBCUpgradeContinuityEvidence) Validate() error {
	var validationErrors []error
	if e.Phase != ibcUpgradeContinuityPhase {
		validationErrors = append(validationErrors, fmt.Errorf("IBC upgrade continuity phase = %q, want %q", e.Phase, ibcUpgradeContinuityPhase))
	}
	if err := e.OriginalChannel.Validate(); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.InFlight.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("validate in-flight checkpoint: %w", err))
	}
	if e.InFlight.Channel != e.OriginalChannel {
		validationErrors = append(validationErrors, errors.New("in-flight packet does not use the original handshake"))
	}
	if err := e.PanaceaUpgrade.Validate(); err != nil {
		validationErrors = append(validationErrors, err)
	}
	for _, snapshot := range []IBCLinkStateSnapshot{
		e.PostUpgradeBeforeRelay,
		e.AfterInFlightRelay,
		e.FinalAfterHermesRestart,
	} {
		if err := snapshot.ValidateAgainst(e.OriginalChannel); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}
	if e.PostUpgradeBeforeRelay.PanaceaNextSequenceSend != e.InFlight.AfterSendState.PanaceaNextSequenceSend ||
		e.PostUpgradeBeforeRelay.OsmosisNextSequenceSend != e.InFlight.AfterSendState.OsmosisNextSequenceSend {
		validationErrors = append(validationErrors, errors.New("IBC send sequences changed during the Panacea upgrade"))
	}
	if err := e.InFlightRelay.validate(e.InFlight); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if e.AfterInFlightRelay.PanaceaNextSequenceSend != e.PostUpgradeBeforeRelay.PanaceaNextSequenceSend ||
		e.AfterInFlightRelay.OsmosisNextSequenceSend != e.PostUpgradeBeforeRelay.OsmosisNextSequenceSend {
		validationErrors = append(validationErrors, errors.New("relaying the in-flight packet changed a send sequence"))
	}
	if err := e.Timeout.validate(e.OriginalChannel); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if e.Timeout.BeforeSendState != e.AfterInFlightRelay {
		validationErrors = append(validationErrors, errors.New("timeout packet did not start from the post-relay link state"))
	}
	if err := e.PostUpgradeTransfers.Validate(); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if e.PostUpgradeTransfers.Channel != e.OriginalChannel {
		validationErrors = append(validationErrors, errors.New("post-upgrade transfers do not use the original handshake"))
	}
	postByDirection := make(map[string]IBCPacketLifecycleEvidence, len(e.PostUpgradeTransfers.Transfers))
	for _, transfer := range e.PostUpgradeTransfers.Transfers {
		postByDirection[transfer.Direction] = transfer
	}
	panaceaPost, panaceaOK := postByDirection[ibcPanaceaToOsmosis]
	osmosisPost, osmosisOK := postByDirection[ibcOsmosisToPanacea]
	if !panaceaOK || panaceaPost.Sequence != e.Timeout.Packet.Sequence+1 {
		validationErrors = append(validationErrors, errors.New("post-upgrade Panacea packet sequence is not contiguous with the timeout packet"))
	}
	if !osmosisOK || osmosisPost.Sequence != e.Timeout.AfterTimeoutState.OsmosisNextSequenceSend {
		validationErrors = append(validationErrors, errors.New("post-upgrade Osmosis packet sequence is not contiguous"))
	}
	if panaceaOK && e.FinalAfterHermesRestart.PanaceaNextSequenceSend != panaceaPost.Sequence+1 {
		validationErrors = append(validationErrors, errors.New("final Panacea next send sequence is incorrect"))
	}
	if osmosisOK && e.FinalAfterHermesRestart.OsmosisNextSequenceSend != osmosisPost.Sequence+1 {
		validationErrors = append(validationErrors, errors.New("final Osmosis next send sequence is incorrect"))
	}
	if len(e.HermesRestarts) != 2 {
		validationErrors = append(validationErrors, fmt.Errorf("Hermes restart evidence count = %d, want 2", len(e.HermesRestarts)))
	}
	for _, restart := range e.HermesRestarts {
		if err := restart.validate(); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}
	if err := validateChainNodeRestarts("Panacea", e.PanaceaNodeRestarts, len(e.PanaceaUpgrade.PanaceaAfter)); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := validateChainNodeRestarts("Osmosis", e.OsmosisNodeRestarts, len(e.PanaceaUpgrade.OsmosisAfter)); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := validateNodeRestartSemantics(e.NodeRestartSemantics, e); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if len(e.FinalPacketStates) != 2 {
		validationErrors = append(validationErrors, fmt.Errorf("post-restart packet state count = %d, want 2", len(e.FinalPacketStates)))
	} else {
		seenKinds := make(map[string]struct{}, 2)
		for _, terminal := range e.FinalPacketStates {
			if _, duplicate := seenKinds[terminal.Kind]; duplicate {
				validationErrors = append(validationErrors, fmt.Errorf("post-restart packet terminal kind %q is duplicated", terminal.Kind))
			}
			seenKinds[terminal.Kind] = struct{}{}
			if err := terminal.validate(e.InFlight.Packet.Sequence, e.Timeout.Packet.Sequence, e.InFlightRelay.DestinationAcknowledgement); err != nil {
				validationErrors = append(validationErrors, err)
			}
		}
	}
	if len(e.FinalDenomTraces) != 2 {
		validationErrors = append(validationErrors, fmt.Errorf("post-restart denom trace count = %d, want 2", len(e.FinalDenomTraces)))
	} else {
		seenChains := make(map[string]struct{}, 2)
		for _, trace := range e.FinalDenomTraces {
			if _, duplicate := seenChains[trace.ChainID]; duplicate {
				validationErrors = append(validationErrors, fmt.Errorf("post-restart denom trace chain %q is duplicated", trace.ChainID))
			}
			seenChains[trace.ChainID] = struct{}{}
			if err := trace.validate(e.OriginalChannel); err != nil {
				validationErrors = append(validationErrors, err)
			}
		}
	}
	if err := validateFinalRestartBalances(e.PostUpgradeTransfers.FinalBalances, e.FinalBalances); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := validateDenomTraceContinuity(e.DenomTraceContinuity, e.OriginalChannel); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := validateFinalEscrowBalances(e.PostUpgradeTransfers.EscrowBalances, e.FinalEscrowBalances, e.OriginalChannel); err != nil {
		validationErrors = append(validationErrors, err)
	}
	return errors.Join(validationErrors...)
}

func validateNodeRestartSemantics(groups []IBCChainRestartEvidence, continuity IBCUpgradeContinuityEvidence) error {
	if len(groups) != 2 {
		return fmt.Errorf("IBC post-restart node semantics has %d chains, want 2", len(groups))
	}
	wantCount := map[string]int{
		continuity.OriginalChannel.Panacea.ChainID: len(continuity.PanaceaUpgrade.PanaceaAfter),
		continuity.OriginalChannel.Osmosis.ChainID: len(continuity.PanaceaUpgrade.OsmosisAfter),
	}
	wantSequence := map[string]uint64{
		continuity.OriginalChannel.Panacea.ChainID: continuity.FinalAfterHermesRestart.PanaceaNextSequenceSend,
		continuity.OriginalChannel.Osmosis.ChainID: continuity.FinalAfterHermesRestart.OsmosisNextSequenceSend,
	}
	seenChains := make(map[string]struct{}, 2)
	for _, group := range groups {
		count, ok := wantCount[group.ChainID]
		if !ok {
			return fmt.Errorf("IBC post-restart semantics uses unknown chain %q", group.ChainID)
		}
		if _, duplicate := seenChains[group.ChainID]; duplicate {
			return fmt.Errorf("IBC post-restart semantics duplicates chain %q", group.ChainID)
		}
		seenChains[group.ChainID] = struct{}{}
		if len(group.Nodes) != count {
			return fmt.Errorf("IBC post-restart semantics for %q has %d nodes, want %d", group.ChainID, len(group.Nodes), count)
		}
		seenNodes := make(map[string]struct{}, count)
		for _, node := range group.Nodes {
			if node.ChainID != group.ChainID || node.IdentityBefore != node.IdentityAfter || node.Restart.Node != node.IdentityAfter.Name ||
				node.ObservedHeight < node.Restart.After.Height || !ibcStateEquals(node.ClientStatus, "ACTIVE") ||
				!ibcStateEquals(node.ConnectionState, "OPEN") || !ibcStateEquals(node.ChannelState, "OPEN") ||
				node.NextSequenceSend != wantSequence[group.ChainID] {
				return fmt.Errorf("IBC post-restart node semantics for %q/%q is incomplete", group.ChainID, node.IdentityAfter.Name)
			}
			if _, duplicate := seenNodes[node.IdentityAfter.Name]; duplicate {
				return fmt.Errorf("IBC post-restart semantics duplicates node %q", node.IdentityAfter.Name)
			}
			seenNodes[node.IdentityAfter.Name] = struct{}{}
			if err := node.DenomTrace.validate(continuity.OriginalChannel); err != nil {
				return err
			}
			packets := node.PacketSemantics
			if packets.SuccessSequence != continuity.InFlight.Packet.Sequence || packets.TimeoutSequence != continuity.Timeout.Packet.Sequence {
				return errors.New("IBC post-restart node packet sequences do not match continuity packets")
			}
			switch group.ChainID {
			case continuity.OriginalChannel.Panacea.ChainID:
				if packets.Role != "source" || packets.SuccessCommitmentPresent || packets.TimeoutCommitmentPresent ||
					packets.SuccessReceipt || packets.TimeoutReceipt || packets.SuccessAcknowledgement != "" || packets.TimeoutAcknowledgement != "" {
					return fmt.Errorf("Panacea node %q packet source terminal semantics are invalid", node.IdentityAfter.Name)
				}
			case continuity.OriginalChannel.Osmosis.ChainID:
				if packets.Role != "destination" || packets.SuccessCommitmentPresent || packets.TimeoutCommitmentPresent ||
					!packets.SuccessReceipt || packets.TimeoutReceipt || packets.SuccessAcknowledgement != continuity.InFlightRelay.DestinationAcknowledgement ||
					packets.TimeoutAcknowledgement != "" {
					return fmt.Errorf("Osmosis node %q packet destination terminal semantics are invalid", node.IdentityAfter.Name)
				}
			}
		}
	}
	return nil
}

func validateDenomTraceContinuity(snapshots []IBCDenomTraceSnapshot, channel IBCChannelHandshake) error {
	wantPhases := []string{ibcPhasePreUpgrade, ibcPhasePostUpgrade, "post-restart"}
	if len(snapshots) != len(wantPhases) {
		return fmt.Errorf("IBC denom trace continuity has %d snapshots, want %d", len(snapshots), len(wantPhases))
	}
	var baseline map[string]IBCDenomTraceEvidence
	for index, phase := range wantPhases {
		snapshot := snapshots[index]
		if err := snapshot.validate(channel, phase); err != nil {
			return err
		}
		current := make(map[string]IBCDenomTraceEvidence, 2)
		for _, trace := range snapshot.Traces {
			current[trace.ChainID] = trace
		}
		if index == 0 {
			baseline = current
			continue
		}
		for chainID, original := range baseline {
			if current[chainID] != original {
				return fmt.Errorf("IBC denom trace for %q changed between pre-upgrade and %s", chainID, phase)
			}
		}
	}
	return nil
}

func validateFinalEscrowBalances(before, after []IBCEscrowBalanceEvidence, channel IBCChannelHandshake) error {
	if len(before) != 2 || len(after) != 2 {
		return fmt.Errorf("post-restart escrow balances have before:%d after:%d entries, want 2 each", len(before), len(after))
	}
	beforeByChain := make(map[string]IBCEscrowBalanceEvidence, 2)
	for _, escrow := range before {
		beforeByChain[escrow.ChainID] = escrow
	}
	seen := make(map[string]struct{}, 2)
	for _, escrow := range after {
		var endpoint IBCChannelEndpoint
		switch escrow.ChainID {
		case channel.Panacea.ChainID:
			endpoint = channel.Panacea
		case channel.Osmosis.ChainID:
			endpoint = channel.Osmosis
		default:
			return fmt.Errorf("post-restart escrow uses unknown chain %q", escrow.ChainID)
		}
		if err := escrow.validate(endpoint, "post-restart"); err != nil {
			return err
		}
		prior, ok := beforeByChain[escrow.ChainID]
		if !ok || escrow.ExpectedDelta != "0" || escrow.Before != prior.After || escrow.After != prior.After ||
			escrow.Address != prior.Address || escrow.Denom != prior.Denom {
			return fmt.Errorf("post-restart escrow for %q did not preserve the post-transfer value", escrow.ChainID)
		}
		if _, duplicate := seen[escrow.ChainID]; duplicate {
			return fmt.Errorf("post-restart escrow for %q is duplicated", escrow.ChainID)
		}
		seen[escrow.ChainID] = struct{}{}
	}
	return nil
}

func validateSourceSendBalance(balance IBCBalanceEvidence, amount, fee string) error {
	before, beforeOK := new(big.Int).SetString(balance.Before, 10)
	transfer, transferOK := new(big.Int).SetString(amount, 10)
	gasFee, feeOK := new(big.Int).SetString(fee, 10)
	if !beforeOK || !transferOK || !feeOK {
		return errors.New("source send balance inputs are invalid")
	}
	expected := new(big.Int).Sub(before, transfer)
	expected.Sub(expected, gasFee)
	if expected.Sign() < 0 || expected.String() != balance.After {
		return fmt.Errorf("source send balance after = %s, want %s", balance.After, expected.String())
	}
	return nil
}

func validateSourceRefundBalance(balance IBCBalanceEvidence, fee string) error {
	before, beforeOK := new(big.Int).SetString(balance.Before, 10)
	gasFee, feeOK := new(big.Int).SetString(fee, 10)
	if !beforeOK || !feeOK {
		return errors.New("source refund balance inputs are invalid")
	}
	expected := new(big.Int).Sub(before, gasFee)
	if expected.Sign() < 0 || expected.String() != balance.After {
		return fmt.Errorf("source refund balance after = %s, want %s", balance.After, expected.String())
	}
	return nil
}

func acknowledgementCommitmentBase64(encodedAcknowledgement string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encodedAcknowledgement)
	if err != nil {
		return "", fmt.Errorf("decode acknowledgement before commitment: %w", err)
	}
	if err := validateSuccessfulAcknowledgement(encodedAcknowledgement); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(channeltypes.CommitAcknowledgement(raw)), nil
}

func validateFinalRestartBalances(before, after []IBCBalanceEvidence) error {
	if len(before) != 4 || len(after) != 4 {
		return fmt.Errorf("post-restart balances have before:%d after:%d entries, want 4 each", len(before), len(after))
	}
	beforeByKey := make(map[string]IBCBalanceEvidence, len(before))
	for _, balance := range before {
		beforeByKey[strings.Join([]string{balance.ChainID, balance.Address, balance.Denom}, "\x00")] = balance
	}
	seen := make(map[string]struct{}, len(after))
	for _, balance := range after {
		if err := balance.validate(); err != nil {
			return err
		}
		key := strings.Join([]string{balance.ChainID, balance.Address, balance.Denom}, "\x00")
		prior, ok := beforeByKey[key]
		if !ok || balance.Before != prior.After || balance.After != prior.After {
			return fmt.Errorf("post-restart balance %q did not preserve the post-transfer value", key)
		}
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("post-restart balance %q is duplicated", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateDestinationReceiveBalance(balance IBCBalanceEvidence, amount string) error {
	before, beforeOK := new(big.Int).SetString(balance.Before, 10)
	transfer, transferOK := new(big.Int).SetString(amount, 10)
	if !beforeOK || !transferOK {
		return errors.New("destination receive balance inputs are invalid")
	}
	expected := new(big.Int).Add(before, transfer)
	if expected.String() != balance.After {
		return fmt.Errorf("destination receive balance after = %s, want %s", balance.After, expected.String())
	}
	return nil
}

func lifecycleMatchesSend(lifecycle IBCPacketLifecycleEvidence, send IBCPacketSendEvidence) bool {
	return lifecycle.Direction == send.Direction &&
		lifecycle.SourceChainID == send.SourceChainID &&
		lifecycle.DestinationChainID == send.DestinationChainID &&
		lifecycle.TxHash == send.TxHash &&
		lifecycle.TxHeight == send.TxHeight &&
		lifecycle.Sequence == send.Sequence &&
		lifecycle.SourcePort == send.SourcePort &&
		lifecycle.SourceChannel == send.SourceChannel &&
		lifecycle.DestinationPort == send.DestinationPort &&
		lifecycle.DestinationChannel == send.DestinationChannel &&
		lifecycle.Denom == send.Denom &&
		lifecycle.Amount == send.Amount
}

func validateChangedNodeIdentities(before, after []IBCNodeRuntimeIdentity, from, to ImageRef) error {
	if len(before) == 0 || len(before) != len(after) {
		return errors.New("Panacea node identity sets are incomplete")
	}
	afterByName, err := nodeIdentityMap(after)
	if err != nil {
		return err
	}
	for _, oldNode := range before {
		newNode, ok := afterByName[oldNode.Name]
		if !ok {
			return fmt.Errorf("upgraded Panacea node %s is missing", oldNode.Name)
		}
		if oldNode.Image != from || newNode.Image != to || oldNode.ContainerID == newNode.ContainerID {
			return fmt.Errorf("Panacea node %s did not change only from the declared old image to current", oldNode.Name)
		}
	}
	_, err = nodeIdentityMap(before)
	return err
}

func validateUnchangedNodeIdentities(before, after []IBCNodeRuntimeIdentity) error {
	if len(before) == 0 || len(before) != len(after) {
		return errors.New("Osmosis node identity sets are incomplete")
	}
	beforeByName, err := nodeIdentityMap(before)
	if err != nil {
		return err
	}
	afterByName, err := nodeIdentityMap(after)
	if err != nil {
		return err
	}
	for name, oldNode := range beforeByName {
		if newNode, ok := afterByName[name]; !ok || newNode != oldNode {
			return fmt.Errorf("Osmosis node %s changed during the Panacea-only upgrade", name)
		}
	}
	return nil
}

func nodeIdentityMap(nodes []IBCNodeRuntimeIdentity) (map[string]IBCNodeRuntimeIdentity, error) {
	result := make(map[string]IBCNodeRuntimeIdentity, len(nodes))
	for _, node := range nodes {
		if strings.TrimSpace(node.Name) == "" || strings.TrimSpace(node.ContainerID) == "" ||
			strings.TrimSpace(node.Image.Repository) == "" || strings.TrimSpace(node.Image.Version) == "" {
			return nil, errors.New("IBC node runtime identity is incomplete")
		}
		if _, duplicate := result[node.Name]; duplicate {
			return nil, fmt.Errorf("IBC node runtime identity %s is duplicated", node.Name)
		}
		result[node.Name] = node
	}
	return result, nil
}

func imageRefFromDocker(image ibc.DockerImage) ImageRef {
	return ImageRef{Repository: image.Repository, Version: image.Version}
}
