package params

import (
	"context"
	"encoding/json"
	"fmt"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"
)

// legacyAminoJSONCompatHandler preserves the v2.2.1 bare JSON representation
// for selected top-level messages. Bare JSON does not cryptographically bind a
// message's type, so this handler must be removed after clients migrate to
// SIGN_MODE_DIRECT.
type legacyAminoJSONCompatHandler struct {
	base                txsigning.SignModeHandler
	bareMessageTypeURLs map[string]struct{}
}

func newLegacyAminoJSONCompatHandler(base txsigning.SignModeHandler, typeURLs []string) *legacyAminoJSONCompatHandler {
	bareMessageTypeURLs := make(map[string]struct{}, len(typeURLs))
	for _, typeURL := range typeURLs {
		bareMessageTypeURLs[typeURL] = struct{}{}
	}

	return &legacyAminoJSONCompatHandler{
		base:                base,
		bareMessageTypeURLs: bareMessageTypeURLs,
	}
}

func (h legacyAminoJSONCompatHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
}

func (h legacyAminoJSONCompatHandler) GetSignBytes(
	ctx context.Context,
	signerData txsigning.SignerData,
	txData txsigning.TxData,
) ([]byte, error) {
	signBytes, err := h.base.GetSignBytes(ctx, signerData, txData)
	if err != nil {
		return nil, err
	}

	if txData.Body == nil || !h.containsBareMessage(txData) {
		return signBytes, nil
	}

	var signDoc map[string]json.RawMessage
	if err := json.Unmarshal(signBytes, &signDoc); err != nil {
		return nil, fmt.Errorf("decode amino JSON sign document: %w", err)
	}

	var messages []json.RawMessage
	if err := json.Unmarshal(signDoc["msgs"], &messages); err != nil {
		return nil, fmt.Errorf("decode amino JSON messages: %w", err)
	}
	if len(messages) != len(txData.Body.Messages) {
		return nil, fmt.Errorf(
			"amino JSON message count mismatch: sign document has %d, transaction has %d",
			len(messages),
			len(txData.Body.Messages),
		)
	}

	for i, message := range txData.Body.Messages {
		if message == nil {
			continue
		}
		if _, ok := h.bareMessageTypeURLs[message.TypeUrl]; !ok {
			continue
		}

		value, err := bareAminoJSONValue(messages[i])
		if err != nil {
			return nil, fmt.Errorf("decode amino JSON message %d (%s): %w", i, message.TypeUrl, err)
		}
		messages[i] = value
	}

	signDoc["msgs"], err = json.Marshal(messages)
	if err != nil {
		return nil, fmt.Errorf("encode amino JSON messages: %w", err)
	}

	compatibleSignBytes, err := json.Marshal(signDoc)
	if err != nil {
		return nil, fmt.Errorf("encode amino JSON sign document: %w", err)
	}
	return compatibleSignBytes, nil
}

func (h legacyAminoJSONCompatHandler) containsBareMessage(txData txsigning.TxData) bool {
	for _, message := range txData.Body.Messages {
		if message == nil {
			continue
		}
		if _, ok := h.bareMessageTypeURLs[message.TypeUrl]; ok {
			return true
		}
	}
	return false
}

func bareAminoJSONValue(message json.RawMessage) (json.RawMessage, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(message, &envelope); err != nil {
		return nil, err
	}

	var aminoType string
	if err := json.Unmarshal(envelope["type"], &aminoType); err != nil {
		return nil, fmt.Errorf("decode type: %w", err)
	}
	if aminoType == "" {
		return nil, fmt.Errorf("empty type")
	}

	value, ok := envelope["value"]
	if !ok || len(value) == 0 {
		return nil, fmt.Errorf("missing value")
	}
	return value, nil
}

var _ txsigning.SignModeHandler = (*legacyAminoJSONCompatHandler)(nil)
