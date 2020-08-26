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

func Sign(data Signable, seq Sequence, privKey crypto.PrivKey) ([]byte, error) {
	return privKey.Sign(mustGetSignBytesWithSeq(data, seq))
}

func Verify(sig []byte, data Signable, seq Sequence, pubKey crypto.PubKey) (Sequence, bool) {
	signBytes := mustGetSignBytesWithSeq(data, seq)
	if !pubKey.VerifyBytes(signBytes, sig) {
		return 0, false
	}
	return seq.next(), true
}

// dataWithSeq is a combination of Seq and any kind of Data.
// The signing is done on this struct.
type dataWithSeq struct {
	Data json.RawMessage `json:"data"`
	Seq  Sequence        `json:"sequence"`
}

// mustGetSignBytesWithSeq returns a byte slice which is the combination of data and seq.
// The return value is deterministic, so that it can be used for signing.
func mustGetSignBytesWithSeq(data Signable, seq Sequence) []byte {
	return sdk.MustSortJSON(didCodec.MustMarshalJSON(dataWithSeq{
		Data: data.GetSignBytes(),
		Seq:  seq,
	}))
}

// Sequence is a preventative measure to distinguish replayed transactions (replay attack).
type Sequence uint64

func NewSequence() Sequence {
	return 0
}

func (s Sequence) next() Sequence {
	return s + 1
}
