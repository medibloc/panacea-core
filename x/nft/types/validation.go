package types

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	maxLocalClassIDBytes     = 64
	maxFullClassIDBytes      = 131
	maxClassNameBytes        = 128
	maxClassSymbolBytes      = 32
	maxClassDescriptionBytes = 1024
	maxURIBytes              = 256
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
