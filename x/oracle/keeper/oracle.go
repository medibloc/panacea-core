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

	params := k.GetParams(ctx)
	if params.UniqueId != msg.UniqueId {
		return sdkerrors.Wrapf(types.ErrRegisterOracle, types.ErrInvalidUniqueID.Error())
	}

	existing, err := k.GetOracleRegistration(ctx, msg.GetUniqueId(), msg.GetOracleAddress())
	if err != nil && !errors.Is(err, types.ErrOracleRegistrationNotFound) {
		return sdkerrors.Wrapf(types.ErrRegisterOracle, err.Error())
	}
	if existing != nil {
		return sdkerrors.Wrapf(types.ErrRegisterOracle, fmt.Sprintf("already registered oracle. address(%s)", msg.OracleAddress))
	}

	if err := k.SetOracleRegistration(ctx, oracleRegistration); err != nil {
		return sdkerrors.Wrapf(types.ErrRegisterOracle, err.Error())
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
	// validate approval for oracle registration
	if err := k.validateApprovalSharingOracleKey(ctx, msg.GetApprovalSharingOracleKey(), msg.GetSignature()); err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleRegistration, err.Error())
	}

	// get oracle registration
	oracleRegistration, err := k.GetOracleRegistration(ctx, msg.ApprovalSharingOracleKey.UniqueId, msg.ApprovalSharingOracleKey.TargetOracleAddress)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleRegistration, err.Error())
	}

	// if EncryptedOraclePrivKey is already set, return error
	if oracleRegistration.EncryptedOraclePrivKey != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleRegistration, "already approved oracle registration. if you want to be shared oracle private key again, please register oracle again")
	}

	oracleRegistration.EncryptedOraclePrivKey = msg.ApprovalSharingOracleKey.EncryptedOraclePrivKey

	// add an encrypted oracle private key to oracleRegistration
	if err := k.SetOracleRegistration(ctx, oracleRegistration); err != nil {
		return sdkerrors.Wrapf(types.ErrApproveOracleRegistration, err.Error())
	}

	newOracle := types.NewOracle(
		msg.ApprovalSharingOracleKey.TargetOracleAddress,
		msg.ApprovalSharingOracleKey.UniqueId,
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
			sdk.NewAttribute(types.AttributeKeyOracleAddress, msg.ApprovalSharingOracleKey.TargetOracleAddress),
			sdk.NewAttribute(types.AttributeKeyUniqueID, msg.ApprovalSharingOracleKey.UniqueId),
		),
	)

	return nil

}

// validateApprovalSharingOracleKey validate unique ID of ApprovalSharingOracleKey and its signature
func (k Keeper) validateApprovalSharingOracleKey(ctx sdk.Context, approval *types.ApprovalSharingOracleKey, signature []byte) error {
	params := k.GetParams(ctx)

	// check unique id
	if approval.UniqueId != params.UniqueId {
		return types.ErrInvalidUniqueID
	}

	// check if the approver oracle exists
	existing, err := k.HasOracle(ctx, approval.ApproverOracleAddress)
	if err != nil || !existing {
		return fmt.Errorf("failed to check if the approver oracle exists or not. address(%s)", approval.ApproverOracleAddress)
	}

	// verify signature
	if err := k.VerifyOracleSignature(ctx, approval, signature); err != nil {
		return err
	}

	if approval.EncryptedOraclePrivKey == nil {
		return fmt.Errorf("encrypted oracle private key is empty")
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
		return nil, sdkerrors.Wrapf(types.ErrGetOracleRegistration, err.Error())
	}
	key := types.GetOracleRegistrationKey(uniqueID, accAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrOracleRegistrationNotFound
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
