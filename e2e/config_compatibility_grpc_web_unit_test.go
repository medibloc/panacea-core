package e2e_test

import (
	"encoding/binary"
	"testing"

	"github.com/cometbft/cometbft/proto/tendermint/p2p"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

func TestDecodeConfigCompatibilityGRPCWebUnaryResponse(t *testing.T) {
	t.Parallel()

	const currentPatchVersion = "2.3.1"
	want := &cmtservice.GetNodeInfoResponse{
		DefaultNodeInfo: &p2p.DefaultNodeInfo{Network: "panacea-test"},
		ApplicationVersion: &cmtservice.VersionInfo{
			AppName: configCompatibilityGRPCWebAppName,
			Version: currentPatchVersion,
		},
	}
	payload, err := proto.Marshal(want)
	require.NoError(t, err)
	body := appendConfigCompatibilityGRPCWebFrame(nil, 0, payload)
	body = appendConfigCompatibilityGRPCWebFrame(body, 0x80, []byte("grpc-status: 0\r\ngrpc-message: \r\n"))

	gotPayload, trailers, frameCount, err := decodeConfigCompatibilityGRPCWebUnaryResponse(body)
	require.NoError(t, err)
	require.Equal(t, 2, frameCount)
	require.Equal(t, "0", trailers["grpc-status"])
	var got cmtservice.GetNodeInfoResponse
	require.NoError(t, proto.Unmarshal(gotPayload, &got))
	require.Equal(t, want.GetDefaultNodeInfo().Network, got.GetDefaultNodeInfo().Network)
	require.Equal(t, want.GetApplicationVersion().Version, got.GetApplicationVersion().Version)
	network, appName, appVersion, err := validateConfigCompatibilityGRPCWebNodeInfo(
		&got,
		"panacea-test",
		currentPatchVersion,
	)
	require.NoError(t, err)
	require.Equal(t, "panacea-test", network)
	require.Equal(t, configCompatibilityGRPCWebAppName, appName)
	require.Equal(t, currentPatchVersion, appVersion)

	_, _, _, err = decodeConfigCompatibilityGRPCWebUnaryResponse(
		appendConfigCompatibilityGRPCWebFrame(nil, 0, payload),
	)
	require.ErrorContains(t, err, "missing its trailer")

	badStatus := appendConfigCompatibilityGRPCWebFrame(nil, 0, payload)
	badStatus = appendConfigCompatibilityGRPCWebFrame(badStatus, 0x80, []byte("grpc-status: 13\r\n"))
	_, _, _, err = decodeConfigCompatibilityGRPCWebUnaryResponse(badStatus)
	require.ErrorContains(t, err, "want 0")

	duplicateStatus := appendConfigCompatibilityGRPCWebFrame(nil, 0, payload)
	duplicateStatus = appendConfigCompatibilityGRPCWebFrame(
		duplicateStatus,
		0x80,
		[]byte("grpc-status: 13\r\ngrpc-status: 0\r\n"),
	)
	_, _, _, err = decodeConfigCompatibilityGRPCWebUnaryResponse(duplicateStatus)
	require.ErrorContains(t, err, "duplicate")

	_, _, _, err = decodeConfigCompatibilityGRPCWebUnaryResponse([]byte{0, 0, 0})
	require.ErrorContains(t, err, "truncated")

	require.NoError(t, validateConfigCompatibilityGRPCWebContentType("application/grpc-web+proto; charset=utf-8"))
	require.NoError(t, validateConfigCompatibilityGRPCWebContentType("application/grpc-web"))
	require.Error(t, validateConfigCompatibilityGRPCWebContentType("application/grpc-webfoo"))
	require.Error(t, validateConfigCompatibilityGRPCWebContentType("application/grpc-web; broken"))

	wrongApp := proto.Clone(want).(*cmtservice.GetNodeInfoResponse)
	wrongApp.ApplicationVersion.AppName = "otherd"
	_, _, _, err = validateConfigCompatibilityGRPCWebNodeInfo(wrongApp, "panacea-test", currentPatchVersion)
	require.ErrorContains(t, err, "otherd")
	_, _, _, err = validateConfigCompatibilityGRPCWebNodeInfo(want, "other-chain", currentPatchVersion)
	require.ErrorContains(t, err, "other-chain")
	_, _, _, err = validateConfigCompatibilityGRPCWebNodeInfo(want, "panacea-test", "")
	require.ErrorContains(t, err, "PANACEA_E2E_CURRENT_BINARY_VERSION")
}

func appendConfigCompatibilityGRPCWebFrame(body []byte, flags byte, payload []byte) []byte {
	header := make([]byte, 5)
	header[0] = flags
	binary.BigEndian.PutUint32(header[1:], uint32(len(payload)))
	body = append(body, header...)
	return append(body, payload...)
}
