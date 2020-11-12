package internal

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/token/types"
)

// mockKeeper implements the token.Keeper interface
type mockKeeper struct {
	tokens map[types.Symbol]types.Token
}

func NewMockKeeper() *mockKeeper {
	return &mockKeeper{tokens: make(map[types.Symbol]types.Token)}
}

func (k mockKeeper) Codec() *codec.Codec {
	return codec.New()
}

func (k mockKeeper) SetToken(_ sdk.Context, symbol types.Symbol, token types.Token) {
	k.tokens[symbol] = token
}

func (k mockKeeper) GetToken(_ sdk.Context, symbol types.Symbol) types.Token {
	return k.tokens[symbol]
}

func (k mockKeeper) ListTokens(sdk.Context) []types.Symbol {
	symbols := make([]types.Symbol, 0)
	for symbol := range k.tokens {
		symbols = append(symbols, symbol)
	}
	return symbols
}
