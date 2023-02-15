package keeper

import (
	"bytes"
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/medibloc/panacea-core/v2/x/did/types"
)

func (m msgServer) CreateDID(goCtx context.Context, msg *types.MsgCreateDID) (*types.MsgCreateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	cur := keeper.GetDIDDocument(ctx, msg.Did)
	if !cur.Empty() {
		if cur.Deactivated {
			return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
		}
		return nil, sdkerrors.Wrapf(types.ErrDIDExists, "DID: %s", msg.Did)
	}

	if err := types.ValidateDIDDocument(msg.Did, msg.Document); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "error: %v", err)
	}

	keeper.SetDIDDocument(ctx, msg.Did, msg.Document)
	return &types.MsgCreateDIDResponse{Did: msg.Did}, nil
}

func (m msgServer) UpdateDID(goCtx context.Context, msg *types.MsgUpdateDID) (*types.MsgUpdateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := types.ValidateDIDDocument(msg.Did, msg.Document); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "error: %v", err)
	}

	prevDocument := keeper.GetDIDDocument(ctx, msg.Did)
	if prevDocument.Empty() {
		return nil, sdkerrors.Wrapf(types.ErrDIDNotFound, "DID: %s", msg.Did)
	}
	if prevDocument.Deactivated {
		return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
	}

	if err := VerifyDIDOwnership(msg.Document.Document, prevDocument.Document); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrVerifyOwnershipFailed, "error: %v", err)
	}

	keeper.SetDIDDocument(ctx, msg.Did, msg.Document)

	return &types.MsgUpdateDIDResponse{}, nil
}

func (m msgServer) DeactivateDID(goCtx context.Context, msg *types.MsgDeactivateDID) (*types.MsgDeactivateDIDResponse, error) {
	keeper := m.Keeper
	ctx := sdk.UnwrapSDKContext(goCtx)

	prevDocument := keeper.GetDIDDocument(ctx, msg.Did)
	if prevDocument.Empty() {
		return nil, sdkerrors.Wrapf(types.ErrDIDNotFound, "DID: %s", msg.Did)
	}
	if prevDocument.Deactivated {
		return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
	}

	if err := VerifyDIDOwnership(msg.Document.Document, prevDocument.Document); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrVerifyOwnershipFailed, "error: %v", err)
	}

	deactivatedDIDDocument := &types.DIDDocument{
		Document:         nil,
		DocumentDataType: "",
		Deactivated:      true,
	}

	keeper.SetDIDDocument(ctx, msg.Did, deactivatedDIDDocument)
	return &types.MsgDeactivateDIDResponse{}, nil

}

func VerifyDIDOwnership(newDocument, prevDocument []byte) error {

	prevDoc, err := ariesdid.ParseDocument(prevDocument)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrParseDocument, "failed to parse did document")
	}
	newDoc, err := ariesdid.ParseDocument(newDocument)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrParseDocument, "failed to parse did document")
	}

	// todo: Currently, proof is assumed to have only one element. It can be changed to have multiple proofs.

	// get previous document proof verification method
	proofMethodID := newDoc.Proof[0].Creator

	prevVM, err := GetVerificationMethodByID(prevDoc, proofMethodID)
	if err != nil {
		return err
	}
	newVM, err := GetVerificationMethodByID(newDoc, proofMethodID)
	if err != nil {
		return err
	}

	if !IsEqualVerificationMethod(prevVM, newVM) {
		return sdkerrors.Wrapf(types.ErrInvalidProof, "verification method does not match.")
	}

	// check newDoc sequence
	prevSequence, err := strconv.ParseUint(prevDoc.Proof[0].Domain, 10, 64)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidSequence, "can not parse previous document sequence")
	}
	newSequence, err := strconv.ParseUint(newDoc.Proof[0].Domain, 10, 64)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidSequence, "can not parse new document sequence")
	}
	if newSequence != prevSequence+1 {
		return sdkerrors.Wrapf(types.ErrInvalidSequence, "invalid sequence")
	}

	// verify proof itself
	if err := types.VerifyProof(*newDoc); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidProof, "error: %v", err)
	}

	return nil
}

func IsEqualVerificationMethod(vm1, vm2 *ariesdid.VerificationMethod) bool {
	return vm1.ID == vm2.ID && vm1.Type == vm2.Type && vm1.Controller == vm2.Controller && bytes.Equal(vm1.Value, vm2.Value)
}

func GetVerificationMethodByID(doc *ariesdid.Doc, vmID string) (*ariesdid.VerificationMethod, error) {
	for _, vm := range doc.VerificationMethod {
		if vm.ID == vmID {
			return &vm, nil
		}
	}
	return nil, sdkerrors.Wrapf(types.ErrInvalidProof, "can not get VerificationMethod. ID: %v", vmID)
}
