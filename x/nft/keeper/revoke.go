package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

func (k Keeper) revokeNFT(
	ctx sdk.Context,
	classID string,
	nftID string,
	controller string,
) error {
	if classID == "" || nftID == "" || controller == "" {
		return sdkerrors.ErrInvalidRequest.Wrap(
			"class_id, nft_id, and controller must not be empty",
		)
	}
	state, record, err := k.loadLiveNFTState(ctx, classID, nftID)
	if err != nil {
		return err
	}
	if state.lifecycle.Revocation != nil {
		return types.ErrNFTRevoked.Wrapf(
			"nft %s in class %s is already revoked",
			nftID,
			classID,
		)
	}

	if record.Policy.Controller != controller {
		return sdkerrors.ErrUnauthorized.Wrapf(
			"account %s does not control class %s",
			controller,
			classID,
		)
	}
	if !record.Policy.Revocable {
		return types.ErrRevocationNotAllowed.Wrapf(
			"class %s does not allow revocation",
			classID,
		)
	}

	revokedAt := ctx.BlockTime()
	if revokedAt.IsZero() {
		return fmt.Errorf("cannot revoke nft %s in class %s at zero block time", nftID, classID)
	}
	if revokedAt.Before(state.lifecycle.Mint.MintedAt) {
		return fmt.Errorf(
			"cannot revoke nft %s in class %s before its mint time",
			nftID,
			classID,
		)
	}
	lifecycle := state.lifecycle
	lifecycle.Revocation = &types.Revocation{
		RevokedAt: revokedAt,
		RevokedBy: controller,
	}

	cacheCtx, writeCache := ctx.CacheContext()
	if err := k.lifecycles.Set(cacheCtx, collections.Join(classID, nftID), lifecycle); err != nil {
		return fmt.Errorf("save nft revocation: %w", err)
	}
	if err := cacheCtx.EventManager().EmitTypedEvent(&types.EventNFTRevoked{
		ClassId:    classID,
		NftId:      nftID,
		Controller: controller,
	}); err != nil {
		return fmt.Errorf("emit nft revoked event: %w", err)
	}
	writeCache()
	return nil
}
