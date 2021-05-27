// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: token/query.proto

package types

import (
	context "context"
	fmt "fmt"
	query "github.com/cosmos/cosmos-sdk/types/query"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// this line is used by starport scaffolding # 3
type QueryGetTokenRequest struct {
	Id uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (m *QueryGetTokenRequest) Reset()         { *m = QueryGetTokenRequest{} }
func (m *QueryGetTokenRequest) String() string { return proto.CompactTextString(m) }
func (*QueryGetTokenRequest) ProtoMessage()    {}
func (*QueryGetTokenRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ec043bcd18c4056e, []int{0}
}
func (m *QueryGetTokenRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryGetTokenRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryGetTokenRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryGetTokenRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryGetTokenRequest.Merge(m, src)
}
func (m *QueryGetTokenRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryGetTokenRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryGetTokenRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryGetTokenRequest proto.InternalMessageInfo

func (m *QueryGetTokenRequest) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

type QueryGetTokenResponse struct {
	Token *Token `protobuf:"bytes,1,opt,name=Token,proto3" json:"Token,omitempty"`
}

func (m *QueryGetTokenResponse) Reset()         { *m = QueryGetTokenResponse{} }
func (m *QueryGetTokenResponse) String() string { return proto.CompactTextString(m) }
func (*QueryGetTokenResponse) ProtoMessage()    {}
func (*QueryGetTokenResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ec043bcd18c4056e, []int{1}
}
func (m *QueryGetTokenResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryGetTokenResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryGetTokenResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryGetTokenResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryGetTokenResponse.Merge(m, src)
}
func (m *QueryGetTokenResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryGetTokenResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryGetTokenResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryGetTokenResponse proto.InternalMessageInfo

func (m *QueryGetTokenResponse) GetToken() *Token {
	if m != nil {
		return m.Token
	}
	return nil
}

type QueryAllTokenRequest struct {
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (m *QueryAllTokenRequest) Reset()         { *m = QueryAllTokenRequest{} }
func (m *QueryAllTokenRequest) String() string { return proto.CompactTextString(m) }
func (*QueryAllTokenRequest) ProtoMessage()    {}
func (*QueryAllTokenRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ec043bcd18c4056e, []int{2}
}
func (m *QueryAllTokenRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryAllTokenRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryAllTokenRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryAllTokenRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryAllTokenRequest.Merge(m, src)
}
func (m *QueryAllTokenRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryAllTokenRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryAllTokenRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryAllTokenRequest proto.InternalMessageInfo

func (m *QueryAllTokenRequest) GetPagination() *query.PageRequest {
	if m != nil {
		return m.Pagination
	}
	return nil
}

type QueryAllTokenResponse struct {
	Token      []*Token            `protobuf:"bytes,1,rep,name=Token,proto3" json:"Token,omitempty"`
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (m *QueryAllTokenResponse) Reset()         { *m = QueryAllTokenResponse{} }
func (m *QueryAllTokenResponse) String() string { return proto.CompactTextString(m) }
func (*QueryAllTokenResponse) ProtoMessage()    {}
func (*QueryAllTokenResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ec043bcd18c4056e, []int{3}
}
func (m *QueryAllTokenResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryAllTokenResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryAllTokenResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryAllTokenResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryAllTokenResponse.Merge(m, src)
}
func (m *QueryAllTokenResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryAllTokenResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryAllTokenResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryAllTokenResponse proto.InternalMessageInfo

func (m *QueryAllTokenResponse) GetToken() []*Token {
	if m != nil {
		return m.Token
	}
	return nil
}

func (m *QueryAllTokenResponse) GetPagination() *query.PageResponse {
	if m != nil {
		return m.Pagination
	}
	return nil
}

func init() {
	proto.RegisterType((*QueryGetTokenRequest)(nil), "medibloc.panaceacore.token.QueryGetTokenRequest")
	proto.RegisterType((*QueryGetTokenResponse)(nil), "medibloc.panaceacore.token.QueryGetTokenResponse")
	proto.RegisterType((*QueryAllTokenRequest)(nil), "medibloc.panaceacore.token.QueryAllTokenRequest")
	proto.RegisterType((*QueryAllTokenResponse)(nil), "medibloc.panaceacore.token.QueryAllTokenResponse")
}

func init() { proto.RegisterFile("token/query.proto", fileDescriptor_ec043bcd18c4056e) }

var fileDescriptor_ec043bcd18c4056e = []byte{
	// 405 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0x4d, 0x4b, 0x1b, 0x41,
	0x18, 0xc7, 0x33, 0xdb, 0xa6, 0x94, 0x29, 0x14, 0x3a, 0xb4, 0x50, 0x96, 0xb2, 0x34, 0x5b, 0x48,
	0xd3, 0x42, 0x67, 0x9a, 0x78, 0xf0, 0x1c, 0x0f, 0xc6, 0x63, 0x0c, 0x9e, 0x3c, 0x08, 0xb3, 0xbb,
	0xc3, 0x3a, 0xb8, 0xd9, 0xd9, 0x64, 0x26, 0x62, 0x10, 0x2f, 0x7e, 0x02, 0xc1, 0x93, 0xe2, 0xb7,
	0xf1, 0xe2, 0x31, 0xe0, 0xc5, 0xa3, 0x24, 0x7e, 0x10, 0xc9, 0xcc, 0x84, 0x6c, 0x12, 0x35, 0xc9,
	0x65, 0x0f, 0x0f, 0xff, 0x97, 0xdf, 0xb3, 0xcf, 0x2e, 0xfc, 0xa2, 0xc4, 0x11, 0x4b, 0x49, 0xa7,
	0xc7, 0xba, 0x7d, 0x9c, 0x75, 0x85, 0x12, 0xc8, 0x6d, 0xb3, 0x88, 0x07, 0x89, 0x08, 0x71, 0x46,
	0x53, 0x1a, 0x32, 0x1a, 0x8a, 0x2e, 0xc3, 0x5a, 0xe7, 0xfe, 0x88, 0x85, 0x88, 0x13, 0x46, 0x68,
	0xc6, 0x09, 0x4d, 0x53, 0xa1, 0xa8, 0xe2, 0x22, 0x95, 0xc6, 0xe9, 0xfe, 0x0d, 0x85, 0x6c, 0x0b,
	0x49, 0x02, 0x2a, 0x99, 0x89, 0x24, 0xc7, 0xd5, 0x80, 0x29, 0x5a, 0x25, 0x19, 0x8d, 0x79, 0xaa,
	0xc5, 0x56, 0x6b, 0x8b, 0xf5, 0xd3, 0x8c, 0xfc, 0x32, 0xfc, 0xba, 0x3b, 0x36, 0x35, 0x98, 0xda,
	0x1b, 0x8f, 0x5b, 0xac, 0xd3, 0x63, 0x52, 0xa1, 0xcf, 0xd0, 0xe1, 0xd1, 0x77, 0xf0, 0x13, 0x54,
	0xde, 0xb7, 0x1c, 0x1e, 0xf9, 0x4d, 0xf8, 0x6d, 0x4e, 0x27, 0x33, 0x91, 0x4a, 0x86, 0x36, 0x61,
	0x51, 0x0f, 0xb4, 0xf6, 0x53, 0xad, 0x84, 0x5f, 0xdf, 0x04, 0x1b, 0xa7, 0xd1, 0xfb, 0x07, 0xb6,
	0xb9, 0x9e, 0x24, 0x33, 0xcd, 0xdb, 0x10, 0x4e, 0xc1, 0x6d, 0x6a, 0x19, 0x9b, 0x2d, 0xf1, 0x78,
	0x4b, 0x6c, 0x5e, 0x9c, 0xdd, 0x12, 0x37, 0x69, 0xcc, 0xac, 0xb7, 0x95, 0x73, 0xfa, 0x57, 0xc0,
	0x22, 0x4f, 0x0b, 0x16, 0x91, 0xdf, 0xad, 0x83, 0x8c, 0x1a, 0x33, 0x68, 0x8e, 0x46, 0xfb, 0xbd,
	0x14, 0xcd, 0xb4, 0xe6, 0xd9, 0x6a, 0xb7, 0x0e, 0x2c, 0x6a, 0x36, 0x74, 0x03, 0x2c, 0x0c, 0xfa,
	0xff, 0x16, 0xc6, 0x4b, 0x37, 0x72, 0xab, 0x6b, 0x38, 0x0c, 0x84, 0x8f, 0xcf, 0xef, 0x9f, 0x2e,
	0x9d, 0x0a, 0x2a, 0x93, 0x89, 0x95, 0xe4, 0xac, 0x24, 0xf7, 0x7d, 0x90, 0x53, 0x1e, 0x9d, 0xa1,
	0x6b, 0x00, 0x3f, 0xea, 0x84, 0x7a, 0x92, 0xac, 0x40, 0x38, 0x77, 0xcb, 0x15, 0x08, 0xe7, 0x8f,
	0xe3, 0xff, 0xd1, 0x84, 0xbf, 0x50, 0x69, 0x29, 0xe1, 0xd6, 0xce, 0xdd, 0xd0, 0x03, 0x83, 0xa1,
	0x07, 0x1e, 0x87, 0x1e, 0xb8, 0x18, 0x79, 0x85, 0xc1, 0xc8, 0x2b, 0x3c, 0x8c, 0xbc, 0xc2, 0x3e,
	0x8e, 0xb9, 0x3a, 0xec, 0x05, 0x38, 0x14, 0xed, 0x85, 0x98, 0x7f, 0x3a, 0xe7, 0x64, 0x92, 0xd4,
	0xcf, 0x98, 0x0c, 0x3e, 0xe8, 0x9f, 0x61, 0xe3, 0x39, 0x00, 0x00, 0xff, 0xff, 0x15, 0x3b, 0x36,
	0x6c, 0x9a, 0x03, 0x00, 0x00,
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
	// this line is used by starport scaffolding # 2
	Token(ctx context.Context, in *QueryGetTokenRequest, opts ...grpc.CallOption) (*QueryGetTokenResponse, error)
	TokenAll(ctx context.Context, in *QueryAllTokenRequest, opts ...grpc.CallOption) (*QueryAllTokenResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Token(ctx context.Context, in *QueryGetTokenRequest, opts ...grpc.CallOption) (*QueryGetTokenResponse, error) {
	out := new(QueryGetTokenResponse)
	err := c.cc.Invoke(ctx, "/medibloc.panaceacore.token.Query/Token", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) TokenAll(ctx context.Context, in *QueryAllTokenRequest, opts ...grpc.CallOption) (*QueryAllTokenResponse, error) {
	out := new(QueryAllTokenResponse)
	err := c.cc.Invoke(ctx, "/medibloc.panaceacore.token.Query/TokenAll", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// this line is used by starport scaffolding # 2
	Token(context.Context, *QueryGetTokenRequest) (*QueryGetTokenResponse, error)
	TokenAll(context.Context, *QueryAllTokenRequest) (*QueryAllTokenResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) Token(ctx context.Context, req *QueryGetTokenRequest) (*QueryGetTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Token not implemented")
}
func (*UnimplementedQueryServer) TokenAll(ctx context.Context, req *QueryAllTokenRequest) (*QueryAllTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TokenAll not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Token_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryGetTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Token(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/medibloc.panaceacore.token.Query/Token",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Token(ctx, req.(*QueryGetTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_TokenAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAllTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).TokenAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/medibloc.panaceacore.token.Query/TokenAll",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).TokenAll(ctx, req.(*QueryAllTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "medibloc.panaceacore.token.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Token",
			Handler:    _Query_Token_Handler,
		},
		{
			MethodName: "TokenAll",
			Handler:    _Query_TokenAll_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "token/query.proto",
}

func (m *QueryGetTokenRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryGetTokenRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryGetTokenRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Id != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryGetTokenResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryGetTokenResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryGetTokenResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Token != nil {
		{
			size, err := m.Token.MarshalToSizedBuffer(dAtA[:i])
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

func (m *QueryAllTokenRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryAllTokenRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryAllTokenRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Pagination != nil {
		{
			size, err := m.Pagination.MarshalToSizedBuffer(dAtA[:i])
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

func (m *QueryAllTokenResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryAllTokenResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryAllTokenResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Pagination != nil {
		{
			size, err := m.Pagination.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintQuery(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.Token) > 0 {
		for iNdEx := len(m.Token) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Token[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
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
func (m *QueryGetTokenRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovQuery(uint64(m.Id))
	}
	return n
}

func (m *QueryGetTokenResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Token != nil {
		l = m.Token.Size()
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryAllTokenRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Pagination != nil {
		l = m.Pagination.Size()
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryAllTokenResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Token) > 0 {
		for _, e := range m.Token {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	if m.Pagination != nil {
		l = m.Pagination.Size()
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
func (m *QueryGetTokenRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryGetTokenRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryGetTokenRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint64(b&0x7F) << shift
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
func (m *QueryGetTokenResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryGetTokenResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryGetTokenResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token", wireType)
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
			if m.Token == nil {
				m.Token = &Token{}
			}
			if err := m.Token.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryAllTokenRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryAllTokenRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryAllTokenRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pagination", wireType)
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
			if m.Pagination == nil {
				m.Pagination = &query.PageRequest{}
			}
			if err := m.Pagination.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryAllTokenResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryAllTokenResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryAllTokenResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Token", wireType)
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
			m.Token = append(m.Token, &Token{})
			if err := m.Token[len(m.Token)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pagination", wireType)
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
			if m.Pagination == nil {
				m.Pagination = &query.PageResponse{}
			}
			if err := m.Pagination.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
