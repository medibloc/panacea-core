// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/oracle/v2alpha2/oracle.proto

package types2

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// OracleStatus enumerates the status of oracle.
type OracleStatus int32

const (
	// ORACLE_STATUS_UNSPECIFIED
	ORACLE_STATUS_UNSPECIFIED OracleStatus = 0
	// ACTIVE defines the oracle status that is active
	ORACLE_STATUS_ACTIVE OracleStatus = 1
	// JAILED defines the oracle status that has been jailed
	ORACLE_STATUS_JAILED OracleStatus = 2
	// MIGRATING defines the oracle status that is in migrating to new version of oracle
	ORACLE_STATUS_MIGRATING OracleStatus = 3
)

var OracleStatus_name = map[int32]string{
	0: "ORACLE_STATUS_UNSPECIFIED",
	1: "ORACLE_STATUS_ACTIVE",
	2: "ORACLE_STATUS_JAILED",
	3: "ORACLE_STATUS_MIGRATING",
}

var OracleStatus_value = map[string]int32{
	"ORACLE_STATUS_UNSPECIFIED": 0,
	"ORACLE_STATUS_ACTIVE":      1,
	"ORACLE_STATUS_JAILED":      2,
	"ORACLE_STATUS_MIGRATING":   3,
}

func (x OracleStatus) String() string {
	return proto.EnumName(OracleStatus_name, int32(x))
}

func (OracleStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_5566fafdf0dde25a, []int{0}
}

// Oracle defines a detail of oracle.
type Oracle struct {
	Address  string       `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Status   OracleStatus `protobuf:"varint,2,opt,name=status,proto3,enum=panacea.oracle.v2alpha2.OracleStatus" json:"status,omitempty"`
	UpTime   uint64       `protobuf:"varint,3,opt,name=up_time,json=upTime,proto3" json:"up_time,omitempty"`
	JailedAt *time.Time   `protobuf:"bytes,4,opt,name=jailed_at,json=jailedAt,proto3,stdtime,stdduration" json:"jailed_at,omitempty" yaml:"jailed_at"`
}

func (m *Oracle) Reset()         { *m = Oracle{} }
func (m *Oracle) String() string { return proto.CompactTextString(m) }
func (*Oracle) ProtoMessage()    {}
func (*Oracle) Descriptor() ([]byte, []int) {
	return fileDescriptor_5566fafdf0dde25a, []int{0}
}
func (m *Oracle) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Oracle) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Oracle.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Oracle) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Oracle.Merge(m, src)
}
func (m *Oracle) XXX_Size() int {
	return m.Size()
}
func (m *Oracle) XXX_DiscardUnknown() {
	xxx_messageInfo_Oracle.DiscardUnknown(m)
}

var xxx_messageInfo_Oracle proto.InternalMessageInfo

func (m *Oracle) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *Oracle) GetStatus() OracleStatus {
	if m != nil {
		return m.Status
	}
	return ORACLE_STATUS_UNSPECIFIED
}

func (m *Oracle) GetUpTime() uint64 {
	if m != nil {
		return m.UpTime
	}
	return 0
}

func (m *Oracle) GetJailedAt() *time.Time {
	if m != nil {
		return m.JailedAt
	}
	return nil
}

func init() {
	proto.RegisterEnum("panacea.oracle.v2alpha2.OracleStatus", OracleStatus_name, OracleStatus_value)
	proto.RegisterType((*Oracle)(nil), "panacea.oracle.v2alpha2.Oracle")
}

func init() {
	proto.RegisterFile("panacea/oracle/v2alpha2/oracle.proto", fileDescriptor_5566fafdf0dde25a)
}

var fileDescriptor_5566fafdf0dde25a = []byte{
	// 410 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x29, 0x48, 0xcc, 0x4b,
	0x4c, 0x4e, 0x4d, 0xd4, 0xcf, 0x2f, 0x4a, 0x4c, 0xce, 0x49, 0xd5, 0x2f, 0x33, 0x4a, 0xcc, 0x29,
	0xc8, 0x48, 0x34, 0x82, 0xf2, 0xf5, 0x0a, 0x8a, 0xf2, 0x4b, 0xf2, 0x85, 0xc4, 0xa1, 0xaa, 0xf4,
	0xa0, 0xa2, 0x30, 0x55, 0x52, 0x22, 0xe9, 0xf9, 0xe9, 0xf9, 0x60, 0x35, 0xfa, 0x20, 0x16, 0x44,
	0xb9, 0x94, 0x7c, 0x7a, 0x7e, 0x7e, 0x7a, 0x4e, 0xaa, 0x3e, 0x98, 0x97, 0x54, 0x9a, 0xa6, 0x5f,
	0x92, 0x99, 0x9b, 0x5a, 0x5c, 0x92, 0x98, 0x5b, 0x00, 0x51, 0xa0, 0x74, 0x95, 0x91, 0x8b, 0xcd,
	0x1f, 0x6c, 0x94, 0x90, 0x04, 0x17, 0x7b, 0x62, 0x4a, 0x4a, 0x51, 0x6a, 0x71, 0xb1, 0x04, 0xa3,
	0x02, 0xa3, 0x06, 0x67, 0x10, 0x8c, 0x2b, 0x64, 0xcb, 0xc5, 0x56, 0x5c, 0x92, 0x58, 0x52, 0x5a,
	0x2c, 0xc1, 0xa4, 0xc0, 0xa8, 0xc1, 0x67, 0xa4, 0xaa, 0x87, 0xc3, 0x15, 0x7a, 0x10, 0xa3, 0x82,
	0xc1, 0x8a, 0x83, 0xa0, 0x9a, 0x84, 0xc4, 0xb9, 0xd8, 0x4b, 0x0b, 0xe2, 0x41, 0x36, 0x4b, 0x30,
	0x2b, 0x30, 0x6a, 0xb0, 0x04, 0xb1, 0x95, 0x16, 0x84, 0x64, 0xe6, 0xa6, 0x0a, 0x45, 0x72, 0x71,
	0x66, 0x25, 0x66, 0xe6, 0xa4, 0xa6, 0xc4, 0x27, 0x96, 0x48, 0xb0, 0x28, 0x30, 0x6a, 0x70, 0x1b,
	0x49, 0xe9, 0x41, 0x5c, 0xac, 0x07, 0x73, 0xb1, 0x5e, 0x08, 0xcc, 0xc5, 0x4e, 0x0a, 0x27, 0xee,
	0xc9, 0x33, 0x7e, 0xba, 0x27, 0x2f, 0x50, 0x99, 0x98, 0x9b, 0x63, 0xa5, 0x04, 0xd7, 0xaa, 0x34,
	0xe1, 0xbe, 0x3c, 0xe3, 0x8c, 0xfb, 0xf2, 0x8c, 0x41, 0x1c, 0x10, 0x31, 0xc7, 0x12, 0xad, 0x16,
	0x46, 0x2e, 0x1e, 0x64, 0xc7, 0x08, 0xc9, 0x72, 0x49, 0xfa, 0x07, 0x39, 0x3a, 0xfb, 0xb8, 0xc6,
	0x07, 0x87, 0x38, 0x86, 0x84, 0x06, 0xc7, 0x87, 0xfa, 0x05, 0x07, 0xb8, 0x3a, 0x7b, 0xba, 0x79,
	0xba, 0xba, 0x08, 0x30, 0x08, 0x49, 0x70, 0x89, 0xa0, 0x4a, 0x3b, 0x3a, 0x87, 0x78, 0x86, 0xb9,
	0x0a, 0x30, 0x62, 0xca, 0x78, 0x39, 0x7a, 0xfa, 0xb8, 0xba, 0x08, 0x30, 0x09, 0x49, 0x73, 0x89,
	0xa3, 0xca, 0xf8, 0x7a, 0xba, 0x07, 0x39, 0x86, 0x78, 0xfa, 0xb9, 0x0b, 0x30, 0x4b, 0xb1, 0x74,
	0x2c, 0x96, 0x63, 0x70, 0xf2, 0x3a, 0xf1, 0x48, 0x8e, 0xf1, 0xc2, 0x23, 0x39, 0xc6, 0x07, 0x8f,
	0xe4, 0x18, 0x27, 0x3c, 0x96, 0x63, 0xb8, 0xf0, 0x58, 0x8e, 0xe1, 0xc6, 0x63, 0x39, 0x86, 0x28,
	0x83, 0xf4, 0xcc, 0x92, 0x8c, 0xd2, 0x24, 0xbd, 0xe4, 0xfc, 0x5c, 0xfd, 0xdc, 0xd4, 0x94, 0xcc,
	0xa4, 0x9c, 0xfc, 0x64, 0x7d, 0x68, 0xb0, 0xea, 0x26, 0xe7, 0x17, 0xa5, 0xea, 0x57, 0xc0, 0x52,
	0x42, 0x49, 0x65, 0x41, 0x6a, 0xb1, 0x51, 0x12, 0x1b, 0x38, 0x48, 0x8c, 0x01, 0x01, 0x00, 0x00,
	0xff, 0xff, 0xf3, 0xd1, 0xb7, 0x32, 0x29, 0x02, 0x00, 0x00,
}

func (m *Oracle) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Oracle) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Oracle) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.JailedAt != nil {
		n1, err1 := github_com_gogo_protobuf_types.StdTimeMarshalTo(*m.JailedAt, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdTime(*m.JailedAt):])
		if err1 != nil {
			return 0, err1
		}
		i -= n1
		i = encodeVarintOracle(dAtA, i, uint64(n1))
		i--
		dAtA[i] = 0x22
	}
	if m.UpTime != 0 {
		i = encodeVarintOracle(dAtA, i, uint64(m.UpTime))
		i--
		dAtA[i] = 0x18
	}
	if m.Status != 0 {
		i = encodeVarintOracle(dAtA, i, uint64(m.Status))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintOracle(dAtA []byte, offset int, v uint64) int {
	offset -= sovOracle(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Oracle) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	if m.Status != 0 {
		n += 1 + sovOracle(uint64(m.Status))
	}
	if m.UpTime != 0 {
		n += 1 + sovOracle(uint64(m.UpTime))
	}
	if m.JailedAt != nil {
		l = github_com_gogo_protobuf_types.SizeOfStdTime(*m.JailedAt)
		n += 1 + l + sovOracle(uint64(l))
	}
	return n
}

func sovOracle(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozOracle(x uint64) (n int) {
	return sovOracle(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Oracle) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOracle
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
			return fmt.Errorf("proto: Oracle: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Oracle: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOracle
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
				return ErrInvalidLengthOracle
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOracle
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			m.Status = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOracle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Status |= OracleStatus(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpTime", wireType)
			}
			m.UpTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOracle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpTime |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field JailedAt", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOracle
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
				return ErrInvalidLengthOracle
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthOracle
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.JailedAt == nil {
				m.JailedAt = new(time.Time)
			}
			if err := github_com_gogo_protobuf_types.StdTimeUnmarshal(m.JailedAt, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOracle(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthOracle
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
func skipOracle(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowOracle
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
					return 0, ErrIntOverflowOracle
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
					return 0, ErrIntOverflowOracle
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
				return 0, ErrInvalidLengthOracle
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupOracle
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthOracle
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthOracle        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowOracle          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupOracle = fmt.Errorf("proto: unexpected end of group")
)
