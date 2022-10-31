package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

func TestOracleRegistrationVoteValidateBasic(t *testing.T) {
	accAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	vote := OracleRegistrationVote{
		UniqueId:               "uniqueID",
		VoterAddress:           accAddr.String(),
		VotingTargetAddress:    accAddr.String(),
		VoteOption:             VOTE_OPTION_YES,
		EncryptedOraclePrivKey: []byte("encrypted"),
	}

	require.NoError(t, vote.ValidateBasic())

	vote.UniqueId = ""
	require.Error(t, vote.ValidateBasic(), "uniqueID is empty")

	vote.UniqueId = "uniqueID"
	vote.VoterAddress = ""
	require.Error(t, vote.ValidateBasic(), "voterAddress is invalid.")

	vote.VoterAddress = accAddr.String()
	vote.VotingTargetAddress = ""
	require.Error(t, vote.ValidateBasic(), "votingTargetAddress is invalid.")

	vote.VotingTargetAddress = accAddr.String()
	vote.VoteOption = VOTE_OPTION_NO
	vote.EncryptedOraclePrivKey = nil
	require.NoError(t, vote.ValidateBasic())

	vote.VotingTargetAddress = accAddr.String()
	vote.VoteOption = VOTE_OPTION_UNSPECIFIED
	require.Error(t, vote.ValidateBasic(), "voteOption is invalid")
}
