package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewMsgServiceMintPNFTRequest(
	denomId string,
	id string,
	name string,
	description string,
	uri string,
	uriHash string,
	creator string,
	data string,
) *MsgServiceMintPNFTRequest {
	return &MsgServiceMintPNFTRequest{
		DenomId:     denomId,
		Id:          id,
		Name:        name,
		Description: description,
		Uri:         uri,
		UriHash:     uriHash,
		Data:        data,
		Creator:     creator,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgServiceMintPNFTRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceMintPNFTRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgServiceMintPNFTRequest) ValidateBasic() error {
	if msg.DenomId == "" {
		return fmt.Errorf("denomId cannot be empty")
	}

	if msg.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if msg.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return err
	}

	return nil
}

func NewMsgServiceTransferPNFTRequest(
	denomId string,
	id string,
	sender string,
	receiver string,
) *MsgServiceTransferPNFTRequest {
	return &MsgServiceTransferPNFTRequest{
		DenomId:  denomId,
		Id:       id,
		Sender:   sender,
		Receiver: receiver,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgServiceTransferPNFTRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceTransferPNFTRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgServiceTransferPNFTRequest) ValidateBasic() error {
	if msg.DenomId == "" {
		return fmt.Errorf("denomId cannot be empty")
	}

	if msg.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if msg.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return err
	}

	if msg.Receiver == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Receiver); err != nil {
		return err
	}

	return nil
}

func NewMsgServiceBurnPNFTRequest(
	denomId string,
	id string,
	bunner string,
) *MsgServiceBurnPNFTRequest {
	return &MsgServiceBurnPNFTRequest{
		DenomId: denomId,
		Id:      id,
		Burner:  bunner,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgServiceBurnPNFTRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgServiceBurnPNFTRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Burner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgServiceBurnPNFTRequest) ValidateBasic() error {
	if msg.DenomId == "" {
		return fmt.Errorf("denomId cannot be empty")
	}

	if msg.Id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	if msg.Burner == "" {
		return fmt.Errorf("bunner cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Burner); err != nil {
		return err
	}

	return nil
}
