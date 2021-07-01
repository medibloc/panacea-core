package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/medibloc/panacea-core/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/x/did/types"
)

func (m msgServer) CreateDID(goCtx context.Context, msg *types.MsgCreateDID) (*types.MsgCreateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	cur := keeper.GetDIDDocument(ctx, msg.DID)
	if !cur.Empty() {
		if cur.Deactivated() {
			return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.DID)
		}
		return nil, sdkerrors.Wrapf(types.ErrDIDExists, "DID: %s", msg.DID)
	}

	seq := types.InitialSequence
	_, err := VerifyDIDOwnership(msg.Document, seq, msg.Document, msg.VerificationMethodID, msg.Signature)
	if err != nil {
		return nil, err
	}

	docWithSeq := types.NewDIDDocumentWithSeq(msg.Document, uint64(seq))
	keeper.SetDIDDocument(ctx, msg.DID, docWithSeq)
	return &types.MsgCreateDIDResponse{}, nil
}

func (m msgServer) UpdateDID(goCtx context.Context, msg *types.MsgUpdateDID) (*types.MsgUpdateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	docWithSeq := keeper.GetDIDDocument(ctx, msg.DID)
	if docWithSeq.Empty() {
		return nil, sdkerrors.Wrapf(types.ErrDIDNotFound, "DID: %s", msg.DID)
	}
	if docWithSeq.Deactivated() {
		return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.DID)
	}

	newSeq, err := VerifyDIDOwnership(msg.Document, docWithSeq.Seq, docWithSeq.Document, msg.VerificationMethodID, msg.Signature)
	if err != nil {
		return nil, err
	}

	newDocWithSeq := types.NewDIDDocumentWithSeq(msg.Document, newSeq)
	keeper.SetDIDDocument(ctx, msg.DID, newDocWithSeq)
	return &types.MsgUpdateDIDResponse{}, nil
}

func (m msgServer) DeactivateDID(goCtx context.Context, msg *types.MsgDeactivateDID) (*types.MsgDeactivateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	docWithSeq := keeper.GetDIDDocument(ctx, msg.DID)
	if docWithSeq.Empty() {
		return nil, sdkerrors.Wrapf(types.ErrDIDNotFound, "DID: %s", msg.DID)
	}
	if docWithSeq.Deactivated() {
		return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.DID)
	}

	doc := types.DIDDocument{
		ID: msg.DID,
	}

	newSeq, err := VerifyDIDOwnership(&doc, docWithSeq.Seq, docWithSeq.Document, msg.VerificationMethodID, msg.Signature)
	if err != nil {
		return nil, err
	}

	keeper.SetDIDDocument(ctx, msg.DID, docWithSeq.Deactivate(newSeq))
	return &types.MsgDeactivateDIDResponse{}, nil

}

func VerifyDIDOwnership(signData *types.DIDDocument, seq uint64, doc *types.DIDDocument, verificationMethodID string, sig []byte) (uint64, error) {
	verificationMethod, ok := doc.VerificationMethodFrom(doc.Authentications, verificationMethodID)
	if !ok {
		return 0, sdkerrors.Wrapf(types.ErrVerificationMethodIDNotFound, "VerificationMethodId: %s", verificationMethodID)
	}

	// TODO: Currently, only ES256K1 is supported to verify DID ownership.
	//       It makes sense for now, since a DID is derived from a Secp256k1 public key.
	//       But, need to support other key types (according to verificationMethod.Type).
	if verificationMethod.Type != types.ES256K_2019 && verificationMethod.Type != types.ES256K_2018 {
		return 0, sdkerrors.Wrapf(types.ErrVerificationMethodKeyTypeNotImplemented, "VerificationMethod: %v", verificationMethod.Type)
	}
	pubKeySecp256k1, err := secp256k1util.PubKeyFromBase58(verificationMethod.PubKeyBase58)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrInvalidSecp256k1PublicKey, "PublicKey: %v", verificationMethod.PubKeyBase58)
	}
	newSeq, ok := types.Verify(sig, signData, seq, pubKeySecp256k1)
	if !ok {
		return 0, types.ErrSigVerificationFailed
	}
	return newSeq, nil
}