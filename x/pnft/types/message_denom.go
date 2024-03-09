package types

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewMsgServiceCreateDenomRequest(
	id string,
	symbol string,
	name string,
	description string,
	uri string,
	uriHash string,
	creator string,
	data string,
) *MsgServiceCreateDenomRequest {
	return &MsgServiceCreateDenomRequest{
		Id:          id,
		Name:        name,
		Symbol:      symbol,
		Description: description,
		Uri:         uri,
		UriHash:     uriHash,
		Data:        data,
		Creator:     creator,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgServiceCreateDenomRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceCreateDenomRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgServiceCreateDenomRequest) ValidateBasic() error {
	if msg.Id == "" {
		return errors.New("id cannot be empty")
	}

	if msg.Name == "" {
		return errors.New("name cannot be empty")
	}

	if msg.Symbol == "" {
		return errors.New("symbol cannot be empty")
	}

	if msg.Creator == "" {
		return errors.New("creator cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return err
	}

	return nil
}

func NewMsgServiceUpdateDenomRequest(
	id string,
	symbol string,
	name string,
	description string,
	uri string,
	uriHash string,
	data string,
	update string,
) *MsgServiceUpdateDenomRequest {
	return &MsgServiceUpdateDenomRequest{
		Id:          id,
		Name:        name,
		Symbol:      symbol,
		Description: description,
		Uri:         uri,
		UriHash:     uriHash,
		Data:        data,
		Updater:     update,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgServiceUpdateDenomRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceUpdateDenomRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Updater)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgServiceUpdateDenomRequest) ValidateBasic() error {
	if msg.Id == "" {
		return errors.New("Id cannot be empty")
	}

	if msg.Updater == "" {
		return errors.New("updater cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Updater); err != nil {
		return err
	}
	return nil
}

func NewMsgServiceDeleteDenomRequest(
	id string,
	remover string,
) *MsgServiceDeleteDenomRequest {
	return &MsgServiceDeleteDenomRequest{
		Id:      id,
		Remover: remover,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgServiceDeleteDenomRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceDeleteDenomRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Remover)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgServiceDeleteDenomRequest) ValidateBasic() error {
	if msg.Id == "" {
		return errors.New("id cannot be empty")
	}

	if msg.Remover == "" {
		return errors.New("remover cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Remover); err != nil {
		return err
	}

	return nil
}

func NewMsgServiceTransferRequest(
	id string,
	sender string,
	receiver string,
) *MsgServiceTransferDenomRequest {
	return &MsgServiceTransferDenomRequest{
		Id:       id,
		Sender:   sender,
		Receiver: receiver,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgServiceTransferDenomRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceTransferDenomRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgServiceTransferDenomRequest) ValidateBasic() error {
	if msg.Id == "" {
		return errors.New("id cannot be empty")
	}

	if msg.Sender == "" {
		return errors.New("sender cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}

	if msg.Receiver == "" {
		return errors.New("receiver cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Receiver); err != nil {
		return err
	}

	return nil
}
