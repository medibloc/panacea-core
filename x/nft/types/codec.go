package types

import (
	upstreamnft "cosmossdk.io/x/nft"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/proto"
)

// NFTData is the closed set of metadata types accepted by the Panacea NFT
// module. Keeping it separate from sdk.Msg prevents metadata from becoming an
// executable transaction message merely to make it resolvable from Any.
type NFTData interface {
	proto.Message
	isNFTData()
}

func (*BasicNFTData) isNFTData() {}

// RegisterInterfaces registers the standard and Panacea NFT wire types.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	upstreamnft.RegisterInterfaces(registry)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateClassRequest{},
		&MsgUpdateControllerRequest{},
		&MsgMintRequest{},
		&MsgRevokeRequest{},
		&MsgBurnRequest{},
	)
	registry.RegisterInterface(NFTDataInterfaceName, (*NFTData)(nil), &BasicNFTData{})

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
