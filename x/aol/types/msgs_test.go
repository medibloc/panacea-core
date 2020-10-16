package types

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTopic(t *testing.T) {
	assert.Nil(t, validateTopic("a.B_c-D123"))

	assert.Equal(t, ErrInvalidTopic(""), validateTopic(""))
	assert.Equal(t, ErrInvalidTopic("a$"), validateTopic("a$"))
	assert.Equal(t, ErrInvalidTopic("a b"), validateTopic("a b"))
	assert.Equal(t, ErrInvalidTopic(" ab"), validateTopic(" ab"))
	assert.Equal(t, ErrInvalidTopic("ab "), validateTopic("ab "))

	var buf bytes.Buffer
	for i := 0; i < MaxTopicLength+1; i++ {
		buf.WriteByte('a')
	}
	assert.Equal(t, ErrMessageTooLarge("topic", MaxTopicLength+1, MaxTopicLength), validateTopic(buf.String()))
}

func TestValidateMoniker(t *testing.T) {
	assert.Nil(t, validateMoniker("a.B_c-D123"))
	assert.Nil(t, validateMoniker(""))

	assert.Equal(t, ErrInvalidMoniker("a$"), validateMoniker("a$"))
	assert.Equal(t, ErrInvalidMoniker("a b"), validateMoniker("a b"))
	assert.Equal(t, ErrInvalidMoniker(" ab"), validateMoniker(" ab"))
	assert.Equal(t, ErrInvalidMoniker("ab "), validateMoniker("ab "))

	var buf bytes.Buffer
	for i := 0; i < MaxMonikerLength+1; i++ {
		buf.WriteByte('a')
	}
	assert.Equal(t, ErrMessageTooLarge("moniker", MaxMonikerLength+1, MaxMonikerLength), validateMoniker(buf.String()))
}

func TestValidateDescription(t *testing.T) {
	assert.Nil(t, validateDescription(""))
	assert.Nil(t, validateDescription("abc"))

	var buf bytes.Buffer
	for i := 0; i < MaxDescriptionLength+1; i++ {
		buf.WriteByte('a')
	}
	assert.Equal(t, ErrMessageTooLarge("description", MaxDescriptionLength+1, MaxDescriptionLength), validateDescription(buf.String()))
}

func TestValidateRecordKey(t *testing.T) {
	assert.Nil(t, validateRecordKey([]byte{}))
	assert.Nil(t, validateRecordKey([]byte("abc")))

	var buf bytes.Buffer
	for i := 0; i < MaxRecordKeyLength+1; i++ {
		buf.WriteByte('a')
	}
	assert.Equal(t, ErrMessageTooLarge("key", MaxRecordKeyLength+1, MaxRecordKeyLength), validateRecordKey(buf.Bytes()))
}

func TestValidateRecordValue(t *testing.T) {
	assert.Nil(t, validateRecordValue([]byte{}))
	assert.Nil(t, validateRecordValue([]byte("abc")))

	var buf bytes.Buffer
	for i := 0; i < MaxRecordValueLength+1; i++ {
		buf.WriteByte('a')
	}
	assert.Equal(t, ErrMessageTooLarge("value", MaxRecordValueLength+1, MaxRecordValueLength), validateRecordValue(buf.Bytes()))
}
