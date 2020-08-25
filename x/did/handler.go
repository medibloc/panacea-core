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
		case MsgDeleteDID:
			return handleMsgDeleteDID(ctx, keeper, msg)
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

	keeper.SetDIDDocument(ctx, msg.DID, msg.Document)
	return sdk.Result{}
}

func handleMsgUpdateDID(ctx sdk.Context, keeper Keeper, msg MsgUpdateDID) sdk.Result {
	// TODO: prevent the double-spending: https://github.com/medibloc/panacea-core/issues/28
	err := verifyDIDOwnership(ctx, keeper, msg.DID, msg.SigKeyID, msg.Signature, msg.Document.GetSignBytes())
	if err != nil {
		return err.Result()
	}

	keeper.SetDIDDocument(ctx, msg.DID, msg.Document)
	return sdk.Result{}
}

func handleMsgDeleteDID(ctx sdk.Context, keeper Keeper, msg MsgDeleteDID) sdk.Result {
	// TODO: prevent the double-spending: https://github.com/medibloc/panacea-core/issues/28
	err := verifyDIDOwnership(ctx, keeper, msg.DID, msg.SigKeyID, msg.Signature, msg.DID.GetSignBytes())
	if err != nil {
		return err.Result()
	}

	keeper.DeleteDID(ctx, msg.DID)
	return sdk.Result{}
}

func verifyDIDOwnership(ctx sdk.Context, keeper Keeper, did types.DID, keyID types.KeyID, sig, data []byte) sdk.Error {
	doc := keeper.GetDID(ctx, did)
	if doc.Empty() {
		return types.ErrDIDNotFound(did)
	}

	pubKey, ok := doc.PubKeyByID(keyID)
	if !ok {
		return types.ErrKeyIDNotFound(keyID)
	}

	pubKeySecp256k1, err := types.NewPubKeyFromBase58(pubKey.KeyBase58)
	if err != nil {
		return types.ErrInvalidSecp256k1PublicKey(err)
	}

	if !pubKeySecp256k1.VerifyBytes(data, sig) {
		return types.ErrSigVerificationFailed()
	}

	return nil
}
