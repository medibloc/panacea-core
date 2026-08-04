package e2e_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeCompatibleSignedTxUnsignedPath = "upgrade/auth-bank/pre-signed-compatible-unsigned.json"
	upgradeCompatibleSignedTxPath         = "upgrade/auth-bank/pre-signed-compatible-signed.json"
	upgradeCompatibleSignedTxAmount       = "1000000"
	upgradeCompatibleSignedTxFee          = "2500000"
	upgradeCompatibleSignedTxGas          = "500000"
	upgradeCompatibleBankSendTypeURL      = "/cosmos.bank.v1beta1.MsgSend"
)

type upgradeCompatibleSignedTxFixture struct {
	Signer              ibc.Wallet              `json:"-"`
	SignerKeyName       string                  `json:"signer_key_name"`
	SignerAddress       string                  `json:"signer_address"`
	RecipientAddress    string                  `json:"recipient_address"`
	AccountNumber       uint64                  `json:"account_number"`
	Sequence            uint64                  `json:"sequence"`
	InitialBalance      string                  `json:"initial_balance"`
	InitialRecipientBal string                  `json:"initial_recipient_balance"`
	Amount              string                  `json:"amount"`
	Fee                 string                  `json:"fee"`
	GasLimit            string                  `json:"gas_limit"`
	ChainID             string                  `json:"chain_id"`
	SignMode            string                  `json:"sign_mode"`
	SignedTxPath        string                  `json:"signed_tx_path"`
	SignatureValidation string                  `json:"signature_validation"`
	PreparedAt          time.Time               `json:"prepared_at"`
	PreUpgrade          upgradeBankAccountState `json:"pre_upgrade"`
	PostUpgrade         upgradeBankAccountState `json:"post_upgrade"`
	PostRecipientBal    string                  `json:"post_recipient_balance"`
	TxHash              string                  `json:"tx_hash"`
}

type upgradeCompatibleSignedTxDecoded struct {
	SignerAddress    string
	RecipientAddress string
	Amount           string
	Fee              string
	GasLimit         string
	Sequence         uint64
	Signature        string
}

func prepareV221CompatibleSignedBankTx(
	ctx context.Context,
	network *harness.Network,
) (upgradeCompatibleSignedTxFixture, error) {
	if network == nil || network.Chain == nil || len(network.Chain.Validators) == 0 {
		return upgradeCompatibleSignedTxFixture{}, errors.New("compatible signed transaction requires an upgrade validator")
	}
	signer, err := network.BuildWallet(ctx, "upgrade-compatible-old-signer", "")
	if err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}
	recipient, err := network.BuildWallet(ctx, "upgrade-compatible-old-recipient", "")
	if err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}
	// Funding is the only pre-signing transaction involving this account. The
	// resulting sequence is captured and then held unchanged until broadcast.
	if _, err := network.BroadcastAndWaitTx(
		ctx,
		"fund-upgrade-compatible-old-signer",
		network.Chain.Validators[0],
		"faucet",
		"bank", "send", "faucet", signer.FormattedAddress(), "20000000umed",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	); err != nil {
		return upgradeCompatibleSignedTxFixture{}, fmt.Errorf("fund compatible old signer: %w", err)
	}
	initial, err := queryUpgradeBankAccount(ctx, network, "upgrade-compatible-old-signer-initial", signer.FormattedAddress())
	if err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}
	initialRecipient, err := network.QueryFullNodeBalance(ctx, recipient.FormattedAddress(), "umed")
	if err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}

	node := network.Chain.Validators[0]
	unsigned, stderr, err := node.Exec(ctx, node.TxCommand(
		signer.KeyName(),
		"bank", "send", signer.FormattedAddress(), recipient.FormattedAddress(), upgradeCompatibleSignedTxAmount+"umed",
		"--generate-only",
		"--fees", upgradeCompatibleSignedTxFee+"umed",
		"--gas", upgradeCompatibleSignedTxGas,
	), node.Chain.Config().Env)
	if err != nil {
		return upgradeCompatibleSignedTxFixture{}, fmt.Errorf("generate compatible old transaction: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	if !json.Valid(unsigned) {
		return upgradeCompatibleSignedTxFixture{}, errors.New("v2.2.1 compatible unsigned transaction is not valid JSON")
	}
	if err := node.WriteFile(ctx, unsigned, upgradeCompatibleSignedTxUnsignedPath); err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}
	if err := network.WriteArtifact(upgradeCompatibleSignedTxUnsignedPath, unsigned); err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}

	unsignedPath := path.Join(node.HomeDir(), upgradeCompatibleSignedTxUnsignedPath)
	signed, stderr, err := node.Exec(ctx, node.NodeCommand(
		"tx", "sign", unsignedPath,
		"--from", signer.KeyName(),
		"--keyring-backend", "test",
		"--chain-id", node.Chain.Config().ChainID,
		"--account-number", strconv.FormatUint(initial.AccountNumber, 10),
		"--sequence", strconv.FormatUint(initial.Sequence, 10),
		"--sign-mode", "direct",
		"--output", "json",
	), node.Chain.Config().Env)
	if err != nil {
		return upgradeCompatibleSignedTxFixture{}, fmt.Errorf("sign compatible transaction with v2.2.1: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	decoded, err := decodeUpgradeCompatibleSignedTx(signed)
	if err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}
	if decoded.SignerAddress != signer.FormattedAddress() || decoded.RecipientAddress != recipient.FormattedAddress() ||
		decoded.Amount != upgradeCompatibleSignedTxAmount || decoded.Fee != upgradeCompatibleSignedTxFee ||
		decoded.GasLimit != upgradeCompatibleSignedTxGas || decoded.Sequence != initial.Sequence {
		return upgradeCompatibleSignedTxFixture{}, fmt.Errorf("v2.2.1 signed MsgSend does not match the fixed signing contract: %+v", decoded)
	}
	if err := node.WriteFile(ctx, signed, upgradeCompatibleSignedTxPath); err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}
	if err := network.WriteArtifact(upgradeCompatibleSignedTxPath, signed); err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}

	signedPath := path.Join(node.HomeDir(), upgradeCompatibleSignedTxPath)
	validation, stderr, err := node.Exec(ctx, node.NodeCommand(
		"tx", "validate-signatures", signedPath,
		"--chain-id", node.Chain.Config().ChainID,
	), node.Chain.Config().Env)
	if err != nil {
		return upgradeCompatibleSignedTxFixture{}, fmt.Errorf("validate v2.2.1 compatible signature: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	if !strings.Contains(string(validation), "[OK]") {
		return upgradeCompatibleSignedTxFixture{}, fmt.Errorf("v2.2.1 signature validation did not report OK: %s", strings.TrimSpace(string(validation)))
	}
	if err := network.WriteArtifact("upgrade/auth-bank/pre-signed-compatible-signature-validation.txt", validation); err != nil {
		return upgradeCompatibleSignedTxFixture{}, err
	}
	fixture := upgradeCompatibleSignedTxFixture{
		Signer:              signer,
		SignerKeyName:       signer.KeyName(),
		SignerAddress:       signer.FormattedAddress(),
		RecipientAddress:    recipient.FormattedAddress(),
		AccountNumber:       initial.AccountNumber,
		Sequence:            initial.Sequence,
		InitialBalance:      initial.Balance,
		InitialRecipientBal: initialRecipient.String(),
		Amount:              upgradeCompatibleSignedTxAmount,
		Fee:                 upgradeCompatibleSignedTxFee,
		GasLimit:            upgradeCompatibleSignedTxGas,
		ChainID:             node.Chain.Config().ChainID,
		SignMode:            "SIGN_MODE_DIRECT",
		SignedTxPath:        upgradeCompatibleSignedTxPath,
		SignatureValidation: strings.TrimSpace(string(validation)),
		PreparedAt:          time.Now().UTC(),
	}
	if err := network.WriteArtifactJSON("upgrade/auth-bank/pre-signed-compatible-preparation.json", fixture); err != nil {
		return fixture, err
	}
	return fixture, nil
}

func captureV221CompatibleSignedBankTxPreUpgrade(
	ctx context.Context,
	network *harness.Network,
	fixture *upgradeCompatibleSignedTxFixture,
) error {
	state, err := queryUpgradeBankAccount(ctx, network, "upgrade-compatible-old-signer-pre-upgrade", fixture.SignerAddress)
	if err != nil {
		return err
	}
	if state.AccountNumber != fixture.AccountNumber || state.Sequence != fixture.Sequence || state.Balance != fixture.InitialBalance {
		return fmt.Errorf("compatible old signer changed before upgrade: got %+v", state)
	}
	fixture.PreUpgrade = state
	return network.WriteArtifactJSON("upgrade/auth-bank/pre-signed-compatible-pre-upgrade.json", map[string]any{
		"fixture": fixture,
		"state":   state,
	})
}

func broadcastV221CompatibleSignedBankTxAfterUpgrade(
	ctx context.Context,
	network *harness.Network,
	fixture *upgradeCompatibleSignedTxFixture,
) error {
	preserved, err := queryUpgradeBankAccount(ctx, network, "upgrade-compatible-old-signer-post-upgrade-preservation", fixture.SignerAddress)
	if err != nil {
		return err
	}
	if preserved != fixture.PreUpgrade {
		return fmt.Errorf("compatible old signer was not preserved across upgrade: before=%+v after=%+v", fixture.PreUpgrade, preserved)
	}
	if err := network.WriteArtifactJSON("upgrade/auth-bank/pre-signed-compatible-post-upgrade-preservation.json", preserved); err != nil {
		return err
	}
	result, err := network.BroadcastSignedTxFileAndWait(
		ctx,
		"upgrade-compatible-old-signed-bank-send",
		network.Chain.Validators[0],
		fixture.SignedTxPath,
	)
	if err != nil {
		return err
	}
	after, err := queryUpgradeBankAccount(ctx, network, "upgrade-compatible-old-signer-post-upgrade-mutation", fixture.SignerAddress)
	if err != nil {
		return err
	}
	recipientBalance, err := network.QueryFullNodeBalance(ctx, fixture.RecipientAddress, "umed")
	if err != nil {
		return err
	}
	wantSignerBalance, err := subtractUpgradeAmounts(fixture.InitialBalance, fixture.Amount, fixture.Fee)
	if err != nil {
		return err
	}
	wantRecipientBalance, err := addUpgradeAmounts(fixture.InitialRecipientBal, fixture.Amount)
	if err != nil {
		return err
	}
	if after.AccountNumber != fixture.AccountNumber || after.Sequence != fixture.Sequence+1 || after.Balance != wantSignerBalance {
		return fmt.Errorf("compatible old-signed MsgSend signer result mismatch: got %+v want account=%d sequence=%d balance=%s", after, fixture.AccountNumber, fixture.Sequence+1, wantSignerBalance)
	}
	if recipientBalance.String() != wantRecipientBalance {
		return fmt.Errorf("compatible old-signed MsgSend recipient balance=%s, want %s", recipientBalance, wantRecipientBalance)
	}
	fixture.PostUpgrade = after
	fixture.PostRecipientBal = recipientBalance.String()
	fixture.TxHash = result.TxHash
	if err := network.WriteArtifactJSON("upgrade/auth-bank/pre-signed-compatible-post-upgrade-mutation.json", map[string]any{
		"fixture":   fixture,
		"tx_result": result,
	}); err != nil {
		return err
	}
	return recordUpgradeHistoricalTx(ctx, network, "post-upgrade-compatible-old-signed", result.TxHash)
}

func assertV221CompatibleSignedBankTxAfterRestart(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeCompatibleSignedTxFixture,
) error {
	state, err := queryUpgradeBankAccount(ctx, network, "upgrade-compatible-old-signer-post-restart", fixture.SignerAddress)
	if err != nil {
		return err
	}
	if state != fixture.PostUpgrade {
		return fmt.Errorf("compatible old-signed signer changed after restart: got %+v want %+v", state, fixture.PostUpgrade)
	}
	recipientBalance, err := network.QueryFullNodeBalance(ctx, fixture.RecipientAddress, "umed")
	if err != nil {
		return err
	}
	if recipientBalance.String() != fixture.PostRecipientBal {
		return fmt.Errorf("compatible old-signed recipient changed after restart: got %s want %s", recipientBalance, fixture.PostRecipientBal)
	}
	if err := recordUpgradeHistoricalTx(ctx, network, "post-restart-compatible-old-signed", fixture.TxHash); err != nil {
		return err
	}
	return network.WriteArtifactJSON("upgrade/auth-bank/pre-signed-compatible-post-restart.json", map[string]any{
		"fixture":           fixture,
		"signer_state":      state,
		"recipient_balance": recipientBalance.String(),
	})
}

func decodeUpgradeCompatibleSignedTx(raw []byte) (upgradeCompatibleSignedTxDecoded, error) {
	var transaction struct {
		Body struct {
			Messages []struct {
				TypeURL string `json:"@type"`
				From    string `json:"from_address"`
				To      string `json:"to_address"`
				Amount  []struct {
					Denom  string `json:"denom"`
					Amount string `json:"amount"`
				} `json:"amount"`
			} `json:"messages"`
		} `json:"body"`
		AuthInfo struct {
			SignerInfos []struct {
				ModeInfo struct {
					Single struct {
						Mode string `json:"mode"`
					} `json:"single"`
				} `json:"mode_info"`
				Sequence string `json:"sequence"`
			} `json:"signer_infos"`
			Fee struct {
				Amount []struct {
					Denom  string `json:"denom"`
					Amount string `json:"amount"`
				} `json:"amount"`
				GasLimit string `json:"gas_limit"`
			} `json:"fee"`
		} `json:"auth_info"`
		Signatures []string `json:"signatures"`
	}
	if err := json.Unmarshal(raw, &transaction); err != nil {
		return upgradeCompatibleSignedTxDecoded{}, fmt.Errorf("decode v2.2.1 compatible signed transaction: %w", err)
	}
	if len(transaction.Body.Messages) != 1 || transaction.Body.Messages[0].TypeURL != upgradeCompatibleBankSendTypeURL {
		return upgradeCompatibleSignedTxDecoded{}, errors.New("compatible signed transaction must contain exactly one MsgSend")
	}
	message := transaction.Body.Messages[0]
	if len(message.Amount) != 1 || message.Amount[0].Denom != "umed" {
		return upgradeCompatibleSignedTxDecoded{}, errors.New("compatible signed MsgSend must contain exactly one umed amount")
	}
	if len(transaction.AuthInfo.SignerInfos) != 1 || transaction.AuthInfo.SignerInfos[0].ModeInfo.Single.Mode != "SIGN_MODE_DIRECT" {
		return upgradeCompatibleSignedTxDecoded{}, errors.New("compatible signed MsgSend must use exactly one SIGN_MODE_DIRECT signer")
	}
	sequence, err := strconv.ParseUint(transaction.AuthInfo.SignerInfos[0].Sequence, 10, 64)
	if err != nil {
		return upgradeCompatibleSignedTxDecoded{}, fmt.Errorf("decode compatible signed sequence: %w", err)
	}
	if len(transaction.AuthInfo.Fee.Amount) != 1 || transaction.AuthInfo.Fee.Amount[0].Denom != "umed" {
		return upgradeCompatibleSignedTxDecoded{}, errors.New("compatible signed MsgSend must contain exactly one umed fee")
	}
	if len(transaction.Signatures) != 1 {
		return upgradeCompatibleSignedTxDecoded{}, errors.New("compatible signed MsgSend must contain exactly one signature")
	}
	signature, err := base64.StdEncoding.DecodeString(transaction.Signatures[0])
	if err != nil || len(signature) == 0 {
		return upgradeCompatibleSignedTxDecoded{}, errors.New("compatible signed MsgSend signature must be non-empty base64")
	}
	return upgradeCompatibleSignedTxDecoded{
		SignerAddress:    message.From,
		RecipientAddress: message.To,
		Amount:           message.Amount[0].Amount,
		Fee:              transaction.AuthInfo.Fee.Amount[0].Amount,
		GasLimit:         transaction.AuthInfo.Fee.GasLimit,
		Sequence:         sequence,
		Signature:        transaction.Signatures[0],
	}, nil
}

func subtractUpgradeAmounts(minuend string, subtrahends ...string) (string, error) {
	value, ok := new(big.Int).SetString(minuend, 10)
	if !ok {
		return "", fmt.Errorf("invalid amount %q", minuend)
	}
	for _, subtrahend := range subtrahends {
		amount, parsed := new(big.Int).SetString(subtrahend, 10)
		if !parsed {
			return "", fmt.Errorf("invalid amount %q", subtrahend)
		}
		value.Sub(value, amount)
	}
	if value.Sign() < 0 {
		return "", errors.New("amount subtraction is negative")
	}
	return value.String(), nil
}

func addUpgradeAmounts(addends ...string) (string, error) {
	value := new(big.Int)
	for _, addend := range addends {
		amount, parsed := new(big.Int).SetString(addend, 10)
		if !parsed {
			return "", fmt.Errorf("invalid amount %q", addend)
		}
		value.Add(value, amount)
	}
	return value.String(), nil
}

func TestDecodeUpgradeCompatibleSignedTxRequiresDirectFixedContract(t *testing.T) {
	t.Parallel()
	raw := []byte(`{
		"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"panacea1from","to_address":"panacea1to","amount":[{"denom":"umed","amount":"1000000"}]}]},
		"auth_info":{"signer_infos":[{"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"7"}],"fee":{"amount":[{"denom":"umed","amount":"2500000"}],"gas_limit":"500000"}},
		"signatures":["AQID"]
	}`)
	decoded, err := decodeUpgradeCompatibleSignedTx(raw)
	require.NoError(t, err)
	require.Equal(t, uint64(7), decoded.Sequence)
	require.Equal(t, "1000000", decoded.Amount)
	require.Equal(t, "2500000", decoded.Fee)

	var document map[string]any
	require.NoError(t, json.Unmarshal(raw, &document))
	authInfo := document["auth_info"].(map[string]any)
	signerInfos := authInfo["signer_infos"].([]any)
	modeInfo := signerInfos[0].(map[string]any)["mode_info"].(map[string]any)
	modeInfo["single"].(map[string]any)["mode"] = "SIGN_MODE_LEGACY_AMINO_JSON"
	changed, err := json.Marshal(document)
	require.NoError(t, err)
	_, err = decodeUpgradeCompatibleSignedTx(changed)
	require.ErrorContains(t, err, "SIGN_MODE_DIRECT")
}
