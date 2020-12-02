package aol

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	KeyDelimiter = []byte{0x00}

	OwnerKeyPrefix  = []byte{0x11} // {Prefix}{Owner}
	TopicKeyPrefix  = []byte{0x12} // {Prefix}{Owner}{Topic}
	ACLWriterPrefix = []byte{0x13} // {Prefix}{Owner}{Topic}{Writer}
	RecordKeyPrefix = []byte{0x14} // {Prefix}{Owner}{Topic}{Offset}
)

func OwnerKey(ownerAddr sdk.AccAddress) []byte {
	return bytes.Join([][]byte{
		OwnerKeyPrefix,
		ownerAddr.Bytes(),
	}, KeyDelimiter)
}

func TopicKey(ownerAddr sdk.AccAddress, topic string) []byte {
	return bytes.Join([][]byte{
		TopicKeyPrefix,
		ownerAddr.Bytes(),
		[]byte(topic),
	}, KeyDelimiter)
}

func ACLWriterKey(ownerAddr sdk.AccAddress, topic string, writerAddr sdk.Address) []byte {
	return bytes.Join([][]byte{
		ACLWriterPrefix,
		ownerAddr.Bytes(),
		[]byte(topic),
		writerAddr.Bytes(),
	}, KeyDelimiter)
}

func RecordKey(ownerAddr sdk.AccAddress, topic string, offset uint64) []byte {
	return bytes.Join([][]byte{
		RecordKeyPrefix,
		ownerAddr.Bytes(),
		[]byte(topic),
		sdk.Uint64ToBigEndian(offset),
	}, KeyDelimiter)
}

func getLastElement(key, prefix []byte) []byte {
	return key[len(prefix):]
}
