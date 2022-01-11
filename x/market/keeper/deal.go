package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/medibloc/panacea-core/v2/x/market/types"
)

func (k Keeper) CreateNewDeal(ctx sdk.Context, owner sdk.AccAddress, deal types.Deal) (uint64, error) {
	dealId := k.GetNextDealNumberAndIncrement(ctx)

	newDeal, err := k.newDeal(dealId, deal)
	if err != nil {
		return 0, err
	}

	var coins sdk.Coins
	coins = append(coins, *deal.GetBudget())

	acc := k.accountKeeper.GetAccount(ctx, sdk.AccAddress(newDeal.GetDealAddress()))
	if acc != nil {
		return 0, sdkerrors.Wrapf(types.ErrDealAlreadyExist, "deal %d already exist", dealId)
	}

	err = k.SetDeal(ctx, newDeal)

	acc = k.accountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(
			sdk.AccAddress(newDeal.GetDealAddress()),
		),
		newDeal.DealAddress),
	)
	k.accountKeeper.SetAccount(ctx, acc)

	err = k.bankKeeper.SendCoins(ctx, owner, sdk.AccAddress(newDeal.GetDealAddress()), coins)
	return newDeal.GetDealId(), nil
}

func (k Keeper) newDeal(dealId uint64, deal types.Deal) (types.Deal, error) {

	dealAddress := types.NewDealAddress(dealId)

	newDeal := &types.Deal{
		DealId:               dealId,
		DealAddress:          dealAddress.String(),
		DataSchema:           deal.GetDataSchema(),
		Budget:               deal.GetBudget(),
		TrustedDataValidator: deal.GetTrustedDataValidator(),
		WantDataCount:        deal.GetWantDataCount(),
		CompleteDataCount:    0,
		Owner:                deal.GetOwner(),
		Status:               "ACTIVE",
	}

	return *newDeal, nil
}

func (k Keeper) SetNextDealNumber(ctx sdk.Context, dealNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.UInt64Value{Value: dealNumber})
	store.Set(types.KeyDealNextNumber, bz)
}

func (k Keeper) GetNextDealNumberAndIncrement(ctx sdk.Context) uint64 {
	var dealNumber uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyDealNextNumber)
	if bz == nil {
		panic(fmt.Errorf("deal has not been initialized -- Should have been done in InitGenesis"))
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &val)
		if err != nil {
			panic(err)
		}

		dealNumber = val.GetValue()
	}

	k.SetNextDealNumber(ctx, dealNumber+1)
	return dealNumber
}

func (k Keeper) GetDeal(ctx sdk.Context, dealId uint64) (types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetKeyPrefixDeals(dealId)
	if !store.Has(dealKey) {
		return types.Deal{}, fmt.Errorf("deal with ID %d does not exist", dealId)
	}

	bz := store.Get(dealKey)

	var deal types.Deal
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &deal)
	if err != nil {
		return types.Deal{}, err
	}

	return deal, nil
}

func (k Keeper) SetDeal(ctx sdk.Context, deal types.Deal) error {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetKeyPrefixDeals(deal.GetDealId())
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&deal)
	store.Set(dealKey, bz)

	return nil
}
