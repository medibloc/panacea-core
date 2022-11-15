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
		distrKeeper   types.DistrKeeper
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey,
	memKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	stakingKeeper types.StakingKeeper,
	distrKeeper types.DistrKeeper,
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
		distrKeeper:   distrKeeper,
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

func (k Keeper) IterateClosedOracleRegistrationQueue(ctx sdk.Context, endTime time.Time, cb func(oracleRegistration *types.OracleRegistration) (stop bool)) {
	iter := k.GetClosedOracleRegistrationQueueIterator(ctx, endTime)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		_, uniqueID, accAddr, err := types.SplitOracleRegistrationVoteQueueKey(iter.Key())
		if err != nil {
			panic(fmt.Errorf("failed split by oracleRegistfationVoteQueue key. err: %w", err))
		}

		votingTargetAddress := accAddr.String()

		oracleRegistration, err := k.GetOracleRegistration(ctx, uniqueID, votingTargetAddress)

		if err != nil {
			panic(fmt.Errorf("failed get oracleRegistration. err: %w", err))
		}

		if cb(oracleRegistration) {
			break
		}
	}
}

func (k Keeper) IterateOracleRegistrationVote(ctx sdk.Context, uniqueID, votingTargetAddress string, cb func(vote *types.OracleRegistrationVote) (stop bool)) {
	iter := k.GetOracleRegistrationVoteIterator(ctx, uniqueID, votingTargetAddress)

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var oracleRegistrationVote types.OracleRegistrationVote
		err := k.cdc.UnmarshalLengthPrefixed(bz, &oracleRegistrationVote)
		if err != nil {
			panic(fmt.Errorf("failed get oracleRegistrationVote. err: %w", err))
		}

		if cb(&oracleRegistrationVote) {
			break
		}
	}
}

func (k Keeper) IterateOracleValidator(ctx sdk.Context, cb func(info *types.OracleValidatorInfo) bool) {
	oracles, err := k.GetAllOracleList(ctx)
	if err != nil {
		panic(err)
	}

	for _, oracle := range oracles {
		accAddr, err := sdk.AccAddressFromBech32(oracle.Address)
		if err != nil {
			panic(err)
		}

		oracleValAddr := sdk.ValAddress(accAddr.Bytes())

		validator, ok := k.stakingKeeper.GetValidator(ctx, oracleValAddr)
		if !ok {
			panic(fmt.Sprintf("failed to retrieve validator information. address: %s", oracle.Address))
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

func (k Keeper) GetAllOracleRegistrationVoteQueueElements(ctx sdk.Context) ([]types.OracleRegistrationVoteQueueElement, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.OracleRegistrationsQueueKey)
	defer iterator.Close()

	queues := make([]types.OracleRegistrationVoteQueueElement, 0)

	for ; iterator.Valid(); iterator.Next() {
		votingEndTime, uniqueID, address, err := types.SplitOracleRegistrationVoteQueueKey(iterator.Key())
		if err != nil {
			return nil, err
		}

		queueElement := types.OracleRegistrationVoteQueueElement{
			UniqueId:      uniqueID,
			Address:       address,
			VotingEndTime: votingEndTime,
		}

		queues = append(queues, queueElement)
	}

	return queues, nil
}
