// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: syllabus/v1/learning_material.proto

package sspb

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
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

// lm_type: only ListTodo need - 2022/11/14
type BookTree struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BookId              string               `protobuf:"bytes,1,opt,name=book_id,json=bookId,proto3" json:"book_id,omitempty"`
	ChapterId           string               `protobuf:"bytes,2,opt,name=chapter_id,json=chapterId,proto3" json:"chapter_id,omitempty"`
	ChapterDisplayOrder int32                `protobuf:"varint,3,opt,name=chapter_display_order,json=chapterDisplayOrder,proto3" json:"chapter_display_order,omitempty"`
	TopicId             string               `protobuf:"bytes,4,opt,name=topic_id,json=topicId,proto3" json:"topic_id,omitempty"`
	TopicDisplayOrder   int32                `protobuf:"varint,5,opt,name=topic_display_order,json=topicDisplayOrder,proto3" json:"topic_display_order,omitempty"`
	LearningMaterialId  string               `protobuf:"bytes,6,opt,name=learning_material_id,json=learningMaterialId,proto3" json:"learning_material_id,omitempty"`
	LmDisplayOrder      int32                `protobuf:"varint,7,opt,name=lm_display_order,json=lmDisplayOrder,proto3" json:"lm_display_order,omitempty"`
	LmType              LearningMaterialType `protobuf:"varint,8,opt,name=lm_type,json=lmType,proto3,enum=syllabus.v1.LearningMaterialType" json:"lm_type,omitempty"`
}

func (x *BookTree) Reset() {
	*x = BookTree{}
	if protoimpl.UnsafeEnabled {
		mi := &file_syllabus_v1_learning_material_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BookTree) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BookTree) ProtoMessage() {}

func (x *BookTree) ProtoReflect() protoreflect.Message {
	mi := &file_syllabus_v1_learning_material_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BookTree.ProtoReflect.Descriptor instead.
func (*BookTree) Descriptor() ([]byte, []int) {
	return file_syllabus_v1_learning_material_proto_rawDescGZIP(), []int{0}
}

func (x *BookTree) GetBookId() string {
	if x != nil {
		return x.BookId
	}
	return ""
}

func (x *BookTree) GetChapterId() string {
	if x != nil {
		return x.ChapterId
	}
	return ""
}

func (x *BookTree) GetChapterDisplayOrder() int32 {
	if x != nil {
		return x.ChapterDisplayOrder
	}
	return 0
}

func (x *BookTree) GetTopicId() string {
	if x != nil {
		return x.TopicId
	}
	return ""
}

func (x *BookTree) GetTopicDisplayOrder() int32 {
	if x != nil {
		return x.TopicDisplayOrder
	}
	return 0
}

func (x *BookTree) GetLearningMaterialId() string {
	if x != nil {
		return x.LearningMaterialId
	}
	return ""
}

func (x *BookTree) GetLmDisplayOrder() int32 {
	if x != nil {
		return x.LmDisplayOrder
	}
	return 0
}

func (x *BookTree) GetLmType() LearningMaterialType {
	if x != nil {
		return x.LmType
	}
	return LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE
}

// LearningMaterialBase is a central or inheritance message to other learning material
// types can inherit.
type LearningMaterialBase struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// learning_material_id is generated learning material ULID
	LearningMaterialId string `protobuf:"bytes,1,opt,name=learning_material_id,json=learningMaterialId,proto3" json:"learning_material_id,omitempty"`
	// topic_id is 1-1 mapped topic ULID
	TopicId string `protobuf:"bytes,2,opt,name=topic_id,json=topicId,proto3" json:"topic_id,omitempty"`
	// name is learning material name
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// type is described at LearningMaterialType enum
	Type string `protobuf:"bytes,4,opt,name=type,proto3" json:"type,omitempty"`
	// display_order for LIST<...>
	DisplayOrder *wrapperspb.Int32Value     `protobuf:"bytes,5,opt,name=display_order,json=displayOrder,proto3" json:"display_order,omitempty"`
	VendorType   LearningMaterialVendorType `protobuf:"varint,6,opt,name=vendor_type,json=vendorType,proto3,enum=syllabus.v1.LearningMaterialVendorType" json:"vendor_type,omitempty"`
}

func (x *LearningMaterialBase) Reset() {
	*x = LearningMaterialBase{}
	if protoimpl.UnsafeEnabled {
		mi := &file_syllabus_v1_learning_material_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LearningMaterialBase) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LearningMaterialBase) ProtoMessage() {}

func (x *LearningMaterialBase) ProtoReflect() protoreflect.Message {
	mi := &file_syllabus_v1_learning_material_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LearningMaterialBase.ProtoReflect.Descriptor instead.
func (*LearningMaterialBase) Descriptor() ([]byte, []int) {
	return file_syllabus_v1_learning_material_proto_rawDescGZIP(), []int{1}
}

func (x *LearningMaterialBase) GetLearningMaterialId() string {
	if x != nil {
		return x.LearningMaterialId
	}
	return ""
}

func (x *LearningMaterialBase) GetTopicId() string {
	if x != nil {
		return x.TopicId
	}
	return ""
}

func (x *LearningMaterialBase) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *LearningMaterialBase) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *LearningMaterialBase) GetDisplayOrder() *wrapperspb.Int32Value {
	if x != nil {
		return x.DisplayOrder
	}
	return nil
}

func (x *LearningMaterialBase) GetVendorType() LearningMaterialVendorType {
	if x != nil {
		return x.VendorType
	}
	return LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE
}

var File_syllabus_v1_learning_material_proto protoreflect.FileDescriptor

var file_syllabus_v1_learning_material_proto_rawDesc = []byte{
	0x0a, 0x23, 0x73, 0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x6c, 0x65,
	0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x73, 0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2e,
	0x76, 0x31, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x17, 0x73, 0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2f, 0x76, 0x31, 0x2f,
	0x65, 0x6e, 0x75, 0x6d, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd9, 0x02, 0x0a, 0x08,
	0x42, 0x6f, 0x6f, 0x6b, 0x54, 0x72, 0x65, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x62, 0x6f, 0x6f, 0x6b,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x6f, 0x6f, 0x6b, 0x49,
	0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x68, 0x61, 0x70, 0x74, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x68, 0x61, 0x70, 0x74, 0x65, 0x72, 0x49, 0x64,
	0x12, 0x32, 0x0a, 0x15, 0x63, 0x68, 0x61, 0x70, 0x74, 0x65, 0x72, 0x5f, 0x64, 0x69, 0x73, 0x70,
	0x6c, 0x61, 0x79, 0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x13, 0x63, 0x68, 0x61, 0x70, 0x74, 0x65, 0x72, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4f,
	0x72, 0x64, 0x65, 0x72, 0x12, 0x19, 0x0a, 0x08, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x5f, 0x69, 0x64,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x49, 0x64, 0x12,
	0x2e, 0x0a, 0x13, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x5f, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79,
	0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x11, 0x74, 0x6f,
	0x70, 0x69, 0x63, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12,
	0x30, 0x0a, 0x14, 0x6c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x61, 0x74, 0x65,
	0x72, 0x69, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x6c,
	0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x49,
	0x64, 0x12, 0x28, 0x0a, 0x10, 0x6c, 0x6d, 0x5f, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f,
	0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x6c, 0x6d, 0x44,
	0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x3a, 0x0a, 0x07, 0x6c,
	0x6d, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x73,
	0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x65, 0x61, 0x72, 0x6e,
	0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x06, 0x6c, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x22, 0x97, 0x02, 0x0a, 0x14, 0x4c, 0x65, 0x61, 0x72,
	0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x42, 0x61, 0x73, 0x65,
	0x12, 0x30, 0x0a, 0x14, 0x6c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x61, 0x74,
	0x65, 0x72, 0x69, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12,
	0x6c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c,
	0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x49, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x40, 0x0a, 0x0d, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79,
	0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49,
	0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c,
	0x61, 0x79, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x48, 0x0a, 0x0b, 0x76, 0x65, 0x6e, 0x64, 0x6f,
	0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x27, 0x2e, 0x73,
	0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x65, 0x61, 0x72, 0x6e,
	0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x56, 0x65, 0x6e, 0x64, 0x6f,
	0x72, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x76, 0x65, 0x6e, 0x64, 0x6f, 0x72, 0x54, 0x79, 0x70,
	0x65, 0x42, 0x3d, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6d, 0x61, 0x6e, 0x61, 0x62, 0x69, 0x65, 0x2d, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x75, 0x66, 0x2f,
	0x73, 0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2f, 0x76, 0x31, 0x3b, 0x73, 0x73, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_syllabus_v1_learning_material_proto_rawDescOnce sync.Once
	file_syllabus_v1_learning_material_proto_rawDescData = file_syllabus_v1_learning_material_proto_rawDesc
)

func file_syllabus_v1_learning_material_proto_rawDescGZIP() []byte {
	file_syllabus_v1_learning_material_proto_rawDescOnce.Do(func() {
		file_syllabus_v1_learning_material_proto_rawDescData = protoimpl.X.CompressGZIP(file_syllabus_v1_learning_material_proto_rawDescData)
	})
	return file_syllabus_v1_learning_material_proto_rawDescData
}

var file_syllabus_v1_learning_material_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_syllabus_v1_learning_material_proto_goTypes = []interface{}{
	(*BookTree)(nil),                // 0: syllabus.v1.BookTree
	(*LearningMaterialBase)(nil),    // 1: syllabus.v1.LearningMaterialBase
	(LearningMaterialType)(0),       // 2: syllabus.v1.LearningMaterialType
	(*wrapperspb.Int32Value)(nil),   // 3: google.protobuf.Int32Value
	(LearningMaterialVendorType)(0), // 4: syllabus.v1.LearningMaterialVendorType
}
var file_syllabus_v1_learning_material_proto_depIdxs = []int32{
	2, // 0: syllabus.v1.BookTree.lm_type:type_name -> syllabus.v1.LearningMaterialType
	3, // 1: syllabus.v1.LearningMaterialBase.display_order:type_name -> google.protobuf.Int32Value
	4, // 2: syllabus.v1.LearningMaterialBase.vendor_type:type_name -> syllabus.v1.LearningMaterialVendorType
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_syllabus_v1_learning_material_proto_init() }
func file_syllabus_v1_learning_material_proto_init() {
	if File_syllabus_v1_learning_material_proto != nil {
		return
	}
	file_syllabus_v1_enums_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_syllabus_v1_learning_material_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BookTree); i {
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
		file_syllabus_v1_learning_material_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LearningMaterialBase); i {
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
			RawDescriptor: file_syllabus_v1_learning_material_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_syllabus_v1_learning_material_proto_goTypes,
		DependencyIndexes: file_syllabus_v1_learning_material_proto_depIdxs,
		MessageInfos:      file_syllabus_v1_learning_material_proto_msgTypes,
	}.Build()
	File_syllabus_v1_learning_material_proto = out.File
	file_syllabus_v1_learning_material_proto_rawDesc = nil
	file_syllabus_v1_learning_material_proto_goTypes = nil
	file_syllabus_v1_learning_material_proto_depIdxs = nil
}
