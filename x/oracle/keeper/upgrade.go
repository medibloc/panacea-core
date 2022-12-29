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

func (k Keeper) UpgradeOracle(ctx sdk.Context, msg *types.MsgUpgradeOracle) error {
	oracleUpgrade := types.NewUpgradeOracle(msg)

	if err := oracleUpgrade.ValidateBasic(); err != nil {
		return err
	}

	upgradeInfo, err := k.GetOracleUpgradeInfo(ctx)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrUpgradeOracle, "failed to get oracle upgrade info")
	}
	if oracleUpgrade.UniqueId != upgradeInfo.UniqueId {
		return sdkerrors.Wrapf(types.ErrUpgradeOracle, "does not match the upgrade uniqueID")
	}

	if _, err := k.GetOracle(ctx, oracleUpgrade.OracleAddress); err != nil {
		return sdkerrors.Wrapf(types.ErrUpgradeOracle, "is not registered oracle")
	}

	if err := k.SetOracleUpgrade(ctx, oracleUpgrade); err != nil {
		return sdkerrors.Wrapf(types.ErrUpgradeOracle, "failed to set oracle upgrade info")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpgrade,
			sdk.NewAttribute(types.AttributeKeyUniqueID, oracleUpgrade.UniqueId),
			sdk.NewAttribute(types.AttributeKeyOracleAddress, oracleUpgrade.OracleAddress),
		),
	)
	return nil
}

func (k Keeper) GetOracleUpgrade(ctx sdk.Context, uniqueID, address string) (*types.OracleUpgrade, error) {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}
	key := types.GetOracleUpgradeKey(uniqueID, accAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrGetOracleUpgrade, "oracle registration not found")
	}

	oracleUpgrade := &types.OracleUpgrade{}

	if err := k.cdc.UnmarshalLengthPrefixed(bz, oracleUpgrade); err != nil {
		return nil, err
	}

	return oracleUpgrade, nil
}

func (k Keeper) SetOracleUpgrade(ctx sdk.Context, upgrade *types.OracleUpgrade) error {
	store := ctx.KVStore(k.storeKey)

	accAddr, err := sdk.AccAddressFromBech32(upgrade.OracleAddress)
	if err != nil {
		return err
	}
	key := types.GetOracleUpgradeKey(upgrade.UniqueId, accAddr)
	bz, err := k.cdc.MarshalLengthPrefixed(upgrade)
	if err != nil {
		return err
	}

	store.Set(key, bz)
	return nil
}

func (k Keeper) GetAllOracleUpgradeList(ctx sdk.Context) ([]types.OracleUpgrade, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.OracleUpgradeKey)
	defer iterator.Close()

	oracleUpgrades := make([]types.OracleUpgrade, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var oracleUpgrade types.OracleUpgrade

		err := k.cdc.UnmarshalLengthPrefixed(bz, &oracleUpgrade)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetOracleUpgrade, err.Error())
		}

		oracleUpgrades = append(oracleUpgrades, oracleUpgrade)
	}

	return oracleUpgrades, nil
}
