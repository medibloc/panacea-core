package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
)

func (k Keeper) SetDIDDocument(ctx sdk.Context, did string, doc types.DIDDocumentWithSeq) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DIDKeyPrefix)
	key := DIDDocumentKey(did)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&doc)
	store.Set(key, bz)
}

func (k Keeper) GetDIDDocument(ctx sdk.Context, did string) types.DIDDocumentWithSeq {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DIDKeyPrefix)
	key := DIDDocumentKey(did)
	bz := store.Get(key)
	if bz == nil {
		return types.DIDDocumentWithSeq{}
	}

	var doc types.DIDDocumentWithSeq
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &doc)
	return doc
}

func (k Keeper) ListDIDs(ctx sdk.Context) []string {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DIDKeyPrefix)
	dids := make([]string, 0)

	prefix := DIDDocumentKey("")
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		did := getLastElement(iter.Key(), prefix)
		dids = append(dids, string(did))
	}
	return dids
}
