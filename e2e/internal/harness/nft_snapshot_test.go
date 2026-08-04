package harness_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func TestSemanticJSONCanonicalizesEquivalentQueryResponses(t *testing.T) {
	firstInput := []byte(`{"supply":{"amount":"1"},"records":[{"id":"nft.1"}]}`)
	secondInput := []byte("{\n  \"records\": [{\"id\": \"nft.1\"}],\n  \"supply\": {\"amount\": \"1\"}\n}")

	first, err := harness.NewSemanticJSON(firstInput)
	require.NoError(t, err)
	second, err := harness.NewSemanticJSON(secondInput)
	require.NoError(t, err)

	require.Equal(t, first, second)
	require.JSONEq(t, string(first), string(second))

	firstInput[0] = '['
	require.JSONEq(t, string(second), string(first))
}

func TestSemanticJSONRejectsInvalidOrTrailingResponses(t *testing.T) {
	for _, input := range [][]byte{
		[]byte(`{"unterminated":`),
		[]byte(`{"first":true} {"second":true}`),
	} {
		value, err := harness.NewSemanticJSON(input)

		require.Error(t, err)
		require.Nil(t, value)
	}
}
