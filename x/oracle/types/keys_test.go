package types_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestSplitOracleRegistrationVoteQueueKey(t *testing.T) {
	uniqueID := "uniqueID"
	now := time.Now().UTC()
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	key := types.GetOracleRegistrationQueueKey(uniqueID, addr, now)

	votingEndTime, splitUniqueID, splitAddr, err := types.SplitOracleRegistrationVoteQueueKey(key)
	require.NoError(t, err)
	require.Equal(t, now, votingEndTime)
	require.Equal(t, uniqueID, splitUniqueID)
	require.Equal(t, addr, splitAddr)
}

func TestSplitOracleRegistrationVoteQueueKeyTimeZero(t *testing.T) {
	uniqueID := "uniqueID"
	now := time.Time{}
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	key := types.GetOracleRegistrationQueueKey(uniqueID, addr, now)

	votingEndTime, splitUniqueID, splitAddr, err := types.SplitOracleRegistrationVoteQueueKey(key)
	require.NoError(t, err)
	require.Equal(t, now, votingEndTime)
	require.True(t, votingEndTime.IsZero())
	require.Equal(t, uniqueID, splitUniqueID)
	require.Equal(t, addr, splitAddr)
}
