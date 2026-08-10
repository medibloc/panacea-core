package keeper

import (
	"errors"
	"fmt"
	"math"

	"cosmossdk.io/collections"
	upstreamnft "cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

type storedNFTState struct {
	nft          upstreamnft.NFT
	lifecycle    types.LifecycleRecord
	tombstone    types.BurnTombstone
	hasNFT       bool
	hasLifecycle bool
	hasTombstone bool
}

func (k Keeper) mintNFT(
	ctx sdk.Context,
	token upstreamnft.NFT,
	controller string,
	recipient sdk.AccAddress,
) error {
	if token.ClassId == "" || token.Id == "" || controller == "" || len(recipient) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("invalid mint relationship")
	}
	record, err := k.getClassRecord(ctx, token.ClassId)
	if err != nil {
		return err
	}
	if record.Policy.Controller != controller {
		return sdkerrors.ErrUnauthorized.Wrapf(
			"account %s does not control class %s",
			controller,
			token.ClassId,
		)
	}
	state, err := k.loadNFTState(ctx, token.ClassId, token.Id)
	if err != nil {
		return err
	}
	if err := state.ensureAvailableForMint(token.ClassId, token.Id); err != nil {
		return err
	}
	liveSupply := k.nftKeeper.GetTotalSupply(ctx, token.ClassId)
	if liveSupply > record.MintedCount {
		return fmt.Errorf(
			"class %s supply %d exceeds minted count %d",
			token.ClassId,
			liveSupply,
			record.MintedCount,
		)
	}
	nextCount, err := nextMintedCount(record.MintedCount, record.Policy.MaxSupply)
	if err != nil {
		return err
	}
	mintedAt := ctx.BlockTime()
	if mintedAt.IsZero() {
		return fmt.Errorf(
			"cannot mint nft %s in class %s at zero block time",
			token.Id,
			token.ClassId,
		)
	}

	lifecycle := types.LifecycleRecord{
		ClassId: token.ClassId,
		NftId:   token.Id,
		Mint: types.MintRecord{
			MintedAt: mintedAt,
			MintedBy: controller,
		},
	}
	key := collections.Join(token.ClassId, token.Id)
	cacheCtx, writeCache := ctx.CacheContext()
	if err := k.nftKeeper.Mint(cacheCtx, token, recipient); err != nil {
		return fmt.Errorf("mint standard nft: %w", err)
	}
	if err := k.lifecycles.Set(cacheCtx, key, lifecycle); err != nil {
		return fmt.Errorf("save nft lifecycle: %w", err)
	}
	if err := k.mintedCounts.Set(cacheCtx, token.ClassId, nextCount); err != nil {
		return fmt.Errorf("update minted count: %w", err)
	}
	if err := k.incrementOwnerClassCount(cacheCtx, token.ClassId, recipient); err != nil {
		return err
	}
	writeCache()
	return nil
}

func (k Keeper) getLiveNFTRecord(
	ctx sdk.Context,
	classID string,
	nftID string,
) (*types.LiveNFTRecord, error) {
	state, classRecord, err := k.loadLiveNFTState(ctx, classID, nftID)
	if err != nil {
		return nil, err
	}
	liveSupply := k.nftKeeper.GetTotalSupply(ctx, classID)
	return k.buildLiveNFTRecord(ctx, classID, nftID, state, classRecord, liveSupply)
}

func (k Keeper) buildLiveNFTRecord(
	ctx sdk.Context,
	classID string,
	nftID string,
	state storedNFTState,
	classRecord *types.ClassRecord,
	liveSupply uint64,
) (*types.LiveNFTRecord, error) {
	if liveSupply == 0 || classRecord.MintedCount == 0 || liveSupply > classRecord.MintedCount {
		return nil, fmt.Errorf(
			"nft %s in class %s has inconsistent supply %d and minted count %d",
			nftID,
			classID,
			liveSupply,
			classRecord.MintedCount,
		)
	}

	if err := types.ValidateURI(state.nft.Uri, state.nft.UriHash); err != nil {
		return nil, fmt.Errorf("nft %s has invalid stored URI metadata: %w", nftID, err)
	}
	canonicalData, err := types.CanonicalizeNFTData(k.cdc, state.nft.Data)
	if err != nil {
		return nil, fmt.Errorf("nft %s has invalid stored data: %w", nftID, err)
	}
	state.nft.Data = canonicalData

	ownerAddress := k.nftKeeper.GetOwner(ctx, classID, nftID)
	if len(ownerAddress) == 0 {
		return nil, fmt.Errorf("nft %s in class %s has no owner", nftID, classID)
	}
	owner, err := k.addressCodec.BytesToString(ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("encode owner for nft %s: %w", nftID, err)
	}
	canonicalOwner, _, err := k.canonicalNonModuleAccount("stored owner", owner)
	if err != nil {
		return nil, fmt.Errorf("nft %s has invalid stored owner: %w", nftID, err)
	}
	status := types.LiveNFTStatus_LIVE_NFT_STATUS_ACTIVE
	if state.lifecycle.Revocation != nil {
		status = types.LiveNFTStatus_LIVE_NFT_STATUS_REVOKED
	}

	return &types.LiveNFTRecord{
		Nft:        &state.nft,
		Owner:      canonicalOwner,
		Status:     status,
		Mint:       state.lifecycle.Mint,
		Revocation: state.lifecycle.Revocation,
	}, nil
}

func (k Keeper) getNFTRecord(
	ctx sdk.Context,
	classID string,
	nftID string,
) (*types.NFTRecord, error) {
	state, err := k.loadNFTState(ctx, classID, nftID)
	if err != nil {
		return nil, err
	}
	if err := state.validateLiveCombination(classID, nftID); err != nil {
		return nil, err
	}
	if state.hasNFT {
		live, err := k.getLiveNFTRecord(ctx, classID, nftID)
		if err != nil {
			return nil, err
		}
		return &types.NFTRecord{
			Record: &types.NFTRecord_Live{Live: live},
		}, nil
	}
	if state.hasTombstone {
		tombstone, err := k.getBurnTombstone(ctx, classID, nftID, state.tombstone)
		if err != nil {
			return nil, err
		}
		return &types.NFTRecord{
			Record: &types.NFTRecord_BurnTombstone{BurnTombstone: tombstone},
		}, nil
	}
	return nil, upstreamnft.ErrNFTNotExists.Wrapf(
		"nft %s in class %s not found",
		nftID,
		classID,
	)
}

func (k Keeper) getBurnTombstone(
	ctx sdk.Context,
	classID string,
	nftID string,
	tombstone types.BurnTombstone,
) (*types.BurnTombstone, error) {
	classRecord, err := k.getClassRecord(ctx, classID)
	if errors.Is(err, upstreamnft.ErrClassNotExists) {
		return nil, fmt.Errorf(
			"nft tombstone %s in class %s references missing class state",
			nftID,
			classID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("load class state for nft tombstone %s: %w", nftID, err)
	}
	liveSupply := k.nftKeeper.GetTotalSupply(ctx, classID)
	if classRecord.MintedCount == 0 || liveSupply >= classRecord.MintedCount {
		return nil, fmt.Errorf(
			"nft tombstone %s in class %s has inconsistent supply %d and minted count %d",
			nftID,
			classID,
			liveSupply,
			classRecord.MintedCount,
		)
	}

	canonical, err := k.canonicalBurnTombstone(classID, nftID, tombstone)
	if err != nil {
		return nil, err
	}
	return &canonical, nil
}

func (k Keeper) canonicalBurnTombstone(
	classID string,
	nftID string,
	tombstone types.BurnTombstone,
) (types.BurnTombstone, error) {
	if tombstone.ClassId != classID || tombstone.NftId != nftID {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone key does not match value for %s/%s",
			classID,
			nftID,
		)
	}
	if tombstone.Mint.MintedAt.IsZero() && tombstone.Mint.MintedBy == "" {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone has no mint record for %s/%s",
			classID,
			nftID,
		)
	}
	if tombstone.Mint.MintedAt.IsZero() {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone has no mint time for %s/%s",
			classID,
			nftID,
		)
	}
	canonicalMinter, _, err := k.canonicalNonModuleAccount(
		"stored tombstone minted_by",
		tombstone.Mint.MintedBy,
	)
	if err != nil {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone has invalid minter for %s/%s: %w",
			classID,
			nftID,
			err,
		)
	}
	if canonicalMinter != tombstone.Mint.MintedBy {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone minter is not canonical for %s/%s",
			classID,
			nftID,
		)
	}
	if err := types.ValidateURI(tombstone.Uri, tombstone.UriHash); err != nil {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone %s has invalid stored URI metadata: %w",
			nftID,
			err,
		)
	}
	canonicalData, err := types.CanonicalizeNFTData(k.cdc, tombstone.Data)
	if err != nil {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone %s has invalid stored data: %w",
			nftID,
			err,
		)
	}
	if tombstone.Revocation != nil {
		if tombstone.Revocation.RevokedAt.IsZero() {
			return types.BurnTombstone{}, fmt.Errorf(
				"nft tombstone has no revocation time for %s/%s",
				classID,
				nftID,
			)
		}
		if tombstone.Revocation.RevokedAt.Before(tombstone.Mint.MintedAt) {
			return types.BurnTombstone{}, fmt.Errorf(
				"nft tombstone revocation predates mint for %s/%s",
				classID,
				nftID,
			)
		}
		canonicalRevoker, _, err := k.canonicalNonModuleAccount(
			"stored tombstone revoked_by",
			tombstone.Revocation.RevokedBy,
		)
		if err != nil {
			return types.BurnTombstone{}, fmt.Errorf(
				"nft tombstone has invalid revoker for %s/%s: %w",
				classID,
				nftID,
				err,
			)
		}
		if canonicalRevoker != tombstone.Revocation.RevokedBy {
			return types.BurnTombstone{}, fmt.Errorf(
				"nft tombstone revoker is not canonical for %s/%s",
				classID,
				nftID,
			)
		}
	}
	if tombstone.BurnedAt.IsZero() {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone has no burn time for %s/%s",
			classID,
			nftID,
		)
	}
	if tombstone.BurnedAt.Before(tombstone.Mint.MintedAt) {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone burn predates mint for %s/%s",
			classID,
			nftID,
		)
	}
	if tombstone.Revocation != nil && tombstone.BurnedAt.Before(tombstone.Revocation.RevokedAt) {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone burn predates revocation for %s/%s",
			classID,
			nftID,
		)
	}
	canonicalBurner, _, err := k.canonicalNonModuleAccount(
		"stored tombstone burned_by",
		tombstone.BurnedBy,
	)
	if err != nil {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone has invalid burner for %s/%s: %w",
			classID,
			nftID,
			err,
		)
	}
	if canonicalBurner != tombstone.BurnedBy {
		return types.BurnTombstone{}, fmt.Errorf(
			"nft tombstone burner is not canonical for %s/%s",
			classID,
			nftID,
		)
	}

	tombstone.Data = canonicalData
	return tombstone, nil
}

func (k Keeper) ensureNFTTransferAllowed(ctx sdk.Context, classID, nftID string) error {
	state, record, err := k.loadLiveNFTState(ctx, classID, nftID)
	if err != nil {
		return err
	}
	if state.lifecycle.Revocation != nil {
		return types.ErrNFTRevoked.Wrapf(
			"nft %s in class %s is revoked",
			nftID,
			classID,
		)
	}

	if record.Policy.TransferPolicy != types.TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE {
		return types.ErrTransferNotAllowed.Wrapf(
			"class %s does not allow owner transfers",
			classID,
		)
	}
	return nil
}

func (k Keeper) loadLiveNFTState(
	ctx sdk.Context,
	classID string,
	nftID string,
) (storedNFTState, *types.ClassRecord, error) {
	state, err := k.loadLiveNFTStateOnly(ctx, classID, nftID)
	if err != nil {
		return storedNFTState{}, nil, err
	}
	record, err := k.getClassRecord(ctx, classID)
	if errors.Is(err, upstreamnft.ErrClassNotExists) {
		return storedNFTState{}, nil, fmt.Errorf(
			"live nft %s in class %s references missing class state",
			nftID,
			classID,
		)
	}
	if err != nil {
		return storedNFTState{}, nil, fmt.Errorf("load class state for nft %s: %w", nftID, err)
	}
	return state, record, nil
}

func (k Keeper) loadLiveNFTStateOnly(
	ctx sdk.Context,
	classID string,
	nftID string,
) (storedNFTState, error) {
	state, err := k.loadNFTState(ctx, classID, nftID)
	if err != nil {
		return storedNFTState{}, err
	}
	if err := state.validateLiveCombination(classID, nftID); err != nil {
		return storedNFTState{}, err
	}
	if !state.hasNFT {
		return storedNFTState{}, upstreamnft.ErrNFTNotExists.Wrapf(
			"nft %s in class %s not found",
			nftID,
			classID,
		)
	}
	return state, nil
}

func (k Keeper) loadNFTState(
	ctx sdk.Context,
	classID string,
	nftID string,
) (storedNFTState, error) {
	state := storedNFTState{}
	state.nft, state.hasNFT = k.nftKeeper.GetNFT(ctx, classID, nftID)

	key := collections.Join(classID, nftID)
	lifecycle, err := k.lifecycles.Get(ctx, key)
	switch {
	case err == nil:
		state.lifecycle = lifecycle
		state.hasLifecycle = true
	case errors.Is(err, collections.ErrNotFound):
	default:
		return storedNFTState{}, fmt.Errorf("load nft lifecycle: %w", err)
	}
	tombstone, err := k.tombstones.Get(ctx, key)
	switch {
	case err == nil:
		state.tombstone = tombstone
		state.hasTombstone = true
	case errors.Is(err, collections.ErrNotFound):
	default:
		return storedNFTState{}, fmt.Errorf("load nft tombstone: %w", err)
	}

	if state.hasNFT && (state.nft.ClassId != classID || state.nft.Id != nftID) {
		return storedNFTState{}, fmt.Errorf(
			"standard nft key does not match value for %s/%s",
			classID,
			nftID,
		)
	}
	if state.hasLifecycle &&
		(state.lifecycle.ClassId != classID || state.lifecycle.NftId != nftID) {
		return storedNFTState{}, fmt.Errorf(
			"nft lifecycle key does not match value for %s/%s",
			classID,
			nftID,
		)
	}
	if state.hasLifecycle &&
		state.lifecycle.Mint.MintedAt.IsZero() && state.lifecycle.Mint.MintedBy == "" {
		return storedNFTState{}, fmt.Errorf("nft lifecycle has no mint record for %s/%s", classID, nftID)
	}
	if state.hasLifecycle && state.lifecycle.Mint.MintedAt.IsZero() {
		return storedNFTState{}, fmt.Errorf("nft lifecycle has no mint time for %s/%s", classID, nftID)
	}
	if state.hasLifecycle {
		canonicalMinter, _, err := k.canonicalNonModuleAccount("stored minted_by", state.lifecycle.Mint.MintedBy)
		if err != nil {
			return storedNFTState{}, fmt.Errorf("nft lifecycle has invalid minter for %s/%s: %w", classID, nftID, err)
		}
		if canonicalMinter != state.lifecycle.Mint.MintedBy {
			return storedNFTState{}, fmt.Errorf(
				"nft lifecycle minter is not canonical for %s/%s",
				classID,
				nftID,
			)
		}
	}
	if state.hasLifecycle && state.lifecycle.Revocation != nil {
		revocation := state.lifecycle.Revocation
		if revocation.RevokedAt.IsZero() {
			return storedNFTState{}, fmt.Errorf(
				"nft lifecycle has no revocation time for %s/%s",
				classID,
				nftID,
			)
		}
		if revocation.RevokedAt.Before(state.lifecycle.Mint.MintedAt) {
			return storedNFTState{}, fmt.Errorf(
				"nft lifecycle revocation predates mint for %s/%s",
				classID,
				nftID,
			)
		}
		canonicalRevoker, _, err := k.canonicalNonModuleAccount("stored revoked_by", revocation.RevokedBy)
		if err != nil {
			return storedNFTState{}, fmt.Errorf(
				"nft lifecycle has invalid revoker for %s/%s: %w",
				classID,
				nftID,
				err,
			)
		}
		if canonicalRevoker != revocation.RevokedBy {
			return storedNFTState{}, fmt.Errorf(
				"nft lifecycle revoker is not canonical for %s/%s",
				classID,
				nftID,
			)
		}
	}
	if state.hasTombstone &&
		(state.tombstone.ClassId != classID || state.tombstone.NftId != nftID) {
		return storedNFTState{}, fmt.Errorf(
			"nft tombstone key does not match value for %s/%s",
			classID,
			nftID,
		)
	}
	return state, nil
}

func (state storedNFTState) ensureAvailableForMint(classID, nftID string) error {
	if err := state.validateLiveCombination(classID, nftID); err != nil {
		return err
	}
	if state.hasTombstone {
		return types.ErrNFTIDPermanentlyUsed.Wrapf(
			"nft %s in class %s was already minted",
			nftID,
			classID,
		)
	}
	if state.hasNFT {
		return upstreamnft.ErrNFTExists.Wrapf(
			"nft %s in class %s already exists",
			nftID,
			classID,
		)
	}
	return nil
}

func (state storedNFTState) validateLiveCombination(classID, nftID string) error {
	if state.hasNFT != state.hasLifecycle ||
		(state.hasTombstone && (state.hasNFT || state.hasLifecycle)) {
		return fmt.Errorf(
			"nft %s in class %s has inconsistent standard, lifecycle, and tombstone state",
			nftID,
			classID,
		)
	}
	return nil
}

func nextMintedCount(current, maxSupply uint64) (uint64, error) {
	if current == math.MaxUint64 || (maxSupply != 0 && current >= maxSupply) {
		return 0, types.ErrMaxSupplyReached.Wrap("class lifetime mint limit reached")
	}
	return current + 1, nil
}
