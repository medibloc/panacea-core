package compkey_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/libs/rand"

	"github.com/stretchr/testify/require"

	"github.com/medibloc/panacea-core/types/compkey"
)

const (
	maxUint8 = int(^uint8(0))
)

func TestEncodeDecode(t *testing.T) {
	input := myKey{
		str: rand.Str(maxUint8),
		num: uint64(100),
	}

	encoded, err := compkey.Encode(&input)
	require.NoError(t, err)

	var decoded myKey
	err = compkey.Decode(encoded, &decoded)
	require.NoError(t, err)
	require.EqualValues(t, input, decoded)
}

func TestEncodeFailure(t *testing.T) {
	input := myKey{
		str: rand.Str(maxUint8 + 1),
		num: uint64(100),
	}

	_, err := compkey.Encode(&input)
	require.Error(t, err)

	require.Panics(t, func() {
		compkey.MustEncode(&input)
	})
}

func TestDecodeFailure(t *testing.T) {
	input := myKey{
		str: rand.Str(maxUint8),
		num: uint64(100),
	}
	encoded := compkey.MustEncode(&input)

	// violates encoded bytes
	encoded = encoded[:len(encoded)-1]

	var decoded myKey
	require.Error(t, compkey.Decode(encoded, &decoded))
	require.Panics(t, func() {
		compkey.MustDecode(encoded, &decoded)
	})
}

func TestPartialEncode(t *testing.T) {
	input := myKey{
		str: rand.Str(maxUint8),
		num: uint64(100),
	}
	partial, err := compkey.PartialEncode(&input, 1)
	require.NoError(t, err)
	entire, err := compkey.Encode(&input)
	require.NoError(t, err)

	require.Equal(t, entire[:len(entire)-1-8], partial) // sizeUint8 is 1 and sizeUint64 is 8
}

func TestPartialEncodeFailure(t *testing.T) {
	input := myKey{
		str: rand.Str(maxUint8),
		num: uint64(100),
	}

	_, err := compkey.PartialEncode(&input, 3)
	require.Error(t, err)

	require.Panics(t, func() {
		compkey.MustPartialEncode(&input, 3)
	})
}

func TestEncodeDecodeString(t *testing.T) {
	input := myKey{
		str: "hello",
		num: uint64(100),
	}

	encoded := compkey.EncodeToString(&input, "/")
	require.Equal(t, "hello/100", encoded)

	var decoded myKey
	err := compkey.DecodeFromString(encoded, "/", &decoded)
	require.NoError(t, err)
	require.EqualValues(t, input, decoded)
}

func TestDecodeFromStringFailure(t *testing.T) {
	encoded := "hello" // doesn't contain the 2nd value

	var decoded myKey
	err := compkey.DecodeFromString(encoded, "/", &decoded)
	require.Error(t, err)
	require.Panics(t, func() {
		compkey.MustDecodeFromString(encoded, "/", &decoded)
	})
}

var _ compkey.CompositeKey = &myKey{}

type myKey struct {
	str string
	num uint64
}

func (m myKey) ByteSlices() [][]byte {
	return [][]byte{
		[]byte(m.str),
		types.Uint64ToBigEndian(m.num),
	}
}

func (m *myKey) FromByteSlices(bzs [][]byte) error {
	if len(bzs) != 2 {
		return fmt.Errorf("invalid len of byte slices")
	}

	m.str = string(bzs[0])
	m.num = types.BigEndianToUint64(bzs[1])
	return nil
}

func (m myKey) Strings() []string {
	return []string{
		m.str,
		strconv.FormatUint(m.num, 10),
	}
}

func (m *myKey) FromStrings(strings []string) error {
	if len(strings) != 2 {
		return fmt.Errorf("invalid len of strings")
	}

	num, err := strconv.ParseUint(strings[1], 10, 64)
	if err != nil {
		return err
	}

	m.str = strings[0]
	m.num = num
	return nil
}
