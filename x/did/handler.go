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
	cur := keeper.GetDIDDocument(ctx, msg.DID)
	if !cur.Empty() {
		return types.ErrDIDExists(msg.DID).Result()
	}

	seq := types.NewSequence()

	_, err := verifyDIDOwnership(msg.Document, seq, msg.Document, msg.SigKeyID, msg.Signature)
	if err != nil {
		return err.Result()
	}

	docWithSeq := types.NewDIDDocumentWithSeq(msg.Document, seq)
	keeper.SetDIDDocument(ctx, msg.DID, docWithSeq)
	return sdk.Result{}
}

func handleMsgUpdateDID(ctx sdk.Context, keeper Keeper, msg MsgUpdateDID) sdk.Result {
	docWithSeq := keeper.GetDIDDocument(ctx, msg.DID)
	if docWithSeq.Empty() {
		return types.ErrDIDNotFound(msg.DID).Result()
	}

	newSeq, err := verifyDIDOwnership(msg.Document, docWithSeq.Seq, docWithSeq.Document, msg.SigKeyID, msg.Signature)
	if err != nil {
		return err.Result()
	}

	newDocWithSeq := types.NewDIDDocumentWithSeq(msg.Document, newSeq)
	keeper.SetDIDDocument(ctx, msg.DID, newDocWithSeq)
	return sdk.Result{}
}

func handleMsgDeleteDID(ctx sdk.Context, keeper Keeper, msg MsgDeleteDID) sdk.Result {
	docWithSeq := keeper.GetDIDDocument(ctx, msg.DID)
	if docWithSeq.Empty() {
		return types.ErrDIDNotFound(msg.DID).Result()
	}

	_, err := verifyDIDOwnership(msg.DID, docWithSeq.Seq, docWithSeq.Document, msg.SigKeyID, msg.Signature)
	if err != nil {
		return err.Result()
	}

	keeper.DeleteDID(ctx, msg.DID)
	return sdk.Result{}
}

// verifyDIDOwnership verifies the DID ownership from a sig which is based on the data.
// It fetches a public key from a doc using keyID. It also uses a seq to verifyDIDOwnership the sig.
// If the verification is successful, it returns a new sequence. If not, it returns an error.
func verifyDIDOwnership(data types.Signable, seq types.Sequence, doc types.DIDDocument, keyID types.KeyID, sig []byte) (types.Sequence, sdk.Error) {
	pubKey, ok := doc.PubKeyByID(keyID)
	if !ok {
		return 0, types.ErrKeyIDNotFound(keyID)
	}
	pubKeySecp256k1, err := types.NewPubKeyFromBase58(pubKey.KeyBase58)
	if err != nil {
		return 0, types.ErrInvalidSecp256k1PublicKey(err)
	}
	newSeq, ok := types.Verify(sig, data, seq, pubKeySecp256k1)
	if !ok {
		return 0, types.ErrSigVerificationFailed()
	}
	return newSeq, nil
}
