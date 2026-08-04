package harness

import (
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunConcurrentTxBatchStartsEveryBroadcastBeforeReturning(t *testing.T) {
	t.Parallel()

	const batchSize = 3
	started := make(chan int, batchSize)
	release := make(chan struct{})
	done := make(chan []int, 1)

	go func() {
		done <- runConcurrentTxBatch(batchSize, func(index int) int {
			started <- index
			<-release
			return index
		})
	}()

	observed := make([]int, 0, batchSize)
	for len(observed) < batchSize {
		select {
		case index := <-started:
			observed = append(observed, index)
		case <-time.After(time.Second):
			t.Fatalf("only %d/%d broadcasts started before the batch blocked", len(observed), batchSize)
		}
	}
	close(release)

	select {
	case results := <-done:
		require.Equal(t, []int{0, 1, 2}, results)
	case <-time.After(time.Second):
		t.Fatal("concurrent transaction batch did not finish")
	}
	sort.Ints(observed)
	require.Equal(t, []int{0, 1, 2}, observed)
}

func TestRunConcurrentTxBatchKeepsInputOrder(t *testing.T) {
	t.Parallel()

	results := runConcurrentTxBatch(3, func(index int) string {
		time.Sleep(time.Duration(3-index) * time.Millisecond)
		return string(rune('a' + index))
	})

	require.Equal(t, []string{"a", "b", "c"}, results)
}

func TestClassifyExpectedCLIRejection(t *testing.T) {
	t.Parallel()

	require.NoError(t, classifyExpectedCLIRejection(
		errors.New("exit code 1"),
		nil,
		[]byte("hrp does not match bech32 prefix: expected 'panacea' got 'cosmos'"),
		"hrp does not match bech32 prefix",
	))
	require.ErrorContains(t, classifyExpectedCLIRejection(nil, []byte(`{"txhash":"ABC"}`), nil, "hrp does not match"), "unexpectedly succeeded")
	require.ErrorContains(t, classifyExpectedCLIRejection(errors.New("exit code 1"), nil, []byte("connection refused"), "hrp does not match"), "missing expected diagnostic")
	require.ErrorContains(t, classifyExpectedCLIRejection(
		errors.New("exit code 1: NFT data is invalid"),
		[]byte(`{"txhash":"ABC"}`),
		nil,
		"NFT data",
	), "returned transaction output")
}

func TestClassifyExpectedCheckTxFailure(t *testing.T) {
	t.Parallel()

	rejected := TxResult{
		Height:    "0",
		TxHash:    "ABC123",
		Codespace: "sdk",
		Code:      32,
		RawLog:    "account sequence mismatch",
	}
	require.NoError(t, classifyExpectedCheckTxFailure(rejected, "sdk", 32))

	passed := rejected
	passed.Code = 0
	passed.Codespace = ""
	require.ErrorContains(t, classifyExpectedCheckTxFailure(passed, "sdk", 32), "unexpectedly passed CheckTx")

	wrongCode := rejected
	wrongCode.Code = 5
	require.ErrorContains(t, classifyExpectedCheckTxFailure(wrongCode, "sdk", 32), "want codespace=sdk code=32")

	committed := rejected
	committed.Height = "17"
	require.ErrorContains(t, classifyExpectedCheckTxFailure(committed, "sdk", 32), "committed height")

	malformedHeight := rejected
	malformedHeight.Height = "not-a-height"
	require.ErrorContains(t, classifyExpectedCheckTxFailure(malformedHeight, "sdk", 32), "broadcast height")
}

func TestParseTxResultPreservesCheckTxAndCommittedEvidence(t *testing.T) {
	broadcast, err := parseTxResult([]byte(`{
      "height":"0",
      "txhash":"ABC123",
      "codespace":"",
      "code":0,
      "raw_log":""
    }`))
	require.NoError(t, err)
	require.Equal(t, "ABC123", broadcast.TxHash)
	require.Zero(t, broadcast.Code)
	require.Equal(t, "0", broadcast.Height)

	committed, err := parseTxResult([]byte(`{
      "height":"17",
      "txhash":"ABC123",
      "codespace":"",
      "code":0,
      "raw_log":"",
      "events":[{
        "type":"panacea.nft.v1.EventClassCreated",
        "attributes":[
          {"key":"class_id","value":"\"panacea1creator:certificate\"","index":true},
          {"key":"creator","value":"\"panacea1creator\"","index":true}
        ]
      }]
    }`))
	require.NoError(t, err)
	require.Equal(t, int64(17), committed.HeightInt64())
	event, ok := committed.FindEvent("panacea.nft.v1.EventClassCreated")
	require.True(t, ok)
	require.Equal(t, "panacea1creator:certificate", event.Attribute("class_id"))
	require.Equal(t, "panacea1creator", event.Attribute("creator"))
}

func TestTxLifecycleResultPreservesBroadcastAndWaitTxCompatibility(t *testing.T) {
	t.Parallel()

	checkTxAccepted := &TxResult{Height: "0", TxHash: "ACCEPTED", Code: 0}
	checkTxRejected := &TxResult{Height: "0", TxHash: "REJECTED", Code: 7}
	committed := &TxResult{Height: "17", TxHash: "ACCEPTED", Code: 9}

	tests := []struct {
		name      string
		lifecycle *TxLifecycleResult
		want      *TxResult
	}{
		{name: "nil lifecycle", lifecycle: nil, want: nil},
		{name: "missing CheckTx evidence", lifecycle: &TxLifecycleResult{}, want: nil},
		{name: "accepted but not observed", lifecycle: &TxLifecycleResult{CheckTx: checkTxAccepted}, want: nil},
		{name: "CheckTx rejection", lifecycle: &TxLifecycleResult{CheckTx: checkTxRejected}, want: checkTxRejected},
		{name: "committed result wins", lifecycle: &TxLifecycleResult{CheckTx: checkTxAccepted, Committed: committed}, want: committed},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Same(t, test.want, test.lifecycle.Result())
		})
	}
}

func TestParseTxResultRejectsMissingHashAndInvalidHeight(t *testing.T) {
	_, err := parseTxResult([]byte(`{"height":"1","code":0}`))
	require.ErrorContains(t, err, "txhash")

	_, err = parseTxResult([]byte(`{"height":"1","txhash":"ABC"}`))
	require.ErrorContains(t, err, "code")

	result, err := parseTxResult([]byte(`{"height":"not-a-height","txhash":"ABC","code":0}`))
	require.NoError(t, err)
	require.Equal(t, int64(0), result.HeightInt64())
}

func TestTxEventAttributeSupportsQuotedAndPlainValues(t *testing.T) {
	event := TxEvent{Attributes: []TxEventAttribute{
		{Key: "quoted", Value: `"value"`},
		{Key: "plain", Value: "value"},
	}}
	require.Equal(t, "value", event.Attribute("quoted"))
	require.Equal(t, "value", event.Attribute("plain"))
	require.Empty(t, event.Attribute("missing"))
}
