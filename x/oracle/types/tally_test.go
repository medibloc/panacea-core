package types_test

import (
	"container/heap"
	"encoding/base64"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
)

var threshold = sdk.NewDec(2).Quo(sdk.NewDec(3))

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
	votes[0].VoteOption = types.VOTE_OPTION_YES
	votes[0].EncryptedOraclePrivKey = consensusValue
	votes[1].VoteOption = types.VOTE_OPTION_YES
	votes[1].EncryptedOraclePrivKey = consensusValue
	votes[2].VoteOption = types.VOTE_OPTION_YES
	votes[2].EncryptedOraclePrivKey = consensusValue

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(70)
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

	tallyResult := tally.CalculateTallyResult(threshold)
	require.Equal(t, sdk.NewInt(100), tallyResult.Yes)
	require.True(t, tallyResult.No.IsZero())
	require.Equal(t, 0, len(tallyResult.InvalidYes))
	require.Equal(t, consensusValue, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(100), tallyResult.Total)
}

func TestTallyResultAllInValid(t *testing.T) {
	uniqueID := "unique"

	votes := makeSampleVotes(3, uniqueID)
	votes[0].VoteOption = types.VOTE_OPTION_NO
	votes[1].VoteOption = types.VOTE_OPTION_NO
	votes[2].VoteOption = types.VOTE_OPTION_NO

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(70)
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

	tallyResult := tally.CalculateTallyResult(threshold)
	require.True(t, tallyResult.Yes.IsZero())
	require.Equal(t, sdk.NewInt(100), tallyResult.No)
	require.Equal(t, 0, len(tallyResult.InvalidYes))
	require.Nil(t, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(100), tallyResult.Total)
}

func TestTallyResultDifferentConsensusValueSuccessConsensus(t *testing.T) {
	uniqueID := "unique"
	consensusValue := []byte("encPriv1")
	consensusValue2 := []byte("encPriv2")
	consensusValue3 := []byte("encPriv3")

	votes := makeSampleVotes(3, uniqueID)
	votes[0].VoteOption = types.VOTE_OPTION_YES
	votes[0].EncryptedOraclePrivKey = consensusValue
	votes[1].VoteOption = types.VOTE_OPTION_YES
	votes[1].EncryptedOraclePrivKey = consensusValue2
	votes[2].VoteOption = types.VOTE_OPTION_YES
	votes[2].EncryptedOraclePrivKey = consensusValue3

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(70)
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

	// If the consensusValue exceeds the threshold, the consensus is successful.
	tallyResult := tally.CalculateTallyResult(threshold)
	require.Equal(t, sdk.NewInt(70), tallyResult.Yes)
	require.True(t, tallyResult.No.IsZero())
	require.Equal(t, 2, len(tallyResult.InvalidYes))
	require.Equal(t, consensusValue2, tallyResult.InvalidYes[0].ConsensusValue)
	require.Equal(t, sdk.NewInt(20), tallyResult.InvalidYes[0].VotingAmount)
	require.Equal(t, consensusValue3, tallyResult.InvalidYes[1].ConsensusValue)
	require.Equal(t, sdk.NewInt(10), tallyResult.InvalidYes[1].VotingAmount)
	require.Equal(t, consensusValue, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(100), tallyResult.Total)
}

func TestTallyResultDifferentConsensusValueFailedConsensus(t *testing.T) {
	uniqueID := "unique"
	consensusValue := []byte("encPriv1")
	consensusValue2 := []byte("encPriv2")
	consensusValue3 := []byte("encPriv3")

	votes := makeSampleVotes(3, uniqueID)
	votes[0].VoteOption = types.VOTE_OPTION_YES
	votes[0].EncryptedOraclePrivKey = consensusValue
	votes[1].VoteOption = types.VOTE_OPTION_YES
	votes[1].EncryptedOraclePrivKey = consensusValue2
	votes[2].VoteOption = types.VOTE_OPTION_YES
	votes[2].EncryptedOraclePrivKey = consensusValue3

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(50)
	infos[votes[1].VoterAddress].BondedTokens = sdk.NewInt(30)
	infos[votes[2].VoterAddress].BondedTokens = sdk.NewInt(20)

	tally := types.NewTally()
	tally.OracleValidatorInfos = infos
	err := tally.Add(votes[0])
	require.NoError(t, err)
	err = tally.Add(votes[1])
	require.NoError(t, err)
	err = tally.Add(votes[2])
	require.NoError(t, err)

	// If the consensusValue exceeds the threshold, the consensus is successful.
	tallyResult := tally.CalculateTallyResult(threshold)
	require.True(t, tallyResult.Yes.IsZero())
	require.True(t, tallyResult.No.IsZero())
	require.Equal(t, 3, len(tallyResult.InvalidYes))
	require.Equal(t, consensusValue, tallyResult.InvalidYes[0].ConsensusValue)
	require.Equal(t, sdk.NewInt(50), tallyResult.InvalidYes[0].VotingAmount)
	require.Equal(t, consensusValue2, tallyResult.InvalidYes[1].ConsensusValue)
	require.Equal(t, sdk.NewInt(30), tallyResult.InvalidYes[1].VotingAmount)
	require.Equal(t, consensusValue3, tallyResult.InvalidYes[2].ConsensusValue)
	require.Equal(t, sdk.NewInt(20), tallyResult.InvalidYes[2].VotingAmount)
	require.Nil(t, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(100), tallyResult.Total)
}

func TestTallyResultLessThenThreshold(t *testing.T) {
	uniqueID := "unique"
	consensusValue := []byte("encPriv1")

	votes := makeSampleVotes(3, uniqueID)
	votes[2].VoteOption = types.VOTE_OPTION_YES
	votes[2].EncryptedOraclePrivKey = consensusValue

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(30)
	infos[votes[1].VoterAddress].BondedTokens = sdk.NewInt(20)
	infos[votes[2].VoterAddress].BondedTokens = sdk.NewInt(10)

	tally := types.NewTally()
	tally.OracleValidatorInfos = infos
	err := tally.Add(votes[2])
	require.NoError(t, err)

	tallyResult := tally.CalculateTallyResult(threshold)
	require.True(t, tallyResult.Yes.IsZero())
	require.True(t, tallyResult.No.IsZero())
	require.Equal(t, 1, len(tallyResult.InvalidYes))
	require.Equal(t, consensusValue, tallyResult.InvalidYes[0].ConsensusValue)
	require.Equal(t, sdk.NewInt(10), tallyResult.InvalidYes[0].VotingAmount)
	require.Nil(t, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(60), tallyResult.Total)
}

func TestTallyResultNumberOfAllVotes(t *testing.T) {
	uniqueID := "unique"
	consensusValue := []byte("encPriv1")
	consensusValue2 := []byte("encPriv2")

	votes := makeSampleVotes(4, uniqueID)
	votes[0].VoteOption = types.VOTE_OPTION_YES
	votes[0].EncryptedOraclePrivKey = consensusValue
	votes[1].VoteOption = types.VOTE_OPTION_NO
	votes[2].VoteOption = types.VOTE_OPTION_YES
	votes[2].EncryptedOraclePrivKey = consensusValue2

	infos := makeSampleOracleValidatorInfoMap(votes)
	infos[votes[0].VoterAddress].BondedTokens = sdk.NewInt(70)
	infos[votes[1].VoterAddress].BondedTokens = sdk.NewInt(15)
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
	tallyResult := tally.CalculateTallyResult(threshold)
	require.Equal(t, sdk.NewInt(70), tallyResult.Yes)
	require.Equal(t, sdk.NewInt(15), tallyResult.No)
	require.Equal(t, 1, len(tallyResult.InvalidYes))
	require.Equal(t, consensusValue2, tallyResult.InvalidYes[0].ConsensusValue)
	require.Equal(t, sdk.NewInt(10), tallyResult.InvalidYes[0].VotingAmount)
	require.Equal(t, consensusValue, tallyResult.ConsensusValue)
	require.Equal(t, sdk.NewInt(100), tallyResult.Total)
}

func TestConsensusTallyMaxHeap(t *testing.T) {
	tally1 := &types.ConsensusTally{
		ConsensusValue: []byte("key1"),
		VotingAmount:   sdk.NewInt(1000),
	}
	tally2 := &types.ConsensusTally{
		ConsensusValue: []byte("key2"),
		VotingAmount:   sdk.NewInt(500),
	}
	tally3 := &types.ConsensusTally{
		ConsensusValue: []byte("key3"),
		VotingAmount:   sdk.NewInt(400),
	}
	tally4 := &types.ConsensusTally{
		ConsensusValue: []byte("key4"),
		VotingAmount:   sdk.NewInt(2000),
	}
	tally5 := &types.ConsensusTally{
		ConsensusValue: []byte("key5"),
		VotingAmount:   sdk.NewInt(1001),
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

func Test(t *testing.T) {
	priv, err := btcec.NewPrivateKey(btcec.S256())
	require.NoError(t, err)
	fmt.Println("priv: ", base64.StdEncoding.EncodeToString(priv.Serialize()))
	fmt.Println("pub: ", base64.StdEncoding.EncodeToString(priv.PubKey().SerializeCompressed()))
}
