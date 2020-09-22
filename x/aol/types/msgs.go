package types

import (
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
	if len(msg.TopicName) > MaxTopicLength {
		return ErrMessageTooLarge("topic_name", len(msg.TopicName), MaxTopicLength)
	}
	if len(msg.Description) > MaxDescriptionLength {
		return ErrMessageTooLarge("description", len(msg.Description), MaxDescriptionLength)
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
	if len(msg.TopicName) > MaxTopicLength {
		return ErrMessageTooLarge("topic_name", len(msg.TopicName), MaxTopicLength)
	}
	if len(msg.Moniker) > MaxMonikerLength {
		return ErrMessageTooLarge("moniker", len(msg.Moniker), MaxMonikerLength)
	}
	if len(msg.Description) > MaxDescriptionLength {
		return ErrMessageTooLarge("description", len(msg.Description), MaxDescriptionLength)
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
	// TODO Empty Topic error
	if len(msg.TopicName) > MaxTopicLength {
		return ErrMessageTooLarge("topic_name", len(msg.TopicName), MaxTopicLength)
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
	if len(msg.TopicName) > MaxTopicLength {
		return ErrMessageTooLarge("topic", len(msg.TopicName), MaxTopicLength)
	}
	if len(msg.Key) > MaxRecordKeyLength {
		return ErrMessageTooLarge("key", len(msg.Key), MaxRecordKeyLength)
	}
	if len(msg.Value) > MaxRecordValueLength {
		return ErrMessageTooLarge("value", len(msg.Value), MaxRecordValueLength)
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
