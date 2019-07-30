package types

import (
	"fmt"
	"strings"
)

type Topic struct {
	Description  string `json:"description"`
	TotalRecords uint64 `json:"total_records"`
	TotalWriters uint64 `json:"total_writers"`
}

func NewTopic(description string) Topic {
	return Topic{Description: description}
}

func (t Topic) NextRecordOffset() uint64 {
	return t.TotalRecords
}

func (t Topic) IncreaseTotalRecords() Topic {
	return Topic{
		TotalRecords: t.TotalRecords + 1,
		TotalWriters: t.TotalWriters,
	}
}

func (t Topic) IncreaseTotalWriters() Topic {
	return Topic{
		TotalRecords: t.TotalRecords,
		TotalWriters: t.TotalWriters + 1,
	}
}

func (t Topic) DecreaseTotalWriters() Topic {
	return Topic{
		TotalRecords: t.TotalRecords,
		TotalWriters: t.TotalWriters - 1,
	}
}

func (t Topic) String() string {
	return fmt.Sprintf(`Topic:
	Description: %s
	TotalRecords: %d
	TotalWriters: %d`,
		t.Description,
		t.TotalRecords,
		t.TotalWriters,
	)
}

type Topics []string

func (t Topics) String() (out string) {
	for _, topic := range t {
		out += topic + "\n"
	}
	return strings.TrimSpace(out)
}
