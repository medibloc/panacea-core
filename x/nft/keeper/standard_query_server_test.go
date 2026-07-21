package keeper

import (
	"bytes"
	"strings"
	"testing"

	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStandardQueryClass(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{61}, 20)))
	request := validCreateClassRequest(creator)
	_, err := NewMsgServer(fixture.keeper).CreateClass(
		sdk.WrapSDKContext(fixture.ctx),
		request,
	)
	require.NoError(t, err)

	classID := creator + ":" + request.LocalClassId
	response, err := NewStandardQueryServer(fixture.keeper).Class(
		sdk.WrapSDKContext(fixture.ctx),
		&upstreamnft.QueryClassRequest{ClassId: classID},
	)
	require.NoError(t, err)
	require.NotNil(t, response.Class)
	require.Equal(t, classID, response.Class.Id)
	require.Equal(t, request.Name, response.Class.Name)
	require.Equal(t, request.Symbol, response.Class.Symbol)
	require.Equal(t, request.Description, response.Class.Description)
	require.Equal(t, request.Uri, response.Class.Uri)
	require.Equal(t, request.UriHash, response.Class.UriHash)
	require.Nil(t, response.Class.Data)
}

func TestStandardQueryClassErrorMapping(t *testing.T) {
	fixture := newKeeperFixture(t, true, true)
	server := NewStandardQueryServer(fixture.keeper)
	creator := fixture.accountAddress(t, sdk.AccAddress(bytes.Repeat([]byte{62}, 20)))
	goCtx := sdk.WrapSDKContext(fixture.ctx)

	_, err := server.Class(goCtx, nil)
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Class(goCtx, &upstreamnft.QueryClassRequest{ClassId: "invalid"})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Class(goCtx, &upstreamnft.QueryClassRequest{
		ClassId: strings.ToUpper(creator) + ":class",
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = server.Class(goCtx, &upstreamnft.QueryClassRequest{
		ClassId: creator + ":missing",
	})
	require.Equal(t, codes.NotFound, status.Code(err))

	orphanClassID := creator + ":orphan"
	require.NoError(t, fixture.keeper.nftKeeper.SaveClass(
		fixture.ctx,
		upstreamnft.Class{Id: orphanClassID},
	))
	_, err = server.Class(goCtx, &upstreamnft.QueryClassRequest{ClassId: orphanClassID})
	require.Equal(t, codes.Internal, status.Code(err))
}
