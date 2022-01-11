// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/market/v2/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

//MsgCreateDeal defines the Msg/CreateDeal request type.
type MsgCreateDeal struct {
	DataSchema           []string    `protobuf:"bytes,1,rep,name=data_schema,json=dataSchema,proto3" json:"data_schema,omitempty"`
	Budget               *types.Coin `protobuf:"bytes,2,opt,name=budget,proto3" json:"budget,omitempty"`
	WantDataCount        uint64      `protobuf:"varint,3,opt,name=want_data_count,json=wantDataCount,proto3" json:"want_data_count,omitempty"`
	TrustedDataValidator []string    `protobuf:"bytes,4,rep,name=trusted_data_validator,json=trustedDataValidator,proto3" json:"trusted_data_validator,omitempty"`
	Owner                string      `protobuf:"bytes,5,opt,name=owner,proto3" json:"owner,omitempty"`
}

func (m *MsgCreateDeal) Reset()         { *m = MsgCreateDeal{} }
func (m *MsgCreateDeal) String() string { return proto.CompactTextString(m) }
func (*MsgCreateDeal) ProtoMessage()    {}
func (*MsgCreateDeal) Descriptor() ([]byte, []int) {
	return fileDescriptor_06be686555d66029, []int{0}
}
func (m *MsgCreateDeal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateDeal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateDeal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateDeal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateDeal.Merge(m, src)
}
func (m *MsgCreateDeal) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateDeal) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateDeal.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateDeal proto.InternalMessageInfo

func (m *MsgCreateDeal) GetDataSchema() []string {
	if m != nil {
		return m.DataSchema
	}
	return nil
}

func (m *MsgCreateDeal) GetBudget() *types.Coin {
	if m != nil {
		return m.Budget
	}
	return nil
}

func (m *MsgCreateDeal) GetWantDataCount() uint64 {
	if m != nil {
		return m.WantDataCount
	}
	return 0
}

func (m *MsgCreateDeal) GetTrustedDataValidator() []string {
	if m != nil {
		return m.TrustedDataValidator
	}
	return nil
}

func (m *MsgCreateDeal) GetOwner() string {
	if m != nil {
		return m.Owner
	}
	return ""
}

//MsgCreateDealResponse defines the Msg/CreateDeal response type.
type MsgCreateDealResponse struct {
	DealId uint64 `protobuf:"varint,1,opt,name=deal_id,json=dealId,proto3" json:"deal_id,omitempty"`
}

func (m *MsgCreateDealResponse) Reset()         { *m = MsgCreateDealResponse{} }
func (m *MsgCreateDealResponse) String() string { return proto.CompactTextString(m) }
func (*MsgCreateDealResponse) ProtoMessage()    {}
func (*MsgCreateDealResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_06be686555d66029, []int{1}
}
func (m *MsgCreateDealResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCreateDealResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCreateDealResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCreateDealResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCreateDealResponse.Merge(m, src)
}
func (m *MsgCreateDealResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgCreateDealResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCreateDealResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCreateDealResponse proto.InternalMessageInfo

func (m *MsgCreateDealResponse) GetDealId() uint64 {
	if m != nil {
		return m.DealId
	}
	return 0
}

func init() {
	proto.RegisterType((*MsgCreateDeal)(nil), "panacea.market.v2.MsgCreateDeal")
	proto.RegisterType((*MsgCreateDealResponse)(nil), "panacea.market.v2.MsgCreateDealResponse")
}

func init() { proto.RegisterFile("panacea/market/v2/tx.proto", fileDescriptor_06be686555d66029) }

var fileDescriptor_06be686555d66029 = []byte{
	// 378 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x92, 0xbf, 0xae, 0xd3, 0x30,
	0x14, 0xc6, 0x6b, 0x7a, 0x6f, 0xd1, 0xf5, 0xd5, 0x15, 0xc2, 0x2a, 0x10, 0x32, 0x84, 0xa8, 0x03,
	0xca, 0x82, 0x4d, 0x03, 0x4f, 0x40, 0xbb, 0x74, 0xe8, 0x12, 0x24, 0x84, 0x58, 0xa2, 0x13, 0xfb,
	0x28, 0x8d, 0x48, 0xe2, 0x28, 0x76, 0xd2, 0xf2, 0x16, 0x3c, 0x16, 0x63, 0x37, 0x18, 0x51, 0xfb,
	0x22, 0x28, 0x7f, 0x2a, 0x51, 0x31, 0xb0, 0xe5, 0x7c, 0xbf, 0x2f, 0xe7, 0x3b, 0x3e, 0x36, 0x75,
	0x2b, 0x28, 0x41, 0x22, 0x88, 0x02, 0xea, 0xaf, 0x68, 0x45, 0x1b, 0x0a, 0x7b, 0xe0, 0x55, 0xad,
	0xad, 0x66, 0x4f, 0x47, 0xc6, 0x07, 0xc6, 0xdb, 0xd0, 0x9d, 0xa7, 0x3a, 0xd5, 0x3d, 0x15, 0xdd,
	0xd7, 0x60, 0x74, 0x3d, 0xa9, 0x4d, 0xa1, 0x8d, 0x48, 0xc0, 0xa0, 0x68, 0x97, 0x09, 0x5a, 0x58,
	0x0a, 0xa9, 0xb3, 0x72, 0xe0, 0x8b, 0x9f, 0x84, 0x3e, 0x6c, 0x4d, 0xba, 0xaa, 0x11, 0x2c, 0xae,
	0x11, 0x72, 0xf6, 0x8a, 0xde, 0x2b, 0xb0, 0x10, 0x1b, 0xb9, 0xc3, 0x02, 0x1c, 0xe2, 0x4f, 0x83,
	0xbb, 0x88, 0x76, 0xd2, 0xc7, 0x5e, 0x61, 0x4b, 0x3a, 0x4b, 0x1a, 0x95, 0xa2, 0x75, 0x1e, 0xf9,
	0x24, 0xb8, 0x0f, 0x5f, 0xf2, 0x21, 0x83, 0x77, 0x19, 0x7c, 0xcc, 0xe0, 0x2b, 0x9d, 0x95, 0xd1,
	0x68, 0x64, 0xaf, 0xe9, 0x93, 0x3d, 0x94, 0x36, 0xee, 0x1b, 0x4b, 0xdd, 0x94, 0xd6, 0x99, 0xfa,
	0x24, 0xb8, 0x89, 0x1e, 0x3a, 0x79, 0x0d, 0x16, 0x56, 0x9d, 0xc8, 0xde, 0xd3, 0xe7, 0xb6, 0x6e,
	0x8c, 0x45, 0x35, 0x58, 0x5b, 0xc8, 0x33, 0x05, 0x56, 0xd7, 0xce, 0x4d, 0x3f, 0xc6, 0x7c, 0xa4,
	0xdd, 0x1f, 0x9f, 0x2e, 0x8c, 0xcd, 0xe9, 0xad, 0xde, 0x97, 0x58, 0x3b, 0xb7, 0x3e, 0x09, 0xee,
	0xa2, 0xa1, 0x58, 0xbc, 0xa5, 0xcf, 0xae, 0x0e, 0x16, 0xa1, 0xa9, 0x74, 0x69, 0x90, 0xbd, 0xa0,
	0x8f, 0x15, 0x42, 0x1e, 0x67, 0xca, 0x21, 0xfd, 0x10, 0xb3, 0xae, 0xdc, 0xa8, 0x30, 0xa6, 0xd3,
	0xad, 0x49, 0xd9, 0x67, 0x4a, 0xff, 0x5a, 0x87, 0xcf, 0xff, 0x59, 0x35, 0xbf, 0xea, 0xeb, 0x06,
	0xff, 0x73, 0x5c, 0x92, 0x3f, 0x6c, 0x7e, 0x9c, 0x3c, 0x72, 0x3c, 0x79, 0xe4, 0xf7, 0xc9, 0x23,
	0xdf, 0xcf, 0xde, 0xe4, 0x78, 0xf6, 0x26, 0xbf, 0xce, 0xde, 0xe4, 0x8b, 0x48, 0x33, 0xbb, 0x6b,
	0x12, 0x2e, 0x75, 0x21, 0x0a, 0x54, 0x59, 0x92, 0x6b, 0x29, 0xc6, 0xb6, 0x6f, 0xa4, 0xae, 0x51,
	0x1c, 0x2e, 0xcf, 0xc0, 0x7e, 0xab, 0xd0, 0x24, 0xb3, 0xfe, 0xfa, 0xde, 0xfd, 0x09, 0x00, 0x00,
	0xff, 0xff, 0x81, 0x07, 0x92, 0xd9, 0x25, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// CreateDeal defines a method for creating a deal.
	CreateDeal(ctx context.Context, in *MsgCreateDeal, opts ...grpc.CallOption) (*MsgCreateDealResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) CreateDeal(ctx context.Context, in *MsgCreateDeal, opts ...grpc.CallOption) (*MsgCreateDealResponse, error) {
	out := new(MsgCreateDealResponse)
	err := c.cc.Invoke(ctx, "/panacea.market.v2.Msg/CreateDeal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// CreateDeal defines a method for creating a deal.
	CreateDeal(context.Context, *MsgCreateDeal) (*MsgCreateDealResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) CreateDeal(ctx context.Context, req *MsgCreateDeal) (*MsgCreateDealResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDeal not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_CreateDeal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateDeal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateDeal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/panacea.market.v2.Msg/CreateDeal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateDeal(ctx, req.(*MsgCreateDeal))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "panacea.market.v2.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateDeal",
			Handler:    _Msg_CreateDeal_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "panacea/market/v2/tx.proto",
}

func (m *MsgCreateDeal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateDeal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateDeal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Owner) > 0 {
		i -= len(m.Owner)
		copy(dAtA[i:], m.Owner)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Owner)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.TrustedDataValidator) > 0 {
		for iNdEx := len(m.TrustedDataValidator) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.TrustedDataValidator[iNdEx])
			copy(dAtA[i:], m.TrustedDataValidator[iNdEx])
			i = encodeVarintTx(dAtA, i, uint64(len(m.TrustedDataValidator[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	if m.WantDataCount != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.WantDataCount))
		i--
		dAtA[i] = 0x18
	}
	if m.Budget != nil {
		{
			size, err := m.Budget.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTx(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.DataSchema) > 0 {
		for iNdEx := len(m.DataSchema) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.DataSchema[iNdEx])
			copy(dAtA[i:], m.DataSchema[iNdEx])
			i = encodeVarintTx(dAtA, i, uint64(len(m.DataSchema[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *MsgCreateDealResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCreateDealResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCreateDealResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.DealId != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.DealId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgCreateDeal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.DataSchema) > 0 {
		for _, s := range m.DataSchema {
			l = len(s)
			n += 1 + l + sovTx(uint64(l))
		}
	}
	if m.Budget != nil {
		l = m.Budget.Size()
		n += 1 + l + sovTx(uint64(l))
	}
	if m.WantDataCount != 0 {
		n += 1 + sovTx(uint64(m.WantDataCount))
	}
	if len(m.TrustedDataValidator) > 0 {
		for _, s := range m.TrustedDataValidator {
			l = len(s)
			n += 1 + l + sovTx(uint64(l))
		}
	}
	l = len(m.Owner)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgCreateDealResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.DealId != 0 {
		n += 1 + sovTx(uint64(m.DealId))
	}
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgCreateDeal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCreateDeal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateDeal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DataSchema", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.DataSchema = append(m.DataSchema, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Budget", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Budget == nil {
				m.Budget = &types.Coin{}
			}
			if err := m.Budget.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field WantDataCount", wireType)
			}
			m.WantDataCount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.WantDataCount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TrustedDataValidator", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TrustedDataValidator = append(m.TrustedDataValidator, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Owner", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Owner = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgCreateDealResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCreateDealResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCreateDealResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DealId", wireType)
			}
			m.DealId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
