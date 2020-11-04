package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	_ sdk.Msg = &MsgIssueCurrency{}
)

// MsgIssueCurrency defines a CreateCurrency message.
type MsgIssueCurrency struct {
	Issuance
}

// NewMsgIssueCurrency is a constructor of MsgIssueCurrency.
func NewMsgIssueCurrency(amount sdk.Coin, issuerAddress sdk.AccAddress) MsgIssueCurrency {
	return MsgIssueCurrency{Issuance{
		Amount:        amount,
		IssuerAddress: issuerAddress,
	}}
}

// Route returns the name of the module.
func (msg MsgIssueCurrency) Route() string {
	return RouterKey
}

// Type returns the name of the action.
func (msg MsgIssueCurrency) Type() string {
	return "issue_currency"
}

// ValidateBasic runs stateless checks on the message.
func (msg MsgIssueCurrency) ValidateBasic() sdk.Error {
	return msg.Valid()
}

// GetSignBytes returns the canonical byte representation of the message. Used to generate a signature.
func (msg MsgIssueCurrency) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners return the addresses of signers that must sign.
func (msg MsgIssueCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.IssuerAddress}
}
