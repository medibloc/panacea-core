package types

import (
	"strings"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestValidateLocalClassID(t *testing.T) {
	for _, valid := range []string{
		"a",
		".",
		"..",
		"class-1_v2.0",
		strings.Repeat("a", maxLocalClassIDBytes),
	} {
		require.NoError(t, ValidateLocalClassID(valid), valid)
	}

	for _, invalid := range []string{
		"",
		strings.Repeat("a", maxLocalClassIDBytes+1),
		"Class",
		"classé",
		"class id",
		"class\x00id",
		"class/id",
		"creator:class",
	} {
		err := ValidateLocalClassID(invalid)
		require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest, invalid)
	}
}

func TestParseClassID(t *testing.T) {
	creator := strings.Repeat("p", 66)
	localClassID := strings.Repeat("a", maxLocalClassIDBytes)
	classID := creator + ":" + localClassID

	parsedCreator, parsedLocalClassID, err := ParseClassID(classID)
	require.NoError(t, err)
	require.Equal(t, creator, parsedCreator)
	require.Equal(t, localClassID, parsedLocalClassID)
	require.Len(t, classID, maxFullClassIDBytes)

	for _, invalid := range []string{
		"",
		"missing-separator",
		":missing-creator",
		creator + ":" + localClassID + "a",
		creator + ":UPPERCASE",
		creator + ":nested:separator",
	} {
		_, _, err := ParseClassID(invalid)
		require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest, invalid)
	}
}

func TestValidateClassMetadata(t *testing.T) {
	validHash := "sha256:" + strings.Repeat("a", 64)
	require.NoError(t, ValidateClassMetadata(
		strings.Repeat("n", maxClassNameBytes),
		strings.Repeat("s", maxClassSymbolBytes),
		strings.Repeat("d", maxClassDescriptionBytes),
		strings.Repeat("u", maxURIBytes),
		validHash,
	))
	require.NoError(t, ValidateClassMetadata("Café", "CERT", "Résumé", "https://example.test/café", ""))
	require.NoError(t, ValidateClassMetadata("", "", "", "", ""))

	testCases := []struct {
		name        string
		className   string
		symbol      string
		description string
		uri         string
		uriHash     string
	}{
		{name: "name too long", className: strings.Repeat("n", maxClassNameBytes+1)},
		{name: "symbol too long", symbol: strings.Repeat("s", maxClassSymbolBytes+1)},
		{name: "description too long", description: strings.Repeat("d", maxClassDescriptionBytes+1)},
		{name: "uri too long", uri: strings.Repeat("u", maxURIBytes+1)},
		{name: "invalid UTF-8", className: string([]byte{0xff})},
		{name: "control character", description: "line\nbreak"},
		{name: "hash without URI", uriHash: validHash},
		{name: "uppercase hash", uri: "uri", uriHash: "sha256:" + strings.Repeat("A", 64)},
		{name: "wrong hash prefix", uri: "uri", uriHash: "sha512:" + strings.Repeat("a", 64)},
		{name: "short hash", uri: "uri", uriHash: "sha256:a"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateClassMetadata(tc.className, tc.symbol, tc.description, tc.uri, tc.uriHash)
			require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
		})
	}
}

func TestValidateTransferPolicy(t *testing.T) {
	require.NoError(t, ValidateTransferPolicy(TransferPolicy_TRANSFER_POLICY_LOCKED))
	require.NoError(t, ValidateTransferPolicy(TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE))
	require.ErrorIs(
		t,
		ValidateTransferPolicy(TransferPolicy_TRANSFER_POLICY_UNSPECIFIED),
		sdkerrors.ErrInvalidRequest,
	)
	require.ErrorIs(t, ValidateTransferPolicy(TransferPolicy(99)), sdkerrors.ErrInvalidRequest)
}
