package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/market/types"
)

func (m msgServer) CreateDeal(goCtx context.Context, msg *types.MsgCreateDeal) (*types.MsgCreateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var deal = types.Deal{
		DataSchema:            msg.DataSchema,
		Budget:                msg.Budget,
		MaxNumData:            msg.MaxNumData,
		TrustedDataValidators: msg.GetTrustedDataValidators(),
		Owner:                 msg.Owner,
	}

	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	newDealId, err := m.Keeper.CreateNewDeal(ctx, owner, deal)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreateDealResponse{DealId: newDealId}, nil
}

func (m msgServer) SellData(goCtx context.Context, msg *types.MsgSellData) (*types.MsgSellDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seller, err := sdk.AccAddressFromBech32(msg.Seller)
	if err != nil {
		return nil, err
	}

	validatorAddr, err := sdk.AccAddressFromBech32(msg.Cert.UnsignedCert.DataValidatorAddress)
	if err != nil {
		return nil, err
	}

	unSignedCert := types.UnsignedDataValidationCertificate{
		DealId:               msg.Cert.UnsignedCert.DealId,
		DataHash:             msg.Cert.UnsignedCert.DataHash,
		EncryptedDataUrl:     msg.Cert.UnsignedCert.EncryptedDataUrl,
		DataValidatorAddress: msg.Cert.UnsignedCert.DataValidatorAddress,
		RequesterAddress:     msg.Cert.UnsignedCert.RequesterAddress,
	}

	cert := types.DataValidationCertificate{
		UnsignedCert: &unSignedCert,
		Signature:    msg.Cert.Signature,
	}

	_, err = m.Keeper.Verify(ctx, validatorAddr, cert)
	if err != nil {
		return nil, err
	}

	reward, err := m.Keeper.SellOwnData(ctx, seller, cert)
	if err != nil {
		return nil, err
	}

	return &types.MsgSellDataResponse{Reward: &reward}, nil
}
<<<<<<< HEAD

func (m msgServer) DeactivateDeal(goCtx context.Context, msg *types.MsgDeactivateDeal) (*types.MsgDeactivateDealResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	requester, err := sdk.AccAddressFromBech32(msg.DeactivateRequester)
	if err != nil {
		return nil, err
	}

	deactivateResponse, err := m.Keeper.DeactivateDeal(ctx, msg.DealId, requester)
	if err != nil {
		return nil, err
	}

	return &deactivateResponse, nil
}
=======
>>>>>>> fd49aba48464f8990c70e5142af535a11a8a793f
