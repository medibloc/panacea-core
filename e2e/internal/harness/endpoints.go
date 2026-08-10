package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const maxEndpointBody = 4 << 20

// RequireRPCStatus verifies the host TCP RPC boundary and its committed height.
func RequireRPCStatus(ctx context.Context, address string, minimumHeight int64) error {
	client := &http.Client{}
	body, err := getJSON(ctx, client, strings.TrimRight(address, "/")+"/status")
	if err != nil {
		return fmt.Errorf("CometBFT RPC status: %w", err)
	}
	var response struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHeight string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("decode CometBFT RPC status: %w", err)
	}
	height, err := strconv.ParseInt(response.Result.SyncInfo.LatestBlockHeight, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid RPC latest height %q: %w", response.Result.SyncInfo.LatestBlockHeight, err)
	}
	if height < minimumHeight {
		return fmt.Errorf("RPC latest height %d is below required %d", height, minimumHeight)
	}
	return nil
}

// RequireRESTNodeInfo verifies the host REST gateway boundary.
func RequireRESTNodeInfo(ctx context.Context, client *http.Client, address string) error {
	body, err := getJSON(ctx, client, strings.TrimRight(address, "/")+"/cosmos/base/tendermint/v1beta1/node_info")
	if err != nil {
		return fmt.Errorf("REST node info: %w", err)
	}
	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("decode REST node info: %w", err)
	}
	if len(response) == 0 {
		return fmt.Errorf("REST node info returned an empty object")
	}
	return nil
}

func getJSON(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	body, _, err := getJSONAtHeight(ctx, client, url, 0)
	return body, err
}

// getJSONAtHeight performs a bounded REST query and optionally pins it to one
// committed height. Response metadata is returned so callers can prove which
// height the gRPC gateway served.
func getJSONAtHeight(
	ctx context.Context,
	client *http.Client,
	url string,
	height int64,
) ([]byte, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	if height > 0 {
		req.Header.Set("x-cosmos-block-height", strconv.FormatInt(height, 10))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxEndpointBody+1))
	if err != nil {
		return nil, resp.Header.Clone(), err
	}
	if len(body) > maxEndpointBody {
		return nil, resp.Header.Clone(), fmt.Errorf("response exceeds %d bytes", maxEndpointBody)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.Header.Clone(), fmt.Errorf("HTTP %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return body, resp.Header.Clone(), nil
}
