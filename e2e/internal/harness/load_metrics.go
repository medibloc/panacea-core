package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// LoadQuerySample is one REST or gRPC observation from the resource baseline.
// Absolute timestamps make the raw artifact independently useful; summaries
// are derived from them without imposing an SLA.
type LoadQuerySample struct {
	RequestID     string    `json:"request_id"`
	Boundary      string    `json:"boundary"`
	Dataset       string    `json:"dataset"`
	StartedAt     time.Time `json:"started_at"`
	FinishedAt    time.Time `json:"finished_at"`
	Success       bool      `json:"success"`
	TimedOut      bool      `json:"timed_out"`
	StatusCode    int       `json:"status_code,omitempty"`
	Status        string    `json:"status,omitempty"`
	ResponseBytes int       `json:"response_bytes,omitempty"`
	Error         string    `json:"error,omitempty"`
}

// LoadTxSample records the CheckTx and committed execution lifecycle for one
// write submitted while queries are in flight.
type LoadTxSample struct {
	Operation       string    `json:"operation"`
	SubmittedAt     time.Time `json:"submitted_at"`
	FinishedAt      time.Time `json:"finished_at"`
	CheckTxAccepted bool      `json:"check_tx_accepted"`
	Committed       bool      `json:"committed"`
	Failed          bool      `json:"failed"`
	TxHash          string    `json:"tx_hash,omitempty"`
	Height          int64     `json:"height,omitempty"`
	GasWanted       int64     `json:"gas_wanted,omitempty"`
	GasUsed         int64     `json:"gas_used,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// LoadBlockSample is an exact committed-block observation from the validator
// during the full-node query window.
type LoadBlockSample struct {
	Height               int64     `json:"height"`
	Time                 time.Time `json:"time"`
	IntervalMilliseconds float64   `json:"interval_milliseconds"`
	Transactions         int       `json:"transactions"`
	FailedTransactions   int       `json:"failed_transactions"`
	GasWanted            int64     `json:"gas_wanted"`
	GasUsed              int64     `json:"gas_used"`
	CommitSignatures     int       `json:"commit_signatures"`
	MissedSignatures     int       `json:"missed_signatures"`
}

// LoadLatencyDistribution uses nearest-rank percentiles over observed wall
// time. Values are milliseconds so JSON consumers need not infer Go duration
// units.
type LoadLatencyDistribution struct {
	Minimum float64 `json:"minimum"`
	P50     float64 `json:"p50"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
	Maximum float64 `json:"maximum"`
}

// LoadQuerySummary is a measurement, not a pass/fail threshold.
type LoadQuerySummary struct {
	Requests            int                     `json:"requests"`
	Successes           int                     `json:"successes"`
	Failures            int                     `json:"failures"`
	Timeouts            int                     `json:"timeouts"`
	WindowMilliseconds  int64                   `json:"window_milliseconds"`
	ThroughputPerSecond float64                 `json:"throughput_per_second"`
	LatencyMilliseconds LoadLatencyDistribution `json:"latency_milliseconds"`
}

// LoadTxSummary preserves the four lifecycle counts requested by the goal and
// aggregates committed gas without asserting an arbitrary performance SLA.
type LoadTxSummary struct {
	Submitted                 int                     `json:"submitted"`
	CheckTxAccepted           int                     `json:"check_tx_accepted"`
	Committed                 int                     `json:"committed"`
	Failed                    int                     `json:"failed"`
	GasWanted                 int64                   `json:"gas_wanted"`
	GasUsed                   int64                   `json:"gas_used"`
	CommitLatencyMilliseconds LoadLatencyDistribution `json:"commit_latency_milliseconds"`
}

// LoadBlockSummary aggregates raw block evidence without a timing threshold.
type LoadBlockSummary struct {
	Blocks               int                     `json:"blocks"`
	Transactions         int                     `json:"transactions"`
	FailedTransactions   int                     `json:"failed_transactions"`
	GasWanted            int64                   `json:"gas_wanted"`
	GasUsed              int64                   `json:"gas_used"`
	MissedSignatures     int                     `json:"missed_signatures"`
	IntervalMilliseconds LoadLatencyDistribution `json:"interval_milliseconds"`
}

// SummarizeLoadQueries derives counts, throughput, and latency percentiles
// from raw observations. It rejects invalid clocks rather than silently
// emitting misleading metrics.
func SummarizeLoadQueries(samples []LoadQuerySample) (LoadQuerySummary, error) {
	var summary LoadQuerySummary
	if len(samples) == 0 {
		return summary, nil
	}

	latencies := make([]time.Duration, 0, len(samples))
	windowStart := samples[0].StartedAt
	windowEnd := samples[0].FinishedAt
	for index, sample := range samples {
		latency, err := observedDuration(sample.StartedAt, sample.FinishedAt)
		if err != nil {
			return LoadQuerySummary{}, fmt.Errorf("query sample %d %w", index, err)
		}
		latencies = append(latencies, latency)
		if sample.StartedAt.Before(windowStart) {
			windowStart = sample.StartedAt
		}
		if sample.FinishedAt.After(windowEnd) {
			windowEnd = sample.FinishedAt
		}
		if sample.Success {
			summary.Successes++
		} else {
			summary.Failures++
		}
		if sample.TimedOut {
			summary.Timeouts++
		}
	}

	summary.Requests = len(samples)
	window := windowEnd.Sub(windowStart)
	summary.WindowMilliseconds = window.Milliseconds()
	if window > 0 {
		summary.ThroughputPerSecond = float64(len(samples)) / window.Seconds()
	}
	summary.LatencyMilliseconds = summarizeDurations(latencies)
	return summary, nil
}

// SummarizeLoadTransactions reports lifecycle counts and commit latency over
// only transactions that reached a committed result.
func SummarizeLoadTransactions(samples []LoadTxSample) (LoadTxSummary, error) {
	var summary LoadTxSummary
	latencies := make([]time.Duration, 0, len(samples))
	for index, sample := range samples {
		latency, err := observedDuration(sample.SubmittedAt, sample.FinishedAt)
		if err != nil {
			return LoadTxSummary{}, fmt.Errorf("transaction sample %d %w", index, err)
		}
		summary.Submitted++
		if sample.CheckTxAccepted {
			summary.CheckTxAccepted++
		}
		if sample.Committed {
			summary.Committed++
			latencies = append(latencies, latency)
			summary.GasWanted += sample.GasWanted
			summary.GasUsed += sample.GasUsed
		}
		if sample.Failed || !sample.Committed {
			summary.Failed++
		}
	}
	summary.CommitLatencyMilliseconds = summarizeDurations(latencies)
	return summary, nil
}

// SummarizeLoadBlocks aggregates block gas, transaction, interval, and missed
// signature evidence collected from the validator.
func SummarizeLoadBlocks(samples []LoadBlockSample) LoadBlockSummary {
	var summary LoadBlockSummary
	intervals := make([]time.Duration, 0, len(samples))
	for _, sample := range samples {
		summary.Blocks++
		summary.Transactions += sample.Transactions
		summary.FailedTransactions += sample.FailedTransactions
		summary.GasWanted += sample.GasWanted
		summary.GasUsed += sample.GasUsed
		summary.MissedSignatures += sample.MissedSignatures
		if sample.IntervalMilliseconds > 0 {
			intervals = append(intervals, time.Duration(sample.IntervalMilliseconds*float64(time.Millisecond)))
		}
	}
	summary.IntervalMilliseconds = summarizeDurations(intervals)
	return summary
}

func observedDuration(startedAt, finishedAt time.Time) (time.Duration, error) {
	if startedAt.IsZero() || finishedAt.IsZero() {
		return 0, errors.New("must have non-zero start and finish timestamps")
	}
	if finishedAt.Before(startedAt) {
		return 0, errors.New("finishes before it starts")
	}
	return finishedAt.Sub(startedAt), nil
}

func summarizeDurations(values []time.Duration) LoadLatencyDistribution {
	if len(values) == 0 {
		return LoadLatencyDistribution{}
	}
	sorted := append([]time.Duration(nil), values...)
	sort.Slice(sorted, func(left, right int) bool { return sorted[left] < sorted[right] })
	toMilliseconds := func(value time.Duration) float64 {
		return float64(value) / float64(time.Millisecond)
	}
	nearestRank := func(percentile float64) float64 {
		rank := int(math.Ceil(percentile*float64(len(sorted)))) - 1
		if rank < 0 {
			rank = 0
		}
		return toMilliseconds(sorted[rank])
	}
	return LoadLatencyDistribution{
		Minimum: toMilliseconds(sorted[0]),
		P50:     nearestRank(0.50),
		P95:     nearestRank(0.95),
		P99:     nearestRank(0.99),
		Maximum: toMilliseconds(sorted[len(sorted)-1]),
	}
}

// DecodeTxGas accepts the string-or-number JSON representation emitted by
// Cosmos SDK query-tx versions and returns the committed gas evidence.
func DecodeTxGas(raw json.RawMessage) (gasWanted, gasUsed int64, err error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var response map[string]any
	if decodeErr := decoder.Decode(&response); decodeErr != nil {
		return 0, 0, fmt.Errorf("decode transaction gas JSON: %w", decodeErr)
	}
	gasWanted, err = jsonInt64(response, "gas_wanted")
	if err != nil {
		return 0, 0, err
	}
	gasUsed, err = jsonInt64(response, "gas_used")
	if err != nil {
		return 0, 0, err
	}
	return gasWanted, gasUsed, nil
}

func jsonInt64(values map[string]any, key string) (int64, error) {
	value, ok := values[key]
	if !ok {
		return 0, fmt.Errorf("transaction result is missing %s", key)
	}
	switch typed := value.(type) {
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("decode %s: %w", key, err)
		}
		return parsed, nil
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return 0, fmt.Errorf("decode %s: %w", key, err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("transaction result %s has unsupported type %T", key, value)
	}
}

// RewriteQueryGasLimit updates exactly one top-level app.toml setting while
// preserving every other byte. The original value is returned for evidence.
func RewriteQueryGasLimit(contents []byte, newLimit uint64) ([]byte, uint64, error) {
	lines := strings.Split(string(contents), "\n")
	matches := 0
	var previous uint64
	for index, line := range lines {
		uncommented, comment, _ := strings.Cut(line, "#")
		left, value, found := strings.Cut(uncommented, "=")
		if !found || strings.TrimSpace(left) != "query-gas-limit" {
			continue
		}
		matches++
		parsed, parseErr := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		if parseErr != nil {
			return nil, 0, fmt.Errorf("query-gas-limit must be numeric: %w", parseErr)
		}
		previous = parsed
		prefix := uncommented[:strings.Index(uncommented, "=")+1]
		trailingWhitespace := value[len(strings.TrimRight(value, " \t")):]
		lines[index] = prefix + " " + strconv.FormatUint(newLimit, 10) + trailingWhitespace
		if comment != "" {
			lines[index] += "#" + comment
		}
	}
	if matches != 1 {
		return nil, 0, fmt.Errorf("app.toml must contain exactly one query-gas-limit setting, found %d", matches)
	}
	return []byte(strings.Join(lines, "\n")), previous, nil
}
