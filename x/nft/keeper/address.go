package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/nft/types"
)

func (k Keeper) canonicalNonModuleAccount(
	ctx context.Context,
	field string,
	value string,
) (string, sdk.AccAddress, error) {
	canonical, address, err := k.canonicalAddress(field, value)
	if err != nil {
		return "", nil, err
	}
	if _, isModuleAccount := k.accountKeeper.GetAccount(ctx, address).(sdk.ModuleAccountI); isModuleAccount {
		return "", nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"%s must not be a module account",
			field,
		)
	}
	return canonical, address, nil
}

func (k Keeper) validateCanonicalClassID(classID string) error {
	creator, _, err := types.ParseClassID(classID)
	if err != nil {
		return err
	}
	canonicalCreator, _, err := k.canonicalAddress("class creator", creator)
	if err != nil {
		return err
	}
	if creator != canonicalCreator {
		return sdkerrors.ErrInvalidRequest.Wrap("class_id creator must use its canonical address")
	}
	return nil
}

func (k Keeper) canonicalAddress(field, value string) (string, sdk.AccAddress, error) {
	addressBytes, err := k.addressCodec.StringToBytes(value)
	if err != nil {
		return "", nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidAddress,
			"invalid %s address: %v",
			field,
			err,
		)
	}
	canonical, err := k.addressCodec.BytesToString(addressBytes)
	if err != nil {
		return "", nil, fmt.Errorf("encode canonical %s address: %w", field, err)
	}
	return canonical, sdk.AccAddress(addressBytes), nil
}
