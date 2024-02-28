package keeper

import (
	"context"
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

func (m msgServer) CreateDID(goCtx context.Context, msg *types.MsgServiceCreateDIDRequest) (*types.MsgServiceCreateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	cur := keeper.GetDIDDocument(ctx, msg.Did)
	if !cur.Empty() {
		if cur.Deactivated() {
			return nil, errors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
		}
		return nil, errors.Wrapf(types.ErrDIDExists, "DID: %s", msg.Did)
	}

	seq := types.InitialSequence
	_, err := VerifyDIDOwnership(msg.Document, seq, msg.Document, msg.VerificationMethodId, msg.Signature)
	if err != nil {
		return nil, err
	}

	docWithSeq := types.NewDIDDocumentWithSeq(msg.Document, uint64(seq))
	keeper.SetDIDDocument(ctx, msg.Did, docWithSeq)
	return &types.MsgServiceCreateDIDResponse{}, nil
}

func (m msgServer) UpdateDID(goCtx context.Context, msg *types.MsgServiceUpdateDIDRequest) (*types.MsgServiceUpdateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	docWithSeq := keeper.GetDIDDocument(ctx, msg.Did)
	if docWithSeq.Empty() {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "DID: %s", msg.Did)
	}
	if docWithSeq.Deactivated() {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
	}

	newSeq, err := VerifyDIDOwnership(msg.Document, docWithSeq.Sequence, docWithSeq.Document, msg.VerificationMethodId, msg.Signature)
	if err != nil {
		return nil, err
	}

	newDocWithSeq := types.NewDIDDocumentWithSeq(msg.Document, newSeq)
	keeper.SetDIDDocument(ctx, msg.Did, newDocWithSeq)
	return &types.MsgServiceUpdateDIDResponse{}, nil
}

func (m msgServer) DeactivateDID(goCtx context.Context, msg *types.MsgServiceDeactivateDIDRequest) (*types.MsgServiceDeactivateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	docWithSeq := keeper.GetDIDDocument(ctx, msg.Did)
	if docWithSeq.Empty() {
		return nil, errors.Wrapf(types.ErrDIDNotFound, "DID: %s", msg.Did)
	}
	if docWithSeq.Deactivated() {
		return nil, errors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
	}

	doc := types.DIDDocument{
		Id: msg.Did,
	}

	newSeq, err := VerifyDIDOwnership(&doc, docWithSeq.Sequence, docWithSeq.Document, msg.VerificationMethodId, msg.Signature)
	if err != nil {
		return nil, err
	}

	keeper.SetDIDDocument(ctx, msg.Did, docWithSeq.Deactivate(newSeq))
	return &types.MsgServiceDeactivateDIDResponse{}, nil

}

func VerifyDIDOwnership(signData *types.DIDDocument, seq uint64, doc *types.DIDDocument, verificationMethodID string, sig []byte) (uint64, error) {
	verificationMethod, ok := doc.VerificationMethodFrom(doc.Authentications, verificationMethodID)
	if !ok {
		return 0, errors.Wrapf(types.ErrVerificationMethodIDNotFound, "VerificationMethodId: %s", verificationMethodID)
	}

	// TODO: Currently, only ES256K1 is supported to verify DID ownership.
	//       It makes sense for now, since a DID is derived from a Secp256k1 public key.
	//       But, need to support other key types (according to verificationMethod.Type).
	if verificationMethod.Type != types.ES256K_2019 && verificationMethod.Type != types.ES256K_2018 {
		return 0, errors.Wrapf(types.ErrVerificationMethodKeyTypeNotImplemented, "VerificationMethod: %v", verificationMethod.Type)
	}
	pubKeySecp256k1, err := secp256k1util.PubKeyFromBase58(verificationMethod.PublicKeyBase58)
	if err != nil {
		return 0, errors.Wrapf(types.ErrInvalidSecp256k1PublicKey, "PublicKey: %v", verificationMethod.PublicKeyBase58)
	}
	newSeq, ok := types.Verify(sig, signData, seq, pubKeySecp256k1)
	if !ok {
		return 0, types.ErrSigVerificationFailed
	}
	return newSeq, nil
}
