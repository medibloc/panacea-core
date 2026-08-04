package harness

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc/metadata"
)

const queryResponseMaxBytes = 4 << 20

type restQueryRequestEvidence struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Height int64  `json:"height"`
}

// FullNodeCLIQuery executes a JSON query against the explicit full node and
// records the public request and raw response in the run's evidence bundle.
func (n *Network) FullNodeCLIQuery(ctx context.Context, step string, command ...string) (json.RawMessage, error) {
	return n.fullNodeCommandQuery(ctx, "cli", step, command...)
}

// FullNodeGRPCQuery executes the same query command with an explicit gRPC
// endpoint on the selected full node. It provides a custom-module gRPC
// boundary without coupling the E2E module to Panacea's generated Go types.
func (n *Network) FullNodeGRPCQuery(ctx context.Context, step string, command ...string) (json.RawMessage, error) {
	if len(n.Chain.FullNodes) == 0 {
		return nil, errors.New("network has no full node")
	}
	fullNode := n.Chain.FullNodes[0]
	command = append(
		append([]string(nil), command...),
		"--grpc-addr", fullNode.HostName()+":9090",
		"--grpc-insecure",
	)
	return n.fullNodeCommandQuery(ctx, "grpc", step, command...)
}

// FullNodeGRPCQueryExpectedError executes a custom-module query through the
// explicit full-node gRPC endpoint and treats one stable diagnostic as the
// successful outcome. This keeps negative query evidence in results.jsonl
// without turning an expected NotFound response into a run failure.
func (n *Network) FullNodeGRPCQueryExpectedError(
	ctx context.Context,
	step string,
	expectedDiagnostic string,
	command ...string,
) error {
	if strings.TrimSpace(step) == "" {
		return errors.New("query step is required")
	}
	if strings.TrimSpace(expectedDiagnostic) == "" {
		return errors.New("expected query diagnostic is required")
	}
	if len(command) == 0 {
		return errors.New("query command is required")
	}
	if len(n.Chain.FullNodes) == 0 {
		return errors.New("network has no full node")
	}

	fullNode := n.Chain.FullNodes[0]
	command = append(
		append([]string(nil), command...),
		"--grpc-addr", fullNode.HostName()+":9090",
		"--grpc-insecure",
	)
	stdout, stderr, queryErr := fullNode.ExecQuery(ctx, command...)
	classificationErr := classifyExpectedCLIRejection(
		queryErr,
		stdout,
		stderr,
		expectedDiagnostic,
	)
	height := queryCommandHeight(command)
	responseEvidence := jsonOrString(stdout)
	if strings.TrimSpace(string(stdout)) == "" {
		responseEvidence = map[string]any{
			"expected_error": true,
			"diagnostic":     boundedString(stderr, txStderrMaxBytes),
		}
	}
	recordErr := n.recordQuery(queryRecord{
		Boundary:         "grpc",
		Step:             step,
		Height:           height,
		HistoricalHeight: height > 0,
		Request: map[string]any{
			"arguments": append([]string(nil), command...),
		},
		Response: responseEvidence,
		Metadata: map[string]any{
			"expected_error":      true,
			"expected_diagnostic": expectedDiagnostic,
			"observed":            classificationErr == nil,
		},
		Stderr: boundedString(stderr, txStderrMaxBytes),
		Error:  errorString(classificationErr),
	})
	if classificationErr != nil {
		classificationErr = fmt.Errorf("full-node gRPC query %s: %w", step, classificationErr)
		n.artifacts.recordFailure("full-node-grpc-query-"+step, classificationErr)
	}
	return errors.Join(classificationErr, recordErr)
}

func (n *Network) fullNodeCommandQuery(
	ctx context.Context,
	boundary string,
	step string,
	command ...string,
) (json.RawMessage, error) {
	if strings.TrimSpace(step) == "" {
		return nil, errors.New("query step is required")
	}
	if len(command) == 0 {
		return nil, errors.New("query command is required")
	}
	if len(n.Chain.FullNodes) == 0 {
		return nil, errors.New("network has no full node")
	}

	stdout, stderr, queryErr := n.Chain.FullNodes[0].ExecQuery(ctx, command...)
	height := queryCommandHeight(command)
	recordErr := n.recordQuery(queryRecord{
		Boundary:         boundary,
		Step:             step,
		Height:           height,
		HistoricalHeight: height > 0,
		Request: map[string]any{
			"arguments": append([]string(nil), command...),
		},
		Response: jsonOrString(stdout),
		Stderr:   boundedString(stderr, txStderrMaxBytes),
		Error:    errorString(queryErr),
	})
	if recordErr != nil {
		return nil, errors.Join(queryErr, recordErr)
	}
	if queryErr != nil {
		err := fmt.Errorf("full-node %s query %s: %w: %s", boundary, step, queryErr, boundedString(stderr, txStderrMaxBytes))
		n.artifacts.recordFailure("full-node-"+boundary+"-query-"+step, err)
		return nil, err
	}
	trimmed := strings.TrimSpace(string(stdout))
	if !json.Valid([]byte(trimmed)) {
		err := fmt.Errorf("full-node %s query %s returned invalid JSON", boundary, step)
		n.artifacts.recordFailure("full-node-"+boundary+"-query-"+step, err)
		return nil, err
	}
	return append(json.RawMessage(nil), trimmed...), nil
}

// FullNodeRESTGet performs one bounded REST query through the explicit full
// node. The supplied path is retained verbatim so tests can distinguish raw
// and percent-encoded identifiers.
func (n *Network) FullNodeRESTGet(
	ctx context.Context,
	client *http.Client,
	step string,
	path string,
) (json.RawMessage, error) {
	return n.FullNodeRESTGetAtHeight(ctx, client, step, path, 0)
}

// FullNodeRESTGetAtHeight pins one REST query to a committed application
// height through the Cosmos SDK gateway metadata contract.
func (n *Network) FullNodeRESTGetAtHeight(
	ctx context.Context,
	client *http.Client,
	step string,
	path string,
	height int64,
) (json.RawMessage, error) {
	if strings.TrimSpace(step) == "" {
		return nil, errors.New("query step is required")
	}
	address, err := n.FullNodeHostAddress(ctx, "1317/tcp")
	if err != nil {
		return nil, err
	}
	request, requestEvidence, err := newFullNodeRESTRequest(ctx, address, path, height)
	if err != nil {
		return nil, fmt.Errorf("REST query %s request: %w", step, err)
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	response, requestErr := client.Do(request)
	if requestErr != nil {
		recordErr := n.recordQuery(queryRecord{
			Boundary:         "rest",
			Step:             step,
			Height:           height,
			HistoricalHeight: height > 0,
			Request:          requestEvidence,
			Error:            requestErr.Error(),
		})
		return nil, errors.Join(fmt.Errorf("REST query %s: %w", step, requestErr), recordErr)
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(response.Body, queryResponseMaxBytes+1))
	tooLarge := len(body) > queryResponseMaxBytes
	if tooLarge {
		body = body[:queryResponseMaxBytes]
	}
	var responseHeightErr error
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		responseHeightErr = validateRESTQueryResponseHeight(height, response.Header)
	}
	responseErr := fullNodeRESTResponseError(
		step,
		response.StatusCode,
		body,
		tooLarge,
		readErr,
		responseHeightErr,
	)
	recordErr := n.recordQuery(queryRecord{
		Boundary:         "rest",
		Step:             step,
		Height:           height,
		HistoricalHeight: height > 0,
		Request:          requestEvidence,
		Response:         jsonOrString(body),
		Status:           response.StatusCode,
		Metadata: map[string]string{
			"grpc_block_height": response.Header.Get("Grpc-Metadata-X-Cosmos-Block-Height"),
		},
		Error: errorString(responseErr),
	})
	if recordErr != nil || responseErr != nil {
		return nil, errors.Join(responseErr, recordErr)
	}
	return append(json.RawMessage(nil), body...), nil
}

func fullNodeRESTResponseError(
	step string,
	statusCode int,
	body []byte,
	tooLarge bool,
	readErr error,
	heightErr error,
) error {
	var responseErrors []error
	if readErr != nil {
		responseErrors = append(responseErrors, readErr)
	}
	if tooLarge {
		responseErrors = append(responseErrors, fmt.Errorf(
			"REST query %s response exceeds %d bytes",
			step,
			queryResponseMaxBytes,
		))
	}
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		responseErrors = append(responseErrors, fmt.Errorf(
			"REST query %s returned HTTP %d: %s",
			step,
			statusCode,
			boundedString(body, txStderrMaxBytes),
		))
		return errors.Join(responseErrors...)
	}
	if heightErr != nil {
		responseErrors = append(responseErrors, heightErr)
	}
	// A truncated or partially read body already has the more precise size/read
	// failure. Do not add a derivative JSON error that obscures the cause.
	if !tooLarge && readErr == nil && !json.Valid(body) {
		responseErrors = append(responseErrors, fmt.Errorf("REST query %s returned invalid JSON", step))
	}
	return errors.Join(responseErrors...)
}

func validateRESTQueryResponseHeight(requestedHeight int64, headers http.Header) error {
	if requestedHeight <= 0 {
		return nil
	}
	want := strconv.FormatInt(requestedHeight, 10)
	got := strings.TrimSpace(headers.Get("Grpc-Metadata-X-Cosmos-Block-Height"))
	if got != want {
		return fmt.Errorf("REST response height %q, want pinned height %s", got, want)
	}
	return nil
}

func newFullNodeRESTRequest(
	ctx context.Context,
	address string,
	path string,
	height int64,
) (*http.Request, restQueryRequestEvidence, error) {
	requestURL, err := restURL(address, path)
	if err != nil {
		return nil, restQueryRequestEvidence{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, restQueryRequestEvidence{}, err
	}
	if height > 0 {
		request.Header.Set("x-cosmos-block-height", strconv.FormatInt(height, 10))
	}
	return request, restQueryRequestEvidence{
		Method: http.MethodGet,
		Path:   path,
		Height: height,
	}, nil
}

// QueryNFTClassGRPC queries the standard NFT class service on the full node.
func (n *Network) QueryNFTClassGRPC(
	ctx context.Context,
	step string,
	classID string,
) (*upstreamnft.QueryClassResponse, error) {
	client, err := n.standardNFTQueryClient()
	if err != nil {
		return nil, err
	}
	request := &upstreamnft.QueryClassRequest{ClassId: classID}
	response, queryErr := client.Class(ctx, request)
	return response, n.finishGRPCQueryAtContext(ctx, step, request, response, queryErr)
}

// QueryNFTGRPC queries one standard NFT on the full node.
func (n *Network) QueryNFTGRPC(
	ctx context.Context,
	step string,
	classID string,
	nftID string,
) (*upstreamnft.QueryNFTResponse, error) {
	client, err := n.standardNFTQueryClient()
	if err != nil {
		return nil, err
	}
	request := &upstreamnft.QueryNFTRequest{ClassId: classID, Id: nftID}
	response, queryErr := client.NFT(ctx, request)
	return response, n.finishGRPCQueryAtContext(ctx, step, request, response, queryErr)
}

// QueryNFTOwnerGRPC queries the owner index on the full node.
func (n *Network) QueryNFTOwnerGRPC(
	ctx context.Context,
	step string,
	classID string,
	nftID string,
) (*upstreamnft.QueryOwnerResponse, error) {
	client, err := n.standardNFTQueryClient()
	if err != nil {
		return nil, err
	}
	request := &upstreamnft.QueryOwnerRequest{ClassId: classID, Id: nftID}
	response, queryErr := client.Owner(ctx, request)
	return response, n.finishGRPCQueryAtContext(ctx, step, request, response, queryErr)
}

// QueryNFTSupplyGRPC queries standard total supply on the full node.
func (n *Network) QueryNFTSupplyGRPC(
	ctx context.Context,
	step string,
	classID string,
) (*upstreamnft.QuerySupplyResponse, error) {
	client, err := n.standardNFTQueryClient()
	if err != nil {
		return nil, err
	}
	request := &upstreamnft.QuerySupplyRequest{ClassId: classID}
	response, queryErr := client.Supply(ctx, request)
	return response, n.finishGRPCQueryAtContext(ctx, step, request, response, queryErr)
}

// QueryNFTBalanceGRPC queries one owner's standard class balance.
func (n *Network) QueryNFTBalanceGRPC(
	ctx context.Context,
	step string,
	classID string,
	owner string,
) (*upstreamnft.QueryBalanceResponse, error) {
	client, err := n.standardNFTQueryClient()
	if err != nil {
		return nil, err
	}
	request := &upstreamnft.QueryBalanceRequest{ClassId: classID, Owner: owner}
	response, queryErr := client.Balance(ctx, request)
	return response, n.finishGRPCQueryAtContext(ctx, step, request, response, queryErr)
}

// QueryNFTsGRPC queries the standard NFT list using the caller's pagination.
func (n *Network) QueryNFTsGRPC(
	ctx context.Context,
	step string,
	request *upstreamnft.QueryNFTsRequest,
) (*upstreamnft.QueryNFTsResponse, error) {
	client, err := n.standardNFTQueryClient()
	if err != nil {
		return nil, err
	}
	response, queryErr := client.NFTs(ctx, request)
	return response, n.finishGRPCQueryAtContext(ctx, step, request, response, queryErr)
}

// QueryNFTClassesGRPC queries the standard class list using the caller's pagination.
func (n *Network) QueryNFTClassesGRPC(
	ctx context.Context,
	step string,
	request *upstreamnft.QueryClassesRequest,
) (*upstreamnft.QueryClassesResponse, error) {
	client, err := n.standardNFTQueryClient()
	if err != nil {
		return nil, err
	}
	response, queryErr := client.Classes(ctx, request)
	return response, n.finishGRPCQueryAtContext(ctx, step, request, response, queryErr)
}

type queryRecord struct {
	RecordedAt       time.Time `json:"recorded_at"`
	Boundary         string    `json:"boundary"`
	Step             string    `json:"step"`
	Height           int64     `json:"height"`
	HistoricalHeight bool      `json:"historical_height"`
	Request          any       `json:"request,omitempty"`
	Response         any       `json:"response,omitempty"`
	Status           int       `json:"status,omitempty"`
	Metadata         any       `json:"metadata,omitempty"`
	Stderr           string    `json:"stderr,omitempty"`
	Error            string    `json:"error,omitempty"`
}

// ContextAtHeight pins a typed gRPC query to the same committed application
// height used by the CLI and REST assertions.
func ContextAtHeight(ctx context.Context, height int64) context.Context {
	if height <= 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "x-cosmos-block-height", strconv.FormatInt(height, 10))
}

func (n *Network) standardNFTQueryClient() (upstreamnft.QueryClient, error) {
	if len(n.Chain.FullNodes) == 0 {
		return nil, errors.New("network has no full node")
	}
	return upstreamnft.NewQueryClient(n.Chain.FullNodes[0].GrpcConn), nil
}

func (n *Network) finishGRPCQuery(step string, request, response any, queryErr error) error {
	return n.finishGRPCQueryAtContext(context.Background(), step, request, response, queryErr)
}

func (n *Network) finishGRPCQueryAtContext(ctx context.Context, step string, request, response any, queryErr error) error {
	height := grpcQueryContextHeight(ctx)
	queryMetadata := map[string]any{}
	if height > 0 {
		queryMetadata["request_height"] = height
	}
	recordErr := n.recordQuery(queryRecord{
		Boundary:         "grpc",
		Step:             step,
		Height:           height,
		HistoricalHeight: height > 0,
		Request:          grpcQueryEvidence(request),
		Response:         grpcQueryEvidence(response),
		Metadata:         queryMetadata,
		Error:            errorString(queryErr),
	})
	if queryErr != nil {
		queryErr = fmt.Errorf("full-node gRPC query %s: %w", step, queryErr)
		n.artifacts.recordFailure("full-node-grpc-query-"+step, queryErr)
	}
	return errors.Join(queryErr, recordErr)
}

func queryCommandHeight(command []string) int64 {
	for index, argument := range command {
		var value string
		switch {
		case argument == "--height" && index+1 < len(command):
			value = command[index+1]
		case strings.HasPrefix(argument, "--height="):
			value = strings.TrimPrefix(argument, "--height=")
		default:
			continue
		}
		height, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err == nil && height > 0 {
			return height
		}
		return 0
	}
	return 0
}

func grpcQueryContextHeight(ctx context.Context) int64 {
	values, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return 0
	}
	heights := values.Get("x-cosmos-block-height")
	if len(heights) != 1 {
		return 0
	}
	height, err := strconv.ParseInt(strings.TrimSpace(heights[0]), 10, 64)
	if err != nil || height <= 0 {
		return 0
	}
	return height
}

func grpcQueryEvidence(value any) any {
	message, ok := value.(proto.Message)
	if !ok {
		return value
	}
	if message == nil {
		return nil
	}
	messageValue := reflect.ValueOf(message)
	if messageValue.Kind() == reflect.Ptr && messageValue.IsNil() {
		return nil
	}
	encoded, err := proto.Marshal(message)
	if err != nil {
		return map[string]any{
			"protobuf_type": proto.MessageName(message),
			"protobuf_text": message.String(),
			"marshal_error": err.Error(),
		}
	}
	return map[string]any{
		"protobuf_type":   proto.MessageName(message),
		"protobuf_base64": base64.StdEncoding.EncodeToString(encoded),
		"protobuf_text":   message.String(),
	}
}

func (n *Network) recordQuery(record queryRecord) error {
	if record.RecordedAt.IsZero() {
		record.RecordedAt = time.Now().UTC()
	}
	if err := n.artifacts.appendJSONLine("queries/results.jsonl", record); err != nil {
		n.artifacts.recordFailure("record-query-"+record.Step, err)
		return fmt.Errorf("record query %s: %w", record.Step, err)
	}
	return nil
}

func restURL(address, path string) (string, error) {
	base, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	if base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("REST endpoint must include scheme and host: %q", address)
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "", fmt.Errorf("REST path must be an absolute path without authority: %q", path)
	}
	reference, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	if reference.IsAbs() || reference.Host != "" || reference.User != nil {
		return "", fmt.Errorf("REST path must not replace endpoint authority: %q", path)
	}
	base.Path = ""
	base.RawPath = ""
	base.RawQuery = ""
	base.Fragment = ""
	return base.ResolveReference(reference).String(), nil
}
