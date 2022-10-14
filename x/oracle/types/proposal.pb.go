// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/oracle/v2alpha2/proposal.proto

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

// Plan defines upgrade plan information.
type Plan struct {
	UniqueId string `protobuf:"bytes,1,opt,name=unique_id,json=uniqueId,proto3" json:"unique_id,omitempty"`
	Height   int64  `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *Plan) Reset()         { *m = Plan{} }
func (m *Plan) String() string { return proto.CompactTextString(m) }
func (*Plan) ProtoMessage()    {}
func (*Plan) Descriptor() ([]byte, []int) {
	return fileDescriptor_b33a42f0588c1055, []int{0}
}
func (m *Plan) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Plan) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Plan.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Plan) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Plan.Merge(m, src)
}
func (m *Plan) XXX_Size() int {
	return m.Size()
}
func (m *Plan) XXX_DiscardUnknown() {
	xxx_messageInfo_Plan.DiscardUnknown(m)
}

var xxx_messageInfo_Plan proto.InternalMessageInfo

// OracleUpgradeProposal defines the information required for a proposal.
type OracleUpgradeProposal struct {
	Title       string `protobuf:"bytes,1,opt,name=title,proto3" json:"title,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Plan        Plan   `protobuf:"bytes,3,opt,name=plan,proto3" json:"plan"`
}

func (m *OracleUpgradeProposal) Reset()         { *m = OracleUpgradeProposal{} }
func (m *OracleUpgradeProposal) String() string { return proto.CompactTextString(m) }
func (*OracleUpgradeProposal) ProtoMessage()    {}
func (*OracleUpgradeProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_b33a42f0588c1055, []int{1}
}
func (m *OracleUpgradeProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OracleUpgradeProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OracleUpgradeProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OracleUpgradeProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OracleUpgradeProposal.Merge(m, src)
}
func (m *OracleUpgradeProposal) XXX_Size() int {
	return m.Size()
}
func (m *OracleUpgradeProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_OracleUpgradeProposal.DiscardUnknown(m)
}

var xxx_messageInfo_OracleUpgradeProposal proto.InternalMessageInfo

func init() {
	proto.RegisterType((*Plan)(nil), "panacea.oracle.v2alpha2.Plan")
	proto.RegisterType((*OracleUpgradeProposal)(nil), "panacea.oracle.v2alpha2.OracleUpgradeProposal")
}

func init() {
	proto.RegisterFile("panacea/oracle/v2alpha2/proposal.proto", fileDescriptor_b33a42f0588c1055)
}

var fileDescriptor_b33a42f0588c1055 = []byte{
	// 306 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0xbf, 0x4e, 0xf3, 0x30,
	0x14, 0xc5, 0xe3, 0xaf, 0xf9, 0x2a, 0xea, 0x6e, 0x56, 0x81, 0x0a, 0x84, 0x1b, 0x75, 0x40, 0x5d,
	0xb0, 0xa5, 0x30, 0x20, 0xb1, 0xd1, 0x8d, 0x89, 0x2a, 0x12, 0x0b, 0x0b, 0x72, 0x1d, 0x2b, 0xb1,
	0xe4, 0xc6, 0xc6, 0x71, 0x10, 0xbc, 0x04, 0xe2, 0x11, 0x78, 0x9c, 0x8c, 0x1d, 0x99, 0x10, 0x24,
	0x0b, 0x8f, 0x81, 0xf2, 0xa7, 0x12, 0x0b, 0x9b, 0xef, 0xd5, 0xcf, 0xe7, 0x9c, 0x7b, 0xe0, 0xa9,
	0x61, 0x19, 0xe3, 0x82, 0x51, 0x6d, 0x19, 0x57, 0x82, 0x3e, 0x86, 0x4c, 0x99, 0x94, 0x85, 0xd4,
	0x58, 0x6d, 0x74, 0xce, 0x14, 0x31, 0x56, 0x3b, 0x8d, 0x0e, 0x7b, 0x8e, 0x74, 0x1c, 0xd9, 0x71,
	0x47, 0x93, 0x44, 0x27, 0xba, 0x65, 0x68, 0xf3, 0xea, 0xf0, 0xf9, 0x15, 0xf4, 0x57, 0x8a, 0x65,
	0xe8, 0x18, 0x8e, 0x8a, 0x4c, 0x3e, 0x14, 0xe2, 0x5e, 0xc6, 0x53, 0x10, 0x80, 0xc5, 0x28, 0xda,
	0xeb, 0x16, 0xd7, 0x31, 0x3a, 0x80, 0xc3, 0x54, 0xc8, 0x24, 0x75, 0xd3, 0x7f, 0x01, 0x58, 0x0c,
	0xa2, 0x7e, 0xba, 0xf4, 0xbf, 0xdf, 0x66, 0x60, 0xfe, 0x02, 0xe0, 0xfe, 0x4d, 0x6b, 0x76, 0x6b,
	0x12, 0xcb, 0x62, 0xb1, 0xea, 0x13, 0xa1, 0x09, 0xfc, 0xef, 0xa4, 0x53, 0xa2, 0x17, 0xec, 0x06,
	0x14, 0xc0, 0x71, 0x2c, 0x72, 0x6e, 0xa5, 0x71, 0x52, 0x67, 0xad, 0xe4, 0x28, 0xfa, 0xbd, 0x42,
	0x17, 0xd0, 0x37, 0x8a, 0x65, 0xd3, 0x41, 0x00, 0x16, 0xe3, 0xf0, 0x84, 0xfc, 0x71, 0x12, 0x69,
	0x92, 0x2f, 0xfd, 0xf2, 0x63, 0xe6, 0x45, 0xed, 0x87, 0x2e, 0xd0, 0x72, 0x55, 0x7e, 0x61, 0xaf,
	0xac, 0x30, 0xd8, 0x56, 0x18, 0x7c, 0x56, 0x18, 0xbc, 0xd6, 0xd8, 0xdb, 0xd6, 0xd8, 0x7b, 0xaf,
	0xb1, 0x77, 0x17, 0x26, 0xd2, 0xa5, 0xc5, 0x9a, 0x70, 0xbd, 0xa1, 0x1b, 0x11, 0xcb, 0xb5, 0xd2,
	0x9c, 0xf6, 0x0e, 0x67, 0x5c, 0xdb, 0xa6, 0x5a, 0xfa, 0xb4, 0xab, 0xd9, 0x3d, 0x1b, 0x91, 0xaf,
	0x87, 0x6d, 0x59, 0xe7, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x2b, 0x52, 0x09, 0x97, 0x85, 0x01,
	0x00, 0x00,
}

func (this *Plan) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Plan)
	if !ok {
		that2, ok := that.(Plan)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.UniqueId != that1.UniqueId {
		return false
	}
	if this.Height != that1.Height {
		return false
	}
	return true
}
func (this *OracleUpgradeProposal) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*OracleUpgradeProposal)
	if !ok {
		that2, ok := that.(OracleUpgradeProposal)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Title != that1.Title {
		return false
	}
	if this.Description != that1.Description {
		return false
	}
	if !this.Plan.Equal(&that1.Plan) {
		return false
	}
	return true
}
func (m *Plan) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Plan) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Plan) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintProposal(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x10
	}
	if len(m.UniqueId) > 0 {
		i -= len(m.UniqueId)
		copy(dAtA[i:], m.UniqueId)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.UniqueId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *OracleUpgradeProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OracleUpgradeProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OracleUpgradeProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Plan.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintProposal(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Title) > 0 {
		i -= len(m.Title)
		copy(dAtA[i:], m.Title)
		i = encodeVarintProposal(dAtA, i, uint64(len(m.Title)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintProposal(dAtA []byte, offset int, v uint64) int {
	offset -= sovProposal(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Plan) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.UniqueId)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	if m.Height != 0 {
		n += 1 + sovProposal(uint64(m.Height))
	}
	return n
}

func (m *OracleUpgradeProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Title)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovProposal(uint64(l))
	}
	l = m.Plan.Size()
	n += 1 + l + sovProposal(uint64(l))
	return n
}

func sovProposal(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozProposal(x uint64) (n int) {
	return sovProposal(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Plan) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
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
			return fmt.Errorf("proto: Plan: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Plan: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UniqueId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.UniqueId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
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
func (m *OracleUpgradeProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowProposal
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
			return fmt.Errorf("proto: OracleUpgradeProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OracleUpgradeProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Title", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Title = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Plan", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowProposal
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
				return ErrInvalidLengthProposal
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthProposal
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Plan.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipProposal(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthProposal
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
func skipProposal(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowProposal
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
					return 0, ErrIntOverflowProposal
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
					return 0, ErrIntOverflowProposal
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
				return 0, ErrInvalidLengthProposal
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupProposal
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthProposal
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthProposal        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowProposal          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupProposal = fmt.Errorf("proto: unexpected end of group")
)