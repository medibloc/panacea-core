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
	if err := k.VerifySignature(ctx, msg.ApproveOracleRegistration, msg.Signature); err != nil {
		return err
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

	blockTime := ctx.BlockHeader().Time

	if !oracle.OracleCommissionRate.Equal(msg.OracleCommissionRate) {
		if err := oracle.ValidateOracleCommission(blockTime, msg.OracleCommissionRate); err != nil {
			return sdkerrors.Wrapf(types.ErrUpdateOracle, err.Error())
		}
	}

	oracle.UpdateTime = blockTime
	if len(msg.Endpoint) > 0 {
		oracle.Endpoint = msg.Endpoint
	}
	oracle.OracleCommissionRate = msg.OracleCommissionRate

	if err := k.SetOracle(ctx, oracle); err != nil {
		return sdkerrors.Wrapf(types.ErrUpdateOracle, err.Error())
	}
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
