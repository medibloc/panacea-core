package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (k Keeper) RegisterDataValidator(ctx sdk.Context, dataValidator types.DataValidator) error {
	validatorAddr, err := sdk.AccAddressFromBech32(dataValidator.Address)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(validatorAddr)
	if store.Has(dataValidatorKey) {
		return sdkerrors.Wrapf(types.ErrDataValidatorAlreadyExist, "the data validator %s is already exists", dataValidator.Address)
	}

	validatorPubKey, err := k.accountKeeper.GetPubKey(ctx, validatorAddr)
	if err != nil {
		return err
	}
	if validatorPubKey == nil {
		return sdkerrors.ErrKeyNotFound
	}

	err = k.SetDataValidator(ctx, dataValidator)
	if err != nil {
		return err
	}

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

func (k Keeper) SetDataValidator(ctx sdk.Context, dataValidator types.DataValidator) error {
	validatorAddr, err := sdk.AccAddressFromBech32(dataValidator.Address)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(validatorAddr)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&dataValidator)
	store.Set(dataValidatorKey, bz)
	return nil
}
