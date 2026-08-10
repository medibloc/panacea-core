package types

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	maxLocalClassIDBytes     = 64
	maxFullClassIDBytes      = 131
	maxClassNameBytes        = 128
	maxClassSymbolBytes      = 32
	maxClassDescriptionBytes = 1024
	maxURIBytes              = 256
	maxNFTIDBytes            = 64
	maxBasicNFTDataBytes     = 1024
)

var uriHashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// ValidateLocalClassID enforces the creator-local portion of a class ID.
func ValidateLocalClassID(localClassID string) error {
	if len(localClassID) == 0 || len(localClassID) > maxLocalClassIDBytes {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"local_class_id must be between 1 and %d bytes",
			maxLocalClassIDBytes,
		)
	}
	for i := 0; i < len(localClassID); i++ {
		character := localClassID[i]
		if (character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') ||
			character == '.' || character == '_' || character == '-' {
			continue
		}
		return sdkerrors.ErrInvalidRequest.Wrap(
			"local_class_id must contain only lowercase letters, digits, '.', '_', or '-'",
		)
	}
	return nil
}

// ValidateNFTID enforces the class-local immutable NFT identifier contract.
func ValidateNFTID(nftID string) error {
	if len(nftID) == 0 || len(nftID) > maxNFTIDBytes {
		return sdkerrors.ErrInvalidRequest.Wrapf(
			"nft_id must be between 1 and %d bytes",
			maxNFTIDBytes,
		)
	}
	if nftID == "." || nftID == ".." {
		return sdkerrors.ErrInvalidRequest.Wrap("nft_id must not be '.' or '..'")
	}
	for i := 0; i < len(nftID); i++ {
		character := nftID[i]
		if (character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') ||
			character == '.' || character == '_' || character == '-' {
			continue
		}
		return sdkerrors.ErrInvalidRequest.Wrap(
			"nft_id must contain only lowercase letters, digits, '.', '_', or '-'",
		)
	}
	return nil
}

// ParseClassID checks the chain-generated class ID shape and returns its two
// components. Address decoding and canonicalization remain codec-dependent.
func ParseClassID(classID string) (creator, localClassID string, err error) {
	if len(classID) == 0 || len(classID) > maxFullClassIDBytes {
		return "", "", sdkerrors.ErrInvalidRequest.Wrapf(
			"class_id must be between 1 and %d bytes",
			maxFullClassIDBytes,
		)
	}
	creator, localClassID, found := strings.Cut(classID, ":")
	if !found || creator == "" {
		return "", "", sdkerrors.ErrInvalidRequest.Wrap(
			"class_id must contain a creator namespace and local class ID",
		)
	}
	if err := ValidateLocalClassID(localClassID); err != nil {
		return "", "", err
	}
	return creator, localClassID, nil
}

// ValidateClassMetadata enforces protocol limits on immutable class metadata.
func ValidateClassMetadata(name, symbol, description, uri, uriHash string) error {
	if err := validateText("name", name, maxClassNameBytes); err != nil {
		return err
	}
	if err := validateText("symbol", symbol, maxClassSymbolBytes); err != nil {
		return err
	}
	if err := validateText("description", description, maxClassDescriptionBytes); err != nil {
		return err
	}
	return ValidateURI(uri, uriHash)
}

// ValidateURI enforces the common URI and content-hash contract.
func ValidateURI(uri, uriHash string) error {
	if err := validateText("uri", uri, maxURIBytes); err != nil {
		return err
	}
	if uriHash == "" {
		return nil
	}
	if uri == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("uri_hash requires uri")
	}
	if !uriHashPattern.MatchString(uriHash) {
		return sdkerrors.ErrInvalidRequest.Wrap(
			"uri_hash must use sha256 followed by 64 lowercase hexadecimal characters",
		)
	}
	return nil
}

// CanonicalizeNFTData validates the closed NFTData interface and returns a
// fresh Any containing only its canonical type URL and deterministic bytes.
// Copying the wire fields before unpacking deliberately ignores cached values.
func CanonicalizeNFTData(
	unpacker cdctypes.AnyUnpacker,
	data *cdctypes.Any,
) (*cdctypes.Any, error) {
	if data == nil {
		return nil, nil
	}
	if unpacker == nil {
		return nil, fmt.Errorf("nft data validation requires an Any unpacker")
	}
	if data.TypeUrl != BasicNFTDataTypeURL {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf(
			"data type_url must be %s",
			BasicNFTDataTypeURL,
		)
	}

	wireData := &cdctypes.Any{
		TypeUrl: data.TypeUrl,
		Value:   append([]byte(nil), data.Value...),
	}
	var unpacked NFTData
	if err := unpacker.UnpackAny(wireData, &unpacked); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("unpack data: %v", err)
	}
	metadata, ok := unpacked.(*BasicNFTData)
	if !ok {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf(
			"unsupported data implementation %T",
			unpacked,
		)
	}
	if metadata.Name == "" && metadata.Description == "" && metadata.ImageUri == "" {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("data must contain at least one field")
	}
	if err := validateText("data.name", metadata.Name, maxBasicNFTDataBytes); err != nil {
		return nil, err
	}
	if err := validateText("data.description", metadata.Description, maxBasicNFTDataBytes); err != nil {
		return nil, err
	}
	if err := ValidateURI(metadata.ImageUri, ""); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid data.image_uri: %v", err)
	}

	canonicalValue, err := metadata.Marshal()
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("encode data: %v", err)
	}
	if len(canonicalValue) > maxBasicNFTDataBytes {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf(
			"data must not exceed %d encoded bytes",
			maxBasicNFTDataBytes,
		)
	}
	if !bytes.Equal(data.Value, canonicalValue) {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(
			"data must use canonical deterministic protobuf encoding",
		)
	}

	return &cdctypes.Any{
		TypeUrl: BasicNFTDataTypeURL,
		Value:   canonicalValue,
	}, nil
}

// ValidateTransferPolicy rejects the protobuf default and unknown values.
func ValidateTransferPolicy(policy TransferPolicy) error {
	switch policy {
	case TransferPolicy_TRANSFER_POLICY_LOCKED,
		TransferPolicy_TRANSFER_POLICY_OWNER_TRANSFERABLE:
		return nil
	default:
		return sdkerrors.ErrInvalidRequest.Wrap("invalid transfer_policy")
	}
}

func validateText(field, value string, maxBytes int) error {
	if !utf8.ValidString(value) {
		return sdkerrors.ErrInvalidRequest.Wrapf("%s must be valid UTF-8", field)
	}
	if len(value) > maxBytes {
		return sdkerrors.ErrInvalidRequest.Wrapf("%s must not exceed %d bytes", field, maxBytes)
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return sdkerrors.ErrInvalidRequest.Wrapf("%s must not contain control characters", field)
		}
	}
	return nil
}
