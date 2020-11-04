package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Issuance struct {
	Amount        sdk.Coin       `json:"amount"`
	IssuerAddress sdk.AccAddress `json:"issuer_address"`
}

func (i Issuance) Empty() bool {
	return i.IssuerAddress.Empty()
}

func (i Issuance) Valid() sdk.Error {
	if !i.Amount.IsValid() {
		return sdk.ErrInvalidCoins("amount is invalid: " + i.Amount.String())
	}
	if !i.Amount.IsPositive() {
		return sdk.ErrInsufficientCoins("amount must be positive")
	}
	if i.IssuerAddress.Empty() {
		return sdk.ErrInvalidAddress(i.IssuerAddress.String())
	}
	return nil
}

func (i Issuance) String() string {
	bz, _ := json.Marshal(i)
	return string(bz)
}
