package types_test

import (
	"testing"

	"github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisDIDDocumentKey(t *testing.T) {
	key := types.GenesisDIDDocumentKey{DID: "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"}

	var newKey types.GenesisDIDDocumentKey
	require.NoError(t, newKey.Unmarshal(key.Marshal()))
	require.Equal(t, key, newKey)
}

func TestGenesisDIDDocumentKey_InvalidDID(t *testing.T) {
	invalidKey := types.GenesisDIDDocumentKey{DID: "invalid_did"}.Marshal()

	var key types.GenesisDIDDocumentKey
	require.Error(t, key.Unmarshal(invalidKey))
}
