package keeper

import (
	"fmt"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"
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
		return types.DataValidator{}, err
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

func (k Keeper) IsRegisteredDataValidator(ctx sdk.Context, dataValidatorAddress sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(dataValidatorAddress)
	return store.Has(dataValidatorKey)
}

func (k Keeper) CreatePool(ctx sdk.Context, curator sdk.AccAddress, poolParams types.PoolParams) (uint64, error) {
	// Get the next pool id
	poolID := k.GetNextPoolNumberAndIncrement(ctx)

	// new pool
	newPool := types.NewPool(poolID, curator, poolParams)

	// pool address for deposit
	poolAddress, err := types.AccPoolAddressFromBech32(newPool.GetPoolAddress())
	if err != nil {
		return 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address of pool %s", newPool.GetPoolAddress())
	}

	// set new account for pool
	acc := k.accountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(poolAddress),
		newPool.GetPoolAddress(),
	))
	k.accountKeeper.SetAccount(ctx, acc)

	// check if the trusted_data_validators are registered
	for _, dataValidator := range poolParams.TrustedDataValidators {
		accAddr, _ := sdk.AccAddressFromBech32(dataValidator)
		if !k.IsRegisteredDataValidator(ctx, accAddr) {
			return 0, sdkerrors.Wrapf(types.ErrNotRegisteredDataValidator, "the data validator %s is not registered", dataValidator)
		}
	}

	// curator send deposit to pool for creation of pool

	deposit := sdk.NewCoins(*newPool.PoolParams.Deposit)
	err = k.bankKeeper.SendCoins(ctx, curator, poolAddress, deposit)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrNotEnoughPoolDeposit, "The curator's balance is not enough to make a data pool")
	}

	// store pool
	k.SetPool(ctx, newPool)

	// TODO: issue curation NFT

	return newPool.GetPoolId(), nil
}

func (k Keeper) GetNextPoolNumberAndIncrement(ctx sdk.Context) uint64 {
	poolNumber := k.GetNextPoolNumber(ctx)
	k.SetPoolNumber(ctx, poolNumber+1)
	return poolNumber
}

func (k Keeper) GetNextPoolNumber(ctx sdk.Context) uint64 {
	var poolNumber uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyPoolNextNumber)
	if bz == nil {
		panic(fmt.Errorf("pool has not been initialized -- Should have been done in InitGenesis"))
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &val)
		if err != nil {
			panic(err)
		}

		poolNumber = val.GetValue()
	}
	return poolNumber
}

func (k Keeper) SetPoolNumber(ctx sdk.Context, poolNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.UInt64Value{Value: poolNumber})
	store.Set(types.KeyPoolNextNumber, bz)
}

func (k Keeper) SetPool(ctx sdk.Context, pool *types.Pool) {
	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(pool.GetPoolId())
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(pool)
	store.Set(poolKey, bz)
}
