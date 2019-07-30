package aol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisOwnerKey
type GenesisOwnerKey struct {
	OwnerAddress sdk.AccAddress `json:"owner_address"`
}

func (k GenesisOwnerKey) Marshal() string {
	return k.OwnerAddress.String()
}

func (k *GenesisOwnerKey) Unmarshal(key string) error {
	addr, err := sdk.AccAddressFromBech32(key)
	if err != nil {
		return err
	}
	k.OwnerAddress = addr
	return nil
}

// GenesisTopicKey
type GenesisTopicKey struct {
	OwnerAddress sdk.AccAddress `json:"owner_address"`
	TopicName    string         `json:"topic_name"`
}

func (k GenesisTopicKey) Marshal() string {
	return fmt.Sprintf("%s/%s", k.OwnerAddress, k.TopicName)
}

func (k *GenesisTopicKey) Unmarshal(key string) error {
	sp := strings.Split(key, "/")
	if len(sp) != 2 {
		return errors.New("invalid format for genesis topic key")
	}
	addr, err := sdk.AccAddressFromBech32(sp[0])
	if err != nil {
		return err
	}
	topic := sp[1]

	k.OwnerAddress = addr
	k.TopicName = topic
	return nil
}

// GenesisWriterKey
type GenesisWriterKey struct {
	OwnerAddress  sdk.AccAddress `json:"owner_address"`
	TopicName     string         `json:"topic_name"`
	WriterAddress sdk.AccAddress `json:"writer_address"`
}

func (k GenesisWriterKey) Marshal() string {
	return fmt.Sprintf("%s/%s/%s", k.OwnerAddress, k.TopicName, k.WriterAddress)
}

func (k *GenesisWriterKey) Unmarshal(key string) error {
	sp := strings.Split(key, "/")
	if len(sp) != 3 {
		return errors.New("invalid format for genesis writer key")
	}
	ownerAddr, err := sdk.AccAddressFromBech32(sp[0])
	if err != nil {
		return err
	}
	topic := sp[1]
	writerAddr, err := sdk.AccAddressFromBech32(sp[2])
	if err != nil {
		return err
	}

	k.OwnerAddress = ownerAddr
	k.TopicName = topic
	k.WriterAddress = writerAddr
	return nil
}

// GenesisRecordKey
type GenesisRecordKey struct {
	OwnerAddress sdk.AccAddress `json:"owner_address"`
	TopicName    string         `json:"topic_name"`
	Offset       uint64         `json:"offset"`
}

func (k GenesisRecordKey) Marshal() string {
	return fmt.Sprintf("%s/%s/%d", k.OwnerAddress, k.TopicName, k.Offset)
}

func (k *GenesisRecordKey) Unmarshal(key string) error {
	sp := strings.Split(key, "/")
	if len(sp) != 3 {
		return errors.New("invalid format for genesis record key")
	}
	ownerAddr, err := sdk.AccAddressFromBech32(sp[0])
	if err != nil {
		return err
	}
	topic := sp[1]
	offset, err := strconv.ParseUint(sp[2], 10, 64)
	if err != nil {
		return err
	}

	k.OwnerAddress = ownerAddr
	k.TopicName = topic
	k.Offset = offset
	return nil
}
