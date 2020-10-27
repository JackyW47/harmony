// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: downloader.proto

package downloader

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type DownloaderRequest_RequestType int32

const (
	DownloaderRequest_BLOCKHASH       DownloaderRequest_RequestType = 0
	DownloaderRequest_BLOCK           DownloaderRequest_RequestType = 1
	DownloaderRequest_NEWBLOCK        DownloaderRequest_RequestType = 2
	DownloaderRequest_BLOCKHEIGHT     DownloaderRequest_RequestType = 3
	DownloaderRequest_REGISTER        DownloaderRequest_RequestType = 4
	DownloaderRequest_REGISTERTIMEOUT DownloaderRequest_RequestType = 5
	DownloaderRequest_UNKNOWN         DownloaderRequest_RequestType = 6
	DownloaderRequest_BLOCKHEADER     DownloaderRequest_RequestType = 7
)

// Enum value maps for DownloaderRequest_RequestType.
var (
	DownloaderRequest_RequestType_name = map[int32]string{
		0: "BLOCKHASH",
		1: "BLOCK",
		2: "NEWBLOCK",
		3: "BLOCKHEIGHT",
		4: "REGISTER",
		5: "REGISTERTIMEOUT",
		6: "UNKNOWN",
		7: "BLOCKHEADER",
	}
	DownloaderRequest_RequestType_value = map[string]int32{
		"BLOCKHASH":       0,
		"BLOCK":           1,
		"NEWBLOCK":        2,
		"BLOCKHEIGHT":     3,
		"REGISTER":        4,
		"REGISTERTIMEOUT": 5,
		"UNKNOWN":         6,
		"BLOCKHEADER":     7,
	}
)

func (x DownloaderRequest_RequestType) Enum() *DownloaderRequest_RequestType {
	p := new(DownloaderRequest_RequestType)
	*p = x
	return p
}

func (x DownloaderRequest_RequestType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DownloaderRequest_RequestType) Descriptor() protoreflect.EnumDescriptor {
	return file_downloader_proto_enumTypes[0].Descriptor()
}

func (DownloaderRequest_RequestType) Type() protoreflect.EnumType {
	return &file_downloader_proto_enumTypes[0]
}

func (x DownloaderRequest_RequestType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DownloaderRequest_RequestType.Descriptor instead.
func (DownloaderRequest_RequestType) EnumDescriptor() ([]byte, []int) {
	return file_downloader_proto_rawDescGZIP(), []int{0, 0}
}

type DownloaderResponse_RegisterResponseType int32

const (
	DownloaderResponse_SUCCESS DownloaderResponse_RegisterResponseType = 0
	DownloaderResponse_FAIL    DownloaderResponse_RegisterResponseType = 1
	DownloaderResponse_INSYNC  DownloaderResponse_RegisterResponseType = 2 // node is now in sync, remove it from the broadcast list
)

// Enum value maps for DownloaderResponse_RegisterResponseType.
var (
	DownloaderResponse_RegisterResponseType_name = map[int32]string{
		0: "SUCCESS",
		1: "FAIL",
		2: "INSYNC",
	}
	DownloaderResponse_RegisterResponseType_value = map[string]int32{
		"SUCCESS": 0,
		"FAIL":    1,
		"INSYNC":  2,
	}
)

func (x DownloaderResponse_RegisterResponseType) Enum() *DownloaderResponse_RegisterResponseType {
	p := new(DownloaderResponse_RegisterResponseType)
	*p = x
	return p
}

func (x DownloaderResponse_RegisterResponseType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DownloaderResponse_RegisterResponseType) Descriptor() protoreflect.EnumDescriptor {
	return file_downloader_proto_enumTypes[1].Descriptor()
}

func (DownloaderResponse_RegisterResponseType) Type() protoreflect.EnumType {
	return &file_downloader_proto_enumTypes[1]
}

func (x DownloaderResponse_RegisterResponseType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DownloaderResponse_RegisterResponseType.Descriptor instead.
func (DownloaderResponse_RegisterResponseType) EnumDescriptor() ([]byte, []int) {
	return file_downloader_proto_rawDescGZIP(), []int{1, 0}
}

// DownloaderRequest is the generic download request.
type DownloaderRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Request type.
	Type DownloaderRequest_RequestType `protobuf:"varint,1,opt,name=type,proto3,enum=downloader.DownloaderRequest_RequestType" json:"type,omitempty"`
	// The hashes of the blocks we want to download.
	Hashes    [][]byte `protobuf:"bytes,2,rep,name=hashes,proto3" json:"hashes,omitempty"`
	PeerHash  []byte   `protobuf:"bytes,3,opt,name=peerHash,proto3" json:"peerHash,omitempty"`
	BlockHash []byte   `protobuf:"bytes,4,opt,name=blockHash,proto3" json:"blockHash,omitempty"`
	Ip        string   `protobuf:"bytes,5,opt,name=ip,proto3" json:"ip,omitempty"`
	Port      string   `protobuf:"bytes,6,opt,name=port,proto3" json:"port,omitempty"`
	Size      uint32   `protobuf:"varint,7,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *DownloaderRequest) Reset() {
	*x = DownloaderRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_downloader_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloaderRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloaderRequest) ProtoMessage() {}

func (x *DownloaderRequest) ProtoReflect() protoreflect.Message {
	mi := &file_downloader_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloaderRequest.ProtoReflect.Descriptor instead.
func (*DownloaderRequest) Descriptor() ([]byte, []int) {
	return file_downloader_proto_rawDescGZIP(), []int{0}
}

func (x *DownloaderRequest) GetType() DownloaderRequest_RequestType {
	if x != nil {
		return x.Type
	}
	return DownloaderRequest_BLOCKHASH
}

func (x *DownloaderRequest) GetHashes() [][]byte {
	if x != nil {
		return x.Hashes
	}
	return nil
}

func (x *DownloaderRequest) GetPeerHash() []byte {
	if x != nil {
		return x.PeerHash
	}
	return nil
}

func (x *DownloaderRequest) GetBlockHash() []byte {
	if x != nil {
		return x.BlockHash
	}
	return nil
}

func (x *DownloaderRequest) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *DownloaderRequest) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *DownloaderRequest) GetSize() uint32 {
	if x != nil {
		return x.Size
	}
	return 0
}

// DownloaderResponse is the generic response of DownloaderRequest.
type DownloaderResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// payload of Block.
	Payload [][]byte `protobuf:"bytes,1,rep,name=payload,proto3" json:"payload,omitempty"`
	// response of registration request
	Type        DownloaderResponse_RegisterResponseType `protobuf:"varint,2,opt,name=type,proto3,enum=downloader.DownloaderResponse_RegisterResponseType" json:"type,omitempty"`
	BlockHeight uint64                                  `protobuf:"varint,3,opt,name=blockHeight,proto3" json:"blockHeight,omitempty"`
}

func (x *DownloaderResponse) Reset() {
	*x = DownloaderResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_downloader_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloaderResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloaderResponse) ProtoMessage() {}

func (x *DownloaderResponse) ProtoReflect() protoreflect.Message {
	mi := &file_downloader_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloaderResponse.ProtoReflect.Descriptor instead.
func (*DownloaderResponse) Descriptor() ([]byte, []int) {
	return file_downloader_proto_rawDescGZIP(), []int{1}
}

func (x *DownloaderResponse) GetPayload() [][]byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *DownloaderResponse) GetType() DownloaderResponse_RegisterResponseType {
	if x != nil {
		return x.Type
	}
	return DownloaderResponse_SUCCESS
}

func (x *DownloaderResponse) GetBlockHeight() uint64 {
	if x != nil {
		return x.BlockHeight
	}
	return 0
}

var File_downloader_proto protoreflect.FileDescriptor

var file_downloader_proto_rawDesc = []byte{
	0x0a, 0x10, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0a, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x22, 0xe6,
	0x02, 0x0a, 0x11, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x3d, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x29, 0x2e, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x2e,
	0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x61, 0x73, 0x68, 0x65, 0x73, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0c, 0x52, 0x06, 0x68, 0x61, 0x73, 0x68, 0x65, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x70,
	0x65, 0x65, 0x72, 0x48, 0x61, 0x73, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x70,
	0x65, 0x65, 0x72, 0x48, 0x61, 0x73, 0x68, 0x12, 0x1c, 0x0a, 0x09, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x48, 0x61, 0x73, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x48, 0x61, 0x73, 0x68, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a,
	0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x87, 0x01,
	0x0a, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0d, 0x0a,
	0x09, 0x42, 0x4c, 0x4f, 0x43, 0x4b, 0x48, 0x41, 0x53, 0x48, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05,
	0x42, 0x4c, 0x4f, 0x43, 0x4b, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x4e, 0x45, 0x57, 0x42, 0x4c,
	0x4f, 0x43, 0x4b, 0x10, 0x02, 0x12, 0x0f, 0x0a, 0x0b, 0x42, 0x4c, 0x4f, 0x43, 0x4b, 0x48, 0x45,
	0x49, 0x47, 0x48, 0x54, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08, 0x52, 0x45, 0x47, 0x49, 0x53, 0x54,
	0x45, 0x52, 0x10, 0x04, 0x12, 0x13, 0x0a, 0x0f, 0x52, 0x45, 0x47, 0x49, 0x53, 0x54, 0x45, 0x52,
	0x54, 0x49, 0x4d, 0x45, 0x4f, 0x55, 0x54, 0x10, 0x05, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b,
	0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x06, 0x12, 0x0f, 0x0a, 0x0b, 0x42, 0x4c, 0x4f, 0x43, 0x4b, 0x48,
	0x45, 0x41, 0x44, 0x45, 0x52, 0x10, 0x07, 0x22, 0xd4, 0x01, 0x0a, 0x12, 0x44, 0x6f, 0x77, 0x6e,
	0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18,
	0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52,
	0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x47, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x33, 0x2e, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61,
	0x64, 0x65, 0x72, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x20, 0x0a, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69,
	0x67, 0x68, 0x74, 0x22, 0x39, 0x0a, 0x14, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x53,
	0x55, 0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x46, 0x41, 0x49, 0x4c,
	0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x49, 0x4e, 0x53, 0x59, 0x4e, 0x43, 0x10, 0x02, 0x32, 0x56,
	0x0a, 0x0a, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x12, 0x48, 0x0a, 0x05,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x1d, 0x2e, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64,
	0x65, 0x72, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65,
	0x72, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_downloader_proto_rawDescOnce sync.Once
	file_downloader_proto_rawDescData = file_downloader_proto_rawDesc
)

func file_downloader_proto_rawDescGZIP() []byte {
	file_downloader_proto_rawDescOnce.Do(func() {
		file_downloader_proto_rawDescData = protoimpl.X.CompressGZIP(file_downloader_proto_rawDescData)
	})
	return file_downloader_proto_rawDescData
}

var file_downloader_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_downloader_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_downloader_proto_goTypes = []interface{}{
	(DownloaderRequest_RequestType)(0),           // 0: downloader.DownloaderRequest.RequestType
	(DownloaderResponse_RegisterResponseType)(0), // 1: downloader.DownloaderResponse.RegisterResponseType
	(*DownloaderRequest)(nil),                    // 2: downloader.DownloaderRequest
	(*DownloaderResponse)(nil),                   // 3: downloader.DownloaderResponse
}
var file_downloader_proto_depIdxs = []int32{
	0, // 0: downloader.DownloaderRequest.type:type_name -> downloader.DownloaderRequest.RequestType
	1, // 1: downloader.DownloaderResponse.type:type_name -> downloader.DownloaderResponse.RegisterResponseType
	2, // 2: downloader.Downloader.Query:input_type -> downloader.DownloaderRequest
	3, // 3: downloader.Downloader.Query:output_type -> downloader.DownloaderResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_downloader_proto_init() }
func file_downloader_proto_init() {
	if File_downloader_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_downloader_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloaderRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_downloader_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloaderResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_downloader_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_downloader_proto_goTypes,
		DependencyIndexes: file_downloader_proto_depIdxs,
		EnumInfos:         file_downloader_proto_enumTypes,
		MessageInfos:      file_downloader_proto_msgTypes,
	}.Build()
	File_downloader_proto = out.File
	file_downloader_proto_rawDesc = nil
	file_downloader_proto_goTypes = nil
	file_downloader_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// DownloaderClient is the client API for Downloader service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DownloaderClient interface {
	Query(ctx context.Context, in *DownloaderRequest, opts ...grpc.CallOption) (*DownloaderResponse, error)
}

type downloaderClient struct {
	cc grpc.ClientConnInterface
}

func NewDownloaderClient(cc grpc.ClientConnInterface) DownloaderClient {
	return &downloaderClient{cc}
}

func (c *downloaderClient) Query(ctx context.Context, in *DownloaderRequest, opts ...grpc.CallOption) (*DownloaderResponse, error) {
	out := new(DownloaderResponse)
	err := c.cc.Invoke(ctx, "/downloader.Downloader/Query", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DownloaderServer is the server API for Downloader service.
type DownloaderServer interface {
	Query(context.Context, *DownloaderRequest) (*DownloaderResponse, error)
}

// UnimplementedDownloaderServer can be embedded to have forward compatible implementations.
type UnimplementedDownloaderServer struct {
}

func (*UnimplementedDownloaderServer) Query(context.Context, *DownloaderRequest) (*DownloaderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
}

func RegisterDownloaderServer(s *grpc.Server, srv DownloaderServer) {
	s.RegisterService(&_Downloader_serviceDesc, srv)
}

func _Downloader_Query_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloaderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DownloaderServer).Query(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/downloader.Downloader/Query",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DownloaderServer).Query(ctx, req.(*DownloaderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Downloader_serviceDesc = grpc.ServiceDesc{
	ServiceName: "downloader.Downloader",
	HandlerType: (*DownloaderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Query",
			Handler:    _Downloader_Query_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "downloader.proto",
}
