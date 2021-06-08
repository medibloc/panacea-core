// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: token/token.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
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

type Token struct {
	Name         string     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Symbol       string     `protobuf:"bytes,2,opt,name=symbol,proto3" json:"symbol,omitempty"`
	TotalSupply  types.Coin `protobuf:"bytes,3,opt,name=totalSupply,proto3" json:"totalSupply"`
	Mintable     bool       `protobuf:"varint,4,opt,name=mintable,proto3" json:"mintable,omitempty"`
	OwnerAddress string     `protobuf:"bytes,5,opt,name=ownerAddress,proto3" json:"ownerAddress,omitempty"`
}

func (m *Token) Reset()         { *m = Token{} }
func (m *Token) String() string { return proto.CompactTextString(m) }
func (*Token) ProtoMessage()    {}
func (*Token) Descriptor() ([]byte, []int) {
	return fileDescriptor_6e2ef433bb3fdc80, []int{0}
}
func (m *Token) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Token) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Token.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Token) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Token.Merge(m, src)
}
func (m *Token) XXX_Size() int {
	return m.Size()
}
func (m *Token) XXX_DiscardUnknown() {
	xxx_messageInfo_Token.DiscardUnknown(m)
}

var xxx_messageInfo_Token proto.InternalMessageInfo

func (m *Token) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Token) GetSymbol() string {
	if m != nil {
		return m.Symbol
	}
	return ""
}

func (m *Token) GetTotalSupply() types.Coin {
	if m != nil {
		return m.TotalSupply
	}
	return types.Coin{}
}

func (m *Token) GetMintable() bool {
	if m != nil {
		return m.Mintable
	}
	return false
}

func (m *Token) GetOwnerAddress() string {
	if m != nil {
		return m.OwnerAddress
	}
	return ""
}

func init() {
	proto.RegisterType((*Token)(nil), "medibloc.panaceacore.token.Token")
}

func init() { proto.RegisterFile("token/token.proto", fileDescriptor_6e2ef433bb3fdc80) }

var fileDescriptor_6e2ef433bb3fdc80 = []byte{
	// 288 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0x31, 0x4e, 0xc3, 0x30,
	0x14, 0x86, 0x63, 0x68, 0xab, 0xe2, 0xb2, 0x60, 0x21, 0x14, 0x32, 0x98, 0xaa, 0x53, 0x17, 0x6c,
	0x15, 0x4e, 0xd0, 0xb2, 0x30, 0x17, 0x26, 0x36, 0xdb, 0xb1, 0x42, 0x44, 0xec, 0x17, 0xc5, 0x2e,
	0x90, 0x5b, 0x70, 0x1a, 0xce, 0xd0, 0xb1, 0x23, 0x13, 0x42, 0xc9, 0x45, 0x50, 0xdc, 0x80, 0x60,
	0xb1, 0xde, 0x7b, 0xfa, 0x7f, 0x7f, 0xd2, 0x87, 0x4f, 0x3c, 0x3c, 0x69, 0xcb, 0xc3, 0xcb, 0xca,
	0x0a, 0x3c, 0x90, 0xc4, 0xe8, 0x34, 0x97, 0x05, 0x28, 0x56, 0x0a, 0x2b, 0x94, 0x16, 0x0a, 0x2a,
	0xcd, 0x42, 0x22, 0x39, 0xcd, 0x20, 0x83, 0x10, 0xe3, 0xdd, 0xb4, 0x6f, 0x24, 0x54, 0x81, 0x33,
	0xe0, 0xb8, 0x14, 0x4e, 0xf3, 0xe7, 0x85, 0xd4, 0x5e, 0x2c, 0xb8, 0x82, 0xbc, 0xff, 0x71, 0xf6,
	0x8e, 0xf0, 0xf0, 0xbe, 0xeb, 0x13, 0x82, 0x07, 0x56, 0x18, 0x1d, 0xa3, 0x29, 0x9a, 0x1f, 0xad,
	0xc3, 0x4c, 0xce, 0xf0, 0xc8, 0xd5, 0x46, 0x42, 0x11, 0x1f, 0x84, 0x6b, 0xbf, 0x91, 0x25, 0x9e,
	0x78, 0xf0, 0xa2, 0xb8, 0xdb, 0x94, 0x65, 0x51, 0xc7, 0x87, 0x53, 0x34, 0x9f, 0x5c, 0x9d, 0xb3,
	0x3d, 0x8b, 0x75, 0x2c, 0xd6, 0xb3, 0xd8, 0x0d, 0xe4, 0x76, 0x35, 0xd8, 0x7e, 0x5e, 0x44, 0xeb,
	0xbf, 0x1d, 0x92, 0xe0, 0xb1, 0xc9, 0xad, 0x17, 0xb2, 0xd0, 0xf1, 0x60, 0x8a, 0xe6, 0xe3, 0xf5,
	0xef, 0x4e, 0x66, 0xf8, 0x18, 0x5e, 0xac, 0xae, 0x96, 0x69, 0x5a, 0x69, 0xe7, 0xe2, 0x61, 0x80,
	0xff, 0xbb, 0xad, 0x6e, 0xb7, 0x0d, 0x45, 0xbb, 0x86, 0xa2, 0xaf, 0x86, 0xa2, 0xb7, 0x96, 0x46,
	0xbb, 0x96, 0x46, 0x1f, 0x2d, 0x8d, 0x1e, 0x58, 0x96, 0xfb, 0xc7, 0x8d, 0x64, 0x0a, 0x0c, 0xff,
	0xf1, 0xc5, 0x7b, 0x5f, 0x97, 0x9d, 0x30, 0xfe, 0xca, 0x7b, 0xb5, 0x75, 0xa9, 0x9d, 0x1c, 0x05,
	0x13, 0xd7, 0xdf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xf2, 0xa6, 0x36, 0x6b, 0x70, 0x01, 0x00, 0x00,
}

func (m *Token) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Token) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Token) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.OwnerAddress) > 0 {
		i -= len(m.OwnerAddress)
		copy(dAtA[i:], m.OwnerAddress)
		i = encodeVarintToken(dAtA, i, uint64(len(m.OwnerAddress)))
		i--
		dAtA[i] = 0x2a
	}
	if m.Mintable {
		i--
		if m.Mintable {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x20
	}
	{
		size, err := m.TotalSupply.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintToken(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.Symbol) > 0 {
		i -= len(m.Symbol)
		copy(dAtA[i:], m.Symbol)
		i = encodeVarintToken(dAtA, i, uint64(len(m.Symbol)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintToken(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintToken(dAtA []byte, offset int, v uint64) int {
	offset -= sovToken(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Token) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	l = len(m.Symbol)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	l = m.TotalSupply.Size()
	n += 1 + l + sovToken(uint64(l))
	if m.Mintable {
		n += 2
	}
	l = len(m.OwnerAddress)
	if l > 0 {
		n += 1 + l + sovToken(uint64(l))
	}
	return n
}

func sovToken(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozToken(x uint64) (n int) {
	return sovToken(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Token) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowToken
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
			return fmt.Errorf("proto: Token: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Token: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Symbol", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Symbol = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalSupply", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TotalSupply.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Mintable", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Mintable = bool(v != 0)
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OwnerAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowToken
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
				return ErrInvalidLengthToken
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthToken
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OwnerAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipToken(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthToken
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
func skipToken(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowToken
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
					return 0, ErrIntOverflowToken
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
					return 0, ErrIntOverflowToken
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
				return 0, ErrInvalidLengthToken
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupToken
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthToken
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthToken        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowToken          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupToken = fmt.Errorf("proto: unexpected end of group")
)
