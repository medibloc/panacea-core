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
	DataValidators []DataValidator `protobuf:"bytes,1,rep,name=data_validators,json=dataValidators,proto3" json:"data_validators"`
	NextPoolNumber uint64          `protobuf:"varint,2,opt,name=next_pool_number,json=nextPoolNumber,proto3" json:"next_pool_number,omitempty"`
	Pools          []Pool          `protobuf:"bytes,3,rep,name=pools,proto3" json:"pools"`
	Params         Params          `protobuf:"bytes,4,opt,name=params,proto3" json:"params"`
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

// Params define parameters of datapool module
type Params struct {
	DataPoolCommissionRate     github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,1,opt,name=data_pool_commission_rate,json=dataPoolCommissionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"data_pool_commission_rate" yaml:"data_pool_commission_rate"`
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
	proto.RegisterType((*Params)(nil), "panacea.datapool.v2.Params")
}

func init() { proto.RegisterFile("panacea/datapool/v2/genesis.proto", fileDescriptor_c5ac074515cc3a48) }

var fileDescriptor_c5ac074515cc3a48 = []byte{
	// 441 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x92, 0x41, 0x6b, 0xd4, 0x40,
	0x14, 0xc7, 0x77, 0xda, 0x75, 0xa1, 0x53, 0xa9, 0x1a, 0x45, 0xd2, 0x15, 0xb2, 0x31, 0x07, 0x89,
	0x87, 0x26, 0xb8, 0xe2, 0x41, 0x6f, 0x6e, 0x0b, 0x22, 0x42, 0xd1, 0x08, 0x1e, 0xbc, 0x84, 0xb7,
	0x33, 0xd3, 0x35, 0x98, 0xc9, 0x0b, 0x33, 0xd3, 0xa5, 0xfd, 0x04, 0x9e, 0x04, 0x3f, 0x56, 0x8f,
	0x3d, 0x8a, 0x87, 0x45, 0x76, 0xbf, 0x81, 0x77, 0x41, 0x66, 0x92, 0x75, 0xf7, 0x90, 0x9e, 0x12,
	0xe6, 0xff, 0xfb, 0xbd, 0x97, 0xf7, 0x32, 0xf4, 0x71, 0x0d, 0x15, 0x30, 0x01, 0x29, 0x07, 0x03,
	0x35, 0x62, 0x99, 0xce, 0xc7, 0xe9, 0x4c, 0x54, 0x42, 0x17, 0x3a, 0xa9, 0x15, 0x1a, 0xf4, 0xee,
	0xb7, 0x48, 0xb2, 0x46, 0x92, 0xf9, 0x78, 0xf8, 0x60, 0x86, 0x33, 0x74, 0x79, 0x6a, 0xdf, 0x1a,
	0x74, 0x18, 0x74, 0x55, 0x73, 0x8a, 0xcb, 0xa3, 0xbf, 0x84, 0xde, 0x7e, 0xd3, 0x14, 0xff, 0x68,
	0xc0, 0x08, 0xef, 0x03, 0xbd, 0x63, 0xd1, 0x7c, 0x0e, 0x65, 0xc1, 0xc1, 0xa0, 0xd2, 0x3e, 0x09,
	0x77, 0xe3, 0xfd, 0x71, 0x94, 0x74, 0x74, 0x4d, 0x4e, 0xc0, 0xc0, 0xa7, 0x35, 0x3a, 0xe9, 0x5f,
	0x2d, 0x46, 0xbd, 0xec, 0x80, 0x6f, 0x1f, 0x6a, 0x2f, 0xa6, 0x77, 0x2b, 0x71, 0x61, 0x72, 0xeb,
	0xe4, 0xd5, 0xb9, 0x9c, 0x0a, 0xe5, 0xef, 0x84, 0x24, 0xee, 0x67, 0x07, 0xf6, 0xfc, 0x3d, 0x62,
	0x79, 0xea, 0x4e, 0xbd, 0x17, 0xf4, 0x96, 0x85, 0xb4, 0xbf, 0xeb, 0x5a, 0x1e, 0x76, 0xb6, 0xb4,
	0x7c, 0xdb, 0xa9, 0xa1, 0xbd, 0x97, 0x74, 0x50, 0x83, 0x02, 0xa9, 0xfd, 0x7e, 0x48, 0xe2, 0xfd,
	0xf1, 0xa3, 0x6e, 0xcf, 0x21, 0xad, 0xd9, 0x0a, 0xd1, 0xb7, 0x1d, 0x3a, 0x68, 0x02, 0xef, 0x3b,
	0xa1, 0x87, 0x6e, 0x74, 0xf7, 0x9d, 0x0c, 0xa5, 0x2c, 0xb4, 0x2e, 0xb0, 0xca, 0x15, 0x18, 0xe1,
	0x93, 0x90, 0xc4, 0x7b, 0x93, 0xcc, 0xca, 0xbf, 0x16, 0xa3, 0x27, 0xb3, 0xc2, 0x7c, 0x39, 0x9f,
	0x26, 0x0c, 0x65, 0xca, 0x50, 0x4b, 0xd4, 0xed, 0xe3, 0x48, 0xf3, 0xaf, 0xa9, 0xb9, 0xac, 0x85,
	0x4e, 0x4e, 0x04, 0xfb, 0xb3, 0x18, 0x85, 0x97, 0x20, 0xcb, 0x57, 0xd1, 0x8d, 0x85, 0xa3, 0xec,
	0xa1, 0xcd, 0xec, 0x50, 0xc7, 0xff, 0x93, 0xcc, 0xfe, 0x89, 0xa7, 0xf4, 0xde, 0xb6, 0xc5, 0x45,
	0x5e, 0xf0, 0xf5, 0xde, 0x36, 0x0a, 0x17, 0x6f, 0xb9, 0x37, 0xa1, 0xc1, 0x06, 0xad, 0xce, 0x4c,
	0xce, 0xb0, 0x32, 0x0a, 0x98, 0xc9, 0x81, 0x73, 0x25, 0xb4, 0x5d, 0x28, 0x89, 0xf7, 0xb2, 0xe1,
	0xda, 0x3b, 0x3d, 0x33, 0xc7, 0x2d, 0xf2, 0xba, 0x21, 0x26, 0xef, 0xae, 0x96, 0x01, 0xb9, 0x5e,
	0x06, 0xe4, 0xf7, 0x32, 0x20, 0x3f, 0x56, 0x41, 0xef, 0x7a, 0x15, 0xf4, 0x7e, 0xae, 0x82, 0xde,
	0xe7, 0x67, 0x5b, 0xc3, 0x4a, 0xc1, 0x8b, 0x69, 0x89, 0x2c, 0x6d, 0x37, 0x7c, 0xc4, 0x50, 0x89,
	0xf4, 0x62, 0x73, 0xbd, 0xdc, 0xec, 0xd3, 0x81, 0xbb, 0x5d, 0xcf, 0xff, 0x05, 0x00, 0x00, 0xff,
	0xff, 0xc9, 0x22, 0x67, 0x8c, 0xcd, 0x02, 0x00, 0x00,
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
		size := m.DataPoolCommissionRate.Size()
		i -= size
		if _, err := m.DataPoolCommissionRate.MarshalTo(dAtA[i:]); err != nil {
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
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.DataPoolCommissionRate.Size()
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
				return fmt.Errorf("proto: wrong wireType = %d for field DataPoolCommissionRate", wireType)
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
			if err := m.DataPoolCommissionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
