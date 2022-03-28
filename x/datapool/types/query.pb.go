// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/datapool/v2/query.proto

package types

import (
	context "context"
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// QueryPoolRequest is the request type for the Query/Pool RPC method.
type QueryPoolRequest struct {
	PoolId uint64 `protobuf:"varint,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
}

func (m *QueryPoolRequest) Reset()         { *m = QueryPoolRequest{} }
func (m *QueryPoolRequest) String() string { return proto.CompactTextString(m) }
func (*QueryPoolRequest) ProtoMessage()    {}
func (*QueryPoolRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f3fda93d2b4f4508, []int{0}
}
func (m *QueryPoolRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPoolRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPoolRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPoolRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPoolRequest.Merge(m, src)
}
func (m *QueryPoolRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryPoolRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPoolRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPoolRequest proto.InternalMessageInfo

func (m *QueryPoolRequest) GetPoolId() uint64 {
	if m != nil {
		return m.PoolId
	}
	return 0
}

// QueryPoolResponse is the response type for the Query/Pool RPC method.
type QueryPoolResponse struct {
	Pool *Pool `protobuf:"bytes,1,opt,name=pool,proto3" json:"pool,omitempty"`
}

func (m *QueryPoolResponse) Reset()         { *m = QueryPoolResponse{} }
func (m *QueryPoolResponse) String() string { return proto.CompactTextString(m) }
func (*QueryPoolResponse) ProtoMessage()    {}
func (*QueryPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f3fda93d2b4f4508, []int{1}
}
func (m *QueryPoolResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPoolResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPoolResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPoolResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPoolResponse.Merge(m, src)
}
func (m *QueryPoolResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryPoolResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPoolResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPoolResponse proto.InternalMessageInfo

func (m *QueryPoolResponse) GetPool() *Pool {
	if m != nil {
		return m.Pool
	}
	return nil
}

// QueryContractRequest
type QueryContractRequest struct {
}

func (m *QueryContractRequest) Reset()         { *m = QueryContractRequest{} }
func (m *QueryContractRequest) String() string { return proto.CompactTextString(m) }
func (*QueryContractRequest) ProtoMessage()    {}
func (*QueryContractRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f3fda93d2b4f4508, []int{2}
}
func (m *QueryContractRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryContractRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryContractRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryContractRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryContractRequest.Merge(m, src)
}
func (m *QueryContractRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryContractRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryContractRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryContractRequest proto.InternalMessageInfo

// QueryContractResponse
type QueryContractResponse struct {
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
}

func (m *QueryContractResponse) Reset()         { *m = QueryContractResponse{} }
func (m *QueryContractResponse) String() string { return proto.CompactTextString(m) }
func (*QueryContractResponse) ProtoMessage()    {}
func (*QueryContractResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f3fda93d2b4f4508, []int{3}
}
func (m *QueryContractResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryContractResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryContractResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryContractResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryContractResponse.Merge(m, src)
}
func (m *QueryContractResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryContractResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryContractResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryContractResponse proto.InternalMessageInfo

func (m *QueryContractResponse) GetContractAddress() string {
	if m != nil {
		return m.ContractAddress
	}
	return ""
}

func init() {
	proto.RegisterType((*QueryPoolRequest)(nil), "panacea.datapool.v2.QueryPoolRequest")
	proto.RegisterType((*QueryPoolResponse)(nil), "panacea.datapool.v2.QueryPoolResponse")
	proto.RegisterType((*QueryContractRequest)(nil), "panacea.datapool.v2.QueryContractRequest")
	proto.RegisterType((*QueryContractResponse)(nil), "panacea.datapool.v2.QueryContractResponse")
}

func init() { proto.RegisterFile("panacea/datapool/v2/query.proto", fileDescriptor_f3fda93d2b4f4508) }

var fileDescriptor_f3fda93d2b4f4508 = []byte{
	// 366 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xcd, 0x4e, 0xc2, 0x40,
	0x14, 0x85, 0x29, 0x41, 0xd4, 0x71, 0x21, 0x8e, 0xff, 0x8d, 0x19, 0x4d, 0x23, 0x28, 0x2a, 0x9d,
	0x58, 0x9f, 0x40, 0x5c, 0x19, 0x37, 0xca, 0xd2, 0x0d, 0x19, 0xda, 0x09, 0x36, 0x29, 0xbd, 0xa5,
	0x33, 0x10, 0x89, 0x71, 0xa3, 0x2f, 0x40, 0xe2, 0x0b, 0xf8, 0x38, 0x2e, 0x49, 0xdc, 0xb8, 0x34,
	0xe0, 0x83, 0x98, 0x0e, 0x6d, 0x34, 0x4d, 0x89, 0x2e, 0x7b, 0xe7, 0xbb, 0xe7, 0x9c, 0x7b, 0x52,
	0xb4, 0x1b, 0x30, 0x9f, 0xd9, 0x9c, 0x51, 0x87, 0x49, 0x16, 0x00, 0x78, 0xb4, 0x6f, 0xd1, 0x6e,
	0x8f, 0x87, 0x03, 0x33, 0x08, 0x41, 0x02, 0x5e, 0x8d, 0x01, 0x33, 0x01, 0xcc, 0xbe, 0xa5, 0xef,
	0xb4, 0x01, 0xda, 0x1e, 0xa7, 0x2c, 0x70, 0x29, 0xf3, 0x7d, 0x90, 0x4c, 0xba, 0xe0, 0x8b, 0xe9,
	0x8a, 0x4e, 0xb2, 0x34, 0xd5, 0xaa, 0x7a, 0x37, 0x8e, 0x51, 0xe9, 0x26, 0x72, 0xb8, 0x06, 0xf0,
	0x1a, 0xbc, 0xdb, 0xe3, 0x42, 0xe2, 0x4d, 0x34, 0x1f, 0x11, 0x4d, 0xd7, 0xd9, 0xd2, 0xf6, 0xb4,
	0xc3, 0x42, 0xa3, 0x18, 0x7d, 0x5e, 0x3a, 0x46, 0x1d, 0xad, 0xfc, 0x82, 0x45, 0x00, 0xbe, 0xe0,
	0xb8, 0x86, 0x0a, 0xd1, 0xb3, 0x42, 0x97, 0xac, 0x6d, 0x33, 0x23, 0xa3, 0xa9, 0x16, 0x14, 0x66,
	0x6c, 0xa0, 0x35, 0xa5, 0x71, 0x01, 0xbe, 0x0c, 0x99, 0x2d, 0x63, 0x53, 0xa3, 0x8e, 0xd6, 0x53,
	0xf3, 0x58, 0xbf, 0x8a, 0x4a, 0x76, 0x3c, 0x6b, 0x32, 0xc7, 0x09, 0xb9, 0x10, 0xca, 0x6b, 0xb1,
	0xb1, 0x9c, 0xcc, 0xcf, 0xa7, 0x63, 0xeb, 0x35, 0x8f, 0xe6, 0x94, 0x08, 0x7e, 0xd6, 0x50, 0x21,
	0x32, 0xc5, 0xe5, 0xcc, 0x3c, 0xe9, 0x93, 0xf5, 0xca, 0x5f, 0xd8, 0x34, 0x8c, 0x71, 0xf2, 0xf4,
	0xfe, 0xf5, 0x92, 0xaf, 0xe0, 0x7d, 0x3a, 0xab, 0x57, 0x41, 0x1f, 0xe2, 0xf2, 0x1e, 0xf1, 0x50,
	0x43, 0x0b, 0xc9, 0x3d, 0xb8, 0x3a, 0xdb, 0x22, 0xd5, 0x85, 0x7e, 0xf4, 0x1f, 0x34, 0x4e, 0x54,
	0x53, 0x89, 0x0e, 0x70, 0x39, 0x33, 0x51, 0xba, 0xb9, 0xfa, 0xd5, 0xdb, 0x98, 0x68, 0xa3, 0x31,
	0xd1, 0x3e, 0xc7, 0x44, 0x1b, 0x4e, 0x48, 0x6e, 0x34, 0x21, 0xb9, 0x8f, 0x09, 0xc9, 0xdd, 0x9e,
	0xb6, 0x5d, 0x79, 0xd7, 0x6b, 0x99, 0x36, 0x74, 0x68, 0x87, 0x3b, 0x6e, 0xcb, 0x03, 0x3b, 0xd1,
	0xac, 0xd9, 0x10, 0x72, 0x7a, 0xff, 0x23, 0x2d, 0x07, 0x01, 0x17, 0xad, 0xa2, 0xfa, 0x87, 0xce,
	0xbe, 0x03, 0x00, 0x00, 0xff, 0xff, 0x69, 0x96, 0xe7, 0x6e, 0xb9, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// Pool returns a Pool.
	Pool(ctx context.Context, in *QueryPoolRequest, opts ...grpc.CallOption) (*QueryPoolResponse, error)
	// Contract returns a contract address registered
	Contract(ctx context.Context, in *QueryContractRequest, opts ...grpc.CallOption) (*QueryContractResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Pool(ctx context.Context, in *QueryPoolRequest, opts ...grpc.CallOption) (*QueryPoolResponse, error) {
	out := new(QueryPoolResponse)
	err := c.cc.Invoke(ctx, "/panacea.datapool.v2.Query/Pool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Contract(ctx context.Context, in *QueryContractRequest, opts ...grpc.CallOption) (*QueryContractResponse, error) {
	out := new(QueryContractResponse)
	err := c.cc.Invoke(ctx, "/panacea.datapool.v2.Query/Contract", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Pool returns a Pool.
	Pool(context.Context, *QueryPoolRequest) (*QueryPoolResponse, error)
	// Contract returns a contract address registered
	Contract(context.Context, *QueryContractRequest) (*QueryContractResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) Pool(ctx context.Context, req *QueryPoolRequest) (*QueryPoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Pool not implemented")
}
func (*UnimplementedQueryServer) Contract(ctx context.Context, req *QueryContractRequest) (*QueryContractResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Contract not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Pool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPoolRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Pool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/panacea.datapool.v2.Query/Pool",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Pool(ctx, req.(*QueryPoolRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Contract_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryContractRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Contract(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/panacea.datapool.v2.Query/Contract",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Contract(ctx, req.(*QueryContractRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "panacea.datapool.v2.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Pool",
			Handler:    _Query_Pool_Handler,
		},
		{
			MethodName: "Contract",
			Handler:    _Query_Contract_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "panacea/datapool/v2/query.proto",
}

func (m *QueryPoolRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPoolRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPoolRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.PoolId != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.PoolId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryPoolResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPoolResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPoolResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Pool != nil {
		{
			size, err := m.Pool.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintQuery(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryContractRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryContractRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryContractRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryContractResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryContractResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryContractResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ContractAddress) > 0 {
		i -= len(m.ContractAddress)
		copy(dAtA[i:], m.ContractAddress)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ContractAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryPoolRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PoolId != 0 {
		n += 1 + sovQuery(uint64(m.PoolId))
	}
	return n
}

func (m *QueryPoolResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Pool != nil {
		l = m.Pool.Size()
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryContractRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryContractResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContractAddress)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryPoolRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryPoolRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPoolRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolId", wireType)
			}
			m.PoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PoolId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryPoolResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryPoolResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPoolResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pool", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Pool == nil {
				m.Pool = &Pool{}
			}
			if err := m.Pool.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryContractRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryContractRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryContractRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryContractResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryContractResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryContractResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
