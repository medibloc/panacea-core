// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/datapool/v2/genesis.proto

package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// GenesisState defines the datapool module's genesis state.
type GenesisState struct {
	DataValidators           []DataValidator          `protobuf:"bytes,1,rep,name=data_validators,json=dataValidators,proto3" json:"data_validators"`
	NextPoolNumber           uint64                   `protobuf:"varint,2,opt,name=next_pool_number,json=nextPoolNumber,proto3" json:"next_pool_number,omitempty"`
	Pools                    []Pool                   `protobuf:"bytes,3,rep,name=pools,proto3" json:"pools"`
	Params                   Params                   `protobuf:"bytes,4,opt,name=params,proto3" json:"params"`
	InstantRevenueDistribute InstantRevenueDistribute `protobuf:"bytes,5,opt,name=instant_revenue_distribute,json=instantRevenueDistribute,proto3" json:"instant_revenue_distribute"`
	SalesHistory             map[string]SalesHistory  `protobuf:"bytes,6,rep,name=sales_history,json=salesHistory,proto3" json:"sales_history" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_c5ac074515cc3a48, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetDataValidators() []DataValidator {
	if m != nil {
		return m.DataValidators
	}
	return nil
}

func (m *GenesisState) GetNextPoolNumber() uint64 {
	if m != nil {
		return m.NextPoolNumber
	}
	return 0
}

func (m *GenesisState) GetPools() []Pool {
	if m != nil {
		return m.Pools
	}
	return nil
}

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetInstantRevenueDistribute() InstantRevenueDistribute {
	if m != nil {
		return m.InstantRevenueDistribute
	}
	return InstantRevenueDistribute{}
}

func (m *GenesisState) GetSalesHistory() map[string]SalesHistory {
	if m != nil {
		return m.SalesHistory
	}
	return nil
}

// Params define parameters of datapool module
type Params struct {
	DataPoolDepositRate        github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,1,opt,name=data_pool_deposit_rate,json=dataPoolDepositRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"data_pool_deposit_rate" yaml:"data_pool_deposit_rate"`
	DataPoolCodeId             uint64                                 `protobuf:"varint,2,opt,name=data_pool_code_id,json=dataPoolCodeId,proto3" json:"data_pool_code_id,omitempty"`
	DataPoolNftContractAddress string                                 `protobuf:"bytes,3,opt,name=data_pool_nft_contract_address,json=dataPoolNftContractAddress,proto3" json:"data_pool_nft_contract_address,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_c5ac074515cc3a48, []int{1}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetDataPoolCodeId() uint64 {
	if m != nil {
		return m.DataPoolCodeId
	}
	return 0
}

func (m *Params) GetDataPoolNftContractAddress() string {
	if m != nil {
		return m.DataPoolNftContractAddress
	}
	return ""
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "panacea.datapool.v2.GenesisState")
	proto.RegisterMapType((map[string]SalesHistory)(nil), "panacea.datapool.v2.GenesisState.SalesHistoryEntry")
	proto.RegisterType((*Params)(nil), "panacea.datapool.v2.Params")
}

func init() { proto.RegisterFile("panacea/datapool/v2/genesis.proto", fileDescriptor_c5ac074515cc3a48) }

var fileDescriptor_c5ac074515cc3a48 = []byte{
	// 563 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x93, 0xcf, 0x6e, 0xd3, 0x4c,
	0x14, 0xc5, 0xe3, 0xe6, 0x8f, 0xf4, 0x4d, 0xfa, 0x95, 0xd6, 0x45, 0xc8, 0x04, 0xe1, 0xa4, 0x59,
	0xa0, 0xb0, 0x88, 0x2d, 0x52, 0x21, 0xa0, 0x3b, 0xd2, 0x20, 0xa8, 0x90, 0x0a, 0xb8, 0x12, 0x0b,
	0x16, 0x58, 0x63, 0xcf, 0x6d, 0x3a, 0xaa, 0xed, 0x31, 0x33, 0x93, 0xa8, 0xd9, 0xf3, 0x00, 0x3c,
	0x0c, 0x0f, 0xd1, 0x65, 0x97, 0x88, 0x45, 0x84, 0x92, 0x37, 0xe0, 0x01, 0x10, 0x9a, 0x19, 0x87,
	0x44, 0xc2, 0xac, 0x3c, 0xba, 0xf7, 0x77, 0xce, 0xf1, 0xcc, 0xdc, 0x41, 0x07, 0x39, 0xce, 0x70,
	0x0c, 0xd8, 0x27, 0x58, 0xe2, 0x9c, 0xb1, 0xc4, 0x9f, 0x0e, 0xfc, 0x31, 0x64, 0x20, 0xa8, 0xf0,
	0x72, 0xce, 0x24, 0xb3, 0xf7, 0x0b, 0xc4, 0x5b, 0x21, 0xde, 0x74, 0xd0, 0xba, 0x3d, 0x66, 0x63,
	0xa6, 0xfb, 0xbe, 0x5a, 0x19, 0xb4, 0xe5, 0x96, 0xb9, 0x69, 0x89, 0xee, 0x77, 0xbf, 0xd6, 0xd0,
	0xf6, 0x4b, 0x63, 0x7e, 0x26, 0xb1, 0x04, 0xfb, 0x1d, 0xba, 0xa5, 0xd0, 0x70, 0x8a, 0x13, 0x4a,
	0xb0, 0x64, 0x5c, 0x38, 0x56, 0xa7, 0xda, 0x6b, 0x0e, 0xba, 0x5e, 0x49, 0xaa, 0x37, 0xc2, 0x12,
	0xbf, 0x5f, 0xa1, 0xc3, 0xda, 0xf5, 0xbc, 0x5d, 0x09, 0x76, 0xc8, 0x66, 0x51, 0xd8, 0x3d, 0xb4,
	0x9b, 0xc1, 0x95, 0x0c, 0x95, 0x26, 0xcc, 0x26, 0x69, 0x04, 0xdc, 0xd9, 0xea, 0x58, 0xbd, 0x5a,
	0xb0, 0xa3, 0xea, 0x6f, 0x19, 0x4b, 0x4e, 0x75, 0xd5, 0x7e, 0x8c, 0xea, 0x0a, 0x12, 0x4e, 0x55,
	0x47, 0xde, 0x2d, 0x8d, 0x54, 0x7c, 0x91, 0x64, 0x68, 0xfb, 0x19, 0x6a, 0xe4, 0x98, 0xe3, 0x54,
	0x38, 0xb5, 0x8e, 0xd5, 0x6b, 0x0e, 0xee, 0x95, 0xeb, 0x34, 0x52, 0x28, 0x0b, 0x81, 0xfd, 0x09,
	0xb5, 0x68, 0x26, 0x24, 0xce, 0x64, 0xc8, 0x61, 0x0a, 0xd9, 0x04, 0x42, 0x42, 0x85, 0xe4, 0x34,
	0x9a, 0x48, 0x70, 0xea, 0xda, 0xae, 0x5f, 0x6a, 0x77, 0x62, 0x64, 0x81, 0x51, 0x8d, 0xfe, 0x88,
	0x8a, 0x00, 0x87, 0xfe, 0xa3, 0x6f, 0x7f, 0x44, 0xff, 0x0b, 0x9c, 0x80, 0x08, 0x2f, 0xa8, 0x90,
	0x8c, 0xcf, 0x9c, 0x86, 0xde, 0xec, 0x61, 0x69, 0xca, 0xe6, 0xdd, 0x78, 0x67, 0x4a, 0xf6, 0xca,
	0xa8, 0x5e, 0x64, 0x92, 0xcf, 0x8a, 0xac, 0x6d, 0xb1, 0xd1, 0x68, 0x45, 0x68, 0xef, 0x2f, 0xd0,
	0xde, 0x45, 0xd5, 0x4b, 0x98, 0x39, 0x56, 0xc7, 0xea, 0xfd, 0x17, 0xa8, 0xa5, 0xfd, 0x04, 0xd5,
	0xa7, 0x38, 0x99, 0x80, 0xbe, 0x8a, 0xe6, 0xe0, 0xa0, 0x34, 0x7e, 0xd3, 0x28, 0x30, 0xfc, 0xd1,
	0xd6, 0x53, 0xab, 0xfb, 0xcb, 0x42, 0x0d, 0x73, 0x9e, 0xf6, 0x67, 0x0b, 0xdd, 0xd1, 0x13, 0xa3,
	0xaf, 0x97, 0x40, 0xce, 0x04, 0x95, 0x21, 0xc7, 0x12, 0x4c, 0xda, 0xf0, 0x8d, 0xfa, 0xc7, 0xef,
	0xf3, 0xf6, 0x83, 0x31, 0x95, 0x17, 0x93, 0xc8, 0x8b, 0x59, 0xea, 0xc7, 0x4c, 0xa4, 0x4c, 0x14,
	0x9f, 0xbe, 0x20, 0x97, 0xbe, 0x9c, 0xe5, 0x20, 0xbc, 0x11, 0xc4, 0x3f, 0xe7, 0xed, 0xfb, 0x33,
	0x9c, 0x26, 0x47, 0xdd, 0x72, 0xd7, 0x6e, 0xb0, 0xaf, 0x1a, 0x6a, 0x0a, 0x46, 0xa6, 0x1c, 0xa8,
	0xb9, 0x7d, 0x88, 0xf6, 0xd6, 0x7c, 0xcc, 0x08, 0x84, 0x94, 0xac, 0xa6, 0x6c, 0xc5, 0x1f, 0x33,
	0x02, 0x27, 0xc4, 0x1e, 0x22, 0x77, 0x8d, 0x66, 0xe7, 0x32, 0x8c, 0x59, 0x26, 0x39, 0x8e, 0x65,
	0x88, 0x09, 0xe1, 0x20, 0xd4, 0xf8, 0xa9, 0x63, 0x6a, 0xad, 0x74, 0xa7, 0xe7, 0xf2, 0xb8, 0x40,
	0x9e, 0x1b, 0x62, 0xf8, 0xfa, 0x7a, 0xe1, 0x5a, 0x37, 0x0b, 0xd7, 0xfa, 0xb1, 0x70, 0xad, 0x2f,
	0x4b, 0xb7, 0x72, 0xb3, 0x74, 0x2b, 0xdf, 0x96, 0x6e, 0xe5, 0xc3, 0xa3, 0x8d, 0x6d, 0xa6, 0x40,
	0x68, 0x94, 0xb0, 0xd8, 0x2f, 0xce, 0xb6, 0x1f, 0x33, 0x0e, 0xfe, 0xd5, 0xfa, 0x31, 0xea, 0x5d,
	0x47, 0x0d, 0xfd, 0x16, 0x0f, 0x7f, 0x07, 0x00, 0x00, 0xff, 0xff, 0xec, 0xd5, 0x26, 0xdd, 0xfb,
	0x03, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SalesHistory) > 0 {
		for k := range m.SalesHistory {
			v := m.SalesHistory[k]
			baseI := i
			{
				size, err := (&v).MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintGenesis(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintGenesis(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x32
		}
	}
	{
		size, err := m.InstantRevenueDistribute.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if len(m.Pools) > 0 {
		for iNdEx := len(m.Pools) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Pools[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if m.NextPoolNumber != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.NextPoolNumber))
		i--
		dAtA[i] = 0x10
	}
	if len(m.DataValidators) > 0 {
		for iNdEx := len(m.DataValidators) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DataValidators[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.DataPoolNftContractAddress) > 0 {
		i -= len(m.DataPoolNftContractAddress)
		copy(dAtA[i:], m.DataPoolNftContractAddress)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.DataPoolNftContractAddress)))
		i--
		dAtA[i] = 0x1a
	}
	if m.DataPoolCodeId != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.DataPoolCodeId))
		i--
		dAtA[i] = 0x10
	}
	{
		size := m.DataPoolDepositRate.Size()
		i -= size
		if _, err := m.DataPoolDepositRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.DataValidators) > 0 {
		for _, e := range m.DataValidators {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.NextPoolNumber != 0 {
		n += 1 + sovGenesis(uint64(m.NextPoolNumber))
	}
	if len(m.Pools) > 0 {
		for _, e := range m.Pools {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.InstantRevenueDistribute.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.SalesHistory) > 0 {
		for k, v := range m.SalesHistory {
			_ = k
			_ = v
			l = v.Size()
			mapEntrySize := 1 + len(k) + sovGenesis(uint64(len(k))) + 1 + l + sovGenesis(uint64(l))
			n += mapEntrySize + 1 + sovGenesis(uint64(mapEntrySize))
		}
	}
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.DataPoolDepositRate.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if m.DataPoolCodeId != 0 {
		n += 1 + sovGenesis(uint64(m.DataPoolCodeId))
	}
	l = len(m.DataPoolNftContractAddress)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataValidators", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DataValidators = append(m.DataValidators, DataValidator{})
			if err := m.DataValidators[len(m.DataValidators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextPoolNumber", wireType)
			}
			m.NextPoolNumber = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NextPoolNumber |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pools", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Pools = append(m.Pools, Pool{})
			if err := m.Pools[len(m.Pools)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InstantRevenueDistribute", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.InstantRevenueDistribute.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SalesHistory", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.SalesHistory == nil {
				m.SalesHistory = make(map[string]SalesHistory)
			}
			var mapkey string
			mapvalue := &SalesHistory{}
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGenesis
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGenesis
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthGenesis
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthGenesis
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGenesis
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return ErrInvalidLengthGenesis
					}
					postmsgIndex := iNdEx + mapmsglen
					if postmsgIndex < 0 {
						return ErrInvalidLengthGenesis
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &SalesHistory{}
					if err := mapvalue.Unmarshal(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipGenesis(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthGenesis
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.SalesHistory[mapkey] = *mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataPoolDepositRate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DataPoolDepositRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataPoolCodeId", wireType)
			}
			m.DataPoolCodeId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.DataPoolCodeId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataPoolNftContractAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DataPoolNftContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
