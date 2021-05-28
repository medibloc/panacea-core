package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

//nolint:unused,deadcode
func listToken(ctx sdk.Context, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	tokens := keeper.GetAllToken(ctx)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, tokens)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

//nolint:unused,deadcode
func getToken(ctx sdk.Context, symbol string, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	if !keeper.HasToken(ctx, symbol) {
		return nil, sdkerrors.ErrKeyNotFound
	}

	token := keeper.GetToken(ctx, symbol)

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, token)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
