package harness

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// DirectedP2PBaselineEvidence records the deterministic one-validator,
// one-full-node topology used by the network-fault suite. The validator only
// accepts connections and the full node is the sole persistent dialer. This
// avoids Interchaintest's all-to-all peer list (which includes each node
// itself) and makes repeated full-node restarts test one P2P path at a time.
type DirectedP2PBaselineEvidence struct {
	ValidatorNode          string          `json:"validator_node"`
	ValidatorNodeID        string          `json:"validator_node_id"`
	FullNode               string          `json:"full_node"`
	FullNodeNodeID         string          `json:"full_node_node_id"`
	ValidatorIPAddress     string          `json:"validator_ip_address"`
	FullNodePersistentPeer string          `json:"full_node_persistent_peer"`
	ValidatorOriginalHash  string          `json:"validator_original_sha256"`
	ValidatorDirectedHash  string          `json:"validator_directed_sha256"`
	FullNodeOriginalHash   string          `json:"full_node_original_sha256"`
	FullNodeDirectedHash   string          `json:"full_node_directed_sha256"`
	PeerExchangeEnabled    bool            `json:"peer_exchange_enabled"`
	MaxNonPersistentDials  int64           `json:"max_non_persistent_dials"`
	Agreement              QuorumAgreement `json:"agreement"`
	AppliedAt              time.Time       `json:"applied_at"`
}

func (e DirectedP2PBaselineEvidence) Validate() error {
	var validationErrors []error
	for name, value := range map[string]string{
		"validator node":            e.ValidatorNode,
		"validator node ID":         e.ValidatorNodeID,
		"full node":                 e.FullNode,
		"full-node node ID":         e.FullNodeNodeID,
		"validator IP address":      e.ValidatorIPAddress,
		"full-node persistent peer": e.FullNodePersistentPeer,
		"validator original hash":   e.ValidatorOriginalHash,
		"validator directed hash":   e.ValidatorDirectedHash,
		"full-node original hash":   e.FullNodeOriginalHash,
		"full-node directed hash":   e.FullNodeDirectedHash,
	} {
		if strings.TrimSpace(value) == "" {
			validationErrors = append(validationErrors, fmt.Errorf("%s is required", name))
		}
	}
	if e.PeerExchangeEnabled {
		validationErrors = append(validationErrors, errors.New("directed P2P baseline must disable peer exchange"))
	}
	if e.MaxNonPersistentDials != 0 {
		validationErrors = append(validationErrors, fmt.Errorf(
			"directed P2P baseline must disable non-persistent dials, got %d",
			e.MaxNonPersistentDials,
		))
	}
	if strings.Contains(e.FullNodePersistentPeer, e.FullNodeNodeID+"@") {
		validationErrors = append(validationErrors, errors.New("full-node persistent peers contain its own node ID"))
	}
	if e.AppliedAt.IsZero() {
		validationErrors = append(validationErrors, errors.New("directed P2P applied_at is required"))
	}
	return errors.Join(validationErrors...)
}

// EstablishDirectedNetworkFaultP2P replaces Interchaintest's symmetric,
// self-inclusive persistent peer list with one explicit full-node-to-validator
// connection. It restarts both nodes, then requires peer-aware agreement before
// any fault is injected.
func (n *Network) EstablishDirectedNetworkFaultP2P(
	ctx context.Context,
) (DirectedP2PBaselineEvidence, error) {
	if n == nil || n.Chain == nil || len(n.Chain.Validators) != 1 || len(n.Chain.FullNodes) != 1 {
		return DirectedP2PBaselineEvidence{}, errors.New("directed network-fault P2P requires exactly one validator and one full node")
	}
	n.txMu.Lock()
	defer n.txMu.Unlock()

	validator := n.Chain.Validators[0]
	fullNode := n.Chain.FullNodes[0]
	validatorNodeID, err := validator.NodeID(ctx)
	if err != nil {
		return DirectedP2PBaselineEvidence{}, fmt.Errorf("query directed P2P validator node ID: %w", err)
	}
	fullNodeNodeID, err := fullNode.NodeID(ctx)
	if err != nil {
		return DirectedP2PBaselineEvidence{}, fmt.Errorf("query directed P2P full-node node ID: %w", err)
	}
	validatorOriginal, err := validator.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return DirectedP2PBaselineEvidence{}, fmt.Errorf("read directed P2P validator config: %w", err)
	}
	fullNodeOriginal, err := fullNode.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return DirectedP2PBaselineEvidence{}, fmt.Errorf("read directed P2P full-node config: %w", err)
	}
	validatorIPAddress, err := n.RunNetworkIPAddress(ctx, validator.ContainerID())
	if err != nil {
		return DirectedP2PBaselineEvidence{}, fmt.Errorf("query directed P2P validator address before restart: %w", err)
	}
	validatorMutation, _, err := directedNetworkFaultP2PMutations(validatorNodeID, validatorIPAddress)
	if err != nil {
		return DirectedP2PBaselineEvidence{}, err
	}
	validatorDirected, err := rewriteCometP2PRoute(validatorOriginal, validatorMutation)
	if err != nil {
		return DirectedP2PBaselineEvidence{}, fmt.Errorf("rewrite directed P2P validator config: %w", err)
	}
	rollback := func(operationErr error) error {
		rollbackErr := restartNodesWithConfigDocuments(
			ctx,
			validator,
			fullNode,
			validatorOriginal,
			fullNodeOriginal,
		)
		return errors.Join(operationErr, rollbackErr)
	}
	if err := fullNode.StopContainer(ctx); err != nil {
		return DirectedP2PBaselineEvidence{}, fmt.Errorf("stop full node for directed P2P baseline: %w", err)
	}
	if err := validator.StopContainer(ctx); err != nil {
		return DirectedP2PBaselineEvidence{}, rollback(fmt.Errorf("stop validator for directed P2P baseline: %w", err))
	}
	if err := validator.WriteFile(ctx, validatorDirected, "config/config.toml"); err != nil {
		return DirectedP2PBaselineEvidence{}, rollback(fmt.Errorf("write directed P2P validator config: %w", err))
	}
	if err := validator.StartContainer(ctx); err != nil {
		return DirectedP2PBaselineEvidence{}, rollback(fmt.Errorf("start directed P2P validator: %w", err))
	}
	validatorIPAddress, err = n.RunNetworkIPAddress(ctx, validator.ContainerID())
	if err != nil {
		return DirectedP2PBaselineEvidence{}, rollback(fmt.Errorf("query directed P2P validator address after restart: %w", err))
	}
	_, fullNodeMutation, err := directedNetworkFaultP2PMutations(validatorNodeID, validatorIPAddress)
	if err != nil {
		return DirectedP2PBaselineEvidence{}, rollback(err)
	}
	fullNodeDirected, err := rewriteCometP2PRoute(fullNodeOriginal, fullNodeMutation)
	if err != nil {
		return DirectedP2PBaselineEvidence{}, rollback(fmt.Errorf("rewrite directed P2P full-node config: %w", err))
	}
	if err := fullNode.WriteFile(ctx, fullNodeDirected, "config/config.toml"); err != nil {
		return DirectedP2PBaselineEvidence{}, rollback(fmt.Errorf("write directed P2P full-node config: %w", err))
	}
	if err := fullNode.StartContainer(ctx); err != nil {
		return DirectedP2PBaselineEvidence{}, rollback(fmt.Errorf("start directed P2P full node: %w", err))
	}

	evidence := DirectedP2PBaselineEvidence{
		ValidatorNode:          validator.Name(),
		ValidatorNodeID:        validatorNodeID,
		FullNode:               fullNode.Name(),
		FullNodeNodeID:         fullNodeNodeID,
		ValidatorIPAddress:     validatorIPAddress,
		FullNodePersistentPeer: fullNodeMutation.persistentPeers,
		ValidatorOriginalHash:  networkFaultSHA256(validatorOriginal),
		ValidatorDirectedHash:  networkFaultSHA256(validatorDirected),
		FullNodeOriginalHash:   networkFaultSHA256(fullNodeOriginal),
		FullNodeDirectedHash:   networkFaultSHA256(fullNodeDirected),
		PeerExchangeEnabled:    false,
		MaxNonPersistentDials:  0,
		AppliedAt:              time.Now().UTC(),
	}
	if err := evidence.Validate(); err != nil {
		return evidence, rollback(err)
	}
	validatorApplied, err := validator.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return evidence, fmt.Errorf("read applied directed P2P validator config: %w", err)
	}
	fullNodeApplied, err := fullNode.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return evidence, fmt.Errorf("read applied directed P2P full-node config: %w", err)
	}
	if actual := networkFaultSHA256(validatorApplied); actual != evidence.ValidatorDirectedHash {
		return evidence, fmt.Errorf("applied validator P2P hash %s does not match directed hash %s", actual, evidence.ValidatorDirectedHash)
	}
	if actual := networkFaultSHA256(fullNodeApplied); actual != evidence.FullNodeDirectedHash {
		return evidence, fmt.Errorf("applied full-node P2P hash %s does not match directed hash %s", actual, evidence.FullNodeDirectedHash)
	}
	targetHeight, err := validator.Height(ctx)
	if err != nil {
		return evidence, fmt.Errorf("query directed P2P validator height: %w", err)
	}
	evidence.Agreement, err = n.WaitForQuorumAgreement(
		ctx,
		"network-fault-directed-p2p-baseline",
		targetHeight,
		validator,
		fullNode,
	)
	if err != nil {
		return evidence, fmt.Errorf("verify directed P2P peer-aware agreement: %w", err)
	}
	if err := n.artifacts.write("network-faults/config/validator-directed-config.toml", validatorDirected); err != nil {
		return evidence, err
	}
	if err := n.artifacts.write("network-faults/config/full-node-directed-config.toml", fullNodeDirected); err != nil {
		return evidence, err
	}
	if err := n.artifacts.writeJSON("network-faults/p2p-directed-baseline.json", evidence); err != nil {
		return evidence, err
	}
	return evidence, nil
}

func directedNetworkFaultP2PMutations(
	validatorNodeID string,
	validatorIPAddress string,
) (cometP2PRouteMutation, cometP2PRouteMutation, error) {
	validatorNodeID = strings.TrimSpace(validatorNodeID)
	if validatorNodeID == "" {
		return cometP2PRouteMutation{}, cometP2PRouteMutation{}, errors.New("validator node ID is required")
	}
	validatorIPAddress = strings.TrimSpace(validatorIPAddress)
	parsedValidatorIP := net.ParseIP(validatorIPAddress)
	if parsedValidatorIP == nil || parsedValidatorIP.To4() == nil {
		return cometP2PRouteMutation{}, cometP2PRouteMutation{}, fmt.Errorf("validator IPv4 address %q is invalid", validatorIPAddress)
	}
	if strings.ContainsAny(validatorNodeID, "@, ") {
		return cometP2PRouteMutation{}, cometP2PRouteMutation{}, fmt.Errorf("validator node ID %q contains a peer-list delimiter", validatorNodeID)
	}
	noNonPersistentOutbound := int64(0)
	return cometP2PRouteMutation{
			persistentPeers:  "",
			peerExchange:     false,
			maxOutboundPeers: &noNonPersistentOutbound,
		}, cometP2PRouteMutation{
			persistentPeers:  validatorNodeID + "@" + parsedValidatorIP.String() + ":26656",
			peerExchange:     false,
			maxOutboundPeers: &noNonPersistentOutbound,
		}, nil
}

type P2PProxyRouteEvidence struct {
	ProxyAlias            string    `json:"proxy_alias"`
	ProxyTargetAddress    string    `json:"proxy_target_address"`
	TargetNode            string    `json:"target_node"`
	TargetNodeID          string    `json:"target_node_id"`
	FullNode              string    `json:"full_node"`
	AppliedAt             time.Time `json:"applied_at"`
	ValidatorOriginalHash string    `json:"validator_original_sha256"`
	ValidatorModifiedHash string    `json:"validator_modified_sha256"`
	FullNodeOriginalHash  string    `json:"full_node_original_sha256"`
	FullNodeModifiedHash  string    `json:"full_node_modified_sha256"`

	validatorOriginal []byte
	fullNodeOriginal  []byte
}

func (e P2PProxyRouteEvidence) Validate() error {
	var validationErrors []error
	for name, value := range map[string]string{
		"proxy alias":             e.ProxyAlias,
		"proxy target address":    e.ProxyTargetAddress,
		"target node":             e.TargetNode,
		"target node ID":          e.TargetNodeID,
		"full node":               e.FullNode,
		"validator original hash": e.ValidatorOriginalHash,
		"validator modified hash": e.ValidatorModifiedHash,
		"full-node original hash": e.FullNodeOriginalHash,
		"full-node modified hash": e.FullNodeModifiedHash,
	} {
		if strings.TrimSpace(value) == "" {
			validationErrors = append(validationErrors, fmt.Errorf("%s is required", name))
		}
	}
	if e.AppliedAt.IsZero() {
		validationErrors = append(validationErrors, errors.New("route applied_at is required"))
	}
	if len(e.validatorOriginal) == 0 || len(e.fullNodeOriginal) == 0 {
		validationErrors = append(validationErrors, errors.New("route rollback documents are required"))
	}
	if e.ValidatorOriginalHash == e.ValidatorModifiedHash {
		validationErrors = append(validationErrors, errors.New("validator P2P config was not changed"))
	}
	if e.FullNodeOriginalHash == e.FullNodeModifiedHash {
		validationErrors = append(validationErrors, errors.New("full-node P2P config was not changed"))
	}
	return errors.Join(validationErrors...)
}

// RouteFullNodeP2PThroughProxy disables peer exchange and every direct
// persistent peer on validator-0, then configures full-node-0 to dial only the
// supplied proxy alias. The proxy must already be listening.
func (n *Network) RouteFullNodeP2PThroughProxy(
	ctx context.Context,
	proxyAlias string,
) (P2PProxyRouteEvidence, error) {
	if n == nil || n.Chain == nil || len(n.Chain.Validators) == 0 || len(n.Chain.FullNodes) == 0 {
		return P2PProxyRouteEvidence{}, errors.New("P2P proxy route requires a validator and full node")
	}
	if !networkFaultNamePattern.MatchString(proxyAlias) {
		return P2PProxyRouteEvidence{}, fmt.Errorf("P2P proxy alias %q must match %s", proxyAlias, networkFaultNamePattern)
	}
	n.txMu.Lock()
	defer n.txMu.Unlock()
	validator := n.Chain.Validators[0]
	fullNode := n.Chain.FullNodes[0]
	targetNodeID, err := validator.NodeID(ctx)
	if err != nil {
		return P2PProxyRouteEvidence{}, fmt.Errorf("query proxy target node ID: %w", err)
	}
	validatorOriginal, err := validator.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return P2PProxyRouteEvidence{}, fmt.Errorf("read validator P2P config: %w", err)
	}
	fullNodeOriginal, err := fullNode.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return P2PProxyRouteEvidence{}, fmt.Errorf("read full-node P2P config: %w", err)
	}
	validatorMutation, fullNodeMutation := cometP2PProxyRouteMutations(targetNodeID, proxyAlias)
	validatorModified, err := rewriteCometP2PRoute(validatorOriginal, validatorMutation)
	if err != nil {
		return P2PProxyRouteEvidence{}, fmt.Errorf("rewrite validator P2P config: %w", err)
	}
	fullNodeModified, err := rewriteCometP2PRoute(fullNodeOriginal, fullNodeMutation)
	if err != nil {
		return P2PProxyRouteEvidence{}, fmt.Errorf("rewrite full-node P2P config: %w", err)
	}
	evidence := P2PProxyRouteEvidence{
		ProxyAlias:            proxyAlias,
		ProxyTargetAddress:    fmt.Sprintf("%s:%d", validator.HostName(), P2PFaultProxyTargetPort),
		TargetNode:            validator.Name(),
		TargetNodeID:          targetNodeID,
		FullNode:              fullNode.Name(),
		AppliedAt:             time.Now().UTC(),
		ValidatorOriginalHash: networkFaultSHA256(validatorOriginal),
		ValidatorModifiedHash: networkFaultSHA256(validatorModified),
		FullNodeOriginalHash:  networkFaultSHA256(fullNodeOriginal),
		FullNodeModifiedHash:  networkFaultSHA256(fullNodeModified),
		validatorOriginal:     append([]byte(nil), validatorOriginal...),
		fullNodeOriginal:      append([]byte(nil), fullNodeOriginal...),
	}
	if err := evidence.Validate(); err != nil {
		return P2PProxyRouteEvidence{}, err
	}
	operationErr := restartNodesWithConfigDocuments(
		ctx,
		validator,
		fullNode,
		validatorModified,
		fullNodeModified,
	)
	if operationErr != nil {
		rollbackErr := restartNodesWithConfigDocuments(
			ctx,
			validator,
			fullNode,
			validatorOriginal,
			fullNodeOriginal,
		)
		return evidence, errors.Join(fmt.Errorf("apply P2P proxy route: %w", operationErr), rollbackErr)
	}
	if err := n.artifacts.writeJSON("network-faults/p2p-route-applied.json", evidence); err != nil {
		return evidence, fmt.Errorf("record P2P proxy route: %w", err)
	}
	if err := n.artifacts.write("network-faults/config/validator-proxied-config.toml", validatorModified); err != nil {
		return evidence, err
	}
	if err := n.artifacts.write("network-faults/config/full-node-proxied-config.toml", fullNodeModified); err != nil {
		return evidence, err
	}
	return evidence, nil
}

// RestoreFullNodeP2PRoute restores both original config documents and proves
// the exact hashes were returned before fault-proxy cleanup.
func (n *Network) RestoreFullNodeP2PRoute(ctx context.Context, evidence P2PProxyRouteEvidence) error {
	if n == nil || n.Chain == nil || len(n.Chain.Validators) == 0 || len(n.Chain.FullNodes) == 0 {
		return errors.New("P2P proxy route restore requires a validator and full node")
	}
	if err := evidence.Validate(); err != nil {
		return err
	}
	n.txMu.Lock()
	defer n.txMu.Unlock()
	validator := n.Chain.Validators[0]
	fullNode := n.Chain.FullNodes[0]
	if err := restartNodesWithConfigDocuments(
		ctx,
		validator,
		fullNode,
		evidence.validatorOriginal,
		evidence.fullNodeOriginal,
	); err != nil {
		return fmt.Errorf("restore P2P route: %w", err)
	}
	validatorRestored, err := validator.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return err
	}
	fullNodeRestored, err := fullNode.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return err
	}
	validatorMatches := networkFaultSHA256(validatorRestored) == evidence.ValidatorOriginalHash
	fullNodeMatches := networkFaultSHA256(fullNodeRestored) == evidence.FullNodeOriginalHash
	result := map[string]any{
		"restored_at":               time.Now().UTC(),
		"validator_sha256":          networkFaultSHA256(validatorRestored),
		"full_node_sha256":          networkFaultSHA256(fullNodeRestored),
		"validator_hash_matches":    validatorMatches,
		"full_node_hash_matches":    fullNodeMatches,
		"host_firewall_was_changed": false,
	}
	if !validatorMatches || !fullNodeMatches {
		return errors.New("restored P2P config hashes do not match originals")
	}
	return n.artifacts.writeJSON("network-faults/p2p-route-restored.json", result)
}

type cometP2PRouteMutation struct {
	persistentPeers  string
	peerExchange     bool
	listenAddress    string
	maxOutboundPeers *int64
}

func cometP2PProxyRouteMutations(targetNodeID, proxyAlias string) (cometP2PRouteMutation, cometP2PRouteMutation) {
	noOutboundPeers := int64(0)
	return cometP2PRouteMutation{
			peerExchange:     false,
			listenAddress:    fmt.Sprintf("tcp://0.0.0.0:%d", P2PFaultProxyTargetPort),
			maxOutboundPeers: &noOutboundPeers,
		}, cometP2PRouteMutation{
			persistentPeers:  targetNodeID + "@" + proxyAlias + ":26656",
			peerExchange:     false,
			maxOutboundPeers: &noOutboundPeers,
		}
}

func rewriteCometP2PRoute(contents []byte, mutation cometP2PRouteMutation) ([]byte, error) {
	var document map[string]any
	if _, err := toml.Decode(string(contents), &document); err != nil {
		return nil, fmt.Errorf("decode config.toml: %w", err)
	}
	p2pValue, ok := document["p2p"]
	if !ok {
		return nil, errors.New("config.toml has no p2p section")
	}
	p2p, ok := p2pValue.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("config.toml p2p section has type %T", p2pValue)
	}
	p2p["persistent_peers"] = mutation.persistentPeers
	p2p["seeds"] = ""
	p2p["pex"] = mutation.peerExchange
	p2p["unconditional_peer_ids"] = ""
	if mutation.listenAddress != "" {
		p2p["laddr"] = mutation.listenAddress
		p2p["external_address"] = ""
	}
	if mutation.maxOutboundPeers != nil {
		p2p["max_num_outbound_peers"] = *mutation.maxOutboundPeers
	}
	document["p2p"] = p2p
	var output bytes.Buffer
	if err := toml.NewEncoder(&output).Encode(document); err != nil {
		return nil, fmt.Errorf("encode config.toml: %w", err)
	}
	return output.Bytes(), nil
}

func restartNodesWithConfigDocuments(
	ctx context.Context,
	validator *cosmos.ChainNode,
	fullNode *cosmos.ChainNode,
	validatorConfig []byte,
	fullNodeConfig []byte,
) error {
	if validator == nil || fullNode == nil {
		return errors.New("validator and full node are required")
	}
	fullStopErr := fullNode.StopContainer(ctx)
	validatorStopErr := validator.StopContainer(ctx)
	if err := errors.Join(fullStopErr, validatorStopErr); err != nil {
		return err
	}
	validatorWriteErr := validator.WriteFile(ctx, validatorConfig, "config/config.toml")
	fullWriteErr := fullNode.WriteFile(ctx, fullNodeConfig, "config/config.toml")
	if err := errors.Join(validatorWriteErr, fullWriteErr); err != nil {
		validatorRestartErr := validator.StartContainer(ctx)
		fullNodeRestartErr := fullNode.StartContainer(ctx)
		return errors.Join(err, validatorRestartErr, fullNodeRestartErr)
	}
	validatorStartErr := validator.StartContainer(ctx)
	fullNodeStartErr := fullNode.StartContainer(ctx)
	return errors.Join(validatorStartErr, fullNodeStartErr)
}

func networkFaultSHA256(contents []byte) string {
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:])
}
