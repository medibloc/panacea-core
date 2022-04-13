package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
)

func (k Keeper) RegisterDataValidator(ctx sdk.Context, dataValidator types.DataValidator) error {
	validatorAddr, err := sdk.AccAddressFromBech32(dataValidator.Address)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(validatorAddr)
	if store.Has(dataValidatorKey) {
		return sdkerrors.Wrapf(types.ErrDataValidatorAlreadyExist, "the data validator %s is already exists", dataValidator.Address)
	}

	validatorPubKey, err := k.accountKeeper.GetPubKey(ctx, validatorAddr)
	if err != nil {
		return err
	}
	if validatorPubKey == nil {
		return sdkerrors.ErrKeyNotFound
	}

	err = k.SetDataValidator(ctx, dataValidator)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) GetAllDataValidators(ctx sdk.Context) ([]types.DataValidator, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixDataValidators)
	defer iterator.Close()

	dataValidators := make([]types.DataValidator, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dataValidator types.DataValidator

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &dataValidator)
		if err != nil {
			return []types.DataValidator{}, err
		}

		dataValidators = append(dataValidators, dataValidator)
	}

	return dataValidators, nil
}

func (k Keeper) GetDataValidator(ctx sdk.Context, dataValidatorAddress sdk.AccAddress) (types.DataValidator, error) {
	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(dataValidatorAddress)
	if !store.Has(dataValidatorKey) {
		return types.DataValidator{}, sdkerrors.Wrapf(types.ErrDataValidatorNotFound, "the data validator %s is not found", dataValidatorAddress)
	}
	bz := store.Get(dataValidatorKey)

	var dataValidator types.DataValidator
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &dataValidator)
	if err != nil {
		return types.DataValidator{}, err
	}

	return dataValidator, nil
}

func (k Keeper) SetDataValidator(ctx sdk.Context, dataValidator types.DataValidator) error {
	validatorAddr, err := sdk.AccAddressFromBech32(dataValidator.Address)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(validatorAddr)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&dataValidator)
	store.Set(dataValidatorKey, bz)
	return nil
}

func (k Keeper) isRegisteredDataValidator(ctx sdk.Context, dataValidatorAddress sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	dataValidatorKey := types.GetKeyPrefixDataValidator(dataValidatorAddress)
	return store.Has(dataValidatorKey)
}

func (k Keeper) UpdateDataValidator(ctx sdk.Context, address sdk.AccAddress, endpoint string) error {
	validator, err := k.GetDataValidator(ctx, address)
	if err != nil {
		return err
	}

	validator.Endpoint = endpoint

	err = k.SetDataValidator(ctx, validator)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) CreatePool(ctx sdk.Context, curator sdk.AccAddress, poolParams types.PoolParams) (uint64, error) {
	// Get the next pool id
	poolID := k.GetNextPoolNumberAndIncrement(ctx)

	// new pool
	newPool := types.NewPool(poolID, curator, poolParams)

	newPoolAddr := newPool.GetPoolAddress()

	// pool address for deposit
	poolAddress, err := sdk.AccAddressFromBech32(newPoolAddr)
	if err != nil {
		return 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address of pool %s", newPoolAddr)
	}

	// set new account for pool
	acc := k.accountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(poolAddress),
		newPoolAddr,
	))
	k.accountKeeper.SetAccount(ctx, acc)

	// check if the trusted_data_validators are registered
	for _, dataValidator := range poolParams.TrustedDataValidators {
		accAddr, _ := sdk.AccAddressFromBech32(dataValidator)
		if !k.isRegisteredDataValidator(ctx, accAddr) {
			return 0, sdkerrors.Wrapf(types.ErrNotRegisteredDataValidator, "the data validator %s is not registered", dataValidator)
		}
	}

	// curator send deposit to pool for creation of pool
	params := k.GetParams(ctx)

	err = k.bankKeeper.SendCoins(ctx, curator, poolAddress, sdk.NewCoins(params.DataPoolDeposit))
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrNotEnoughPoolDeposit, "The curator's balance is not enough to make a data pool")
	}

	// mint curator NFT
	nftContractAddrParam := params.DataPoolNftContractAddress
	if nftContractAddrParam == "" {
		return 0, sdkerrors.Wrapf(types.ErrNoRegisteredNFTContract, "failed to get NFT contract address")
	}

	nftContractAddr, err := sdk.AccAddressFromBech32(nftContractAddrParam)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "invalid contract address")
	}

	zeroFund, err := sdk.ParseCoinsNormalized("0umed")
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "error in parsing coin")
	}

	mintMsg := types.NewMsgMintNFT(newPool.GetPoolId(), curator.String())
	mintMsgBz, err := json.Marshal(mintMsg)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to marshal mint NFT msg")
	}

	moduleAddr := types.GetModuleAddress()

	_, err = k.wasmKeeper.Execute(ctx, nftContractAddr, moduleAddr, mintMsgBz, zeroFund)

	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to mint curator NFT")
	}

	poolName := "pool_" + strconv.FormatUint(newPool.GetPoolId(), 10)
	symbol := "DATA" + strconv.FormatUint(newPool.GetPoolId(), 10)

	instantiateMsg := types.NewInstantiateNFTMsg(poolName, symbol, newPoolAddr)
	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to marshal instantiateMsg")
	}

	codeID := k.GetParams(ctx).DataPoolCodeId

	if err != nil {
		return 0, sdkerrors.Wrapf(err, "invalid new pool address")
	}

	// instantiate NFT contract for minting data access NFT (set admin to module)
	poolNFTContractAddr, _, err := k.wasmKeeper.Instantiate(ctx, codeID, moduleAddr, poolAddress, instantiateMsgBz, "data access NFT", nil)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to instantiate contract")
	}

	newPool.NftContractAddr = poolNFTContractAddr.String()

	// store pool
	k.SetPool(ctx, newPool)

	return newPool.GetPoolId(), nil
}

func (k Keeper) GetNextPoolNumberAndIncrement(ctx sdk.Context) uint64 {
	poolNumber := k.GetNextPoolNumber(ctx)
	k.SetPoolNumber(ctx, poolNumber+1)
	return poolNumber
}

func (k Keeper) GetNextPoolNumber(ctx sdk.Context) uint64 {
	var poolNumber uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyPoolNextNumber)
	if bz == nil {
		panic(fmt.Errorf("pool has not been initialized -- Should have been done in InitGenesis"))
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &val)
		if err != nil {
			panic(err)
		}

		poolNumber = val.GetValue()
	}
	return poolNumber
}

func (k Keeper) SetPoolNumber(ctx sdk.Context, poolNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.UInt64Value{Value: poolNumber})
	store.Set(types.KeyPoolNextNumber, bz)
}

func (k Keeper) SetPool(ctx sdk.Context, pool *types.Pool) {
	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(pool.GetPoolId())
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(pool)
	store.Set(poolKey, bz)
}

func (k Keeper) GetAllPools(ctx sdk.Context) ([]types.Pool, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixPools)
	defer iterator.Close()

	pools := make([]types.Pool, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var pool types.Pool

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &pool)
		if err != nil {
			return []types.Pool{}, err
		}

		pools = append(pools, pool)
	}

	return pools, nil
}

func (k Keeper) GetPool(ctx sdk.Context, poolID uint64) (*types.Pool, error) {
	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(poolID)
	bz := store.Get(poolKey)
	if bz == nil {
		return nil, types.ErrPoolNotFound
	}
	pool := &types.Pool{}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, pool)

	return pool, nil
}

// CreateNFTContract stores NFT contract
func (k Keeper) CreateNFTContract(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte) (uint64, error) {
	// access configuration of only for creator address
	accessConfig := &wasmtypes.AccessConfig{
		Permission: wasmtypes.AccessTypeOnlyAddress,
		Address:    creator.String(),
	}

	// create contract
	codeID, err := k.wasmKeeper.Create(ctx, creator, wasmCode, accessConfig)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to create contract")
	}

	return codeID, nil
}

// DeployAndRegisterNFTContract creates, instantiate contract and store contract address. only used in test code
func (k Keeper) DeployAndRegisterNFTContract(ctx sdk.Context, wasmCode []byte) (sdk.AccAddress, error) {
	moduleAddr := types.GetModuleAddress()

	codeID, err := k.CreateNFTContract(ctx, moduleAddr, wasmCode)
	if err != nil {
		return nil, err
	}

	initMsg := types.NewInstantiateNFTMsg("curation", "CUR", moduleAddr.String())
	initMsgBz, err := json.Marshal(initMsg)
	if err != nil {
		return nil, err
	}

	// instantiate contract (set admin to module)
	contractAddr, _, err := k.wasmKeeper.Instantiate(ctx, codeID, moduleAddr, moduleAddr, initMsgBz, "curator NFT", nil)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed to instantiate contract")
	}

	return contractAddr, nil
}
