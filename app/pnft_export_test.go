package app_test

import (
	"encoding/json"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	panaceaapp "github.com/medibloc/panacea-core/v2/app"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
)

func TestLegacyPNFTGovernanceProposalExportsAndRedecodes(t *testing.T) {
	testApp, proposer, ctx := newPNFTProposalTestApp(t)
	legacyMsg := newLegacyProposalPNFTMsg(authtypes.NewModuleAddress(govtypes.ModuleName).String())

	proposal, err := testApp.GovKeeper.SubmitProposal(
		ctx,
		[]sdk.Msg{legacyMsg},
		"ipfs://legacy-pnft-proposal",
		"Legacy PNFT proposal",
		"Preserve a pending legacy PNFT message during export",
		proposer,
		false,
	)
	require.NoError(t, err)
	require.NoError(t, testApp.GovKeeper.ActivateVotingPeriod(ctx, proposal))
	storedProposal, err := testApp.GovKeeper.Proposals.Get(ctx, proposal.Id)
	require.NoError(t, err)
	require.Equal(t, govv1.StatusVotingPeriod, storedProposal.Status)

	_, err = testApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 2,
		Time:   ctx.BlockTime().Add(time.Second),
	})
	require.NoError(t, err)
	_, err = testApp.Commit()
	require.NoError(t, err)

	exported, err := testApp.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)

	var appState panaceaapp.GenesisState
	require.NoError(t, json.Unmarshal(exported.AppState, &appState))
	govStateJSON, ok := appState[govtypes.ModuleName]
	require.True(t, ok)
	require.Contains(t, string(govStateJSON), "/panacea.pnft.v2.MsgMintPNFTRequest")

	var govState govv1.GenesisState
	require.NoError(t, testApp.AppCodec().UnmarshalJSON(govStateJSON, &govState))
	require.Len(t, govState.Proposals, 1)

	exportedProposal := govState.Proposals[0]
	require.Equal(t, storedProposal.Id, exportedProposal.Id)
	require.Equal(t, govv1.StatusVotingPeriod, exportedProposal.Status)
	require.Equal(t, storedProposal.Metadata, exportedProposal.Metadata)
	require.Equal(t, storedProposal.Title, exportedProposal.Title)
	require.Equal(t, storedProposal.Summary, exportedProposal.Summary)
	require.Equal(t, proposer.String(), exportedProposal.Proposer)
	require.Equal(t, storedProposal.SubmitTime, exportedProposal.SubmitTime)
	require.Equal(t, storedProposal.DepositEndTime, exportedProposal.DepositEndTime)
	require.Equal(t, storedProposal.VotingStartTime, exportedProposal.VotingStartTime)
	require.Equal(t, storedProposal.VotingEndTime, exportedProposal.VotingEndTime)
	require.Len(t, exportedProposal.Messages, 1)
	require.Equal(t, "/panacea.pnft.v2.MsgMintPNFTRequest", exportedProposal.Messages[0].TypeUrl)

	decodedMsg, ok := exportedProposal.Messages[0].GetCachedValue().(*pnfttypes.MsgMintPNFTRequest)
	require.True(t, ok)
	require.True(t, proto.Equal(legacyMsg, decodedMsg))
}

func TestLegacyPNFTGroupProposalExportsAndRedecodes(t *testing.T) {
	testApp, member, ctx := newPNFTProposalTestApp(t)
	createGroup := &group.MsgCreateGroupWithPolicy{
		Admin: member.String(),
		Members: []group.MemberRequest{{
			Address:  member.String(),
			Weight:   "1",
			Metadata: "legacy-member",
		}},
		GroupMetadata:       "ipfs://legacy-group",
		GroupPolicyMetadata: "ipfs://legacy-group-policy",
	}
	require.NoError(t, createGroup.SetDecisionPolicy(
		group.NewThresholdDecisionPolicy("1", time.Hour, 0),
	))
	created, err := testApp.GroupKeeper.CreateGroupWithPolicy(ctx, createGroup)
	require.NoError(t, err)

	legacyMsg := newLegacyProposalPNFTMsg(created.GroupPolicyAddress)
	submit, err := group.NewMsgSubmitProposal(
		created.GroupPolicyAddress,
		[]string{member.String()},
		[]sdk.Msg{legacyMsg},
		"ipfs://legacy-pnft-group-proposal",
		group.Exec_EXEC_UNSPECIFIED,
		"Legacy PNFT group proposal",
		"Preserve a pending legacy PNFT group message during export",
	)
	require.NoError(t, err)
	submitted, err := testApp.GroupKeeper.SubmitProposal(ctx, submit)
	require.NoError(t, err)
	storedResponse, err := testApp.GroupKeeper.Proposal(ctx, &group.QueryProposalRequest{
		ProposalId: submitted.ProposalId,
	})
	require.NoError(t, err)
	storedProposal := storedResponse.Proposal
	require.Equal(t, group.PROPOSAL_STATUS_SUBMITTED, storedProposal.Status)
	require.Equal(t, group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN, storedProposal.ExecutorResult)

	_, err = testApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 2,
		Time:   ctx.BlockTime().Add(time.Second),
	})
	require.NoError(t, err)
	_, err = testApp.Commit()
	require.NoError(t, err)

	exported, err := testApp.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)

	var appState panaceaapp.GenesisState
	require.NoError(t, json.Unmarshal(exported.AppState, &appState))
	groupStateJSON, ok := appState[group.ModuleName]
	require.True(t, ok)
	require.Contains(t, string(groupStateJSON), "/panacea.pnft.v2.MsgMintPNFTRequest")

	var groupState group.GenesisState
	require.NoError(t, testApp.AppCodec().UnmarshalJSON(groupStateJSON, &groupState))
	require.Len(t, groupState.Groups, 1)
	require.Len(t, groupState.GroupMembers, 1)
	require.Len(t, groupState.GroupPolicies, 1)
	require.Len(t, groupState.Proposals, 1)
	require.Empty(t, groupState.Votes)

	require.Equal(t, created.GroupId, groupState.Groups[0].Id)
	require.Equal(t, createGroup.GroupMetadata, groupState.Groups[0].Metadata)
	require.Equal(t, created.GroupId, groupState.GroupMembers[0].GroupId)
	require.Equal(t, member.String(), groupState.GroupMembers[0].Member.Address)
	require.Equal(t, created.GroupPolicyAddress, groupState.GroupPolicies[0].Address)
	require.Equal(t, created.GroupId, groupState.GroupPolicies[0].GroupId)
	require.Equal(t, createGroup.GroupPolicyMetadata, groupState.GroupPolicies[0].Metadata)

	exportedProposal := groupState.Proposals[0]
	require.Equal(t, storedProposal.Id, exportedProposal.Id)
	require.Equal(t, storedProposal.GroupPolicyAddress, exportedProposal.GroupPolicyAddress)
	require.Equal(t, storedProposal.Metadata, exportedProposal.Metadata)
	require.Equal(t, storedProposal.Proposers, exportedProposal.Proposers)
	require.Equal(t, storedProposal.SubmitTime, exportedProposal.SubmitTime)
	require.Equal(t, storedProposal.GroupVersion, exportedProposal.GroupVersion)
	require.Equal(t, storedProposal.GroupPolicyVersion, exportedProposal.GroupPolicyVersion)
	require.Equal(t, group.PROPOSAL_STATUS_SUBMITTED, exportedProposal.Status)
	require.Equal(t, storedProposal.FinalTallyResult, exportedProposal.FinalTallyResult)
	require.Equal(t, storedProposal.VotingPeriodEnd, exportedProposal.VotingPeriodEnd)
	require.Equal(t, group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN, exportedProposal.ExecutorResult)
	require.Equal(t, storedProposal.Title, exportedProposal.Title)
	require.Equal(t, storedProposal.Summary, exportedProposal.Summary)
	require.Len(t, exportedProposal.Messages, 1)
	require.Equal(t, "/panacea.pnft.v2.MsgMintPNFTRequest", exportedProposal.Messages[0].TypeUrl)

	decodedMsg, ok := exportedProposal.Messages[0].GetCachedValue().(*pnfttypes.MsgMintPNFTRequest)
	require.True(t, ok)
	require.True(t, proto.Equal(legacyMsg, decodedMsg))
}
