package types_test

import (
	"container/heap"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func makeSampleVotes(count int, uniqueID string) []*types.OracleRegistrationVote {
	votes := make([]*types.OracleRegistrationVote, 0)
	for i := 0; i < count; i++ {
		votes = append(votes, &types.OracleRegistrationVote{
			UniqueId:            uniqueID,
			VoterAddress:        fmt.Sprintf("voter%v", i),
			VotingTargetAddress: "newVoter",
		})
	}

	return votes
}

func makeSampleOracleValidatorInfoMap(votes []*types.OracleRegistrationVote) map[string]*types.OracleValidatorInfo {
	infos := make(map[string]*types.OracleValidatorInfo)
	for _, vote := range votes {

		info := &types.OracleValidatorInfo{
			Address:         vote.VoterAddress,
			OracleActivated: true,
			ValidatorJailed: false,
		}
		infos[info.Address] = info
	}

	return infos
}

func TestTallyResultAllValid(t *testing.T) {
	uniqueID := "unique"
	consensusValue := []byte("encPriv1")

	votes := makeSampleVotes(3, uniqueID)
	votes[0].VoteOption = types.VOTE_OPTION_VALID
	votes[0].EncryptedOraclePrivKey = consensusValue
	votes[1].VoteOption = types.VOTE_OPTION_VALID
	votes[1].EncryptedOraclePrivKey = consensusValue
	votes[2].VoteOption = types.VOTE_OPTION_VALID
	votes[2].EncryptedOraclePrivKey = consensusValue

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(30)
	infos[votes[1].VoterAddress].BondedTokens = sdk.NewInt(20)
	infos[votes[2].VoterAddress].BondedTokens = sdk.NewInt(10)

	tally := types.NewTally()
	tally.OracleValidatorInfos = infos
	err := tally.Add(votes[0])
	require.NoError(t, err)
	err = tally.Add(votes[1])
	require.NoError(t, err)
	err = tally.Add(votes[2])
	require.NoError(t, err)

	tallyResult := tally.CalculateTallyResult(sdk.NewDec(1).Quo(sdk.NewDec(3)))
	require.Equal(t, sdk.NewInt(60), tallyResult.Yes)
	require.True(t, tallyResult.No.IsZero())
	require.True(t, tallyResult.InvalidYes.IsZero())
	require.Equal(t, consensusValue, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(60), tallyResult.Total)
}

func TestTallyResultAllInValid(t *testing.T) {
	uniqueID := "unique"

	votes := makeSampleVotes(3, uniqueID)
	votes[0].VoteOption = types.VOTE_OPTION_INVALID
	votes[1].VoteOption = types.VOTE_OPTION_INVALID
	votes[2].VoteOption = types.VOTE_OPTION_INVALID

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(30)
	infos[votes[1].VoterAddress].BondedTokens = sdk.NewInt(20)
	infos[votes[2].VoterAddress].BondedTokens = sdk.NewInt(10)

	tally := types.NewTally()
	tally.OracleValidatorInfos = infos
	err := tally.Add(votes[0])
	require.NoError(t, err)
	err = tally.Add(votes[1])
	require.NoError(t, err)
	err = tally.Add(votes[2])
	require.NoError(t, err)

	tallyResult := tally.CalculateTallyResult(sdk.NewDec(1).Quo(sdk.NewDec(3)))
	require.True(t, tallyResult.Yes.IsZero())
	require.Equal(t, sdk.NewInt(60), tallyResult.No)
	require.True(t, tallyResult.InvalidYes.IsZero())
	require.Nil(t, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(60), tallyResult.Total)
}

func TestTallyResultDifferentConsensusValue(t *testing.T) {
	uniqueID := "unique"

	votes := makeSampleVotes(3, uniqueID)
	votes[0].VoteOption = types.VOTE_OPTION_VALID
	votes[0].EncryptedOraclePrivKey = []byte("encPriv1")
	votes[1].VoteOption = types.VOTE_OPTION_VALID
	votes[1].EncryptedOraclePrivKey = []byte("encPriv2")
	votes[2].VoteOption = types.VOTE_OPTION_VALID
	votes[2].EncryptedOraclePrivKey = []byte("encPriv3")

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(30)
	infos[votes[1].VoterAddress].BondedTokens = sdk.NewInt(20)
	infos[votes[2].VoterAddress].BondedTokens = sdk.NewInt(10)

	tally := types.NewTally()
	tally.OracleValidatorInfos = infos
	err := tally.Add(votes[0])
	require.NoError(t, err)
	err = tally.Add(votes[1])
	require.NoError(t, err)
	err = tally.Add(votes[2])
	require.NoError(t, err)

	// ConsensusValue with the highest number of Yes votes selected.
	// All others counted as invalid votes
	tallyResult := tally.CalculateTallyResult(sdk.NewDec(1).Quo(sdk.NewDec(3)))
	require.Equal(t, sdk.NewInt(30), tallyResult.Yes)
	require.True(t, tallyResult.No.IsZero())
	require.Equal(t, sdk.NewInt(30), tallyResult.InvalidYes)
	require.Equal(t, votes[0].EncryptedOraclePrivKey, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(60), tallyResult.Total)
}

func TestTallyResultLessThenQuorum(t *testing.T) {
	uniqueID := "unique"

	votes := makeSampleVotes(3, uniqueID)
	votes[2].VoteOption = types.VOTE_OPTION_VALID
	votes[2].EncryptedOraclePrivKey = []byte("encPriv1")

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(30)
	infos[votes[1].VoterAddress].BondedTokens = sdk.NewInt(20)
	infos[votes[2].VoterAddress].BondedTokens = sdk.NewInt(10)

	tally := types.NewTally()
	tally.OracleValidatorInfos = infos
	err := tally.Add(votes[2])
	require.NoError(t, err)

	tallyResult := tally.CalculateTallyResult(sdk.NewDec(1).Quo(sdk.NewDec(3)))
	require.Equal(t, sdk.NewInt(10), tallyResult.Yes)
	require.True(t, tallyResult.No.IsZero())
	require.True(t, tallyResult.InvalidYes.IsZero())
	require.Nil(t, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(60), tallyResult.Total)
}

func TestTallyResultNumberOfAllVotes(t *testing.T) {
	uniqueID := "unique"

	votes := makeSampleVotes(4, uniqueID)
	votes[0].VoteOption = types.VOTE_OPTION_VALID
	votes[0].EncryptedOraclePrivKey = []byte("encPriv1")
	votes[1].VoteOption = types.VOTE_OPTION_INVALID
	votes[2].VoteOption = types.VOTE_OPTION_VALID
	votes[2].EncryptedOraclePrivKey = []byte("encPriv3")

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(30)
	infos[votes[1].VoterAddress].BondedTokens = sdk.NewInt(20)
	infos[votes[2].VoterAddress].BondedTokens = sdk.NewInt(10)
	infos[votes[3].VoterAddress].BondedTokens = sdk.NewInt(5)

	tally := types.NewTally()
	tally.OracleValidatorInfos = infos
	err := tally.Add(votes[0])
	require.NoError(t, err)
	err = tally.Add(votes[1])
	require.NoError(t, err)
	err = tally.Add(votes[2])
	require.NoError(t, err)

	// ConsensusValue with the highest number of Yes votes selected.
	// All others counted as invalid votes
	tallyResult := tally.CalculateTallyResult(sdk.NewDec(1).Quo(sdk.NewDec(3)))
	require.Equal(t, sdk.NewInt(30), tallyResult.Yes)
	require.Equal(t, sdk.NewInt(20), tallyResult.No)
	require.Equal(t, sdk.NewInt(10), tallyResult.InvalidYes)
	require.Equal(t, votes[0].EncryptedOraclePrivKey, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(65), tallyResult.Total)
}

func TestConsensusTallyMaxHeap(t *testing.T) {
	tally1 := &types.ConsensusTally{
		ConsensusKey: []byte("key1"),
		VotingAmount: sdk.NewInt(1000),
	}
	tally2 := &types.ConsensusTally{
		ConsensusKey: []byte("key2"),
		VotingAmount: sdk.NewInt(500),
	}
	tally3 := &types.ConsensusTally{
		ConsensusKey: []byte("key3"),
		VotingAmount: sdk.NewInt(400),
	}
	tally4 := &types.ConsensusTally{
		ConsensusKey: []byte("key4"),
		VotingAmount: sdk.NewInt(2000),
	}
	tally5 := &types.ConsensusTally{
		ConsensusKey: []byte("key5"),
		VotingAmount: sdk.NewInt(1001),
	}
	items := []*types.ConsensusTally{tally1, tally2, tally3, tally4, tally5}

	consensusHeap := types.NewConsensusTallyMaxHeap()
	for _, item := range items {
		heap.Push(&consensusHeap, item)
	}

	require.Equal(t, tally4, heap.Pop(&consensusHeap).(*types.ConsensusTally))
	require.Equal(t, tally5, heap.Pop(&consensusHeap).(*types.ConsensusTally))
	require.Equal(t, tally1, heap.Pop(&consensusHeap).(*types.ConsensusTally))
	require.Equal(t, tally2, heap.Pop(&consensusHeap).(*types.ConsensusTally))
	require.Equal(t, tally3, heap.Pop(&consensusHeap).(*types.ConsensusTally))
}