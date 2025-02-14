// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: calendar/v1/user.proto

package v1

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

type StaffInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Email string `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
}

func (x *StaffInfo) Reset() {
	*x = StaffInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_v1_user_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StaffInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StaffInfo) ProtoMessage() {}

func (x *StaffInfo) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_v1_user_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StaffInfo.ProtoReflect.Descriptor instead.
func (*StaffInfo) Descriptor() ([]byte, []int) {
	return file_calendar_v1_user_proto_rawDescGZIP(), []int{0}
}

func (x *StaffInfo) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *StaffInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *StaffInfo) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

type GetStaffsByLocationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LocationId string `protobuf:"bytes,1,opt,name=location_id,json=locationId,proto3" json:"location_id,omitempty"`
}

func (x *GetStaffsByLocationRequest) Reset() {
	*x = GetStaffsByLocationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_v1_user_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStaffsByLocationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStaffsByLocationRequest) ProtoMessage() {}

func (x *GetStaffsByLocationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_v1_user_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStaffsByLocationRequest.ProtoReflect.Descriptor instead.
func (*GetStaffsByLocationRequest) Descriptor() ([]byte, []int) {
	return file_calendar_v1_user_proto_rawDescGZIP(), []int{1}
}

func (x *GetStaffsByLocationRequest) GetLocationId() string {
	if x != nil {
		return x.LocationId
	}
	return ""
}

type GetStaffsByLocationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Staffs []*GetStaffsByLocationResponse_StaffInfo `protobuf:"bytes,1,rep,name=staffs,proto3" json:"staffs,omitempty"`
}

func (x *GetStaffsByLocationResponse) Reset() {
	*x = GetStaffsByLocationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_v1_user_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStaffsByLocationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStaffsByLocationResponse) ProtoMessage() {}

func (x *GetStaffsByLocationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_v1_user_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStaffsByLocationResponse.ProtoReflect.Descriptor instead.
func (*GetStaffsByLocationResponse) Descriptor() ([]byte, []int) {
	return file_calendar_v1_user_proto_rawDescGZIP(), []int{2}
}

func (x *GetStaffsByLocationResponse) GetStaffs() []*GetStaffsByLocationResponse_StaffInfo {
	if x != nil {
		return x.Staffs
	}
	return nil
}

type GetStaffsByLocationIDsAndNameOrEmailRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LocationIds        []string `protobuf:"bytes,1,rep,name=location_ids,json=locationIds,proto3" json:"location_ids,omitempty"`
	Keyword            string   `protobuf:"bytes,2,opt,name=keyword,proto3" json:"keyword,omitempty"`
	FilteredTeacherIds []string `protobuf:"bytes,3,rep,name=filtered_teacher_ids,json=filteredTeacherIds,proto3" json:"filtered_teacher_ids,omitempty"`
	Limit              uint32   `protobuf:"varint,4,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (x *GetStaffsByLocationIDsAndNameOrEmailRequest) Reset() {
	*x = GetStaffsByLocationIDsAndNameOrEmailRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_v1_user_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStaffsByLocationIDsAndNameOrEmailRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStaffsByLocationIDsAndNameOrEmailRequest) ProtoMessage() {}

func (x *GetStaffsByLocationIDsAndNameOrEmailRequest) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_v1_user_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStaffsByLocationIDsAndNameOrEmailRequest.ProtoReflect.Descriptor instead.
func (*GetStaffsByLocationIDsAndNameOrEmailRequest) Descriptor() ([]byte, []int) {
	return file_calendar_v1_user_proto_rawDescGZIP(), []int{3}
}

func (x *GetStaffsByLocationIDsAndNameOrEmailRequest) GetLocationIds() []string {
	if x != nil {
		return x.LocationIds
	}
	return nil
}

func (x *GetStaffsByLocationIDsAndNameOrEmailRequest) GetKeyword() string {
	if x != nil {
		return x.Keyword
	}
	return ""
}

func (x *GetStaffsByLocationIDsAndNameOrEmailRequest) GetFilteredTeacherIds() []string {
	if x != nil {
		return x.FilteredTeacherIds
	}
	return nil
}

func (x *GetStaffsByLocationIDsAndNameOrEmailRequest) GetLimit() uint32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

type GetStaffsByLocationIDsAndNameOrEmailResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Staffs []*StaffInfo `protobuf:"bytes,1,rep,name=staffs,proto3" json:"staffs,omitempty"`
}

func (x *GetStaffsByLocationIDsAndNameOrEmailResponse) Reset() {
	*x = GetStaffsByLocationIDsAndNameOrEmailResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_v1_user_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStaffsByLocationIDsAndNameOrEmailResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStaffsByLocationIDsAndNameOrEmailResponse) ProtoMessage() {}

func (x *GetStaffsByLocationIDsAndNameOrEmailResponse) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_v1_user_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStaffsByLocationIDsAndNameOrEmailResponse.ProtoReflect.Descriptor instead.
func (*GetStaffsByLocationIDsAndNameOrEmailResponse) Descriptor() ([]byte, []int) {
	return file_calendar_v1_user_proto_rawDescGZIP(), []int{4}
}

func (x *GetStaffsByLocationIDsAndNameOrEmailResponse) GetStaffs() []*StaffInfo {
	if x != nil {
		return x.Staffs
	}
	return nil
}

type GetStaffsByLocationResponse_StaffInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Email string `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
}

func (x *GetStaffsByLocationResponse_StaffInfo) Reset() {
	*x = GetStaffsByLocationResponse_StaffInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_calendar_v1_user_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStaffsByLocationResponse_StaffInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStaffsByLocationResponse_StaffInfo) ProtoMessage() {}

func (x *GetStaffsByLocationResponse_StaffInfo) ProtoReflect() protoreflect.Message {
	mi := &file_calendar_v1_user_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStaffsByLocationResponse_StaffInfo.ProtoReflect.Descriptor instead.
func (*GetStaffsByLocationResponse_StaffInfo) Descriptor() ([]byte, []int) {
	return file_calendar_v1_user_proto_rawDescGZIP(), []int{2, 0}
}

func (x *GetStaffsByLocationResponse_StaffInfo) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *GetStaffsByLocationResponse_StaffInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetStaffsByLocationResponse_StaffInfo) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

var File_calendar_v1_user_proto protoreflect.FileDescriptor

var file_calendar_v1_user_proto_rawDesc = []byte{
	0x0a, 0x16, 0x63, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x73,
	0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x63, 0x61, 0x6c, 0x65, 0x6e, 0x64,
	0x61, 0x72, 0x2e, 0x76, 0x31, 0x22, 0x45, 0x0a, 0x09, 0x53, 0x74, 0x61, 0x66, 0x66, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x22, 0x3d, 0x0a, 0x1a,
	0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x66, 0x66, 0x73, 0x42, 0x79, 0x4c, 0x6f, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x6c, 0x6f,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x22, 0xb0, 0x01, 0x0a, 0x1b,
	0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x66, 0x66, 0x73, 0x42, 0x79, 0x4c, 0x6f, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4a, 0x0a, 0x06, 0x73,
	0x74, 0x61, 0x66, 0x66, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x63, 0x61,
	0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61,
	0x66, 0x66, 0x73, 0x42, 0x79, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x66, 0x66, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x06, 0x73, 0x74, 0x61, 0x66, 0x66, 0x73, 0x1a, 0x45, 0x0a, 0x09, 0x53, 0x74, 0x61, 0x66, 0x66,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69,
	0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x22, 0xb2,
	0x01, 0x0a, 0x2b, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x66, 0x66, 0x73, 0x42, 0x79, 0x4c, 0x6f,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x73, 0x41, 0x6e, 0x64, 0x4e, 0x61, 0x6d, 0x65,
	0x4f, 0x72, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21,
	0x0a, 0x0c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64,
	0x73, 0x12, 0x18, 0x0a, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x30, 0x0a, 0x14, 0x66,
	0x69, 0x6c, 0x74, 0x65, 0x72, 0x65, 0x64, 0x5f, 0x74, 0x65, 0x61, 0x63, 0x68, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x12, 0x66, 0x69, 0x6c, 0x74, 0x65,
	0x72, 0x65, 0x64, 0x54, 0x65, 0x61, 0x63, 0x68, 0x65, 0x72, 0x49, 0x64, 0x73, 0x12, 0x14, 0x0a,
	0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x6c, 0x69,
	0x6d, 0x69, 0x74, 0x22, 0x5e, 0x0a, 0x2c, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x66, 0x66, 0x73,
	0x42, 0x79, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x73, 0x41, 0x6e, 0x64,
	0x4e, 0x61, 0x6d, 0x65, 0x4f, 0x72, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x66, 0x66, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x74, 0x61, 0x66, 0x66, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x06, 0x73, 0x74, 0x61,
	0x66, 0x66, 0x73, 0x32, 0x9b, 0x02, 0x0a, 0x11, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x9b, 0x01, 0x0a, 0x24, 0x47, 0x65,
	0x74, 0x53, 0x74, 0x61, 0x66, 0x66, 0x73, 0x42, 0x79, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x49, 0x44, 0x73, 0x41, 0x6e, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x4f, 0x72, 0x45, 0x6d, 0x61,
	0x69, 0x6c, 0x12, 0x38, 0x2e, 0x63, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x66, 0x66, 0x73, 0x42, 0x79, 0x4c, 0x6f, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x73, 0x41, 0x6e, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x4f, 0x72,
	0x45, 0x6d, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x39, 0x2e, 0x63,
	0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74,
	0x61, 0x66, 0x66, 0x73, 0x42, 0x79, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44,
	0x73, 0x41, 0x6e, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x4f, 0x72, 0x45, 0x6d, 0x61, 0x69, 0x6c, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x68, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x53, 0x74,
	0x61, 0x66, 0x66, 0x73, 0x42, 0x79, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x27,
	0x2e, 0x63, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74,
	0x53, 0x74, 0x61, 0x66, 0x66, 0x73, 0x42, 0x79, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x63, 0x61, 0x6c, 0x65, 0x6e, 0x64,
	0x61, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x61, 0x66, 0x66, 0x73, 0x42,
	0x79, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x38, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6d, 0x61, 0x6e, 0x61, 0x62, 0x69, 0x65, 0x2d, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x75, 0x66, 0x2f,
	0x63, 0x61, 0x6c, 0x65, 0x6e, 0x64, 0x61, 0x72, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_calendar_v1_user_proto_rawDescOnce sync.Once
	file_calendar_v1_user_proto_rawDescData = file_calendar_v1_user_proto_rawDesc
)

func file_calendar_v1_user_proto_rawDescGZIP() []byte {
	file_calendar_v1_user_proto_rawDescOnce.Do(func() {
		file_calendar_v1_user_proto_rawDescData = protoimpl.X.CompressGZIP(file_calendar_v1_user_proto_rawDescData)
	})
	return file_calendar_v1_user_proto_rawDescData
}

var file_calendar_v1_user_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_calendar_v1_user_proto_goTypes = []interface{}{
	(*StaffInfo)(nil),                                    // 0: calendar.v1.StaffInfo
	(*GetStaffsByLocationRequest)(nil),                   // 1: calendar.v1.GetStaffsByLocationRequest
	(*GetStaffsByLocationResponse)(nil),                  // 2: calendar.v1.GetStaffsByLocationResponse
	(*GetStaffsByLocationIDsAndNameOrEmailRequest)(nil),  // 3: calendar.v1.GetStaffsByLocationIDsAndNameOrEmailRequest
	(*GetStaffsByLocationIDsAndNameOrEmailResponse)(nil), // 4: calendar.v1.GetStaffsByLocationIDsAndNameOrEmailResponse
	(*GetStaffsByLocationResponse_StaffInfo)(nil),        // 5: calendar.v1.GetStaffsByLocationResponse.StaffInfo
}
var file_calendar_v1_user_proto_depIdxs = []int32{
	5, // 0: calendar.v1.GetStaffsByLocationResponse.staffs:type_name -> calendar.v1.GetStaffsByLocationResponse.StaffInfo
	0, // 1: calendar.v1.GetStaffsByLocationIDsAndNameOrEmailResponse.staffs:type_name -> calendar.v1.StaffInfo
	3, // 2: calendar.v1.UserReaderService.GetStaffsByLocationIDsAndNameOrEmail:input_type -> calendar.v1.GetStaffsByLocationIDsAndNameOrEmailRequest
	1, // 3: calendar.v1.UserReaderService.GetStaffsByLocation:input_type -> calendar.v1.GetStaffsByLocationRequest
	4, // 4: calendar.v1.UserReaderService.GetStaffsByLocationIDsAndNameOrEmail:output_type -> calendar.v1.GetStaffsByLocationIDsAndNameOrEmailResponse
	2, // 5: calendar.v1.UserReaderService.GetStaffsByLocation:output_type -> calendar.v1.GetStaffsByLocationResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_calendar_v1_user_proto_init() }
func file_calendar_v1_user_proto_init() {
	if File_calendar_v1_user_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_calendar_v1_user_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StaffInfo); i {
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
		file_calendar_v1_user_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStaffsByLocationRequest); i {
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
		file_calendar_v1_user_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStaffsByLocationResponse); i {
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
		file_calendar_v1_user_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStaffsByLocationIDsAndNameOrEmailRequest); i {
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
		file_calendar_v1_user_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStaffsByLocationIDsAndNameOrEmailResponse); i {
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
		file_calendar_v1_user_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStaffsByLocationResponse_StaffInfo); i {
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
			RawDescriptor: file_calendar_v1_user_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_calendar_v1_user_proto_goTypes,
		DependencyIndexes: file_calendar_v1_user_proto_depIdxs,
		MessageInfos:      file_calendar_v1_user_proto_msgTypes,
	}.Build()
	File_calendar_v1_user_proto = out.File
	file_calendar_v1_user_proto_rawDesc = nil
	file_calendar_v1_user_proto_goTypes = nil
	file_calendar_v1_user_proto_depIdxs = nil
}
