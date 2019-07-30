package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Writer struct {
	Moniker       string `json:"moniker"`
	Description   string `json:"description"`
	NanoTimestamp int64  `json:"nano_timestamp"`
}

func NewWriter(moniker string, description string, nanoTimestamp int64) Writer {
	return Writer{
		Moniker:       moniker,
		Description:   description,
		NanoTimestamp: nanoTimestamp,
	}
}

func (w Writer) String() string {
	return fmt.Sprintf(`WriterAddress:
	Moniker: %v
	Description: %v
	Registered Time: %v`,
		w.Moniker,
		w.Description,
		sdk.FormatTimeBytes(time.Unix(0, w.NanoTimestamp)),
	)
}

type Writers []string

func (w Writers) String() (out string) {
	for _, writer := range w {
		out += writer + "\n"
	}
	return strings.TrimSpace(out)
}
