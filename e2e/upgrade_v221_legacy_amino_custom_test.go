package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	// Keep the lowest fixture (the DID update at 1,000,000 gas) at the
	// network's fixed minimum gas price of 5umed.
	upgradeV221LegacyAminoFee         = "5000000"
	upgradeV221LegacyAminoAOLGas      = "500000"
	upgradeV221LegacyAminoDIDGas      = "1000000"
	upgradeV221LegacyAminoFunding     = "30000000"
	upgradeV221LegacyAminoSignMode    = "SIGN_MODE_LEGACY_AMINO_JSON"
	upgradeV221LegacyAminoAOLTypeURL  = "/panacea.aol.v2.MsgCreateTopicRequest"
	upgradeV221LegacyAminoDIDTypeURL  = "/panacea.did.v2.MsgUpdateDIDRequest"
	upgradeV221LegacyAminoSDKCode     = uint32(4)
	upgradeV221LegacyAminoSDKSpace    = "sdk"
	upgradeV221LegacyAminoArtifactDir = "upgrade/legacy-amino-custom"
)

type upgradeV221LegacyAminoCustomTxKind string

const (
	upgradeV221LegacyAminoAOLCreateTopic upgradeV221LegacyAminoCustomTxKind = "aol-create-topic"
	upgradeV221LegacyAminoDIDUpdate      upgradeV221LegacyAminoCustomTxKind = "did-update"
)

type upgradeV221LegacyAminoCustomTxDecoded struct {
	Kind                 upgradeV221LegacyAminoCustomTxKind
	TypeURL              string
	SignerAddress        string
	StateObjectID        string
	Sequence             uint64
	Fee                  string
	GasLimit             string
	Signature            string
	TopicName            string
	TopicDescription     string
	DID                  string
	VerificationMethodID string
	DIDDocument          json.RawMessage
}

type upgradeV221LegacyAminoSemanticState struct {
	Kind             upgradeV221LegacyAminoCustomTxKind `json:"kind"`
	TopicNames       []string                           `json:"topic_names,omitempty"`
	TopicName        string                             `json:"topic_name,omitempty"`
	TopicDescription string                             `json:"topic_description,omitempty"`
	TopicRecords     uint64                             `json:"topic_records,omitempty"`
	TopicWriters     uint64                             `json:"topic_writers,omitempty"`
	DID              string                             `json:"did,omitempty"`
	DIDDocument      json.RawMessage                    `json:"did_document,omitempty"`
	DIDSequence      uint64                             `json:"did_sequence,omitempty"`
}

type upgradeV221LegacyAminoCustomTxFixture struct {
	Signer                  ibc.Wallet                          `json:"-"`
	Kind                    upgradeV221LegacyAminoCustomTxKind  `json:"kind"`
	SignerKeyName           string                              `json:"signer_key_name"`
	SignerAddress           string                              `json:"signer_address"`
	AccountNumber           uint64                              `json:"account_number"`
	Sequence                uint64                              `json:"sequence"`
	InitialBalance          string                              `json:"initial_balance"`
	Fee                     string                              `json:"fee"`
	GasLimit                string                              `json:"gas_limit"`
	ChainID                 string                              `json:"chain_id"`
	SignMode                string                              `json:"sign_mode"`
	StateObjectID           string                              `json:"state_object_id"`
	UnsignedTxPath          string                              `json:"unsigned_tx_path"`
	SignedTxPath            string                              `json:"signed_tx_path"`
	TamperedTxPath          string                              `json:"tampered_tx_path"`
	SignedJSONSHA256        string                              `json:"signed_json_sha256"`
	SignatureValidation     string                              `json:"signature_validation"`
	PreparedAt              time.Time                           `json:"prepared_at"`
	BeforeSemantic          upgradeV221LegacyAminoSemanticState `json:"before_semantic"`
	ExpectedSemantic        upgradeV221LegacyAminoSemanticState `json:"expected_semantic"`
	PreUpgrade              upgradeBankAccountState             `json:"pre_upgrade"`
	TamperedCheckTx         *harness.TxResult                   `json:"tampered_check_tx,omitempty"`
	PostUpgrade             upgradeBankAccountState             `json:"post_upgrade"`
	PostSemantic            upgradeV221LegacyAminoSemanticState `json:"post_semantic"`
	TxHash                  string                              `json:"tx_hash"`
	DIDVerificationMethodID string                              `json:"did_verification_method_id,omitempty"`
}

type upgradeV221LegacyAminoCustomTxsFixture struct {
	AOL upgradeV221LegacyAminoCustomTxFixture `json:"aol"`
	DID upgradeV221LegacyAminoCustomTxFixture `json:"did"`
}

// prepareV221LegacyAminoCustomTxs must run while every node still uses the
// v2.2.1 image. It creates two isolated accounts and asks that exact CLI to
// generate and amino-sign one AOL and one DID transaction without broadcasting
// either signed transaction.
func prepareV221LegacyAminoCustomTxs(
	ctx context.Context,
	network *harness.Network,
) (upgradeV221LegacyAminoCustomTxsFixture, error) {
	if err := validateV221LegacyAminoNetwork(network); err != nil {
		return upgradeV221LegacyAminoCustomTxsFixture{}, err
	}
	node := network.Chain.Validators[0]
	aol, err := prepareV221LegacyAminoAOLTx(ctx, network, node)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxsFixture{}, err
	}
	did, err := prepareV221LegacyAminoDIDTx(ctx, network, node)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxsFixture{}, err
	}
	fixture := upgradeV221LegacyAminoCustomTxsFixture{AOL: aol, DID: did}
	if err := network.WriteArtifactJSON(upgradeV221LegacyAminoArtifactDir+"/preparation.json", fixture); err != nil {
		return fixture, err
	}
	return fixture, nil
}

func prepareV221LegacyAminoAOLTx(
	ctx context.Context,
	network *harness.Network,
	node *cosmos.ChainNode,
) (upgradeV221LegacyAminoCustomTxFixture, error) {
	signer, initial, err := buildAndFundV221LegacyAminoSigner(ctx, network, "upgrade-legacy-amino-aol")
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	topicName := fmt.Sprintf("legacy-amino-%d", time.Now().UTC().UnixNano())
	description := "v2.2.1 legacy-amino AOL transaction broadcast after upgrade"
	unsigned, stderr, err := node.Exec(ctx, node.TxCommand(
		signer.KeyName(),
		"aol", "create-topic", topicName,
		"--description", description,
		"--generate-only",
		"--fees", upgradeV221LegacyAminoFee+"umed",
		"--gas", upgradeV221LegacyAminoAOLGas,
	), node.Chain.Config().Env)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf(
			"generate v2.2.1 amino AOL transaction: %w: %s",
			err,
			strings.TrimSpace(string(stderr)),
		)
	}
	before, err := captureV221LegacyAminoAOLSemantic(ctx, network, "prepare", signer.FormattedAddress(), topicName, false)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	expected := upgradeV221LegacyAminoSemanticState{
		Kind:             upgradeV221LegacyAminoAOLCreateTopic,
		TopicNames:       []string{topicName},
		TopicName:        topicName,
		TopicDescription: description,
	}
	return finishV221LegacyAminoCustomTx(
		ctx,
		network,
		node,
		signer,
		initial,
		upgradeV221LegacyAminoAOLCreateTopic,
		unsigned,
		upgradeV221LegacyAminoAOLGas,
		before,
		expected,
		"",
	)
}

func prepareV221LegacyAminoDIDTx(
	ctx context.Context,
	network *harness.Network,
	node *cosmos.ChainNode,
) (upgradeV221LegacyAminoCustomTxFixture, error) {
	beforeIDs, err := network.DIDVerificationMethodIDs(ctx, node)
	if err != nil {
		// A fresh standalone scenario legitimately has no DID keystore yet.
		// Preserve every other read/format failure instead of weakening it.
		if !strings.Contains(err.Error(), "has no public verification method identifiers") {
			return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf(
				"list existing DID identifiers before legacy-amino fixture: %w",
				err,
			)
		}
		beforeIDs = nil
	}
	signer, _, err := buildAndFundV221LegacyAminoSigner(ctx, network, "upgrade-legacy-amino-did")
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	if _, err := network.BroadcastDIDCreateAndWaitTx(
		ctx,
		"upgrade-legacy-amino-did-create",
		node,
		signer.KeyName(),
		"did", "create-did",
		"--gas", upgradeV221LegacyAminoDIDGas,
		"--broadcast-mode", "sync",
	); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf("create dedicated legacy-amino DID: %w", err)
	}
	afterIDs, err := network.DIDVerificationMethodIDs(ctx, node)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	verificationMethodID, err := onlyNewV221LegacyAminoDIDIdentifier(beforeIDs, afterIDs)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	did, err := didFromVerificationMethodID(verificationMethodID)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	initial, err := queryUpgradeBankAccount(ctx, network, "legacy-amino-did-initial-account", signer.FormattedAddress())
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	beforeRaw, err := network.FullNodeCLIQuery(ctx, "legacy-amino-did-before", "did", "get-did", did)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	beforeState, err := decodeUpgradeDIDQueryState(beforeRaw)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	updateDocument, _, err := makeUpgradeDIDUpdateDocument(beforeRaw)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	documentPath := upgradeV221LegacyAminoArtifactDir + "/did/update-document.json"
	if err := node.WriteFile(ctx, updateDocument, documentPath); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	if err := network.WriteArtifact(documentPath, updateDocument); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	unsigned, err := network.GenerateDIDAuthenticatedTx(
		ctx,
		"upgrade-legacy-amino-did-generate",
		node,
		signer.KeyName(),
		"did", "update-did", did, verificationMethodID, path.Join(node.HomeDir(), documentPath),
		"--fees", upgradeV221LegacyAminoFee+"umed",
		"--gas", upgradeV221LegacyAminoDIDGas,
	)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	documentRaw, err := json.Marshal(beforeState.Document)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	before := upgradeV221LegacyAminoSemanticState{
		Kind:        upgradeV221LegacyAminoDIDUpdate,
		DID:         did,
		DIDDocument: documentRaw,
		DIDSequence: beforeState.Sequence,
	}
	updatedDocument := make(map[string]any)
	if err := json.Unmarshal(updateDocument, &updatedDocument); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	updatedRaw, err := json.Marshal(updatedDocument)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	expected := upgradeV221LegacyAminoSemanticState{
		Kind:        upgradeV221LegacyAminoDIDUpdate,
		DID:         did,
		DIDDocument: updatedRaw,
		DIDSequence: beforeState.Sequence + 1,
	}
	return finishV221LegacyAminoCustomTx(
		ctx,
		network,
		node,
		signer,
		initial,
		upgradeV221LegacyAminoDIDUpdate,
		unsigned,
		upgradeV221LegacyAminoDIDGas,
		before,
		expected,
		verificationMethodID,
	)
}

func buildAndFundV221LegacyAminoSigner(
	ctx context.Context,
	network *harness.Network,
	keyName string,
) (ibc.Wallet, upgradeBankAccountState, error) {
	signer, err := network.BuildWallet(ctx, keyName, "")
	if err != nil {
		return nil, upgradeBankAccountState{}, err
	}
	if _, err := network.BroadcastAndWaitTx(
		ctx,
		"fund-"+keyName,
		network.Chain.Validators[0],
		"faucet",
		"bank", "send", "faucet", signer.FormattedAddress(), upgradeV221LegacyAminoFunding+"umed",
		"--gas", "500000",
		"--broadcast-mode", "sync",
	); err != nil {
		return nil, upgradeBankAccountState{}, fmt.Errorf("fund %s: %w", keyName, err)
	}
	state, err := queryUpgradeBankAccount(ctx, network, keyName+"-account", signer.FormattedAddress())
	if err != nil {
		return nil, upgradeBankAccountState{}, err
	}
	return signer, state, nil
}

func finishV221LegacyAminoCustomTx(
	ctx context.Context,
	network *harness.Network,
	node *cosmos.ChainNode,
	signer ibc.Wallet,
	initial upgradeBankAccountState,
	kind upgradeV221LegacyAminoCustomTxKind,
	unsigned []byte,
	gas string,
	before upgradeV221LegacyAminoSemanticState,
	expected upgradeV221LegacyAminoSemanticState,
	verificationMethodID string,
) (upgradeV221LegacyAminoCustomTxFixture, error) {
	if !json.Valid(unsigned) {
		return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf("v2.2.1 %s generate-only output is not JSON", kind)
	}
	slug := string(kind)
	unsignedPath := upgradeV221LegacyAminoArtifactDir + "/" + slug + "/unsigned.json"
	signedPath := upgradeV221LegacyAminoArtifactDir + "/" + slug + "/signed.json"
	tamperedPath := upgradeV221LegacyAminoArtifactDir + "/" + slug + "/tampered.json"
	if err := node.WriteFile(ctx, unsigned, unsignedPath); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	if err := network.WriteArtifact(unsignedPath, unsigned); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}

	signed, stderr, err := node.Exec(ctx, node.NodeCommand(
		"tx", "sign", path.Join(node.HomeDir(), unsignedPath),
		"--from", signer.KeyName(),
		"--keyring-backend", "test",
		"--chain-id", node.Chain.Config().ChainID,
		"--account-number", strconv.FormatUint(initial.AccountNumber, 10),
		"--sequence", strconv.FormatUint(initial.Sequence, 10),
		"--sign-mode", "amino-json",
		"--output", "json",
	), node.Chain.Config().Env)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf(
			"sign v2.2.1 %s transaction in legacy amino JSON mode: %w: %s",
			kind,
			err,
			strings.TrimSpace(string(stderr)),
		)
	}
	decoded, err := decodeV221LegacyAminoCustomTx(signed)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	if decoded.Kind != kind || decoded.SignerAddress != signer.FormattedAddress() ||
		decoded.Sequence != initial.Sequence || decoded.Fee != upgradeV221LegacyAminoFee ||
		decoded.GasLimit != gas {
		return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf(
			"v2.2.1 %s signed transaction violates fixed contract: %+v",
			kind,
			decoded,
		)
	}
	switch kind {
	case upgradeV221LegacyAminoAOLCreateTopic:
		if decoded.TopicName != expected.TopicName || decoded.TopicDescription != expected.TopicDescription ||
			decoded.StateObjectID != signer.FormattedAddress()+"/"+expected.TopicName {
			return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf(
				"v2.2.1 amino AOL message does not match requested semantic state: %+v",
				decoded,
			)
		}
	case upgradeV221LegacyAminoDIDUpdate:
		if decoded.DID != expected.DID || decoded.VerificationMethodID != verificationMethodID {
			return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf(
				"v2.2.1 amino DID message identity does not match requested semantic state: %+v",
				decoded,
			)
		}
		documentState := upgradeV221LegacyAminoSemanticState{
			Kind:        upgradeV221LegacyAminoDIDUpdate,
			DID:         expected.DID,
			DIDDocument: decoded.DIDDocument,
			DIDSequence: expected.DIDSequence,
		}
		if err := assertV221LegacyAminoSemanticEqual(expected, documentState); err != nil {
			return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf(
				"v2.2.1 amino DID message does not match requested document: %w",
				err,
			)
		}
	}
	if err := node.WriteFile(ctx, signed, signedPath); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	if err := network.WriteArtifact(signedPath, signed); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	tampered, err := tamperV221LegacyAminoSignedTx(signed)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	if err := node.WriteFile(ctx, tampered, tamperedPath); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	if err := network.WriteArtifact(tamperedPath, tampered); err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, err
	}
	validation, stderr, err := node.Exec(ctx, node.NodeCommand(
		"tx", "validate-signatures", path.Join(node.HomeDir(), signedPath),
		"--chain-id", node.Chain.Config().ChainID,
	), node.Chain.Config().Env)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf(
			"validate v2.2.1 %s amino signature: %w: %s",
			kind,
			err,
			strings.TrimSpace(string(stderr)),
		)
	}
	if !strings.Contains(string(validation), "[OK]") {
		return upgradeV221LegacyAminoCustomTxFixture{}, fmt.Errorf(
			"v2.2.1 %s signature validation did not report OK: %s",
			kind,
			strings.TrimSpace(string(validation)),
		)
	}
	digest := sha256.Sum256(signed)
	return upgradeV221LegacyAminoCustomTxFixture{
		Signer:                  signer,
		Kind:                    kind,
		SignerKeyName:           signer.KeyName(),
		SignerAddress:           signer.FormattedAddress(),
		AccountNumber:           initial.AccountNumber,
		Sequence:                initial.Sequence,
		InitialBalance:          initial.Balance,
		Fee:                     upgradeV221LegacyAminoFee,
		GasLimit:                gas,
		ChainID:                 node.Chain.Config().ChainID,
		SignMode:                upgradeV221LegacyAminoSignMode,
		StateObjectID:           decoded.StateObjectID,
		UnsignedTxPath:          unsignedPath,
		SignedTxPath:            signedPath,
		TamperedTxPath:          tamperedPath,
		SignedJSONSHA256:        hex.EncodeToString(digest[:]),
		SignatureValidation:     strings.TrimSpace(string(validation)),
		PreparedAt:              time.Now().UTC(),
		BeforeSemantic:          before,
		ExpectedSemantic:        expected,
		DIDVerificationMethodID: verificationMethodID,
	}, nil
}

func captureV221LegacyAminoCustomTxsPreUpgrade(
	ctx context.Context,
	network *harness.Network,
	fixture *upgradeV221LegacyAminoCustomTxsFixture,
) error {
	if fixture == nil {
		return errors.New("legacy-amino custom fixture is required")
	}
	for _, item := range []*upgradeV221LegacyAminoCustomTxFixture{&fixture.AOL, &fixture.DID} {
		state, err := queryUpgradeBankAccount(
			ctx,
			network,
			"legacy-amino-"+string(item.Kind)+"-pre-upgrade-account",
			item.SignerAddress,
		)
		if err != nil {
			return err
		}
		want := upgradeBankAccountState{
			Address:       item.SignerAddress,
			AccountNumber: item.AccountNumber,
			Sequence:      item.Sequence,
			Balance:       item.InitialBalance,
		}
		if state != want {
			return fmt.Errorf("legacy-amino %s signer changed before upgrade: got=%+v want=%+v", item.Kind, state, want)
		}
		semantic, err := captureV221LegacyAminoSemantic(ctx, network, "pre-upgrade", *item, false)
		if err != nil {
			return err
		}
		if err := assertV221LegacyAminoSemanticEqual(item.BeforeSemantic, semantic); err != nil {
			return fmt.Errorf("legacy-amino %s state changed before upgrade: %w", item.Kind, err)
		}
		item.PreUpgrade = state
	}
	return network.WriteArtifactJSON(upgradeV221LegacyAminoArtifactDir+"/pre-upgrade.json", fixture)
}

func broadcastV221LegacyAminoCustomTxsAfterUpgrade(
	ctx context.Context,
	network *harness.Network,
	fixture *upgradeV221LegacyAminoCustomTxsFixture,
) error {
	if fixture == nil {
		return errors.New("legacy-amino custom fixture is required")
	}
	if err := validateV221LegacyAminoNetwork(network); err != nil {
		return err
	}
	node := network.Chain.Validators[0]
	for _, item := range []*upgradeV221LegacyAminoCustomTxFixture{&fixture.AOL, &fixture.DID} {
		preserved, err := queryUpgradeBankAccount(
			ctx,
			network,
			"legacy-amino-"+string(item.Kind)+"-post-upgrade-preservation-account",
			item.SignerAddress,
		)
		if err != nil {
			return err
		}
		if preserved != item.PreUpgrade {
			return fmt.Errorf("legacy-amino %s signer was not preserved: before=%+v after=%+v", item.Kind, item.PreUpgrade, preserved)
		}
		semanticBefore, err := captureV221LegacyAminoSemantic(ctx, network, "post-upgrade-before", *item, false)
		if err != nil {
			return err
		}
		if err := assertV221LegacyAminoSemanticEqual(item.BeforeSemantic, semanticBefore); err != nil {
			return fmt.Errorf("legacy-amino %s semantic state was not preserved: %w", item.Kind, err)
		}

		tampered, err := network.BroadcastSignedTxFileExpectCheckTxFailure(
			ctx,
			"legacy-amino-"+string(item.Kind)+"-tampered",
			node,
			item.TamperedTxPath,
			upgradeV221LegacyAminoSDKSpace,
			upgradeV221LegacyAminoSDKCode,
		)
		if err != nil {
			return err
		}
		unchanged, err := queryUpgradeBankAccount(
			ctx,
			network,
			"legacy-amino-"+string(item.Kind)+"-after-tamper-account",
			item.SignerAddress,
		)
		if err != nil {
			return err
		}
		if unchanged != item.PreUpgrade {
			return fmt.Errorf("legacy-amino %s tampered CheckTx changed signer: before=%+v after=%+v", item.Kind, item.PreUpgrade, unchanged)
		}
		semanticUnchanged, err := captureV221LegacyAminoSemantic(ctx, network, "after-tamper", *item, false)
		if err != nil {
			return err
		}
		if err := assertV221LegacyAminoSemanticEqual(item.BeforeSemantic, semanticUnchanged); err != nil {
			return fmt.Errorf("legacy-amino %s tampered CheckTx changed semantic state: %w", item.Kind, err)
		}

		result, err := network.BroadcastSignedTxFileAndWait(
			ctx,
			"legacy-amino-"+string(item.Kind)+"-valid",
			node,
			item.SignedTxPath,
		)
		if err != nil {
			return err
		}
		after, err := queryUpgradeBankAccount(
			ctx,
			network,
			"legacy-amino-"+string(item.Kind)+"-post-upgrade-account",
			item.SignerAddress,
		)
		if err != nil {
			return err
		}
		wantBalance, err := subtractUpgradeAmounts(item.InitialBalance, item.Fee)
		if err != nil {
			return err
		}
		if after.AccountNumber != item.AccountNumber || after.Sequence != item.Sequence+1 || after.Balance != wantBalance {
			return fmt.Errorf(
				"legacy-amino %s signer result mismatch: got=%+v want account=%d sequence=%d balance=%s",
				item.Kind,
				after,
				item.AccountNumber,
				item.Sequence+1,
				wantBalance,
			)
		}
		semanticAfter, err := captureV221LegacyAminoSemantic(ctx, network, "post-upgrade-after", *item, true)
		if err != nil {
			return err
		}
		if err := assertV221LegacyAminoSemanticEqual(item.ExpectedSemantic, semanticAfter); err != nil {
			return fmt.Errorf("legacy-amino %s semantic result mismatch: %w", item.Kind, err)
		}
		item.TamperedCheckTx = tampered
		item.PostUpgrade = after
		item.PostSemantic = semanticAfter
		item.TxHash = result.TxHash
		if err := recordUpgradeHistoricalTx(ctx, network, "legacy-amino-"+string(item.Kind), result.TxHash); err != nil {
			return err
		}
	}
	return network.WriteArtifactJSON(upgradeV221LegacyAminoArtifactDir+"/post-upgrade.json", fixture)
}

func assertV221LegacyAminoCustomTxsAfterRestart(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeV221LegacyAminoCustomTxsFixture,
) error {
	for _, item := range []upgradeV221LegacyAminoCustomTxFixture{fixture.AOL, fixture.DID} {
		state, err := queryUpgradeBankAccount(
			ctx,
			network,
			"legacy-amino-"+string(item.Kind)+"-post-restart-account",
			item.SignerAddress,
		)
		if err != nil {
			return err
		}
		if state != item.PostUpgrade {
			return fmt.Errorf("legacy-amino %s signer changed after restart: got=%+v want=%+v", item.Kind, state, item.PostUpgrade)
		}
		semantic, err := captureV221LegacyAminoSemantic(ctx, network, "post-restart", item, true)
		if err != nil {
			return err
		}
		if err := assertV221LegacyAminoSemanticEqual(item.PostSemantic, semantic); err != nil {
			return fmt.Errorf("legacy-amino %s state changed after restart: %w", item.Kind, err)
		}
		if err := recordUpgradeHistoricalTx(ctx, network, "post-restart-legacy-amino-"+string(item.Kind), item.TxHash); err != nil {
			return err
		}
	}
	return network.WriteArtifactJSON(upgradeV221LegacyAminoArtifactDir+"/post-restart.json", fixture)
}

func validateV221LegacyAminoNetwork(network *harness.Network) error {
	if network == nil || network.Chain == nil || len(network.Chain.Validators) == 0 || len(network.Chain.FullNodes) == 0 {
		return errors.New("legacy-amino custom transactions require validator and full-node boundaries")
	}
	return nil
}

func onlyNewV221LegacyAminoDIDIdentifier(before, after []string) (string, error) {
	known := make(map[string]struct{}, len(before))
	for _, identifier := range before {
		known[identifier] = struct{}{}
	}
	var added []string
	for _, identifier := range after {
		if _, exists := known[identifier]; !exists {
			added = append(added, identifier)
		}
	}
	if len(added) != 1 {
		return "", fmt.Errorf("legacy-amino DID creation added %d public identifiers, want 1", len(added))
	}
	return added[0], nil
}

func captureV221LegacyAminoSemantic(
	ctx context.Context,
	network *harness.Network,
	phase string,
	fixture upgradeV221LegacyAminoCustomTxFixture,
	objectExists bool,
) (upgradeV221LegacyAminoSemanticState, error) {
	switch fixture.Kind {
	case upgradeV221LegacyAminoAOLCreateTopic:
		return captureV221LegacyAminoAOLSemantic(
			ctx,
			network,
			phase,
			fixture.SignerAddress,
			fixture.ExpectedSemantic.TopicName,
			objectExists,
		)
	case upgradeV221LegacyAminoDIDUpdate:
		raw, err := network.FullNodeCLIQuery(
			ctx,
			"legacy-amino-did-"+phase,
			"did", "get-did", fixture.ExpectedSemantic.DID,
		)
		if err != nil {
			return upgradeV221LegacyAminoSemanticState{}, err
		}
		state, err := decodeUpgradeDIDQueryState(raw)
		if err != nil {
			return upgradeV221LegacyAminoSemanticState{}, err
		}
		document, err := json.Marshal(state.Document)
		if err != nil {
			return upgradeV221LegacyAminoSemanticState{}, err
		}
		return upgradeV221LegacyAminoSemanticState{
			Kind:        upgradeV221LegacyAminoDIDUpdate,
			DID:         fixture.ExpectedSemantic.DID,
			DIDDocument: document,
			DIDSequence: state.Sequence,
		}, nil
	default:
		return upgradeV221LegacyAminoSemanticState{}, fmt.Errorf("unsupported legacy-amino semantic kind %q", fixture.Kind)
	}
}

func captureV221LegacyAminoAOLSemantic(
	ctx context.Context,
	network *harness.Network,
	phase string,
	ownerAddress string,
	topicName string,
	objectExists bool,
) (upgradeV221LegacyAminoSemanticState, error) {
	raw, err := network.FullNodeCLIQuery(
		ctx,
		networkSafeStep("legacy-amino-aol-"+phase+"-topics"),
		"aol", "list-topic", ownerAddress,
	)
	if err != nil {
		return upgradeV221LegacyAminoSemanticState{}, err
	}
	var response struct {
		TopicNames []string `json:"topic_names"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return upgradeV221LegacyAminoSemanticState{}, fmt.Errorf("decode legacy-amino AOL topic list: %w", err)
	}
	sort.Strings(response.TopicNames)
	state := upgradeV221LegacyAminoSemanticState{
		Kind:       upgradeV221LegacyAminoAOLCreateTopic,
		TopicNames: response.TopicNames,
		TopicName:  topicName,
	}
	if !objectExists {
		if len(response.TopicNames) != 0 {
			return state, fmt.Errorf("dedicated legacy-amino AOL owner unexpectedly has topics %v", response.TopicNames)
		}
		return state, nil
	}
	if len(response.TopicNames) != 1 || response.TopicNames[0] != topicName {
		return state, fmt.Errorf("legacy-amino AOL owner topics=%v, want only %q", response.TopicNames, topicName)
	}
	topicRaw, err := network.FullNodeCLIQuery(
		ctx,
		networkSafeStep("legacy-amino-aol-"+phase+"-topic"),
		"aol", "get-topic", ownerAddress, topicName,
	)
	if err != nil {
		return upgradeV221LegacyAminoSemanticState{}, err
	}
	var topicResponse struct {
		Topic struct {
			Description  string          `json:"description"`
			TotalRecords json.RawMessage `json:"total_records"`
			TotalWriters json.RawMessage `json:"total_writers"`
		} `json:"topic"`
	}
	if err := json.Unmarshal(topicRaw, &topicResponse); err != nil {
		return upgradeV221LegacyAminoSemanticState{}, fmt.Errorf("decode legacy-amino AOL topic: %w", err)
	}
	records, err := decodeV221LegacyAminoUint(topicResponse.Topic.TotalRecords)
	if err != nil {
		return upgradeV221LegacyAminoSemanticState{}, fmt.Errorf("decode legacy-amino AOL total records: %w", err)
	}
	writers, err := decodeV221LegacyAminoUint(topicResponse.Topic.TotalWriters)
	if err != nil {
		return upgradeV221LegacyAminoSemanticState{}, fmt.Errorf("decode legacy-amino AOL total writers: %w", err)
	}
	state.TopicDescription = topicResponse.Topic.Description
	state.TopicRecords = records
	state.TopicWriters = writers
	return state, nil
}

func networkSafeStep(value string) string {
	return strings.ReplaceAll(value, "_", "-")
}

func decodeV221LegacyAminoUint(raw json.RawMessage) (uint64, error) {
	text := strings.Trim(strings.TrimSpace(string(raw)), "\"")
	if text == "" || text == "null" {
		return 0, nil
	}
	return strconv.ParseUint(text, 10, 64)
}

func assertV221LegacyAminoSemanticEqual(
	want upgradeV221LegacyAminoSemanticState,
	got upgradeV221LegacyAminoSemanticState,
) error {
	if want.Kind != got.Kind || !reflect.DeepEqual(want.TopicNames, got.TopicNames) ||
		want.TopicName != got.TopicName || want.TopicDescription != got.TopicDescription ||
		want.TopicRecords != got.TopicRecords || want.TopicWriters != got.TopicWriters ||
		want.DID != got.DID || want.DIDSequence != got.DIDSequence {
		return fmt.Errorf("semantic state differs: want=%+v got=%+v", want, got)
	}
	if len(want.DIDDocument) == 0 && len(got.DIDDocument) == 0 {
		return nil
	}
	var wantDocument any
	var gotDocument any
	if err := json.Unmarshal(want.DIDDocument, &wantDocument); err != nil {
		return fmt.Errorf("decode expected DID document: %w", err)
	}
	if err := json.Unmarshal(got.DIDDocument, &gotDocument); err != nil {
		return fmt.Errorf("decode actual DID document: %w", err)
	}
	if !reflect.DeepEqual(wantDocument, gotDocument) {
		return fmt.Errorf("DID document differs: want=%s got=%s", want.DIDDocument, got.DIDDocument)
	}
	return nil
}

func decodeV221LegacyAminoCustomTx(raw []byte) (upgradeV221LegacyAminoCustomTxDecoded, error) {
	var transaction struct {
		Body struct {
			Messages []json.RawMessage `json:"messages"`
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
		return upgradeV221LegacyAminoCustomTxDecoded{}, fmt.Errorf("decode v2.2.1 legacy-amino custom transaction: %w", err)
	}
	if len(transaction.Body.Messages) != 1 {
		return upgradeV221LegacyAminoCustomTxDecoded{}, errors.New("legacy-amino custom transaction must contain exactly one message")
	}
	if len(transaction.AuthInfo.SignerInfos) != 1 ||
		transaction.AuthInfo.SignerInfos[0].ModeInfo.Single.Mode != upgradeV221LegacyAminoSignMode {
		return upgradeV221LegacyAminoCustomTxDecoded{}, errors.New("legacy-amino custom transaction must use exactly one SIGN_MODE_LEGACY_AMINO_JSON signer")
	}
	sequence, err := strconv.ParseUint(transaction.AuthInfo.SignerInfos[0].Sequence, 10, 64)
	if err != nil {
		return upgradeV221LegacyAminoCustomTxDecoded{}, fmt.Errorf("decode legacy-amino custom sequence: %w", err)
	}
	if len(transaction.AuthInfo.Fee.Amount) != 1 || transaction.AuthInfo.Fee.Amount[0].Denom != "umed" ||
		strings.TrimSpace(transaction.AuthInfo.Fee.Amount[0].Amount) == "" {
		return upgradeV221LegacyAminoCustomTxDecoded{}, errors.New("legacy-amino custom transaction must contain exactly one umed fee")
	}
	if strings.TrimSpace(transaction.AuthInfo.Fee.GasLimit) == "" {
		return upgradeV221LegacyAminoCustomTxDecoded{}, errors.New("legacy-amino custom transaction gas limit is required")
	}
	if len(transaction.Signatures) != 1 {
		return upgradeV221LegacyAminoCustomTxDecoded{}, errors.New("legacy-amino custom transaction must contain exactly one signature")
	}
	outerSignature, err := base64.StdEncoding.DecodeString(transaction.Signatures[0])
	if err != nil || len(outerSignature) == 0 {
		return upgradeV221LegacyAminoCustomTxDecoded{}, errors.New("legacy-amino custom transaction signature must be non-empty base64")
	}
	var envelope struct {
		TypeURL string `json:"@type"`
	}
	if err := json.Unmarshal(transaction.Body.Messages[0], &envelope); err != nil {
		return upgradeV221LegacyAminoCustomTxDecoded{}, err
	}
	decoded := upgradeV221LegacyAminoCustomTxDecoded{
		TypeURL:   envelope.TypeURL,
		Sequence:  sequence,
		Fee:       transaction.AuthInfo.Fee.Amount[0].Amount,
		GasLimit:  transaction.AuthInfo.Fee.GasLimit,
		Signature: transaction.Signatures[0],
	}
	switch envelope.TypeURL {
	case upgradeV221LegacyAminoAOLTypeURL:
		var message struct {
			TopicName    string `json:"topic_name"`
			Description  string `json:"description"`
			OwnerAddress string `json:"owner_address"`
		}
		if err := json.Unmarshal(transaction.Body.Messages[0], &message); err != nil {
			return upgradeV221LegacyAminoCustomTxDecoded{}, err
		}
		if strings.TrimSpace(message.TopicName) == "" || strings.TrimSpace(message.OwnerAddress) == "" {
			return upgradeV221LegacyAminoCustomTxDecoded{}, errors.New("legacy-amino AOL create-topic payload is incomplete")
		}
		decoded.Kind = upgradeV221LegacyAminoAOLCreateTopic
		decoded.SignerAddress = message.OwnerAddress
		decoded.StateObjectID = message.OwnerAddress + "/" + message.TopicName
		decoded.TopicName = message.TopicName
		decoded.TopicDescription = message.Description
	case upgradeV221LegacyAminoDIDTypeURL:
		var message struct {
			DID                  string         `json:"did"`
			Document             map[string]any `json:"document"`
			VerificationMethodID string         `json:"verification_method_id"`
			Signature            string         `json:"signature"`
			FromAddress          string         `json:"from_address"`
		}
		if err := json.Unmarshal(transaction.Body.Messages[0], &message); err != nil {
			return upgradeV221LegacyAminoCustomTxDecoded{}, err
		}
		innerSignature, signatureErr := base64.StdEncoding.DecodeString(message.Signature)
		if strings.TrimSpace(message.DID) == "" || strings.TrimSpace(message.FromAddress) == "" ||
			strings.TrimSpace(message.VerificationMethodID) == "" || len(message.Document) == 0 ||
			signatureErr != nil || len(innerSignature) == 0 {
			return upgradeV221LegacyAminoCustomTxDecoded{}, errors.New("legacy-amino DID update payload is incomplete")
		}
		documentID, _ := message.Document["id"].(string)
		if documentID != message.DID {
			return upgradeV221LegacyAminoCustomTxDecoded{}, errors.New("legacy-amino DID update document ID does not match DID")
		}
		document, err := json.Marshal(message.Document)
		if err != nil {
			return upgradeV221LegacyAminoCustomTxDecoded{}, err
		}
		decoded.Kind = upgradeV221LegacyAminoDIDUpdate
		decoded.SignerAddress = message.FromAddress
		decoded.StateObjectID = message.DID
		decoded.DID = message.DID
		decoded.VerificationMethodID = message.VerificationMethodID
		decoded.DIDDocument = document
	default:
		return upgradeV221LegacyAminoCustomTxDecoded{}, fmt.Errorf("unsupported legacy-amino custom message type %q", envelope.TypeURL)
	}
	return decoded, nil
}

func tamperV221LegacyAminoSignedTx(raw []byte) ([]byte, error) {
	var transaction map[string]any
	if err := json.Unmarshal(raw, &transaction); err != nil {
		return nil, fmt.Errorf("decode legacy-amino transaction for tampering: %w", err)
	}
	body, ok := transaction["body"].(map[string]any)
	if !ok {
		return nil, errors.New("legacy-amino transaction has no body")
	}
	messages, ok := body["messages"].([]any)
	if !ok || len(messages) != 1 {
		return nil, errors.New("legacy-amino transaction must have one message for tampering")
	}
	message, ok := messages[0].(map[string]any)
	if !ok {
		return nil, errors.New("legacy-amino transaction message is not an object")
	}
	typeURL, _ := message["@type"].(string)
	switch typeURL {
	case upgradeV221LegacyAminoAOLTypeURL:
		topicName, _ := message["topic_name"].(string)
		if strings.TrimSpace(topicName) == "" {
			return nil, errors.New("legacy-amino AOL message has no topic name")
		}
		message["topic_name"] = topicName + "-tampered"
	case upgradeV221LegacyAminoDIDTypeURL:
		document, ok := message["document"].(map[string]any)
		if !ok {
			return nil, errors.New("legacy-amino DID message has no document")
		}
		services, ok := document["services"].([]any)
		if !ok || len(services) == 0 {
			return nil, errors.New("legacy-amino DID document has no service to tamper")
		}
		service, ok := services[0].(map[string]any)
		if !ok {
			return nil, errors.New("legacy-amino DID service is not an object")
		}
		endpoint, _ := service["service_endpoint"].(string)
		if strings.TrimSpace(endpoint) == "" {
			return nil, errors.New("legacy-amino DID service endpoint is empty")
		}
		service["service_endpoint"] = strings.TrimRight(endpoint, "/") + "/tampered"
	default:
		return nil, fmt.Errorf("unsupported legacy-amino message type %q for tampering", typeURL)
	}
	return json.MarshalIndent(transaction, "", "  ")
}
