package types

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/types/compkey"
)

const (
	// ModuleName defines the module name
	ModuleName = "aol"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_capability"

	// this line is used by starport scaffolding # ibc/keys/name
)

// this line is used by starport scaffolding # ibc/keys/port

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	OwnerKey  = "Owner-value-"
	TopicKey  = "Topic-value-"
	WriterKey = "Writer-value-"
	RecordKey = "Record-value-"
)

var (
	_ compkey.CompositeKey = &OwnerCompositeKey{}
	_ compkey.CompositeKey = &TopicCompositeKey{}
	_ compkey.CompositeKey = &WriterCompositeKey{}
	_ compkey.CompositeKey = &RecordCompositeKey{}
)

type OwnerCompositeKey struct {
	OwnerAddress sdk.AccAddress
}

func (k OwnerCompositeKey) ByteSlices() [][]byte {
	return [][]byte{k.OwnerAddress.Bytes()}
}

func (k *OwnerCompositeKey) FromByteSlices(bzs [][]byte) error {
	if len(bzs) != 1 {
		return fmt.Errorf("invalid input length")
	}
	if err := sdk.VerifyAddressFormat(bzs[0]); err != nil {
		return fmt.Errorf("invalid account address bytes: %w", err)
	}

	k.OwnerAddress = bzs[0]
	return nil
}

func (k OwnerCompositeKey) Strings() []string {
	return []string{k.OwnerAddress.String()}
}

func (k *OwnerCompositeKey) FromStrings(strings []string) error {
	if len(strings) != 1 {
		return fmt.Errorf("invalid input length")
	}

	addr, err := sdk.AccAddressFromBech32(strings[0])
	if err != nil {
		return fmt.Errorf("invalid account address string: %w", err)
	}

	k.OwnerAddress = addr
	return nil
}

type TopicCompositeKey struct {
	OwnerAddress sdk.AccAddress
	TopicName    string
}

func (k TopicCompositeKey) ByteSlices() [][]byte {
	return [][]byte{
		k.OwnerAddress.Bytes(),
		[]byte(k.TopicName),
	}
}

func (k *TopicCompositeKey) FromByteSlices(bzs [][]byte) error {
	if len(bzs) != 2 {
		return fmt.Errorf("invalid input length")
	}
	if err := sdk.VerifyAddressFormat(bzs[0]); err != nil {
		return fmt.Errorf("invalid account address bytes: %w", err)
	}

	k.OwnerAddress = bzs[0]
	k.TopicName = string(bzs[1])
	return nil
}

func (k TopicCompositeKey) Strings() []string {
	return []string{
		k.OwnerAddress.String(),
		k.TopicName,
	}
}

func (k *TopicCompositeKey) FromStrings(strings []string) error {
	if len(strings) != 2 {
		return fmt.Errorf("invalid input length")
	}

	addr, err := sdk.AccAddressFromBech32(strings[0])
	if err != nil {
		return fmt.Errorf("invalid account address string: %w", err)
	}

	k.OwnerAddress = addr
	k.TopicName = strings[1]
	return nil
}

type WriterCompositeKey struct {
	OwnerAddress  sdk.AccAddress
	TopicName     string
	WriterAddress sdk.AccAddress
}

func (k WriterCompositeKey) ByteSlices() [][]byte {
	return [][]byte{
		k.OwnerAddress.Bytes(),
		[]byte(k.TopicName),
		k.WriterAddress.Bytes(),
	}
}

func (k *WriterCompositeKey) FromByteSlices(bzs [][]byte) error {
	if len(bzs) != 3 {
		return fmt.Errorf("invalid input length")
	}
	if err := sdk.VerifyAddressFormat(bzs[0]); err != nil {
		return fmt.Errorf("invalid account address bytes: %w", err)
	}
	if err := sdk.VerifyAddressFormat(bzs[2]); err != nil {
		return fmt.Errorf("invalid account address bytes: %w", err)
	}

	k.OwnerAddress = bzs[0]
	k.TopicName = string(bzs[1])
	k.WriterAddress = bzs[2]
	return nil
}

func (k WriterCompositeKey) Strings() []string {
	return []string{
		k.OwnerAddress.String(),
		k.TopicName,
		k.WriterAddress.String(),
	}
}

func (k *WriterCompositeKey) FromStrings(strings []string) error {
	if len(strings) != 3 {
		return fmt.Errorf("invalid input length")
	}

	ownerAddr, err := sdk.AccAddressFromBech32(strings[0])
	if err != nil {
		return fmt.Errorf("invalid account address string: %w", err)
	}
	writerAddr, err := sdk.AccAddressFromBech32(strings[2])
	if err != nil {
		return fmt.Errorf("invalid account address string: %w", err)
	}

	k.OwnerAddress = ownerAddr
	k.TopicName = strings[1]
	k.WriterAddress = writerAddr
	return nil
}

type RecordCompositeKey struct {
	OwnerAddress sdk.AccAddress
	TopicName    string
	Offset       uint64
}

func (k RecordCompositeKey) ByteSlices() [][]byte {
	return [][]byte{
		k.OwnerAddress.Bytes(),
		[]byte(k.TopicName),
		sdk.Uint64ToBigEndian(k.Offset),
	}
}

func (k *RecordCompositeKey) FromByteSlices(bzs [][]byte) error {
	if len(bzs) != 3 {
		return fmt.Errorf("invalid input length")
	}
	if err := sdk.VerifyAddressFormat(bzs[0]); err != nil {
		return fmt.Errorf("invalid account address bytes: %w", err)
	}

	k.OwnerAddress = bzs[0]
	k.TopicName = string(bzs[1])
	k.Offset = sdk.BigEndianToUint64(bzs[2])
	return nil
}

func (k RecordCompositeKey) Strings() []string {
	return []string{
		k.OwnerAddress.String(),
		k.TopicName,
		strconv.FormatUint(k.Offset, 10),
	}
}

func (k *RecordCompositeKey) FromStrings(strings []string) error {
	if len(strings) != 3 {
		return fmt.Errorf("invalid input length")
	}

	ownerAddr, err := sdk.AccAddressFromBech32(strings[0])
	if err != nil {
		return fmt.Errorf("invalid account address string: %w", err)
	}
	offset, err := strconv.ParseUint(strings[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid offset string: %w", err)
	}

	k.OwnerAddress = ownerAddr
	k.TopicName = strings[1]
	k.Offset = offset
	return nil
}
