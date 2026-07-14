package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	pnftkeeper "github.com/medibloc/panacea-core/v2/x/pnft/keeper"
	pnfttypes "github.com/medibloc/panacea-core/v2/x/pnft/types"
	"github.com/stretchr/testify/require"
)

func TestMsgServerRetainsBasicValidation(t *testing.T) {
	goCtx := sdk.WrapSDKContext(sdk.Context{})
	// A nil keeper makes the ordering observable: every invalid request must be
	// rejected by ValidateBasic before the MsgServer accesses keeper state.
	msgServer := pnftkeeper.NewMsgServerImpl(nil)

	testCases := []struct {
		name    string
		wantErr error
		call    func() error
	}{
		{
			name:    "CreateDenom",
			wantErr: pnfttypes.ErrCreateDenom,
			call: func() error {
				_, err := msgServer.CreateDenom(goCtx, &pnfttypes.MsgCreateDenomRequest{})
				return err
			},
		},
		{
			name:    "UpdateDenom",
			wantErr: pnfttypes.ErrUpdateDenom,
			call: func() error {
				_, err := msgServer.UpdateDenom(goCtx, &pnfttypes.MsgUpdateDenomRequest{})
				return err
			},
		},
		{
			name:    "DeleteDenom",
			wantErr: pnfttypes.ErrDeleteDenom,
			call: func() error {
				_, err := msgServer.DeleteDenom(goCtx, &pnfttypes.MsgDeleteDenomRequest{})
				return err
			},
		},
		{
			name:    "TransferDenom",
			wantErr: pnfttypes.ErrTransferDenom,
			call: func() error {
				_, err := msgServer.TransferDenom(goCtx, &pnfttypes.MsgTransferDenomRequest{})
				return err
			},
		},
		{
			name:    "MintPNFT",
			wantErr: pnfttypes.ErrMintPNFT,
			call: func() error {
				_, err := msgServer.MintPNFT(goCtx, &pnfttypes.MsgMintPNFTRequest{})
				return err
			},
		},
		{
			name:    "TransferPNFT",
			wantErr: pnfttypes.ErrTransferPNFT,
			call: func() error {
				_, err := msgServer.TransferPNFT(goCtx, &pnfttypes.MsgTransferPNFTRequest{})
				return err
			},
		},
		{
			name:    "BurnPNFT",
			wantErr: pnfttypes.ErrBurnPNFT,
			call: func() error {
				_, err := msgServer.BurnPNFT(goCtx, &pnfttypes.MsgBurnPNFTRequest{})
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			require.NotPanics(t, func() {
				err = tc.call()
			})
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
