package e2e_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

func proposalIDFromCommittedTx(result *harness.TxResult) (uint64, error) {
	if result == nil {
		return 0, errors.New("committed proposal transaction is required")
	}
	ids := make(map[uint64]struct{})
	for _, event := range result.Events {
		for _, attribute := range event.Attributes {
			if attribute.Key != "proposal_id" {
				continue
			}
			value := attribute.Value
			var decoded string
			if json.Unmarshal([]byte(value), &decoded) == nil {
				value = decoded
			}
			id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
			if err != nil || id == 0 {
				return 0, fmt.Errorf("invalid proposal_id event attribute %q", attribute.Value)
			}
			ids[id] = struct{}{}
		}
	}
	if len(ids) != 1 {
		return 0, fmt.Errorf("proposal transaction contains %d distinct proposal IDs, want 1", len(ids))
	}
	for id := range ids {
		return id, nil
	}
	panic("unreachable")
}

func didFromKeyStoreMetadata(contents []byte) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	var metadata map[string]json.RawMessage
	if err := decoder.Decode(&metadata); err != nil {
		return "", fmt.Errorf("decode DID key metadata: %w", err)
	}
	if len(metadata) != 1 {
		return "", errors.New("DID key metadata must contain only its public address")
	}
	var verificationMethodID string
	if err := json.Unmarshal(metadata["address"], &verificationMethodID); err != nil {
		return "", errors.New("DID key metadata has no string address")
	}
	did, fragment, found := strings.Cut(verificationMethodID, "#")
	if !found || fragment == "" || !strings.HasPrefix(did, "did:panacea:") {
		return "", fmt.Errorf("invalid Panacea verification method ID %q", verificationMethodID)
	}
	return did, nil
}
