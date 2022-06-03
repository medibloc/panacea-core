package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (k Keeper) RegisterOracle(ctx sdk.Context, oracle types.Oracle) error {
	oracleAddr, err := sdk.AccAddressFromBech32(oracle.Address)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	oracleKey := types.GetKeyPrefixOracle(oracleAddr)
	if store.Has(oracleKey) {
		return sdkerrors.Wrapf(types.ErrOracleAlreadyExist, "the oracle %s is already exists", oracle.Address)
	}

	oraclePubKey, err := k.accountKeeper.GetPubKey(ctx, oracleAddr)
	if err != nil {
		return err
	}
	if oraclePubKey == nil {
		return sdkerrors.ErrKeyNotFound
	}

	err = k.SetOracle(ctx, oracle)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) GetAllOracles(ctx sdk.Context) ([]types.Oracle, error) {
	// TODO: add pagination
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixOracles)
	defer iterator.Close()

	oracles := make([]types.Oracle, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var oracle types.Oracle

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &oracle)
		if err != nil {
			return nil, err
		}

		oracles = append(oracles, oracle)
	}

	return oracles, nil
}

func (k Keeper) GetOracle(ctx sdk.Context, oracleAddress sdk.AccAddress) (types.Oracle, error) {
	store := ctx.KVStore(k.storeKey)
	oracleKey := types.GetKeyPrefixOracle(oracleAddress)
	if !store.Has(oracleKey) {
		return types.Oracle{}, sdkerrors.Wrapf(types.ErrOracleNotFound, "the oracle %s is not found", oracleAddress)
	}
	bz := store.Get(oracleKey)

	var oracle types.Oracle
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &oracle)
	if err != nil {
		return types.Oracle{}, err
	}

	return oracle, nil
}

func (k Keeper) SetOracle(ctx sdk.Context, oracle types.Oracle) error {
	oracleAddr, err := sdk.AccAddressFromBech32(oracle.Address)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	oracleKey := types.GetKeyPrefixOracle(oracleAddr)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&oracle)
	store.Set(oracleKey, bz)
	return nil
}

func (k Keeper) IsRegisteredOracle(ctx sdk.Context, oracleAddress sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	oracleKey := types.GetKeyPrefixOracle(oracleAddress)
	return store.Has(oracleKey)
}

func (k Keeper) UpdateOracle(ctx sdk.Context, address sdk.AccAddress, endpoint string) error {
	oracle, err := k.GetOracle(ctx, address)
	if err != nil {
		return err
	}

	oracle.Endpoint = endpoint

	err = k.SetOracle(ctx, oracle)
	if err != nil {
		return err
	}

	return nil
}
