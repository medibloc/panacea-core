package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestSplitOracleUpgradeKey(t *testing.T) {
	uniqueID := "uniqueID"
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	key := types.GetOracleUpgradeKey(uniqueID, addr)
	splitUniqueID, splitAddr := types.SplitOracleUpgradeKey(key)
	require.Equal(t, uniqueID, splitUniqueID)
	require.Equal(t, addr, splitAddr)
}
