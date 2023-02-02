package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestMustGetSignBytesWithSeq(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	document := NewDocument(did)
	didDocument, err := NewDIDDocument(document, "aries-framework-go@v1.0.8")
	require.NoError(t, err)

	signBytes := mustGetSignBytesWithSeq(&didDocument, 100)
	require.NotNil(t, signBytes)

	dataWithSeq := DataWithSeq{}
	require.NoError(t, dataWithSeq.Unmarshal(signBytes))

	unmarshalDoc := DIDDocument{}
	require.NoError(t, unmarshalDoc.Unmarshal(dataWithSeq.GetData()))
	require.Equal(t, didDocument.Document, unmarshalDoc.Document)
	require.Equal(t, didDocument.DocumentDataType, unmarshalDoc.DocumentDataType)
	require.Equal(t, uint64(100), dataWithSeq.Sequence)
}

func TestSequence(t *testing.T) {
	seq := InitialSequence
	require.Equal(t, uint64(0), seq)

	nextSeq := nextSequence(seq)
	require.Equal(t, uint64(0), seq)
	require.Equal(t, uint64(1), nextSeq)
}

func TestSignVerify(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	document := NewDocument(did)
	didDocument, err := NewDIDDocument(document, "aries-framework-go@v1.0.8")
	require.NoError(t, err)
	privKey := secp256k1.GenPrivKey()
	seq := uint64(100)

	sig, err := Sign(&didDocument, seq, privKey)
	require.NoError(t, err)
	require.NotNil(t, sig)

	newSeq, ok := Verify(sig, &didDocument, seq, privKey.PubKey())
	require.True(t, ok)
	require.Equal(t, seq+1, newSeq)
}

func TestSignVerify_doInvalid(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	document := NewDocument(did)
	didDocument, err := NewDIDDocument(document, "aries-framework-go@v1.0.8")
	require.NoError(t, err)

	privKey := secp256k1.GenPrivKey()
	anotherPrivKey := secp256k1.GenPrivKey()
	seq := uint64(100)

	sig, err := Sign(&didDocument, seq, privKey)
	require.NoError(t, err)
	require.NotNil(t, sig)

	newSeq, ok := Verify(sig, &didDocument, seq, anotherPrivKey.PubKey())
	require.Equal(t, false, ok)
	require.Equal(t, uint64(0), newSeq)
}
