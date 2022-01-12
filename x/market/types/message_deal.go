package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreateDeal{}

func NewMsgCreateDeal(dataSchema []string, budget *sdk.Coin, wantDataCount uint64, trustedDataValidator []string, owner string) *MsgCreateDeal {
	return &MsgCreateDeal{
		DataSchema:           dataSchema,
		Budget:               budget,
		WantDataCount:        wantDataCount,
		TrustedDataValidator: trustedDataValidator,
		Owner:                owner,
	}
}

func (msg *MsgCreateDeal) Route() string {
	return RouterKey
}

func (msg *MsgCreateDeal) Type() string {
	return "CreateDeal"
}

// ValidateBasic TODO: Validation for Create Deal Msg.
func (msg *MsgCreateDeal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Owner)

	if err != nil {
		return nil
	}
	return nil
}

func (msg *MsgCreateDeal) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateDeal) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
