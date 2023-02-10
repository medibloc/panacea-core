// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/did/v3/query.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/query"
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

// QueryDIDRequest is the request type for the Query/DIDDocument RPC method.
type QueryDIDRequest struct {
	// NOTE: Using base64 due to the URI path cannot contain colons.
	DidBase64 string `protobuf:"bytes,1,opt,name=did_base64,json=didBase64,proto3" json:"did_base64,omitempty"`
}

func (m *QueryDIDRequest) Reset()         { *m = QueryDIDRequest{} }
func (m *QueryDIDRequest) String() string { return proto.CompactTextString(m) }
func (*QueryDIDRequest) ProtoMessage()    {}
func (*QueryDIDRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1966fee92a319ca6, []int{0}
}
func (m *QueryDIDRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryDIDRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryDIDRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryDIDRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDIDRequest.Merge(m, src)
}
func (m *QueryDIDRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryDIDRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDIDRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDIDRequest proto.InternalMessageInfo

func (m *QueryDIDRequest) GetDidBase64() string {
	if m != nil {
		return m.DidBase64
	}
	return ""
}

// QueryDIDResponse is the response type for the Query/DIDDocument RPC method.
type QueryDIDResponse struct {
	DidDocument *DIDDocument `protobuf:"bytes,1,opt,name=did_document,json=didDocument,proto3" json:"did_document,omitempty"`
}

func (m *QueryDIDResponse) Reset()         { *m = QueryDIDResponse{} }
func (m *QueryDIDResponse) String() string { return proto.CompactTextString(m) }
func (*QueryDIDResponse) ProtoMessage()    {}
func (*QueryDIDResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1966fee92a319ca6, []int{1}
}
func (m *QueryDIDResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryDIDResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryDIDResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryDIDResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDIDResponse.Merge(m, src)
}
func (m *QueryDIDResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryDIDResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDIDResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDIDResponse proto.InternalMessageInfo

func (m *QueryDIDResponse) GetDidDocument() *DIDDocument {
	if m != nil {
		return m.DidDocument
	}
	return nil
}

func init() {
	proto.RegisterType((*QueryDIDRequest)(nil), "panacea.did.v3.QueryDIDRequest")
	proto.RegisterType((*QueryDIDResponse)(nil), "panacea.did.v3.QueryDIDResponse")
}

func init() { proto.RegisterFile("panacea/did/v3/query.proto", fileDescriptor_1966fee92a319ca6) }

var fileDescriptor_1966fee92a319ca6 = []byte{
	// 332 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0x41, 0x4f, 0xf2, 0x40,
	0x10, 0x86, 0xe9, 0xf7, 0x45, 0x13, 0x16, 0xa3, 0xa6, 0x27, 0x52, 0xb5, 0x22, 0x5e, 0xd4, 0x68,
	0x57, 0xc0, 0x78, 0xf4, 0x40, 0xf6, 0x20, 0x37, 0xed, 0xd1, 0x8b, 0xd9, 0x76, 0x27, 0x75, 0x13,
	0xba, 0x53, 0xd8, 0x2d, 0x91, 0xa8, 0x17, 0x7f, 0x81, 0x89, 0x7f, 0xca, 0x23, 0x89, 0x17, 0x8f,
	0x06, 0xfc, 0x21, 0xa6, 0x0b, 0x84, 0x88, 0xf1, 0xd8, 0xce, 0xfb, 0x3c, 0x33, 0xef, 0x12, 0x2f,
	0xe3, 0x8a, 0xc7, 0xc0, 0xa9, 0x90, 0x82, 0x0e, 0x5a, 0xb4, 0x97, 0x43, 0x7f, 0x18, 0x64, 0x7d,
	0x34, 0xe8, 0xae, 0xcf, 0x66, 0x81, 0x90, 0x22, 0x18, 0xb4, 0xbc, 0xed, 0x04, 0x31, 0xe9, 0x02,
	0xe5, 0x99, 0xa4, 0x5c, 0x29, 0x34, 0xdc, 0x48, 0x54, 0x7a, 0x9a, 0xf6, 0x8e, 0x62, 0xd4, 0x29,
	0x6a, 0x1a, 0x71, 0x0d, 0x53, 0x0d, 0x1d, 0x34, 0x22, 0x30, 0xbc, 0x41, 0x33, 0x9e, 0x48, 0x65,
	0xc3, 0xb3, 0x6c, 0x75, 0x69, 0x6b, 0xb1, 0xc0, 0x4e, 0xea, 0xa7, 0x64, 0xe3, 0xba, 0x60, 0x59,
	0x87, 0x85, 0xd0, 0xcb, 0x41, 0x1b, 0x77, 0x87, 0x10, 0x21, 0xc5, 0x6d, 0xe1, 0x3d, 0x3f, 0xab,
	0x3a, 0x35, 0xe7, 0xa0, 0x1c, 0x96, 0x85, 0x14, 0x6d, 0xfb, 0xa3, 0x1e, 0x92, 0xcd, 0x05, 0xa1,
	0x33, 0x54, 0x1a, 0xdc, 0x0b, 0xb2, 0x56, 0x20, 0x02, 0xe3, 0x3c, 0x05, 0x65, 0x2c, 0x54, 0x69,
	0x6e, 0x05, 0x3f, 0x0b, 0x05, 0xac, 0xc3, 0xd8, 0x2c, 0x12, 0x56, 0x84, 0x14, 0xf3, 0x8f, 0xe6,
	0x23, 0x59, 0xb1, 0x4e, 0x57, 0x93, 0xff, 0xac, 0xc3, 0xdc, 0xdd, 0x65, 0x72, 0xe9, 0x46, 0xaf,
	0xf6, 0x77, 0x60, 0x7a, 0x52, 0xfd, 0xf0, 0xf9, 0xfd, 0xeb, 0xf5, 0xdf, 0xbe, 0xbb, 0x47, 0x7f,
	0x77, 0xd7, 0xf4, 0x61, 0xd1, 0xf0, 0xa9, 0x7d, 0xf9, 0x36, 0xf6, 0x9d, 0xd1, 0xd8, 0x77, 0x3e,
	0xc7, 0xbe, 0xf3, 0x32, 0xf1, 0x4b, 0xa3, 0x89, 0x5f, 0xfa, 0x98, 0xf8, 0xa5, 0x2b, 0xe7, 0xe6,
	0x38, 0x91, 0xe6, 0x2e, 0x8f, 0x82, 0x18, 0x53, 0x9a, 0x82, 0x90, 0x51, 0x17, 0xe3, 0xb9, 0xf1,
	0x24, 0xc6, 0x3e, 0xd0, 0x7b, 0x2b, 0x36, 0xc3, 0x0c, 0x74, 0xb4, 0x6a, 0x1f, 0xb5, 0xf5, 0x1d,
	0x00, 0x00, 0xff, 0xff, 0xbc, 0x0d, 0x91, 0x44, 0xe6, 0x01, 0x00, 0x00,
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
	// DID returns a DID Document.
	DID(ctx context.Context, in *QueryDIDRequest, opts ...grpc.CallOption) (*QueryDIDResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) DID(ctx context.Context, in *QueryDIDRequest, opts ...grpc.CallOption) (*QueryDIDResponse, error) {
	out := new(QueryDIDResponse)
	err := c.cc.Invoke(ctx, "/panacea.did.v3.Query/DID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// DID returns a DID Document.
	DID(context.Context, *QueryDIDRequest) (*QueryDIDResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) DID(ctx context.Context, req *QueryDIDRequest) (*QueryDIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DID not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_DID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryDIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).DID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/panacea.did.v3.Query/DID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).DID(ctx, req.(*QueryDIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "panacea.did.v3.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DID",
			Handler:    _Query_DID_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "panacea/did/v3/query.proto",
}

func (m *QueryDIDRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryDIDRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryDIDRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.DidBase64) > 0 {
		i -= len(m.DidBase64)
		copy(dAtA[i:], m.DidBase64)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.DidBase64)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryDIDResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryDIDResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryDIDResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.DidDocument != nil {
		{
			size, err := m.DidDocument.MarshalToSizedBuffer(dAtA[:i])
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
func (m *QueryDIDRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.DidBase64)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryDIDResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.DidDocument != nil {
		l = m.DidDocument.Size()
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
func (m *QueryDIDRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryDIDRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryDIDRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DidBase64", wireType)
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
			m.DidBase64 = string(dAtA[iNdEx:postIndex])
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
func (m *QueryDIDResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryDIDResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryDIDResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DidDocument", wireType)
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
			if m.DidDocument == nil {
				m.DidDocument = &DIDDocument{}
			}
			if err := m.DidDocument.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
