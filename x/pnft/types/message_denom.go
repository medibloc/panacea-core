package types

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewMsgCreateDenomRequest(
	id string,
	symbol string,
	name string,
	description string,
	uri string,
	uriHash string,
	creator string,
	data string,
) *MsgCreateDenomRequest {
	return &MsgCreateDenomRequest{
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
func (msg *MsgCreateDenomRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateDenomRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgCreateDenomRequest) ValidateBasic() error {
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

func NewMsgUpdateDenomRequest(
	id string,
	symbol string,
	name string,
	description string,
	uri string,
	uriHash string,
	data string,
	update string,
) *MsgUpdateDenomRequest {
	return &MsgUpdateDenomRequest{
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
func (msg *MsgUpdateDenomRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateDenomRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Updater)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUpdateDenomRequest) ValidateBasic() error {
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

func NewMsgDeleteDenomRequest(
	id string,
	remover string,
) *MsgDeleteDenomRequest {
	return &MsgDeleteDenomRequest{
		Id:      id,
		Remover: remover,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgDeleteDenomRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeleteDenomRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Remover)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgDeleteDenomRequest) ValidateBasic() error {
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

func NewMsgTransferRequest(
	id string,
	sender string,
	receiver string,
) *MsgTransferDenomRequest {
	return &MsgTransferDenomRequest{
		Id:       id,
		Sender:   sender,
		Receiver: receiver,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgTransferDenomRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgTransferDenomRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgTransferDenomRequest) ValidateBasic() error {
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
