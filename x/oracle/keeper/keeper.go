package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc      codec.Codec
		storeKey sdk.StoreKey
		memKey   sdk.StoreKey

		paramSpace paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey,
	memKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramSpace: paramSpace,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) AddOracleRegistrationVoteQueue(ctx sdk.Context, uniqueID string, addr sdk.AccAddress, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetOracleRegistrationVoteQueueKey(uniqueID, addr, endTime), addr)
}

func (k Keeper) GetClosedOracleRegistrationVoteQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.OracleRegistrationVotesQueueKey, sdk.PrefixEndBytes(types.GetOracleRegistrationVoteQueueByTimeKey(endTime)))
}

func (k Keeper) RemoveOracleRegistrationVoteQueue(ctx sdk.Context, uniqueID string, addr sdk.AccAddress, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOracleRegistrationVoteQueueKey(uniqueID, addr, endTime))
}

func (k Keeper) IterateEndOracleRegistrationVotesQueue(ctx sdk.Context, endTime time.Time, cb func() (stop bool)) {
	iter := k.GetEndOracleRegistrationVoteQueueIterator(ctx, endTime)

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {

	}
}
