package did

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
)

// Keeper maintains the link to data storage
// and exposes getter/setter methods for
// the various parts of the state machine
type Keeper struct {
	// Unexposed key to access store from sdk.Context
	storeKey sdk.StoreKey

	cdc *codec.Codec
}

// NewKeeper creates a new instance of the did Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

func (k Keeper) SetDID(ctx sdk.Context, did types.DID, doc types.DIDDocument) {
	store := ctx.KVStore(k.storeKey)
	key := DIDKey(did)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(doc)
	store.Set(key, bz)
}

func (k Keeper) HasDID(ctx sdk.Context, did types.DID) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(DIDKey(did))
}

func (k Keeper) GetDID(ctx sdk.Context, did types.DID) types.DIDDocument {
	store := ctx.KVStore(k.storeKey)
	key := DIDKey(did)
	bz := store.Get(key)
	if bz == nil {
		return types.DIDDocument{}
	}

	var doc types.DIDDocument
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &doc)
	return doc
}
