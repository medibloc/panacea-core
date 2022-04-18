package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (k Keeper) DistributeRevenuePools(ctx sdk.Context) error {
	revenuePools, err := k.getDistributeRevenuePools(ctx)
	if err != nil {
		return err
	}

	if revenuePools.IsEmpty() {
		return nil
	}

	// This lastIndex is to execute the following iteration in the next block if it does not end in one block.
	// But We're currently processing it all in one block. :)
	lastIndex := 0
	// TODO We need to think about how to deal with the pool where distribution has failed.
	for i, poolID := range revenuePools.PoolIds {
		lastIndex = i
		// search current pool info
		pool, err := k.GetPool(ctx, poolID)
		if err != nil {
			return err
		}

		// calculate the revenue to be sent to each seller(totalSalesBalance / totalShareTokenAmount)
		eachDistributionAmount, err := k.getEachDistributionAmount(ctx, pool)
		if err != nil {
			return err
		}

		// send deposit to curator
		err = k.sendDepositToCurator(ctx, pool)
		if err != nil {
			return err
		}

		// distribute of revenue to each seller
		k.bankKeeper.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
			if balance.Denom != types.GetDenomOfShareToken(poolID) {
				return false
			}
			// Seller must receive a distribution equal to the shareTokens it owns.
			distributeCoins := sdk.NewCoins(
				sdk.NewCoin(assets.MicroMedDenom, eachDistributionAmount.Mul(balance.Amount)),
			)

			poolAddress, innerErr := sdk.AccAddressFromBech32(pool.GetPoolAddress())
			// stop distribute when an error occurs
			if innerErr != nil {
				err = innerErr
				return true
			}

			// excluding moduleAccount
			if types.GetModuleAddress().Equals(addr) {
				return false
			}

			err = k.bankKeeper.SendCoins(ctx, poolAddress, addr, distributeCoins)

			if err != nil {
				k.Logger(ctx).Error("failed to distribute coin", err)
				return true
			}
			return false
		})

		// an error occur when distributing revenue to sellers
		if err != nil {
			return err
		}
	}
	// delete this pool from revenuePools
	revenuePools.RemovePreviousIndex(lastIndex)
	err = k.setDistributePoolsRevenue(ctx, revenuePools)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) getEachDistributionAmount(ctx sdk.Context, pool *types.Pool) (sdk.Int, error) {
	// get total amount of shareTokens in the pool
	shareTokenDenom := types.GetDenomOfShareToken(pool.PoolId)
	totalShareTokenAmount := k.bankKeeper.GetSupply(ctx).GetTotal().AmountOf(shareTokenDenom)

	// get the pool's total sales revenue
	poolAddress, err := sdk.AccAddressFromBech32(pool.GetPoolAddress())
	if err != nil {
		return sdk.NewInt(0), err
	}
	poolSalesBalance := k.bankKeeper.GetBalance(ctx, poolAddress, assets.MicroMedDenom)
	// TODO Our deposit can be changed by governance, so we need to put the amount of our deposit in the pool.
	deposit := k.GetParams(ctx).DataPoolDeposit
	totalSalesBalance := poolSalesBalance.Sub(deposit)

	// calculator the amount to be distributed to each seller
	if totalSalesBalance.Amount.Equal(sdk.NewInt(0)) {
		return sdk.NewInt(0), nil
	}
	eachDistributionAmount := totalSalesBalance.Amount.Quo(totalShareTokenAmount)

	return eachDistributionAmount, nil
}

func (k Keeper) sendDepositToCurator(ctx sdk.Context, pool *types.Pool) error {
	// TODO Our deposits can be changed by governance, so we need to get our deposit information from the pool.
	deposit := k.GetParams(ctx).DataPoolDeposit
	curatorAddr, err := sdk.AccAddressFromBech32(pool.Curator)
	if err != nil {
		return err
	}
	poolAddr, err := sdk.AccAddressFromBech32(pool.PoolAddress)
	if err != nil {
		return err
	}
	err = k.bankKeeper.SendCoins(ctx, poolAddr, curatorAddr, sdk.NewCoins(deposit))
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) setDistributePoolsRevenue(ctx sdk.Context, revenuePools *types.DistributeRevenuePools) error {
	if revenuePools == nil {
		return sdkerrors.ErrInvalidRequest
	}

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(revenuePools)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixDistributePoolsRevenue, bz)
	return nil
}

func (k Keeper) getDistributeRevenuePools(ctx sdk.Context) (*types.DistributeRevenuePools, error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.KeyPrefixDistributePoolsRevenue) {
		return &types.DistributeRevenuePools{}, nil
	}

	var distributeRevenuePools types.DistributeRevenuePools
	bz := store.Get(types.KeyPrefixDistributePoolsRevenue)
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &distributeRevenuePools)
	if err != nil {
		return nil, err
	}

	return &distributeRevenuePools, nil
}

func (k Keeper) appendDistributeRevenuePools(ctx sdk.Context, pool *types.Pool) error {
	if types.ACTIVE != pool.Status {
		return nil
	}

	revenuePools, err := k.getDistributeRevenuePools(ctx)
	if err != nil {
		return err
	}
	revenuePools.AppendPoolID(pool.PoolId)

	err = k.setDistributePoolsRevenue(ctx, revenuePools)
	if err != nil {
		return err
	}

	return nil
}
