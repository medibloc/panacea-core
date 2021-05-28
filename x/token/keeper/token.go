package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/token/types"
)

// SetToken set a specific token in the store
func (k Keeper) SetToken(ctx sdk.Context, token types.Token) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKey))
	b := k.cdc.MustMarshalBinaryBare(&token)
	store.Set(GetSymbolBytes(token.Symbol), b)

	// Update the total supply of all coins
	supply := k.bankKeeper.GetSupply(ctx)
	newTotal := supply.GetTotal().Add(token.TotalSupply)
	supply.SetTotal(newTotal)
	k.bankKeeper.SetSupply(ctx, supply)

	// Deposit the total supply of the new coin to the owner account
	ownerAddr, err := sdk.AccAddressFromBech32(token.OwnerAddress)
	if err != nil {
		panic(err)
	}
	if err := k.bankKeeper.AddCoins(ctx, ownerAddr, sdk.NewCoins(token.TotalSupply)); err != nil {
		panic(err)
	}
}

// GetToken returns a token from its id
func (k Keeper) GetToken(ctx sdk.Context, symbol string) types.Token {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKey))
	var token types.Token
	k.cdc.MustUnmarshalBinaryBare(store.Get(GetSymbolBytes(symbol)), &token)
	return token
}

// HasToken checks if the token exists in the store
func (k Keeper) HasToken(ctx sdk.Context, symbol string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKey))
	return store.Has(GetSymbolBytes(symbol))
}

// GetAllToken returns all token
func (k Keeper) GetAllToken(ctx sdk.Context) (list []types.Token) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Token
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetSymbolBytes returns the byte representation of the symbol
func GetSymbolBytes(symbol string) []byte {
	return []byte(symbol)
}
