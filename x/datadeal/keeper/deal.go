package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) CreateDeal(ctx sdk.Context, buyerAddress sdk.AccAddress, msg *types.MsgCreateDeal) (uint64, error) {

	dealID, err := k.GetNextDealNumberAndIncrement(ctx)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to get next deal num")
	}

	newDeal := types.NewDeal(dealID, msg)

	coins := sdk.NewCoins(*msg.Budget)

	dealAddress, err := sdk.AccAddressFromBech32(newDeal.Address)
	if err != nil {
		return 0, err
	}

	acc := k.accountKeeper.GetAccount(ctx, dealAddress)
	if acc != nil {
		return 0, sdkerrors.Wrapf(types.ErrDealAlreadyExist, "deal %d already exist", dealID)
	}

	acc = k.accountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(
			dealAddress,
		),
		strconv.FormatUint(newDeal.Id, 10)),
	)
	k.accountKeeper.SetAccount(ctx, acc)

	if err = k.bankKeeper.SendCoins(ctx, buyerAddress, dealAddress, coins); err != nil {
		return 0, sdkerrors.Wrapf(err, "Failed to send coins to deal account")
	}

	if err = k.SetDeal(ctx, *newDeal); err != nil {
		return 0, err
	}

	return newDeal.Id, nil
}

func (k Keeper) SetNextDealNumber(ctx sdk.Context, dealNumber uint64) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalLengthPrefixed(&gogotypes.UInt64Value{Value: dealNumber})
	if err != nil {
		return sdkerrors.Wrapf(err, "Failed to set next deal number")
	}
	store.Set(types.KeyDealNextNumber, bz)
	return nil
}

func (k Keeper) GetNextDealNumber(ctx sdk.Context) (uint64, error) {
	var dealNumber uint64
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.KeyDealNextNumber) {
		return 0, types.ErrDealNotInitialized
	}

	bz := store.Get(types.KeyDealNextNumber)

	val := gogotypes.UInt64Value{}

	if err := k.cdc.UnmarshalLengthPrefixed(bz, &val); err != nil {
		return 0, sdkerrors.Wrapf(err, "Failed to get next deal number")
	}

	dealNumber = val.GetValue()

	return dealNumber, nil
}

func (k Keeper) GetNextDealNumberAndIncrement(ctx sdk.Context) (uint64, error) {
	dealNumber, err := k.GetNextDealNumber(ctx)
	if err != nil {
		return 0, err
	}

	if err = k.SetNextDealNumber(ctx, dealNumber+1); err != nil {
		return 0, err
	}

	return dealNumber, nil
}

func (k Keeper) GetDeal(ctx sdk.Context, dealID uint64) (types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetDealKey(dealID)

	bz := store.Get(dealKey)
	if bz == nil {
		return types.Deal{}, sdkerrors.Wrapf(types.ErrDealNotFound, "deal with ID %d does not exist", dealID)
	}

	var deal types.Deal
	if err := k.cdc.UnmarshalLengthPrefixed(bz, &deal); err != nil {
		return types.Deal{}, err
	}
	return deal, nil
}

func (k Keeper) SetDeal(ctx sdk.Context, deal types.Deal) error {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetDealKey(deal.GetId())
	bz, err := k.cdc.MarshalLengthPrefixed(&deal)
	if err != nil {
		return sdkerrors.Wrapf(err, "Failed to set deal")
	}
	store.Set(dealKey, bz)
	return nil
}

func (k Keeper) GetAllDeals(ctx sdk.Context) ([]types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixDeals)
	defer iterator.Close()

	deals := make([]types.Deal, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var deal types.Deal

		if err := k.cdc.UnmarshalLengthPrefixed(bz, &deal); err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetDeal, err.Error())
		}

		deals = append(deals, deal)
	}

	return deals, nil
}

func (k Keeper) SellData(ctx sdk.Context, msg *types.MsgSellData) error {
	_, err := sdk.AccAddressFromBech32(msg.SellerAddress)
	if err != nil {
		return err
	}

	deal, err := k.GetDeal(ctx, msg.DealId)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrSellData, err.Error())
	}

	if deal.Status != types.DEAL_STATUS_ACTIVE {
		return sdkerrors.Wrapf(types.ErrSellData, "deal status is not ACTIVE")
	}

	getDataSale, _ := k.GetDataSale(ctx, msg.VerifiableCid, msg.DealId)
	if getDataSale != nil && getDataSale.Status != types.DATA_SALE_STATUS_FAILED {
		return sdkerrors.Wrapf(types.ErrSellData, "data already exists")
	}

	dataSale := types.NewDataSale(msg)
	dataSale.VotingPeriod = k.oracleKeeper.GetVotingPeriod(ctx)

	if err := k.SetDataSale(ctx, dataSale); err != nil {
		return sdkerrors.Wrapf(types.ErrSellData, err.Error())
	}

	//TODO: Add DataSale to VoteDataSale Queue
	//k.AddDataSaleQueue()

	return nil
}

func (k Keeper) GetDataSale(ctx sdk.Context, verifiableCID string, dealID uint64) (*types.DataSale, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDataSaleKey(verifiableCID, dealID)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrDataSaleNotFound
	}

	dataSale := &types.DataSale{}

	err := k.cdc.UnmarshalLengthPrefixed(bz, dataSale)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrGetDataSale, err.Error())
	}

	return dataSale, nil
}

func (k Keeper) SetDataSale(ctx sdk.Context, dataSale *types.DataSale) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDataSaleKey(dataSale.VerifiableCid, dataSale.DealId)

	bz, err := k.cdc.MarshalLengthPrefixed(dataSale)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetAllDataSaleList(ctx sdk.Context) ([]types.DataSale, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DataSaleKey)
	defer iterator.Close()

	dataSales := make([]types.DataSale, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dataSale types.DataSale

		err := k.cdc.UnmarshalLengthPrefixed(bz, &dataSale)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetDataSale, err.Error())
		}

		dataSales = append(dataSales, dataSale)
	}

	return dataSales, nil
}
