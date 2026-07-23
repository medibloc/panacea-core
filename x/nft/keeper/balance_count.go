package keeper

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) incrementOwnerClassCount(
	ctx sdk.Context,
	classID string,
	owner sdk.AccAddress,
) error {
	key, err := k.ownerClassCountKey(classID, owner)
	if err != nil {
		return err
	}
	count, found, err := k.loadOwnerClassCount(ctx, key)
	if err != nil {
		return err
	}
	if found && count == math.MaxUint64 {
		return fmt.Errorf(
			"owner %s class %s balance count overflows",
			key.K2(),
			classID,
		)
	}
	if err := k.ownerClassCounts.Set(ctx, key, count+1); err != nil {
		return fmt.Errorf(
			"increment owner %s class %s balance count: %w",
			key.K2(),
			classID,
			err,
		)
	}
	return nil
}

func (k Keeper) decrementOwnerClassCount(
	ctx sdk.Context,
	classID string,
	owner sdk.AccAddress,
) error {
	key, err := k.ownerClassCountKey(classID, owner)
	if err != nil {
		return err
	}
	count, found, err := k.loadOwnerClassCount(ctx, key)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf(
			"owner %s class %s balance count is missing",
			key.K2(),
			classID,
		)
	}
	if count == 1 {
		if err := k.ownerClassCounts.Remove(ctx, key); err != nil {
			return fmt.Errorf(
				"remove owner %s class %s zero balance count: %w",
				key.K2(),
				classID,
				err,
			)
		}
		return nil
	}
	if err := k.ownerClassCounts.Set(ctx, key, count-1); err != nil {
		return fmt.Errorf(
			"decrement owner %s class %s balance count: %w",
			key.K2(),
			classID,
			err,
		)
	}
	return nil
}

func (k Keeper) transferOwnerClassCount(
	ctx sdk.Context,
	classID string,
	sender sdk.AccAddress,
	receiver sdk.AccAddress,
) error {
	if bytes.Equal(sender, receiver) {
		key, err := k.ownerClassCountKey(classID, sender)
		if err != nil {
			return err
		}
		_, found, err := k.loadOwnerClassCount(ctx, key)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf(
				"owner %s class %s balance count is missing",
				key.K2(),
				classID,
			)
		}
		return nil
	}
	if err := k.decrementOwnerClassCount(ctx, classID, sender); err != nil {
		return err
	}
	if err := k.incrementOwnerClassCount(ctx, classID, receiver); err != nil {
		return err
	}
	return nil
}

func (k Keeper) ownerClassCountKey(
	classID string,
	owner sdk.AccAddress,
) (collections.Pair[string, string], error) {
	ownerString, err := k.addressCodec.BytesToString(owner)
	if err != nil {
		return collections.Pair[string, string]{}, fmt.Errorf("encode balance owner: %w", err)
	}
	return collections.Join(classID, ownerString), nil
}

func (k Keeper) loadOwnerClassCount(
	ctx sdk.Context,
	key collections.Pair[string, string],
) (uint64, bool, error) {
	count, err := k.ownerClassCounts.Get(ctx, key)
	if errors.Is(err, collections.ErrNotFound) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf(
			"load owner %s class %s balance count: %w",
			key.K2(),
			key.K1(),
			err,
		)
	}
	if count == 0 {
		return 0, false, fmt.Errorf(
			"owner %s class %s has a stored zero balance count",
			key.K2(),
			key.K1(),
		)
	}
	return count, true, nil
}
