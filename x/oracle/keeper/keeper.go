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

func (k Keeper) VerifySignature(ctx sdk.Context, msg codec.ProtoMarshaler, sigBz []byte) error {
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
	_, err := sdk.AccAddressFromBech32(oracleAddress)
	if err != nil {
		return err
	}

	// TODO Check is oracle registered?

	// TODO Check is registered oracle's uniqueId correct?

	return nil
}
