package did

import (
	"bytes"

	"github.com/medibloc/panacea-core/x/did/types"
)

var (
	KeyDelimiter = []byte{0x00}

	DIDKeyPrefix = []byte{0x11} // {Prefix}{DID}
)

func DIDKey(did types.DID) []byte {
	return bytes.Join([][]byte{
		DIDKeyPrefix,
		[]byte(did),
	}, KeyDelimiter)
}
