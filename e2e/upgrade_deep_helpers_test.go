package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

type upgradeBankAccountState struct {
	Address       string `json:"address"`
	AccountNumber uint64 `json:"account_number"`
	Sequence      uint64 `json:"sequence"`
	Balance       string `json:"balance"`
}

type upgradeStateCheckpoint struct {
	Phase       string                               `json:"phase"`
	RecordedAt  time.Time                            `json:"recorded_at"`
	Height      int64                                `json:"height"`
	Observation harness.UpgradeCheckpointObservation `json:"observation"`
	Bank        upgradeBankAccountState              `json:"bank"`
	TxHashes    []string                             `json:"tx_hashes,omitempty"`
}

type upgradeRestartNodeEvidence struct {
	Node        string                 `json:"node"`
	ContainerID string                 `json:"container_id"`
	ImageID     string                 `json:"image_id"`
	Before      harness.BlockEvidence  `json:"before"`
	After       *harness.BlockEvidence `json:"after,omitempty"`
}

type upgradeRestartEvidence struct {
	Phase        string                       `json:"phase"`
	StartedAt    time.Time                    `json:"started_at"`
	StoppedAt    *time.Time                   `json:"stopped_at,omitempty"`
	RestartedAt  *time.Time                   `json:"restarted_at,omitempty"`
	CompletedAt  *time.Time                   `json:"completed_at,omitempty"`
	BeforeHeight int64                        `json:"before_height"`
	TargetHeight int64                        `json:"target_height"`
	Nodes        []upgradeRestartNodeEvidence `json:"nodes"`
	StopError    string                       `json:"stop_error,omitempty"`
	StartError   string                       `json:"start_error,omitempty"`
	ObserveError string                       `json:"observe_error,omitempty"`
}

func restartUpgradeNetworkWithEvidence(
	ctx context.Context,
	network *harness.Network,
	phase string,
	advance int64,
) (upgradeRestartEvidence, error) {
	if strings.TrimSpace(phase) == "" {
		return upgradeRestartEvidence{}, errors.New("upgrade restart phase is required")
	}
	if advance < 1 {
		return upgradeRestartEvidence{}, fmt.Errorf("upgrade restart advance must be positive, got %d", advance)
	}
	beforeHeight, err := network.Chain.Height(ctx)
	if err != nil {
		return upgradeRestartEvidence{}, fmt.Errorf("query pre-restart height: %w", err)
	}
	evidence := upgradeRestartEvidence{
		Phase:        phase,
		StartedAt:    time.Now().UTC(),
		BeforeHeight: beforeHeight,
		TargetHeight: beforeHeight + advance,
		Nodes:        make([]upgradeRestartNodeEvidence, len(network.Chain.Nodes())),
	}
	for index, node := range network.Chain.Nodes() {
		block, blockErr := network.NodeLatestBlock(ctx, node)
		if blockErr != nil {
			return upgradeRestartEvidence{}, fmt.Errorf("capture pre-restart block for %s: %w", node.Name(), blockErr)
		}
		inspect, inspectErr := node.DockerClient.ContainerInspect(ctx, node.ContainerID())
		if inspectErr != nil {
			return upgradeRestartEvidence{}, fmt.Errorf("inspect pre-restart container %s: %w", node.Name(), inspectErr)
		}
		evidence.Nodes[index] = upgradeRestartNodeEvidence{
			Node:        node.Name(),
			ContainerID: node.ContainerID(),
			ImageID:     inspect.Image,
			Before:      block,
		}
	}

	var operationErrors []error
	if artifactErr := network.AppendArtifactJSON("upgrade/restart-events.jsonl", map[string]any{
		"event":       "restart-started",
		"recorded_at": evidence.StartedAt,
		"evidence":    evidence,
	}); artifactErr != nil {
		operationErrors = append(operationErrors, artifactErr)
	}

	stopErr := network.Chain.StopAllNodes(ctx)
	stoppedAt := time.Now().UTC()
	evidence.StoppedAt = &stoppedAt
	evidence.StopError = errorText(stopErr)
	if stopErr != nil {
		operationErrors = append(operationErrors, fmt.Errorf("stop all upgrade nodes: %w", stopErr))
	}
	if artifactErr := network.AppendArtifactJSON("upgrade/restart-events.jsonl", map[string]any{
		"event":       "nodes-stopped",
		"recorded_at": stoppedAt,
		"stop_error":  evidence.StopError,
	}); artifactErr != nil {
		operationErrors = append(operationErrors, artifactErr)
	}

	startErr := network.Chain.StartAllNodes(ctx)
	restartedAt := time.Now().UTC()
	evidence.RestartedAt = &restartedAt
	evidence.StartError = errorText(startErr)
	if startErr != nil {
		operationErrors = append(operationErrors, fmt.Errorf("start all upgrade nodes: %w", startErr))
	}

	var observeErrors []error
	if startErr == nil {
		for index, node := range network.Chain.Nodes() {
			if waitErr := network.WaitForNodeHeight(ctx, node, evidence.TargetHeight); waitErr != nil {
				observeErrors = append(observeErrors, fmt.Errorf("wait for restarted %s: %w", node.Name(), waitErr))
				continue
			}
			block, blockErr := network.NodeLatestBlock(ctx, node)
			if blockErr != nil {
				observeErrors = append(observeErrors, fmt.Errorf("capture post-restart block for %s: %w", node.Name(), blockErr))
				continue
			}
			inspect, inspectErr := node.DockerClient.ContainerInspect(ctx, node.ContainerID())
			if inspectErr != nil {
				observeErrors = append(observeErrors, fmt.Errorf("inspect post-restart container %s: %w", node.Name(), inspectErr))
				continue
			}
			evidence.Nodes[index].ContainerID = node.ContainerID()
			evidence.Nodes[index].ImageID = inspect.Image
			evidence.Nodes[index].After = &block
		}
	}
	observeErr := errors.Join(observeErrors...)
	evidence.ObserveError = errorText(observeErr)
	if observeErr != nil {
		operationErrors = append(operationErrors, observeErr)
	}
	completedAt := time.Now().UTC()
	evidence.CompletedAt = &completedAt
	if artifactErr := network.AppendArtifactJSON("upgrade/restart-events.jsonl", map[string]any{
		"event":       "restart-completed",
		"recorded_at": completedAt,
		"evidence":    evidence,
	}); artifactErr != nil {
		operationErrors = append(operationErrors, artifactErr)
	}
	return evidence, errors.Join(operationErrors...)
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func captureUpgradeStateCheckpoint(
	ctx context.Context,
	network *harness.Network,
	phase string,
	address string,
	txHashes []string,
) (upgradeStateCheckpoint, error) {
	if strings.TrimSpace(phase) == "" {
		return upgradeStateCheckpoint{}, errors.New("upgrade checkpoint phase is required")
	}
	fullNode := network.Chain.FullNodes[0]
	observation, err := network.CaptureUpgradeCheckpointObservation(
		ctx,
		"upgrade-"+phase,
		fullNode,
		0,
	)
	if err != nil {
		return upgradeStateCheckpoint{}, fmt.Errorf("observe %s checkpoint block: %w", phase, err)
	}
	account, err := queryUpgradeBankAccountAtHeight(
		ctx,
		network,
		"upgrade-"+phase+"-account",
		address,
		observation.Height,
	)
	if err != nil {
		return upgradeStateCheckpoint{}, err
	}
	checkpoint := upgradeStateCheckpoint{
		Phase:       phase,
		RecordedAt:  observation.ObservedAt,
		Height:      observation.Height,
		Observation: observation,
		Bank:        account,
		TxHashes:    append([]string(nil), txHashes...),
	}
	if err := network.WriteArtifactJSON("state-checkpoints/"+phase+".json", checkpoint); err != nil {
		return upgradeStateCheckpoint{}, fmt.Errorf("record %s state checkpoint: %w", phase, err)
	}
	if err := network.AppendArtifactJSON("upgrade/phases.jsonl", checkpoint); err != nil {
		return upgradeStateCheckpoint{}, fmt.Errorf("record %s phase: %w", phase, err)
	}
	return checkpoint, nil
}

func queryUpgradeBankAccount(
	ctx context.Context,
	network *harness.Network,
	step string,
	address string,
) (upgradeBankAccountState, error) {
	return queryUpgradeBankAccountAtHeight(ctx, network, step, address, 0)
}

func queryUpgradeBankAccountAtHeight(
	ctx context.Context,
	network *harness.Network,
	step string,
	address string,
	height int64,
) (upgradeBankAccountState, error) {
	command := []string{"auth", "account", address}
	queryCtx := ctx
	if height > 0 {
		command = append(command, "--height", strconv.FormatInt(height, 10))
		queryCtx = harness.ContextAtHeight(ctx, height)
	}
	raw, err := network.FullNodeCLIQuery(ctx, step, command...)
	if err != nil {
		return upgradeBankAccountState{}, err
	}
	account, err := decodeUpgradeBankAccount(raw, address)
	if err != nil {
		return upgradeBankAccountState{}, fmt.Errorf("decode %s: %w", step, err)
	}
	balance, err := network.QueryFullNodeBalance(queryCtx, address, "umed")
	if err != nil {
		return upgradeBankAccountState{}, fmt.Errorf("query %s balance: %w", step, err)
	}
	account.Balance = balance.String()
	return account, nil
}

func decodeUpgradeBankAccount(raw []byte, expectedAddress string) (upgradeBankAccountState, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return upgradeBankAccountState{}, fmt.Errorf("decode auth account JSON: %w", err)
	}
	account, ok, err := findUpgradeBaseAccount(value, expectedAddress)
	if err != nil {
		return upgradeBankAccountState{}, err
	}
	if !ok {
		return upgradeBankAccountState{}, fmt.Errorf("auth account response has no base account for %s", expectedAddress)
	}
	return account, nil
}

func findUpgradeBaseAccount(value any, expectedAddress string) (upgradeBankAccountState, bool, error) {
	switch typed := value.(type) {
	case map[string]any:
		address, hasAddress := typed["address"].(string)
		accountNumberValue, hasAccountNumber := typed["account_number"]
		sequenceValue, hasSequence := typed["sequence"]
		if hasAddress && hasAccountNumber && address == expectedAddress {
			accountNumber, err := upgradeUint64(accountNumberValue)
			if err != nil {
				return upgradeBankAccountState{}, false, fmt.Errorf("decode account_number: %w", err)
			}
			sequence := uint64(0)
			if hasSequence {
				sequence, err = upgradeUint64(sequenceValue)
				if err != nil {
					return upgradeBankAccountState{}, false, fmt.Errorf("decode sequence: %w", err)
				}
			}
			return upgradeBankAccountState{
				Address:       address,
				AccountNumber: accountNumber,
				Sequence:      sequence,
			}, true, nil
		}
		for _, child := range typed {
			account, ok, err := findUpgradeBaseAccount(child, expectedAddress)
			if err != nil || ok {
				return account, ok, err
			}
		}
	case []any:
		for _, child := range typed {
			account, ok, err := findUpgradeBaseAccount(child, expectedAddress)
			if err != nil || ok {
				return account, ok, err
			}
		}
	}
	return upgradeBankAccountState{}, false, nil
}

func upgradeUint64(value any) (uint64, error) {
	switch typed := value.(type) {
	case string:
		return strconv.ParseUint(typed, 10, 64)
	case json.Number:
		return strconv.ParseUint(typed.String(), 10, 64)
	default:
		return 0, fmt.Errorf("unexpected JSON value %T", value)
	}
}

func recordUpgradeHistoricalTx(
	ctx context.Context,
	network *harness.Network,
	phase string,
	txHash string,
) error {
	if strings.TrimSpace(txHash) == "" {
		return errors.New("historical transaction hash is required")
	}
	raw, err := network.FullNodeCLIQuery(ctx, "upgrade-"+phase+"-historical-tx", "tx", txHash)
	if err != nil {
		return err
	}
	var result struct {
		Height string `json:"height"`
		TxHash string `json:"txhash"`
		Code   uint32 `json:"code"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return fmt.Errorf("decode historical transaction %s: %w", txHash, err)
	}
	if !strings.EqualFold(result.TxHash, txHash) {
		return fmt.Errorf("historical transaction returned hash %s, want %s", result.TxHash, txHash)
	}
	if result.Code != 0 {
		return fmt.Errorf("historical transaction %s returned code %d", txHash, result.Code)
	}
	height, err := strconv.ParseInt(result.Height, 10, 64)
	if err != nil || height <= 0 {
		return fmt.Errorf("historical transaction %s has invalid height %q", txHash, result.Height)
	}
	return network.AppendArtifactJSON("tx/historical-lookups.jsonl", map[string]any{
		"recorded_at": time.Now().UTC(),
		"phase":       phase,
		"tx_hash":     txHash,
		"height":      height,
		"code":        result.Code,
		"response":    json.RawMessage(raw),
	})
}

func requireUpgradeBankAccountEqual(t *testing.T, expected, actual upgradeBankAccountState) {
	t.Helper()
	require.Equal(t, expected.Address, actual.Address)
	require.Equal(t, expected.AccountNumber, actual.AccountNumber)
	require.Equal(t, expected.Sequence, actual.Sequence)
	require.Equal(t, expected.Balance, actual.Balance)
}

func TestDecodeUpgradeBankAccountAcrossSDKShapes(t *testing.T) {
	t.Parallel()

	const address = "panacea1account"
	for _, tc := range []struct {
		name string
		raw  string
	}{
		{
			name: "direct base account",
			raw:  `{"account":{"@type":"/cosmos.auth.v1beta1.BaseAccount","address":"panacea1account","account_number":"17","sequence":"4"}}`,
		},
		{
			name: "nested vesting base account",
			raw:  `{"account":{"@type":"/cosmos.vesting.v1beta1.ContinuousVestingAccount","base_vesting_account":{"base_account":{"address":"panacea1account","account_number":17,"sequence":4}}}}`,
		},
		{
			name: "current type value base account with omitted zero sequence",
			raw:  `{"account":{"type":"/cosmos.auth.v1beta1.BaseAccount","value":{"address":"panacea1account","account_number":"17"}}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account, err := decodeUpgradeBankAccount([]byte(tc.raw), address)
			require.NoError(t, err)
			require.Equal(t, address, account.Address)
			require.Equal(t, uint64(17), account.AccountNumber)
			if tc.name == "current type value base account with omitted zero sequence" {
				require.Zero(t, account.Sequence)
			} else {
				require.Equal(t, uint64(4), account.Sequence)
			}
		})
	}
}

func TestDecodeUpgradeBankAccountRejectsMissingAndInvalidState(t *testing.T) {
	t.Parallel()

	_, err := decodeUpgradeBankAccount([]byte(`{"account":{"address":"other","account_number":"1","sequence":"0"}}`), "panacea1account")
	require.ErrorContains(t, err, "no base account")

	_, err = decodeUpgradeBankAccount([]byte(`{"account":{"address":"panacea1account","account_number":"bad","sequence":"0"}}`), "panacea1account")
	require.ErrorContains(t, err, "account_number")
}
