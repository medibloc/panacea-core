package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	networkFaultSeed                        int64 = 20260804
	networkFaultWebSocketOperationTimeout         = 12 * time.Second
	networkFaultWebSocketCleanupTimeout           = 2 * time.Second
	networkFaultOversizedBodyMarker               = "request body too large"
	networkFaultGRPCMaxReceiveBytes         uint  = 1024
	networkFaultFirstReadinessStartTimeout        = 20 * time.Second
	networkFaultReadinessObservationTimeout       = 3 * time.Second
)

type networkFaultEndpointMutation struct {
	APIEnabled       *bool
	GRPCEnabled      *bool
	ReadTimeout      *uint
	WriteTimeout     *uint
	MaxBodyBytes     *uint
	GRPCMaxRecvBytes *uint
}

type networkFaultWebSocketTransaction struct {
	Height int64  `json:"height"`
	TxHash string `json:"tx_hash"`
}

type networkFaultWebSocketPhase struct {
	BlockHeights []int64                            `json:"block_heights"`
	Transactions []networkFaultWebSocketTransaction `json:"transactions"`
}

type networkFaultWebSocketContinuity struct {
	Before                       networkFaultWebSocketPhase `json:"before_fault"`
	After                        networkFaultWebSocketPhase `json:"after_fault"`
	FaultMissingBlockHeights     []int64                    `json:"fault_missing_block_heights"`
	DuplicateBlockEvents         int                        `json:"duplicate_block_events"`
	DuplicateTransactionEvents   int                        `json:"duplicate_transaction_events"`
	MissingTransactionHashes     []string                   `json:"missing_transaction_hashes"`
	UnexpectedTransactionHashes  []string                   `json:"unexpected_transaction_hashes,omitempty"`
	ConsecutiveOutsideFault      bool                       `json:"consecutive_outside_fault"`
	ExpectedTransactionsObserved bool                       `json:"expected_transactions_observed"`
}

type networkFaultProxyLogEvent struct {
	Event        string `json:"event"`
	ConnectionID uint64 `json:"connection_id,omitempty"`
	Client       string `json:"client,omitempty"`
	Target       string `json:"target,omitempty"`
	Direction    string `json:"direction,omitempty"`
	Chunk        uint64 `json:"chunk,omitempty"`
	Bytes        int    `json:"bytes,omitempty"`
}

type networkFaultReadinessRecovery struct {
	Phase                    string    `json:"phase"`
	FirstStartTimeout        string    `json:"first_start_timeout"`
	FirstStartError          string    `json:"first_start_error"`
	FirstStartHeight         int64     `json:"first_start_height"`
	FirstStartCatchingUp     bool      `json:"first_start_catching_up"`
	RecoveryRestartAttempted bool      `json:"recovery_restart_attempted"`
	RecoveryRestartSucceeded bool      `json:"recovery_restart_succeeded"`
	RecoveryRestartError     string    `json:"recovery_restart_error,omitempty"`
	RecordedAt               time.Time `json:"recorded_at"`
}

type networkFaultReadinessRuntime struct {
	Start   func(context.Context) error
	Stop    func(context.Context) error
	Observe func(context.Context) (height int64, catchingUp bool, err error)
	Record  func(networkFaultReadinessRecovery) error
}

func TestLocalDockerNetworkAndEndpointFaults(t *testing.T) {
	if os.Getenv("PANACEA_E2E_NETWORK_FAULTS") != "1" {
		t.Skip("set PANACEA_E2E_NETWORK_FAULTS=1 or use ./scripts/e2e/run.sh network-faults")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 18*time.Minute)
	defer cancel()
	network, err := harness.Start(ctx, t, harness.Config{
		Image:                harness.CurrentImage(),
		NumValidators:        1,
		NumFullNodes:         1,
		TimeoutCommit:        "1s",
		SetupFailureCategory: harness.NetworkFaultCategoryEnvironmentPreflight,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.NoError(t, network.WriteArtifactJSON("environment/network-policy.json", map[string]any{
		"scope":                       "run-owned Docker bridge and run-owned proxy containers",
		"seed":                        networkFaultSeed,
		"host_firewall_changed":       false,
		"host_network_namespace_used": false,
		"net_admin_required":          false,
		"faults": []string{
			"Docker bridge partition",
			"TCP proxy delay/jitter/effective stream loss",
			"proxy DNS alias recreation",
			"REST and gRPC listener isolation",
			"slow client, oversized REST and gRPC messages, connection churn",
			"same-client WebSocket NewBlock/Tx reconnect, duplicate and gap detection",
		},
		"failure_category_artifacts": map[string]string{
			"environment_preflight": "environment/network-failure-categories.jsonl",
			"chain_p2p_runtime":     "network-faults/failure-categories.jsonl",
		},
		"excluded_p2": []string{"host firewall", "external ingress", "external DNS", "TLS"},
	}))

	validator := network.Chain.Validators[0]
	fullNode := network.Chain.FullNodes[0]
	preflightErr := network.WaitForHeight(ctx, 5)
	if preflightErr == nil {
		preflightErr = network.WaitForFullNode(ctx, 5)
	}
	preflightEvidence := harness.NetworkFaultCategoryEvidence{
		Category: harness.NetworkFaultCategoryEnvironmentPreflight,
		Phase:    "initial-chain-readiness",
		Outcome:  harness.NetworkFaultOutcomePassed,
		Scope:    harness.NetworkFaultScopeLocalEnvironment,
	}
	if preflightErr != nil {
		preflightEvidence.Outcome = harness.NetworkFaultOutcomeFailed
		preflightEvidence.Error = preflightErr.Error()
	}
	categoryErr := network.RecordNetworkFaultCategory(preflightEvidence)
	require.NoError(t, errors.Join(preflightErr, categoryErr))

	networkFaultExerciseDockerPartition(t, ctx, network, validator, fullNode)
	networkFaultExerciseProxyImpairments(t, ctx, network, validator, fullNode)
	// Container recreation still proves that Docker replaces the full-node IP
	// and updates its DNS alias. Apply the one-way IP route first so that the
	// subsequent P2P recovery does not depend on Interchaintest's symmetric,
	// self-inclusive hostname peer list while that alias is changing.
	_, err = network.EstablishDirectedNetworkFaultP2P(ctx)
	require.NoError(t, err)
	networkFaultExerciseContainerRecreation(t, ctx, network, validator, fullNode)
	networkFaultExerciseEndpointIsolation(t, ctx, network, validator, fullNode)
	networkFaultExerciseGRPCMessageBoundary(t, ctx, network, validator, fullNode)
	networkFaultExerciseHTTPBoundaries(t, ctx, network, validator, fullNode)
	networkFaultExerciseWebSocketReconnect(t, ctx, network, validator, fullNode)

	finalHeight, err := validator.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForFullNode(ctx, finalHeight))
	_, err = network.RequireSameHistoryAtHeight(ctx, finalHeight, validator, fullNode)
	require.NoError(t, err)
	require.NoError(t, network.WriteArtifactJSON("network-faults/result.json", map[string]any{
		"passed":                true,
		"final_height":          finalHeight,
		"same_history":          true,
		"host_firewall_changed": false,
		"recorded_at":           time.Now().UTC(),
	}))
}

func networkFaultExerciseDockerPartition(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	validator *cosmos.ChainNode,
	fullNode *cosmos.ChainNode,
) {
	t.Helper()
	startHeight, err := validator.Height(ctx)
	require.NoError(t, err)
	// Docker/Colima may remove the published-port entry from inspect while a
	// container is detached from its only bridge. Capture the run-owned host
	// endpoint before the fault so the probe tests reachability, not inspect
	// representation during the fault.
	rpcAddress, err := network.FullNodeHostAddress(ctx, "26657/tcp")
	require.NoError(t, err)
	disconnected := true
	require.NoError(t, network.DisconnectNodeFromRunNetwork(ctx, "full-node-partition", fullNode))
	t.Cleanup(func() {
		if !disconnected {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		recordNetworkFaultCleanup(
			t,
			network,
			"full-node-partition",
			network.ReconnectNodeToRunNetwork(cleanupCtx, "cleanup-full-node-partition", fullNode),
		)
	})
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, startHeight+3))
	probeCtx, probeCancel := context.WithTimeout(ctx, time.Second)
	probeErr := harness.RequireRPCStatus(probeCtx, rpcAddress, startHeight)
	probeCancel()
	require.Error(t, probeErr, "partitioned full-node RPC must not be reachable through a stale bridge address")
	require.NoError(t, recordChainP2PRuntimeFault(network, "full-node-partition", harness.NetworkFaultOutcomeExpectedFaultObserved))
	require.NoError(t, network.RecordNetworkFaultCleanup(
		"full-node-reconnect",
		network.ReconnectNodeToRunNetwork(ctx, "full-node-reconnect", fullNode),
	))
	disconnected = false
	require.NoError(t, network.WaitForFullNode(ctx, startHeight+3))
	_, err = network.RequireSameHistoryAtHeight(ctx, startHeight+3, validator, fullNode)
	require.NoError(t, err)
	require.NoError(t, recordChainP2PRuntimeFault(network, "full-node-reconnect", harness.NetworkFaultOutcomeRecovered))
	require.NoError(t, network.WriteArtifactJSON("network-faults/partition-recovery.json", map[string]any{
		"start_height":               startHeight,
		"target_height":              startHeight + 3,
		"rpc_during_partition_error": probeErr.Error(),
		"recovered":                  true,
	}))
}

func networkFaultExerciseProxyImpairments(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	validator *cosmos.ChainNode,
	fullNode *cosmos.ChainNode,
) {
	t.Helper()
	const proxyAlias = "p2p-fault-proxy"
	baseConfig := harness.P2PFaultProxyConfig{
		Alias:         proxyAlias,
		Image:         harness.CurrentImage(),
		TargetAddress: fmt.Sprintf("%s:%d", validator.HostName(), harness.P2PFaultProxyTargetPort),
		Seed:          networkFaultSeed,
	}
	activeProxies := make(map[string]harness.P2PFaultProxy)
	startProxy := func(config harness.P2PFaultProxyConfig) (harness.P2PFaultProxy, error) {
		proxy, startErr := network.StartP2PFaultProxy(ctx, config)
		if strings.TrimSpace(proxy.ContainerID) != "" {
			activeProxies[proxy.ContainerID] = proxy
		}
		return proxy, startErr
	}
	stopProxy := func(recordPhase, logPhase string, proxy harness.P2PFaultProxy) error {
		stopErr := network.StopP2PFaultProxy(ctx, logPhase, proxy)
		if stopErr == nil {
			delete(activeProxies, proxy.ContainerID)
		}
		return network.RecordNetworkFaultCleanup(recordPhase, stopErr)
	}
	var route *harness.P2PProxyRouteEvidence
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cleanupCancel()
		var cleanupErrors []error
		if route != nil {
			cleanupErrors = append(cleanupErrors, network.RestoreFullNodeP2PRoute(cleanupCtx, *route))
		}
		proxies := make([]harness.P2PFaultProxy, 0, len(activeProxies))
		for _, proxy := range activeProxies {
			proxies = append(proxies, proxy)
		}
		sort.Slice(proxies, func(i, j int) bool {
			return proxies[i].Config.Name < proxies[j].Config.Name
		})
		for _, proxy := range proxies {
			cleanupErrors = append(
				cleanupErrors,
				network.StopP2PFaultProxy(cleanupCtx, "cleanup-"+proxy.Config.Name, proxy),
			)
		}
		recordNetworkFaultCleanup(t, network, "p2p-proxy-route", errors.Join(cleanupErrors...))
	})

	cleanConfig := baseConfig
	cleanConfig.Name = "route-clean"
	cleanProxy, err := startProxy(cleanConfig)
	require.NoError(t, err)
	applied, err := network.RouteFullNodeP2PThroughProxy(ctx, proxyAlias)
	if applied.Validate() == nil {
		route = &applied
	}
	require.NoError(t, err)
	target, err := validator.Height(ctx)
	require.NoError(t, err)
	target += 3
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, target))
	require.NoError(t, network.WaitForFullNode(ctx, target))

	require.NoError(t, stopProxy("route-clean-proxy-stop", "route-clean", cleanProxy))
	partitionStart, err := validator.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, partitionStart+2))
	stallCtx, stallCancel := context.WithTimeout(ctx, 8*time.Second)
	partitionedFullNodeHeight, stallErr := network.WaitForStableHeight(stallCtx, fullNode, 2*time.Second)
	stallCancel()
	require.NoError(t, stallErr, "full node must not bypass the stopped P2P proxy")
	validatorAfterPartition, err := validator.Height(ctx)
	require.NoError(t, err)
	require.Less(t, partitionedFullNodeHeight, validatorAfterPartition)
	require.NoError(t, recordChainP2PRuntimeFault(network, "proxy-partition", harness.NetworkFaultOutcomeExpectedFaultObserved))
	require.NoError(t, network.WriteArtifactJSON("network-faults/proxy-partition.json", map[string]any{
		"validator_start_height":   partitionStart,
		"validator_current_height": validatorAfterPartition,
		"full_node_stalled_height": partitionedFullNodeHeight,
		"proxy_stopped":            true,
		"direct_bypass_observed":   false,
	}))

	lossConfig := baseConfig
	lossConfig.Name = "effective-loss"
	lossConfig.Delay = 40 * time.Millisecond
	lossConfig.Jitter = 20 * time.Millisecond
	lossConfig.DropEvery = 1
	lossProxy, err := startProxy(lossConfig)
	require.NoError(t, err)
	require.NoError(t, waitForFaultProxyLog(ctx, network, lossProxy, "\"event\":\"chunk-dropped\"", 15*time.Second))
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, partitionStart+4))
	lossStallCtx, lossStallCancel := context.WithTimeout(ctx, 8*time.Second)
	lossStalledHeight, lossStallErr := network.WaitForStableHeight(lossStallCtx, fullNode, 2*time.Second)
	lossStallCancel()
	require.NoError(t, lossStallErr, "effective stream loss must prevent full-node block progress")
	validatorAfterLoss, err := validator.Height(ctx)
	require.NoError(t, err)
	require.Less(t, lossStalledHeight, validatorAfterLoss)
	require.NoError(t, recordChainP2PRuntimeFault(network, "effective-loss", harness.NetworkFaultOutcomeExpectedFaultObserved))
	require.NoError(t, network.WriteArtifactJSON("network-faults/effective-loss.json", map[string]any{
		"delay":                    lossConfig.Delay.String(),
		"jitter":                   lossConfig.Jitter.String(),
		"drop_every":               lossConfig.DropEvery,
		"chunk_drop_observed":      true,
		"validator_current_height": validatorAfterLoss,
		"full_node_stalled_height": lossStalledHeight,
		"direct_bypass_observed":   false,
	}))
	require.NoError(t, stopProxy("effective-loss-proxy-stop", "effective-loss", lossProxy))

	delayedConfig := baseConfig
	delayedConfig.Name = "delay-jitter"
	delayedConfig.Delay = 60 * time.Millisecond
	delayedConfig.Jitter = 30 * time.Millisecond
	delayedProxy, err := startProxy(delayedConfig)
	require.NoError(t, err)
	delayedTarget, err := validator.Height(ctx)
	require.NoError(t, err)
	delayedTarget += 4
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, delayedTarget))
	require.NoError(t, network.WaitForFullNode(ctx, delayedTarget))
	delayedHistory, err := network.RequireSameHistoryAtHeight(ctx, delayedTarget, validator, fullNode)
	require.NoError(t, err)
	require.NoError(t, recordChainP2PRuntimeFault(network, "delay-jitter", harness.NetworkFaultOutcomeRecovered))
	require.NoError(t, network.WriteArtifactJSON("network-faults/delay-jitter.json", map[string]any{
		"delay":         delayedConfig.Delay.String(),
		"jitter":        delayedConfig.Jitter.String(),
		"target_height": delayedTarget,
		"same_history":  delayedHistory,
	}))
	resolvedBeforeReplacement, err := resolveDockerAliasFromNode(ctx, fullNode, proxyAlias)
	require.NoError(t, err)
	require.Contains(t, resolvedBeforeReplacement, delayedProxy.IPAddress)

	replacementConfig := baseConfig
	replacementConfig.Name = "dns-replacement"
	replacementConfig.Delay = 10 * time.Millisecond
	replacementProxy, err := startProxy(replacementConfig)
	require.NoError(t, err)
	require.NotEqual(t, delayedProxy.ContainerID, replacementProxy.ContainerID)
	require.NotEqual(t, delayedProxy.IPAddress, replacementProxy.IPAddress)
	fullNodeProxyClientIP, err := network.RunNetworkIPAddress(ctx, fullNode.ContainerID())
	require.NoError(t, err)
	require.NoError(t, stopProxy("delay-jitter-proxy-stop", "delay-jitter", delayedProxy))

	oldIPObserverConfig := baseConfig
	oldIPObserverConfig.Name = "dns-old-ip-observer"
	oldIPObserverConfig.Alias = "dns-old-ip-observer"
	oldIPObserverConfig.IPv4Address = delayedProxy.IPAddress
	oldIPObserverConfig.DropEvery = 1
	oldIPObserver, err := startProxy(oldIPObserverConfig)
	require.NoError(t, err)
	require.Equal(t, delayedProxy.IPAddress, oldIPObserver.IPAddress)
	resolvedAfterReplacement, err := waitForDockerAliasResolution(
		ctx,
		fullNode,
		proxyAlias,
		replacementProxy.IPAddress,
		delayedProxy.IPAddress,
		10*time.Second,
	)
	require.NoError(t, err)
	oldIPConnection, err := waitForFaultProxyConnectionFromIP(
		ctx,
		network,
		oldIPObserver,
		fullNodeProxyClientIP,
		15*time.Second,
	)
	require.NoError(t, err)
	require.NoError(t, waitForFaultProxyLog(
		ctx,
		network,
		oldIPObserver,
		`"event":"chunk-dropped"`,
		5*time.Second,
	))
	replacementLogsBeforeRestart, err := network.P2PFaultProxyLogs(ctx, replacementProxy)
	require.NoError(t, err)
	replacementEventsBeforeRestart, err := decodeNetworkFaultProxyLogEvents(replacementLogsBeforeRestart)
	require.NoError(t, err)
	require.Zero(
		t,
		networkFaultProxyEventCount(replacementEventsBeforeRestart, "connection-opened"),
		"replacement proxy must not receive the cached peer connection before full-node restart",
	)
	validatorBeforeDNSStall, err := validator.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, validatorBeforeDNSStall+2))
	dnsStallCtx, dnsStallCancel := context.WithTimeout(ctx, 8*time.Second)
	fullNodeHeightBeforeRestart, dnsStallErr := network.WaitForStableHeight(dnsStallCtx, fullNode, 2*time.Second)
	dnsStallCancel()
	require.NoError(t, dnsStallErr, "CometBFT must expose its cached old peer IP before the bounded restart recovery")
	validatorAfterDNSStall, err := validator.Height(ctx)
	require.NoError(t, err)
	require.Less(t, fullNodeHeightBeforeRestart, validatorAfterDNSStall)
	require.NoError(t, recordChainP2PRuntimeFault(network, "proxy-dns-cached-old-ip", harness.NetworkFaultOutcomeExpectedFaultObserved))

	dnsRecoveryRestart, err := network.GracefulRestartNode(ctx, fullNode)
	require.NoError(t, err)
	resolvedAfterRestart, err := resolveDockerAliasFromNode(ctx, fullNode, proxyAlias)
	require.NoError(t, err)
	require.Contains(t, resolvedAfterRestart, replacementProxy.IPAddress)
	require.NotContains(t, resolvedAfterRestart, delayedProxy.IPAddress)
	replacementConnection, err := waitForFaultProxyConnectionFromIP(
		ctx,
		network,
		replacementProxy,
		fullNodeProxyClientIP,
		10*time.Second,
	)
	require.NoError(t, err)
	replacementTarget, err := validator.Height(ctx)
	require.NoError(t, err)
	replacementTarget += 3
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, replacementTarget))
	require.NoError(t, network.WaitForFullNode(ctx, replacementTarget))
	replacementHistory, err := network.RequireSameHistoryAtHeight(ctx, replacementTarget, validator, fullNode)
	require.NoError(t, err)
	require.NoError(t, recordChainP2PRuntimeFault(network, "proxy-dns-recreation", harness.NetworkFaultOutcomeRecovered))
	require.NoError(t, network.WriteArtifactJSON("network-faults/proxy-recreation.json", map[string]any{
		"old_container_id":           delayedProxy.ContainerID,
		"new_container_id":           replacementProxy.ContainerID,
		"old_ip_address":             delayedProxy.IPAddress,
		"new_ip_address":             replacementProxy.IPAddress,
		"resolved_before":            resolvedBeforeReplacement,
		"resolved_after":             resolvedAfterReplacement,
		"resolved_after_restart":     resolvedAfterRestart,
		"same_dns_alias":             proxyAlias,
		"ip_changed":                 true,
		"hot_dns_refresh_supported":  false,
		"cached_old_ip_stall_height": fullNodeHeightBeforeRestart,
		"validator_after_dns_stall":  validatorAfterDNSStall,
		"old_ip_observer": map[string]any{
			"container_id":  oldIPObserver.ContainerID,
			"ip_address":    oldIPObserver.IPAddress,
			"connection":    oldIPConnection,
			"drop_observed": true,
		},
		"replacement_connections_before_restart": 0,
		"restart_required":                       true,
		"restart":                                dnsRecoveryRestart,
		"replacement_connection":                 replacementConnection,
		"reconnected":                            true,
		"target_height":                          replacementTarget,
		"same_history":                           replacementHistory,
	}))

	require.NoError(t, network.RecordNetworkFaultCleanup(
		"p2p-proxy-route-restore",
		network.RestoreFullNodeP2PRoute(ctx, applied),
	))
	route = nil
	require.NoError(t, stopProxy("dns-old-ip-observer-stop", "dns-old-ip-observer", oldIPObserver))
	require.NoError(t, stopProxy("dns-replacement-proxy-stop", "dns-replacement", replacementProxy))
	directTarget, err := validator.Height(ctx)
	require.NoError(t, err)
	directTarget += 2
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, directTarget))
	require.NoError(t, network.WaitForFullNode(ctx, directTarget))
}

func resolveDockerAliasFromNode(
	ctx context.Context,
	node *cosmos.ChainNode,
	alias string,
) ([]string, error) {
	if node == nil {
		return nil, errors.New("DNS resolution node is required")
	}
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return nil, errors.New("DNS alias is required")
	}
	stdout, stderr, err := node.Exec(ctx, []string{"getent", "ahostsv4", alias}, node.Chain.Config().Env)
	if err != nil {
		return nil, fmt.Errorf("resolve Docker alias %s from %s: %w: %s", alias, node.Name(), err, strings.TrimSpace(string(stderr)))
	}
	addresses := make(map[string]struct{})
	for _, line := range strings.Split(string(stdout), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		parsed := net.ParseIP(fields[0])
		if parsed == nil || parsed.To4() == nil {
			continue
		}
		addresses[parsed.String()] = struct{}{}
	}
	result := make([]string, 0, len(addresses))
	for address := range addresses {
		result = append(result, address)
	}
	sort.Strings(result)
	if len(result) == 0 {
		return nil, fmt.Errorf("Docker alias %s resolved to no IPv4 addresses from %s", alias, node.Name())
	}
	return result, nil
}

func waitForDockerAliasResolution(
	ctx context.Context,
	node *cosmos.ChainNode,
	alias string,
	expectedIP string,
	forbiddenIP string,
	timeout time.Duration,
) ([]string, error) {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	var lastAddresses []string
	var lastErr error
	for {
		lastAddresses, lastErr = resolveDockerAliasFromNode(waitCtx, node, alias)
		if lastErr == nil && slicesContainString(lastAddresses, expectedIP) && !slicesContainString(lastAddresses, forbiddenIP) {
			return lastAddresses, nil
		}
		select {
		case <-waitCtx.Done():
			return nil, fmt.Errorf(
				"wait for Docker alias %s to move from %s to %s: %w; addresses=%v last_error=%v",
				alias,
				forbiddenIP,
				expectedIP,
				waitCtx.Err(),
				lastAddresses,
				lastErr,
			)
		case <-ticker.C:
		}
	}
}

func slicesContainString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func decodeNetworkFaultProxyLogEvents(logs []byte) ([]networkFaultProxyLogEvent, error) {
	decoder := json.NewDecoder(bytes.NewReader(logs))
	events := make([]networkFaultProxyLogEvent, 0)
	for {
		var event networkFaultProxyLogEvent
		if err := decoder.Decode(&event); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decode fault proxy JSON event %d: %w", len(events)+1, err)
		}
		event.Event = strings.TrimSpace(event.Event)
		if event.Event == "" {
			return nil, fmt.Errorf("fault proxy JSON event %d has no event name", len(events)+1)
		}
		events = append(events, event)
	}
	if len(events) == 0 {
		return nil, errors.New("fault proxy logs contain no JSON events")
	}
	return events, nil
}

func networkFaultProxyEventCount(events []networkFaultProxyLogEvent, eventName string) int {
	eventName = strings.TrimSpace(eventName)
	count := 0
	for _, event := range events {
		if event.Event == eventName {
			count++
		}
	}
	return count
}

func networkFaultProxyConnectionFromIP(
	logs []byte,
	expectedIP string,
) (networkFaultProxyLogEvent, error) {
	expected := net.ParseIP(strings.TrimSpace(expectedIP))
	if expected == nil || expected.To4() == nil {
		return networkFaultProxyLogEvent{}, fmt.Errorf("expected fault proxy client IPv4 address %q is invalid", expectedIP)
	}
	events, err := decodeNetworkFaultProxyLogEvents(logs)
	if err != nil {
		return networkFaultProxyLogEvent{}, err
	}
	observedClients := make([]string, 0)
	for _, event := range events {
		if event.Event != "connection-opened" {
			continue
		}
		host, _, splitErr := net.SplitHostPort(strings.TrimSpace(event.Client))
		if splitErr != nil {
			return networkFaultProxyLogEvent{}, fmt.Errorf(
				"fault proxy connection %d has invalid client address %q: %w",
				event.ConnectionID,
				event.Client,
				splitErr,
			)
		}
		clientIP := net.ParseIP(host)
		if clientIP == nil || clientIP.To4() == nil {
			return networkFaultProxyLogEvent{}, fmt.Errorf(
				"fault proxy connection %d client %q is not IPv4",
				event.ConnectionID,
				event.Client,
			)
		}
		observedClients = append(observedClients, clientIP.String())
		if clientIP.Equal(expected) {
			return event, nil
		}
	}
	return networkFaultProxyLogEvent{}, fmt.Errorf(
		"fault proxy has no connection from %s; observed clients=%v",
		expected.String(),
		observedClients,
	)
}

func waitForFaultProxyConnectionFromIP(
	ctx context.Context,
	network *harness.Network,
	proxy harness.P2PFaultProxy,
	expectedIP string,
	timeout time.Duration,
) (networkFaultProxyLogEvent, error) {
	if network == nil {
		return networkFaultProxyLogEvent{}, errors.New("fault proxy network is required")
	}
	if strings.TrimSpace(proxy.ContainerID) == "" {
		return networkFaultProxyLogEvent{}, errors.New("fault proxy container ID is required")
	}
	if timeout <= 0 {
		return networkFaultProxyLogEvent{}, fmt.Errorf("fault proxy connection timeout must be positive, got %s", timeout)
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	var lastErr error
	for {
		logs, logsErr := network.P2PFaultProxyLogs(waitCtx, proxy)
		if logsErr != nil {
			lastErr = logsErr
		} else {
			event, connectionErr := networkFaultProxyConnectionFromIP(logs, expectedIP)
			if connectionErr == nil {
				return event, nil
			}
			lastErr = connectionErr
		}
		select {
		case <-waitCtx.Done():
			return networkFaultProxyLogEvent{}, fmt.Errorf(
				"wait for fault proxy %s connection from %s: %w; last error: %v",
				proxy.Config.Name,
				expectedIP,
				waitCtx.Err(),
				lastErr,
			)
		case <-ticker.C:
		}
	}
}

func waitForFaultProxyLog(
	ctx context.Context,
	network *harness.Network,
	proxy harness.P2PFaultProxy,
	needle string,
	timeout time.Duration,
) error {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		logs, err := network.P2PFaultProxyLogs(waitCtx, proxy)
		if err == nil && bytes.Contains(logs, []byte(needle)) {
			return nil
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("wait for fault proxy log %q: %w; last error: %v", needle, waitCtx.Err(), err)
		case <-ticker.C:
		}
	}
}

func networkFaultExerciseContainerRecreation(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	validator *cosmos.ChainNode,
	fullNode *cosmos.ChainNode,
) {
	t.Helper()
	oldContainerID := fullNode.ContainerID()
	oldIPAddress, err := network.RunNetworkIPAddress(ctx, oldContainerID)
	require.NoError(t, err)
	resolvedBefore, err := resolveDockerAliasFromNode(ctx, validator, fullNode.HostName())
	require.NoError(t, err)
	require.Contains(t, resolvedBefore, oldIPAddress)
	beforeHeight, err := fullNode.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, fullNode.StopContainer(ctx))
	require.NoError(t, fullNode.RemoveContainer(ctx))
	guardConfig := harness.P2PFaultProxyConfig{
		Name:          "full-node-ip-guard",
		Alias:         "full-node-ip-guard",
		Image:         harness.CurrentImage(),
		TargetAddress: validator.HostName() + ":26656",
		IPv4Address:   oldIPAddress,
		Seed:          networkFaultSeed,
	}
	guard, err := network.StartP2PFaultProxy(ctx, guardConfig)
	require.NoError(t, err)
	guardActive := true
	t.Cleanup(func() {
		if !guardActive {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cleanupCancel()
		recordNetworkFaultCleanup(
			t,
			network,
			"full-node-ip-guard",
			network.StopP2PFaultProxy(cleanupCtx, "cleanup-full-node-ip-guard", guard),
		)
	})
	require.Equal(t, oldIPAddress, guard.IPAddress)
	require.NoError(t, fullNode.CreateNodeContainer(ctx))
	require.NoError(t, fullNode.StartContainer(ctx))
	require.NotEqual(t, oldContainerID, fullNode.ContainerID())
	newIPAddress, err := network.RunNetworkIPAddress(ctx, fullNode.ContainerID())
	require.NoError(t, err)
	require.NotEqual(t, oldIPAddress, newIPAddress)
	require.NoError(t, network.RecordNetworkFaultCleanup(
		"full-node-ip-guard-stop",
		network.StopP2PFaultProxy(ctx, "full-node-ip-guard", guard),
	))
	guardActive = false
	resolvedAfter, err := waitForDockerAliasResolution(
		ctx,
		validator,
		fullNode.HostName(),
		newIPAddress,
		oldIPAddress,
		10*time.Second,
	)
	require.NoError(t, err)
	target, err := validator.Height(ctx)
	require.NoError(t, err)
	target += 2
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, target))
	require.NoError(t, network.WaitForFullNode(ctx, target))
	_, err = network.RequireSameHistoryAtHeight(ctx, target, validator, fullNode)
	require.NoError(t, err)
	require.NoError(t, recordChainP2PRuntimeFault(network, "full-node-container-recreation", harness.NetworkFaultOutcomeRecovered))
	require.NoError(t, network.WriteArtifactJSON("network-faults/full-node-container-recreation.json", map[string]any{
		"old_container_id":       oldContainerID,
		"new_container_id":       fullNode.ContainerID(),
		"old_ip_address":         oldIPAddress,
		"new_ip_address":         newIPAddress,
		"resolved_before":        resolvedBefore,
		"resolved_after":         resolvedAfter,
		"before_height":          beforeHeight,
		"recovered_height":       target,
		"hostname":               fullNode.HostName(),
		"ip_changed":             true,
		"dns_resolution_updated": true,
		"p2p_recovered":          true,
	}))
}

func networkFaultExerciseEndpointIsolation(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	validator *cosmos.ChainNode,
	fullNode *cosmos.ChainNode,
) {
	t.Helper()
	original, err := fullNode.ReadFile(ctx, "config/app.toml")
	require.NoError(t, err)
	mutated := false
	t.Cleanup(func() {
		if !mutated {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cleanupCancel()
		recordNetworkFaultCleanup(
			t,
			network,
			"endpoint-app-config",
			restartFullNodeWithAppConfig(cleanupCtx, network, "cleanup-endpoint-app-config", validator, fullNode, original),
		)
	})
	rpcAddress, err := network.FullNodeHostAddress(ctx, "26657/tcp")
	require.NoError(t, err)
	restAddress, err := network.FullNodeHostAddress(ctx, "1317/tcp")
	require.NoError(t, err)
	balanceAddress, err := validator.AccountKeyBech32(ctx, "validator")
	require.NoError(t, err)

	apiDisabled, err := rewriteNetworkFaultAppConfig(original, networkFaultEndpointMutation{APIEnabled: boolPointer(false)})
	require.NoError(t, err)
	mutated = true
	require.NoError(t, restartFullNodeWithAppConfig(ctx, network, "api-disabled", validator, fullNode, apiDisabled))
	restCtx, restCancel := context.WithTimeout(ctx, 2*time.Second)
	restErr := harness.RequireRESTNodeInfo(restCtx, &http.Client{Timeout: time.Second}, restAddress)
	restCancel()
	require.Error(t, restErr)
	grpcAvailableCtx, grpcAvailableCancel := context.WithTimeout(ctx, 3*time.Second)
	_, err = banktypes.NewQueryClient(fullNode.GrpcConn).Balance(grpcAvailableCtx, &banktypes.QueryBalanceRequest{
		Address: balanceAddress,
		Denom:   "umed",
	})
	grpcAvailableCancel()
	require.NoError(t, err)
	height, err := validator.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, height+2))
	rpcAvailableCtx, rpcAvailableCancel := context.WithTimeout(ctx, 3*time.Second)
	require.NoError(t, harness.RequireRPCStatus(rpcAvailableCtx, rpcAddress, height))
	rpcAvailableCancel()

	grpcDisabled, err := rewriteNetworkFaultAppConfig(original, networkFaultEndpointMutation{GRPCEnabled: boolPointer(false)})
	require.NoError(t, err)
	require.NoError(t, restartFullNodeWithAppConfig(ctx, network, "grpc-disabled", validator, fullNode, grpcDisabled))
	grpcCtx, grpcCancel := context.WithTimeout(ctx, 2*time.Second)
	_, grpcErr := banktypes.NewQueryClient(fullNode.GrpcConn).Balance(grpcCtx, &banktypes.QueryBalanceRequest{
		Address: balanceAddress,
		Denom:   "umed",
	})
	grpcCancel()
	require.Error(t, grpcErr)
	require.NoError(t, harness.RequireRESTNodeInfo(ctx, &http.Client{Timeout: 3 * time.Second}, restAddress))
	rpcStillAvailableCtx, rpcStillAvailableCancel := context.WithTimeout(ctx, 3*time.Second)
	require.NoError(t, harness.RequireRPCStatus(rpcStillAvailableCtx, rpcAddress, height))
	rpcStillAvailableCancel()

	require.NoError(t, network.RecordNetworkFaultCleanup(
		"endpoint-app-config",
		restartFullNodeWithAppConfig(ctx, network, "endpoint-app-config-restored", validator, fullNode, original),
	))
	mutated = false
	grpcRestoredCtx, grpcRestoredCancel := context.WithTimeout(ctx, 3*time.Second)
	_, err = banktypes.NewQueryClient(fullNode.GrpcConn).Balance(grpcRestoredCtx, &banktypes.QueryBalanceRequest{
		Address: balanceAddress,
		Denom:   "umed",
	})
	grpcRestoredCancel()
	require.NoError(t, err)
	require.NoError(t, harness.RequireRESTNodeInfo(ctx, &http.Client{Timeout: 3 * time.Second}, restAddress))
	rpcIsolation := networkFaultExerciseRPCListenerIsolation(
		t,
		ctx,
		network,
		validator,
		fullNode,
		rpcAddress,
		restAddress,
	)
	require.NoError(t, network.WriteArtifactJSON("network-faults/endpoint-isolation.json", map[string]any{
		"rest_disabled_error":          restErr.Error(),
		"grpc_disabled_error":          grpcErr.Error(),
		"rpc_listener_fault":           rpcIsolation,
		"rpc_remained_available":       true,
		"consensus_remained_available": true,
		"restored":                     true,
	}))
}

func networkFaultExerciseRPCListenerIsolation(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	validator *cosmos.ChainNode,
	fullNode *cosmos.ChainNode,
	rpcAddress string,
	restAddress string,
) map[string]any {
	t.Helper()
	validatorBefore, err := validator.Height(ctx)
	require.NoError(t, err)
	beforeGRPCCtx, beforeGRPCCancel := context.WithTimeout(ctx, 3*time.Second)
	fullNodeBefore, err := networkFaultLatestGRPCHeight(beforeGRPCCtx, fullNode)
	beforeGRPCCancel()
	require.NoError(t, err)

	var fault harness.FullNodeRPCBoundaryFault
	faultActive := false
	t.Cleanup(func() {
		if !faultActive {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cleanupCancel()
		var cleanupErrors []error
		if cleanupErr := network.RestoreFullNodeRPCBoundaryFault(cleanupCtx, fullNode, fault); cleanupErr != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("restore full-node RPC listener fault: %w", cleanupErr))
		} else {
			validatorHeight, heightErr := validator.Height(cleanupCtx)
			if heightErr != nil {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("query validator height after RPC fault cleanup: %w", heightErr))
			} else if catchUpErr := network.WaitForFullNode(cleanupCtx, validatorHeight); catchUpErr != nil {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("wait for full node after RPC fault cleanup: %w", catchUpErr))
			} else if _, historyErr := network.RequireSameHistoryAtHeight(cleanupCtx, validatorHeight, validator, fullNode); historyErr != nil {
				cleanupErrors = append(cleanupErrors, fmt.Errorf("compare history after RPC fault cleanup: %w", historyErr))
			}
		}
		recordNetworkFaultCleanup(t, network, "rpc-listener", errors.Join(cleanupErrors...))
	})
	fault, err = network.ApplyFullNodeRPCBoundaryFault(ctx, fullNode)
	require.NoError(t, err)
	faultActive = true

	restReadyCtx, restReadyCancel := context.WithTimeout(ctx, 12*time.Second)
	require.NoError(t, waitForNetworkFaultREST(restReadyCtx, restAddress))
	restReadyCancel()
	grpcReadyCtx, grpcReadyCancel := context.WithTimeout(ctx, 12*time.Second)
	_, err = waitForNetworkFaultGRPCHeight(grpcReadyCtx, fullNode, fullNodeBefore)
	grpcReadyCancel()
	require.NoError(t, err)
	rpcFailureCtx, rpcFailureCancel := context.WithTimeout(ctx, 2*time.Second)
	rpcConnectionError, err := probeNetworkFaultTCPConnectionFailure(rpcFailureCtx, rpcAddress)
	rpcFailureCancel()
	require.NoError(t, err)

	validatorTarget := validatorBefore + 3
	validatorProgressCtx, validatorProgressCancel := context.WithTimeout(ctx, 10*time.Second)
	require.NoError(t, network.WaitForNodeHeight(validatorProgressCtx, validator, validatorTarget))
	validatorProgressCancel()
	validatorDuring, err := validator.Height(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, validatorDuring, validatorTarget)
	fullNodeProgressCtx, fullNodeProgressCancel := context.WithTimeout(ctx, 12*time.Second)
	fullNodeDuring, err := waitForNetworkFaultGRPCHeight(fullNodeProgressCtx, fullNode, validatorTarget)
	fullNodeProgressCancel()
	require.NoError(t, err)
	require.Greater(t, fullNodeDuring, fullNodeBefore)
	duringRESTCtx, duringRESTCancel := context.WithTimeout(ctx, 3*time.Second)
	duringRESTErr := waitForNetworkFaultREST(duringRESTCtx, restAddress)
	duringRESTCancel()
	require.NoError(t, duringRESTErr)

	require.NoError(t, network.RecordNetworkFaultCleanup(
		"rpc-listener",
		network.RestoreFullNodeRPCBoundaryFault(ctx, fullNode, fault),
	))
	faultActive = false
	validatorAfter, err := validator.Height(ctx)
	require.NoError(t, err)
	restoredAgreement, err := network.WaitForQuorumAgreement(
		ctx,
		"network-fault-rpc-listener-restored",
		validatorAfter,
		validator,
		fullNode,
	)
	require.NoError(t, err)
	afterGRPCCtx, afterGRPCCancel := context.WithTimeout(ctx, 3*time.Second)
	fullNodeAfter, err := networkFaultLatestGRPCHeight(afterGRPCCtx, fullNode)
	afterGRPCCancel()
	require.NoError(t, err)
	require.GreaterOrEqual(t, fullNodeAfter, validatorAfter)
	rpcRestoredCtx, rpcRestoredCancel := context.WithTimeout(ctx, 3*time.Second)
	require.NoError(t, harness.RequireRPCStatus(rpcRestoredCtx, rpcAddress, validatorAfter))
	rpcRestoredCancel()
	require.NoError(t, harness.RequireRESTNodeInfo(ctx, &http.Client{Timeout: 3 * time.Second}, restAddress))
	return map[string]any{
		"original_listener":            fault.OriginalListener,
		"fault_listener":               fault.FaultListener,
		"restored_listener":            fault.OriginalListener,
		"validator_height_before":      validatorBefore,
		"full_node_grpc_height_before": fullNodeBefore,
		"validator_height_during":      validatorDuring,
		"full_node_grpc_height_during": fullNodeDuring,
		"validator_height_after":       validatorAfter,
		"full_node_grpc_height_after":  fullNodeAfter,
		"host_rpc_connection_error":    rpcConnectionError,
		"host_rpc_unreachable":         true,
		"rest_available_during":        true,
		"grpc_available_during":        true,
		"consensus_progressed":         true,
		"full_node_synced_during":      true,
		"normal_start_restored":        true,
		"same_history_after_restore":   restoredAgreement,
	}
}

func waitForNetworkFaultREST(ctx context.Context, restAddress string) error {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	client := &http.Client{Timeout: 2 * time.Second}
	var lastErr error
	for {
		attemptCtx, attemptCancel := context.WithTimeout(ctx, 2*time.Second)
		lastErr = harness.RequireRESTNodeInfo(attemptCtx, client, restAddress)
		attemptCancel()
		if lastErr == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for REST during RPC listener fault: %w; last error: %v", ctx.Err(), lastErr)
		case <-ticker.C:
		}
	}
}

func waitForNetworkFaultGRPCHeight(
	ctx context.Context,
	node *cosmos.ChainNode,
	target int64,
) (int64, error) {
	if target <= 0 {
		return 0, fmt.Errorf("gRPC height target must be positive, got %d", target)
	}
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	var lastHeight int64
	var lastErr error
	for {
		attemptCtx, attemptCancel := context.WithTimeout(ctx, 2*time.Second)
		lastHeight, lastErr = networkFaultLatestGRPCHeight(attemptCtx, node)
		attemptCancel()
		if lastErr == nil && lastHeight >= target {
			return lastHeight, nil
		}
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("wait for full-node gRPC height %d (last=%d): %w; last error: %v", target, lastHeight, ctx.Err(), lastErr)
		case <-ticker.C:
		}
	}
}

func networkFaultLatestGRPCHeight(ctx context.Context, node *cosmos.ChainNode) (int64, error) {
	if node == nil || node.GrpcConn == nil {
		return 0, errors.New("full-node gRPC connection is unavailable")
	}
	response, err := cmtservice.NewServiceClient(node.GrpcConn).GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
	if err != nil {
		return 0, err
	}
	var height int64
	if response != nil && response.SdkBlock != nil {
		height = response.SdkBlock.Header.Height
	} else if response != nil && response.Block != nil {
		height = response.Block.Header.Height
	}
	if height <= 0 {
		return 0, fmt.Errorf("full-node gRPC latest block returned invalid height %d", height)
	}
	return height, nil
}

func probeNetworkFaultTCPConnectionFailure(ctx context.Context, address string) (string, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("RPC address has no host: %q", address)
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return "", errors.New("host RPC failure probe requires a bounded context")
	}
	connection, dialErr := (&net.Dialer{}).DialContext(ctx, "tcp", parsed.Host)
	if dialErr != nil {
		return classifyNetworkFaultConnectionFailure(ctx, "dial", dialErr)
	}
	defer connection.Close()
	if err := connection.SetDeadline(deadline); err != nil {
		return "", fmt.Errorf("bound host RPC failure probe connection: %w", err)
	}
	request := []byte("GET /status HTTP/1.1\r\nHost: " + parsed.Host + "\r\nConnection: close\r\n\r\n")
	if _, err := connection.Write(request); err != nil {
		return classifyNetworkFaultConnectionFailure(ctx, "write", err)
	}
	var responseByte [1]byte
	if _, err := io.ReadFull(connection, responseByte[:]); err != nil {
		return classifyNetworkFaultConnectionFailure(ctx, "read", err)
	}
	return "", fmt.Errorf("host RPC unexpectedly returned response data (first byte %#x)", responseByte[0])
}

func classifyNetworkFaultConnectionFailure(ctx context.Context, operation string, connectionErr error) (string, error) {
	if ctx.Err() != nil {
		return "", fmt.Errorf("host RPC failure probe exceeded bound during %s: %w", operation, ctx.Err())
	}
	if timeout, ok := connectionErr.(net.Error); ok && timeout.Timeout() {
		return "", fmt.Errorf("host RPC failure probe timed out during %s instead of receiving a connection failure: %w", operation, connectionErr)
	}
	return operation + ": " + connectionErr.Error(), nil
}

func networkFaultExerciseGRPCMessageBoundary(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	validator *cosmos.ChainNode,
	fullNode *cosmos.ChainNode,
) {
	t.Helper()
	original, err := fullNode.ReadFile(ctx, "config/app.toml")
	require.NoError(t, err)
	limited, err := rewriteNetworkFaultAppConfig(original, networkFaultEndpointMutation{
		GRPCMaxRecvBytes: uintPointer(networkFaultGRPCMaxReceiveBytes),
	})
	require.NoError(t, err)
	require.NoError(t, restartFullNodeWithAppConfig(ctx, network, "grpc-message-limited", validator, fullNode, limited))
	mutated := true
	t.Cleanup(func() {
		if !mutated {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cleanupCancel()
		recordNetworkFaultCleanup(
			t,
			network,
			"grpc-message-config",
			restartFullNodeWithAppConfig(cleanupCtx, network, "cleanup-grpc-message-config", validator, fullNode, original),
		)
	})

	address, err := validator.AccountKeyBech32(ctx, "validator")
	require.NoError(t, err)
	baselineCtx, baselineCancel := context.WithTimeout(ctx, 5*time.Second)
	_, err = banktypes.NewQueryClient(fullNode.GrpcConn).AllBalances(
		baselineCtx,
		&banktypes.QueryAllBalancesRequest{Address: address},
	)
	baselineCancel()
	require.NoError(t, err)

	request := &banktypes.QueryAllBalancesRequest{Address: strings.Repeat("A", 4096)}
	requestBytes := proto.Size(request)
	require.Greater(t, requestBytes, int(networkFaultGRPCMaxReceiveBytes))
	heightBefore, err := validator.Height(ctx)
	require.NoError(t, err)
	oversizedCtx, oversizedCancel := context.WithTimeout(ctx, 3*time.Second)
	_, grpcErr := banktypes.NewQueryClient(fullNode.GrpcConn).AllBalances(oversizedCtx, request)
	oversizedCancel()
	require.NoError(t, validateOversizedGRPCRejection(
		grpcErr,
		int(networkFaultGRPCMaxReceiveBytes),
		requestBytes,
	))

	continuityCtx, continuityCancel := context.WithTimeout(ctx, 10*time.Second)
	validatorWaitErr := network.WaitForNodeHeight(continuityCtx, validator, heightBefore+2)
	fullNodeWaitErr := network.WaitForFullNode(continuityCtx, heightBefore+2)
	continuityCancel()
	require.NoError(t, errors.Join(validatorWaitErr, fullNodeWaitErr))
	_, err = network.RequireSameHistoryAtHeight(ctx, heightBefore+2, validator, fullNode)
	require.NoError(t, err)
	validatorHeightAfter, err := validator.Height(ctx)
	require.NoError(t, err)
	fullNodeHeightAfter, err := fullNode.Height(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, validatorHeightAfter, heightBefore+2)
	require.GreaterOrEqual(t, fullNodeHeightAfter, heightBefore+2)
	postFailureCtx, postFailureCancel := context.WithTimeout(ctx, 5*time.Second)
	_, err = banktypes.NewQueryClient(fullNode.GrpcConn).AllBalances(
		postFailureCtx,
		&banktypes.QueryAllBalancesRequest{Address: address},
	)
	postFailureCancel()
	require.NoError(t, err)
	require.NoError(t, network.WriteArtifactJSON("network-faults/oversized-grpc-message.json", map[string]any{
		"configured_max_receive_bytes": networkFaultGRPCMaxReceiveBytes,
		"request_message_bytes":        requestBytes,
		"grpc_code":                    status.Code(grpcErr).String(),
		"rejection":                    grpcErr.Error(),
		"height_before":                heightBefore,
		"validator_height_after":       validatorHeightAfter,
		"full_node_height_after":       fullNodeHeightAfter,
		"normal_query_after_rejection": true,
		"same_history":                 true,
		"bounded_timeout":              "3s",
	}))

	require.NoError(t, network.RecordNetworkFaultCleanup(
		"grpc-message-config",
		restartFullNodeWithAppConfig(ctx, network, "grpc-message-config-restored", validator, fullNode, original),
	))
	mutated = false
}

func networkFaultExerciseHTTPBoundaries(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	validator *cosmos.ChainNode,
	fullNode *cosmos.ChainNode,
) {
	t.Helper()
	original, err := fullNode.ReadFile(ctx, "config/app.toml")
	require.NoError(t, err)
	limited, err := rewriteNetworkFaultAppConfig(original, networkFaultEndpointMutation{
		ReadTimeout:  uintPointer(1),
		WriteTimeout: uintPointer(1),
		MaxBodyBytes: uintPointer(1024),
	})
	require.NoError(t, err)
	require.NoError(t, restartFullNodeWithAppConfig(ctx, network, "http-boundary-limited", validator, fullNode, limited))
	mutated := true
	t.Cleanup(func() {
		if !mutated {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cleanupCancel()
		recordNetworkFaultCleanup(
			t,
			network,
			"http-boundary-config",
			restartFullNodeWithAppConfig(cleanupCtx, network, "cleanup-http-boundary-config", validator, fullNode, original),
		)
	})
	restAddress, err := network.FullNodeHostAddress(ctx, "1317/tcp")
	require.NoError(t, err)
	require.NoError(t, harness.RequireRESTNodeInfo(ctx, &http.Client{Timeout: 3 * time.Second}, restAddress))
	slowEvidence, err := runSlowRESTClient(restAddress, 1500*time.Millisecond)
	require.NoError(t, err)
	slowEvidence["baseline_rest_healthy"] = true
	slowEvidence["configured_read_timeout_seconds"] = 1
	require.NoError(t, network.WriteArtifactJSON("network-faults/slow-client.json", slowEvidence))

	oversizedPayload := "{\"tx_bytes\":\"" + strings.Repeat("A", 4096) + "\",\"mode\":\"BROADCAST_MODE_SYNC\"}"
	heightBeforeOversized, err := fullNode.Height(ctx)
	require.NoError(t, err)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(restAddress, "/")+"/cosmos/tx/v1beta1/txs", strings.NewReader(oversizedPayload))
	require.NoError(t, err)
	request.Header.Set("content-type", "application/json")
	response, err := (&http.Client{Timeout: 3 * time.Second}).Do(request)
	require.NoError(t, err)
	responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 8192))
	closeErr := response.Body.Close()
	require.NoError(t, errors.Join(readErr, closeErr))
	require.NoError(t, validateOversizedRESTRejection(
		response.StatusCode,
		responseBody,
		1024,
		len(oversizedPayload),
	))
	continuityCtx, continuityCancel := context.WithTimeout(ctx, 8*time.Second)
	require.NoError(t, network.WaitForNodeHeight(continuityCtx, fullNode, heightBeforeOversized+2))
	continuityCancel()
	heightAfterOversized, err := fullNode.Height(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, heightAfterOversized, heightBeforeOversized+2)
	require.NoError(t, network.WriteArtifactJSON("network-faults/oversized-body.json", map[string]any{
		"configured_max_body_bytes": 1024,
		"request_body_bytes":        len(oversizedPayload),
		"status_code":               response.StatusCode,
		"rejection_reason":          networkFaultOversizedBodyMarker,
		"response":                  string(responseBody),
		"height_before":             heightBeforeOversized,
		"height_after":              heightAfterOversized,
		"node_progressed":           true,
		"bounded":                   true,
	}))

	require.NoError(t, network.RecordNetworkFaultCleanup(
		"http-boundary-config",
		restartFullNodeWithAppConfig(ctx, network, "http-boundary-config-restored", validator, fullNode, original),
	))
	mutated = false
	rpcAddress, err := network.FullNodeHostAddress(ctx, "26657/tcp")
	require.NoError(t, err)
	const churnRequests = 40
	var wait sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	var churnErrors []string
	var churnHeights []int64
	for index := 0; index < churnRequests; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			transport := &http.Transport{DisableKeepAlives: true}
			client := &http.Client{Timeout: 3 * time.Second, Transport: transport}
			defer transport.CloseIdleConnections()
			churnCtx, churnCancel := context.WithTimeout(ctx, 3*time.Second)
			defer churnCancel()
			request, err := http.NewRequestWithContext(
				churnCtx,
				http.MethodGet,
				strings.TrimRight(rpcAddress, "/")+"/status",
				nil,
			)
			if err == nil {
				var response *http.Response
				response, err = client.Do(request)
				if err == nil {
					body, readErr := io.ReadAll(io.LimitReader(response.Body, (1<<20)+1))
					closeErr := response.Body.Close()
					err = errors.Join(readErr, closeErr)
					if err == nil && len(body) > 1<<20 {
						err = errors.New("RPC churn response exceeded 1 MiB")
					}
					if err == nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
						err = fmt.Errorf("RPC churn response status %s", response.Status)
					}
					var height int64
					if err == nil {
						height, err = networkFaultRPCStatusHeight(body)
					}
					if err == nil {
						mu.Lock()
						churnHeights = append(churnHeights, height)
						mu.Unlock()
					}
				}
			}
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				churnErrors = append(churnErrors, err.Error())
			} else {
				successes++
			}
		}()
	}
	wait.Wait()
	require.Empty(t, churnErrors)
	require.Equal(t, churnRequests, successes)
	require.Len(t, churnHeights, churnRequests)
	minimumHeight := churnHeights[0]
	maximumHeight := churnHeights[0]
	for _, height := range churnHeights[1:] {
		if height < minimumHeight {
			minimumHeight = height
		}
		if height > maximumHeight {
			maximumHeight = height
		}
	}
	require.NoError(t, network.WriteArtifactJSON("network-faults/connection-churn.json", map[string]any{
		"requests":        churnRequests,
		"successes":       successes,
		"errors":          churnErrors,
		"minimum_height":  minimumHeight,
		"maximum_height":  maximumHeight,
		"bounded_timeout": "3s",
	}))
}

func networkFaultRPCStatusHeight(body []byte) (int64, error) {
	var response struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("decode RPC churn status: %w", err)
	}
	height, err := strconv.ParseInt(response.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil || height <= 0 {
		return 0, fmt.Errorf("invalid RPC churn height %q", response.Result.SyncInfo.LatestBlockHeight)
	}
	return height, nil
}

func networkFaultExerciseWebSocketReconnect(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	validator *cosmos.ChainNode,
	fullNode *cosmos.ChainNode,
) {
	t.Helper()
	rpcAddress, err := network.FullNodeHostAddress(ctx, "26657/tcp")
	require.NoError(t, err)
	session, err := startNetworkFaultWebSocketSession(ctx, rpcAddress, "network-fault-reconnect")
	require.NoError(t, err)
	sessionActive := true
	t.Cleanup(func() {
		if !sessionActive {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), networkFaultWebSocketCleanupTimeout)
		defer cleanupCancel()
		recordNetworkFaultCleanup(t, network, "websocket-session", session.Close(cleanupCtx))
	})

	recipient, err := validator.AccountKeyBech32(ctx, "validator")
	require.NoError(t, err)
	before, beforeTx, _, err := collectNetworkFaultWebSocketPhase(
		ctx,
		session,
		0,
		func() (*harness.TxResult, error) {
			return network.BroadcastAndWaitTx(
				ctx,
				"network-fault-websocket-before",
				validator,
				"validator",
				"bank", "send", "validator", recipient, "1umed",
				"--gas", "200000", "--broadcast-mode", "sync",
			)
		},
	)
	require.NoError(t, err)

	var fault harness.FullNodeRPCBoundaryFault
	faultActive := false
	t.Cleanup(func() {
		if !faultActive {
			return
		}
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cleanupCancel()
		recordNetworkFaultCleanup(
			t,
			network,
			"websocket-rpc-listener",
			network.RestoreFullNodeRPCBoundaryFault(cleanupCtx, fullNode, fault),
		)
	})
	fault, err = network.ApplyFullNodeRPCBoundaryFault(ctx, fullNode)
	require.NoError(t, err)
	faultActive = true
	failureCtx, failureCancel := context.WithTimeout(ctx, 2*time.Second)
	rpcConnectionError, err := probeNetworkFaultTCPConnectionFailure(failureCtx, rpcAddress)
	failureCancel()
	require.NoError(t, err)

	faultStartHeight, err := validator.Height(ctx)
	require.NoError(t, err)
	faultTargetHeight := faultStartHeight + 3
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, faultTargetHeight))
	fullNodeFaultCtx, fullNodeFaultCancel := context.WithTimeout(ctx, 12*time.Second)
	fullNodeFaultHeight, err := waitForNetworkFaultGRPCHeight(fullNodeFaultCtx, fullNode, faultTargetHeight)
	fullNodeFaultCancel()
	require.NoError(t, err)

	require.NoError(t, network.RecordNetworkFaultCleanup(
		"websocket-rpc-listener",
		network.RestoreFullNodeRPCBoundaryFault(ctx, fullNode, fault),
	))
	faultActive = false
	after, afterTx, staleBeforeReconnect, err := collectNetworkFaultWebSocketPhase(
		ctx,
		session,
		fullNodeFaultHeight,
		func() (*harness.TxResult, error) {
			return network.BroadcastAndWaitTx(
				ctx,
				"network-fault-websocket-after",
				validator,
				"validator",
				"bank", "send", "validator", recipient, "1umed",
				"--gas", "200000", "--broadcast-mode", "sync",
			)
		},
	)
	require.NoError(t, err)
	before.BlockHeights = append(before.BlockHeights, staleBeforeReconnect...)
	continuity, err := validateNetworkFaultWebSocketReconnect(
		before,
		after,
		[]string{beforeTx.TxHash, afterTx.TxHash},
	)
	require.NoError(t, err)
	require.NotEmpty(t, continuity.FaultMissingBlockHeights, "the induced RPC outage must create an explicitly accounted event gap")

	cleanupCtx, cleanupCancel := context.WithTimeout(ctx, networkFaultWebSocketCleanupTimeout)
	err = session.Close(cleanupCtx)
	cleanupCancel()
	require.NoError(t, network.RecordNetworkFaultCleanup("websocket-session", err))
	sessionActive = false
	require.NoError(t, network.WriteArtifactJSON("network-faults/websocket-reconnect.json", map[string]any{
		"same_client":                         true,
		"subscriptions":                       []string{cmttypes.EventNewBlock, cmttypes.EventTx},
		"fault":                               "full-node-rpc-listener-isolated",
		"fault_listener":                      fault.FaultListener,
		"rpc_connection_error":                rpcConnectionError,
		"fault_start_height":                  faultStartHeight,
		"full_node_height_during_fault":       fullNodeFaultHeight,
		"post_fault_first_height":             after.BlockHeights[0],
		"post_fault_event_above_fault_height": after.BlockHeights[0] > fullNodeFaultHeight,
		"stale_pre_reconnect_heights":         staleBeforeReconnect,
		"continuity":                          continuity,
		"reconnected_and_resubscribed":        true,
	}))
}

type networkFaultWebSocketSession struct {
	client       *rpchttp.HTTP
	subscriber   string
	blocks       <-chan coretypes.ResultEvent
	transactions <-chan coretypes.ResultEvent
	closed       bool
}

func startNetworkFaultWebSocketSession(
	ctx context.Context,
	rpcAddress string,
	subscriber string,
) (*networkFaultWebSocketSession, error) {
	operationCtx, operationCancel := context.WithTimeout(ctx, networkFaultWebSocketOperationTimeout)
	defer operationCancel()
	client, err := rpchttp.New(rpcAddress, "/websocket")
	if err != nil {
		return nil, err
	}
	if err := client.Start(); err != nil {
		return nil, err
	}
	blocks, err := client.Subscribe(
		operationCtx,
		subscriber,
		cmttypes.QueryForEvent(cmttypes.EventNewBlock).String(),
		128,
	)
	if err != nil {
		return nil, errors.Join(err, client.Stop())
	}
	transactions, err := client.Subscribe(
		operationCtx,
		subscriber,
		cmttypes.QueryForEvent(cmttypes.EventTx).String(),
		32,
	)
	if err != nil {
		return nil, errors.Join(err, client.Stop())
	}
	return &networkFaultWebSocketSession{
		client:       client,
		subscriber:   subscriber,
		blocks:       blocks,
		transactions: transactions,
	}, nil
}

func (s *networkFaultWebSocketSession) Close(ctx context.Context) error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	var unsubscribeErr error
	if s.client != nil && s.client.IsRunning() {
		unsubscribeErr = s.client.UnsubscribeAll(ctx, s.subscriber)
		if unsubscribeErr != nil {
			unsubscribeErr = fmt.Errorf("unsubscribe WebSocket client %s: %w", s.subscriber, unsubscribeErr)
		}
	}
	var stopErr error
	if s.client != nil && s.client.IsRunning() {
		stopErr = s.client.Stop()
		if stopErr != nil {
			stopErr = fmt.Errorf("stop WebSocket client %s: %w", s.subscriber, stopErr)
		}
	}
	return errors.Join(unsubscribeErr, stopErr)
}

func collectNetworkFaultWebSocketPhase(
	ctx context.Context,
	session *networkFaultWebSocketSession,
	minimumFirstHeight int64,
	broadcast func() (*harness.TxResult, error),
) (networkFaultWebSocketPhase, *harness.TxResult, []int64, error) {
	var phase networkFaultWebSocketPhase
	if session == nil || session.client == nil || !session.client.IsRunning() {
		return phase, nil, nil, errors.New("running WebSocket session is required")
	}
	if broadcast == nil {
		return phase, nil, nil, errors.New("WebSocket transaction broadcaster is required")
	}
	operationCtx, operationCancel := context.WithTimeout(ctx, networkFaultWebSocketOperationTimeout)
	defer operationCancel()
	firstHeight, stale, err := waitForNetworkFaultWebSocketBlockAbove(operationCtx, session.blocks, minimumFirstHeight)
	if err != nil {
		return phase, nil, stale, err
	}
	phase.BlockHeights = append(phase.BlockHeights, firstHeight)
	transaction, err := broadcast()
	if err != nil {
		return phase, transaction, stale, err
	}
	if transaction == nil || strings.TrimSpace(transaction.TxHash) == "" || transaction.HeightInt64() <= 0 {
		return phase, transaction, stale, errors.New("WebSocket phase broadcast returned no committed transaction identity")
	}
	wantedHash := strings.ToUpper(transaction.TxHash)
	wantedObserved := false
	for !wantedObserved || len(phase.BlockHeights) < 2 || phase.BlockHeights[len(phase.BlockHeights)-1] < transaction.HeightInt64() {
		select {
		case <-operationCtx.Done():
			return phase, transaction, stale, fmt.Errorf("collect WebSocket phase: %w", operationCtx.Err())
		case event, open := <-session.blocks:
			if !open {
				return phase, transaction, stale, errors.New("WebSocket block subscription closed early")
			}
			height, decodeErr := decodeNetworkFaultWebSocketBlock(event)
			if decodeErr != nil {
				return phase, transaction, stale, decodeErr
			}
			phase.BlockHeights = append(phase.BlockHeights, height)
		case event, open := <-session.transactions:
			if !open {
				return phase, transaction, stale, errors.New("WebSocket transaction subscription closed early")
			}
			observed, decodeErr := decodeNetworkFaultWebSocketTransaction(event)
			if decodeErr != nil {
				return phase, transaction, stale, decodeErr
			}
			phase.Transactions = append(phase.Transactions, observed)
			if strings.EqualFold(observed.TxHash, wantedHash) {
				wantedObserved = true
			}
		}
	}
	return phase, transaction, stale, nil
}

func waitForNetworkFaultWebSocketBlockAbove(
	ctx context.Context,
	blocks <-chan coretypes.ResultEvent,
	minimumHeight int64,
) (int64, []int64, error) {
	var stale []int64
	for {
		select {
		case <-ctx.Done():
			return 0, stale, fmt.Errorf("wait for WebSocket block above %d: %w", minimumHeight, ctx.Err())
		case event, open := <-blocks:
			if !open {
				return 0, stale, errors.New("WebSocket block subscription closed early")
			}
			height, err := decodeNetworkFaultWebSocketBlock(event)
			if err != nil {
				return 0, stale, err
			}
			if height <= minimumHeight {
				stale = append(stale, height)
				continue
			}
			return height, stale, nil
		}
	}
}

func decodeNetworkFaultWebSocketBlock(event coretypes.ResultEvent) (int64, error) {
	blockEvent, ok := event.Data.(cmttypes.EventDataNewBlock)
	if !ok || blockEvent.Block == nil || blockEvent.Block.Height <= 0 {
		return 0, fmt.Errorf("unexpected WebSocket block event payload %T", event.Data)
	}
	return blockEvent.Block.Height, nil
}

func decodeNetworkFaultWebSocketTransaction(event coretypes.ResultEvent) (networkFaultWebSocketTransaction, error) {
	txEvent, ok := event.Data.(cmttypes.EventDataTx)
	if !ok || txEvent.Height <= 0 || len(txEvent.Tx) == 0 {
		return networkFaultWebSocketTransaction{}, fmt.Errorf("unexpected WebSocket transaction event payload %T", event.Data)
	}
	return networkFaultWebSocketTransaction{
		Height: txEvent.Height,
		TxHash: strings.ToUpper(fmt.Sprintf("%X", cmttypes.Tx(txEvent.Tx).Hash())),
	}, nil
}

func runSlowRESTClient(restAddress string, hold time.Duration) (map[string]any, error) {
	if hold <= 0 {
		return nil, fmt.Errorf("slow-client hold must be positive, got %s", hold)
	}
	parsed, err := url.Parse(restAddress)
	if err != nil {
		return nil, err
	}
	connection, err := net.DialTimeout("tcp", parsed.Host, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	startedAt := time.Now()
	if err := connection.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return nil, err
	}
	if _, err := io.WriteString(connection, "POST /cosmos/tx/v1beta1/txs HTTP/1.1\r\nHost: "+parsed.Host+"\r\nContent-Length: 100\r\n"); err != nil {
		return nil, err
	}
	time.Sleep(hold)
	if err := connection.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return nil, err
	}
	buffer := make([]byte, 1024)
	read, readErr := connection.Read(buffer)
	elapsed := time.Since(startedAt)
	statusCode := 0
	closedWithoutResponse := false
	if read == 0 {
		if readErr == nil {
			return nil, errors.New("slow client read returned neither response nor connection-close error")
		}
		if timeout, ok := readErr.(net.Error); ok && timeout.Timeout() {
			return nil, fmt.Errorf("slow client was not closed by server within %s", elapsed)
		}
		closedWithoutResponse = true
	} else {
		statusCode, statusErr := networkFaultHTTPStatus(buffer[:read])
		if statusErr != nil {
			return nil, fmt.Errorf("slow client received data instead of a timeout close: %w", statusErr)
		}
		if statusCode < http.StatusBadRequest {
			return nil, fmt.Errorf("slow client received successful HTTP response %d", statusCode)
		}
		return nil, fmt.Errorf("slow client received HTTP response %d instead of a timeout connection close", statusCode)
	}
	if elapsed > 4*time.Second {
		return nil, fmt.Errorf("slow client rejection exceeded bound: %s", elapsed)
	}
	return map[string]any{
		"held_headers_for":  hold.String(),
		"elapsed":           elapsed.String(),
		"response":          string(buffer[:read]),
		"read_error":        errorStringForNetworkFault(readErr),
		"status_code":       statusCode,
		"connection_closed": closedWithoutResponse,
		"bounded":           true,
	}, nil
}

func networkFaultHTTPStatus(response []byte) (int, error) {
	lineEnd := bytes.Index(response, []byte("\r\n"))
	if lineEnd < 0 {
		return 0, errors.New("HTTP response has no complete status line")
	}
	fields := strings.Fields(string(response[:lineEnd]))
	if len(fields) < 2 || !strings.HasPrefix(fields[0], "HTTP/") {
		return 0, fmt.Errorf("invalid HTTP status line %q", response[:lineEnd])
	}
	status, err := strconv.Atoi(fields[1])
	if err != nil || status < 100 || status > 599 {
		return 0, fmt.Errorf("invalid HTTP status code %q", fields[1])
	}
	return status, nil
}

func validateOversizedRESTRejection(status int, response []byte, configuredLimit, requestBytes int) error {
	if configuredLimit <= 0 {
		return fmt.Errorf("configured body limit must be positive, got %d", configuredLimit)
	}
	if requestBytes <= configuredLimit {
		return fmt.Errorf("request body %d does not exceed configured limit %d", requestBytes, configuredLimit)
	}
	if status != http.StatusBadRequest && status != http.StatusRequestEntityTooLarge {
		return fmt.Errorf("oversized body status = %d, want 400 or 413", status)
	}
	if !bytes.Contains(bytes.ToLower(response), []byte(networkFaultOversizedBodyMarker)) {
		return fmt.Errorf("oversized body response does not identify %q: %s", networkFaultOversizedBodyMarker, strings.TrimSpace(string(response)))
	}
	return nil
}

func validateOversizedGRPCRejection(grpcErr error, configuredLimit, requestBytes int) error {
	if configuredLimit <= 0 {
		return fmt.Errorf("configured gRPC receive limit must be positive, got %d", configuredLimit)
	}
	if requestBytes <= configuredLimit {
		return fmt.Errorf("gRPC request message %d does not exceed configured limit %d", requestBytes, configuredLimit)
	}
	if grpcErr == nil {
		return errors.New("oversized gRPC request unexpectedly succeeded")
	}
	if status.Code(grpcErr) != codes.ResourceExhausted {
		return fmt.Errorf("oversized gRPC request code = %s, want ResourceExhausted: %w", status.Code(grpcErr), grpcErr)
	}
	diagnostic := strings.ToLower(grpcErr.Error())
	if !strings.Contains(diagnostic, "message larger than max") &&
		!strings.Contains(diagnostic, "received message larger than max") {
		return fmt.Errorf("oversized gRPC rejection does not identify message size: %w", grpcErr)
	}
	return nil
}

func validateNetworkFaultWebSocketReconnect(
	before networkFaultWebSocketPhase,
	after networkFaultWebSocketPhase,
	expectedTransactionHashes []string,
) (networkFaultWebSocketContinuity, error) {
	continuity := networkFaultWebSocketContinuity{Before: before, After: after}
	var validationErrors []error
	var blockValidationErrors []error
	phaseCases := []struct {
		name  string
		phase networkFaultWebSocketPhase
	}{
		{name: "before fault", phase: before},
		{name: "after fault", phase: after},
	}
	for _, phaseCase := range phaseCases {
		phaseName, phase := phaseCase.name, phaseCase.phase
		if err := validateConsecutiveNetworkFaultWebSocketHeights(phaseName, phase.BlockHeights); err != nil {
			blockValidationErrors = append(blockValidationErrors, err)
		}
	}
	blockCounts := make(map[int64]int, len(before.BlockHeights)+len(after.BlockHeights))
	for _, phase := range []networkFaultWebSocketPhase{before, after} {
		for _, height := range phase.BlockHeights {
			blockCounts[height]++
			if blockCounts[height] > 1 {
				continuity.DuplicateBlockEvents++
			}
		}
	}
	if continuity.DuplicateBlockEvents > 0 {
		blockValidationErrors = append(blockValidationErrors, fmt.Errorf(
			"WebSocket block stream contained %d duplicate events",
			continuity.DuplicateBlockEvents,
		))
	}
	if len(before.BlockHeights) > 0 && len(after.BlockHeights) > 0 {
		lastBefore := before.BlockHeights[len(before.BlockHeights)-1]
		firstAfter := after.BlockHeights[0]
		if firstAfter <= lastBefore {
			blockValidationErrors = append(blockValidationErrors, fmt.Errorf(
				"post-fault WebSocket block height %d does not follow pre-fault height %d",
				firstAfter,
				lastBefore,
			))
		} else {
			const maximumRecordedFaultGap = 10_000
			if firstAfter-lastBefore-1 > maximumRecordedFaultGap {
				blockValidationErrors = append(blockValidationErrors, fmt.Errorf(
					"fault-induced WebSocket height gap %d exceeds artifact bound %d",
					firstAfter-lastBefore-1,
					maximumRecordedFaultGap,
				))
			} else {
				for height := lastBefore + 1; height < firstAfter; height++ {
					continuity.FaultMissingBlockHeights = append(continuity.FaultMissingBlockHeights, height)
				}
			}
		}
	}
	validationErrors = append(validationErrors, blockValidationErrors...)

	expected := make(map[string]struct{}, len(expectedTransactionHashes))
	if len(expectedTransactionHashes) == 0 {
		validationErrors = append(validationErrors, errors.New("expected WebSocket transaction hashes are empty"))
	}
	for _, txHash := range expectedTransactionHashes {
		normalized := strings.ToUpper(strings.TrimSpace(txHash))
		if normalized == "" {
			validationErrors = append(validationErrors, errors.New("expected WebSocket transaction hash is empty"))
			continue
		}
		if _, duplicate := expected[normalized]; duplicate {
			validationErrors = append(validationErrors, fmt.Errorf("expected WebSocket transaction hash %s is duplicated", normalized))
			continue
		}
		expected[normalized] = struct{}{}
	}
	observed := make(map[string]int, len(expected))
	for _, phaseCase := range phaseCases {
		phaseName, phase := phaseCase.name, phaseCase.phase
		phaseBlocks := make(map[int64]struct{}, len(phase.BlockHeights))
		for _, height := range phase.BlockHeights {
			phaseBlocks[height] = struct{}{}
		}
		for _, transaction := range phase.Transactions {
			normalized := strings.ToUpper(strings.TrimSpace(transaction.TxHash))
			if normalized == "" || transaction.Height <= 0 {
				validationErrors = append(validationErrors, fmt.Errorf(
					"%s WebSocket transaction is invalid: height=%d hash=%q",
					phaseName,
					transaction.Height,
					transaction.TxHash,
				))
				continue
			}
			if _, matched := phaseBlocks[transaction.Height]; !matched {
				validationErrors = append(validationErrors, fmt.Errorf(
					"%s WebSocket transaction %s at height %d has no matching block event",
					phaseName,
					normalized,
					transaction.Height,
				))
			}
			observed[normalized]++
			if observed[normalized] > 1 {
				continuity.DuplicateTransactionEvents++
			}
			if _, wanted := expected[normalized]; !wanted {
				continuity.UnexpectedTransactionHashes = append(continuity.UnexpectedTransactionHashes, normalized)
			}
		}
	}
	for txHash := range expected {
		if observed[txHash] == 0 {
			continuity.MissingTransactionHashes = append(continuity.MissingTransactionHashes, txHash)
		}
	}
	sort.Strings(continuity.MissingTransactionHashes)
	sort.Strings(continuity.UnexpectedTransactionHashes)
	if continuity.DuplicateTransactionEvents > 0 {
		validationErrors = append(validationErrors, fmt.Errorf(
			"WebSocket transaction stream contained %d duplicate events",
			continuity.DuplicateTransactionEvents,
		))
	}
	if len(continuity.MissingTransactionHashes) > 0 {
		validationErrors = append(validationErrors, fmt.Errorf(
			"WebSocket transaction stream is missing expected hashes %v",
			continuity.MissingTransactionHashes,
		))
	}
	if len(continuity.UnexpectedTransactionHashes) > 0 {
		validationErrors = append(validationErrors, fmt.Errorf(
			"WebSocket transaction stream contained unexpected hashes %v",
			continuity.UnexpectedTransactionHashes,
		))
	}
	continuity.ConsecutiveOutsideFault = len(blockValidationErrors) == 0
	continuity.ExpectedTransactionsObserved = len(continuity.MissingTransactionHashes) == 0 &&
		continuity.DuplicateTransactionEvents == 0 &&
		len(continuity.UnexpectedTransactionHashes) == 0 &&
		len(expected) > 0
	return continuity, errors.Join(validationErrors...)
}

func validateConsecutiveNetworkFaultWebSocketHeights(phase string, heights []int64) error {
	if len(heights) == 0 {
		return fmt.Errorf("%s WebSocket block heights are empty", phase)
	}
	for index, height := range heights {
		if height <= 0 {
			return fmt.Errorf("%s WebSocket block height must be positive, got %d", phase, height)
		}
		if index == 0 {
			continue
		}
		if height != heights[index-1]+1 {
			return fmt.Errorf(
				"%s WebSocket block heights have an event gap or duplicate: %d after %d",
				phase,
				height,
				heights[index-1],
			)
		}
	}
	return nil
}

func startNetworkFaultNodeWithReadinessRecovery(
	ctx context.Context,
	phase string,
	runtime networkFaultReadinessRuntime,
) error {
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return errors.New("network-fault readiness phase is required")
	}
	if runtime.Start == nil || runtime.Stop == nil || runtime.Observe == nil || runtime.Record == nil {
		return errors.New("network-fault readiness runtime is incomplete")
	}

	firstStartCtx, firstStartCancel := context.WithTimeout(ctx, networkFaultFirstReadinessStartTimeout)
	firstStartErr := runtime.Start(firstStartCtx)
	firstStartCancel()
	if firstStartErr == nil {
		return nil
	}
	if ctx.Err() != nil {
		return fmt.Errorf("start node during %s: %w", phase, errors.Join(firstStartErr, ctx.Err()))
	}

	observationCtx, observationCancel := context.WithTimeout(ctx, networkFaultReadinessObservationTimeout)
	firstHeight, firstCatchingUp, observationErr := runtime.Observe(observationCtx)
	observationCancel()
	if observationErr != nil {
		return fmt.Errorf(
			"start node during %s without a readable readiness observation: %w",
			phase,
			errors.Join(firstStartErr, observationErr),
		)
	}
	// ChainNode.StartContainer's readiness loop only checks Comet RPC and
	// catching_up. The parent context can win the race immediately before a
	// successful retry, so an independently observed caught-up node satisfies
	// the same contract without an unnecessary restart.
	if !firstCatchingUp {
		return nil
	}

	evidence := networkFaultReadinessRecovery{
		Phase:                    phase,
		FirstStartTimeout:        networkFaultFirstReadinessStartTimeout.String(),
		FirstStartError:          firstStartErr.Error(),
		FirstStartHeight:         firstHeight,
		FirstStartCatchingUp:     true,
		RecoveryRestartAttempted: true,
		RecordedAt:               time.Now().UTC(),
	}
	stopErr := runtime.Stop(ctx)
	var recoveryStartErr error
	if stopErr == nil {
		recoveryStartErr = runtime.Start(ctx)
	}
	recoveryErr := errors.Join(stopErr, recoveryStartErr)
	evidence.RecoveryRestartSucceeded = recoveryErr == nil
	if recoveryErr != nil {
		evidence.RecoveryRestartError = recoveryErr.Error()
	}
	recordErr := runtime.Record(evidence)
	if recoveryErr != nil {
		return fmt.Errorf(
			"start node during %s after one bounded readiness restart: %w",
			phase,
			errors.Join(firstStartErr, recoveryErr, recordErr),
		)
	}
	if recordErr != nil {
		return fmt.Errorf("record readiness restart during %s: %w", phase, recordErr)
	}
	return nil
}

func restartFullNodeWithAppConfig(
	ctx context.Context,
	network *harness.Network,
	phase string,
	validator *cosmos.ChainNode,
	node *cosmos.ChainNode,
	contents []byte,
) error {
	if network == nil || validator == nil || node == nil {
		return errors.New("network, validator, and full node are required")
	}
	phase = strings.TrimSpace(phase)
	if phase == "" {
		return errors.New("full-node app-config restart phase is required")
	}
	stopErr := node.StopContainer(ctx)
	if stopErr != nil {
		return stopErr
	}
	if err := node.WriteFile(ctx, contents, "config/app.toml"); err != nil {
		return errors.Join(err, node.StartContainer(ctx))
	}
	runtime := networkFaultReadinessRuntime{
		Start: node.StartContainer,
		Stop:  node.StopContainer,
		Observe: func(observationCtx context.Context) (int64, bool, error) {
			if node.Client == nil {
				return 0, false, errors.New("full-node Comet RPC client is nil")
			}
			status, statusErr := node.Client.Status(observationCtx)
			if statusErr != nil {
				return 0, false, statusErr
			}
			if status == nil {
				return 0, false, errors.New("full-node Comet RPC status is nil")
			}
			return status.SyncInfo.LatestBlockHeight, status.SyncInfo.CatchingUp, nil
		},
		Record: func(evidence networkFaultReadinessRecovery) error {
			return network.AppendArtifactJSON("network-faults/readiness-recoveries.jsonl", evidence)
		},
	}
	if err := startNetworkFaultNodeWithReadinessRecovery(ctx, phase, runtime); err != nil {
		return err
	}
	targetHeight, err := validator.Height(ctx)
	if err != nil {
		return fmt.Errorf("query validator height after %s restart: %w", phase, err)
	}
	_, err = network.WaitForQuorumAgreement(
		ctx,
		"network-fault-"+phase,
		targetHeight,
		validator,
		node,
	)
	if err != nil {
		return fmt.Errorf("verify peer-aware agreement after %s restart: %w", phase, err)
	}
	return nil
}

func recordNetworkFaultCleanup(
	t *testing.T,
	network *harness.Network,
	phase string,
	cleanupErr error,
) {
	t.Helper()
	if err := network.RecordNetworkFaultCleanup(phase, cleanupErr); err != nil {
		t.Errorf("network fault cleanup %s: %v", phase, err)
	}
}

func recordChainP2PRuntimeFault(
	network *harness.Network,
	phase string,
	outcome harness.NetworkFaultOutcome,
) error {
	return network.RecordNetworkFaultCategory(harness.NetworkFaultCategoryEvidence{
		Category: harness.NetworkFaultCategoryChainP2PRuntime,
		Phase:    phase,
		Outcome:  outcome,
		Scope:    harness.NetworkFaultScopeRunOwnedDockerP2P,
	})
}

func rewriteNetworkFaultAppConfig(
	contents []byte,
	mutation networkFaultEndpointMutation,
) ([]byte, error) {
	var document map[string]any
	if _, err := toml.Decode(string(contents), &document); err != nil {
		return nil, err
	}
	apiValue, ok := document["api"]
	if !ok {
		return nil, errors.New("app.toml has no api section")
	}
	api, ok := apiValue.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("app.toml api section has type %T", apiValue)
	}
	grpcValue, ok := document["grpc"]
	if !ok {
		return nil, errors.New("app.toml has no grpc section")
	}
	grpc, ok := grpcValue.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("app.toml grpc section has type %T", grpcValue)
	}
	if mutation.APIEnabled != nil {
		api["enable"] = *mutation.APIEnabled
	}
	if mutation.GRPCEnabled != nil {
		grpc["enable"] = *mutation.GRPCEnabled
	}
	if mutation.GRPCMaxRecvBytes != nil {
		if *mutation.GRPCMaxRecvBytes == 0 {
			return nil, errors.New("gRPC max receive message bytes must be positive")
		}
		grpc["max-recv-msg-size"] = strconv.FormatUint(uint64(*mutation.GRPCMaxRecvBytes), 10)
	}
	if mutation.ReadTimeout != nil {
		api["rpc-read-timeout"] = int64(*mutation.ReadTimeout)
	}
	if mutation.WriteTimeout != nil {
		api["rpc-write-timeout"] = int64(*mutation.WriteTimeout)
	}
	if mutation.MaxBodyBytes != nil {
		api["rpc-max-body-bytes"] = int64(*mutation.MaxBodyBytes)
	}
	document["api"] = api
	document["grpc"] = grpc
	var output bytes.Buffer
	if err := toml.NewEncoder(&output).Encode(document); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func boolPointer(value bool) *bool {
	return &value
}

func uintPointer(value uint) *uint {
	return &value
}

func errorStringForNetworkFault(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
