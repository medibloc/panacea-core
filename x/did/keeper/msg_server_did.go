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
	return &types.MsgCreateDIDResponse{}, nil
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

	document := keeper.GetDIDDocument(ctx, msg.Did)
	if document.Empty() {
		return nil, sdkerrors.Wrapf(types.ErrDIDNotFound, "DID: %s", msg.Did)
	}
	if document.Deactivated {
		return nil, sdkerrors.Wrapf(types.ErrDIDDeactivated, "DID: %s", msg.Did)
	}
	document.Deactivated = true

	keeper.SetDIDDocument(ctx, msg.Did, document)
	return &types.MsgDeactivateDIDResponse{}, nil

}

func VerifyDIDOwnership(newDocument, prevDocument []byte) error {

	doc, err := ariesdid.ParseDocument(prevDocument)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrParseDocument, "failed to parse did document")
	}
	newDoc, err := ariesdid.ParseDocument(newDocument)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrParseDocument, "failed to parse did document")
	}

	// todo: Currently, only the key used in CreateDID is used for ownership verification.
	//       Since the aries framework can create proof array based on multiple keys, this part can be extended.

	// get previous document proof verification method
	proofMethodID := doc.Proof[0].Creator
	var verificationMethod ariesdid.VerificationMethod

	for _, vm := range doc.VerificationMethod {
		if vm.ID == proofMethodID {
			verificationMethod = vm
		}
	}

	// check if the proof was created with the verification method used in the previous document
	check := false
	for _, vm := range newDoc.VerificationMethod {
		if vm.ID == proofMethodID {
			check = IsEqualVerificationMethod(vm, verificationMethod)
		}
	}
	if check == false {
		return sdkerrors.Wrapf(types.ErrInvalidProof, "there is no proof verification method in document")
	}

	// Check if newDoc proof method matches previous document proof method
	if newDoc.Proof[0].Creator != proofMethodID {
		return sdkerrors.Wrapf(types.ErrInvalidProof, "does not match previous proof method")
	}

	// check newDoc sequence
	prevSequence, err := strconv.ParseUint(doc.Proof[0].Domain, 10, 64)
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

func IsEqualVerificationMethod(vm1, vm2 ariesdid.VerificationMethod) bool {
	return vm1.ID == vm2.ID && vm1.Type == vm2.Type && vm1.Controller == vm2.Controller && bytes.Equal(vm1.Value, vm2.Value)
}
