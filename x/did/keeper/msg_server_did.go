package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func (m msgServer) CreateDID(goCtx context.Context, msg *types.MsgCreateDID) (*types.MsgCreateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	cur := keeper.GetDIDDocument(ctx, msg.Did)
	if !cur.Empty() {
		if cur.Deactivated() {
			return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
		}
		return nil, sdkerrors.Wrapf(types.ErrDIDExists, "DID: %s", msg.Did)
	}

	seq := types.InitialSequence
	_, err := VerifyDIDOwnership(msg.Document, seq, msg.Document, msg.Signature)
	if err != nil {
		return nil, err
	}

	docWithSeq := types.NewDIDDocumentWithSeq(msg.Document, seq)
	keeper.SetDIDDocument(ctx, msg.Did, docWithSeq)
	return &types.MsgCreateDIDResponse{}, nil
}

func (m msgServer) UpdateDID(goCtx context.Context, msg *types.MsgUpdateDID) (*types.MsgUpdateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	docWithSeq := keeper.GetDIDDocument(ctx, msg.Did)
	if docWithSeq.Empty() {
		return nil, sdkerrors.Wrapf(types.ErrDIDNotFound, "DID: %s", msg.Did)
	}
	if docWithSeq.Deactivated() {
		return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
	}

	newSeq, err := VerifyDIDOwnership(msg.Document, docWithSeq.Sequence, docWithSeq.Document, msg.Signature)
	if err != nil {
		return nil, err
	}

	newDocWithSeq := types.NewDIDDocumentWithSeq(msg.Document, newSeq)
	keeper.SetDIDDocument(ctx, msg.Did, newDocWithSeq)
	return &types.MsgUpdateDIDResponse{}, nil
}

func (m msgServer) DeactivateDID(goCtx context.Context, msg *types.MsgDeactivateDID) (*types.MsgDeactivateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	docWithSeq := keeper.GetDIDDocument(ctx, msg.Did)
	if docWithSeq.Empty() {
		return nil, sdkerrors.Wrapf(types.ErrDIDNotFound, "DID: %s", msg.Did)
	}
	if docWithSeq.Deactivated() {
		return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
	}

	doc := types.DIDDocument{}

	newSeq, err := VerifyDIDOwnership(&doc, docWithSeq.Sequence, docWithSeq.Document, msg.Signature)
	if err != nil {
		return nil, err
	}

	keeper.SetDIDDocument(ctx, msg.Did, docWithSeq.Deactivate(newSeq))
	return &types.MsgDeactivateDIDResponse{}, nil

}

func VerifyDIDOwnership(signData *types.DIDDocument, seq uint64, doc *types.DIDDocument, sig []byte) (uint64, error) {

	docBz := doc.Document
	document, err := ariesdid.ParseDocument(docBz)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "failed to parse did document")
	}

	// TODO: Currently, the public key is stored in the first verificationMethod and make signature using it.
	//		 We will improve this later by using the document's proof field.
	verificationMethod := document.VerificationMethod[0]

	// TODO: Currently, only ES256K1 is supported to verify DID ownership.
	//       It makes sense for now, since a DID is derived from a Secp256k1 public key.
	//       But, need to support other key types (according to verificationMethod.Type).
	if verificationMethod.Type != types.ES256K_2019 && verificationMethod.Type != types.ES256K_2018 {
		return 0, sdkerrors.Wrapf(types.ErrVerificationMethodKeyTypeNotImplemented, "VerificationMethod: %v", verificationMethod.Type)
	}

	var key secp256k1.PubKey
	key = make([]byte, secp256k1.PubKeySize)
	copy(key[:], document.VerificationMethod[0].Value)

	newSeq, ok := types.Verify(sig, signData, seq, key)
	if !ok {
		return 0, types.ErrSigVerificationFailed
	}
	return newSeq, nil
}
