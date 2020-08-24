package did

import (
	"bytes"

	"github.com/medibloc/panacea-core/x/did/types"
)

var (
	KeyDelimiter = []byte{0x00}

	DIDDocumentKeyPrefix = []byte{0x11} // {Prefix}{DID}
)

func DIDDocumenetKey(did types.DID) []byte {
	return bytes.Join([][]byte{
		DIDDocumentKeyPrefix,
		[]byte(did),
	}, KeyDelimiter)
}
