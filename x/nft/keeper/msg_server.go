package keeper

import (
	"context"

	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

type msgServer struct {
	types.UnimplementedMsgServer
	keeper Keeper
}

// NewMsgServer creates Panacea's policy-aware NFT message server.
func NewMsgServer(k Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

// CreateClass creates the immutable standard class and Panacea policy state.
func (m msgServer) CreateClass(
	goCtx context.Context,
	request *types.MsgCreateClassRequest,
) (*types.MsgCreateClassResponse, error) {
	if request == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}
	if err := types.ValidateLocalClassID(request.LocalClassId); err != nil {
		return nil, err
	}
	if err := types.ValidateClassMetadata(
		request.Name,
		request.Symbol,
		request.Description,
		request.Uri,
		request.UriHash,
	); err != nil {
		return nil, err
	}
	if err := types.ValidateTransferPolicy(request.TransferPolicy); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	creator, _, err := m.keeper.canonicalNonModuleAccount(ctx, "creator", request.Creator)
	if err != nil {
		return nil, err
	}
	classID := creator + ":" + request.LocalClassId
	if _, _, err := types.ParseClassID(classID); err != nil {
		return nil, err
	}

	class := upstreamnft.Class{
		Id:          classID,
		Name:        request.Name,
		Symbol:      request.Symbol,
		Description: request.Description,
		Uri:         request.Uri,
		UriHash:     request.UriHash,
	}
	policy := types.ClassPolicy{
		ClassId:        classID,
		Creator:        creator,
		Controller:     creator,
		TransferPolicy: request.TransferPolicy,
		Revocable:      request.Revocable,
		MaxSupply:      request.MaxSupply,
	}
	if err := m.keeper.createClass(ctx, class, policy); err != nil {
		return nil, err
	}

	return &types.MsgCreateClassResponse{ClassId: classID}, nil
}

// UpdateController transfers class operations to a new non-module account.
func (m msgServer) UpdateController(
	goCtx context.Context,
	request *types.MsgUpdateControllerRequest,
) (*types.MsgUpdateControllerResponse, error) {
	if request == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}
	if err := m.keeper.validateCanonicalClassID(request.ClassId); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	controller, _, err := m.keeper.canonicalAddress("controller", request.Controller)
	if err != nil {
		return nil, err
	}
	newController, _, err := m.keeper.canonicalNonModuleAccount(
		ctx,
		"new controller",
		request.NewController,
	)
	if err != nil {
		return nil, err
	}
	if err := m.keeper.updateController(
		ctx,
		request.ClassId,
		controller,
		newController,
	); err != nil {
		return nil, err
	}

	return &types.MsgUpdateControllerResponse{}, nil
}

// Mint creates one standard NFT and its ACTIVE Panacea lifecycle atomically.
func (m msgServer) Mint(
	goCtx context.Context,
	request *types.MsgMintRequest,
) (*types.MsgMintResponse, error) {
	if request == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}
	if err := m.keeper.validateCanonicalClassID(request.ClassId); err != nil {
		return nil, err
	}
	if err := types.ValidateNFTID(request.NftId); err != nil {
		return nil, err
	}
	if err := types.ValidateURI(request.Uri, request.UriHash); err != nil {
		return nil, err
	}
	data, err := types.CanonicalizeNFTData(m.keeper.cdc, request.Data)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	controller, _, err := m.keeper.canonicalAddress("controller", request.Controller)
	if err != nil {
		return nil, err
	}
	_, recipient, err := m.keeper.canonicalNonModuleAccount(ctx, "recipient", request.Recipient)
	if err != nil {
		return nil, err
	}
	token := upstreamnft.NFT{
		ClassId: request.ClassId,
		Id:      request.NftId,
		Uri:     request.Uri,
		UriHash: request.UriHash,
		Data:    data,
	}
	if err := m.keeper.mintNFT(ctx, token, controller, recipient); err != nil {
		return nil, err
	}

	return &types.MsgMintResponse{}, nil
}
