package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

func (k Keeper) createClass(
	ctx sdk.Context,
	class upstreamnft.Class,
	policy types.ClassPolicy,
) error {
	if class.Id == "" || class.Id != policy.ClassId || policy.Creator != policy.Controller {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid class and policy relationship")
	}
	if class.Data != nil {
		return sdkerrors.ErrInvalidRequest.Wrap("class data must be nil")
	}

	policyExists, err := k.classPolicies.Has(ctx, class.Id)
	if err != nil {
		return fmt.Errorf("check class policy: %w", err)
	}
	mintedCountExists, err := k.mintedCounts.Has(ctx, class.Id)
	if err != nil {
		return fmt.Errorf("check minted count: %w", err)
	}
	if k.nftKeeper.HasClass(ctx, class.Id) || policyExists || mintedCountExists {
		return upstreamnft.ErrClassExists.Wrapf("class %s already exists", class.Id)
	}

	cacheCtx, writeCache := ctx.CacheContext()
	if err := k.nftKeeper.SaveClass(cacheCtx, class); err != nil {
		return errorsmod.Wrap(err, "save standard nft class")
	}
	if err := k.classPolicies.Set(cacheCtx, class.Id, policy); err != nil {
		return fmt.Errorf("save class policy: %w", err)
	}
	if err := k.mintedCounts.Set(cacheCtx, class.Id, 0); err != nil {
		return fmt.Errorf("initialize minted count: %w", err)
	}
	if err := cacheCtx.EventManager().EmitTypedEvent(&types.EventClassCreated{
		ClassId: class.Id,
		Creator: policy.Creator,
	}); err != nil {
		return fmt.Errorf("emit class created event: %w", err)
	}
	writeCache()
	return nil
}

func (k Keeper) getClassRecord(ctx sdk.Context, classID string) (*types.ClassRecord, error) {
	class, hasClass := k.nftKeeper.GetClass(ctx, classID)
	policy, policyErr := k.classPolicies.Get(ctx, classID)
	mintedCount, mintedCountErr := k.mintedCounts.Get(ctx, classID)

	policyMissing := errors.Is(policyErr, collections.ErrNotFound)
	mintedCountMissing := errors.Is(mintedCountErr, collections.ErrNotFound)
	if !hasClass && policyMissing && mintedCountMissing {
		return nil, upstreamnft.ErrClassNotExists.Wrapf("class %s not found", classID)
	}
	if !hasClass || policyMissing || mintedCountMissing {
		return nil, fmt.Errorf("class %s has inconsistent standard and policy state", classID)
	}
	if policyErr != nil {
		return nil, fmt.Errorf("load class policy: %w", policyErr)
	}
	if mintedCountErr != nil {
		return nil, fmt.Errorf("load minted count: %w", mintedCountErr)
	}
	if policy.ClassId != classID {
		return nil, fmt.Errorf("class policy key does not match value for %s", classID)
	}

	return &types.ClassRecord{
		Class:       &class,
		Policy:      &policy,
		MintedCount: mintedCount,
	}, nil
}
