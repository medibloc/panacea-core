package keeper

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
)

func (k Keeper) RegisterOracle(ctx sdk.Context, msg *types.MsgRegisterOracle) error {
	oracleRegistration := types.NewOracleRegistration(msg)

	if err := oracleRegistration.ValidateBasic(); err != nil {
		return err
	}

	params := k.GetParams(ctx)
	if params.UniqueId != oracleRegistration.UniqueId {
		return sdkerrors.Wrapf(types.ErrOracleRegistration, "is not match the currently active uniqueID")
	}

	if oracle, err := k.GetOracle(ctx, oracleRegistration.OracleAddress); !errors.Is(types.ErrOracleNotFound, err) {
		if oracle != nil {
			return sdkerrors.Wrapf(types.ErrOracleRegistration, "already registered oracle. address(%s)", oracleRegistration.OracleAddress)
		} else {
			return sdkerrors.Wrapf(types.ErrOracleRegistration, err.Error())
		}
	}

	if err := k.SetOracleRegistration(ctx, oracleRegistration); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRegistration,
			sdk.NewAttribute(types.AttributeKeyUniqueID, oracleRegistration.UniqueId),
			sdk.NewAttribute(types.AttributeKeyOracleAddress, oracleRegistration.OracleAddress),
		),
	)
	return nil
}

func (k Keeper) ApproveOracleRegistration(ctx sdk.Context, msg *types.MsgApproveOracleRegistration) error {

	if err := k.validateApproveOracleRegistration(ctx, msg); err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleRegistration, err.Error())
	}

	oracleRegistration, err := k.GetOracleRegistration(ctx, msg.ApproveOracleRegistration.UniqueId, msg.ApproveOracleRegistration.TargetOracleAddress)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleRegistration, err.Error())
	}

	oracleRegistration.EncryptedOraclePrivKey = msg.ApproveOracleRegistration.EncryptedOraclePrivKey

	// add an encrypted oracle private key to oracleRegistration
	if err := k.SetOracleRegistration(ctx, oracleRegistration); err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleRegistration, err.Error())
	}

	newOracle := types.NewOracle(
		msg.ApproveOracleRegistration.TargetOracleAddress,
		msg.ApproveOracleRegistration.UniqueId,
		oracleRegistration.Endpoint,
		oracleRegistration.OracleCommissionRate,
		oracleRegistration.OracleCommissionMaxRate,
		oracleRegistration.OracleCommissionMaxChangeRate,
		ctx.BlockTime(),
	)

	// append new oracle info
	if err := k.SetOracle(ctx, newOracle); err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleRegistration, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeApproveOracleRegistration,
			sdk.NewAttribute(types.AttributeKeyOracleAddress, msg.ApproveOracleRegistration.TargetOracleAddress),
			sdk.NewAttribute(types.AttributeKeyUniqueID, msg.ApproveOracleRegistration.UniqueId),
		),
	)

	return nil

}

// validateApproveOracleRegistration checks signature
func (k Keeper) validateApproveOracleRegistration(ctx sdk.Context, msg *types.MsgApproveOracleRegistration) error {

	params := k.GetParams(ctx)
	targetOracleAddress := msg.ApproveOracleRegistration.TargetOracleAddress

	// check unique id
	if msg.ApproveOracleRegistration.UniqueId != params.UniqueId {
		return types.ErrInvalidUniqueID
	}

	// verify signature
	if err := k.VerifyOracleSignature(ctx, msg.ApproveOracleRegistration, msg.Signature); err != nil {
		return err
	}

	if msg.ApproveOracleRegistration.EncryptedOraclePrivKey == nil {
		return fmt.Errorf("encrypted oracle private key is nil")
	}

	// check if the oracle has been already registered
	hasOracle, err := k.HasOracle(ctx, targetOracleAddress)
	if err != nil {
		return err
	}
	if hasOracle {
		return fmt.Errorf("already registered oracle. address(%s)", targetOracleAddress)
	}

	return nil
}

func (k Keeper) UpdateOracleInfo(ctx sdk.Context, msg *types.MsgUpdateOracleInfo) error {
	oracle, err := k.GetOracle(ctx, msg.OracleAddress)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrUpdateOracle, err.Error())
	}

	if msg.OracleCommissionRate != nil && !oracle.OracleCommissionRate.Equal(*msg.OracleCommissionRate) {
		blockTime := ctx.BlockHeader().Time
		if err := oracle.ValidateOracleCommission(blockTime, *msg.OracleCommissionRate); err != nil {
			return sdkerrors.Wrapf(types.ErrUpdateOracle, err.Error())
		}
		oracle.OracleCommissionRate = *msg.OracleCommissionRate
		oracle.UpdateTime = blockTime
	}

	if len(msg.Endpoint) > 0 {
		oracle.Endpoint = msg.Endpoint
	}

	if err := k.SetOracle(ctx, oracle); err != nil {
		return sdkerrors.Wrapf(types.ErrUpdateOracle, err.Error())
	}
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
		return sdkerrors.Wrapf(types.ErrUpgradeOracle, "is not match the upgrade uniqueID")
	}

	if err := k.SetOracleUpgrade(ctx, oracleUpgrade); err != nil {
		return err
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

func (k Keeper) SetOracleRegistration(ctx sdk.Context, regOracle *types.OracleRegistration) error {
	store := ctx.KVStore(k.storeKey)

	accAddr, err := sdk.AccAddressFromBech32(regOracle.OracleAddress)
	if err != nil {
		return err
	}
	key := types.GetOracleRegistrationKey(regOracle.UniqueId, accAddr)
	bz, err := k.cdc.MarshalLengthPrefixed(regOracle)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetAllOracleRegistrationList(ctx sdk.Context) ([]types.OracleRegistration, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.OracleRegistrationKey)
	defer iterator.Close()

	oracleRegistrations := make([]types.OracleRegistration, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var oracleRegistration types.OracleRegistration
		err := k.cdc.UnmarshalLengthPrefixed(bz, &oracleRegistration)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetOracleRegistration, err.Error())
		}

		oracleRegistrations = append(oracleRegistrations, oracleRegistration)
	}

	return oracleRegistrations, nil

}

func (k Keeper) GetOracleRegistration(ctx sdk.Context, uniqueID, address string) (*types.OracleRegistration, error) {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}
	key := types.GetOracleRegistrationKey(uniqueID, accAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrGetOracleRegistration, "oracle registration not found")
	}

	oracleRegistration := &types.OracleRegistration{}

	if err := k.cdc.UnmarshalLengthPrefixed(bz, oracleRegistration); err != nil {
		return nil, err
	}

	return oracleRegistration, nil
}

func (k Keeper) SetOracle(ctx sdk.Context, oracle *types.Oracle) error {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(oracle.OracleAddress)
	if err != nil {
		return err
	}
	key := types.GetOracleKey(accAddr)
	bz, err := k.cdc.MarshalLengthPrefixed(oracle)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}

func (k Keeper) GetAllOracleList(ctx sdk.Context) ([]types.Oracle, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.OraclesKey)
	defer iterator.Close()

	oracles := make([]types.Oracle, 0)

	for ; iterator.Valid(); iterator.Next() {
		bz := iterator.Value()
		var oracle types.Oracle

		err := k.cdc.UnmarshalLengthPrefixed(bz, &oracle)
		if err != nil {
			return nil, sdkerrors.Wrapf(types.ErrGetOracle, err.Error())
		}

		oracles = append(oracles, oracle)
	}

	return oracles, nil
}

func (k Keeper) GetOracle(ctx sdk.Context, address string) (*types.Oracle, error) {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}
	key := types.GetOracleKey(accAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, sdkerrors.Wrapf(types.ErrOracleNotFound, "oracle '%s' does not exist", address)
	}

	oracle := &types.Oracle{}

	err = k.cdc.UnmarshalLengthPrefixed(bz, oracle)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrGetOracle, err.Error())
	}

	return oracle, nil
}

func (k Keeper) HasOracle(ctx sdk.Context, address string) (bool, error) {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return false, err
	}

	return store.Has(types.GetOracleKey(accAddr)), nil
}
