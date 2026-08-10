package app

import (
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AccountAddressPrefix = "panacea"
)

var (
	AccountPubKeyPrefix    = AccountAddressPrefix + "pub"
	ValidatorAddressPrefix = AccountAddressPrefix + "valoper"
	ValidatorPubKeyPrefix  = AccountAddressPrefix + "valoperpub"
	ConsNodeAddressPrefix  = AccountAddressPrefix + "valcons"
	ConsNodePubKeyPrefix   = AccountAddressPrefix + "valconspub"
	setConfigOnce          sync.Once
)

func SetConfig() {
	setConfigOnce.Do(func() {
		config := sdk.GetConfig()
		config.SetPurpose(44)
		config.SetCoinType(371)
		config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountPubKeyPrefix)
		config.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorPubKeyPrefix)
		config.SetBech32PrefixForConsensusNode(ConsNodeAddressPrefix, ConsNodePubKeyPrefix)
		config.Seal()
	})
}
