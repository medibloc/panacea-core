package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

// StartFromExport starts a fresh, isolated chain from an exported application
// genesis. The source chain must already be stopped because the new validator
// volume receives the same test-only consensus key.
func StartFromExport(
	ctx context.Context,
	t *testing.T,
	cfg Config,
	exportedGenesis []byte,
	validatorKey []byte,
) (*Network, error) {
	t.Helper()
	if len(exportedGenesis) == 0 || len(validatorKey) == 0 {
		return nil, errors.New("export bootstrap requires both genesis and validator key")
	}
	cfg.exportedGenesis = append([]byte(nil), exportedGenesis...)
	cfg.validatorKey = append([]byte(nil), validatorKey...)
	return Start(ctx, t, cfg)
}

func configureExportBootstrap(ctx context.Context, spec *interchaintest.ChainSpec, cfg Config) error {
	hasGenesis := len(cfg.exportedGenesis) > 0
	hasValidatorKey := len(cfg.validatorKey) > 0
	if !hasGenesis && !hasValidatorKey {
		return nil
	}
	if !hasGenesis || !hasValidatorKey {
		return errors.New("export bootstrap requires both genesis and validator key")
	}
	if spec == nil {
		return errors.New("export bootstrap requires a chain spec")
	}
	if cfg.NumValidators != 1 {
		return fmt.Errorf("export bootstrap requires exactly one validator, got %d", cfg.NumValidators)
	}
	if cfg.NumFullNodes < 1 {
		return fmt.Errorf("export bootstrap requires at least one full node, got %d", cfg.NumFullNodes)
	}
	chainID, err := validateExportBootstrap(cfg.exportedGenesis, cfg.validatorKey)
	if err != nil {
		return err
	}
	exportedGenesis := append([]byte(nil), cfg.exportedGenesis...)
	validatorKey := append([]byte(nil), cfg.validatorKey...)

	// An export is a continuation of the same chain at a later initial height,
	// not a genesis migration to a new identity.
	spec.ChainConfig.ChainID = chainID
	spec.ChainConfig.SkipGenTx = true
	// Interchaintest mutates the assembled genesis after PreGenesis: it adds
	// configured wallets and replaces every JSON token named "stake" with the
	// configured denomination. The latter corrupts the SDK distribution field
	// starting_info.stake. Restore the byte-exact export in the final modifier
	// hook so neither mutation reaches the fresh chain.
	spec.ChainConfig.ModifyGenesis = func(ibc.ChainConfig, []byte) ([]byte, error) {
		return append([]byte(nil), exportedGenesis...), nil
	}
	spec.ChainConfig.PreGenesis = func(chain ibc.Chain) error {
		cosmosChain, ok := chain.(*cosmos.CosmosChain)
		if !ok {
			return fmt.Errorf("export bootstrap requires CosmosChain, got %T", chain)
		}
		if len(cosmosChain.Validators) != 1 {
			return fmt.Errorf("export bootstrap initialized %d validators, want 1", len(cosmosChain.Validators))
		}
		if err := cosmosChain.Validators[0].OverwriteGenesisFile(ctx, exportedGenesis); err != nil {
			return fmt.Errorf("install exported genesis: %w", err)
		}
		// InitFullNodeFiles already created a fresh height-zero signing state.
		// Replace only the key; never copy the source priv_validator_state.json.
		if err := cosmosChain.Validators[0].OverwritePrivValFile(ctx, validatorKey); err != nil {
			return fmt.Errorf("install exported validator key: %w", err)
		}
		return nil
	}
	return nil
}

func validateExportBootstrap(source, validatorKey []byte) (string, error) {
	document, err := decodeJSONObject(source, "imported genesis")
	if err != nil {
		return "", err
	}
	if err := validateImportedInitialHeight(document["initial_height"]); err != nil {
		return "", err
	}
	if !isJSONObject(document["app_state"]) {
		return "", errors.New("imported genesis app_state must be a JSON object")
	}
	var chainID string
	if err := json.Unmarshal(document["chain_id"], &chainID); err != nil {
		return "", fmt.Errorf("decode imported genesis chain_id: %w", err)
	}
	if strings.TrimSpace(chainID) == "" || len(chainID) > 50 {
		return "", errors.New("imported genesis chain_id must contain 1 to 50 bytes")
	}
	if err := validateImportedValidatorKey(source, validatorKey); err != nil {
		return "", fmt.Errorf("private validator identity is absent from or mismatched with exported genesis: %w", err)
	}
	return chainID, nil
}

func validateImportedInitialHeight(raw json.RawMessage) error {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return errors.New("imported genesis is missing initial_height")
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		height, parseErr := strconv.ParseInt(text, 10, 64)
		if parseErr != nil || height <= 0 {
			return fmt.Errorf("imported genesis has invalid initial_height %q", text)
		}
		return nil
	}
	var height int64
	if err := json.Unmarshal(raw, &height); err != nil || height <= 0 {
		return fmt.Errorf("imported genesis has invalid initial_height %s", bytes.TrimSpace(raw))
	}
	return nil
}

func isJSONObject(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) >= 2 && trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}'
}
