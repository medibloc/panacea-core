package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisTokenKey(t *testing.T) {
	key := GenesisTokenKey{Symbol: Symbol("ABC-0EA")}

	var newKey GenesisTokenKey
	require.NoError(t, newKey.Unmarshal(key.Marshal()))
	require.Equal(t, key, newKey)
}

func TestGenesisDIDDocumentKey_InvalidDID(t *testing.T) {
	invalidKey := GenesisTokenKey{Symbol: Symbol("invalid_symbol")}.Marshal()

	var key GenesisTokenKey
	require.Error(t, key.Unmarshal(invalidKey))
}
