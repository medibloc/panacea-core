package harness

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// QuorumCommitment is one node's block and application commitment at an exact
// height. Comparing both hashes proves that nodes share one consensus history
// and one committed application state.
type QuorumCommitment struct {
	Node      string `json:"node"`
	Height    int64  `json:"height"`
	BlockHash string `json:"block_hash"`
	AppHash   string `json:"app_hash"`
}

// QuorumAgreement is the common commitment reported by all named nodes.
type QuorumAgreement struct {
	Height    int64    `json:"height"`
	BlockHash string   `json:"block_hash"`
	AppHash   string   `json:"app_hash"`
	Nodes     []string `json:"nodes"`
}

// QuorumHeightSample retains every height read used to prove progress or a
// bounded halt. Read errors are evidence too and never count as a stable node.
type QuorumHeightSample struct {
	ObservedAt time.Time `json:"observed_at"`
	Height     int64     `json:"height,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// QuorumHeightWindow is the bounded height evidence returned by an observer.
type QuorumHeightWindow struct {
	StartHeight  int64                `json:"start_height"`
	EndHeight    int64                `json:"end_height"`
	TargetHeight int64                `json:"target_height,omitempty"`
	Samples      []QuorumHeightSample `json:"samples"`
}

// QuorumObserver polls a node boundary at a caller-selected cadence. Live
// suites use a conservative cadence; unit tests can use a much shorter one.
type QuorumObserver struct {
	pollInterval time.Duration
}

// NewQuorumObserver constructs a bounded height observer.
func NewQuorumObserver(pollInterval time.Duration) (*QuorumObserver, error) {
	if pollInterval <= 0 {
		return nil, errors.New("quorum poll interval must be positive")
	}
	return &QuorumObserver{pollInterval: pollInterval}, nil
}

// WaitForProgress waits until the observed height advances by minimumBlocks.
func (o *QuorumObserver) WaitForProgress(
	ctx context.Context,
	startHeight int64,
	minimumBlocks int64,
	readHeight func(context.Context) (int64, error),
) (QuorumHeightWindow, error) {
	window := QuorumHeightWindow{StartHeight: startHeight, EndHeight: startHeight}
	if o == nil || o.pollInterval <= 0 {
		return window, errors.New("quorum observer is not initialized")
	}
	if startHeight < 0 {
		return window, errors.New("start height cannot be negative")
	}
	if minimumBlocks <= 0 {
		return window, errors.New("minimum block progress must be positive")
	}
	if readHeight == nil {
		return window, errors.New("height reader is required")
	}
	if startHeight > math.MaxInt64-minimumBlocks {
		return window, fmt.Errorf("progress target %d + %d overflows int64", startHeight, minimumBlocks)
	}
	window.TargetHeight = startHeight + minimumBlocks

	ticker := time.NewTicker(o.pollInterval)
	defer ticker.Stop()
	var lastErr error
	for {
		height, err := readHeight(ctx)
		sample := QuorumHeightSample{ObservedAt: time.Now().UTC(), Height: height, Error: quorumErrorString(err)}
		window.Samples = append(window.Samples, sample)
		if err == nil {
			window.EndHeight = height
			if height >= window.TargetHeight {
				return window, nil
			}
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return window, fmt.Errorf(
				"wait for quorum progress from %d to %d: last height=%d last error=%v: %w",
				startHeight,
				window.TargetHeight,
				window.EndHeight,
				lastErr,
				ctx.Err(),
			)
		case <-ticker.C:
		}
	}
}

func quorumErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// ObserveStall first waits for a quiet height to absorb any block already in
// flight when quorum is removed, then proves that exact height remains fixed
// for the full observation window. A read error never counts as proof.
func (o *QuorumObserver) ObserveStall(
	ctx context.Context,
	quietPeriod time.Duration,
	observationPeriod time.Duration,
	readHeight func(context.Context) (int64, error),
) (QuorumHeightWindow, error) {
	window := QuorumHeightWindow{}
	if o == nil || o.pollInterval <= 0 {
		return window, errors.New("quorum observer is not initialized")
	}
	if quietPeriod <= 0 {
		return window, errors.New("stall quiet period must be positive")
	}
	if observationPeriod <= 0 {
		return window, errors.New("stall observation period must be positive")
	}
	if readHeight == nil {
		return window, errors.New("height reader is required")
	}

	ticker := time.NewTicker(o.pollInterval)
	defer ticker.Stop()
	var (
		baseline            int64
		quietSince          time.Time
		observationDeadline time.Time
		lastErr             error
	)
	for {
		now := time.Now().UTC()
		height, err := readHeight(ctx)
		window.Samples = append(window.Samples, QuorumHeightSample{
			ObservedAt: now,
			Height:     height,
			Error:      quorumErrorString(err),
		})
		if err != nil {
			lastErr = err
			if observationDeadline.IsZero() {
				// A failed read cannot contribute time toward establishing a
				// quiet baseline. Start the quiet proof over after recovery.
				quietSince = time.Time{}
			} else {
				// Require one complete observation period after the latest
				// visibility gap instead of treating it as evidence of a halt.
				observationDeadline = now.Add(observationPeriod)
			}
		} else if quietSince.IsZero() {
			baseline = height
			window.StartHeight = height
			window.EndHeight = height
			quietSince = now
		} else if observationDeadline.IsZero() {
			window.EndHeight = height
			if height != baseline {
				baseline = height
				window.StartHeight = height
				quietSince = now
			} else if now.Sub(quietSince) >= quietPeriod {
				observationDeadline = now.Add(observationPeriod)
			}
		} else {
			window.EndHeight = height
			if height != baseline {
				return window, fmt.Errorf(
					"quorum advanced during bounded stall window: baseline=%d observed=%d",
					baseline,
					height,
				)
			}
			if !now.Before(observationDeadline) {
				return window, nil
			}
		}

		select {
		case <-ctx.Done():
			return window, fmt.Errorf(
				"observe quorum stall: baseline=%d last height=%d last error=%v: %w",
				baseline,
				window.EndHeight,
				lastErr,
				ctx.Err(),
			)
		case <-ticker.C:
		}
	}
}

// VerifyCommonCommitment requires every supplied node to report the exact
// target height and identical non-empty block and application hashes.
func VerifyCommonCommitment(targetHeight int64, commitments []QuorumCommitment) (QuorumAgreement, error) {
	if targetHeight <= 0 {
		return QuorumAgreement{}, errors.New("commitment height must be positive")
	}
	if len(commitments) < 2 {
		return QuorumAgreement{}, errors.New("at least two node commitments are required")
	}

	first := commitments[0]
	if strings.TrimSpace(first.BlockHash) == "" || strings.TrimSpace(first.AppHash) == "" {
		return QuorumAgreement{}, fmt.Errorf("node %q returned an empty commitment hash", first.Node)
	}
	agreement := QuorumAgreement{
		Height:    targetHeight,
		BlockHash: first.BlockHash,
		AppHash:   first.AppHash,
		Nodes:     make([]string, 0, len(commitments)),
	}
	seenNodes := make(map[string]struct{}, len(commitments))
	for _, commitment := range commitments {
		if strings.TrimSpace(commitment.Node) == "" {
			return QuorumAgreement{}, errors.New("commitment node name is required")
		}
		if _, exists := seenNodes[commitment.Node]; exists {
			return QuorumAgreement{}, fmt.Errorf("duplicate commitment evidence for node %s", commitment.Node)
		}
		seenNodes[commitment.Node] = struct{}{}
		if commitment.Height != targetHeight {
			return QuorumAgreement{}, fmt.Errorf(
				"node %s reported commitment height %d, want %d",
				commitment.Node,
				commitment.Height,
				targetHeight,
			)
		}
		if commitment.BlockHash != agreement.BlockHash || commitment.AppHash != agreement.AppHash {
			return QuorumAgreement{}, fmt.Errorf(
				"node %s commitment differs at height %d: block=%s app=%s, want block=%s app=%s",
				commitment.Node,
				targetHeight,
				commitment.BlockHash,
				commitment.AppHash,
				agreement.BlockHash,
				agreement.AppHash,
			)
		}
		agreement.Nodes = append(agreement.Nodes, commitment.Node)
	}
	return agreement, nil
}
