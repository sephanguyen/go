// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: notificationmgmt/v2/notifications.proto

package npbv2

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

type RetrieveNotificationDetailRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserNotificationId string `protobuf:"bytes,1,opt,name=user_notification_id,json=userNotificationId,proto3" json:"user_notification_id,omitempty"`
}

func (x *RetrieveNotificationDetailRequest) Reset() {
	*x = RetrieveNotificationDetailRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v2_notifications_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveNotificationDetailRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveNotificationDetailRequest) ProtoMessage() {}

func (x *RetrieveNotificationDetailRequest) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v2_notifications_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveNotificationDetailRequest.ProtoReflect.Descriptor instead.
func (*RetrieveNotificationDetailRequest) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v2_notifications_proto_rawDescGZIP(), []int{0}
}

func (x *RetrieveNotificationDetailRequest) GetUserNotificationId() string {
	if x != nil {
		return x.UserNotificationId
	}
	return ""
}

type RetrieveNotificationDetailResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item              *v1.Notification      `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
	UserNotification  *v1.UserNotification  `protobuf:"bytes,2,opt,name=user_notification,json=userNotification,proto3" json:"user_notification,omitempty"`
	UserQuestionnaire *v1.UserQuestionnaire `protobuf:"bytes,3,opt,name=user_questionnaire,json=userQuestionnaire,proto3" json:"user_questionnaire,omitempty"`
}

func (x *RetrieveNotificationDetailResponse) Reset() {
	*x = RetrieveNotificationDetailResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_notificationmgmt_v2_notifications_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveNotificationDetailResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveNotificationDetailResponse) ProtoMessage() {}

func (x *RetrieveNotificationDetailResponse) ProtoReflect() protoreflect.Message {
	mi := &file_notificationmgmt_v2_notifications_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveNotificationDetailResponse.ProtoReflect.Descriptor instead.
func (*RetrieveNotificationDetailResponse) Descriptor() ([]byte, []int) {
	return file_notificationmgmt_v2_notifications_proto_rawDescGZIP(), []int{1}
}

func (x *RetrieveNotificationDetailResponse) GetItem() *v1.Notification {
	if x != nil {
		return x.Item
	}
	return nil
}

func (x *RetrieveNotificationDetailResponse) GetUserNotification() *v1.UserNotification {
	if x != nil {
		return x.UserNotification
	}
	return nil
}

func (x *RetrieveNotificationDetailResponse) GetUserQuestionnaire() *v1.UserQuestionnaire {
	if x != nil {
		return x.UserQuestionnaire
	}
	return nil
}

var File_notificationmgmt_v2_notifications_proto protoreflect.FileDescriptor

var file_notificationmgmt_v2_notifications_proto_rawDesc = []byte{
	0x0a, 0x27, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67,
	0x6d, 0x74, 0x2f, 0x76, 0x32, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x6e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x32, 0x1a, 0x1d,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x55, 0x0a,
	0x21, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x30, 0x0a, 0x14, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x6e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x12, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x49, 0x64, 0x22, 0xe8, 0x01, 0x0a, 0x22, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76,
	0x65, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2b, 0x0a, 0x04, 0x69,
	0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x12, 0x48, 0x0a, 0x11, 0x75, 0x73, 0x65, 0x72,
	0x5f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x55, 0x73, 0x65, 0x72, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x10, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x4b, 0x0a, 0x12, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x69, 0x6f, 0x6e, 0x6e, 0x61, 0x69, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x51,
	0x75, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x6e, 0x61, 0x69, 0x72, 0x65, 0x52, 0x11, 0x75, 0x73,
	0x65, 0x72, 0x51, 0x75, 0x65, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x6e, 0x61, 0x69, 0x72, 0x65, 0x32,
	0xab, 0x01, 0x0a, 0x19, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x61, 0x64, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x8d, 0x01,
	0x0a, 0x1a, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x12, 0x36, 0x2e, 0x6e,
	0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e,
	0x76, 0x32, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x37, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69,
	0x65, 0x76, 0x65, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44,
	0x65, 0x74, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x46, 0x5a,
	0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x6e, 0x61,
	0x62, 0x69, 0x65, 0x2d, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x75, 0x66, 0x2f, 0x6e, 0x6f, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x6d, 0x67, 0x6d, 0x74, 0x2f, 0x76, 0x32, 0x3b,
	0x6e, 0x70, 0x62, 0x76, 0x32, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_notificationmgmt_v2_notifications_proto_rawDescOnce sync.Once
	file_notificationmgmt_v2_notifications_proto_rawDescData = file_notificationmgmt_v2_notifications_proto_rawDesc
)

func file_notificationmgmt_v2_notifications_proto_rawDescGZIP() []byte {
	file_notificationmgmt_v2_notifications_proto_rawDescOnce.Do(func() {
		file_notificationmgmt_v2_notifications_proto_rawDescData = protoimpl.X.CompressGZIP(file_notificationmgmt_v2_notifications_proto_rawDescData)
	})
	return file_notificationmgmt_v2_notifications_proto_rawDescData
}

var file_notificationmgmt_v2_notifications_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_notificationmgmt_v2_notifications_proto_goTypes = []interface{}{
	(*RetrieveNotificationDetailRequest)(nil),  // 0: notificationmgmt.v2.RetrieveNotificationDetailRequest
	(*RetrieveNotificationDetailResponse)(nil), // 1: notificationmgmt.v2.RetrieveNotificationDetailResponse
	(*v1.Notification)(nil),                    // 2: common.v1.Notification
	(*v1.UserNotification)(nil),                // 3: common.v1.UserNotification
	(*v1.UserQuestionnaire)(nil),               // 4: common.v1.UserQuestionnaire
}
var file_notificationmgmt_v2_notifications_proto_depIdxs = []int32{
	2, // 0: notificationmgmt.v2.RetrieveNotificationDetailResponse.item:type_name -> common.v1.Notification
	3, // 1: notificationmgmt.v2.RetrieveNotificationDetailResponse.user_notification:type_name -> common.v1.UserNotification
	4, // 2: notificationmgmt.v2.RetrieveNotificationDetailResponse.user_questionnaire:type_name -> common.v1.UserQuestionnaire
	0, // 3: notificationmgmt.v2.NotificationReaderService.RetrieveNotificationDetail:input_type -> notificationmgmt.v2.RetrieveNotificationDetailRequest
	1, // 4: notificationmgmt.v2.NotificationReaderService.RetrieveNotificationDetail:output_type -> notificationmgmt.v2.RetrieveNotificationDetailResponse
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_notificationmgmt_v2_notifications_proto_init() }
func file_notificationmgmt_v2_notifications_proto_init() {
	if File_notificationmgmt_v2_notifications_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_notificationmgmt_v2_notifications_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetrieveNotificationDetailRequest); i {
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
		file_notificationmgmt_v2_notifications_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetrieveNotificationDetailResponse); i {
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
			RawDescriptor: file_notificationmgmt_v2_notifications_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_notificationmgmt_v2_notifications_proto_goTypes,
		DependencyIndexes: file_notificationmgmt_v2_notifications_proto_depIdxs,
		MessageInfos:      file_notificationmgmt_v2_notifications_proto_msgTypes,
	}.Build()
	File_notificationmgmt_v2_notifications_proto = out.File
	file_notificationmgmt_v2_notifications_proto_rawDesc = nil
	file_notificationmgmt_v2_notifications_proto_goTypes = nil
	file_notificationmgmt_v2_notifications_proto_depIdxs = nil
}
