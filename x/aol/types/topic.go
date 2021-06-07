package types

import (
	"regexp"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	maxTopicLength       = 70
	maxDescriptionLength = 5000
)

func (t Topic) Validate() error {
	return validateDescription(t.Description)
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

func validateTopicName(topicName string) error {
	if len(topicName) > maxTopicLength {
		return sdkerrors.Wrapf(ErrMessageTooLarge, "topicName (%d > %d)", len(topicName), maxTopicLength)
	}

	// cannot be an empty string
	if !regexp.MustCompile("^[A-Za-z0-9._-]+$").MatchString(topicName) {
		return sdkerrors.Wrapf(ErrInvalidTopic, "topic %s", topicName)
	}

	return nil
}

func validateDescription(description string) error {
	if len(description) > maxDescriptionLength {
		return sdkerrors.Wrapf(ErrMessageTooLarge, "description (%d > %d)", len(description), maxDescriptionLength)
	}
	return nil
}
