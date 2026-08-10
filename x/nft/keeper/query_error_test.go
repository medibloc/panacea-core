package keeper

import (
	"errors"
	"testing"

	upstreamnft "cosmossdk.io/x/nft"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapQueryStateError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "class not found", err: upstreamnft.ErrClassNotExists.Wrap("missing class"), code: codes.NotFound},
		{name: "nft not found", err: upstreamnft.ErrNFTNotExists.Wrap("missing nft"), code: codes.NotFound},
		{name: "invalid stored data", err: sdkerrors.ErrInvalidRequest.Wrap("bad stored data"), code: codes.Internal},
		{name: "invalid stored address", err: sdkerrors.ErrInvalidAddress.Wrap("bad stored owner"), code: codes.Internal},
		{name: "unexpected error", err: errors.New("store failure"), code: codes.Internal},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := mapQueryStateError(testCase.err)

			require.Equal(t, testCase.code, status.Code(err))
			require.Equal(t, testCase.err.Error(), status.Convert(err).Message())
		})
	}
}

func TestMapQueryStateErrorReturnsNilForNil(t *testing.T) {
	require.NoError(t, mapQueryStateError(nil))
}
