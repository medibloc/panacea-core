package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	upstreamnft "cosmossdk.io/x/nft"
	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/go-bip39"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	legacyPNFTLocalClassID     = "legacy.isolation"
	legacyPNFTID               = "legacy.asset.1"
	legacyPNFTMessageTypeURL   = "/panacea.pnft.v2.MsgTransferPNFTRequest"
	legacyPNFTUnsignedFilePath = "upgrade/legacy-pnft-old-unsigned.json"
	legacyPNFTSignedFilePath   = "upgrade/legacy-pnft-old-signed.json"
	legacyPNFTEmptyQueryLimit  = 100
	upgradeRawTxSignerDomain   = "panacea upgrade e2e raw transaction signer: "
)

type upgradeWalletBuilder func(context.Context, string, string) (ibc.Wallet, error)

// buildUpgradeRawTxSigner supplies a deterministic local-test mnemonic to the
// wallet builder and verifies that the returned wallet retains it. In
// particular, Interchaintest's empty-mnemonic key-generation path intentionally
// returns a wallet with no mnemonic, which cannot be used by BroadcastRawTx.
func buildUpgradeRawTxSigner(
	ctx context.Context,
	keyName string,
	buildWallet upgradeWalletBuilder,
) (ibc.Wallet, error) {
	if strings.TrimSpace(keyName) == "" {
		return nil, errors.New("upgrade raw transaction signer key name is required")
	}
	if buildWallet == nil {
		return nil, errors.New("upgrade raw transaction wallet builder is required")
	}
	entropy := sha256.Sum256([]byte(upgradeRawTxSignerDomain + keyName))
	mnemonic, err := bip39.NewMnemonic(entropy[:])
	if err != nil {
		return nil, fmt.Errorf("build upgrade raw transaction signer mnemonic: %w", err)
	}
	wallet, err := buildWallet(ctx, keyName, mnemonic)
	if err != nil {
		return nil, fmt.Errorf("build upgrade raw transaction signer: %w", err)
	}
	if wallet == nil {
		return nil, errors.New("upgrade raw transaction wallet builder returned nil")
	}
	if strings.TrimSpace(wallet.Mnemonic()) == "" {
		return nil, errors.New("upgrade raw transaction signer did not retain mnemonic")
	}
	if wallet.Mnemonic() != mnemonic {
		return nil, errors.New("upgrade raw transaction signer retained a different mnemonic")
	}
	if len(wallet.Address()) == 0 {
		return nil, errors.New("upgrade raw transaction signer has no address")
	}
	return wallet, nil
}

func buildAndFundUpgradeRawTxSigner(
	ctx context.Context,
	network *harness.Network,
	keyName string,
) (ibc.Wallet, error) {
	if network == nil || network.Chain == nil || len(network.Chain.Validators) == 0 {
		return nil, errors.New("upgrade raw transaction signer requires a validator")
	}
	wallet, err := buildUpgradeRawTxSigner(ctx, keyName, network.BuildWallet)
	if err != nil {
		return nil, err
	}
	_, err = network.BroadcastAndWaitTx(
		ctx,
		"fund-"+keyName,
		network.Chain.Validators[0],
		interchaintest.FaucetAccountKeyName,
		"bank", "send",
		interchaintest.FaucetAccountKeyName,
		wallet.FormattedAddress(),
		sdkmath.NewInt(20_000_000).String()+"umed",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return nil, fmt.Errorf("fund upgrade raw transaction signer: %w", err)
	}
	return wallet, nil
}

type legacyPNFTFixture struct {
	LocalClassID string `json:"future_local_class_id"`
	DenomID      string `json:"denom_id"`
	PNFTID       string `json:"pnft_id"`
	Creator      string `json:"creator"`
	Owner        string `json:"owner"`
}

type legacyPNFTDenom struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Owner  string `json:"owner"`
}

type legacyPNFTRecord struct {
	DenomID   string `json:"denom_id"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Creator   string `json:"creator"`
	Owner     string `json:"owner"`
	CreatedAt string `json:"created_at"`
}

type legacyPNFTFixtureEvidence struct {
	Denom legacyPNFTDenom  `json:"denom"`
	PNFT  legacyPNFTRecord `json:"pnft"`
}

type legacySignedTxEvidence struct {
	TypeURL   string `json:"type_url"`
	DenomID   string `json:"denom_id"`
	PNFTID    string `json:"pnft_id,omitempty"`
	Signer    string `json:"signer"`
	Receiver  string `json:"receiver,omitempty"`
	Signature string `json:"-"`
}

type preparedLegacyPNFTFixture struct {
	Fixture             legacyPNFTFixture         `json:"fixture"`
	State               legacyPNFTFixtureEvidence `json:"state"`
	CreateTxHash        string                    `json:"create_tx_hash"`
	MintTxHash          string                    `json:"mint_tx_hash"`
	SignedTxPath        string                    `json:"signed_tx_path"`
	SignedTx            legacySignedTxEvidence    `json:"signed_tx"`
	SignatureValidation string                    `json:"signature_validation"`
}

type legacyPNFTIsolationEvidence struct {
	ClassID      string                          `json:"class_id"`
	NFTID        string                          `json:"nft_id"`
	CreateTxHash string                          `json:"create_tx_hash"`
	MintTxHash   string                          `json:"mint_tx_hash"`
	Class        *upstreamnft.QueryClassResponse `json:"class"`
	NFT          *upstreamnft.QueryNFTResponse   `json:"nft"`
}

type newRawLegacyPNFTRejectionEvidence struct {
	ConstructedAt      time.Time         `json:"constructed_at"`
	TypeURL            string            `json:"type_url"`
	MessageValueBase64 string            `json:"message_value_base64"`
	Signer             string            `json:"signer"`
	Receiver           string            `json:"receiver"`
	SignedBy           string            `json:"signed_by"`
	ExpectedSignMode   string            `json:"expected_sign_mode"`
	TransactionResult  *harness.TxResult `json:"transaction_result"`
}

func newLegacyPNFTFixture(creator string) (legacyPNFTFixture, error) {
	creator = strings.TrimSpace(creator)
	if creator == "" {
		return legacyPNFTFixture{}, errors.New("legacy PNFT creator is required")
	}
	return legacyPNFTFixture{
		LocalClassID: legacyPNFTLocalClassID,
		DenomID:      creator + ":" + legacyPNFTLocalClassID,
		PNFTID:       legacyPNFTID,
		Creator:      creator,
		Owner:        creator,
	}, nil
}

// prepareV221LegacyPNFTFixture creates a non-empty legacy store through the
// v2.2.1 public transaction/query boundaries. The pending transfer is signed
// and cryptographically checked by that same old binary, but intentionally
// left unbroadcast so an upgrade test can submit the exact bytes afterwards.
func prepareV221LegacyPNFTFixture(
	ctx context.Context,
	network *harness.Network,
	creator ibc.Wallet,
	transferRecipient string,
) (preparedLegacyPNFTFixture, error) {
	if network == nil || network.Chain == nil {
		return preparedLegacyPNFTFixture{}, errors.New("legacy PNFT network is required")
	}
	if creator == nil {
		return preparedLegacyPNFTFixture{}, errors.New("legacy PNFT creator wallet is required")
	}
	transferRecipient = strings.TrimSpace(transferRecipient)
	if transferRecipient == "" {
		return preparedLegacyPNFTFixture{}, errors.New("legacy PNFT transfer recipient is required")
	}
	if len(network.Chain.Validators) == 0 {
		return preparedLegacyPNFTFixture{}, errors.New("legacy PNFT network has no validator")
	}

	fixture, err := newLegacyPNFTFixture(creator.FormattedAddress())
	if err != nil {
		return preparedLegacyPNFTFixture{}, err
	}
	node := network.Chain.Validators[0]
	created, err := network.BroadcastAndWaitTx(
		ctx,
		"v221-legacy-pnft-create-denom",
		node,
		creator.KeyName(),
		"pnft", "create-denom",
		"--denom-id", fixture.DenomID,
		"--denom-symbol", "LEGACY",
		"--denom-name", "Legacy Isolation",
		"--denom-description", "v2.2.1 adversarial non-migration fixture",
		"--denom-data", `{"source":"v2.2.1","contract":"do-not-migrate"}`,
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("create v2.2.1 legacy denom: %w", err)
	}
	minted, err := network.BroadcastAndWaitTx(
		ctx,
		"v221-legacy-pnft-mint",
		node,
		creator.KeyName(),
		"pnft", "mint-pnft", fixture.DenomID, fixture.PNFTID,
		"--pnft-name", "Legacy Asset",
		"--pnft-description", "v2.2.1 adversarial non-migration fixture",
		"--pnft-data", `{"source":"v2.2.1","contract":"do-not-migrate"}`,
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("mint v2.2.1 legacy PNFT: %w", err)
	}

	denomJSON, err := network.FullNodeCLIQuery(
		ctx,
		"v221-legacy-pnft-denom",
		"pnft", "get-denom", fixture.DenomID,
	)
	if err != nil {
		return preparedLegacyPNFTFixture{}, err
	}
	pnftJSON, err := network.FullNodeCLIQuery(
		ctx,
		"v221-legacy-pnft-record",
		"pnft", "get-pnft", fixture.DenomID, fixture.PNFTID,
	)
	if err != nil {
		return preparedLegacyPNFTFixture{}, err
	}
	state, err := decodeLegacyPNFTFixtureEvidence(denomJSON, pnftJSON, fixture)
	if err != nil {
		return preparedLegacyPNFTFixture{}, err
	}

	unsignedJSON, stderr, err := node.Exec(ctx, node.TxCommand(
		creator.KeyName(),
		"pnft", "transfer-pnft", fixture.DenomID, fixture.PNFTID, transferRecipient,
		"--generate-only",
		"--gas", "500000",
	), node.Chain.Config().Env)
	if err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("generate v2.2.1 legacy PNFT transfer: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	if !json.Valid(unsignedJSON) {
		return preparedLegacyPNFTFixture{}, errors.New("v2.2.1 produced invalid unsigned legacy PNFT JSON")
	}
	if err := node.WriteFile(ctx, unsignedJSON, legacyPNFTUnsignedFilePath); err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("write unsigned legacy PNFT transaction: %w", err)
	}
	if err := network.WriteArtifact(legacyPNFTUnsignedFilePath, unsignedJSON); err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("record unsigned legacy PNFT transaction: %w", err)
	}

	unsignedContainerPath := path.Join(node.HomeDir(), legacyPNFTUnsignedFilePath)
	signedJSON, stderr, err := node.Exec(ctx, node.NodeCommand(
		"tx", "sign", unsignedContainerPath,
		"--from", creator.KeyName(),
		"--keyring-backend", "test",
		"--chain-id", node.Chain.Config().ChainID,
		"--output", "json",
	), node.Chain.Config().Env)
	if err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("sign legacy PNFT transaction with v2.2.1 binary: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	signedEvidence, err := decodeLegacySignedTxEvidence(signedJSON, fixture.DenomID, fixture.Creator)
	if err != nil {
		return preparedLegacyPNFTFixture{}, err
	}
	if signedEvidence.PNFTID != fixture.PNFTID {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("old-binary signed transaction PNFT id %q, want %q", signedEvidence.PNFTID, fixture.PNFTID)
	}
	if signedEvidence.Receiver != transferRecipient {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("old-binary signed transaction receiver %q, want %q", signedEvidence.Receiver, transferRecipient)
	}
	if err := node.WriteFile(ctx, signedJSON, legacyPNFTSignedFilePath); err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("write signed legacy PNFT transaction: %w", err)
	}
	if err := network.WriteArtifact(legacyPNFTSignedFilePath, signedJSON); err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("record signed legacy PNFT transaction: %w", err)
	}

	signedContainerPath := path.Join(node.HomeDir(), legacyPNFTSignedFilePath)
	validationOutput, stderr, err := node.Exec(ctx, node.NodeCommand(
		"tx", "validate-signatures", signedContainerPath,
		"--chain-id", node.Chain.Config().ChainID,
	), node.Chain.Config().Env)
	if err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("validate v2.2.1 legacy PNFT signature: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	if !strings.Contains(string(validationOutput), "[OK]") {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("v2.2.1 legacy PNFT signature validation did not report OK: %s", strings.TrimSpace(string(validationOutput)))
	}
	if err := network.WriteArtifact("upgrade/legacy-pnft-old-signature-validation.txt", validationOutput); err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("record legacy PNFT signature validation: %w", err)
	}

	prepared := preparedLegacyPNFTFixture{
		Fixture:             fixture,
		State:               state,
		CreateTxHash:        created.TxHash,
		MintTxHash:          minted.TxHash,
		SignedTxPath:        legacyPNFTSignedFilePath,
		SignedTx:            signedEvidence,
		SignatureValidation: strings.TrimSpace(string(validationOutput)),
	}
	if err := network.WriteArtifactJSON("upgrade/legacy-pnft-v221-fixture.json", prepared); err != nil {
		return preparedLegacyPNFTFixture{}, fmt.Errorf("record v2.2.1 legacy PNFT fixture: %w", err)
	}
	return prepared, nil
}

func decodeLegacyPNFTFixtureEvidence(
	denomJSON []byte,
	pnftJSON []byte,
	want legacyPNFTFixture,
) (legacyPNFTFixtureEvidence, error) {
	if strings.TrimSpace(want.DenomID) == "" || strings.TrimSpace(want.PNFTID) == "" {
		return legacyPNFTFixtureEvidence{}, errors.New("legacy denom and PNFT ids are required")
	}
	if strings.TrimSpace(want.Creator) == "" || strings.TrimSpace(want.Owner) == "" {
		return legacyPNFTFixtureEvidence{}, errors.New("legacy PNFT creator and owner are required")
	}

	var denomResponse struct {
		Denom legacyPNFTDenom `json:"denom"`
	}
	if err := json.Unmarshal(denomJSON, &denomResponse); err != nil {
		return legacyPNFTFixtureEvidence{}, fmt.Errorf("decode legacy denom query: %w", err)
	}
	var pnftResponse struct {
		PNFT legacyPNFTRecord `json:"pnft"`
	}
	if err := json.Unmarshal(pnftJSON, &pnftResponse); err != nil {
		return legacyPNFTFixtureEvidence{}, fmt.Errorf("decode legacy PNFT query: %w", err)
	}

	if denomResponse.Denom.ID != want.DenomID {
		return legacyPNFTFixtureEvidence{}, fmt.Errorf("legacy denom id %q, want %q", denomResponse.Denom.ID, want.DenomID)
	}
	if denomResponse.Denom.Owner != want.Owner {
		return legacyPNFTFixtureEvidence{}, fmt.Errorf("legacy denom owner %q, want %q", denomResponse.Denom.Owner, want.Owner)
	}
	if pnftResponse.PNFT.DenomID != want.DenomID {
		return legacyPNFTFixtureEvidence{}, fmt.Errorf("legacy PNFT denom id %q, want %q", pnftResponse.PNFT.DenomID, want.DenomID)
	}
	if pnftResponse.PNFT.ID != want.PNFTID {
		return legacyPNFTFixtureEvidence{}, fmt.Errorf("legacy PNFT id %q, want %q", pnftResponse.PNFT.ID, want.PNFTID)
	}
	if pnftResponse.PNFT.Creator != want.Creator {
		return legacyPNFTFixtureEvidence{}, fmt.Errorf("legacy PNFT creator %q, want %q", pnftResponse.PNFT.Creator, want.Creator)
	}
	if pnftResponse.PNFT.Owner != want.Owner {
		return legacyPNFTFixtureEvidence{}, fmt.Errorf("legacy PNFT owner %q, want %q", pnftResponse.PNFT.Owner, want.Owner)
	}

	return legacyPNFTFixtureEvidence{
		Denom: denomResponse.Denom,
		PNFT:  pnftResponse.PNFT,
	}, nil
}

func decodeLegacySignedTxEvidence(raw []byte, wantDenomID, wantSigner string) (legacySignedTxEvidence, error) {
	var transaction struct {
		Body struct {
			Messages []struct {
				TypeURL  string `json:"@type"`
				DenomID  string `json:"denom_id"`
				ID       string `json:"id"`
				Sender   string `json:"sender"`
				Receiver string `json:"receiver"`
			} `json:"messages"`
		} `json:"body"`
		AuthInfo struct {
			SignerInfos []json.RawMessage `json:"signer_infos"`
		} `json:"auth_info"`
		Signatures []string `json:"signatures"`
	}
	if err := json.Unmarshal(raw, &transaction); err != nil {
		return legacySignedTxEvidence{}, fmt.Errorf("decode old-binary signed transaction: %w", err)
	}
	if len(transaction.Body.Messages) != 1 {
		return legacySignedTxEvidence{}, fmt.Errorf("old-binary signed transaction has %d messages, want 1", len(transaction.Body.Messages))
	}
	message := transaction.Body.Messages[0]
	if message.TypeURL != legacyPNFTMessageTypeURL {
		return legacySignedTxEvidence{}, fmt.Errorf("old-binary signed transaction type URL %q, want %q", message.TypeURL, legacyPNFTMessageTypeURL)
	}
	if message.DenomID != wantDenomID {
		return legacySignedTxEvidence{}, fmt.Errorf("old-binary signed transaction denom id %q, want %q", message.DenomID, wantDenomID)
	}
	if message.Sender != wantSigner {
		return legacySignedTxEvidence{}, fmt.Errorf("old-binary signed transaction signer %q, want %q", message.Sender, wantSigner)
	}
	if len(transaction.AuthInfo.SignerInfos) != 1 {
		return legacySignedTxEvidence{}, fmt.Errorf("old-binary signed transaction has %d signer infos, want 1", len(transaction.AuthInfo.SignerInfos))
	}
	if len(transaction.Signatures) != 1 || strings.TrimSpace(transaction.Signatures[0]) == "" {
		return legacySignedTxEvidence{}, errors.New("old-binary signed transaction must contain one signature")
	}
	if signature, err := base64.StdEncoding.DecodeString(transaction.Signatures[0]); err != nil || len(signature) == 0 {
		return legacySignedTxEvidence{}, errors.New("old-binary signed transaction signature is not non-empty base64")
	}

	return legacySignedTxEvidence{
		TypeURL:   message.TypeURL,
		DenomID:   message.DenomID,
		PNFTID:    message.ID,
		Signer:    message.Sender,
		Receiver:  message.Receiver,
		Signature: transaction.Signatures[0],
	}, nil
}

func assertV221LegacyPNFTStoreEmpty(ctx context.Context, network *harness.Network) error {
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 {
		return errors.New("v2.2.1 legacy PNFT empty-store query requires a full node")
	}
	height, err := network.Chain.FullNodes[0].Height(ctx)
	if err != nil {
		return fmt.Errorf("read v2.2.1 legacy PNFT checkpoint height: %w", err)
	}
	if height <= 0 {
		return fmt.Errorf("v2.2.1 legacy PNFT checkpoint height must be positive, got %d", height)
	}
	observation, err := network.CaptureUpgradeCheckpointObservation(
		ctx,
		"v221-legacy-pnft-empty",
		network.Chain.FullNodes[0],
		height,
	)
	if err != nil {
		return err
	}
	path := legacyPNFTEmptyRESTPath()
	raw, err := network.FullNodeRESTGetAtHeight(
		ctx,
		nil,
		"v221-legacy-pnft-empty",
		path,
		height,
	)
	if err != nil {
		return err
	}
	denomCount, err := decodeCompleteLegacyPNFTDenomPage(raw)
	if err != nil {
		return fmt.Errorf("decode v2.2.1 legacy PNFT denom list: %w", err)
	}
	if denomCount != 0 {
		return fmt.Errorf("v2.2.1 legacy PNFT store has %d denoms, want empty", denomCount)
	}
	return network.WriteArtifactJSON("upgrade/legacy-pnft-normal-empty.json", map[string]any{
		"recorded_at":         observation.ObservedAt,
		"phase":               "v2.2.1-preparation",
		"query_boundary":      "rest",
		"query_height":        height,
		"observation":         observation,
		"query_path":          path,
		"pagination_limit":    legacyPNFTEmptyQueryLimit,
		"pagination_complete": true,
		"denom_count":         denomCount,
		"response":            json.RawMessage(raw),
	})
}

func legacyPNFTEmptyRESTPath() string {
	pageQuery := url.Values{
		"pagination.count_total": []string{"true"},
		"pagination.limit":       []string{strconv.Itoa(legacyPNFTEmptyQueryLimit)},
	}.Encode()
	return "/panacea/pnft/v2/denoms?" + pageQuery
}

func decodeCompleteLegacyPNFTDenomPage(raw json.RawMessage) (int, error) {
	var response struct {
		Denoms     []json.RawMessage `json:"denoms"`
		Pagination *struct {
			NextKey json.RawMessage `json:"next_key"`
			Total   json.RawMessage `json:"total"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return 0, err
	}
	if response.Pagination == nil {
		return 0, errors.New("bounded legacy PNFT denom response is missing pagination")
	}
	nextKey := strings.TrimSpace(string(response.Pagination.NextKey))
	if nextKey != "null" && nextKey != `""` {
		return 0, fmt.Errorf("bounded legacy PNFT denom page is incomplete: next_key=%s", nextKey)
	}
	totalText := strings.Trim(strings.TrimSpace(string(response.Pagination.Total)), `"`)
	total, err := strconv.ParseUint(totalText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("decode bounded legacy PNFT denom total %q: %w", totalText, err)
	}
	if total != uint64(len(response.Denoms)) {
		return 0, fmt.Errorf("bounded legacy PNFT denom page total=%d, returned=%d", total, len(response.Denoms))
	}
	if len(response.Denoms) > legacyPNFTEmptyQueryLimit {
		return 0, fmt.Errorf(
			"bounded legacy PNFT denom page returned %d entries above limit %d",
			len(response.Denoms),
			legacyPNFTEmptyQueryLimit,
		)
	}
	return len(response.Denoms), nil
}

func assertCurrentNFTStoreEmpty(
	ctx context.Context,
	network *harness.Network,
	phase string,
	canonicalClassID string,
) error {
	canonicalClassID = strings.TrimSpace(canonicalClassID)
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 {
		return errors.New("current NFT empty-store query requires a full node")
	}
	if canonicalClassID == "" {
		return errors.New("current NFT empty-store query requires a canonical class ID")
	}
	if err := validateCanonicalNFTClassIDForEmptyQuery(canonicalClassID); err != nil {
		return err
	}
	observation, err := network.CaptureUpgradeCheckpointObservation(
		ctx,
		"upgrade-"+phase+"-empty-nft",
		network.Chain.FullNodes[0],
		0,
	)
	if err != nil {
		return err
	}
	queryCtx := harness.ContextAtHeight(ctx, observation.Height)
	page := currentNFTEmptyPageRequest()
	classes, err := network.QueryNFTClassesGRPC(
		queryCtx,
		"upgrade-"+phase+"-empty-nft-classes",
		&upstreamnft.QueryClassesRequest{Pagination: page},
	)
	if err != nil {
		return err
	}
	nfts, err := network.QueryNFTsGRPC(
		queryCtx,
		"upgrade-"+phase+"-empty-nfts",
		&upstreamnft.QueryNFTsRequest{
			ClassId:    canonicalClassID,
			Pagination: currentNFTEmptyPageRequest(),
		},
	)
	if err != nil {
		return err
	}
	heightText := strconv.FormatInt(observation.Height, 10)
	panaceaRecords, err := network.FullNodeGRPCQuery(
		ctx,
		"upgrade-"+phase+"-empty-panacea-nft-records",
		"nft", "nft-records",
		"--class-id", canonicalClassID,
		"--limit", strconv.Itoa(legacyPNFTEmptyQueryLimit),
		"--height", heightText,
	)
	if err != nil {
		return err
	}
	if err := validateCompleteEmptyCurrentNFTPage("classes", len(classes.GetClasses()), classes.GetPagination()); err != nil {
		return fmt.Errorf("current NFT store is not empty during %s: %w", phase, err)
	}
	if err := validateCompleteEmptyCurrentNFTPage("legacy class NFTs", len(nfts.GetNfts()), nfts.GetPagination()); err != nil {
		return fmt.Errorf("current NFT store is not empty during %s: %w", phase, err)
	}
	panaceaRecordCount, err := decodeCompleteEmptyCurrentPanaceaNFTRecords(panaceaRecords)
	if err != nil {
		return fmt.Errorf("current Panacea NFT store is not empty during %s: %w", phase, err)
	}
	if panaceaRecordCount != 0 {
		return fmt.Errorf("current Panacea NFT store has %d live records during %s, want zero", panaceaRecordCount, phase)
	}
	if err := network.FullNodeGRPCQueryExpectedError(
		ctx,
		"upgrade-"+phase+"-panacea-class-record-absent",
		"code = NotFound",
		"nft", "class-record", canonicalClassID,
		"--height", heightText,
	); err != nil {
		return err
	}
	if err := network.FullNodeGRPCQueryExpectedError(
		ctx,
		"upgrade-"+phase+"-panacea-nft-record-absent",
		"code = NotFound",
		"nft", "nft-record", canonicalClassID, legacyPNFTID,
		"--height", heightText,
	); err != nil {
		return err
	}
	return network.WriteArtifactJSON("upgrade/nft-empty-"+phase+".json", map[string]any{
		"recorded_at":                   observation.ObservedAt,
		"phase":                         phase,
		"query_height":                  observation.Height,
		"observation":                   observation,
		"canonical_class_id":            canonicalClassID,
		"nft_id":                        legacyPNFTID,
		"standard_class_count":          0,
		"standard_nft_count":            0,
		"panacea_live_record_count":     0,
		"panacea_class_policy_absent":   true,
		"panacea_minted_counter_absent": true,
		"panacea_nft_record_absent":     true,
		"panacea_burn_tombstone_absent": true,
		"absence_contract": map[string]any{
			"class_record_query_step": "upgrade-" + phase + "-panacea-class-record-absent",
			"class_record_not_found_proves": []string{
				"standard_class",
				"panacea_class_policy",
				"panacea_minted_counter",
			},
			"nft_record_query_step": "upgrade-" + phase + "-panacea-nft-record-absent",
			"nft_record_not_found_proves": []string{
				"live_nft_record",
				"burn_tombstone",
			},
			"expected_grpc_code": "NotFound",
		},
		"classes":         classes,
		"nfts":            nfts,
		"panacea_records": panaceaRecords,
	})
}

func validateCanonicalNFTClassIDForEmptyQuery(classID string) error {
	creator, localClassID, found := strings.Cut(strings.TrimSpace(classID), ":")
	if !found || strings.TrimSpace(creator) == "" || strings.TrimSpace(localClassID) == "" {
		return fmt.Errorf(
			"current NFT empty-store class ID %q must contain a creator namespace and local class ID",
			classID,
		)
	}
	return nil
}

func decodeCompleteEmptyCurrentPanaceaNFTRecords(raw json.RawMessage) (int, error) {
	var response struct {
		NFTRecords []json.RawMessage `json:"nft_records"`
		Pagination *struct {
			NextKey json.RawMessage `json:"next_key"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return 0, err
	}
	if response.Pagination == nil {
		return 0, errors.New("bounded Panacea NFT record response is missing pagination")
	}
	nextKey := strings.TrimSpace(string(response.Pagination.NextKey))
	if nextKey != "null" && nextKey != `""` {
		return 0, fmt.Errorf("bounded Panacea NFT record page is incomplete: next_key=%s", nextKey)
	}
	if len(response.NFTRecords) > legacyPNFTEmptyQueryLimit {
		return 0, fmt.Errorf(
			"bounded Panacea NFT record page returned %d entries above limit %d",
			len(response.NFTRecords),
			legacyPNFTEmptyQueryLimit,
		)
	}
	return len(response.NFTRecords), nil
}

func currentNFTEmptyPageRequest() *querytypes.PageRequest {
	// cosmossdk.io/x/nft uses collections pagination, which deliberately rejects
	// count_total. A bounded first page plus an empty next_key proves the result
	// is complete without relying on that unsupported option.
	return &querytypes.PageRequest{Limit: legacyPNFTEmptyQueryLimit}
}

func validateCompleteEmptyCurrentNFTPage(label string, count int, pagination *querytypes.PageResponse) error {
	if count != 0 {
		return fmt.Errorf("%s returned %d entries, want zero", label, count)
	}
	if pagination == nil {
		return fmt.Errorf("%s response is missing pagination", label)
	}
	if len(pagination.NextKey) != 0 {
		return fmt.Errorf("%s response has a next page", label)
	}
	return nil
}

func assertOldBinarySignedLegacyPNFTDisabled(
	ctx context.Context,
	network *harness.Network,
	prepared preparedLegacyPNFTFixture,
) error {
	result, err := network.BroadcastSignedTxFileAndWaitDeliverFailure(
		ctx,
		"upgrade-old-signed-legacy-pnft-disabled",
		network.Chain.Validators[0],
		prepared.SignedTxPath,
		"sdk",
		18,
	)
	if err != nil {
		return err
	}
	if !strings.Contains(result.RawLog, legacyPNFTDisabledMessage) {
		return fmt.Errorf("old-binary signed legacy PNFT rejection %q does not contain %q", result.RawLog, legacyPNFTDisabledMessage)
	}
	return network.WriteArtifactJSON("upgrade/legacy-pnft-old-signed-rejection.json", map[string]any{
		"type_url":  prepared.SignedTx.TypeURL,
		"denom_id":  prepared.SignedTx.DenomID,
		"pnft_id":   prepared.SignedTx.PNFTID,
		"tx_hash":   result.TxHash,
		"codespace": result.Codespace,
		"code":      result.Code,
		"raw_log":   result.RawLog,
	})
}

// assertNewRawLegacyPNFTDisabled constructs the legacy protobuf payload only
// after the upgrade and asks the current raw-tx path to account-query, DIRECT
// sign, broadcast, and wait for it. This is intentionally independent from
// the old-binary signed fixture above: it proves newly signed legacy messages
// are disabled too, rather than merely rejecting stale historical bytes.
func assertNewRawLegacyPNFTDisabled(
	ctx context.Context,
	network *harness.Network,
	signer ibc.Wallet,
	receiver string,
) error {
	if signer == nil || strings.TrimSpace(receiver) == "" {
		return errors.New("new raw legacy PNFT signer and receiver are required")
	}
	messageValue, err := buildNewRawLegacyPNFTTransfer(
		"post-upgrade-disabled-denom",
		"post-upgrade-disabled-nft",
		signer.FormattedAddress(),
		receiver,
	)
	if err != nil {
		return err
	}
	constructedAt := time.Now().UTC()
	result, err := network.BroadcastRawTx(ctx, "upgrade-new-raw-legacy-pnft-disabled", harness.RawTxRequest{
		Signer: signer,
		Message: harness.RawProtoMessage{
			TypeURL: legacyPNFTMessageTypeURL,
			Value:   messageValue,
		},
		GasLimit:  500_000,
		FeeAmount: sdkmath.NewInt(2_500_000),
	})
	if err != nil {
		return err
	}
	if result == nil || result.HeightInt64() <= 0 || result.Codespace != "sdk" || result.Code != 18 {
		return fmt.Errorf("new raw legacy PNFT result=%+v, want committed sdk/18 rejection", result)
	}
	if !strings.Contains(result.RawLog, legacyPNFTDisabledMessage) {
		return fmt.Errorf("new raw legacy PNFT rejection %q does not contain %q", result.RawLog, legacyPNFTDisabledMessage)
	}
	evidence := newRawLegacyPNFTRejectionEvidence{
		ConstructedAt:      constructedAt,
		TypeURL:            legacyPNFTMessageTypeURL,
		MessageValueBase64: base64.StdEncoding.EncodeToString(messageValue),
		Signer:             signer.FormattedAddress(),
		Receiver:           receiver,
		SignedBy:           "current post-upgrade binary raw transaction path",
		ExpectedSignMode:   "SIGN_MODE_DIRECT",
		TransactionResult:  result,
	}
	return network.WriteArtifactJSON("upgrade/legacy-pnft-new-raw-rejection.json", evidence)
}

func buildNewRawLegacyPNFTTransfer(denomID, pnftID, sender, receiver string) ([]byte, error) {
	for label, value := range map[string]string{
		"denom_id": denomID,
		"pnft_id":  pnftID,
		"sender":   sender,
		"receiver": receiver,
	} {
		if strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("raw legacy PNFT %s is required", label)
		}
	}
	var message []byte
	message = appendRawProtoString(message, 1, denomID)
	message = appendRawProtoString(message, 2, pnftID)
	message = appendRawProtoString(message, 3, sender)
	message = appendRawProtoString(message, 4, receiver)
	return message, nil
}

func createStandardNFTAtLegacyIDs(
	ctx context.Context,
	network *harness.Network,
	creator ibc.Wallet,
	prepared preparedLegacyPNFTFixture,
) (legacyPNFTIsolationEvidence, error) {
	if creator == nil {
		return legacyPNFTIsolationEvidence{}, errors.New("standard NFT isolation creator is required")
	}
	fixture := prepared.Fixture
	if creator.FormattedAddress() != fixture.Creator {
		return legacyPNFTIsolationEvidence{}, fmt.Errorf(
			"standard NFT isolation creator %s, want legacy creator %s",
			creator.FormattedAddress(),
			fixture.Creator,
		)
	}
	node := network.Chain.Validators[0]
	created, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-create-standard-class-at-legacy-id",
		node,
		creator.KeyName(),
		"nft", "create-class",
		fixture.LocalClassID,
		"Standard Legacy-ID Isolation",
		"ISOLATED",
		"owner-transferable",
		"true",
		"10",
		"--description", "standard NFT created after non-migration of legacy PNFT state",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return legacyPNFTIsolationEvidence{}, err
	}
	minted, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-mint-standard-nft-at-legacy-id",
		node,
		creator.KeyName(),
		"nft", "mint", fixture.DenomID, fixture.PNFTID, creator.FormattedAddress(),
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return legacyPNFTIsolationEvidence{}, err
	}
	classResponse, err := network.QueryNFTClassGRPC(
		ctx,
		"upgrade-standard-class-at-legacy-id",
		fixture.DenomID,
	)
	if err != nil {
		return legacyPNFTIsolationEvidence{}, err
	}
	if classResponse.GetClass() == nil || classResponse.GetClass().GetId() != fixture.DenomID {
		return legacyPNFTIsolationEvidence{}, errors.New("standard NFT class at legacy ID was not created")
	}
	nftResponse, err := network.QueryNFTGRPC(
		ctx,
		"upgrade-standard-nft-at-legacy-id",
		fixture.DenomID,
		fixture.PNFTID,
	)
	if err != nil {
		return legacyPNFTIsolationEvidence{}, err
	}
	if nftResponse.GetNft() == nil ||
		nftResponse.GetNft().GetClassId() != fixture.DenomID ||
		nftResponse.GetNft().GetId() != fixture.PNFTID {
		return legacyPNFTIsolationEvidence{}, errors.New("standard NFT at legacy IDs was not minted")
	}
	evidence := legacyPNFTIsolationEvidence{
		ClassID:      fixture.DenomID,
		NFTID:        fixture.PNFTID,
		CreateTxHash: created.TxHash,
		MintTxHash:   minted.TxHash,
		Class:        classResponse,
		NFT:          nftResponse,
	}
	if err := network.WriteArtifactJSON("upgrade/legacy-pnft-standard-isolation-post-upgrade.json", evidence); err != nil {
		return evidence, err
	}
	return evidence, nil
}

func assertStandardNFTAtLegacyIDsPersisted(
	ctx context.Context,
	network *harness.Network,
	want legacyPNFTIsolationEvidence,
) error {
	classResponse, err := network.QueryNFTClassGRPC(
		ctx,
		"upgrade-post-restart-standard-class-at-legacy-id",
		want.ClassID,
	)
	if err != nil {
		return err
	}
	nftResponse, err := network.QueryNFTGRPC(
		ctx,
		"upgrade-post-restart-standard-nft-at-legacy-id",
		want.ClassID,
		want.NFTID,
	)
	if err != nil {
		return err
	}
	if classResponse.GetClass() == nil || classResponse.GetClass().GetId() != want.ClassID {
		return errors.New("post-restart standard class at legacy ID is missing")
	}
	if nftResponse.GetNft() == nil ||
		nftResponse.GetNft().GetClassId() != want.ClassID ||
		nftResponse.GetNft().GetId() != want.NFTID {
		return errors.New("post-restart standard NFT at legacy IDs is missing")
	}
	return network.WriteArtifactJSON("upgrade/legacy-pnft-standard-isolation-post-restart.json", map[string]any{
		"class_id": want.ClassID,
		"nft_id":   want.NFTID,
		"class":    classResponse,
		"nft":      nftResponse,
	})
}
