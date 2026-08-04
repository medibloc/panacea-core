package e2e_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	loadClassCount              = 100
	loadDenseNFTCount           = 100
	loadWideOwnerClassCount     = 100
	loadMaximumLocalIDBytes     = 64
	loadMaximumClassNameBytes   = 128
	loadMaximumClassSymbolBytes = 32
	loadMaximumDescriptionBytes = 1024
	loadMaximumURIBytes         = 256
	loadMaximumNFTDataBytes     = 1024
	loadMaximumNFTDataTextBytes = loadMaximumNFTDataBytes - 3
	loadBasicNFTDataTypeURL     = "/panacea.nft.v1.BasicNFTData"
)

type loadClassMetadata struct {
	LocalID     string `json:"local_id"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	URI         string `json:"uri"`
	URIHash     string `json:"uri_hash"`
}

type loadNFTMetadata struct {
	ID       string `json:"id"`
	URI      string `json:"uri"`
	URIHash  string `json:"uri_hash"`
	DataJSON string `json:"data_json"`
}

type loadSeedMint struct {
	ClassIndex int    `json:"class_index"`
	NFTIndex   int    `json:"nft_index"`
	Owner      string `json:"owner"`
}

func loadValidRuntimeSample() harness.LoadNodeRuntimeSample {
	cpu := 1.5
	rss := uint64(1)
	openFiles := 2
	dbSize := uint64(3)
	writes := uint64(4)
	goroutines := uint64(5)
	peers := 0
	catchingUp := false
	mempoolTransactions := 0
	mempoolBytes := int64(0)
	return harness.LoadNodeRuntimeSample{
		Node:                  "full-node",
		CPUPercent:            &cpu,
		RSSBytes:              &rss,
		OpenFiles:             &openFiles,
		DBSizeBytes:           &dbSize,
		BlockDeviceWriteBytes: &writes,
		Goroutines:            &goroutines,
		Peers:                 &peers,
		CatchingUp:            &catchingUp,
		MempoolTransactions:   &mempoolTransactions,
		MempoolBytes:          &mempoolBytes,
	}
}

func loadMaximumClassMetadata(index int) loadClassMetadata {
	return loadClassMetadata{
		LocalID:     loadMaximumIdentifier("c", index),
		Name:        strings.Repeat("n", loadMaximumClassNameBytes),
		Symbol:      strings.Repeat("s", loadMaximumClassSymbolBytes),
		Description: strings.Repeat("d", loadMaximumDescriptionBytes),
		URI:         strings.Repeat("u", loadMaximumURIBytes),
		URIHash:     "sha256:" + strings.Repeat("a", 64),
	}
}

func loadMaximumNFTMetadata(prefix string, index int) loadNFTMetadata {
	data, err := json.Marshal(map[string]string{
		"@type":       loadBasicNFTDataTypeURL,
		"description": strings.Repeat("d", loadMaximumNFTDataTextBytes),
	})
	if err != nil {
		panic(err)
	}
	return loadNFTMetadata{
		ID:       loadMaximumIdentifier(prefix, index),
		URI:      strings.Repeat("u", loadMaximumURIBytes),
		URIHash:  "sha256:" + strings.Repeat("b", 64),
		DataJSON: string(data),
	}
}

func loadMaximumIdentifier(prefix string, index int) string {
	suffix := fmt.Sprintf("%03d", index)
	return strings.Repeat(prefix, loadMaximumLocalIDBytes-len(suffix)) + suffix
}

// loadSeedPlan creates exactly 100 NFTs in class zero and exactly one NFT
// owned by the wide owner in each of 100 classes. The other 99 dense-class
// NFTs have a distinct owner, yielding 199 setup mints without polluting either
// exact query boundary.
func loadSeedPlan() []loadSeedMint {
	plan := make([]loadSeedMint, 0, loadDenseNFTCount+loadWideOwnerClassCount-1)
	for nftIndex := 0; nftIndex < loadDenseNFTCount-1; nftIndex++ {
		plan = append(plan, loadSeedMint{ClassIndex: 0, NFTIndex: nftIndex, Owner: "dense"})
	}
	for classIndex := 0; classIndex < loadWideOwnerClassCount; classIndex++ {
		plan = append(plan, loadSeedMint{ClassIndex: classIndex, NFTIndex: classIndex, Owner: "wide"})
	}
	return plan
}

func TestLoadBoundaryMetadataMatchesProtocolByteLimits(t *testing.T) {
	t.Parallel()

	class := loadMaximumClassMetadata(99)
	require.Len(t, class.LocalID, loadMaximumLocalIDBytes)
	require.Len(t, class.Name, loadMaximumClassNameBytes)
	require.Len(t, class.Symbol, loadMaximumClassSymbolBytes)
	require.Len(t, class.Description, loadMaximumDescriptionBytes)
	require.Len(t, class.URI, loadMaximumURIBytes)
	require.Len(t, class.URIHash, len("sha256:")+64)
	require.Regexp(t, `^[a-z0-9._-]+$`, class.LocalID)
	require.Regexp(t, `^sha256:[0-9a-f]{64}$`, class.URIHash)

	nft := loadMaximumNFTMetadata("n", 99)
	require.Len(t, nft.ID, loadMaximumLocalIDBytes)
	require.Len(t, nft.URI, loadMaximumURIBytes)
	require.Regexp(t, `^[a-z0-9._-]+$`, nft.ID)
	require.Regexp(t, `^sha256:[0-9a-f]{64}$`, nft.URIHash)
	// BasicNFTData.description is protobuf field 2: one tag byte, a
	// two-byte varint length for 1021, and 1021 payload bytes.
	require.Equal(t, loadMaximumNFTDataBytes, 1+2+loadMaximumNFTDataTextBytes)
	var decoded map[string]string
	require.NoError(t, json.Unmarshal([]byte(nft.DataJSON), &decoded))
	require.Equal(t, loadBasicNFTDataTypeURL, decoded["@type"])
	require.Len(t, decoded["description"], loadMaximumNFTDataTextBytes)
}

func TestLoadSeedPlanHasExactBoundaryCardinality(t *testing.T) {
	t.Parallel()

	plan := loadSeedPlan()
	require.Len(t, plan, 199)
	denseCount := 0
	wideCount := 0
	wideClasses := make(map[int]struct{})
	for _, mint := range plan {
		if mint.ClassIndex == 0 {
			denseCount++
		}
		if mint.Owner == "wide" {
			wideCount++
			wideClasses[mint.ClassIndex] = struct{}{}
		}
	}
	require.Equal(t, loadDenseNFTCount, denseCount)
	require.Equal(t, loadWideOwnerClassCount, wideCount)
	require.Len(t, wideClasses, loadWideOwnerClassCount)
}

func TestLoadValidateCoreRuntimeMetricsRequiresOperationalMeasurements(t *testing.T) {
	samples := []harness.LoadNodeRuntimeSample{loadValidRuntimeSample()}
	require.NoError(t, loadValidateCoreRuntimeMetrics(samples))

	samples[0].RSSBytes = nil
	require.ErrorContains(t, loadValidateCoreRuntimeMetrics(samples), "RSS")
	samples[0].RSSBytes = loadValidRuntimeSample().RSSBytes
	samples[0].Goroutines = nil
	require.ErrorContains(t, loadValidateCoreRuntimeMetrics(samples), "goroutines")
}

func TestLoadValidateCoreRuntimeMetricsRequiresValidPeerCount(t *testing.T) {
	sample := loadValidRuntimeSample()
	require.NoError(t, loadValidateCoreRuntimeMetrics([]harness.LoadNodeRuntimeSample{sample}))

	sample.Peers = nil
	require.ErrorContains(t, loadValidateCoreRuntimeMetrics([]harness.LoadNodeRuntimeSample{sample}), "peer count")
	negativePeers := -1
	sample.Peers = &negativePeers
	require.ErrorContains(t, loadValidateCoreRuntimeMetrics([]harness.LoadNodeRuntimeSample{sample}), "peer count")
}

func TestLoadValidateCoreRuntimeMetricsRequiresCatchingUpObservation(t *testing.T) {
	sample := loadValidRuntimeSample()
	require.NoError(t, loadValidateCoreRuntimeMetrics([]harness.LoadNodeRuntimeSample{sample}))

	sample.CatchingUp = nil
	require.ErrorContains(t, loadValidateCoreRuntimeMetrics([]harness.LoadNodeRuntimeSample{sample}), "catching-up")
}

func TestLoadValidateCoreRuntimeMetricsRequiresValidMempoolObservation(t *testing.T) {
	valid := loadValidRuntimeSample()
	require.NoError(t, loadValidateCoreRuntimeMetrics([]harness.LoadNodeRuntimeSample{valid}))

	tests := []struct {
		name   string
		mutate func(*harness.LoadNodeRuntimeSample)
		want   string
	}{
		{
			name:   "missing transaction count",
			mutate: func(sample *harness.LoadNodeRuntimeSample) { sample.MempoolTransactions = nil },
			want:   "mempool transaction count",
		},
		{
			name: "negative transaction count",
			mutate: func(sample *harness.LoadNodeRuntimeSample) {
				value := -1
				sample.MempoolTransactions = &value
			},
			want: "mempool transaction count",
		},
		{
			name:   "missing byte count",
			mutate: func(sample *harness.LoadNodeRuntimeSample) { sample.MempoolBytes = nil },
			want:   "mempool byte count",
		},
		{
			name: "negative byte count",
			mutate: func(sample *harness.LoadNodeRuntimeSample) {
				value := int64(-1)
				sample.MempoolBytes = &value
			},
			want: "mempool byte count",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sample := valid
			test.mutate(&sample)
			require.ErrorContains(t, loadValidateCoreRuntimeMetrics([]harness.LoadNodeRuntimeSample{sample}), test.want)
		})
	}
}

func TestLoadValidateCoreRuntimeMetricsPreservesCaptureDiagnostics(t *testing.T) {
	sample := loadValidRuntimeSample()
	sample.Peers = nil
	sample.Unavailable = map[string]string{
		"peers": "net_info RPC timed out",
	}

	err := loadValidateCoreRuntimeMetrics([]harness.LoadNodeRuntimeSample{sample})
	require.ErrorContains(t, err, "peer count")
	require.ErrorContains(t, err, "net_info RPC timed out")
}

func TestLoadValidateWorkloadHeightWindowRequiresProgressDuringLoad(t *testing.T) {
	t.Parallel()

	require.NoError(t, loadValidateWorkloadHeightWindow(10, 11))
	require.ErrorContains(t, loadValidateWorkloadHeightWindow(10, 10), "did not advance")
	require.ErrorContains(t, loadValidateWorkloadHeightWindow(10, 9), "did not advance")
}

func TestLoadValidateGasProbeBatchRequiresMixedBoundariesAndOutcomes(t *testing.T) {
	t.Parallel()

	valid := []loadQueryGasProbe{
		{Boundary: "rest", ExpectedOutcome: "success", OutcomeMatched: true},
		{Boundary: "rest", ExpectedOutcome: "query_gas_rejected", OutcomeMatched: true},
		{Boundary: "grpc", ExpectedOutcome: "success", OutcomeMatched: true},
		{Boundary: "grpc", ExpectedOutcome: "query_gas_rejected", OutcomeMatched: true},
	}
	require.NoError(t, loadValidateGasProbeBatch(valid))

	require.ErrorContains(t, loadValidateGasProbeBatch(valid[:3]), "grpc query_gas_rejected")
	invalid := append([]loadQueryGasProbe(nil), valid...)
	invalid[0].OutcomeMatched = false
	require.ErrorContains(t, loadValidateGasProbeBatch(invalid), "did not match")
}

func TestLoadIsGasErrorRecognizesCosmosSDKQueryGasPanics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "REST gateway",
			err:  fmt.Errorf(`REST query load-low-gas-rest-over-limit returned HTTP 500: {"code":13, "message":"{ReadPerByte}", "details":[]}`),
		},
		{
			name: "gRPC",
			err:  status.Error(codes.Internal, "{ReadPerByte}"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.True(t, loadIsGasError(test.err))
		})
	}
}

func TestLoadIsGasErrorRejectsUnrelatedInternalFailures(t *testing.T) {
	t.Parallel()

	require.False(t, loadIsGasError(nil))
	require.False(t, loadIsGasError(fmt.Errorf("rpc error: code = Internal desc = database unavailable")))
	require.False(t, loadIsGasError(fmt.Errorf("REST query returned HTTP 500: read failed")))
	require.False(t, loadIsGasError(fmt.Errorf("{ReadPerByte}")))
}
