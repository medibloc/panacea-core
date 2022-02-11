package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/market/types"
)

const (
	ACTIVE    = "ACTIVE"    // When deal is activated.
	INACTIVE  = "INACTIVE"  // When deal is deactivated.
	COMPLETED = "COMPLETED" // When deal is completed.
)

func (k Keeper) CreateNewDeal(ctx sdk.Context, owner sdk.AccAddress, deal types.Deal) (uint64, error) {
	dealId := k.GetNextDealNumberAndIncrement(ctx)

	newDeal := newDeal(dealId, deal)

	var coins sdk.Coins
	coins = append(coins, *deal.GetBudget())

	dealAddress, err := types.AccDealAddressFromBech32(newDeal.GetDealAddress())
	if err != nil {
		return 0, err
	}

	acc := k.accountKeeper.GetAccount(ctx, dealAddress)
	if acc != nil {
		return 0, sdkerrors.Wrapf(types.ErrDealAlreadyExist, "deal %d already exist", dealId)
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

func newDeal(dealId uint64, deal types.Deal) types.Deal {

	dealAddress := types.NewDealAddress(dealId)

	newDeal := &types.Deal{
		DealId:                dealId,
		DealAddress:           dealAddress.String(),
		DataSchema:            deal.GetDataSchema(),
		Budget:                deal.GetBudget(),
		TrustedDataValidators: deal.GetTrustedDataValidators(),
		MaxNumData:            deal.GetMaxNumData(),
		CurNumData:            0,
		Owner:                 deal.GetOwner(),
		Status:                ACTIVE,
	}

	return *newDeal
}

func (k Keeper) SetNextDealNumber(ctx sdk.Context, dealNumber uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&gogotypes.UInt64Value{Value: dealNumber})
	store.Set(types.KeyDealNextNumber, bz)
}

func (k Keeper) GetNextDealNumber(ctx sdk.Context) uint64 {
	var dealNumber uint64
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyDealNextNumber)
	if bz == nil {
		panic(fmt.Errorf("deal has not been initialized -- Should have been done in InitGenesis"))
	} else {
		val := gogotypes.UInt64Value{}

		err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &val)
		if err != nil {
			panic(err)
		}

		dealNumber = val.GetValue()
	}
	return dealNumber
}

func (k Keeper) GetNextDealNumberAndIncrement(ctx sdk.Context) uint64 {
	dealNumber := k.GetNextDealNumber(ctx)
	k.SetNextDealNumber(ctx, dealNumber+1)
	return dealNumber
}

func (k Keeper) GetDeal(ctx sdk.Context, dealId uint64) (types.Deal, error) {
	store := ctx.KVStore(k.storeKey)
	dealKey := types.GetKeyPrefixDeals(dealId)
	if !store.Has(dealKey) {
		return types.Deal{}, sdkerrors.Wrapf(types.ErrDealNotFound, "deal with ID %d does not exist", dealId)
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

func (k Keeper) SellOwnData(ctx sdk.Context, seller sdk.AccAddress, cert types.DataValidationCertificate) (sdk.Coin, error) {
	err := k.isDataCertDuplicate(ctx, cert)
	if err != nil {
		return sdk.Coin{}, err
	}

	findDeal, err := k.GetDeal(ctx, cert.UnsignedCert.GetDealId())
	if err != nil {
		return sdk.Coin{}, err
	}

	if findDeal.GetStatus() != ACTIVE {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidStatus, "%s", findDeal.GetStatus())
	}

	dealAddress, err := types.AccDealAddressFromBech32(findDeal.GetDealAddress())
	if err != nil {
		return sdk.Coin{}, err
	}

	err = k.isTrustedValidator(cert, findDeal)
	if err != nil {
		return sdk.Coin{}, err
	}

	//TODO: Fields max_num_data and cur_num_data will be changed in next data market model.
	totalAmount := findDeal.GetBudget().Amount.Uint64()
	countOfData := findDeal.GetMaxNumData()

	pricePerData := sdk.NewCoin(assets.MicroMedDenom, sdk.NewIntFromUint64(totalAmount/countOfData))

	dealBalance := k.bankKeeper.GetBalance(ctx, dealAddress, assets.MicroMedDenom)
	if dealBalance.IsLT(pricePerData) {
		return sdk.Coin{}, fmt.Errorf("deal's balance is smaller than reward")
	}

	coins := append(sdk.Coins{}, pricePerData)

	err = k.bankKeeper.SendCoins(ctx, dealAddress, seller, coins)
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrNotEnoughBalance, "The deal's balance is not enough to make deal")
	}

	k.SetDataCertificate(ctx, findDeal.GetDealId(), cert)
	SetCurNumData(&findDeal)
	k.SetDeal(ctx, findDeal)
	return pricePerData, nil

}

func (k Keeper) isDataCertDuplicate(ctx sdk.Context, cert types.DataValidationCertificate) error {
	store := ctx.KVStore(k.storeKey)
	dataCertKey := types.GetKeyPrefixDataCertificate(cert.UnsignedCert.GetDealId(), cert.UnsignedCert.GetDataHash())

	if store.Has(dataCertKey) {
		return sdkerrors.Wrapf(types.ErrDataAlreadyExist, "data %s is already exist.", dataCertKey)
	}

	return nil
}

func (k Keeper) isTrustedValidator(cert types.DataValidationCertificate, findDeal types.Deal) error {
	validator := cert.UnsignedCert.GetDataValidatorAddress()
	trustedValidators := findDeal.GetTrustedDataValidators()

	for _, v := range trustedValidators {
		if validator == v {
			return nil
		}
	}
	return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid validator address")
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

func (k Keeper) SetDataCertificate(ctx sdk.Context, dealId uint64, cert types.DataValidationCertificate) {
	store := ctx.KVStore(k.storeKey)
	dataHash := cert.UnsignedCert.GetDataHash()
	dataCertificateKey := types.GetKeyPrefixDataCertificate(dealId, dataHash)
	storedDataCertificate := k.cdc.MustMarshalBinaryLengthPrefixed(&cert)
	store.Set(dataCertificateKey, storedDataCertificate)
}

func SetCurNumData(deal *types.Deal) {
	curNumData := deal.GetCurNumData() + 1
	deal.CurNumData = curNumData
}

func (k Keeper) VerifyDataCertificate(ctx sdk.Context, validatorAddr sdk.AccAddress, cert types.DataValidationCertificate) (bool, error) {
	validatorAcc := k.accountKeeper.GetAccount(ctx, validatorAddr)
	if validatorAcc == nil {
		return false, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid validator address")
	}

	if validatorAcc.GetPubKey() == nil {
		return false, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "the publicKey does not exist in the validator account")
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

	return isValid, nil
}

func (k Keeper) DeactivateDeal(ctx sdk.Context, dealId uint64, requester sdk.AccAddress) (uint64, error) {
	findDeal, err := k.GetDeal(ctx, dealId)
	if err != nil {
		return 0, err
	}

	dealOwner, err := sdk.AccAddressFromBech32(findDeal.GetOwner())
	if err != nil {
		return 0, err
	}

	if !dealOwner.Equals(requester) {
		return 0, fmt.Errorf("the owner of deal and requester is not equal")
	}

	if findDeal.GetStatus() != ACTIVE {
		return 0, sdkerrors.Wrapf(types.ErrInvalidStatus, "%s", findDeal.GetStatus())
	}

	findDealAddress, err := types.AccDealAddressFromBech32(findDeal.GetDealAddress())
	if err != nil {
		return 0, err
	}

	remainDealBalance := k.bankKeeper.GetBalance(ctx, findDealAddress, assets.MicroMedDenom)

	err = k.bankKeeper.SendCoins(ctx, findDealAddress, requester, sdk.Coins{remainDealBalance})
	if err != nil {
		return 0, err
	}

	findDeal.Status = INACTIVE
	k.SetDeal(ctx, findDeal)

	return dealId, nil
}
