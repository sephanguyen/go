// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: syllabus/v1/learning_history_data_sync.proto

package sspb

import (
	proto "github.com/golang/protobuf/proto"
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

type DownloadMappingFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DownloadMappingFileRequest) Reset() {
	*x = DownloadMappingFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_syllabus_v1_learning_history_data_sync_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadMappingFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadMappingFileRequest) ProtoMessage() {}

func (x *DownloadMappingFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_syllabus_v1_learning_history_data_sync_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadMappingFileRequest.ProtoReflect.Descriptor instead.
func (*DownloadMappingFileRequest) Descriptor() ([]byte, []int) {
	return file_syllabus_v1_learning_history_data_sync_proto_rawDescGZIP(), []int{0}
}

type DownloadMappingFileResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MappingCourseIdUrl           string `protobuf:"bytes,1,opt,name=mapping_course_id_url,json=mappingCourseIdUrl,proto3" json:"mapping_course_id_url,omitempty"`
	MappingExamLoIdUrl           string `protobuf:"bytes,2,opt,name=mapping_exam_lo_id_url,json=mappingExamLoIdUrl,proto3" json:"mapping_exam_lo_id_url,omitempty"`
	MappingQuestionTagUrl        string `protobuf:"bytes,3,opt,name=mapping_question_tag_url,json=mappingQuestionTagUrl,proto3" json:"mapping_question_tag_url,omitempty"`
	FailedSyncEmailRecipientsUrl string `protobuf:"bytes,4,opt,name=failed_sync_email_recipients_url,json=failedSyncEmailRecipientsUrl,proto3" json:"failed_sync_email_recipients_url,omitempty"`
}

func (x *DownloadMappingFileResponse) Reset() {
	*x = DownloadMappingFileResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_syllabus_v1_learning_history_data_sync_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadMappingFileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadMappingFileResponse) ProtoMessage() {}

func (x *DownloadMappingFileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_syllabus_v1_learning_history_data_sync_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadMappingFileResponse.ProtoReflect.Descriptor instead.
func (*DownloadMappingFileResponse) Descriptor() ([]byte, []int) {
	return file_syllabus_v1_learning_history_data_sync_proto_rawDescGZIP(), []int{1}
}

func (x *DownloadMappingFileResponse) GetMappingCourseIdUrl() string {
	if x != nil {
		return x.MappingCourseIdUrl
	}
	return ""
}

func (x *DownloadMappingFileResponse) GetMappingExamLoIdUrl() string {
	if x != nil {
		return x.MappingExamLoIdUrl
	}
	return ""
}

func (x *DownloadMappingFileResponse) GetMappingQuestionTagUrl() string {
	if x != nil {
		return x.MappingQuestionTagUrl
	}
	return ""
}

func (x *DownloadMappingFileResponse) GetFailedSyncEmailRecipientsUrl() string {
	if x != nil {
		return x.FailedSyncEmailRecipientsUrl
	}
	return ""
}

type UploadMappingFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MappingCourseId           []byte `protobuf:"bytes,1,opt,name=mapping_course_id,json=mappingCourseId,proto3" json:"mapping_course_id,omitempty"`
	MappingExamLoId           []byte `protobuf:"bytes,2,opt,name=mapping_exam_lo_id,json=mappingExamLoId,proto3" json:"mapping_exam_lo_id,omitempty"`
	MappingQuestionTag        []byte `protobuf:"bytes,3,opt,name=mapping_question_tag,json=mappingQuestionTag,proto3" json:"mapping_question_tag,omitempty"`
	FailedSyncEmailRecipients []byte `protobuf:"bytes,4,opt,name=failed_sync_email_recipients,json=failedSyncEmailRecipients,proto3" json:"failed_sync_email_recipients,omitempty"`
}

func (x *UploadMappingFileRequest) Reset() {
	*x = UploadMappingFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_syllabus_v1_learning_history_data_sync_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadMappingFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadMappingFileRequest) ProtoMessage() {}

func (x *UploadMappingFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_syllabus_v1_learning_history_data_sync_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadMappingFileRequest.ProtoReflect.Descriptor instead.
func (*UploadMappingFileRequest) Descriptor() ([]byte, []int) {
	return file_syllabus_v1_learning_history_data_sync_proto_rawDescGZIP(), []int{2}
}

func (x *UploadMappingFileRequest) GetMappingCourseId() []byte {
	if x != nil {
		return x.MappingCourseId
	}
	return nil
}

func (x *UploadMappingFileRequest) GetMappingExamLoId() []byte {
	if x != nil {
		return x.MappingExamLoId
	}
	return nil
}

func (x *UploadMappingFileRequest) GetMappingQuestionTag() []byte {
	if x != nil {
		return x.MappingQuestionTag
	}
	return nil
}

func (x *UploadMappingFileRequest) GetFailedSyncEmailRecipients() []byte {
	if x != nil {
		return x.FailedSyncEmailRecipients
	}
	return nil
}

type UploadMappingFileResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UploadMappingFileResponse) Reset() {
	*x = UploadMappingFileResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_syllabus_v1_learning_history_data_sync_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadMappingFileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadMappingFileResponse) ProtoMessage() {}

func (x *UploadMappingFileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_syllabus_v1_learning_history_data_sync_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadMappingFileResponse.ProtoReflect.Descriptor instead.
func (*UploadMappingFileResponse) Descriptor() ([]byte, []int) {
	return file_syllabus_v1_learning_history_data_sync_proto_rawDescGZIP(), []int{3}
}

var File_syllabus_v1_learning_history_data_sync_proto protoreflect.FileDescriptor

var file_syllabus_v1_learning_history_data_sync_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x73, 0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x6c, 0x65,
	0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x68, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x5f, 0x64,
	0x61, 0x74, 0x61, 0x5f, 0x73, 0x79, 0x6e, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b,
	0x73, 0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2e, 0x76, 0x31, 0x22, 0x1c, 0x0a, 0x1a, 0x44,
	0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x46, 0x69,
	0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x85, 0x02, 0x0a, 0x1b, 0x44, 0x6f,
	0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x6c,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x31, 0x0a, 0x15, 0x6d, 0x61, 0x70,
	0x70, 0x69, 0x6e, 0x67, 0x5f, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x5f, 0x69, 0x64, 0x5f, 0x75,
	0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e,
	0x67, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x49, 0x64, 0x55, 0x72, 0x6c, 0x12, 0x32, 0x0a, 0x16,
	0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x5f, 0x65, 0x78, 0x61, 0x6d, 0x5f, 0x6c, 0x6f, 0x5f,
	0x69, 0x64, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x6d, 0x61,
	0x70, 0x70, 0x69, 0x6e, 0x67, 0x45, 0x78, 0x61, 0x6d, 0x4c, 0x6f, 0x49, 0x64, 0x55, 0x72, 0x6c,
	0x12, 0x37, 0x0a, 0x18, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x5f, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x61, 0x67, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x15, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x51, 0x75, 0x65, 0x73, 0x74,
	0x69, 0x6f, 0x6e, 0x54, 0x61, 0x67, 0x55, 0x72, 0x6c, 0x12, 0x46, 0x0a, 0x20, 0x66, 0x61, 0x69,
	0x6c, 0x65, 0x64, 0x5f, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x5f, 0x72,
	0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x1c, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x53, 0x79, 0x6e, 0x63, 0x45,
	0x6d, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x55, 0x72,
	0x6c, 0x22, 0xe6, 0x01, 0x0a, 0x18, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61, 0x70, 0x70,
	0x69, 0x6e, 0x67, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2a,
	0x0a, 0x11, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x5f, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f, 0x6d, 0x61, 0x70, 0x70, 0x69,
	0x6e, 0x67, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x49, 0x64, 0x12, 0x2b, 0x0a, 0x12, 0x6d, 0x61,
	0x70, 0x70, 0x69, 0x6e, 0x67, 0x5f, 0x65, 0x78, 0x61, 0x6d, 0x5f, 0x6c, 0x6f, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x45,
	0x78, 0x61, 0x6d, 0x4c, 0x6f, 0x49, 0x64, 0x12, 0x30, 0x0a, 0x14, 0x6d, 0x61, 0x70, 0x70, 0x69,
	0x6e, 0x67, 0x5f, 0x71, 0x75, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x61, 0x67, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x12, 0x6d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x51, 0x75,
	0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x61, 0x67, 0x12, 0x3f, 0x0a, 0x1c, 0x66, 0x61, 0x69,
	0x6c, 0x65, 0x64, 0x5f, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x5f, 0x72,
	0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x19, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x53, 0x79, 0x6e, 0x63, 0x45, 0x6d, 0x61, 0x69, 0x6c,
	0x52, 0x65, 0x63, 0x69, 0x70, 0x69, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x1b, 0x0a, 0x19, 0x55, 0x70,
	0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x6c, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xee, 0x01, 0x0a, 0x1e, 0x4c, 0x65, 0x61, 0x72,
	0x6e, 0x69, 0x6e, 0x67, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x44, 0x61, 0x74, 0x61, 0x53,
	0x79, 0x6e, 0x63, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x68, 0x0a, 0x13, 0x44, 0x6f,
	0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x6c,
	0x65, 0x12, 0x27, 0x2e, 0x73, 0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2e, 0x76, 0x31, 0x2e,
	0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x46,
	0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x73, 0x79, 0x6c,
	0x6c, 0x61, 0x62, 0x75, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61,
	0x64, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x62, 0x0a, 0x11, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61,
	0x70, 0x70, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x25, 0x2e, 0x73, 0x79, 0x6c, 0x6c,
	0x61, 0x62, 0x75, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61,
	0x70, 0x70, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x26, 0x2e, 0x73, 0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x55,
	0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4d, 0x61, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x6c, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x3d, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x69, 0x65, 0x2d, 0x63,
	0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d,
	0x61, 0x6e, 0x61, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x79, 0x6c, 0x6c, 0x61, 0x62, 0x75, 0x73, 0x2f,
	0x76, 0x31, 0x3b, 0x73, 0x73, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_syllabus_v1_learning_history_data_sync_proto_rawDescOnce sync.Once
	file_syllabus_v1_learning_history_data_sync_proto_rawDescData = file_syllabus_v1_learning_history_data_sync_proto_rawDesc
)

func file_syllabus_v1_learning_history_data_sync_proto_rawDescGZIP() []byte {
	file_syllabus_v1_learning_history_data_sync_proto_rawDescOnce.Do(func() {
		file_syllabus_v1_learning_history_data_sync_proto_rawDescData = protoimpl.X.CompressGZIP(file_syllabus_v1_learning_history_data_sync_proto_rawDescData)
	})
	return file_syllabus_v1_learning_history_data_sync_proto_rawDescData
}

var file_syllabus_v1_learning_history_data_sync_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_syllabus_v1_learning_history_data_sync_proto_goTypes = []interface{}{
	(*DownloadMappingFileRequest)(nil),  // 0: syllabus.v1.DownloadMappingFileRequest
	(*DownloadMappingFileResponse)(nil), // 1: syllabus.v1.DownloadMappingFileResponse
	(*UploadMappingFileRequest)(nil),    // 2: syllabus.v1.UploadMappingFileRequest
	(*UploadMappingFileResponse)(nil),   // 3: syllabus.v1.UploadMappingFileResponse
}
var file_syllabus_v1_learning_history_data_sync_proto_depIdxs = []int32{
	0, // 0: syllabus.v1.LearningHistoryDataSyncService.DownloadMappingFile:input_type -> syllabus.v1.DownloadMappingFileRequest
	2, // 1: syllabus.v1.LearningHistoryDataSyncService.UploadMappingFile:input_type -> syllabus.v1.UploadMappingFileRequest
	1, // 2: syllabus.v1.LearningHistoryDataSyncService.DownloadMappingFile:output_type -> syllabus.v1.DownloadMappingFileResponse
	3, // 3: syllabus.v1.LearningHistoryDataSyncService.UploadMappingFile:output_type -> syllabus.v1.UploadMappingFileResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_syllabus_v1_learning_history_data_sync_proto_init() }
func file_syllabus_v1_learning_history_data_sync_proto_init() {
	if File_syllabus_v1_learning_history_data_sync_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_syllabus_v1_learning_history_data_sync_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadMappingFileRequest); i {
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
		file_syllabus_v1_learning_history_data_sync_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadMappingFileResponse); i {
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
		file_syllabus_v1_learning_history_data_sync_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadMappingFileRequest); i {
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
		file_syllabus_v1_learning_history_data_sync_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadMappingFileResponse); i {
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
			RawDescriptor: file_syllabus_v1_learning_history_data_sync_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_syllabus_v1_learning_history_data_sync_proto_goTypes,
		DependencyIndexes: file_syllabus_v1_learning_history_data_sync_proto_depIdxs,
		MessageInfos:      file_syllabus_v1_learning_history_data_sync_proto_msgTypes,
	}.Build()
	File_syllabus_v1_learning_history_data_sync_proto = out.File
	file_syllabus_v1_learning_history_data_sync_proto_rawDesc = nil
	file_syllabus_v1_learning_history_data_sync_proto_goTypes = nil
	file_syllabus_v1_learning_history_data_sync_proto_depIdxs = nil
}
