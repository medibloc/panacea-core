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

func (k Keeper) AddDataVerificationQueue(ctx sdk.Context, verifiableCID string, dealID uint64, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetDataVerificationQueueKey(verifiableCID, dealID, endTime), []byte(verifiableCID))
}

func (k Keeper) GetClosedDataVerificationQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.DataVerificationQueueKey, sdk.PrefixEndBytes(types.GetDataVerificationQueueKeyByTimeKey(endTime)))
}

func (k Keeper) RemoveDataVerificationQueue(ctx sdk.Context, dealID uint64, verifiableCID string, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetDataVerificationQueueKey(verifiableCID, dealID, endTime))
}

func (k Keeper) IterateClosedDataVerificationQueue(ctx sdk.Context, endTime time.Time, cb func(dataSale *types.DataSale) (stop bool)) {
	iter := k.GetClosedDataVerificationQueueIterator(ctx, endTime)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		dealID, verifiableCID := types.SplitDataVerificationQueueKey(iter.Key())

		dataSale, err := k.GetDataSale(ctx, verifiableCID, dealID)
		if err != nil {
			panic(fmt.Errorf("failed get dataSale. err: %w", err))
		}

		if cb(dataSale) {
			break
		}
	}
}

func (k Keeper) AddDataDeliveryQueue(ctx sdk.Context, verifiableCID string, dealID uint64, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetDataDeliveryQueueKey(dealID, verifiableCID, endTime), []byte(verifiableCID))
}
