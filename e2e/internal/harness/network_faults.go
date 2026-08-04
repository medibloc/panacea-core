package harness

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/dockerutil"
)

const (
	networkFaultProxyBinary   = "/usr/bin/panacea-e2e-faultproxy"
	networkFaultProxyLogLimit = 4 << 20
	// P2PFaultProxyTargetPort is deliberately different from CometBFT's
	// default port. Moving the validator listener while proxying makes stale
	// addrbook entries to :26656 fail instead of silently bypassing the proxy.
	P2PFaultProxyTargetPort = 27656
)

var networkFaultNamePattern = regexp.MustCompile("^[a-z0-9][a-z0-9-]{0,40}$")

// P2PFaultProxyConfig describes a run-owned, container-scoped TCP impairment
// proxy. DropEvery closes the stream at each Nth chunk, which creates bounded
// effective loss without touching the host firewall or network namespace.
type P2PFaultProxyConfig struct {
	Name          string        `json:"name"`
	Alias         string        `json:"alias"`
	Image         ImageRef      `json:"image"`
	TargetAddress string        `json:"target_address"`
	IPv4Address   string        `json:"ipv4_address,omitempty"`
	Delay         time.Duration `json:"delay_nanoseconds"`
	Jitter        time.Duration `json:"jitter_nanoseconds"`
	DropEvery     uint64        `json:"drop_every"`
	Seed          int64         `json:"seed"`
}

func (c P2PFaultProxyConfig) Validate() error {
	var validationErrors []error
	if !networkFaultNamePattern.MatchString(c.Name) {
		validationErrors = append(validationErrors, fmt.Errorf("fault proxy name %q must match %s", c.Name, networkFaultNamePattern))
	}
	if !networkFaultNamePattern.MatchString(c.Alias) {
		validationErrors = append(validationErrors, fmt.Errorf("fault proxy alias %q must match %s", c.Alias, networkFaultNamePattern))
	}
	if strings.TrimSpace(c.Image.Repository) == "" || strings.TrimSpace(c.Image.Version) == "" {
		validationErrors = append(validationErrors, errors.New("fault proxy image repository and version are required"))
	}
	if strings.TrimSpace(c.TargetAddress) == "" {
		validationErrors = append(validationErrors, errors.New("fault proxy target address is required"))
	}
	if c.IPv4Address != "" {
		parsed := net.ParseIP(strings.TrimSpace(c.IPv4Address))
		if parsed == nil || parsed.To4() == nil {
			validationErrors = append(validationErrors, fmt.Errorf("fault proxy IPv4 address %q is invalid", c.IPv4Address))
		}
	}
	if c.Delay < 0 || c.Jitter < 0 {
		validationErrors = append(validationErrors, errors.New("fault proxy delay and jitter cannot be negative"))
	}
	if c.Seed == 0 {
		validationErrors = append(validationErrors, errors.New("fault proxy seed must be non-zero"))
	}
	return errors.Join(validationErrors...)
}

// P2PFaultProxy identifies one running proxy without exposing a Docker client.
type P2PFaultProxy struct {
	Config      P2PFaultProxyConfig `json:"config"`
	ContainerID string              `json:"container_id"`
	Name        string              `json:"container_name"`
	IPAddress   string              `json:"ip_address"`
	StartedAt   time.Time           `json:"started_at"`
}

type DockerNetworkFaultEvidence struct {
	Phase            string    `json:"phase"`
	Action           string    `json:"action"`
	Node             string    `json:"node"`
	ContainerID      string    `json:"container_id"`
	NetworkID        string    `json:"network_id"`
	BeforeIPAddress  string    `json:"before_ip_address,omitempty"`
	AfterIPAddress   string    `json:"after_ip_address,omitempty"`
	AttachedBefore   bool      `json:"attached_before"`
	AttachedAfter    bool      `json:"attached_after"`
	TransitionProved bool      `json:"transition_proved"`
	RecordedAt       time.Time `json:"recorded_at"`
	Error            string    `json:"error,omitempty"`
}

// StartP2PFaultProxy starts a proxy on the same run-owned Docker network as
// the chain. Cleanup remains label-driven and therefore panic safe.
func (n *Network) StartP2PFaultProxy(ctx context.Context, config P2PFaultProxyConfig) (P2PFaultProxy, error) {
	if n == nil || n.artifacts == nil || n.artifacts.client == nil {
		return P2PFaultProxy{}, errors.New("network fault Docker client is unavailable")
	}
	if err := config.Validate(); err != nil {
		return P2PFaultProxy{}, err
	}
	if strings.TrimSpace(n.artifacts.networkID) == "" {
		return P2PFaultProxy{}, errors.New("network fault Docker network ID is unavailable")
	}
	containerName := "panacea-e2e-fault-" + n.artifacts.runID + "-" + config.Name
	command := []string{
		"--listen", ":26656",
		"--target", config.TargetAddress,
		"--delay", config.Delay.String(),
		"--jitter", config.Jitter.String(),
		"--drop-every", strconv.FormatUint(config.DropEvery, 10),
		"--seed", strconv.FormatInt(config.Seed, 10),
	}
	endpoint := &dockernetwork.EndpointSettings{Aliases: []string{config.Alias}}
	if config.IPv4Address != "" {
		endpoint.IPAMConfig = &dockernetwork.EndpointIPAMConfig{IPv4Address: strings.TrimSpace(config.IPv4Address)}
	}
	created, err := n.artifacts.client.ContainerCreate(
		ctx,
		&dockercontainer.Config{
			Image:      config.Image.Repository + ":" + config.Image.Version,
			Entrypoint: []string{networkFaultProxyBinary},
			Cmd:        command,
			Labels: map[string]string{
				dockerutil.CleanupLabel:     n.artifacts.runID,
				"panacea.e2e.network-fault": config.Name,
			},
		},
		&dockercontainer.HostConfig{},
		&dockernetwork.NetworkingConfig{EndpointsConfig: map[string]*dockernetwork.EndpointSettings{
			n.artifacts.networkID: endpoint,
		}},
		nil,
		containerName,
	)
	if err != nil {
		return P2PFaultProxy{}, fmt.Errorf("create P2P fault proxy %s: %w", config.Name, err)
	}
	proxy := P2PFaultProxy{
		Config:      config,
		ContainerID: created.ID,
		Name:        containerName,
		StartedAt:   time.Now().UTC(),
	}
	if err := n.artifacts.client.ContainerStart(ctx, created.ID, dockertypes.ContainerStartOptions{}); err != nil {
		return proxy, fmt.Errorf("start P2P fault proxy %s: %w", config.Name, err)
	}
	if err := n.waitForP2PFaultProxy(ctx, proxy); err != nil {
		return proxy, err
	}
	proxy.IPAddress, err = n.RunNetworkIPAddress(ctx, proxy.ContainerID)
	if err != nil {
		return proxy, fmt.Errorf("inspect P2P fault proxy %s run-network address: %w", config.Name, err)
	}
	if config.IPv4Address != "" && proxy.IPAddress != strings.TrimSpace(config.IPv4Address) {
		return proxy, fmt.Errorf(
			"P2P fault proxy %s address = %s, want reserved %s",
			config.Name,
			proxy.IPAddress,
			config.IPv4Address,
		)
	}
	if err := n.artifacts.appendJSONLine("network-faults/proxies.jsonl", map[string]any{
		"action": "started", "proxy": proxy,
	}); err != nil {
		return proxy, fmt.Errorf("record P2P fault proxy start: %w", err)
	}
	return proxy, nil
}

func (n *Network) waitForP2PFaultProxy(ctx context.Context, proxy P2PFaultProxy) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		logs, err := n.P2PFaultProxyLogs(ctx, proxy)
		if err == nil && bytes.Contains(logs, []byte("\"event\":\"listening\"")) {
			return nil
		}
		inspect, inspectErr := n.artifacts.client.ContainerInspect(ctx, proxy.ContainerID)
		if inspectErr == nil && inspect.State != nil && !inspect.State.Running {
			return fmt.Errorf("P2P fault proxy %s exited before readiness: %s", proxy.Config.Name, strings.TrimSpace(string(logs)))
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for P2P fault proxy %s readiness: %w", proxy.Config.Name, ctx.Err())
		case <-ticker.C:
		}
	}
}

// P2PFaultProxyLogs returns Docker-demultiplexed stdout and stderr with a
// strict diagnostic bound.
func (n *Network) P2PFaultProxyLogs(ctx context.Context, proxy P2PFaultProxy) (logs []byte, retErr error) {
	if n == nil || n.artifacts == nil || n.artifacts.client == nil {
		return nil, errors.New("network fault Docker client is unavailable")
	}
	reader, err := n.artifacts.client.ContainerLogs(ctx, proxy.ContainerID, dockertypes.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return nil, fmt.Errorf("read P2P fault proxy %s logs: %w", proxy.Config.Name, err)
	}
	defer func() {
		retErr = errors.Join(retErr, reader.Close())
	}()
	var stdout, stderr bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdout, &stderr, io.LimitReader(reader, networkFaultProxyLogLimit+1)); err != nil {
		return nil, fmt.Errorf("decode P2P fault proxy %s logs: %w", proxy.Config.Name, err)
	}
	logs = append(append([]byte(nil), stdout.Bytes()...), stderr.Bytes()...)
	if len(logs) > networkFaultProxyLogLimit {
		return nil, fmt.Errorf("P2P fault proxy %s logs exceed %d bytes", proxy.Config.Name, networkFaultProxyLogLimit)
	}
	return logs, nil
}

// StopP2PFaultProxy captures final logs, stops and removes the exact run-owned
// container, and records cleanup evidence.
func (n *Network) StopP2PFaultProxy(ctx context.Context, phase string, proxy P2PFaultProxy) error {
	if n == nil || n.artifacts == nil || n.artifacts.client == nil {
		return errors.New("network fault Docker client is unavailable")
	}
	if !networkFaultNamePattern.MatchString(phase) {
		return fmt.Errorf("network fault phase %q must match %s", phase, networkFaultNamePattern)
	}
	logs, logsErr := n.P2PFaultProxyLogs(ctx, proxy)
	if len(logs) > 0 {
		logsErr = errors.Join(logsErr, n.artifacts.write("network-faults/proxy-logs/"+phase+".jsonl", logs))
	}
	timeout := 5
	stopErr := n.artifacts.client.ContainerStop(ctx, proxy.ContainerID, dockercontainer.StopOptions{Timeout: &timeout})
	removeErr := n.artifacts.client.ContainerRemove(ctx, proxy.ContainerID, dockertypes.ContainerRemoveOptions{Force: true})
	recordErr := n.artifacts.appendJSONLine("network-faults/proxies.jsonl", map[string]any{
		"action": "stopped-and-removed", "phase": phase, "proxy": proxy,
		"recorded_at": time.Now().UTC(), "logs_error": errorString(logsErr),
		"stop_error": errorString(stopErr), "remove_error": errorString(removeErr),
	})
	return errors.Join(logsErr, stopErr, removeErr, recordErr)
}

// DisconnectNodeFromRunNetwork applies a complete container-scoped partition.
func (n *Network) DisconnectNodeFromRunNetwork(ctx context.Context, phase string, node *cosmos.ChainNode) error {
	return n.changeNodeRunNetwork(ctx, phase, node, "disconnect")
}

// ReconnectNodeToRunNetwork restores the exact run-owned bridge attachment.
func (n *Network) ReconnectNodeToRunNetwork(ctx context.Context, phase string, node *cosmos.ChainNode) error {
	return n.changeNodeRunNetwork(ctx, phase, node, "reconnect")
}

func (n *Network) changeNodeRunNetwork(ctx context.Context, phase string, node *cosmos.ChainNode, action string) error {
	if n == nil || n.artifacts == nil || n.artifacts.client == nil {
		return errors.New("network fault Docker client is unavailable")
	}
	if !networkFaultNamePattern.MatchString(phase) {
		return fmt.Errorf("network fault phase %q must match %s", phase, networkFaultNamePattern)
	}
	if node == nil || strings.TrimSpace(node.ContainerID()) == "" {
		return errors.New("network fault node container is required")
	}
	if action != "disconnect" && action != "reconnect" {
		return fmt.Errorf("unsupported run-network action %q", action)
	}
	beforeIP, attachedBefore, beforeErr := n.inspectNodeRunNetworkAttachment(ctx, node.ContainerID())
	if beforeErr != nil {
		return fmt.Errorf("inspect node %s before run-network %s: %w", node.Name(), action, beforeErr)
	}
	var operationErr error
	switch action {
	case "disconnect":
		operationErr = n.artifacts.client.NetworkDisconnect(ctx, n.artifacts.networkID, node.ContainerID(), false)
	case "reconnect":
		operationErr = n.artifacts.client.NetworkConnect(ctx, n.artifacts.networkID, node.ContainerID(), nil)
	}
	afterIP, attachedAfter, inspectAfterErr := n.inspectNodeRunNetworkAttachment(ctx, node.ContainerID())
	transitionErr := validateNetworkFaultAttachmentTransition(
		action,
		attachedBefore,
		attachedAfter,
		beforeIP,
		afterIP,
	)
	combinedOperationErr := errors.Join(operationErr, inspectAfterErr, transitionErr)
	evidence := DockerNetworkFaultEvidence{
		Phase:            phase,
		Action:           action,
		Node:             node.Name(),
		ContainerID:      node.ContainerID(),
		NetworkID:        n.artifacts.networkID,
		BeforeIPAddress:  beforeIP,
		AfterIPAddress:   afterIP,
		AttachedBefore:   attachedBefore,
		AttachedAfter:    attachedAfter,
		TransitionProved: combinedOperationErr == nil,
		RecordedAt:       time.Now().UTC(),
		Error:            errorString(combinedOperationErr),
	}
	recordErr := n.artifacts.appendJSONLine("network-faults/docker-network.jsonl", evidence)
	if combinedOperationErr != nil {
		combinedOperationErr = fmt.Errorf("%s node %s from run network: %w", action, node.Name(), combinedOperationErr)
	}
	return errors.Join(combinedOperationErr, recordErr)
}

func (n *Network) inspectNodeRunNetworkAttachment(
	ctx context.Context,
	containerID string,
) (string, bool, error) {
	inspect, err := n.artifacts.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", false, err
	}
	if inspect.NetworkSettings == nil {
		return "", false, errors.New("Docker inspect returned no network settings")
	}
	endpoint, attached := resolveNetworkFaultEndpoint(inspect.NetworkSettings.Networks, n.artifacts.networkID)
	if !attached {
		return "", false, nil
	}
	ipAddress := strings.TrimSpace(endpoint.IPAddress)
	if ipAddress == "" {
		return "", true, fmt.Errorf("run-network endpoint %s has no IPv4 address", n.artifacts.networkID)
	}
	return ipAddress, true, nil
}

// RunNetworkIPAddress returns the exact IPv4 address assigned to a container
// on this test's run-owned bridge. It is intentionally scoped to the bridge so
// callers cannot confuse a host/published address with the peer DNS target.
func (n *Network) RunNetworkIPAddress(ctx context.Context, containerID string) (string, error) {
	if n == nil || n.artifacts == nil || n.artifacts.client == nil {
		return "", errors.New("network fault Docker client is unavailable")
	}
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return "", errors.New("run-network container ID is required")
	}
	ipAddress, attached, err := n.inspectNodeRunNetworkAttachment(ctx, containerID)
	if err != nil {
		return "", err
	}
	if !attached {
		return "", fmt.Errorf("container %s is not attached to run network %s", containerID, n.artifacts.networkID)
	}
	return ipAddress, nil
}

func resolveNetworkFaultEndpoint(
	networks map[string]*dockernetwork.EndpointSettings,
	networkIDOrName string,
) (*dockernetwork.EndpointSettings, bool) {
	wanted := strings.TrimSpace(networkIDOrName)
	if wanted == "" {
		return nil, false
	}
	if endpoint := networks[wanted]; endpoint != nil {
		return endpoint, true
	}
	for name, endpoint := range networks {
		if endpoint == nil {
			continue
		}
		if strings.TrimSpace(name) == wanted || strings.TrimSpace(endpoint.NetworkID) == wanted {
			return endpoint, true
		}
	}
	return nil, false
}

func validateNetworkFaultAttachmentTransition(
	action string,
	attachedBefore bool,
	attachedAfter bool,
	beforeIP string,
	afterIP string,
) error {
	switch action {
	case "disconnect":
		if !attachedBefore || strings.TrimSpace(beforeIP) == "" {
			return errors.New("disconnect did not begin from an attached endpoint with a non-empty IP address")
		}
		if attachedAfter {
			return errors.New("disconnect left the container attached to the run network")
		}
	case "reconnect":
		if attachedBefore {
			return errors.New("reconnect did not begin from a detached endpoint")
		}
		if !attachedAfter || strings.TrimSpace(afterIP) == "" {
			return errors.New("reconnect did not produce an attached endpoint with a non-empty IP address")
		}
	default:
		return fmt.Errorf("unsupported run-network action %q", action)
	}
	return nil
}
