package did

import (
	"testing"

	"github.com/medibloc/panacea-core/x/did/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisDIDDocumentKey(t *testing.T) {
	key := GenesisDIDDocumentKey{DID: types.DID("did:panacea:testnet:KS5zGZt66Me8MCctZBYrP")}

	var newKey GenesisDIDDocumentKey
	require.NoError(t, newKey.Unmarshal(key.Marshal()))
	require.Equal(t, key, newKey)
}

func TestGenesisDIDDocumentKey_InvalidDID(t *testing.T) {
	invalidKey := GenesisDIDDocumentKey{DID: types.DID("invalid_did")}.Marshal()

	var key GenesisDIDDocumentKey
	require.Error(t, key.Unmarshal(invalidKey))
}
