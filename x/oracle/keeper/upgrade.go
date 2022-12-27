package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

// SetOracleUpgradeInfo stores oracle upgrade information.
func (k Keeper) SetOracleUpgradeInfo(ctx sdk.Context, upgradeInfo *types.OracleUpgradeInfo) error {
	store := ctx.KVStore(k.storeKey)

	bz, err := k.cdc.MarshalLengthPrefixed(upgradeInfo)
	if err != nil {
		return err
	}

	store.Set(types.OracleUpgradeInfoKey, bz)

	return nil
}

// GetOracleUpgradeInfo gets oracle upgrade information.
func (k Keeper) GetOracleUpgradeInfo(ctx sdk.Context) (*types.OracleUpgradeInfo, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.OracleUpgradeInfoKey)
	if bz == nil {
		return nil, types.ErrOracleUpgradeInfoNotFound
	}

	var upgradeInfo types.OracleUpgradeInfo
	if err := k.cdc.UnmarshalLengthPrefixed(bz, &upgradeInfo); err != nil {
		return nil, sdkerrors.Wrap(types.ErrGetOracleUpgradeInfo, err.Error())
	}

	return &upgradeInfo, nil
}

func (k Keeper) RemoveOracleUpgradeInfo(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.OracleUpgradeInfoKey)
}

func (k Keeper) ApplyUpgrade(ctx sdk.Context, info *types.OracleUpgradeInfo) error {
	params := k.GetParams(ctx)
	params.UniqueId = info.UniqueId
	if err := params.Validate(); err != nil {
		return err
	}
	k.SetParams(ctx, params)

	//TODO: update `Oracles` that already upgraded

	ctx.Logger().Info("Oracle upgrade was successful.", fmt.Sprintf("uniqueID: %s, height: %v", info.UniqueId, info.Height))
	return nil
}
