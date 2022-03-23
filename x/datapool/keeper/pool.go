package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (k Keeper) RegisterDataValidator(ctx sdk.Context, dataValidatorAddress sdk.AccAddress, dataValidator types.DataValidator) error {
	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(dataValidatorAddress)
	if store.Has(dataValidatorKey) {
		return sdkerrors.Wrapf(types.ErrDataValidatorAlreadyExist, "the data validator %s is already exists", dataValidatorAddress)
	}

	balance := k.bankKeeper.GetBalance(ctx, dataValidatorAddress, assets.MicroMedDenom)
	if balance.IsZero() {
		return sdkerrors.ErrInsufficientFunds
	}

	k.SetDataValidator(ctx, dataValidatorAddress, dataValidator)
	return nil
}

func (k Keeper) GetDataValidator(ctx sdk.Context, dataValidatorAddress sdk.AccAddress) (types.DataValidator, error) {
	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(dataValidatorAddress)
	if !store.Has(dataValidatorKey) {
		return types.DataValidator{}, sdkerrors.Wrapf(types.ErrDataValidatorNotFound, "the data validator %s is not found", dataValidatorAddress)
	}

	bz := store.Get(dataValidatorKey)

	var dataValidator types.DataValidator
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &dataValidator)
	if err != nil {
		return types.DataValidator{}, nil
	}

	return dataValidator, nil
}

func (k Keeper) SetDataValidator(ctx sdk.Context, dataValidatorAddress sdk.AccAddress, dataValidator types.DataValidator) {
	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(dataValidatorAddress)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&dataValidator)
	store.Set(dataValidatorKey, bz)
}
