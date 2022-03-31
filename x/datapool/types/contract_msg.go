package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/types/assets"
)

type mint struct {
	TokenId string   `json:"token_id"`
	Owner   string   `json:"owner"`
	Name    string   `json:"name"`
	Price   sdk.Coin `json:"price"`
}

type MsgMintNFT struct {
	mint `json:"mint"`
}

func NewMsgMintNFT(poolID uint64, owner string) *MsgMintNFT {
	tokenID := "data_pool_" + strconv.FormatUint(poolID, 10)
	zeroFund := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0))

	mint := &mint{
		TokenId: tokenID,
		Owner:   owner,
		Name:    "curator_nft",
		Price:   zeroFund,
	}

	return &MsgMintNFT{mint: *mint}
}

type InstantiateNFTMsg struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Minter string `json:"minter"`
}

func NewInstantiateNFTMsg(name, symbol, minterAddress string) *InstantiateNFTMsg {
	return &InstantiateNFTMsg{
		Name:   name,
		Symbol: symbol,
		Minter: minterAddress,
	}
}

type MigrateContractMsg struct {
	Payout sdk.AccAddress `json:"payout"`
}

func NewMigrateContractMsg(payout sdk.AccAddress) *MigrateContractMsg {
	return &MigrateContractMsg{
		Payout: payout,
	}
}
