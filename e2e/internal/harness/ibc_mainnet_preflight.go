package harness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	osmosisMainnetRPCDefault  = "https://rpc.osmosis.zone"
	osmosisMainnetRESTDefault = "https://lcd.osmosis.zone"

	osmosisMainnetRPCEnv  = "PANACEA_E2E_OSMOSIS_MAINNET_RPC_ENDPOINT"
	osmosisMainnetRESTEnv = "PANACEA_E2E_OSMOSIS_MAINNET_REST_ENDPOINT"

	osmosisMainnetChainID            = "osmosis-1"
	osmosisMainnetPanaceaChainID     = "panacea-3"
	osmosisMainnetChannelID          = "channel-82"
	osmosisMainnetCounterpartyID     = "channel-1"
	osmosisMainnetConnectionID       = "connection-1231"
	osmosisMainnetRegistryCommit     = "a294eea176868b0bf82545050b6f8f76577f406a"
	osmosisMainnetChannelRegistryURL = "https://github.com/cosmos/chain-registry/blob/" +
		osmosisMainnetRegistryCommit + "/_IBC/osmosis-panacea.json"

	osmosisMainnetPreflightArtifactPath = "ibc/mainnet-preflight.json"
	osmosisMainnetPreflightSchema       = "panacea.osmosis-mainnet-preflight/v1"
	osmosisMainnetPreflightTimeout      = 45 * time.Second
	osmosisMainnetRequestTimeout        = 8 * time.Second
	osmosisMainnetMaxResponseBytes      = 4 << 20
	osmosisMainnetMaxBlockAge           = 10 * time.Minute
	osmosisNodeInfoDependencyScope      = "original-module-path-version-only"
)

// OsmosisMainnetPreflightConfig permits an operator to select explicit
// read-only endpoints while keeping chain, release, channel, and middleware
// fixtures immutable. There is deliberately no skip or fallback endpoint.
type OsmosisMainnetPreflightConfig struct {
	RPCEndpoint  string        `json:"rpc_endpoint"`
	RESTEndpoint string        `json:"rest_endpoint"`
	Timeout      time.Duration `json:"timeout"`
}

func resolveOsmosisMainnetPreflightConfig() (OsmosisMainnetPreflightConfig, error) {
	rpcEndpoint := strings.TrimSpace(os.Getenv(osmosisMainnetRPCEnv))
	if rpcEndpoint == "" {
		rpcEndpoint = osmosisMainnetRPCDefault
	}
	restEndpoint := strings.TrimSpace(os.Getenv(osmosisMainnetRESTEnv))
	if restEndpoint == "" {
		restEndpoint = osmosisMainnetRESTDefault
	}
	config := OsmosisMainnetPreflightConfig{
		RPCEndpoint:  rpcEndpoint,
		RESTEndpoint: restEndpoint,
		Timeout:      osmosisMainnetPreflightTimeout,
	}
	if err := config.validate(); err != nil {
		return OsmosisMainnetPreflightConfig{}, err
	}
	return config, nil
}

func (c OsmosisMainnetPreflightConfig) validate() error {
	var validationErrors []error
	for name, endpoint := range map[string]string{
		"RPC":  c.RPCEndpoint,
		"REST": c.RESTEndpoint,
	} {
		parsed, err := url.Parse(strings.TrimSpace(endpoint))
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			validationErrors = append(validationErrors, fmt.Errorf("Osmosis mainnet %s endpoint %q is not an absolute URL", name, endpoint))
			continue
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			validationErrors = append(validationErrors, fmt.Errorf("Osmosis mainnet %s endpoint scheme %q is not HTTP(S)", name, parsed.Scheme))
		}
		if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			validationErrors = append(validationErrors, fmt.Errorf("Osmosis mainnet %s endpoint must not contain credentials, query, or fragment", name))
		}
	}
	if c.Timeout <= 0 || c.Timeout > time.Minute {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis mainnet preflight timeout %s must be within (0, 1m]", c.Timeout))
	}
	return errors.Join(validationErrors...)
}

type OsmosisMainnetExpectedFixture struct {
	ChainID                  string                   `json:"chain_id"`
	CounterpartyChainID      string                   `json:"counterparty_chain_id"`
	ChannelRegistryReference string                   `json:"channel_registry_reference"`
	ChannelID                string                   `json:"channel_id"`
	CounterpartyChannelID    string                   `json:"counterparty_channel_id"`
	ConnectionID             string                   `json:"connection_id"`
	PortID                   string                   `json:"port_id"`
	State                    string                   `json:"state"`
	Ordering                 string                   `json:"ordering"`
	Version                  string                   `json:"version"`
	RegistryStatus           string                   `json:"registry_status"`
	RegistryPreferred        bool                     `json:"registry_preferred"`
	Binary                   IBCBinaryVersionContract `json:"binary"`
}

func expectedOsmosisMainnetFixture() OsmosisMainnetExpectedFixture {
	return OsmosisMainnetExpectedFixture{
		ChainID:                  osmosisMainnetChainID,
		CounterpartyChainID:      osmosisMainnetPanaceaChainID,
		ChannelRegistryReference: osmosisMainnetChannelRegistryURL,
		ChannelID:                osmosisMainnetChannelID,
		CounterpartyChannelID:    osmosisMainnetCounterpartyID,
		ConnectionID:             osmosisMainnetConnectionID,
		PortID:                   "transfer",
		State:                    "STATE_OPEN",
		Ordering:                 "ORDER_UNORDERED",
		Version:                  "ics20-1",
		RegistryStatus:           "ACTIVE",
		RegistryPreferred:        true,
		Binary:                   pinnedOsmosisBinaryContract(),
	}
}

type OsmosisMainnetStatusEvidence struct {
	Available         bool      `json:"available"`
	Network           string    `json:"network,omitempty"`
	CometBFTVersion   string    `json:"cometbft_version,omitempty"`
	LatestBlockHeight int64     `json:"latest_block_height,omitempty"`
	LatestBlockTime   time.Time `json:"latest_block_time,omitempty"`
	CatchingUp        bool      `json:"catching_up"`
	Error             string    `json:"error,omitempty"`
}

type OsmosisMainnetNodeInfoEvidence struct {
	Available                   bool                     `json:"available"`
	Network                     string                   `json:"network,omitempty"`
	CometBFT                    string                   `json:"cometbft_version,omitempty"`
	DependencyObservationScope  string                   `json:"dependency_observation_scope"`
	ReplacementMetadataObserved bool                     `json:"replacement_metadata_observed"`
	Binary                      IBCBinaryVersionIdentity `json:"binary"`
	Error                       string                   `json:"error,omitempty"`
}

type OsmosisMainnetChannelEvidence struct {
	InvestigationStatus   string `json:"investigation_status"`
	ChannelID             string `json:"channel_id,omitempty"`
	PortID                string `json:"port_id,omitempty"`
	State                 string `json:"state,omitempty"`
	Ordering              string `json:"ordering,omitempty"`
	Version               string `json:"version,omitempty"`
	CounterpartyPortID    string `json:"counterparty_port_id,omitempty"`
	CounterpartyChannelID string `json:"counterparty_channel_id,omitempty"`
	ConnectionID          string `json:"connection_id,omitempty"`
	ProofHeight           string `json:"proof_height,omitempty"`
	Error                 string `json:"error,omitempty"`
}

type OsmosisMainnetPreflightEvidence struct {
	SchemaVersion string                             `json:"schema_version"`
	StartedAt     time.Time                          `json:"started_at"`
	CompletedAt   time.Time                          `json:"completed_at"`
	Config        OsmosisMainnetPreflightConfig      `json:"config"`
	Expected      OsmosisMainnetExpectedFixture      `json:"expected"`
	Status        OsmosisMainnetStatusEvidence       `json:"status"`
	NodeInfo      OsmosisMainnetNodeInfoEvidence     `json:"node_info"`
	Channel       OsmosisMainnetChannelEvidence      `json:"channel"`
	Middleware    IBCMiddlewareInvestigationEvidence `json:"middleware"`
	Passed        bool                               `json:"passed"`
	Error         string                             `json:"error,omitempty"`
}

func (e OsmosisMainnetPreflightEvidence) Validate() error {
	var validationErrors []error
	if e.SchemaVersion != osmosisMainnetPreflightSchema || e.StartedAt.IsZero() || e.CompletedAt.IsZero() {
		validationErrors = append(validationErrors, errors.New("Osmosis mainnet preflight metadata is incomplete"))
	}
	if err := e.Config.validate(); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if !expectedMainnetFixtureEqual(e.Expected, expectedOsmosisMainnetFixture()) {
		validationErrors = append(validationErrors, errors.New("Osmosis mainnet expected fixture was mutated"))
	}
	if !e.Status.Available || e.Status.Network != osmosisMainnetChainID || e.Status.CatchingUp ||
		e.Status.LatestBlockHeight < 1 || !sameCometBFTMinorLine(e.Status.CometBFTVersion, PinnedIBCProvenance().Osmosis.CometBFTVersion) {
		validationErrors = append(validationErrors, errors.New("Osmosis mainnet status did not confirm the pinned live chain"))
	}
	if !e.NodeInfo.Available || e.NodeInfo.Network != osmosisMainnetChainID ||
		e.NodeInfo.CometBFT != strings.TrimPrefix(PinnedIBCProvenance().Osmosis.CometBFTVersion, "v") {
		validationErrors = append(validationErrors, errors.New("Osmosis mainnet node-info did not confirm the pinned live chain"))
	}
	if e.NodeInfo.DependencyObservationScope != osmosisNodeInfoDependencyScope || e.NodeInfo.ReplacementMetadataObserved {
		validationErrors = append(validationErrors, errors.New("Osmosis node-info dependency observation scope is overstated"))
	}
	if err := validateIBCBinaryVersionIdentity(e.NodeInfo.Binary, osmosisNodeInfoObservableContract()); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis mainnet binary: %w", err))
	}
	if err := validateOsmosisMainnetChannelEvidence(e.Channel); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if err := e.Middleware.Validate(); err != nil {
		validationErrors = append(validationErrors, err)
	}
	if !e.Passed {
		validationErrors = append(validationErrors, errors.New("Osmosis mainnet preflight is not marked passed"))
	}
	if e.Error != "" {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis mainnet preflight recorded an error: %s", e.Error))
	}
	return errors.Join(validationErrors...)
}

func runOsmosisMainnetPreflight(
	ctx context.Context,
	client *http.Client,
	config OsmosisMainnetPreflightConfig,
	now time.Time,
) (evidence OsmosisMainnetPreflightEvidence, retErr error) {
	evidence = OsmosisMainnetPreflightEvidence{
		SchemaVersion: osmosisMainnetPreflightSchema,
		StartedAt:     now.UTC(),
		Config:        config,
		Expected:      expectedOsmosisMainnetFixture(),
		Channel: OsmosisMainnetChannelEvidence{
			InvestigationStatus: ibcInvestigationUnavailable,
		},
		Middleware: IBCMiddlewareInvestigationEvidence{
			InvestigationStatus: ibcInvestigationUnavailable,
		},
	}
	defer func() {
		evidence.CompletedAt = time.Now().UTC()
		if retErr != nil {
			evidence.Error = retErr.Error()
			evidence.Passed = false
		}
	}()
	if ctx == nil {
		return evidence, errors.New("Osmosis mainnet preflight context is required")
	}
	if client == nil {
		return evidence, errors.New("Osmosis mainnet preflight HTTP client is required")
	}
	if err := config.validate(); err != nil {
		return evidence, err
	}
	if now.IsZero() {
		return evidence, errors.New("Osmosis mainnet preflight clock is required")
	}

	preflightCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	statusURL := joinReadOnlyEndpoint(config.RPCEndpoint, "/status")
	var statusResponse osmosisStatusResponse
	if err := getOsmosisJSON(preflightCtx, client, statusURL, &statusResponse); err != nil {
		evidence.Status.Error = err.Error()
		evidence.Channel.Error = "status preflight unavailable: " + err.Error()
		evidence.Middleware.Error = "status preflight unavailable: " + err.Error()
		return evidence, fmt.Errorf("read Osmosis mainnet status: %w", err)
	}
	status, err := validateOsmosisStatusResponse(statusResponse, now.UTC())
	if err != nil {
		status.Error = err.Error()
		evidence.Status = status
		evidence.Channel.Error = "not investigated because status fixture mismatched: " + err.Error()
		evidence.Middleware.Error = "not investigated because status fixture mismatched: " + err.Error()
		return evidence, err
	}
	evidence.Status = status

	nodeInfoURL := joinReadOnlyEndpoint(config.RESTEndpoint, "/cosmos/base/tendermint/v1beta1/node_info")
	var nodeInfoResponse osmosisNodeInfoResponse
	if err := getOsmosisJSON(preflightCtx, client, nodeInfoURL, &nodeInfoResponse); err != nil {
		evidence.NodeInfo.Error = err.Error()
		evidence.Channel.Error = "node-info preflight unavailable: " + err.Error()
		evidence.Middleware.Error = "node-info preflight unavailable: " + err.Error()
		return evidence, fmt.Errorf("read Osmosis mainnet node-info: %w", err)
	}
	nodeInfo, err := validateOsmosisNodeInfoResponse(nodeInfoResponse)
	if err != nil {
		nodeInfo.Error = err.Error()
		evidence.NodeInfo = nodeInfo
		evidence.Channel.Error = "not investigated because node-info fixture mismatched: " + err.Error()
		evidence.Middleware.InvestigationStatus = ibcInvestigationMismatch
		evidence.Middleware.Error = "node-info fixture mismatch: " + err.Error()
		return evidence, err
	}
	evidence.NodeInfo = nodeInfo

	channelURL := joinReadOnlyEndpoint(
		config.RESTEndpoint,
		"/ibc/core/channel/v1/channels/"+osmosisMainnetChannelID+"/ports/transfer",
	)
	var channelResponse osmosisChannelResponse
	if err := getOsmosisJSON(preflightCtx, client, channelURL, &channelResponse); err != nil {
		evidence.Channel.Error = err.Error()
		evidence.Middleware.Error = "active channel unavailable: " + err.Error()
		return evidence, fmt.Errorf("read active Osmosis/Panacea mainnet channel: %w", err)
	}
	channel := newOsmosisMainnetChannelEvidence(channelResponse)
	if err := validateOsmosisMainnetChannelEvidence(channel); err != nil {
		channel.InvestigationStatus = ibcInvestigationMismatch
		channel.Error = err.Error()
		evidence.Channel = channel
		evidence.Middleware.InvestigationStatus = ibcInvestigationMismatch
		evidence.Middleware.Error = "active channel fixture mismatch: " + err.Error()
		return evidence, err
	}
	evidence.Channel = channel
	evidence.Middleware, err = newOsmosisMiddlewareEvidence(evidence.NodeInfo, evidence.Channel)
	if err != nil {
		evidence.Middleware.InvestigationStatus = ibcInvestigationMismatch
		evidence.Middleware.Error = err.Error()
		return evidence, err
	}
	evidence.Passed = true
	evidence.CompletedAt = time.Now().UTC()
	if err := evidence.Validate(); err != nil {
		return evidence, err
	}
	return evidence, nil
}

func getOsmosisJSON(ctx context.Context, client *http.Client, endpoint string, target any) error {
	for attempt := 1; ; attempt++ {
		requestCtx, cancel := context.WithTimeout(ctx, osmosisMainnetRequestTimeout)
		err := getOsmosisJSONOnce(requestCtx, client, endpoint, target)
		cancel()
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return err
		}
		delay, retry := osmosisMainnetRetryDelay(attempt)
		if !retry || !retryableOsmosisReadError(err) {
			return err
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("retry GET %s after transient error %v: %w", endpoint, err, ctx.Err())
		case <-timer.C:
		}
	}
}

type osmosisHTTPStatusError struct {
	endpoint   string
	statusCode int
	status     string
}

func (e *osmosisHTTPStatusError) Error() string {
	return fmt.Sprintf("GET %s returned %s", e.endpoint, e.status)
}

func retryableOsmosisReadError(err error) bool {
	var statusErr *osmosisHTTPStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.statusCode {
		case http.StatusTooManyRequests,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			return true
		default:
			return false
		}
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var networkErr net.Error
	return errors.As(err, &networkErr) && (networkErr.Timeout() || networkErr.Temporary())
}

func osmosisMainnetRetryDelay(completedAttempt int) (time.Duration, bool) {
	switch completedAttempt {
	case 1:
		return time.Second, true
	case 2:
		return 2 * time.Second, true
	default:
		return 0, false
	}
}

func getOsmosisJSONOnce(ctx context.Context, client *http.Client, endpoint string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "panacea-e2e-ibc-preflight/1")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return &osmosisHTTPStatusError{
			endpoint:   endpoint,
			statusCode: response.StatusCode,
			status:     response.Status,
		}
	}
	limited := io.LimitReader(response.Body, osmosisMainnetMaxResponseBytes+1)
	contents, err := io.ReadAll(limited)
	if err != nil {
		return err
	}
	if len(contents) > osmosisMainnetMaxResponseBytes {
		return fmt.Errorf("GET %s response exceeds %d bytes", endpoint, osmosisMainnetMaxResponseBytes)
	}
	decoder := json.NewDecoder(strings.NewReader(string(contents)))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode GET %s: %w", endpoint, err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return fmt.Errorf("decode GET %s: multiple JSON values", endpoint)
	}
	return nil
}

func joinReadOnlyEndpoint(base, path string) string {
	return strings.TrimRight(strings.TrimSpace(base), "/") + path
}

type osmosisStatusResponse struct {
	Result struct {
		NodeInfo struct {
			Network string `json:"network"`
			Version string `json:"version"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHeight string    `json:"latest_block_height"`
			LatestBlockTime   time.Time `json:"latest_block_time"`
			CatchingUp        bool      `json:"catching_up"`
		} `json:"sync_info"`
	} `json:"result"`
}

func validateOsmosisStatusResponse(response osmosisStatusResponse, now time.Time) (OsmosisMainnetStatusEvidence, error) {
	height, err := strconv.ParseInt(response.Result.SyncInfo.LatestBlockHeight, 10, 64)
	evidence := OsmosisMainnetStatusEvidence{
		Available:         true,
		Network:           response.Result.NodeInfo.Network,
		CometBFTVersion:   response.Result.NodeInfo.Version,
		LatestBlockHeight: height,
		LatestBlockTime:   response.Result.SyncInfo.LatestBlockTime,
		CatchingUp:        response.Result.SyncInfo.CatchingUp,
	}
	var validationErrors []error
	if err != nil || height < 1 {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis mainnet latest height %q is invalid", response.Result.SyncInfo.LatestBlockHeight))
	}
	if evidence.Network != osmosisMainnetChainID {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis status network = %q, want %q", evidence.Network, osmosisMainnetChainID))
	}
	wantComet := PinnedIBCProvenance().Osmosis.CometBFTVersion
	if !sameCometBFTMinorLine(evidence.CometBFTVersion, wantComet) {
		validationErrors = append(validationErrors, fmt.Errorf(
			"Osmosis status CometBFT = %q, want the same major.minor line as %q",
			evidence.CometBFTVersion,
			strings.TrimPrefix(wantComet, "v"),
		))
	}
	if evidence.CatchingUp {
		validationErrors = append(validationErrors, errors.New("Osmosis mainnet status reports catching_up=true"))
	}
	if evidence.LatestBlockTime.IsZero() || now.Sub(evidence.LatestBlockTime) > osmosisMainnetMaxBlockAge || evidence.LatestBlockTime.After(now.Add(2*time.Minute)) {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis latest block time %s is not fresh relative to %s", evidence.LatestBlockTime, now))
	}
	return evidence, errors.Join(validationErrors...)
}

// sameCometBFTMinorLine permits public RPC operators to run different patch
// releases from the pinned Osmosis application binary. Mainnet consensus does
// not require every RPC node to expose the same CometBFT patch version. The
// exact application dependency remains enforced through REST node-info and the
// digest-pinned local counterparty image.
func sameCometBFTMinorLine(observed, pinned string) bool {
	observedMajor, observedMinor, observedOK := cometBFTMajorMinor(observed)
	pinnedMajor, pinnedMinor, pinnedOK := cometBFTMajorMinor(pinned)
	return observedOK && pinnedOK && observedMajor == pinnedMajor && observedMinor == pinnedMinor
}

func cometBFTMajorMinor(version string) (int, int, bool) {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(version), "v"), ".")
	if len(parts) != 3 {
		return 0, 0, false
	}
	values := make([]int, len(parts))
	for index, part := range parts {
		if part == "" {
			return 0, 0, false
		}
		for _, character := range part {
			if character < '0' || character > '9' {
				return 0, 0, false
			}
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			return 0, 0, false
		}
		values[index] = value
	}
	return values[0], values[1], true
}

type osmosisBuildDependency struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

type osmosisNodeInfoResponse struct {
	DefaultNodeInfo struct {
		Network string `json:"network"`
		Version string `json:"version"`
	} `json:"default_node_info"`
	ApplicationVersion struct {
		Name             string                   `json:"name"`
		AppName          string                   `json:"app_name"`
		Version          string                   `json:"version"`
		GitCommit        string                   `json:"git_commit"`
		GoVersion        string                   `json:"go_version"`
		CosmosSDKVersion string                   `json:"cosmos_sdk_version"`
		BuildDeps        []osmosisBuildDependency `json:"build_deps"`
	} `json:"application_version"`
}

func validateOsmosisNodeInfoResponse(response osmosisNodeInfoResponse) (OsmosisMainnetNodeInfoEvidence, error) {
	dependencies := make([]IBCDependencyContract, 0, len(response.ApplicationVersion.BuildDeps))
	seen := make(map[string]struct{}, len(response.ApplicationVersion.BuildDeps))
	for _, dependency := range response.ApplicationVersion.BuildDeps {
		if _, duplicate := seen[dependency.Path]; duplicate {
			return OsmosisMainnetNodeInfoEvidence{}, fmt.Errorf("Osmosis node-info contains duplicate dependency %s", dependency.Path)
		}
		seen[dependency.Path] = struct{}{}
		dependencies = append(dependencies, IBCDependencyContract{Path: dependency.Path, Version: dependency.Version})
	}
	identity := IBCBinaryVersionIdentity{
		Name:             response.ApplicationVersion.Name,
		AppName:          response.ApplicationVersion.AppName,
		Version:          response.ApplicationVersion.Version,
		Commit:           response.ApplicationVersion.GitCommit,
		CosmosSDKVersion: response.ApplicationVersion.CosmosSDKVersion,
		GoVersion:        response.ApplicationVersion.GoVersion,
		Dependencies:     dependencies,
	}
	evidence := OsmosisMainnetNodeInfoEvidence{
		Available:                   true,
		Network:                     response.DefaultNodeInfo.Network,
		CometBFT:                    response.DefaultNodeInfo.Version,
		DependencyObservationScope:  osmosisNodeInfoDependencyScope,
		ReplacementMetadataObserved: false,
		Binary:                      identity,
	}
	var validationErrors []error
	if evidence.Network != osmosisMainnetChainID {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis node-info network = %q, want %q", evidence.Network, osmosisMainnetChainID))
	}
	wantComet := strings.TrimPrefix(PinnedIBCProvenance().Osmosis.CometBFTVersion, "v")
	if evidence.CometBFT != wantComet {
		validationErrors = append(validationErrors, fmt.Errorf("Osmosis node-info CometBFT = %q, want %q", evidence.CometBFT, wantComet))
	}
	if strings.TrimSpace(identity.GoVersion) == "" {
		validationErrors = append(validationErrors, errors.New("Osmosis node-info Go version is empty"))
	}
	if err := validateIBCBinaryVersionIdentity(identity, osmosisNodeInfoObservableContract()); err != nil {
		validationErrors = append(validationErrors, err)
	}
	return evidence, errors.Join(validationErrors...)
}

type osmosisChannelResponse struct {
	Channel struct {
		State        string `json:"state"`
		Ordering     string `json:"ordering"`
		Counterparty struct {
			PortID    string `json:"port_id"`
			ChannelID string `json:"channel_id"`
		} `json:"counterparty"`
		ConnectionHops []string `json:"connection_hops"`
		Version        string   `json:"version"`
	} `json:"channel"`
	ProofHeight struct {
		RevisionNumber string `json:"revision_number"`
		RevisionHeight string `json:"revision_height"`
	} `json:"proof_height"`
}

func newOsmosisMainnetChannelEvidence(response osmosisChannelResponse) OsmosisMainnetChannelEvidence {
	connectionID := ""
	if len(response.Channel.ConnectionHops) == 1 {
		connectionID = response.Channel.ConnectionHops[0]
	}
	return OsmosisMainnetChannelEvidence{
		InvestigationStatus:   ibcInvestigationConfirmed,
		ChannelID:             osmosisMainnetChannelID,
		PortID:                "transfer",
		State:                 response.Channel.State,
		Ordering:              response.Channel.Ordering,
		Version:               response.Channel.Version,
		CounterpartyPortID:    response.Channel.Counterparty.PortID,
		CounterpartyChannelID: response.Channel.Counterparty.ChannelID,
		ConnectionID:          connectionID,
		ProofHeight:           response.ProofHeight.RevisionNumber + "-" + response.ProofHeight.RevisionHeight,
	}
}

func validateOsmosisMainnetChannelEvidence(evidence OsmosisMainnetChannelEvidence) error {
	fixture := expectedOsmosisMainnetFixture()
	var validationErrors []error
	if evidence.InvestigationStatus != ibcInvestigationConfirmed {
		validationErrors = append(validationErrors, fmt.Errorf("mainnet channel investigation status = %q, want confirmed", evidence.InvestigationStatus))
	}
	if evidence.ChannelID != fixture.ChannelID || evidence.PortID != fixture.PortID ||
		evidence.State != fixture.State || evidence.Ordering != fixture.Ordering || evidence.Version != fixture.Version ||
		evidence.CounterpartyPortID != fixture.PortID || evidence.CounterpartyChannelID != fixture.CounterpartyChannelID ||
		evidence.ConnectionID != fixture.ConnectionID {
		validationErrors = append(validationErrors, fmt.Errorf(
			"active mainnet channel = %s/%s %s %s %s -> %s/%s via %s, want registry fixture %s/%s %s %s %s -> %s/%s via %s",
			evidence.PortID, evidence.ChannelID, evidence.State, evidence.Ordering, evidence.Version,
			evidence.CounterpartyPortID, evidence.CounterpartyChannelID, evidence.ConnectionID,
			fixture.PortID, fixture.ChannelID, fixture.State, fixture.Ordering, fixture.Version,
			fixture.PortID, fixture.CounterpartyChannelID, fixture.ConnectionID,
		))
	}
	if evidence.Error != "" {
		validationErrors = append(validationErrors, fmt.Errorf("mainnet channel investigation recorded an error: %s", evidence.Error))
	}
	return errors.Join(validationErrors...)
}

func expectedMainnetFixtureEqual(left, right OsmosisMainnetExpectedFixture) bool {
	return left.ChainID == right.ChainID && left.CounterpartyChainID == right.CounterpartyChainID &&
		left.ChannelRegistryReference == right.ChannelRegistryReference && left.ChannelID == right.ChannelID &&
		left.CounterpartyChannelID == right.CounterpartyChannelID && left.ConnectionID == right.ConnectionID &&
		left.PortID == right.PortID && left.State == right.State && left.Ordering == right.Ordering &&
		left.Version == right.Version && left.RegistryStatus == right.RegistryStatus &&
		left.RegistryPreferred == right.RegistryPreferred && binaryContractsEqual(left.Binary, right.Binary)
}
