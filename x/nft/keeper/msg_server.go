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
