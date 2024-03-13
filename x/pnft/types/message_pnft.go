package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewMsgMintPNFTRequest(
	denomId string,
	id string,
	name string,
	description string,
	uri string,
	uriHash string,
	creator string,
	data string,
) *MsgMintPNFTRequest {
	return &MsgMintPNFTRequest{
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
func (msg *MsgMintPNFTRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgMintPNFTRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgMintPNFTRequest) ValidateBasic() error {
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

func NewMsgTransferPNFTRequest(
	denomId string,
	id string,
	sender string,
	receiver string,
) *MsgTransferPNFTRequest {
	return &MsgTransferPNFTRequest{
		DenomId:  denomId,
		Id:       id,
		Sender:   sender,
		Receiver: receiver,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgTransferPNFTRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgTransferPNFTRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgTransferPNFTRequest) ValidateBasic() error {
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

func NewMsgBurnPNFTRequest(
	denomId string,
	id string,
	bunner string,
) *MsgBurnPNFTRequest {
	return &MsgBurnPNFTRequest{
		DenomId: denomId,
		Id:      id,
		Burner:  bunner,
	}
}

// GetSignBytes returns a byte array which is used to generate a signature for verifying DID ownership.
func (msg *MsgBurnPNFTRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgBurnPNFTRequest) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.Burner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgBurnPNFTRequest) ValidateBasic() error {
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
