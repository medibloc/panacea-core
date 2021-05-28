package keeper

import (
	"context"
	"fmt"

	"github.com/tendermint/tendermint/crypto/tmhash"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/x/token/types"
)

func (k msgServer) IssueToken(goCtx context.Context, msg *types.MsgIssueToken) (*types.MsgIssueTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	txHash := fmt.Sprintf("%X", tmhash.Sum(ctx.TxBytes()))
	token := NewTokenFromMsg(msg, txHash)

	if k.HasToken(ctx, token.Symbol) {
		return nil, sdkerrors.Wrap(types.ErrTokenExists, fmt.Sprintf("symbol: %s", token.Symbol))
	}

	k.SetToken(ctx, token)
	return &types.MsgIssueTokenResponse{}, nil
}

func NewTokenFromMsg(msg *types.MsgIssueToken, txHash string) types.Token {
	symbol := types.NewSymbol(msg.ShortSymbol, txHash)
	return types.Token{
		Name:         msg.Name,
		Symbol:       symbol,
		TotalSupply:  sdk.NewCoin(types.GetMicroDenom(symbol), msg.TotalSupplyMicro.Int),
		Mintable:     msg.Mintable,
		OwnerAddress: msg.OwnerAddress,
	}
}
