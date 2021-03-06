// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pmap.proto

package types // import "github.com/aergoio/aergo/types"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// query to aergomap
type MapQuery struct {
	Status               *Status  `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	AddMe                bool     `protobuf:"varint,2,opt,name=addMe,proto3" json:"addMe,omitempty"`
	Size                 int32    `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	Excludes             [][]byte `protobuf:"bytes,4,rep,name=excludes,proto3" json:"excludes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MapQuery) Reset()         { *m = MapQuery{} }
func (m *MapQuery) String() string { return proto.CompactTextString(m) }
func (*MapQuery) ProtoMessage()    {}
func (*MapQuery) Descriptor() ([]byte, []int) {
	return fileDescriptor_pmap_a3be564b6f5f4c34, []int{0}
}
func (m *MapQuery) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MapQuery.Unmarshal(m, b)
}
func (m *MapQuery) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MapQuery.Marshal(b, m, deterministic)
}
func (dst *MapQuery) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MapQuery.Merge(dst, src)
}
func (m *MapQuery) XXX_Size() int {
	return xxx_messageInfo_MapQuery.Size(m)
}
func (m *MapQuery) XXX_DiscardUnknown() {
	xxx_messageInfo_MapQuery.DiscardUnknown(m)
}

var xxx_messageInfo_MapQuery proto.InternalMessageInfo

func (m *MapQuery) GetStatus() *Status {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *MapQuery) GetAddMe() bool {
	if m != nil {
		return m.AddMe
	}
	return false
}

func (m *MapQuery) GetSize() int32 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *MapQuery) GetExcludes() [][]byte {
	if m != nil {
		return m.Excludes
	}
	return nil
}

type MapResponse struct {
	Status               ResultStatus   `protobuf:"varint,1,opt,name=status,proto3,enum=types.ResultStatus" json:"status,omitempty"`
	Addresses            []*PeerAddress `protobuf:"bytes,2,rep,name=addresses,proto3" json:"addresses,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *MapResponse) Reset()         { *m = MapResponse{} }
func (m *MapResponse) String() string { return proto.CompactTextString(m) }
func (*MapResponse) ProtoMessage()    {}
func (*MapResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_pmap_a3be564b6f5f4c34, []int{1}
}
func (m *MapResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MapResponse.Unmarshal(m, b)
}
func (m *MapResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MapResponse.Marshal(b, m, deterministic)
}
func (dst *MapResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MapResponse.Merge(dst, src)
}
func (m *MapResponse) XXX_Size() int {
	return xxx_messageInfo_MapResponse.Size(m)
}
func (m *MapResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MapResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MapResponse proto.InternalMessageInfo

func (m *MapResponse) GetStatus() ResultStatus {
	if m != nil {
		return m.Status
	}
	return ResultStatus_OK
}

func (m *MapResponse) GetAddresses() []*PeerAddress {
	if m != nil {
		return m.Addresses
	}
	return nil
}

func init() {
	proto.RegisterType((*MapQuery)(nil), "types.MapQuery")
	proto.RegisterType((*MapResponse)(nil), "types.MapResponse")
}

func init() { proto.RegisterFile("pmap.proto", fileDescriptor_pmap_a3be564b6f5f4c34) }

var fileDescriptor_pmap_a3be564b6f5f4c34 = []byte{
	// 236 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x50, 0x4f, 0x4b, 0xc3, 0x30,
	0x14, 0xa7, 0xeb, 0x3a, 0xba, 0x57, 0xf5, 0x10, 0x3d, 0x84, 0x1e, 0x24, 0x0c, 0x84, 0x80, 0xd0,
	0x49, 0xfd, 0x04, 0x7a, 0x2f, 0x68, 0xbc, 0x79, 0xcb, 0x96, 0xc7, 0x2c, 0x74, 0x4b, 0xc8, 0x4b,
	0xd0, 0xfa, 0xe9, 0x85, 0xb4, 0x28, 0xee, 0xf4, 0xde, 0xef, 0x1f, 0xbf, 0xc7, 0x03, 0x70, 0x47,
	0xed, 0x1a, 0xe7, 0x6d, 0xb0, 0xac, 0x08, 0xa3, 0x43, 0xaa, 0xe1, 0x64, 0x0d, 0x4e, 0x54, 0xbd,
	0x76, 0xed, 0xac, 0x6e, 0x3e, 0xa1, 0xec, 0xb4, 0x7b, 0x8d, 0xe8, 0x47, 0x76, 0x07, 0x2b, 0x0a,
	0x3a, 0x44, 0xe2, 0x99, 0xc8, 0x64, 0xd5, 0x5e, 0x36, 0x29, 0xda, 0xbc, 0x25, 0x52, 0xcd, 0x22,
	0xbb, 0x81, 0x42, 0x1b, 0xd3, 0x21, 0x5f, 0x88, 0x4c, 0x96, 0x6a, 0x02, 0x8c, 0xc1, 0x92, 0xfa,
	0x6f, 0xe4, 0xb9, 0xc8, 0x64, 0xa1, 0xd2, 0xce, 0x6a, 0x28, 0xf1, 0x6b, 0x3f, 0x44, 0x83, 0xc4,
	0x97, 0x22, 0x97, 0x17, 0xea, 0x17, 0x6f, 0x06, 0xa8, 0x3a, 0xed, 0x14, 0x92, 0xb3, 0x27, 0x42,
	0x76, 0xff, 0xaf, 0xfb, 0xaa, 0xbd, 0x9e, 0xbb, 0x15, 0x52, 0x1c, 0xc2, 0xd9, 0x05, 0x0f, 0xb0,
	0xd6, 0xc6, 0x78, 0x24, 0x42, 0xe2, 0x0b, 0x91, 0xcb, 0xaa, 0x65, 0xb3, 0xff, 0x05, 0xd1, 0x3f,
	0x4d, 0x9a, 0xfa, 0x33, 0x3d, 0x8b, 0xf7, 0xdb, 0x43, 0x1f, 0x3e, 0xe2, 0xae, 0xd9, 0xdb, 0xe3,
	0x56, 0xa3, 0x3f, 0xd8, 0xde, 0x4e, 0x73, 0x9b, 0x82, 0xbb, 0x55, 0xfa, 0xc7, 0xe3, 0x4f, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x62, 0xd3, 0x65, 0x80, 0x3b, 0x01, 0x00, 0x00,
}
