package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gogotypes "github.com/gogo/protobuf/types"
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
		return types.Deal{}, fmt.Errorf("deal with ID %d does not exist", dealId)
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
	// Is deal exists
	findDeal, err := k.GetDeal(ctx, cert.UnsignedCert.GetDealId())
	if err != nil {
		return sdk.Coin{}, err
	}

	// Is deal activated
	if findDeal.GetStatus() != ACTIVE {
		return sdk.Coin{}, fmt.Errorf("deal is not activated")
	}

	dealAddress, err := types.AccDealAddressFromBech32(findDeal.GetDealAddress())
	if err != nil {
		return sdk.Coin{}, err
	}

	// Is trusted validator
	validator := cert.UnsignedCert.GetDataValidatorAddress()
	trustedValidators := findDeal.GetTrustedDataValidators()

	flag := false
	for _, v := range trustedValidators {
		if validator == v {
			flag = true
			break
		}
	}
	if !flag {
		return sdk.Coin{}, fmt.Errorf("data validator is invalid address")
	}

	// Is deal's balance insufficient
	dealBalance := k.bankKeeper.GetBalance(ctx, dealAddress, "umed")
	sellerBalance := k.bankKeeper.GetBalance(ctx, seller, "umed")
	if dealBalance.IsLT(sellerBalance) {
		return sdk.Coin{}, fmt.Errorf("deal's balance is smaller than seller's balance")
	}

	//TODO: Change the field(max_num_data) to price_per_data
	totalAmount := findDeal.GetBudget().Amount.Uint64()
	countOfData := findDeal.GetMaxNumData()

	amountOfReward := totalAmount / countOfData

	pricePerData := sdk.NewCoin("umed", sdk.NewIntFromUint64(amountOfReward))

	var coins sdk.Coins
	coins = append(coins, pricePerData)

	err = k.bankKeeper.SendCoins(ctx, dealAddress, seller, coins)
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrNotEnoughBalance, "The deal's balance is not enough to make deal")
	}

	k.SetCurNumData(findDeal)
	k.SetData(ctx, findDeal.GetDealId(), cert)
	return pricePerData, nil
}

func (k Keeper) SetData(ctx sdk.Context, dealId uint64, cert types.DataValidationCertificate) {
	store := ctx.KVStore(k.storeKey)
	dataHash := cert.UnsignedCert.DataHash
	dealKeyData := types.GetKeyPrefixData(dealId, dataHash)
	storedData := k.cdc.MustMarshalBinaryLengthPrefixed(cert.GetUnsignedCert())
	store.Set(dealKeyData, storedData)
}

func (k Keeper) SetCurNumData(deal types.Deal) {
	curNumData := deal.GetCurNumData() + 1
	deal.CurNumData = curNumData
}
