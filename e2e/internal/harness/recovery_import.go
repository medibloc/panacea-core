package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type importedConsensusValidator struct {
	Address string `json:"address"`
	PubKey  struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
	Power json.RawMessage `json:"power"`
}

func validateImportedValidatorKey(genesis, validatorKey []byte) error {
	document, err := decodeJSONObject(genesis, "imported genesis")
	if err != nil {
		return err
	}
	validators, err := importedGenesisValidators(document)
	if err != nil {
		return err
	}
	if len(validators) != 1 {
		return fmt.Errorf("imported genesis must contain exactly one consensus validator, got %d", len(validators))
	}
	var key struct {
		Address string `json:"address"`
		PubKey  struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"pub_key"`
		PrivKey struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"priv_key"`
	}
	decoder := json.NewDecoder(bytes.NewReader(validatorKey))
	if err := decoder.Decode(&key); err != nil {
		return fmt.Errorf("decode imported validator key: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("decode imported validator key: multiple JSON values")
		}
		return fmt.Errorf("decode trailing imported validator key: %w", err)
	}
	if strings.TrimSpace(key.PrivKey.Type) == "" || strings.TrimSpace(key.PrivKey.Value) == "" {
		return errors.New("imported validator key has no private key material")
	}
	if strings.TrimSpace(key.Address) == "" ||
		strings.TrimSpace(key.PubKey.Type) == "" ||
		strings.TrimSpace(key.PubKey.Value) == "" {
		return errors.New("imported validator key has no complete public identity")
	}
	validator := validators[0]
	if strings.TrimSpace(validator.Address) == "" ||
		strings.TrimSpace(validator.PubKey.Type) == "" ||
		strings.TrimSpace(validator.PubKey.Value) == "" {
		return errors.New("exported consensus validator has no complete public identity")
	}
	if !strings.EqualFold(key.Address, validator.Address) ||
		key.PubKey.Type != validator.PubKey.Type ||
		key.PubKey.Value != validator.PubKey.Value {
		return errors.New("imported validator key does not match the exported consensus validator")
	}
	return nil
}

func decodeJSONObject(contents []byte, label string) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	var document map[string]json.RawMessage
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode %s: %w", label, err)
	}
	if document == nil {
		return nil, fmt.Errorf("%s must be a JSON object", label)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode %s: multiple JSON values", label)
		}
		return nil, fmt.Errorf("decode trailing %s: %w", label, err)
	}
	return document, nil
}

func importedGenesisValidators(document map[string]json.RawMessage) ([]importedConsensusValidator, error) {
	if rawConsensus := document["consensus"]; len(rawConsensus) > 0 && !bytes.Equal(bytes.TrimSpace(rawConsensus), []byte("null")) {
		var consensus struct {
			Validators []importedConsensusValidator `json:"validators"`
		}
		if err := json.Unmarshal(rawConsensus, &consensus); err != nil {
			return nil, fmt.Errorf("decode imported consensus validators: %w", err)
		}
		return consensus.Validators, nil
	}
	var validators []importedConsensusValidator
	if err := json.Unmarshal(document["validators"], &validators); err != nil {
		return nil, fmt.Errorf("decode imported validators: %w", err)
	}
	return validators, nil
}
