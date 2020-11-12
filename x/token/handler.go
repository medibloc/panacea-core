package token

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/token/keeper"
	"github.com/medibloc/panacea-core/x/token/types"
)

// NewHandler returns a handler for "token" type messages
func NewHandler(keeper keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgIssueToken:
			return handleMsgIssueToken(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized token Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgIssueToken(ctx sdk.Context, keeper keeper.Keeper, msg MsgIssueToken) sdk.Result {
	txHash := fmt.Sprintf("%X", tmhash.Sum(ctx.TxBytes()))
	token := types.NewToken(msg, txHash)

	// Check existence
	if existed := keeper.GetToken(ctx, token.Symbol); !existed.Empty() {
		return types.ErrTokenExists(token.Symbol).Result()
	}

	// Store the token
	keeper.SetToken(ctx, token.Symbol, token)

	return sdk.Result{
		Data: ModuleCdc.MustMarshalJSON(token),
	}
}
