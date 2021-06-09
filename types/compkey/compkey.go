package compkey

import (
	"fmt"
	"strings"
)

const sizeUint8 = 1
const maxUint8 = int(^uint8(0))

// CompositeKey is an interface that can be encoded into a byte slice or a string which can be prefix-searched.
// It is used for storing keys in key-value stores that support prefix searching.
// Instead of using separators for encoding, CompositeKey uses the following scheme:
//
// [size_1][value_1][size_2][value_2]...
//
// To compact the size of an encoded key, the size of each value must be in uint8.
type CompositeKey interface {
	ByteSlices() [][]byte
	FromByteSlices([][]byte) error
	Strings() []string
	FromStrings([]string) error
}

// Encode encodes a CompositeKey into a byte slice.
// If the size of some values are not in uint8, it returns a non-nil error.
func Encode(key CompositeKey) ([]byte, error) {
	return encode(key.ByteSlices())
}

// MustEncode calls Encode internally, but it panics if Encode returns a non-nil error.
func MustEncode(key CompositeKey) []byte {
	bz, err := Encode(key)
	if err != nil {
		panic(err)
	}
	return bz
}

// PartialEncode encodes only a few values of CompositeKey into a byte slice.
// If the size of some values are not in uint8, it returns a non-nil error.
func PartialEncode(key CompositeKey, numValues int) ([]byte, error) {
	values := key.ByteSlices()
	if len(values) < numValues {
		return nil, fmt.Errorf("invalid num of values: %d", numValues)
	}
	return encode(values[:numValues])
}

// MustPartialEncode calls PartialEncode internally, but it panics if PartialEncode returns a non-nil error.
func MustPartialEncode(key CompositeKey, numValues int) []byte {
	bz, err := PartialEncode(key, numValues)
	if err != nil {
		panic(err)
	}
	return bz
}

// encode encodes multiple byte slices into a byte slice.
// If the size of some byte slices are not in uint8, it returns a non-nil error.
func encode(values [][]byte) ([]byte, error) {
	size := 0
	for _, value := range values {
		size += sizeUint8 + len(value)
	}

	bz := make([]byte, size)
	idx := 0
	for _, value := range values {
		if len(value) > maxUint8 {
			return nil, fmt.Errorf("the size of value must be in uint8")
		}
		bz[idx] = uint8(len(value))
		idx += 1
		idx += copy(bz[idx:], value)
	}

	return bz, nil
}

// Decode decode a byte slice into a CompositeKey.
// It returns a non-nil error, if the byte slice follows an unexpected encoding.
func Decode(bz []byte, out CompositeKey) error {
	values := make([][]byte, 0)

	idx := 0
	for idx < len(bz) {
		valueSize := int(bz[idx])
		idx += 1
		exclusiveEnd := idx + valueSize

		if exclusiveEnd > len(bz) {
			return fmt.Errorf("failed to decode composite key")
		}
		value := make([]byte, valueSize)
		idx += copy(value, bz[idx:exclusiveEnd])

		values = append(values, value)
	}

	return out.FromByteSlices(values)
}

// MustDecode calls Decode internally, but it panics if Decode returns a non-nil error.
func MustDecode(bz []byte, out CompositeKey) {
	if err := Decode(bz, out); err != nil {
		panic(err)
	}
}

// EncodeToString encodes a CompositeKey into a string with a separator.
// Choose a separator that is not contained in any values of CompositeKey.Strings().
func EncodeToString(key CompositeKey, separator string) string {
	var builder strings.Builder
	for i, value := range key.Strings() {
		if i > 0 {
			builder.WriteString(separator)
		}
		builder.WriteString(value)
	}
	return builder.String()
}

// DecodeFromString decodes a string into a CompositeKey.
// Use a separator that was used when encoding.
func DecodeFromString(encoded string, separator string, out CompositeKey) error {
	values := strings.Split(encoded, separator)
	return out.FromStrings(values)
}

// MustDecodeToString calls DecodeToString internally, but it panics if DecodeToString returns a non-nil error.
func MustDecodeFromString(encoded string, separator string, out CompositeKey) {
	if err := DecodeFromString(encoded, separator, out); err != nil {
		panic(err)
	}
}
