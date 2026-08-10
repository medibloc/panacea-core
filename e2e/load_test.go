package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/gogoproto/proto"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	loadSeed                 int64  = 20260804
	loadWorkersPerBoundary          = 2
	loadIterationsPerWorker         = 15
	loadNormalQueryGasLimit  uint64 = 10_000_000
	loadReducedQueryGasLimit uint64 = 1_800_000
	loadDatasetClasses              = "maximum_classes"
	loadDatasetDenseNFTs            = "maximum_nfts_one_class"
	loadDatasetWideOwner            = "owner_across_classes"
)

type loadSuiteConfig struct {
	Seed                 int64          `json:"seed"`
	WorkersPerBoundary   int            `json:"workers_per_boundary"`
	IterationsPerWorker  int            `json:"iterations_per_worker"`
	RequestTimeout       time.Duration  `json:"request_timeout_nanoseconds"`
	NormalQueryGasLimit  uint64         `json:"normal_query_gas_limit"`
	ReducedQueryGasLimit uint64         `json:"reduced_query_gas_limit"`
	ClassCount           int            `json:"class_count"`
	DenseClassNFTCount   int            `json:"dense_class_nft_count"`
	WideOwnerClassCount  int            `json:"wide_owner_class_count"`
	SeedMintCount        int            `json:"seed_mint_count"`
	MetadataByteLimits   map[string]int `json:"metadata_byte_limits"`
	PerformanceSLA       string         `json:"performance_sla"`
}

type loadQueryEnvelope struct {
	Classes []json.RawMessage `json:"classes"`
	NFTs    []struct {
		ClassID string `json:"class_id"`
	} `json:"nfts"`
}

type loadQueryGasProbe struct {
	RecordedAt      time.Time `json:"recorded_at"`
	ConfiguredLimit uint64    `json:"configured_limit"`
	Boundary        string    `json:"boundary"`
	Dataset         string    `json:"dataset"`
	ExpectedOutcome string    `json:"expected_outcome"`
	Succeeded       bool      `json:"succeeded"`
	OutcomeMatched  bool      `json:"outcome_matched"`
	HTTPStatus      int       `json:"http_status,omitempty"`
	GRPCStatus      string    `json:"grpc_status,omitempty"`
	Error           string    `json:"error,omitempty"`
}

type loadResourceDelta struct {
	Node                  string `json:"node"`
	DBSizeBytes           int64  `json:"db_size_bytes"`
	BlockDeviceWriteBytes int64  `json:"block_device_write_bytes"`
}

func loadValidateCoreRuntimeMetrics(samples []harness.LoadNodeRuntimeSample) error {
	if len(samples) == 0 {
		return errors.New("runtime metrics contain no node samples")
	}
	var validationErrors []error
	for _, sample := range samples {
		node := sample.Node
		if strings.TrimSpace(node) == "" {
			node = "unknown node"
		}
		for _, metric := range []struct {
			name           string
			available      bool
			unavailableKey string
		}{
			{name: "catching-up observation", available: sample.CatchingUp != nil, unavailableKey: "status"},
			{name: "peer count", available: sample.Peers != nil, unavailableKey: "peers"},
			{name: "mempool transaction count", available: sample.MempoolTransactions != nil, unavailableKey: "mempool"},
			{name: "mempool byte count", available: sample.MempoolBytes != nil, unavailableKey: "mempool"},
			{name: "CPU", available: sample.CPUPercent != nil, unavailableKey: "docker_stats"},
			{name: "RSS", available: sample.RSSBytes != nil, unavailableKey: "docker_stats"},
			{name: "open files", available: sample.OpenFiles != nil, unavailableKey: "open_files"},
			{name: "DB size", available: sample.DBSizeBytes != nil, unavailableKey: "db_size_bytes"},
			{name: "block write bytes", available: sample.BlockDeviceWriteBytes != nil, unavailableKey: "docker_stats"},
			{name: "goroutines", available: sample.Goroutines != nil, unavailableKey: "goroutines"},
		} {
			if !metric.available {
				message := fmt.Sprintf("%s is missing %s", node, metric.name)
				if diagnostic := strings.TrimSpace(sample.Unavailable[metric.unavailableKey]); diagnostic != "" {
					message += fmt.Sprintf(": %s", diagnostic)
				}
				validationErrors = append(validationErrors, errors.New(message))
			}
		}
		if sample.Peers != nil && *sample.Peers < 0 {
			validationErrors = append(validationErrors, fmt.Errorf("%s has invalid peer count %d", node, *sample.Peers))
		}
		if sample.MempoolTransactions != nil && *sample.MempoolTransactions < 0 {
			validationErrors = append(validationErrors, fmt.Errorf(
				"%s has invalid mempool transaction count %d",
				node,
				*sample.MempoolTransactions,
			))
		}
		if sample.MempoolBytes != nil && *sample.MempoolBytes < 0 {
			validationErrors = append(validationErrors, fmt.Errorf(
				"%s has invalid mempool byte count %d",
				node,
				*sample.MempoolBytes,
			))
		}
	}
	return errors.Join(validationErrors...)
}

func loadValidateWorkloadHeightWindow(startHeight, workloadEndHeight int64) error {
	if workloadEndHeight <= startHeight {
		return fmt.Errorf(
			"validator height did not advance during load: start=%d end=%d",
			startHeight,
			workloadEndHeight,
		)
	}
	return nil
}

func TestFullNodeLoadAndResourceBaseline(t *testing.T) {
	if os.Getenv("PANACEA_E2E_LOAD") != "1" {
		t.Skip("set PANACEA_E2E_LOAD=1 to run the Docker load/resource baseline")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	requestTimeout := 10 * time.Second
	config := loadSuiteConfig{
		Seed:                 loadSeed,
		WorkersPerBoundary:   loadWorkersPerBoundary,
		IterationsPerWorker:  loadIterationsPerWorker,
		RequestTimeout:       requestTimeout,
		NormalQueryGasLimit:  loadNormalQueryGasLimit,
		ReducedQueryGasLimit: loadReducedQueryGasLimit,
		ClassCount:           loadClassCount,
		DenseClassNFTCount:   loadDenseNFTCount,
		WideOwnerClassCount:  loadWideOwnerClassCount,
		SeedMintCount:        len(loadSeedPlan()),
		MetadataByteLimits: map[string]int{
			"class_local_id":    loadMaximumLocalIDBytes,
			"class_name":        loadMaximumClassNameBytes,
			"class_symbol":      loadMaximumClassSymbolBytes,
			"class_description": loadMaximumDescriptionBytes,
			"uri":               loadMaximumURIBytes,
			"nft_data_encoded":  loadMaximumNFTDataBytes,
		},
		PerformanceSLA: "none; this suite records an observational release baseline",
	}

	network, err := harness.Start(ctx, t, harness.Config{
		Image:           harness.CurrentImage(),
		NumValidators:   1,
		NumFullNodes:    1,
		TimeoutCommit:   "1s",
		QueryGasLimit:   loadNormalQueryGasLimit,
		EnableTelemetry: true,
	})
	require.NoError(t, err)
	defer network.RecordTestPanic()
	require.NoError(t, network.WriteArtifactJSON("metrics/load-config.json", config))
	require.NoError(t, network.WriteArtifactJSON("metrics/environment.json", map[string]any{
		"recorded_at": time.Now().UTC(),
		"go_version":  runtime.Version(),
		"go_os":       runtime.GOOS,
		"go_arch":     runtime.GOARCH,
		"chain_id":    network.Chain.Config().ChainID,
		"image":       harness.CurrentImage(),
	}))

	creator := loadBuildAndFundWallet(t, ctx, network, "load-creator", "2000000000umed")
	denseOwner := loadBuildAndFundWallet(t, ctx, network, "load-dense-owner", "20000000umed")
	wideOwner := loadBuildAndFundWallet(t, ctx, network, "load-wide-owner", "20000000umed")
	mixedOwner := loadBuildAndFundWallet(t, ctx, network, "load-mixed-owner", "20000000umed")
	mixedRecipient := loadBuildAndFundWallet(t, ctx, network, "load-mixed-recipient", "20000000umed")

	classIDs := loadSeedClasses(t, ctx, network, creator)
	loadSeedNFTs(t, ctx, network, creator, denseOwner, wideOwner, classIDs)
	require.NoError(t, network.WriteArtifactJSON("metrics/dataset.json", map[string]any{
		"class_ids":          classIDs,
		"dense_class_id":     classIDs[0],
		"dense_owner":        denseOwner.FormattedAddress(),
		"wide_owner":         wideOwner.FormattedAddress(),
		"seed_plan":          loadSeedPlan(),
		"exact_class_count":  loadClassCount,
		"exact_dense_nfts":   loadDenseNFTCount,
		"exact_wide_classes": loadWideOwnerClassCount,
	}))
	loadAssertBoundaryDatasets(t, ctx, network, classIDs[0], wideOwner.FormattedAddress())

	runtimeBefore, err := network.CaptureLoadRuntimeMetrics(ctx, "before-load")
	require.NoError(t, err)
	require.NoError(t, loadValidateCoreRuntimeMetrics(runtimeBefore))
	validator := network.Chain.Validators[0]
	startHeight, err := validator.Height(ctx)
	require.NoError(t, err)

	querySamples, txSamples, workloadErr := loadRunConcurrentWorkload(
		ctx,
		network,
		config,
		classIDs,
		wideOwner.FormattedAddress(),
		creator,
		mixedOwner,
		mixedRecipient,
	)
	workloadEndHeight, heightErr := validator.Height(ctx)
	for _, sample := range querySamples {
		require.NoError(t, network.AppendArtifactJSON("metrics/query-samples.jsonl", sample))
	}
	for _, sample := range txSamples {
		require.NoError(t, network.AppendArtifactJSON("metrics/transactions.jsonl", sample))
	}
	require.NoError(t, workloadErr)
	require.NoError(t, heightErr)
	require.NoError(t, loadValidateWorkloadHeightWindow(startHeight, workloadEndHeight))
	require.NoError(t, network.WaitForFullNode(ctx, workloadEndHeight))

	blocks, err := network.CollectLoadBlockMetrics(ctx, validator, startHeight, workloadEndHeight)
	require.NoError(t, err)
	require.NotEmpty(t, blocks)
	querySummaries, err := loadSummarizeQueries(querySamples)
	require.NoError(t, err)
	txSummary, err := harness.SummarizeLoadTransactions(txSamples)
	require.NoError(t, err)
	require.Equal(t, 4, txSummary.Submitted)
	require.Equal(t, 4, txSummary.CheckTxAccepted)
	require.Equal(t, 4, txSummary.Committed)
	require.Zero(t, txSummary.Failed)
	require.NoError(t, network.WriteArtifactJSON("metrics/query-summary.json", querySummaries))
	require.NoError(t, network.WriteArtifactJSON("metrics/transaction-summary.json", txSummary))
	require.NoError(t, network.WriteArtifactJSON("metrics/block-summary.json", harness.SummarizeLoadBlocks(blocks)))

	runtimeAfter, err := network.CaptureLoadRuntimeMetrics(ctx, "after-load")
	require.NoError(t, err)
	require.NoError(t, loadValidateCoreRuntimeMetrics(runtimeAfter))
	require.NoError(t, network.WriteArtifactJSON("metrics/resource-delta.json", loadRuntimeDeltas(runtimeBefore, runtimeAfter)))

	loadExerciseReducedQueryGas(t, ctx, network, classIDs[0], wideOwner.FormattedAddress())
	postGasStart, err := validator.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, postGasStart+2))
	finalHeight, err := validator.Height(ctx)
	require.NoError(t, err)
	require.NoError(t, network.WaitForFullNode(ctx, finalHeight))
	finalRuntime, err := network.CaptureLoadRuntimeMetrics(ctx, "after-query-gas-restore")
	require.NoError(t, err)
	require.NoError(t, loadValidateCoreRuntimeMetrics(finalRuntime))
	for _, sample := range finalRuntime {
		require.NotNil(t, sample.Height, "%s stopped responding after load", sample.Node)
		require.NotNil(t, sample.CatchingUp, "%s did not expose final catching-up state", sample.Node)
		require.False(t, *sample.CatchingUp, "%s remained catching up", sample.Node)
	}
}

func loadSeedClasses(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	creator ibc.Wallet,
) []string {
	t.Helper()
	classIDs := make([]string, loadClassCount)
	for index := 0; index < loadClassCount; index++ {
		metadata := loadMaximumClassMetadata(index)
		maxSupply := "1"
		switch index {
		case 0:
			maxSupply = "100"
		case loadClassCount - 1:
			maxSupply = "2"
		}
		classIDs[index] = creator.FormattedAddress() + ":" + metadata.LocalID
		_, err := network.BroadcastAndWaitTx(
			ctx,
			fmt.Sprintf("load-create-class-%03d", index),
			network.Chain.Validators[0],
			creator.KeyName(),
			"nft", "create-class", metadata.LocalID, metadata.Name, metadata.Symbol,
			"owner-transferable", "true", maxSupply,
			"--description", metadata.Description,
			"--uri", metadata.URI,
			"--uri-hash", metadata.URIHash,
			"--gas", "800000",
			"--broadcast-mode", "sync",
		)
		require.NoError(t, err)
	}
	return classIDs
}

func loadSeedNFTs(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	creator ibc.Wallet,
	denseOwner ibc.Wallet,
	wideOwner ibc.Wallet,
	classIDs []string,
) {
	t.Helper()
	for sequence, mint := range loadSeedPlan() {
		prefix := "n"
		recipient := denseOwner.FormattedAddress()
		if mint.Owner == "wide" {
			prefix = "w"
			recipient = wideOwner.FormattedAddress()
		}
		metadata := loadMaximumNFTMetadata(prefix, mint.NFTIndex)
		_, err := network.BroadcastAndWaitTx(
			ctx,
			fmt.Sprintf("load-mint-seed-%03d", sequence),
			network.Chain.Validators[0],
			creator.KeyName(),
			"nft", "mint", classIDs[mint.ClassIndex], metadata.ID, recipient,
			"--uri", metadata.URI,
			"--uri-hash", metadata.URIHash,
			"--data", metadata.DataJSON,
			"--gas", "800000",
			"--broadcast-mode", "sync",
		)
		require.NoError(t, err)
	}
}

func loadAssertBoundaryDatasets(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	denseClassID string,
	wideOwner string,
) {
	t.Helper()
	page := &query.PageRequest{Limit: 100}
	classes, err := network.QueryNFTClassesGRPC(ctx, "load-verify-classes", &upstreamnft.QueryClassesRequest{Pagination: page})
	require.NoError(t, err)
	require.Len(t, classes.Classes, loadClassCount)
	dense, err := network.QueryNFTsGRPC(ctx, "load-verify-dense-nfts", &upstreamnft.QueryNFTsRequest{
		ClassId: denseClassID, Pagination: &query.PageRequest{Limit: 100},
	})
	require.NoError(t, err)
	require.Len(t, dense.Nfts, loadDenseNFTCount)
	wide, err := network.QueryNFTsGRPC(ctx, "load-verify-wide-owner", &upstreamnft.QueryNFTsRequest{
		Owner: wideOwner, Pagination: &query.PageRequest{Limit: 100},
	})
	require.NoError(t, err)
	require.Len(t, wide.Nfts, loadWideOwnerClassCount)
	classSet := make(map[string]struct{}, len(wide.Nfts))
	for _, nft := range wide.Nfts {
		classSet[nft.ClassId] = struct{}{}
	}
	require.Len(t, classSet, loadWideOwnerClassCount)
}

func loadRunConcurrentWorkload(
	ctx context.Context,
	network *harness.Network,
	config loadSuiteConfig,
	classIDs []string,
	wideOwner string,
	creator ibc.Wallet,
	mixedOwner ibc.Wallet,
	mixedRecipient ibc.Wallet,
) ([]harness.LoadQuerySample, []harness.LoadTxSample, error) {
	httpClient := &http.Client{Timeout: config.RequestTimeout}
	start := make(chan struct{})
	workDone := make(chan struct{})
	queryChannel := make(chan harness.LoadQuerySample, 2*config.WorkersPerBoundary*config.IterationsPerWorker)
	txChannel := make(chan harness.LoadTxSample, 4)
	errorChannel := make(chan error, 2*config.WorkersPerBoundary*config.IterationsPerWorker+8)
	var (
		errorCollector sync.WaitGroup
		workloadErrors []error
	)
	errorCollector.Add(1)
	go func() {
		defer errorCollector.Done()
		for err := range errorChannel {
			if err != nil {
				workloadErrors = append(workloadErrors, err)
			}
		}
	}()

	var work sync.WaitGroup
	for worker := 0; worker < config.WorkersPerBoundary; worker++ {
		work.Add(2)
		go func(worker int) {
			defer work.Done()
			<-start
			loadRunRESTWorker(ctx, network, httpClient, config, worker, classIDs[0], wideOwner, queryChannel)
		}(worker)
		go func(worker int) {
			defer work.Done()
			<-start
			loadRunGRPCWorker(ctx, network, config, worker, classIDs[0], wideOwner, queryChannel)
		}(worker)
	}
	work.Add(1)
	go func() {
		defer work.Done()
		<-start
		loadRunMixedWrites(ctx, network, classIDs[len(classIDs)-1], creator, mixedOwner, mixedRecipient, txChannel, errorChannel)
	}()

	var sampler sync.WaitGroup
	sampler.Add(1)
	go func() {
		defer sampler.Done()
		<-start
		iteration := 0
		for {
			label := fmt.Sprintf("during-load-%02d", iteration)
			samples, captureErr := network.CaptureLoadRuntimeMetrics(ctx, label)
			validationErr := loadValidateCoreRuntimeMetrics(samples)
			if err := errors.Join(captureErr, validationErr); err != nil {
				errorChannel <- fmt.Errorf("capture and validate runtime metrics %s: %w", label, err)
			}
			iteration++
			timer := time.NewTimer(time.Second)
			select {
			case <-workDone:
				if !timer.Stop() {
					<-timer.C
				}
				return
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				errorChannel <- ctx.Err()
				return
			case <-timer.C:
			}
		}
	}()

	close(start)
	work.Wait()
	close(workDone)
	sampler.Wait()
	close(queryChannel)
	close(txChannel)
	close(errorChannel)
	errorCollector.Wait()

	queries := make([]harness.LoadQuerySample, 0, cap(queryChannel))
	for sample := range queryChannel {
		queries = append(queries, sample)
	}
	txs := make([]harness.LoadTxSample, 0, 4)
	for sample := range txChannel {
		txs = append(txs, sample)
	}
	for _, sample := range queries {
		if !sample.Success {
			workloadErrors = append(workloadErrors, fmt.Errorf("%s %s: %s", sample.Boundary, sample.Dataset, sample.Error))
		}
	}
	return queries, txs, errors.Join(workloadErrors...)
}

func loadRunRESTWorker(
	ctx context.Context,
	network *harness.Network,
	httpClient *http.Client,
	config loadSuiteConfig,
	worker int,
	denseClassID string,
	wideOwner string,
	output chan<- harness.LoadQuerySample,
) {
	order := rand.New(rand.NewSource(config.Seed + int64(worker))).Perm(3)
	for iteration := 0; iteration < config.IterationsPerWorker; iteration++ {
		dataset := loadDatasetName(order[iteration%len(order)])
		requestCtx, cancel := context.WithTimeout(ctx, config.RequestTimeout)
		started := time.Now().UTC()
		body, err := network.FullNodeRESTGet(
			requestCtx,
			httpClient,
			fmt.Sprintf("load-rest-%02d-%03d", worker, iteration),
			loadRESTPath(dataset, denseClassID, wideOwner),
		)
		if err == nil {
			err = loadValidateRESTPayload(dataset, body)
		}
		finished := time.Now().UTC()
		cancel()
		output <- harness.LoadQuerySample{
			RequestID:     fmt.Sprintf("rest-%02d-%03d", worker, iteration),
			Boundary:      "rest",
			Dataset:       dataset,
			StartedAt:     started,
			FinishedAt:    finished,
			Success:       err == nil,
			TimedOut:      loadTimedOut(err),
			StatusCode:    loadRESTStatus(err),
			Status:        fmt.Sprintf("HTTP %d", loadRESTStatus(err)),
			ResponseBytes: len(body),
			Error:         loadErrorString(err),
		}
	}
}

func loadRunGRPCWorker(
	ctx context.Context,
	network *harness.Network,
	config loadSuiteConfig,
	worker int,
	denseClassID string,
	wideOwner string,
	output chan<- harness.LoadQuerySample,
) {
	order := rand.New(rand.NewSource(config.Seed + 1000 + int64(worker))).Perm(3)
	for iteration := 0; iteration < config.IterationsPerWorker; iteration++ {
		dataset := loadDatasetName(order[iteration%len(order)])
		requestCtx, cancel := context.WithTimeout(ctx, config.RequestTimeout)
		started := time.Now().UTC()
		responseBytes, err := loadGRPCQuery(
			requestCtx,
			network,
			fmt.Sprintf("load-grpc-%02d-%03d", worker, iteration),
			dataset,
			denseClassID,
			wideOwner,
		)
		finished := time.Now().UTC()
		cancel()
		output <- harness.LoadQuerySample{
			RequestID:     fmt.Sprintf("grpc-%02d-%03d", worker, iteration),
			Boundary:      "grpc",
			Dataset:       dataset,
			StartedAt:     started,
			FinishedAt:    finished,
			Success:       err == nil,
			TimedOut:      loadTimedOut(err),
			Status:        status.Code(err).String(),
			ResponseBytes: responseBytes,
			Error:         loadErrorString(err),
		}
	}
}

func loadGRPCQuery(
	ctx context.Context,
	network *harness.Network,
	step string,
	dataset string,
	denseClassID string,
	wideOwner string,
) (int, error) {
	switch dataset {
	case loadDatasetClasses:
		response, err := network.QueryNFTClassesGRPC(ctx, step, &upstreamnft.QueryClassesRequest{
			Pagination: &query.PageRequest{Limit: 100},
		})
		if err != nil {
			return 0, err
		}
		if len(response.Classes) != loadClassCount {
			return proto.Size(response), fmt.Errorf("classes returned %d records, want %d", len(response.Classes), loadClassCount)
		}
		return proto.Size(response), nil
	case loadDatasetDenseNFTs:
		response, err := network.QueryNFTsGRPC(ctx, step, &upstreamnft.QueryNFTsRequest{
			ClassId: denseClassID, Pagination: &query.PageRequest{Limit: 100},
		})
		if err != nil {
			return 0, err
		}
		if len(response.Nfts) != loadDenseNFTCount {
			return proto.Size(response), fmt.Errorf("dense class returned %d records, want %d", len(response.Nfts), loadDenseNFTCount)
		}
		return proto.Size(response), nil
	case loadDatasetWideOwner:
		response, err := network.QueryNFTsGRPC(ctx, step, &upstreamnft.QueryNFTsRequest{
			Owner: wideOwner, Pagination: &query.PageRequest{Limit: 100},
		})
		if err != nil {
			return 0, err
		}
		classes := make(map[string]struct{}, len(response.Nfts))
		for _, nft := range response.Nfts {
			classes[nft.ClassId] = struct{}{}
		}
		if len(response.Nfts) != loadWideOwnerClassCount || len(classes) != loadWideOwnerClassCount {
			return proto.Size(response), fmt.Errorf("wide owner returned %d NFTs across %d classes", len(response.Nfts), len(classes))
		}
		return proto.Size(response), nil
	default:
		return 0, fmt.Errorf("unknown load dataset %q", dataset)
	}
}

func loadRunMixedWrites(
	ctx context.Context,
	network *harness.Network,
	classID string,
	creator ibc.Wallet,
	mixedOwner ibc.Wallet,
	mixedRecipient ibc.Wallet,
	output chan<- harness.LoadTxSample,
	errorsOut chan<- error,
) {
	metadata := loadMaximumNFTMetadata("m", 1)
	operations := []struct {
		name    string
		signer  ibc.Wallet
		command []string
	}{
		{
			name:   "mint",
			signer: creator,
			command: []string{
				"nft", "mint", classID, metadata.ID, mixedOwner.FormattedAddress(),
				"--uri", metadata.URI, "--uri-hash", metadata.URIHash, "--data", metadata.DataJSON,
			},
		},
		{
			name: "send", signer: mixedOwner,
			command: []string{"nft", "send", classID, metadata.ID, mixedRecipient.FormattedAddress()},
		},
		{
			name: "revoke", signer: creator,
			command: []string{"nft", "revoke", classID, metadata.ID},
		},
		{
			name: "burn", signer: mixedRecipient,
			command: []string{"nft", "burn", classID, metadata.ID},
		},
	}
	for _, operation := range operations {
		command := append(append([]string(nil), operation.command...), "--gas", "800000", "--broadcast-mode", "sync")
		submittedAt := time.Now().UTC()
		lifecycle, err := network.BroadcastAndWaitTxLifecycle(
			ctx,
			"load-mixed-"+operation.name,
			network.Chain.Validators[0],
			operation.signer.KeyName(),
			command...,
		)
		sample, err := loadTxSampleFromLifecycle(
			operation.name,
			submittedAt,
			time.Now().UTC(),
			lifecycle,
			err,
		)
		output <- sample
		if err != nil {
			errorsOut <- err
			return
		}
	}
}

func loadTxSampleFromLifecycle(
	operation string,
	submittedAt time.Time,
	finishedAt time.Time,
	lifecycle *harness.TxLifecycleResult,
	txErr error,
) (harness.LoadTxSample, error) {
	sample := harness.LoadTxSample{
		Operation:   operation,
		SubmittedAt: submittedAt,
		FinishedAt:  finishedAt,
		Failed:      txErr != nil,
	}
	if lifecycle == nil {
		sample.Error = loadErrorString(txErr)
		return sample, txErr
	}

	sample.CheckTxAccepted = lifecycle.CheckTx != nil && lifecycle.CheckTx.Code == 0
	if lifecycle.Committed == nil {
		sample.Error = loadErrorString(txErr)
		return sample, txErr
	}

	committed := lifecycle.Committed
	sample.TxHash = committed.TxHash
	sample.Height = committed.HeightInt64()
	sample.Committed = sample.Height > 0
	if committed.Code != 0 {
		sample.Failed = true
	}
	gasWanted, gasUsed, gasErr := harness.DecodeTxGas(committed.Raw)
	if gasErr != nil {
		txErr = errors.Join(txErr, fmt.Errorf("decode %s transaction gas: %w", operation, gasErr))
		sample.Failed = true
	} else {
		sample.GasWanted = gasWanted
		sample.GasUsed = gasUsed
	}
	sample.Error = loadErrorString(txErr)
	return sample, txErr
}

func loadExerciseReducedQueryGas(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	denseClassID string,
	wideOwner string,
) {
	t.Helper()
	validator := network.Chain.Validators[0]
	validatorStart, err := validator.Height(ctx)
	require.NoError(t, err)
	override, err := network.ApplyFullNodeQueryGasLimit(ctx, loadReducedQueryGasLimit)
	require.NoError(t, err)
	t.Cleanup(func() {
		restoreCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		require.NoError(t, override.Restore(restoreCtx))
	})

	probes, err := loadRunReducedGasMixedPhase(ctx, network, denseClassID, wideOwner)
	require.NoError(t, err)

	postRejection, err := network.QueryNFTClassesGRPC(
		ctx,
		"load-low-gas-grpc-post-rejection-success",
		&upstreamnft.QueryClassesRequest{Pagination: &query.PageRequest{Limit: 1}},
	)
	postProbe := loadQueryGasProbe{
		RecordedAt: time.Now().UTC(), ConfiguredLimit: loadReducedQueryGasLimit,
		Boundary: "grpc", Dataset: loadDatasetClasses, ExpectedOutcome: "post_rejection_success", Succeeded: err == nil,
		OutcomeMatched: err == nil && len(postRejection.GetClasses()) == 1,
		GRPCStatus:     status.Code(err).String(), Error: loadErrorString(err),
	}
	probes = append(probes, postProbe)
	require.NoError(t, network.AppendArtifactJSON("metrics/query-gas-probes.jsonl", postProbe))
	require.NoError(t, err)
	require.Len(t, postRejection.Classes, 1)
	require.NoError(t, network.WaitForNodeHeight(ctx, validator, validatorStart+2))
	validatorEnd, err := validator.Height(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, validatorEnd, validatorStart+2)
	require.NoError(t, network.WriteArtifactJSON("metrics/query-gas-probes.json", probes))
	require.NoError(t, network.WriteArtifactJSON("metrics/query-gas-validator-progress.json", map[string]any{
		"start_height": validatorStart,
		"end_height":   validatorEnd,
	}))

	require.NoError(t, override.Restore(ctx))
	ownerResponse, err := network.QueryNFTsGRPC(ctx, "load-restored-gas-owner-success", &upstreamnft.QueryNFTsRequest{
		Owner: wideOwner, Pagination: &query.PageRequest{Limit: 100},
	})
	require.NoError(t, err)
	require.Len(t, ownerResponse.Nfts, loadWideOwnerClassCount)
}

func loadRunReducedGasMixedPhase(
	ctx context.Context,
	network *harness.Network,
	denseClassID string,
	wideOwner string,
) ([]loadQueryGasProbe, error) {
	const normalLimit = 1

	type operation struct {
		name string
		run  func() loadQueryGasProbe
	}
	httpClient := &http.Client{Timeout: 10 * time.Second}
	grpcClient := upstreamnft.NewQueryClient(network.Chain.FullNodes[0].GrpcConn)
	operations := []operation{
		{
			name: "rest-normal",
			run: func() loadQueryGasProbe {
				path := "/cosmos/nft/v1beta1/classes?" + url.Values{
					"pagination.limit": {fmt.Sprint(normalLimit)},
				}.Encode()
				body, err := network.FullNodeRESTGet(ctx, httpClient, "load-low-gas-rest-normal", path)
				if err == nil {
					var response loadQueryEnvelope
					if decodeErr := json.Unmarshal(body, &response); decodeErr != nil {
						err = fmt.Errorf("decode normal REST query: %w", decodeErr)
					} else if len(response.Classes) != normalLimit {
						err = fmt.Errorf("normal REST query returned %d classes, want %d", len(response.Classes), normalLimit)
					}
				}
				return loadQueryGasProbe{
					RecordedAt: time.Now().UTC(), ConfiguredLimit: loadReducedQueryGasLimit,
					Boundary: "rest", Dataset: loadDatasetClasses, ExpectedOutcome: "success",
					Succeeded: err == nil, OutcomeMatched: err == nil,
					HTTPStatus: loadRESTStatus(err), Error: loadErrorString(err),
				}
			},
		},
		{
			name: "grpc-normal",
			run: func() loadQueryGasProbe {
				response, err := grpcClient.Classes(ctx, &upstreamnft.QueryClassesRequest{
					Pagination: &query.PageRequest{Limit: normalLimit},
				})
				if err == nil && len(response.GetClasses()) != normalLimit {
					err = fmt.Errorf("normal gRPC query returned %d classes, want %d", len(response.GetClasses()), normalLimit)
				}
				return loadQueryGasProbe{
					RecordedAt: time.Now().UTC(), ConfiguredLimit: loadReducedQueryGasLimit,
					Boundary: "grpc", Dataset: loadDatasetClasses, ExpectedOutcome: "success",
					Succeeded: err == nil, OutcomeMatched: err == nil,
					GRPCStatus: status.Code(err).String(), Error: loadErrorString(err),
				}
			},
		},
		{
			name: "rest-over-gas",
			run: func() loadQueryGasProbe {
				_, err := network.FullNodeRESTGet(
					ctx,
					httpClient,
					"load-low-gas-rest-over-limit",
					loadRESTPath(loadDatasetWideOwner, denseClassID, wideOwner),
				)
				return loadQueryGasProbe{
					RecordedAt: time.Now().UTC(), ConfiguredLimit: loadReducedQueryGasLimit,
					Boundary: "rest", Dataset: loadDatasetWideOwner, ExpectedOutcome: "query_gas_rejected",
					Succeeded: err == nil, OutcomeMatched: err != nil && loadIsGasError(err),
					HTTPStatus: loadRESTStatus(err), Error: loadErrorString(err),
				}
			},
		},
		{
			name: "grpc-over-gas",
			run: func() loadQueryGasProbe {
				_, err := grpcClient.NFTs(ctx, &upstreamnft.QueryNFTsRequest{
					Owner: wideOwner, Pagination: &query.PageRequest{Limit: 100},
				})
				return loadQueryGasProbe{
					RecordedAt: time.Now().UTC(), ConfiguredLimit: loadReducedQueryGasLimit,
					Boundary: "grpc", Dataset: loadDatasetWideOwner, ExpectedOutcome: "query_gas_rejected",
					Succeeded: err == nil, OutcomeMatched: err != nil && loadIsGasError(err),
					GRPCStatus: status.Code(err).String(), Error: loadErrorString(err),
				}
			},
		},
	}

	start := make(chan struct{})
	results := make(chan loadQueryGasProbe, len(operations))
	var work sync.WaitGroup
	for _, operation := range operations {
		operation := operation
		work.Add(1)
		go func() {
			defer work.Done()
			<-start
			results <- operation.run()
		}()
	}
	close(start)
	work.Wait()
	close(results)

	probes := make([]loadQueryGasProbe, 0, len(operations))
	var artifactErrors []error
	for probe := range results {
		probes = append(probes, probe)
		if err := network.AppendArtifactJSON("metrics/query-gas-probes.jsonl", probe); err != nil {
			artifactErrors = append(artifactErrors, err)
		}
	}
	return probes, errors.Join(errors.Join(artifactErrors...), loadValidateGasProbeBatch(probes))
}

func loadValidateGasProbeBatch(probes []loadQueryGasProbe) error {
	required := map[string]bool{
		"rest success":            false,
		"rest query_gas_rejected": false,
		"grpc success":            false,
		"grpc query_gas_rejected": false,
	}
	var validationErrors []error
	for _, probe := range probes {
		key := probe.Boundary + " " + probe.ExpectedOutcome
		if _, ok := required[key]; ok {
			required[key] = true
		}
		if !probe.OutcomeMatched {
			validationErrors = append(validationErrors, fmt.Errorf(
				"%s %s outcome did not match: %s",
				probe.Boundary,
				probe.ExpectedOutcome,
				probe.Error,
			))
		}
	}
	for key, observed := range required {
		if !observed {
			validationErrors = append(validationErrors, fmt.Errorf("mixed query-gas phase is missing %s", key))
		}
	}
	return errors.Join(validationErrors...)
}

func loadDatasetName(index int) string {
	return []string{loadDatasetClasses, loadDatasetDenseNFTs, loadDatasetWideOwner}[index]
}

func loadRESTPath(dataset, denseClassID, wideOwner string) string {
	parameters := url.Values{"pagination.limit": {"100"}}
	switch dataset {
	case loadDatasetClasses:
		return "/cosmos/nft/v1beta1/classes?" + parameters.Encode()
	case loadDatasetDenseNFTs:
		parameters.Set("class_id", denseClassID)
	case loadDatasetWideOwner:
		parameters.Set("owner", wideOwner)
	}
	return "/cosmos/nft/v1beta1/nfts?" + parameters.Encode()
}

func loadValidateRESTPayload(dataset string, body []byte) error {
	var envelope loadQueryEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("decode %s REST response: %w", dataset, err)
	}
	switch dataset {
	case loadDatasetClasses:
		if len(envelope.Classes) != loadClassCount {
			return fmt.Errorf("classes REST response has %d records, want %d", len(envelope.Classes), loadClassCount)
		}
	case loadDatasetDenseNFTs:
		if len(envelope.NFTs) != loadDenseNFTCount {
			return fmt.Errorf("dense REST response has %d records, want %d", len(envelope.NFTs), loadDenseNFTCount)
		}
	case loadDatasetWideOwner:
		classes := make(map[string]struct{}, len(envelope.NFTs))
		for _, nft := range envelope.NFTs {
			classes[nft.ClassID] = struct{}{}
		}
		if len(envelope.NFTs) != loadWideOwnerClassCount || len(classes) != loadWideOwnerClassCount {
			return fmt.Errorf("wide-owner REST response has %d NFTs across %d classes", len(envelope.NFTs), len(classes))
		}
	default:
		return fmt.Errorf("unknown load dataset %q", dataset)
	}
	return nil
}

func loadSummarizeQueries(samples []harness.LoadQuerySample) (map[string]harness.LoadQuerySummary, error) {
	groups := map[string][]harness.LoadQuerySample{"all": samples}
	for _, sample := range samples {
		key := sample.Boundary + "/" + sample.Dataset
		groups[key] = append(groups[key], sample)
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	summaries := make(map[string]harness.LoadQuerySummary, len(groups))
	for _, key := range keys {
		summary, err := harness.SummarizeLoadQueries(groups[key])
		if err != nil {
			return nil, fmt.Errorf("summarize %s: %w", key, err)
		}
		summaries[key] = summary
	}
	return summaries, nil
}

func loadRuntimeDeltas(before, after []harness.LoadNodeRuntimeSample) []loadResourceDelta {
	beforeByNode := make(map[string]harness.LoadNodeRuntimeSample, len(before))
	for _, sample := range before {
		beforeByNode[sample.Node] = sample
	}
	deltas := make([]loadResourceDelta, 0, len(after))
	for _, sample := range after {
		previous := beforeByNode[sample.Node]
		delta := loadResourceDelta{Node: sample.Node}
		if previous.DBSizeBytes != nil && sample.DBSizeBytes != nil {
			delta.DBSizeBytes = int64(*sample.DBSizeBytes) - int64(*previous.DBSizeBytes)
		}
		if previous.BlockDeviceWriteBytes != nil && sample.BlockDeviceWriteBytes != nil {
			delta.BlockDeviceWriteBytes = int64(*sample.BlockDeviceWriteBytes) - int64(*previous.BlockDeviceWriteBytes)
		}
		deltas = append(deltas, delta)
	}
	return deltas
}

func loadBuildAndFundWallet(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	keyName string,
	amount string,
) ibc.Wallet {
	t.Helper()
	wallet, err := network.BuildWallet(ctx, keyName, "")
	require.NoError(t, err)
	_, err = network.BroadcastAndWaitTx(
		ctx,
		"load-fund-"+keyName,
		network.Chain.Validators[0],
		interchaintest.FaucetAccountKeyName,
		"bank", "send", interchaintest.FaucetAccountKeyName, wallet.FormattedAddress(), amount,
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	return wallet
}

func loadRESTStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	words := strings.Fields(err.Error())
	for index := 0; index+1 < len(words); index++ {
		if words[index] != "HTTP" {
			continue
		}
		var statusCode int
		if _, scanErr := fmt.Sscanf(words[index+1], "%d", &statusCode); scanErr == nil {
			return statusCode
		}
	}
	return 0
}

func loadIsGasError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	humanReadableGasError := strings.Contains(message, "out of gas") ||
		strings.Contains(message, "gas limit") ||
		strings.Contains(message, "query gas")
	// Cosmos SDK query execution currently forwards the recovered
	// storetypes.ErrorOutOfGas struct using its default rendering. The
	// maximum-size owner query exhausts gas while reading a value, so gRPC
	// exposes Internal/{ReadPerByte} and the REST gateway maps it to HTTP 500.
	sdkReadGasPanic := strings.Contains(message, "{readperbyte}") &&
		(status.Code(err) == codes.Internal || loadRESTStatus(err) == http.StatusInternalServerError)
	return humanReadableGasError || sdkReadGasPanic
}

func loadTimedOut(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) || status.Code(err) == codes.DeadlineExceeded {
		return true
	}
	var networkError net.Error
	return errors.As(err, &networkError) && networkError.Timeout()
}

func loadErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
