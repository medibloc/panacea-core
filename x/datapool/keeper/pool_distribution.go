package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

// addSalesHistory stores sales information in history.
func (k Keeper) addSalesHistory(ctx sdk.Context, poolID, round uint64, addr sdk.AccAddress, dataHash []byte) {
	salesHistory := k.GetSalesHistory(ctx, poolID, round, addr.String())
	if salesHistory == nil {
		salesHistory = &types.SalesHistory{
			PoolId:        poolID,
			Round:         round,
			SellerAddress: addr.String(),
			DataHashes:    [][]byte{dataHash},
			PaidCoin:      &types.ZeroFund,
		}
	} else {
		salesHistory.AddDataHash(dataHash)
	}

	k.SetSalesHistory(ctx, salesHistory)
}

// SetSalesHistory stores sales history.
func (k Keeper) SetSalesHistory(ctx sdk.Context, salesHistory *types.SalesHistory) {
	key := types.GetKeyPrefixSalesHistory(
		salesHistory.PoolId,
		salesHistory.Round,
		salesHistory.SellerAddress,
	)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, k.cdc.MustMarshalLengthPrefixed(salesHistory))
}

// GetSalesHistory returns the sales history. If there is no value, it responds nil.
func (k Keeper) GetSalesHistory(ctx sdk.Context, poolID, round uint64, sellerAddress string) *types.SalesHistory {
	key := types.GetKeyPrefixSalesHistory(poolID, round, sellerAddress)
	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return nil
	}
	bz := store.Get(key)
	var salesHistory types.SalesHistory
	k.cdc.MustUnmarshalLengthPrefixed(bz, &salesHistory)
	return &salesHistory
}

func (k Keeper) GetSalesHistories(ctx sdk.Context, poolID, round uint64) []*types.SalesHistory {
	key := types.GetKeyPrefixSalesHistories(poolID, round)
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, key)
	var histories []*types.SalesHistory
	for ; iter.Valid(); iter.Next() {
		history := &types.SalesHistory{}
		bz := iter.Value()
		k.cdc.MustUnmarshalLengthPrefixed(bz, history)
		histories = append(histories, history)
	}
	return histories
}

func (k Keeper) GetAllSalesHistories(ctx sdk.Context) []*types.SalesHistory {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixSalesHistory)
	var histories []*types.SalesHistory
	for ; iter.Valid(); iter.Next() {
		history := &types.SalesHistory{}
		bz := iter.Value()
		k.cdc.MustUnmarshalLengthPrefixed(bz, history)
		histories = append(histories, history)
	}
	return histories
}

// SetInstantRevenueDistribution stores the poolID to which the revenue should be distributed immediately.
func (k Keeper) SetInstantRevenueDistribution(ctx sdk.Context, instantRevenueDistribution *types.InstantRevenueDistribution) {
	bz := k.cdc.MustMarshalLengthPrefixed(instantRevenueDistribution)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixInstantRevenueDistribution, bz)
}

func (k Keeper) GetInstantRevenueDistribution(ctx sdk.Context) *types.InstantRevenueDistribution {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.KeyPrefixInstantRevenueDistribution) {
		return &types.InstantRevenueDistribution{}
	}
	bz := store.Get(types.KeyPrefixInstantRevenueDistribution)
	var delayedRevenueDistribution types.InstantRevenueDistribution
	k.cdc.MustUnmarshalLengthPrefixed(bz, &delayedRevenueDistribution)
	return &delayedRevenueDistribution
}

// addInstantRevenueDistribution adds the poolID to distribution. If there are duplicate poolIDs, they are not added.
func (k Keeper) addInstantRevenueDistribution(ctx sdk.Context, poolID uint64) {
	instantRevenueDistribution := k.GetInstantRevenueDistribution(ctx)
	instantRevenueDistribution.AppendPoolID(poolID)
	k.SetInstantRevenueDistribution(ctx, instantRevenueDistribution)
}

func (k Keeper) DistributionRevenuePools(ctx sdk.Context) error {
	instantRevenueDistribution := k.GetInstantRevenueDistribution(ctx)

	if instantRevenueDistribution.IsEmpty() {
		return nil
	}

	// This lastIndex is to execute the following iteration in the next block if it does not end in one block.
	// But We're currently processing it all in one block. :)
	lastIndex := 0
	// TODO We need to think about how to deal with the pool where distribution has failed.
	for i, poolID := range instantRevenueDistribution.PoolIds {
		lastIndex = i
		// search current pool info
		pool, err := k.GetPool(ctx, poolID)
		if err != nil {
			return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
		}

		// send deposit to curator
		err = k.sendDepositToCurator(ctx, pool)
		if err != nil {
			return err
		}

		// distribution of revenue to each seller
		err = k.executeRevenueDistribution(ctx, pool)
		if err != nil {
			return err
		}
	}
	// delete this pool from instantRevenueDistribution
	instantRevenueDistribution.TruncateFromBeginning(lastIndex)
	k.SetInstantRevenueDistribution(ctx, instantRevenueDistribution)

	return nil
}

// sendDepositToCurator returns the deposit if the pool status is ACTIVE but the deposit has not yet been returned.
func (k Keeper) sendDepositToCurator(ctx sdk.Context, pool *types.Pool) error {
	if pool.Status != types.ACTIVE || pool.WasDepositReturned {
		return nil
	}

	curatorAddr, err := sdk.AccAddressFromBech32(pool.Curator)
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
	}
	poolAddr, err := sdk.AccAddressFromBech32(pool.PoolAddress)
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
	}
	err = k.bankKeeper.SendCoins(ctx, poolAddr, curatorAddr, sdk.NewCoins(pool.Deposit))
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
	}
	pool.WasDepositReturned = true
	k.SetPool(ctx, pool)

	return nil
}

func (k Keeper) executeRevenueDistribution(ctx sdk.Context, pool *types.Pool) error {
	availablePoolCoinAmount, err := k.getAvailablePoolCoinAmount(ctx, pool)
	if err != nil {
		return err
	}

	// if there is no coin available, proceed no further.
	if availablePoolCoinAmount.Equal(sdk.NewInt(0)) {
		return nil
	}

	// calculate the revenue to be sent to each seller and curator
	eachSellerDistributionAmount, curatorCommissionAmount := k.getEachDistributionAmount(pool)

	err = k.distributionToEachSeller(ctx, pool, eachSellerDistributionAmount)
	if err != nil {
		return err
	}

	err = k.sendCommissionToCurator(ctx, pool, curatorCommissionAmount)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) getAvailablePoolCoinAmount(ctx sdk.Context, pool *types.Pool) (*sdk.Int, error) {
	poolAddr, err := sdk.AccAddressFromBech32(pool.PoolAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
	}
	poolBalance := k.bankKeeper.GetBalance(ctx, poolAddr, assets.MicroMedDenom)
	if pool.WasDepositReturned {
		return &poolBalance.Amount, nil
	} else {
		amount := poolBalance.Amount.Sub(pool.Deposit.Amount)
		return &amount, nil
	}
}

func (k Keeper) distributionToEachSeller(ctx sdk.Context, pool *types.Pool, eachSellerDistributionAmount sdk.Int) error {
	poolAddress, err := sdk.AccAddressFromBech32(pool.GetPoolAddress())
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
	}

	salesHistories := k.GetSalesHistories(ctx, pool.PoolId, pool.Round)
	for _, history := range salesHistories {
		paymentAmount := calculatePaymentAmountToSeller(history, eachSellerDistributionAmount)
		paymentCoin := sdk.NewCoin(assets.MicroMedDenom, paymentAmount)

		sellerAddr, err := sdk.AccAddressFromBech32(history.SellerAddress)
		if err != nil {
			return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
		}

		err = k.bankKeeper.SendCoins(
			ctx,
			poolAddress,
			sellerAddr,
			sdk.NewCoins(paymentCoin))
		if err != nil {
			return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
		}
		*history.PaidCoin = history.PaidCoin.Add(paymentCoin)

		k.SetSalesHistory(ctx, history)
	}
	return nil
}

// calculatePaymentAmountToSeller calculates the amount to be paid to the seller.
func calculatePaymentAmountToSeller(history *types.SalesHistory, eachSellerDistributionAmount sdk.Int) sdk.Int {
	paidCoinAmount := history.PaidCoin.Amount
	dataCount := sdk.NewInt(int64(history.DataCount()))
	distributedAmount := eachSellerDistributionAmount.Mul(dataCount)
	paymentAmount := distributedAmount.Sub(paidCoinAmount)
	return paymentAmount
}

// getEachDistributionAmount returns the amount to be distributed to each seller and curator.
func (k Keeper) getEachDistributionAmount(pool *types.Pool) (sdk.Int, sdk.Int) {
	nftPriceAmount := pool.PoolParams.NftPrice.Amount.ToDec()
	numIssuedNFTs := sdk.NewIntFromUint64(pool.NumIssuedNfts).ToDec()
	targetNumData := sdk.NewIntFromUint64(pool.PoolParams.TargetNumData).ToDec()
	curatorCommissionRate := pool.CuratorCommissionRate

	// ((nftPriceAmount * numIssuedNFTs) / targetNumData) * (1 - curatorCommissionRate)
	sellerAmount := nftPriceAmount.Mul(numIssuedNFTs).Quo(targetNumData).Mul(sdk.NewDec(1).Sub(curatorCommissionRate)).TruncateInt()
	// nftPriceAmount * numIssuedNFTs * curatorCommissionRate
	curatorCommissionAmount := nftPriceAmount.Mul(numIssuedNFTs).Mul(curatorCommissionRate).TruncateInt()

	return sellerAmount, curatorCommissionAmount
}

// sendCommissionToCurator sends a commission to the curator
func (k Keeper) sendCommissionToCurator(ctx sdk.Context, pool *types.Pool, curatorCommissionAmount sdk.Int) error {
	poolAddress, err := sdk.AccAddressFromBech32(pool.GetPoolAddress())
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
	}

	paidCuratorCommissionAmount := pool.CuratorCommission[pool.Round].Amount
	paymentAmount := curatorCommissionAmount.Sub(paidCuratorCommissionAmount)
	paymentCoin := sdk.NewCoin(assets.MicroMedDenom, paymentAmount)

	curatorAddr, err := sdk.AccAddressFromBech32(pool.Curator)
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
	}

	err = k.bankKeeper.SendCoins(
		ctx,
		poolAddress,
		curatorAddr,
		sdk.NewCoins(paymentCoin))
	if err != nil {
		return sdkerrors.Wrap(types.ErrRevenueDistribution, err.Error())
	}

	pool.CuratorCommission[pool.Round] = paymentCoin.Add(sdk.NewCoin(assets.MicroMedDenom, paidCuratorCommissionAmount))
	k.SetPool(ctx, pool)

	return nil
}
