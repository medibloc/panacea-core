package keeper

import (
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
)

func (k Keeper) MintPNFT(
	ctx sdk.Context,
	pnft *types.PNFT,
) error {
	denom, err := k.GetDenom(ctx, pnft.DenomId)
	if err != nil {
		return err
	}

	if denom.Owner != pnft.Creator {
		return fmt.Errorf("permission denied. %s does not have permission to mint pnft. Only possible to owner of denom", pnft.Creator)
	}

	meta, err := codectypes.NewAnyWithValue(
		&types.PNFTMeta{
			Name:        pnft.Name,
			Description: pnft.Description,
			Creator:     pnft.Creator,
			CreatedAt:   pnft.CreatedAt,
			Data:        pnft.Data,
		})
	if err != nil {
		return err
	}

	sdkNFT := nft.NFT{
		ClassId: denom.Id,
		Id:      pnft.Id,
		Uri:     pnft.Uri,
		UriHash: pnft.UriHash,
		Data:    meta,
	}

	receiver, err := sdk.AccAddressFromBech32(pnft.Creator)
	if err != nil {
		return err
	}

	if err := k.nftKeeper.Mint(ctx, sdkNFT, receiver); err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventMintPNFT{
		DenomId: pnft.DenomId,
		Id:      pnft.Id,
		Creator: pnft.Creator,
	})
}

func (k Keeper) TransferPNFT(
	ctx sdk.Context,
	denomId string,
	id string,
	sender string,
	receiver string,
) error {
	pnft, err := k.GetPNFT(ctx, denomId, id)
	if err != nil {
		return err
	}

	if sender != pnft.Owner {
		return fmt.Errorf("permission denied. %s does not have permission to transfer pnft. Only possible to owner of pnft", sender)
	}

	receiverAddr, err := sdk.AccAddressFromBech32(receiver)
	if err != nil {
		return err
	}

	if err := k.nftKeeper.Transfer(ctx, denomId, id, receiverAddr); err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventTransferPNFT{
		DenomId:  pnft.DenomId,
		Id:       pnft.Id,
		Sender:   sender,
		Receiver: receiver,
	})
}

func (k Keeper) BurnPNFT(
	ctx sdk.Context,
	denomId string,
	id string,
	burner string,
) error {
	pnft, err := k.GetPNFT(ctx, denomId, id)
	if err != nil {
		return err
	}

	if burner != pnft.Owner {
		return fmt.Errorf("permission denied. %s does not have permission to burn pnft. Only possible to owner of pnft", burner)
	}

	if err := k.nftKeeper.Burn(ctx, denomId, id); err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventBurnPNFT{
		DenomId: pnft.DenomId,
		Id:      pnft.Id,
		Burner:  burner,
	})
}

func (k Keeper) GetPNFTsByDenomId(ctx sdk.Context, denomId string) ([]*types.PNFT, error) {
	var pnfts []*types.PNFT
	for _, n := range k.nftKeeper.GetNFTsOfClass(ctx, denomId) {
		var meta types.PNFTMeta
		if err := k.cdc.Unmarshal(n.Data.GetValue(), &meta); err != nil {
			return nil, err
		}

		ownerAddr := k.nftKeeper.GetOwner(ctx, denomId, n.Id)

		pnfts = append(pnfts, &types.PNFT{
			DenomId:     n.ClassId,
			Id:          n.Id,
			Name:        meta.Name,
			Description: meta.Description,
			Uri:         n.Uri,
			UriHash:     n.UriHash,
			Data:        meta.Data,
			Creator:     meta.Creator,
			Owner:       ownerAddr.String(),
			CreatedAt:   meta.CreatedAt,
		})
	}

	return pnfts, nil
}

func (k Keeper) GetPNFTsByDenomIdAndOwner(ctx sdk.Context, denomId, owner string) ([]*types.PNFT, error) {
	var pnfts []*types.PNFT
	ownerAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return nil, err
	}

	for _, n := range k.nftKeeper.GetNFTsOfClassByOwner(ctx, denomId, ownerAddr) {
		var meta types.PNFTMeta
		if err := k.cdc.Unmarshal(n.Data.GetValue(), &meta); err != nil {
			return nil, err
		}

		ownerAddr := k.nftKeeper.GetOwner(ctx, denomId, n.Id)

		pnfts = append(pnfts, &types.PNFT{
			DenomId:     n.ClassId,
			Id:          n.Id,
			Name:        meta.Name,
			Description: meta.Description,
			Uri:         n.Uri,
			UriHash:     n.UriHash,
			Data:        meta.Data,
			Creator:     meta.Creator,
			Owner:       ownerAddr.String(),
			CreatedAt:   meta.CreatedAt,
		})
	}

	return pnfts, nil
}

func (k Keeper) GetPNFT(
	ctx sdk.Context,
	denomId string,
	id string,
) (*types.PNFT, error) {
	nft, exist := k.nftKeeper.GetNFT(ctx, denomId, id)
	if !exist {
		return nil, fmt.Errorf("cannot found pnft. denomId: %s, pnftId: %s", denomId, id)
	}

	ownerAddr := k.nftKeeper.GetOwner(ctx, denomId, id)

	var meta types.PNFTMeta
	if err := k.cdc.Unmarshal(nft.Data.GetValue(), &meta); err != nil {
		return nil, err
	}
	return &types.PNFT{
		DenomId:     nft.ClassId,
		Id:          nft.Id,
		Name:        meta.Name,
		Description: meta.Description,
		Uri:         nft.Uri,
		UriHash:     nft.UriHash,
		Data:        meta.Data,
		Creator:     meta.Creator,
		Owner:       ownerAddr.String(),
		CreatedAt:   meta.CreatedAt,
	}, nil
}
