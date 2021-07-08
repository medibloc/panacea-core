package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestMustGetSignBytesWithSeq(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	doc := DIDDocument{
		Id: did,
	}

	signBytes := mustGetSignBytesWithSeq(&doc, 100)
	require.NotNil(t, signBytes)

	dataWithSeq := DataWithSeq{}
	require.NoError(t, dataWithSeq.Unmarshal(signBytes))

	unmarshalDoc := DIDDocument{}
	require.NoError(t, unmarshalDoc.Unmarshal(dataWithSeq.GetData()))
	require.Equal(t, doc.Id, unmarshalDoc.Id)
	require.Equal(t, uint64(100), dataWithSeq.Seq)
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
	doc := DIDDocument{
		Id: did,
	}
	privKey := secp256k1.GenPrivKey()
	seq := uint64(100)

	sig, err := Sign(&doc, seq, privKey)
	require.NoError(t, err)
	require.NotNil(t, sig)

	newSeq, ok := Verify(sig, &doc, seq, privKey.PubKey())
	require.True(t, ok)
	require.Equal(t, seq+1, newSeq)
}

func TestSignVerify_doInvalid(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	doc := DIDDocument{
		Id: did,
	}
	privKey := secp256k1.GenPrivKey()
	anotherPrivKey := secp256k1.GenPrivKey()
	seq := uint64(100)

	sig, err := Sign(&doc, seq, privKey)
	require.NoError(t, err)
	require.NotNil(t, sig)

	newSeq, ok := Verify(sig, &doc, seq, anotherPrivKey.PubKey())
	require.Equal(t, false, ok)
	require.Equal(t, uint64(0), newSeq)
}
