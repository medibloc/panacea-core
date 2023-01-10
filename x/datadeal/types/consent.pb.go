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

// Consent defines a consent that includes a certificate
type Consent struct {
	Certificate *Certificate `protobuf:"bytes,1,opt,name=certificate,proto3" json:"certificate,omitempty"`
}

func (m *Consent) Reset()         { *m = Consent{} }
func (m *Consent) String() string { return proto.CompactTextString(m) }
func (*Consent) ProtoMessage()    {}
func (*Consent) Descriptor() ([]byte, []int) {
	return fileDescriptor_b5d80581f65c4381, []int{0}
}
func (m *Consent) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Consent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Consent.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Consent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Consent.Merge(m, src)
}
func (m *Consent) XXX_Size() int {
	return m.Size()
}
func (m *Consent) XXX_DiscardUnknown() {
	xxx_messageInfo_Consent.DiscardUnknown(m)
}

var xxx_messageInfo_Consent proto.InternalMessageInfo

func (m *Consent) GetCertificate() *Certificate {
	if m != nil {
		return m.Certificate
	}
	return nil
}

// Certificate defines a certificate with signature
type Certificate struct {
	UnsignedCertificate *UnsignedCertificate `protobuf:"bytes,1,opt,name=unsigned_certificate,json=unsignedCertificate,proto3" json:"unsigned_certificate,omitempty"`
	Signature           []byte               `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *Certificate) Reset()         { *m = Certificate{} }
func (m *Certificate) String() string { return proto.CompactTextString(m) }
func (*Certificate) ProtoMessage()    {}
func (*Certificate) Descriptor() ([]byte, []int) {
	return fileDescriptor_b5d80581f65c4381, []int{1}
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

func (m *Certificate) GetUnsignedCertificate() *UnsignedCertificate {
	if m != nil {
		return m.UnsignedCertificate
	}
	return nil
}

func (m *Certificate) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

// UnsignedCertificate defines a certificate information
type UnsignedCertificate struct {
	Cid             string `protobuf:"bytes,1,opt,name=cid,proto3" json:"cid,omitempty"`
	UniqueId        string `protobuf:"bytes,2,opt,name=unique_id,json=uniqueId,proto3" json:"unique_id,omitempty"`
	OracleAddress   string `protobuf:"bytes,3,opt,name=oracle_address,json=oracleAddress,proto3" json:"oracle_address,omitempty"`
	DealId          uint64 `protobuf:"varint,4,opt,name=deal_id,json=dealId,proto3" json:"deal_id,omitempty"`
	ProviderAddress string `protobuf:"bytes,5,opt,name=provider_address,json=providerAddress,proto3" json:"provider_address,omitempty"`
	DataHash        string `protobuf:"bytes,6,opt,name=data_hash,json=dataHash,proto3" json:"data_hash,omitempty"`
}

func (m *UnsignedCertificate) Reset()         { *m = UnsignedCertificate{} }
func (m *UnsignedCertificate) String() string { return proto.CompactTextString(m) }
func (*UnsignedCertificate) ProtoMessage()    {}
func (*UnsignedCertificate) Descriptor() ([]byte, []int) {
	return fileDescriptor_b5d80581f65c4381, []int{2}
}
func (m *UnsignedCertificate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *UnsignedCertificate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_UnsignedCertificate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *UnsignedCertificate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UnsignedCertificate.Merge(m, src)
}
func (m *UnsignedCertificate) XXX_Size() int {
	return m.Size()
}
func (m *UnsignedCertificate) XXX_DiscardUnknown() {
	xxx_messageInfo_UnsignedCertificate.DiscardUnknown(m)
}

var xxx_messageInfo_UnsignedCertificate proto.InternalMessageInfo

func (m *UnsignedCertificate) GetCid() string {
	if m != nil {
		return m.Cid
	}
	return ""
}

func (m *UnsignedCertificate) GetUniqueId() string {
	if m != nil {
		return m.UniqueId
	}
	return ""
}

func (m *UnsignedCertificate) GetOracleAddress() string {
	if m != nil {
		return m.OracleAddress
	}
	return ""
}

func (m *UnsignedCertificate) GetDealId() uint64 {
	if m != nil {
		return m.DealId
	}
	return 0
}

func (m *UnsignedCertificate) GetProviderAddress() string {
	if m != nil {
		return m.ProviderAddress
	}
	return ""
}

func (m *UnsignedCertificate) GetDataHash() string {
	if m != nil {
		return m.DataHash
	}
	return ""
}

func init() {
	proto.RegisterType((*Consent)(nil), "panacea.datadeal.v2.Consent")
	proto.RegisterType((*Certificate)(nil), "panacea.datadeal.v2.Certificate")
	proto.RegisterType((*UnsignedCertificate)(nil), "panacea.datadeal.v2.UnsignedCertificate")
}

func init() { proto.RegisterFile("panacea/datadeal/v2/consent.proto", fileDescriptor_b5d80581f65c4381) }

var fileDescriptor_b5d80581f65c4381 = []byte{
	// 376 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x91, 0x4f, 0x6e, 0xe2, 0x30,
	0x14, 0x87, 0xf1, 0xc0, 0xc0, 0xc4, 0xcc, 0x1f, 0x64, 0x90, 0x26, 0x82, 0x51, 0x94, 0x41, 0x1a,
	0x29, 0xb3, 0x68, 0xa2, 0xd2, 0x13, 0x14, 0x36, 0x45, 0x55, 0x37, 0x91, 0xba, 0x69, 0x17, 0x91,
	0xb1, 0xdd, 0xc4, 0x12, 0xc4, 0xa9, 0xe3, 0xa0, 0xf6, 0x06, 0x5d, 0xf6, 0x58, 0x5d, 0x74, 0xc1,
	0xb2, 0xcb, 0x0a, 0x2e, 0x52, 0xd9, 0x81, 0x82, 0x44, 0x76, 0xf6, 0xf7, 0x7e, 0xef, 0xf3, 0x93,
	0x1f, 0xfc, 0x9b, 0xe1, 0x14, 0x13, 0x86, 0x03, 0x8a, 0x15, 0xa6, 0x0c, 0xcf, 0x83, 0xe5, 0x28,
	0x20, 0x22, 0xcd, 0x59, 0xaa, 0xfc, 0x4c, 0x0a, 0x25, 0x50, 0x77, 0x1b, 0xf1, 0x77, 0x11, 0x7f,
	0x39, 0xea, 0xf7, 0x62, 0x11, 0x0b, 0x53, 0x0f, 0xf4, 0xa9, 0x8c, 0xf6, 0x9d, 0x2a, 0x9b, 0x69,
	0x31, 0xf5, 0xe1, 0x15, 0x6c, 0x4d, 0x4a, 0x37, 0x1a, 0xc3, 0x36, 0x61, 0x52, 0xf1, 0x3b, 0x4e,
	0xb0, 0x62, 0x36, 0x70, 0x81, 0xd7, 0x1e, 0xb9, 0x7e, 0xc5, 0x5b, 0xfe, 0x64, 0x9f, 0x0b, 0x0f,
	0x9b, 0x86, 0x4f, 0x00, 0xb6, 0x0f, 0x8a, 0xe8, 0x16, 0xf6, 0x8a, 0x34, 0xe7, 0x71, 0xca, 0x68,
	0x74, 0x2c, 0xf7, 0x2a, 0xe5, 0xd7, 0xdb, 0x86, 0xc3, 0x47, 0xba, 0xc5, 0x31, 0x44, 0x7f, 0xa0,
	0xa5, 0x21, 0x56, 0x85, 0x64, 0xf6, 0x17, 0x17, 0x78, 0xdf, 0xc3, 0x3d, 0x18, 0xbe, 0x02, 0xd8,
	0xad, 0x50, 0xa1, 0x0e, 0xac, 0x13, 0x4e, 0xcd, 0x04, 0x56, 0xa8, 0x8f, 0x68, 0x00, 0xad, 0x22,
	0xe5, 0xf7, 0x05, 0x8b, 0x38, 0x35, 0x1e, 0x2b, 0xfc, 0x56, 0x82, 0x29, 0x45, 0xff, 0xe0, 0x4f,
	0x21, 0x31, 0x99, 0xb3, 0x08, 0x53, 0x2a, 0x59, 0x9e, 0xdb, 0x75, 0x93, 0xf8, 0x51, 0xd2, 0xf3,
	0x12, 0xa2, 0xdf, 0xb0, 0xa5, 0xe7, 0xd7, 0x86, 0x86, 0x0b, 0xbc, 0x46, 0xd8, 0xd4, 0xd7, 0x29,
	0x45, 0xff, 0x61, 0x27, 0x93, 0x62, 0xc9, 0x29, 0x93, 0x9f, 0x86, 0xaf, 0xc6, 0xf0, 0x6b, 0xc7,
	0x77, 0x8e, 0x01, 0xb4, 0xf4, 0x3f, 0x44, 0x09, 0xce, 0x13, 0xbb, 0x59, 0xce, 0xa1, 0xc1, 0x05,
	0xce, 0x93, 0xf1, 0xe5, 0xcb, 0xda, 0x01, 0xab, 0xb5, 0x03, 0xde, 0xd7, 0x0e, 0x78, 0xde, 0x38,
	0xb5, 0xd5, 0xc6, 0xa9, 0xbd, 0x6d, 0x9c, 0xda, 0xcd, 0x69, 0xcc, 0x55, 0x52, 0xcc, 0x7c, 0x22,
	0x16, 0xc1, 0x82, 0x51, 0x3e, 0x9b, 0x0b, 0x12, 0x6c, 0x3f, 0xf6, 0x84, 0x08, 0xc9, 0x82, 0x87,
	0xfd, 0xf6, 0xd5, 0x63, 0xc6, 0xf2, 0x59, 0xd3, 0x2c, 0xff, 0xec, 0x23, 0x00, 0x00, 0xff, 0xff,
	0x98, 0x32, 0x79, 0xbd, 0x6c, 0x02, 0x00, 0x00,
}

func (m *Consent) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Consent) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Consent) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Certificate != nil {
		{
			size, err := m.Certificate.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintConsent(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
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
	if len(m.Signature) > 0 {
		i -= len(m.Signature)
		copy(dAtA[i:], m.Signature)
		i = encodeVarintConsent(dAtA, i, uint64(len(m.Signature)))
		i--
		dAtA[i] = 0x12
	}
	if m.UnsignedCertificate != nil {
		{
			size, err := m.UnsignedCertificate.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintConsent(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *UnsignedCertificate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UnsignedCertificate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *UnsignedCertificate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.DataHash) > 0 {
		i -= len(m.DataHash)
		copy(dAtA[i:], m.DataHash)
		i = encodeVarintConsent(dAtA, i, uint64(len(m.DataHash)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.ProviderAddress) > 0 {
		i -= len(m.ProviderAddress)
		copy(dAtA[i:], m.ProviderAddress)
		i = encodeVarintConsent(dAtA, i, uint64(len(m.ProviderAddress)))
		i--
		dAtA[i] = 0x2a
	}
	if m.DealId != 0 {
		i = encodeVarintConsent(dAtA, i, uint64(m.DealId))
		i--
		dAtA[i] = 0x20
	}
	if len(m.OracleAddress) > 0 {
		i -= len(m.OracleAddress)
		copy(dAtA[i:], m.OracleAddress)
		i = encodeVarintConsent(dAtA, i, uint64(len(m.OracleAddress)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.UniqueId) > 0 {
		i -= len(m.UniqueId)
		copy(dAtA[i:], m.UniqueId)
		i = encodeVarintConsent(dAtA, i, uint64(len(m.UniqueId)))
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
func (m *Consent) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Certificate != nil {
		l = m.Certificate.Size()
		n += 1 + l + sovConsent(uint64(l))
	}
	return n
}

func (m *Certificate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.UnsignedCertificate != nil {
		l = m.UnsignedCertificate.Size()
		n += 1 + l + sovConsent(uint64(l))
	}
	l = len(m.Signature)
	if l > 0 {
		n += 1 + l + sovConsent(uint64(l))
	}
	return n
}

func (m *UnsignedCertificate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Cid)
	if l > 0 {
		n += 1 + l + sovConsent(uint64(l))
	}
	l = len(m.UniqueId)
	if l > 0 {
		n += 1 + l + sovConsent(uint64(l))
	}
	l = len(m.OracleAddress)
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
func (m *Consent) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: Consent: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Consent: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Certificate", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConsent
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
				return ErrInvalidLengthConsent
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthConsent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Certificate == nil {
				m.Certificate = &Certificate{}
			}
			if err := m.Certificate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
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
				return fmt.Errorf("proto: wrong wireType = %d for field UnsignedCertificate", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConsent
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
				return ErrInvalidLengthConsent
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthConsent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.UnsignedCertificate == nil {
				m.UnsignedCertificate = &UnsignedCertificate{}
			}
			if err := m.UnsignedCertificate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConsent
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthConsent
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthConsent
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signature = append(m.Signature[:0], dAtA[iNdEx:postIndex]...)
			if m.Signature == nil {
				m.Signature = []byte{}
			}
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
func (m *UnsignedCertificate) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: UnsignedCertificate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UnsignedCertificate: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field UniqueId", wireType)
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
			m.UniqueId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OracleAddress", wireType)
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
			m.OracleAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
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
				m.DealId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
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
		case 6:
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