package types

import (
	"strings"
	"testing"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
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

func TestValidateNFTID(t *testing.T) {
	for _, valid := range []string{
		"a",
		"nft-1_v2.0",
		strings.Repeat("a", maxNFTIDBytes),
	} {
		require.NoError(t, ValidateNFTID(valid), valid)
	}

	for _, invalid := range []string{
		"",
		".",
		"..",
		strings.Repeat("a", maxNFTIDBytes+1),
		"NFT",
		"nfté",
		"nft id",
		"nft\x00id",
		"nft/id",
	} {
		err := ValidateNFTID(invalid)
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

func TestCanonicalizeNFTData(t *testing.T) {
	_, cdc := newTestCodec()
	metadata := &BasicNFTData{
		Name:        "Certificate #1",
		Description: "Completion certificate",
		ImageUri:    "https://example.test/image.png",
	}
	packed, err := cdctypes.NewAnyWithValue(metadata)
	require.NoError(t, err)

	canonical, err := CanonicalizeNFTData(cdc, packed)
	require.NoError(t, err)
	require.NotSame(t, packed, canonical)
	require.Equal(t, BasicNFTDataTypeURL, canonical.TypeUrl)
	require.Equal(t, packed.Value, canonical.Value)
	require.Nil(t, canonical.GetCachedValue())

	canonical, err = CanonicalizeNFTData(cdc, nil)
	require.NoError(t, err)
	require.Nil(t, canonical)

	maximum, err := cdctypes.NewAnyWithValue(&BasicNFTData{
		Name: strings.Repeat("n", maxBasicNFTDataBytes-3),
	})
	require.NoError(t, err)
	require.Len(t, maximum.Value, maxBasicNFTDataBytes)
	_, err = CanonicalizeNFTData(cdc, maximum)
	require.NoError(t, err)
}

func TestCanonicalizeNFTDataRejectsInvalidWireData(t *testing.T) {
	_, cdc := newTestCodec()
	valid, err := cdctypes.NewAnyWithValue(&BasicNFTData{Name: "name"})
	require.NoError(t, err)
	invalidUTF8, err := cdctypes.NewAnyWithValue(&BasicNFTData{Name: string([]byte{0xff})})
	require.NoError(t, err)
	controlCharacter, err := cdctypes.NewAnyWithValue(&BasicNFTData{Description: "line\nbreak"})
	require.NoError(t, err)
	longImageURI, err := cdctypes.NewAnyWithValue(&BasicNFTData{
		ImageUri: strings.Repeat("u", maxURIBytes+1),
	})
	require.NoError(t, err)
	tooLarge, err := cdctypes.NewAnyWithValue(&BasicNFTData{
		Name: strings.Repeat("n", maxBasicNFTDataBytes-2),
	})
	require.NoError(t, err)
	require.Greater(t, len(tooLarge.Value), maxBasicNFTDataBytes)
	cachedMismatch, err := cdctypes.NewAnyWithValue(&BasicNFTData{Name: "cached"})
	require.NoError(t, err)
	cachedMismatch.Value = nil

	testCases := []struct {
		name string
		data *cdctypes.Any
	}{
		{
			name: "unknown type URL",
			data: &cdctypes.Any{
				TypeUrl: "/panacea.nft.v1.UnknownNFTData",
				Value:   valid.Value,
			},
		},
		{
			name: "malformed protobuf",
			data: &cdctypes.Any{TypeUrl: BasicNFTDataTypeURL, Value: []byte{0xff}},
		},
		{
			name: "empty metadata",
			data: &cdctypes.Any{TypeUrl: BasicNFTDataTypeURL},
		},
		{name: "invalid UTF-8", data: invalidUTF8},
		{name: "control character", data: controlCharacter},
		{name: "image URI too long", data: longImageURI},
		{name: "encoded data too large", data: tooLarge},
		{
			name: "non-canonical field order",
			data: &cdctypes.Any{
				TypeUrl: BasicNFTDataTypeURL,
				Value:   []byte{0x12, 0x01, 'd', 0x0a, 0x01, 'n'},
			},
		},
		{name: "cached value disagrees with wire bytes", data: cachedMismatch},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := CanonicalizeNFTData(cdc, tc.data)
			require.ErrorIs(t, err, sdkerrors.ErrInvalidRequest)
		})
	}
}
