package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) CreateDeal(ctx sdk.Context, buyerAddress sdk.AccAddress, msg *types.MsgCreateDeal) (uint64, error) {

	dealID, err := k.GetNextDealNumberAndIncrement(ctx)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to get next deal num")
	}

	newDeal := types.NewDeal(dealID, msg)

	coins := sdk.NewCoins(*msg.Budget)

	dealAddress, err := sdk.AccAddressFromBech32(newDeal.Address)
	if err != nil {
		return 0, err
	}

	acc := k.accountKeeper.GetAccount(ctx, dealAddress)
	if acc != nil {
		return 0, sdkerrors.Wrapf(types.ErrDealAlreadyExist, "deal %d already exist", dealID)
	}

	acc = k.accountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(
			dealAddress,
		),
		"deal"+strconv.FormatUint(newDeal.Id, 10)),
	)
	k.accountKeeper.SetAccount(ctx, acc)

	if err = k.bankKeeper.SendCoins(ctx, buyerAddress, dealAddress, coins); err != nil {
		return 0, sdkerrors.Wrapf(err, "Failed to send coins to deal account")
	}

	if err = k.SetDeal(ctx, newDeal); err != nil {
		return 0, err
	}

	return newDeal.Id, nil
}

func (k Keeper) SetNextDealNumber(ctx sdk.Context, dealNumber uint64) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalLengthPrefixed(&gogotypes.UInt64Value{Value: dealNumber})
	if err != nil {
		return sdkerrors.Wrapf(err, "Failed to set next deal number")
	}
	store.Set(types.KeyDealNextNumber, bz)
	return nil
}

func (k Keeper) GetNextDealNumber(ctx sdk.Context) (uint64, error) {
	var dealNumber uint64
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.KeyDealNextNumber) {
		return 0, types.ErrDealNotInitialized
	}

	bz := store.Get(types.KeyDealNextNumber)

	val := gogotypes.UInt64Value{}

	if err := k.cdc.UnmarshalLengthPrefixed(bz, &val); err != nil {
		return 0, sdkerrors.Wrapf(err, "Failed to get next deal number")
	}

	dealNumber = val.GetValue()

	return dealNumber, nil
}

func (k Keeper) GetNextDealNumberAndIncrement(ctx sdk.Context) (uint64, error) {
	dealNumber, err := k.GetNextDealNumber(ctx)
	if err != nil {
		return 0, err
	}

	if err = k.SetNextDealNumber(ctx, dealNumber+1); err != nil {
		return 0, err
	}

	return dealNumber, nil
}

func (k Keeper) GetDeal(ctx sdk.Context, dealID uint64) (*types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetDealKey(dealID)

	bz := store.Get(dealKey)
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrDealNotFound, "deal with ID %d does not exist", dealID)
	}

	deal := &types.Deal{}
	if err := k.cdc.UnmarshalLengthPrefixed(bz, deal); err != nil {
		return nil, err
	}
	return deal, nil
}

func (k Keeper) SetDeal(ctx sdk.Context, deal *types.Deal) error {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetDealKey(deal.GetId())
	bz, err := k.cdc.MarshalLengthPrefixed(deal)
	if err != nil {
		return sdkerrors.Wrapf(err, "Failed to set deal")
	}
	store.Set(dealKey, bz)
	return nil
}

func (k Keeper) GetAllDeals(ctx sdk.Context) ([]types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixDeals)
	defer iterator.Close()

	deals := make([]types.Deal, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var deal types.Deal

		if err := k.cdc.UnmarshalLengthPrefixed(bz, &deal); err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetDeal, err.Error())
		}

		deals = append(deals, deal)
	}

	return deals, nil
}

func (k Keeper) SellData(ctx sdk.Context, msg *types.MsgSellData) error {
	_, err := sdk.AccAddressFromBech32(msg.SellerAddress)
	if err != nil {
		return err
	}

	deal, err := k.GetDeal(ctx, msg.DealId)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrSellData, err.Error())
	}

	if deal.Status != types.DEAL_STATUS_ACTIVE {
		return sdkerrors.Wrapf(types.ErrSellData, "deal status is not ACTIVE")
	}

	getDataSale, _ := k.GetDataSale(ctx, msg.VerifiableCid, msg.DealId)
	if getDataSale != nil && getDataSale.Status != types.DATA_SALE_STATUS_VERIFICATION_FAILED {
		return sdkerrors.Wrapf(types.ErrSellData, "data already exists")
	}

	dataSale := types.NewDataSale(msg)
	dataSale.VotingPeriod = k.oracleKeeper.GetVotingPeriod(ctx)

	if err := k.SetDataSale(ctx, dataSale); err != nil {
		return sdkerrors.Wrapf(types.ErrSellData, err.Error())
	}

	k.AddDataVerificationQueue(ctx, dataSale.VerifiableCid, dataSale.DealId, dataSale.VotingPeriod.VotingEndTime)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDataVerificationVote,
			sdk.NewAttribute(types.AttributeKeyVoteStatus, types.AttributeValueVoteStatusStarted),
			sdk.NewAttribute(types.AttributeKeyVerifiableCID, dataSale.VerifiableCid),
		),
	)

	return nil
}

func (k Keeper) GetDataSale(ctx sdk.Context, verifiableCID string, dealID uint64) (*types.DataSale, error) {
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

func (k Keeper) GetAllDataSaleList(ctx sdk.Context) ([]types.DataSale, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DataSaleKey)
	defer iterator.Close()

	dataSales := make([]types.DataSale, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dataSale types.DataSale

		err := k.cdc.UnmarshalLengthPrefixed(bz, &dataSale)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetDataSale, err.Error())
		}

		dataSales = append(dataSales, dataSale)
	}

	return dataSales, nil
}

func (k Keeper) VoteDataVerification(ctx sdk.Context, vote *types.DataVerificationVote, signature []byte) error {
	if err := vote.ValidateBasic(); err != nil {
		return sdkerrors.Wrapf(types.ErrDataVerificationVote, err.Error())
	}

	if !k.verifyDataVerificationVoteSignature(ctx, vote, signature) {
		return sdkerrors.Wrap(oracletypes.ErrDetectionMaliciousBehavior, "")
	}

	if err := k.validateDataVerificationVote(ctx, vote); err != nil {
		return sdkerrors.Wrap(types.ErrDataVerificationVote, err.Error())
	}

	if err := k.SetDataVerificationVote(ctx, vote); err != nil {
		return sdkerrors.Wrap(types.ErrDataVerificationVote, err.Error())
	}

	return nil
}

func (k Keeper) VoteDataDelivery(ctx sdk.Context, vote *types.DataDeliveryVote, signature []byte) error {
	if err := vote.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(types.ErrDataDeliveryVote, err.Error())
	}

	if !k.verifyDataDeliveryVoteSignature(ctx, vote, signature) {
		return sdkerrors.Wrap(oracletypes.ErrDetectionMaliciousBehavior, "")
	}

	// Check if the dataSale vote status
	if err := k.validateDataDeliveryVote(ctx, vote); err != nil {
		return sdkerrors.Wrap(types.ErrDataDeliveryVote, err.Error())
	}

	// When all validations pass, the vote is saved.
	if err := k.SetDataDeliveryVote(ctx, vote); err != nil {
		return sdkerrors.Wrap(types.ErrDataDeliveryVote, err.Error())
	}

	return nil
}

// verifyVoteSignature defines to check for malicious requests.
func (k Keeper) verifyDataVerificationVoteSignature(ctx sdk.Context, vote *types.DataVerificationVote, signature []byte) bool {
	voteBz := k.cdc.MustMarshal(vote)

	oraclePubKeyBz := k.oracleKeeper.GetParams(ctx).MustDecodeOraclePublicKey()
	return secp256k1.PubKey(oraclePubKeyBz).VerifySignature(voteBz, signature)
}

func (k Keeper) verifyDataDeliveryVoteSignature(ctx sdk.Context, vote *types.DataDeliveryVote, signature []byte) bool {
	voteBz := k.cdc.MustMarshal(vote)

	oraclePubKeyBz := k.oracleKeeper.GetParams(ctx).MustDecodeOraclePublicKey()
	return secp256k1.PubKey(oraclePubKeyBz).VerifySignature(voteBz, signature)
}

// validateDataVerificationVote checks the data/verification status in the Panacea to ensure that the data can be voted to be verified.
func (k Keeper) validateDataVerificationVote(ctx sdk.Context, vote *types.DataVerificationVote) error {
	oracle, err := k.oracleKeeper.GetOracle(ctx, vote.VoterAddress)
	if err != nil {
		return err
	}

	if oracle.Status != oracletypes.ORACLE_STATUS_ACTIVE {
		return types.ErrOracleNotActive
	}

	dataSale, err := k.GetDataSale(ctx, vote.VerifiableCid, vote.DealId)
	if err != nil {
		return err
	}

	if dataSale.Status != types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD {
		return fmt.Errorf("the current voted data's status is not 'VERIFICATION_VOTING_PERIOD'")
	}

	return nil
}

func (k Keeper) validateDataDeliveryVote(ctx sdk.Context, vote *types.DataDeliveryVote) error {
	oracle, err := k.oracleKeeper.GetOracle(ctx, vote.VoterAddress)
	if err != nil {
		return err
	}

	if oracle.Status != oracletypes.ORACLE_STATUS_ACTIVE {
		return types.ErrOracleNotActive
	}

	dataSale, err := k.GetDataSale(ctx, vote.VerifiableCid, vote.DealId)
	if err != nil {
		return err
	}

	if dataSale.Status != types.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD {
		return fmt.Errorf("the current voted data's status is not 'DELIVERY_VOTING_PERIOD'")
	}

	return nil
}

func (k Keeper) GetDataVerificationVote(ctx sdk.Context, verifiableCID, voterAddress string, dealID uint64) (*types.DataVerificationVote, error) {
	store := ctx.KVStore(k.storeKey)
	voterAccAddr, err := sdk.AccAddressFromBech32(voterAddress)
	if err != nil {
		return nil, err
	}
	key := types.GetDataVerificationVoteKey(verifiableCID, voterAccAddr, dealID)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("oracle does not exist. verifiableCID: %s, voterAddress: %s, dealID: %d", verifiableCID, voterAddress, dealID)
	}

	vote := &types.DataVerificationVote{}
	err = k.cdc.UnmarshalLengthPrefixed(bz, vote)
	if err != nil {
		return nil, err
	}

	return vote, nil
}

func (k Keeper) SetDataVerificationVote(ctx sdk.Context, vote *types.DataVerificationVote) error {
	store := ctx.KVStore(k.storeKey)

	voterAccAddr, err := sdk.AccAddressFromBech32(vote.VoterAddress)
	if err != nil {
		return err
	}

	key := types.GetDataVerificationVoteKey(vote.VerifiableCid, voterAccAddr, vote.DealId)
	bz, err := k.cdc.MarshalLengthPrefixed(vote)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetDataVerificationVoteIterator(ctx sdk.Context, dealID uint64, verifiableCid string) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, types.GetDataVerificationVotesKey(verifiableCid, dealID))
}

func (k Keeper) GetAllDataVerificationVoteList(ctx sdk.Context) ([]types.DataVerificationVote, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DataVerificationVoteKey)
	defer iterator.Close()

	dataVerificationVotes := make([]types.DataVerificationVote, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dataVerificationVote types.DataVerificationVote
		err := k.cdc.UnmarshalLengthPrefixed(bz, &dataVerificationVote)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrDataVerificationVote, err.Error())
		}

		dataVerificationVotes = append(dataVerificationVotes, dataVerificationVote)
	}

	return dataVerificationVotes, nil
}

func (k Keeper) RemoveDataVerificationVote(ctx sdk.Context, vote *types.DataVerificationVote) error {
	store := ctx.KVStore(k.storeKey)
	voterAccAddr, err := sdk.AccAddressFromBech32(vote.VoterAddress)
	if err != nil {
		return err
	}
	key := types.GetDataVerificationVoteKey(vote.VerifiableCid, voterAccAddr, vote.DealId)

	store.Delete(key)

	return nil
}

func (k Keeper) SetDataDeliveryVote(ctx sdk.Context, vote *types.DataDeliveryVote) error {
	store := ctx.KVStore(k.storeKey)

	voterAccAddr, err := sdk.AccAddressFromBech32(vote.VoterAddress)
	if err != nil {
		return err
	}
	key := types.GetDataDeliveryVoteKey(vote.DealId, vote.VerifiableCid, voterAccAddr)

	bz, err := k.cdc.MarshalLengthPrefixed(vote)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

func (k Keeper) GetDataDeliveryVote(ctx sdk.Context, verifiableCID, voterAddress string, dealID uint64) (*types.DataDeliveryVote, error) {
	store := ctx.KVStore(k.storeKey)
	voterAccAddr, err := sdk.AccAddressFromBech32(voterAddress)
	if err != nil {
		return nil, err
	}
	key := types.GetDataDeliveryVoteKey(dealID, verifiableCID, voterAccAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("DataSale does not exist. dealID: %d, voterAddress: %s, verifiableCID: %s", dealID, voterAddress, verifiableCID)
	}
	vote := &types.DataDeliveryVote{}
	err = k.cdc.UnmarshalLengthPrefixed(bz, vote)

	if err != nil {
		return nil, err
	}

	return vote, nil
}

func (k Keeper) GetDataDeliveryVoteIterator(ctx sdk.Context, dealID uint64, verifiableCid string) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, types.GetDataDeliveryVotesKey(dealID, verifiableCid))
}

func (k Keeper) GetAllDataDeliveryVoteList(ctx sdk.Context) ([]types.DataDeliveryVote, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.DataDeliveryVoteKey)
	defer iterator.Close()

	dataDeliveryVotes := make([]types.DataDeliveryVote, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var dataDeliveryVote types.DataDeliveryVote

		err := k.cdc.UnmarshalLengthPrefixed(bz, &dataDeliveryVote)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetDataSale, err.Error())
		}

		dataDeliveryVotes = append(dataDeliveryVotes, dataDeliveryVote)
	}

	return dataDeliveryVotes, nil
}

func (k Keeper) RemoveDataDeliveryVote(ctx sdk.Context, vote *types.DataDeliveryVote) error {
	store := ctx.KVStore(k.storeKey)
	voterAccAddr, err := sdk.AccAddressFromBech32(vote.VoterAddress)
	if err != nil {
		return err
	}
	key := types.GetDataDeliveryVoteKey(vote.DealId, vote.VerifiableCid, voterAccAddr)

	store.Delete(key)

	return nil
}
