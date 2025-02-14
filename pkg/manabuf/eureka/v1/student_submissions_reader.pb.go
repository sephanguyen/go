// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: eureka/v1/student_submissions_reader.proto

package epb

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

type RetrieveStudentSubmissionHistoryByLoIDsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LoIds []string `protobuf:"bytes,1,rep,name=lo_ids,json=loIds,proto3" json:"lo_ids,omitempty"`
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsRequest) Reset() {
	*x = RetrieveStudentSubmissionHistoryByLoIDsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v1_student_submissions_reader_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveStudentSubmissionHistoryByLoIDsRequest) ProtoMessage() {}

func (x *RetrieveStudentSubmissionHistoryByLoIDsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v1_student_submissions_reader_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveStudentSubmissionHistoryByLoIDsRequest.ProtoReflect.Descriptor instead.
func (*RetrieveStudentSubmissionHistoryByLoIDsRequest) Descriptor() ([]byte, []int) {
	return file_eureka_v1_student_submissions_reader_proto_rawDescGZIP(), []int{0}
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsRequest) GetLoIds() []string {
	if x != nil {
		return x.LoIds
	}
	return nil
}

type RetrieveStudentSubmissionHistoryByLoIDsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Submissions []*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory `protobuf:"bytes,1,rep,name=submissions,proto3" json:"submissions,omitempty"`
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse) Reset() {
	*x = RetrieveStudentSubmissionHistoryByLoIDsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v1_student_submissions_reader_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveStudentSubmissionHistoryByLoIDsResponse) ProtoMessage() {}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v1_student_submissions_reader_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveStudentSubmissionHistoryByLoIDsResponse.ProtoReflect.Descriptor instead.
func (*RetrieveStudentSubmissionHistoryByLoIDsResponse) Descriptor() ([]byte, []int) {
	return file_eureka_v1_student_submissions_reader_proto_rawDescGZIP(), []int{1}
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse) GetSubmissions() []*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory {
	if x != nil {
		return x.Submissions
	}
	return nil
}

type RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LoId          string                                                                                `protobuf:"bytes,1,opt,name=lo_id,json=loId,proto3" json:"lo_id,omitempty"`
	Results       []*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult `protobuf:"bytes,2,rep,name=results,proto3" json:"results,omitempty"` // correspond with LearningObjective's ID
	TotalQuestion int32                                                                                 `protobuf:"varint,3,opt,name=total_question,json=totalQuestion,proto3" json:"total_question,omitempty"`
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory) Reset() {
	*x = RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v1_student_submissions_reader_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory) ProtoMessage() {}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v1_student_submissions_reader_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory.ProtoReflect.Descriptor instead.
func (*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory) Descriptor() ([]byte, []int) {
	return file_eureka_v1_student_submissions_reader_proto_rawDescGZIP(), []int{1, 0}
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory) GetLoId() string {
	if x != nil {
		return x.LoId
	}
	return ""
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory) GetResults() []*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult {
	if x != nil {
		return x.Results
	}
	return nil
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory) GetTotalQuestion() int32 {
	if x != nil {
		return x.TotalQuestion
	}
	return 0
}

type RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	QuestionId string `protobuf:"bytes,1,opt,name=question_id,json=questionId,proto3" json:"question_id,omitempty"`
	Correct    bool   `protobuf:"varint,2,opt,name=correct,proto3" json:"correct,omitempty"`
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult) Reset() {
	*x = RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_eureka_v1_student_submissions_reader_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult) ProtoMessage() {
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult) ProtoReflect() protoreflect.Message {
	mi := &file_eureka_v1_student_submissions_reader_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult.ProtoReflect.Descriptor instead.
func (*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult) Descriptor() ([]byte, []int) {
	return file_eureka_v1_student_submissions_reader_proto_rawDescGZIP(), []int{1, 0, 0}
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult) GetQuestionId() string {
	if x != nil {
		return x.QuestionId
	}
	return ""
}

func (x *RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult) GetCorrect() bool {
	if x != nil {
		return x.Correct
	}
	return false
}

var File_eureka_v1_student_submissions_reader_proto protoreflect.FileDescriptor

var file_eureka_v1_student_submissions_reader_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x75, 0x64,
	0x65, 0x6e, 0x74, 0x5f, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x5f,
	0x72, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x65, 0x75,
	0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x31, 0x22, 0x47, 0x0a, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69,
	0x65, 0x76, 0x65, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x79, 0x4c, 0x6f, 0x49,
	0x44, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x6c, 0x6f, 0x5f,
	0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x6f, 0x49, 0x64, 0x73,
	0x22, 0xbb, 0x03, 0x0a, 0x2f, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x53, 0x74, 0x75,
	0x64, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x69,
	0x73, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x79, 0x4c, 0x6f, 0x49, 0x44, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6e, 0x0a, 0x0b, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x4c, 0x2e, 0x65, 0x75, 0x72, 0x65,
	0x6b, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x53, 0x74,
	0x75, 0x64, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x48,
	0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x79, 0x4c, 0x6f, 0x49, 0x44, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x0b, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x97, 0x02, 0x0a, 0x11, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x13, 0x0a, 0x05, 0x6c, 0x6f,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6c, 0x6f, 0x49, 0x64, 0x12,
	0x77, 0x0a, 0x07, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x5d, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74,
	0x72, 0x69, 0x65, 0x76, 0x65, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x62, 0x6d,
	0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x79, 0x4c,
	0x6f, 0x49, 0x44, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x53, 0x75, 0x62,
	0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x53,
	0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52,
	0x07, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x73, 0x12, 0x25, 0x0a, 0x0e, 0x74, 0x6f, 0x74, 0x61,
	0x6c, 0x5f, 0x71, 0x75, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0d, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x51, 0x75, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x1a,
	0x4d, 0x0a, 0x10, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x71, 0x75, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x71, 0x75, 0x65, 0x73, 0x74, 0x69,
	0x6f, 0x6e, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x72, 0x72, 0x65, 0x63, 0x74, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x63, 0x6f, 0x72, 0x72, 0x65, 0x63, 0x74, 0x32, 0xc3,
	0x01, 0x0a, 0x1e, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x61, 0x64, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0xa0, 0x01, 0x0a, 0x27, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x53, 0x74,
	0x75, 0x64, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x48,
	0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x79, 0x4c, 0x6f, 0x49, 0x44, 0x73, 0x12, 0x39, 0x2e,
	0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65,
	0x76, 0x65, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x79, 0x4c, 0x6f, 0x49, 0x44,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x3a, 0x2e, 0x65, 0x75, 0x72, 0x65, 0x6b,
	0x61, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x53, 0x74, 0x75,
	0x64, 0x65, 0x6e, 0x74, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x69,
	0x73, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x79, 0x4c, 0x6f, 0x49, 0x44, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x42, 0x3a, 0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x69, 0x65, 0x2d, 0x63, 0x6f, 0x6d, 0x2f, 0x62,
	0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62,
	0x75, 0x66, 0x2f, 0x65, 0x75, 0x72, 0x65, 0x6b, 0x61, 0x2f, 0x76, 0x31, 0x3b, 0x65, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_eureka_v1_student_submissions_reader_proto_rawDescOnce sync.Once
	file_eureka_v1_student_submissions_reader_proto_rawDescData = file_eureka_v1_student_submissions_reader_proto_rawDesc
)

func file_eureka_v1_student_submissions_reader_proto_rawDescGZIP() []byte {
	file_eureka_v1_student_submissions_reader_proto_rawDescOnce.Do(func() {
		file_eureka_v1_student_submissions_reader_proto_rawDescData = protoimpl.X.CompressGZIP(file_eureka_v1_student_submissions_reader_proto_rawDescData)
	})
	return file_eureka_v1_student_submissions_reader_proto_rawDescData
}

var file_eureka_v1_student_submissions_reader_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_eureka_v1_student_submissions_reader_proto_goTypes = []interface{}{
	(*RetrieveStudentSubmissionHistoryByLoIDsRequest)(nil),                                     // 0: eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsRequest
	(*RetrieveStudentSubmissionHistoryByLoIDsResponse)(nil),                                    // 1: eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsResponse
	(*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory)(nil),                  // 2: eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsResponse.SubmissionHistory
	(*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult)(nil), // 3: eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsResponse.SubmissionHistory.SubmissionResult
}
var file_eureka_v1_student_submissions_reader_proto_depIdxs = []int32{
	2, // 0: eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsResponse.submissions:type_name -> eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsResponse.SubmissionHistory
	3, // 1: eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsResponse.SubmissionHistory.results:type_name -> eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsResponse.SubmissionHistory.SubmissionResult
	0, // 2: eureka.v1.StudentSubmissionReaderService.RetrieveStudentSubmissionHistoryByLoIDs:input_type -> eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsRequest
	1, // 3: eureka.v1.StudentSubmissionReaderService.RetrieveStudentSubmissionHistoryByLoIDs:output_type -> eureka.v1.RetrieveStudentSubmissionHistoryByLoIDsResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_eureka_v1_student_submissions_reader_proto_init() }
func file_eureka_v1_student_submissions_reader_proto_init() {
	if File_eureka_v1_student_submissions_reader_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_eureka_v1_student_submissions_reader_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetrieveStudentSubmissionHistoryByLoIDsRequest); i {
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
		file_eureka_v1_student_submissions_reader_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetrieveStudentSubmissionHistoryByLoIDsResponse); i {
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
		file_eureka_v1_student_submissions_reader_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory); i {
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
		file_eureka_v1_student_submissions_reader_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetrieveStudentSubmissionHistoryByLoIDsResponse_SubmissionHistory_SubmissionResult); i {
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
			RawDescriptor: file_eureka_v1_student_submissions_reader_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_eureka_v1_student_submissions_reader_proto_goTypes,
		DependencyIndexes: file_eureka_v1_student_submissions_reader_proto_depIdxs,
		MessageInfos:      file_eureka_v1_student_submissions_reader_proto_msgTypes,
	}.Build()
	File_eureka_v1_student_submissions_reader_proto = out.File
	file_eureka_v1_student_submissions_reader_proto_rawDesc = nil
	file_eureka_v1_student_submissions_reader_proto_goTypes = nil
	file_eureka_v1_student_submissions_reader_proto_depIdxs = nil
}
