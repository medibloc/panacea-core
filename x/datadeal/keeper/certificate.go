package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/types/assets"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

func (k Keeper) SubmitConsent(ctx sdk.Context, cert *types.Certificate) error {
	unsignedCert := cert.UnsignedCertificate
	if err := k.oracleKeeper.VerifyOracleSignature(ctx, unsignedCert, cert.Signature); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	if err := k.oracleKeeper.VerifyOracle(ctx, unsignedCert.OracleAddress); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	deal, err := k.GetDeal(ctx, unsignedCert.DealId)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, "failed to get deal. %v", err)
	} else if deal.Status != types.DEAL_STATUS_ACTIVE {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, "deal status is not ACTIVE")
	}

	if err := k.verifyUnsignedCertificate(ctx, unsignedCert); err != nil {
		return sdkerrors.Wrapf(types.ErrSubmitConsent, err.Error())
	}

	if err := k.SetCertificate(ctx, cert); err != nil {
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
		return fmt.Errorf("already provided certificate")
	}
	return nil
}

func (k Keeper) isProvidedCertificate(ctx sdk.Context, dealID uint64, dataHash string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetCertificateKey(dealID, dataHash))
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

func (k Keeper) SetCertificate(ctx sdk.Context, cert *types.Certificate) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCertificateKey(cert.UnsignedCertificate.DealId, cert.UnsignedCertificate.DataHash)

	bz, err := k.cdc.MarshalLengthPrefixed(cert)

	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetCertificate(ctx sdk.Context, dealID uint64, dataHash string) (*types.Certificate, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCertificateKey(dealID, dataHash)

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrCertificateNotFound
	}

	certificate := &types.Certificate{}

	err := k.cdc.UnmarshalLengthPrefixed(bz, certificate)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrGetCertificate, err.Error())
	}

	return certificate, nil
}
