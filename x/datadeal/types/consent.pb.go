// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/datadeal/v2/consent.proto

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

// Certificate defines a certificate
type Certificate struct {
	Cid             string `protobuf:"bytes,1,opt,name=cid,proto3" json:"cid,omitempty"`
	OperatorAddress string `protobuf:"bytes,2,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`
	DealId          int64  `protobuf:"varint,3,opt,name=deal_id,json=dealId,proto3" json:"deal_id,omitempty"`
	ProviderAddress string `protobuf:"bytes,4,opt,name=provider_address,json=providerAddress,proto3" json:"provider_address,omitempty"`
	DataHash        string `protobuf:"bytes,5,opt,name=data_hash,json=dataHash,proto3" json:"data_hash,omitempty"`
}

func (m *Certificate) Reset()         { *m = Certificate{} }
func (m *Certificate) String() string { return proto.CompactTextString(m) }
func (*Certificate) ProtoMessage()    {}
func (*Certificate) Descriptor() ([]byte, []int) {
	return fileDescriptor_b5d80581f65c4381, []int{0}
}
func (m *Certificate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Certificate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Certificate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Certificate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Certificate.Merge(m, src)
}
func (m *Certificate) XXX_Size() int {
	return m.Size()
}
func (m *Certificate) XXX_DiscardUnknown() {
	xxx_messageInfo_Certificate.DiscardUnknown(m)
}

var xxx_messageInfo_Certificate proto.InternalMessageInfo

func (m *Certificate) GetCid() string {
	if m != nil {
		return m.Cid
	}
	return ""
}

func (m *Certificate) GetOperatorAddress() string {
	if m != nil {
		return m.OperatorAddress
	}
	return ""
}

func (m *Certificate) GetDealId() int64 {
	if m != nil {
		return m.DealId
	}
	return 0
}

func (m *Certificate) GetProviderAddress() string {
	if m != nil {
		return m.ProviderAddress
	}
	return ""
}

func (m *Certificate) GetDataHash() string {
	if m != nil {
		return m.DataHash
	}
	return ""
}

func init() {
	proto.RegisterType((*Certificate)(nil), "panacea.datadeal.v2.Certificate")
}

func init() { proto.RegisterFile("panacea/datadeal/v2/consent.proto", fileDescriptor_b5d80581f65c4381) }

var fileDescriptor_b5d80581f65c4381 = []byte{
	// 279 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0xb1, 0x4e, 0xeb, 0x30,
	0x14, 0x86, 0xeb, 0xdb, 0x4b, 0xa1, 0x66, 0xa0, 0x0a, 0x48, 0x44, 0x45, 0xb2, 0x0a, 0x53, 0x19,
	0x88, 0x45, 0x79, 0x02, 0x60, 0x01, 0xb1, 0x75, 0x64, 0x89, 0x1c, 0xfb, 0x90, 0x58, 0x6a, 0x73,
	0x22, 0xdb, 0x44, 0xf0, 0x16, 0xbc, 0x07, 0x2f, 0xc2, 0xd8, 0x91, 0x11, 0x25, 0x2f, 0x82, 0xec,
	0x34, 0xea, 0xc2, 0x76, 0xfc, 0xff, 0x9f, 0x3f, 0xd9, 0x87, 0x9e, 0x57, 0xa2, 0x14, 0x12, 0x04,
	0x57, 0xc2, 0x09, 0x05, 0x62, 0xc5, 0xeb, 0x05, 0x97, 0x58, 0x5a, 0x28, 0x5d, 0x52, 0x19, 0x74,
	0x18, 0x1d, 0x6f, 0x91, 0xa4, 0x47, 0x92, 0x7a, 0x31, 0x3d, 0xc9, 0x31, 0xc7, 0xd0, 0x73, 0x3f,
	0x75, 0xe8, 0x94, 0xfd, 0x65, 0x0b, 0x57, 0x42, 0x7f, 0xf1, 0x49, 0xe8, 0xe1, 0x3d, 0x18, 0xa7,
	0x5f, 0xb4, 0x14, 0x0e, 0xa2, 0x09, 0x1d, 0x4a, 0xad, 0x62, 0x32, 0x23, 0xf3, 0xf1, 0xd2, 0x8f,
	0xd1, 0x25, 0x9d, 0x60, 0x05, 0x46, 0x38, 0x34, 0xa9, 0x50, 0xca, 0x80, 0xb5, 0xf1, 0xbf, 0x50,
	0x1f, 0xf5, 0xf9, 0x6d, 0x17, 0x47, 0xa7, 0x74, 0xdf, 0xab, 0x53, 0xad, 0xe2, 0xe1, 0x8c, 0xcc,
	0x87, 0xcb, 0x91, 0x3f, 0x3e, 0x06, 0x47, 0x65, 0xb0, 0xd6, 0x0a, 0x76, 0x8e, 0xff, 0x9d, 0xa3,
	0xcf, 0x7b, 0xc7, 0x19, 0x1d, 0xfb, 0xa7, 0xa6, 0x85, 0xb0, 0x45, 0xbc, 0x17, 0x98, 0x03, 0x1f,
	0x3c, 0x08, 0x5b, 0xdc, 0x3d, 0x7d, 0x35, 0x8c, 0x6c, 0x1a, 0x46, 0x7e, 0x1a, 0x46, 0x3e, 0x5a,
	0x36, 0xd8, 0xb4, 0x6c, 0xf0, 0xdd, 0xb2, 0xc1, 0xf3, 0x75, 0xae, 0x5d, 0xf1, 0x9a, 0x25, 0x12,
	0xd7, 0x7c, 0x0d, 0x4a, 0x67, 0x2b, 0x94, 0x7c, 0xfb, 0xf7, 0x2b, 0x89, 0x06, 0xf8, 0xdb, 0x6e,
	0x05, 0xee, 0xbd, 0x02, 0x9b, 0x8d, 0xc2, 0x06, 0x6e, 0x7e, 0x03, 0x00, 0x00, 0xff, 0xff, 0x2b,
	0xb1, 0x98, 0x17, 0x71, 0x01, 0x00, 0x00,
}

func (m *Certificate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Certificate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Certificate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.DataHash) > 0 {
		i -= len(m.DataHash)
		copy(dAtA[i:], m.DataHash)
		i = encodeVarintConsent(dAtA, i, uint64(len(m.DataHash)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.ProviderAddress) > 0 {
		i -= len(m.ProviderAddress)
		copy(dAtA[i:], m.ProviderAddress)
		i = encodeVarintConsent(dAtA, i, uint64(len(m.ProviderAddress)))
		i--
		dAtA[i] = 0x22
	}
	if m.DealId != 0 {
		i = encodeVarintConsent(dAtA, i, uint64(m.DealId))
		i--
		dAtA[i] = 0x18
	}
	if len(m.OperatorAddress) > 0 {
		i -= len(m.OperatorAddress)
		copy(dAtA[i:], m.OperatorAddress)
		i = encodeVarintConsent(dAtA, i, uint64(len(m.OperatorAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Cid) > 0 {
		i -= len(m.Cid)
		copy(dAtA[i:], m.Cid)
		i = encodeVarintConsent(dAtA, i, uint64(len(m.Cid)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintConsent(dAtA []byte, offset int, v uint64) int {
	offset -= sovConsent(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Certificate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Cid)
	if l > 0 {
		n += 1 + l + sovConsent(uint64(l))
	}
	l = len(m.OperatorAddress)
	if l > 0 {
		n += 1 + l + sovConsent(uint64(l))
	}
	if m.DealId != 0 {
		n += 1 + sovConsent(uint64(m.DealId))
	}
	l = len(m.ProviderAddress)
	if l > 0 {
		n += 1 + l + sovConsent(uint64(l))
	}
	l = len(m.DataHash)
	if l > 0 {
		n += 1 + l + sovConsent(uint64(l))
	}
	return n
}

func sovConsent(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozConsent(x uint64) (n int) {
	return sovConsent(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Certificate) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConsent
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
			return fmt.Errorf("proto: Certificate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Certificate: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cid", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConsent
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
				return ErrInvalidLengthConsent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConsent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Cid = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OperatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConsent
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
				return ErrInvalidLengthConsent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConsent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OperatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DealId", wireType)
			}
			m.DealId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConsent
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.DealId |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProviderAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConsent
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
				return ErrInvalidLengthConsent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConsent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ProviderAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataHash", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConsent
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
				return ErrInvalidLengthConsent
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConsent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DataHash = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipConsent(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthConsent
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
func skipConsent(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowConsent
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
					return 0, ErrIntOverflowConsent
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
					return 0, ErrIntOverflowConsent
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
				return 0, ErrInvalidLengthConsent
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupConsent
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthConsent
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthConsent        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowConsent          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupConsent = fmt.Errorf("proto: unexpected end of group")
)
