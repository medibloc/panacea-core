package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

// SemanticJSON is a canonical JSON query response. Object keys are sorted by
// encoding/json and numbers are retained as json.Number values, so snapshots
// compare application state rather than CLI whitespace or object-key order.
type SemanticJSON []byte

// NewSemanticJSON validates and defensively canonicalizes one JSON response.
func NewSemanticJSON(contents []byte) (SemanticJSON, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("decode semantic JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("multiple JSON values")
		}
		return nil, fmt.Errorf("decode semantic JSON trailing content: %w", err)
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode semantic JSON: %w", err)
	}
	return append(SemanticJSON(nil), canonical...), nil
}

// NFTStateSnapshotRequest selects only NFT module state. Bank balances,
// account sequence, and fee effects are deliberately outside this snapshot.
type NFTStateSnapshotRequest struct {
	ClassID string
	NFTIDs  []string
	Owners  []string
}

// NFTStateSnapshot is the semantic NFT state needed to prove that a rejected
// transaction made no partial Class, lifecycle, supply, or owner-index write.
// Every field is queried through the explicit full node.
type NFTStateSnapshot struct {
	ClassRecord      SemanticJSON            `json:"class_record"`
	NFTRecords       map[string]SemanticJSON `json:"nft_records"`
	LiveByClass      SemanticJSON            `json:"live_by_class"`
	LiveByOwner      map[string]SemanticJSON `json:"live_by_owner"`
	StandardSupply   SemanticJSON            `json:"standard_supply"`
	StandardBalances map[string]SemanticJSON `json:"standard_balances"`
}

// SnapshotNFTState queries coupled Panacea and standard NFT views without
// observing bank/auth state that legitimately changes when a failed DeliverTx
// consumes fees and sequence.
func (n *Network) SnapshotNFTState(
	ctx context.Context,
	step string,
	request NFTStateSnapshotRequest,
) (NFTStateSnapshot, error) {
	if strings.TrimSpace(step) == "" {
		return NFTStateSnapshot{}, errors.New("NFT snapshot step is required")
	}
	if strings.TrimSpace(request.ClassID) == "" {
		return NFTStateSnapshot{}, errors.New("NFT snapshot class ID is required")
	}
	nftIDs, err := sortedUniqueSnapshotValues("NFT ID", request.NFTIDs)
	if err != nil {
		return NFTStateSnapshot{}, err
	}
	owners, err := sortedUniqueSnapshotValues("owner", request.Owners)
	if err != nil {
		return NFTStateSnapshot{}, err
	}

	snapshot := NFTStateSnapshot{
		NFTRecords:       make(map[string]SemanticJSON, len(nftIDs)),
		LiveByOwner:      make(map[string]SemanticJSON, len(owners)),
		StandardBalances: make(map[string]SemanticJSON, len(owners)),
	}
	snapshot.ClassRecord, err = n.snapshotNFTQuery(
		ctx,
		step+"-class-record",
		"nft", "class-record", request.ClassID,
	)
	if err != nil {
		return NFTStateSnapshot{}, err
	}
	for index, nftID := range nftIDs {
		snapshot.NFTRecords[nftID], err = n.snapshotNFTQuery(
			ctx,
			fmt.Sprintf("%s-nft-record-%d", step, index),
			"nft", "nft-record", request.ClassID, nftID,
		)
		if err != nil {
			return NFTStateSnapshot{}, err
		}
	}
	snapshot.LiveByClass, err = n.snapshotNFTQuery(
		ctx,
		step+"-live-by-class",
		"nft", "nft-records", "--class-id", request.ClassID, "--limit", "100",
	)
	if err != nil {
		return NFTStateSnapshot{}, err
	}
	snapshot.StandardSupply, err = n.snapshotNFTQuery(
		ctx,
		step+"-standard-supply",
		"nft", "supply", request.ClassID,
	)
	if err != nil {
		return NFTStateSnapshot{}, err
	}
	for index, owner := range owners {
		snapshot.LiveByOwner[owner], err = n.snapshotNFTQuery(
			ctx,
			fmt.Sprintf("%s-live-by-owner-%d", step, index),
			"nft", "nft-records",
			"--class-id", request.ClassID,
			"--owner", owner,
			"--limit", "100",
		)
		if err != nil {
			return NFTStateSnapshot{}, err
		}
		snapshot.StandardBalances[owner], err = n.snapshotNFTQuery(
			ctx,
			fmt.Sprintf("%s-standard-balance-%d", step, index),
			"nft", "balance", owner, request.ClassID,
		)
		if err != nil {
			return NFTStateSnapshot{}, err
		}
	}
	return snapshot, nil
}

func (n *Network) snapshotNFTQuery(
	ctx context.Context,
	step string,
	command ...string,
) (SemanticJSON, error) {
	response, err := n.FullNodeCLIQuery(ctx, step, command...)
	if err != nil {
		return nil, fmt.Errorf("snapshot NFT state %s: %w", step, err)
	}
	semantic, err := NewSemanticJSON(response)
	if err != nil {
		return nil, fmt.Errorf("snapshot NFT state %s: %w", step, err)
	}
	return semantic, nil
}

func sortedUniqueSnapshotValues(name string, values []string) ([]string, error) {
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	for index, value := range copyValues {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("NFT snapshot %s must not be empty", name)
		}
		if index > 0 && value == copyValues[index-1] {
			return nil, fmt.Errorf("NFT snapshot %s %q is duplicated", name, value)
		}
	}
	return copyValues, nil
}
