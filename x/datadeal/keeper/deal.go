package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) CreateDeal(ctx sdk.Context, owner sdk.AccAddress, deal types.Deal) (uint64, error) {
	dealID, err := k.GetNextDealNumberAndIncrement(ctx)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to get next deal num")
	}

	newDeal := types.NewDeal(dealID, deal)

	var coins sdk.Coins
	coins = append(coins, *deal.GetBudget())

	dealAddress, err := sdk.AccAddressFromBech32(newDeal.GetDealAddress())
	if err != nil {
		return 0, err
	}

	acc := k.accountKeeper.GetAccount(ctx, dealAddress)
	if acc != nil {
		return 0, sdkerrors.Wrapf(types.ErrDealAlreadyExist, "deal %d already exist", dealID)
	}

	k.SetDeal(ctx, newDeal)

	acc = k.accountKeeper.NewAccount(ctx, authtypes.NewModuleAccount(
		authtypes.NewBaseAccountWithAddress(
			dealAddress,
		),
		newDeal.GetDealAddress()),
	)
	k.accountKeeper.SetAccount(ctx, acc)

	err = k.bankKeeper.SendCoins(ctx, owner, dealAddress, coins)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrNotEnoughBalance, "The owner's balance is not enough to make deal")
	}

	return newDeal.GetDealId(), nil
}

func (k Keeper) SetNextDealNumber(ctx sdk.Context, dealNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.UInt64Value{Value: dealNumber})
	store.Set(types.KeyDealNextNumber, bz)
}

func (k Keeper) GetNextDealNumber(ctx sdk.Context) (uint64, error) {
	var dealNumber uint64
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.KeyDealNextNumber) {
		return 0, types.ErrDealNotInitialized
	}

	bz := store.Get(types.KeyDealNextNumber)

	val := gogotypes.UInt64Value{}

	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &val)
	if err != nil {
		return 0, err
	}

	dealNumber = val.GetValue()

	return dealNumber, nil
}

func (k Keeper) GetNextDealNumberAndIncrement(ctx sdk.Context) (uint64, error) {
	dealNumber, err := k.GetNextDealNumber(ctx)
	if err != nil {
		return 0, err
	}

	k.SetNextDealNumber(ctx, dealNumber+1)

	return dealNumber, nil
}

func (k Keeper) GetDeal(ctx sdk.Context, dealID uint64) (types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetKeyPrefixDeals(dealID)
	if !store.Has(dealKey) {
		return types.Deal{}, sdkerrors.Wrapf(types.ErrDealNotFound, "deal with ID %d does not exist", dealID)
	}

	bz := store.Get(dealKey)

	var deal types.Deal
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &deal)
	if err != nil {
		return types.Deal{}, err
	}

	return deal, nil
}

func (k Keeper) SetDeal(ctx sdk.Context, deal types.Deal) {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetKeyPrefixDeals(deal.GetDealId())
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&deal)
	store.Set(dealKey, bz)
}

func (k Keeper) ListDeals(ctx sdk.Context) ([]types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	deals := make([]types.Deal, 0)

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixDeals)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var deal types.Deal

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &deal)
		if err != nil {
			return []types.Deal{}, err
		}

		deals = append(deals, deal)
	}

	return deals, nil
}

func (k Keeper) SellData(ctx sdk.Context, seller sdk.AccAddress, cert types.DataValidationCertificate) (sdk.Coin, error) {
	if k.isDuplicatedData(ctx, cert) {
		return sdk.Coin{}, types.ErrDataAlreadyExist
	}

	deal, err := k.GetDeal(ctx, cert.UnsignedCert.GetDealId())
	if err != nil {
		return sdk.Coin{}, err
	}

	if deal.GetStatus() != types.ACTIVE {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidStatus, "%s", deal.GetStatus())
	}

	dealAddress, err := sdk.AccAddressFromBech32(deal.GetDealAddress())
	if err != nil {
		return sdk.Coin{}, err
	}

	if !k.isTrustedDataValidator(cert, deal) {
		return sdk.Coin{}, types.ErrInvalidDataVal
	}

	//TODO: Fields max_num_data and cur_num_data will be changed in next data datadeal model.
	totalBudget := deal.GetBudget().Amount.ToDec()
	maxNumData := sdk.NewIntFromUint64(deal.GetMaxNumData()).ToDec()
	pricePerData := totalBudget.Quo(maxNumData).TruncateInt()

	reward := sdk.NewCoin(assets.MicroMedDenom, pricePerData)

	dealBalance := k.bankKeeper.GetBalance(ctx, dealAddress, assets.MicroMedDenom)
	if dealBalance.IsLT(reward) {
		return sdk.Coin{}, fmt.Errorf("deal's balance is smaller than reward")
	}

	coins := append(sdk.Coins{}, reward)

	err = k.bankKeeper.SendCoins(ctx, dealAddress, seller, coins)
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrNotEnoughBalance, "The deal's balance is not enough to make deal")
	}

	k.SetDataCertificate(ctx, deal.GetDealId(), cert)
	SetCurNumData(&deal)

	if deal.GetCurNumData() == deal.GetMaxNumData() {
		SetStatusCompleted(&deal)
	}

	k.SetDeal(ctx, deal)

	return reward, nil
}

func (k Keeper) isDuplicatedData(ctx sdk.Context, cert types.DataValidationCertificate) bool {
	store := ctx.KVStore(k.storeKey)
	dataCertKey := types.GetKeyPrefixDataCertificate(cert.UnsignedCert.GetDealId(), cert.UnsignedCert.GetDataHash())

	return store.Has(dataCertKey)
}

func (k Keeper) isTrustedDataValidator(cert types.DataValidationCertificate, deal types.Deal) bool {
	validator := cert.UnsignedCert.GetDataValidatorAddress()
	trustedValidators := deal.GetTrustedDataValidators()

	if len(trustedValidators) == 0 {
		return true
	}

	for _, v := range trustedValidators {
		if validator == v {
			return true
		}
	}

	return false
}

func (k Keeper) GetDataCertificate(ctx sdk.Context, cert types.DataValidationCertificate) (types.DataValidationCertificate, error) {
	store := ctx.KVStore(k.storeKey)
	dataCertificateKey := types.GetKeyPrefixDataCertificate(cert.UnsignedCert.DealId, cert.UnsignedCert.DataHash)
	if !store.Has(dataCertificateKey) {
		return types.DataValidationCertificate{}, sdkerrors.Wrapf(types.ErrDataNotFound, "data with ID %s does not exist", dataCertificateKey)
	}

	bz := store.Get(dataCertificateKey)

	var dataCertificate types.DataValidationCertificate
	err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &dataCertificate)
	if err != nil {
		return types.DataValidationCertificate{}, err
	}

	return dataCertificate, nil
}

func (k Keeper) ListDataCertificates(ctx sdk.Context) ([]types.DataValidationCertificate, error) {
	store := ctx.KVStore(k.storeKey)
	dataCertificates := make([]types.DataValidationCertificate, 0)

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixDataCertificateStore)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var dataCertificate types.DataValidationCertificate

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &dataCertificate)
		if err != nil {
			return []types.DataValidationCertificate{}, err
		}
		dataCertificates = append(dataCertificates, dataCertificate)
	}
	return dataCertificates, nil
}

func (k Keeper) SetDataCertificate(ctx sdk.Context, dealID uint64, cert types.DataValidationCertificate) {
	store := ctx.KVStore(k.storeKey)
	dataHash := cert.UnsignedCert.GetDataHash()
	dataCertificateKey := types.GetKeyPrefixDataCertificate(dealID, dataHash)
	storedDataCertificate := k.cdc.MustMarshalBinaryLengthPrefixed(&cert)
	store.Set(dataCertificateKey, storedDataCertificate)
}

func SetCurNumData(deal *types.Deal) {
	curNumData := deal.GetCurNumData() + 1
	deal.CurNumData = curNumData
}

func SetStatusCompleted(deal *types.Deal) {
	deal.Status = types.COMPLETED
}

func (k Keeper) VerifyDataCertificate(ctx sdk.Context, validatorAddr sdk.AccAddress, cert types.DataValidationCertificate) (bool, error) {
	validatorAcc := k.accountKeeper.GetAccount(ctx, validatorAddr)
	if validatorAcc == nil {
		return false, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address")
	}

	unSignedMarshaled, err := cert.UnsignedCert.Marshal()
	if err != nil {
		return false, sdkerrors.Wrapf(err, "invalid marshaled value")
	}

	validatorPubKey := validatorAcc.GetPubKey()
	if validatorPubKey == nil {
		return false, sdkerrors.Wrapf(err, "validator has no public key")
	}

	isValid := validatorPubKey.VerifySignature(unSignedMarshaled, cert.GetSignature())
	if !isValid {
		return false, sdkerrors.Wrapf(types.ErrInvalidSignature, "%s", cert.GetSignature())
	}

	return true, nil
}

func (k Keeper) DeactivateDeal(ctx sdk.Context, dealID uint64, requester sdk.AccAddress) (uint64, error) {
	deal, err := k.GetDeal(ctx, dealID)
	if err != nil {
		return 0, err
	}

	dealOwner, err := sdk.AccAddressFromBech32(deal.GetOwner())
	if err != nil {
		return 0, err
	}

	if !dealOwner.Equals(requester) {
		return 0, fmt.Errorf("the owner of deal and requester is not equal")
	}

	if deal.GetStatus() != types.ACTIVE {
		return 0, sdkerrors.Wrapf(types.ErrInvalidStatus, "%s", deal.GetStatus())
	}

	dealAddress, err := sdk.AccAddressFromBech32(deal.GetDealAddress())
	if err != nil {
		return 0, err
	}

	remainDealBalance := k.bankKeeper.GetBalance(ctx, dealAddress, assets.MicroMedDenom)

	err = k.bankKeeper.SendCoins(ctx, dealAddress, requester, sdk.Coins{remainDealBalance})
	if err != nil {
		return 0, err
	}

	deal.Status = types.INACTIVE
	k.SetDeal(ctx, deal)

	return dealID, nil
}
