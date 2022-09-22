package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) SellData(ctx sdk.Context, msg *types.MsgSellData) error {
	_, err := sdk.AccAddressFromBech32(msg.SellerAddress)
	if err != nil {
		return err
	}

	//TODO: Check the Deal which seller want to sell data to the deal exists
	//k.GetDeal()

	if err = k.checkDataSaleStatus(ctx, msg.VerifiableCid, msg.DealId); err != nil {
		return sdkerrors.Wrapf(types.ErrSellData, err.Error())
	}

	getDataSale, _ := k.GetDataSale(ctx, msg.VerifiableCid, msg.DealId)
	if getDataSale != nil {
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

func (k Keeper) GetDataSale(ctx sdk.Context, verifiableCID []byte, dealID uint64) (*types.DataSale, error) {
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

func (k Keeper) checkDataSaleStatus(ctx sdk.Context, verifiableCID []byte, dealID uint64) error {
	existDataSale, err := k.GetDataSale(ctx, verifiableCID, dealID)
	if err != nil {
		return err
	}

	switch existDataSale.Status {
	case types.DATA_SALE_STATUS_COMPLETED:
		return nil
	case types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD:
		return fmt.Errorf("in verification voting period")
	case types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD:
		return fmt.Errorf("in delivery voting period")
	case types.DATA_SALE_STATUS_FAILED:
		return nil
	default:
		return fmt.Errorf("unexpected state. status: %s", existDataSale.Status)
	}
}
