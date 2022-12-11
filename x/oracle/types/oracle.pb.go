// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/oracle/v2/oracle.proto

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

// Oracle defines a detail of oracle.
type Oracle struct {
	OracleAddress        string                                 `protobuf:"bytes,1,opt,name=oracle_address,json=oracleAddress,proto3" json:"oracle_address,omitempty"`
	UniqueId             string                                 `protobuf:"bytes,2,opt,name=unique_id,json=uniqueId,proto3" json:"unique_id,omitempty"`
	Endpoint             string                                 `protobuf:"bytes,3,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	OracleCommissionRate github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,4,opt,name=oracle_commission_rate,json=oracleCommissionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"oracle_commission_rate"`
}

func (m *Oracle) Reset()         { *m = Oracle{} }
func (m *Oracle) String() string { return proto.CompactTextString(m) }
func (*Oracle) ProtoMessage()    {}
func (*Oracle) Descriptor() ([]byte, []int) {
	return fileDescriptor_35c1a1e2fbbbc7ea, []int{0}
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

func (m *Oracle) GetOracleAddress() string {
	if m != nil {
		return m.OracleAddress
	}
	return ""
}

func (m *Oracle) GetUniqueId() string {
	if m != nil {
		return m.UniqueId
	}
	return ""
}

func (m *Oracle) GetEndpoint() string {
	if m != nil {
		return m.Endpoint
	}
	return ""
}

// OracleRegistration defines the detailed states of the registration of oracle.
type OracleRegistration struct {
	UniqueId      string `protobuf:"bytes,1,opt,name=unique_id,json=uniqueId,proto3" json:"unique_id,omitempty"`
	OracleAddress string `protobuf:"bytes,2,opt,name=oracle_address,json=oracleAddress,proto3" json:"oracle_address,omitempty"`
	// Node public key is a pair with a node private key which is generated in SGX by each oracle.
	// This key is used to share the oracle private key from other oracles.
	NodePubKey []byte `protobuf:"bytes,3,opt,name=node_pub_key,json=nodePubKey,proto3" json:"node_pub_key,omitempty"`
	// Anyone can validate that the node key pair is generated in SGX using this node key remote report.
	NodePubKeyRemoteReport []byte `protobuf:"bytes,4,opt,name=node_pub_key_remote_report,json=nodePubKeyRemoteReport,proto3" json:"node_pub_key_remote_report,omitempty"`
	// The trusted block info is required for light client.
	// Other oracle can validate whether the oracle set correct trusted block info.
	TrustedBlockHeight     int64                                  `protobuf:"varint,5,opt,name=trusted_block_height,json=trustedBlockHeight,proto3" json:"trusted_block_height,omitempty"`
	TrustedBlockHash       []byte                                 `protobuf:"bytes,6,opt,name=trusted_block_hash,json=trustedBlockHash,proto3" json:"trusted_block_hash,omitempty"`
	Endpoint               string                                 `protobuf:"bytes,7,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	OracleCommissionRate   github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,8,opt,name=oracle_commission_rate,json=oracleCommissionRate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"oracle_commission_rate"`
	EncryptedOraclePrivKey []byte                                 `protobuf:"bytes,9,opt,name=encrypted_oracle_priv_key,json=encryptedOraclePrivKey,proto3" json:"encrypted_oracle_priv_key,omitempty"`
}

func (m *OracleRegistration) Reset()         { *m = OracleRegistration{} }
func (m *OracleRegistration) String() string { return proto.CompactTextString(m) }
func (*OracleRegistration) ProtoMessage()    {}
func (*OracleRegistration) Descriptor() ([]byte, []int) {
	return fileDescriptor_35c1a1e2fbbbc7ea, []int{1}
}
func (m *OracleRegistration) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OracleRegistration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OracleRegistration.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OracleRegistration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OracleRegistration.Merge(m, src)
}
func (m *OracleRegistration) XXX_Size() int {
	return m.Size()
}
func (m *OracleRegistration) XXX_DiscardUnknown() {
	xxx_messageInfo_OracleRegistration.DiscardUnknown(m)
}

var xxx_messageInfo_OracleRegistration proto.InternalMessageInfo

func (m *OracleRegistration) GetUniqueId() string {
	if m != nil {
		return m.UniqueId
	}
	return ""
}

func (m *OracleRegistration) GetOracleAddress() string {
	if m != nil {
		return m.OracleAddress
	}
	return ""
}

func (m *OracleRegistration) GetNodePubKey() []byte {
	if m != nil {
		return m.NodePubKey
	}
	return nil
}

func (m *OracleRegistration) GetNodePubKeyRemoteReport() []byte {
	if m != nil {
		return m.NodePubKeyRemoteReport
	}
	return nil
}

func (m *OracleRegistration) GetTrustedBlockHeight() int64 {
	if m != nil {
		return m.TrustedBlockHeight
	}
	return 0
}

func (m *OracleRegistration) GetTrustedBlockHash() []byte {
	if m != nil {
		return m.TrustedBlockHash
	}
	return nil
}

func (m *OracleRegistration) GetEndpoint() string {
	if m != nil {
		return m.Endpoint
	}
	return ""
}

func (m *OracleRegistration) GetEncryptedOraclePrivKey() []byte {
	if m != nil {
		return m.EncryptedOraclePrivKey
	}
	return nil
}

func init() {
	proto.RegisterType((*Oracle)(nil), "panacea.oracle.v2.Oracle")
	proto.RegisterType((*OracleRegistration)(nil), "panacea.oracle.v2.OracleRegistration")
}

func init() { proto.RegisterFile("panacea/oracle/v2/oracle.proto", fileDescriptor_35c1a1e2fbbbc7ea) }

var fileDescriptor_35c1a1e2fbbbc7ea = []byte{
	// 454 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0x41, 0x6f, 0xd3, 0x30,
	0x14, 0x6e, 0x56, 0x28, 0xad, 0x55, 0x10, 0x58, 0xd5, 0x14, 0x8a, 0x94, 0x55, 0x93, 0x40, 0x3b,
	0xb0, 0x04, 0x8d, 0x13, 0xdc, 0x28, 0x1c, 0x98, 0x76, 0x60, 0xf2, 0x91, 0x8b, 0xe5, 0xd8, 0x4f,
	0x89, 0xd5, 0x25, 0x0e, 0xb6, 0x13, 0x91, 0x7f, 0xc1, 0xcf, 0xda, 0x09, 0x4d, 0xe2, 0x82, 0x38,
	0x4c, 0xa8, 0xfd, 0x23, 0x28, 0x76, 0x36, 0xd6, 0xc1, 0x81, 0x03, 0xa7, 0x38, 0xdf, 0xf7, 0xe5,
	0xf9, 0xfb, 0xde, 0x7b, 0x41, 0x51, 0xc5, 0x4a, 0xc6, 0x81, 0x25, 0x4a, 0x33, 0x7e, 0x06, 0x49,
	0x73, 0xd4, 0x9f, 0xe2, 0x4a, 0x2b, 0xab, 0xf0, 0xa3, 0x9e, 0x8f, 0x7b, 0xb4, 0x39, 0x9a, 0xcf,
	0x32, 0x95, 0x29, 0xc7, 0x26, 0xdd, 0xc9, 0x0b, 0xf7, 0xbf, 0x06, 0x68, 0xf4, 0xc1, 0x69, 0xf0,
	0x53, 0xf4, 0xc0, 0xab, 0x29, 0x13, 0x42, 0x83, 0x31, 0x61, 0xb0, 0x08, 0x0e, 0x26, 0xe4, 0xbe,
	0x47, 0xdf, 0x78, 0x10, 0x3f, 0x41, 0x93, 0xba, 0x94, 0x9f, 0x6a, 0xa0, 0x52, 0x84, 0x3b, 0x4e,
	0x31, 0xf6, 0xc0, 0xb1, 0xc0, 0x73, 0x34, 0x86, 0x52, 0x54, 0x4a, 0x96, 0x36, 0x1c, 0x7a, 0xee,
	0xea, 0x1d, 0x0b, 0xb4, 0xdb, 0xd7, 0xe7, 0xaa, 0x28, 0xa4, 0x31, 0x52, 0x95, 0x54, 0x33, 0x0b,
	0xe1, 0x9d, 0x4e, 0xb9, 0x8c, 0xcf, 0x2f, 0xf7, 0x06, 0x3f, 0x2e, 0xf7, 0x9e, 0x65, 0xd2, 0xe6,
	0x75, 0x1a, 0x73, 0x55, 0x24, 0x5c, 0x99, 0x42, 0x99, 0xfe, 0x71, 0x68, 0xc4, 0x2a, 0xb1, 0x6d,
	0x05, 0x26, 0x7e, 0x07, 0x9c, 0xcc, 0x7c, 0xb5, 0xb7, 0xd7, 0xc5, 0x08, 0xb3, 0xb0, 0xff, 0x6d,
	0x88, 0xb0, 0x0f, 0x44, 0x20, 0x93, 0xc6, 0x6a, 0x66, 0xa5, 0x2a, 0xb7, 0x5d, 0x07, 0xb7, 0x5c,
	0xff, 0x99, 0x7c, 0xe7, 0x6f, 0xc9, 0x17, 0x68, 0x5a, 0x2a, 0x01, 0xb4, 0xaa, 0x53, 0xba, 0x82,
	0xd6, 0x05, 0x9c, 0x12, 0xd4, 0x61, 0xa7, 0x75, 0x7a, 0x02, 0x2d, 0x7e, 0x8d, 0xe6, 0x37, 0x15,
	0x54, 0x43, 0xa1, 0x2c, 0x50, 0x0d, 0x95, 0xd2, 0xd6, 0xc5, 0x9c, 0x92, 0xdd, 0xdf, 0x7a, 0xe2,
	0x68, 0xe2, 0x58, 0xfc, 0x02, 0xcd, 0xac, 0xae, 0x8d, 0x05, 0x41, 0xd3, 0x33, 0xc5, 0x57, 0x34,
	0x07, 0x99, 0xe5, 0x36, 0xbc, 0xbb, 0x08, 0x0e, 0x86, 0x04, 0xf7, 0xdc, 0xb2, 0xa3, 0xde, 0x3b,
	0x06, 0x3f, 0x47, 0xf8, 0xd6, 0x17, 0xcc, 0xe4, 0xe1, 0xc8, 0xdd, 0xf2, 0x70, 0x4b, 0xcf, 0x4c,
	0xbe, 0x35, 0x9a, 0x7b, 0xff, 0x3c, 0x9a, 0xf1, 0xff, 0x1b, 0x0d, 0x7e, 0x85, 0x1e, 0x43, 0xc9,
	0x75, 0x5b, 0x75, 0x8e, 0xfb, 0xfb, 0x2a, 0x2d, 0x1b, 0xd7, 0xcc, 0x89, 0x6f, 0xce, 0xb5, 0xc0,
	0xcf, 0xf0, 0x54, 0xcb, 0xe6, 0x04, 0xda, 0xe5, 0xf1, 0xf9, 0x3a, 0x0a, 0x2e, 0xd6, 0x51, 0xf0,
	0x73, 0x1d, 0x05, 0x5f, 0x36, 0xd1, 0xe0, 0x62, 0x13, 0x0d, 0xbe, 0x6f, 0xa2, 0xc1, 0xc7, 0xe4,
	0x86, 0xa5, 0x02, 0x84, 0xec, 0x3a, 0x91, 0xf4, 0xdb, 0x7f, 0xc8, 0x95, 0x86, 0xe4, 0xf3, 0xd5,
	0x4f, 0xe2, 0xfc, 0xa5, 0x23, 0xb7, 0xf8, 0x2f, 0x7f, 0x05, 0x00, 0x00, 0xff, 0xff, 0xa2, 0x6a,
	0x7f, 0xbd, 0x43, 0x03, 0x00, 0x00,
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
	{
		size := m.OracleCommissionRate.Size()
		i -= size
		if _, err := m.OracleCommissionRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintOracle(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if len(m.Endpoint) > 0 {
		i -= len(m.Endpoint)
		copy(dAtA[i:], m.Endpoint)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.Endpoint)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.UniqueId) > 0 {
		i -= len(m.UniqueId)
		copy(dAtA[i:], m.UniqueId)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.UniqueId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.OracleAddress) > 0 {
		i -= len(m.OracleAddress)
		copy(dAtA[i:], m.OracleAddress)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.OracleAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *OracleRegistration) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OracleRegistration) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OracleRegistration) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.EncryptedOraclePrivKey) > 0 {
		i -= len(m.EncryptedOraclePrivKey)
		copy(dAtA[i:], m.EncryptedOraclePrivKey)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.EncryptedOraclePrivKey)))
		i--
		dAtA[i] = 0x4a
	}
	{
		size := m.OracleCommissionRate.Size()
		i -= size
		if _, err := m.OracleCommissionRate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintOracle(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x42
	if len(m.Endpoint) > 0 {
		i -= len(m.Endpoint)
		copy(dAtA[i:], m.Endpoint)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.Endpoint)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.TrustedBlockHash) > 0 {
		i -= len(m.TrustedBlockHash)
		copy(dAtA[i:], m.TrustedBlockHash)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.TrustedBlockHash)))
		i--
		dAtA[i] = 0x32
	}
	if m.TrustedBlockHeight != 0 {
		i = encodeVarintOracle(dAtA, i, uint64(m.TrustedBlockHeight))
		i--
		dAtA[i] = 0x28
	}
	if len(m.NodePubKeyRemoteReport) > 0 {
		i -= len(m.NodePubKeyRemoteReport)
		copy(dAtA[i:], m.NodePubKeyRemoteReport)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.NodePubKeyRemoteReport)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.NodePubKey) > 0 {
		i -= len(m.NodePubKey)
		copy(dAtA[i:], m.NodePubKey)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.NodePubKey)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.OracleAddress) > 0 {
		i -= len(m.OracleAddress)
		copy(dAtA[i:], m.OracleAddress)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.OracleAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.UniqueId) > 0 {
		i -= len(m.UniqueId)
		copy(dAtA[i:], m.UniqueId)
		i = encodeVarintOracle(dAtA, i, uint64(len(m.UniqueId)))
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
	l = len(m.OracleAddress)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	l = len(m.UniqueId)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	l = len(m.Endpoint)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	l = m.OracleCommissionRate.Size()
	n += 1 + l + sovOracle(uint64(l))
	return n
}

func (m *OracleRegistration) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.UniqueId)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	l = len(m.OracleAddress)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	l = len(m.NodePubKey)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	l = len(m.NodePubKeyRemoteReport)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	if m.TrustedBlockHeight != 0 {
		n += 1 + sovOracle(uint64(m.TrustedBlockHeight))
	}
	l = len(m.TrustedBlockHash)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	l = len(m.Endpoint)
	if l > 0 {
		n += 1 + l + sovOracle(uint64(l))
	}
	l = m.OracleCommissionRate.Size()
	n += 1 + l + sovOracle(uint64(l))
	l = len(m.EncryptedOraclePrivKey)
	if l > 0 {
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
				return fmt.Errorf("proto: wrong wireType = %d for field OracleAddress", wireType)
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
			m.OracleAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UniqueId", wireType)
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
			m.UniqueId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Endpoint", wireType)
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
			m.Endpoint = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OracleCommissionRate", wireType)
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
			if err := m.OracleCommissionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *OracleRegistration) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: OracleRegistration: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OracleRegistration: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UniqueId", wireType)
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
			m.UniqueId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OracleAddress", wireType)
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
			m.OracleAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NodePubKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOracle
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
				return ErrInvalidLengthOracle
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthOracle
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NodePubKey = append(m.NodePubKey[:0], dAtA[iNdEx:postIndex]...)
			if m.NodePubKey == nil {
				m.NodePubKey = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NodePubKeyRemoteReport", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOracle
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
				return ErrInvalidLengthOracle
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthOracle
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.NodePubKeyRemoteReport = append(m.NodePubKeyRemoteReport[:0], dAtA[iNdEx:postIndex]...)
			if m.NodePubKeyRemoteReport == nil {
				m.NodePubKeyRemoteReport = []byte{}
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TrustedBlockHeight", wireType)
			}
			m.TrustedBlockHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOracle
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TrustedBlockHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TrustedBlockHash", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOracle
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
				return ErrInvalidLengthOracle
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthOracle
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TrustedBlockHash = append(m.TrustedBlockHash[:0], dAtA[iNdEx:postIndex]...)
			if m.TrustedBlockHash == nil {
				m.TrustedBlockHash = []byte{}
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Endpoint", wireType)
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
			m.Endpoint = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OracleCommissionRate", wireType)
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
			if err := m.OracleCommissionRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EncryptedOraclePrivKey", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOracle
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
				return ErrInvalidLengthOracle
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthOracle
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.EncryptedOraclePrivKey = append(m.EncryptedOraclePrivKey[:0], dAtA[iNdEx:postIndex]...)
			if m.EncryptedOraclePrivKey == nil {
				m.EncryptedOraclePrivKey = []byte{}
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
