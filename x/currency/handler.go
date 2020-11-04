package currency

import (
	"fmt"

	"github.com/medibloc/panacea-core/types/assets"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/currency/keeper"
	"github.com/medibloc/panacea-core/x/currency/types"
)

// NewHandler returns a handler for "currency" type messages
func NewHandler(keeper keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgIssueCurrency:
			return handleMsgIssueCurrency(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized currency Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgIssueCurrency(ctx sdk.Context, keeper keeper.Keeper, msg MsgIssueCurrency) sdk.Result {
	if msg.Amount.Denom == assets.MicroMedDenom {
		return types.ErrDenomNotAllowed(msg.Amount.Denom).Result()
	}

	cur := keeper.GetIssuance(ctx, msg.Amount.Denom)
	if !cur.Empty() {
		return types.ErrDenomExists(msg.Amount.Denom).Result()
	}

	keeper.SetIssuance(ctx, msg.Amount.Denom, msg)
	return sdk.Result{}
}
