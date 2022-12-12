package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) CreateDeal(ctx sdk.Context, msg *types.MsgCreateDeal) (uint64, error) {
	dealID, err := k.GetAndIncreaseNextDealNumber(ctx)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to get next deal num")
	}

	newDeal := types.NewDeal(dealID, msg)

	budget := sdk.NewCoins(*msg.Budget)

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
		"deal"+strconv.FormatUint(newDeal.Id, 10)),
	)
	k.accountKeeper.SetAccount(ctx, acc)

	consumerAccAddr, err := sdk.AccAddressFromBech32(msg.ConsumerAddress)
	if err != nil {
		return 0, nil
	}

	if err = k.bankKeeper.SendCoins(ctx, consumerAccAddr, dealAddress, budget); err != nil {
		return 0, sdkerrors.Wrapf(err, "Failed to send budget to deal account")
	}

	if err = k.SetDeal(ctx, newDeal); err != nil {
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
	store.Set(types.DealNextNumberKey, bz)
	return nil
}

func (k Keeper) GetNextDealNumber(ctx sdk.Context) (uint64, error) {
	var dealNumber uint64
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.DealNextNumberKey) {
		return 0, types.ErrDealNotInitialized
	}

	bz := store.Get(types.DealNextNumberKey)

	val := gogotypes.UInt64Value{}

	if err := k.cdc.UnmarshalLengthPrefixed(bz, &val); err != nil {
		return 0, sdkerrors.Wrapf(err, "Failed to get next deal number")
	}

	dealNumber = val.GetValue()

	return dealNumber, nil
}

func (k Keeper) GetAndIncreaseNextDealNumber(ctx sdk.Context) (uint64, error) {
	dealNumber, err := k.GetNextDealNumber(ctx)
	if err != nil {
		return 0, err
	}

	if err = k.SetNextDealNumber(ctx, dealNumber+1); err != nil {
		return 0, err
	}

	return dealNumber, nil
}

func (k Keeper) GetDeal(ctx sdk.Context, dealID uint64) (*types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetDealKey(dealID)

	bz := store.Get(dealKey)
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrDealNotFound, "deal with ID %d does not exist", dealID)
	}

	deal := &types.Deal{}
	if err := k.cdc.UnmarshalLengthPrefixed(bz, deal); err != nil {
		return nil, err
	}
	return deal, nil
}

func (k Keeper) SetDeal(ctx sdk.Context, deal *types.Deal) error {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetDealKey(deal.GetId())
	bz, err := k.cdc.MarshalLengthPrefixed(deal)
	if err != nil {
		return sdkerrors.Wrapf(err, "Failed to set deal")
	}
	store.Set(dealKey, bz)
	return nil
}

func (k Keeper) GetAllDeals(ctx sdk.Context) ([]types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DealKey)
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

func (k Keeper) IncreaseCurNumDataOfDeal(ctx sdk.Context, dealID uint64) error {
	deal, err := k.GetDeal(ctx, dealID)
	if err != nil {
		return err
	}
	deal.CurNumData = deal.CurNumData + 1
	if deal.CurNumData == deal.MaxNumData {
		deal.Status = types.DEAL_STATUS_COMPLETED
	}
	if err = k.SetDeal(ctx, deal); err != nil {
		return err
	}
	return nil
}

func (k Keeper) DeactivateDeal(ctx sdk.Context, msg *types.MsgDeactivateDeal) error {
	deal, err := k.GetDeal(ctx, msg.DealId)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrDeactivateDeal, err.Error())
	}

	if deal.ConsumerAddress != msg.RequesterAddress {
		return sdkerrors.Wrapf(types.ErrDeactivateDeal, "only consumer can deactivate the deal")
	}

	if deal.Status != types.DEAL_STATUS_ACTIVE {
		return sdkerrors.Wrapf(types.ErrDeactivateDeal, "deal's status is not 'ACTIVE'")
	}

	deal.Status = types.DEAL_STATUS_INACTIVE

	err = k.SetDeal(ctx, deal)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrDeactivateDeal, err.Error())
	}

	dealAccAddr, err := sdk.AccAddressFromBech32(deal.Address)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrDeactivateDeal, err.Error())
	}

	consumerAccAddr, err := sdk.AccAddressFromBech32(deal.ConsumerAddress)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrDeactivateDeal, err.Error())
	}

	// refund a budget to consumer
	err = k.bankKeeper.SendCoins(ctx, dealAccAddr, consumerAccAddr, sdk.NewCoins(*deal.Budget))
	if err != nil {
		return sdkerrors.Wrapf(types.ErrDeactivateDeal, err.Error())
	}

	return nil
}
