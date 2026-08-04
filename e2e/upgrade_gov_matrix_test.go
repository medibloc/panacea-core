package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/v2/e2e/internal/harness"
)

type upgradeGovCoin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type upgradeGovParams struct {
	MinDeposit                 []upgradeGovCoin `json:"min_deposit"`
	MaxDepositPeriod           string           `json:"max_deposit_period"`
	VotingPeriod               string           `json:"voting_period"`
	Quorum                     string           `json:"quorum"`
	Threshold                  string           `json:"threshold"`
	VetoThreshold              string           `json:"veto_threshold"`
	ExpeditedMinDeposit        []upgradeGovCoin `json:"expedited_min_deposit,omitempty"`
	ExpeditedVotingPeriod      string           `json:"expedited_voting_period,omitempty"`
	ProposalCancelRatio        string           `json:"proposal_cancel_ratio,omitempty"`
	ProposalCancelDestination  string           `json:"proposal_cancel_destination,omitempty"`
	MinInitialDepositRatio     string           `json:"min_initial_deposit_ratio,omitempty"`
	BurnVoteQuorum             bool             `json:"burn_vote_quorum,omitempty"`
	BurnProposalDepositPrevote bool             `json:"burn_proposal_deposit_prevote,omitempty"`
	BurnVoteVeto               bool             `json:"burn_vote_veto,omitempty"`
}

type upgradeGovTally struct {
	Yes        string `json:"yes"`
	Abstain    string `json:"abstain"`
	No         string `json:"no"`
	NoWithVeto string `json:"no_with_veto"`
}

type upgradeGovCheckpoint struct {
	Phase       string                               `json:"phase"`
	Observation harness.UpgradeCheckpointObservation `json:"observation"`
	ProposalID  uint64                               `json:"proposal_id"`
	Params      upgradeGovParams                     `json:"params"`
	Tally       upgradeGovTally                      `json:"tally"`
}

type upgradeGovProposalState struct {
	ID           uint64           `json:"id"`
	Status       string           `json:"status"`
	TotalDeposit []upgradeGovCoin `json:"total_deposit"`
	Expedited    bool             `json:"expedited"`
	Title        string           `json:"title"`
	Summary      string           `json:"summary"`
}

type upgradeExpeditedGovEvidence struct {
	ProposalID      uint64                                `json:"proposal_id"`
	Deposit         string                                `json:"deposit"`
	ExpectedCoins   []upgradeGovCoin                      `json:"expected_coins"`
	SubmitTxHash    string                                `json:"submit_tx_hash"`
	SubmitHeight    string                                `json:"submit_height"`
	VoteTxHashes    []string                              `json:"vote_tx_hashes"`
	AfterSubmission upgradeGovProposalState               `json:"after_submission"`
	AfterVoting     upgradeGovProposalState               `json:"after_voting"`
	PostRestart     upgradeGovProposalState               `json:"post_restart,omitempty"`
	Observation     *harness.UpgradeCheckpointObservation `json:"observation,omitempty"`
}

func captureUpgradeGovCheckpoint(
	ctx context.Context,
	network *harness.Network,
	phase string,
	proposalID uint64,
) (upgradeGovCheckpoint, error) {
	if network == nil || network.Chain == nil || len(network.Chain.FullNodes) == 0 {
		return upgradeGovCheckpoint{}, errors.New("governance checkpoint requires a full node")
	}
	observation, err := network.CaptureUpgradeCheckpointObservation(
		ctx,
		"upgrade-gov-"+phase,
		network.Chain.FullNodes[0],
		0,
	)
	if err != nil {
		return upgradeGovCheckpoint{}, err
	}
	height := strconv.FormatInt(observation.Height, 10)
	paramsRaw, err := network.FullNodeCLIQuery(ctx, "upgrade-gov-"+phase+"-params", "gov", "params", "--height", height)
	if err != nil {
		return upgradeGovCheckpoint{}, err
	}
	params, err := decodeUpgradeGovParams(paramsRaw)
	if err != nil {
		return upgradeGovCheckpoint{}, fmt.Errorf("decode %s gov params: %w", phase, err)
	}
	tallyRaw, err := network.FullNodeCLIQuery(
		ctx,
		"upgrade-gov-"+phase+"-tally",
		"gov", "tally", strconv.FormatUint(proposalID, 10), "--height", height,
	)
	if err != nil {
		return upgradeGovCheckpoint{}, err
	}
	tally, err := decodeUpgradeGovTally(tallyRaw)
	if err != nil {
		return upgradeGovCheckpoint{}, fmt.Errorf("decode %s gov tally: %w", phase, err)
	}
	checkpoint := upgradeGovCheckpoint{
		Phase:       phase,
		Observation: observation,
		ProposalID:  proposalID,
		Params:      params,
		Tally:       tally,
	}
	if err := network.WriteArtifactJSON("state-checkpoints/gov-"+phase+".json", checkpoint); err != nil {
		return upgradeGovCheckpoint{}, err
	}
	return checkpoint, nil
}

func submitPostUpgradeGovProposal(
	t *testing.T,
	ctx context.Context,
	network *harness.Network,
	keyName string,
) upgradeExpeditedGovEvidence {
	t.Helper()
	paramsRaw, err := network.FullNodeCLIQuery(ctx, "upgrade-expedited-gov-params", "gov", "params")
	require.NoError(t, err)
	params, err := decodeUpgradeGovParams(paramsRaw)
	require.NoError(t, err)
	require.NotEmpty(t, params.ExpeditedMinDeposit)
	deposit, err := formatUpgradeGovCoins(params.ExpeditedMinDeposit)
	require.NoError(t, err)
	const proposalFile = "upgrade/gov-expedited-proposal-input.json"
	proposalInput := map[string]any{
		"messages":  []any{},
		"metadata":  "post-upgrade expedited governance continuity",
		"deposit":   deposit,
		"title":     "Post-upgrade expedited governance continuity",
		"summary":   "Exercise migrated expedited governance with the exact minimum deposit",
		"expedited": true,
	}
	proposalJSON, err := json.Marshal(proposalInput)
	require.NoError(t, err)
	node := network.Chain.Validators[0]
	require.NoError(t, node.WriteFile(ctx, proposalJSON, proposalFile))
	require.NoError(t, network.WriteArtifact(proposalFile, proposalJSON))
	result, err := network.BroadcastAndWaitTx(
		ctx,
		"upgrade-post-gov-expedited-proposal",
		node,
		keyName,
		"gov", "submit-proposal", path.Join(node.HomeDir(), proposalFile),
		"--gas", "500000",
		"--broadcast-mode", "sync",
	)
	require.NoError(t, err)
	proposalID, err := proposalIDFromCommittedTx(result)
	require.NoError(t, err)
	afterSubmission, err := queryUpgradeGovProposal(ctx, network, "upgrade-expedited-after-submission", proposalID)
	require.NoError(t, err)
	require.True(t, afterSubmission.Expedited)
	require.Equal(t, params.ExpeditedMinDeposit, afterSubmission.TotalDeposit)
	require.Contains(t, []string{"PROPOSAL_STATUS_VOTING_PERIOD", "VOTING_PERIOD"}, strings.ToUpper(afterSubmission.Status))

	voteRequests := make([]harness.TxBatchRequest, len(network.Chain.Validators))
	for index, validator := range network.Chain.Validators {
		voteRequests[index] = harness.TxBatchRequest{
			Step:    "upgrade-expedited-gov-vote-" + strconv.Itoa(index),
			Node:    validator,
			KeyName: "validator",
			Command: []string{
				"gov", "vote", strconv.FormatUint(proposalID, 10), "yes",
				"--gas", "500000",
				"--broadcast-mode", "sync",
			},
		}
	}
	votes, err := network.BroadcastAndWaitTxBatch(ctx, voteRequests)
	require.NoError(t, err)
	voteTxHashes := make([]string, 0, len(votes))
	for _, vote := range votes {
		require.NotNil(t, vote)
		voteTxHashes = append(voteTxHashes, vote.TxHash)
	}
	require.NoError(t, waitForProposalPassed(ctx, network, proposalID))
	afterVoting, err := queryUpgradeGovProposal(ctx, network, "upgrade-expedited-after-voting", proposalID)
	require.NoError(t, err)
	require.True(t, afterVoting.Expedited)
	require.Equal(t, params.ExpeditedMinDeposit, afterVoting.TotalDeposit)
	require.Contains(t, []string{"PROPOSAL_STATUS_PASSED", "PASSED"}, strings.ToUpper(afterVoting.Status))
	evidence := upgradeExpeditedGovEvidence{
		ProposalID:      proposalID,
		Deposit:         deposit,
		ExpectedCoins:   append([]upgradeGovCoin(nil), params.ExpeditedMinDeposit...),
		SubmitTxHash:    result.TxHash,
		SubmitHeight:    result.Height,
		VoteTxHashes:    voteTxHashes,
		AfterSubmission: afterSubmission,
		AfterVoting:     afterVoting,
	}
	require.NoError(t, network.WriteArtifactJSON("upgrade/post-governance-expedited-proposal.json", evidence))
	return evidence
}

func capturePostRestartExpeditedGovProposal(
	ctx context.Context,
	network *harness.Network,
	evidence *upgradeExpeditedGovEvidence,
) error {
	if evidence == nil || evidence.ProposalID == 0 {
		return errors.New("expedited governance evidence is required")
	}
	proposal, err := queryUpgradeGovProposal(ctx, network, "upgrade-expedited-post-restart", evidence.ProposalID)
	if err != nil {
		return err
	}
	if !proposal.Expedited || !upgradeGovCoinsEqual(proposal.TotalDeposit, evidence.ExpectedCoins) {
		return fmt.Errorf("post-restart expedited proposal flag/deposit changed: %+v", proposal)
	}
	if status := strings.ToUpper(proposal.Status); status != "PROPOSAL_STATUS_PASSED" && status != "PASSED" {
		return fmt.Errorf("post-restart expedited proposal status=%q, want passed", proposal.Status)
	}
	evidence.PostRestart = proposal
	observation, err := network.CaptureUpgradeCheckpointObservation(
		ctx,
		"upgrade-expedited-post-restart",
		network.Chain.FullNodes[0],
		0,
	)
	if err != nil {
		return err
	}
	evidence.Observation = &observation
	return network.WriteArtifactJSON("state-checkpoints/gov-expedited-post-restart.json", evidence)
}

func queryUpgradeGovProposal(
	ctx context.Context,
	network *harness.Network,
	step string,
	proposalID uint64,
) (upgradeGovProposalState, error) {
	raw, err := network.FullNodeCLIQuery(ctx, step, "gov", "proposal", strconv.FormatUint(proposalID, 10))
	if err != nil {
		return upgradeGovProposalState{}, err
	}
	return decodeUpgradeGovProposal(raw)
}

func decodeUpgradeGovProposal(raw []byte) (upgradeGovProposalState, error) {
	type proposalJSON struct {
		ID           json.RawMessage  `json:"id"`
		ProposalID   json.RawMessage  `json:"proposal_id"`
		Status       string           `json:"status"`
		TotalDeposit []upgradeGovCoin `json:"total_deposit"`
		Expedited    bool             `json:"expedited"`
		Title        string           `json:"title"`
		Summary      string           `json:"summary"`
	}
	var response struct {
		Proposal *proposalJSON `json:"proposal"`
		proposalJSON
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return upgradeGovProposalState{}, err
	}
	proposal := response.proposalJSON
	if response.Proposal != nil {
		proposal = *response.Proposal
	}
	idRaw := proposal.ID
	if len(idRaw) == 0 || string(idRaw) == "null" {
		idRaw = proposal.ProposalID
	}
	idText := strings.Trim(strings.TrimSpace(string(idRaw)), `"`)
	id, err := strconv.ParseUint(idText, 10, 64)
	if err != nil || id == 0 {
		return upgradeGovProposalState{}, fmt.Errorf("decode governance proposal id %q", idText)
	}
	if strings.TrimSpace(proposal.Status) == "" || len(proposal.TotalDeposit) == 0 {
		return upgradeGovProposalState{}, errors.New("governance proposal status and total deposit are required")
	}
	if _, err := formatUpgradeGovCoins(proposal.TotalDeposit); err != nil {
		return upgradeGovProposalState{}, fmt.Errorf("decode governance proposal deposit: %w", err)
	}
	return upgradeGovProposalState{
		ID:           id,
		Status:       proposal.Status,
		TotalDeposit: proposal.TotalDeposit,
		Expedited:    proposal.Expedited,
		Title:        proposal.Title,
		Summary:      proposal.Summary,
	}, nil
}

func formatUpgradeGovCoins(coins []upgradeGovCoin) (string, error) {
	if len(coins) == 0 {
		return "", errors.New("governance coin list is empty")
	}
	formatted := make([]string, len(coins))
	for index, coin := range coins {
		amount, ok := new(big.Int).SetString(coin.Amount, 10)
		if strings.TrimSpace(coin.Denom) == "" || !ok || amount.Sign() <= 0 {
			return "", fmt.Errorf("invalid governance coin %+v", coin)
		}
		formatted[index] = amount.String() + coin.Denom
	}
	return strings.Join(formatted, ","), nil
}

func upgradeGovCoinsEqual(left, right []upgradeGovCoin) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func assertUpgradeGovMigration(
	t *testing.T,
	before upgradeGovCheckpoint,
	after upgradeGovCheckpoint,
) {
	t.Helper()
	require.Equal(t, before.ProposalID, after.ProposalID)
	require.Equal(t, before.Tally, after.Tally)
	require.Equal(t, before.Params.MinDeposit, after.Params.MinDeposit)
	require.Equal(t, before.Params.MaxDepositPeriod, after.Params.MaxDepositPeriod)
	require.Equal(t, before.Params.VotingPeriod, after.Params.VotingPeriod)
	require.Equal(t, before.Params.Quorum, after.Params.Quorum)
	require.Equal(t, before.Params.Threshold, after.Params.Threshold)
	require.Equal(t, before.Params.VetoThreshold, after.Params.VetoThreshold)
	require.Equal(t, multiplyUpgradeGovCoins(t, before.Params.MinDeposit, 5), after.Params.ExpeditedMinDeposit)
}

func multiplyUpgradeGovCoins(t *testing.T, coins []upgradeGovCoin, multiplier int64) []upgradeGovCoin {
	t.Helper()
	require.Positive(t, multiplier)
	result := make([]upgradeGovCoin, len(coins))
	for index, coin := range coins {
		amount, ok := new(big.Int).SetString(coin.Amount, 10)
		require.True(t, ok, "invalid governance deposit amount %q", coin.Amount)
		result[index] = upgradeGovCoin{
			Denom:  coin.Denom,
			Amount: amount.Mul(amount, big.NewInt(multiplier)).String(),
		}
	}
	return result
}

func decodeUpgradeGovParams(raw []byte) (upgradeGovParams, error) {
	var response struct {
		Params        upgradeGovParams `json:"params"`
		DepositParams struct {
			MinDeposit       []upgradeGovCoin `json:"min_deposit"`
			MaxDepositPeriod string           `json:"max_deposit_period"`
		} `json:"deposit_params"`
		VotingParams struct {
			VotingPeriod string `json:"voting_period"`
		} `json:"voting_params"`
		TallyParams struct {
			Quorum        string `json:"quorum"`
			Threshold     string `json:"threshold"`
			VetoThreshold string `json:"veto_threshold"`
		} `json:"tally_params"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return upgradeGovParams{}, err
	}
	params := response.Params
	if len(params.MinDeposit) == 0 {
		params.MinDeposit = response.DepositParams.MinDeposit
		params.MaxDepositPeriod = response.DepositParams.MaxDepositPeriod
		params.VotingPeriod = response.VotingParams.VotingPeriod
		params.Quorum = response.TallyParams.Quorum
		params.Threshold = response.TallyParams.Threshold
		params.VetoThreshold = response.TallyParams.VetoThreshold
	}
	if len(params.MinDeposit) == 0 {
		return upgradeGovParams{}, errors.New("governance params have no minimum deposit")
	}
	for _, coin := range params.MinDeposit {
		if strings.TrimSpace(coin.Denom) == "" || strings.TrimSpace(coin.Amount) == "" {
			return upgradeGovParams{}, errors.New("governance minimum deposit coin is incomplete")
		}
	}
	if params.VotingPeriod == "" || params.Quorum == "" || params.Threshold == "" || params.VetoThreshold == "" {
		return upgradeGovParams{}, errors.New("governance params are incomplete")
	}
	return params, nil
}

func decodeUpgradeGovTally(raw []byte) (upgradeGovTally, error) {
	var response struct {
		TallyResult *struct {
			YesCount        string `json:"yes_count"`
			AbstainCount    string `json:"abstain_count"`
			NoCount         string `json:"no_count"`
			NoWithVetoCount string `json:"no_with_veto_count"`
		} `json:"tally"`
		YesCount        string `json:"yes_count"`
		AbstainCount    string `json:"abstain_count"`
		NoCount         string `json:"no_count"`
		NoWithVetoCount string `json:"no_with_veto_count"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return upgradeGovTally{}, err
	}
	if response.TallyResult != nil {
		response.YesCount = response.TallyResult.YesCount
		response.AbstainCount = response.TallyResult.AbstainCount
		response.NoCount = response.TallyResult.NoCount
		response.NoWithVetoCount = response.TallyResult.NoWithVetoCount
	}
	if response.YesCount == "" || response.AbstainCount == "" || response.NoCount == "" || response.NoWithVetoCount == "" {
		return upgradeGovTally{}, errors.New("governance tally response is incomplete")
	}
	return upgradeGovTally{
		Yes:        response.YesCount,
		Abstain:    response.AbstainCount,
		No:         response.NoCount,
		NoWithVeto: response.NoWithVetoCount,
	}, nil
}

func TestDecodeUpgradeGovParamsSupportsV047AndV050Shapes(t *testing.T) {
	legacy, err := decodeUpgradeGovParams([]byte(`{
		"deposit_params":{"min_deposit":[{"denom":"umed","amount":"1"}],"max_deposit_period":"30s"},
		"voting_params":{"voting_period":"30s"},
		"tally_params":{"quorum":"0.334","threshold":"0.5","veto_threshold":"0.334"}
	}`))
	require.NoError(t, err)
	require.Equal(t, "1", legacy.MinDeposit[0].Amount)
	current, err := decodeUpgradeGovParams([]byte(`{
		"params":{"min_deposit":[{"denom":"umed","amount":"1"}],"max_deposit_period":"30s","voting_period":"30s","quorum":"0.334","threshold":"0.5","veto_threshold":"0.334","expedited_min_deposit":[{"denom":"umed","amount":"5"}]}
	}`))
	require.NoError(t, err)
	require.Equal(t, "5", current.ExpeditedMinDeposit[0].Amount)
}

func TestDecodeUpgradeGovTallySupportsDirectAndWrappedShapes(t *testing.T) {
	for _, raw := range []string{
		`{"yes_count":"4","abstain_count":"0","no_count":"0","no_with_veto_count":"0"}`,
		`{"tally":{"yes_count":"4","abstain_count":"0","no_count":"0","no_with_veto_count":"0"}}`,
	} {
		tally, err := decodeUpgradeGovTally([]byte(raw))
		require.NoError(t, err)
		require.Equal(t, "4", tally.Yes)
	}
}

func TestDecodeUpgradeGovProposalRequiresExpeditedEvidenceFields(t *testing.T) {
	t.Parallel()
	proposal, err := decodeUpgradeGovProposal([]byte(`{
		"proposal":{
			"id":"9",
			"status":"PROPOSAL_STATUS_PASSED",
			"total_deposit":[{"denom":"umed","amount":"5"}],
			"expedited":true,
			"title":"title",
			"summary":"summary"
		}
	}`))
	require.NoError(t, err)
	require.Equal(t, uint64(9), proposal.ID)
	require.True(t, proposal.Expedited)
	require.Equal(t, []upgradeGovCoin{{Denom: "umed", Amount: "5"}}, proposal.TotalDeposit)
	deposit, err := formatUpgradeGovCoins(proposal.TotalDeposit)
	require.NoError(t, err)
	require.Equal(t, "5umed", deposit)

	_, err = decodeUpgradeGovProposal([]byte(`{"proposal":{"id":"9","status":"PASSED","total_deposit":[]}}`))
	require.ErrorContains(t, err, "status and total deposit")
}
