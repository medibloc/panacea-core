package keeper

import (
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/medibloc/panacea-core/v2/x/pnft/types"
)

func (k Keeper) SaveDenom(
	ctx sdk.Context,
	denom *types.Denom,
) error {
	class, err := types.NewClassFromDenom(k.cdc, denom)
	if err != nil {
		return err
	}

	if err := k.nftKeeper.SaveClass(ctx, *class); err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventSaveDenom{
		Id:      denom.Id,
		Creator: denom.Creator,
	})
}

func (k Keeper) ParseDenoms(classes []*nft.Class) ([]*types.Denom, error) {
	denoms := []*types.Denom{}
	for _, class := range classes {
		denom, err := types.NewDenomFromClass(k.cdc, class)
		if err != nil {
			return nil, err
		}

		denoms = append(denoms, denom)
	}

	return denoms, nil
}

func (k Keeper) UpdateDenom(ctx sdk.Context, msg *types.Denom, updater string) error {
	denom, err := k.GetDenom(ctx, msg.GetId())
	if err != nil {
		return err
	}

	if updater != denom.Creator {
		return errors.New(fmt.Sprintf("permission denied: %s does not have permission to modify this resource.", updater))
	}

	if msg.Name != "" {
		denom.Name = msg.Name
	}
	if msg.Symbol != "" {
		denom.Symbol = msg.Symbol
	}
	if msg.Description != "" {
		denom.Description = msg.Description
	}
	if msg.Uri != "" {
		denom.Uri = msg.Uri
	}
	if msg.UriHash != "" {
		denom.UriHash = msg.UriHash
	}
	if msg.Data != "" {
		denom.Data = msg.Name
	}

	class, err := types.NewClassFromDenom(k.cdc, denom)
	if err != nil {
		return err
	}

	if err := k.nftKeeper.UpdateClass(ctx, *class); err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventUpdateDenom{
		Id:      denom.Id,
		Updater: updater,
	})
}

func (k Keeper) DeleteDenom(ctx sdk.Context, id string, remover string) error {
	denom, err := k.GetDenom(ctx, id)
	if err != nil {
		return err
	}

	if remover != denom.Creator {
		return errors.New(fmt.Sprintf("permission denied: %s does not have permission to remove this resource.", remover))
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(classStoreKey(id))

	return ctx.EventManager().EmitTypedEvent(&types.EventDeleteDenom{
		Id:      denom.Id,
		Remover: remover,
	})
}

func (k Keeper) TransferDenomOwner(
	ctx sdk.Context,
	id string,
	sender string,
	receiver string,
) error {
	denom, err := k.GetDenom(ctx, id)
	if err != nil {
		return err
	}

	if sender != denom.Creator {
		return errors.New(fmt.Sprintf("%s is not allowed transfer denom to %s", sender, receiver))
	}

	denom.Creator = receiver
	class, err := types.NewClassFromDenom(k.cdc, denom)
	if err != nil {
		return err
	}

	if err := k.nftKeeper.UpdateClass(ctx, *class); err != nil {
		return err
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventTransferDenom{
		Id:       denom.Id,
		Sender:   sender,
		Receiver: receiver,
	})
}

func (k Keeper) GetAllDenoms(ctx sdk.Context) ([]*types.Denom, error) {
	var denoms []*types.Denom
	classes := k.nftKeeper.GetClasses(ctx)

	for _, class := range classes {
		denom, err := types.NewDenomFromClass(k.cdc, class)
		if err != nil {
			return nil, err
		}
		denoms = append(denoms, denom)
	}
	return denoms, nil
}

func (k Keeper) GetDenom(ctx sdk.Context, id string) (*types.Denom, error) {
	class, found := k.nftKeeper.GetClass(ctx, id)
	if !found {
		return nil, errors.New("not found class.")
	}

	return types.NewDenomFromClass(k.cdc, &class)
}
