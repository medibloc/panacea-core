package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
)

func (k Keeper) SetDIDDocument(ctx sdk.Context, did string, doc types.DIDDocumentWithSeq) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DIDKeyPrefix)
	key := []byte(did)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(&doc)
	store.Set(key, bz)
}

func (k Keeper) GetDIDDocument(ctx sdk.Context, did string) types.DIDDocumentWithSeq {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DIDKeyPrefix)
	key := []byte(did)
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

	iter := sdk.KVStorePrefixIterator(store, []byte{})
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		did := string(iter.Key())
		dids = append(dids, did)
	}
	return dids
}
