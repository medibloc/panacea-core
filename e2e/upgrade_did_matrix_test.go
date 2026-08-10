package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

const (
	upgradeDIDControllerKey         = "upgrade-did-controller"
	upgradeDIDUpdateDocumentPath    = "upgrade/did/update-document.json"
	upgradeDIDServiceID             = "upgrade-continuity-service"
	upgradeDIDServiceType           = "LinkedDomains"
	upgradeDIDServiceEndpoint       = "https://upgrade.example.test/did-continuity"
	upgradeDIDDeactivatedGRPCCode   = "NotFound"
	upgradeDIDDeactivatedDiagnostic = "code = NotFound desc = DID deactivated"
)

type upgradeDIDFixture struct {
	DID                  string          `json:"did"`
	VerificationMethodID string          `json:"verification_method_id"`
	CreateTxHash         string          `json:"create_tx_hash"`
	Response             json.RawMessage `json:"response"`
}

type upgradeDIDFixtures struct {
	RecordedAt  time.Time                            `json:"recorded_at"`
	QueryHeight int64                                `json:"query_height"`
	Observation harness.UpgradeCheckpointObservation `json:"observation"`
	KeyName     string                               `json:"key_name"`
	Updated     upgradeDIDFixture                    `json:"updated"`
	Deactivated upgradeDIDFixture                    `json:"deactivated"`
}

type upgradeDIDMutationEvidence struct {
	RecordedAt       time.Time                            `json:"recorded_at"`
	QueryHeight      int64                                `json:"query_height"`
	Observation      harness.UpgradeCheckpointObservation `json:"observation"`
	UpdateTxHash     string                               `json:"update_tx_hash"`
	DeactivateTxHash string                               `json:"deactivate_tx_hash"`
	UpdatedResponse  json.RawMessage                      `json:"updated_response"`
	DeactivatedQuery upgradeDIDDeactivatedQueryEvidence   `json:"deactivated_query"`
}

type upgradeDIDPersistenceEvidence struct {
	RecordedAt       time.Time                            `json:"recorded_at"`
	QueryHeight      int64                                `json:"query_height"`
	Observation      harness.UpgradeCheckpointObservation `json:"observation"`
	UpdatedResponse  json.RawMessage                      `json:"updated_response"`
	DeactivatedQuery upgradeDIDDeactivatedQueryEvidence   `json:"deactivated_query"`
}

type upgradeDIDDeactivatedQueryEvidence struct {
	DID                string `json:"did"`
	QueryHeight        int64  `json:"query_height"`
	ExpectedGRPCCode   string `json:"expected_grpc_code"`
	ExpectedDiagnostic string `json:"expected_diagnostic"`
	Observed           bool   `json:"observed"`
}

type upgradeDIDExpectedErrorQuery func(
	context.Context,
	string,
	string,
	...string,
) error

type upgradeDIDQueryState struct {
	Document map[string]any
	Sequence uint64
}

func validateUpgradeDIDCheckpointContract(
	recordedAt time.Time,
	queryHeight int64,
	observation harness.UpgradeCheckpointObservation,
) error {
	if err := observation.Validate(); err != nil {
		return fmt.Errorf("DID checkpoint observation: %w", err)
	}
	if queryHeight != observation.Height {
		return fmt.Errorf("DID checkpoint query height %d does not match observation height %d", queryHeight, observation.Height)
	}
	if recordedAt.IsZero() || !recordedAt.Equal(observation.ObservedAt) {
		return errors.New("DID checkpoint observed_at does not match recorded_at")
	}
	return nil
}

func prepareUpgradeDIDFixtures(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
) upgradeDIDFixtures {
	t.Helper()
	wallet := buildAndFundNFTWallet(t, ctx, network, upgradeDIDControllerKey)
	node := network.Chain.Validators[0]

	updateCreateTx, err := network.BroadcastDIDCreateAndWaitTx(
		ctx,
		"upgrade-create-did-for-update",
		node,
		wallet.KeyName(),
		"did", "create-did",
		"--gas", "1000000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	identifiers, err := network.DIDVerificationMethodIDs(ctx, node)
	require.NoError(t, err)
	require.Len(t, identifiers, 1)
	updateVerificationMethodID := identifiers[0]
	updateDID, err := didFromVerificationMethodID(updateVerificationMethodID)
	require.NoError(t, err)
	updateRaw, err := network.FullNodeCLIQuery(
		ctx,
		"upgrade-pre-did-for-update",
		"did", "get-did", updateDID,
	)
	require.NoError(t, err)
	_, err = decodeUpgradeDIDQueryState(updateRaw)
	require.NoError(t, err)

	deactivateCreateTx, err := network.BroadcastDIDCreateAndWaitTx(
		ctx,
		"upgrade-create-did-for-deactivate",
		node,
		wallet.KeyName(),
		"did", "create-did",
		"--gas", "1000000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	identifiers, err = network.DIDVerificationMethodIDs(ctx, node)
	require.NoError(t, err)
	require.Len(t, identifiers, 2)
	deactivateVerificationMethodID, err := distinctUpgradeDIDIdentifier(identifiers, updateVerificationMethodID)
	require.NoError(t, err)
	deactivateDID, err := didFromVerificationMethodID(deactivateVerificationMethodID)
	require.NoError(t, err)
	deactivateRaw, err := network.FullNodeCLIQuery(
		ctx,
		"upgrade-pre-did-for-deactivate",
		"did", "get-did", deactivateDID,
	)
	require.NoError(t, err)
	_, err = decodeUpgradeDIDQueryState(deactivateRaw)
	require.NoError(t, err)

	observation, err := network.CaptureUpgradeCheckpointObservation(
		ctx,
		"upgrade-did-pre-upgrade",
		network.Chain.FullNodes[0],
		0,
	)
	require.NoError(t, err)
	updateRaw, err = queryUpgradeDIDAtHeight(
		ctx,
		network,
		"upgrade-pre-checkpoint-did-for-update",
		updateDID,
		observation.Height,
	)
	require.NoError(t, err)
	_, err = decodeUpgradeDIDQueryState(updateRaw)
	require.NoError(t, err)
	deactivateRaw, err = queryUpgradeDIDAtHeight(
		ctx,
		network,
		"upgrade-pre-checkpoint-did-for-deactivate",
		deactivateDID,
		observation.Height,
	)
	require.NoError(t, err)
	_, err = decodeUpgradeDIDQueryState(deactivateRaw)
	require.NoError(t, err)

	fixtures := upgradeDIDFixtures{
		RecordedAt:  observation.ObservedAt,
		QueryHeight: observation.Height,
		Observation: observation,
		KeyName:     wallet.KeyName(),
		Updated: upgradeDIDFixture{
			DID:                  updateDID,
			VerificationMethodID: updateVerificationMethodID,
			CreateTxHash:         updateCreateTx.TxHash,
			Response:             append(json.RawMessage(nil), updateRaw...),
		},
		Deactivated: upgradeDIDFixture{
			DID:                  deactivateDID,
			VerificationMethodID: deactivateVerificationMethodID,
			CreateTxHash:         deactivateCreateTx.TxHash,
			Response:             append(json.RawMessage(nil), deactivateRaw...),
		},
	}
	require.NoError(t, validateUpgradeDIDCheckpointContract(fixtures.RecordedAt, fixtures.QueryHeight, fixtures.Observation))
	require.NoError(t, network.WriteArtifactJSON("upgrade/did/pre-upgrade.json", fixtures))
	return fixtures
}

func assertUpgradeDIDFixturesPreserved(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	fixtures upgradeDIDFixtures,
) {
	t.Helper()
	for name, fixture := range map[string]upgradeDIDFixture{
		"update":     fixtures.Updated,
		"deactivate": fixtures.Deactivated,
	} {
		raw, err := network.FullNodeCLIQuery(
			ctx,
			"upgrade-post-preservation-did-"+name,
			"did", "get-did", fixture.DID,
		)
		require.NoError(t, err)
		require.JSONEq(t, string(fixture.Response), string(raw))
	}
}

func runPostUpgradeDIDMutations(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	fixtures upgradeDIDFixtures,
) upgradeDIDMutationEvidence {
	t.Helper()
	node := network.Chain.Validators[0]
	updateDocument, service, err := makeUpgradeDIDUpdateDocument(fixtures.Updated.Response)
	require.NoError(t, err)
	require.NoError(t, node.WriteFile(ctx, updateDocument, upgradeDIDUpdateDocumentPath))
	require.NoError(t, network.WriteArtifact(upgradeDIDUpdateDocumentPath, updateDocument))

	updateTx, err := network.BroadcastDIDAuthenticatedAndWaitTx(
		ctx,
		"upgrade-post-update-did",
		node,
		fixtures.KeyName,
		"did", "update-did",
		fixtures.Updated.DID,
		fixtures.Updated.VerificationMethodID,
		path.Join(node.HomeDir(), upgradeDIDUpdateDocumentPath),
		"--gas", "1000000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	deactivateTx, err := network.BroadcastDIDAuthenticatedAndWaitTx(
		ctx,
		"upgrade-post-deactivate-did",
		node,
		fixtures.KeyName,
		"did", "deactivate-did",
		fixtures.Deactivated.DID,
		fixtures.Deactivated.VerificationMethodID,
		"--gas", "1000000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)

	observation, err := network.CaptureUpgradeCheckpointObservation(
		ctx,
		"upgrade-did-post-upgrade",
		network.Chain.FullNodes[0],
		0,
	)
	require.NoError(t, err)
	updatedRaw, err := queryUpgradeDIDAtHeight(
		ctx,
		network,
		"upgrade-post-mutated-did",
		fixtures.Updated.DID,
		observation.Height,
	)
	require.NoError(t, err)
	preUpdate, err := decodeUpgradeDIDQueryState(fixtures.Updated.Response)
	require.NoError(t, err)
	updated, err := decodeUpgradeDIDQueryState(updatedRaw)
	require.NoError(t, err)
	require.Equal(t, preUpdate.Sequence+1, updated.Sequence)
	require.Equal(t, service, upgradeDIDServiceFromDocument(updated.Document, upgradeDIDServiceID))

	deactivatedQuery, err := queryUpgradeDIDDeactivatedAtHeight(
		ctx,
		network.FullNodeGRPCQueryExpectedError,
		"upgrade-post-deactivated-did",
		fixtures.Deactivated.DID,
		observation.Height,
	)
	require.NoError(t, err)

	evidence := upgradeDIDMutationEvidence{
		RecordedAt:       observation.ObservedAt,
		QueryHeight:      observation.Height,
		Observation:      observation,
		UpdateTxHash:     updateTx.TxHash,
		DeactivateTxHash: deactivateTx.TxHash,
		UpdatedResponse:  append(json.RawMessage(nil), updatedRaw...),
		DeactivatedQuery: deactivatedQuery,
	}
	require.NoError(t, validateUpgradeDIDCheckpointContract(evidence.RecordedAt, evidence.QueryHeight, evidence.Observation))
	require.NoError(t, network.WriteArtifactJSON("upgrade/did/post-upgrade.json", evidence))
	return evidence
}

func assertUpgradeDIDMutationsPersisted(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	fixtures upgradeDIDFixtures,
	want upgradeDIDMutationEvidence,
) {
	t.Helper()
	observation, err := network.CaptureUpgradeCheckpointObservation(
		ctx,
		"upgrade-did-post-restart",
		network.Chain.FullNodes[0],
		0,
	)
	require.NoError(t, err)
	updatedRaw, err := queryUpgradeDIDAtHeight(
		ctx,
		network,
		"upgrade-post-restart-updated-did",
		fixtures.Updated.DID,
		observation.Height,
	)
	require.NoError(t, err)
	require.JSONEq(t, string(want.UpdatedResponse), string(updatedRaw))
	deactivatedQuery, err := queryUpgradeDIDDeactivatedAtHeight(
		ctx,
		network.FullNodeGRPCQueryExpectedError,
		"upgrade-post-restart-deactivated-did",
		fixtures.Deactivated.DID,
		observation.Height,
	)
	require.NoError(t, err)
	require.Equal(t, want.DeactivatedQuery.DID, deactivatedQuery.DID)
	require.Equal(t, want.DeactivatedQuery.ExpectedGRPCCode, deactivatedQuery.ExpectedGRPCCode)
	require.Equal(t, want.DeactivatedQuery.ExpectedDiagnostic, deactivatedQuery.ExpectedDiagnostic)
	require.True(t, want.DeactivatedQuery.Observed)
	require.True(t, deactivatedQuery.Observed)
	evidence := upgradeDIDPersistenceEvidence{
		RecordedAt:       observation.ObservedAt,
		QueryHeight:      observation.Height,
		Observation:      observation,
		UpdatedResponse:  append(json.RawMessage(nil), updatedRaw...),
		DeactivatedQuery: deactivatedQuery,
	}
	require.NoError(t, validateUpgradeDIDCheckpointContract(evidence.RecordedAt, evidence.QueryHeight, evidence.Observation))
	require.NoError(t, network.WriteArtifactJSON("upgrade/did/post-restart.json", evidence))
}

func queryUpgradeDIDAtHeight(
	ctx context.Context,
	network *harness.Network,
	step string,
	did string,
	height int64,
) (json.RawMessage, error) {
	if height <= 0 {
		return nil, fmt.Errorf("DID checkpoint query height must be positive, got %d", height)
	}
	return network.FullNodeCLIQuery(
		ctx,
		step,
		"did", "get-did", did,
		"--height", strconv.FormatInt(height, 10),
	)
}

func queryUpgradeDIDDeactivatedAtHeight(
	ctx context.Context,
	query upgradeDIDExpectedErrorQuery,
	step string,
	did string,
	height int64,
) (upgradeDIDDeactivatedQueryEvidence, error) {
	evidence := upgradeDIDDeactivatedQueryEvidence{
		DID:                did,
		QueryHeight:        height,
		ExpectedGRPCCode:   upgradeDIDDeactivatedGRPCCode,
		ExpectedDiagnostic: upgradeDIDDeactivatedDiagnostic,
	}
	if height <= 0 {
		return evidence, fmt.Errorf("deactivated DID query height must be positive, got %d", height)
	}
	if strings.TrimSpace(did) == "" {
		return evidence, errors.New("deactivated DID query requires a DID")
	}
	if query == nil {
		return evidence, errors.New("deactivated DID expected-error query is required")
	}
	if err := query(
		ctx,
		step,
		upgradeDIDDeactivatedDiagnostic,
		"did", "get-did", did,
		"--height", strconv.FormatInt(height, 10),
	); err != nil {
		return evidence, fmt.Errorf("query deactivated DID %q at height %d: %w", did, height, err)
	}
	evidence.Observed = true
	return evidence, nil
}

func didFromVerificationMethodID(identifier string) (string, error) {
	did, fragment, found := strings.Cut(strings.TrimSpace(identifier), "#")
	if !found || fragment == "" || !strings.HasPrefix(did, "did:panacea:") {
		return "", fmt.Errorf("invalid Panacea verification method ID %q", identifier)
	}
	return did, nil
}

func distinctUpgradeDIDIdentifier(identifiers []string, existing string) (string, error) {
	var distinct string
	for _, identifier := range identifiers {
		if identifier == existing {
			continue
		}
		if distinct != "" {
			return "", errors.New("multiple new DID verification method identifiers")
		}
		distinct = identifier
	}
	if distinct == "" {
		return "", errors.New("new DID verification method identifier was not found")
	}
	return distinct, nil
}

func decodeUpgradeDIDQueryState(raw []byte) (upgradeDIDQueryState, error) {
	var response struct {
		DIDDocumentWithSeq struct {
			Document map[string]any  `json:"document"`
			Sequence json.RawMessage `json:"sequence"`
		} `json:"did_document_with_seq"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return upgradeDIDQueryState{}, fmt.Errorf("decode DID query: %w", err)
	}
	if response.DIDDocumentWithSeq.Document == nil {
		return upgradeDIDQueryState{}, errors.New("DID query has no document")
	}
	sequenceText := strings.Trim(strings.TrimSpace(string(response.DIDDocumentWithSeq.Sequence)), `"`)
	sequence, err := strconv.ParseUint(sequenceText, 10, 64)
	if err != nil {
		return upgradeDIDQueryState{}, fmt.Errorf("decode DID sequence %q: %w", sequenceText, err)
	}
	return upgradeDIDQueryState{
		Document: response.DIDDocumentWithSeq.Document,
		Sequence: sequence,
	}, nil
}

func makeUpgradeDIDUpdateDocument(raw []byte) ([]byte, map[string]any, error) {
	state, err := decodeUpgradeDIDQueryState(raw)
	if err != nil {
		return nil, nil, err
	}
	if upgradeDIDDocumentDeactivated(state.Document) {
		return nil, nil, errors.New("cannot update a deactivated DID document")
	}
	service := map[string]any{
		"id":               upgradeDIDServiceID,
		"type":             upgradeDIDServiceType,
		"service_endpoint": upgradeDIDServiceEndpoint,
	}
	state.Document["services"] = []any{service}
	document, err := json.MarshalIndent(state.Document, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("encode updated DID document: %w", err)
	}
	return document, service, nil
}

func upgradeDIDServiceFromDocument(document map[string]any, serviceID string) map[string]any {
	services, _ := document["services"].([]any)
	for _, value := range services {
		service, _ := value.(map[string]any)
		if service["id"] == serviceID {
			return service
		}
	}
	return nil
}

func upgradeDIDDocumentDeactivated(document map[string]any) bool {
	id, _ := document["id"].(string)
	return strings.TrimSpace(id) == ""
}

func TestUpgradeDIDDocumentMutation(t *testing.T) {
	t.Parallel()
	active := []byte(`{
		"did_document_with_seq": {
			"document": {
				"contexts":["https://www.w3.org/ns/did/v1"],
				"id":"did:panacea:fixture",
				"verification_methods":[],
				"authentications":[],
				"services":[]
			},
			"sequence":"7"
		}
	}`)
	document, service, err := makeUpgradeDIDUpdateDocument(active)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(document, &decoded))
	require.Equal(t, service, upgradeDIDServiceFromDocument(decoded, upgradeDIDServiceID))

}

func TestDistinctUpgradeDIDIdentifier(t *testing.T) {
	t.Parallel()
	identifier, err := distinctUpgradeDIDIdentifier(
		[]string{"did:panacea:first#key1", "did:panacea:second#key1"},
		"did:panacea:first#key1",
	)
	require.NoError(t, err)
	require.Equal(t, "did:panacea:second#key1", identifier)
	_, err = distinctUpgradeDIDIdentifier([]string{"did:panacea:first#key1"}, "did:panacea:first#key1")
	require.Error(t, err)
}
