// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: yasuo/v1/upload_reader.proto

package ypb

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
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

type RetrieveUploadInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Endpoint string `protobuf:"bytes,1,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	Bucket   string `protobuf:"bytes,2,opt,name=bucket,proto3" json:"bucket,omitempty"`
}

func (x *RetrieveUploadInfoResponse) Reset() {
	*x = RetrieveUploadInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_yasuo_v1_upload_reader_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveUploadInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveUploadInfoResponse) ProtoMessage() {}

func (x *RetrieveUploadInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_yasuo_v1_upload_reader_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveUploadInfoResponse.ProtoReflect.Descriptor instead.
func (*RetrieveUploadInfoResponse) Descriptor() ([]byte, []int) {
	return file_yasuo_v1_upload_reader_proto_rawDescGZIP(), []int{0}
}

func (x *RetrieveUploadInfoResponse) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *RetrieveUploadInfoResponse) GetBucket() string {
	if x != nil {
		return x.Bucket
	}
	return ""
}

var File_yasuo_v1_upload_reader_proto protoreflect.FileDescriptor

var file_yasuo_v1_upload_reader_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x79, 0x61, 0x73, 0x75, 0x6f, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x5f, 0x72, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08,
	0x79, 0x61, 0x73, 0x75, 0x6f, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x50, 0x0a, 0x1a, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76,
	0x65, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12,
	0x16, 0x0a, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x32, 0x69, 0x0a, 0x13, 0x55, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x52, 0x65, 0x61, 0x64, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x52,
	0x0a, 0x12, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x24, 0x2e, 0x79,
	0x61, 0x73, 0x75, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65,
	0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x69, 0x65, 0x2d, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x63,
	0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x75, 0x66,
	0x2f, 0x79, 0x61, 0x73, 0x75, 0x6f, 0x2f, 0x76, 0x31, 0x3b, 0x79, 0x70, 0x62, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_yasuo_v1_upload_reader_proto_rawDescOnce sync.Once
	file_yasuo_v1_upload_reader_proto_rawDescData = file_yasuo_v1_upload_reader_proto_rawDesc
)

func file_yasuo_v1_upload_reader_proto_rawDescGZIP() []byte {
	file_yasuo_v1_upload_reader_proto_rawDescOnce.Do(func() {
		file_yasuo_v1_upload_reader_proto_rawDescData = protoimpl.X.CompressGZIP(file_yasuo_v1_upload_reader_proto_rawDescData)
	})
	return file_yasuo_v1_upload_reader_proto_rawDescData
}

var file_yasuo_v1_upload_reader_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_yasuo_v1_upload_reader_proto_goTypes = []interface{}{
	(*RetrieveUploadInfoResponse)(nil), // 0: yasuo.v1.RetrieveUploadInfoResponse
	(*emptypb.Empty)(nil),              // 1: google.protobuf.Empty
}
var file_yasuo_v1_upload_reader_proto_depIdxs = []int32{
	1, // 0: yasuo.v1.UploadReaderService.RetrieveUploadInfo:input_type -> google.protobuf.Empty
	0, // 1: yasuo.v1.UploadReaderService.RetrieveUploadInfo:output_type -> yasuo.v1.RetrieveUploadInfoResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_yasuo_v1_upload_reader_proto_init() }
func file_yasuo_v1_upload_reader_proto_init() {
	if File_yasuo_v1_upload_reader_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_yasuo_v1_upload_reader_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetrieveUploadInfoResponse); i {
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
			RawDescriptor: file_yasuo_v1_upload_reader_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_yasuo_v1_upload_reader_proto_goTypes,
		DependencyIndexes: file_yasuo_v1_upload_reader_proto_depIdxs,
		MessageInfos:      file_yasuo_v1_upload_reader_proto_msgTypes,
	}.Build()
	File_yasuo_v1_upload_reader_proto = out.File
	file_yasuo_v1_upload_reader_proto_rawDesc = nil
	file_yasuo_v1_upload_reader_proto_goTypes = nil
	file_yasuo_v1_upload_reader_proto_depIdxs = nil
}
