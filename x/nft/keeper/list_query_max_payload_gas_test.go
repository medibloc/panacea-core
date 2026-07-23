package keeper

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	upstreamnft "cosmossdk.io/x/nft"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	nfttypes "github.com/medibloc/panacea-core/v2/x/nft/types"
	"github.com/stretchr/testify/require"
)

const (
	// These values mirror the protocol bounds in types/validation.go. A bound
	// change must deliberately update this fixture and its query gas goldens.
	maximumPayloadLocalIDBytes          = 64
	maximumPayloadClassNameBytes        = 128
	maximumPayloadClassSymbolBytes      = 32
	maximumPayloadClassDescriptionBytes = 1024
	maximumPayloadURIBytes              = 256
	maximumPayloadNFTDataBytes          = 1024
)

func TestMaximumPayloadPageQueryGas(t *testing.T) {
	t.Run("standard classes", func(t *testing.T) {
		fixture := newMaximumPayloadClassesFixture(t)
		goCtx, meter := contextWithQueryGasMeter(fixture.goCtx)
		response, err := fixture.standard.Classes(
			goCtx,
			&upstreamnft.QueryClassesRequest{
				Pagination: &query.PageRequest{Limit: maximumQueryPageLimit},
			},
		)
		require.NoError(t, err)
		require.Len(t, response.Classes, int(maximumQueryPageLimit))
		require.NotEmpty(t, response.Pagination.NextKey)
		require.Equal(t, uint64(1546800), meter.GasConsumed())
	})

	fixture := newMaximumPayloadNFTFixture(t)
	expectedGas := map[string]uint64{
		"class":       1763525,
		"owner":       1928454,
		"class_owner": 1928454,
	}
	for _, filter := range maximumPageNFTFilters() {
		t.Run("standard nfts/"+filter.name, func(t *testing.T) {
			goCtx, meter := contextWithQueryGasMeter(fixture.goCtx)
			response, err := fixture.standard.NFTs(
				goCtx,
				filter.standardRequest(fixture),
			)
			require.NoError(t, err)
			require.Len(t, response.Nfts, int(maximumQueryPageLimit))
			require.NotEmpty(t, response.Pagination.NextKey)
			require.Equal(t, expectedGas[filter.name], meter.GasConsumed())
		})

		t.Run("panacea nft records/"+filter.name, func(t *testing.T) {
			goCtx, meter := contextWithQueryGasMeter(fixture.goCtx)
			response, err := fixture.panacea.NFTRecords(
				goCtx,
				filter.panaceaRequest(fixture),
			)
			require.NoError(t, err)
			require.Len(t, response.NftRecords, int(maximumQueryPageLimit))
			require.NotEmpty(t, response.Pagination.NextKey)
			require.Equal(t, expectedGas[filter.name], meter.GasConsumed())
		})
	}
}

func newMaximumPayloadClassesFixture(t testing.TB) maximumPageQueryFixture {
	t.Helper()
	fixture := newKeeperFixture(t, true, true)
	fixture.ctx = fixture.ctx.WithBlockTime(
		time.Date(2026, time.July, 23, 0, 0, 0, 0, time.UTC),
	)
	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{118}, 32))
	creator := fixture.accountAddress(t, creatorAddress)
	require.Len(t, creator, 66)
	fixture.accountKeeper.accounts[string(creatorAddress)] =
		authtypes.NewBaseAccountWithAddress(creatorAddress)

	server := NewMsgServer(fixture.keeper)
	for index := 0; index < maximumPageBenchmarkStateSize; index++ {
		request := maximumPayloadClassRequest(creator, index)
		response, err := server.CreateClass(fixture.ctx, request)
		require.NoError(t, err)
		require.Len(t, response.ClassId, 131)
	}
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return maximumPageQueryFixture{
		goCtx:    fixture.ctx,
		standard: NewStandardQueryServer(fixture.keeper),
		panacea:  NewQueryServer(fixture.keeper),
	}
}

func newMaximumPayloadNFTFixture(t testing.TB) maximumPageQueryFixture {
	t.Helper()
	fixture := newKeeperFixture(t, true, true)
	fixture.ctx = fixture.ctx.WithBlockTime(
		time.Date(2026, time.July, 23, 0, 0, 0, 0, time.UTC),
	)
	creatorAddress := sdk.AccAddress(bytes.Repeat([]byte{119}, 32))
	ownerAddress := sdk.AccAddress(bytes.Repeat([]byte{120}, 32))
	creator := fixture.accountAddress(t, creatorAddress)
	owner := fixture.accountAddress(t, ownerAddress)
	require.Len(t, creator, 66)
	require.Len(t, owner, 66)
	for _, address := range []sdk.AccAddress{creatorAddress, ownerAddress} {
		fixture.accountKeeper.accounts[string(address)] =
			authtypes.NewBaseAccountWithAddress(address)
	}

	classRequest := maximumPayloadClassRequest(creator, 0)
	classRequest.MaxSupply = 0
	classResponse, err := NewMsgServer(fixture.keeper).CreateClass(
		fixture.ctx,
		classRequest,
	)
	require.NoError(t, err)
	require.Len(t, classResponse.ClassId, 131)

	data, err := cdctypes.NewAnyWithValue(&nfttypes.BasicNFTData{
		Description: strings.Repeat("d", maximumPayloadNFTDataBytes-3),
	})
	require.NoError(t, err)
	require.Equal(t, nfttypes.BasicNFTDataTypeURL, data.TypeUrl)
	require.Len(t, data.Value, maximumPayloadNFTDataBytes)

	server := NewMsgServer(fixture.keeper)
	for index := 0; index < maximumPageBenchmarkStateSize; index++ {
		request := validMintRequest(classResponse.ClassId, creator, owner)
		request.NftId = maximumPayloadIdentifier("n", index)
		request.Uri = strings.Repeat("u", maximumPayloadURIBytes)
		request.Data = &cdctypes.Any{
			TypeUrl: data.TypeUrl,
			Value:   append([]byte(nil), data.Value...),
		}
		_, err := server.Mint(fixture.ctx, request)
		require.NoError(t, err)
	}
	fixture.ctx = fixture.ctx.WithEventManager(sdk.NewEventManager())
	return maximumPageQueryFixture{
		goCtx:    fixture.ctx,
		classID:  classResponse.ClassId,
		owner:    owner,
		standard: NewStandardQueryServer(fixture.keeper),
		panacea:  NewQueryServer(fixture.keeper),
	}
}

func maximumPayloadClassRequest(
	creator string,
	index int,
) *nfttypes.MsgCreateClassRequest {
	request := validCreateClassRequest(creator)
	request.LocalClassId = maximumPayloadIdentifier("c", index)
	request.Name = strings.Repeat("n", maximumPayloadClassNameBytes)
	request.Symbol = strings.Repeat("s", maximumPayloadClassSymbolBytes)
	request.Description = strings.Repeat("d", maximumPayloadClassDescriptionBytes)
	request.Uri = strings.Repeat("u", maximumPayloadURIBytes)
	return request
}

func maximumPayloadIdentifier(prefix string, index int) string {
	suffix := fmt.Sprintf("%03d", index)
	return strings.Repeat(prefix, maximumPayloadLocalIDBytes-len(suffix)) + suffix
}

func contextWithQueryGasMeter(
	goCtx context.Context,
) (context.Context, storetypes.GasMeter) {
	meter := storetypes.NewInfiniteGasMeter()
	return sdk.UnwrapSDKContext(goCtx).WithGasMeter(meter),
		meter
}
