package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeHaltMempoolTxCount     = 2
	upgradeHaltMempoolAmount      = "1000000"
	upgradeHaltMempoolFee         = "2500000"
	upgradeHaltMempoolGas         = "500000"
	upgradeHaltMempoolFunding     = "30000000"
	upgradeHaltMempoolArtifactDir = "upgrade/halt-mempool-carrier"
)

type upgradeHaltMempoolReconcilePlan struct {
	InitialSequence  uint64 `json:"initial_sequence"`
	ObservedSequence uint64 `json:"observed_sequence"`
	TransactionCount int    `json:"transaction_count"`
	CommittedPrefix  int    `json:"committed_prefix"`
	MissingSuffix    []int  `json:"missing_suffix,omitempty"`
}

type upgradeHaltMempoolSignedTx struct {
	Index             int               `json:"index"`
	RecipientAddress  string            `json:"recipient_address"`
	InitialRecipient  string            `json:"initial_recipient_balance"`
	Sequence          uint64            `json:"sequence"`
	Amount            string            `json:"amount"`
	Fee               string            `json:"fee"`
	GasLimit          string            `json:"gas_limit"`
	UnsignedTxPath    string            `json:"unsigned_tx_path"`
	SignedTxPath      string            `json:"signed_tx_path"`
	SignedJSONSHA256  string            `json:"signed_json_sha256"`
	SignatureEvidence string            `json:"signature_evidence"`
	CheckTx           *harness.TxResult `json:"check_tx,omitempty"`
	Committed         *harness.TxResult `json:"committed,omitempty"`
}

type upgradeHaltMempoolFixture struct {
	Signer            ibc.Wallet                      `json:"-"`
	SignerKeyName     string                          `json:"signer_key_name"`
	SignerAddress     string                          `json:"signer_address"`
	AccountNumber     uint64                          `json:"account_number"`
	InitialSequence   uint64                          `json:"initial_sequence"`
	InitialBalance    string                          `json:"initial_balance"`
	ChainID           string                          `json:"chain_id"`
	PreparedAt        time.Time                       `json:"prepared_at"`
	Transactions      []upgradeHaltMempoolSignedTx    `json:"transactions"`
	PreUpgrade        upgradeBankAccountState         `json:"pre_upgrade"`
	CarrierNode       string                          `json:"carrier_node"`
	SubmittedAt       time.Time                       `json:"submitted_at"`
	Reconciliation    upgradeHaltMempoolReconcilePlan `json:"reconciliation"`
	PostUpgrade       upgradeBankAccountState         `json:"post_upgrade"`
	PostRecipientBals []string                        `json:"post_recipient_balances"`
}

// prepareV221UpgradeHaltMempoolTxs asks v2.2.1 to generate and DIRECT-sign two
// same-account MsgSend transactions with consecutive sequences. Exact signed
// bytes are copied to every old-image node before the upgrade starts.
func prepareV221UpgradeHaltMempoolTxs(
	ctx context.Context,
	network *harness.Network,
) (upgradeHaltMempoolFixture, error) {
	if network == nil || network.Chain == nil || len(network.Chain.Validators) == 0 || len(network.Chain.FullNodes) == 0 {
		return upgradeHaltMempoolFixture{}, errors.New("upgrade-halt mempool fixture requires validator and full-node boundaries")
	}
	signer, err := network.BuildWallet(ctx, "upgrade-halt-mempool-signer", "")
	if err != nil {
		return upgradeHaltMempoolFixture{}, err
	}
	if _, err := network.BroadcastAndWaitTx(
		ctx,
		"fund-upgrade-halt-mempool-signer",
		network.Chain.Validators[0],
		"faucet",
		"bank", "send", "faucet", signer.FormattedAddress(), upgradeHaltMempoolFunding+"umed",
		"--gas", upgradeHaltMempoolGas,
		"--broadcast-mode", "sync",
	); err != nil {
		return upgradeHaltMempoolFixture{}, fmt.Errorf("fund upgrade-halt mempool signer: %w", err)
	}
	initial, err := queryUpgradeBankAccount(ctx, network, "upgrade-halt-mempool-initial-account", signer.FormattedAddress())
	if err != nil {
		return upgradeHaltMempoolFixture{}, err
	}

	recipients := make([]ibc.Wallet, upgradeHaltMempoolTxCount)
	initialRecipients := make([]string, upgradeHaltMempoolTxCount)
	for index := range recipients {
		recipients[index], err = network.BuildWallet(ctx, fmt.Sprintf("upgrade-halt-mempool-recipient-%d", index), "")
		if err != nil {
			return upgradeHaltMempoolFixture{}, err
		}
		balance, balanceErr := network.QueryFullNodeBalance(ctx, recipients[index].FormattedAddress(), "umed")
		if balanceErr != nil {
			return upgradeHaltMempoolFixture{}, balanceErr
		}
		initialRecipients[index] = balance.String()
	}

	node := network.Chain.Validators[0]
	fixture := upgradeHaltMempoolFixture{
		Signer:          signer,
		SignerKeyName:   signer.KeyName(),
		SignerAddress:   signer.FormattedAddress(),
		AccountNumber:   initial.AccountNumber,
		InitialSequence: initial.Sequence,
		InitialBalance:  initial.Balance,
		ChainID:         node.Chain.Config().ChainID,
		PreparedAt:      time.Now().UTC(),
		Transactions:    make([]upgradeHaltMempoolSignedTx, upgradeHaltMempoolTxCount),
	}
	signedDocuments := make([][]byte, upgradeHaltMempoolTxCount)
	for index := 0; index < upgradeHaltMempoolTxCount; index++ {
		unsigned, stderr, generateErr := node.Exec(ctx, node.TxCommand(
			signer.KeyName(),
			"bank", "send", signer.FormattedAddress(), recipients[index].FormattedAddress(), upgradeHaltMempoolAmount+"umed",
			"--generate-only",
			"--fees", upgradeHaltMempoolFee+"umed",
			"--gas", upgradeHaltMempoolGas,
		), node.Chain.Config().Env)
		if generateErr != nil {
			return fixture, fmt.Errorf(
				"generate v2.2.1 halt-mempool transaction %d: %w: %s",
				index,
				generateErr,
				strings.TrimSpace(string(stderr)),
			)
		}
		if !json.Valid(unsigned) {
			return fixture, fmt.Errorf("v2.2.1 halt-mempool unsigned transaction %d is not JSON", index)
		}
		unsignedPath := fmt.Sprintf("%s/tx-%d-unsigned.json", upgradeHaltMempoolArtifactDir, index)
		signedPath := fmt.Sprintf("%s/tx-%d-signed.json", upgradeHaltMempoolArtifactDir, index)
		if err := node.WriteFile(ctx, unsigned, unsignedPath); err != nil {
			return fixture, err
		}
		if err := network.WriteArtifact(unsignedPath, unsigned); err != nil {
			return fixture, err
		}
		sequence := initial.Sequence + uint64(index)
		signArgs := upgradeHaltMempoolSignArguments(
			path.Join(node.HomeDir(), unsignedPath),
			signer.KeyName(),
			node.Chain.Config().ChainID,
			initial.AccountNumber,
			sequence,
		)
		signed, signStderr, signErr := node.Exec(ctx, node.NodeCommand(signArgs...), node.Chain.Config().Env)
		if signErr != nil {
			return fixture, fmt.Errorf(
				"sign v2.2.1 halt-mempool transaction %d: %w: %s",
				index,
				signErr,
				strings.TrimSpace(string(signStderr)),
			)
		}
		if writeErr := node.WriteFile(ctx, signed, signedPath); writeErr != nil {
			return fixture, writeErr
		}
		validation, validationStderr, validationErr := node.Exec(ctx, node.NodeCommand(
			"tx", "validate-signatures", path.Join(node.HomeDir(), signedPath),
			"--chain-id", node.Chain.Config().ChainID,
		), node.Chain.Config().Env)
		if validationErr != nil {
			return fixture, fmt.Errorf(
				"validate v2.2.1 halt-mempool signature %d: %w: %s",
				index,
				validationErr,
				strings.TrimSpace(string(validationStderr)),
			)
		}
		if !strings.Contains(string(validation), "[OK]") {
			return fixture, fmt.Errorf("v2.2.1 halt-mempool signature %d did not validate: %s", index, strings.TrimSpace(string(validation)))
		}
		if err := network.WriteArtifact(signedPath, signed); err != nil {
			return fixture, err
		}
		for _, destination := range network.Chain.Nodes() {
			if destination == node {
				continue
			}
			if err := destination.WriteFile(ctx, signed, signedPath); err != nil {
				return fixture, fmt.Errorf("copy halt-mempool tx %d to %s: %w", index, destination.Name(), err)
			}
		}
		digest := sha256.Sum256(signed)
		fixture.Transactions[index] = upgradeHaltMempoolSignedTx{
			Index:             index,
			RecipientAddress:  recipients[index].FormattedAddress(),
			InitialRecipient:  initialRecipients[index],
			Sequence:          sequence,
			Amount:            upgradeHaltMempoolAmount,
			Fee:               upgradeHaltMempoolFee,
			GasLimit:          upgradeHaltMempoolGas,
			UnsignedTxPath:    unsignedPath,
			SignedTxPath:      signedPath,
			SignedJSONSHA256:  hex.EncodeToString(digest[:]),
			SignatureEvidence: strings.TrimSpace(string(validation)),
		}
		signedDocuments[index] = signed
	}
	if _, err := decodeUpgradeHaltMempoolSignedBankPair(signedDocuments[0], signedDocuments[1]); err != nil {
		return fixture, err
	}
	if err := network.WriteArtifactJSON(upgradeHaltMempoolArtifactDir+"/preparation.json", fixture); err != nil {
		return fixture, err
	}
	return fixture, nil
}

func upgradeHaltMempoolSignArguments(
	unsignedPath string,
	keyName string,
	chainID string,
	accountNumber uint64,
	sequence uint64,
) []string {
	return []string{
		"tx", "sign", unsignedPath,
		"--from", keyName,
		"--keyring-backend", "test",
		"--chain-id", chainID,
		"--account-number", strconv.FormatUint(accountNumber, 10),
		"--sequence", strconv.FormatUint(sequence, 10),
		// Without offline mode the v2.2.1 CLI queries the current account and
		// silently replaces a future sequence, producing two sequence-N txs.
		"--offline",
		"--sign-mode", "direct",
		"--output", "json",
	}
}

func captureV221UpgradeHaltMempoolTxsPreUpgrade(
	ctx context.Context,
	network *harness.Network,
	fixture *upgradeHaltMempoolFixture,
) error {
	if fixture == nil {
		return errors.New("upgrade-halt mempool fixture is required")
	}
	state, err := queryUpgradeBankAccount(ctx, network, "upgrade-halt-mempool-pre-upgrade-account", fixture.SignerAddress)
	if err != nil {
		return err
	}
	want := upgradeBankAccountState{
		Address:       fixture.SignerAddress,
		AccountNumber: fixture.AccountNumber,
		Sequence:      fixture.InitialSequence,
		Balance:       fixture.InitialBalance,
	}
	if state != want {
		return fmt.Errorf("upgrade-halt mempool signer changed before submission: got=%+v want=%+v", state, want)
	}
	for index, transaction := range fixture.Transactions {
		balance, balanceErr := network.QueryFullNodeBalance(ctx, transaction.RecipientAddress, "umed")
		if balanceErr != nil {
			return balanceErr
		}
		if balance.String() != transaction.InitialRecipient {
			return fmt.Errorf(
				"upgrade-halt mempool recipient %d changed before submission: got=%s want=%s",
				index,
				balance,
				transaction.InitialRecipient,
			)
		}
	}
	fixture.PreUpgrade = state
	return network.WriteArtifactJSON(upgradeHaltMempoolArtifactDir+"/pre-upgrade.json", fixture)
}

// submitV221UpgradeHaltMempoolTxsAtHalt must be called only after the planned
// height has stopped committing blocks. carrier remains on v2.2.1 until both
// exact hashes have been observed committed by an upgraded validator.
func submitV221UpgradeHaltMempoolTxsAtHalt(
	ctx context.Context,
	network *harness.Network,
	carrier *cosmos.ChainNode,
	fixture *upgradeHaltMempoolFixture,
) error {
	if fixture == nil {
		return errors.New("upgrade-halt mempool fixture is required")
	}
	if carrier == nil {
		return errors.New("upgrade-halt mempool carrier node is required")
	}
	seenHashes := make(map[string]struct{}, len(fixture.Transactions))
	for index := range fixture.Transactions {
		transaction := &fixture.Transactions[index]
		checkTx, err := network.BroadcastSignedTxFileCheckTx(
			ctx,
			fmt.Sprintf("upgrade-halt-mempool-checktx-%d", index),
			carrier,
			transaction.SignedTxPath,
		)
		if err != nil {
			return err
		}
		if checkTx == nil || checkTx.Code != 0 || strings.TrimSpace(checkTx.Height) != "0" || strings.TrimSpace(checkTx.TxHash) == "" {
			return fmt.Errorf("upgrade-halt mempool CheckTx %d returned incomplete acceptance: %+v", index, checkTx)
		}
		normalized := strings.ToUpper(checkTx.TxHash)
		if _, duplicate := seenHashes[normalized]; duplicate {
			return fmt.Errorf("upgrade-halt mempool CheckTx %d duplicated hash %s", index, checkTx.TxHash)
		}
		seenHashes[normalized] = struct{}{}
		transaction.CheckTx = checkTx
	}
	fixture.CarrierNode = carrier.Name()
	fixture.SubmittedAt = time.Now().UTC()
	return network.WriteArtifactJSON(upgradeHaltMempoolArtifactDir+"/halt-checktx.json", fixture)
}

// assertV221UpgradeHaltMempoolTxsCommittedOnNode proves carrier continuity on
// an upgraded validator while the old full node is still running. Missing
// suffixes are evidence failures; this helper never hides eviction by
// rebuilding or rebroadcasting a transaction.
func assertV221UpgradeHaltMempoolTxsCommittedOnNode(
	ctx context.Context,
	network *harness.Network,
	queryNode *cosmos.ChainNode,
	fixture *upgradeHaltMempoolFixture,
) error {
	if fixture == nil {
		return errors.New("upgrade-halt mempool fixture is required")
	}
	if queryNode == nil {
		return errors.New("upgraded transaction query node is required")
	}
	for index := range fixture.Transactions {
		transaction := &fixture.Transactions[index]
		if transaction.CheckTx == nil {
			return fmt.Errorf("upgrade-halt mempool transaction %d has no CheckTx evidence", index)
		}
		committed, err := network.WaitForCommittedTxOnNode(
			ctx,
			fmt.Sprintf("upgrade-halt-mempool-committed-%d", index),
			queryNode,
			transaction.CheckTx.TxHash,
		)
		if err != nil {
			return fmt.Errorf("carrier did not deliver exact halt-mempool transaction %d: %w", index, err)
		}
		if committed.Code != 0 || !strings.EqualFold(committed.TxHash, transaction.CheckTx.TxHash) {
			return fmt.Errorf("upgrade-halt mempool transaction %d committed with unexpected result %+v", index, committed)
		}
		transaction.Committed = committed
	}

	state, err := queryUpgradeBankAccountOnNode(
		ctx,
		network,
		queryNode,
		"upgrade-halt-mempool-post-upgrade-account",
		fixture.SignerAddress,
	)
	if err != nil {
		return err
	}
	plan, err := planUpgradeHaltMempoolReconciliation(
		fixture.InitialSequence,
		state.Sequence,
		len(fixture.Transactions),
	)
	if err != nil {
		return err
	}
	if len(plan.MissingSuffix) != 0 {
		return fmt.Errorf(
			"upgrade-halt mempool carrier lost signed transaction suffix %v; fallback rebroadcast is forbidden",
			plan.MissingSuffix,
		)
	}
	wantBalance := fixture.InitialBalance
	for _, transaction := range fixture.Transactions {
		wantBalance, err = subtractUpgradeAmounts(wantBalance, transaction.Amount, transaction.Fee)
		if err != nil {
			return err
		}
	}
	if state.AccountNumber != fixture.AccountNumber || state.Balance != wantBalance {
		return fmt.Errorf(
			"upgrade-halt mempool signer final state mismatch: got=%+v want account=%d sequence=%d balance=%s",
			state,
			fixture.AccountNumber,
			fixture.InitialSequence+uint64(len(fixture.Transactions)),
			wantBalance,
		)
	}
	recipientBalances := make([]string, len(fixture.Transactions))
	for index, transaction := range fixture.Transactions {
		balance, balanceErr := queryUpgradeBalanceOnNode(
			ctx,
			network,
			queryNode,
			fmt.Sprintf("upgrade-halt-mempool-recipient-%d", index),
			transaction.RecipientAddress,
		)
		if balanceErr != nil {
			return balanceErr
		}
		wantRecipient, addErr := addUpgradeAmounts(transaction.InitialRecipient, transaction.Amount)
		if addErr != nil {
			return addErr
		}
		if balance != wantRecipient {
			return fmt.Errorf("upgrade-halt mempool recipient %d balance=%s, want %s", index, balance, wantRecipient)
		}
		recipientBalances[index] = balance
	}
	fixture.Reconciliation = plan
	fixture.PostUpgrade = state
	fixture.PostRecipientBals = recipientBalances
	return network.WriteArtifactJSON(upgradeHaltMempoolArtifactDir+"/carrier-commit.json", fixture)
}

func assertV221UpgradeHaltMempoolTxsAfterRestart(
	ctx context.Context,
	network *harness.Network,
	fixture upgradeHaltMempoolFixture,
) error {
	state, err := queryUpgradeBankAccount(ctx, network, "upgrade-halt-mempool-post-restart-account", fixture.SignerAddress)
	if err != nil {
		return err
	}
	if state != fixture.PostUpgrade {
		return fmt.Errorf("upgrade-halt mempool signer changed after restart: got=%+v want=%+v", state, fixture.PostUpgrade)
	}
	if len(fixture.PostRecipientBals) != len(fixture.Transactions) {
		return errors.New("upgrade-halt mempool post-upgrade recipient evidence is incomplete")
	}
	for index, transaction := range fixture.Transactions {
		balance, balanceErr := network.QueryFullNodeBalance(ctx, transaction.RecipientAddress, "umed")
		if balanceErr != nil {
			return balanceErr
		}
		if balance.String() != fixture.PostRecipientBals[index] {
			return fmt.Errorf(
				"upgrade-halt mempool recipient %d changed after restart: got=%s want=%s",
				index,
				balance,
				fixture.PostRecipientBals[index],
			)
		}
		if transaction.Committed == nil || transaction.CheckTx == nil ||
			!strings.EqualFold(transaction.Committed.TxHash, transaction.CheckTx.TxHash) {
			return fmt.Errorf("upgrade-halt mempool transaction %d lacks exact hash lineage", index)
		}
		if err := recordUpgradeHistoricalTx(
			ctx,
			network,
			fmt.Sprintf("post-restart-upgrade-halt-mempool-%d", index),
			transaction.CheckTx.TxHash,
		); err != nil {
			return err
		}
	}
	return network.WriteArtifactJSON(upgradeHaltMempoolArtifactDir+"/post-restart.json", fixture)
}

func planUpgradeHaltMempoolReconciliation(
	initialSequence uint64,
	observedSequence uint64,
	transactionCount int,
) (upgradeHaltMempoolReconcilePlan, error) {
	if transactionCount <= 0 {
		return upgradeHaltMempoolReconcilePlan{}, errors.New("upgrade-halt mempool transaction count must be positive")
	}
	maximum := initialSequence + uint64(transactionCount)
	if maximum < initialSequence || observedSequence < initialSequence || observedSequence > maximum {
		return upgradeHaltMempoolReconcilePlan{}, fmt.Errorf(
			"observed sequence %d is outside upgrade-halt mempool range [%d,%d]",
			observedSequence,
			initialSequence,
			maximum,
		)
	}
	committed := int(observedSequence - initialSequence)
	missing := make([]int, 0, transactionCount-committed)
	for index := committed; index < transactionCount; index++ {
		missing = append(missing, index)
	}
	if len(missing) == 0 {
		missing = nil
	}
	return upgradeHaltMempoolReconcilePlan{
		InitialSequence:  initialSequence,
		ObservedSequence: observedSequence,
		TransactionCount: transactionCount,
		CommittedPrefix:  committed,
		MissingSuffix:    missing,
	}, nil
}

func decodeUpgradeHaltMempoolSignedBankPair(
	first []byte,
	second []byte,
) ([]upgradeCompatibleSignedTxDecoded, error) {
	decoded := make([]upgradeCompatibleSignedTxDecoded, 2)
	var err error
	decoded[0], err = decodeUpgradeCompatibleSignedTx(first)
	if err != nil {
		return nil, fmt.Errorf("decode first upgrade-halt mempool transaction: %w", err)
	}
	decoded[1], err = decodeUpgradeCompatibleSignedTx(second)
	if err != nil {
		return nil, fmt.Errorf("decode second upgrade-halt mempool transaction: %w", err)
	}
	if decoded[0].SignerAddress != decoded[1].SignerAddress {
		return nil, errors.New("upgrade-halt mempool transactions must use the same signer")
	}
	if decoded[1].Sequence != decoded[0].Sequence+1 {
		return nil, errors.New("upgrade-halt mempool transactions must use consecutive sequences")
	}
	if decoded[0].RecipientAddress == decoded[1].RecipientAddress {
		return nil, errors.New("upgrade-halt mempool transactions must use distinct recipients")
	}
	for index, transaction := range decoded {
		if transaction.Amount != upgradeHaltMempoolAmount || transaction.Fee != upgradeHaltMempoolFee ||
			transaction.GasLimit != upgradeHaltMempoolGas {
			return nil, fmt.Errorf("upgrade-halt mempool transaction %d violates fixed amount/fee/gas contract", index)
		}
	}
	return decoded, nil
}

func queryUpgradeBankAccountOnNode(
	ctx context.Context,
	network *harness.Network,
	node *cosmos.ChainNode,
	step string,
	address string,
) (upgradeBankAccountState, error) {
	raw, err := network.NodeCLIQuery(ctx, step+"-auth", node, "auth", "account", address)
	if err != nil {
		return upgradeBankAccountState{}, err
	}
	state, err := decodeUpgradeBankAccount(raw, address)
	if err != nil {
		return upgradeBankAccountState{}, err
	}
	balance, err := queryUpgradeBalanceOnNode(ctx, network, node, step+"-balance", address)
	if err != nil {
		return upgradeBankAccountState{}, err
	}
	state.Balance = balance
	return state, nil
}

func queryUpgradeBalanceOnNode(
	ctx context.Context,
	network *harness.Network,
	node *cosmos.ChainNode,
	step string,
	address string,
) (string, error) {
	raw, err := network.NodeCLIQuery(ctx, step, node, "bank", "balance", address, "umed")
	if err != nil {
		return "", err
	}
	var response struct {
		Balance struct {
			Amount string `json:"amount"`
			Denom  string `json:"denom"`
		} `json:"balance"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", fmt.Errorf("decode node bank balance: %w", err)
	}
	if response.Balance.Denom != "umed" || strings.TrimSpace(response.Balance.Amount) == "" {
		return "", fmt.Errorf("node bank balance returned %+v, want umed", response.Balance)
	}
	return response.Balance.Amount, nil
}
