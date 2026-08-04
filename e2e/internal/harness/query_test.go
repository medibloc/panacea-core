package harness

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	upstreamnft "cosmossdk.io/x/nft"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"
)

func TestNewFullNodeRESTRequestPinsHeightAndPreservesArtifactPath(t *testing.T) {
	const path = "/cosmos/mint/v1beta1/annual_provisions"
	request, evidence, err := newFullNodeRESTRequest(
		context.Background(),
		"http://127.0.0.1:1317",
		path,
		82,
	)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:1317"+path, request.URL.String())
	require.Equal(t, "82", request.Header.Get("x-cosmos-block-height"))
	require.Equal(t, restQueryRequestEvidence{Method: "GET", Path: path, Height: 82}, evidence)

	encoded, err := json.Marshal(evidence)
	require.NoError(t, err)
	require.JSONEq(t, `{"method":"GET","path":"/cosmos/mint/v1beta1/annual_provisions","height":82}`, string(encoded))
}

func TestValidateRESTQueryResponseHeightRequiresServerConfirmation(t *testing.T) {
	matching := http.Header{"Grpc-Metadata-X-Cosmos-Block-Height": []string{"82"}}
	require.NoError(t, validateRESTQueryResponseHeight(82, matching))
	require.NoError(t, validateRESTQueryResponseHeight(0, nil))
	require.ErrorContains(t, validateRESTQueryResponseHeight(82, http.Header{}), `height ""`)
	require.ErrorContains(
		t,
		validateRESTQueryResponseHeight(82, http.Header{"Grpc-Metadata-X-Cosmos-Block-Height": []string{"83"}}),
		`height "83"`,
	)
}

func TestQueryEvidenceHeightExtraction(t *testing.T) {
	t.Parallel()
	require.Equal(t, int64(82), queryCommandHeight([]string{"staking", "pool", "--height", "82"}))
	require.Equal(t, int64(83), queryCommandHeight([]string{"staking", "pool", "--height=83"}))
	require.Zero(t, queryCommandHeight([]string{"staking", "pool"}))
	require.Zero(t, queryCommandHeight([]string{"staking", "pool", "--height", "invalid"}))

	pinned := ContextAtHeight(context.Background(), 84)
	require.Equal(t, int64(84), grpcQueryContextHeight(pinned))
	require.Zero(t, grpcQueryContextHeight(context.Background()))
}

func TestFullNodeRESTResponseErrorRecordsEveryTerminalFailure(t *testing.T) {
	t.Parallel()

	require.NoError(t, fullNodeRESTResponseError("ok", 200, []byte(`{"ok":true}`), false, nil, nil))

	readErr := errors.New("read response body")
	err := fullNodeRESTResponseError("read", 200, []byte(`{"partial"`), false, readErr, nil)
	require.ErrorIs(t, err, readErr)
	require.NotContains(t, err.Error(), "invalid JSON")

	err = fullNodeRESTResponseError("large", 200, []byte(`{"partial"`), true, nil, nil)
	require.ErrorContains(t, err, "response exceeds")
	require.NotContains(t, err.Error(), "invalid JSON")

	err = fullNodeRESTResponseError("status", 503, []byte(`{"error":"unavailable"}`), false, nil, nil)
	require.ErrorContains(t, err, `returned HTTP 503: {"error":"unavailable"}`)

	err = fullNodeRESTResponseError("json", 200, []byte(`<html>ok</html>`), false, nil, nil)
	require.ErrorContains(t, err, "returned invalid JSON")

	heightErr := errors.New("REST response height mismatch")
	err = fullNodeRESTResponseError("height", 200, []byte(`{"ok":true}`), false, nil, heightErr)
	require.ErrorIs(t, err, heightErr)
}

func TestRESTURLPreservesRawAndEncodedNFTIdentifiers(t *testing.T) {
	testCases := []struct {
		name string
		path string
		want string
	}{
		{
			name: "raw class delimiter",
			path: "/panacea/nft/v1/classes/panacea1creator:certificate",
			want: "http://127.0.0.1:1317/panacea/nft/v1/classes/panacea1creator:certificate",
		},
		{
			name: "percent encoded class delimiter and dotted nft id",
			path: "/panacea/nft/v1/nfts/panacea1creator%3Acertificate/certificate.1",
			want: "http://127.0.0.1:1317/panacea/nft/v1/nfts/panacea1creator%3Acertificate/certificate.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := restURL("http://127.0.0.1:1317", tc.path)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestRESTURLRejectsAuthorityAndNonAbsolutePaths(t *testing.T) {
	for _, path := range []string{"relative", "//other-host/path", "https://other-host/path"} {
		_, err := restURL("http://127.0.0.1:1317", path)
		require.Error(t, err, path)
	}
}

func TestGRPCQueryEvidenceDoesNotRemarshalUnknownAny(t *testing.T) {
	response := &upstreamnft.QueryNFTResponse{
		Nft: &upstreamnft.NFT{
			ClassId: "creator:class",
			Id:      "certificate.1",
			Data: &codectypes.Any{
				TypeUrl: "/panacea.nft.v1.BasicNFTData",
				Value:   []byte{0x0a, 0x03, 'n', 'f', 't'},
			},
		},
	}

	evidence := grpcQueryEvidence(response)
	contents, err := json.Marshal(evidence)
	require.NoError(t, err)
	require.Contains(t, string(contents), "QueryNFTResponse")
	require.Contains(t, string(contents), "BasicNFTData")
}

func TestFinishGRPCQueryPreservesTypedNilQueryAndArtifactErrors(t *testing.T) {
	t.Parallel()

	root := trustedArtifactTempDir(t)
	artifactFile := filepath.Join(root, "artifact-file")
	require.NoError(t, os.WriteFile(artifactFile, []byte("not a directory"), 0o600))

	queryErr := errors.New("upstream gRPC query failed")
	var response *upstreamnft.QueryClassResponse
	network := &Network{artifacts: &artifactStore{dir: artifactFile}}

	err := network.finishGRPCQuery("typed-nil", &upstreamnft.QueryClassRequest{}, response, queryErr)
	require.Error(t, err)
	require.ErrorIs(t, err, queryErr)
	require.ErrorContains(t, err, "full-node gRPC query typed-nil")
	require.ErrorContains(t, err, "record query typed-nil")
}
