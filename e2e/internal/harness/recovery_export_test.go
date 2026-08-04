package harness

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalGenesisDigestIgnoresJSONFormattingAndObjectOrder(t *testing.T) {
	first := []byte(`{"chain_id":"panacea-test","app_state":{"bank":{"supply":[]}},"initial_height":"42"}`)
	second := []byte(`{
  "initial_height": "42",
  "app_state": {"bank": {"supply": []}},
  "chain_id": "panacea-test"
}`)

	firstDigest, err := CanonicalGenesisDigest(first)
	require.NoError(t, err)
	secondDigest, err := CanonicalGenesisDigest(second)
	require.NoError(t, err)
	require.Equal(t, firstDigest, secondDigest)
}

func TestExportedGenesisAppHeightResolvesNextInitialHeight(t *testing.T) {
	height, err := exportedGenesisAppHeight([]byte(`{"initial_height":"42"}`))
	require.NoError(t, err)
	require.Equal(t, int64(41), height)

	height, err = exportedGenesisAppHeight([]byte(`{"initial_height":42}`))
	require.NoError(t, err)
	require.Equal(t, int64(41), height)

	_, err = exportedGenesisAppHeight([]byte(`{"initial_height":"0"}`))
	require.ErrorContains(t, err, "initial_height")
}

func TestPrivateValidatorLastSignedHeightAcceptsCometEncoding(t *testing.T) {
	height, err := privateValidatorLastSignedHeight([]byte(`{"height":"41","round":0,"step":3}`))
	require.NoError(t, err)
	require.Equal(t, int64(41), height)

	height, err = privateValidatorLastSignedHeight([]byte(`{"height":41,"round":0,"step":3}`))
	require.NoError(t, err)
	require.Equal(t, int64(41), height)

	_, err = privateValidatorLastSignedHeight([]byte(`{"height":"-1"}`))
	require.ErrorContains(t, err, "height")
}
