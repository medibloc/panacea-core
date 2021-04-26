package did

import (
	"fmt"

	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"

	"github.com/medibloc/panacea-core/x/did/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/x/did/types"
)

// NewHandler returns a handler for "did" type messages
func NewHandler(keeper keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreateDID:
			return handleMsgCreateDID(ctx, keeper, msg)
		case MsgUpdateDID:
			return handleMsgUpdateDID(ctx, keeper, msg)
		case MsgDeactivateDID:
			return handleMsgDeactivateDID(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized did Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateDID(ctx sdk.Context, keeper keeper.Keeper, msg MsgCreateDID) sdk.Result {
	cur := keeper.GetDIDDocument(ctx, msg.DID)
	if !cur.Empty() {
		if cur.Deactivated() {
			return types.ErrDIDDeactivated(msg.DID).Result()
		}
		return types.ErrDIDExists(msg.DID).Result()
	}

	seq := types.InitialSequence

	_, err := verifyDIDOwnership(msg.Document, seq, msg.Document, msg.VerificationMethodID, msg.Signature)
	if err != nil {
		return err.Result()
	}

	docWithSeq := types.NewDIDDocumentWithSeq(msg.Document, seq)
	keeper.SetDIDDocument(ctx, msg.DID, docWithSeq)
	return sdk.Result{}
}

func handleMsgUpdateDID(ctx sdk.Context, keeper keeper.Keeper, msg MsgUpdateDID) sdk.Result {
	docWithSeq := keeper.GetDIDDocument(ctx, msg.DID)
	if docWithSeq.Empty() {
		return types.ErrDIDNotFound(msg.DID).Result()
	}
	if docWithSeq.Deactivated() {
		return types.ErrDIDDeactivated(msg.DID).Result()
	}

	newSeq, err := verifyDIDOwnership(msg.Document, docWithSeq.Seq, docWithSeq.Document, msg.VerificationMethodID, msg.Signature)
	if err != nil {
		return err.Result()
	}

	newDocWithSeq := types.NewDIDDocumentWithSeq(msg.Document, newSeq)
	keeper.SetDIDDocument(ctx, msg.DID, newDocWithSeq)
	return sdk.Result{}
}

func handleMsgDeactivateDID(ctx sdk.Context, keeper keeper.Keeper, msg MsgDeactivateDID) sdk.Result {
	docWithSeq := keeper.GetDIDDocument(ctx, msg.DID)
	if docWithSeq.Empty() {
		return types.ErrDIDNotFound(msg.DID).Result()
	}
	if docWithSeq.Deactivated() {
		return types.ErrDIDDeactivated(msg.DID).Result()
	}

	newSeq, err := verifyDIDOwnership(msg.DID, docWithSeq.Seq, docWithSeq.Document, msg.VerificationMethodID, msg.Signature)
	if err != nil {
		return err.Result()
	}

	// put a tombstone instead of deletion
	keeper.SetDIDDocument(ctx, msg.DID, docWithSeq.Deactivate(newSeq))
	return sdk.Result{}
}

// verifyDIDOwnership verifies the DID ownership from a sig which is covers a DID Document + a sequence number.
// A public key is taken from one of verificationMethods in the DID Document.
// If the verification is successful, it returns a new sequence. If not, it returns an error.
func verifyDIDOwnership(data types.Signable, seq types.Sequence, doc types.DIDDocument, verificationMethodID types.VerificationMethodID, sig []byte) (types.Sequence, sdk.Error) {
	verificationMethod, ok := doc.VerificationMethodByID(verificationMethodID)
	if !ok {
		return 0, types.ErrVerificationMethodIDNotFound(verificationMethodID)
	}

	// TODO: Currently, only ES256K1 is supported to verify DID ownership.
	//       It makes sense for now, since a DID is derived from a Secp256k1 public key.
	//       But, need to support other key types (according to verificationMethod.Type).
	if verificationMethod.Type != types.ES256K_2019 && verificationMethod.Type != types.ES256K_2018 {
		return 0, types.ErrVerificationMethodKeyTypeNotImplemented(verificationMethod.Type)
	}
	pubKeySecp256k1, err := secp256k1util.PubKeyFromBase58(verificationMethod.PubKeyBase58)
	if err != nil {
		return 0, types.ErrInvalidSecp256k1PublicKey(err)
	}
	newSeq, ok := types.Verify(sig, data, seq, pubKeySecp256k1)
	if !ok {
		return 0, types.ErrSigVerificationFailed()
	}
	return newSeq, nil
}
