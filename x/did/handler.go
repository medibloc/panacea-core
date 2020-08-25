package did

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
)

// NewHandler returns a handler for "did" type messages
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreateDID:
			return handleMsgCreateDID(ctx, keeper, msg)
		case MsgUpdateDID:
			return handleMsgUpdateDID(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized did Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateDID(ctx sdk.Context, keeper Keeper, msg MsgCreateDID) sdk.Result {
	if keeper.HasDID(ctx, msg.DID) {
		return types.ErrDIDExists(msg.DID).Result()
	}

	keeper.SetDID(ctx, msg.DID, msg.Document)
	return sdk.Result{}
}

func handleMsgUpdateDID(ctx sdk.Context, keeper Keeper, msg MsgUpdateDID) sdk.Result {
	curDoc := keeper.GetDID(ctx, msg.DID)
	if curDoc.Empty() {
		return types.ErrDIDNotFound(msg.DID).Result()
	}

	pubKey, ok := curDoc.PubKeyByID(msg.SigKeyID)
	if !ok {
		return types.ErrKeyIDNotFound(msg.SigKeyID).Result()
	}

	pubKeySecp256k1, err := types.NewPubKeyFromBase58(pubKey.KeyBase58)
	if err != nil {
		return types.ErrInvalidSecp256k1PublicKey(err).Result()
	}
	if !pubKeySecp256k1.VerifyBytes(msg.Document.GetSignBytes(), msg.Signature) {
		return types.ErrSigVerificationFailed().Result()
	}

	keeper.SetDID(ctx, msg.DID, msg.Document)
	return sdk.Result{}
}
