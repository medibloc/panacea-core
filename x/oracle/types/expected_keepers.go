package types

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountKeeper interface {
	GetPubKey(sdk.Context, sdk.AccAddress) (cryptotypes.PubKey, error)
}
