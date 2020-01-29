package types

import (
	"encoding/base64"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Record struct {
	Key           []byte         `json:"key"`
	Value         []byte         `json:"value"`
	NanoTimestamp int64          `json:"nano_timestamp"`
	WriterAddress sdk.AccAddress `json:"writer_address"`
}

func NewRecord(key []byte, value []byte, nanoTimestamp int64, writer sdk.AccAddress) Record {
	return Record{
		Key:           key,
		Value:         value,
		NanoTimestamp: nanoTimestamp,
		WriterAddress: writer,
	}
}

func (r Record) String() string {
	return fmt.Sprintf(`Record:
	Key: %s
	Value: %s
	Accepted Time: %s
	WriterAddress: %s`,
		base64.StdEncoding.EncodeToString(r.Key),
		base64.StdEncoding.EncodeToString(r.Value),
		sdk.FormatTimeBytes(time.Unix(0, r.NanoTimestamp)),
		r.WriterAddress,
	)
}

func (r Record) IsEmpty() bool {
	return r.WriterAddress.Empty()
}
