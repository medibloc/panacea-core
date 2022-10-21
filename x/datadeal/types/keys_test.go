package types_test

import (
	"testing"
	"time"

	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/stretchr/testify/require"
)

func TestSplitDataVerificationVoteQueueKey(t *testing.T) {
	dealID := uint64(1)
	dataHash := "dataHash"
	now := time.Now()

	key := types.GetDataVerificationQueueKey(dataHash, dealID, now)

	splitDealID, splitCID := types.SplitDataVerificationQueueKey(key)

	require.Equal(t, dealID, splitDealID)
	require.Equal(t, dataHash, splitCID)
}

func TestSpliDataDeliveryVoteQueueKey(t *testing.T) {
	dealID := uint64(1)
	dataHash := "dataHash"
	now := time.Now()

	key := types.GetDataDeliveryQueueKey(dealID, dataHash, now)

	splitDealID, splitCID := types.SplitDataDeliveryQueueKey(key)

	require.Equal(t, dealID, splitDealID)
	require.Equal(t, dataHash, splitCID)
}
