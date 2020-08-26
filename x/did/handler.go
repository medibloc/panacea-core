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

	docWithSeq := types.NewDIDDocumentWithSeq(msg.Document, types.NewSequence())
	keeper.SetDIDDocument(ctx, msg.DID, docWithSeq)
	return sdk.Result{}
}

func handleMsgUpdateDID(ctx sdk.Context, keeper Keeper, msg MsgUpdateDID) sdk.Result {
	newSeq, err := verifyDIDOwnership(ctx, keeper, msg.DID, msg.SigKeyID, msg.Signature, msg.Document)
	if err != nil {
		return err.Result()
	}

	newDocWithSeq := types.NewDIDDocumentWithSeq(msg.Document, newSeq)
	keeper.SetDIDDocument(ctx, msg.DID, newDocWithSeq)
	return sdk.Result{}
}

func handleMsgDeleteDID(ctx sdk.Context, keeper Keeper, msg MsgDeleteDID) sdk.Result {
	_, err := verifyDIDOwnership(ctx, keeper, msg.DID, msg.SigKeyID, msg.Signature, msg.DID)
	if err != nil {
		return err.Result()
	}

	keeper.DeleteDID(ctx, msg.DID)
	return sdk.Result{}
}

func verifyDIDOwnership(ctx sdk.Context, keeper Keeper, did types.DID, keyID types.KeyID, sig []byte, data types.Signable) (types.Sequence, sdk.Error) {
	docWithSeq := keeper.GetDIDDocument(ctx, did)
	if docWithSeq.Empty() {
		return 0, types.ErrDIDNotFound(did)
	}

	pubKey, ok := docWithSeq.Document.PubKeyByID(keyID)
	if !ok {
		return 0, types.ErrKeyIDNotFound(keyID)
	}

	pubKeySecp256k1, err := types.NewPubKeyFromBase58(pubKey.KeyBase58)
	if err != nil {
		return 0, types.ErrInvalidSecp256k1PublicKey(err)
	}

	newSeq, ok := types.Verify(sig, data, docWithSeq.Seq, pubKeySecp256k1)
	if !ok {
		return 0, types.ErrSigVerificationFailed()
	}
	return newSeq, nil
}
