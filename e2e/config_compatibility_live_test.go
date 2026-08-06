package e2e_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/gogoproto/proto"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const configCompatibilityTimeout = 12 * time.Minute

const (
	configCompatibilityGRPCWebMethod       = "/cosmos.base.tendermint.v1beta1.Service/GetNodeInfo"
	configCompatibilityGRPCWebMaxBodyBytes = 1 << 20
	configCompatibilityGRPCWebAppName      = "panacead"
)

type configCompatibilityCommandEvidence struct {
	RecordedAt time.Time `json:"recorded_at"`
	Name       string    `json:"name"`
	Arguments  []string  `json:"arguments"`
	Stdout     string    `json:"stdout,omitempty"`
	Stderr     string    `json:"stderr,omitempty"`
	Error      string    `json:"error,omitempty"`
	Expected   string    `json:"expected"`
}

type configCompatibilityGRPCWebEvidence struct {
	RecordedAt         time.Time         `json:"recorded_at"`
	URL                string            `json:"url"`
	HTTPStatus         int               `json:"http_status"`
	ContentType        string            `json:"content_type"`
	FrameCount         int               `json:"frame_count"`
	Trailers           map[string]string `json:"trailers"`
	Network            string            `json:"network"`
	ApplicationName    string            `json:"application_name"`
	ApplicationVersion string            `json:"application_version"`
	Error              string            `json:"error,omitempty"`
}

// TestV047NodeHomeConfigCompatibility performs a real v2.2.1 -> current
// software upgrade while retaining the old full-node volume. It separately
// proves (1) the current binary starts without rewriting the legacy files and
// (2) the operator-invoked v0.50 confix migration preserves bounded endpoint,
// database, mempool, indexer, and listener overrides.
func TestV047NodeHomeConfigCompatibility(t *testing.T) {
	if os.Getenv("PANACEA_E2E_CONFIG_COMPAT") != "1" {
		t.Skip("set PANACEA_E2E_CONFIG_COMPAT=1 or use ./scripts/e2e/run.sh config-compat")
	}

	ctx, cancel := context.WithTimeout(context.Background(), configCompatibilityTimeout)
	defer cancel()

	network, err := harness.Start(ctx, t, harness.Config{
		Image:         harness.V221Image(),
		NumValidators: 1,
		NumFullNodes:  1,
		TimeoutCommit: "1s",
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.Len(t, network.Chain.Validators, 1)
	require.Len(t, network.Chain.FullNodes, 1)
	toolchainEvidence, err := network.CaptureAndRecordGoToolchainEvidence(ctx)
	require.NoError(t, err)
	require.Equal(t, "go1.26.5", toolchainEvidence.GOVersion)

	fullNode := network.Chain.FullNodes[0]
	generated, err := readConfigCompatibilityNodeHome(ctx, fullNode)
	require.NoError(t, err)
	require.NoError(t, recordConfigCompatibilitySnapshot(network, "generated-v221", generated))

	prepared, err := harness.PrepareV047NodeHomeConfig(
		generated.App,
		generated.Client,
		generated.Comet,
		network.Chain.Config().ChainID,
	)
	require.NoError(t, err)
	beforeWriteHeight, err := fullNode.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, fullNode.StopContainer(ctx))
	require.NoError(t, writeConfigCompatibilityNodeHome(ctx, fullNode, prepared))
	require.NoError(t, fullNode.StartContainer(ctx))
	require.NoError(t, network.WaitForNodeHeight(ctx, fullNode, beforeWriteHeight+1))
	preparedFromDisk, err := readConfigCompatibilityNodeHome(ctx, fullNode)
	require.NoError(t, err)
	require.NoError(t, harness.ValidatePreservedV047NodeHome(prepared, preparedFromDisk))
	require.NoError(t, recordConfigCompatibilitySnapshot(network, "prepared-v221", preparedFromDisk))

	proposer := buildAndFundNFTWallet(t, ctx, network, "config-compat-upgrade-proposer")
	upgradeHeight := scheduleConfigCompatibilityUpgrade(t, ctx, network, proposer.KeyName())
	haltEvidence, err := waitForOldBinaryUpgradeHalt(ctx, network, upgradeHeight)
	require.NoError(t, network.WriteArtifactJSON("config-compat/old-binary-halt.json", haltEvidence))
	require.NoError(t, err)

	allNodes := []*cosmos.ChainNode{network.Chain.Validators[0], fullNode}
	switches, err := network.SwitchNodeImagesTogether(ctx, "config-compat-current", allNodes, harness.CurrentImage())
	require.NoError(t, err)
	require.Len(t, switches, len(allNodes))
	require.NoError(t, network.WriteArtifactJSON("config-compat/image-switches.json", switches))
	for _, node := range allNodes {
		require.NoError(t, network.WaitForNodeHeight(ctx, node, upgradeHeight+3))
		stdout, stderr, versionErr := node.ExecBin(ctx, "version", "--long")
		require.NoError(t, versionErr, "%s", stderr)
		require.Contains(t, string(stdout), "version: 2.3.0")
		require.NoError(t, network.WriteArtifact("config-compat/versions/"+node.Name()+".txt", stdout))
	}
	_, err = network.RequireSameHistoryAtHeight(ctx, upgradeHeight+2, allNodes...)
	require.NoError(t, err)

	postStart, err := readConfigCompatibilityNodeHome(ctx, fullNode)
	require.NoError(t, err)
	require.NoError(t, harness.ValidatePreservedV047NodeHome(preparedFromDisk, postStart))
	require.NoError(t, recordConfigCompatibilitySnapshot(network, "current-started-unmigrated", postStart))
	unmigratedGRPCWeb := probeConfigCompatibilityGRPCWeb(
		ctx,
		network,
		"current-started-unmigrated",
		network.Chain.Config().ChainID,
	)
	require.NoError(t, network.WriteArtifactJSON("config-compat/grpc-web/current-started-unmigrated.json", unmigratedGRPCWeb))
	require.Empty(t, unmigratedGRPCWeb.Error)

	viewApp := runConfigCompatibilityCommand(ctx, fullNode, "view-old-app", "success", "config", "view", "app")
	require.NoError(t, recordConfigCompatibilityCommand(network, viewApp))
	require.Empty(t, viewApp.Error)
	require.NotEmpty(t, strings.TrimSpace(viewApp.Stdout))

	diffBefore := runConfigCompatibilityCommand(ctx, fullNode, "diff-v050-before", "success with migration diff", "config", "diff", "v0.50")
	require.NoError(t, recordConfigCompatibilityCommand(network, diffBefore))
	require.Empty(t, diffBefore.Error)
	diffDiagnostics := diffBefore.Stdout + diffBefore.Stderr
	require.Contains(t, diffDiagnostics, "query-gas-limit=7654321")
	require.Contains(t, diffDiagnostics, "api.rpc-write-timeout=17")

	preview := runConfigCompatibilityCommand(ctx, fullNode, "migrate-v050-preview", "success without file mutation", "config", "migrate", "v0.50", "--stdout")
	require.NoError(t, recordConfigCompatibilityCommand(network, preview))
	require.Empty(t, preview.Error)
	require.NotEmpty(t, strings.TrimSpace(preview.Stdout))
	require.NoError(t, network.WriteArtifact("config-compat/migration/app.toml.v050.preview", []byte(preview.Stdout)))
	previewSnapshot := harness.NewNodeHomeConfigSnapshot([]byte(preview.Stdout), postStart.Client, postStart.Comet)
	require.NoError(t, harness.ValidateMigratedV050NodeHome(postStart, previewSnapshot))

	unchangedAfterPreview, err := readConfigCompatibilityNodeHome(ctx, fullNode)
	require.NoError(t, err)
	require.NoError(t, harness.ValidatePreservedV047NodeHome(postStart, unchangedAfterPreview))

	apply := runConfigCompatibilityCommand(ctx, fullNode, "migrate-v050-apply", "success with in-place app.toml update", "config", "migrate", "v0.50", "--verbose")
	require.NoError(t, recordConfigCompatibilityCommand(network, apply))
	require.Empty(t, apply.Error)
	migrated, err := readConfigCompatibilityNodeHome(ctx, fullNode)
	require.NoError(t, err)
	require.NoError(t, harness.ValidateMigratedV050NodeHome(postStart, migrated))
	require.NoError(t, recordConfigCompatibilitySnapshot(network, "migrated-v050", migrated))

	diffAfter := runConfigCompatibilityCommand(ctx, fullNode, "diff-v050-after", "success with only intentional override differences", "config", "diff", "v0.50")
	require.NoError(t, recordConfigCompatibilityCommand(network, diffAfter))
	require.Empty(t, diffAfter.Error)

	requireConfigCompatibilityCommandContracts(t, ctx, network, fullNode)

	restart, err := network.GracefulRestartNode(ctx, fullNode)
	require.NoError(t, err)
	require.Greater(t, restart.After.Height, restart.Before.Height)
	afterRestart, err := readConfigCompatibilityNodeHome(ctx, fullNode)
	require.NoError(t, err)
	require.NoError(t, harness.ValidatePreservedV047NodeHome(migrated, afterRestart))
	require.NoError(t, recordConfigCompatibilitySnapshot(network, "migrated-v050-after-restart", afterRestart))

	restAddress, err := network.FullNodeHostAddress(ctx, "1317/tcp")
	require.NoError(t, err)
	require.NoError(t, harness.RequireRESTNodeInfo(ctx, &http.Client{Timeout: 5 * time.Second}, restAddress))
	migratedGRPCWeb := probeConfigCompatibilityGRPCWeb(
		ctx,
		network,
		"migrated-v050-after-restart",
		network.Chain.Config().ChainID,
	)
	require.NoError(t, network.WriteArtifactJSON("config-compat/grpc-web/migrated-v050-after-restart.json", migratedGRPCWeb))
	require.Empty(t, migratedGRPCWeb.Error)
	balance, err := network.QueryFullNodeBalance(ctx, proposer.FormattedAddress(), "umed")
	require.NoError(t, err)
	require.True(t, balance.IsPositive())
	require.NoError(t, network.WriteArtifactJSON("config-compat/result.json", map[string]any{
		"upgrade_height":        upgradeHeight,
		"current_started":       true,
		"dry_run_validated":     true,
		"migration_applied":     true,
		"restart_validated":     true,
		"rest_validated":        true,
		"grpc_validated":        true,
		"grpc_web_unmigrated":   unmigratedGRPCWeb.Error == "",
		"grpc_web_migrated":     migratedGRPCWeb.Error == "",
		"proposer_balance_umed": balance.String(),
	}))
}

func probeConfigCompatibilityGRPCWeb(
	ctx context.Context,
	network *harness.Network,
	phase string,
	expectedNetwork string,
) configCompatibilityGRPCWebEvidence {
	evidence := configCompatibilityGRPCWebEvidence{
		RecordedAt: time.Now().UTC(),
		Trailers:   map[string]string{},
	}
	fail := func(err error) configCompatibilityGRPCWebEvidence {
		evidence.Error = err.Error()
		return evidence
	}
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 {
		return fail(errors.New("gRPC-Web probe requires a full node"))
	}
	address, err := network.FullNodeHostAddress(ctx, "1317/tcp")
	if err != nil {
		return fail(fmt.Errorf("resolve %s gRPC-Web API address: %w", phase, err))
	}
	evidence.URL = strings.TrimRight(address, "/") + configCompatibilityGRPCWebMethod
	requestBody := []byte{0, 0, 0, 0, 0} // one uncompressed, empty protobuf request frame
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, evidence.URL, bytes.NewReader(requestBody))
	if err != nil {
		return fail(fmt.Errorf("build %s gRPC-Web request: %w", phase, err))
	}
	request.Header.Set("Content-Type", "application/grpc-web+proto")
	request.Header.Set("Accept", "application/grpc-web+proto")
	request.Header.Set("X-Grpc-Web", "1")
	request.Header.Set("TE", "trailers")
	response, err := (&http.Client{Timeout: 5 * time.Second}).Do(request)
	if err != nil {
		return fail(fmt.Errorf("call %s gRPC-Web endpoint: %w", phase, err))
	}
	defer response.Body.Close()
	evidence.HTTPStatus = response.StatusCode
	evidence.ContentType = response.Header.Get("Content-Type")
	body, err := io.ReadAll(io.LimitReader(response.Body, configCompatibilityGRPCWebMaxBodyBytes+1))
	if err != nil {
		return fail(fmt.Errorf("read %s gRPC-Web response: %w", phase, err))
	}
	if len(body) > configCompatibilityGRPCWebMaxBodyBytes {
		return fail(fmt.Errorf("%s gRPC-Web response exceeds %d bytes", phase, configCompatibilityGRPCWebMaxBodyBytes))
	}
	if response.StatusCode != http.StatusOK {
		return fail(fmt.Errorf("%s gRPC-Web HTTP status = %d, want 200", phase, response.StatusCode))
	}
	if err := validateConfigCompatibilityGRPCWebContentType(evidence.ContentType); err != nil {
		return fail(fmt.Errorf("%s gRPC-Web content type: %w", phase, err))
	}
	payload, trailers, frameCount, err := decodeConfigCompatibilityGRPCWebUnaryResponse(body)
	evidence.FrameCount = frameCount
	evidence.Trailers = trailers
	if err != nil {
		return fail(fmt.Errorf("decode %s gRPC-Web response: %w", phase, err))
	}
	var nodeInfo cmtservice.GetNodeInfoResponse
	if err := proto.Unmarshal(payload, &nodeInfo); err != nil {
		return fail(fmt.Errorf("decode %s gRPC-Web node info protobuf: %w", phase, err))
	}
	if nodeInfo.GetDefaultNodeInfo() == nil || nodeInfo.GetApplicationVersion() == nil {
		return fail(fmt.Errorf("%s gRPC-Web node info is incomplete", phase))
	}
	evidence.Network, evidence.ApplicationName, evidence.ApplicationVersion, err =
		validateConfigCompatibilityGRPCWebNodeInfo(&nodeInfo, expectedNetwork)
	if err != nil {
		return fail(fmt.Errorf("%s gRPC-Web node info: %w", phase, err))
	}
	return evidence
}

func validateConfigCompatibilityGRPCWebContentType(contentType string) error {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("parse %q: %w", contentType, err)
	}
	switch strings.ToLower(mediaType) {
	case "application/grpc-web", "application/grpc-web+proto":
		return nil
	default:
		return fmt.Errorf("got %q, want binary application/grpc-web or application/grpc-web+proto", mediaType)
	}
}

func validateConfigCompatibilityGRPCWebNodeInfo(
	nodeInfo *cmtservice.GetNodeInfoResponse,
	expectedNetwork string,
) (string, string, string, error) {
	if nodeInfo == nil || nodeInfo.GetDefaultNodeInfo() == nil || nodeInfo.GetApplicationVersion() == nil {
		return "", "", "", errors.New("response is incomplete")
	}
	network := nodeInfo.GetDefaultNodeInfo().Network
	applicationName := nodeInfo.GetApplicationVersion().AppName
	applicationVersion := nodeInfo.GetApplicationVersion().Version
	if network != expectedNetwork {
		return network, applicationName, applicationVersion, fmt.Errorf("network = %q, want %q", network, expectedNetwork)
	}
	if applicationName != configCompatibilityGRPCWebAppName || applicationVersion != upgradeBinaryVersion {
		return network, applicationName, applicationVersion, fmt.Errorf(
			"application identity = %q@%q, want %q@%q",
			applicationName,
			applicationVersion,
			configCompatibilityGRPCWebAppName,
			upgradeBinaryVersion,
		)
	}
	return network, applicationName, applicationVersion, nil
}

func decodeConfigCompatibilityGRPCWebUnaryResponse(body []byte) ([]byte, map[string]string, int, error) {
	var payload []byte
	trailers := make(map[string]string)
	frameCount := 0
	seenTrailers := false
	for len(body) > 0 {
		if len(body) < 5 {
			return nil, trailers, frameCount, errors.New("truncated gRPC-Web frame header")
		}
		flags := body[0]
		length := int(binary.BigEndian.Uint32(body[1:5]))
		body = body[5:]
		if length < 0 || length > len(body) {
			return nil, trailers, frameCount, errors.New("truncated gRPC-Web frame payload")
		}
		frame := body[:length]
		body = body[length:]
		frameCount++
		switch flags {
		case 0:
			if seenTrailers || payload != nil {
				return nil, trailers, frameCount, errors.New("gRPC-Web unary response contains an unexpected data frame")
			}
			payload = append([]byte(nil), frame...)
		case 0x80:
			if seenTrailers {
				return nil, trailers, frameCount, errors.New("gRPC-Web unary response contains duplicate trailers")
			}
			seenTrailers = true
			for _, line := range strings.Split(string(frame), "\r\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				name, value, ok := strings.Cut(line, ":")
				if !ok || strings.TrimSpace(name) == "" {
					return nil, trailers, frameCount, fmt.Errorf("malformed gRPC-Web trailer %q", line)
				}
				name = strings.ToLower(strings.TrimSpace(name))
				if _, duplicate := trailers[name]; duplicate {
					return nil, trailers, frameCount, fmt.Errorf("duplicate gRPC-Web trailer %q", name)
				}
				trailers[name] = strings.TrimSpace(value)
			}
		default:
			return nil, trailers, frameCount, fmt.Errorf("unsupported gRPC-Web frame flags 0x%02x", flags)
		}
	}
	if payload == nil {
		return nil, trailers, frameCount, errors.New("gRPC-Web unary response is missing its data frame")
	}
	if !seenTrailers {
		return nil, trailers, frameCount, errors.New("gRPC-Web unary response is missing its trailer frame")
	}
	if trailers["grpc-status"] != "0" {
		return nil, trailers, frameCount, fmt.Errorf("gRPC-Web grpc-status = %q, want 0", trailers["grpc-status"])
	}
	return payload, trailers, frameCount, nil
}

func scheduleConfigCompatibilityUpgrade(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	proposerKey string,
) int64 {
	t.Helper()
	baseHeight, err := network.Chain.Height(ctx)
	require.NoError(t, err)
	upgradeHeight := baseHeight + 55
	proposalTx, err := network.BroadcastAndWaitTx(
		ctx,
		"config-compat-submit-upgrade",
		network.Chain.Validators[0],
		proposerKey,
		"gov", "submit-legacy-proposal", "software-upgrade", upgradeName,
		"--title", "Panacea v2.3.0 config compatibility upgrade",
		"--description", "v0.47 node-home and v0.50 confix compatibility",
		"--deposit", "1umed",
		"--upgrade-height", strconv.FormatInt(upgradeHeight, 10),
		"--upgrade-info", "{}",
		"--no-validate",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	proposalID, err := proposalIDFromCommittedTx(proposalTx)
	require.NoError(t, err)
	_, err = network.BroadcastAndWaitTx(
		ctx,
		"config-compat-vote-upgrade",
		network.Chain.Validators[0],
		"validator",
		"gov", "vote", strconv.FormatUint(proposalID, 10), "yes",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	require.NoError(t, waitForProposalPassed(ctx, network, proposalID))
	require.NoError(t, network.WriteArtifactJSON("config-compat/upgrade-plan.json", map[string]any{
		"proposal_id":    proposalID,
		"upgrade_height": upgradeHeight,
		"tx_hash":        proposalTx.TxHash,
	}))
	return upgradeHeight
}

func readConfigCompatibilityNodeHome(ctx context.Context, node *cosmos.ChainNode) (harness.NodeHomeConfigSnapshot, error) {
	if node == nil {
		return harness.NodeHomeConfigSnapshot{}, errors.New("config compatibility node is required")
	}
	app, err := node.ReadFile(ctx, "config/app.toml")
	if err != nil {
		return harness.NodeHomeConfigSnapshot{}, fmt.Errorf("read app.toml: %w", err)
	}
	client, err := node.ReadFile(ctx, "config/client.toml")
	if err != nil {
		return harness.NodeHomeConfigSnapshot{}, fmt.Errorf("read client.toml: %w", err)
	}
	comet, err := node.ReadFile(ctx, "config/config.toml")
	if err != nil {
		return harness.NodeHomeConfigSnapshot{}, fmt.Errorf("read config.toml: %w", err)
	}
	snapshot := harness.NewNodeHomeConfigSnapshot(app, client, comet)
	if err := snapshot.Validate(); err != nil {
		return harness.NodeHomeConfigSnapshot{}, err
	}
	return snapshot, nil
}

func writeConfigCompatibilityNodeHome(ctx context.Context, node *cosmos.ChainNode, snapshot harness.NodeHomeConfigSnapshot) error {
	if node == nil {
		return errors.New("config compatibility node is required")
	}
	if err := snapshot.Validate(); err != nil {
		return err
	}
	for _, document := range []struct {
		path     string
		contents []byte
	}{
		{path: "config/app.toml", contents: snapshot.App},
		{path: "config/client.toml", contents: snapshot.Client},
		{path: "config/config.toml", contents: snapshot.Comet},
	} {
		if err := node.WriteFile(ctx, document.contents, document.path); err != nil {
			return fmt.Errorf("write %s: %w", document.path, err)
		}
	}
	return nil
}

func recordConfigCompatibilitySnapshot(network *harness.Network, phase string, snapshot harness.NodeHomeConfigSnapshot) error {
	if network == nil {
		return errors.New("config compatibility network is required")
	}
	if err := snapshot.Validate(); err != nil {
		return err
	}
	base := "config-compat/" + phase + "/"
	if err := network.WriteArtifact(base+"app.toml", snapshot.App); err != nil {
		return err
	}
	if err := network.WriteArtifact(base+"client.toml", snapshot.Client); err != nil {
		return err
	}
	if err := network.WriteArtifact(base+"config.toml", snapshot.Comet); err != nil {
		return err
	}
	return network.WriteArtifactJSON(base+"sha256.json", snapshot.SHA256)
}

func runConfigCompatibilityCommand(
	ctx context.Context,
	node *cosmos.ChainNode,
	name string,
	expected string,
	arguments ...string,
) configCompatibilityCommandEvidence {
	evidence := configCompatibilityCommandEvidence{
		RecordedAt: time.Now().UTC(),
		Name:       name,
		Arguments:  append([]string(nil), arguments...),
		Expected:   expected,
	}
	commandCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	stdout, stderr, err := node.ExecBin(commandCtx, arguments...)
	evidence.Stdout = string(stdout)
	evidence.Stderr = string(stderr)
	if err != nil {
		evidence.Error = err.Error()
	}
	return evidence
}

func recordConfigCompatibilityCommand(network *harness.Network, evidence configCompatibilityCommandEvidence) error {
	return network.AppendArtifactJSON("config-compat/commands.jsonl", evidence)
}

func requireConfigCompatibilityCommandContracts(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	node *cosmos.ChainNode,
) {
	t.Helper()
	oldClient := runConfigCompatibilityCommand(ctx, node, "old-client-command", "clear unknown-command failure", "config", "chain-id")
	require.NoError(t, recordConfigCompatibilityCommand(network, oldClient))
	require.NotEmpty(t, oldClient.Error)
	require.Contains(t, strings.ToLower(oldClient.Stderr+oldClient.Stdout+oldClient.Error), "unknown command")

	newClient := runConfigCompatibilityCommand(ctx, node, "new-client-command", "success", "config", "get", "client", "chain-id")
	require.NoError(t, recordConfigCompatibilityCommand(network, newClient))
	require.Empty(t, newClient.Error)
	chainID := strings.Trim(strings.TrimSpace(newClient.Stdout), `"`)
	require.Equal(t, network.Chain.Config().ChainID, chainID)

	removedFlag := runConfigCompatibilityCommand(
		ctx,
		node,
		"removed-grpc-web-address-flag",
		"clear unknown-flag failure",
		"start",
		"--grpc-web.address=localhost:9091",
		"--help",
	)
	require.NoError(t, recordConfigCompatibilityCommand(network, removedFlag))
	require.NotEmpty(t, removedFlag.Error)
	removedDiagnostics := strings.ToLower(removedFlag.Stdout + removedFlag.Stderr + removedFlag.Error)
	require.Contains(t, removedDiagnostics, "unknown flag")
	require.Contains(t, removedDiagnostics, "grpc-web.address")

	newFlag := runConfigCompatibilityCommand(
		ctx,
		node,
		"current-api-address-flag",
		"success",
		"start",
		"--api.address=tcp://127.0.0.1:1317",
		"--help",
	)
	require.NoError(t, recordConfigCompatibilityCommand(network, newFlag))
	require.Empty(t, newFlag.Error)
	require.Contains(t, newFlag.Stdout+newFlag.Stderr, "--api.address")

	for _, item := range []struct {
		name string
		path string
		want string
	}{
		{name: "query-gas-limit", path: "query-gas-limit", want: "7654321"},
		{name: "api-read-timeout", path: "api.rpc-read-timeout", want: "13"},
		{name: "api-write-timeout", path: "api.rpc-write-timeout", want: "17"},
		{name: "api-max-body", path: "api.rpc-max-body-bytes", want: "765432"},
		{name: "grpc-max-recv", path: "grpc.max-recv-msg-size", want: "8388608"},
		{name: "grpc-max-send", path: "grpc.max-send-msg-size", want: "9437184"},
	} {
		evidence := runConfigCompatibilityCommand(ctx, node, "get-"+item.name, "success with preserved value", "config", "get", "app", item.path)
		require.NoError(t, recordConfigCompatibilityCommand(network, evidence))
		require.Empty(t, evidence.Error)
		value := strings.Trim(strings.TrimSpace(evidence.Stdout), `"`)
		parsed, parseErr := strconv.ParseInt(value, 10, 64)
		require.NoError(t, parseErr, "%s returned %q", item.path, evidence.Stdout)
		want, parseErr := strconv.ParseInt(item.want, 10, 64)
		require.NoError(t, parseErr)
		require.Equal(t, want, parsed, item.path)
	}
}
