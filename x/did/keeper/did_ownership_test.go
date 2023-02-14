package keeper_test

import (
	"testing"

	"github.com/btcsuite/btcd/btcec"
	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/medibloc/panacea-core/v2/x/did/keeper"
	"github.com/medibloc/panacea-core/v2/x/did/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestVerifyDIDOwnership(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	didDocument, btcecPrivKey, vmID := makeTestDIDDocument(did)
	docBz := didDocument.Document

	signedDoc, err := types.SignDocument(docBz, vmID, 0, btcecPrivKey)
	require.NoError(t, err)

	// add new verification method
	doc, err := ariesdid.ParseDocument(docBz)
	require.NoError(t, err)

	newPrivKey := secp256k1.GenPrivKey()
	_, btcecNewPubKey := btcec.PrivKeyFromBytes(btcec.S256(), newPrivKey.Bytes())

	newVerificationMethodID := types.NewVerificationMethodID(did, "key2")
	newVerificationMethod := types.NewVerificationMethod(newVerificationMethodID, types.ES256K_2019, did, btcecNewPubKey.SerializeUncompressed())
	doc.VerificationMethod = append(doc.VerificationMethod, newVerificationMethod)

	newDocBz, err := doc.JSONBytes()
	require.NoError(t, err)

	// sign updated document with previous verification method and private key
	// updated document have to signed with previous sequence+1
	updatedSignedDoc, err := types.SignDocument(newDocBz, vmID, 1, btcecPrivKey)
	require.NoError(t, err)

	err = keeper.VerifyDIDOwnership(updatedSignedDoc, signedDoc)
	require.NoError(t, err)
}

func TestVerifyDIDOwnershipInvalidSequence(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	didDocument, btcecPrivKey, vmID := makeTestDIDDocument(did)
	docBz := didDocument.Document

	signedDoc, err := types.SignDocument(docBz, vmID, 0, btcecPrivKey)
	require.NoError(t, err)

	// sign same document with invalid sequence
	updatedSignedDoc, err := types.SignDocument(docBz, vmID, 2, btcecPrivKey)
	require.NoError(t, err)
	err = keeper.VerifyDIDOwnership(updatedSignedDoc, signedDoc)
	require.Error(t, err)
}

func TestVerifyDIDOwnershipDeleteVerificationMethod(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	didDocument, btcecPrivKey, vmID := makeTestDIDDocument(did)
	docBz := didDocument.Document

	signedDoc, err := types.SignDocument(docBz, vmID, 0, btcecPrivKey)
	require.NoError(t, err)

	// change verification method
	doc, err := ariesdid.ParseDocument(docBz)
	require.NoError(t, err)

	newPrivKey := secp256k1.GenPrivKey()
	btcecNewPrivKey, btcecNewPubKey := btcec.PrivKeyFromBytes(btcec.S256(), newPrivKey.Bytes())

	newVerificationMethodID := types.NewVerificationMethodID(did, "key2")
	newVerificationMethod := types.NewVerificationMethod(newVerificationMethodID, types.ES256K_2019, did, btcecNewPubKey.SerializeUncompressed())
	doc.VerificationMethod = []ariesdid.VerificationMethod{newVerificationMethod}
	doc.Authentication = nil

	newDocBz, err := doc.JSONBytes()
	require.NoError(t, err)

	// sign updated document with previous verification method and private key
	// updated document have to signed with previous sequence+1
	updatedSignedDoc, err := types.SignDocument(newDocBz, vmID, 1, btcecNewPrivKey)
	require.NoError(t, err)

	err = keeper.VerifyDIDOwnership(updatedSignedDoc, signedDoc)
	require.Error(t, err)
}
