package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) SubmitConsent(ctx sdk.Context, consent *types.Consent) error {
	unsignedCert := consent.Certificate.UnsignedCertificate
	if err := k.oracleKeeper.VerifyOracleSignature(ctx, unsignedCert, consent.Certificate.Signature); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	if err := k.oracleKeeper.VerifyOracle(ctx, unsignedCert.OracleAddress); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	deal, err := k.GetDeal(ctx, consent.DealId)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, "failed to get deal. %v", err)
	} else if deal.Status != types.DEAL_STATUS_ACTIVE {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, "deal status is not ACTIVE")
	}

	if err := k.verifyUnsignedCertificate(ctx, unsignedCert); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	if err := k.ValidateAgreements(deal.AgreementTerms, consent.Agreements); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	if err := k.SetConsent(ctx, consent); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	if err := k.sendReward(ctx, deal, unsignedCert); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	if err := k.postProcessingOfDeal(ctx, deal); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	return nil
}

func (k Keeper) verifyUnsignedCertificate(ctx sdk.Context, unsignedCert *types.UnsignedCertificate) error {
	activeUniqueID := k.oracleKeeper.GetParams(ctx).UniqueId
	if activeUniqueID != unsignedCert.UniqueId {
		return fmt.Errorf("does not match active uniqueID. certificateUniqueID(%s) activeUniqueID(%s)", unsignedCert.UniqueId, activeUniqueID)
	}

	if k.isProvidedCertificate(ctx, unsignedCert.DealId, unsignedCert.DataHash) {
		return fmt.Errorf("already provided consent")
	}
	return nil
}

func (k Keeper) ValidateAgreements(terms []*types.AgreementTerm, agreements []*types.Agreement) error {
	if len(terms) != len(agreements) {
		return fmt.Errorf("invalid count(%v) of agreements: expected:%v", len(agreements), len(terms))
	}

	for _, term := range terms {
		agreement := findAgreement(agreements, term.Id)
		if agreement == nil {
			return fmt.Errorf("cannot find agreement for term %v", term.Id)
		}
		if term.Required && !agreement.Agreement {
			return fmt.Errorf("disagreed to the required agreement term %v", term.Id)
		}
	}

	return nil
}

func findAgreement(agreements []*types.Agreement, termID uint32) *types.Agreement {
	for _, agreement := range agreements {
		if agreement.TermId == termID {
			return agreement
		}
	}
	return nil
}

func (k Keeper) isProvidedCertificate(ctx sdk.Context, dealID uint64, dataHash string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetConsentKey(dealID, dataHash))
}

func (k Keeper) sendReward(ctx sdk.Context, deal *types.Deal, unsignedCert *types.UnsignedCertificate) error {
	dealAccAddr, err := sdk.AccAddressFromBech32(deal.GetAddress())
	if err != nil {
		return err
	}

	providerAccAddr, err := sdk.AccAddressFromBech32(unsignedCert.ProviderAddress)
	if err != nil {
		return err
	}

	pricePerData := deal.GetPricePerData()

	dealBalance := k.bankKeeper.GetBalance(ctx, dealAccAddr, assets.MicroMedDenom)
	if dealBalance.IsLT(sdk.NewCoin(assets.MicroMedDenom, pricePerData.TruncateInt())) {
		return fmt.Errorf("not enough balance in deal")
	}

	oracle, err := k.oracleKeeper.GetOracle(ctx, unsignedCert.OracleAddress)
	if err != nil {
		return fmt.Errorf("failed to get oracle. %w", err)
	}
	oracleCommissionRate := oracle.OracleCommissionRate

	oracleReward := sdk.NewCoin(assets.MicroMedDenom, pricePerData.Mul(oracleCommissionRate).TruncateInt())
	providerReward := sdk.NewCoin(assets.MicroMedDenom, pricePerData.Mul(sdk.OneDec().Sub(oracleCommissionRate)).TruncateInt())
	if err := k.bankKeeper.SendCoins(ctx, dealAccAddr, providerAccAddr, sdk.NewCoins(providerReward)); err != nil {
		return fmt.Errorf("failed to send reward to provider. %w", err)
	}

	// We already do oracle address verification above.
	oracleAccAddr, _ := sdk.AccAddressFromBech32(unsignedCert.OracleAddress)
	if err := k.bankKeeper.SendCoins(ctx, dealAccAddr, oracleAccAddr, sdk.NewCoins(oracleReward)); err != nil {
		return fmt.Errorf("failed to send reward to oracle. %w", err)
	}
	return nil
}

func (k Keeper) postProcessingOfDeal(ctx sdk.Context, deal *types.Deal) error {
	deal.IncreaseCurNumData()

	if deal.CurNumData == deal.MaxNumData {
		deal.Status = types.DEAL_STATUS_COMPLETED
	}

	if err := k.SetDeal(ctx, deal); err != nil {
		return fmt.Errorf("failed to set deal. %w", err)
	}
	return nil
}

func (k Keeper) SetConsent(ctx sdk.Context, consent *types.Consent) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetConsentKey(consent.Certificate.UnsignedCertificate.DealId, consent.Certificate.UnsignedCertificate.DataHash)
	bz, err := k.cdc.MarshalLengthPrefixed(consent)

	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetConsent(ctx sdk.Context, dealID uint64, dataHash string) (*types.Consent, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetConsentKey(dealID, dataHash)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrConsentNotFound
	}

	consent := &types.Consent{}

	err := k.cdc.UnmarshalLengthPrefixed(bz, consent)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrGetConsent, err.Error())
	}

	return consent, nil
}

func (k Keeper) GetAllConsents(ctx sdk.Context) ([]types.Consent, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ConsentKey)
	defer iterator.Close()

	consents := make([]types.Consent, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var consent types.Consent

		if err := k.cdc.UnmarshalLengthPrefixed(bz, &consent); err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetConsent, err.Error())
		}

		consents = append(consents, consent)
	}

	return consents, nil
}
