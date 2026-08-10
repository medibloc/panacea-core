package harness

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetJSONAtHeightPinsRequestAndReturnsResponseMetadata(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, "41", request.Header.Get("x-cosmos-block-height"))
		return &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Grpc-Metadata-X-Cosmos-Block-Height": []string{"41"},
				"Content-Type":                        []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, nil
	})}

	body, headers, err := getJSONAtHeight(context.Background(), client, "http://fullnode.test/query", 41)
	require.NoError(t, err)
	require.JSONEq(t, `{"ok":true}`, string(body))
	require.Equal(t, "41", headers.Get("Grpc-Metadata-X-Cosmos-Block-Height"))
}

func TestGetJSONAtHeightPreservesBoundedHTTPFailure(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "404 Not Found",
			StatusCode: http.StatusNotFound,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"code":5,"message":"missing"}`)),
		}, nil
	})}

	_, _, err := getJSONAtHeight(context.Background(), client, "http://fullnode.test/query", 0)
	require.ErrorContains(t, err, "HTTP 404")
	require.ErrorContains(t, err, "missing")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}
