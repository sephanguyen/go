// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: eureka/v2/book.proto

package epb

import (
	proto "github.com/golang/protobuf/proto"
	common "github.com/manabie-com/backend/pkg/manabuf/eureka/v2/common"
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

type UpsertBooksRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Books []*UpsertBooksRequest_Book `protobuf:"bytes,1,rep,name=books,proto3" json:"books,omitempty"`
}

func (x *UpsertBooksRequest) Reset() {
	*x = UpsertBooksRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertBooksRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertBooksRequest) ProtoMessage() {}

func (x *UpsertBooksRequest) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertBooksRequest.ProtoReflect.Descriptor instead.
func (*UpsertBooksRequest) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{0}
}

func (x *UpsertBooksRequest) GetBooks() []*UpsertBooksRequest_Book {
	if x != nil {
		return x.Books
	}
	return nil
}

type UpsertBooksResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BookIds []string `protobuf:"bytes,1,rep,name=book_ids,json=bookIds,proto3" json:"book_ids,omitempty"`
}

func (x *UpsertBooksResponse) Reset() {
	*x = UpsertBooksResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertBooksResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertBooksResponse) ProtoMessage() {}

func (x *UpsertBooksResponse) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertBooksResponse.ProtoReflect.Descriptor instead.
func (*UpsertBooksResponse) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{1}
}

func (x *UpsertBooksResponse) GetBookIds() []string {
	if x != nil {
		return x.BookIds
	}
	return nil
}

type GetBookContentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string                            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name     string                            `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Chapters []*GetBookContentResponse_Chapter `protobuf:"bytes,3,rep,name=chapters,proto3" json:"chapters,omitempty"`
}

func (x *GetBookContentResponse) Reset() {
	*x = GetBookContentResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBookContentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookContentResponse) ProtoMessage() {}

func (x *GetBookContentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookContentResponse.ProtoReflect.Descriptor instead.
func (*GetBookContentResponse) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{2}
}

func (x *GetBookContentResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *GetBookContentResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetBookContentResponse) GetChapters() []*GetBookContentResponse_Chapter {
	if x != nil {
		return x.Chapters
	}
	return nil
}

type GetBookContentRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BookId string `protobuf:"bytes,1,opt,name=book_id,json=bookId,proto3" json:"book_id,omitempty"`
}

func (x *GetBookContentRequest) Reset() {
	*x = GetBookContentRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBookContentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookContentRequest) ProtoMessage() {}

func (x *GetBookContentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookContentRequest.ProtoReflect.Descriptor instead.
func (*GetBookContentRequest) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{3}
}

func (x *GetBookContentRequest) GetBookId() string {
	if x != nil {
		return x.BookId
	}
	return ""
}

type GetBookHierarchyFlattenByLearningMaterialIDRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LearningMaterialId string `protobuf:"bytes,1,opt,name=learning_material_id,json=learningMaterialId,proto3" json:"learning_material_id,omitempty"`
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDRequest) Reset() {
	*x = GetBookHierarchyFlattenByLearningMaterialIDRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookHierarchyFlattenByLearningMaterialIDRequest) ProtoMessage() {}

func (x *GetBookHierarchyFlattenByLearningMaterialIDRequest) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookHierarchyFlattenByLearningMaterialIDRequest.ProtoReflect.Descriptor instead.
func (*GetBookHierarchyFlattenByLearningMaterialIDRequest) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{4}
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDRequest) GetLearningMaterialId() string {
	if x != nil {
		return x.LearningMaterialId
	}
	return ""
}

type GetBookHierarchyFlattenByLearningMaterialIDResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BookHierarchyFlatten *GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten `protobuf:"bytes,1,opt,name=book_hierarchy_flatten,json=bookHierarchyFlatten,proto3" json:"book_hierarchy_flatten,omitempty"`
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse) Reset() {
	*x = GetBookHierarchyFlattenByLearningMaterialIDResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookHierarchyFlattenByLearningMaterialIDResponse) ProtoMessage() {}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookHierarchyFlattenByLearningMaterialIDResponse.ProtoReflect.Descriptor instead.
func (*GetBookHierarchyFlattenByLearningMaterialIDResponse) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{5}
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse) GetBookHierarchyFlatten() *GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten {
	if x != nil {
		return x.BookHierarchyFlatten
	}
	return nil
}

type UpsertBooksRequest_Book struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BookId string `protobuf:"bytes,1,opt,name=book_id,json=bookId,proto3" json:"book_id,omitempty"`
	Name   string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *UpsertBooksRequest_Book) Reset() {
	*x = UpsertBooksRequest_Book{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertBooksRequest_Book) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertBooksRequest_Book) ProtoMessage() {}

func (x *UpsertBooksRequest_Book) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertBooksRequest_Book.ProtoReflect.Descriptor instead.
func (*UpsertBooksRequest_Book) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{0, 0}
}

func (x *UpsertBooksRequest_Book) GetBookId() string {
	if x != nil {
		return x.BookId
	}
	return ""
}

func (x *UpsertBooksRequest_Book) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type GetBookContentResponse_Chapter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string                          `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name         string                          `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	DisplayOrder int32                           `protobuf:"varint,3,opt,name=display_order,json=displayOrder,proto3" json:"display_order,omitempty"`
	Topics       []*GetBookContentResponse_Topic `protobuf:"bytes,4,rep,name=topics,proto3" json:"topics,omitempty"`
}

func (x *GetBookContentResponse_Chapter) Reset() {
	*x = GetBookContentResponse_Chapter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBookContentResponse_Chapter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookContentResponse_Chapter) ProtoMessage() {}

func (x *GetBookContentResponse_Chapter) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookContentResponse_Chapter.ProtoReflect.Descriptor instead.
func (*GetBookContentResponse_Chapter) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{2, 0}
}

func (x *GetBookContentResponse_Chapter) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *GetBookContentResponse_Chapter) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetBookContentResponse_Chapter) GetDisplayOrder() int32 {
	if x != nil {
		return x.DisplayOrder
	}
	return 0
}

func (x *GetBookContentResponse_Chapter) GetTopics() []*GetBookContentResponse_Topic {
	if x != nil {
		return x.Topics
	}
	return nil
}

type GetBookContentResponse_Topic struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id                string                                     `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name              string                                     `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	DisplayOrder      int32                                      `protobuf:"varint,3,opt,name=display_order,json=displayOrder,proto3" json:"display_order,omitempty"`
	IconUrl           string                                     `protobuf:"bytes,4,opt,name=icon_url,json=iconUrl,proto3" json:"icon_url,omitempty"`
	LearningMaterials []*GetBookContentResponse_LearningMaterial `protobuf:"bytes,5,rep,name=learning_materials,json=learningMaterials,proto3" json:"learning_materials,omitempty"`
}

func (x *GetBookContentResponse_Topic) Reset() {
	*x = GetBookContentResponse_Topic{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBookContentResponse_Topic) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookContentResponse_Topic) ProtoMessage() {}

func (x *GetBookContentResponse_Topic) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookContentResponse_Topic.ProtoReflect.Descriptor instead.
func (*GetBookContentResponse_Topic) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{2, 1}
}

func (x *GetBookContentResponse_Topic) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *GetBookContentResponse_Topic) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetBookContentResponse_Topic) GetDisplayOrder() int32 {
	if x != nil {
		return x.DisplayOrder
	}
	return 0
}

func (x *GetBookContentResponse_Topic) GetIconUrl() string {
	if x != nil {
		return x.IconUrl
	}
	return ""
}

func (x *GetBookContentResponse_Topic) GetLearningMaterials() []*GetBookContentResponse_LearningMaterial {
	if x != nil {
		return x.LearningMaterials
	}
	return nil
}

type GetBookContentResponse_LearningMaterial struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string                      `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	DisplayOrder int32                       `protobuf:"varint,2,opt,name=display_order,json=displayOrder,proto3" json:"display_order,omitempty"`
	Name         string                      `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Type         common.LearningMaterialType `protobuf:"varint,4,opt,name=type,proto3,enum=eureka.v2.common.LearningMaterialType" json:"type,omitempty"`
}

func (x *GetBookContentResponse_LearningMaterial) Reset() {
	*x = GetBookContentResponse_LearningMaterial{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBookContentResponse_LearningMaterial) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookContentResponse_LearningMaterial) ProtoMessage() {}

func (x *GetBookContentResponse_LearningMaterial) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookContentResponse_LearningMaterial.ProtoReflect.Descriptor instead.
func (*GetBookContentResponse_LearningMaterial) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{2, 2}
}

func (x *GetBookContentResponse_LearningMaterial) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *GetBookContentResponse_LearningMaterial) GetDisplayOrder() int32 {
	if x != nil {
		return x.DisplayOrder
	}
	return 0
}

func (x *GetBookContentResponse_LearningMaterial) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetBookContentResponse_LearningMaterial) GetType() common.LearningMaterialType {
	if x != nil {
		return x.Type
	}
	return common.LearningMaterialType_LEARNING_MATERIAL_LEARNING_OBJECTIVE
}

type GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BookId             string `protobuf:"bytes,1,opt,name=book_id,json=bookId,proto3" json:"book_id,omitempty"`
	ChapterId          string `protobuf:"bytes,2,opt,name=chapter_id,json=chapterId,proto3" json:"chapter_id,omitempty"`
	TopicId            string `protobuf:"bytes,3,opt,name=topic_id,json=topicId,proto3" json:"topic_id,omitempty"`
	LearningMaterialId string `protobuf:"bytes,4,opt,name=learning_material_id,json=learningMaterialId,proto3" json:"learning_material_id,omitempty"`
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten) Reset() {
	*x = GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v2_book_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten) ProtoMessage() {}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v2_book_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten.ProtoReflect.Descriptor instead.
func (*GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten) Descriptor() ([]byte, []int) {
	return file_eureka_v2_book_proto_rawDescGZIP(), []int{5, 0}
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten) GetBookId() string {
	if x != nil {
		return x.BookId
	}
	return ""
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten) GetChapterId() string {
	if x != nil {
		return x.ChapterId
	}
	return ""
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten) GetTopicId() string {
	if x != nil {
		return x.TopicId
	}
	return ""
}

func (x *GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten) GetLearningMaterialId() string {
	if x != nil {
		return x.LearningMaterialId
	}
	return ""
}

var File_eureka_v2_book_proto protoreflect.FileDescriptor

var file_eureka_v2_book_proto_rawDesc = []byte{
	0x0a, 0x14, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2f, 0x76, 0x32, 0x2f, 0x62, 0x6f, 0x6f, 0x6b,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76,
	0x32, 0x1a, 0x1c, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2f, 0x76, 0x32, 0x2f, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2f, 0x65, 0x6e, 0x75, 0x6d, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x83, 0x01, 0x0a, 0x12, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x38, 0x0a, 0x05, 0x62, 0x6f, 0x6f, 0x6b, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76,
	0x32, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x2e, 0x42, 0x6f, 0x6f, 0x6b, 0x52, 0x05, 0x62, 0x6f, 0x6f, 0x6b, 0x73,
	0x1a, 0x33, 0x0a, 0x04, 0x42, 0x6f, 0x6f, 0x6b, 0x12, 0x17, 0x0a, 0x07, 0x62, 0x6f, 0x6f, 0x6b,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x6f, 0x6f, 0x6b, 0x49,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x30, 0x0a, 0x13, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x42,
	0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x19, 0x0a, 0x08,
	0x62, 0x6f, 0x6f, 0x6b, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07,
	0x62, 0x6f, 0x6f, 0x6b, 0x49, 0x64, 0x73, 0x22, 0x84, 0x05, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x42,
	0x6f, 0x6f, 0x6b, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x45, 0x0a, 0x08, 0x63, 0x68, 0x61, 0x70, 0x74, 0x65,
	0x72, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b,
	0x61, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x43, 0x6f, 0x6e, 0x74,
	0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x43, 0x68, 0x61, 0x70,
	0x74, 0x65, 0x72, 0x52, 0x08, 0x63, 0x68, 0x61, 0x70, 0x74, 0x65, 0x72, 0x73, 0x1a, 0x93, 0x01,
	0x0a, 0x07, 0x43, 0x68, 0x61, 0x70, 0x74, 0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x23, 0x0a,
	0x0d, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4f, 0x72, 0x64,
	0x65, 0x72, 0x12, 0x3f, 0x0a, 0x06, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x73, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x27, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x32, 0x2e, 0x47,
	0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x52, 0x06, 0x74, 0x6f, 0x70,
	0x69, 0x63, 0x73, 0x1a, 0xce, 0x01, 0x0a, 0x05, 0x54, 0x6f, 0x70, 0x69, 0x63, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x23, 0x0a, 0x0d, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6f, 0x72, 0x64,
	0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61,
	0x79, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x19, 0x0a, 0x08, 0x69, 0x63, 0x6f, 0x6e, 0x5f, 0x75,
	0x72, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x69, 0x63, 0x6f, 0x6e, 0x55, 0x72,
	0x6c, 0x12, 0x61, 0x0a, 0x12, 0x6c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x61,
	0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32, 0x2e,
	0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f,
	0x6b, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x2e, 0x4c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61,
	0x6c, 0x52, 0x11, 0x6c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65, 0x72,
	0x69, 0x61, 0x6c, 0x73, 0x1a, 0x97, 0x01, 0x0a, 0x10, 0x4c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e,
	0x67, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x64, 0x69, 0x73,
	0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x3a, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x26, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x32, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x4c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65,
	0x72, 0x69, 0x61, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x30,
	0x0a, 0x15, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x62, 0x6f, 0x6f, 0x6b, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x6f, 0x6f, 0x6b, 0x49, 0x64,
	0x22, 0x66, 0x0a, 0x32, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x48, 0x69, 0x65, 0x72, 0x61,
	0x72, 0x63, 0x68, 0x79, 0x46, 0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x42, 0x79, 0x4c, 0x65, 0x61,
	0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x49, 0x44, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x14, 0x6c, 0x65, 0x61, 0x72, 0x6e, 0x69,
	0x6e, 0x67, 0x5f, 0x6d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x6c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61,
	0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x49, 0x64, 0x22, 0xdf, 0x02, 0x0a, 0x33, 0x47, 0x65, 0x74,
	0x42, 0x6f, 0x6f, 0x6b, 0x48, 0x69, 0x65, 0x72, 0x61, 0x72, 0x63, 0x68, 0x79, 0x46, 0x6c, 0x61,
	0x74, 0x74, 0x65, 0x6e, 0x42, 0x79, 0x4c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61,
	0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x89, 0x01, 0x0a, 0x16, 0x62, 0x6f, 0x6f, 0x6b, 0x5f, 0x68, 0x69, 0x65, 0x72, 0x61, 0x72,
	0x63, 0x68, 0x79, 0x5f, 0x66, 0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x53, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x65,
	0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x48, 0x69, 0x65, 0x72, 0x61, 0x72, 0x63, 0x68, 0x79, 0x46, 0x6c,
	0x61, 0x74, 0x74, 0x65, 0x6e, 0x42, 0x79, 0x4c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d,
	0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x2e, 0x42, 0x6f, 0x6f, 0x6b, 0x48, 0x69, 0x65, 0x72, 0x61, 0x72, 0x63, 0x68, 0x79, 0x46,
	0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x52, 0x14, 0x62, 0x6f, 0x6f, 0x6b, 0x48, 0x69, 0x65, 0x72,
	0x61, 0x72, 0x63, 0x68, 0x79, 0x46, 0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x1a, 0x9b, 0x01, 0x0a,
	0x14, 0x42, 0x6f, 0x6f, 0x6b, 0x48, 0x69, 0x65, 0x72, 0x61, 0x72, 0x63, 0x68, 0x79, 0x46, 0x6c,
	0x61, 0x74, 0x74, 0x65, 0x6e, 0x12, 0x17, 0x0a, 0x07, 0x62, 0x6f, 0x6f, 0x6b, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x6f, 0x6f, 0x6b, 0x49, 0x64, 0x12, 0x1d,
	0x0a, 0x0a, 0x63, 0x68, 0x61, 0x70, 0x74, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x63, 0x68, 0x61, 0x70, 0x74, 0x65, 0x72, 0x49, 0x64, 0x12, 0x19, 0x0a,
	0x08, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x49, 0x64, 0x12, 0x30, 0x0a, 0x14, 0x6c, 0x65, 0x61, 0x72,
	0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x5f, 0x69, 0x64,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x6c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67,
	0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x49, 0x64, 0x32, 0xe1, 0x02, 0x0a, 0x0b, 0x42,
	0x6f, 0x6f, 0x6b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4c, 0x0a, 0x0b, 0x55, 0x70,
	0x73, 0x65, 0x72, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73, 0x12, 0x1d, 0x2e, 0x65, 0x75, 0x72, 0x65,
	0x6b, 0x61, 0x2e, 0x76, 0x32, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x42, 0x6f, 0x6f, 0x6b,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b,
	0x61, 0x2e, 0x76, 0x32, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x55, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x42,
	0x6f, 0x6f, 0x6b, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x20, 0x2e, 0x65, 0x75, 0x72,
	0x65, 0x6b, 0x61, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x65,
	0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b,
	0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0xac, 0x01, 0x0a, 0x2b, 0x47, 0x65, 0x74, 0x42, 0x6f, 0x6f, 0x6b, 0x48, 0x69, 0x65, 0x72, 0x61,
	0x72, 0x63, 0x68, 0x79, 0x46, 0x6c, 0x61, 0x74, 0x74, 0x65, 0x6e, 0x42, 0x79, 0x4c, 0x65, 0x61,
	0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x49, 0x44, 0x12,
	0x3d, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x65, 0x74, 0x42,
	0x6f, 0x6f, 0x6b, 0x48, 0x69, 0x65, 0x72, 0x61, 0x72, 0x63, 0x68, 0x79, 0x46, 0x6c, 0x61, 0x74,
	0x74, 0x65, 0x6e, 0x42, 0x79, 0x4c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74,
	0x65, 0x72, 0x69, 0x61, 0x6c, 0x49, 0x44, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x3e,
	0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x6f,
	0x6f, 0x6b, 0x48, 0x69, 0x65, 0x72, 0x61, 0x72, 0x63, 0x68, 0x79, 0x46, 0x6c, 0x61, 0x74, 0x74,
	0x65, 0x6e, 0x42, 0x79, 0x4c, 0x65, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x65,
	0x72, 0x69, 0x61, 0x6c, 0x49, 0x44, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x3a,
	0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x6e,
	0x61, 0x62, 0x69, 0x65, 0x2d, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64,
	0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x75, 0x72,
	0x65, 0x6b, 0x61, 0x2f, 0x76, 0x32, 0x3b, 0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_eureka_v2_book_proto_rawDescOnce sync.Once
	file_eureka_v2_book_proto_rawDescData = file_eureka_v2_book_proto_rawDesc
)

func file_eureka_v2_book_proto_rawDescGZIP() []byte {
	file_eureka_v2_book_proto_rawDescOnce.Do(func() {
		file_eureka_v2_book_proto_rawDescData = protoimpl.X.CompressGZIP(file_eureka_v2_book_proto_rawDescData)
	})
	return file_eureka_v2_book_proto_rawDescData
}

var file_eureka_v2_book_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_eureka_v2_book_proto_goTypes = []interface{}{
	(*UpsertBooksRequest)(nil),                                                       // 0: eureka.v2.UpsertBooksRequest
	(*UpsertBooksResponse)(nil),                                                      // 1: eureka.v2.UpsertBooksResponse
	(*GetBookContentResponse)(nil),                                                   // 2: eureka.v2.GetBookContentResponse
	(*GetBookContentRequest)(nil),                                                    // 3: eureka.v2.GetBookContentRequest
	(*GetBookHierarchyFlattenByLearningMaterialIDRequest)(nil),                       // 4: eureka.v2.GetBookHierarchyFlattenByLearningMaterialIDRequest
	(*GetBookHierarchyFlattenByLearningMaterialIDResponse)(nil),                      // 5: eureka.v2.GetBookHierarchyFlattenByLearningMaterialIDResponse
	(*UpsertBooksRequest_Book)(nil),                                                  // 6: eureka.v2.UpsertBooksRequest.Book
	(*GetBookContentResponse_Chapter)(nil),                                           // 7: eureka.v2.GetBookContentResponse.Chapter
	(*GetBookContentResponse_Topic)(nil),                                             // 8: eureka.v2.GetBookContentResponse.Topic
	(*GetBookContentResponse_LearningMaterial)(nil),                                  // 9: eureka.v2.GetBookContentResponse.LearningMaterial
	(*GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten)(nil), // 10: eureka.v2.GetBookHierarchyFlattenByLearningMaterialIDResponse.BookHierarchyFlatten
	(common.LearningMaterialType)(0),                                                 // 11: eureka.v2.common.LearningMaterialType
}
var file_eureka_v2_book_proto_depIdxs = []int32{
	6,  // 0: eureka.v2.UpsertBooksRequest.books:type_name -> eureka.v2.UpsertBooksRequest.Book
	7,  // 1: eureka.v2.GetBookContentResponse.chapters:type_name -> eureka.v2.GetBookContentResponse.Chapter
	10, // 2: eureka.v2.GetBookHierarchyFlattenByLearningMaterialIDResponse.book_hierarchy_flatten:type_name -> eureka.v2.GetBookHierarchyFlattenByLearningMaterialIDResponse.BookHierarchyFlatten
	8,  // 3: eureka.v2.GetBookContentResponse.Chapter.topics:type_name -> eureka.v2.GetBookContentResponse.Topic
	9,  // 4: eureka.v2.GetBookContentResponse.Topic.learning_materials:type_name -> eureka.v2.GetBookContentResponse.LearningMaterial
	11, // 5: eureka.v2.GetBookContentResponse.LearningMaterial.type:type_name -> eureka.v2.common.LearningMaterialType
	0,  // 6: eureka.v2.BookService.UpsertBooks:input_type -> eureka.v2.UpsertBooksRequest
	3,  // 7: eureka.v2.BookService.GetBookContent:input_type -> eureka.v2.GetBookContentRequest
	4,  // 8: eureka.v2.BookService.GetBookHierarchyFlattenByLearningMaterialID:input_type -> eureka.v2.GetBookHierarchyFlattenByLearningMaterialIDRequest
	1,  // 9: eureka.v2.BookService.UpsertBooks:output_type -> eureka.v2.UpsertBooksResponse
	2,  // 10: eureka.v2.BookService.GetBookContent:output_type -> eureka.v2.GetBookContentResponse
	5,  // 11: eureka.v2.BookService.GetBookHierarchyFlattenByLearningMaterialID:output_type -> eureka.v2.GetBookHierarchyFlattenByLearningMaterialIDResponse
	9,  // [9:12] is the sub-list for method output_type
	6,  // [6:9] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_eureka_v2_book_proto_init() }
func file_eureka_v2_book_proto_init() {
	if File_eureka_v2_book_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_eureka_v2_book_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertBooksRequest); i {
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
		file_eureka_v2_book_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertBooksResponse); i {
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
		file_eureka_v2_book_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBookContentResponse); i {
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
		file_eureka_v2_book_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBookContentRequest); i {
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
		file_eureka_v2_book_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBookHierarchyFlattenByLearningMaterialIDRequest); i {
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
		file_eureka_v2_book_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBookHierarchyFlattenByLearningMaterialIDResponse); i {
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
		file_eureka_v2_book_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertBooksRequest_Book); i {
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
		file_eureka_v2_book_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBookContentResponse_Chapter); i {
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
		file_eureka_v2_book_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBookContentResponse_Topic); i {
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
		file_eureka_v2_book_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBookContentResponse_LearningMaterial); i {
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
		file_eureka_v2_book_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetBookHierarchyFlattenByLearningMaterialIDResponse_BookHierarchyFlatten); i {
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
			RawDescriptor: file_eureka_v2_book_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_eureka_v2_book_proto_goTypes,
		DependencyIndexes: file_eureka_v2_book_proto_depIdxs,
		MessageInfos:      file_eureka_v2_book_proto_msgTypes,
	}.Build()
	File_eureka_v2_book_proto = out.File
	file_eureka_v2_book_proto_rawDesc = nil
	file_eureka_v2_book_proto_goTypes = nil
	file_eureka_v2_book_proto_depIdxs = nil
}
