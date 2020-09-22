package keeper

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
)

// Keeper maintains the link to data storage
// and exposes getter/setter methods for
// the various parts of the state machine
type Keeper interface {
	Codec() *codec.Codec
	SetDIDDocument(ctx sdk.Context, did types.DID, doc types.DIDDocumentWithSeq)
	GetDIDDocument(ctx sdk.Context, did types.DID) types.DIDDocumentWithSeq
	ListDIDs(ctx sdk.Context) []types.DID
}

// didKeeper implements the Keeper interface
type didKeeper struct {
	// Unexposed key to access store from sdk.Context
	storeKey sdk.StoreKey

	cdc *codec.Codec
}

// NewKeeper creates a new instance of the did Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return didKeeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

func (k didKeeper) Codec() *codec.Codec {
	return k.cdc
}

func (k didKeeper) SetDIDDocument(ctx sdk.Context, did types.DID, doc types.DIDDocumentWithSeq) {
	store := ctx.KVStore(k.storeKey)
	key := DIDDocumentKey(did)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(doc)
	store.Set(key, bz)
}

func (k didKeeper) GetDIDDocument(ctx sdk.Context, did types.DID) types.DIDDocumentWithSeq {
	store := ctx.KVStore(k.storeKey)
	key := DIDDocumentKey(did)
	bz := store.Get(key)
	if bz == nil {
		return types.DIDDocumentWithSeq{}
	}

	var doc types.DIDDocumentWithSeq
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &doc)
	return doc
}

func (k didKeeper) ListDIDs(ctx sdk.Context) []types.DID {
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
