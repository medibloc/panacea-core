package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
		return 0, sdkerrors.Wrapf(err, "invalid address of pool %s", newPoolAddr)
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
		return 0, sdkerrors.Wrapf(err, "the balance is not enough to make a data pool")
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

	mintMsg := types.NewMsgMintCuratorNFT(newPool.GetPoolId(), curator.String())
	mintMsgBz, err := json.Marshal(mintMsg)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to marshal mint NFT msg")
	}

	moduleAddr := types.GetModuleAddress()

	_, err = k.wasmKeeper.Execute(ctx, nftContractAddr, moduleAddr, mintMsgBz, sdk.NewCoins(types.ZeroFund))
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to mint curator NFT")
	}

	poolName := "data_pool_" + strconv.FormatUint(newPool.GetPoolId(), 10)
	symbol := "DATA" + strconv.FormatUint(newPool.GetPoolId(), 10)

	instantiateMsg := types.NewInstantiateNFTMsg(poolName, symbol, newPoolAddr)
	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to marshal instantiate contract Msg")
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

	// mint tokens as many as targetNumData
	k.setInitialSupply(ctx, poolID)
	err = k.mintPoolShareToken(ctx, poolID, poolParams.TargetNumData)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to mint share token")
	}

	return newPool.GetPoolId(), nil
}

// setInitialSupply defines supply to be initialized for tokens to be minted.
func (k Keeper) setInitialSupply(ctx sdk.Context, poolID uint64) {
	supply := banktypes.Supply{
		Total: sdk.NewCoins(types.GetAccumPoolShareToken(poolID, 0)),
	}

	k.bankKeeper.SetSupply(ctx, &supply)
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

// SellData verifies the certificate against the pool information and stores it using a key combined with poolID, dataHash, and round.
// Seller is paid shareToken.
func (k Keeper) SellData(ctx sdk.Context, seller sdk.AccAddress, cert types.DataValidationCertificate) (*sdk.Coin, error) {
	if cert.UnsignedCert.Requester != seller.String() {
		return nil, types.ErrNotEqualsSeller
	}

	if err := k.verifySignature(ctx, cert); err != nil {
		return nil, err
	}

	if k.isDuplicatedCertificate(ctx, cert) {
		return nil, types.ErrExistSameDataHash
	}

	pool, err := k.GetPool(ctx, cert.UnsignedCert.PoolId)
	if err != nil {
		return nil, err
	}

	if err := k.validateCertificateByPool(cert, pool); err != nil {
		return nil, err
	}

	k.SetDataValidationCertificate(ctx, cert)

	k.increaseCurNumAndUpdatePool(ctx, pool)

	shareToken := types.GetAccumPoolShareToken(pool.PoolId, 1)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seller, sdk.NewCoins(shareToken))
	if err != nil {
		return nil, err
	}

	return &shareToken, nil
}

// verifySignature verifies that the signature of the dataValidator is correct
func (k Keeper) verifySignature(ctx sdk.Context, dataValidatorCert types.DataValidationCertificate) error {
	dataValidator := dataValidatorCert.UnsignedCert.DataValidator
	unsignedCert := dataValidatorCert.UnsignedCert
	sign := dataValidatorCert.Signature

	valAddr, err := sdk.AccAddressFromBech32(dataValidator)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalidSignature, err.Error())
	}
	pubKey, err := k.accountKeeper.GetPubKey(ctx, valAddr)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalidSignature, err.Error())
	}
	unsignedCertBinary, err := k.cdc.MarshalBinaryBare(unsignedCert)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalidSignature, err.Error())
	}
	if !pubKey.VerifySignature(unsignedCertBinary, sign) {
		return sdkerrors.Wrap(types.ErrInvalidSignature, "invalid signature")
	}
	return nil
}

// isDuplicatedCertificate goes through all rounds and checks for data duplication.
func (k Keeper) isDuplicatedCertificate(ctx sdk.Context, cert types.DataValidationCertificate) bool {
	store := ctx.KVStore(k.storeKey)
	unsignedCert := cert.UnsignedCert
	for round := uint64(1); round <= unsignedCert.Round; round++ {
		key := types.GetKeyPrefixDataValidateCert(unsignedCert.GetPoolId(), round, unsignedCert.GetDataHash())
		if store.Has(key) {
			return true
		}
	}
	return false
}

// validateCertificateByPool verifies the pool and certificate data
func (k Keeper) validateCertificateByPool(cert types.DataValidationCertificate, pool *types.Pool) error {
	validator := cert.UnsignedCert.DataValidator
	trustedValidators := pool.PoolParams.TrustedDataValidators

	if !contains(trustedValidators, validator) {
		return sdkerrors.Wrap(types.ErrInvalidDataValidationCert, "the data validator is not trusted")
	}

	if pool.Status != types.PENDING {
		return sdkerrors.Wrap(types.ErrInvalidDataValidationCert, "the status of the pool is not 'PENDING'")
	}

	if pool.Round != cert.UnsignedCert.Round {
		return sdkerrors.Wrap(types.ErrInvalidDataValidationCert, fmt.Sprintf("pool round do not matched. pool round: %v", pool.Round))
	}

	return nil
}

func (k Keeper) SetDataValidationCertificate(ctx sdk.Context, cert types.DataValidationCertificate) {
	unsignedCert := cert.UnsignedCert
	key := types.GetKeyPrefixDataValidateCert(unsignedCert.PoolId, unsignedCert.Round, unsignedCert.DataHash)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, k.cdc.MustMarshalBinaryLengthPrefixed(&cert))
}

func (k Keeper) increaseCurNumAndUpdatePool(ctx sdk.Context, pool *types.Pool) {
	pool.CurNumData += 1

	if pool.CurNumData == pool.PoolParams.TargetNumData {
		pool.Status = types.ACTIVE
	}

	k.SetPool(ctx, pool)
}

func (k Keeper) mintPoolShareToken(ctx sdk.Context, poolID, amount uint64) error {
	shareToken := types.GetAccumPoolShareToken(poolID, amount)
	shareTokens := sdk.NewCoins(shareToken)
	err := k.bankKeeper.MintCoins(ctx, types.ModuleName, shareTokens)
	if err != nil {
		return sdkerrors.Wrap(types.ErrFailedMintShareToken, err.Error())
	}

	return nil
}

func contains(validators []string, validator string) bool {
	for _, v := range validators {
		if validator == v {
			return true
		}
	}
	return false
}

func (k Keeper) GetDataValidationCertificate(ctx sdk.Context, poolID, round uint64, dataHash []byte) (types.DataValidationCertificate, error) {
	key := types.GetKeyPrefixDataValidateCert(poolID, round, dataHash)
	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return types.DataValidationCertificate{}, sdkerrors.Wrap(types.ErrGetDataValidationCert, "certification is not exist")
	}

	bz := store.Get(key)

	var cert types.DataValidationCertificate
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &cert)
	if err != nil {
		return types.DataValidationCertificate{}, sdkerrors.Wrap(types.ErrGetDataValidationCert, err.Error())
	}

	return cert, nil
}

func (k Keeper) BuyDataAccessNFT(ctx sdk.Context, buyer sdk.AccAddress, poolID, round uint64, payment sdk.Coin) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to get pool %d", poolID)
	}

	if pool.GetNumIssuedNfts() == pool.GetPoolParams().GetMaxNftSupply() {
		return types.ErrNFTAllIssued
	}

	if pool.GetRound() != round {
		return types.ErrRoundNotMatched
	}

	if !payment.Equal(pool.GetPoolParams().GetNftPrice()) {
		return types.ErrPaymentNotMatched
	}

	poolAcc, err := sdk.AccAddressFromBech32(pool.GetPoolAddress())
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid pool address")
	}

	err = k.bankKeeper.SendCoins(ctx, buyer, poolAcc, sdk.NewCoins(payment))
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to pay")
	}

	//mint data access NFT when pool is activated
	if pool.GetStatus() == types.ACTIVE {
		contractAddr := pool.GetNftContractAddr()

		contractAcc, err := sdk.AccAddressFromBech32(contractAddr)
		if err != nil {
			return sdkerrors.Wrapf(err, "invalid NFT contract address")
		}

		mintMsg := types.NewMsgMintDataAccessNFT(pool.GetNumIssuedNfts()+1, buyer.String())
		mintMsgBz, err := json.Marshal(mintMsg)
		if err != nil {
			return sdkerrors.Wrapf(err, "failed to marshal mint NFT Msg")
		}

		_, err = k.wasmKeeper.Execute(ctx, contractAcc, poolAcc, mintMsgBz, sdk.NewCoins(types.ZeroFund))
		if err != nil {
			return sdkerrors.Wrapf(err, "failed to mint NFT")
		}
	} else {
		// if pool is PENDING state, add buyer to white list
		k.AddToWhiteList(ctx, poolID, buyer)
	}

	k.increaseIssuedNFT(ctx, pool)

	return nil
}

func (k Keeper) increaseIssuedNFT(ctx sdk.Context, pool *types.Pool) {
	pool.NumIssuedNfts += 1

	k.SetPool(ctx, pool)
}

func (k Keeper) AddToWhiteList(ctx sdk.Context, poolID uint64, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyWhiteList(poolID, addr)
	whiteList := &types.WhiteList{
		PoolId:  poolID,
		Address: addr.String(),
	}

	store.Set(key, k.cdc.MustMarshalBinaryLengthPrefixed(whiteList))
}

func (k Keeper) GetAllWhiteLists(ctx sdk.Context) ([]types.WhiteList, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixPoolWhiteList)
	defer iterator.Close()

	whiteLists := make([]types.WhiteList, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var list types.WhiteList

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &list)
		if err != nil {
			return []types.WhiteList{}, err
		}

		whiteLists = append(whiteLists, list)
	}

	return whiteLists, nil
}

func (k Keeper) GetWhiteList(ctx sdk.Context, poolID uint64) ([]types.WhiteList, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetKeyPoolWhiteList(poolID))
	defer iterator.Close()

	whiteLists := make([]types.WhiteList, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var list types.WhiteList

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &list)
		if err != nil {
			return []types.WhiteList{}, err
		}

		whiteLists = append(whiteLists, list)
	}

	return whiteLists, nil
}
