// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/datadeal/v2alpha2/genesis.proto

package types

import (
	fmt "fmt"
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

// GenesisState defines the datadeal module's genesis state.
type GenesisState struct {
	Deals                 []Deal                 `protobuf:"bytes,1,rep,name=deals,proto3" json:"deals"`
	NextDealNumber        uint64                 `protobuf:"varint,2,opt,name=next_deal_number,json=nextDealNumber,proto3" json:"next_deal_number,omitempty"`
	DataSales             []DataSale             `protobuf:"bytes,3,rep,name=data_sales,json=dataSales,proto3" json:"data_sales"`
	DataVerificationVotes []DataVerificationVote `protobuf:"bytes,4,rep,name=data_verification_votes,json=dataVerificationVotes,proto3" json:"data_verification_votes"`
	DataDeliveryVotes     []DataDeliveryVote     `protobuf:"bytes,5,rep,name=data_delivery_votes,json=dataDeliveryVotes,proto3" json:"data_delivery_votes"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_148a7361fee02e04, []int{0}
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

func (m *GenesisState) GetDeals() []Deal {
	if m != nil {
		return m.Deals
	}
	return nil
}

func (m *GenesisState) GetNextDealNumber() uint64 {
	if m != nil {
		return m.NextDealNumber
	}
	return 0
}

func (m *GenesisState) GetDataSales() []DataSale {
	if m != nil {
		return m.DataSales
	}
	return nil
}

func (m *GenesisState) GetDataVerificationVotes() []DataVerificationVote {
	if m != nil {
		return m.DataVerificationVotes
	}
	return nil
}

func (m *GenesisState) GetDataDeliveryVotes() []DataDeliveryVote {
	if m != nil {
		return m.DataDeliveryVotes
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "panacea.datadeal.v2alpha2.GenesisState")
}

func init() {
	proto.RegisterFile("panacea/datadeal/v2alpha2/genesis.proto", fileDescriptor_148a7361fee02e04)
}

var fileDescriptor_148a7361fee02e04 = []byte{
	// 364 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0xc1, 0x4a, 0xeb, 0x40,
	0x14, 0x86, 0x93, 0xdb, 0xf6, 0xc2, 0x9d, 0x2b, 0xa2, 0x51, 0x31, 0x76, 0x91, 0x96, 0x2a, 0x18,
	0x10, 0x33, 0x10, 0xdd, 0xb9, 0x2b, 0x05, 0x5d, 0x75, 0xd1, 0x42, 0x17, 0x6e, 0xc2, 0x49, 0x72,
	0x4c, 0x03, 0x49, 0x26, 0x64, 0xa6, 0xa1, 0x7d, 0x0b, 0x7d, 0xab, 0x2e, 0xbb, 0x74, 0x25, 0xd2,
	0xbe, 0x88, 0x64, 0x92, 0x62, 0x11, 0xdb, 0xee, 0x32, 0x27, 0xdf, 0xff, 0x7f, 0x33, 0x1c, 0x72,
	0x9d, 0x42, 0x02, 0x1e, 0x02, 0xf5, 0x41, 0x80, 0x8f, 0x10, 0xd1, 0xdc, 0x86, 0x28, 0x1d, 0x83,
	0x4d, 0x03, 0x4c, 0x90, 0x87, 0xdc, 0x4a, 0x33, 0x26, 0x98, 0x76, 0x51, 0x81, 0xd6, 0x1a, 0xb4,
	0xd6, 0x60, 0xf3, 0x34, 0x60, 0x01, 0x93, 0x14, 0x2d, 0xbe, 0xca, 0x40, 0xf3, 0x6a, 0x7b, 0xb3,
	0x8c, 0x97, 0x94, 0xb9, 0x83, 0x02, 0x01, 0x1c, 0x22, 0xac, 0xc8, 0xce, 0x76, 0x52, 0x4c, 0x4b,
	0xa6, 0xf3, 0x56, 0x23, 0x07, 0x8f, 0xe5, 0xb5, 0x87, 0x02, 0x04, 0x6a, 0x0f, 0xa4, 0x51, 0xa0,
	0x5c, 0x57, 0xdb, 0x35, 0xf3, 0xbf, 0xdd, 0xb2, 0xb6, 0xbe, 0xc2, 0xea, 0x21, 0x44, 0xdd, 0xfa,
	0xfc, 0xa3, 0xa5, 0x0c, 0xca, 0x8c, 0x66, 0x92, 0xa3, 0x04, 0xa7, 0xc2, 0x29, 0x4e, 0x4e, 0x32,
	0x89, 0x5d, 0xcc, 0xf4, 0x3f, 0x6d, 0xd5, 0xac, 0x0f, 0x0e, 0x8b, 0x79, 0x11, 0xe8, 0xcb, 0xa9,
	0xf6, 0x44, 0x48, 0x51, 0xe8, 0x14, 0xd7, 0xe5, 0x7a, 0x4d, 0xba, 0x2e, 0x77, 0xb9, 0x40, 0xc0,
	0x10, 0x22, 0xac, 0x7c, 0xff, 0xfc, 0xea, 0xcc, 0xb5, 0x98, 0x9c, 0xcb, 0xa6, 0x1c, 0xb3, 0xf0,
	0x25, 0xf4, 0x40, 0x84, 0x2c, 0x71, 0x72, 0x26, 0x90, 0xeb, 0x75, 0x59, 0x4b, 0xf7, 0xd4, 0x8e,
	0x36, 0x82, 0x23, 0x26, 0xd6, 0x8a, 0x33, 0xff, 0x97, 0x7f, 0x5c, 0x03, 0x72, 0x22, 0x75, 0x3e,
	0x46, 0x61, 0x8e, 0xd9, 0xac, 0x52, 0x35, 0xa4, 0xea, 0x66, 0x8f, 0xaa, 0x57, 0x85, 0x36, 0x34,
	0xc7, 0xfe, 0x8f, 0x39, 0xef, 0xf6, 0xe7, 0x4b, 0x43, 0x5d, 0x2c, 0x0d, 0xf5, 0x73, 0x69, 0xa8,
	0xaf, 0x2b, 0x43, 0x59, 0xac, 0x0c, 0xe5, 0x7d, 0x65, 0x28, 0xcf, 0xf7, 0x41, 0x28, 0xc6, 0x13,
	0xd7, 0xf2, 0x58, 0x4c, 0x63, 0xf4, 0x43, 0x37, 0x62, 0x1e, 0xad, 0x94, 0xb7, 0x1e, 0xcb, 0x90,
	0xe6, 0x36, 0x9d, 0x7e, 0xef, 0x5b, 0xcc, 0x52, 0xe4, 0xee, 0x5f, 0xb9, 0xea, 0xbb, 0xaf, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xea, 0x78, 0xb6, 0x80, 0xba, 0x02, 0x00, 0x00,
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
	if len(m.DataDeliveryVotes) > 0 {
		for iNdEx := len(m.DataDeliveryVotes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DataDeliveryVotes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.DataVerificationVotes) > 0 {
		for iNdEx := len(m.DataVerificationVotes) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DataVerificationVotes[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.DataSales) > 0 {
		for iNdEx := len(m.DataSales) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.DataSales[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if m.NextDealNumber != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.NextDealNumber))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Deals) > 0 {
		for iNdEx := len(m.Deals) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Deals[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.Deals) > 0 {
		for _, e := range m.Deals {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.NextDealNumber != 0 {
		n += 1 + sovGenesis(uint64(m.NextDealNumber))
	}
	if len(m.DataSales) > 0 {
		for _, e := range m.DataSales {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.DataVerificationVotes) > 0 {
		for _, e := range m.DataVerificationVotes {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.DataDeliveryVotes) > 0 {
		for _, e := range m.DataDeliveryVotes {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
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
				return fmt.Errorf("proto: wrong wireType = %d for field Deals", wireType)
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
			m.Deals = append(m.Deals, Deal{})
			if err := m.Deals[len(m.Deals)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextDealNumber", wireType)
			}
			m.NextDealNumber = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NextDealNumber |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataSales", wireType)
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
			m.DataSales = append(m.DataSales, DataSale{})
			if err := m.DataSales[len(m.DataSales)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataVerificationVotes", wireType)
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
			m.DataVerificationVotes = append(m.DataVerificationVotes, DataVerificationVote{})
			if err := m.DataVerificationVotes[len(m.DataVerificationVotes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataDeliveryVotes", wireType)
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
			m.DataDeliveryVotes = append(m.DataDeliveryVotes, DataDeliveryVote{})
			if err := m.DataDeliveryVotes[len(m.DataDeliveryVotes)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
