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

type MsgMintNft struct {
	mint `json:"mint"`
}

func NewMsgMintNft(poolID uint64, owner string) *MsgMintNft {
	tokenID := "data_pool_" + strconv.FormatUint(poolID, 10)
	zeroFund := sdk.NewCoin(assets.MicroMedDenom, sdk.NewInt(0))

	mint := &mint{
		TokenId: tokenID,
		Owner:   owner,
		Name:    "curator_nft",
		Price:   zeroFund,
	}

	return &MsgMintNft{mint: *mint}
}

type InstantiateNftMsg struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Minter string `json:"minter"`
}

func NewInstantiateNftMsg(name, symbol, minterAddress string) *InstantiateNftMsg {
	return &InstantiateNftMsg{
		Name:   name,
		Symbol: symbol,
		Minter: minterAddress,
	}
}
