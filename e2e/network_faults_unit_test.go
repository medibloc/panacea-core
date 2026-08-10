package e2e_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStartNetworkFaultNodeWithReadinessRecoveryRestartsOneCatchingUpProcess(t *testing.T) {
	t.Parallel()

	startCalls := 0
	stopCalls := 0
	var recorded networkFaultReadinessRecovery
	err := startNetworkFaultNodeWithReadinessRecovery(
		context.Background(),
		"endpoint-grpc-disabled",
		networkFaultReadinessRuntime{
			Start: func(context.Context) error {
				startCalls++
				if startCalls == 1 {
					return errors.New("still catching up")
				}
				return nil
			},
			Stop: func(context.Context) error {
				stopCalls++
				return nil
			},
			Observe: func(context.Context) (int64, bool, error) {
				return 101, true, nil
			},
			Record: func(evidence networkFaultReadinessRecovery) error {
				recorded = evidence
				return nil
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, 2, startCalls)
	require.Equal(t, 1, stopCalls)
	require.Equal(t, "endpoint-grpc-disabled", recorded.Phase)
	require.EqualValues(t, 101, recorded.FirstStartHeight)
	require.True(t, recorded.FirstStartCatchingUp)
	require.True(t, recorded.RecoveryRestartSucceeded)
}

func TestStartNetworkFaultNodeWithReadinessRecoveryFailsClosedAfterOneRetry(t *testing.T) {
	t.Parallel()

	startCalls := 0
	stopCalls := 0
	err := startNetworkFaultNodeWithReadinessRecovery(
		context.Background(),
		"endpoint-grpc-disabled",
		networkFaultReadinessRuntime{
			Start: func(context.Context) error {
				startCalls++
				return errors.New("still catching up")
			},
			Stop: func(context.Context) error {
				stopCalls++
				return nil
			},
			Observe: func(context.Context) (int64, bool, error) {
				return 101, true, nil
			},
			Record: func(networkFaultReadinessRecovery) error { return nil },
		},
	)
	require.ErrorContains(t, err, "after one bounded readiness restart")
	require.Equal(t, 2, startCalls)
	require.Equal(t, 1, stopCalls)
}

func TestStartNetworkFaultNodeWithReadinessRecoveryAcceptsCaughtUpTimeoutRace(t *testing.T) {
	t.Parallel()

	stopCalls := 0
	err := startNetworkFaultNodeWithReadinessRecovery(
		context.Background(),
		"endpoint-api-disabled",
		networkFaultReadinessRuntime{
			Start: func(context.Context) error { return context.DeadlineExceeded },
			Stop: func(context.Context) error {
				stopCalls++
				return nil
			},
			Observe: func(context.Context) (int64, bool, error) {
				return 102, false, nil
			},
			Record: func(networkFaultReadinessRecovery) error {
				return errors.New("record must not run without a recovery restart")
			},
		},
	)
	require.NoError(t, err)
	require.Zero(t, stopCalls)
}

func TestRunSlowRESTClientRejectsSuccessfulResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, listener.Close()) })

	serverDone := make(chan error, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverDone <- acceptErr
			return
		}
		defer connection.Close()
		_, writeErr := connection.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\nConnection: close\r\n\r\nOK"))
		serverDone <- writeErr
	}()

	_, err = runSlowRESTClient("http://"+listener.Addr().String(), 10*time.Millisecond)
	require.ErrorContains(t, err, "successful HTTP response")
	require.NoError(t, <-serverDone)
}

func TestValidateOversizedRESTRejectionRequiresBodyLimitReason(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateOversizedRESTRejection(
		400,
		[]byte(`{"error":"error reading request body: http: request body too large"}`),
		1024,
		4096,
	))
	require.ErrorContains(t, validateOversizedRESTRejection(
		400,
		[]byte(`{"error":"invalid tx bytes"}`),
		1024,
		4096,
	), "does not identify")
	require.Error(t, validateOversizedRESTRejection(
		200,
		[]byte(`request body too large`),
		1024,
		4096,
	))
	require.Error(t, validateOversizedRESTRejection(
		400,
		[]byte(`request body too large`),
		4096,
		1024,
	))
}

func TestValidateOversizedGRPCRejectionRequiresResourceExhaustedSizeDiagnostic(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateOversizedGRPCRejection(
		status.Error(codes.ResourceExhausted, "grpc: received message larger than max (4099 vs. 1024)"),
		1024,
		4099,
	))
	require.ErrorContains(t, validateOversizedGRPCRejection(
		status.Error(codes.InvalidArgument, "invalid address"),
		1024,
		4099,
	), "ResourceExhausted")
	require.ErrorContains(t, validateOversizedGRPCRejection(
		status.Error(codes.ResourceExhausted, "quota exceeded"),
		1024,
		4099,
	), "message size")
}

func TestValidateNetworkFaultWebSocketReconnectAccountsForFaultGapAndTransactions(t *testing.T) {
	t.Parallel()

	continuity, err := validateNetworkFaultWebSocketReconnect(
		networkFaultWebSocketPhase{
			BlockHeights: []int64{10, 11},
			Transactions: []networkFaultWebSocketTransaction{{Height: 11, TxHash: "AABB"}},
		},
		networkFaultWebSocketPhase{
			BlockHeights: []int64{14, 15},
			Transactions: []networkFaultWebSocketTransaction{{Height: 15, TxHash: "CCDD"}},
		},
		[]string{"AABB", "CCDD"},
	)
	require.NoError(t, err)
	require.Equal(t, []int64{12, 13}, continuity.FaultMissingBlockHeights)
	require.Zero(t, continuity.DuplicateBlockEvents)
	require.Zero(t, continuity.DuplicateTransactionEvents)
	require.Empty(t, continuity.MissingTransactionHashes)
}

func TestValidateNetworkFaultWebSocketReconnectRejectsUnaccountedEventDefects(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		before   networkFaultWebSocketPhase
		after    networkFaultWebSocketPhase
		expected []string
		contains string
	}{
		"intra-phase block gap": {
			before: networkFaultWebSocketPhase{
				BlockHeights: []int64{10, 12},
				Transactions: []networkFaultWebSocketTransaction{{Height: 12, TxHash: "AABB"}},
			},
			after: networkFaultWebSocketPhase{
				BlockHeights: []int64{14, 15},
				Transactions: []networkFaultWebSocketTransaction{{Height: 15, TxHash: "CCDD"}},
			},
			expected: []string{"AABB", "CCDD"},
			contains: "event gap or duplicate",
		},
		"duplicate block": {
			before: networkFaultWebSocketPhase{
				BlockHeights: []int64{10, 10},
				Transactions: []networkFaultWebSocketTransaction{{Height: 10, TxHash: "AABB"}},
			},
			after: networkFaultWebSocketPhase{
				BlockHeights: []int64{14, 15},
				Transactions: []networkFaultWebSocketTransaction{{Height: 15, TxHash: "CCDD"}},
			},
			expected: []string{"AABB", "CCDD"},
			contains: "duplicate events",
		},
		"duplicate transaction": {
			before: networkFaultWebSocketPhase{
				BlockHeights: []int64{10, 11},
				Transactions: []networkFaultWebSocketTransaction{
					{Height: 11, TxHash: "AABB"},
					{Height: 11, TxHash: "AABB"},
				},
			},
			after: networkFaultWebSocketPhase{
				BlockHeights: []int64{14, 15},
				Transactions: []networkFaultWebSocketTransaction{{Height: 15, TxHash: "CCDD"}},
			},
			expected: []string{"AABB", "CCDD"},
			contains: "duplicate events",
		},
		"missing transaction": {
			before: networkFaultWebSocketPhase{
				BlockHeights: []int64{10, 11},
				Transactions: []networkFaultWebSocketTransaction{{Height: 11, TxHash: "AABB"}},
			},
			after: networkFaultWebSocketPhase{
				BlockHeights: []int64{14, 15},
			},
			expected: []string{"AABB", "CCDD"},
			contains: "missing expected hashes",
		},
		"transaction without matching block": {
			before: networkFaultWebSocketPhase{
				BlockHeights: []int64{10, 11},
				Transactions: []networkFaultWebSocketTransaction{{Height: 9, TxHash: "AABB"}},
			},
			after: networkFaultWebSocketPhase{
				BlockHeights: []int64{14, 15},
				Transactions: []networkFaultWebSocketTransaction{{Height: 15, TxHash: "CCDD"}},
			},
			expected: []string{"AABB", "CCDD"},
			contains: "has no matching block event",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := validateNetworkFaultWebSocketReconnect(test.before, test.after, test.expected)
			require.ErrorContains(t, err, test.contains)
		})
	}
}

func TestNetworkFaultRPCStatusHeightRejectsGenericHTTP200Body(t *testing.T) {
	t.Parallel()

	height, err := networkFaultRPCStatusHeight([]byte(`{"result":{"sync_info":{"latest_block_height":"17"}}}`))
	require.NoError(t, err)
	require.EqualValues(t, 17, height)
	_, err = networkFaultRPCStatusHeight([]byte(`<html>OK</html>`))
	require.ErrorContains(t, err, "decode RPC churn status")
	_, err = networkFaultRPCStatusHeight([]byte(`{"result":{"sync_info":{"latest_block_height":"0"}}}`))
	require.ErrorContains(t, err, "invalid RPC churn height")
}

func TestProbeNetworkFaultTCPConnectionFailureAcceptsForwarderEOF(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, listener.Close()) })

	serverDone := make(chan error, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverDone <- acceptErr
			return
		}
		serverDone <- connection.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	diagnostic, err := probeNetworkFaultTCPConnectionFailure(ctx, "http://"+listener.Addr().String())
	require.NoError(t, err)
	require.NotEmpty(t, diagnostic)
	require.NoError(t, <-serverDone)
}

func TestProbeNetworkFaultTCPConnectionFailureRejectsHTTPResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, listener.Close()) })

	serverDone := make(chan error, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverDone <- acceptErr
			return
		}
		defer connection.Close()
		_, writeErr := connection.Write([]byte("HTTP/1.1 503 Service Unavailable\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"))
		serverDone <- writeErr
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = probeNetworkFaultTCPConnectionFailure(ctx, "http://"+listener.Addr().String())
	require.ErrorContains(t, err, "unexpectedly returned response data")
	require.NoError(t, <-serverDone)
}

func TestNetworkFaultProxyConnectionFromIPRequiresExactJSONClient(t *testing.T) {
	t.Parallel()

	logs := []byte("" +
		`{"event":"listening","target":"validator:27656"}` + "\n" +
		`{"event":"connection-opened","connection_id":7,"client":"172.30.0.5:43123","target":"validator:27656"}` + "\n" +
		`{"event":"chunk-dropped","connection_id":7,"direction":"client-to-target","chunk":1,"bytes":128}` + "\n")
	events, err := decodeNetworkFaultProxyLogEvents(logs)
	require.NoError(t, err)
	require.Equal(t, 1, networkFaultProxyEventCount(events, "connection-opened"))
	require.Equal(t, 1, networkFaultProxyEventCount(events, "chunk-dropped"))

	connection, err := networkFaultProxyConnectionFromIP(logs, "172.30.0.5")
	require.NoError(t, err)
	require.EqualValues(t, 7, connection.ConnectionID)
	require.Equal(t, "172.30.0.5:43123", connection.Client)

	_, err = networkFaultProxyConnectionFromIP(logs, "172.30.0.6")
	require.ErrorContains(t, err, "no connection from 172.30.0.6")
	_, err = networkFaultProxyConnectionFromIP(logs, "not-an-ip")
	require.ErrorContains(t, err, "is invalid")
}

func TestDecodeNetworkFaultProxyLogEventsRejectsMalformedOrAmbiguousEvidence(t *testing.T) {
	t.Parallel()

	_, err := decodeNetworkFaultProxyLogEvents(nil)
	require.ErrorContains(t, err, "no JSON events")
	_, err = decodeNetworkFaultProxyLogEvents([]byte(`{"event":"listening"}` + "\nnot-json\n"))
	require.ErrorContains(t, err, "decode fault proxy JSON event 2")
	_, err = decodeNetworkFaultProxyLogEvents([]byte(`{"target":"validator:27656"}`))
	require.ErrorContains(t, err, "has no event name")
	_, err = networkFaultProxyConnectionFromIP(
		[]byte(`{"event":"connection-opened","connection_id":8,"client":"missing-port"}`),
		"172.30.0.5",
	)
	require.ErrorContains(t, err, "invalid client address")
}

func TestRewriteNetworkFaultAppConfigMutatesOnlyRequestedBoundaries(t *testing.T) {
	t.Parallel()
	input := []byte("[api]\nenable = true\nrpc-read-timeout = 10\nrpc-write-timeout = 10\nrpc-max-body-bytes = 1000000\n[grpc]\nenable = true\naddress = \"0.0.0.0:9090\"\n")
	output, err := rewriteNetworkFaultAppConfig(input, networkFaultEndpointMutation{
		APIEnabled:       boolPointer(false),
		ReadTimeout:      uintPointer(1),
		MaxBodyBytes:     uintPointer(1024),
		GRPCMaxRecvBytes: uintPointer(2048),
	})
	require.NoError(t, err)
	var decoded map[string]any
	_, err = toml.Decode(string(output), &decoded)
	require.NoError(t, err)
	api := decoded["api"].(map[string]any)
	grpc := decoded["grpc"].(map[string]any)
	require.Equal(t, false, api["enable"])
	require.EqualValues(t, 1, api["rpc-read-timeout"])
	require.EqualValues(t, 1024, api["rpc-max-body-bytes"])
	require.EqualValues(t, 10, api["rpc-write-timeout"])
	require.Equal(t, true, grpc["enable"])
	require.Equal(t, "0.0.0.0:9090", grpc["address"])
	require.Equal(t, "2048", grpc["max-recv-msg-size"])
}
