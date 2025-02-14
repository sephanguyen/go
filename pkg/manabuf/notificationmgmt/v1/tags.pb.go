// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: notificationmgmt/v1/tags.proto

package npb

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

type UpsertTagRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TagId string `protobuf:"bytes,1,opt,name=tag_id,json=tagId,proto3" json:"tag_id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *UpsertTagRequest) Reset() {
	*x = UpsertTagRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertTagRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertTagRequest) ProtoMessage() {}

func (x *UpsertTagRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertTagRequest.ProtoReflect.Descriptor instead.
func (*UpsertTagRequest) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{0}
}

func (x *UpsertTagRequest) GetTagId() string {
	if x != nil {
		return x.TagId
	}
	return ""
}

func (x *UpsertTagRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type UpsertTagResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TagId string `protobuf:"bytes,1,opt,name=tag_id,json=tagId,proto3" json:"tag_id,omitempty"`
}

func (x *UpsertTagResponse) Reset() {
	*x = UpsertTagResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertTagResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertTagResponse) ProtoMessage() {}

func (x *UpsertTagResponse) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertTagResponse.ProtoReflect.Descriptor instead.
func (*UpsertTagResponse) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{1}
}

func (x *UpsertTagResponse) GetTagId() string {
	if x != nil {
		return x.TagId
	}
	return ""
}

type DeleteTagRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TagId string `protobuf:"bytes,1,opt,name=tag_id,json=tagId,proto3" json:"tag_id,omitempty"`
}

func (x *DeleteTagRequest) Reset() {
	*x = DeleteTagRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteTagRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteTagRequest) ProtoMessage() {}

func (x *DeleteTagRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteTagRequest.ProtoReflect.Descriptor instead.
func (*DeleteTagRequest) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{2}
}

func (x *DeleteTagRequest) GetTagId() string {
	if x != nil {
		return x.TagId
	}
	return ""
}

type DeleteTagResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteTagResponse) Reset() {
	*x = DeleteTagResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteTagResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteTagResponse) ProtoMessage() {}

func (x *DeleteTagResponse) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteTagResponse.ProtoReflect.Descriptor instead.
func (*DeleteTagResponse) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{3}
}

type ImportTagsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Payload []byte `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *ImportTagsRequest) Reset() {
	*x = ImportTagsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImportTagsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImportTagsRequest) ProtoMessage() {}

func (x *ImportTagsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImportTagsRequest.ProtoReflect.Descriptor instead.
func (*ImportTagsRequest) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{4}
}

func (x *ImportTagsRequest) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

type ImportTagsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Errors []*ImportTagsResponse_ImportTagsError `protobuf:"bytes,1,rep,name=errors,proto3" json:"errors,omitempty"`
}

func (x *ImportTagsResponse) Reset() {
	*x = ImportTagsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImportTagsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImportTagsResponse) ProtoMessage() {}

func (x *ImportTagsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImportTagsResponse.ProtoReflect.Descriptor instead.
func (*ImportTagsResponse) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{5}
}

func (x *ImportTagsResponse) GetErrors() []*ImportTagsResponse_ImportTagsError {
	if x != nil {
		return x.Errors
	}
	return nil
}

type CheckExistTagNameRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TagName string `protobuf:"bytes,1,opt,name=tag_name,json=tagName,proto3" json:"tag_name,omitempty"`
}

func (x *CheckExistTagNameRequest) Reset() {
	*x = CheckExistTagNameRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckExistTagNameRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckExistTagNameRequest) ProtoMessage() {}

func (x *CheckExistTagNameRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckExistTagNameRequest.ProtoReflect.Descriptor instead.
func (*CheckExistTagNameRequest) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{6}
}

func (x *CheckExistTagNameRequest) GetTagName() string {
	if x != nil {
		return x.TagName
	}
	return ""
}

type CheckExistTagNameResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsExist bool `protobuf:"varint,1,opt,name=is_exist,json=isExist,proto3" json:"is_exist,omitempty"`
}

func (x *CheckExistTagNameResponse) Reset() {
	*x = CheckExistTagNameResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckExistTagNameResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckExistTagNameResponse) ProtoMessage() {}

func (x *CheckExistTagNameResponse) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckExistTagNameResponse.ProtoReflect.Descriptor instead.
func (*CheckExistTagNameResponse) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{7}
}

func (x *CheckExistTagNameResponse) GetIsExist() bool {
	if x != nil {
		return x.IsExist
	}
	return false
}

type GetTagsByFilterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Keyword string     `protobuf:"bytes,1,opt,name=keyword,proto3" json:"keyword,omitempty"`
	Paging  *v1.Paging `protobuf:"bytes,2,opt,name=paging,proto3" json:"paging,omitempty"`
}

func (x *GetTagsByFilterRequest) Reset() {
	*x = GetTagsByFilterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTagsByFilterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTagsByFilterRequest) ProtoMessage() {}

func (x *GetTagsByFilterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTagsByFilterRequest.ProtoReflect.Descriptor instead.
func (*GetTagsByFilterRequest) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{8}
}

func (x *GetTagsByFilterRequest) GetKeyword() string {
	if x != nil {
		return x.Keyword
	}
	return ""
}

func (x *GetTagsByFilterRequest) GetPaging() *v1.Paging {
	if x != nil {
		return x.Paging
	}
	return nil
}

type GetTagsByFilterResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tags         []*GetTagsByFilterResponse_Tag `protobuf:"bytes,1,rep,name=tags,proto3" json:"tags,omitempty"`
	TotalItems   uint32                         `protobuf:"varint,2,opt,name=total_items,json=totalItems,proto3" json:"total_items,omitempty"`
	NextPage     *v1.Paging                     `protobuf:"bytes,3,opt,name=next_page,json=nextPage,proto3" json:"next_page,omitempty"`
	PreviousPage *v1.Paging                     `protobuf:"bytes,4,opt,name=previous_page,json=previousPage,proto3" json:"previous_page,omitempty"`
}

func (x *GetTagsByFilterResponse) Reset() {
	*x = GetTagsByFilterResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTagsByFilterResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTagsByFilterResponse) ProtoMessage() {}

func (x *GetTagsByFilterResponse) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTagsByFilterResponse.ProtoReflect.Descriptor instead.
func (*GetTagsByFilterResponse) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{9}
}

func (x *GetTagsByFilterResponse) GetTags() []*GetTagsByFilterResponse_Tag {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *GetTagsByFilterResponse) GetTotalItems() uint32 {
	if x != nil {
		return x.TotalItems
	}
	return 0
}

func (x *GetTagsByFilterResponse) GetNextPage() *v1.Paging {
	if x != nil {
		return x.NextPage
	}
	return nil
}

func (x *GetTagsByFilterResponse) GetPreviousPage() *v1.Paging {
	if x != nil {
		return x.PreviousPage
	}
	return nil
}

type ExportTagsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ExportTagsRequest) Reset() {
	*x = ExportTagsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExportTagsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExportTagsRequest) ProtoMessage() {}

func (x *ExportTagsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExportTagsRequest.ProtoReflect.Descriptor instead.
func (*ExportTagsRequest) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{10}
}

type ExportTagsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ExportTagsResponse) Reset() {
	*x = ExportTagsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExportTagsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExportTagsResponse) ProtoMessage() {}

func (x *ExportTagsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExportTagsResponse.ProtoReflect.Descriptor instead.
func (*ExportTagsResponse) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{11}
}

func (x *ExportTagsResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type ImportTagsResponse_ImportTagsError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RowNumber int32  `protobuf:"varint,1,opt,name=row_number,json=rowNumber,proto3" json:"row_number,omitempty"`
	Error     string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *ImportTagsResponse_ImportTagsError) Reset() {
	*x = ImportTagsResponse_ImportTagsError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImportTagsResponse_ImportTagsError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImportTagsResponse_ImportTagsError) ProtoMessage() {}

func (x *ImportTagsResponse_ImportTagsError) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImportTagsResponse_ImportTagsError.ProtoReflect.Descriptor instead.
func (*ImportTagsResponse_ImportTagsError) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{5, 0}
}

func (x *ImportTagsResponse_ImportTagsError) GetRowNumber() int32 {
	if x != nil {
		return x.RowNumber
	}
	return 0
}

func (x *ImportTagsResponse_ImportTagsError) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type GetTagsByFilterResponse_Tag struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TagId string `protobuf:"bytes,1,opt,name=tag_id,json=tagId,proto3" json:"tag_id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *GetTagsByFilterResponse_Tag) Reset() {
	*x = GetTagsByFilterResponse_Tag{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v1_tags_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTagsByFilterResponse_Tag) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTagsByFilterResponse_Tag) ProtoMessage() {}

func (x *GetTagsByFilterResponse_Tag) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v1_tags_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTagsByFilterResponse_Tag.ProtoReflect.Descriptor instead.
func (*GetTagsByFilterResponse_Tag) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v1_tags_proto_rawDescGZIP(), []int{9, 0}
}

func (x *GetTagsByFilterResponse_Tag) GetTagId() string {
	if x != nil {
		return x.TagId
	}
	return ""
}

func (x *GetTagsByFilterResponse_Tag) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

var File_notificationmgmt_v1_tags_proto protoreflect.FileDescriptor

var file_notificationmgmt_v1_tags_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67,
	0x6d, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x61, 0x67, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x13, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67,
	0x6d, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x18, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31,
	0x2f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x3d, 0x0a, 0x10, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x54, 0x61, 0x67, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x74, 0x61, 0x67, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x61, 0x67, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x2a,
	0x0a, 0x11, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x54, 0x61, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x74, 0x61, 0x67, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x61, 0x67, 0x49, 0x64, 0x22, 0x29, 0x0a, 0x10, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x54, 0x61, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15,
	0x0a, 0x06, 0x74, 0x61, 0x67, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x61, 0x67, 0x49, 0x64, 0x22, 0x13, 0x0a, 0x11, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54,
	0x61, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2d, 0x0a, 0x11, 0x49, 0x6d,
	0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x67, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0xad, 0x01, 0x0a, 0x12, 0x49, 0x6d,
	0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x4f, 0x0a, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x37, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d,
	0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x67,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74,
	0x54, 0x61, 0x67, 0x73, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x06, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x73, 0x1a, 0x46, 0x0a, 0x0f, 0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x67, 0x73, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x6f, 0x77, 0x5f, 0x6e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x72, 0x6f, 0x77, 0x4e, 0x75, 0x6d,
	0x62, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x35, 0x0a, 0x18, 0x43, 0x68, 0x65,
	0x63, 0x6b, 0x45, 0x78, 0x69, 0x73, 0x74, 0x54, 0x61, 0x67, 0x4e, 0x61, 0x6d, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x74, 0x61, 0x67, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x74, 0x61, 0x67, 0x4e, 0x61, 0x6d, 0x65,
	0x22, 0x36, 0x0a, 0x19, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x45, 0x78, 0x69, 0x73, 0x74, 0x54, 0x61,
	0x67, 0x4e, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x19, 0x0a,
	0x08, 0x69, 0x73, 0x5f, 0x65, 0x78, 0x69, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x07, 0x69, 0x73, 0x45, 0x78, 0x69, 0x73, 0x74, 0x22, 0x5d, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x54,
	0x61, 0x67, 0x73, 0x42, 0x79, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x29, 0x0a, 0x06,
	0x70, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x52,
	0x06, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x22, 0x9a, 0x02, 0x0a, 0x17, 0x47, 0x65, 0x74, 0x54,
	0x61, 0x67, 0x73, 0x42, 0x79, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x44, 0x0a, 0x04, 0x74, 0x61, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x30, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x67, 0x73, 0x42,
	0x79, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e,
	0x54, 0x61, 0x67, 0x52, 0x04, 0x74, 0x61, 0x67, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x74, 0x6f, 0x74,
	0x61, 0x6c, 0x5f, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a,
	0x74, 0x6f, 0x74, 0x61, 0x6c, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x12, 0x2e, 0x0a, 0x09, 0x6e, 0x65,
	0x78, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x67,
	0x52, 0x08, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65, 0x12, 0x36, 0x0a, 0x0d, 0x70, 0x72,
	0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x11, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61,
	0x67, 0x69, 0x6e, 0x67, 0x52, 0x0c, 0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x50, 0x61,
	0x67, 0x65, 0x1a, 0x30, 0x0a, 0x03, 0x54, 0x61, 0x67, 0x12, 0x15, 0x0a, 0x06, 0x74, 0x61, 0x67,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x61, 0x67, 0x49, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x22, 0x13, 0x0a, 0x11, 0x45, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61,
	0x67, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x28, 0x0a, 0x12, 0x45, 0x78, 0x70,
	0x6f, 0x72, 0x74, 0x54, 0x61, 0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x32, 0xaf, 0x02, 0x0a, 0x16, 0x54, 0x61, 0x67, 0x4d, 0x67, 0x6d, 0x74, 0x4d,
	0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x5a,
	0x0a, 0x09, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x54, 0x61, 0x67, 0x12, 0x25, 0x2e, 0x6e, 0x6f,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76,
	0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x54, 0x61, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x26, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x54,
	0x61, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5a, 0x0a, 0x09, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x54, 0x61, 0x67, 0x12, 0x25, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x54, 0x61, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26,
	0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x61, 0x67, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5d, 0x0a, 0x0a, 0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74,
	0x54, 0x61, 0x67, 0x73, 0x12, 0x26, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6d, 0x70, 0x6f, 0x72,
	0x74, 0x54, 0x61, 0x67, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x6e,
	0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e,
	0x76, 0x31, 0x2e, 0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x67, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xd7, 0x02, 0x0a, 0x14, 0x54, 0x61, 0x67, 0x4d, 0x67, 0x6d,
	0x74, 0x52, 0x65, 0x61, 0x64, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x72,
	0x0a, 0x11, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x45, 0x78, 0x69, 0x73, 0x74, 0x54, 0x61, 0x67, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x2d, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x45,
	0x78, 0x69, 0x73, 0x74, 0x54, 0x61, 0x67, 0x4e, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x2e, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x45, 0x78,
	0x69, 0x73, 0x74, 0x54, 0x61, 0x67, 0x4e, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x6c, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x54, 0x61, 0x67, 0x73, 0x42, 0x79, 0x46,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x2b, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x54,
	0x61, 0x67, 0x73, 0x42, 0x79, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x67, 0x73,
	0x42, 0x79, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x5d, 0x0a, 0x0a, 0x45, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x67, 0x73, 0x12, 0x26,
	0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x67, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x70,
	0x6f, 0x72, 0x74, 0x54, 0x61, 0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42,
	0x44, 0x5a, 0x42, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61,
	0x6e, 0x61, 0x62, 0x69, 0x65, 0x2d, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e,
	0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x75, 0x66, 0x2f, 0x6e, 0x6f,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2f, 0x76,
	0x31, 0x3b, 0x6e, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_notificationmgmt_v1_tags_proto_rawDescOnce sync.Once
	file_notificationmgmt_v1_tags_proto_rawDescData = file_notificationmgmt_v1_tags_proto_rawDesc
)

func file_notificationmgmt_v1_tags_proto_rawDescGZIP() []byte {
	file_notificationmgmt_v1_tags_proto_rawDescOnce.Do(func() {
		file_notificationmgmt_v1_tags_proto_rawDescData = protoimpl.X.CompressGZIP(file_notificationmgmt_v1_tags_proto_rawDescData)
	})
	return file_notificationmgmt_v1_tags_proto_rawDescData
}

var file_notificationmgmt_v1_tags_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_notificationmgmt_v1_tags_proto_goTypes = []interface{}{
	(*UpsertTagRequest)(nil),                   // 0: notificationmgmt.v1.UpsertTagRequest
	(*UpsertTagResponse)(nil),                  // 1: notificationmgmt.v1.UpsertTagResponse
	(*DeleteTagRequest)(nil),                   // 2: notificationmgmt.v1.DeleteTagRequest
	(*DeleteTagResponse)(nil),                  // 3: notificationmgmt.v1.DeleteTagResponse
	(*ImportTagsRequest)(nil),                  // 4: notificationmgmt.v1.ImportTagsRequest
	(*ImportTagsResponse)(nil),                 // 5: notificationmgmt.v1.ImportTagsResponse
	(*CheckExistTagNameRequest)(nil),           // 6: notificationmgmt.v1.CheckExistTagNameRequest
	(*CheckExistTagNameResponse)(nil),          // 7: notificationmgmt.v1.CheckExistTagNameResponse
	(*GetTagsByFilterRequest)(nil),             // 8: notificationmgmt.v1.GetTagsByFilterRequest
	(*GetTagsByFilterResponse)(nil),            // 9: notificationmgmt.v1.GetTagsByFilterResponse
	(*ExportTagsRequest)(nil),                  // 10: notificationmgmt.v1.ExportTagsRequest
	(*ExportTagsResponse)(nil),                 // 11: notificationmgmt.v1.ExportTagsResponse
	(*ImportTagsResponse_ImportTagsError)(nil), // 12: notificationmgmt.v1.ImportTagsResponse.ImportTagsError
	(*GetTagsByFilterResponse_Tag)(nil),        // 13: notificationmgmt.v1.GetTagsByFilterResponse.Tag
	(*v1.Paging)(nil),                          // 14: common.v1.Paging
}
var file_notificationmgmt_v1_tags_proto_depIdxs = []int32{
	12, // 0: notificationmgmt.v1.ImportTagsResponse.errors:type_name -> notificationmgmt.v1.ImportTagsResponse.ImportTagsError
	14, // 1: notificationmgmt.v1.GetTagsByFilterRequest.paging:type_name -> common.v1.Paging
	13, // 2: notificationmgmt.v1.GetTagsByFilterResponse.tags:type_name -> notificationmgmt.v1.GetTagsByFilterResponse.Tag
	14, // 3: notificationmgmt.v1.GetTagsByFilterResponse.next_page:type_name -> common.v1.Paging
	14, // 4: notificationmgmt.v1.GetTagsByFilterResponse.previous_page:type_name -> common.v1.Paging
	0,  // 5: notificationmgmt.v1.TagMgmtModifierService.UpsertTag:input_type -> notificationmgmt.v1.UpsertTagRequest
	2,  // 6: notificationmgmt.v1.TagMgmtModifierService.DeleteTag:input_type -> notificationmgmt.v1.DeleteTagRequest
	4,  // 7: notificationmgmt.v1.TagMgmtModifierService.ImportTags:input_type -> notificationmgmt.v1.ImportTagsRequest
	6,  // 8: notificationmgmt.v1.TagMgmtReaderService.CheckExistTagName:input_type -> notificationmgmt.v1.CheckExistTagNameRequest
	8,  // 9: notificationmgmt.v1.TagMgmtReaderService.GetTagsByFilter:input_type -> notificationmgmt.v1.GetTagsByFilterRequest
	10, // 10: notificationmgmt.v1.TagMgmtReaderService.ExportTags:input_type -> notificationmgmt.v1.ExportTagsRequest
	1,  // 11: notificationmgmt.v1.TagMgmtModifierService.UpsertTag:output_type -> notificationmgmt.v1.UpsertTagResponse
	3,  // 12: notificationmgmt.v1.TagMgmtModifierService.DeleteTag:output_type -> notificationmgmt.v1.DeleteTagResponse
	5,  // 13: notificationmgmt.v1.TagMgmtModifierService.ImportTags:output_type -> notificationmgmt.v1.ImportTagsResponse
	7,  // 14: notificationmgmt.v1.TagMgmtReaderService.CheckExistTagName:output_type -> notificationmgmt.v1.CheckExistTagNameResponse
	9,  // 15: notificationmgmt.v1.TagMgmtReaderService.GetTagsByFilter:output_type -> notificationmgmt.v1.GetTagsByFilterResponse
	11, // 16: notificationmgmt.v1.TagMgmtReaderService.ExportTags:output_type -> notificationmgmt.v1.ExportTagsResponse
	11, // [11:17] is the sub-list for method output_type
	5,  // [5:11] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_notificationmgmt_v1_tags_proto_init() }
func file_notificationmgmt_v1_tags_proto_init() {
	if File_notificationmgmt_v1_tags_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_notificationmgmt_v1_tags_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertTagRequest); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertTagResponse); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteTagRequest); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteTagResponse); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImportTagsRequest); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImportTagsResponse); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckExistTagNameRequest); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckExistTagNameResponse); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTagsByFilterRequest); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTagsByFilterResponse); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExportTagsRequest); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExportTagsResponse); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImportTagsResponse_ImportTagsError); i {
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
		file_notificationmgmt_v1_tags_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTagsByFilterResponse_Tag); i {
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
			RawDescriptor: file_notificationmgmt_v1_tags_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_notificationmgmt_v1_tags_proto_goTypes,
		DependencyIndexes: file_notificationmgmt_v1_tags_proto_depIdxs,
		MessageInfos:      file_notificationmgmt_v1_tags_proto_msgTypes,
	}.Build()
	File_notificationmgmt_v1_tags_proto = out.File
	file_notificationmgmt_v1_tags_proto_rawDesc = nil
	file_notificationmgmt_v1_tags_proto_goTypes = nil
	file_notificationmgmt_v1_tags_proto_depIdxs = nil
}
