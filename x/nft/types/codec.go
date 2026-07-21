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

// UnpackInterfaces resolves nested NFT metadata during transaction decoding.
func (m *MsgMintRequest) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if m == nil || m.Data == nil {
		return nil
	}
	return unpackNFTData(unpacker, m.Data)
}

// UnpackInterfaces resolves metadata restored from the policy store or
// genesis without relying on an Any cached value.
func (m *BurnTombstone) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if m == nil || m.Data == nil {
		return nil
	}
	return unpackNFTData(unpacker, m.Data)
}

// UnpackInterfaces resolves metadata nested in a live combined record.
func (m *LiveNFTRecord) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if m == nil || m.Nft == nil || m.Nft.Data == nil {
		return nil
	}
	return unpackNFTData(unpacker, m.Nft.Data)
}

// UnpackInterfaces resolves metadata in either NFTRecord variant.
func (m *NFTRecord) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if m == nil {
		return nil
	}
	if live := m.GetLive(); live != nil {
		return live.UnpackInterfaces(unpacker)
	}
	if tombstone := m.GetBurnTombstone(); tombstone != nil {
		return tombstone.UnpackInterfaces(unpacker)
	}
	return nil
}

// UnpackInterfaces resolves metadata in a point-query response.
func (m *QueryNFTRecordResponse) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if m == nil || m.NftRecord == nil {
		return nil
	}
	return m.NftRecord.UnpackInterfaces(unpacker)
}

// UnpackInterfaces resolves metadata in live list-query responses.
func (m *QueryNFTRecordsResponse) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if m == nil {
		return nil
	}
	for _, record := range m.NftRecords {
		if err := record.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

// UnpackInterfaces resolves all NFT metadata nested in combined genesis.
func (m *GenesisState) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if m == nil {
		return nil
	}
	if m.NftState != nil {
		for _, entry := range m.NftState.Entries {
			if entry == nil {
				continue
			}
			for _, token := range entry.Nfts {
				if token != nil && token.Data != nil {
					if err := unpackNFTData(unpacker, token.Data); err != nil {
						return err
					}
				}
			}
		}
	}
	for _, tombstone := range m.Tombstones {
		if tombstone != nil {
			if err := tombstone.UnpackInterfaces(unpacker); err != nil {
				return err
			}
		}
	}
	return nil
}

func unpackNFTData(unpacker cdctypes.AnyUnpacker, packed *cdctypes.Any) error {
	var data NFTData
	return unpacker.UnpackAny(packed, &data)
}

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
