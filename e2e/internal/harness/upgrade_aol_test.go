package harness

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func TestAOLUpgradeFixtureValidate(t *testing.T) {
	fixture := testAOLUpgradeFixture()
	require.NoError(t, fixture.Validate())
	require.Equal(t, []string{
		"aol-owner:panacea1owner",
		"aol-topic:panacea1owner/upgrade-aol",
		"aol-writer:panacea1owner/upgrade-aol/panacea1writer",
		"aol-writer:panacea1owner/upgrade-aol/panacea1mutation",
		"aol-record:panacea1owner/upgrade-aol/0",
		"aol-record:panacea1owner/upgrade-aol/1",
	}, fixture.StateObjectIDs())

	tests := []struct {
		name   string
		mutate func(*AOLUpgradeFixture)
		want   string
	}{
		{
			name: "missing owner",
			mutate: func(fixture *AOLUpgradeFixture) {
				fixture.OwnerAddress = ""
			},
			want: "owner_address is required",
		},
		{
			name: "identities not distinct",
			mutate: func(fixture *AOLUpgradeFixture) {
				fixture.MutationWriterAddress = fixture.WriterAddress
			},
			want: "must be distinct",
		},
		{
			name: "invalid topic",
			mutate: func(fixture *AOLUpgradeFixture) {
				fixture.TopicName = "contains spaces"
			},
			want: "topic_name",
		},
		{
			name: "invalid writer moniker",
			mutate: func(fixture *AOLUpgradeFixture) {
				fixture.WriterMoniker = "not allowed"
			},
			want: "writer_moniker",
		},
		{
			name: "empty initial record",
			mutate: func(fixture *AOLUpgradeFixture) {
				fixture.InitialRecord.Key = ""
			},
			want: "initial_record key length",
		},
		{
			name: "oversized mutation value",
			mutate: func(fixture *AOLUpgradeFixture) {
				fixture.MutationRecord.Value = string(make([]byte, 5001))
			},
			want: "mutation_record value length",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invalid := fixture
			test.mutate(&invalid)
			require.ErrorContains(t, invalid.Validate(), test.want)
		})
	}
}

func TestDecodeAOLUpgradeQueries(t *testing.T) {
	topic, err := decodeAOLUpgradeTopic(
		json.RawMessage(`{"topic":{"description":"preserved","total_records":"2","total_writers":2}}`),
		"upgrade-aol",
	)
	require.NoError(t, err)
	require.Equal(t, AOLUpgradeTopicState{
		Name:         "upgrade-aol",
		Description:  "preserved",
		TotalRecords: 2,
		TotalWriters: 2,
	}, topic)

	topics, err := decodeAOLUpgradeTopics(json.RawMessage(`{"topic_names":["zeta","alpha"]}`))
	require.NoError(t, err)
	require.Equal(t, []string{"alpha", "zeta"}, topics)

	writers, err := decodeAOLUpgradeWriterAddresses(
		json.RawMessage(`{"writer_addresses":["panacea1writerz","panacea1writera"]}`),
	)
	require.NoError(t, err)
	require.Equal(t, []string{"panacea1writera", "panacea1writerz"}, writers)

	writer, err := decodeAOLUpgradeWriter(
		json.RawMessage(`{"writer":{"moniker":"primary","description":"writer","nano_timestamp":"123"}}`),
		"panacea1writer",
	)
	require.NoError(t, err)
	require.Equal(t, int64(123), writer.NanoTimestamp)
	require.Equal(t, "panacea1writer", writer.Address)

	record, err := decodeAOLUpgradeRecord(
		json.RawMessage(`{"record":{"key":"a2V5","value":"dmFsdWU=","nano_timestamp":456,"writer_address":"panacea1writer"}}`),
		7,
	)
	require.NoError(t, err)
	require.Equal(t, uint64(7), record.Offset)
	require.Equal(t, []byte("key"), record.Key)
	require.Equal(t, []byte("value"), record.Value)
	require.Equal(t, int64(456), record.NanoTimestamp)

	_, err = decodeAOLUpgradeTopics(json.RawMessage(`{"topic_names":["duplicate","duplicate"]}`))
	require.ErrorContains(t, err, "uniquely sorted")
	_, err = decodeAOLUpgradeRecord(json.RawMessage(`{"record":null}`), 0)
	require.ErrorContains(t, err, "missing record")
}

func TestDecodeCompleteAOLUpgradePageRejectsCursorOrTotalMismatch(t *testing.T) {
	_, err := decodeCompleteAOLUpgradeTopicsPage(json.RawMessage(
		`{"topic_names":["upgrade-aol"],"pagination":{"next_key":"bmV4dA==","total":"2"}}`,
	))
	require.ErrorContains(t, err, "incomplete")

	_, err = decodeCompleteAOLUpgradeWritersPage(json.RawMessage(
		`{"writer_addresses":["panacea1writer"],"pagination":{"next_key":null,"total":"2"}}`,
	))
	require.ErrorContains(t, err, "total=2, returned=1")
}

func TestAssertAOLUpgradeCheckpointEqual(t *testing.T) {
	want := testAOLUpgradeCheckpoint(UpgradeCoveragePhasePreUpgradeCheckpoint)
	got := want
	got.RecordedAt = want.RecordedAt.Add(time.Hour)
	got.QueryHeight = want.QueryHeight + 100
	got.Phase = UpgradeCoveragePhasePostUpgradePreservation
	got.Observation.ObservedAt = got.RecordedAt
	got.Observation.Height = got.QueryHeight
	got.Observation.BlockID = "EEFF"
	got.Observation.AppHash = "0011"
	require.NoError(t, AssertAOLUpgradeCheckpointEqual(want, got))

	got.Records = cloneAOLUpgradeRecords(got.Records)
	got.Records[0].Value = []byte("changed")
	require.ErrorContains(t, AssertAOLUpgradeCheckpointEqual(want, got), "semantic state changed")

	invalid := want
	invalid.QueryHeight = 0
	require.ErrorContains(t, AssertAOLUpgradeCheckpointEqual(invalid, want), "query_height")
}

func TestPrepareAOLUpgradeFixtureRecordsCommandsAndArtifact(t *testing.T) {
	fixture := testAOLUpgradeFixture()
	type request struct {
		step    string
		keyName string
		command []string
	}
	var requests []request
	var waitedHeight int64
	var artifactPath string
	var artifact AOLUpgradePreparationEvidence
	runtime := aolUpgradeRuntime{
		broadcast: func(_ context.Context, step string, _ *cosmos.ChainNode, keyName string, command ...string) (*TxResult, error) {
			requests = append(requests, request{step: step, keyName: keyName, command: append([]string(nil), command...)})
			height := int64(10 + len(requests))
			return &TxResult{Height: strconv.FormatInt(height, 10), TxHash: fmt.Sprintf("HASH%d", height)}, nil
		},
		waitFullNodeHeight: func(_ context.Context, height int64) error {
			waitedHeight = height
			return nil
		},
		writeJSON: func(path string, value any) error {
			artifactPath = path
			artifact = value.(AOLUpgradePreparationEvidence)
			return nil
		},
	}

	evidence, err := prepareAOLUpgradeFixture(context.Background(), runtime, &cosmos.ChainNode{}, fixture)
	require.NoError(t, err)
	require.Equal(t, AOLUpgradePreparationArtifactPath, artifactPath)
	require.Equal(t, int64(13), waitedHeight)
	require.Equal(t, evidence, artifact)
	require.Equal(t, UpgradeCoveragePhaseV221Preparation, evidence.Phase)
	require.Len(t, evidence.Transactions, 3)
	require.Equal(t, []string{
		"upgrade-aol-v221-create-topic",
		"upgrade-aol-v221-add-writer",
		"upgrade-aol-v221-add-record",
	}, []string{requests[0].step, requests[1].step, requests[2].step})
	require.Equal(t, fixture.OwnerKeyName, requests[0].keyName)
	require.Equal(t, fixture.OwnerKeyName, requests[1].keyName)
	require.Equal(t, fixture.WriterKeyName, requests[2].keyName)
	require.Equal(t, []string{
		"aol", "add-record", fixture.OwnerAddress, fixture.TopicName,
		fixture.InitialRecord.Key, fixture.InitialRecord.Value,
		"--gas", aolUpgradeTxGas, "--broadcast-mode", "sync",
	}, requests[2].command)
}

func TestCaptureAOLUpgradeCheckpointUsesOnePinnedHeight(t *testing.T) {
	fixture := testAOLUpgradeFixture()
	state := newFakeAOLUpgradeState(fixture)
	var commands [][]string
	var restPaths []string
	runtime := aolUpgradeRuntime{
		fullNodeHeight:     func(context.Context) (int64, error) { return 77, nil },
		captureObservation: fakeAOLCheckpointObservation,
		query: func(_ context.Context, _ string, command ...string) (json.RawMessage, error) {
			commands = append(commands, append([]string(nil), command...))
			return state.query(command)
		},
		queryREST: func(_ context.Context, _ string, path string, height int64) (json.RawMessage, error) {
			require.Equal(t, int64(77), height)
			restPaths = append(restPaths, path)
			return state.queryREST(path)
		},
	}

	checkpoint, err := captureAOLUpgradeCheckpoint(
		context.Background(),
		runtime,
		UpgradeCoveragePhasePreUpgradeCheckpoint,
		fixture,
	)
	require.NoError(t, err)
	require.Equal(t, int64(77), checkpoint.QueryHeight)
	require.Equal(t, fixture.OwnerAddress, checkpoint.Owner.Address)
	require.Equal(t, fixture.TopicDescription, checkpoint.Topic.Description)
	require.Equal(t, uint64(1), checkpoint.Topic.TotalWriters)
	require.Equal(t, uint64(1), checkpoint.Topic.TotalRecords)
	require.Equal(t, []byte(fixture.InitialRecord.Key), checkpoint.Records[0].Key)
	require.Len(t, commands, 3)
	require.Len(t, restPaths, 2)
	for _, command := range commands {
		require.GreaterOrEqual(t, len(command), 2)
		require.Equal(t, []string{"--height", "77"}, command[len(command)-2:])
	}
}

func TestCaptureAOLUpgradeCheckpointUsesBoundedRESTPaginationAcrossCLIVersions(t *testing.T) {
	fixture := testAOLUpgradeFixture()
	state := newFakeAOLUpgradeState(fixture)
	var cliCommands [][]string
	var restPaths []string
	runtime := aolUpgradeRuntime{
		fullNodeHeight:     func(context.Context) (int64, error) { return 77, nil },
		captureObservation: fakeAOLCheckpointObservation,
		query: func(_ context.Context, _ string, command ...string) (json.RawMessage, error) {
			cliCommands = append(cliCommands, append([]string(nil), command...))
			if len(command) > 1 && (command[1] == "list-topic" || command[1] == "list-writer") {
				return nil, fmt.Errorf("source and target CLIs do not expose pagination flags: %v", command)
			}
			return state.query(command)
		},
		queryREST: func(_ context.Context, _ string, path string, height int64) (json.RawMessage, error) {
			require.Equal(t, int64(77), height)
			restPaths = append(restPaths, path)
			switch {
			case strings.Contains(path, "/writers?"):
				return json.RawMessage(`{"writer_addresses":["` + fixture.WriterAddress + `"],"pagination":{"next_key":null,"total":"1"}}`), nil
			case strings.Contains(path, "/topics?"):
				return json.RawMessage(`{"topic_names":["` + fixture.TopicName + `"],"pagination":{"next_key":null,"total":"1"}}`), nil
			default:
				return nil, fmt.Errorf("unexpected paginated AOL REST path %q", path)
			}
		},
	}

	checkpoint, err := captureAOLUpgradeCheckpoint(
		context.Background(),
		runtime,
		UpgradeCoveragePhasePreUpgradeCheckpoint,
		fixture,
	)
	require.NoError(t, err)
	require.Equal(t, []string{fixture.TopicName}, checkpoint.Owner.TopicNames)
	require.Len(t, checkpoint.Writers, 1)
	require.Equal(t, []string{
		"/panacea/aol/v2/owners/" + fixture.OwnerAddress + "/topics?pagination.count_total=true&pagination.limit=100",
		"/panacea/aol/v2/owners/" + fixture.OwnerAddress + "/topics/" + fixture.TopicName + "/writers?pagination.count_total=true&pagination.limit=100",
	}, restPaths)
	for _, command := range cliCommands {
		require.NotContains(t, command, "list-topic")
		require.NotContains(t, command, "list-writer")
		require.NotContains(t, command, "--limit")
		require.Equal(t, []string{"--height", "77"}, command[len(command)-2:])
	}
}

func TestMutateAOLAfterUpgradeProvesRecordAndWriterTransitions(t *testing.T) {
	fixture := testAOLUpgradeFixture()
	before := testAOLUpgradeCheckpoint(UpgradeCoveragePhasePostUpgradePreservation)
	state := newFakeAOLUpgradeState(fixture)
	height := int64(100)
	var waited []int64
	var artifactPath string
	var artifact AOLUpgradeMutationEvidence
	var steps []string
	runtime := aolUpgradeRuntime{
		broadcast: func(_ context.Context, step string, _ *cosmos.ChainNode, _ string, command ...string) (*TxResult, error) {
			steps = append(steps, step)
			height++
			if err := state.apply(command); err != nil {
				return nil, err
			}
			return &TxResult{Height: strconv.FormatInt(height, 10), TxHash: fmt.Sprintf("HASH%d", height)}, nil
		},
		query: func(_ context.Context, _ string, command ...string) (json.RawMessage, error) {
			return state.query(command)
		},
		queryREST: func(_ context.Context, _ string, path string, _ int64) (json.RawMessage, error) {
			return state.queryREST(path)
		},
		fullNodeHeight:     func(context.Context) (int64, error) { return height, nil },
		captureObservation: fakeAOLCheckpointObservation,
		waitFullNodeHeight: func(_ context.Context, target int64) error {
			waited = append(waited, target)
			return nil
		},
		writeJSON: func(path string, value any) error {
			artifactPath = path
			artifact = value.(AOLUpgradeMutationEvidence)
			return nil
		},
	}

	evidence, err := mutateAOLAfterUpgrade(context.Background(), runtime, &cosmos.ChainNode{}, fixture, before)
	require.NoError(t, err)
	require.Equal(t, AOLUpgradePostMutationArtifactPath, artifactPath)
	require.Equal(t, evidence, artifact)
	require.Equal(t, uint64(1), evidence.MutationRecordOffset)
	require.Equal(t, uint64(2), evidence.AfterWriterAdd.Topic.TotalWriters)
	require.Equal(t, uint64(2), evidence.AfterWriterAdd.Topic.TotalRecords)
	require.Equal(t, uint64(1), evidence.Final.Topic.TotalWriters)
	require.Equal(t, uint64(2), evidence.Final.Topic.TotalRecords)
	require.Equal(t, []byte(fixture.MutationRecord.Value), evidence.Final.Records[1].Value)
	require.Equal(t, []string{
		"upgrade-aol-post-add-record",
		"upgrade-aol-post-add-writer",
		"upgrade-aol-post-delete-writer",
	}, steps)
	require.Equal(t, []int64{102, 103}, waited)
	require.NoError(t, AssertAOLUpgradeCheckpointEqual(evidence.Final, withAOLObservation(
		evidence.Final,
		UpgradeCoveragePhasePostRestart,
		999,
	)))
}

func fakeAOLCheckpointObservation(_ context.Context, _ string, height int64) (UpgradeCheckpointObservation, error) {
	return UpgradeCheckpointObservation{
		ObservedAt:    time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC),
		Node:          "fullnode-0",
		QueryBoundary: UpgradeCheckpointQueryBoundaryCometBFTRPC,
		Height:        height,
		BlockID:       "AABB",
		AppHash:       "CCDD",
	}, nil
}

func TestValidateAOLUpgradeMutationRejectsWrongOffsetAndWriterDeletion(t *testing.T) {
	fixture := testAOLUpgradeFixture()
	before := testAOLUpgradeCheckpoint(UpgradeCoveragePhasePostUpgradePreservation)
	afterAdd := testAOLAfterAddCheckpoint(fixture, before)
	require.NoError(t, validateAOLUpgradeAfterAdd(fixture, before, afterAdd))

	wrongRecord := afterAdd
	wrongRecord.Records = cloneAOLUpgradeRecords(afterAdd.Records)
	wrongRecord.Records[1].Value = []byte("wrong")
	require.ErrorContains(t, validateAOLUpgradeAfterAdd(fixture, before, wrongRecord), "mutation record")

	final := testAOLFinalCheckpoint(before, afterAdd)
	require.NoError(t, validateAOLUpgradeAfterDelete(before, afterAdd, final))
	writerStillPresent := afterAdd
	require.ErrorContains(t, validateAOLUpgradeAfterDelete(before, afterAdd, writerStillPresent), "semantic state changed")
}

func testAOLUpgradeFixture() AOLUpgradeFixture {
	return AOLUpgradeFixture{
		OwnerAddress:              "panacea1owner",
		OwnerKeyName:              "aol-owner-key",
		WriterAddress:             "panacea1writer",
		WriterKeyName:             "aol-writer-key",
		MutationWriterAddress:     "panacea1mutation",
		TopicName:                 "upgrade-aol",
		TopicDescription:          "state retained across upgrade",
		WriterMoniker:             "primary-writer",
		WriterDescription:         "persistent writer",
		MutationWriterMoniker:     "mutation-writer",
		MutationWriterDescription: "temporary writer",
		InitialRecord: AOLUpgradeRecordInput{
			Key:   "before-key",
			Value: "before-value",
		},
		MutationRecord: AOLUpgradeRecordInput{
			Key:   "after-key",
			Value: "after-value",
		},
	}
}

func testAOLUpgradeCheckpoint(phase UpgradeCoveragePhaseName) AOLUpgradeCheckpoint {
	fixture := testAOLUpgradeFixture()
	return AOLUpgradeCheckpoint{
		SchemaVersion: AOLUpgradeEvidenceSchemaVersion,
		RecordedAt:    time.Unix(100, 0).UTC(),
		Phase:         phase,
		QueryHeight:   10,
		Observation: UpgradeCheckpointObservation{
			ObservedAt:    time.Unix(100, 0).UTC(),
			Node:          "fullnode-0",
			QueryBoundary: UpgradeCheckpointQueryBoundaryCometBFTRPC,
			Height:        10,
			BlockID:       "AABB",
			AppHash:       "CCDD",
		},
		Owner: AOLUpgradeOwnerState{
			Address:    fixture.OwnerAddress,
			TopicNames: []string{fixture.TopicName},
		},
		Topic: AOLUpgradeTopicState{
			Name:         fixture.TopicName,
			Description:  fixture.TopicDescription,
			TotalRecords: 1,
			TotalWriters: 1,
		},
		Writers: []AOLUpgradeWriterState{{
			Address:       fixture.WriterAddress,
			Moniker:       fixture.WriterMoniker,
			Description:   fixture.WriterDescription,
			NanoTimestamp: 101,
		}},
		Records: []AOLUpgradeRecordState{{
			Offset:        0,
			Key:           []byte(fixture.InitialRecord.Key),
			Value:         []byte(fixture.InitialRecord.Value),
			NanoTimestamp: 102,
			WriterAddress: fixture.WriterAddress,
		}},
	}
}

func testAOLAfterAddCheckpoint(fixture AOLUpgradeFixture, before AOLUpgradeCheckpoint) AOLUpgradeCheckpoint {
	after := withAOLObservation(before, UpgradeCoveragePhasePostUpgradeMutation, before.QueryHeight+1)
	after.Topic.TotalRecords++
	after.Topic.TotalWriters++
	after.Writers = append(after.Writers, AOLUpgradeWriterState{
		Address:       fixture.MutationWriterAddress,
		Moniker:       fixture.MutationWriterMoniker,
		Description:   fixture.MutationWriterDescription,
		NanoTimestamp: 201,
	})
	sort.Slice(after.Writers, func(i, j int) bool { return after.Writers[i].Address < after.Writers[j].Address })
	after.Records = append(after.Records, AOLUpgradeRecordState{
		Offset:        1,
		Key:           []byte(fixture.MutationRecord.Key),
		Value:         []byte(fixture.MutationRecord.Value),
		NanoTimestamp: 202,
		WriterAddress: fixture.WriterAddress,
	})
	return after
}

func testAOLFinalCheckpoint(before, afterAdd AOLUpgradeCheckpoint) AOLUpgradeCheckpoint {
	final := withAOLObservation(before, UpgradeCoveragePhasePostUpgradeMutation, afterAdd.QueryHeight+1)
	final.Topic.TotalRecords = afterAdd.Topic.TotalRecords
	final.Records = cloneAOLUpgradeRecords(afterAdd.Records)
	return final
}

func withAOLObservation(checkpoint AOLUpgradeCheckpoint, phase UpgradeCoveragePhaseName, height int64) AOLUpgradeCheckpoint {
	checkpoint.RecordedAt = checkpoint.RecordedAt.Add(time.Minute)
	checkpoint.Phase = phase
	checkpoint.QueryHeight = height
	checkpoint.Observation.ObservedAt = checkpoint.RecordedAt
	checkpoint.Observation.Height = height
	checkpoint.Observation.BlockID = fmt.Sprintf("BLOCK-%d", height)
	checkpoint.Observation.AppHash = fmt.Sprintf("APP-%d", height)
	checkpoint.Owner.TopicNames = append([]string(nil), checkpoint.Owner.TopicNames...)
	checkpoint.Writers = append([]AOLUpgradeWriterState(nil), checkpoint.Writers...)
	checkpoint.Records = cloneAOLUpgradeRecords(checkpoint.Records)
	return checkpoint
}

type fakeAOLUpgradeState struct {
	fixture AOLUpgradeFixture
	writers map[string]AOLUpgradeWriterState
	records []AOLUpgradeRecordState
}

func newFakeAOLUpgradeState(fixture AOLUpgradeFixture) *fakeAOLUpgradeState {
	return &fakeAOLUpgradeState{
		fixture: fixture,
		writers: map[string]AOLUpgradeWriterState{
			fixture.WriterAddress: {
				Address:       fixture.WriterAddress,
				Moniker:       fixture.WriterMoniker,
				Description:   fixture.WriterDescription,
				NanoTimestamp: 101,
			},
		},
		records: []AOLUpgradeRecordState{{
			Offset:        0,
			Key:           []byte(fixture.InitialRecord.Key),
			Value:         []byte(fixture.InitialRecord.Value),
			NanoTimestamp: 102,
			WriterAddress: fixture.WriterAddress,
		}},
	}
}

func (state *fakeAOLUpgradeState) apply(command []string) error {
	if len(command) < 2 || command[0] != "aol" {
		return fmt.Errorf("unexpected AOL command: %v", command)
	}
	switch command[1] {
	case "add-record":
		state.records = append(state.records, AOLUpgradeRecordState{
			Offset:        uint64(len(state.records)),
			Key:           []byte(command[4]),
			Value:         []byte(command[5]),
			NanoTimestamp: 202,
			WriterAddress: state.fixture.WriterAddress,
		})
	case "add-writer":
		address := command[3]
		state.writers[address] = AOLUpgradeWriterState{
			Address:       address,
			Moniker:       flagValue(command, "--moniker"),
			Description:   flagValue(command, "--description"),
			NanoTimestamp: 201,
		}
	case "delete-writer":
		delete(state.writers, command[3])
	default:
		return fmt.Errorf("unexpected AOL mutation command: %v", command)
	}
	return nil
}

func (state *fakeAOLUpgradeState) query(command []string) (json.RawMessage, error) {
	if len(command) < 2 || command[0] != "aol" {
		return nil, fmt.Errorf("unexpected AOL query: %v", command)
	}
	marshal := func(value any) (json.RawMessage, error) {
		encoded, err := json.Marshal(value)
		return json.RawMessage(encoded), err
	}
	switch command[1] {
	case "get-topic":
		return marshal(map[string]any{"topic": map[string]any{
			"description":   state.fixture.TopicDescription,
			"total_records": strconv.Itoa(len(state.records)),
			"total_writers": strconv.Itoa(len(state.writers)),
		}})
	case "list-topic":
		return marshal(map[string]any{"topic_names": []string{state.fixture.TopicName}})
	case "list-writer":
		addresses := make([]string, 0, len(state.writers))
		for address := range state.writers {
			addresses = append(addresses, address)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(addresses)))
		return marshal(map[string]any{"writer_addresses": addresses})
	case "get-writer":
		writer, ok := state.writers[command[4]]
		if !ok {
			return nil, fmt.Errorf("writer %s is absent", command[4])
		}
		return marshal(map[string]any{"writer": map[string]any{
			"moniker":        writer.Moniker,
			"description":    writer.Description,
			"nano_timestamp": strconv.FormatInt(writer.NanoTimestamp, 10),
		}})
	case "get-record":
		offset, err := strconv.Atoi(command[4])
		if err != nil || offset < 0 || offset >= len(state.records) {
			return nil, fmt.Errorf("record offset %q is absent", command[4])
		}
		record := state.records[offset]
		return marshal(map[string]any{"record": map[string]any{
			"key":            record.Key,
			"value":          record.Value,
			"nano_timestamp": strconv.FormatInt(record.NanoTimestamp, 10),
			"writer_address": record.WriterAddress,
		}})
	default:
		return nil, fmt.Errorf("unexpected AOL query command: %v", command)
	}
}

func (state *fakeAOLUpgradeState) queryREST(path string) (json.RawMessage, error) {
	marshal := func(value any) (json.RawMessage, error) {
		encoded, err := json.Marshal(value)
		return json.RawMessage(encoded), err
	}
	pagination := map[string]any{"next_key": nil}
	switch {
	case strings.Contains(path, "/writers?"):
		addresses := make([]string, 0, len(state.writers))
		for address := range state.writers {
			addresses = append(addresses, address)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(addresses)))
		pagination["total"] = strconv.Itoa(len(addresses))
		return marshal(map[string]any{"writer_addresses": addresses, "pagination": pagination})
	case strings.Contains(path, "/topics?"):
		pagination["total"] = "1"
		return marshal(map[string]any{
			"topic_names": []string{state.fixture.TopicName},
			"pagination":  pagination,
		})
	default:
		return nil, fmt.Errorf("unexpected AOL REST query: %s", path)
	}
}

func flagValue(command []string, flag string) string {
	for index := 0; index+1 < len(command); index++ {
		if command[index] == flag {
			return command[index+1]
		}
	}
	return ""
}
