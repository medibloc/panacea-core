package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/market/util/address"
)

func NewDealAddress(dealId uint64) sdk.AccAddress {
	key := append([]byte("deal"), sdk.Uint64ToBigEndian(dealId)...)
	return address.Module(ModuleName, key)
}
