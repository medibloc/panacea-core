package types

import (
	"cosmossdk.io/errors"
	"regexp"
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
		Description:  t.Description,
	}
}

func (t Topic) IncreaseTotalWriters() Topic {
	return Topic{
		TotalRecords: t.TotalRecords,
		TotalWriters: t.TotalWriters + 1,
		Description:  t.Description,
	}
}

func (t Topic) DecreaseTotalWriters() Topic {
	return Topic{
		TotalRecords: t.TotalRecords,
		TotalWriters: t.TotalWriters - 1,
		Description:  t.Description,
	}
}

func validateTopicName(topicName string) error {
	if len(topicName) > maxTopicLength {
		return errors.Wrapf(ErrMessageTooLarge, "topicName (%d > %d)", len(topicName), maxTopicLength)
	}

	// cannot be an empty string
	if !regexp.MustCompile("^[A-Za-z0-9._-]+$").MatchString(topicName) {
		return errors.Wrapf(ErrInvalidTopic, "topic %s", topicName)
	}

	return nil
}

func validateDescription(description string) error {
	if len(description) > maxDescriptionLength {
		return errors.Wrapf(ErrMessageTooLarge, "description (%d > %d)", len(description), maxDescriptionLength)
	}
	return nil
}
