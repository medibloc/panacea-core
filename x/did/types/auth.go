package types

import (
	"fmt"

	sdkcodec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/crypto"
)

func Sign(signableData sdkcodec.ProtoMarshaler, seq uint64, privKey crypto.PrivKey) ([]byte, error) {
	return privKey.Sign(mustGetSignBytesWithSeq(signableData, seq))
}

func Verify(signature []byte, signableData sdkcodec.ProtoMarshaler, seq uint64, pubKey crypto.PubKey) (uint64, bool) {
	signBytes := mustGetSignBytesWithSeq(signableData, seq)

	if !pubKey.VerifySignature(signBytes, signature) {
		return 0, false
	}
	return nextSequence(seq), true
}

// mustGetSignBytesWithSeq returns a byte slice which is the combination of data and seq.
// The return value is deterministic, so that it can be used for signing.
func mustGetSignBytesWithSeq(signableData sdkcodec.ProtoMarshaler, seq uint64) []byte {
	dAtA, err := signableData.Marshal()
	if err != nil {
		panic(fmt.Sprintf("marshal failed: %s, signableData: %s", err.Error(), signableData))
	}
	dataWithSeq := DataWithSeq{
		Data:     dAtA,
		Sequence: seq,
	}

	dAtA, err = dataWithSeq.Marshal()

	if err != nil {
		panic(fmt.Sprintf("marshal failed: %s, dataWithSeq: %v", err.Error(), dataWithSeq))
	}

	return dAtA
}

// Sequence is a preventative measure to distinguish replayed transactions (replay attack).
const InitialSequence uint64 = 0

func nextSequence(seq uint64) uint64 {
	return seq + 1
}
