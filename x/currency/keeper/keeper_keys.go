package keeper

import (
	"bytes"
)

var (
	KeyDelimiter = []byte{0x00}

	IssuanceKeyPrefix = []byte{0x11} // {Prefix}{denom}
)

func IssuanceKey(denom string) []byte {
	return bytes.Join([][]byte{
		IssuanceKeyPrefix,
		[]byte(denom),
	}, KeyDelimiter)
}

func getLastElement(key, prefix []byte) []byte {
	return key[len(prefix):]
}
