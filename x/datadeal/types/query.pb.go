// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/datadeal/v2/query.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/cosmos-sdk/types/query"
	_ "github.com/gogo/protobuf/gogoproto"
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

// QueryDealRequest is the request type for Query/Deal RPC method.
type QueryDealRequest struct {
	DealId uint64 `protobuf:"varint,1,opt,name=deal_id,json=dealId,proto3" json:"deal_id,omitempty"`
}

func (m *QueryDealRequest) Reset()         { *m = QueryDealRequest{} }
func (m *QueryDealRequest) String() string { return proto.CompactTextString(m) }
func (*QueryDealRequest) ProtoMessage()    {}
func (*QueryDealRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c7a445ecc4b9161, []int{0}
}
func (m *QueryDealRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryDealRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryDealRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryDealRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDealRequest.Merge(m, src)
}
func (m *QueryDealRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryDealRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDealRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDealRequest proto.InternalMessageInfo

func (m *QueryDealRequest) GetDealId() uint64 {
	if m != nil {
		return m.DealId
	}
	return 0
}

// QueryDealResponse is the response type for the Query/Deal RPC method.
type QueryDealResponse struct {
	Deal *Deal `protobuf:"bytes,1,opt,name=deal,proto3" json:"deal,omitempty"`
}

func (m *QueryDealResponse) Reset()         { *m = QueryDealResponse{} }
func (m *QueryDealResponse) String() string { return proto.CompactTextString(m) }
func (*QueryDealResponse) ProtoMessage()    {}
func (*QueryDealResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4c7a445ecc4b9161, []int{1}
}
func (m *QueryDealResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryDealResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryDealResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryDealResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDealResponse.Merge(m, src)
}
func (m *QueryDealResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryDealResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDealResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDealResponse proto.InternalMessageInfo

func (m *QueryDealResponse) GetDeal() *Deal {
	if m != nil {
		return m.Deal
	}
	return nil
}

func init() {
	proto.RegisterType((*QueryDealRequest)(nil), "panacea.datadeal.v2.QueryDealRequest")
	proto.RegisterType((*QueryDealResponse)(nil), "panacea.datadeal.v2.QueryDealResponse")
}

func init() { proto.RegisterFile("panacea/datadeal/v2/query.proto", fileDescriptor_4c7a445ecc4b9161) }

var fileDescriptor_4c7a445ecc4b9161 = []byte{
	// 341 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x90, 0x41, 0x4b, 0xeb, 0x40,
	0x10, 0xc7, 0xbb, 0x8f, 0xbe, 0x3e, 0xd8, 0x77, 0x79, 0x2f, 0x0a, 0xda, 0x20, 0xab, 0x14, 0x2d,
	0xa2, 0x36, 0x43, 0xe3, 0x37, 0x28, 0x5e, 0xc4, 0x93, 0x3d, 0x7a, 0x91, 0x49, 0xb2, 0xc6, 0x40,
	0xba, 0x93, 0x76, 0x37, 0xc5, 0x22, 0x5e, 0xf4, 0x2c, 0x08, 0x7e, 0x29, 0x8f, 0x05, 0x2f, 0x1e,
	0xa5, 0xf5, 0x83, 0x48, 0x36, 0x29, 0x8a, 0x54, 0xbc, 0xcd, 0xec, 0xfc, 0xe6, 0xff, 0x9f, 0xfd,
	0xf3, 0xcd, 0x0c, 0x15, 0x86, 0x12, 0x21, 0x42, 0x83, 0x91, 0xc4, 0x14, 0xc6, 0x3e, 0x0c, 0x73,
	0x39, 0x9a, 0x78, 0xd9, 0x88, 0x0c, 0x39, 0x2b, 0x15, 0xe0, 0x2d, 0x00, 0x6f, 0xec, 0xbb, 0x1b,
	0x31, 0x51, 0x9c, 0x4a, 0xc0, 0x2c, 0x01, 0x54, 0x8a, 0x0c, 0x9a, 0x84, 0x94, 0x2e, 0x57, 0xdc,
	0xbd, 0x90, 0xf4, 0x80, 0x34, 0x04, 0xa8, 0x65, 0xa9, 0x05, 0xe3, 0x6e, 0x20, 0x0d, 0x76, 0x21,
	0xc3, 0x38, 0x51, 0x16, 0xae, 0x58, 0xb1, 0xcc, 0xdf, 0xda, 0x94, 0xf3, 0x66, 0xe5, 0x64, 0xbb,
	0x20, 0xbf, 0x00, 0x54, 0xd5, 0x65, 0xee, 0x6a, 0x4c, 0x31, 0xd9, 0x12, 0x8a, 0xaa, 0x7c, 0x6d,
	0xed, 0xf3, 0x7f, 0xa7, 0x85, 0xe5, 0x91, 0xc4, 0xb4, 0x2f, 0x87, 0xb9, 0xd4, 0xc6, 0x59, 0xe3,
	0x7f, 0x0a, 0xc9, 0xf3, 0x24, 0x5a, 0x67, 0x5b, 0x6c, 0xb7, 0xde, 0x6f, 0x14, 0xed, 0x71, 0xd4,
	0xea, 0xf1, 0xff, 0x9f, 0x60, 0x9d, 0x91, 0xd2, 0xd2, 0xe9, 0xf0, 0x7a, 0x31, 0xb6, 0xe8, 0x5f,
	0xbf, 0xe9, 0x2d, 0x09, 0xc0, 0xb3, 0x0b, 0x16, 0xf3, 0xef, 0x19, 0xff, 0x6d, 0x45, 0x9c, 0x3b,
	0xc6, 0xeb, 0xc5, 0xc0, 0xd9, 0x59, 0xba, 0xf3, 0xf5, 0x2c, 0xb7, 0xfd, 0x13, 0x56, 0x1e, 0xd4,
	0x3a, 0xb8, 0x7d, 0x7e, 0x7b, 0xfc, 0xd5, 0x76, 0xb6, 0xe1, 0xbb, 0xb0, 0x34, 0x5c, 0x57, 0x1f,
	0xbc, 0xe9, 0x9d, 0x3c, 0xcd, 0x04, 0x9b, 0xce, 0x04, 0x7b, 0x9d, 0x09, 0xf6, 0x30, 0x17, 0xb5,
	0xe9, 0x5c, 0xd4, 0x5e, 0xe6, 0xa2, 0x76, 0xd6, 0x8d, 0x13, 0x73, 0x99, 0x07, 0x5e, 0x48, 0x03,
	0x18, 0xc8, 0x28, 0x09, 0x52, 0x0a, 0x17, 0x92, 0x9d, 0x90, 0x46, 0x12, 0xae, 0x3e, 0x94, 0xcd,
	0x24, 0x93, 0x3a, 0x68, 0xd8, 0x50, 0x0f, 0xdf, 0x03, 0x00, 0x00, 0xff, 0xff, 0x9b, 0x53, 0x16,
	0xd6, 0x27, 0x02, 0x00, 0x00,
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
	// Deal returns a Deal.
	Deal(ctx context.Context, in *QueryDealRequest, opts ...grpc.CallOption) (*QueryDealResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Deal(ctx context.Context, in *QueryDealRequest, opts ...grpc.CallOption) (*QueryDealResponse, error) {
	out := new(QueryDealResponse)
	err := c.cc.Invoke(ctx, "/panacea.datadeal.v2.Query/Deal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// Deal returns a Deal.
	Deal(context.Context, *QueryDealRequest) (*QueryDealResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) Deal(ctx context.Context, req *QueryDealRequest) (*QueryDealResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Deal not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Deal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryDealRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Deal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/panacea.datadeal.v2.Query/Deal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Deal(ctx, req.(*QueryDealRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "panacea.datadeal.v2.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Deal",
			Handler:    _Query_Deal_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "panacea/datadeal/v2/query.proto",
}

func (m *QueryDealRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryDealRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryDealRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.DealId != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.DealId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryDealResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryDealResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryDealResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Deal != nil {
		{
			size, err := m.Deal.MarshalToSizedBuffer(dAtA[:i])
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
func (m *QueryDealRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.DealId != 0 {
		n += 1 + sovQuery(uint64(m.DealId))
	}
	return n
}

func (m *QueryDealResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Deal != nil {
		l = m.Deal.Size()
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
func (m *QueryDealRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryDealRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryDealRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DealId", wireType)
			}
			m.DealId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
func (m *QueryDealResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryDealResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryDealResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Deal", wireType)
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
			if m.Deal == nil {
				m.Deal = &Deal{}
			}
			if err := m.Deal.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
