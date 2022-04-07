package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (k Keeper) DistributeRevenuePools(ctx sdk.Context) error {
	delayedRevenueDistribute, err := k.GetDelayedRevenueDistribute(ctx)
	if err != nil {
		return err
	}

	if delayedRevenueDistribute.IsEmpty() {
		return nil
	}

	// This lastIndex is to execute the following iteration in the next block if it does not end in one block.
	// But We're currently processing it all in one block. :)
	lastIndex := 0
	// TODO We need to think about how to deal with the pool where distribution has failed.
	for i, poolID := range delayedRevenueDistribute.PoolIds {
		lastIndex = i
		// search current pool info
		pool, err := k.GetPool(ctx, poolID)
		if err != nil {
			return err
		}

		// send deposit to curator
		err = k.sendDepositToCurator(ctx, pool)
		if err != nil {
			return err
		}

		poolAddress, err := sdk.AccAddressFromBech32(pool.GetPoolAddress())
		if err != nil {
			return err
		}
		// if the sales amount is 0, the processing of this pool is terminated.
		availablePoolCoinAmount := k.getAvailablePoolCoinAmount(ctx, poolAddress, pool.IsPaidDeposit)
		if availablePoolCoinAmount.Equal(sdk.NewInt(0)) {
			continue
		}

		// calculate the revenue to be sent to each seller(totalSalesBalance / totalShareTokenAmount)
		eachDistributionAmount, err := k.getEachDistributionAmount(ctx, pool)
		if err != nil {
			return err
		}

		// distribute of revenue to each seller
		revenueDistributeTargets := k.ListRevenueDistributeTargetByRound(ctx, pool.PoolId, pool.Round)
		for _, revenueDistributeTarget := range revenueDistributeTargets {
			paidCoin := revenueDistributeTarget.PaidCoin
			// if it has already been distributed, pass it.
			if eachDistributionAmount.Equal(paidCoin.Amount) {
				continue
			}

			paymentAmount := eachDistributionAmount.Mul(paidCoin.Amount)
			sellerAddr, err := sdk.AccAddressFromBech32(revenueDistributeTarget.Address)
			if err != nil {
				return err
			}
			paymentCoin := sdk.NewCoin(assets.MicroMedDenom, paymentAmount)
			err = k.bankKeeper.SendCoins(
				ctx,
				poolAddress,
				sellerAddr,
				sdk.NewCoins(paymentCoin))
			revenueDistributeTarget.PaidCoin.Add(paymentCoin)
		}

		/*k.bankKeeper.IterateAllBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
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
		})*/

		// an error occur when distributing revenue to sellers
		if err != nil {
			return err
		}
	}
	// delete this pool from delayedRevenueDistribute
	delayedRevenueDistribute.RemovePreviousIndex(lastIndex)
	err = k.SetDelayedRevenueDistribute(ctx, delayedRevenueDistribute)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) getAvailablePoolCoinAmount(ctx sdk.Context, poolAddress sdk.AccAddress, isPaidDeposit bool) sdk.Int {
	poolBalance := k.bankKeeper.GetBalance(ctx, poolAddress, assets.MicroMedDenom)
	if isPaidDeposit {
		return poolBalance.Amount
	} else {
		deposit := k.GetParams(ctx).DataPoolDeposit
		return poolBalance.Amount.Quo(deposit.Amount)
	}
}

func (k Keeper) getEachDistributionAmount(ctx sdk.Context, pool *types.Pool) (sdk.Int, error) {
	// get total amount of shareTokens in the pool
	shareTokenDenom := types.GetDenomOfShareToken(pool.PoolId)
	totalShareTokenAmount := k.bankKeeper.GetSupply(ctx).GetTotal().AmountOf(shareTokenDenom)

	zeroAmount := sdk.NewInt(0)

	// get the pool's total sales revenue
	poolAddress, err := sdk.AccAddressFromBech32(pool.GetPoolAddress())
	if err != nil {
		return zeroAmount, err
	}
	poolSalesBalance := k.bankKeeper.GetBalance(ctx, poolAddress, assets.MicroMedDenom)
	// TODO Our deposit can be changed by governance, so we need to put the amount of our deposit in the pool.
	deposit := k.GetParams(ctx).DataPoolDeposit
	totalSalesBalance := poolSalesBalance.Sub(deposit)

	// calculator the amount to be distributed to each seller
	if totalSalesBalance.Amount.Equal(zeroAmount) {
		return zeroAmount, nil
	}
	eachDistributionAmount := totalSalesBalance.Amount.Quo(totalShareTokenAmount)

	return eachDistributionAmount, nil
}

func (k Keeper) sendDepositToCurator(ctx sdk.Context, pool *types.Pool) error {
	// TODO Our deposits can be changed by governance, so we need to get our deposit information from the pool.
	if pool.IsPaidDeposit {
		return nil
	}

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
	pool.IsPaidDeposit = true
	k.SetPool(ctx, pool)

	return nil
}

func (k Keeper) SetDelayedRevenueDistribute(ctx sdk.Context, delayedRevenueDistribute *types.DelayedRevenueDistribute) error {
	if delayedRevenueDistribute == nil {
		return sdkerrors.ErrInvalidRequest
	}

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(delayedRevenueDistribute)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixDelayedRevenueDistribute, bz)
	return nil
}

func (k Keeper) GetDelayedRevenueDistribute(ctx sdk.Context) (*types.DelayedRevenueDistribute, error) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.KeyPrefixDelayedRevenueDistribute) {
		return &types.DelayedRevenueDistribute{}, nil
	}

	var delayedRevenueDistribute types.DelayedRevenueDistribute
	bz := store.Get(types.KeyPrefixDelayedRevenueDistribute)
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &delayedRevenueDistribute)
	if err != nil {
		return nil, err
	}

	return &delayedRevenueDistribute, nil
}

func (k Keeper) appendDistributeRevenuePools(ctx sdk.Context, pool *types.Pool) error {
	if types.ACTIVE != pool.Status {
		return nil
	}

	delayedRevenueDistribute, err := k.GetDelayedRevenueDistribute(ctx)
	if err != nil {
		return err
	}
	delayedRevenueDistribute.AppendPoolID(pool.PoolId)

	err = k.SetDelayedRevenueDistribute(ctx, delayedRevenueDistribute)
	if err != nil {
		return err
	}

	return nil
}

// initRevenueDistributeTarget initializes the revenue distribution target information.
func (k Keeper) initRevenueDistributeTarget(ctx sdk.Context, poolID, round uint64, addr sdk.AccAddress) {
	zeroFund := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0))
	revenueDistributeTarget := types.RevenueDistributeTarget{
		PoolId:   poolID,
		Round:    round,
		Address:  addr.String(),
		PaidCoin: &zeroFund,
	}
	k.setRevenueDistributeTarget(ctx, revenueDistributeTarget)
}

func (k Keeper) setRevenueDistributeTarget(ctx sdk.Context, revenueDistributeTarget types.RevenueDistributeTarget) {
	key := types.GetKeyPrefixDistributeRevenueTarget(
		revenueDistributeTarget.PoolId,
		revenueDistributeTarget.Round,
		revenueDistributeTarget.Address,
	)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, k.cdc.MustMarshalBinaryLengthPrefixed(&revenueDistributeTarget))
}

func (k Keeper) GetRevenueDistributeTarget(ctx sdk.Context, poolID, round uint64, addr sdk.AccAddress) *types.RevenueDistributeTarget {
	key := types.GetKeyPrefixDistributeRevenueTarget(poolID, round, addr.String())
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	var revenueDistributeTarget types.RevenueDistributeTarget
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &revenueDistributeTarget)

	return &revenueDistributeTarget
}

func (k Keeper) ListRevenueDistributeTargetByRound(ctx sdk.Context, poolID, round uint64) []types.RevenueDistributeTarget {
	key := types.GetKeyPrefixDistributeRevenueTargetByRound(poolID, round)
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, key)
	defer iter.Close()

	targetList := make([]types.RevenueDistributeTarget, 0)
	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var revenueDistributeTarget types.RevenueDistributeTarget
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &revenueDistributeTarget)
		targetList = append(targetList, revenueDistributeTarget)
	}
	return targetList
}
