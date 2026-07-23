package keeper

import (
	"testing"

	upstreamnft "cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

func TestMaximumPayloadPageResponseSize(t *testing.T) {
	t.Run("standard classes", func(t *testing.T) {
		fixture := newMaximumPayloadClassesFixture(t)
		response, err := fixture.standard.Classes(
			fixture.goCtx,
			&upstreamnft.QueryClassesRequest{
				Pagination: &query.PageRequest{Limit: maximumQueryPageLimit},
			},
		)
		require.NoError(t, err)
		require.Len(t, response.Classes, int(maximumQueryPageLimit))
		require.NotEmpty(t, response.Pagination.NextKey)
		encoded, err := response.Marshal()
		require.NoError(t, err)
		require.Len(t, encoded, 166237)
	})

	fixture := newMaximumPayloadNFTFixture(t)
	expectedSizes := map[string]struct {
		standard int
		panacea  int
	}{
		"class":       {standard: 159568, panacea: 174668},
		"owner":       {standard: 159702, panacea: 174802},
		"class_owner": {standard: 159568, panacea: 174668},
	}
	for _, filter := range maximumPageNFTFilters() {
		expected, exists := expectedSizes[filter.name]
		require.True(t, exists, "missing response-size golden for %s", filter.name)

		t.Run("standard nfts/"+filter.name, func(t *testing.T) {
			response, err := fixture.standard.NFTs(
				fixture.goCtx,
				filter.standardRequest(fixture),
			)
			require.NoError(t, err)
			require.Len(t, response.Nfts, int(maximumQueryPageLimit))
			require.NotEmpty(t, response.Pagination.NextKey)
			encoded, err := response.Marshal()
			require.NoError(t, err)
			require.Len(t, encoded, expected.standard)
		})

		t.Run("panacea nft records/"+filter.name, func(t *testing.T) {
			response, err := fixture.panacea.NFTRecords(
				fixture.goCtx,
				filter.panaceaRequest(fixture),
			)
			require.NoError(t, err)
			require.Len(t, response.NftRecords, int(maximumQueryPageLimit))
			require.NotEmpty(t, response.Pagination.NextKey)
			encoded, err := response.Marshal()
			require.NoError(t, err)
			require.Len(t, encoded, expected.panacea)
		})
	}
}
