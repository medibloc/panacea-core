package keeper

import (
	"bytes"
)

var (
	KeyDelimiter = []byte{0x00}

	DIDDocumentKeyPrefix = []byte{0x11} // {Prefix}{DID}
)

func DIDDocumentKey(did string) []byte {
	return bytes.Join([][]byte{
		DIDDocumentKeyPrefix,
		[]byte(did),
	}, KeyDelimiter)
}

func getLastElement(key, prefix []byte) []byte {
	return key[len(prefix):]
}