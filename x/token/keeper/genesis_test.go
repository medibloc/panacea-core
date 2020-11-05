package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/token/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	ctx := sdk.Context{}

	// prepare a keeper with some data
	keeper := newMockKeeper()
	token1 := types.Token{
		Name:         "my token 1",
		Symbol:       "KAI-0EA",
		TotalSupply:  sdk.NewCoin("ukai0ea", sdk.NewInt(1000000000)),
		Mintable:     true,
		OwnerAddress: sdk.AccAddress("panacea126r28pr7sstg7yfmedv3qq4st4a4exlwccx2vc"),
	}
	keeper.SetToken(ctx, token1.Symbol, token1)
	token2 := types.Token{
		Name:         "my token 2",
		Symbol:       "LOV-6AC",
		TotalSupply:  sdk.NewCoin("ulov6ac", sdk.NewInt(1000000000)),
		Mintable:     true,
		OwnerAddress: sdk.AccAddress("panacea126r28pr7sstg7yfmedv3qq4st4a4exlwccx2vc"),
	}
	keeper.SetToken(ctx, token2.Symbol, token2)

	// export a genesis
	state := ExportGenesis(ctx, keeper)
	require.Equal(t, 2, len(state.Tokens))
	require.Equal(t, token1, state.Tokens[newGenesisTokenKey(token1.Symbol)])
	require.Equal(t, token2, state.Tokens[newGenesisTokenKey(token2.Symbol)])

	// check if the exported genesis is valid
	require.NoError(t, types.ValidateGenesis(state))

	// import it to a new keeper
	newK := newMockKeeper()
	InitGenesis(ctx, newK, state)
	require.Equal(t, 2, len(newK.ListTokens(ctx)))
	require.Equal(t, token1, newK.GetToken(ctx, token1.Symbol))
	require.Equal(t, token2, newK.GetToken(ctx, token2.Symbol))
}

func newGenesisTokenKey(symbol types.Symbol) string {
	return types.GenesisTokenKey{Symbol: symbol}.Marshal()
}

// mockKeeper implements the token.Keeper interface
type mockKeeper struct {
	tokens map[types.Symbol]types.Token
}

func newMockKeeper() *mockKeeper {
	return &mockKeeper{tokens: make(map[types.Symbol]types.Token)}
}

func (k mockKeeper) Codec() *codec.Codec {
	return codec.New()
}

func (k mockKeeper) SetToken(ctx sdk.Context, symbol types.Symbol, token types.Token) {
	k.tokens[symbol] = token
}

func (k mockKeeper) GetToken(ctx sdk.Context, symbol types.Symbol) types.Token {
	return k.tokens[symbol]
}

func (k mockKeeper) ListTokens(ctx sdk.Context) []types.Symbol {
	symbols := make([]types.Symbol, 0)
	for symbol := range k.tokens {
		symbols = append(symbols, symbol)
	}
	return symbols
}
