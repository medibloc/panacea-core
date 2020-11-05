package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgIssueToken{}
)

// MsgIssueToken defines a IssueToken message.
type MsgIssueToken struct {
	Name         Name           `json:"name"`
	Symbol       ShortSymbol    `json:"symbol"`
	TotalSupply  sdk.Int        `json:"total_supply"`
	Mintable     bool           `json:"mintable"`
	OwnerAddress sdk.AccAddress `json:"owner_address"`
}

// NewMsgIssueToken is a constructor of MsgIssueToken.
func NewMsgIssueToken(name Name, symbol ShortSymbol, totalSupply sdk.Int, mintable bool, ownerAddress sdk.AccAddress) MsgIssueToken {
	return MsgIssueToken{
		Name:         name,
		Symbol:       symbol,
		TotalSupply:  totalSupply,
		Mintable:     mintable,
		OwnerAddress: ownerAddress,
	}
}

// Route returns the name of the module.
func (msg MsgIssueToken) Route() string {
	return RouterKey
}

// Type returns the name of the action.
func (msg MsgIssueToken) Type() string {
	return "issue_token"
}

// ValidateBasic runs stateless checks on the message.
func (msg MsgIssueToken) ValidateBasic() sdk.Error {
	if err := msg.Name.validate(); err != nil {
		return err
	}
	if err := msg.Symbol.validate(); err != nil {
		return err
	}
	if err := validateTotalSupplyAmount(msg.TotalSupply); err != nil {
		return err
	}
	if msg.OwnerAddress.Empty() {
		return sdk.ErrInvalidAddress(msg.OwnerAddress.String())
	}
	return nil
}

// GetSignBytes returns the canonical byte representation of the message. Used to generate a signature.
func (msg MsgIssueToken) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg MsgIssueToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.OwnerAddress}
}
