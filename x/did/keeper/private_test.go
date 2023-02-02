package keeper

import (
	"testing"
	"time"

	ariesdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/medibloc/panacea-core/v2/x/did/internal/secp256k1util"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/medibloc/panacea-core/v2/x/did/types"
)

func TestVerifyDIDOwnership(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	docWithSeq, privKey := newDIDDocumentWithSeq(did)
	doc := docWithSeq.Document

	sig, _ := types.Sign(doc, docWithSeq.Sequence, privKey)

	newSeq, err := VerifyDIDOwnership(doc, docWithSeq.Sequence, docWithSeq.Document, sig)
	require.NoError(t, err)
	require.Equal(t, docWithSeq.Sequence+1, newSeq)
}

func TestVerifyDIDOwnership_SigVerificationFailed(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	docWithSeq, privKey := newDIDDocumentWithSeq(did)
	doc := docWithSeq.Document

	sig, _ := types.Sign(doc, docWithSeq.Sequence+11234, privKey)

	_, err := VerifyDIDOwnership(doc, docWithSeq.Sequence, docWithSeq.Document, sig)
	require.ErrorIs(t, types.ErrSigVerificationFailed, err)
}

func newDIDDocumentWithSeq(id string) (types.DIDDocumentWithSeq, crypto.PrivKey) {
	privKey := secp256k1.GenPrivKey()
	pubKey := secp256k1util.PubKeyBytes(secp256k1util.DerivePubKey(privKey))
	verificationMethodID := types.NewVerificationMethodID(id, "key1")
	verificationMethod := []ariesdid.VerificationMethod{
		{
			ID:         verificationMethodID,
			Type:       types.ES256K_2019,
			Controller: id,
			Value:      pubKey,
		},
		{
			ID:         verificationMethodID,
			Type:       types.BLS1281G2_2020,
			Controller: id,
			Value:      []byte("dummy BBS+ pub key"),
		},
	}

	authentication := []ariesdid.Verification{
		{VerificationMethod: *ariesdid.NewVerificationMethodFromBytes(verificationMethodID,
			types.ES256K_2019,
			id,
			pubKey), Relationship: ariesdid.Authentication},
		{VerificationMethod: ariesdid.VerificationMethod{
			ID:         verificationMethodID,
			Type:       types.ES256K_2019,
			Controller: id,
			Value:      pubKey,
		}},
	}

	createdTime := time.Now()

	doc := &ariesdid.Doc{
		Context:            []string{types.ContextDIDV1},
		ID:                 id,
		VerificationMethod: verificationMethod,
		Authentication:     authentication,
		Created:            &createdTime,
	}
	docBz, _ := doc.JSONBytes()

	document := &types.DIDDocument{
		Document:         docBz,
		DocumentDataType: "aries-framework-go@v0.1.8",
	}

	docWithSeq := types.NewDIDDocumentWithSeq(
		document,
		types.InitialSequence,
	)
	return docWithSeq, privKey
}
