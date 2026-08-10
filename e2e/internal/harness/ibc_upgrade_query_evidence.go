package harness

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ibcUpgradeGRPCClientStatusStep    = "ibc-post-restart-client-status"
	ibcUpgradeGRPCConnectionStep      = "ibc-post-restart-connection"
	ibcUpgradeGRPCChannelStep         = "ibc-post-restart-channel"
	ibcUpgradeGRPCReceiptStep         = "ibc-post-restart-packet-receipt"
	ibcUpgradeGRPCAcknowledgementStep = "ibc-post-restart-packet-acknowledgement"
	ibcUpgradeGRPCQueryArtifactPath   = "queries/results.jsonl"
)

// IBCUpgradeGRPCQueryCoverage describes the structured latest-state gRPC
// evidence emitted after the local Osmosis nodes have restarted. The node
// observation height is retained as metadata, but it is not promoted to a
// historical-height claim because the request was not height-pinned.
func IBCUpgradeGRPCQueryCoverage() UpgradeQueryCoverage {
	steps := []string{
		ibcUpgradeGRPCClientStatusStep,
		ibcUpgradeGRPCConnectionStep,
		ibcUpgradeGRPCChannelStep,
		ibcUpgradeGRPCReceiptStep,
		ibcUpgradeGRPCAcknowledgementStep,
	}
	references := make([]UpgradeQueryEvidenceReference, 0, len(steps))
	for _, step := range steps {
		references = append(references, UpgradeQueryEvidenceReference{
			ArtifactPath: ibcUpgradeGRPCQueryArtifactPath,
			Boundary:     UpgradeQueryBoundaryGRPC,
			Step:         step,
		})
	}
	return UpgradeQueryCoverage{
		Boundary:                      UpgradeQueryBoundaryGRPC,
		Supported:                     true,
		Exercised:                     true,
		Reason:                        "post-restart typed IBC gRPC queries prove the retained client, connection, channel, packet receipt, and acknowledgement state",
		EvidencePaths:                 []string{ibcUpgradeGRPCQueryArtifactPath},
		Evidence:                      references,
		HistoricalHeightSupported:     false,
		HistoricalHeightExercised:     false,
		HistoricalHeightReason:        "the lane records latest-state gRPC responses and does not claim a server-confirmed historical response height",
		HistoricalHeightEvidencePaths: []string{UpgradeQueryCoverageArtifactPath},
	}
}

func buildIBCUpgradeGRPCQueryRecords(evidence IBCUpgradeContinuityEvidence) ([]queryRecord, error) {
	if err := evidence.OriginalChannel.Validate(); err != nil {
		return nil, fmt.Errorf("validate IBC channel for gRPC query evidence: %w", err)
	}
	destination := evidence.OriginalChannel.Osmosis
	node, err := selectIBCUpgradeGRPCDestinationObservation(evidence.NodeRestartSemantics, destination.ChainID)
	if err != nil {
		return nil, err
	}
	packets := node.PacketSemantics
	if node.ObservedHeight < 1 {
		return nil, errors.New("IBC gRPC query evidence requires a positive post-restart observation height")
	}
	if !ibcStateEquals(node.ClientStatus, "ACTIVE") ||
		!ibcStateEquals(node.ConnectionState, "OPEN") ||
		!ibcStateEquals(node.ChannelState, "OPEN") {
		return nil, fmt.Errorf(
			"IBC post-restart gRPC link state is not active/open: client=%q connection=%q channel=%q",
			node.ClientStatus,
			node.ConnectionState,
			node.ChannelState,
		)
	}
	if packets.SuccessSequence == 0 || !packets.SuccessReceipt || strings.TrimSpace(packets.SuccessAcknowledgement) == "" {
		return nil, errors.New("IBC post-restart gRPC packet receipt or acknowledgement evidence is incomplete")
	}
	if _, err := base64.StdEncoding.DecodeString(packets.SuccessAcknowledgement); err != nil {
		return nil, fmt.Errorf("decode IBC post-restart acknowledgement evidence: %w", err)
	}

	recordedAt := time.Now().UTC()
	metadata := func(query string) map[string]any {
		return map[string]any{
			"query":                        query,
			"transport":                    "typed-grpc",
			"node":                         node.IdentityAfter.Name,
			"chain_id":                     node.ChainID,
			"post_restart_observed_height": node.ObservedHeight,
			"request_height":               int64(0),
		}
	}
	record := func(step string, request, response, recordMetadata any) queryRecord {
		return queryRecord{
			RecordedAt:       recordedAt,
			Boundary:         "grpc",
			Step:             step,
			Height:           0,
			HistoricalHeight: false,
			Request:          request,
			Response:         response,
			Metadata:         recordMetadata,
		}
	}

	return []queryRecord{
		record(
			ibcUpgradeGRPCClientStatusStep,
			map[string]any{"client_id": destination.ClientID},
			map[string]any{"status": node.ClientStatus},
			metadata("ibc.core.client.v1.Query/ClientStatus"),
		),
		record(
			ibcUpgradeGRPCConnectionStep,
			map[string]any{"connection_id": destination.ConnectionID},
			map[string]any{
				"state":                      node.ConnectionState,
				"client_id":                  destination.ClientID,
				"counterparty_connection_id": destination.CounterpartyConnectionID,
			},
			metadata("ibc.core.connection.v1.Query/Connection"),
		),
		record(
			ibcUpgradeGRPCChannelStep,
			map[string]any{"port_id": destination.PortID, "channel_id": destination.ChannelID},
			map[string]any{
				"state":                   node.ChannelState,
				"connection_id":           destination.ConnectionID,
				"counterparty_port_id":    destination.PortID,
				"counterparty_channel_id": destination.CounterpartyChannelID,
				"version":                 destination.Version,
			},
			metadata("ibc.core.channel.v1.Query/Channel"),
		),
		record(
			ibcUpgradeGRPCReceiptStep,
			map[string]any{
				"port_id": destination.PortID, "channel_id": destination.ChannelID,
				"sequence": packets.SuccessSequence,
			},
			map[string]any{"received": packets.SuccessReceipt},
			metadata("ibc.core.channel.v1.Query/PacketReceipt"),
		),
		record(
			ibcUpgradeGRPCAcknowledgementStep,
			map[string]any{
				"port_id": destination.PortID, "channel_id": destination.ChannelID,
				"sequence": packets.SuccessSequence,
			},
			map[string]any{"acknowledgement_base64": packets.SuccessAcknowledgement},
			metadata("ibc.core.channel.v1.Query/PacketAcknowledgement"),
		),
	}, nil
}

func selectIBCUpgradeGRPCDestinationObservation(
	groups []IBCChainRestartEvidence,
	destinationChainID string,
) (IBCNodePostRestartEvidence, error) {
	destinationChainID = strings.TrimSpace(destinationChainID)
	if destinationChainID == "" {
		return IBCNodePostRestartEvidence{}, errors.New("IBC gRPC destination chain ID is required")
	}
	for _, group := range groups {
		if group.ChainID != destinationChainID {
			continue
		}
		for _, node := range group.Nodes {
			if node.ChainID == destinationChainID && node.PacketSemantics.Role == "destination" {
				if strings.TrimSpace(node.IdentityAfter.Name) == "" {
					return IBCNodePostRestartEvidence{}, errors.New("IBC gRPC destination observation has no node identity")
				}
				return node, nil
			}
		}
	}
	return IBCNodePostRestartEvidence{}, fmt.Errorf(
		"IBC post-restart destination gRPC observation for chain %q was not found",
		destinationChainID,
	)
}

func (n *IBCTopology) recordIBCUpgradeGRPCQueryRecords(evidence IBCUpgradeContinuityEvidence) error {
	if n == nil || n.artifacts == nil || n.artifacts.base == nil {
		return errors.New("IBC topology artifacts are required to record gRPC query evidence")
	}
	records, err := buildIBCUpgradeGRPCQueryRecords(evidence)
	if err != nil {
		return err
	}
	for _, record := range records {
		if err := n.artifacts.base.appendJSONLine(ibcUpgradeGRPCQueryArtifactPath, record); err != nil {
			return fmt.Errorf("record IBC gRPC query %s: %w", record.Step, err)
		}
	}
	return nil
}
