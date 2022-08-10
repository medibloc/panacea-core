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

		paramSpace    paramtypes.Subspace
		stakingKeeper types.StakingKeeper
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey,
	memKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	stakingKeeper types.StakingKeeper,
) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramSpace:    paramSpace,
		stakingKeeper: stakingKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) AddOracleRegistrationQueue(ctx sdk.Context, uniqueID string, addr sdk.AccAddress, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetOracleRegistrationQueueKey(uniqueID, addr, endTime), addr)
}

func (k Keeper) GetClosedOracleRegistrationQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.OracleRegistrationsQueueKey, sdk.PrefixEndBytes(types.GetOracleRegistrationVoteQueueByTimeKey(endTime)))
}

func (k Keeper) RemoveOracleRegistrationQueue(ctx sdk.Context, uniqueID string, addr sdk.AccAddress, endTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetOracleRegistrationQueueKey(uniqueID, addr, endTime))
}

func (k Keeper) IterateOracleValidator(ctx sdk.Context, cb func(info *types.OracleValidatorInfo) bool) {
	oracles, err := k.GetAllOracleList(ctx)
	if err != nil {
		panic(fmt.Sprintf(""))
	}

	for _, oracle := range oracles {
		accAddr, err := sdk.AccAddressFromBech32(oracle.Address)
		if err != nil {
			panic(fmt.Sprintf(""))
		}

		oracleValAddr := sdk.ValAddress(accAddr.Bytes())

		validator, ok := k.stakingKeeper.GetValidator(ctx, oracleValAddr)
		if !ok {
			panic(fmt.Sprintf(""))
		}

		oracleValidatorInfo := &types.OracleValidatorInfo{
			Address:         oracle.Address,
			OracleActivated: oracle.IsActivated(),
			BondedTokens:    validator.BondedTokens(),
			ValidatorJailed: validator.IsJailed(),
		}

		if cb(oracleValidatorInfo) {
			break
		}
	}
}
