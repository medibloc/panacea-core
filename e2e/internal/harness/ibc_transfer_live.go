package harness

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

const ibcPacketPollHeightLimit int64 = 100

// IBCPreUpgradeTransferRequest supplies one independently funded wallet per
// chain. Each wallet sends the chain's native denomination and receives the
// counterparty voucher; private wallet material is never written to artifacts.
type IBCPreUpgradeTransferRequest struct {
	PanaceaUser ibc.Wallet
	OsmosisUser ibc.Wallet
	Amount      sdkmath.Int
}

type ibcPacketStageEvidence struct {
	Stage      string                     `json:"stage"`
	RecordedAt string                     `json:"recorded_at"`
	Transfer   IBCPacketLifecycleEvidence `json:"transfer"`
}

type ibcPacketAckAtHeight struct {
	Height int64
	Ack    ibc.PacketAcknowledgement
}

// OpenTransferChannel performs the real Hermes client, connection, and
// channel handshakes and snapshots both chain views. It can be attempted only
// once on a topology so a failed partial handshake cannot be hidden by
// creating replacement identifiers in the same run.
func (n *IBCTopology) OpenTransferChannel(ctx context.Context) (IBCChannelHandshake, error) {
	var zero IBCChannelHandshake
	if ctx == nil {
		return zero, errors.New("IBC channel context is required")
	}
	if n == nil {
		return zero, errors.New("IBC topology is required")
	}

	n.lifecycleMu.Lock()
	defer n.lifecycleMu.Unlock()
	if err := n.validateTransferRuntime(); err != nil {
		return zero, err
	}
	if n.channel != nil {
		return zero, errors.New("IBC transfer channel is already open")
	}
	if n.channelAttempted {
		return zero, errors.New("IBC transfer channel handshake was already attempted; start a fresh topology instead of replacing its identifiers")
	}
	n.channelAttempted = true

	panaceaChainID := n.Panacea.Config().ChainID
	osmosisChainID := n.Osmosis.Config().ChainID
	steps := []struct {
		stage string
		run   func() error
	}{
		{
			stage: "generate-path",
			run: func() error {
				return n.Relayer.GeneratePath(ctx, n.execReporter, panaceaChainID, osmosisChainID, n.Path)
			},
		},
		{
			stage: "create-clients",
			run: func() error {
				return n.Relayer.CreateClients(ctx, n.execReporter, n.Path, ibc.DefaultClientOpts())
			},
		},
		{
			stage: "wait-for-client-blocks",
			run: func() error {
				return testutil.WaitForBlocks(ctx, 2, n.Panacea, n.Osmosis)
			},
		},
		{
			stage: "create-connections",
			run: func() error {
				return n.Relayer.CreateConnections(ctx, n.execReporter, n.Path)
			},
		},
		{
			stage: "create-channel",
			run: func() error {
				return n.Relayer.CreateChannel(ctx, n.execReporter, n.Path, ibc.DefaultChannelOpts())
			},
		},
	}
	for _, step := range steps {
		if err := step.run(); err != nil {
			return zero, n.ibcOperationError("channel-"+step.stage, err)
		}
	}

	panaceaEndpoint, err := snapshotIBCChannelEndpoint(
		ctx,
		n.Relayer,
		n.execReporter,
		panaceaChainID,
		osmosisChainID,
	)
	if err != nil {
		return zero, n.ibcOperationError("channel-query-panacea", err)
	}
	osmosisEndpoint, err := snapshotIBCChannelEndpoint(
		ctx,
		n.Relayer,
		n.execReporter,
		osmosisChainID,
		panaceaChainID,
	)
	if err != nil {
		return zero, n.ibcOperationError("channel-query-osmosis", err)
	}
	handshake := IBCChannelHandshake{
		Path:    n.Path,
		Panacea: panaceaEndpoint,
		Osmosis: osmosisEndpoint,
	}
	if err := handshake.Validate(); err != nil {
		return zero, n.ibcOperationError("channel-validate", err)
	}
	if err := n.artifacts.base.writeJSON("ibc/state/pre-upgrade-channel.json", handshake); err != nil {
		return zero, n.ibcOperationError("channel-artifact", err)
	}
	if err := n.artifacts.pinHermesTransferChannels(ctx, n.hermes, n.execReporter, handshake); err != nil {
		return zero, n.ibcOperationError("channel-filter-hermes", err)
	}

	n.channel = &handshake
	return handshake, nil
}

// TransferChannel returns a copy of the exact handshake retained by this
// topology. Future post-upgrade slices use it to reject replacement paths.
func (n *IBCTopology) TransferChannel() (IBCChannelHandshake, bool) {
	if n == nil {
		return IBCChannelHandshake{}, false
	}
	n.lifecycleMu.Lock()
	defer n.lifecycleMu.Unlock()
	if n.channel == nil {
		return IBCChannelHandshake{}, false
	}
	return *n.channel, true
}

func snapshotIBCChannelEndpoint(
	ctx context.Context,
	relayer ibc.Relayer,
	reporter ibc.RelayerExecReporter,
	chainID string,
	counterpartyChainID string,
) (IBCChannelEndpoint, error) {
	var zero IBCChannelEndpoint
	channel, err := ibc.GetTransferChannel(ctx, relayer, reporter, chainID, counterpartyChainID)
	if err != nil {
		return zero, fmt.Errorf("get unique transfer channel on %s: %w", chainID, err)
	}
	if len(channel.ConnectionHops) != 1 {
		return zero, fmt.Errorf("transfer channel %s/%s has %d connection hops, want 1", channel.PortID, channel.ChannelID, len(channel.ConnectionHops))
	}

	connections, err := relayer.GetConnections(ctx, reporter, chainID)
	if err != nil {
		return zero, fmt.Errorf("get connections on %s: %w", chainID, err)
	}
	var connection *ibc.ConnectionOutput
	for _, candidate := range connections {
		if candidate != nil && candidate.ID == channel.ConnectionHops[0] {
			if connection != nil {
				return zero, fmt.Errorf("multiple connections named %s on %s", candidate.ID, chainID)
			}
			connection = candidate
		}
	}
	if connection == nil {
		return zero, fmt.Errorf("connection %s for transfer channel %s/%s was not found on %s", channel.ConnectionHops[0], channel.PortID, channel.ChannelID, chainID)
	}
	if connection.Counterparty == nil {
		return zero, fmt.Errorf("connection %s on %s has no counterparty", connection.ID, chainID)
	}

	clients, err := relayer.GetClients(ctx, reporter, chainID)
	if err != nil {
		return zero, fmt.Errorf("get clients on %s: %w", chainID, err)
	}
	matchedClient := false
	for _, client := range clients {
		if client == nil || client.ClientID != connection.ClientID {
			continue
		}
		if matchedClient {
			return zero, fmt.Errorf("multiple clients named %s on %s", connection.ClientID, chainID)
		}
		matchedClient = true
		if client.ClientState.ChainID != counterpartyChainID {
			return zero, fmt.Errorf("client %s on %s tracks %s, want %s", client.ClientID, chainID, client.ClientState.ChainID, counterpartyChainID)
		}
	}
	if !matchedClient {
		return zero, fmt.Errorf("client %s for connection %s was not found on %s", connection.ClientID, connection.ID, chainID)
	}

	return IBCChannelEndpoint{
		ChainID:                  chainID,
		CounterpartyChainID:      counterpartyChainID,
		ClientID:                 connection.ClientID,
		CounterpartyClientID:     connection.Counterparty.ClientId,
		ConnectionID:             connection.ID,
		CounterpartyConnectionID: connection.Counterparty.ConnectionId,
		ConnectionState:          connection.State,
		PortID:                   channel.PortID,
		ChannelID:                channel.ChannelID,
		CounterpartyChannelID:    channel.Counterparty.ChannelID,
		ChannelState:             channel.State,
		Ordering:                 channel.Ordering,
		Version:                  channel.Version,
	}, nil
}

// RunPreUpgradeBidirectionalTransfers relays one real ICS-20 transfer in each
// direction over the retained channel. It records packet sequence, committed
// receive/acknowledgement heights, raw acknowledgement, and exact balances.
func (n *IBCTopology) RunPreUpgradeBidirectionalTransfers(
	ctx context.Context,
	req IBCPreUpgradeTransferRequest,
) (IBCPreUpgradeTransferEvidence, error) {
	var zero IBCPreUpgradeTransferEvidence
	if ctx == nil {
		return zero, errors.New("IBC transfer context is required")
	}
	if n == nil {
		return zero, errors.New("IBC topology is required")
	}

	n.lifecycleMu.Lock()
	defer n.lifecycleMu.Unlock()
	if err := n.validateTransferRuntime(); err != nil {
		return zero, err
	}
	if n.channel == nil {
		return zero, errors.New("open the IBC transfer channel before sending packets")
	}
	if n.preUpgradeTransferAttempted {
		return zero, errors.New("pre-upgrade bidirectional IBC transfer was already attempted on this topology")
	}
	if req.PanaceaUser == nil || req.OsmosisUser == nil {
		return zero, errors.New("one funded IBC user per chain is required")
	}
	if req.Amount.IsNil() || !req.Amount.IsPositive() {
		return zero, errors.New("IBC transfer amount must be positive")
	}

	handshake := *n.channel
	panaceaAddress := req.PanaceaUser.FormattedAddress()
	osmosisAddress := req.OsmosisUser.FormattedAddress()
	if strings.TrimSpace(panaceaAddress) == "" || strings.TrimSpace(osmosisAddress) == "" {
		return zero, n.ibcOperationError("pre-upgrade-wallet-address", errors.New("IBC wallet formatted addresses are required"))
	}
	panaceaDenom := n.Panacea.Config().Denom
	osmosisDenom := n.Osmosis.Config().Denom
	medOnOsmosis := ibcVoucherDenom(handshake.Osmosis, panaceaDenom)
	osmoOnPanacea := ibcVoucherDenom(handshake.Panacea, osmosisDenom)

	panaceaNativeBefore, err := n.Panacea.GetBalance(ctx, panaceaAddress, panaceaDenom)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-balance-panacea-native-before", err)
	}
	panaceaVoucherBefore, err := n.Panacea.GetBalance(ctx, panaceaAddress, osmoOnPanacea)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-balance-panacea-voucher-before", err)
	}
	osmosisNativeBefore, err := n.Osmosis.GetBalance(ctx, osmosisAddress, osmosisDenom)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-balance-osmosis-native-before", err)
	}
	osmosisVoucherBefore, err := n.Osmosis.GetBalance(ctx, osmosisAddress, medOnOsmosis)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-balance-osmosis-voucher-before", err)
	}
	panaceaEscrowAddress, panaceaEscrowBefore, err := queryIBCEscrowBalance(ctx, n.Panacea, handshake.Panacea, panaceaDenom)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-escrow-panacea-before", err)
	}
	osmosisEscrowAddress, osmosisEscrowBefore, err := queryIBCEscrowBalance(ctx, n.Osmosis, handshake.Osmosis, osmosisDenom)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-escrow-osmosis-before", err)
	}

	n.preUpgradeTransferAttempted = true
	if err := n.Relayer.StartRelayer(ctx, n.execReporter, n.Path); err != nil {
		return zero, n.ibcOperationError("pre-upgrade-hermes-start", err)
	}
	n.relayerStarted = true

	panaceaTransfer, panaceaFee, err := n.executeIBCTransfer(
		ctx,
		ibcPhasePreUpgrade,
		ibcPanaceaToOsmosis,
		n.Panacea,
		n.Osmosis,
		req.PanaceaUser,
		req.OsmosisUser,
		handshake.Panacea,
		handshake.Osmosis,
		panaceaDenom,
		req.Amount,
	)
	if err != nil {
		return zero, err
	}
	osmosisTransfer, osmosisFee, err := n.executeIBCTransfer(
		ctx,
		ibcPhasePreUpgrade,
		ibcOsmosisToPanacea,
		n.Osmosis,
		n.Panacea,
		req.OsmosisUser,
		req.PanaceaUser,
		handshake.Osmosis,
		handshake.Panacea,
		osmosisDenom,
		req.Amount,
	)
	if err != nil {
		return zero, err
	}

	panaceaNativeExpected := panaceaNativeBefore.Sub(req.Amount).SubRaw(panaceaFee)
	osmosisNativeExpected := osmosisNativeBefore.Sub(req.Amount).SubRaw(osmosisFee)
	panaceaVoucherExpected := panaceaVoucherBefore.Add(req.Amount)
	osmosisVoucherExpected := osmosisVoucherBefore.Add(req.Amount)
	if panaceaNativeExpected.IsNegative() || osmosisNativeExpected.IsNegative() {
		return zero, n.ibcOperationError("pre-upgrade-balance-expected", errors.New("IBC source balance is insufficient for transfer amount and gas fee"))
	}

	for _, expected := range []struct {
		name    string
		chain   *cosmos.CosmosChain
		balance ibc.WalletAmount
	}{
		{
			name:  "panacea-native",
			chain: n.Panacea,
			balance: ibc.WalletAmount{
				Address: panaceaAddress,
				Denom:   panaceaDenom,
				Amount:  panaceaNativeExpected,
			},
		},
		{
			name:  "panacea-osmosis-voucher",
			chain: n.Panacea,
			balance: ibc.WalletAmount{
				Address: panaceaAddress,
				Denom:   osmoOnPanacea,
				Amount:  panaceaVoucherExpected,
			},
		},
		{
			name:  "osmosis-native",
			chain: n.Osmosis,
			balance: ibc.WalletAmount{
				Address: osmosisAddress,
				Denom:   osmosisDenom,
				Amount:  osmosisNativeExpected,
			},
		},
		{
			name:  "osmosis-panacea-voucher",
			chain: n.Osmosis,
			balance: ibc.WalletAmount{
				Address: osmosisAddress,
				Denom:   medOnOsmosis,
				Amount:  osmosisVoucherExpected,
			},
		},
	} {
		if err := cosmos.PollForBalance(ctx, expected.chain, 50, expected.balance); err != nil {
			return zero, n.ibcOperationError("pre-upgrade-balance-poll-"+expected.name, err)
		}
	}

	panaceaNativeAfter, err := n.Panacea.GetBalance(ctx, panaceaAddress, panaceaDenom)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-balance-panacea-native-after", err)
	}
	panaceaVoucherAfter, err := n.Panacea.GetBalance(ctx, panaceaAddress, osmoOnPanacea)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-balance-panacea-voucher-after", err)
	}
	osmosisNativeAfter, err := n.Osmosis.GetBalance(ctx, osmosisAddress, osmosisDenom)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-balance-osmosis-native-after", err)
	}
	osmosisVoucherAfter, err := n.Osmosis.GetBalance(ctx, osmosisAddress, medOnOsmosis)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-balance-osmosis-voucher-after", err)
	}
	_, panaceaEscrowAfter, err := queryIBCEscrowBalance(ctx, n.Panacea, handshake.Panacea, panaceaDenom)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-escrow-panacea-after", err)
	}
	_, osmosisEscrowAfter, err := queryIBCEscrowBalance(ctx, n.Osmosis, handshake.Osmosis, osmosisDenom)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-escrow-osmosis-after", err)
	}
	preUpgradeTraces, err := n.queryDenomTraceSnapshot(ctx, ibcPhasePreUpgrade)
	if err != nil {
		return zero, n.ibcOperationError("pre-upgrade-denom-traces", err)
	}

	evidence := IBCPreUpgradeTransferEvidence{
		Phase:     ibcPhasePreUpgrade,
		Channel:   handshake,
		Transfers: []IBCPacketLifecycleEvidence{panaceaTransfer, osmosisTransfer},
		FinalBalances: []IBCBalanceEvidence{
			{
				ChainID:       handshake.Panacea.ChainID,
				Address:       panaceaAddress,
				Denom:         panaceaDenom,
				Before:        panaceaNativeBefore.String(),
				After:         panaceaNativeAfter.String(),
				ExpectedAfter: panaceaNativeExpected.String(),
			},
			{
				ChainID:       handshake.Panacea.ChainID,
				Address:       panaceaAddress,
				Denom:         osmoOnPanacea,
				Before:        panaceaVoucherBefore.String(),
				After:         panaceaVoucherAfter.String(),
				ExpectedAfter: panaceaVoucherExpected.String(),
			},
			{
				ChainID:       handshake.Osmosis.ChainID,
				Address:       osmosisAddress,
				Denom:         osmosisDenom,
				Before:        osmosisNativeBefore.String(),
				After:         osmosisNativeAfter.String(),
				ExpectedAfter: osmosisNativeExpected.String(),
			},
			{
				ChainID:       handshake.Osmosis.ChainID,
				Address:       osmosisAddress,
				Denom:         medOnOsmosis,
				Before:        osmosisVoucherBefore.String(),
				After:         osmosisVoucherAfter.String(),
				ExpectedAfter: osmosisVoucherExpected.String(),
			},
		},
		EscrowBalances: []IBCEscrowBalanceEvidence{
			newIBCEscrowBalanceEvidence(ibcPhasePreUpgrade, handshake.Panacea, panaceaEscrowAddress, panaceaDenom, panaceaEscrowBefore, panaceaEscrowAfter, req.Amount),
			newIBCEscrowBalanceEvidence(ibcPhasePreUpgrade, handshake.Osmosis, osmosisEscrowAddress, osmosisDenom, osmosisEscrowBefore, osmosisEscrowAfter, req.Amount),
		},
		DenomTraces: preUpgradeTraces,
	}
	if err := evidence.Validate(); err != nil {
		return zero, n.ibcOperationError("pre-upgrade-evidence-validate", err)
	}
	if err := n.artifacts.base.writeJSON("ibc/state/pre-upgrade-bidirectional.json", evidence); err != nil {
		return zero, n.ibcOperationError("pre-upgrade-evidence-artifact", err)
	}
	n.preUpgradeTransferComplete = true
	n.preUpgradeTransferEvidence = &evidence
	return evidence, nil
}

func (n *IBCTopology) executeIBCTransfer(
	ctx context.Context,
	phase string,
	direction string,
	source *cosmos.CosmosChain,
	destination *cosmos.CosmosChain,
	sourceUser ibc.Wallet,
	destinationUser ibc.Wallet,
	sourceEndpoint IBCChannelEndpoint,
	destinationEndpoint IBCChannelEndpoint,
	denom string,
	amount sdkmath.Int,
) (IBCPacketLifecycleEvidence, int64, error) {
	var zero IBCPacketLifecycleEvidence
	destinationStartHeight, err := destination.Height(ctx)
	if err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-destination-height", err)
	}
	tx, err := source.SendIBCTransfer(ctx, sourceEndpoint.ChannelID, sourceUser.KeyName(), ibc.WalletAmount{
		Address: destinationUser.FormattedAddress(),
		Denom:   denom,
		Amount:  amount,
	}, ibc.TransferOptions{})
	if err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-send", err)
	}
	if err := tx.Validate(); err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-send-validate", err)
	}
	if tx.Packet.SourcePort != sourceEndpoint.PortID ||
		tx.Packet.SourceChannel != sourceEndpoint.ChannelID ||
		tx.Packet.DestPort != destinationEndpoint.PortID ||
		tx.Packet.DestChannel != destinationEndpoint.ChannelID {
		return zero, 0, n.ibcOperationError(
			phase+"-"+direction+"-packet-channel",
			fmt.Errorf(
				"send packet endpoint = %s/%s -> %s/%s, want %s/%s -> %s/%s",
				tx.Packet.SourcePort,
				tx.Packet.SourceChannel,
				tx.Packet.DestPort,
				tx.Packet.DestChannel,
				sourceEndpoint.PortID,
				sourceEndpoint.ChannelID,
				destinationEndpoint.PortID,
				destinationEndpoint.ChannelID,
			),
		)
	}

	transfer := IBCPacketLifecycleEvidence{
		Direction:          direction,
		SourceChainID:      sourceEndpoint.ChainID,
		DestinationChainID: destinationEndpoint.ChainID,
		TxHash:             tx.TxHash,
		TxHeight:           tx.Height,
		Sequence:           tx.Packet.Sequence,
		SourcePort:         tx.Packet.SourcePort,
		SourceChannel:      tx.Packet.SourceChannel,
		DestinationPort:    tx.Packet.DestPort,
		DestinationChannel: tx.Packet.DestChannel,
		Denom:              denom,
		Amount:             amount.String(),
		PacketData:         base64.StdEncoding.EncodeToString(tx.Packet.Data),
		TimeoutHeight:      tx.Packet.TimeoutHeight,
		TimeoutTimestamp:   uint64(tx.Packet.TimeoutTimestamp),
	}
	if err := n.recordIBCPacketStage(phase, "send", transfer); err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-send-artifact", err)
	}

	recvHeight, err := pollForRecvPacket(
		ctx,
		destination,
		destinationStartHeight,
		destinationStartHeight+ibcPacketPollHeightLimit,
		tx.Packet,
	)
	if err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-recv", err)
	}
	transfer.Recv = IBCPacketObservation{
		Observed: true,
		ChainID:  destinationEndpoint.ChainID,
		Height:   recvHeight,
	}
	if err := n.recordIBCPacketStage(phase, "recv", transfer); err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-recv-artifact", err)
	}

	ackAtHeight, err := pollForPacketAcknowledgement(
		ctx,
		source,
		tx.Height,
		tx.Height+ibcPacketPollHeightLimit,
		tx.Packet,
	)
	if err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-ack", err)
	}
	if err := ackAtHeight.Ack.Validate(); err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-ack-validate", err)
	}
	transfer.Ack = IBCAcknowledgementObservation{
		Observed:        true,
		ChainID:         sourceEndpoint.ChainID,
		Height:          ackAtHeight.Height,
		Acknowledgement: base64.StdEncoding.EncodeToString(ackAtHeight.Ack.Acknowledgement),
	}
	if err := validateSuccessfulAcknowledgement(transfer.Ack.Acknowledgement); err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-ack-result", err)
	}
	if err := n.recordIBCPacketStage(phase, "ack", transfer); err != nil {
		return zero, 0, n.ibcOperationError(phase+"-"+direction+"-ack-artifact", err)
	}
	return transfer, source.GetGasFeesInNativeDenom(tx.GasSpent), nil
}

func pollForRecvPacket(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	startHeight int64,
	maxHeight int64,
	want ibc.Packet,
) (int64, error) {
	poller := testutil.BlockPoller[int64]{
		CurrentHeight: stableIBCQueryHeight(chain.Height),
		PollFunc: func(ctx context.Context, height int64) (int64, error) {
			values, err := successfulIBCMessageValues(ctx, chain, height, "/ibc.core.channel.v1.MsgRecvPacket")
			if err != nil {
				return 0, err
			}
			for _, value := range values {
				var recv channeltypes.MsgRecvPacket
				if err := proto.Unmarshal(value, &recv); err != nil {
					return 0, fmt.Errorf("decode successful MsgRecvPacket at height %d: %w", height, err)
				}
				if recvPacketMatches(recv.Packet, want) {
					return height, nil
				}
			}
			return 0, testutil.ErrNotFound
		},
	}
	height, err := poller.DoPoll(ctx, startHeight, maxHeight)
	if err != nil {
		return 0, fmt.Errorf("find receive for packet sequence %d between heights %d and %d: %w", want.Sequence, startHeight, maxHeight, err)
	}
	return height, nil
}

func pollForPacketAcknowledgement(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	startHeight int64,
	maxHeight int64,
	want ibc.Packet,
) (ibcPacketAckAtHeight, error) {
	var zero ibcPacketAckAtHeight
	poller := testutil.BlockPoller[ibcPacketAckAtHeight]{
		CurrentHeight: stableIBCQueryHeight(chain.Height),
		PollFunc: func(ctx context.Context, height int64) (ibcPacketAckAtHeight, error) {
			values, err := successfulIBCMessageValues(ctx, chain, height, "/ibc.core.channel.v1.MsgAcknowledgement")
			if err != nil {
				return zero, err
			}
			for _, value := range values {
				var message channeltypes.MsgAcknowledgement
				if err := proto.Unmarshal(value, &message); err != nil {
					return zero, fmt.Errorf("decode successful MsgAcknowledgement at height %d: %w", height, err)
				}
				ack := ibc.PacketAcknowledgement{
					Packet:          toInterchaintestPacket(message.Packet),
					Acknowledgement: append([]byte(nil), message.Acknowledgement...),
				}
				if ack.Packet.Equal(want) {
					return ibcPacketAckAtHeight{Height: height, Ack: ack}, nil
				}
			}
			return zero, testutil.ErrNotFound
		},
	}
	ack, err := poller.DoPoll(ctx, startHeight, maxHeight)
	if err != nil {
		return zero, fmt.Errorf("find acknowledgement for packet sequence %d between heights %d and %d: %w", want.Sequence, startHeight, maxHeight, err)
	}
	return ack, nil
}

// stableIBCQueryHeight keeps the block currently reported as the RPC tip out
// of packet scans. CometBFT exposes the new height before the corresponding
// BlockResults record is durably queryable. Advancing a BlockPoller cursor on
// that transient error permanently skips the block that contains the packet.
func stableIBCQueryHeight(
	currentHeight func(context.Context) (int64, error),
) func(context.Context) (int64, error) {
	return func(ctx context.Context) (int64, error) {
		height, err := currentHeight(ctx)
		if err != nil {
			return 0, err
		}
		return stableIBCScanEndHeight(height), nil
	}
}

func stableIBCScanEndHeight(reportedHeight int64) int64 {
	if reportedHeight <= 1 {
		return reportedHeight
	}
	return reportedHeight - 1
}

// successfulIBCMessageValues correlates transaction bodies with DeliverTx
// results. Looking at block messages alone can count a message from a failed
// transaction as a receive or acknowledgement; only Code==0 transactions are
// evidence of committed IBC state transitions.
func successfulIBCMessageValues(
	ctx context.Context,
	chain *cosmos.CosmosChain,
	height int64,
	wantTypeURL string,
) ([][]byte, error) {
	block, err := chain.GetFullNode().Client.Block(ctx, &height)
	if err != nil {
		return nil, fmt.Errorf("query block %d: %w", height, err)
	}
	results, err := chain.GetFullNode().Client.BlockResults(ctx, &height)
	if err != nil {
		return nil, fmt.Errorf("query block results %d: %w", height, err)
	}
	if len(block.Block.Txs) != len(results.TxsResults) {
		return nil, fmt.Errorf(
			"block %d has %d transactions but %d transaction results",
			height,
			len(block.Block.Txs),
			len(results.TxsResults),
		)
	}

	var values [][]byte
	for i, txBytes := range block.Block.Txs {
		result := results.TxsResults[i]
		if result == nil {
			return nil, fmt.Errorf("block %d transaction %d has no result", height, i)
		}
		if result.Code != 0 {
			continue
		}
		var raw txtypes.TxRaw
		if err := proto.Unmarshal(txBytes, &raw); err != nil {
			return nil, fmt.Errorf("decode successful transaction %d at height %d: %w", i, height, err)
		}
		var body txtypes.TxBody
		if err := proto.Unmarshal(raw.BodyBytes, &body); err != nil {
			return nil, fmt.Errorf("decode successful transaction body %d at height %d: %w", i, height, err)
		}
		for _, message := range body.Messages {
			if message == nil || !ibcTypeURLEqual(message.TypeUrl, wantTypeURL) {
				continue
			}
			values = append(values, append([]byte(nil), message.Value...))
		}
	}
	return values, nil
}

func ibcTypeURLEqual(left, right string) bool {
	return strings.TrimPrefix(left, "/") == strings.TrimPrefix(right, "/")
}

func toInterchaintestPacket(packet channeltypes.Packet) ibc.Packet {
	return ibc.Packet{
		Sequence:         packet.Sequence,
		SourcePort:       packet.SourcePort,
		SourceChannel:    packet.SourceChannel,
		DestPort:         packet.DestinationPort,
		DestChannel:      packet.DestinationChannel,
		Data:             append([]byte(nil), packet.Data...),
		TimeoutHeight:    packet.TimeoutHeight.String(),
		TimeoutTimestamp: ibc.Nanoseconds(packet.TimeoutTimestamp),
	}
}

func recvPacketMatches(got channeltypes.Packet, want ibc.Packet) bool {
	packet := toInterchaintestPacket(got)
	return packet.Sequence == want.Sequence &&
		packet.SourcePort == want.SourcePort &&
		packet.SourceChannel == want.SourceChannel &&
		packet.DestPort == want.DestPort &&
		packet.DestChannel == want.DestChannel &&
		bytes.Equal(packet.Data, want.Data)
}

func ibcVoucherDenom(destination IBCChannelEndpoint, baseDenom string) string {
	trace := strings.Join([]string{destination.PortID, destination.ChannelID, baseDenom}, "/")
	hash := sha256.Sum256([]byte(trace))
	return "ibc/" + strings.ToUpper(hex.EncodeToString(hash[:]))
}

func (n *IBCTopology) recordIBCPacketStage(phase, stage string, transfer IBCPacketLifecycleEvidence) error {
	return n.artifacts.base.appendJSONLine("ibc/packets/"+phase+".jsonl", ibcPacketStageEvidence{
		Stage:      stage,
		RecordedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Transfer:   transfer,
	})
}

func (n *IBCTopology) validateTransferRuntime() error {
	if n.Panacea == nil || n.Osmosis == nil || n.Relayer == nil || n.hermes == nil || n.execReporter == nil {
		return errors.New("IBC topology runtime is not initialized")
	}
	if n.artifacts == nil || n.artifacts.base == nil {
		return errors.New("IBC topology artifact store is not initialized")
	}
	if strings.TrimSpace(n.Path) == "" {
		return errors.New("IBC topology path is required")
	}
	return nil
}

func (n *IBCTopology) ibcOperationError(stage string, err error) error {
	wrapped := fmt.Errorf("IBC %s: %w", stage, err)
	if n != nil && n.artifacts != nil && n.artifacts.base != nil {
		n.artifacts.base.recordFailure("ibc-"+stage, wrapped)
	}
	return wrapped
}
