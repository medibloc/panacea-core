package types

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestMustGetSignBytesWithSeq(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	signableDID := SignableDID(did)
	signBytes := mustGetSignBytesWithSeq(signableDID, 100)
	require.NotNil(t, signBytes)

	var obj dataWithSeq
	require.NoError(t, ModuleCdc.Amino.UnmarshalJSON(signBytes, &obj))
	require.Equal(t, ModuleCdc.Amino.MustMarshalJSON(signableDID), []byte(obj.Data))
	require.Equal(t, uint64(100), obj.Seq)
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
	signableDID := SignableDID(did)
	privKey := secp256k1.GenPrivKey()
	seq := uint64(100)

	sig, err := Sign(signableDID, seq, privKey)
	require.NoError(t, err)
	require.NotNil(t, sig)

	newSeq, ok := Verify(sig, signableDID, seq, privKey.PubKey())
	require.True(t, ok)
	require.Equal(t, seq+1, newSeq)
}

func TestSignVerify_doInvalid(t *testing.T) {
	did := "did:panacea:7Prd74ry1Uct87nZqL3ny7aR7Cg46JamVbJgk8azVgUm"
	signableDID := SignableDID(did)
	privKey := secp256k1.GenPrivKey()
	anotherPrivKey := secp256k1.GenPrivKey()
	seq := uint64(100)

	sig, err := Sign(signableDID, seq, privKey)
	require.NoError(t, err)
	require.NotNil(t, sig)

	newSeq, ok := Verify(sig, signableDID, seq, anotherPrivKey.PubKey())
	require.Equal(t, false, ok)
	require.Equal(t, uint64(0), newSeq)
}
