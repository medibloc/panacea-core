package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/token/types"
)

// Keeper maintains the link to data storage
// and exposes getter/setter methods for
// the various parts of the state machine
type Keeper interface {
	Codec() *codec.Codec
	SetToken(ctx sdk.Context, symbol types.Symbol, token types.Token)
	GetToken(ctx sdk.Context, symbol types.Symbol) types.Token
	ListTokens(ctx sdk.Context) []types.Symbol
}

// tokenKeeper implements the Keeper interface
type tokenKeeper struct {
	// Unexposed key to access store from sdk.Context
	storeKey sdk.StoreKey

	cdc *codec.Codec

	// External keepers
	bankKeeper types.BankKeeper
}

// NewKeeper creates a new instance of the token Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, bankKeeper types.BankKeeper) Keeper {
	return tokenKeeper{
		storeKey:   storeKey,
		cdc:        cdc,
		bankKeeper: bankKeeper,
	}
}

func (k tokenKeeper) Codec() *codec.Codec {
	return k.cdc
}

func (k tokenKeeper) SetToken(ctx sdk.Context, symbol types.Symbol, token types.Token) {
	store := ctx.KVStore(k.storeKey)
	key := TokenKey(symbol)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(token)
	store.Set(key, bz)

	// Deposit the total supply to the owner account
	if _, err := k.bankKeeper.AddCoins(ctx, token.OwnerAddress, sdk.NewCoins(token.TotalSupply)); err != nil {
		panic(err)
	}
}

func (k tokenKeeper) GetToken(ctx sdk.Context, symbol types.Symbol) types.Token {
	store := ctx.KVStore(k.storeKey)
	key := TokenKey(symbol)
	bz := store.Get(key)
	if bz == nil {
		return types.Token{}
	}

	var token types.Token
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &token)
	return token
}

func (k tokenKeeper) ListTokens(ctx sdk.Context) []types.Symbol {
	store := ctx.KVStore(k.storeKey)
	symbols := make([]types.Symbol, 0)

	prefix := TokenKey("")
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		symbol := getLastElement(iter.Key(), prefix)
		symbols = append(symbols, types.Symbol(symbol))
	}
	return symbols
}
