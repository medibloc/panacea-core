package harness

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type NetworkFaultCategory string

const (
	NetworkFaultCategoryEnvironmentPreflight NetworkFaultCategory = "environment-preflight"
	NetworkFaultCategoryChainP2PRuntime      NetworkFaultCategory = "chain-p2p-runtime"
)

type NetworkFaultOutcome string

const (
	NetworkFaultOutcomePassed                NetworkFaultOutcome = "passed"
	NetworkFaultOutcomeFailed                NetworkFaultOutcome = "failed"
	NetworkFaultOutcomeExpectedFaultObserved NetworkFaultOutcome = "expected-fault-observed"
	NetworkFaultOutcomeRecovered             NetworkFaultOutcome = "recovered"
)

const (
	NetworkFaultScopeLocalEnvironment  = "local-environment"
	NetworkFaultScopeRunOwnedDockerP2P = "run-owned-docker-p2p"
)

// NetworkFaultCategoryEvidence prevents Docker/sandbox/tooling setup failures
// from being reported as evidence of a chain P2P fault.
type NetworkFaultCategoryEvidence struct {
	Category   NetworkFaultCategory `json:"category"`
	Phase      string               `json:"phase"`
	Outcome    NetworkFaultOutcome  `json:"outcome"`
	Scope      string               `json:"scope"`
	Error      string               `json:"error,omitempty"`
	RecordedAt time.Time            `json:"recorded_at"`
}

func (e NetworkFaultCategoryEvidence) Validate() error {
	if !networkFaultNamePattern.MatchString(e.Phase) {
		return fmt.Errorf("network fault category phase %q must match %s", e.Phase, networkFaultNamePattern)
	}
	if e.RecordedAt.IsZero() {
		return errors.New("network fault category recorded_at is required")
	}
	switch e.Outcome {
	case NetworkFaultOutcomePassed, NetworkFaultOutcomeFailed,
		NetworkFaultOutcomeExpectedFaultObserved, NetworkFaultOutcomeRecovered:
	default:
		return fmt.Errorf("unsupported network fault outcome %q", e.Outcome)
	}
	if e.Outcome == NetworkFaultOutcomeFailed && strings.TrimSpace(e.Error) == "" {
		return errors.New("failed network fault category requires an error")
	}
	if e.Outcome != NetworkFaultOutcomeFailed && strings.TrimSpace(e.Error) != "" {
		return fmt.Errorf("network fault outcome %q cannot carry an error", e.Outcome)
	}
	switch e.Category {
	case NetworkFaultCategoryEnvironmentPreflight:
		if e.Scope != NetworkFaultScopeLocalEnvironment {
			return fmt.Errorf("environment preflight scope %q, want %q", e.Scope, NetworkFaultScopeLocalEnvironment)
		}
		if e.Outcome != NetworkFaultOutcomePassed && e.Outcome != NetworkFaultOutcomeFailed {
			return fmt.Errorf("environment preflight outcome %q must be passed or failed", e.Outcome)
		}
	case NetworkFaultCategoryChainP2PRuntime:
		if e.Scope != NetworkFaultScopeRunOwnedDockerP2P {
			return fmt.Errorf("chain P2P runtime scope %q, want %q", e.Scope, NetworkFaultScopeRunOwnedDockerP2P)
		}
		if e.Outcome == NetworkFaultOutcomePassed {
			return errors.New("chain P2P runtime must identify an observed fault or recovery")
		}
	default:
		return fmt.Errorf("unsupported network fault category %q", e.Category)
	}
	return nil
}

// RecordNetworkFaultCategory writes environment and chain-runtime observations
// to distinct paths. Only an actual failed outcome enters the manifest failure
// list; an expected injected fault is evidence, not a test failure.
func (n *Network) RecordNetworkFaultCategory(evidence NetworkFaultCategoryEvidence) error {
	if n == nil || n.artifacts == nil {
		return errors.New("network fault artifact store is unavailable")
	}
	if evidence.RecordedAt.IsZero() {
		evidence.RecordedAt = time.Now().UTC()
	}
	if err := evidence.Validate(); err != nil {
		return err
	}
	path := "network-faults/failure-categories.jsonl"
	if evidence.Category == NetworkFaultCategoryEnvironmentPreflight {
		path = "environment/network-failure-categories.jsonl"
	}
	recordErr := n.artifacts.appendJSONLine(path, evidence)
	if recordErr != nil {
		n.artifacts.recordFailure("network-fault-category-artifact-"+evidence.Phase, recordErr)
		return recordErr
	}
	if evidence.Outcome == NetworkFaultOutcomeFailed {
		n.artifacts.recordFailure(string(evidence.Category)+"-"+evidence.Phase, errors.New(evidence.Error))
	}
	return nil
}

func (s *artifactStore) recordNetworkFaultSetupFailure(category NetworkFaultCategory, phase string, setupErr error) error {
	if s == nil || setupErr == nil || category == "" {
		return nil
	}
	network := &Network{artifacts: s}
	return network.RecordNetworkFaultCategory(NetworkFaultCategoryEvidence{
		Category: category,
		Phase:    phase,
		Outcome:  NetworkFaultOutcomeFailed,
		Scope:    NetworkFaultScopeLocalEnvironment,
		Error:    setupErr.Error(),
	})
}
