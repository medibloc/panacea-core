package token

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/medibloc/panacea-core/x/token/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	txBytes, _   = hex.DecodeString("7B226E616D65223A226D7920616C6C222C2273796D626F6C223A224B41492D304541222C22746F74616C5")
	txHashPrefix = "BAF"
)

func TestHandleMsgIssueToken(t *testing.T) {
	ctx := sdk.Context{}.WithTxBytes(txBytes)
	keeper := newMockKeeper()
	msg := types.NewMsgIssueToken("my token", "LOV", sdk.NewInt(1000000000), true, sdk.AccAddress{})

	res := handleMsgIssueToken(ctx, keeper, msg)
	require.True(t, res.IsOK())

	symbol := types.Symbol(fmt.Sprintf("LOV-%s", txHashPrefix))

	var token types.Token
	require.NoError(t, ModuleCdc.UnmarshalJSON(res.Data, &token))
	require.Equal(t, msg.Name, token.Name)
	require.Equal(t, symbol, token.Symbol)
	require.Equal(t, sdk.NewCoin(symbol.MicroDenom(), msg.TotalSupply), token.TotalSupply)
	require.True(t, token.Mintable)
	require.Equal(t, sdk.AccAddress{}, token.OwnerAddress)

	require.False(t, keeper.GetToken(ctx, symbol).Empty())
}

func TestHandleMsgIssueToken_Exists(t *testing.T) {
	ctx := sdk.Context{}.WithTxBytes(txBytes)
	keeper := newMockKeeper()
	msg := types.NewMsgIssueToken("my token", "LOV", sdk.NewInt(1000000000), true, sdk.AccAddress{})

	res := handleMsgIssueToken(ctx, keeper, msg)
	t.Log(res)
	require.True(t, res.IsOK())

	res = handleMsgIssueToken(ctx, keeper, msg)
	require.False(t, res.IsOK())
	require.Equal(t, types.DefaultCodespace, res.Codespace)
	require.Equal(t, types.CodeTokenExists, res.Code)
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
