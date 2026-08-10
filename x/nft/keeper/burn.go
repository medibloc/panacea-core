package keeper

import (
	"bytes"
	"fmt"

	"cosmossdk.io/collections"
	upstreamnft "cosmossdk.io/x/nft"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

func (k Keeper) burnNFT(
	ctx sdk.Context,
	classID string,
	nftID string,
	owner string,
	ownerAddress sdk.AccAddress,
) error {
	if classID == "" || nftID == "" || owner == "" || len(ownerAddress) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap(
			"class_id, nft_id, and owner must not be empty",
		)
	}

	state, classRecord, err := k.loadLiveNFTState(ctx, classID, nftID)
	if err != nil {
		return err
	}
	liveSupply := k.nftKeeper.GetTotalSupply(ctx, classID)
	if liveSupply == 0 || classRecord.MintedCount == 0 || liveSupply > classRecord.MintedCount {
		return fmt.Errorf(
			"nft %s in class %s has inconsistent supply %d and minted count %d",
			nftID,
			classID,
			liveSupply,
			classRecord.MintedCount,
		)
	}

	storedOwner := k.nftKeeper.GetOwner(ctx, classID, nftID)
	if len(storedOwner) == 0 {
		return fmt.Errorf("nft %s in class %s has no owner", nftID, classID)
	}
	storedOwnerString, err := k.addressCodec.BytesToString(storedOwner)
	if err != nil {
		return fmt.Errorf("encode owner for nft %s: %w", nftID, err)
	}
	canonicalStoredOwner, canonicalStoredAddress, err := k.canonicalNonModuleAccount(
		"stored owner",
		storedOwnerString,
	)
	if err != nil {
		return fmt.Errorf("nft %s has invalid stored owner: %w", nftID, err)
	}
	if canonicalStoredOwner != storedOwnerString {
		return fmt.Errorf("nft %s in class %s has non-canonical stored owner", nftID, classID)
	}
	if !bytes.Equal(canonicalStoredAddress, ownerAddress) {
		return sdkerrors.ErrUnauthorized.Wrapf(
			"account %s does not own nft %s in class %s",
			owner,
			nftID,
			classID,
		)
	}

	if err := types.ValidateURI(state.nft.Uri, state.nft.UriHash); err != nil {
		return fmt.Errorf("nft %s has invalid stored URI metadata: %w", nftID, err)
	}
	canonicalData, err := types.CanonicalizeNFTData(k.cdc, state.nft.Data)
	if err != nil {
		return fmt.Errorf("nft %s has invalid stored data: %w", nftID, err)
	}

	burnedAt := ctx.BlockTime()
	if burnedAt.IsZero() {
		return fmt.Errorf("cannot burn nft %s in class %s at zero block time", nftID, classID)
	}
	if burnedAt.Before(state.lifecycle.Mint.MintedAt) {
		return fmt.Errorf(
			"cannot burn nft %s in class %s before its mint time",
			nftID,
			classID,
		)
	}
	if state.lifecycle.Revocation != nil && burnedAt.Before(state.lifecycle.Revocation.RevokedAt) {
		return fmt.Errorf(
			"cannot burn nft %s in class %s before its revocation time",
			nftID,
			classID,
		)
	}

	tombstone := types.BurnTombstone{
		ClassId:    classID,
		NftId:      nftID,
		Mint:       state.lifecycle.Mint,
		Uri:        state.nft.Uri,
		UriHash:    state.nft.UriHash,
		Data:       canonicalData,
		Revocation: state.lifecycle.Revocation,
		BurnedAt:   burnedAt,
		BurnedBy:   owner,
	}
	key := collections.Join(classID, nftID)
	cacheCtx, writeCache := ctx.CacheContext()
	if err := k.nftKeeper.Burn(cacheCtx, classID, nftID); err != nil {
		return fmt.Errorf("burn standard nft: %w", err)
	}
	if err := validateBurnEvent(
		cacheCtx.EventManager().ABCIEvents(),
		classID,
		nftID,
		owner,
	); err != nil {
		return err
	}
	if err := k.lifecycles.Remove(cacheCtx, key); err != nil {
		return fmt.Errorf("remove nft lifecycle: %w", err)
	}
	if err := k.tombstones.Set(cacheCtx, key, tombstone); err != nil {
		return fmt.Errorf("save nft tombstone: %w", err)
	}
	if err := k.decrementOwnerClassCount(cacheCtx, classID, ownerAddress); err != nil {
		return err
	}
	writeCache()
	return nil
}

func validateBurnEvent(events []abci.Event, classID, nftID, owner string) error {
	burnEventType := proto.MessageName(&upstreamnft.EventBurn{})
	burnEventCount := 0
	for _, event := range events {
		if event.Type != burnEventType {
			continue
		}
		burnEventCount++
		parsedEvent, err := sdk.ParseTypedEvent(event)
		if err != nil {
			return fmt.Errorf("decode standard nft burn event: %w", err)
		}
		burnEvent, ok := parsedEvent.(*upstreamnft.EventBurn)
		if !ok || burnEvent.ClassId != classID || burnEvent.Id != nftID || burnEvent.Owner != owner {
			return fmt.Errorf("burn standard nft emitted an unexpected event")
		}
	}
	if burnEventCount != 1 {
		return fmt.Errorf(
			"burn standard nft emitted %d EventBurn events instead of one",
			burnEventCount,
		)
	}
	return nil
}
