// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/oracle/v2/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/protobuf/types/known/timestamppb"
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

// GenesisState defines the oracle module's genesis state.
type GenesisState struct {
	Oracles             []Oracle             `protobuf:"bytes,1,rep,name=oracles,proto3" json:"oracles"`
	OracleRegistrations []OracleRegistration `protobuf:"bytes,2,rep,name=oracle_registrations,json=oracleRegistrations,proto3" json:"oracle_registrations"`
	OracleUpgrades      []OracleUpgrade      `protobuf:"bytes,3,rep,name=oracle_upgrades,json=oracleUpgrades,proto3" json:"oracle_upgrades"`
	OracleUpgradeInfo   *OracleUpgradeInfo   `protobuf:"bytes,4,opt,name=oracle_upgrade_info,json=oracleUpgradeInfo,proto3" json:"oracle_upgrade_info,omitempty"`
	Params              Params               `protobuf:"bytes,5,opt,name=params,proto3" json:"params"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_321e5b447e46614a, []int{0}
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

func (m *GenesisState) GetOracles() []Oracle {
	if m != nil {
		return m.Oracles
	}
	return nil
}

func (m *GenesisState) GetOracleRegistrations() []OracleRegistration {
	if m != nil {
		return m.OracleRegistrations
	}
	return nil
}

func (m *GenesisState) GetOracleUpgrades() []OracleUpgrade {
	if m != nil {
		return m.OracleUpgrades
	}
	return nil
}

func (m *GenesisState) GetOracleUpgradeInfo() *OracleUpgradeInfo {
	if m != nil {
		return m.OracleUpgradeInfo
	}
	return nil
}

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

// Params defines the oracle module's params.
type Params struct {
	// A base64-encoded oracle public key which is paired with an oracle private key generated in SGX by the first oracle.
	// This key is used to encrypt data, so that the data can be decrypted and verified securely only in SGX
	OraclePublicKey string `protobuf:"bytes,1,opt,name=oracle_public_key,json=oraclePublicKey,proto3" json:"oracle_public_key,omitempty"`
	// A base64-encoded SGX remote report which contains an oracle public key.
	// Using this report, anyone can validate that the oracle key pair was generated in SGX.
	OraclePubKeyRemoteReport string `protobuf:"bytes,2,opt,name=oracle_pub_key_remote_report,json=oraclePubKeyRemoteReport,proto3" json:"oracle_pub_key_remote_report,omitempty"`
	// The unique ID of current version of oracle.
	UniqueId string `protobuf:"bytes,3,opt,name=unique_id,json=uniqueId,proto3" json:"unique_id,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_321e5b447e46614a, []int{1}
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

func (m *Params) GetOraclePublicKey() string {
	if m != nil {
		return m.OraclePublicKey
	}
	return ""
}

func (m *Params) GetOraclePubKeyRemoteReport() string {
	if m != nil {
		return m.OraclePubKeyRemoteReport
	}
	return ""
}

func (m *Params) GetUniqueId() string {
	if m != nil {
		return m.UniqueId
	}
	return ""
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "panacea.oracle.v2.GenesisState")
	proto.RegisterType((*Params)(nil), "panacea.oracle.v2.Params")
}

func init() { proto.RegisterFile("panacea/oracle/v2/genesis.proto", fileDescriptor_321e5b447e46614a) }

var fileDescriptor_321e5b447e46614a = []byte{
	// 430 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0x41, 0x6f, 0xd3, 0x30,
	0x14, 0xc7, 0x9b, 0x75, 0x14, 0xe6, 0x21, 0xd0, 0xbc, 0x1d, 0xcc, 0x40, 0x69, 0x35, 0x81, 0x54,
	0x21, 0x91, 0x48, 0xe1, 0x80, 0xb8, 0x70, 0xd8, 0x05, 0x4d, 0x43, 0xda, 0x14, 0xe0, 0xc2, 0x81,
	0xc8, 0x49, 0x5f, 0x8d, 0x45, 0x93, 0x67, 0x6c, 0x67, 0x22, 0x1f, 0x81, 0x1b, 0x7c, 0xab, 0x1d,
	0x77, 0xe4, 0x84, 0x50, 0xfb, 0x45, 0x50, 0x6d, 0x0f, 0x06, 0xa5, 0xda, 0xed, 0xe5, 0xfd, 0x7f,
	0xfe, 0xd9, 0xb1, 0x1f, 0x19, 0x2a, 0xde, 0xf0, 0x0a, 0x78, 0x8a, 0x9a, 0x57, 0x33, 0x48, 0xcf,
	0xb2, 0x54, 0x40, 0x03, 0x46, 0x9a, 0x44, 0x69, 0xb4, 0x48, 0x77, 0x02, 0x90, 0x78, 0x20, 0x39,
	0xcb, 0xf6, 0xf7, 0x04, 0x0a, 0x74, 0x69, 0xba, 0xac, 0x3c, 0xb8, 0x3f, 0x14, 0x88, 0x62, 0x06,
	0xa9, 0xfb, 0x2a, 0xdb, 0x69, 0x6a, 0x65, 0x0d, 0xc6, 0xf2, 0x5a, 0x05, 0x20, 0x5e, 0xdd, 0x2a,
	0x38, 0x5d, 0x7e, 0xf0, 0xa5, 0x4f, 0x6e, 0xbf, 0xf4, 0x7b, 0xbf, 0xb6, 0xdc, 0x02, 0x7d, 0x4e,
	0x6e, 0x7a, 0xc0, 0xb0, 0x68, 0xd4, 0x1f, 0x6f, 0x67, 0xf7, 0x92, 0x95, 0xc3, 0x24, 0x27, 0xae,
	0x3a, 0xdc, 0x3c, 0xff, 0x31, 0xec, 0xe5, 0x97, 0x3c, 0x7d, 0x4f, 0xf6, 0x7c, 0x59, 0x68, 0x10,
	0xd2, 0x58, 0xcd, 0xad, 0xc4, 0xc6, 0xb0, 0x0d, 0xe7, 0x79, 0xb4, 0xd6, 0x93, 0x5f, 0xa1, 0x83,
	0x73, 0x17, 0x57, 0x12, 0x43, 0x4f, 0xc8, 0xdd, 0xe0, 0x6f, 0x95, 0xd0, 0x7c, 0x02, 0x86, 0xf5,
	0x9d, 0x7a, 0xb4, 0x56, 0xfd, 0xd6, 0x83, 0xc1, 0x7a, 0x07, 0xaf, 0x36, 0x0d, 0x7d, 0x43, 0x76,
	0xff, 0x16, 0x16, 0xb2, 0x99, 0x22, 0xdb, 0x1c, 0x45, 0xe3, 0xed, 0xec, 0xe1, 0x75, 0xd2, 0xa3,
	0x66, 0x8a, 0xf9, 0x0e, 0xfe, 0xdb, 0xa2, 0xcf, 0xc8, 0x40, 0x71, 0xcd, 0x6b, 0xc3, 0x6e, 0x38,
	0xd1, 0xff, 0x2e, 0xf0, 0xd4, 0x01, 0xe1, 0x58, 0x01, 0x3f, 0xf8, 0x16, 0x91, 0x81, 0x0f, 0xe8,
	0x63, 0x12, 0xc4, 0x85, 0x6a, 0xcb, 0x99, 0xac, 0x8a, 0x8f, 0xd0, 0xb1, 0x68, 0x14, 0x8d, 0xb7,
	0xf2, 0x70, 0x07, 0xa7, 0xae, 0x7f, 0x0c, 0x1d, 0x7d, 0x41, 0x1e, 0xfc, 0x61, 0x97, 0x60, 0xa1,
	0xa1, 0x46, 0xbb, 0x7c, 0x05, 0x85, 0xda, 0xb2, 0x0d, 0xb7, 0x8c, 0xfd, 0x5e, 0x76, 0x0c, 0x5d,
	0xee, 0x80, 0xdc, 0xe5, 0xf4, 0x3e, 0xd9, 0x6a, 0x1b, 0xf9, 0xa9, 0x85, 0x42, 0x4e, 0x58, 0xdf,
	0xc1, 0xb7, 0x7c, 0xe3, 0x68, 0x72, 0xf8, 0xea, 0x7c, 0x1e, 0x47, 0x17, 0xf3, 0x38, 0xfa, 0x39,
	0x8f, 0xa3, 0xaf, 0x8b, 0xb8, 0x77, 0xb1, 0x88, 0x7b, 0xdf, 0x17, 0x71, 0xef, 0x5d, 0x26, 0xa4,
	0xfd, 0xd0, 0x96, 0x49, 0x85, 0x75, 0x5a, 0xc3, 0x44, 0x96, 0x33, 0xac, 0xd2, 0xf0, 0xa7, 0x4f,
	0x2a, 0xd4, 0x6e, 0xd6, 0x3e, 0x5f, 0xce, 0x9d, 0xed, 0x14, 0x98, 0x72, 0xe0, 0x86, 0xee, 0xe9,
	0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xf4, 0xd8, 0x52, 0x4f, 0x01, 0x03, 0x00, 0x00,
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
	dAtA[i] = 0x2a
	if m.OracleUpgradeInfo != nil {
		{
			size, err := m.OracleUpgradeInfo.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x22
	}
	if len(m.OracleUpgrades) > 0 {
		for iNdEx := len(m.OracleUpgrades) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.OracleUpgrades[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.OracleRegistrations) > 0 {
		for iNdEx := len(m.OracleRegistrations) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.OracleRegistrations[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Oracles) > 0 {
		for iNdEx := len(m.Oracles) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Oracles[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.UniqueId) > 0 {
		i -= len(m.UniqueId)
		copy(dAtA[i:], m.UniqueId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.UniqueId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.OraclePubKeyRemoteReport) > 0 {
		i -= len(m.OraclePubKeyRemoteReport)
		copy(dAtA[i:], m.OraclePubKeyRemoteReport)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.OraclePubKeyRemoteReport)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.OraclePublicKey) > 0 {
		i -= len(m.OraclePublicKey)
		copy(dAtA[i:], m.OraclePublicKey)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.OraclePublicKey)))
		i--
		dAtA[i] = 0xa
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
	if len(m.Oracles) > 0 {
		for _, e := range m.Oracles {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.OracleRegistrations) > 0 {
		for _, e := range m.OracleRegistrations {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.OracleUpgrades) > 0 {
		for _, e := range m.OracleUpgrades {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.OracleUpgradeInfo != nil {
		l = m.OracleUpgradeInfo.Size()
		n += 1 + l + sovGenesis(uint64(l))
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
	l = len(m.OraclePublicKey)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.OraclePubKeyRemoteReport)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.UniqueId)
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
				return fmt.Errorf("proto: wrong wireType = %d for field Oracles", wireType)
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
			m.Oracles = append(m.Oracles, Oracle{})
			if err := m.Oracles[len(m.Oracles)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OracleRegistrations", wireType)
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
			m.OracleRegistrations = append(m.OracleRegistrations, OracleRegistration{})
			if err := m.OracleRegistrations[len(m.OracleRegistrations)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OracleUpgrades", wireType)
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
			m.OracleUpgrades = append(m.OracleUpgrades, OracleUpgrade{})
			if err := m.OracleUpgrades[len(m.OracleUpgrades)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OracleUpgradeInfo", wireType)
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
			if m.OracleUpgradeInfo == nil {
				m.OracleUpgradeInfo = &OracleUpgradeInfo{}
			}
			if err := m.OracleUpgradeInfo.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
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
				return fmt.Errorf("proto: wrong wireType = %d for field OraclePublicKey", wireType)
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
			m.OraclePublicKey = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OraclePubKeyRemoteReport", wireType)
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
			m.OraclePubKeyRemoteReport = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UniqueId", wireType)
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
			m.UniqueId = string(dAtA[iNdEx:postIndex])
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
