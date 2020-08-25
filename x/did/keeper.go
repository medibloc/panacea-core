package did

import (
	"bytes"

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

func (k Keeper) SetDIDDocument(ctx sdk.Context, did types.DID, doc types.DIDDocument) {
	store := ctx.KVStore(k.storeKey)
	key := DIDDocumentKey(did)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(doc)
	store.Set(key, bz)
}

func (k Keeper) HasDID(ctx sdk.Context, did types.DID) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(DIDDocumentKey(did))
}

func (k Keeper) GetDIDDocument(ctx sdk.Context, did types.DID) types.DIDDocument {
	store := ctx.KVStore(k.storeKey)
	key := DIDDocumentKey(did)
	bz := store.Get(key)
	if bz == nil {
		return types.DIDDocument{}
	}

	var doc types.DIDDocument
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &doc)
	return doc
}

func (k Keeper) ListDIDs(ctx sdk.Context) []types.DID {
	store := ctx.KVStore(k.storeKey)
	dids := make([]types.DID, 0)

	firstKey := DIDDocumentKey("")
	iter := sdk.KVStorePrefixIterator(store, firstKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		did := types.DID(bytes.Split(iter.Key(), firstKey)[1])
		dids = append(dids, did)
	}
	return dids
}

func (k Keeper) DeleteDID(ctx sdk.Context, did types.DID) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(DIDDocumentKey(did))
}
