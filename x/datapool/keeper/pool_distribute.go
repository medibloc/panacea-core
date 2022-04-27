package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

// addSalesHistory stores sales information in history.
func (k Keeper) addSalesHistory(ctx sdk.Context, poolID, round uint64, addr sdk.AccAddress, dataHash []byte) {
	zeroFund := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0))
	info := types.SalesInfo{
		PoolId:   poolID,
		Round:    round,
		Address:  addr.String(),
		DataHash: dataHash,
		PaidCoin: &zeroFund,
	}

	salesHistory := k.GetSalesHistory(ctx, poolID, round)
	salesHistory.SalesInfos = append(salesHistory.SalesInfos, &info)

	k.SetSalesHistory(ctx, poolID, round, salesHistory)
}

// SetSalesHistory stores sales history.
func (k Keeper) SetSalesHistory(ctx sdk.Context, poolID, round uint64, salesHistory *types.SalesHistory) {
	key := types.GetKeyPrefixSalesHistory(
		poolID,
		round,
	)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, k.cdc.MustMarshalBinaryLengthPrefixed(salesHistory))
}

// GetSalesHistory returns the sales history. If there is no value, it responds by initializing it.
func (k Keeper) GetSalesHistory(ctx sdk.Context, poolID, round uint64) *types.SalesHistory {
	key := types.GetKeyPrefixSalesHistory(poolID, round)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	if bz == nil {
		return &types.SalesHistory{}
	}
	var salesHistory types.SalesHistory
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &salesHistory)
	return &salesHistory
}

// SetInstantRevenueDistribute stores the poolID to which the revenue should be distributed immediately.
func (k Keeper) SetInstantRevenueDistribute(ctx sdk.Context, instantRevenueDistribute *types.InstantRevenueDistribute) {
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(instantRevenueDistribute)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixInstantRevenueDistribute, bz)
}

func (k Keeper) GetInstantRevenueDistribute(ctx sdk.Context) *types.InstantRevenueDistribute {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.KeyPrefixInstantRevenueDistribute) {
		return &types.InstantRevenueDistribute{}
	}
	bz := store.Get(types.KeyPrefixInstantRevenueDistribute)
	var delayedRevenueDistribute types.InstantRevenueDistribute
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &delayedRevenueDistribute)
	return &delayedRevenueDistribute
}

// addInstantRevenueDistribute adds the poolID to distribute. If there are duplicate poolIDs, they are not added.
func (k Keeper) addInstantRevenueDistribute(ctx sdk.Context, poolID uint64) error {
	instantRevenueDistribute := k.GetInstantRevenueDistribute(ctx)
	instantRevenueDistribute.AppendPoolID(poolID)
	k.SetInstantRevenueDistribute(ctx, instantRevenueDistribute)
	return nil
}

func (k Keeper) DistributeRevenuePools(ctx sdk.Context) error {
	instantRevenueDistribute := k.GetInstantRevenueDistribute(ctx)

	if instantRevenueDistribute.IsEmpty() {
		return nil
	}

	// This lastIndex is to execute the following iteration in the next block if it does not end in one block.
	// But We're currently processing it all in one block. :)
	lastIndex := 0
	// TODO We need to think about how to deal with the pool where distribution has failed.
	for i, poolID := range instantRevenueDistribute.PoolIds {
		lastIndex = i
		// search current pool info
		pool, err := k.GetPool(ctx, poolID)
		if err != nil {
			return sdkerrors.Wrap(types.ErrRevenueDistribute, err.Error())
		}

		// send deposit to curator
		err = k.sendDepositToCurator(ctx, pool)
		if err != nil {
			return err
		}

		// distribute of revenue to each seller
		err = k.executeRevenueDistribute(ctx, pool)
		if err != nil {
			return err
		}
	}
	// delete this pool from instantRevenueDistribute
	instantRevenueDistribute.RemovePreviousIndex(lastIndex)
	k.SetInstantRevenueDistribute(ctx, instantRevenueDistribute)

	return nil
}

// sendDepositToCurator returns to the Curator if the pool status is ACTIVE but the deposit has not yet been returned.
func (k Keeper) sendDepositToCurator(ctx sdk.Context, pool *types.Pool) error {
	if pool.Status != types.ACTIVE || pool.IsPaidDeposit {
		return nil
	}

	curatorAddr, err := sdk.AccAddressFromBech32(pool.Curator)
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribute, err.Error())
	}
	poolAddr, err := sdk.AccAddressFromBech32(pool.PoolAddress)
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribute, err.Error())
	}
	err = k.bankKeeper.SendCoins(ctx, poolAddr, curatorAddr, sdk.NewCoins(pool.Deposit))
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribute, err.Error())
	}
	pool.IsPaidDeposit = true
	k.SetPool(ctx, pool)

	return nil
}

func (k Keeper) getAvailablePoolCoinAmount(ctx sdk.Context, pool *types.Pool) (*sdk.Int, error) {
	poolAddr, err := sdk.AccAddressFromBech32(pool.PoolAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrRevenueDistribute, err.Error())
	}
	poolBalance := k.bankKeeper.GetBalance(ctx, poolAddr, assets.MicroMedDenom)
	if pool.IsPaidDeposit {
		return &poolBalance.Amount, nil
	} else {
		amount := poolBalance.Amount.Sub(pool.Deposit.Amount)
		return &amount, nil
	}
}

func (k Keeper) executeRevenueDistribute(ctx sdk.Context, pool *types.Pool) error {
	availablePoolCoinAmount, err := k.getAvailablePoolCoinAmount(ctx, pool)
	if err != nil {
		return err
	}

	// calculate the revenue to be sent to each seller
	eachDistributionAmount := k.getEachDistributionAmount(pool)

	poolAddress, err := sdk.AccAddressFromBech32(pool.GetPoolAddress())
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribute, err.Error())
	}

	salesHistory := k.GetSalesHistory(ctx, pool.PoolId, pool.Round)
	for _, salesInfo := range salesHistory.GetSalesInfos() {
		// if there is no coin available, proceed no further.
		if availablePoolCoinAmount.Equal(sdk.NewInt(0)) {
			break
		}

		paidCoin := salesInfo.PaidCoin

		// if it has already been distributed, pass it.
		if eachDistributionAmount.Equal(paidCoin.Amount) {
			continue
		}

		paymentAmount := eachDistributionAmount.Sub(paidCoin.Amount)
		// If the transferable amount is less than the payable amount, it is replaced with the transferable amount.
		if availablePoolCoinAmount.LT(paymentAmount) {
			paymentAmount = *availablePoolCoinAmount
		}
		sellerAddr, err := sdk.AccAddressFromBech32(salesInfo.Address)
		if err != nil {
			return sdkerrors.Wrap(types.ErrRevenueDistribute, err.Error())
		}
		paymentCoin := sdk.NewCoin(assets.MicroMedDenom, paymentAmount)
		err = k.bankKeeper.SendCoins(
			ctx,
			poolAddress,
			sellerAddr,
			sdk.NewCoins(paymentCoin))
		if err != nil {
			return sdkerrors.Wrap(types.ErrRevenueDistribute, err.Error())
		}
		*salesInfo.PaidCoin = salesInfo.PaidCoin.Add(paymentCoin)

		// Subtract the amount paid from the transferable amount.
		*availablePoolCoinAmount = availablePoolCoinAmount.Sub(paymentAmount)
	}
	k.SetSalesHistory(ctx, pool.PoolId, pool.Round, salesHistory)
	return nil
}

func (k Keeper) getEachDistributionAmount(pool *types.Pool) sdk.Int {
	totalNftPriceAmount := pool.PoolParams.MaxNftSupply * pool.PoolParams.NftPrice.Amount.Uint64()
	return sdk.NewIntFromUint64(totalNftPriceAmount / pool.PoolParams.TargetNumData)
}
