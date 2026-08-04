package harness

import (
	"fmt"
	"strings"
)

// ClassifyUnsupportedDBBackendStartup recognizes the explicit Cosmos SDK
// v0.50 rejection emitted before a node can open an unsupported application
// database. A timeout or generic container failure is deliberately not enough.
func ClassifyUnsupportedDBBackendStartup(backend, diagnostics string) (string, bool) {
	backend = strings.TrimSpace(backend)
	if backend == "" {
		return "", false
	}
	marker := fmt.Sprintf(
		"invalid app-db-backend %q, use %q, %q, %q instead",
		backend,
		"goleveldb",
		"pebbledb",
		"rocksdb",
	)
	return marker, strings.Contains(diagnostics, marker)
}
