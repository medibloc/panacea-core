package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc      codec.Codec
		storeKey sdk.StoreKey
		memKey   sdk.StoreKey

		paramSpace paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey,
	memKey sdk.StoreKey,
	paramSpace paramtypes.Subspace,
) *Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramSpace: paramSpace,
	}
}

func (k Keeper) VerifyOracleSignature(ctx sdk.Context, msg codec.ProtoMarshaler, sigBz []byte) error {
	bz, err := k.cdc.Marshal(msg)
	if err != nil {
		return err
	}

	oraclePubKeyBz := k.GetParams(ctx).MustDecodeOraclePubKey()
	oraclePubKey, err := btcec.ParsePubKey(oraclePubKeyBz, btcec.S256())
	if err != nil {
		return err
	}

	signature, err := btcec.ParseSignature(sigBz, btcec.S256())
	if err != nil {
		return err
	}

	if !signature.Verify(bz, oraclePubKey) {
		return fmt.Errorf("failed to signature validation")
	}

	return nil
}

func (k Keeper) VerifyOracle(ctx sdk.Context, oracleAddress string) error {
	oracle, err := k.GetOracle(ctx, oracleAddress)
	if err != nil {
		return fmt.Errorf("failed to oracle validation. address(%s) %w", oracleAddress, err)
	}

	activeUniqueID := k.GetParams(ctx).UniqueId
	if activeUniqueID != oracle.UniqueId {
		return fmt.Errorf("is not active an oracle. oracleAddress(%s), oracleUniqueID(%s), activeUniqueID(%s)",
			oracle.OracleAddress,
			oracle.UniqueId,
			activeUniqueID,
		)
	}

	return nil
}
