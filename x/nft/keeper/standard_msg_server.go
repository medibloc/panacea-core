package keeper

import (
	"context"

	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

var _ upstreamnft.MsgServer = standardMsgServer{}

type standardMsgServer struct {
	keeper Keeper
}

// NewStandardMsgServer creates the policy-aware implementation of the
// standard cosmos.nft.v1beta1.Msg service.
func NewStandardMsgServer(k Keeper) upstreamnft.MsgServer {
	return standardMsgServer{keeper: k}
}

// Send validates Panacea lifecycle and transfer policy before delegating
// ownership authentication, state mutation, and EventSend to the SDK keeper.
func (m standardMsgServer) Send(
	goCtx context.Context,
	request *upstreamnft.MsgSend,
) (*upstreamnft.MsgSendResponse, error) {
	if request == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("empty request")
	}
	if err := m.keeper.validateCanonicalClassID(request.ClassId); err != nil {
		return nil, err
	}
	if err := types.ValidateNFTID(request.Id); err != nil {
		return nil, err
	}
	sender, senderAddress, err := m.keeper.canonicalAddress("sender", request.Sender)
	if err != nil {
		return nil, err
	}
	receiver, receiverAddress, err := m.keeper.canonicalNonModuleAccount("receiver", request.Receiver)
	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.keeper.ensureNFTTransferAllowed(ctx, request.ClassId, request.Id); err != nil {
		return nil, err
	}

	canonicalRequest := *request
	canonicalRequest.Sender = sender
	canonicalRequest.Receiver = receiver
	cacheCtx, writeCache := ctx.CacheContext()
	response, err := m.keeper.nftKeeper.Send(sdk.WrapSDKContext(cacheCtx), &canonicalRequest)
	if err != nil {
		return nil, err
	}
	if err := m.keeper.transferOwnerClassCount(
		cacheCtx,
		request.ClassId,
		senderAddress,
		receiverAddress,
	); err != nil {
		return nil, err
	}
	writeCache()
	return response, nil
}
