package keeper

import (
	"bytes"
	"sort"
	"testing"

	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStandardQueryClassesReturnsEmptyPage(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	request := &upstreamnft.QueryClassesRequest{}

	response, err := NewStandardQueryServer(fixture.keeper).Classes(
		fixture.ctx,
		request,
	)

	require.NoError(t, err)
	require.Empty(t, response.Classes)
	require.NotNil(t, response.Pagination)
	require.Empty(t, response.Pagination.NextKey)
	require.Zero(t, response.Pagination.Total)
	require.Nil(t, request.Pagination)
}

func TestStandardQueryClassesPaginatesInClassIDOrder(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classIDs := createClassesForStandardQueryTest(t, &fixture, "gamma", "alpha", "beta")
	server := NewStandardQueryServer(fixture.keeper)
	goCtx := fixture.ctx

	first, err := server.Classes(goCtx, &upstreamnft.QueryClassesRequest{
		Pagination: &query.PageRequest{Limit: 2},
	})
	require.NoError(t, err)
	require.Equal(t, []string{classIDs[0], classIDs[1]}, standardClassIDs(first.Classes))
	require.NotEmpty(t, first.Pagination.NextKey)
	require.Zero(t, first.Pagination.Total)

	second, err := server.Classes(goCtx, &upstreamnft.QueryClassesRequest{
		Pagination: &query.PageRequest{Key: first.Pagination.NextKey, Limit: 2},
	})
	require.NoError(t, err)
	require.Equal(t, []string{classIDs[2]}, standardClassIDs(second.Classes))
	require.Empty(t, second.Pagination.NextKey)
	require.Zero(t, second.Pagination.Total)
}

func TestStandardQueryClassesSupportsReversePagination(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	classIDs := createClassesForStandardQueryTest(t, &fixture, "gamma", "alpha", "beta")
	server := NewStandardQueryServer(fixture.keeper)
	goCtx := fixture.ctx

	first, err := server.Classes(goCtx, &upstreamnft.QueryClassesRequest{
		Pagination: &query.PageRequest{Limit: 2, Reverse: true},
	})
	require.NoError(t, err)
	require.Equal(t, []string{classIDs[2], classIDs[1]}, standardClassIDs(first.Classes))
	require.NotEmpty(t, first.Pagination.NextKey)

	second, err := server.Classes(goCtx, &upstreamnft.QueryClassesRequest{
		Pagination: &query.PageRequest{
			Key:     first.Pagination.NextKey,
			Limit:   2,
			Reverse: true,
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{classIDs[0]}, standardClassIDs(second.Classes))
	require.Empty(t, second.Pagination.NextKey)
}

func TestStandardQueryClassesErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	server := NewStandardQueryServer(fixture.keeper)
	goCtx := fixture.ctx

	_, err := server.Classes(goCtx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	for _, pagination := range []*query.PageRequest{
		{Limit: 101},
		{Offset: 1},
		{CountTotal: true},
	} {
		_, err = server.Classes(goCtx, &upstreamnft.QueryClassesRequest{Pagination: pagination})
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	}

	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{92}, 20)))
	orphanClassID := creator + ":orphan"
	require.NoError(t, fixture.keeper.nftKeeper.SaveClass(
		fixture.ctx,
		upstreamnft.Class{Id: orphanClassID},
	))
	_, err = server.Classes(goCtx, &upstreamnft.QueryClassesRequest{})
	require.Equal(t, codes.Internal, status.Code(err))
}

func createClassesForStandardQueryTest(
	t *testing.T,
	fixture *keeperFixture,
	localClassIDs ...string,
) []string {
	t.Helper()
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{91}, 20)))
	server := NewMsgServer(fixture.keeper)
	classIDs := make([]string, 0, len(localClassIDs))
	for _, localClassID := range localClassIDs {
		request := validCreateClassRequest(creator)
		request.LocalClassId = localClassID
		response, err := server.CreateClass(fixture.ctx, request)
		require.NoError(t, err)
		classIDs = append(classIDs, response.ClassId)
	}
	sort.Strings(classIDs)
	return classIDs
}

func standardClassIDs(classes []*upstreamnft.Class) []string {
	classIDs := make([]string, len(classes))
	for index, class := range classes {
		classIDs[index] = class.Id
	}
	return classIDs
}
