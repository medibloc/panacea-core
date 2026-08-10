package harness

import (
	"errors"
	"fmt"
	"time"
)

// NetworkFaultCleanupEvidence records every suite-owned restore/stop cleanup,
// including failures that must also make the enclosing test fail.
type NetworkFaultCleanupEvidence struct {
	Phase      string    `json:"phase"`
	Result     string    `json:"result"`
	Error      string    `json:"error,omitempty"`
	RecordedAt time.Time `json:"recorded_at"`
}

// RecordNetworkFaultCleanup keeps a cleanup failure visible at all three
// boundaries: the caller receives the original error, the cleanup timeline
// records it, and the run manifest receives a typed failure stage.
func (n *Network) RecordNetworkFaultCleanup(phase string, cleanupErr error) error {
	if n == nil || n.artifacts == nil {
		return errors.Join(cleanupErr, errors.New("network fault artifact store is unavailable"))
	}
	if !networkFaultNamePattern.MatchString(phase) {
		return errors.Join(cleanupErr, fmt.Errorf("network fault cleanup phase %q must match %s", phase, networkFaultNamePattern))
	}
	evidence := NetworkFaultCleanupEvidence{
		Phase:      phase,
		Result:     "succeeded",
		Error:      errorString(cleanupErr),
		RecordedAt: time.Now().UTC(),
	}
	if cleanupErr != nil {
		evidence.Result = "failed"
	}
	recordErr := n.artifacts.appendJSONLine("network-faults/cleanup.jsonl", evidence)
	if cleanupErr != nil {
		n.artifacts.recordFailure("network-fault-cleanup-"+phase, cleanupErr)
	}
	if recordErr != nil {
		n.artifacts.recordFailure("network-fault-cleanup-artifact-"+phase, recordErr)
	}
	return errors.Join(cleanupErr, recordErr)
}
