// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: invoicemgmt/v1/payment_detail.proto

package invoice_pb

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/timestamppb"
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

type BillingAddress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BillingAddressId string `protobuf:"bytes,1,opt,name=billing_address_id,json=billingAddressId,proto3" json:"billing_address_id,omitempty"`
	PostalCode       string `protobuf:"bytes,2,opt,name=postal_code,json=postalCode,proto3" json:"postal_code,omitempty"`
	// Deprecated: Do not use.
	Prefecture string `protobuf:"bytes,3,opt,name=prefecture,proto3" json:"prefecture,omitempty"`
	City       string `protobuf:"bytes,4,opt,name=city,proto3" json:"city,omitempty"`
	Street1    string `protobuf:"bytes,5,opt,name=street1,proto3" json:"street1,omitempty"`
	Street2    string `protobuf:"bytes,6,opt,name=street2,proto3" json:"street2,omitempty"`
	// Deprecated: Do not use.
	PrefectureId   string `protobuf:"bytes,7,opt,name=prefecture_id,json=prefectureId,proto3" json:"prefecture_id,omitempty"`
	PrefectureCode string `protobuf:"bytes,8,opt,name=prefecture_code,json=prefectureCode,proto3" json:"prefecture_code,omitempty"`
}

func (x *BillingAddress) Reset() {
	*x = BillingAddress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BillingAddress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BillingAddress) ProtoMessage() {}

func (x *BillingAddress) ProtoReflect() protoreflect.Message {
	mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BillingAddress.ProtoReflect.Descriptor instead.
func (*BillingAddress) Descriptor() ([]byte, []int) {
	return file_invoicemgmt_v1_payment_detail_proto_rawDescGZIP(), []int{0}
}

func (x *BillingAddress) GetBillingAddressId() string {
	if x != nil {
		return x.BillingAddressId
	}
	return ""
}

func (x *BillingAddress) GetPostalCode() string {
	if x != nil {
		return x.PostalCode
	}
	return ""
}

// Deprecated: Do not use.
func (x *BillingAddress) GetPrefecture() string {
	if x != nil {
		return x.Prefecture
	}
	return ""
}

func (x *BillingAddress) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

func (x *BillingAddress) GetStreet1() string {
	if x != nil {
		return x.Street1
	}
	return ""
}

func (x *BillingAddress) GetStreet2() string {
	if x != nil {
		return x.Street2
	}
	return ""
}

// Deprecated: Do not use.
func (x *BillingAddress) GetPrefectureId() string {
	if x != nil {
		return x.PrefectureId
	}
	return ""
}

func (x *BillingAddress) GetPrefectureCode() string {
	if x != nil {
		return x.PrefectureCode
	}
	return ""
}

type BillingInformation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StudentPaymentDetailId string          `protobuf:"bytes,1,opt,name=student_payment_detail_id,json=studentPaymentDetailId,proto3" json:"student_payment_detail_id,omitempty"`
	PayerName              string          `protobuf:"bytes,2,opt,name=payer_name,json=payerName,proto3" json:"payer_name,omitempty"`
	PayerPhoneNumber       string          `protobuf:"bytes,3,opt,name=payer_phone_number,json=payerPhoneNumber,proto3" json:"payer_phone_number,omitempty"`
	BillingAddress         *BillingAddress `protobuf:"bytes,4,opt,name=billing_address,json=billingAddress,proto3" json:"billing_address,omitempty"`
}

func (x *BillingInformation) Reset() {
	*x = BillingInformation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BillingInformation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BillingInformation) ProtoMessage() {}

func (x *BillingInformation) ProtoReflect() protoreflect.Message {
	mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BillingInformation.ProtoReflect.Descriptor instead.
func (*BillingInformation) Descriptor() ([]byte, []int) {
	return file_invoicemgmt_v1_payment_detail_proto_rawDescGZIP(), []int{1}
}

func (x *BillingInformation) GetStudentPaymentDetailId() string {
	if x != nil {
		return x.StudentPaymentDetailId
	}
	return ""
}

func (x *BillingInformation) GetPayerName() string {
	if x != nil {
		return x.PayerName
	}
	return ""
}

func (x *BillingInformation) GetPayerPhoneNumber() string {
	if x != nil {
		return x.PayerPhoneNumber
	}
	return ""
}

func (x *BillingInformation) GetBillingAddress() *BillingAddress {
	if x != nil {
		return x.BillingAddress
	}
	return nil
}

type BankAccountInformation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	BankAccountId     string          `protobuf:"bytes,1,opt,name=bank_account_id,json=bankAccountId,proto3" json:"bank_account_id,omitempty"`
	BankId            string          `protobuf:"bytes,2,opt,name=bank_id,json=bankId,proto3" json:"bank_id,omitempty"`
	BankBranchId      string          `protobuf:"bytes,3,opt,name=bank_branch_id,json=bankBranchId,proto3" json:"bank_branch_id,omitempty"`
	BankAccountNumber string          `protobuf:"bytes,4,opt,name=bank_account_number,json=bankAccountNumber,proto3" json:"bank_account_number,omitempty"`
	BankAccountHolder string          `protobuf:"bytes,5,opt,name=bank_account_holder,json=bankAccountHolder,proto3" json:"bank_account_holder,omitempty"`
	BankAccountType   BankAccountType `protobuf:"varint,6,opt,name=bank_account_type,json=bankAccountType,proto3,enum=invoicemgmt.v1.BankAccountType" json:"bank_account_type,omitempty"`
	IsVerified        bool            `protobuf:"varint,7,opt,name=is_verified,json=isVerified,proto3" json:"is_verified,omitempty"`
}

func (x *BankAccountInformation) Reset() {
	*x = BankAccountInformation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BankAccountInformation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BankAccountInformation) ProtoMessage() {}

func (x *BankAccountInformation) ProtoReflect() protoreflect.Message {
	mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BankAccountInformation.ProtoReflect.Descriptor instead.
func (*BankAccountInformation) Descriptor() ([]byte, []int) {
	return file_invoicemgmt_v1_payment_detail_proto_rawDescGZIP(), []int{2}
}

func (x *BankAccountInformation) GetBankAccountId() string {
	if x != nil {
		return x.BankAccountId
	}
	return ""
}

func (x *BankAccountInformation) GetBankId() string {
	if x != nil {
		return x.BankId
	}
	return ""
}

func (x *BankAccountInformation) GetBankBranchId() string {
	if x != nil {
		return x.BankBranchId
	}
	return ""
}

func (x *BankAccountInformation) GetBankAccountNumber() string {
	if x != nil {
		return x.BankAccountNumber
	}
	return ""
}

func (x *BankAccountInformation) GetBankAccountHolder() string {
	if x != nil {
		return x.BankAccountHolder
	}
	return ""
}

func (x *BankAccountInformation) GetBankAccountType() BankAccountType {
	if x != nil {
		return x.BankAccountType
	}
	return BankAccountType_SAVINGS_ACCOUNT
}

func (x *BankAccountInformation) GetIsVerified() bool {
	if x != nil {
		return x.IsVerified
	}
	return false
}

type UpsertStudentPaymentInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StudentId       string                  `protobuf:"bytes,1,opt,name=student_id,json=studentId,proto3" json:"student_id,omitempty"`
	BillingInfo     *BillingInformation     `protobuf:"bytes,2,opt,name=billing_info,json=billingInfo,proto3" json:"billing_info,omitempty"`
	BankAccountInfo *BankAccountInformation `protobuf:"bytes,3,opt,name=bank_account_info,json=bankAccountInfo,proto3" json:"bank_account_info,omitempty"`
}

func (x *UpsertStudentPaymentInfoRequest) Reset() {
	*x = UpsertStudentPaymentInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertStudentPaymentInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertStudentPaymentInfoRequest) ProtoMessage() {}

func (x *UpsertStudentPaymentInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertStudentPaymentInfoRequest.ProtoReflect.Descriptor instead.
func (*UpsertStudentPaymentInfoRequest) Descriptor() ([]byte, []int) {
	return file_invoicemgmt_v1_payment_detail_proto_rawDescGZIP(), []int{3}
}

func (x *UpsertStudentPaymentInfoRequest) GetStudentId() string {
	if x != nil {
		return x.StudentId
	}
	return ""
}

func (x *UpsertStudentPaymentInfoRequest) GetBillingInfo() *BillingInformation {
	if x != nil {
		return x.BillingInfo
	}
	return nil
}

func (x *UpsertStudentPaymentInfoRequest) GetBankAccountInfo() *BankAccountInformation {
	if x != nil {
		return x.BankAccountInfo
	}
	return nil
}

type UpsertStudentPaymentInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Successful bool `protobuf:"varint,1,opt,name=successful,proto3" json:"successful,omitempty"`
}

func (x *UpsertStudentPaymentInfoResponse) Reset() {
	*x = UpsertStudentPaymentInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpsertStudentPaymentInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpsertStudentPaymentInfoResponse) ProtoMessage() {}

func (x *UpsertStudentPaymentInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpsertStudentPaymentInfoResponse.ProtoReflect.Descriptor instead.
func (*UpsertStudentPaymentInfoResponse) Descriptor() ([]byte, []int) {
	return file_invoicemgmt_v1_payment_detail_proto_rawDescGZIP(), []int{4}
}

func (x *UpsertStudentPaymentInfoResponse) GetSuccessful() bool {
	if x != nil {
		return x.Successful
	}
	return false
}

type UpdateStudentPaymentMethodRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StudentId              string        `protobuf:"bytes,1,opt,name=student_id,json=studentId,proto3" json:"student_id,omitempty"`
	StudentPaymentDetailId string        `protobuf:"bytes,2,opt,name=student_payment_detail_id,json=studentPaymentDetailId,proto3" json:"student_payment_detail_id,omitempty"`
	PaymentMethod          PaymentMethod `protobuf:"varint,3,opt,name=payment_method,json=paymentMethod,proto3,enum=invoicemgmt.v1.PaymentMethod" json:"payment_method,omitempty"`
}

func (x *UpdateStudentPaymentMethodRequest) Reset() {
	*x = UpdateStudentPaymentMethodRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateStudentPaymentMethodRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateStudentPaymentMethodRequest) ProtoMessage() {}

func (x *UpdateStudentPaymentMethodRequest) ProtoReflect() protoreflect.Message {
	mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateStudentPaymentMethodRequest.ProtoReflect.Descriptor instead.
func (*UpdateStudentPaymentMethodRequest) Descriptor() ([]byte, []int) {
	return file_invoicemgmt_v1_payment_detail_proto_rawDescGZIP(), []int{5}
}

func (x *UpdateStudentPaymentMethodRequest) GetStudentId() string {
	if x != nil {
		return x.StudentId
	}
	return ""
}

func (x *UpdateStudentPaymentMethodRequest) GetStudentPaymentDetailId() string {
	if x != nil {
		return x.StudentPaymentDetailId
	}
	return ""
}

func (x *UpdateStudentPaymentMethodRequest) GetPaymentMethod() PaymentMethod {
	if x != nil {
		return x.PaymentMethod
	}
	return PaymentMethod_DIRECT_DEBIT
}

type UpdateStudentPaymentMethodResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Successful bool `protobuf:"varint,1,opt,name=successful,proto3" json:"successful,omitempty"`
}

func (x *UpdateStudentPaymentMethodResponse) Reset() {
	*x = UpdateStudentPaymentMethodResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateStudentPaymentMethodResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateStudentPaymentMethodResponse) ProtoMessage() {}

func (x *UpdateStudentPaymentMethodResponse) ProtoReflect() protoreflect.Message {
	mi := &file_invoicemgmt_v1_payment_detail_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateStudentPaymentMethodResponse.ProtoReflect.Descriptor instead.
func (*UpdateStudentPaymentMethodResponse) Descriptor() ([]byte, []int) {
	return file_invoicemgmt_v1_payment_detail_proto_rawDescGZIP(), []int{6}
}

func (x *UpdateStudentPaymentMethodResponse) GetSuccessful() bool {
	if x != nil {
		return x.Successful
	}
	return false
}

var File_invoicemgmt_v1_payment_detail_proto protoreflect.FileDescriptor

var file_invoicemgmt_v1_payment_detail_proto_rawDesc = []byte{
	0x0a, 0x23, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x6d, 0x67, 0x6d, 0x74, 0x2f, 0x76, 0x31,
	0x2f, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x6d, 0x67,
	0x6d, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x6d,
	0x67, 0x6d, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x65, 0x6e, 0x75, 0x6d, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x9d, 0x02, 0x0a, 0x0e, 0x42, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x2c, 0x0a, 0x12, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67,
	0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x10, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x6f, 0x73, 0x74, 0x61, 0x6c, 0x5f, 0x63, 0x6f,
	0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x6f, 0x73, 0x74, 0x61, 0x6c,
	0x43, 0x6f, 0x64, 0x65, 0x12, 0x22, 0x0a, 0x0a, 0x70, 0x72, 0x65, 0x66, 0x65, 0x63, 0x74, 0x75,
	0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x02, 0x18, 0x01, 0x52, 0x0a, 0x70, 0x72,
	0x65, 0x66, 0x65, 0x63, 0x74, 0x75, 0x72, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x69, 0x74, 0x79,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x69, 0x74, 0x79, 0x12, 0x18, 0x0a, 0x07,
	0x73, 0x74, 0x72, 0x65, 0x65, 0x74, 0x31, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73,
	0x74, 0x72, 0x65, 0x65, 0x74, 0x31, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x72, 0x65, 0x65, 0x74,
	0x32, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x74, 0x72, 0x65, 0x65, 0x74, 0x32,
	0x12, 0x27, 0x0a, 0x0d, 0x70, 0x72, 0x65, 0x66, 0x65, 0x63, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x69,
	0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x42, 0x02, 0x18, 0x01, 0x52, 0x0c, 0x70, 0x72, 0x65,
	0x66, 0x65, 0x63, 0x74, 0x75, 0x72, 0x65, 0x49, 0x64, 0x12, 0x27, 0x0a, 0x0f, 0x70, 0x72, 0x65,
	0x66, 0x65, 0x63, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0e, 0x70, 0x72, 0x65, 0x66, 0x65, 0x63, 0x74, 0x75, 0x72, 0x65, 0x43, 0x6f,
	0x64, 0x65, 0x22, 0xe5, 0x01, 0x0a, 0x12, 0x42, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x39, 0x0a, 0x19, 0x73, 0x74, 0x75,
	0x64, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x16, 0x73, 0x74,
	0x75, 0x64, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x44, 0x65, 0x74, 0x61,
	0x69, 0x6c, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x61, 0x79, 0x65, 0x72, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x2c, 0x0a, 0x12, 0x70, 0x61, 0x79, 0x65, 0x72, 0x5f, 0x70, 0x68, 0x6f,
	0x6e, 0x65, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x10, 0x70, 0x61, 0x79, 0x65, 0x72, 0x50, 0x68, 0x6f, 0x6e, 0x65, 0x4e, 0x75, 0x6d, 0x62, 0x65,
	0x72, 0x12, 0x47, 0x0a, 0x0f, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x69, 0x6e, 0x76,
	0x6f, 0x69, 0x63, 0x65, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x69, 0x6c, 0x6c,
	0x69, 0x6e, 0x67, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x0e, 0x62, 0x69, 0x6c, 0x6c,
	0x69, 0x6e, 0x67, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0xcd, 0x02, 0x0a, 0x16, 0x42,
	0x61, 0x6e, 0x6b, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x72, 0x6d,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x26, 0x0a, 0x0f, 0x62, 0x61, 0x6e, 0x6b, 0x5f, 0x61, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d,
	0x62, 0x61, 0x6e, 0x6b, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x17, 0x0a,
	0x07, 0x62, 0x61, 0x6e, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x62, 0x61, 0x6e, 0x6b, 0x49, 0x64, 0x12, 0x24, 0x0a, 0x0e, 0x62, 0x61, 0x6e, 0x6b, 0x5f, 0x62,
	0x72, 0x61, 0x6e, 0x63, 0x68, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c,
	0x62, 0x61, 0x6e, 0x6b, 0x42, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x49, 0x64, 0x12, 0x2e, 0x0a, 0x13,
	0x62, 0x61, 0x6e, 0x6b, 0x5f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x6e, 0x75, 0x6d,
	0x62, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x62, 0x61, 0x6e, 0x6b, 0x41,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x2e, 0x0a, 0x13,
	0x62, 0x61, 0x6e, 0x6b, 0x5f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x68, 0x6f, 0x6c,
	0x64, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x62, 0x61, 0x6e, 0x6b, 0x41,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x48, 0x6f, 0x6c, 0x64, 0x65, 0x72, 0x12, 0x4b, 0x0a, 0x11,
	0x62, 0x61, 0x6e, 0x6b, 0x5f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1f, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63,
	0x65, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x61, 0x6e, 0x6b, 0x41, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0f, 0x62, 0x61, 0x6e, 0x6b, 0x41, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x73, 0x5f,
	0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a,
	0x69, 0x73, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64, 0x22, 0xdb, 0x01, 0x0a, 0x1f, 0x55,
	0x70, 0x73, 0x65, 0x72, 0x74, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x79, 0x6d,
	0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d,
	0x0a, 0x0a, 0x73, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x73, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x45, 0x0a,
	0x0c, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x6d, 0x67, 0x6d,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x66, 0x6f,
	0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x52, 0x0a, 0x11, 0x62, 0x61, 0x6e, 0x6b, 0x5f, 0x61, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x26, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31,
	0x2e, 0x42, 0x61, 0x6e, 0x6b, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f,
	0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0f, 0x62, 0x61, 0x6e, 0x6b, 0x41, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x42, 0x0a, 0x20, 0x55, 0x70, 0x73, 0x65,
	0x72, 0x74, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a, 0x0a,
	0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x66, 0x75, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x0a, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x66, 0x75, 0x6c, 0x22, 0xc3, 0x01, 0x0a,
	0x21, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x50, 0x61,
	0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x49,
	0x64, 0x12, 0x39, 0x0a, 0x19, 0x73, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x5f, 0x70, 0x61, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x16, 0x73, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x49, 0x64, 0x12, 0x44, 0x0a, 0x0e,
	0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x1d, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x6d, 0x67,
	0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x52, 0x0d, 0x70, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x22, 0x44, 0x0a, 0x22, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x75, 0x64,
	0x65, 0x6e, 0x74, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x75, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x66, 0x75, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x66, 0x75, 0x6c, 0x32, 0x9f, 0x02, 0x0a, 0x18, 0x45, 0x64, 0x69,
	0x74, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x7d, 0x0a, 0x18, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x53,
	0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66,
	0x6f, 0x12, 0x2f, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x6d, 0x67, 0x6d, 0x74, 0x2e,
	0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e, 0x74,
	0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x30, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x6d, 0x67, 0x6d, 0x74,
	0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x73, 0x65, 0x72, 0x74, 0x53, 0x74, 0x75, 0x64, 0x65, 0x6e,
	0x74, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x83, 0x01, 0x0a, 0x1a, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53,
	0x74, 0x75, 0x64, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x12, 0x31, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x6d, 0x67, 0x6d,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x75, 0x64, 0x65,
	0x6e, 0x74, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x32, 0x2e, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65,
	0x6d, 0x67, 0x6d, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74,
	0x75, 0x64, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x46, 0x5a, 0x44, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x69, 0x65,
	0x2d, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x6d, 0x61, 0x6e, 0x61, 0x62, 0x75, 0x66, 0x2f, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65,
	0x6d, 0x67, 0x6d, 0x74, 0x2f, 0x76, 0x31, 0x3b, 0x69, 0x6e, 0x76, 0x6f, 0x69, 0x63, 0x65, 0x5f,
	0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_invoicemgmt_v1_payment_detail_proto_rawDescOnce sync.Once
	file_invoicemgmt_v1_payment_detail_proto_rawDescData = file_invoicemgmt_v1_payment_detail_proto_rawDesc
)

func file_invoicemgmt_v1_payment_detail_proto_rawDescGZIP() []byte {
	file_invoicemgmt_v1_payment_detail_proto_rawDescOnce.Do(func() {
		file_invoicemgmt_v1_payment_detail_proto_rawDescData = protoimpl.X.CompressGZIP(file_invoicemgmt_v1_payment_detail_proto_rawDescData)
	})
	return file_invoicemgmt_v1_payment_detail_proto_rawDescData
}

var file_invoicemgmt_v1_payment_detail_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_invoicemgmt_v1_payment_detail_proto_goTypes = []interface{}{
	(*BillingAddress)(nil),                     // 0: invoicemgmt.v1.BillingAddress
	(*BillingInformation)(nil),                 // 1: invoicemgmt.v1.BillingInformation
	(*BankAccountInformation)(nil),             // 2: invoicemgmt.v1.BankAccountInformation
	(*UpsertStudentPaymentInfoRequest)(nil),    // 3: invoicemgmt.v1.UpsertStudentPaymentInfoRequest
	(*UpsertStudentPaymentInfoResponse)(nil),   // 4: invoicemgmt.v1.UpsertStudentPaymentInfoResponse
	(*UpdateStudentPaymentMethodRequest)(nil),  // 5: invoicemgmt.v1.UpdateStudentPaymentMethodRequest
	(*UpdateStudentPaymentMethodResponse)(nil), // 6: invoicemgmt.v1.UpdateStudentPaymentMethodResponse
	(BankAccountType)(0),                       // 7: invoicemgmt.v1.BankAccountType
	(PaymentMethod)(0),                         // 8: invoicemgmt.v1.PaymentMethod
}
var file_invoicemgmt_v1_payment_detail_proto_depIdxs = []int32{
	0, // 0: invoicemgmt.v1.BillingInformation.billing_address:type_name -> invoicemgmt.v1.BillingAddress
	7, // 1: invoicemgmt.v1.BankAccountInformation.bank_account_type:type_name -> invoicemgmt.v1.BankAccountType
	1, // 2: invoicemgmt.v1.UpsertStudentPaymentInfoRequest.billing_info:type_name -> invoicemgmt.v1.BillingInformation
	2, // 3: invoicemgmt.v1.UpsertStudentPaymentInfoRequest.bank_account_info:type_name -> invoicemgmt.v1.BankAccountInformation
	8, // 4: invoicemgmt.v1.UpdateStudentPaymentMethodRequest.payment_method:type_name -> invoicemgmt.v1.PaymentMethod
	3, // 5: invoicemgmt.v1.EditPaymentDetailService.UpsertStudentPaymentInfo:input_type -> invoicemgmt.v1.UpsertStudentPaymentInfoRequest
	5, // 6: invoicemgmt.v1.EditPaymentDetailService.UpdateStudentPaymentMethod:input_type -> invoicemgmt.v1.UpdateStudentPaymentMethodRequest
	4, // 7: invoicemgmt.v1.EditPaymentDetailService.UpsertStudentPaymentInfo:output_type -> invoicemgmt.v1.UpsertStudentPaymentInfoResponse
	6, // 8: invoicemgmt.v1.EditPaymentDetailService.UpdateStudentPaymentMethod:output_type -> invoicemgmt.v1.UpdateStudentPaymentMethodResponse
	7, // [7:9] is the sub-list for method output_type
	5, // [5:7] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_invoicemgmt_v1_payment_detail_proto_init() }
func file_invoicemgmt_v1_payment_detail_proto_init() {
	if File_invoicemgmt_v1_payment_detail_proto != nil {
		return
	}
	file_invoicemgmt_v1_enums_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_invoicemgmt_v1_payment_detail_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BillingAddress); i {
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
		file_invoicemgmt_v1_payment_detail_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BillingInformation); i {
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
		file_invoicemgmt_v1_payment_detail_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BankAccountInformation); i {
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
		file_invoicemgmt_v1_payment_detail_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertStudentPaymentInfoRequest); i {
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
		file_invoicemgmt_v1_payment_detail_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpsertStudentPaymentInfoResponse); i {
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
		file_invoicemgmt_v1_payment_detail_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateStudentPaymentMethodRequest); i {
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
		file_invoicemgmt_v1_payment_detail_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateStudentPaymentMethodResponse); i {
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
			RawDescriptor: file_invoicemgmt_v1_payment_detail_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_invoicemgmt_v1_payment_detail_proto_goTypes,
		DependencyIndexes: file_invoicemgmt_v1_payment_detail_proto_depIdxs,
		MessageInfos:      file_invoicemgmt_v1_payment_detail_proto_msgTypes,
	}.Build()
	File_invoicemgmt_v1_payment_detail_proto = out.File
	file_invoicemgmt_v1_payment_detail_proto_rawDesc = nil
	file_invoicemgmt_v1_payment_detail_proto_goTypes = nil
	file_invoicemgmt_v1_payment_detail_proto_depIdxs = nil
}
