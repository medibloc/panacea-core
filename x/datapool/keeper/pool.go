package keeper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/medibloc/panacea-core/v2/x/datapool/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (k Keeper) CreatePool(ctx sdk.Context, curator sdk.AccAddress, deposit sdk.Coin, poolParams types.PoolParams) (uint64, error) {
	if len(poolParams.GetTrustedOracles()) == 0 {
		return 0, types.ErrNoTrustedOracle
	}
	// Get the next pool id
	poolID := k.GetNextPoolNumberAndIncrement(ctx)

	// new pool
	newPool := types.NewPool(poolID, curator, poolParams)

	newPoolAddr := newPool.GetPoolAddress()

	// pool address for deposit
	poolAddress, err := sdk.AccAddressFromBech32(newPoolAddr)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrCreatePool, "invalid address of pool %s", newPoolAddr)
	}

	// set new account for pool
	acc := k.accountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(poolAddress),
		newPoolAddr,
	))
	k.accountKeeper.SetAccount(ctx, acc)

	// check if the trusted_oracles are registered
	for _, oracle := range poolParams.TrustedOracles {
		accAddr, err := sdk.AccAddressFromBech32(oracle)
		if err != nil {
			return 0, err
		}

		if !k.oracleKeeper.IsRegisteredOracle(ctx, accAddr) {
			return 0, sdkerrors.Wrapf(oracletypes.ErrNotRegisteredOracle, "oracle %s is not registered", accAddr)
		}
	}

	// curator deposit check.
	params := k.GetParams(ctx)

	NFTPriceDec := poolParams.GetNftPrice().Amount.ToDec()
	NFTTotalSupplyDec := sdk.NewDecFromInt(sdk.NewIntFromUint64(poolParams.GetMaxNftSupply()))
	expectedTotalSalesDec := NFTPriceDec.Mul(NFTTotalSupplyDec)
	requiredDeposit := expectedTotalSalesDec.Mul(params.DataPoolCommissionRate)
	if deposit.Amount.ToDec().LT(requiredDeposit) {
		return 0, types.ErrNotEnoughPoolDeposit
	}

	newPool.Deposit = deposit
	newPool.CuratorCommissionRate = params.DataPoolCommissionRate
	newPool.CuratorCommission[newPool.Round] = types.ZeroFund

	// send deposit to pool
	err = k.bankKeeper.SendCoins(ctx, curator, poolAddress, sdk.NewCoins(deposit))
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrCreatePool, err.Error())
	}

	// mint curator NFT
	nftContractAddrParam := params.DataPoolNftContractAddress
	if nftContractAddrParam == "" {
		return 0, types.ErrNoRegisteredNFTContract
	}

	nftContractAddr, err := sdk.AccAddressFromBech32(nftContractAddrParam)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrCreatePool, "invalid contract address: %s", nftContractAddrParam)
	}

	mintMsg := types.NewMsgMintCuratorNFT(newPool.GetPoolId(), curator.String())
	mintMsgBz, err := json.Marshal(mintMsg)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrMintNFT, err.Error())
	}

	moduleAddr := types.GetModuleAddress()

	_, err = k.wasmKeeper.Execute(ctx, nftContractAddr, moduleAddr, mintMsgBz, sdk.NewCoins(types.ZeroFund))
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrMintNFT, err.Error())
	}

	poolName := "data_pool_" + strconv.FormatUint(newPool.GetPoolId(), 10)
	symbol := "DATA" + strconv.FormatUint(newPool.GetPoolId(), 10)

	instantiateMsg := types.NewInstantiateNFTMsg(poolName, symbol, newPoolAddr)
	instantiateMsgBz, err := json.Marshal(instantiateMsg)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrInstantiateContract, err.Error())
	}

	codeID := k.GetParams(ctx).DataPoolCodeId

	// instantiate NFT contract for minting data pass (set admin to module)
	poolNFTContractAddr, _, err := k.wasmKeeper.Instantiate(ctx, codeID, moduleAddr, poolAddress, instantiateMsgBz, "data pass", nil)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrInstantiateContract, err.Error())
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

		err := k.cdc.UnmarshalLengthPrefixed(bz, &val)
		if err != nil {
			panic(err)
		}

		poolNumber = val.GetValue()
	}
	return poolNumber
}

func (k Keeper) SetPoolNumber(ctx sdk.Context, poolNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalLengthPrefixed(&gogotypes.UInt64Value{Value: poolNumber})
	store.Set(types.KeyPoolNextNumber, bz)
}

func (k Keeper) SetPool(ctx sdk.Context, pool *types.Pool) {
	store := ctx.KVStore(k.storeKey)
	poolKey := types.GetKeyPrefixPools(pool.GetPoolId())
	bz := k.cdc.MustMarshalLengthPrefixed(pool)
	store.Set(poolKey, bz)
}

func (k Keeper) GetAllPools(ctx sdk.Context) ([]types.Pool, error) {
	// TODO: add pagination
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixPools)
	defer iterator.Close()

	pools := make([]types.Pool, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var pool types.Pool

		err := k.cdc.UnmarshalLengthPrefixed(bz, &pool)
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
	k.cdc.MustUnmarshalLengthPrefixed(bz, pool)

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
func (k Keeper) SellData(ctx sdk.Context, seller sdk.AccAddress, cert types.DataCert) error {
	if cert.UnsignedCert.Requester != seller.String() {
		return types.ErrNotEqualsSeller
	}

	if err := k.verifySignature(ctx, cert); err != nil {
		return err
	}

	if k.isDuplicatedCert(ctx, cert) {
		return types.ErrExistSameDataHash
	}

	poolID := cert.UnsignedCert.PoolId
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return err
	}

	if err := k.validateCertByPool(cert, pool); err != nil {
		return err
	}

	k.SetDataCert(ctx, cert)

	k.increaseCurNumAndUpdatePool(ctx, pool)

	k.addInstantRevenueDistribution(ctx, poolID)

	k.addSalesHistory(ctx, pool.PoolId, pool.Round, seller, cert.UnsignedCert.DataHash)

	return nil
}

// verifySignature verifies that the signature of the oracle is correct
func (k Keeper) verifySignature(ctx sdk.Context, oracleCert types.DataCert) error {
	ora := oracleCert.UnsignedCert.Oracle
	unsignedCert := oracleCert.UnsignedCert
	sign := oracleCert.Signature

	valAddr, err := sdk.AccAddressFromBech32(ora)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalidSignature, err.Error())
	}
	pubKey, err := k.accountKeeper.GetPubKey(ctx, valAddr)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalidSignature, err.Error())
	}
	unsignedCertBinary, err := k.cdc.Marshal(unsignedCert)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalidSignature, err.Error())
	}
	if !pubKey.VerifySignature(unsignedCertBinary, sign) {
		return sdkerrors.Wrap(types.ErrInvalidSignature, "invalid signature")
	}
	return nil
}

// isDuplicatedCert goes through all rounds and checks for data duplication.
func (k Keeper) isDuplicatedCert(ctx sdk.Context, cert types.DataCert) bool {
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

// validateCertByPool verifies the pool and certificate data
func (k Keeper) validateCertByPool(cert types.DataCert, pool *types.Pool) error {
	oracle := cert.UnsignedCert.Oracle
	trustedOracles := pool.PoolParams.TrustedOracles

	if !contains(trustedOracles, oracle) {
		return sdkerrors.Wrap(types.ErrInvalidDataCert, "the oracle is not trusted")
	}

	if pool.Status != types.PENDING {
		return sdkerrors.Wrap(types.ErrInvalidDataCert, "the status of the pool is not 'PENDING'")
	}

	if pool.Round != cert.UnsignedCert.Round {
		return sdkerrors.Wrap(types.ErrInvalidDataCert, fmt.Sprintf("pool round do not matched. pool round: %v", pool.Round))
	}

	return nil
}

func (k Keeper) SetDataCert(ctx sdk.Context, cert types.DataCert) {
	unsignedCert := cert.UnsignedCert
	key := types.GetKeyPrefixDataValidateCert(unsignedCert.PoolId, unsignedCert.Round, unsignedCert.DataHash)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, k.cdc.MustMarshalLengthPrefixed(&cert))
}

func (k Keeper) increaseCurNumAndUpdatePool(ctx sdk.Context, pool *types.Pool) {
	pool.CurNumData += 1

	if pool.CurNumData == pool.PoolParams.TargetNumData {
		pool.Status = types.ACTIVE
	}

	k.SetPool(ctx, pool)
}

func (k Keeper) GetDataCert(ctx sdk.Context, poolID, round uint64, dataHash []byte) (types.DataCert, error) {
	key := types.GetKeyPrefixDataValidateCert(poolID, round, dataHash)
	store := ctx.KVStore(k.storeKey)

	if !store.Has(key) {
		return types.DataCert{}, sdkerrors.Wrap(types.ErrGetDataCert, "certification is not exist")
	}

	bz := store.Get(key)

	var cert types.DataCert
	err := k.cdc.UnmarshalLengthPrefixed(bz, &cert)
	if err != nil {
		return types.DataCert{}, sdkerrors.Wrap(types.ErrGetDataCert, err.Error())
	}

	return cert, nil
}

func (k Keeper) BuyDataPass(ctx sdk.Context, buyer sdk.AccAddress, poolID, round uint64, payment sdk.Coin) error {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrBuyDataPass, err.Error())
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
		return sdkerrors.Wrapf(types.ErrBuyDataPass, err.Error())
	}

	err = k.bankKeeper.SendCoins(ctx, buyer, poolAcc, sdk.NewCoins(payment))
	if err != nil {
		return sdkerrors.Wrapf(types.ErrBuyDataPass, err.Error())
	}

	//mint data pass when pool is activated
	contractAddr := pool.GetNftContractAddr()

	contractAcc, err := sdk.AccAddressFromBech32(contractAddr)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, err.Error())
	}

	mintMsg := types.NewMsgMintDataPass(pool.GetNumIssuedNfts()+1, buyer.String())
	mintMsgBz, err := json.Marshal(mintMsg)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrMintNFT, err.Error())
	}

	_, err = k.wasmKeeper.Execute(ctx, contractAcc, poolAcc, mintMsgBz, sdk.NewCoins(types.ZeroFund))
	if err != nil {
		return sdkerrors.Wrapf(types.ErrMintNFT, err.Error())
	}

	k.increaseNumIssuedNFT(ctx, pool)

	k.addInstantRevenueDistribution(ctx, poolID)

	return nil
}

func (k Keeper) increaseNumIssuedNFT(ctx sdk.Context, pool *types.Pool) {
	pool.NumIssuedNfts += 1

	k.SetPool(ctx, pool)
}

func (k Keeper) RedeemDataPass(ctx sdk.Context, redeemNFT types.MsgRedeemDataPass) (*types.DataPassRedeemReceipt, error) {
	pool, err := k.GetPool(ctx, redeemNFT.PoolId)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrRedeemDataPass, err.Error())
	}

	if pool.GetRound() != redeemNFT.Round {
		return nil, types.ErrRoundNotMatched
	}

	nftContractAcc, err := sdk.AccAddressFromBech32(pool.NftContractAddr)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}

	moduleAddr := types.GetModuleAddress()
	transferNFTMsg := types.NewTransferNFTMsg(moduleAddr.String(), strconv.FormatUint(redeemNFT.DataPassId, 10))

	transferNFTMsgBz, err := json.Marshal(transferNFTMsg)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrRedeemDataPass, err.Error())
	}

	redeemerAcc, err := sdk.AccAddressFromBech32(redeemNFT.Redeemer)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrRedeemDataPass, err.Error())
	}

	redeemTokenIds, err := k.GetRedeemerDataPassByAddr(ctx, pool.PoolId, redeemerAcc)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrRedeemDataPass, err.Error())
	}

	if !contains(redeemTokenIds, strconv.FormatUint(redeemNFT.DataPassId, 10)) {
		return nil, types.ErrNotOwnedRedeemerNft
	}

	zeroFund := sdk.NewCoins(types.ZeroFund)

	_, err = k.wasmKeeper.Execute(ctx, nftContractAcc, redeemerAcc, transferNFTMsgBz, zeroFund)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrRedeemDataPass, err.Error())
	}

	nftRedeemReceipt := types.NewDataPassRedeemReceipt(redeemNFT.PoolId, redeemNFT.Round, redeemNFT.DataPassId, redeemNFT.Redeemer)

	k.SetDataPassRedeemReceipt(ctx, *nftRedeemReceipt)
	err = k.appendDataPassRedeemHistory(ctx, *nftRedeemReceipt)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrRedeemDataPass, err.Error())
	}

	return nftRedeemReceipt, nil
}

func (k Keeper) GetAllDataPassRedeemReceipts(ctx sdk.Context) ([]types.DataPassRedeemReceipt, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixNFTRedeemReceipts)
	defer iterator.Close()

	dataPassRedeemReceipts := make([]types.DataPassRedeemReceipt, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dataPassRedeemReceipt types.DataPassRedeemReceipt

		err := k.cdc.UnmarshalLengthPrefixed(bz, &dataPassRedeemReceipt)
		if err != nil {
			return nil, err
		}

		dataPassRedeemReceipts = append(dataPassRedeemReceipts, dataPassRedeemReceipt)
	}

	return dataPassRedeemReceipts, nil
}

func (k Keeper) GetDataPassRedeemReceipt(ctx sdk.Context, poolID, round, dataPassID uint64) (types.DataPassRedeemReceipt, error) {
	key := types.GetKeyPrefixNFTRedeemReceipt(poolID, round, dataPassID)
	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return types.DataPassRedeemReceipt{}, types.ErrRedeemDataPassNotFound
	}

	bz := store.Get(key)

	var receipt types.DataPassRedeemReceipt
	err := k.cdc.UnmarshalLengthPrefixed(bz, &receipt)
	if err != nil {
		return types.DataPassRedeemReceipt{}, sdkerrors.Wrap(types.ErrGetDataPassRedeemReceipt, err.Error())
	}

	return receipt, nil
}

func (k Keeper) SetDataPassRedeemReceipt(ctx sdk.Context, redeemReceipt types.DataPassRedeemReceipt) {
	store := ctx.KVStore(k.storeKey)
	receiptKey := types.GetKeyPrefixNFTRedeemReceipt(redeemReceipt.PoolId, redeemReceipt.Round, redeemReceipt.DataPassId)
	bz := k.cdc.MustMarshalLengthPrefixed(&redeemReceipt)
	store.Set(receiptKey, bz)
}

func (k Keeper) GetDataPassRedeemHistory(ctx sdk.Context, redeemer string, poolID uint64) (types.DataPassRedeemHistory, error) {
	store := ctx.KVStore(k.storeKey)

	key := types.GetKeyPrefixDataPassRedeemHistoryByPool(redeemer, poolID)

	if !store.Has(key) {
		return types.DataPassRedeemHistory{}, types.ErrRedeemHistoryNotFound
	}

	bz := store.Get(key)

	var history types.DataPassRedeemHistory
	err := k.cdc.UnmarshalLengthPrefixed(bz, &history)
	if err != nil {
		return types.DataPassRedeemHistory{}, err
	}

	return history, nil
}

func (k Keeper) SetDataPassRedeemHistory(ctx sdk.Context, redeemHistory types.DataPassRedeemHistory) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetKeyPrefixDataPassRedeemHistoryByPool(redeemHistory.Redeemer, redeemHistory.PoolId)
	bz := k.cdc.MustMarshalLengthPrefixed(&redeemHistory)
	store.Set(key, bz)
}

func (k Keeper) appendDataPassRedeemHistory(ctx sdk.Context, redeemReceipt types.DataPassRedeemReceipt) error {
	redeemer := redeemReceipt.Redeemer
	poolID := redeemReceipt.PoolId

	redeemHistory, err := k.GetDataPassRedeemHistory(ctx, redeemer, poolID)
	if err != nil && !errors.Is(err, types.ErrRedeemHistoryNotFound) {
		return err
	}

	if errors.Is(err, types.ErrRedeemHistoryNotFound) {
		redeemHistory = types.DataPassRedeemHistory{
			Redeemer:               redeemer,
			PoolId:                 poolID,
			DataPassRedeemReceipts: []types.DataPassRedeemReceipt{redeemReceipt},
		}
	} else {
		redeemHistory.AppendRedeemReceipt(redeemReceipt)
	}

	k.SetDataPassRedeemHistory(ctx, redeemHistory)

	return nil
}

func (k Keeper) GetAllDataPassRedeemHistory(ctx sdk.Context) ([]types.DataPassRedeemHistory, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixDataPassRedeemHistory)
	defer iterator.Close()

	redeemHistory := make([]types.DataPassRedeemHistory, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var history types.DataPassRedeemHistory

		err := k.cdc.UnmarshalLengthPrefixed(bz, &history)
		if err != nil {
			return nil, err
		}

		redeemHistory = append(redeemHistory, history)
	}

	return redeemHistory, nil
}

func (k Keeper) GetRedeemerDataPassByAddr(ctx sdk.Context, poolID uint64, redeemer sdk.AccAddress) ([]string, error) {
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	acc, err := sdk.AccAddressFromBech32(pool.NftContractAddr)
	if err != nil {
		return nil, err
	}

	tokens, err := k.GetRedeemerDataPassWithNFTContractAcc(ctx, acc, redeemer)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (k Keeper) GetRedeemerDataPassWithNFTContractAcc(ctx sdk.Context, nftContractAcc, redeemer sdk.AccAddress) ([]string, error) {
	query := types.NewQueryTokensRequest(redeemer.String())
	queryBz, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	data, err := k.viewKeeper.QuerySmart(ctx, nftContractAcc, queryBz)
	if err != nil {
		return nil, err
	}

	var res types.QueryTokensResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return res.Tokens, nil
}

func contains(slices []string, component string) bool {
	for _, c := range slices {
		if component == c {
			return true
		}
	}
	return false
}
