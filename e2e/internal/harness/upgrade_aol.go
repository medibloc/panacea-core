package harness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

const (
	AOLUpgradeEvidenceSchemaVersion = "2"

	AOLUpgradePreparationArtifactPath      = "upgrade/aol/v2.2.1-preparation.json"
	AOLUpgradePreCheckpointArtifactPath    = "upgrade/aol/pre-upgrade-checkpoint.json"
	AOLUpgradePostPreservationArtifactPath = "upgrade/aol/post-upgrade-preservation.json"
	AOLUpgradePostMutationArtifactPath     = "upgrade/aol/post-upgrade-mutation.json"
	AOLUpgradePostRestartArtifactPath      = "upgrade/aol/post-restart.json"

	aolUpgradeQueryLimit = 100
	aolUpgradeTxGas      = "500000"
)

// AOLUpgradeFixture describes one dedicated owner/topic/writer graph that is
// created on v2.2.1 and then exercised through all five upgrade phases. The
// caller owns wallet creation and funding so the same identities remain in the
// keyring when the container image changes.
type AOLUpgradeFixture struct {
	OwnerAddress              string                `json:"owner_address"`
	OwnerKeyName              string                `json:"owner_key_name"`
	WriterAddress             string                `json:"writer_address"`
	WriterKeyName             string                `json:"writer_key_name"`
	MutationWriterAddress     string                `json:"mutation_writer_address"`
	TopicName                 string                `json:"topic_name"`
	TopicDescription          string                `json:"topic_description"`
	WriterMoniker             string                `json:"writer_moniker"`
	WriterDescription         string                `json:"writer_description"`
	MutationWriterMoniker     string                `json:"mutation_writer_moniker"`
	MutationWriterDescription string                `json:"mutation_writer_description"`
	InitialRecord             AOLUpgradeRecordInput `json:"initial_record"`
	MutationRecord            AOLUpgradeRecordInput `json:"mutation_record"`
}

// AOLUpgradeRecordInput is passed verbatim to the AOL CLI. The live query
// checkpoint later decodes the protobuf-JSON base64 bytes back into []byte.
type AOLUpgradeRecordInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Validate rejects fixtures that cannot exercise distinct owner, persistent
// writer, transient writer, and non-empty record state.
func (f AOLUpgradeFixture) Validate() error {
	required := []struct {
		name  string
		value string
	}{
		{name: "owner_address", value: f.OwnerAddress},
		{name: "owner_key_name", value: f.OwnerKeyName},
		{name: "writer_address", value: f.WriterAddress},
		{name: "writer_key_name", value: f.WriterKeyName},
		{name: "mutation_writer_address", value: f.MutationWriterAddress},
		{name: "topic_name", value: f.TopicName},
		{name: "topic_description", value: f.TopicDescription},
		{name: "writer_moniker", value: f.WriterMoniker},
		{name: "writer_description", value: f.WriterDescription},
		{name: "mutation_writer_moniker", value: f.MutationWriterMoniker},
		{name: "mutation_writer_description", value: f.MutationWriterDescription},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("AOL upgrade fixture %s is required", field.name)
		}
	}
	if f.OwnerAddress == f.WriterAddress ||
		f.OwnerAddress == f.MutationWriterAddress ||
		f.WriterAddress == f.MutationWriterAddress {
		return errors.New("AOL upgrade fixture owner, writer, and mutation writer addresses must be distinct")
	}
	if len(f.TopicName) > 70 || !regexp.MustCompile(`^[A-Za-z0-9._-]+$`).MatchString(f.TopicName) {
		return fmt.Errorf("AOL upgrade fixture topic_name %q is invalid", f.TopicName)
	}
	if len(f.TopicDescription) > 5000 || len(f.WriterDescription) > 5000 || len(f.MutationWriterDescription) > 5000 {
		return errors.New("AOL upgrade fixture description exceeds 5000 bytes")
	}
	if err := validateAOLUpgradeMoniker("writer_moniker", f.WriterMoniker); err != nil {
		return err
	}
	if err := validateAOLUpgradeMoniker("mutation_writer_moniker", f.MutationWriterMoniker); err != nil {
		return err
	}
	if err := f.InitialRecord.validate("initial_record"); err != nil {
		return err
	}
	if err := f.MutationRecord.validate("mutation_record"); err != nil {
		return err
	}
	return nil
}

func validateAOLUpgradeMoniker(field, value string) error {
	if len(value) > 70 || !regexp.MustCompile(`^[A-Za-z0-9._-]+$`).MatchString(value) {
		return fmt.Errorf("AOL upgrade fixture %s %q is invalid", field, value)
	}
	return nil
}

func (r AOLUpgradeRecordInput) validate(field string) error {
	if len(r.Key) == 0 || len(r.Key) > 70 {
		return fmt.Errorf("AOL upgrade fixture %s key length must be between 1 and 70 bytes", field)
	}
	if len(r.Value) == 0 || len(r.Value) > 5000 {
		return fmt.Errorf("AOL upgrade fixture %s value length must be between 1 and 5000 bytes", field)
	}
	return nil
}

// StateObjectIDs returns stable identifiers suitable for the AOL row in the
// upgrade coverage matrix.
func (f AOLUpgradeFixture) StateObjectIDs() []string {
	return []string{
		"aol-owner:" + f.OwnerAddress,
		fmt.Sprintf("aol-topic:%s/%s", f.OwnerAddress, f.TopicName),
		fmt.Sprintf("aol-writer:%s/%s/%s", f.OwnerAddress, f.TopicName, f.WriterAddress),
		fmt.Sprintf("aol-writer:%s/%s/%s", f.OwnerAddress, f.TopicName, f.MutationWriterAddress),
		fmt.Sprintf("aol-record:%s/%s/0", f.OwnerAddress, f.TopicName),
		fmt.Sprintf("aol-record:%s/%s/1", f.OwnerAddress, f.TopicName),
	}
}

// AOLUpgradeTransactionEvidence is the stable committed subset retained by
// the phase artifact. Complete raw lifecycle evidence remains in tx/*.jsonl.
type AOLUpgradeTransactionEvidence struct {
	Step   string `json:"step"`
	TxHash string `json:"tx_hash"`
	Height int64  `json:"height"`
}

// AOLUpgradePreparationEvidence binds the three source-version commits to the
// concrete AOL state-object identifiers used by later phases.
type AOLUpgradePreparationEvidence struct {
	SchemaVersion       string                          `json:"schema_version"`
	RecordedAt          time.Time                       `json:"recorded_at"`
	Phase               UpgradeCoveragePhaseName        `json:"phase"`
	OwnerAddress        string                          `json:"owner_address"`
	TopicName           string                          `json:"topic_name"`
	WriterAddress       string                          `json:"writer_address"`
	InitialRecordOffset uint64                          `json:"initial_record_offset"`
	StateObjectIDs      []string                        `json:"state_object_ids"`
	Transactions        []AOLUpgradeTransactionEvidence `json:"transactions"`
}

// AOLUpgradeOwnerState is the owner-address and topic-index projection.
type AOLUpgradeOwnerState struct {
	Address    string   `json:"address"`
	TopicNames []string `json:"topic_names"`
}

// AOLUpgradeTopicState retains the topic value and append-only counters.
type AOLUpgradeTopicState struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	TotalRecords uint64 `json:"total_records"`
	TotalWriters uint64 `json:"total_writers"`
}

// AOLUpgradeWriterState retains one address-keyed writer value.
type AOLUpgradeWriterState struct {
	Address       string `json:"address"`
	Moniker       string `json:"moniker"`
	Description   string `json:"description"`
	NanoTimestamp int64  `json:"nano_timestamp"`
}

// AOLUpgradeRecordState makes the store offset explicit next to its value.
type AOLUpgradeRecordState struct {
	Offset        uint64 `json:"offset"`
	Key           []byte `json:"key"`
	Value         []byte `json:"value"`
	NanoTimestamp int64  `json:"nano_timestamp"`
	WriterAddress string `json:"writer_address"`
}

// AOLUpgradeCheckpoint is a semantic, height-pinned projection of the owner
// index, topic counters, every current writer, and every record offset.
type AOLUpgradeCheckpoint struct {
	SchemaVersion string                       `json:"schema_version"`
	RecordedAt    time.Time                    `json:"recorded_at"`
	Phase         UpgradeCoveragePhaseName     `json:"phase"`
	QueryHeight   int64                        `json:"query_height"`
	Observation   UpgradeCheckpointObservation `json:"observation"`
	Owner         AOLUpgradeOwnerState         `json:"owner"`
	Topic         AOLUpgradeTopicState         `json:"topic"`
	Writers       []AOLUpgradeWriterState      `json:"writers"`
	Records       []AOLUpgradeRecordState      `json:"records"`
}

func (c AOLUpgradeCheckpoint) Validate() error {
	if c.SchemaVersion != AOLUpgradeEvidenceSchemaVersion {
		return fmt.Errorf("AOL checkpoint schema version %q, want %q", c.SchemaVersion, AOLUpgradeEvidenceSchemaVersion)
	}
	if c.RecordedAt.IsZero() {
		return errors.New("AOL checkpoint recorded_at is required")
	}
	if !isAOLCheckpointPhase(c.Phase) {
		return fmt.Errorf("AOL checkpoint phase %q is invalid", c.Phase)
	}
	if c.QueryHeight <= 0 {
		return fmt.Errorf("AOL checkpoint query_height must be positive, got %d", c.QueryHeight)
	}
	if err := c.Observation.Validate(); err != nil {
		return fmt.Errorf("AOL checkpoint observation: %w", err)
	}
	if c.Observation.Height != c.QueryHeight || !c.Observation.ObservedAt.Equal(c.RecordedAt) {
		return errors.New("AOL checkpoint observation height/time does not match checkpoint")
	}
	if strings.TrimSpace(c.Owner.Address) == "" || strings.TrimSpace(c.Topic.Name) == "" {
		return errors.New("AOL checkpoint owner address and topic name are required")
	}
	if err := validateSortedUniqueStrings(c.Owner.TopicNames, "AOL checkpoint owner topic_names"); err != nil {
		return err
	}
	if !containsString(c.Owner.TopicNames, c.Topic.Name) {
		return fmt.Errorf("AOL checkpoint owner topics do not contain %q", c.Topic.Name)
	}
	if uint64(len(c.Writers)) != c.Topic.TotalWriters {
		return fmt.Errorf("AOL checkpoint has %d writers, topic reports %d", len(c.Writers), c.Topic.TotalWriters)
	}
	for index, writer := range c.Writers {
		if strings.TrimSpace(writer.Address) == "" || writer.NanoTimestamp <= 0 {
			return fmt.Errorf("AOL checkpoint writer %d is incomplete", index)
		}
		if index > 0 && c.Writers[index-1].Address >= writer.Address {
			return errors.New("AOL checkpoint writers must be uniquely sorted by address")
		}
	}
	if uint64(len(c.Records)) != c.Topic.TotalRecords {
		return fmt.Errorf("AOL checkpoint has %d records, topic reports %d", len(c.Records), c.Topic.TotalRecords)
	}
	for index, record := range c.Records {
		if record.Offset != uint64(index) {
			return fmt.Errorf("AOL checkpoint record %d has offset %d", index, record.Offset)
		}
		if len(record.Key) == 0 || len(record.Value) == 0 || record.NanoTimestamp <= 0 || strings.TrimSpace(record.WriterAddress) == "" {
			return fmt.Errorf("AOL checkpoint record offset %d is incomplete", record.Offset)
		}
	}
	return nil
}

func isAOLCheckpointPhase(phase UpgradeCoveragePhaseName) bool {
	switch phase {
	case UpgradeCoveragePhasePreUpgradeCheckpoint,
		UpgradeCoveragePhasePostUpgradePreservation,
		UpgradeCoveragePhasePostUpgradeMutation,
		UpgradeCoveragePhasePostRestart:
		return true
	default:
		return false
	}
}

func validateSortedUniqueStrings(values []string, field string) error {
	for index, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s contains an empty value", field)
		}
		if index > 0 && values[index-1] >= value {
			return fmt.Errorf("%s must be uniquely sorted", field)
		}
	}
	return nil
}

// AOLUpgradeMutationEvidence proves that one record was appended at the next
// offset and that a second writer could be added and removed after upgrade.
type AOLUpgradeMutationEvidence struct {
	SchemaVersion        string                          `json:"schema_version"`
	RecordedAt           time.Time                       `json:"recorded_at"`
	Phase                UpgradeCoveragePhaseName        `json:"phase"`
	MutationRecordOffset uint64                          `json:"mutation_record_offset"`
	Transactions         []AOLUpgradeTransactionEvidence `json:"transactions"`
	Before               AOLUpgradeCheckpoint            `json:"before"`
	AfterWriterAdd       AOLUpgradeCheckpoint            `json:"after_writer_add"`
	Final                AOLUpgradeCheckpoint            `json:"final"`
}

type aolUpgradeRuntime struct {
	broadcast          func(context.Context, string, *cosmos.ChainNode, string, ...string) (*TxResult, error)
	query              func(context.Context, string, ...string) (json.RawMessage, error)
	queryREST          func(context.Context, string, string, int64) (json.RawMessage, error)
	fullNodeHeight     func(context.Context) (int64, error)
	captureObservation func(context.Context, string, int64) (UpgradeCheckpointObservation, error)
	waitFullNodeHeight func(context.Context, int64) error
	writeJSON          func(string, any) error
}

func (n *Network) aolUpgradeRuntime() (aolUpgradeRuntime, error) {
	if n == nil || n.Chain == nil {
		return aolUpgradeRuntime{}, errors.New("AOL upgrade network is required")
	}
	if n.artifacts == nil {
		return aolUpgradeRuntime{}, errors.New("AOL upgrade artifact store is required")
	}
	if len(n.Chain.FullNodes) == 0 {
		return aolUpgradeRuntime{}, errors.New("AOL upgrade network has no full node")
	}
	fullNode := n.Chain.FullNodes[0]
	return aolUpgradeRuntime{
		broadcast: n.BroadcastAndWaitTx,
		query:     n.FullNodeCLIQuery,
		queryREST: func(ctx context.Context, step, path string, height int64) (json.RawMessage, error) {
			return n.FullNodeRESTGetAtHeight(ctx, nil, step, path, height)
		},
		fullNodeHeight: fullNode.Height,
		captureObservation: func(ctx context.Context, step string, height int64) (UpgradeCheckpointObservation, error) {
			return n.CaptureUpgradeCheckpointObservation(ctx, step, fullNode, height)
		},
		waitFullNodeHeight: func(ctx context.Context, height int64) error {
			return n.WaitForNodeHeight(ctx, fullNode, height)
		},
		writeJSON: n.WriteArtifactJSON,
	}, nil
}

// PrepareAOLUpgradeFixture creates the topic, persistent writer, and offset-0
// record on the source image and writes the preparation phase artifact.
func (n *Network) PrepareAOLUpgradeFixture(
	ctx context.Context,
	node *cosmos.ChainNode,
	fixture AOLUpgradeFixture,
) (AOLUpgradePreparationEvidence, error) {
	runtime, err := n.aolUpgradeRuntime()
	if err != nil {
		return AOLUpgradePreparationEvidence{}, err
	}
	return prepareAOLUpgradeFixture(ctx, runtime, node, fixture)
}

func prepareAOLUpgradeFixture(
	ctx context.Context,
	runtime aolUpgradeRuntime,
	node *cosmos.ChainNode,
	fixture AOLUpgradeFixture,
) (AOLUpgradePreparationEvidence, error) {
	if err := fixture.Validate(); err != nil {
		return AOLUpgradePreparationEvidence{}, err
	}
	if node == nil {
		return AOLUpgradePreparationEvidence{}, errors.New("AOL upgrade transaction node is required")
	}

	requests := []struct {
		step    string
		keyName string
		command []string
	}{
		{
			step:    "upgrade-aol-v221-create-topic",
			keyName: fixture.OwnerKeyName,
			command: []string{"aol", "create-topic", fixture.TopicName, "--description", fixture.TopicDescription, "--gas", aolUpgradeTxGas, "--broadcast-mode", "sync"},
		},
		{
			step:    "upgrade-aol-v221-add-writer",
			keyName: fixture.OwnerKeyName,
			command: []string{"aol", "add-writer", fixture.TopicName, fixture.WriterAddress, "--moniker", fixture.WriterMoniker, "--description", fixture.WriterDescription, "--gas", aolUpgradeTxGas, "--broadcast-mode", "sync"},
		},
		{
			step:    "upgrade-aol-v221-add-record",
			keyName: fixture.WriterKeyName,
			command: []string{"aol", "add-record", fixture.OwnerAddress, fixture.TopicName, fixture.InitialRecord.Key, fixture.InitialRecord.Value, "--gas", aolUpgradeTxGas, "--broadcast-mode", "sync"},
		},
	}

	transactions := make([]AOLUpgradeTransactionEvidence, 0, len(requests))
	var lastHeight int64
	for _, request := range requests {
		result, err := runtime.broadcast(ctx, request.step, node, request.keyName, request.command...)
		if err != nil {
			return AOLUpgradePreparationEvidence{}, fmt.Errorf("%s: %w", request.step, err)
		}
		tx, err := newAOLUpgradeTransactionEvidence(request.step, result)
		if err != nil {
			return AOLUpgradePreparationEvidence{}, err
		}
		transactions = append(transactions, tx)
		lastHeight = tx.Height
	}
	if err := runtime.waitFullNodeHeight(ctx, lastHeight); err != nil {
		return AOLUpgradePreparationEvidence{}, fmt.Errorf("wait for AOL preparation on full node: %w", err)
	}

	evidence := AOLUpgradePreparationEvidence{
		SchemaVersion:       AOLUpgradeEvidenceSchemaVersion,
		RecordedAt:          time.Now().UTC(),
		Phase:               UpgradeCoveragePhaseV221Preparation,
		OwnerAddress:        fixture.OwnerAddress,
		TopicName:           fixture.TopicName,
		WriterAddress:       fixture.WriterAddress,
		InitialRecordOffset: 0,
		StateObjectIDs:      fixture.StateObjectIDs(),
		Transactions:        transactions,
	}
	if err := runtime.writeJSON(AOLUpgradePreparationArtifactPath, evidence); err != nil {
		return evidence, fmt.Errorf("record AOL preparation artifact: %w", err)
	}
	return evidence, nil
}

func newAOLUpgradeTransactionEvidence(step string, result *TxResult) (AOLUpgradeTransactionEvidence, error) {
	if result == nil {
		return AOLUpgradeTransactionEvidence{}, fmt.Errorf("AOL transaction %s returned no result", step)
	}
	height := result.HeightInt64()
	if height <= 0 || strings.TrimSpace(result.TxHash) == "" {
		return AOLUpgradeTransactionEvidence{}, fmt.Errorf("AOL transaction %s returned incomplete commit evidence", step)
	}
	return AOLUpgradeTransactionEvidence{Step: step, TxHash: result.TxHash, Height: height}, nil
}

// CaptureAOLUpgradePreUpgradeCheckpoint records the source-version semantic
// checkpoint used as the preservation oracle after the binary upgrade.
func (n *Network) CaptureAOLUpgradePreUpgradeCheckpoint(
	ctx context.Context,
	fixture AOLUpgradeFixture,
) (AOLUpgradeCheckpoint, error) {
	runtime, err := n.aolUpgradeRuntime()
	if err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	checkpoint, err := captureAOLUpgradeCheckpoint(ctx, runtime, UpgradeCoveragePhasePreUpgradeCheckpoint, fixture)
	if err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	if err := runtime.writeJSON(AOLUpgradePreCheckpointArtifactPath, checkpoint); err != nil {
		return checkpoint, fmt.Errorf("record AOL pre-upgrade checkpoint: %w", err)
	}
	return checkpoint, nil
}

// CaptureAndAssertAOLUpgradePreserved queries the upgraded binary, records the
// actual result even on a semantic mismatch, and compares it with the source
// checkpoint while ignoring observation time, phase, and query height.
func (n *Network) CaptureAndAssertAOLUpgradePreserved(
	ctx context.Context,
	fixture AOLUpgradeFixture,
	want AOLUpgradeCheckpoint,
) (AOLUpgradeCheckpoint, error) {
	return n.captureAndAssertAOLUpgradeCheckpoint(
		ctx,
		fixture,
		want,
		UpgradeCoveragePhasePostUpgradePreservation,
		AOLUpgradePostPreservationArtifactPath,
	)
}

// CaptureAndAssertAOLAfterRestart proves that the final post-mutation state is
// unchanged after all nodes restart and writes the fifth-phase artifact.
func (n *Network) CaptureAndAssertAOLAfterRestart(
	ctx context.Context,
	fixture AOLUpgradeFixture,
	want AOLUpgradeCheckpoint,
) (AOLUpgradeCheckpoint, error) {
	return n.captureAndAssertAOLUpgradeCheckpoint(
		ctx,
		fixture,
		want,
		UpgradeCoveragePhasePostRestart,
		AOLUpgradePostRestartArtifactPath,
	)
}

func (n *Network) captureAndAssertAOLUpgradeCheckpoint(
	ctx context.Context,
	fixture AOLUpgradeFixture,
	want AOLUpgradeCheckpoint,
	phase UpgradeCoveragePhaseName,
	artifactPath string,
) (AOLUpgradeCheckpoint, error) {
	runtime, err := n.aolUpgradeRuntime()
	if err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	got, err := captureAOLUpgradeCheckpoint(ctx, runtime, phase, fixture)
	if err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	recordErr := runtime.writeJSON(artifactPath, got)
	compareErr := AssertAOLUpgradeCheckpointEqual(want, got)
	if combined := errors.Join(compareErr, recordErr); combined != nil {
		return got, fmt.Errorf("verify AOL checkpoint phase %s: %w", phase, combined)
	}
	return got, nil
}

// AssertAOLUpgradeCheckpointEqual compares durable AOL state, including every
// record offset and timestamp, without requiring observations at equal heights.
func AssertAOLUpgradeCheckpointEqual(want, got AOLUpgradeCheckpoint) error {
	if err := want.Validate(); err != nil {
		return fmt.Errorf("validate expected AOL checkpoint: %w", err)
	}
	if err := got.Validate(); err != nil {
		return fmt.Errorf("validate actual AOL checkpoint: %w", err)
	}
	wantState := aolUpgradeSemanticState(want)
	gotState := aolUpgradeSemanticState(got)
	if reflect.DeepEqual(wantState, gotState) {
		return nil
	}
	wantJSON, _ := json.Marshal(wantState)
	gotJSON, _ := json.Marshal(gotState)
	return fmt.Errorf("AOL semantic state changed: want=%s got=%s", wantJSON, gotJSON)
}

type aolUpgradeState struct {
	Owner   AOLUpgradeOwnerState    `json:"owner"`
	Topic   AOLUpgradeTopicState    `json:"topic"`
	Writers []AOLUpgradeWriterState `json:"writers"`
	Records []AOLUpgradeRecordState `json:"records"`
}

func aolUpgradeSemanticState(checkpoint AOLUpgradeCheckpoint) aolUpgradeState {
	return aolUpgradeState{
		Owner:   checkpoint.Owner,
		Topic:   checkpoint.Topic,
		Writers: checkpoint.Writers,
		Records: checkpoint.Records,
	}
}

// MutateAOLAfterUpgrade appends one record at the next offset, proves a new
// writer is queryable, deletes that writer, and returns the state oracle for
// the later restart assertion.
func (n *Network) MutateAOLAfterUpgrade(
	ctx context.Context,
	node *cosmos.ChainNode,
	fixture AOLUpgradeFixture,
	before AOLUpgradeCheckpoint,
) (AOLUpgradeMutationEvidence, error) {
	runtime, err := n.aolUpgradeRuntime()
	if err != nil {
		return AOLUpgradeMutationEvidence{}, err
	}
	return mutateAOLAfterUpgrade(ctx, runtime, node, fixture, before)
}

func mutateAOLAfterUpgrade(
	ctx context.Context,
	runtime aolUpgradeRuntime,
	node *cosmos.ChainNode,
	fixture AOLUpgradeFixture,
	before AOLUpgradeCheckpoint,
) (AOLUpgradeMutationEvidence, error) {
	if err := fixture.Validate(); err != nil {
		return AOLUpgradeMutationEvidence{}, err
	}
	if err := before.Validate(); err != nil {
		return AOLUpgradeMutationEvidence{}, fmt.Errorf("validate AOL pre-mutation checkpoint: %w", err)
	}
	if node == nil {
		return AOLUpgradeMutationEvidence{}, errors.New("AOL upgrade transaction node is required")
	}

	transactions := make([]AOLUpgradeTransactionEvidence, 0, 3)
	run := func(step, keyName string, command ...string) (AOLUpgradeTransactionEvidence, error) {
		result, err := runtime.broadcast(ctx, step, node, keyName, command...)
		if err != nil {
			return AOLUpgradeTransactionEvidence{}, fmt.Errorf("%s: %w", step, err)
		}
		return newAOLUpgradeTransactionEvidence(step, result)
	}

	addRecord, err := run(
		"upgrade-aol-post-add-record",
		fixture.WriterKeyName,
		"aol", "add-record", fixture.OwnerAddress, fixture.TopicName,
		fixture.MutationRecord.Key, fixture.MutationRecord.Value,
		"--gas", aolUpgradeTxGas, "--broadcast-mode", "sync",
	)
	if err != nil {
		return AOLUpgradeMutationEvidence{}, err
	}
	transactions = append(transactions, addRecord)
	addWriter, err := run(
		"upgrade-aol-post-add-writer",
		fixture.OwnerKeyName,
		"aol", "add-writer", fixture.TopicName, fixture.MutationWriterAddress,
		"--moniker", fixture.MutationWriterMoniker,
		"--description", fixture.MutationWriterDescription,
		"--gas", aolUpgradeTxGas, "--broadcast-mode", "sync",
	)
	if err != nil {
		return AOLUpgradeMutationEvidence{}, err
	}
	transactions = append(transactions, addWriter)
	if err := runtime.waitFullNodeHeight(ctx, addWriter.Height); err != nil {
		return AOLUpgradeMutationEvidence{}, fmt.Errorf("wait for AOL add mutations on full node: %w", err)
	}
	afterAdd, err := captureAOLUpgradeCheckpoint(ctx, runtime, UpgradeCoveragePhasePostUpgradeMutation, fixture)
	if err != nil {
		return AOLUpgradeMutationEvidence{}, err
	}
	if err := validateAOLUpgradeAfterAdd(fixture, before, afterAdd); err != nil {
		return AOLUpgradeMutationEvidence{}, err
	}

	deleteWriter, err := run(
		"upgrade-aol-post-delete-writer",
		fixture.OwnerKeyName,
		"aol", "delete-writer", fixture.TopicName, fixture.MutationWriterAddress,
		"--gas", aolUpgradeTxGas, "--broadcast-mode", "sync",
	)
	if err != nil {
		return AOLUpgradeMutationEvidence{}, err
	}
	transactions = append(transactions, deleteWriter)
	if err := runtime.waitFullNodeHeight(ctx, deleteWriter.Height); err != nil {
		return AOLUpgradeMutationEvidence{}, fmt.Errorf("wait for AOL writer deletion on full node: %w", err)
	}
	finalCheckpoint, err := captureAOLUpgradeCheckpoint(ctx, runtime, UpgradeCoveragePhasePostUpgradeMutation, fixture)
	if err != nil {
		return AOLUpgradeMutationEvidence{}, err
	}
	transitionErr := validateAOLUpgradeAfterDelete(before, afterAdd, finalCheckpoint)

	evidence := AOLUpgradeMutationEvidence{
		SchemaVersion:        AOLUpgradeEvidenceSchemaVersion,
		RecordedAt:           time.Now().UTC(),
		Phase:                UpgradeCoveragePhasePostUpgradeMutation,
		MutationRecordOffset: before.Topic.TotalRecords,
		Transactions:         transactions,
		Before:               before,
		AfterWriterAdd:       afterAdd,
		Final:                finalCheckpoint,
	}
	recordErr := runtime.writeJSON(AOLUpgradePostMutationArtifactPath, evidence)
	if combined := errors.Join(transitionErr, recordErr); combined != nil {
		return evidence, fmt.Errorf("verify AOL post-upgrade mutation: %w", combined)
	}
	return evidence, nil
}

func validateAOLUpgradeAfterAdd(
	fixture AOLUpgradeFixture,
	before AOLUpgradeCheckpoint,
	after AOLUpgradeCheckpoint,
) error {
	if err := assertAOLUpgradeIdentityAndPrefix(before, after); err != nil {
		return fmt.Errorf("AOL add mutation did not preserve existing state: %w", err)
	}
	if after.Topic.TotalRecords != before.Topic.TotalRecords+1 {
		return fmt.Errorf("AOL record count after add is %d, want %d", after.Topic.TotalRecords, before.Topic.TotalRecords+1)
	}
	if after.Topic.TotalWriters != before.Topic.TotalWriters+1 {
		return fmt.Errorf("AOL writer count after add is %d, want %d", after.Topic.TotalWriters, before.Topic.TotalWriters+1)
	}
	writer, ok := findAOLUpgradeWriter(after.Writers, fixture.MutationWriterAddress)
	if !ok {
		return fmt.Errorf("AOL mutation writer %s is absent after add", fixture.MutationWriterAddress)
	}
	if writer.Moniker != fixture.MutationWriterMoniker || writer.Description != fixture.MutationWriterDescription {
		return fmt.Errorf("AOL mutation writer metadata changed: %+v", writer)
	}
	offset := before.Topic.TotalRecords
	record := after.Records[offset]
	if record.Offset != offset ||
		!reflect.DeepEqual(record.Key, []byte(fixture.MutationRecord.Key)) ||
		!reflect.DeepEqual(record.Value, []byte(fixture.MutationRecord.Value)) ||
		record.WriterAddress != fixture.WriterAddress {
		return fmt.Errorf("AOL mutation record at offset %d is unexpected: %+v", offset, record)
	}
	return nil
}

func assertAOLUpgradeIdentityAndPrefix(before, after AOLUpgradeCheckpoint) error {
	if err := before.Validate(); err != nil {
		return err
	}
	if err := after.Validate(); err != nil {
		return err
	}
	if !reflect.DeepEqual(before.Owner, after.Owner) {
		return errors.New("owner topic index changed")
	}
	if before.Topic.Name != after.Topic.Name || before.Topic.Description != after.Topic.Description {
		return errors.New("topic identity or description changed")
	}
	for _, wantWriter := range before.Writers {
		gotWriter, ok := findAOLUpgradeWriter(after.Writers, wantWriter.Address)
		if !ok || !reflect.DeepEqual(wantWriter, gotWriter) {
			return fmt.Errorf("writer %s changed", wantWriter.Address)
		}
	}
	if len(after.Records) < len(before.Records) || !reflect.DeepEqual(before.Records, after.Records[:len(before.Records)]) {
		return errors.New("existing record offsets changed")
	}
	return nil
}

func validateAOLUpgradeAfterDelete(before, afterAdd, final AOLUpgradeCheckpoint) error {
	expected := AOLUpgradeCheckpoint{
		SchemaVersion: AOLUpgradeEvidenceSchemaVersion,
		RecordedAt:    final.RecordedAt,
		Phase:         UpgradeCoveragePhasePostUpgradeMutation,
		QueryHeight:   final.QueryHeight,
		Observation:   final.Observation,
		Owner:         before.Owner,
		Topic: AOLUpgradeTopicState{
			Name:         before.Topic.Name,
			Description:  before.Topic.Description,
			TotalRecords: afterAdd.Topic.TotalRecords,
			TotalWriters: before.Topic.TotalWriters,
		},
		Writers: append([]AOLUpgradeWriterState(nil), before.Writers...),
		Records: cloneAOLUpgradeRecords(afterAdd.Records),
	}
	return AssertAOLUpgradeCheckpointEqual(expected, final)
}

func findAOLUpgradeWriter(writers []AOLUpgradeWriterState, address string) (AOLUpgradeWriterState, bool) {
	index := sort.Search(len(writers), func(index int) bool { return writers[index].Address >= address })
	if index == len(writers) || writers[index].Address != address {
		return AOLUpgradeWriterState{}, false
	}
	return writers[index], true
}

func cloneAOLUpgradeRecords(records []AOLUpgradeRecordState) []AOLUpgradeRecordState {
	clone := make([]AOLUpgradeRecordState, len(records))
	for index, record := range records {
		clone[index] = record
		clone[index].Key = append([]byte(nil), record.Key...)
		clone[index].Value = append([]byte(nil), record.Value...)
	}
	return clone
}

func captureAOLUpgradeCheckpoint(
	ctx context.Context,
	runtime aolUpgradeRuntime,
	phase UpgradeCoveragePhaseName,
	fixture AOLUpgradeFixture,
) (AOLUpgradeCheckpoint, error) {
	if err := fixture.Validate(); err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	if !isAOLCheckpointPhase(phase) {
		return AOLUpgradeCheckpoint{}, fmt.Errorf("cannot capture AOL checkpoint for phase %q", phase)
	}
	height, err := runtime.fullNodeHeight(ctx)
	if err != nil {
		return AOLUpgradeCheckpoint{}, fmt.Errorf("read AOL checkpoint full-node height: %w", err)
	}
	if height <= 0 {
		return AOLUpgradeCheckpoint{}, fmt.Errorf("AOL checkpoint full-node height must be positive, got %d", height)
	}
	if runtime.captureObservation == nil {
		return AOLUpgradeCheckpoint{}, errors.New("AOL checkpoint block observation is required")
	}
	observation, err := runtime.captureObservation(ctx, "upgrade-aol-"+string(phase), height)
	if err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	query := func(label string, command ...string) (json.RawMessage, error) {
		command = append(command, "--height", strconv.FormatInt(height, 10))
		step := fmt.Sprintf("upgrade-aol-%s-%s", strings.ReplaceAll(string(phase), ".", ""), label)
		return runtime.query(ctx, step, command...)
	}

	topicRaw, err := query("topic", "aol", "get-topic", fixture.OwnerAddress, fixture.TopicName)
	if err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	topic, err := decodeAOLUpgradeTopic(topicRaw, fixture.TopicName)
	if err != nil {
		return AOLUpgradeCheckpoint{}, fmt.Errorf("decode AOL topic checkpoint: %w", err)
	}
	if runtime.queryREST == nil {
		return AOLUpgradeCheckpoint{}, errors.New("AOL checkpoint bounded REST query is required")
	}
	pageQuery := url.Values{
		"pagination.count_total": []string{"true"},
		"pagination.limit":       []string{strconv.Itoa(aolUpgradeQueryLimit)},
	}.Encode()
	topicsBasePath := "/panacea/aol/v2/owners/" + url.PathEscape(fixture.OwnerAddress) + "/topics"
	topicsPath := topicsBasePath + "?" + pageQuery
	topicsRaw, err := runtime.queryREST(
		ctx,
		fmt.Sprintf("upgrade-aol-%s-owner-topics", strings.ReplaceAll(string(phase), ".", "")),
		topicsPath,
		height,
	)
	if err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	topicNames, err := decodeCompleteAOLUpgradeTopicsPage(topicsRaw)
	if err != nil {
		return AOLUpgradeCheckpoint{}, fmt.Errorf("decode AOL owner topic index: %w", err)
	}
	if len(topicNames) != 1 || topicNames[0] != fixture.TopicName {
		return AOLUpgradeCheckpoint{}, fmt.Errorf(
			"dedicated AOL fixture owner topics=%v, want only %q",
			topicNames,
			fixture.TopicName,
		)
	}

	writersPath := topicsBasePath + "/" + url.PathEscape(fixture.TopicName) + "/writers?" + pageQuery
	writersRaw, err := runtime.queryREST(
		ctx,
		fmt.Sprintf("upgrade-aol-%s-writers", strings.ReplaceAll(string(phase), ".", "")),
		writersPath,
		height,
	)
	if err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	writerAddresses, err := decodeCompleteAOLUpgradeWritersPage(writersRaw)
	if err != nil {
		return AOLUpgradeCheckpoint{}, fmt.Errorf("decode AOL writer index: %w", err)
	}
	if uint64(len(writerAddresses)) != topic.TotalWriters {
		return AOLUpgradeCheckpoint{}, fmt.Errorf("AOL writer query returned %d entries, topic reports %d", len(writerAddresses), topic.TotalWriters)
	}
	if !containsString(writerAddresses, fixture.WriterAddress) {
		return AOLUpgradeCheckpoint{}, fmt.Errorf("AOL writer index does not contain persistent writer %s", fixture.WriterAddress)
	}
	writers := make([]AOLUpgradeWriterState, 0, len(writerAddresses))
	for index, address := range writerAddresses {
		raw, err := query(
			fmt.Sprintf("writer-%d", index),
			"aol", "get-writer", fixture.OwnerAddress, fixture.TopicName, address,
		)
		if err != nil {
			return AOLUpgradeCheckpoint{}, err
		}
		writer, err := decodeAOLUpgradeWriter(raw, address)
		if err != nil {
			return AOLUpgradeCheckpoint{}, fmt.Errorf("decode AOL writer %s: %w", address, err)
		}
		writers = append(writers, writer)
	}

	records := make([]AOLUpgradeRecordState, 0, topic.TotalRecords)
	for offset := uint64(0); offset < topic.TotalRecords; offset++ {
		raw, err := query(
			fmt.Sprintf("record-%d", offset),
			"aol", "get-record", fixture.OwnerAddress, fixture.TopicName, strconv.FormatUint(offset, 10),
		)
		if err != nil {
			return AOLUpgradeCheckpoint{}, err
		}
		record, err := decodeAOLUpgradeRecord(raw, offset)
		if err != nil {
			return AOLUpgradeCheckpoint{}, fmt.Errorf("decode AOL record offset %d: %w", offset, err)
		}
		records = append(records, record)
	}

	checkpoint := AOLUpgradeCheckpoint{
		SchemaVersion: AOLUpgradeEvidenceSchemaVersion,
		RecordedAt:    observation.ObservedAt,
		Phase:         phase,
		QueryHeight:   height,
		Observation:   observation,
		Owner: AOLUpgradeOwnerState{
			Address:    fixture.OwnerAddress,
			TopicNames: topicNames,
		},
		Topic:   topic,
		Writers: writers,
		Records: records,
	}
	if err := checkpoint.Validate(); err != nil {
		return AOLUpgradeCheckpoint{}, err
	}
	return checkpoint, nil
}

type aolJSONUint64 uint64

func (value *aolJSONUint64) UnmarshalJSON(data []byte) error {
	parsed, err := parseJSONUint64(data)
	if err != nil {
		return err
	}
	*value = aolJSONUint64(parsed)
	return nil
}

type aolJSONInt64 int64

func (value *aolJSONInt64) UnmarshalJSON(data []byte) error {
	data = []byte(strings.Trim(string(data), `"`))
	parsed, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("parse JSON int64 %q: %w", data, err)
	}
	*value = aolJSONInt64(parsed)
	return nil
}

func parseJSONUint64(data []byte) (uint64, error) {
	data = []byte(strings.Trim(string(data), `"`))
	parsed, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse JSON uint64 %q: %w", data, err)
	}
	return parsed, nil
}

func decodeAOLUpgradeTopic(raw json.RawMessage, topicName string) (AOLUpgradeTopicState, error) {
	var response struct {
		Topic *struct {
			Description  string        `json:"description"`
			TotalRecords aolJSONUint64 `json:"total_records"`
			TotalWriters aolJSONUint64 `json:"total_writers"`
		} `json:"topic"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return AOLUpgradeTopicState{}, err
	}
	if response.Topic == nil {
		return AOLUpgradeTopicState{}, errors.New("topic response is missing topic")
	}
	return AOLUpgradeTopicState{
		Name:         topicName,
		Description:  response.Topic.Description,
		TotalRecords: uint64(response.Topic.TotalRecords),
		TotalWriters: uint64(response.Topic.TotalWriters),
	}, nil
}

func decodeAOLUpgradeTopics(raw json.RawMessage) ([]string, error) {
	var response struct {
		TopicNames []string `json:"topic_names"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}
	return sortedUniqueAOLValues(response.TopicNames, "topic_names")
}

type aolUpgradePageResponse struct {
	NextKey json.RawMessage `json:"next_key"`
	Total   aolJSONUint64   `json:"total"`
}

func decodeCompleteAOLUpgradeTopicsPage(raw json.RawMessage) ([]string, error) {
	var response struct {
		TopicNames []string                `json:"topic_names"`
		Pagination *aolUpgradePageResponse `json:"pagination"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}
	values, err := sortedUniqueAOLValues(response.TopicNames, "topic_names")
	if err != nil {
		return nil, err
	}
	if err := validateCompleteAOLUpgradePage(response.Pagination, len(values), "topics"); err != nil {
		return nil, err
	}
	return values, nil
}

func decodeAOLUpgradeWriterAddresses(raw json.RawMessage) ([]string, error) {
	var response struct {
		WriterAddresses []string `json:"writer_addresses"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}
	return sortedUniqueAOLValues(response.WriterAddresses, "writer_addresses")
}

func decodeCompleteAOLUpgradeWritersPage(raw json.RawMessage) ([]string, error) {
	var response struct {
		WriterAddresses []string                `json:"writer_addresses"`
		Pagination      *aolUpgradePageResponse `json:"pagination"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}
	values, err := sortedUniqueAOLValues(response.WriterAddresses, "writer_addresses")
	if err != nil {
		return nil, err
	}
	if err := validateCompleteAOLUpgradePage(response.Pagination, len(values), "writers"); err != nil {
		return nil, err
	}
	return values, nil
}

func validateCompleteAOLUpgradePage(page *aolUpgradePageResponse, count int, label string) error {
	if page == nil {
		return fmt.Errorf("bounded AOL %s response is missing pagination", label)
	}
	nextKey := strings.TrimSpace(string(page.NextKey))
	if nextKey != "null" && nextKey != `""` {
		return fmt.Errorf("bounded AOL %s page is incomplete: next_key=%s", label, nextKey)
	}
	if uint64(page.Total) != uint64(count) {
		return fmt.Errorf("bounded AOL %s page total=%d, returned=%d", label, page.Total, count)
	}
	if count > aolUpgradeQueryLimit {
		return fmt.Errorf("bounded AOL %s page returned %d entries above limit %d", label, count, aolUpgradeQueryLimit)
	}
	return nil
}

func sortedUniqueAOLValues(values []string, field string) ([]string, error) {
	values = append([]string(nil), values...)
	sort.Strings(values)
	if err := validateSortedUniqueStrings(values, "AOL query "+field); err != nil {
		return nil, err
	}
	return values, nil
}

func decodeAOLUpgradeWriter(raw json.RawMessage, address string) (AOLUpgradeWriterState, error) {
	var response struct {
		Writer *struct {
			Moniker       string       `json:"moniker"`
			Description   string       `json:"description"`
			NanoTimestamp aolJSONInt64 `json:"nano_timestamp"`
		} `json:"writer"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return AOLUpgradeWriterState{}, err
	}
	if response.Writer == nil {
		return AOLUpgradeWriterState{}, errors.New("writer response is missing writer")
	}
	return AOLUpgradeWriterState{
		Address:       address,
		Moniker:       response.Writer.Moniker,
		Description:   response.Writer.Description,
		NanoTimestamp: int64(response.Writer.NanoTimestamp),
	}, nil
}

func decodeAOLUpgradeRecord(raw json.RawMessage, offset uint64) (AOLUpgradeRecordState, error) {
	var response struct {
		Record *struct {
			Key           []byte       `json:"key"`
			Value         []byte       `json:"value"`
			NanoTimestamp aolJSONInt64 `json:"nano_timestamp"`
			WriterAddress string       `json:"writer_address"`
		} `json:"record"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return AOLUpgradeRecordState{}, err
	}
	if response.Record == nil {
		return AOLUpgradeRecordState{}, errors.New("record response is missing record")
	}
	return AOLUpgradeRecordState{
		Offset:        offset,
		Key:           append([]byte(nil), response.Record.Key...),
		Value:         append([]byte(nil), response.Record.Value...),
		NanoTimestamp: int64(response.Record.NanoTimestamp),
		WriterAddress: response.Record.WriterAddress,
	}, nil
}

func containsString(values []string, want string) bool {
	index := sort.SearchStrings(values, want)
	return index < len(values) && values[index] == want
}
