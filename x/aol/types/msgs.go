package types

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgCreateTopic{}
	_ sdk.Msg = &MsgAddWriter{}
	_ sdk.Msg = &MsgDeleteWriter{}
	_ sdk.Msg = &MsgAddRecord{}
)

const (
	MaxTopicLength       = 70
	MaxMonikerLength     = 70
	MaxDescriptionLength = 5000
	MaxRecordKeyLength   = 70
	MaxRecordValueLength = 5000
)

// MsgCreateTopic
type MsgCreateTopic struct {
	TopicName    string         `json:"topic_name"`
	Description  string         `json:"description"`
	OwnerAddress sdk.AccAddress `json:"owner_address"`
}

func (msg MsgCreateTopic) Route() string { return RouterKey }

func (msg MsgCreateTopic) Type() string { return "create_topic" }

func (msg MsgCreateTopic) ValidateBasic() sdk.Error {
	if err := validateTopic(msg.TopicName); err != nil {
		return err
	}
	if err := validateDescription(msg.Description); err != nil {
		return err
	}
	if msg.OwnerAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.OwnerAddress.String())
	}
	return nil
}

func (msg MsgCreateTopic) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCreateTopic) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.OwnerAddress}
}

// MsgAddWriter
type MsgAddWriter struct {
	TopicName     string         `json:"topic_name"`
	Moniker       string         `json:"moniker"`
	Description   string         `json:"description"`
	WriterAddress sdk.AccAddress `json:"writer_address"`
	OwnerAddress  sdk.AccAddress `json:"owner_address"`
}

func (msg MsgAddWriter) Route() string { return RouterKey }

func (msg MsgAddWriter) Type() string { return "add_writer" }

func (msg MsgAddWriter) ValidateBasic() sdk.Error {
	if err := validateTopic(msg.TopicName); err != nil {
		return err
	}
	if err := validateMoniker(msg.Moniker); err != nil {
		return err
	}
	if err := validateDescription(msg.Description); err != nil {
		return err
	}
	if msg.WriterAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.WriterAddress.String())
	}
	if msg.OwnerAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.OwnerAddress.String())
	}
	return nil
}

func (msg MsgAddWriter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgAddWriter) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.OwnerAddress}
}

// MsgDeleteWriter
type MsgDeleteWriter struct {
	TopicName     string         `json:"topic_name"`
	WriterAddress sdk.AccAddress `json:"writer_address"`
	OwnerAddress  sdk.AccAddress `json:"owner_address"`
}

func (msg MsgDeleteWriter) Route() string { return RouterKey }

func (msg MsgDeleteWriter) Type() string { return "delete_writer" }

func (msg MsgDeleteWriter) ValidateBasic() sdk.Error {
	if err := validateTopic(msg.TopicName); err != nil {
		return err
	}
	if msg.WriterAddress.Empty() {
		// TODO Error Message
		return sdk.ErrInvalidAddress(msg.WriterAddress.String())
	}
	if msg.OwnerAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.OwnerAddress.String())
	}
	return nil
}

func (msg MsgDeleteWriter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgDeleteWriter) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.OwnerAddress}
}

// MsgAddRecord
type MsgAddRecord struct {
	TopicName       string         `json:"topic_name"`
	Key             []byte         `json:"key"`
	Value           []byte         `json:"value"`
	WriterAddress   sdk.AccAddress `json:"writer_address"`
	OwnerAddress    sdk.AccAddress `json:"owner_address"`
	FeePayerAddress sdk.AccAddress `json:"fee_payer_address"`
}

func (msg MsgAddRecord) Route() string { return RouterKey }

func (msg MsgAddRecord) Type() string { return "add_record" }

func (msg MsgAddRecord) ValidateBasic() sdk.Error {
	if err := validateTopic(msg.TopicName); err != nil {
		return err
	}
	if err := validateRecordKey(msg.Key); err != nil {
		return err
	}
	if err := validateRecordValue(msg.Value); err != nil {
		return err
	}
	if msg.WriterAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.WriterAddress.String())
	}
	if msg.OwnerAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.OwnerAddress.String())
	}
	return nil
}

func (msg MsgAddRecord) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgAddRecord) GetSigners() []sdk.AccAddress {
	if msg.FeePayerAddress.Empty() {
		return []sdk.AccAddress{msg.WriterAddress}
	}
	return []sdk.AccAddress{msg.FeePayerAddress, msg.WriterAddress}
}

func validateTopic(topic string) sdk.Error {
	if len(topic) > MaxTopicLength {
		return ErrMessageTooLarge("topic", len(topic), MaxTopicLength)
	}

	// cannot be an empty string
	if !regexp.MustCompile("^[A-Za-z0-9._-]+$").MatchString(topic) {
		return ErrInvalidTopic(topic)
	}

	return nil
}

func validateMoniker(moniker string) sdk.Error {
	if len(moniker) > MaxMonikerLength {
		return ErrMessageTooLarge("moniker", len(moniker), MaxMonikerLength)
	}

	// can be an empty string
	if !regexp.MustCompile("^[A-Za-z0-9._-]*$").MatchString(moniker) {
		return ErrInvalidMoniker(moniker)
	}

	return nil
}

func validateDescription(description string) sdk.Error {
	if len(description) > MaxDescriptionLength {
		return ErrMessageTooLarge("description", len(description), MaxDescriptionLength)
	}
	return nil
}

func validateRecordKey(key []byte) sdk.Error {
	if len(key) > MaxRecordKeyLength {
		return ErrMessageTooLarge("key", len(key), MaxRecordKeyLength)
	}
	return nil
}

func validateRecordValue(value []byte) sdk.Error {
	if len(value) > MaxRecordValueLength {
		return ErrMessageTooLarge("value", len(value), MaxRecordValueLength)
	}
	return nil
}
