// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: api/v1/grpc/storymetadata.proto

package storymetadata_v1

import (
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

type GetStoryMetadataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NumberOfPages int32 `protobuf:"varint,1,opt,name=numberOfPages,proto3" json:"numberOfPages,omitempty"`
}

func (x *GetStoryMetadataRequest) Reset() {
	*x = GetStoryMetadataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_grpc_storymetadata_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStoryMetadataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStoryMetadataRequest) ProtoMessage() {}

func (x *GetStoryMetadataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_grpc_storymetadata_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStoryMetadataRequest.ProtoReflect.Descriptor instead.
func (*GetStoryMetadataRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_grpc_storymetadata_proto_rawDescGZIP(), []int{0}
}

func (x *GetStoryMetadataRequest) GetNumberOfPages() int32 {
	if x != nil {
		return x.NumberOfPages
	}
	return 0
}

type GetStoryMetadataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Stories []byte `protobuf:"bytes,1,opt,name=stories,proto3" json:"stories,omitempty"`
	Errs    []byte `protobuf:"bytes,2,opt,name=errs,proto3" json:"errs,omitempty"`
}

func (x *GetStoryMetadataResponse) Reset() {
	*x = GetStoryMetadataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_grpc_storymetadata_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStoryMetadataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStoryMetadataResponse) ProtoMessage() {}

func (x *GetStoryMetadataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_grpc_storymetadata_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStoryMetadataResponse.ProtoReflect.Descriptor instead.
func (*GetStoryMetadataResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_grpc_storymetadata_proto_rawDescGZIP(), []int{1}
}

func (x *GetStoryMetadataResponse) GetStories() []byte {
	if x != nil {
		return x.Stories
	}
	return nil
}

func (x *GetStoryMetadataResponse) GetErrs() []byte {
	if x != nil {
		return x.Errs
	}
	return nil
}

var File_api_v1_grpc_storymetadata_proto protoreflect.FileDescriptor

var file_api_v1_grpc_storymetadata_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x73, 0x74,
	0x6f, 0x72, 0x79, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x10, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x2e, 0x76, 0x31, 0x22, 0x3f, 0x0a, 0x17, 0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x79, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x24,
	0x0a, 0x0d, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4f, 0x66, 0x50, 0x61, 0x67, 0x65, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x4f, 0x66, 0x50,
	0x61, 0x67, 0x65, 0x73, 0x22, 0x48, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x79,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x65, 0x72,
	0x72, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x65, 0x72, 0x72, 0x73, 0x32, 0x77,
	0x0a, 0x0d, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x66, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x29,
	0x2e, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x73, 0x74, 0x6f, 0x72,
	0x79, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74,
	0x53, 0x74, 0x6f, 0x72, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x43, 0x5a, 0x41, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6b, 0x6f, 0x6e, 0x65, 0x61, 0x6c, 0x32, 0x30, 0x31, 0x33,
	0x2f, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x67, 0x65,
	0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x74, 0x6f, 0x72,
	0x79, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_v1_grpc_storymetadata_proto_rawDescOnce sync.Once
	file_api_v1_grpc_storymetadata_proto_rawDescData = file_api_v1_grpc_storymetadata_proto_rawDesc
)

func file_api_v1_grpc_storymetadata_proto_rawDescGZIP() []byte {
	file_api_v1_grpc_storymetadata_proto_rawDescOnce.Do(func() {
		file_api_v1_grpc_storymetadata_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_v1_grpc_storymetadata_proto_rawDescData)
	})
	return file_api_v1_grpc_storymetadata_proto_rawDescData
}

var file_api_v1_grpc_storymetadata_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_v1_grpc_storymetadata_proto_goTypes = []interface{}{
	(*GetStoryMetadataRequest)(nil),  // 0: storymetadata.v1.GetStoryMetadataRequest
	(*GetStoryMetadataResponse)(nil), // 1: storymetadata.v1.GetStoryMetadataResponse
}
var file_api_v1_grpc_storymetadata_proto_depIdxs = []int32{
	0, // 0: storymetadata.v1.storymetadata.GetMetadata:input_type -> storymetadata.v1.GetStoryMetadataRequest
	1, // 1: storymetadata.v1.storymetadata.GetMetadata:output_type -> storymetadata.v1.GetStoryMetadataResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_v1_grpc_storymetadata_proto_init() }
func file_api_v1_grpc_storymetadata_proto_init() {
	if File_api_v1_grpc_storymetadata_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_v1_grpc_storymetadata_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStoryMetadataRequest); i {
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
		file_api_v1_grpc_storymetadata_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStoryMetadataResponse); i {
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
			RawDescriptor: file_api_v1_grpc_storymetadata_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_v1_grpc_storymetadata_proto_goTypes,
		DependencyIndexes: file_api_v1_grpc_storymetadata_proto_depIdxs,
		MessageInfos:      file_api_v1_grpc_storymetadata_proto_msgTypes,
	}.Build()
	File_api_v1_grpc_storymetadata_proto = out.File
	file_api_v1_grpc_storymetadata_proto_rawDesc = nil
	file_api_v1_grpc_storymetadata_proto_goTypes = nil
	file_api_v1_grpc_storymetadata_proto_depIdxs = nil
}
