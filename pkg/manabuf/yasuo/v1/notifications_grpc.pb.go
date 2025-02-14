// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package ypb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// NotificationModifierServiceClient is the client API for NotificationModifierService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NotificationModifierServiceClient interface {
	// Deprecated: Do not use.
	CreateNotification(ctx context.Context, in *CreateNotificationRequest, opts ...grpc.CallOption) (*CreateNotificationResponse, error)
	UpsertNotification(ctx context.Context, in *UpsertNotificationRequest, opts ...grpc.CallOption) (*UpsertNotificationResponse, error)
	SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*SendNotificationResponse, error)
	DiscardNotification(ctx context.Context, in *DiscardNotificationRequest, opts ...grpc.CallOption) (*DiscardNotificationResponse, error)
	NotifyUnreadUser(ctx context.Context, in *NotifyUnreadUserRequest, opts ...grpc.CallOption) (*NotifyUnreadUserResponse, error)
	SendScheduledNotification(ctx context.Context, in *SendScheduledNotificationRequest, opts ...grpc.CallOption) (*SendScheduledNotificationResponse, error)
	SubmitQuestionnaire(ctx context.Context, in *SubmitQuestionnaireRequest, opts ...grpc.CallOption) (*SubmitQuestionnaireResponse, error)
}

type notificationModifierServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNotificationModifierServiceClient(cc grpc.ClientConnInterface) NotificationModifierServiceClient {
	return &notificationModifierServiceClient{cc}
}

// Deprecated: Do not use.
func (c *notificationModifierServiceClient) CreateNotification(ctx context.Context, in *CreateNotificationRequest, opts ...grpc.CallOption) (*CreateNotificationResponse, error) {
	out := new(CreateNotificationResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.NotificationModifierService/CreateNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationModifierServiceClient) UpsertNotification(ctx context.Context, in *UpsertNotificationRequest, opts ...grpc.CallOption) (*UpsertNotificationResponse, error) {
	out := new(UpsertNotificationResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.NotificationModifierService/UpsertNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationModifierServiceClient) SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*SendNotificationResponse, error) {
	out := new(SendNotificationResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.NotificationModifierService/SendNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationModifierServiceClient) DiscardNotification(ctx context.Context, in *DiscardNotificationRequest, opts ...grpc.CallOption) (*DiscardNotificationResponse, error) {
	out := new(DiscardNotificationResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.NotificationModifierService/DiscardNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationModifierServiceClient) NotifyUnreadUser(ctx context.Context, in *NotifyUnreadUserRequest, opts ...grpc.CallOption) (*NotifyUnreadUserResponse, error) {
	out := new(NotifyUnreadUserResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.NotificationModifierService/NotifyUnreadUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationModifierServiceClient) SendScheduledNotification(ctx context.Context, in *SendScheduledNotificationRequest, opts ...grpc.CallOption) (*SendScheduledNotificationResponse, error) {
	out := new(SendScheduledNotificationResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.NotificationModifierService/SendScheduledNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationModifierServiceClient) SubmitQuestionnaire(ctx context.Context, in *SubmitQuestionnaireRequest, opts ...grpc.CallOption) (*SubmitQuestionnaireResponse, error) {
	out := new(SubmitQuestionnaireResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.NotificationModifierService/SubmitQuestionnaire", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NotificationModifierServiceServer is the server API for NotificationModifierService service.
// All implementations should embed UnimplementedNotificationModifierServiceServer
// for forward compatibility
type NotificationModifierServiceServer interface {
	// Deprecated: Do not use.
	CreateNotification(context.Context, *CreateNotificationRequest) (*CreateNotificationResponse, error)
	UpsertNotification(context.Context, *UpsertNotificationRequest) (*UpsertNotificationResponse, error)
	SendNotification(context.Context, *SendNotificationRequest) (*SendNotificationResponse, error)
	DiscardNotification(context.Context, *DiscardNotificationRequest) (*DiscardNotificationResponse, error)
	NotifyUnreadUser(context.Context, *NotifyUnreadUserRequest) (*NotifyUnreadUserResponse, error)
	SendScheduledNotification(context.Context, *SendScheduledNotificationRequest) (*SendScheduledNotificationResponse, error)
	SubmitQuestionnaire(context.Context, *SubmitQuestionnaireRequest) (*SubmitQuestionnaireResponse, error)
}

// UnimplementedNotificationModifierServiceServer should be embedded to have forward compatible implementations.
type UnimplementedNotificationModifierServiceServer struct {
}

func (UnimplementedNotificationModifierServiceServer) CreateNotification(context.Context, *CreateNotificationRequest) (*CreateNotificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateNotification not implemented")
}
func (UnimplementedNotificationModifierServiceServer) UpsertNotification(context.Context, *UpsertNotificationRequest) (*UpsertNotificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertNotification not implemented")
}
func (UnimplementedNotificationModifierServiceServer) SendNotification(context.Context, *SendNotificationRequest) (*SendNotificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendNotification not implemented")
}
func (UnimplementedNotificationModifierServiceServer) DiscardNotification(context.Context, *DiscardNotificationRequest) (*DiscardNotificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DiscardNotification not implemented")
}
func (UnimplementedNotificationModifierServiceServer) NotifyUnreadUser(context.Context, *NotifyUnreadUserRequest) (*NotifyUnreadUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyUnreadUser not implemented")
}
func (UnimplementedNotificationModifierServiceServer) SendScheduledNotification(context.Context, *SendScheduledNotificationRequest) (*SendScheduledNotificationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendScheduledNotification not implemented")
}
func (UnimplementedNotificationModifierServiceServer) SubmitQuestionnaire(context.Context, *SubmitQuestionnaireRequest) (*SubmitQuestionnaireResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitQuestionnaire not implemented")
}

// UnsafeNotificationModifierServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NotificationModifierServiceServer will
// result in compilation errors.
type UnsafeNotificationModifierServiceServer interface {
	mustEmbedUnimplementedNotificationModifierServiceServer()
}

func RegisterNotificationModifierServiceServer(s grpc.ServiceRegistrar, srv NotificationModifierServiceServer) {
	s.RegisterService(&_NotificationModifierService_serviceDesc, srv)
}

func _NotificationModifierService_CreateNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationModifierServiceServer).CreateNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.NotificationModifierService/CreateNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationModifierServiceServer).CreateNotification(ctx, req.(*CreateNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationModifierService_UpsertNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpsertNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationModifierServiceServer).UpsertNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.NotificationModifierService/UpsertNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationModifierServiceServer).UpsertNotification(ctx, req.(*UpsertNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationModifierService_SendNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationModifierServiceServer).SendNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.NotificationModifierService/SendNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationModifierServiceServer).SendNotification(ctx, req.(*SendNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationModifierService_DiscardNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DiscardNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationModifierServiceServer).DiscardNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.NotificationModifierService/DiscardNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationModifierServiceServer).DiscardNotification(ctx, req.(*DiscardNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationModifierService_NotifyUnreadUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NotifyUnreadUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationModifierServiceServer).NotifyUnreadUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.NotificationModifierService/NotifyUnreadUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationModifierServiceServer).NotifyUnreadUser(ctx, req.(*NotifyUnreadUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationModifierService_SendScheduledNotification_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendScheduledNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationModifierServiceServer).SendScheduledNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.NotificationModifierService/SendScheduledNotification",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationModifierServiceServer).SendScheduledNotification(ctx, req.(*SendScheduledNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationModifierService_SubmitQuestionnaire_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitQuestionnaireRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationModifierServiceServer).SubmitQuestionnaire(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.NotificationModifierService/SubmitQuestionnaire",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationModifierServiceServer).SubmitQuestionnaire(ctx, req.(*SubmitQuestionnaireRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _NotificationModifierService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "yasuo.v1.NotificationModifierService",
	HandlerType: (*NotificationModifierServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateNotification",
			Handler:    _NotificationModifierService_CreateNotification_Handler,
		},
		{
			MethodName: "UpsertNotification",
			Handler:    _NotificationModifierService_UpsertNotification_Handler,
		},
		{
			MethodName: "SendNotification",
			Handler:    _NotificationModifierService_SendNotification_Handler,
		},
		{
			MethodName: "DiscardNotification",
			Handler:    _NotificationModifierService_DiscardNotification_Handler,
		},
		{
			MethodName: "NotifyUnreadUser",
			Handler:    _NotificationModifierService_NotifyUnreadUser_Handler,
		},
		{
			MethodName: "SendScheduledNotification",
			Handler:    _NotificationModifierService_SendScheduledNotification_Handler,
		},
		{
			MethodName: "SubmitQuestionnaire",
			Handler:    _NotificationModifierService_SubmitQuestionnaire_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "yasuo/v1/notifications.proto",
}

// NotificationReaderServiceClient is the client API for NotificationReaderService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NotificationReaderServiceClient interface {
	RetrieveNotificationDetail(ctx context.Context, in *RetrieveNotificationDetailRequest, opts ...grpc.CallOption) (*RetrieveNotificationDetailResponse, error)
}

type notificationReaderServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNotificationReaderServiceClient(cc grpc.ClientConnInterface) NotificationReaderServiceClient {
	return &notificationReaderServiceClient{cc}
}

func (c *notificationReaderServiceClient) RetrieveNotificationDetail(ctx context.Context, in *RetrieveNotificationDetailRequest, opts ...grpc.CallOption) (*RetrieveNotificationDetailResponse, error) {
	out := new(RetrieveNotificationDetailResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.NotificationReaderService/RetrieveNotificationDetail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NotificationReaderServiceServer is the server API for NotificationReaderService service.
// All implementations should embed UnimplementedNotificationReaderServiceServer
// for forward compatibility
type NotificationReaderServiceServer interface {
	RetrieveNotificationDetail(context.Context, *RetrieveNotificationDetailRequest) (*RetrieveNotificationDetailResponse, error)
}

// UnimplementedNotificationReaderServiceServer should be embedded to have forward compatible implementations.
type UnimplementedNotificationReaderServiceServer struct {
}

func (UnimplementedNotificationReaderServiceServer) RetrieveNotificationDetail(context.Context, *RetrieveNotificationDetailRequest) (*RetrieveNotificationDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RetrieveNotificationDetail not implemented")
}

// UnsafeNotificationReaderServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NotificationReaderServiceServer will
// result in compilation errors.
type UnsafeNotificationReaderServiceServer interface {
	mustEmbedUnimplementedNotificationReaderServiceServer()
}

func RegisterNotificationReaderServiceServer(s grpc.ServiceRegistrar, srv NotificationReaderServiceServer) {
	s.RegisterService(&_NotificationReaderService_serviceDesc, srv)
}

func _NotificationReaderService_RetrieveNotificationDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RetrieveNotificationDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationReaderServiceServer).RetrieveNotificationDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.NotificationReaderService/RetrieveNotificationDetail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationReaderServiceServer).RetrieveNotificationDetail(ctx, req.(*RetrieveNotificationDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _NotificationReaderService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "yasuo.v1.NotificationReaderService",
	HandlerType: (*NotificationReaderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RetrieveNotificationDetail",
			Handler:    _NotificationReaderService_RetrieveNotificationDetail_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "yasuo/v1/notifications.proto",
}
