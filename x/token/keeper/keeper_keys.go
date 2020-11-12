package keeper

import (
	"bytes"

	"github.com/medibloc/panacea-core/x/token/types"
)

var (
	keyDelimiter = []byte{0x00}

	tokenKeyPrefix = []byte{0x11} // {Prefix}{symbol}
)

func TokenKey(symbol types.Symbol) []byte {
	return bytes.Join([][]byte{
		tokenKeyPrefix,
		[]byte(symbol.ToUpper()), // Symbol is case-insensitive.
	}, keyDelimiter)
}

func getLastElement(key, prefix []byte) []byte {
	return key[len(prefix):]
}
