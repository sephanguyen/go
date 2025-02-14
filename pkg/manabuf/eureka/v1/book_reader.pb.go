// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: eureka/v1/book_reader.proto

package epb

import (
	proto "github.com/golang/protobuf/proto"
	v1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
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

type ListBooksRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Paging *v1.Paging       `protobuf:"bytes,1,opt,name=paging,proto3" json:"paging,omitempty"`
	Filter *v1.CommonFilter `protobuf:"bytes,2,opt,name=filter,proto3" json:"filter,omitempty"`
}

func (x *ListBooksRequest) Reset() {
	*x = ListBooksRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v1_book_reader_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListBooksRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListBooksRequest) ProtoMessage() {}

func (x *ListBooksRequest) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v1_book_reader_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListBooksRequest.ProtoReflect.Descriptor instead.
func (*ListBooksRequest) Descriptor() ([]byte, []int) {
	return file_eureka_v1_book_reader_proto_rawDescGZIP(), []int{0}
}

func (x *ListBooksRequest) GetPaging() *v1.Paging {
	if x != nil {
		return x.Paging
	}
	return nil
}

func (x *ListBooksRequest) GetFilter() *v1.CommonFilter {
	if x != nil {
		return x.Filter
	}
	return nil
}

type ListBooksResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NextPage *v1.Paging `protobuf:"bytes,1,opt,name=next_page,json=nextPage,proto3" json:"next_page,omitempty"`
	Items    []*v1.Book `protobuf:"bytes,2,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *ListBooksResponse) Reset() {
	*x = ListBooksResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v1_book_reader_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListBooksResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListBooksResponse) ProtoMessage() {}

func (x *ListBooksResponse) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v1_book_reader_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListBooksResponse.ProtoReflect.Descriptor instead.
func (*ListBooksResponse) Descriptor() ([]byte, []int) {
	return file_eureka_v1_book_reader_proto_rawDescGZIP(), []int{1}
}

func (x *ListBooksResponse) GetNextPage() *v1.Paging {
	if x != nil {
		return x.NextPage
	}
	return nil
}

func (x *ListBooksResponse) GetItems() []*v1.Book {
	if x != nil {
		return x.Items
	}
	return nil
}

var File_eureka_v1_book_reader_proto protoreflect.FileDescriptor

var file_eureka_v1_book_reader_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x62, 0x6f, 0x6f, 0x6b,
	0x5f, 0x72, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x65,
	0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x31, 0x1a, 0x18, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x18, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x6e, 0x0a, 0x10,
	0x4c, 0x69, 0x73, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x29, 0x0a, 0x06, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x11, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x67,
	0x69, 0x6e, 0x67, 0x52, 0x06, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x12, 0x2f, 0x0a, 0x06, 0x66,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x46, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x52, 0x06, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x22, 0x6a, 0x0a, 0x11,
	0x4c, 0x69, 0x73, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x2e, 0x0a, 0x09, 0x6e, 0x65, 0x78, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31,
	0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x52, 0x08, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67,
	0x65, 0x12, 0x25, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x0f, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x6f, 0x6f,
	0x6b, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x32, 0x5b, 0x0a, 0x11, 0x42, 0x6f, 0x6f, 0x6b,
	0x52, 0x65, 0x61, 0x64, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x46, 0x0a,
	0x09, 0x4c, 0x69, 0x73, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73, 0x12, 0x1b, 0x2e, 0x65, 0x75, 0x72,
	0x65, 0x6b, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61,
	0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x3a, 0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x69, 0x65, 0x2d, 0x63, 0x6f, 0x6d, 0x2f,
	0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x61, 0x6e, 0x61,
	0x62, 0x75, 0x66, 0x2f, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2f, 0x76, 0x31, 0x3b, 0x65, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_eureka_v1_book_reader_proto_rawDescOnce sync.Once
	file_eureka_v1_book_reader_proto_rawDescData = file_eureka_v1_book_reader_proto_rawDesc
)

func file_eureka_v1_book_reader_proto_rawDescGZIP() []byte {
	file_eureka_v1_book_reader_proto_rawDescOnce.Do(func() {
		file_eureka_v1_book_reader_proto_rawDescData = protoimpl.X.CompressGZIP(file_eureka_v1_book_reader_proto_rawDescData)
	})
	return file_eureka_v1_book_reader_proto_rawDescData
}

var file_eureka_v1_book_reader_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_eureka_v1_book_reader_proto_goTypes = []interface{}{
	(*ListBooksRequest)(nil),  // 0: eureka.v1.ListBooksRequest
	(*ListBooksResponse)(nil), // 1: eureka.v1.ListBooksResponse
	(*v1.Paging)(nil),         // 2: common.v1.Paging
	(*v1.CommonFilter)(nil),   // 3: common.v1.CommonFilter
	(*v1.Book)(nil),           // 4: common.v1.Book
}
var file_eureka_v1_book_reader_proto_depIdxs = []int32{
	2, // 0: eureka.v1.ListBooksRequest.paging:type_name -> common.v1.Paging
	3, // 1: eureka.v1.ListBooksRequest.filter:type_name -> common.v1.CommonFilter
	2, // 2: eureka.v1.ListBooksResponse.next_page:type_name -> common.v1.Paging
	4, // 3: eureka.v1.ListBooksResponse.items:type_name -> common.v1.Book
	0, // 4: eureka.v1.BookReaderService.ListBooks:input_type -> eureka.v1.ListBooksRequest
	1, // 5: eureka.v1.BookReaderService.ListBooks:output_type -> eureka.v1.ListBooksResponse
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_eureka_v1_book_reader_proto_init() }
func file_eureka_v1_book_reader_proto_init() {
	if File_eureka_v1_book_reader_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_eureka_v1_book_reader_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListBooksRequest); i {
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
		file_eureka_v1_book_reader_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListBooksResponse); i {
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
			RawDescriptor: file_eureka_v1_book_reader_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_eureka_v1_book_reader_proto_goTypes,
		DependencyIndexes: file_eureka_v1_book_reader_proto_depIdxs,
		MessageInfos:      file_eureka_v1_book_reader_proto_msgTypes,
	}.Build()
	File_eureka_v1_book_reader_proto = out.File
	file_eureka_v1_book_reader_proto_rawDesc = nil
	file_eureka_v1_book_reader_proto_goTypes = nil
	file_eureka_v1_book_reader_proto_depIdxs = nil
}
