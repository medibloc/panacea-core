package keeper

import (
	"fmt"
	"time"

	oraclekeeper "github.com/medibloc/panacea-core/v2/x/oracle/keeper"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

type (
	Keeper struct {
		cdc           codec.Codec
		storeKey      sdk.StoreKey
		memKey        sdk.StoreKey
		oracleKeeper  oraclekeeper.Keeper
		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey,
	memKey sdk.StoreKey,
	oracleKeeper oraclekeeper.Keeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,

) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		oracleKeeper:  oracleKeeper,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetOracleKeeper() oraclekeeper.Keeper {
	return k.oracleKeeper
}

func (k Keeper) AddDataVerificationQueue(ctx sdk.Context, dataHash string, dealID uint64, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetDataVerificationQueueKey(dataHash, dealID, endTime), []byte(dataHash))
}

func (k Keeper) GetAllDataVerificationQueueElements(ctx sdk.Context) ([]types.DataVerificationQueueElement, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DataVerificationQueueKey)
	defer iterator.Close()

	dataVerificationQueues := make([]types.DataVerificationQueueElement, 0)

	for ; iterator.Valid(); iterator.Next() {
		votingEndTime, dealID, dataHash, err := types.SplitDataQueueKey(iterator.Key())
		if err != nil {
			panic(err)
		}

		dataVerificationQueue := types.DataVerificationQueueElement{
			DataHash:      dataHash,
			DealId:        dealID,
			VotingEndTime: *votingEndTime,
		}

		dataVerificationQueues = append(dataVerificationQueues, dataVerificationQueue)
	}

	return dataVerificationQueues, nil
}

func (k Keeper) GetClosedDataVerificationQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.DataVerificationQueueKey, sdk.PrefixEndBytes(types.GetDataVerificationQueueKeyByTimeKey(endTime)))
}

func (k Keeper) RemoveDataVerificationQueue(ctx sdk.Context, dealID uint64, dataHash string, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDataVerificationQueueKey(dataHash, dealID, endTime))
}

func (k Keeper) IterateClosedDataVerificationQueue(ctx sdk.Context, endTime time.Time, cb func(dataSale *types.DataSale) (stop bool)) {
	iter := k.GetClosedDataVerificationQueueIterator(ctx, endTime)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		_, dealID, dataHash, _ := types.SplitDataQueueKey(iter.Key())

		dataSale, err := k.GetDataSale(ctx, dataHash, dealID)

		if err != nil {
			panic(fmt.Errorf("failed get dataSale. err: %w", err))
		}

		if cb(dataSale) {
			break
		}
	}
}

func (k Keeper) AddDataDeliveryQueue(ctx sdk.Context, dataHash string, dealID uint64, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetDataDeliveryQueueKey(dealID, dataHash, endTime), []byte(dataHash))
}

func (k Keeper) GetAllDataDeliveryQueueElements(ctx sdk.Context) ([]types.DataDeliveryQueueElement, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DataDeliveryQueueKey)
	defer iterator.Close()

	dataDeliveryQueues := make([]types.DataDeliveryQueueElement, 0)

	for ; iterator.Valid(); iterator.Next() {
		votingEndTime, dealID, dataHash, err := types.SplitDataQueueKey(iterator.Key())
		if err != nil {
			panic(err)
		}

		dataDeliveryQueue := types.DataDeliveryQueueElement{
			DataHash:      dataHash,
			DealId:        dealID,
			VotingEndTime: *votingEndTime,
		}

		dataDeliveryQueues = append(dataDeliveryQueues, dataDeliveryQueue)
	}

	return dataDeliveryQueues, nil
}

func (k Keeper) GetClosedDataDeliveryQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.DataDeliveryQueueKey, sdk.PrefixEndBytes(types.GetDataDeliveryQueueByTimeKey(endTime)))
}

func (k Keeper) RemoveDataDeliveryQueue(ctx sdk.Context, dealId uint64, dataHash string, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDataDeliveryQueueKey(dealId, dataHash, endTime))
}

func (k Keeper) IterateClosedDataDeliveryQueue(ctx sdk.Context, endTime time.Time, cb func(dataSale *types.DataSale) (stop bool)) {
	iter := k.GetClosedDataDeliveryQueueIterator(ctx, endTime)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		_, dealId, dataHash, err := types.SplitDataQueueKey(iter.Key())
		if err != nil {
			panic(err)
		}

		dataSale, err := k.GetDataSale(ctx, dataHash, dealId)

		if err != nil {
			panic(fmt.Errorf("failed get dataSale. err: %w", err))
		}

		if cb(dataSale) {
			break
		}
	}
}
