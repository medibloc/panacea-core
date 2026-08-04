package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"google.golang.org/grpc/metadata"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeBurnAddress                     = "panacea100000000000000000000000000000000nqmafp"
	upgradeSystemBurnAmount                = "2000000"
	upgradeSystemDenom                     = "umed"
	upgradeSystemPreUpgradeCheckpointPhase = "pre-upgrade-checkpoint"
	upgradeSystemLocalhostClientID         = "09-localhost"
	upgradeSystemLocalhostConnectionID     = "connection-localhost"
	upgradeSystemV221IBCGoVersion          = "github.com/cosmos/ibc-go/v7@v7.3.2"
	upgradeSystemLocalhostProjectionReason = "ibc-go v7.3.2 exports its runtime localhost connection but validates only connection-{N}; InitGenesis deterministically recreates this sentinel"
	upgradeSystemMintRESTBoundary          = "rest"
)

var upgradeSystemModuleNames = []string{
	"bank",
	"burn",
	"capability",
	"consensus",
	"crisis",
	"mint",
}

type upgradeSystemModuleExport struct {
	Height        int64                      `json:"height"`
	ModuleNames   []string                   `json:"module_names"`
	Modules       map[string]json.RawMessage `json:"modules"`
	ModuleDigests map[string]string          `json:"module_digests"`
}

type upgradeSystemPreparation struct {
	BurnerAddress string     `json:"burner_address"`
	BurnerKeyName string     `json:"burner_key_name"`
	BurnAddress   string     `json:"burn_address"`
	Amount        string     `json:"amount"`
	BurnTxHash    string     `json:"burn_tx_hash"`
	Wallet        ibc.Wallet `json:"-"`
}

type upgradeSystemModuleCheckpoint struct {
	Phase                 string                                 `json:"phase"`
	RecordedAt            time.Time                              `json:"recorded_at"`
	Height                int64                                  `json:"height"`
	Observation           harness.UpgradeCheckpointObservation   `json:"observation"`
	MintParams            minttypes.Params                       `json:"mint_params"`
	Inflation             string                                 `json:"inflation"`
	AnnualProvisions      string                                 `json:"annual_provisions"`
	ConsensusParams       json.RawMessage                        `json:"consensus_params"`
	LegacyParams          map[string]map[string]string           `json:"legacy_params"`
	Supply                string                                 `json:"supply"`
	BurnAddressBalance    string                                 `json:"burn_address_balance"`
	Export                upgradeSystemModuleExport              `json:"export"`
	ExportDigest          string                                 `json:"export_digest"`
	GenesisValidation     upgradeSystemGenesisValidationEvidence `json:"genesis_validation"`
	ExportValidated       bool                                   `json:"export_validated"`
	HistoricalPublicQuery bool                                   `json:"historical_public_query"`
	MintQueryBoundary     string                                 `json:"mint_query_boundary"`
	MintQueryHeight       int64                                  `json:"mint_query_height"`
	MintQueryPaths        []string                               `json:"mint_query_paths"`
}

type upgradeSystemMintRESTState struct {
	Params           minttypes.Params
	Inflation        string
	AnnualProvisions string
}

type upgradeSystemConsensusState struct {
	BlockMaxBytes              int64    `json:"block_max_bytes"`
	BlockMaxGas                int64    `json:"block_max_gas"`
	EvidenceMaxAgeNumBlocks    int64    `json:"evidence_max_age_num_blocks"`
	EvidenceMaxAgeDuration     int64    `json:"evidence_max_age_duration"`
	EvidenceMaxBytes           int64    `json:"evidence_max_bytes"`
	ValidatorPubKeyTypes       []string `json:"validator_pub_key_types"`
	VersionApp                 uint64   `json:"version_app"`
	VoteExtensionsEnableHeight int64    `json:"vote_extensions_enable_height"`
}

type upgradeSystemGenesisValidationEvidence struct {
	Projected             bool                              `json:"projected"`
	SourceDigest          string                            `json:"source_digest"`
	ValidationInputDigest string                            `json:"validation_input_digest"`
	RemovedConnectionID   string                            `json:"removed_connection_id,omitempty"`
	RemovedConnection     json.RawMessage                   `json:"removed_connection,omitempty"`
	Reason                string                            `json:"reason,omitempty"`
	UpstreamVersion       string                            `json:"upstream_version,omitempty"`
	PublicValidation      harness.GenesisValidationEvidence `json:"public_validation"`
}

type upgradeSystemMutation struct {
	BurnTxHash      string                        `json:"burn_tx_hash"`
	BurnAmount      string                        `json:"burn_amount"`
	StartHeight     int64                         `json:"start_height"`
	EndHeight       int64                         `json:"end_height"`
	BlocksCommitted int64                         `json:"blocks_committed"`
	Checkpoint      upgradeSystemModuleCheckpoint `json:"checkpoint"`
}

func prepareUpgradeSystemModules(
	ctx context.Context,
	network *harness.Network,
) (upgradeSystemPreparation, error) {
	if network == nil || network.Chain == nil || len(network.Chain.Validators) == 0 {
		return upgradeSystemPreparation{}, errors.New("system-module preparation requires a validator")
	}
	wallet, err := network.BuildWallet(ctx, "upgrade-system-burner", "")
	if err != nil {
		return upgradeSystemPreparation{}, fmt.Errorf("build system-module burner: %w", err)
	}
	if _, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-system-fund-burner",
		network.Chain.Validators[0],
		interchaintest.FaucetAccountKeyName,
		"bank", "send",
		interchaintest.FaucetAccountKeyName,
		wallet.FormattedAddress(),
		"10000000"+upgradeSystemDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	); err != nil {
		return upgradeSystemPreparation{}, fmt.Errorf("fund system-module burner: %w", err)
	}
	burnTx, err := broadcastUpgradeSystemBurn(ctx, network, wallet, "pre-upgrade", upgradeSystemBurnAmount)
	if err != nil {
		return upgradeSystemPreparation{}, err
	}
	preparation := upgradeSystemPreparation{
		BurnerAddress: wallet.FormattedAddress(),
		BurnerKeyName: wallet.KeyName(),
		BurnAddress:   upgradeBurnAddress,
		Amount:        upgradeSystemBurnAmount,
		BurnTxHash:    burnTx,
		Wallet:        wallet,
	}
	if err := network.WriteArtifactJSON("upgrade/system-modules/preparation.json", preparation); err != nil {
		return preparation, fmt.Errorf("record system-module preparation: %w", err)
	}
	return preparation, nil
}

func broadcastUpgradeSystemBurn(
	ctx context.Context,
	network *harness.Network,
	wallet ibc.Wallet,
	phase string,
	amount string,
) (string, error) {
	if value, ok := sdkmath.NewIntFromString(amount); !ok || !value.IsPositive() {
		return "", fmt.Errorf("invalid burn amount %q", amount)
	}
	tx, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-system-burn-"+phase,
		network.Chain.Validators[0],
		wallet.KeyName(),
		"bank", "send", wallet.KeyName(), upgradeBurnAddress, amount+upgradeSystemDenom,
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	if err != nil {
		return "", fmt.Errorf("%s system burn transaction: %w", phase, err)
	}
	balance, err := network.QueryFullNodeBalance(ctx, upgradeBurnAddress, upgradeSystemDenom)
	if err != nil {
		return "", fmt.Errorf("%s burn-address balance: %w", phase, err)
	}
	if !balance.IsZero() {
		return "", fmt.Errorf(
			"%s burn-address balance = %s%s after committed send, want zero after burn EndBlock",
			phase,
			balance,
			upgradeSystemDenom,
		)
	}
	return tx.TxHash, nil
}

func upgradeSystemMintRESTQueryPaths() []string {
	return []string{
		"/cosmos/mint/v1beta1/params",
		"/cosmos/mint/v1beta1/inflation",
		"/cosmos/mint/v1beta1/annual_provisions",
	}
}

func decodeUpgradeSystemMintRESTResponses(
	paramsResponse []byte,
	inflationResponse []byte,
	annualProvisionsResponse []byte,
) (upgradeSystemMintRESTState, error) {
	paramsEnvelope, err := decodeUpgradeSystemJSONObject(paramsResponse, "mint params response")
	if err != nil {
		return upgradeSystemMintRESTState{}, err
	}
	paramsRaw, err := requiredUpgradeSystemJSONField(paramsEnvelope, "mint params response", "params")
	if err != nil {
		return upgradeSystemMintRESTState{}, err
	}
	params, err := decodeUpgradeSystemMintParams(paramsRaw)
	if err != nil {
		return upgradeSystemMintRESTState{}, err
	}

	inflation, err := decodeUpgradeSystemMintRESTDecimal(inflationResponse, "inflation", "inflation")
	if err != nil {
		return upgradeSystemMintRESTState{}, err
	}
	annualProvisions, err := decodeUpgradeSystemMintRESTDecimal(
		annualProvisionsResponse,
		"annual provisions",
		"annual_provisions",
		"annualProvisions",
	)
	if err != nil {
		return upgradeSystemMintRESTState{}, err
	}
	return upgradeSystemMintRESTState{
		Params:           params,
		Inflation:        inflation,
		AnnualProvisions: annualProvisions,
	}, nil
}

func decodeUpgradeSystemMintParams(paramsRaw []byte) (minttypes.Params, error) {
	paramsObject, err := decodeUpgradeSystemJSONObject(paramsRaw, "mint params")
	if err != nil {
		return minttypes.Params{}, err
	}
	stringField := func(label string, names ...string) (string, error) {
		raw, fieldErr := requiredUpgradeSystemJSONField(paramsObject, "mint params", names...)
		if fieldErr != nil {
			return "", fieldErr
		}
		var value string
		if unmarshalErr := json.Unmarshal(raw, &value); unmarshalErr != nil || strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("mint params %s must be a non-empty string", label)
		}
		return strings.TrimSpace(value), nil
	}
	mintDenom, err := stringField("mint denom", "mint_denom", "mintDenom")
	if err != nil {
		return minttypes.Params{}, err
	}
	decodeDec := func(label string, names ...string) (sdkmath.LegacyDec, error) {
		text, fieldErr := stringField(label, names...)
		if fieldErr != nil {
			return sdkmath.LegacyDec{}, fieldErr
		}
		value, parseErr := sdkmath.LegacyNewDecFromStr(text)
		if parseErr != nil {
			return sdkmath.LegacyDec{}, fmt.Errorf("mint params %s %q is invalid: %w", label, text, parseErr)
		}
		return value, nil
	}
	inflationRateChange, err := decodeDec("inflation rate change", "inflation_rate_change", "inflationRateChange")
	if err != nil {
		return minttypes.Params{}, err
	}
	inflationMax, err := decodeDec("inflation max", "inflation_max", "inflationMax")
	if err != nil {
		return minttypes.Params{}, err
	}
	inflationMin, err := decodeDec("inflation min", "inflation_min", "inflationMin")
	if err != nil {
		return minttypes.Params{}, err
	}
	goalBonded, err := decodeDec("goal bonded", "goal_bonded", "goalBonded")
	if err != nil {
		return minttypes.Params{}, err
	}
	blocksRaw, err := requiredUpgradeSystemJSONField(paramsObject, "mint params", "blocks_per_year", "blocksPerYear")
	if err != nil {
		return minttypes.Params{}, err
	}
	blocksPerYear, err := decodeUpgradeSystemJSONUint64(blocksRaw, "mint params blocks_per_year")
	if err != nil {
		return minttypes.Params{}, err
	}
	if blocksPerYear == 0 {
		return minttypes.Params{}, errors.New("mint params blocks_per_year must be positive")
	}
	params := minttypes.Params{
		MintDenom:           mintDenom,
		InflationRateChange: inflationRateChange,
		InflationMax:        inflationMax,
		InflationMin:        inflationMin,
		GoalBonded:          goalBonded,
		BlocksPerYear:       blocksPerYear,
	}
	if err := params.Validate(); err != nil {
		return minttypes.Params{}, fmt.Errorf("validate mint params: %w", err)
	}
	return params, nil
}

func decodeUpgradeSystemMintRESTDecimal(response []byte, label string, names ...string) (string, error) {
	object, err := decodeUpgradeSystemJSONObject(response, "mint "+label+" response")
	if err != nil {
		return "", err
	}
	raw, err := requiredUpgradeSystemJSONField(object, "mint "+label+" response", names...)
	if err != nil {
		return "", err
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("mint %s response must contain a non-empty decimal string", label)
	}
	value = strings.TrimSpace(value)
	if _, err := sdkmath.LegacyNewDecFromStr(value); err != nil {
		return "", fmt.Errorf("mint %s response %q is invalid: %w", label, value, err)
	}
	return value, nil
}

func decodeUpgradeSystemJSONObject(contents []byte, label string) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()
	var object map[string]json.RawMessage
	if err := decoder.Decode(&object); err != nil {
		return nil, fmt.Errorf("decode %s: %w", label, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode %s: multiple JSON values", label)
		}
		return nil, fmt.Errorf("decode %s trailing value: %w", label, err)
	}
	if object == nil {
		return nil, fmt.Errorf("decode %s: object is required", label)
	}
	return object, nil
}

func requiredUpgradeSystemJSONField(
	object map[string]json.RawMessage,
	label string,
	names ...string,
) (json.RawMessage, error) {
	var result json.RawMessage
	var found string
	for _, name := range names {
		raw, ok := object[name]
		if !ok || len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			continue
		}
		if found != "" {
			return nil, fmt.Errorf("%s contains both %q and %q", label, found, name)
		}
		found = name
		result = raw
	}
	if found == "" {
		return nil, fmt.Errorf("%s is missing one of %v", label, names)
	}
	return result, nil
}

func optionalUpgradeSystemJSONField(
	object map[string]json.RawMessage,
	label string,
	names ...string,
) (json.RawMessage, bool, error) {
	var result json.RawMessage
	var found string
	for _, name := range names {
		raw, ok := object[name]
		if !ok || len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			continue
		}
		if found != "" {
			return nil, false, fmt.Errorf("%s contains both %q and %q", label, found, name)
		}
		found = name
		result = raw
	}
	return result, found != "", nil
}

func rejectUnknownUpgradeSystemJSONFields(
	object map[string]json.RawMessage,
	label string,
	allowed ...string,
) error {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}
	for name := range object {
		if _, ok := allowedSet[name]; !ok {
			return fmt.Errorf("%s contains unsupported field %q", label, name)
		}
	}
	return nil
}

func decodeUpgradeSystemConsensusState(raw []byte) (upgradeSystemConsensusState, error) {
	object, err := decodeUpgradeSystemJSONObject(raw, "consensus params")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}
	if err := rejectUnknownUpgradeSystemJSONFields(object, "consensus params", "block", "evidence", "validator", "version", "abci"); err != nil {
		return upgradeSystemConsensusState{}, err
	}
	decodeObject := func(label string, names ...string) (map[string]json.RawMessage, error) {
		field, fieldErr := requiredUpgradeSystemJSONField(object, "consensus params", names...)
		if fieldErr != nil {
			return nil, fieldErr
		}
		return decodeUpgradeSystemJSONObject(field, label)
	}
	decodeInt64 := func(container map[string]json.RawMessage, label string, names ...string) (int64, error) {
		field, fieldErr := requiredUpgradeSystemJSONField(container, label, names...)
		if fieldErr != nil {
			return 0, fieldErr
		}
		return decodeUpgradeSystemJSONInt64(field, label+" "+names[0])
	}

	block, err := decodeObject("consensus block params", "block")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}
	if err := rejectUnknownUpgradeSystemJSONFields(block, "consensus block params", "max_bytes", "maxBytes", "max_gas", "maxGas"); err != nil {
		return upgradeSystemConsensusState{}, err
	}
	blockMaxBytes, err := decodeInt64(block, "consensus block params", "max_bytes", "maxBytes")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}
	blockMaxGas, err := decodeInt64(block, "consensus block params", "max_gas", "maxGas")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}

	evidence, err := decodeObject("consensus evidence params", "evidence")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}
	if err := rejectUnknownUpgradeSystemJSONFields(
		evidence,
		"consensus evidence params",
		"max_age_num_blocks", "maxAgeNumBlocks",
		"max_age_duration", "maxAgeDuration",
		"max_bytes", "maxBytes",
	); err != nil {
		return upgradeSystemConsensusState{}, err
	}
	evidenceMaxAgeNumBlocks, err := decodeInt64(evidence, "consensus evidence params", "max_age_num_blocks", "maxAgeNumBlocks")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}
	evidenceMaxAgeDuration, err := decodeInt64(evidence, "consensus evidence params", "max_age_duration", "maxAgeDuration")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}
	evidenceMaxBytes, err := decodeInt64(evidence, "consensus evidence params", "max_bytes", "maxBytes")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}

	validator, err := decodeObject("consensus validator params", "validator")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}
	if err := rejectUnknownUpgradeSystemJSONFields(validator, "consensus validator params", "pub_key_types", "pubKeyTypes"); err != nil {
		return upgradeSystemConsensusState{}, err
	}
	pubKeyTypesRaw, err := requiredUpgradeSystemJSONField(validator, "consensus validator params", "pub_key_types", "pubKeyTypes")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}
	var pubKeyTypes []string
	if err := json.Unmarshal(pubKeyTypesRaw, &pubKeyTypes); err != nil || len(pubKeyTypes) == 0 {
		return upgradeSystemConsensusState{}, errors.New("consensus validator pub_key_types must be a non-empty string array")
	}
	for _, pubKeyType := range pubKeyTypes {
		if strings.TrimSpace(pubKeyType) == "" {
			return upgradeSystemConsensusState{}, errors.New("consensus validator pub_key_types contains an empty value")
		}
	}

	version, err := decodeObject("consensus version params", "version")
	if err != nil {
		return upgradeSystemConsensusState{}, err
	}
	if err := rejectUnknownUpgradeSystemJSONFields(version, "consensus version params", "app"); err != nil {
		return upgradeSystemConsensusState{}, err
	}
	var versionApp uint64
	if versionAppRaw, ok, fieldErr := optionalUpgradeSystemJSONField(version, "consensus version params", "app"); fieldErr != nil {
		return upgradeSystemConsensusState{}, fieldErr
	} else if ok {
		versionApp, err = decodeUpgradeSystemJSONUint64(versionAppRaw, "consensus version app")
		if err != nil {
			return upgradeSystemConsensusState{}, err
		}
	}

	var voteExtensionsEnableHeight int64
	if abciRaw, ok, fieldErr := optionalUpgradeSystemJSONField(object, "consensus params", "abci"); fieldErr != nil {
		return upgradeSystemConsensusState{}, fieldErr
	} else if ok {
		abci, decodeErr := decodeUpgradeSystemJSONObject(abciRaw, "consensus ABCI params")
		if decodeErr != nil {
			return upgradeSystemConsensusState{}, decodeErr
		}
		if err := rejectUnknownUpgradeSystemJSONFields(abci, "consensus ABCI params", "vote_extensions_enable_height", "voteExtensionsEnableHeight"); err != nil {
			return upgradeSystemConsensusState{}, err
		}
		if heightRaw, present, heightErr := optionalUpgradeSystemJSONField(
			abci,
			"consensus ABCI params",
			"vote_extensions_enable_height",
			"voteExtensionsEnableHeight",
		); heightErr != nil {
			return upgradeSystemConsensusState{}, heightErr
		} else if present {
			voteExtensionsEnableHeight, err = decodeUpgradeSystemJSONInt64(heightRaw, "consensus ABCI vote_extensions_enable_height")
			if err != nil {
				return upgradeSystemConsensusState{}, err
			}
		}
	}

	return upgradeSystemConsensusState{
		BlockMaxBytes:              blockMaxBytes,
		BlockMaxGas:                blockMaxGas,
		EvidenceMaxAgeNumBlocks:    evidenceMaxAgeNumBlocks,
		EvidenceMaxAgeDuration:     evidenceMaxAgeDuration,
		EvidenceMaxBytes:           evidenceMaxBytes,
		ValidatorPubKeyTypes:       append([]string(nil), pubKeyTypes...),
		VersionApp:                 versionApp,
		VoteExtensionsEnableHeight: voteExtensionsEnableHeight,
	}, nil
}

func upgradeSystemConsensusStatesEqual(left, right []byte) (bool, error) {
	leftState, err := decodeUpgradeSystemConsensusState(left)
	if err != nil {
		return false, fmt.Errorf("decode left consensus params: %w", err)
	}
	rightState, err := decodeUpgradeSystemConsensusState(right)
	if err != nil {
		return false, fmt.Errorf("decode right consensus params: %w", err)
	}
	return canonicalUpgradeSystemValuesEqual(leftState, rightState), nil
}

func decodeUpgradeSystemJSONUint64(raw json.RawMessage, label string) (uint64, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return 0, fmt.Errorf("%s is required", label)
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var number json.Number
		if decodeErr := decoder.Decode(&number); decodeErr != nil {
			return 0, fmt.Errorf("%s is invalid: %s", label, bytes.TrimSpace(raw))
		}
		text = number.String()
	}
	value, err := strconv.ParseUint(strings.TrimSpace(text), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %s", label, bytes.TrimSpace(raw))
	}
	return value, nil
}

func prepareUpgradeSystemGenesisValidationInput(
	phase string,
	contents []byte,
) ([]byte, upgradeSystemGenesisValidationEvidence, error) {
	sourceDigest, err := harness.CanonicalGenesisDigest(contents)
	if err != nil {
		return nil, upgradeSystemGenesisValidationEvidence{}, fmt.Errorf("digest system-module source export: %w", err)
	}
	evidence := upgradeSystemGenesisValidationEvidence{
		SourceDigest:          sourceDigest,
		ValidationInputDigest: sourceDigest,
	}
	if phase != upgradeSystemPreUpgradeCheckpointPhase {
		return append([]byte(nil), contents...), evidence, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()
	var document map[string]any
	if err := decoder.Decode(&document); err != nil {
		return nil, evidence, fmt.Errorf("decode v2.2.1 validation projection source: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, evidence, errors.New("decode v2.2.1 validation projection source: multiple JSON values")
		}
		return nil, evidence, fmt.Errorf("decode v2.2.1 validation projection trailing value: %w", err)
	}
	appState, err := upgradeSystemObjectField(document, "app_state", "v2.2.1 exported genesis")
	if err != nil {
		return nil, evidence, err
	}
	ibcState, err := upgradeSystemObjectField(appState, "ibc", "v2.2.1 app_state")
	if err != nil {
		return nil, evidence, err
	}
	connectionGenesis, err := upgradeSystemObjectField(ibcState, "connection_genesis", "v2.2.1 ibc state")
	if err != nil {
		return nil, evidence, err
	}
	connections, ok := connectionGenesis["connections"].([]any)
	if !ok {
		return nil, evidence, errors.New("v2.2.1 IBC connection_genesis connections must be an array")
	}

	sentinelIndex := -1
	var sentinel map[string]any
	for index, value := range connections {
		connection, ok := value.(map[string]any)
		if !ok {
			return nil, evidence, fmt.Errorf("v2.2.1 IBC connection %d must be an object", index)
		}
		if !upgradeSystemReferencesLocalhostConnection(connection) {
			continue
		}
		if sentinelIndex >= 0 {
			return nil, evidence, errors.New("v2.2.1 export contains multiple or partial localhost connections")
		}
		if err := validateUpgradeSystemV221LocalhostConnection(connection); err != nil {
			return nil, evidence, err
		}
		sentinelIndex = index
		sentinel = connection
	}
	if sentinelIndex < 0 {
		return nil, evidence, errors.New("v2.2.1 export must contain exactly one canonical runtime localhost connection")
	}

	projectedConnections := make([]any, 0, len(connections)-1)
	projectedConnections = append(projectedConnections, connections[:sentinelIndex]...)
	projectedConnections = append(projectedConnections, connections[sentinelIndex+1:]...)
	connectionGenesis["connections"] = projectedConnections
	validationInput, err := json.Marshal(document)
	if err != nil {
		return nil, evidence, fmt.Errorf("encode v2.2.1 validation projection: %w", err)
	}
	validationInputDigest, err := harness.CanonicalGenesisDigest(validationInput)
	if err != nil {
		return nil, evidence, fmt.Errorf("digest v2.2.1 validation projection: %w", err)
	}
	removedConnection, err := json.Marshal(sentinel)
	if err != nil {
		return nil, evidence, fmt.Errorf("encode removed v2.2.1 localhost connection: %w", err)
	}
	evidence.Projected = true
	evidence.ValidationInputDigest = validationInputDigest
	evidence.RemovedConnectionID = upgradeSystemLocalhostConnectionID
	evidence.RemovedConnection = removedConnection
	evidence.Reason = upgradeSystemLocalhostProjectionReason
	evidence.UpstreamVersion = upgradeSystemV221IBCGoVersion
	return validationInput, evidence, nil
}

func upgradeSystemObjectField(object map[string]any, field, label string) (map[string]any, error) {
	value, ok := object[field].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s %s must be an object", label, field)
	}
	return value, nil
}

func upgradeSystemReferencesLocalhostConnection(connection map[string]any) bool {
	if upgradeSystemStringField(connection, "id") == upgradeSystemLocalhostConnectionID ||
		upgradeSystemStringField(connection, "client_id") == upgradeSystemLocalhostClientID {
		return true
	}
	counterparty, _ := connection["counterparty"].(map[string]any)
	return upgradeSystemStringField(counterparty, "client_id") == upgradeSystemLocalhostClientID ||
		upgradeSystemStringField(counterparty, "connection_id") == upgradeSystemLocalhostConnectionID
}

func validateUpgradeSystemV221LocalhostConnection(connection map[string]any) error {
	if err := requireUpgradeSystemObjectKeys(
		connection,
		"v2.2.1 localhost connection",
		"client_id", "counterparty", "delay_period", "id", "state", "versions",
	); err != nil {
		return err
	}
	if got := upgradeSystemStringField(connection, "id"); got != upgradeSystemLocalhostConnectionID {
		return fmt.Errorf("v2.2.1 localhost connection id %q, want %q", got, upgradeSystemLocalhostConnectionID)
	}
	if got := upgradeSystemStringField(connection, "client_id"); got != upgradeSystemLocalhostClientID {
		return fmt.Errorf("v2.2.1 localhost connection client_id %q, want %q", got, upgradeSystemLocalhostClientID)
	}
	if got := upgradeSystemStringField(connection, "state"); got != "STATE_OPEN" {
		return fmt.Errorf("v2.2.1 localhost connection state %q, want STATE_OPEN", got)
	}
	if got := upgradeSystemStringField(connection, "delay_period"); got != "0" {
		return fmt.Errorf("v2.2.1 localhost connection delay_period %q, want 0", got)
	}
	counterparty, err := upgradeSystemObjectField(connection, "counterparty", "v2.2.1 localhost connection")
	if err != nil {
		return err
	}
	if err := requireUpgradeSystemObjectKeys(
		counterparty,
		"v2.2.1 localhost counterparty",
		"client_id", "connection_id", "prefix",
	); err != nil {
		return err
	}
	if got := upgradeSystemStringField(counterparty, "client_id"); got != upgradeSystemLocalhostClientID {
		return fmt.Errorf("v2.2.1 localhost counterparty client_id %q, want %q", got, upgradeSystemLocalhostClientID)
	}
	if got := upgradeSystemStringField(counterparty, "connection_id"); got != upgradeSystemLocalhostConnectionID {
		return fmt.Errorf("v2.2.1 localhost counterparty connection_id %q, want %q", got, upgradeSystemLocalhostConnectionID)
	}
	prefix, err := upgradeSystemObjectField(counterparty, "prefix", "v2.2.1 localhost counterparty")
	if err != nil {
		return err
	}
	if err := requireUpgradeSystemObjectKeys(prefix, "v2.2.1 localhost prefix", "key_prefix"); err != nil {
		return err
	}
	if got := upgradeSystemStringField(prefix, "key_prefix"); got != "aWJj" {
		return fmt.Errorf("v2.2.1 localhost key prefix %q, want aWJj", got)
	}
	versions, ok := connection["versions"].([]any)
	if !ok || len(versions) != 1 {
		return fmt.Errorf("v2.2.1 localhost connection must have exactly one version, got %T length %d", connection["versions"], len(versions))
	}
	version, ok := versions[0].(map[string]any)
	if !ok {
		return errors.New("v2.2.1 localhost connection version must be an object")
	}
	if err := requireUpgradeSystemObjectKeys(version, "v2.2.1 localhost version", "features", "identifier"); err != nil {
		return err
	}
	if got := upgradeSystemStringField(version, "identifier"); got != "1" {
		return fmt.Errorf("v2.2.1 localhost version identifier %q, want 1", got)
	}
	features, ok := version["features"].([]any)
	if !ok || len(features) != 2 || features[0] != "ORDER_ORDERED" || features[1] != "ORDER_UNORDERED" {
		return fmt.Errorf("v2.2.1 localhost version features = %v, want [ORDER_ORDERED ORDER_UNORDERED]", version["features"])
	}
	return nil
}

func requireUpgradeSystemObjectKeys(object map[string]any, label string, keys ...string) error {
	if len(object) != len(keys) {
		return fmt.Errorf("%s has %d fields, want exactly %d", label, len(object), len(keys))
	}
	for _, key := range keys {
		if _, ok := object[key]; !ok {
			return fmt.Errorf("%s is missing %q", label, key)
		}
	}
	return nil
}

func upgradeSystemStringField(object map[string]any, field string) string {
	value, _ := object[field].(string)
	return value
}

func captureUpgradeSystemModuleCheckpoint(
	ctx context.Context,
	network *harness.Network,
	phase string,
) (upgradeSystemModuleCheckpoint, error) {
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 || len(network.Chain.Validators) < 4 {
		return upgradeSystemModuleCheckpoint{}, errors.New("system-module checkpoint requires four validators and a full node")
	}
	if !systemModulePhaseIsSafe(phase) {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("invalid system-module phase %q", phase)
	}
	exportNode := network.Chain.Validators[3]
	height, err := exportNode.Height(ctx)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("query system-module export height: %w", err)
	}
	if height <= 1 {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("system-module export height must exceed 1, got %d", height)
	}
	if err := network.WaitForFullNode(ctx, height); err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("wait full node for system-module height %d: %w", height, err)
	}
	fullNode := network.Chain.FullNodes[0]
	observation, err := network.CaptureUpgradeCheckpointObservation(ctx, "upgrade-system-"+phase, fullNode, height)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, err
	}
	exportEvidence, err := network.ExportValidatorGenesisDeterministically(
		ctx,
		"upgrade-system-"+phase,
		3,
		height,
	)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("%s system-module export: %w", phase, err)
	}
	if err := network.WaitForNodeHeight(ctx, exportNode, height+1); err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("wait for system-module export node after %s: %w", phase, err)
	}
	exported, err := decodeUpgradeSystemModuleExport(exportEvidence.Contents)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("%s system-module export state: %w", phase, err)
	}
	if exported.Height != height || exportEvidence.Height != height {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf(
			"%s system-module export height mismatch: query=%d decoded=%d evidence=%d",
			phase,
			height,
			exported.Height,
			exportEvidence.Height,
		)
	}
	if err := network.WriteArtifact(
		"upgrade/system-modules/exports/"+phase+".json",
		exportEvidence.Contents,
	); err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("record %s system-module export: %w", phase, err)
	}
	validationInput, genesisValidation, err := prepareUpgradeSystemGenesisValidationInput(phase, exportEvidence.Contents)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("prepare %s system-module genesis validation input: %w", phase, err)
	}
	if genesisValidation.SourceDigest != exportEvidence.Digest {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf(
			"%s validation source digest changed: export=%s validation-source=%s",
			phase,
			exportEvidence.Digest,
			genesisValidation.SourceDigest,
		)
	}
	if genesisValidation.Projected {
		if err := network.WriteArtifact(
			"upgrade/system-modules/exports/"+phase+"-validation-input.json",
			validationInput,
		); err != nil {
			return upgradeSystemModuleCheckpoint{}, fmt.Errorf("record %s projected genesis validation input: %w", phase, err)
		}
	}
	validationEvidence, err := network.ValidateGenesisDocument(ctx, phase, exportNode, validationInput)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("validate %s system-module export: %w", phase, err)
	}
	if validationEvidence.CanonicalDigest != genesisValidation.ValidationInputDigest {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf(
			"%s validation input digest changed: expected=%s public-validation=%s",
			phase,
			genesisValidation.ValidationInputDigest,
			validationEvidence.CanonicalDigest,
		)
	}
	genesisValidation.PublicValidation = validationEvidence
	if err := network.WriteArtifactJSON(
		"upgrade/system-modules/exports/"+phase+"-validation-contract.json",
		genesisValidation,
	); err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("record %s genesis validation contract: %w", phase, err)
	}

	mintPaths := upgradeSystemMintRESTQueryPaths()
	mintSteps := []string{"params", "inflation", "annual-provisions"}
	mintResponses := make([]json.RawMessage, len(mintPaths))
	for index, path := range mintPaths {
		mintResponses[index], err = network.FullNodeRESTGetAtHeight(
			ctx,
			nil,
			"upgrade-system-"+phase+"-mint-"+mintSteps[index],
			path,
			height,
		)
		if err != nil {
			return upgradeSystemModuleCheckpoint{}, fmt.Errorf(
				"query %s mint %s at height %d: %w",
				phase,
				mintSteps[index],
				height,
				err,
			)
		}
	}
	mintState, err := decodeUpgradeSystemMintRESTResponses(
		mintResponses[0],
		mintResponses[1],
		mintResponses[2],
	)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("decode %s mint REST state at height %d: %w", phase, height, err)
	}

	pinnedCtx := metadata.AppendToOutgoingContext(ctx, "x-cosmos-block-height", strconv.FormatInt(height, 10))
	consensusResponse, err := consensustypes.NewQueryClient(fullNode.GrpcConn).Params(
		pinnedCtx,
		&consensustypes.QueryParamsRequest{},
	)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("query %s consensus params at height %d: %w", phase, height, err)
	}
	if consensusResponse.GetParams() == nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("query %s consensus params returned nil", phase)
	}
	consensusJSON, err := json.Marshal(consensusResponse.GetParams())
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("encode %s consensus params: %w", phase, err)
	}
	consensusJSON, err = canonicalizeUpgradeSystemJSON(consensusJSON)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("canonicalize %s consensus params: %w", phase, err)
	}
	bankClient := banktypes.NewQueryClient(fullNode.GrpcConn)
	supplyResponse, err := bankClient.SupplyOf(
		pinnedCtx,
		&banktypes.QuerySupplyOfRequest{Denom: upgradeSystemDenom},
	)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("query %s bank supply at height %d: %w", phase, height, err)
	}
	if !supplyResponse.GetAmount().IsValid() || supplyResponse.GetAmount().Denom != upgradeSystemDenom {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("query %s bank supply returned invalid amount %s", phase, supplyResponse.GetAmount())
	}
	burnBalanceResponse, err := bankClient.Balance(
		pinnedCtx,
		&banktypes.QueryBalanceRequest{Address: upgradeBurnAddress, Denom: upgradeSystemDenom},
	)
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("query %s burn balance at height %d: %w", phase, height, err)
	}
	burnBalance := sdkmath.ZeroInt()
	if burnBalanceResponse.GetBalance() != nil {
		burnBalance = burnBalanceResponse.GetBalance().Amount
	}
	legacyParams, err := queryAllUpgradeLegacyParams(pinnedCtx, paramsproposal.NewQueryClient(fullNode.GrpcConn))
	if err != nil {
		return upgradeSystemModuleCheckpoint{}, fmt.Errorf("query %s legacy params at height %d: %w", phase, height, err)
	}
	checkpoint := upgradeSystemModuleCheckpoint{
		Phase:                 phase,
		RecordedAt:            observation.ObservedAt,
		Height:                height,
		Observation:           observation,
		MintParams:            mintState.Params,
		Inflation:             mintState.Inflation,
		AnnualProvisions:      mintState.AnnualProvisions,
		ConsensusParams:       consensusJSON,
		LegacyParams:          legacyParams,
		Supply:                supplyResponse.GetAmount().Amount.String(),
		BurnAddressBalance:    burnBalance.String(),
		Export:                exported,
		ExportDigest:          exportEvidence.Digest,
		GenesisValidation:     genesisValidation,
		ExportValidated:       true,
		HistoricalPublicQuery: true,
		MintQueryBoundary:     upgradeSystemMintRESTBoundary,
		MintQueryHeight:       height,
		MintQueryPaths:        mintPaths,
	}
	if err := validateUpgradeSystemModuleCheckpoint(checkpoint); err != nil {
		return checkpoint, fmt.Errorf("validate %s system-module checkpoint: %w", phase, err)
	}
	if err := network.WriteArtifactJSON(
		"upgrade/system-modules/checkpoints/"+phase+".json",
		checkpoint,
	); err != nil {
		return checkpoint, fmt.Errorf("record %s system-module checkpoint: %w", phase, err)
	}
	if err := network.AppendArtifactJSON("upgrade/system-modules/phases.jsonl", checkpoint); err != nil {
		return checkpoint, fmt.Errorf("append %s system-module phase: %w", phase, err)
	}
	return checkpoint, nil
}

func validateUpgradeSystemModuleCheckpoint(checkpoint upgradeSystemModuleCheckpoint) error {
	var validationErrors []error
	if !systemModulePhaseIsSafe(checkpoint.Phase) {
		validationErrors = append(validationErrors, fmt.Errorf("invalid phase %q", checkpoint.Phase))
	}
	if checkpoint.Height <= 1 {
		validationErrors = append(validationErrors, fmt.Errorf("height must exceed 1, got %d", checkpoint.Height))
	}
	if err := checkpoint.Observation.Validate(); err != nil {
		validationErrors = append(validationErrors, err)
	} else if checkpoint.Observation.Height != checkpoint.Height {
		validationErrors = append(validationErrors, fmt.Errorf(
			"observation height %d does not match checkpoint %d",
			checkpoint.Observation.Height,
			checkpoint.Height,
		))
	}
	if checkpoint.Export.Height != checkpoint.Height {
		validationErrors = append(validationErrors, fmt.Errorf("export height %d does not match checkpoint %d", checkpoint.Export.Height, checkpoint.Height))
	}
	if checkpoint.MintParams.MintDenom != upgradeSystemDenom {
		validationErrors = append(validationErrors, fmt.Errorf("mint denom %q, want %q", checkpoint.MintParams.MintDenom, upgradeSystemDenom))
	}
	if err := checkpoint.MintParams.Validate(); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("mint params are invalid: %w", err))
	}
	inflation, err := sdkmath.LegacyNewDecFromStr(checkpoint.Inflation)
	if err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("inflation %q is invalid: %w", checkpoint.Inflation, err))
	} else if inflation.LT(checkpoint.MintParams.InflationMin) || inflation.GT(checkpoint.MintParams.InflationMax) {
		validationErrors = append(validationErrors, fmt.Errorf(
			"inflation %s is outside [%s,%s]",
			inflation,
			checkpoint.MintParams.InflationMin,
			checkpoint.MintParams.InflationMax,
		))
	}
	annual, err := sdkmath.LegacyNewDecFromStr(checkpoint.AnnualProvisions)
	if err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("annual provisions %q are invalid: %w", checkpoint.AnnualProvisions, err))
	} else if annual.IsNegative() {
		validationErrors = append(validationErrors, fmt.Errorf("annual provisions %s are negative", annual))
	}
	supply, ok := sdkmath.NewIntFromString(checkpoint.Supply)
	if !ok || !supply.IsPositive() {
		validationErrors = append(validationErrors, fmt.Errorf("supply %q must be a positive integer", checkpoint.Supply))
	}
	burnBalance, ok := sdkmath.NewIntFromString(checkpoint.BurnAddressBalance)
	if !ok || !burnBalance.IsZero() {
		validationErrors = append(validationErrors, fmt.Errorf("burn-address balance %q must be zero", checkpoint.BurnAddressBalance))
	}
	if len(checkpoint.ConsensusParams) == 0 || bytes.Equal(bytes.TrimSpace(checkpoint.ConsensusParams), []byte("null")) {
		validationErrors = append(validationErrors, errors.New("consensus params are required"))
	}
	if len(checkpoint.LegacyParams) == 0 {
		validationErrors = append(validationErrors, errors.New("legacy params state is required"))
	}
	if len(checkpoint.Export.ModuleNames) != len(upgradeSystemModuleNames) {
		validationErrors = append(validationErrors, fmt.Errorf("export has %d module names, want %d", len(checkpoint.Export.ModuleNames), len(upgradeSystemModuleNames)))
	}
	for _, moduleName := range upgradeSystemModuleNames {
		if len(checkpoint.Export.Modules[moduleName]) == 0 || checkpoint.Export.ModuleDigests[moduleName] == "" {
			validationErrors = append(validationErrors, fmt.Errorf("export lacks %s module evidence", moduleName))
		}
	}
	if !checkpoint.HistoricalPublicQuery {
		validationErrors = append(validationErrors, errors.New("historical-height public query evidence is required"))
	}
	if checkpoint.MintQueryBoundary != upgradeSystemMintRESTBoundary || checkpoint.MintQueryHeight != checkpoint.Height {
		validationErrors = append(validationErrors, fmt.Errorf(
			"mint query evidence boundary=%q height=%d, want REST at checkpoint height %d",
			checkpoint.MintQueryBoundary,
			checkpoint.MintQueryHeight,
			checkpoint.Height,
		))
	}
	if strings.Join(checkpoint.MintQueryPaths, "\x00") != strings.Join(upgradeSystemMintRESTQueryPaths(), "\x00") {
		validationErrors = append(validationErrors, fmt.Errorf("mint REST query paths = %v", checkpoint.MintQueryPaths))
	}
	if !checkpoint.ExportValidated || checkpoint.ExportDigest == "" {
		validationErrors = append(validationErrors, errors.New("deterministic export validation evidence is required"))
	}
	validation := checkpoint.GenesisValidation
	if validation.SourceDigest == "" || validation.SourceDigest != checkpoint.ExportDigest {
		validationErrors = append(validationErrors, fmt.Errorf(
			"genesis validation source digest %q does not match export %q",
			validation.SourceDigest,
			checkpoint.ExportDigest,
		))
	}
	if validation.ValidationInputDigest == "" ||
		validation.PublicValidation.CanonicalDigest != validation.ValidationInputDigest {
		validationErrors = append(validationErrors, errors.New("public genesis validation digest evidence is required"))
	}
	if checkpoint.Phase == upgradeSystemPreUpgradeCheckpointPhase {
		if !validation.Projected ||
			validation.RemovedConnectionID != upgradeSystemLocalhostConnectionID ||
			len(validation.RemovedConnection) == 0 ||
			validation.Reason != upgradeSystemLocalhostProjectionReason ||
			validation.UpstreamVersion != upgradeSystemV221IBCGoVersion ||
			validation.SourceDigest == validation.ValidationInputDigest {
			validationErrors = append(validationErrors, errors.New("v2.2.1 localhost genesis validation projection evidence is incomplete"))
		}
	} else if validation.Projected ||
		validation.RemovedConnectionID != "" ||
		len(validation.RemovedConnection) != 0 ||
		validation.Reason != "" ||
		validation.UpstreamVersion != "" ||
		validation.SourceDigest != validation.ValidationInputDigest {
		validationErrors = append(validationErrors, errors.New("genesis validation projection is forbidden after the v2.2.1 pre-upgrade checkpoint"))
	}
	if err := validateUpgradeSystemExportBoundary(checkpoint); err != nil {
		validationErrors = append(validationErrors, err)
	}
	return errors.Join(validationErrors...)
}

func validateUpgradeSystemExportBoundary(checkpoint upgradeSystemModuleCheckpoint) error {
	var validationErrors []error
	if !bytes.Equal(checkpoint.Export.Modules["burn"], []byte("{}")) {
		validationErrors = append(validationErrors, fmt.Errorf("burn genesis state = %s, want {}", checkpoint.Export.Modules["burn"]))
	}

	var capabilityState struct {
		Index json.RawMessage `json:"index"`
	}
	if err := json.Unmarshal(checkpoint.Export.Modules["capability"], &capabilityState); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("decode capability export: %w", err))
	} else if index, err := decodeUpgradeSystemJSONInt64(capabilityState.Index, "capability index"); err != nil {
		validationErrors = append(validationErrors, err)
	} else if index <= 0 {
		validationErrors = append(validationErrors, fmt.Errorf("capability index must be positive, got %d", index))
	}

	var crisisState struct {
		ConstantFee struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"constant_fee"`
	}
	if err := json.Unmarshal(checkpoint.Export.Modules["crisis"], &crisisState); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("decode crisis export: %w", err))
	} else {
		amount, ok := sdkmath.NewIntFromString(crisisState.ConstantFee.Amount)
		if crisisState.ConstantFee.Denom != upgradeSystemDenom || !ok || !amount.IsPositive() {
			validationErrors = append(validationErrors, fmt.Errorf(
				"crisis constant fee = %s%s, want a positive %s coin",
				crisisState.ConstantFee.Amount,
				crisisState.ConstantFee.Denom,
				upgradeSystemDenom,
			))
		}
	}

	var bankState struct {
		Supply []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"supply"`
	}
	if err := json.Unmarshal(checkpoint.Export.Modules["bank"], &bankState); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("decode bank export: %w", err))
	} else {
		var exportSupply string
		for _, coin := range bankState.Supply {
			if coin.Denom != upgradeSystemDenom {
				continue
			}
			if exportSupply != "" {
				validationErrors = append(validationErrors, errors.New("bank export contains duplicate umed supply"))
			}
			exportSupply = coin.Amount
		}
		if exportSupply != checkpoint.Supply {
			validationErrors = append(validationErrors, fmt.Errorf("bank export supply %q, gRPC supply %q", exportSupply, checkpoint.Supply))
		}
	}

	var mintState struct {
		Minter struct {
			Inflation        string `json:"inflation"`
			AnnualProvisions string `json:"annual_provisions"`
		} `json:"minter"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(checkpoint.Export.Modules["mint"], &mintState); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("decode mint export: %w", err))
	} else {
		if mintState.Minter.Inflation != checkpoint.Inflation {
			validationErrors = append(validationErrors, fmt.Errorf("mint export inflation %q, REST inflation %q", mintState.Minter.Inflation, checkpoint.Inflation))
		}
		if mintState.Minter.AnnualProvisions != checkpoint.AnnualProvisions {
			validationErrors = append(validationErrors, fmt.Errorf("mint export annual provisions %q, REST value %q", mintState.Minter.AnnualProvisions, checkpoint.AnnualProvisions))
		}
		exportParams, err := decodeUpgradeSystemMintParams(mintState.Params)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("decode mint export params: %w", err))
		} else if !canonicalUpgradeSystemValuesEqual(exportParams, checkpoint.MintParams) {
			validationErrors = append(validationErrors, fmt.Errorf(
				"mint export params differ from REST params: export=%v REST=%v",
				exportParams,
				checkpoint.MintParams,
			))
		}
	}

	if exportConsensus := checkpoint.Export.Modules["consensus"]; len(exportConsensus) == 0 {
		validationErrors = append(validationErrors, errors.New("consensus export params are missing"))
	} else if equal, err := upgradeSystemConsensusStatesEqual(exportConsensus, checkpoint.ConsensusParams); err != nil {
		validationErrors = append(validationErrors, fmt.Errorf("compare consensus export and gRPC params: %w", err))
	} else if !equal {
		validationErrors = append(validationErrors, errors.New("consensus export params differ semantically from gRPC params"))
	}
	return errors.Join(validationErrors...)
}

func validateUpgradeSystemModulePreservation(
	before upgradeSystemModuleCheckpoint,
	after upgradeSystemModuleCheckpoint,
) error {
	if err := validateUpgradeSystemModuleCheckpoint(before); err != nil {
		return fmt.Errorf("before checkpoint: %w", err)
	}
	if err := validateUpgradeSystemModuleCheckpoint(after); err != nil {
		return fmt.Errorf("after checkpoint: %w", err)
	}
	if after.Height <= before.Height {
		return fmt.Errorf("system-module height did not advance: before=%d after=%d", before.Height, after.Height)
	}
	if !canonicalUpgradeSystemValuesEqual(before.MintParams, after.MintParams) {
		return errors.New("mint params changed")
	}
	consensusEqual, err := upgradeSystemConsensusStatesEqual(before.ConsensusParams, after.ConsensusParams)
	if err != nil {
		return fmt.Errorf("compare consensus params across upgrade: %w", err)
	}
	if !consensusEqual {
		return errors.New("consensus params changed")
	}
	if !canonicalUpgradeSystemValuesEqual(before.LegacyParams, after.LegacyParams) {
		return errors.New("legacy params state changed")
	}
	beforeSupply, _ := sdkmath.NewIntFromString(before.Supply)
	afterSupply, _ := sdkmath.NewIntFromString(after.Supply)
	if afterSupply.LT(beforeSupply) {
		return fmt.Errorf("supply decreased across preservation boundary: before=%s after=%s", beforeSupply, afterSupply)
	}
	for _, moduleName := range []string{"burn", "capability", "crisis"} {
		if before.Export.ModuleDigests[moduleName] != after.Export.ModuleDigests[moduleName] {
			return fmt.Errorf(
				"%s module state changed: before=%s after=%s",
				moduleName,
				before.Export.ModuleDigests[moduleName],
				after.Export.ModuleDigests[moduleName],
			)
		}
	}
	return nil
}

func queryAllUpgradeLegacyParams(
	ctx context.Context,
	client paramsproposal.QueryClient,
) (map[string]map[string]string, error) {
	if client == nil {
		return nil, errors.New("legacy params query client is required")
	}
	response, err := client.Subspaces(ctx, &paramsproposal.QuerySubspacesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list legacy params subspaces: %w", err)
	}
	if len(response.GetSubspaces()) == 0 {
		return nil, errors.New("legacy params query returned no subspaces")
	}
	state := make(map[string]map[string]string, len(response.GetSubspaces()))
	for _, subspace := range response.GetSubspaces() {
		if subspace == nil || strings.TrimSpace(subspace.GetSubspace()) == "" {
			return nil, errors.New("legacy params query returned an unnamed subspace")
		}
		name := subspace.GetSubspace()
		if _, exists := state[name]; exists {
			return nil, fmt.Errorf("legacy params query returned duplicate subspace %q", name)
		}
		keys := make(map[string]string, len(subspace.GetKeys()))
		for _, key := range subspace.GetKeys() {
			if strings.TrimSpace(key) == "" {
				return nil, fmt.Errorf("legacy params subspace %q returned an empty key", name)
			}
			if _, exists := keys[key]; exists {
				return nil, fmt.Errorf("legacy params subspace %q returned duplicate key %q", name, key)
			}
			paramResponse, err := client.Params(ctx, &paramsproposal.QueryParamsRequest{
				Subspace: name,
				Key:      key,
			})
			if err != nil {
				return nil, fmt.Errorf("query legacy param %s/%s: %w", name, key, err)
			}
			param := paramResponse.GetParam()
			if param.Subspace != name || param.Key != key {
				return nil, fmt.Errorf(
					"legacy param response identity %s/%s, want %s/%s",
					param.Subspace,
					param.Key,
					name,
					key,
				)
			}
			keys[key] = param.Value
		}
		state[name] = keys
	}
	return state, nil
}

func mutateUpgradeSystemModules(
	ctx context.Context,
	network *harness.Network,
	preparation upgradeSystemPreparation,
) (upgradeSystemMutation, error) {
	startHeight, err := network.Chain.Height(ctx)
	if err != nil {
		return upgradeSystemMutation{}, fmt.Errorf("query system-module mutation start height: %w", err)
	}
	wallet := preparation.Wallet
	if wallet == nil {
		return upgradeSystemMutation{}, errors.New("system-module burner wallet is unavailable")
	}
	if wallet.FormattedAddress() != preparation.BurnerAddress {
		return upgradeSystemMutation{}, fmt.Errorf(
			"system-module burner changed: got %s want %s",
			wallet.FormattedAddress(),
			preparation.BurnerAddress,
		)
	}
	burnTxHash, err := broadcastUpgradeSystemBurn(ctx, network, wallet, "post-upgrade", upgradeSystemBurnAmount)
	if err != nil {
		return upgradeSystemMutation{}, err
	}
	if err := network.WaitForHeight(ctx, startHeight+3); err != nil {
		return upgradeSystemMutation{}, fmt.Errorf("wait for normal post-upgrade block lifecycle: %w", err)
	}
	checkpoint, err := captureUpgradeSystemModuleCheckpoint(ctx, network, "post-upgrade-mutation")
	if err != nil {
		return upgradeSystemMutation{}, err
	}
	mutation := upgradeSystemMutation{
		BurnTxHash:      burnTxHash,
		BurnAmount:      upgradeSystemBurnAmount,
		StartHeight:     startHeight,
		EndHeight:       checkpoint.Height,
		BlocksCommitted: checkpoint.Height - startHeight,
		Checkpoint:      checkpoint,
	}
	if mutation.BlocksCommitted < 3 {
		return mutation, fmt.Errorf("normal block lifecycle committed %d blocks, want at least 3", mutation.BlocksCommitted)
	}
	if err := network.WriteArtifactJSON("upgrade/system-modules/mutation.json", mutation); err != nil {
		return mutation, fmt.Errorf("record system-module mutation: %w", err)
	}
	return mutation, nil
}

func systemModulePhaseIsSafe(phase string) bool {
	if phase == "" || strings.ContainsAny(phase, "/\\") {
		return false
	}
	for _, part := range strings.Split(phase, "-") {
		if part == "" {
			return false
		}
		for _, char := range part {
			if (char < 'a' || char > 'z') && (char < '0' || char > '9') {
				return false
			}
		}
	}
	return true
}

func canonicalUpgradeSystemValuesEqual(left, right any) bool {
	leftJSON, err := json.Marshal(left)
	if err != nil {
		return false
	}
	rightJSON, err := json.Marshal(right)
	if err != nil {
		return false
	}
	leftCanonical, err := canonicalizeUpgradeSystemJSON(leftJSON)
	if err != nil {
		return false
	}
	rightCanonical, err := canonicalizeUpgradeSystemJSON(rightJSON)
	if err != nil {
		return false
	}
	return bytes.Equal(leftCanonical, rightCanonical)
}

func decodeUpgradeSystemModuleExport(contents []byte) (upgradeSystemModuleExport, error) {
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.UseNumber()
	var document struct {
		InitialHeight json.RawMessage            `json:"initial_height"`
		AppState      map[string]json.RawMessage `json:"app_state"`
		Consensus     json.RawMessage            `json:"consensus"`
		ConsensusOld  json.RawMessage            `json:"consensus_params"`
	}
	if err := decoder.Decode(&document); err != nil {
		return upgradeSystemModuleExport{}, fmt.Errorf("decode exported genesis: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return upgradeSystemModuleExport{}, errors.New("decode exported genesis: multiple JSON values")
		}
		return upgradeSystemModuleExport{}, fmt.Errorf("decode exported genesis trailing value: %w", err)
	}
	initialHeight, err := decodeUpgradeSystemJSONInt64(document.InitialHeight, "initial_height")
	if err != nil {
		return upgradeSystemModuleExport{}, err
	}
	if initialHeight <= 1 {
		return upgradeSystemModuleExport{}, fmt.Errorf("initial_height must be greater than 1, got %d", initialHeight)
	}
	if len(document.AppState) == 0 {
		return upgradeSystemModuleExport{}, errors.New("exported genesis app_state is empty")
	}

	result := upgradeSystemModuleExport{
		Height:        initialHeight - 1,
		ModuleNames:   append([]string(nil), upgradeSystemModuleNames...),
		Modules:       make(map[string]json.RawMessage, len(upgradeSystemModuleNames)),
		ModuleDigests: make(map[string]string, len(upgradeSystemModuleNames)),
	}
	for _, name := range upgradeSystemModuleNames {
		raw, ok := document.AppState[name]
		if name == "consensus" {
			raw, ok = exportedConsensusParams(raw, document.Consensus, document.ConsensusOld)
		}
		if !ok || len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return upgradeSystemModuleExport{}, fmt.Errorf("exported genesis is missing non-null %s module state", name)
		}
		canonical, err := canonicalizeUpgradeSystemJSON(raw)
		if err != nil {
			return upgradeSystemModuleExport{}, fmt.Errorf("canonicalize %s module state: %w", name, err)
		}
		digest, err := harness.CanonicalGenesisDigest(canonical)
		if err != nil {
			return upgradeSystemModuleExport{}, fmt.Errorf("digest %s module state: %w", name, err)
		}
		result.Modules[name] = canonical
		result.ModuleDigests[name] = digest
	}
	return result, nil
}

func exportedConsensusParams(appState, current, legacy json.RawMessage) (json.RawMessage, bool) {
	for index, candidate := range []json.RawMessage{appState, current} {
		if len(bytes.TrimSpace(candidate)) == 0 || bytes.Equal(bytes.TrimSpace(candidate), []byte("null")) {
			continue
		}
		var envelope struct {
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(candidate, &envelope); err == nil && len(bytes.TrimSpace(envelope.Params)) > 0 && !bytes.Equal(bytes.TrimSpace(envelope.Params), []byte("null")) {
			return envelope.Params, true
		}
		if index == 0 {
			return candidate, true
		}
	}
	if len(bytes.TrimSpace(legacy)) > 0 && !bytes.Equal(bytes.TrimSpace(legacy), []byte("null")) {
		return legacy, true
	}
	return nil, false
}

func decodeUpgradeSystemJSONInt64(raw json.RawMessage, label string) (int64, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return 0, fmt.Errorf("%s is required", label)
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		value, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%s is invalid: %q", label, text)
		}
		return value, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var number json.Number
	if err := decoder.Decode(&number); err != nil {
		return 0, fmt.Errorf("%s is invalid: %s", label, bytes.TrimSpace(raw))
	}
	value, err := number.Int64()
	if err != nil {
		return 0, fmt.Errorf("%s is invalid: %s", label, bytes.TrimSpace(raw))
	}
	return value, nil
}

func canonicalizeUpgradeSystemJSON(raw []byte) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("multiple JSON values")
		}
		return nil, err
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(canonical), nil
}
