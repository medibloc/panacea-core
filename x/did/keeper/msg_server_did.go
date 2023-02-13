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

	if !msg.Document.Valid() {
		return nil, sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "DID: %s", msg.Did)
	}

	keeper.SetDIDDocument(ctx, msg.Did, msg.Document)
	return &types.MsgCreateDIDResponse{}, nil
}

func (m msgServer) UpdateDID(goCtx context.Context, msg *types.MsgUpdateDID) (*types.MsgUpdateDIDResponse, error) {
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
		return &types.MsgUpdateDIDResponse{}, err
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
		return sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "failed to parse did document")
	}

	// get previous document proof verification method
	proofMehtodID := doc.Proof[0].Creator
	var verificationMethod ariesdid.VerificationMethod

	for _, vm := range doc.VerificationMethod {
		if vm.ID == proofMehtodID {
			verificationMethod = vm
		}
	}

	// check if the proof was created with the verification method used in the previous document
	newDoc, err := ariesdid.ParseDocument(newDocument)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "failed to parse did document")
	}

	check := false
	for _, vm := range newDoc.VerificationMethod {
		if vm.ID == proofMehtodID {
			check = IsEqualVerificationMethod(vm, verificationMethod)
		}
	}
	if check == false {
		return sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "there is no proof verification method in document")
	}

	// check newDoc proof method id
	if newDoc.Proof[0].Creator != proofMehtodID {
		return sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "does not match previous proof method")
	}

	// check newDoc domain
	prevSequence, err := strconv.ParseUint(doc.Proof[0].Domain, 10, 64)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "invalid sequence")
	}
	newSequence, err := strconv.ParseUint(newDoc.Proof[0].Domain, 10, 64)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "invalid sequence")
	}
	if newSequence != prevSequence+1 {
		return sdkerrors.Wrapf(types.ErrInvalidDIDDocument, "invalid sequence")
	}

	// verify proof
	if err := types.VerifyProof(*newDoc); err != nil {
		return err
	}

	return nil
}

func IsEqualVerificationMethod(vm1, vm2 ariesdid.VerificationMethod) bool {
	return vm1.ID == vm2.ID && vm1.Type == vm2.Type && vm1.Controller == vm2.Controller && bytes.Equal(vm1.Value, vm2.Value)
}
