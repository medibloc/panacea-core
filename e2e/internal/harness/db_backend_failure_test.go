package harness

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyUnsupportedDBBackendStartupRequiresClearSDKDiagnostic(t *testing.T) {
	t.Parallel()

	diagnostic := `failed to start chain: panic: invalid app-db-backend "badgerdb", use "goleveldb", "pebbledb", "rocksdb" instead`
	marker, ok := ClassifyUnsupportedDBBackendStartup("badgerdb", diagnostic)
	require.True(t, ok)
	require.Equal(t, `invalid app-db-backend "badgerdb", use "goleveldb", "pebbledb", "rocksdb" instead`, marker)

	_, ok = ClassifyUnsupportedDBBackendStartup("badgerdb", "context deadline exceeded")
	require.False(t, ok, "a bounded timeout without the SDK diagnostic is not a clear backend rejection")
}
