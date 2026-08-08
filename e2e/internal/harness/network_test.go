package harness

import (
	"strings"
	"testing"
)

func TestValidateCommittedTx(t *testing.T) {
	const hash = "AABBCCDD"
	tests := []struct {
		name       string
		result     string
		wantHeight int64
		wantError  string
	}{
		{name: "numeric code", result: `{"height":"17","txhash":"aabbccdd","code":0}`, wantHeight: 17},
		{name: "quoted code", result: `{"height":18,"txhash":"AABBCCDD","code":"0"}`, wantHeight: 18},
		{name: "missing code", result: `{"height":"17","txhash":"AABBCCDD"}`, wantError: "missing code"},
		{name: "deliver failure", result: `{"height":"17","txhash":"AABBCCDD","code":7}`, wantError: "non-zero code 7"},
		{name: "wrong hash", result: `{"height":"17","txhash":"FFFF","code":0}`, wantError: "hash"},
		{name: "missing height", result: `{"txhash":"AABBCCDD","code":0}`, wantError: "height"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			height, err := validateCommittedTx([]byte(test.result), hash)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("validateCommittedTx error = %v, want containing %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("validateCommittedTx: %v", err)
			}
			if height != test.wantHeight {
				t.Fatalf("height = %d, want %d", height, test.wantHeight)
			}
		})
	}
}

func TestNormalizeHostAddress(t *testing.T) {
	tests := map[string]string{
		"http://0.0.0.0:26657":  "http://127.0.0.1:26657",
		"http://[::]:1317":      "http://127.0.0.1:1317",
		"http://127.0.0.1:9090": "http://127.0.0.1:9090",
	}
	for input, want := range tests {
		got, err := normalizeHostAddress(input)
		if err != nil {
			t.Fatalf("normalizeHostAddress(%q): %v", input, err)
		}
		if got != want {
			t.Fatalf("normalizeHostAddress(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestDetachedCometStateSyncRPCClientNormalizesWildcardBinding(t *testing.T) {
	client, address, err := newDetachedCometStateSyncRPCClient("http://0.0.0.0:26657")
	if err != nil {
		t.Fatalf("newDetachedCometStateSyncRPCClient: %v", err)
	}
	if client == nil {
		t.Fatal("newDetachedCometStateSyncRPCClient returned a nil client")
	}
	if address != "http://127.0.0.1:26657" {
		t.Fatalf("RPC address = %q, want %q", address, "http://127.0.0.1:26657")
	}
}
