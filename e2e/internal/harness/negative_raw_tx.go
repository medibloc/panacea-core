package harness

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

// RawProtoMessage is an sdk.Msg envelope whose protobuf payload is supplied
// verbatim. It exists for negative wire-compatibility tests that the normal
// JSON CLI correctly refuses to construct.
type RawProtoMessage struct {
	TypeURL string
	Value   []byte
}

// RawTxRequest describes one directly encoded signed transaction. The signer
// mnemonic is consumed in memory and never included in artifacts.
type RawTxRequest struct {
	Signer               ibc.Wallet
	SignerAccountAddress string
	FeePayer             ibc.Wallet
	Message              RawProtoMessage
	GasLimit             uint64
	FeeAmount            sdkmath.Int
	// Sequence overrides the queried signer sequence when a test must sign a
	// deliberately stale transaction. Nil uses the current on-chain value.
	Sequence *uint64
}

func parseRawBroadcastResponse(contents []byte) (TxResult, error) {
	var response struct {
		Result *struct {
			Code      uint32 `json:"code"`
			Log       string `json:"log"`
			Codespace string `json:"codespace"`
			Hash      string `json:"hash"`
		} `json:"result"`
		Error *struct {
			Code    int             `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal(contents, &response); err != nil {
		return TxResult{}, fmt.Errorf("decode raw transaction broadcast JSON: %w", err)
	}
	if response.Error != nil {
		return TxResult{}, fmt.Errorf(
			"raw transaction RPC error %d: %s: %s",
			response.Error.Code,
			response.Error.Message,
			response.Error.Data,
		)
	}
	if response.Result == nil {
		return TxResult{}, errors.New("raw transaction broadcast returned no result")
	}
	if strings.TrimSpace(response.Result.Hash) == "" {
		return TxResult{}, errors.New("raw transaction broadcast returned no hash")
	}
	return TxResult{
		Height:    "0",
		TxHash:    response.Result.Hash,
		Codespace: response.Result.Codespace,
		Code:      response.Result.Code,
		RawLog:    response.Result.Log,
		Raw:       append(json.RawMessage(nil), contents...),
	}, nil
}

type rawTxBuildRequest struct {
	Signer                ibc.Wallet
	Message               RawProtoMessage
	ChainID               string
	CoinType              uint32
	Denom                 string
	GasLimit              uint64
	FeeAmount             sdkmath.Int
	AccountNumber         uint64
	Sequence              uint64
	FeePayer              ibc.Wallet
	FeePayerAccountNumber uint64
	FeePayerSequence      uint64
}

func deriveRawTxPrivateKey(wallet ibc.Wallet, coinType uint32) (cryptotypes.PrivKey, error) {
	if wallet == nil || strings.TrimSpace(wallet.Mnemonic()) == "" {
		return nil, errors.New("raw transaction signer mnemonic is required")
	}
	hdPath := hd.CreateHDPath(coinType, 0, 0).String()
	derived, err := hd.Secp256k1.Derive()(wallet.Mnemonic(), "", hdPath)
	if err != nil {
		return nil, fmt.Errorf("derive raw transaction signer: %w", err)
	}
	privateKey := hd.Secp256k1.Generate()(derived)
	if !bytes.Equal(privateKey.PubKey().Address(), wallet.Address()) {
		return nil, errors.New("derived raw transaction signer does not match wallet address")
	}
	return privateKey, nil
}

func buildSignedRawTx(request rawTxBuildRequest) ([]byte, error) {
	if strings.TrimSpace(request.Message.TypeURL) == "" {
		return nil, errors.New("raw transaction message type URL is required")
	}
	if strings.TrimSpace(request.ChainID) == "" {
		return nil, errors.New("raw transaction chain ID is required")
	}
	if strings.TrimSpace(request.Denom) == "" {
		return nil, errors.New("raw transaction fee denom is required")
	}
	if request.GasLimit == 0 {
		return nil, errors.New("raw transaction gas limit must be positive")
	}
	if !request.FeeAmount.IsPositive() {
		return nil, errors.New("raw transaction fee amount must be positive")
	}

	privateKey, err := deriveRawTxPrivateKey(request.Signer, request.CoinType)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := proto.Marshal(&txtypes.TxBody{
		Messages: []*types.Any{{
			TypeUrl: request.Message.TypeURL,
			Value:   append([]byte(nil), request.Message.Value...),
		}},
	})
	if err != nil {
		return nil, fmt.Errorf("encode raw transaction body: %w", err)
	}
	publicKey, err := types.NewAnyWithValue(privateKey.PubKey())
	if err != nil {
		return nil, fmt.Errorf("encode raw transaction public key: %w", err)
	}
	signerInfos := []*txtypes.SignerInfo{{
		PublicKey: publicKey,
		ModeInfo: &txtypes.ModeInfo{Sum: &txtypes.ModeInfo_Single_{
			Single: &txtypes.ModeInfo_Single{Mode: signing.SignMode_SIGN_MODE_DIRECT},
		}},
		Sequence: request.Sequence,
	}}
	type signingSpec struct {
		privateKey    cryptotypes.PrivKey
		accountNumber uint64
	}
	signingSpecs := []signingSpec{{
		privateKey:    privateKey,
		accountNumber: request.AccountNumber,
	}}
	feePayerAddress := ""
	if request.FeePayer != nil {
		feePayerPrivateKey, deriveErr := deriveRawTxPrivateKey(request.FeePayer, request.CoinType)
		if deriveErr != nil {
			return nil, fmt.Errorf("derive raw transaction fee payer: %w", deriveErr)
		}
		feePayerPublicKey, anyErr := types.NewAnyWithValue(feePayerPrivateKey.PubKey())
		if anyErr != nil {
			return nil, fmt.Errorf("encode raw transaction fee payer public key: %w", anyErr)
		}
		signerInfos = append(signerInfos, &txtypes.SignerInfo{
			PublicKey: feePayerPublicKey,
			ModeInfo: &txtypes.ModeInfo{Sum: &txtypes.ModeInfo_Single_{
				Single: &txtypes.ModeInfo_Single{Mode: signing.SignMode_SIGN_MODE_DIRECT},
			}},
			Sequence: request.FeePayerSequence,
		})
		signingSpecs = append(signingSpecs, signingSpec{
			privateKey:    feePayerPrivateKey,
			accountNumber: request.FeePayerAccountNumber,
		})
		feePayerAddress = request.FeePayer.FormattedAddress()
	}
	authInfoBytes, err := proto.Marshal(&txtypes.AuthInfo{
		SignerInfos: signerInfos,
		Fee: &txtypes.Fee{
			Amount:   sdk.NewCoins(sdk.NewCoin(request.Denom, request.FeeAmount)),
			GasLimit: request.GasLimit,
			Payer:    feePayerAddress,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("encode raw transaction auth info: %w", err)
	}
	signatures := make([][]byte, 0, len(signingSpecs))
	for _, signer := range signingSpecs {
		signBytes, marshalErr := proto.Marshal(&txtypes.SignDoc{
			BodyBytes:     bodyBytes,
			AuthInfoBytes: authInfoBytes,
			ChainId:       request.ChainID,
			AccountNumber: signer.accountNumber,
		})
		if marshalErr != nil {
			return nil, fmt.Errorf("encode raw transaction sign doc: %w", marshalErr)
		}
		signature, signErr := signer.privateKey.Sign(signBytes)
		if signErr != nil {
			return nil, fmt.Errorf("sign raw transaction: %w", signErr)
		}
		signatures = append(signatures, signature)
	}
	encoded, err := proto.Marshal(&txtypes.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		Signatures:    signatures,
	})
	if err != nil {
		return nil, fmt.Errorf("encode signed raw transaction: %w", err)
	}
	return encoded, nil
}

// BroadcastRawTx signs a verbatim sdk.Msg protobuf payload, broadcasts it by
// CometBFT JSON-RPC, and returns either the CheckTx rejection or the committed
// DeliverTx result. It is intentionally limited to negative wire-boundary
// tests that cannot be expressed through the canonicalizing JSON CLI.
func (n *Network) BroadcastRawTx(
	ctx context.Context,
	step string,
	request RawTxRequest,
) (*TxResult, error) {
	if n == nil || n.Chain == nil {
		return nil, errors.New("raw transaction network is required")
	}
	if strings.TrimSpace(step) == "" {
		return nil, errors.New("raw transaction step is required")
	}
	if len(n.Chain.FullNodes) == 0 {
		return nil, errors.New("raw transaction requires a full node")
	}
	if request.Signer == nil {
		return nil, errors.New("raw transaction signer is required")
	}

	signerAccountAddress := strings.TrimSpace(request.SignerAccountAddress)
	if signerAccountAddress == "" {
		signerAccountAddress = request.Signer.FormattedAddress()
	}
	accountClient := authtypes.NewQueryClient(n.Chain.FullNodes[0].GrpcConn)
	account, err := accountClient.AccountInfo(
		ctx,
		&authtypes.QueryAccountInfoRequest{Address: signerAccountAddress},
	)
	if err != nil {
		return nil, fmt.Errorf("query raw transaction signer account: %w", err)
	}
	if account.Info == nil {
		return nil, errors.New("raw transaction signer query returned no account info")
	}
	var feePayerAccount *authtypes.QueryAccountInfoResponse
	if request.FeePayer != nil {
		if request.FeePayer.FormattedAddress() == signerAccountAddress {
			return nil, errors.New("explicit raw transaction fee payer must differ from the message signer account")
		}
		feePayerAccount, err = accountClient.AccountInfo(
			ctx,
			&authtypes.QueryAccountInfoRequest{Address: request.FeePayer.FormattedAddress()},
		)
		if err != nil {
			return nil, fmt.Errorf("query raw transaction fee payer account: %w", err)
		}
		if feePayerAccount.Info == nil {
			return nil, errors.New("raw transaction fee payer query returned no account info")
		}
	}
	coinType, err := strconv.ParseUint(n.Chain.Config().CoinType, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse raw transaction coin type: %w", err)
	}
	signingSequence := account.Info.Sequence
	if request.Sequence != nil {
		signingSequence = *request.Sequence
	}
	buildRequest := rawTxBuildRequest{
		Signer:        request.Signer,
		Message:       request.Message,
		ChainID:       n.Chain.Config().ChainID,
		CoinType:      uint32(coinType),
		Denom:         n.Chain.Config().Denom,
		GasLimit:      request.GasLimit,
		FeeAmount:     request.FeeAmount,
		AccountNumber: account.Info.AccountNumber,
		Sequence:      signingSequence,
		FeePayer:      request.FeePayer,
	}
	if feePayerAccount != nil {
		buildRequest.FeePayerAccountNumber = feePayerAccount.Info.AccountNumber
		buildRequest.FeePayerSequence = feePayerAccount.Info.Sequence
	}
	encoded, err := buildSignedRawTx(buildRequest)
	if err != nil {
		return nil, err
	}

	requestID := fmt.Sprintf("raw-%s-%d", step, time.Now().UTC().UnixNano())
	requestEvidence := map[string]any{
		"request_id":             requestID,
		"recorded_at":            time.Now().UTC(),
		"step":                   step,
		"signing_key_address":    request.Signer.FormattedAddress(),
		"signer_account_address": signerAccountAddress,
		"message_type_url":       request.Message.TypeURL,
		"message_value_base64":   base64.StdEncoding.EncodeToString(request.Message.Value),
		"gas_limit":              request.GasLimit,
		"fee_amount":             request.FeeAmount.String() + n.Chain.Config().Denom,
		"account_number":         account.Info.AccountNumber,
		"account_sequence":       account.Info.Sequence,
		"signed_sequence":        signingSequence,
	}
	if feePayerAccount != nil {
		requestEvidence["fee_payer"] = request.FeePayer.FormattedAddress()
		requestEvidence["fee_payer_account_number"] = feePayerAccount.Info.AccountNumber
		requestEvidence["fee_payer_sequence"] = feePayerAccount.Info.Sequence
	}
	if err := n.artifacts.appendJSONLine("tx/raw-requests.jsonl", requestEvidence); err != nil {
		return nil, fmt.Errorf("record raw transaction request: %w", err)
	}

	rpcAddress, err := n.FullNodeHostAddress(ctx, "26657/tcp")
	if err != nil {
		return nil, err
	}
	rpcPayload, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      requestID,
		"method":  "broadcast_tx_sync",
		"params": map[string]string{
			"tx": base64.StdEncoding.EncodeToString(encoded),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("encode raw transaction RPC request: %w", err)
	}
	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(rpcAddress, "/"),
		bytes.NewReader(rpcPayload),
	)
	if err != nil {
		return nil, fmt.Errorf("create raw transaction RPC request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("broadcast raw transaction: %w", err)
	}
	defer response.Body.Close()
	responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, queryResponseMaxBytes+1))
	if readErr != nil {
		return nil, fmt.Errorf("read raw transaction broadcast response: %w", readErr)
	}
	if len(responseBody) > queryResponseMaxBytes {
		return nil, fmt.Errorf("raw transaction broadcast response exceeds %d bytes", queryResponseMaxBytes)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"raw transaction broadcast returned HTTP %d: %s",
			response.StatusCode,
			boundedString(responseBody, txStderrMaxBytes),
		)
	}
	if err := n.artifacts.appendJSONLine("tx/raw-broadcast-results.jsonl", map[string]any{
		"request_id":  requestID,
		"recorded_at": time.Now().UTC(),
		"step":        step,
		"response":    jsonOrString(responseBody),
	}); err != nil {
		return nil, fmt.Errorf("record raw transaction broadcast: %w", err)
	}
	checkTx, err := parseRawBroadcastResponse(responseBody)
	if err != nil {
		return nil, err
	}
	if checkTx.Code != 0 {
		return &checkTx, nil
	}
	return n.waitForCommittedTx(ctx, requestID, step, checkTx.TxHash)
}
