package types

import (
	"encoding/json"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Signable is an interface for data which can be converted to a sign-able byte slice (deterministic).
type Signable interface {
	// Get the canonical byte representation of the data.
	GetSignBytes() []byte
}

func Sign(data Signable, seq uint64, privKey crypto.PrivKey) ([]byte, error) {
	return privKey.Sign(mustGetSignBytesWithSeq(data, seq))
}

func Verify(sig []byte, data Signable, seq uint64, pubKey crypto.PubKey) (uint64, bool) {
	signBytes := mustGetSignBytesWithSeq(data, seq)
	if !pubKey.VerifySignature(signBytes, sig) {
		return 0, false
	}
	return nextSequence(seq), true
}

// dataWithSeq is a combination of Seq and any kind of Data.
// The signing is done on this struct.
type dataWithSeq struct {
	Data json.RawMessage `json:"data"`
	Seq  uint64          `json:"sequence"`
}

// mustGetSignBytesWithSeq returns a byte slice which is the combination of data and seq.
// The return value is deterministic, so that it can be used for signing.
func mustGetSignBytesWithSeq(data Signable, seq uint64) []byte {
	return sdk.MustSortJSON(ModuleCdc.Amino.MustMarshalJSON(dataWithSeq{
		Data: data.GetSignBytes(),
		Seq:  seq,
	}))
}

// Sequence is a preventative measure to distinguish replayed transactions (replay attack).
const InitialSequence uint64 = 0

func nextSequence(seq uint64) uint64 {
	return seq + 1
}
