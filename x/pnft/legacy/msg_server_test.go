package legacy

import (
	"context"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/stretchr/testify/require"
)

func TestMsgServerRejectsEveryLegacyMessage(t *testing.T) {
	server := NewMsgServer()
	testCases := []struct {
		name string
		call func(context.Context) (any, error)
	}{
		{
			name: "create denom",
			call: func(ctx context.Context) (any, error) {
				return server.CreateDenom(ctx, &types.MsgCreateDenomRequest{})
			},
		},
		{
			name: "update denom",
			call: func(ctx context.Context) (any, error) {
				return server.UpdateDenom(ctx, &types.MsgUpdateDenomRequest{})
			},
		},
		{
			name: "delete denom",
			call: func(ctx context.Context) (any, error) {
				return server.DeleteDenom(ctx, &types.MsgDeleteDenomRequest{})
			},
		},
		{
			name: "transfer denom",
			call: func(ctx context.Context) (any, error) {
				return server.TransferDenom(ctx, &types.MsgTransferDenomRequest{})
			},
		},
		{
			name: "mint PNFT",
			call: func(ctx context.Context) (any, error) {
				return server.MintPNFT(ctx, &types.MsgMintPNFTRequest{})
			},
		},
		{
			name: "transfer PNFT",
			call: func(ctx context.Context) (any, error) {
				return server.TransferPNFT(ctx, &types.MsgTransferPNFTRequest{})
			},
		},
		{
			name: "burn PNFT",
			call: func(ctx context.Context) (any, error) {
				return server.BurnPNFT(ctx, &types.MsgBurnPNFTRequest{})
			},
		},
	}

	var expectedError string
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := tc.call(context.Background())
			require.Nil(t, response)
			require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
			require.ErrorContains(t, err, DisabledErrorMessage)

			if expectedError == "" {
				expectedError = err.Error()
			}
			require.Equal(t, expectedError, err.Error())
		})
	}
}
