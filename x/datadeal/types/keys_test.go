package types_test

import (
	"testing"
	"time"

	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/require"
)

func TestSplitOracleRegistrationVoteQueueKey(t *testing.T) {
	dealID := uint64(1)
	dataHash := "dataHash"
	now := time.Now()

	key := types.GetDataDeliveryQueueKey(dealID, dataHash, now)

	splitDealID, splitCID := types.SplitDataDeliveryQueueKey(key)

	require.Equal(t, dealID, splitDealID)
	require.Equal(t, dataHash, splitCID)
}
