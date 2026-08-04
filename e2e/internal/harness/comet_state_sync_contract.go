package harness

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
)

const (
	defaultCometStateSyncTrustHeightLag       = int64(2)
	defaultCometStateSyncTrustPeriod          = 168 * time.Hour
	defaultCometStateSyncDiscoveryTime        = 5 * time.Second
	defaultCometStateSyncChunkRequestTimeout  = 10 * time.Second
	defaultCometStateSyncChunkFetchers        = int32(4)
	defaultCometStateSyncProviderWaitTimeout  = 30 * time.Second
	defaultCometStateSyncCompletionTimeout    = 2 * time.Minute
	defaultCometStateSyncBadTrustHashTimeout  = 30 * time.Second
	maximumCometStateSyncBadTrustHashTimeout  = 90 * time.Second
	cometStateSyncBadTrustHashDeadlineSlack   = 6 * time.Second
	cometStateSyncMaximumMatchedArtifactLines = 64
)

var (
	stateSyncLogHeightPattern = regexp.MustCompile(`(?i)height["']?\s*[:=]\s*"?([0-9]+)`)
	terminalANSIPattern       = regexp.MustCompile(`\x1b\[[0-9;]*[[:alpha:]]`)
)

type cometStateSyncPlanOptions struct {
	TrustPeriod         time.Duration
	DiscoveryTime       time.Duration
	ChunkRequestTimeout time.Duration
	ChunkFetchers       int32
}

type cometStateSyncPlan struct {
	RPCServers          []string      `json:"rpc_servers"`
	TrustHeight         int64         `json:"trust_height"`
	TrustHash           string        `json:"trust_hash"`
	TrustPeriod         time.Duration `json:"-"`
	DiscoveryTime       time.Duration `json:"-"`
	ChunkRequestTimeout time.Duration `json:"-"`
	ChunkFetchers       int32         `json:"chunk_fetchers"`
}

// CometStateSyncConfigEvidence is the exact [statesync] configuration written
// to the detached node's config.toml. String duration fields match CometBFT's
// persisted configuration rather than Go's JSON duration representation.
type CometStateSyncConfigEvidence struct {
	Enabled             bool     `json:"enabled"`
	RPCServers          []string `json:"rpc_servers"`
	TrustHeight         int64    `json:"trust_height"`
	TrustHash           string   `json:"trust_hash"`
	TrustPeriod         string   `json:"trust_period"`
	DiscoveryTime       string   `json:"discovery_time"`
	ChunkRequestTimeout string   `json:"chunk_request_timeout"`
	ChunkFetchers       int32    `json:"chunk_fetchers"`
}

func newCometStateSyncPlan(
	rpcServers []string,
	trustHeight int64,
	trustHash string,
	options cometStateSyncPlanOptions,
) (cometStateSyncPlan, error) {
	if len(rpcServers) != 2 {
		return cometStateSyncPlan{}, fmt.Errorf("CometBFT state sync requires exactly two RPC servers, got %d", len(rpcServers))
	}
	servers := make([]string, len(rpcServers))
	seenServers := make(map[string]struct{}, len(rpcServers))
	for index, server := range rpcServers {
		server = strings.TrimSpace(server)
		if !strings.HasPrefix(server, "http://") && !strings.HasPrefix(server, "https://") {
			return cometStateSyncPlan{}, fmt.Errorf("state-sync RPC server %d must use http:// or https://: %q", index, server)
		}
		if _, duplicate := seenServers[server]; duplicate {
			return cometStateSyncPlan{}, fmt.Errorf("state-sync RPC server %q is duplicated", server)
		}
		seenServers[server] = struct{}{}
		servers[index] = server
	}
	if trustHeight <= 0 {
		return cometStateSyncPlan{}, fmt.Errorf("state-sync trust height must be positive, got %d", trustHeight)
	}
	trustHash = strings.ToUpper(strings.TrimSpace(trustHash))
	decodedHash, err := hex.DecodeString(trustHash)
	if err != nil {
		return cometStateSyncPlan{}, fmt.Errorf("decode state-sync trust hash: %w", err)
	}
	if len(decodedHash) != 32 {
		return cometStateSyncPlan{}, fmt.Errorf("state-sync trust hash must be 32 bytes, got %d", len(decodedHash))
	}

	if options.TrustPeriod == 0 {
		options.TrustPeriod = defaultCometStateSyncTrustPeriod
	}
	if options.DiscoveryTime == 0 {
		options.DiscoveryTime = defaultCometStateSyncDiscoveryTime
	}
	if options.ChunkRequestTimeout == 0 {
		options.ChunkRequestTimeout = defaultCometStateSyncChunkRequestTimeout
	}
	if options.ChunkFetchers == 0 {
		options.ChunkFetchers = defaultCometStateSyncChunkFetchers
	}
	if options.TrustPeriod <= 0 {
		return cometStateSyncPlan{}, errors.New("state-sync trust period must be positive")
	}
	if options.DiscoveryTime < 5*time.Second {
		return cometStateSyncPlan{}, errors.New("state-sync discovery time must be at least 5s")
	}
	if options.ChunkRequestTimeout < 5*time.Second {
		return cometStateSyncPlan{}, errors.New("state-sync chunk request timeout must be at least 5s")
	}
	if options.ChunkFetchers <= 0 {
		return cometStateSyncPlan{}, errors.New("state-sync chunk fetchers must be positive")
	}

	return cometStateSyncPlan{
		RPCServers:          servers,
		TrustHeight:         trustHeight,
		TrustHash:           trustHash,
		TrustPeriod:         options.TrustPeriod,
		DiscoveryTime:       options.DiscoveryTime,
		ChunkRequestTimeout: options.ChunkRequestTimeout,
		ChunkFetchers:       options.ChunkFetchers,
	}, nil
}

func (plan cometStateSyncPlan) configEvidence() CometStateSyncConfigEvidence {
	return CometStateSyncConfigEvidence{
		Enabled:             true,
		RPCServers:          append([]string(nil), plan.RPCServers...),
		TrustHeight:         plan.TrustHeight,
		TrustHash:           plan.TrustHash,
		TrustPeriod:         plan.TrustPeriod.String(),
		DiscoveryTime:       plan.DiscoveryTime.String(),
		ChunkRequestTimeout: plan.ChunkRequestTimeout.String(),
		ChunkFetchers:       plan.ChunkFetchers,
	}
}

func (plan cometStateSyncPlan) tomlOverride() testutil.Toml {
	return testutil.Toml{
		"statesync": testutil.Toml{
			"enable":                true,
			"rpc_servers":           strings.Join(plan.RPCServers, ","),
			"trust_height":          plan.TrustHeight,
			"trust_hash":            plan.TrustHash,
			"trust_period":          plan.TrustPeriod.String(),
			"discovery_time":        plan.DiscoveryTime.String(),
			"chunk_request_timeout": plan.ChunkRequestTimeout.String(),
			"chunk_fetchers":        strconv.FormatInt(int64(plan.ChunkFetchers), 10),
		},
	}
}

func validateRenderedCometStateSyncConfig(contents []byte, plan cometStateSyncPlan) error {
	tree, err := toml.LoadBytes(contents)
	if err != nil {
		return fmt.Errorf("parse rendered state-sync config.toml: %w", err)
	}
	expected := plan.configEvidence()
	checks := []struct {
		key  string
		want any
	}{
		{key: "statesync.enable", want: expected.Enabled},
		{key: "statesync.rpc_servers", want: strings.Join(expected.RPCServers, ",")},
		{key: "statesync.trust_height", want: expected.TrustHeight},
		{key: "statesync.trust_hash", want: expected.TrustHash},
		{key: "statesync.trust_period", want: expected.TrustPeriod},
		{key: "statesync.discovery_time", want: expected.DiscoveryTime},
		{key: "statesync.chunk_request_timeout", want: expected.ChunkRequestTimeout},
	}
	for _, check := range checks {
		got := tree.Get(check.key)
		if fmt.Sprint(got) != fmt.Sprint(check.want) {
			return fmt.Errorf("rendered %s=%v, want %v", check.key, got, check.want)
		}
	}
	gotFetchers, err := strconv.ParseInt(fmt.Sprint(tree.Get("statesync.chunk_fetchers")), 10, 32)
	if err != nil {
		return fmt.Errorf("parse rendered statesync.chunk_fetchers: %w", err)
	}
	if int32(gotFetchers) != expected.ChunkFetchers {
		return fmt.Errorf("rendered statesync.chunk_fetchers=%d, want %d", gotFetchers, expected.ChunkFetchers)
	}
	return nil
}

func validateRenderedCometStateSyncPEXDisabled(contents []byte) error {
	tree, err := toml.LoadBytes(contents)
	if err != nil {
		return fmt.Errorf("parse rendered state-sync config.toml for PEX: %w", err)
	}
	if got := tree.Get("p2p.pex"); got != false {
		return fmt.Errorf("rendered p2p.pex=%v, want false", got)
	}
	return nil
}

func mutateCometStateSyncTrustHash(trustHash string) (string, error) {
	trustHash = strings.ToUpper(strings.TrimSpace(trustHash))
	decoded, err := hex.DecodeString(trustHash)
	if err != nil {
		return "", fmt.Errorf("decode trust hash before mutation: %w", err)
	}
	if len(decoded) != 32 {
		return "", fmt.Errorf("trust hash must be 32 bytes before mutation, got %d", len(decoded))
	}
	if decoded[0] == 0 {
		decoded[0] = 1
	} else {
		decoded[0] ^= 1
	}
	mutated := strings.ToUpper(hex.EncodeToString(decoded))
	if strings.EqualFold(mutated, trustHash) {
		return "", errors.New("trust hash mutation did not change the hash")
	}
	return mutated, nil
}

func parseStateSyncDataInventory(nodeHome string, output []byte) ([]string, error) {
	cleanHome := path.Clean(nodeHome)
	if !path.IsAbs(cleanHome) || cleanHome != nodeHome {
		return nil, fmt.Errorf("state-sync node home must be a clean absolute path: %q", nodeHome)
	}
	dataRoot := path.Join(cleanHome, "data")
	prefix := dataRoot + "/"
	var inventory []string
	for _, line := range strings.Split(string(output), "\n") {
		candidate := strings.TrimSpace(line)
		if candidate == "" {
			continue
		}
		cleaned := path.Clean(candidate)
		if !strings.HasPrefix(cleaned, prefix) {
			return nil, fmt.Errorf("fresh state-sync inventory path escapes data directory: %q", candidate)
		}
		relative := strings.TrimPrefix(cleaned, cleanHome+"/")
		inventory = append(inventory, relative)
	}
	sort.Strings(inventory)
	for _, entry := range inventory {
		// panacead init writes only this height-zero safety file. Any database,
		// snapshot, WAL, or other entry means this was not an empty bootstrap.
		if entry != "data/priv_validator_state.json" {
			return inventory, fmt.Errorf("fresh state-sync data directory contains pre-existing state: %s", entry)
		}
	}
	return inventory, nil
}

// CometStateSyncLogEvidence retains the required CometBFT state-sync markers.
// All markers must be present; accepting app snapshot/import or ordinary block
// sync as an equivalent path is intentionally impossible.
type CometStateSyncLogEvidence struct {
	DiscoveredSnapshot bool     `json:"discovered_snapshot"`
	AcceptedSnapshot   bool     `json:"accepted_snapshot"`
	FetchedChunks      int      `json:"fetched_chunks"`
	AppliedChunks      int      `json:"applied_chunks"`
	VerifiedABCIApp    bool     `json:"verified_abci_app"`
	RestoredSnapshot   bool     `json:"restored_snapshot"`
	SnapshotHeight     int64    `json:"snapshot_height"`
	MatchedLines       []string `json:"matched_lines"`
}

func parseCometStateSyncLogs(contents []byte) CometStateSyncLogEvidence {
	var evidence CometStateSyncLogEvidence
	for _, rawLine := range strings.Split(string(contents), "\n") {
		line := plainStateSyncLogLine(rawLine)
		if line == "" {
			continue
		}
		matched := false
		switch {
		case strings.Contains(line, "Discovered new snapshot"):
			evidence.DiscoveredSnapshot = true
			matched = true
		case strings.Contains(line, "Snapshot accepted, restoring"):
			evidence.AcceptedSnapshot = true
			matched = true
		case strings.Contains(line, "Fetching snapshot chunk"):
			evidence.FetchedChunks++
			matched = true
		case strings.Contains(line, "Applied snapshot chunk to ABCI app"):
			evidence.AppliedChunks++
			matched = true
		case strings.Contains(line, "Verified ABCI app"):
			evidence.VerifiedABCIApp = true
			matched = true
		case strings.Contains(line, "Snapshot restored"):
			evidence.RestoredSnapshot = true
			matched = true
			if height, ok := stateSyncHeightFromLogLine(line); ok {
				evidence.SnapshotHeight = height
			}
		}
		if matched && len(evidence.MatchedLines) < cometStateSyncMaximumMatchedArtifactLines {
			evidence.MatchedLines = append(evidence.MatchedLines, boundedLine(line))
		}
	}
	return evidence
}

func (evidence CometStateSyncLogEvidence) validate() error {
	var missing []string
	if !evidence.DiscoveredSnapshot {
		missing = append(missing, "Discovered new snapshot")
	}
	if !evidence.AcceptedSnapshot {
		missing = append(missing, "Snapshot accepted, restoring")
	}
	if evidence.FetchedChunks < 1 {
		missing = append(missing, "Fetching snapshot chunk")
	}
	if evidence.AppliedChunks < 1 {
		missing = append(missing, "Applied snapshot chunk to ABCI app")
	}
	if !evidence.VerifiedABCIApp {
		missing = append(missing, "Verified ABCI app")
	}
	if !evidence.RestoredSnapshot {
		missing = append(missing, "Snapshot restored")
	}
	if evidence.SnapshotHeight <= 1 {
		missing = append(missing, "snapshot height greater than 1")
	}
	if len(missing) > 0 {
		return fmt.Errorf("state-sync log evidence is incomplete; missing %s", strings.Join(missing, ", "))
	}
	return nil
}

// CometStateSyncProviderEvidence proves that a selected peer completed at
// least one Cosmos SDK state snapshot and therefore could serve chunks.
type CometStateSyncProviderEvidence struct {
	Node                     string    `json:"node"`
	Since                    time.Time `json:"since"`
	CompletedSnapshotHeights []int64   `json:"completed_snapshot_heights,omitempty"`
	MatchedLines             []string  `json:"matched_lines,omitempty"`
	LogError                 string    `json:"log_error,omitempty"`
	RawLogArtifact           string    `json:"raw_log_artifact"`
}

func parseCometStateSyncProviderLogs(node string, since time.Time, artifactPath string, contents []byte) CometStateSyncProviderEvidence {
	evidence := CometStateSyncProviderEvidence{Node: node, Since: since, RawLogArtifact: artifactPath}
	heights := make(map[int64]struct{})
	for _, rawLine := range strings.Split(string(contents), "\n") {
		line := plainStateSyncLogLine(rawLine)
		if !strings.Contains(strings.ToLower(line), "completed state snapshot") {
			continue
		}
		if len(evidence.MatchedLines) < cometStateSyncMaximumMatchedArtifactLines {
			evidence.MatchedLines = append(evidence.MatchedLines, boundedLine(line))
		}
		if height, ok := stateSyncHeightFromLogLine(line); ok && height > 0 {
			heights[height] = struct{}{}
		}
	}
	for height := range heights {
		evidence.CompletedSnapshotHeights = append(evidence.CompletedSnapshotHeights, height)
	}
	sort.Slice(evidence.CompletedSnapshotHeights, func(i, j int) bool {
		return evidence.CompletedSnapshotHeights[i] < evidence.CompletedSnapshotHeights[j]
	})
	return evidence
}

func usableProviderSnapshotHeight(providers []CometStateSyncProviderEvidence, commonLatest int64) (int64, bool) {
	var selected int64
	for _, provider := range providers {
		for _, height := range provider.CompletedSnapshotHeights {
			// CometBFT's state provider verifies H, H+1, and H+2.
			if height > 1 && height+2 <= commonLatest && height > selected {
				selected = height
			}
		}
	}
	return selected, selected > 0
}

// CometStateSyncBadTrustLogEvidence proves that the light client rejected the
// deliberately mutated hash, and that no snapshot restore succeeded.
type CometStateSyncBadTrustLogEvidence struct {
	RejectedTrustHash bool     `json:"rejected_trust_hash"`
	UnexpectedSuccess bool     `json:"unexpected_success"`
	MatchedLines      []string `json:"matched_lines,omitempty"`
}

func parseCometStateSyncBadTrustLogs(contents []byte) CometStateSyncBadTrustLogEvidence {
	var evidence CometStateSyncBadTrustLogEvidence
	for _, rawLine := range strings.Split(string(contents), "\n") {
		line := plainStateSyncLogLine(rawLine)
		lower := strings.ToLower(line)
		if strings.Contains(lower, "expected header's hash") && strings.Contains(lower, "but got") ||
			strings.Contains(lower, "trusted header hash") && strings.Contains(lower, "does not match") {
			evidence.RejectedTrustHash = true
			if len(evidence.MatchedLines) < cometStateSyncMaximumMatchedArtifactLines {
				evidence.MatchedLines = append(evidence.MatchedLines, boundedLine(line))
			}
		}
		if strings.Contains(line, "Snapshot restored") || strings.Contains(line, "Verified ABCI app") {
			evidence.UnexpectedSuccess = true
		}
	}
	return evidence
}

func validateCometStateSyncBadTrustFailure(
	logs CometStateSyncBadTrustLogEvidence,
	elapsed time.Duration,
	limit time.Duration,
) error {
	if limit <= 0 || limit > maximumCometStateSyncBadTrustHashTimeout {
		return fmt.Errorf("bad trust-hash failure limit must be within (0,%s], got %s", maximumCometStateSyncBadTrustHashTimeout, limit)
	}
	if elapsed > limit+cometStateSyncBadTrustHashDeadlineSlack {
		return fmt.Errorf("bad trust-hash failure exceeded bounded deadline: elapsed=%s limit=%s slack=%s", elapsed, limit, cometStateSyncBadTrustHashDeadlineSlack)
	}
	if logs.UnexpectedSuccess {
		return errors.New("bad trust-hash node unexpectedly restored and verified a snapshot")
	}
	if !logs.RejectedTrustHash {
		return errors.New("bad trust-hash node logs do not prove light-client hash rejection")
	}
	return nil
}

func stateSyncHeightFromLogLine(line string) (int64, bool) {
	matches := stateSyncLogHeightPattern.FindStringSubmatch(line)
	if len(matches) != 2 {
		return 0, false
	}
	height, err := strconv.ParseInt(matches[1], 10, 64)
	return height, err == nil
}

func boundedLine(line string) string {
	const maximum = 4 << 10
	if len(line) <= maximum {
		return line
	}
	return line[:maximum] + "...[truncated]"
}

func plainStateSyncLogLine(line string) string {
	return strings.TrimSpace(terminalANSIPattern.ReplaceAllString(line, ""))
}

func validateCometStateSyncQueryCommand(command []string) error {
	if len(command) == 0 {
		return errors.New("state-sync query command is required")
	}
	for _, argument := range command {
		if strings.TrimSpace(argument) == "" {
			return errors.New("state-sync query command contains an empty argument")
		}
		for _, harnessFlag := range []string{"--height", "--node", "--home", "--output"} {
			if argument == harnessFlag || strings.HasPrefix(argument, harnessFlag+"=") {
				return fmt.Errorf("state-sync query command must not override harness flag %s", harnessFlag)
			}
		}
	}
	return nil
}

func validateCometStateSyncQueryContinuity(before, after []byte, kind string) error {
	beforeSemantic, err := NewSemanticJSON(before)
	if err != nil {
		return fmt.Errorf("canonicalize %s query before restart: %w", kind, err)
	}
	afterSemantic, err := NewSemanticJSON(after)
	if err != nil {
		return fmt.Errorf("canonicalize %s query after restart: %w", kind, err)
	}
	if !bytes.Equal(beforeSemantic, afterSemantic) {
		return fmt.Errorf("%s query changed across state-sync node restart: before=%s after=%s", kind, beforeSemantic, afterSemantic)
	}
	return nil
}
