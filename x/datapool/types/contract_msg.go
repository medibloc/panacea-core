package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

func NewMsgMintCuratorNFT(poolID uint64, owner string) *MsgMintNFT {
	mint := &mint{
		TokenId: strconv.FormatUint(poolID, 10),
		Owner:   owner,
		Name:    "curator_nft",
		Price:   ZeroFund,
	}

	return &MsgMintNFT{mint: *mint}
}

func NewMsgMintDataAccessNFT(numNFT uint64, owner string) *MsgMintNFT {
	mint := &mint{
		TokenId: strconv.FormatUint(numNFT, 10),
		Owner:   owner,
		Name:    "data_access_nft",
		Price:   ZeroFund,
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
