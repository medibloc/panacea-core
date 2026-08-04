package harness

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// CometStateSyncUnavailableProviderLogEvidence captures the light-client
// setup boundary and the underlying RPC transport failure.
type CometStateSyncUnavailableProviderLogEvidence struct {
	LightClientSetupFailed   bool     `json:"light_client_setup_failed"`
	ProviderTransportFailure bool     `json:"provider_transport_failure"`
	UnexpectedSuccess        bool     `json:"unexpected_success"`
	MatchedLines             []string `json:"matched_lines,omitempty"`
}

func parseCometStateSyncUnavailableProviderLogs(contents []byte) CometStateSyncUnavailableProviderLogEvidence {
	var evidence CometStateSyncUnavailableProviderLogEvidence
	for _, rawLine := range strings.Split(string(contents), "\n") {
		line := plainStateSyncLogLine(rawLine)
		lower := strings.ToLower(line)
		matched := false
		if strings.Contains(lower, "failed to set up light client state provider") {
			evidence.LightClientSetupFailed = true
			matched = true
		}
		if strings.Contains(lower, "connection refused") ||
			strings.Contains(lower, "connect: network is unreachable") ||
			strings.Contains(lower, "connect: no route to host") ||
			(strings.Contains(lower, "dial tcp") &&
				(strings.Contains(lower, "127.0.0.1:1") || strings.Contains(lower, "127.0.0.1:2"))) {
			evidence.ProviderTransportFailure = true
			matched = true
		}
		if strings.Contains(line, "Snapshot restored") || strings.Contains(line, "Verified ABCI app") {
			evidence.UnexpectedSuccess = true
			matched = true
		}
		if matched && len(evidence.MatchedLines) < cometStateSyncMaximumMatchedArtifactLines {
			evidence.MatchedLines = append(evidence.MatchedLines, boundedLine(line))
		}
	}
	return evidence
}

func validateCometStateSyncUnavailableProviderFailure(
	logs CometStateSyncUnavailableProviderLogEvidence,
	elapsed time.Duration,
	limit time.Duration,
) error {
	if err := validateCometStateSyncFaultDeadline(elapsed, limit, "unavailable-provider"); err != nil {
		return err
	}
	if logs.UnexpectedSuccess {
		return errors.New("unavailable-provider node unexpectedly restored and verified a snapshot")
	}
	if !logs.LightClientSetupFailed {
		return errors.New("unavailable-provider logs do not prove light-client state-provider setup failure")
	}
	if !logs.ProviderTransportFailure {
		return errors.New("unavailable-provider logs do not prove an RPC transport failure")
	}
	return nil
}

// CometStateSyncCorruptedChunkLogEvidence captures both the Cosmos SDK
// checksum contract and CometBFT's exhausted snapshot-provider contract.
type CometStateSyncCorruptedChunkLogEvidence struct {
	ChecksumMismatches int      `json:"checksum_mismatches"`
	NoValidPeers       bool     `json:"no_valid_peers"`
	UnexpectedSuccess  bool     `json:"unexpected_success"`
	MatchedLines       []string `json:"matched_lines,omitempty"`
}

func parseCometStateSyncCorruptedChunkLogs(contents []byte) CometStateSyncCorruptedChunkLogEvidence {
	var evidence CometStateSyncCorruptedChunkLogEvidence
	for _, rawLine := range strings.Split(string(contents), "\n") {
		line := plainStateSyncLogLine(rawLine)
		lower := strings.ToLower(line)
		matched := false
		if strings.Contains(lower, "chunk checksum mismatch") &&
			strings.Contains(lower, "rejecting sender") &&
			strings.Contains(lower, "requesting refetch") {
			evidence.ChecksumMismatches++
			matched = true
		}
		if strings.Contains(lower, "no valid peers found for snapshot") {
			evidence.NoValidPeers = true
			matched = true
		}
		if strings.Contains(line, "Snapshot restored") || strings.Contains(line, "Verified ABCI app") {
			evidence.UnexpectedSuccess = true
			matched = true
		}
		if matched && len(evidence.MatchedLines) < cometStateSyncMaximumMatchedArtifactLines {
			evidence.MatchedLines = append(evidence.MatchedLines, boundedLine(line))
		}
	}
	return evidence
}

func validateCometStateSyncCorruptedChunkFailure(
	logs CometStateSyncCorruptedChunkLogEvidence,
	elapsed time.Duration,
	limit time.Duration,
) error {
	if err := validateCometStateSyncFaultDeadline(elapsed, limit, "corrupted-chunk"); err != nil {
		return err
	}
	if logs.UnexpectedSuccess {
		return errors.New("corrupted-chunk node unexpectedly restored and verified a snapshot")
	}
	if logs.ChecksumMismatches < 2 {
		return fmt.Errorf("corrupted-chunk logs contain %d checksum mismatches, want at least 2 for two providers", logs.ChecksumMismatches)
	}
	if !logs.NoValidPeers {
		return errors.New("corrupted-chunk logs do not prove that both configured snapshot providers were exhausted")
	}
	return nil
}

func validateCometStateSyncFaultDeadline(elapsed, limit time.Duration, kind string) error {
	if limit < 10*time.Second || limit > maximumCometStateSyncFaultTimeout {
		return fmt.Errorf("%s failure limit must be within [10s,%s], got %s", kind, maximumCometStateSyncFaultTimeout, limit)
	}
	if elapsed > limit+cometStateSyncFaultDeadlineSlack {
		return fmt.Errorf(
			"%s failure exceeded bounded deadline: elapsed=%s limit=%s slack=%s",
			kind,
			elapsed,
			limit,
			cometStateSyncFaultDeadlineSlack,
		)
	}
	return nil
}
