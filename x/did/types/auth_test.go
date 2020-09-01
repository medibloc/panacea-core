package types

import (
	"testing"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/require"
)

func TestMustGetSignBytesWithSeq(t *testing.T) {
	did := DID("did:panacea:testnet:UCqMYBSt8efvLQ59wNmMo")
	signBytes := mustGetSignBytesWithSeq(did, Sequence(100))
	require.NotNil(t, signBytes)

	var obj dataWithSeq
	require.NoError(t, codec.New().UnmarshalJSON(signBytes, &obj))
	require.Equal(t, did.GetSignBytes(), []byte(obj.Data))
	require.Equal(t, Sequence(100), obj.Seq)
}

func TestSequence(t *testing.T) {
	seq := InitialSequence
	require.Equal(t, Sequence(0), seq)

	nextSeq := seq.next()
	require.Equal(t, Sequence(0), seq)
	require.Equal(t, Sequence(1), nextSeq)
}

func TestSignVerify(t *testing.T) {
	did := DID("did:panacea:testnet:UCqMYBSt8efvLQ59wNmMo")
	privKey := secp256k1.GenPrivKey()
	seq := Sequence(100)

	sig, err := Sign(did, seq, privKey)
	require.NoError(t, err)
	require.NotNil(t, sig)

	newSeq, ok := Verify(sig, did, seq, privKey.PubKey())
	require.True(t, ok)
	require.Equal(t, seq+1, newSeq)
}
