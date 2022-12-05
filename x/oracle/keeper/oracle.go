package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func (k Keeper) HasOracle(ctx sdk.Context, address string) (bool, error) {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return false, err
	}

	return store.Has(types.GetOracleKey(accAddr)), nil
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

func (k Keeper) GetOracleRegistration(ctx sdk.Context, uniqueID, address string) (*types.OracleRegistration, error) {
	store := ctx.KVStore(k.storeKey)
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}
	key := types.GetOracleRegistrationKey(uniqueID, accAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrOracleRegistrationNotFound
	}

	oracleRegistration := &types.OracleRegistration{}

	err = k.cdc.UnmarshalLengthPrefixed(bz, oracleRegistration)
	if err != nil {
		return nil, err
	}

	return oracleRegistration, nil
}

func (k Keeper) ApproveOracleRegistration(ctx sdk.Context, msg *types.MsgApproveOracleRegistration) error {

	if err := k.validateApproveOracleRegistration(ctx, msg); err != nil {
		return fmt.Errorf("invalid oracle private key signature")
	}

	oracleRegistration, err := k.GetOracleRegistration(ctx, msg.ApproveOracleRegistration.UniqueId, msg.ApproveOracleRegistration.TargetOracleAddress)
	if err != nil {
		return err
	}

	newOracle := &types.Oracle{
		OracleAddress:        msg.ApproveOracleRegistration.TargetOracleAddress,
		UniqueId:             msg.ApproveOracleRegistration.UniqueId,
		Endpoint:             oracleRegistration.Endpoint,
		OracleCommissionRate: oracleRegistration.OracleCommissionRate,
	}

	if err := k.SetOracle(ctx, newOracle); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeApproveOracleRegistration,
			sdk.NewAttribute(types.AttributeKeyOracleAddress, msg.ApproveOracleRegistration.TargetOracleAddress),
			sdk.NewAttribute(types.AttributeKeyEncryptedOraclePrivKey, string(msg.ApproveOracleRegistration.EncryptedOraclePrivKey)),
		),
	)

	return nil

}

// validateApproveOracleRegistration checks signature
func (k Keeper) validateApproveOracleRegistration(ctx sdk.Context, msg *types.MsgApproveOracleRegistration) interface{} {

	params := k.GetParams(ctx)
	targetOracleAddress := msg.ApproveOracleRegistration.TargetOracleAddress

	if msg.ApproveOracleRegistration.UniqueId != params.UniqueId {
		return types.ErrInvalidUniqueId
	}

	ApproveOracleRegistrationBz := k.cdc.MustMarshal(msg.ApproveOracleRegistration)

	oraclePubKeyBz := k.GetParams(ctx).MustDecodeOraclePublicKey()
	if !secp256k1.PubKey(oraclePubKeyBz).VerifySignature(ApproveOracleRegistrationBz, msg.Signature) {
		return fmt.Errorf("failed to signature validation")
	}

	hasOracle, err := k.HasOracle(ctx, targetOracleAddress)
	if err != nil {
		return err
	}
	if hasOracle {
		return fmt.Errorf("oracle address(%v) is already registered", targetOracleAddress)
	}

	return nil
}
