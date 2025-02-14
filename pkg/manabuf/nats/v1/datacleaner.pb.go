// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: nats/v1/datacleaner.proto

package npb

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

type ExtraCond struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Table     string `protobuf:"bytes,1,opt,name=table,proto3" json:"table,omitempty"`
	Condition string `protobuf:"bytes,2,opt,name=condition,proto3" json:"condition,omitempty"`
}

func (x *ExtraCond) Reset() {
	*x = ExtraCond{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nats_v1_datacleaner_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExtraCond) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExtraCond) ProtoMessage() {}

func (x *ExtraCond) ProtoReflect() protoreflect.Message {
	mi := &file_nats_v1_datacleaner_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExtraCond.ProtoReflect.Descriptor instead.
func (*ExtraCond) Descriptor() ([]byte, []int) {
	return file_nats_v1_datacleaner_proto_rawDescGZIP(), []int{0}
}

func (x *ExtraCond) GetTable() string {
	if x != nil {
		return x.Table
	}
	return ""
}

func (x *ExtraCond) GetCondition() string {
	if x != nil {
		return x.Condition
	}
	return ""
}

type EventDataClean struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Service   string       `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	SchoolId  string       `protobuf:"bytes,2,opt,name=school_id,json=schoolId,proto3" json:"school_id,omitempty"`
	Tables    string       `protobuf:"bytes,3,opt,name=tables,proto3" json:"tables,omitempty"`
	BeforeAt  string       `protobuf:"bytes,4,opt,name=before_at,json=beforeAt,proto3" json:"before_at,omitempty"`
	AfterAt   string       `protobuf:"bytes,5,opt,name=after_at,json=afterAt,proto3" json:"after_at,omitempty"`
	PerBatch  int32        `protobuf:"varint,6,opt,name=per_batch,json=perBatch,proto3" json:"per_batch,omitempty"`
	ExtraCond []*ExtraCond `protobuf:"bytes,7,rep,name=extra_cond,json=extraCond,proto3" json:"extra_cond,omitempty"`
}

func (x *EventDataClean) Reset() {
	*x = EventDataClean{}
	if protoimpl.UnsafeEnabled {
		mi := &file_nats_v1_datacleaner_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventDataClean) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventDataClean) ProtoMessage() {}

func (x *EventDataClean) ProtoReflect() protoreflect.Message {
	mi := &file_nats_v1_datacleaner_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventDataClean.ProtoReflect.Descriptor instead.
func (*EventDataClean) Descriptor() ([]byte, []int) {
	return file_nats_v1_datacleaner_proto_rawDescGZIP(), []int{1}
}

func (x *EventDataClean) GetService() string {
	if x != nil {
		return x.Service
	}
	return ""
}

func (x *EventDataClean) GetSchoolId() string {
	if x != nil {
		return x.SchoolId
	}
	return ""
}

func (x *EventDataClean) GetTables() string {
	if x != nil {
		return x.Tables
	}
	return ""
}

func (x *EventDataClean) GetBeforeAt() string {
	if x != nil {
		return x.BeforeAt
	}
	return ""
}

func (x *EventDataClean) GetAfterAt() string {
	if x != nil {
		return x.AfterAt
	}
	return ""
}

func (x *EventDataClean) GetPerBatch() int32 {
	if x != nil {
		return x.PerBatch
	}
	return 0
}

func (x *EventDataClean) GetExtraCond() []*ExtraCond {
	if x != nil {
		return x.ExtraCond
	}
	return nil
}

var File_nats_v1_datacleaner_proto protoreflect.FileDescriptor

var file_nats_v1_datacleaner_proto_rawDesc = []byte{
	0x0a, 0x19, 0x6e, 0x61, 0x74, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x63, 0x6c,
	0x65, 0x61, 0x6e, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x6e, 0x61, 0x74,
	0x73, 0x2e, 0x76, 0x31, 0x22, 0x3f, 0x0a, 0x09, 0x45, 0x78, 0x74, 0x72, 0x61, 0x43, 0x6f, 0x6e,
	0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x6f, 0x6e, 0x64, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x64,
	0x69, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xe7, 0x01, 0x0a, 0x0e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x44,
	0x61, 0x74, 0x61, 0x43, 0x6c, 0x65, 0x61, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x63, 0x68, 0x6f, 0x6f, 0x6c, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x63, 0x68, 0x6f, 0x6f, 0x6c, 0x49, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x62, 0x65, 0x66, 0x6f, 0x72,
	0x65, 0x5f, 0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x62, 0x65, 0x66, 0x6f,
	0x72, 0x65, 0x41, 0x74, 0x12, 0x19, 0x0a, 0x08, 0x61, 0x66, 0x74, 0x65, 0x72, 0x5f, 0x61, 0x74,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x66, 0x74, 0x65, 0x72, 0x41, 0x74, 0x12,
	0x1b, 0x0a, 0x09, 0x70, 0x65, 0x72, 0x5f, 0x62, 0x61, 0x74, 0x63, 0x68, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x08, 0x70, 0x65, 0x72, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12, 0x31, 0x0a, 0x0a,
	0x65, 0x78, 0x74, 0x72, 0x61, 0x5f, 0x63, 0x6f, 0x6e, 0x64, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x12, 0x2e, 0x6e, 0x61, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x74, 0x72, 0x61,
	0x43, 0x6f, 0x6e, 0x64, 0x52, 0x09, 0x65, 0x78, 0x74, 0x72, 0x61, 0x43, 0x6f, 0x6e, 0x64, 0x42,
	0x38, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61,
	0x6e, 0x61, 0x62, 0x69, 0x65, 0x2d, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e,
	0x64, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x75, 0x66, 0x2f, 0x6e, 0x61,
	0x74, 0x73, 0x2f, 0x76, 0x31, 0x3b, 0x6e, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_nats_v1_datacleaner_proto_rawDescOnce sync.Once
	file_nats_v1_datacleaner_proto_rawDescData = file_nats_v1_datacleaner_proto_rawDesc
)

func file_nats_v1_datacleaner_proto_rawDescGZIP() []byte {
	file_nats_v1_datacleaner_proto_rawDescOnce.Do(func() {
		file_nats_v1_datacleaner_proto_rawDescData = protoimpl.X.CompressGZIP(file_nats_v1_datacleaner_proto_rawDescData)
	})
	return file_nats_v1_datacleaner_proto_rawDescData
}

var file_nats_v1_datacleaner_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_nats_v1_datacleaner_proto_goTypes = []interface{}{
	(*ExtraCond)(nil),      // 0: nats.v1.ExtraCond
	(*EventDataClean)(nil), // 1: nats.v1.EventDataClean
}
var file_nats_v1_datacleaner_proto_depIdxs = []int32{
	0, // 0: nats.v1.EventDataClean.extra_cond:type_name -> nats.v1.ExtraCond
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_nats_v1_datacleaner_proto_init() }
func file_nats_v1_datacleaner_proto_init() {
	if File_nats_v1_datacleaner_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_nats_v1_datacleaner_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExtraCond); i {
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
		file_nats_v1_datacleaner_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventDataClean); i {
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
			RawDescriptor: file_nats_v1_datacleaner_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_nats_v1_datacleaner_proto_goTypes,
		DependencyIndexes: file_nats_v1_datacleaner_proto_depIdxs,
		MessageInfos:      file_nats_v1_datacleaner_proto_msgTypes,
	}.Build()
	File_nats_v1_datacleaner_proto = out.File
	file_nats_v1_datacleaner_proto_rawDesc = nil
	file_nats_v1_datacleaner_proto_goTypes = nil
	file_nats_v1_datacleaner_proto_depIdxs = nil
}
