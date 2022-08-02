package types_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSplitOracleRegistrationVoteQueueKey(t *testing.T) {
	uniqueID := "uniqueID"
	now := time.Now()
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	key := types.GetOracleRegistrationVoteQueueKey(uniqueID, addr, now)

	splitUniqueID, splitAddr := types.SplitOracleRegistrationVoteQueueKey(key)

	require.Equal(t, uniqueID, splitUniqueID)
	require.Equal(t, addr, splitAddr)
}
