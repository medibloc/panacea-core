// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: panacea/market/v2/query.proto

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
	return fileDescriptor_b4f233a9a9d75b16, []int{0}
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
	return fileDescriptor_b4f233a9a9d75b16, []int{1}
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
	proto.RegisterType((*QueryDealRequest)(nil), "panacea.market.v2.QueryDealRequest")
	proto.RegisterType((*QueryDealResponse)(nil), "panacea.market.v2.QueryDealResponse")
}

func init() { proto.RegisterFile("panacea/market/v2/query.proto", fileDescriptor_b4f233a9a9d75b16) }

var fileDescriptor_b4f233a9a9d75b16 = []byte{
	// 342 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x50, 0xc1, 0x4a, 0xeb, 0x50,
	0x10, 0x6d, 0x1e, 0x7d, 0x7d, 0x90, 0xb7, 0x79, 0x0d, 0x0f, 0xaa, 0xa1, 0x86, 0x12, 0x5d, 0x14,
	0x8b, 0x19, 0x1a, 0x7f, 0x40, 0xc4, 0x4d, 0x97, 0x76, 0xe9, 0x46, 0x6e, 0x92, 0x31, 0x06, 0x93,
	0x3b, 0x69, 0xee, 0x4d, 0xb1, 0x14, 0x37, 0xe2, 0x07, 0x08, 0xfe, 0x94, 0xcb, 0x82, 0x1b, 0x97,
	0xd2, 0xfa, 0x21, 0x72, 0x6f, 0x12, 0x10, 0x2b, 0xee, 0x66, 0xe6, 0x9c, 0x39, 0x67, 0xe6, 0x98,
	0x7b, 0x39, 0xe3, 0x2c, 0x44, 0x06, 0x19, 0x2b, 0x6e, 0x50, 0xc2, 0xdc, 0x87, 0x59, 0x89, 0xc5,
	0xc2, 0xcb, 0x0b, 0x92, 0x64, 0x75, 0x6b, 0xd8, 0xab, 0x60, 0x6f, 0xee, 0xdb, 0xfd, 0x98, 0x28,
	0x4e, 0x11, 0x58, 0x9e, 0x00, 0xe3, 0x9c, 0x24, 0x93, 0x09, 0x71, 0x51, 0x2d, 0xd8, 0x87, 0x21,
	0x89, 0x8c, 0x04, 0x04, 0x4c, 0x60, 0xa5, 0x04, 0xf3, 0x71, 0x80, 0x92, 0x8d, 0x21, 0x67, 0x71,
	0xc2, 0x35, 0xb9, 0xe6, 0xf6, 0xb7, 0xbd, 0x23, 0x64, 0x69, 0x8d, 0xee, 0xd6, 0x3e, 0xba, 0x0b,
	0xca, 0x2b, 0x60, 0xbc, 0xbe, 0xca, 0xfe, 0x1f, 0x53, 0x4c, 0xba, 0x04, 0x55, 0x55, 0x53, 0x77,
	0x64, 0xfe, 0x3b, 0x57, 0x86, 0x67, 0xc8, 0xd2, 0x29, 0xce, 0x4a, 0x14, 0xd2, 0xea, 0x99, 0x7f,
	0x94, 0xe4, 0x65, 0x12, 0xed, 0x18, 0x03, 0x63, 0xd8, 0x9e, 0x76, 0x54, 0x3b, 0x89, 0xdc, 0x13,
	0xb3, 0xfb, 0x89, 0x2c, 0x72, 0xe2, 0x02, 0xad, 0x91, 0xd9, 0x56, 0xb0, 0xa6, 0xfe, 0xf5, 0x7b,
	0xde, 0xd6, 0xf3, 0x9e, 0xa6, 0x6b, 0x92, 0xff, 0x60, 0x98, 0xbf, 0xb5, 0x84, 0xb5, 0x34, 0xdb,
	0x6a, 0x6e, 0xed, 0x7f, 0xb3, 0xf0, 0xf5, 0x22, 0xfb, 0xe0, 0x67, 0x52, 0x75, 0x89, 0x3b, 0xbc,
	0x7f, 0x79, 0x7f, 0xfa, 0xe5, 0x5a, 0x03, 0x68, 0x32, 0x52, 0x9e, 0x4d, 0x42, 0x02, 0x96, 0xf5,
	0x57, 0x77, 0xa7, 0x93, 0xe7, 0xb5, 0x63, 0xac, 0xd6, 0x8e, 0xf1, 0xb6, 0x76, 0x8c, 0xc7, 0x8d,
	0xd3, 0x5a, 0x6d, 0x9c, 0xd6, 0xeb, 0xc6, 0x69, 0x5d, 0x40, 0x9c, 0xc8, 0xeb, 0x32, 0xf0, 0x42,
	0xca, 0x20, 0xc3, 0x28, 0x09, 0x52, 0x0a, 0x1b, 0xb9, 0xa3, 0x90, 0x0a, 0x84, 0xdb, 0x26, 0x79,
	0xb9, 0xc8, 0x51, 0x04, 0x1d, 0x9d, 0xe3, 0xf1, 0x47, 0x00, 0x00, 0x00, 0xff, 0xff, 0x35, 0x23,
	0x24, 0x67, 0x14, 0x02, 0x00, 0x00,
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
	err := c.cc.Invoke(ctx, "/panacea.market.v2.Query/Deal", in, out, opts...)
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
		FullMethod: "/panacea.market.v2.Query/Deal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Deal(ctx, req.(*QueryDealRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "panacea.market.v2.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Deal",
			Handler:    _Query_Deal_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "panacea/market/v2/query.proto",
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