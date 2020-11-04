package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/currency/types"
)

// Keeper maintains the link to data storage
// and exposes getter/setter methods for
// the various parts of the state machine
type Keeper interface {
	Codec() *codec.Codec
	SetIssuance(ctx sdk.Context, denom string, issuance types.Issuance)
	GetIssuance(ctx sdk.Context, denom string) types.Issuance
	ListIssuedDenoms(ctx sdk.Context) []string
}

// currencyKeeper implements the Keeper interface
type currencyKeeper struct {
	// Unexposed key to access store from sdk.Context
	storeKey sdk.StoreKey

	cdc *codec.Codec

	// External keepers
	bankKeeper types.BankKeeper
}

// NewKeeper creates a new instance of the currency Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, bankKeeper types.BankKeeper) Keeper {
	return currencyKeeper{
		storeKey:   storeKey,
		cdc:        cdc,
		bankKeeper: bankKeeper,
	}
}

func (k currencyKeeper) Codec() *codec.Codec {
	return k.cdc
}

func (k currencyKeeper) SetIssuance(ctx sdk.Context, denom string, issuance types.Issuance) {
	store := ctx.KVStore(k.storeKey)
	key := IssuanceKey(denom)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(issuance)
	store.Set(key, bz)

	if _, err := k.bankKeeper.AddCoins(ctx, issuance.IssuerAddress, sdk.NewCoins(issuance.Amount)); err != nil {
		panic(err)
	}
}

func (k currencyKeeper) GetIssuance(ctx sdk.Context, denom string) types.Issuance {
	store := ctx.KVStore(k.storeKey)
	key := IssuanceKey(denom)
	bz := store.Get(key)
	if bz == nil {
		return types.Issuance{}
	}

	var issuance types.Issuance
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &issuance)
	return issuance
}

func (k currencyKeeper) ListIssuedDenoms(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)
	denoms := make([]string, 0)

	prefix := IssuanceKey("")
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		denom := getLastElement(iter.Key(), prefix)
		denoms = append(denoms, string(denom))
	}
	return denoms
}
