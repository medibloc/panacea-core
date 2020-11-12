package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestName_validate(t *testing.T) {
	require.NoError(t, Name("my token").validate())
	require.Equal(t, ErrInvalidName(""), Name("").validate())

	longName := "aaaaaaaaaaaaaaaaaaaaaa"
	require.NoError(t, Name(longName).validate())
	longName = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	require.Equal(t, ErrInvalidName(longName), Name(longName).validate())
}

func TestShortSymbol_validate(t *testing.T) {
	require.NoError(t, ShortSymbol("LoV").validate())
	require.Equal(t, ErrInvalidSymbol("1ov"), ShortSymbol("1ov").validate())
	require.Equal(t, ErrInvalidSymbol("Lo"), ShortSymbol("Lo").validate())

	longSymbol := "aaaaaaaaaaaaa"
	require.NoError(t, ShortSymbol(longSymbol).validate())
	longSymbol = "aaaaaaaaaaaaaa"
	require.Equal(t, ErrInvalidSymbol(longSymbol), ShortSymbol(longSymbol).validate())

	require.Equal(t, ErrSymbolNotAllowed("med"), ShortSymbol("med").validate())
}

func TestSymbol_validate(t *testing.T) {
	require.NoError(t, Symbol("LoV-0eA").validate())
	require.Equal(t, ErrInvalidSymbol("LoV0eA"), Symbol("LoV0eA").validate())
	require.Equal(t, ErrInvalidSymbol("LoV-0eAB"), Symbol("LoV-0eAB").validate())
}
