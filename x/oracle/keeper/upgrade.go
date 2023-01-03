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

func (k Keeper) ApplyUpgrade(ctx sdk.Context, info *types.OracleUpgradeInfo) error {

	params := k.GetParams(ctx)
	params.UniqueId = info.UniqueId
	if err := params.Validate(); err != nil {
		return err
	}
	k.SetParams(ctx, params)

	iterator := k.GetOracleUpgradeIterator(ctx, info.UniqueId)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		uniqueID, address := types.SplitOracleUpgradeKey(iterator.Key())
		oracleUpgrade, err := k.GetOracleUpgrade(ctx, uniqueID, address.String())
		if err != nil {
			return err
		}
		if oracleUpgrade.EncryptedOraclePrivKey != nil {
			oracle, err := k.GetOracle(ctx, address.String())
			if err != nil {
				return err
			}
			oracle.UniqueId = info.UniqueId
			if err := k.SetOracle(ctx, oracle); err != nil {
				return err
			}
		}
	}

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

func (k Keeper) ApproveOracleUpgrade(ctx sdk.Context, msg *types.MsgApproveOracleUpgrade) error {
	approval := msg.GetApprovalSharingOracleKey()

	// validate approval for oracle upgrade
	if err := k.validateApprovalSharingOracleKey(ctx, approval, msg.GetSignature()); err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleUpgrade, err.Error())
	}

	// get oracle upgrade and upgrade info
	upgradeInfo, err := k.GetOracleUpgradeInfo(ctx)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleUpgrade, err.Error())
	}

	oracleUpgrade, err := k.GetOracleUpgrade(ctx, upgradeInfo.GetUniqueId(), approval.GetTargetOracleAddress())
	if err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleUpgrade, err.Error())
	}

	// check unique ID
	if approval.TargetUniqueId != upgradeInfo.GetUniqueId() {
		return sdkerrors.Wrapf(types.ErrApproveOracleUpgrade, types.ErrInvalidUniqueID.Error())
	}

	// if EncryptedOraclePrivKey is already set, return error
	if oracleUpgrade.EncryptedOraclePrivKey != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleUpgrade, "already approved oracle upgrade. if you want to be shared oracle private key again, please upgrade oracle again")
	}

	// update encrypted oracle private key
	oracleUpgrade.EncryptedOraclePrivKey = approval.EncryptedOraclePrivKey

	// set oracle upgrade
	if err := k.SetOracleUpgrade(ctx, oracleUpgrade); err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleUpgrade, err.Error())
	}

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeApproveOracleUpgrade,
			sdk.NewAttribute(types.AttributeKeyOracleAddress, approval.GetTargetOracleAddress()),
			sdk.NewAttribute(types.AttributeKeyUniqueID, upgradeInfo.GetUniqueId()),
		),
	)

	return nil
}

func (k Keeper) GetOracleUpgradeIterator(ctx sdk.Context, uniqueID string) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, types.GetOracleUpgradeByUniqueIDKey(uniqueID))
}
