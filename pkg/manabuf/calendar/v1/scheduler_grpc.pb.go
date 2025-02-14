// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// SchedulerModifierServiceClient is the client API for SchedulerModifierService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SchedulerModifierServiceClient interface {
	CreateScheduler(ctx context.Context, in *CreateSchedulerRequest, opts ...grpc.CallOption) (*CreateSchedulerResponse, error)
	UpdateScheduler(ctx context.Context, in *UpdateSchedulerRequest, opts ...grpc.CallOption) (*UpdateSchedulerResponse, error)
	CreateManySchedulers(ctx context.Context, in *CreateManySchedulersRequest, opts ...grpc.CallOption) (*CreateManySchedulersResponse, error)
}

type schedulerModifierServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSchedulerModifierServiceClient(cc grpc.ClientConnInterface) SchedulerModifierServiceClient {
	return &schedulerModifierServiceClient{cc}
}

func (c *schedulerModifierServiceClient) CreateScheduler(ctx context.Context, in *CreateSchedulerRequest, opts ...grpc.CallOption) (*CreateSchedulerResponse, error) {
	out := new(CreateSchedulerResponse)
	err := c.cc.Invoke(ctx, "/calendar.v1.SchedulerModifierService/CreateScheduler", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerModifierServiceClient) UpdateScheduler(ctx context.Context, in *UpdateSchedulerRequest, opts ...grpc.CallOption) (*UpdateSchedulerResponse, error) {
	out := new(UpdateSchedulerResponse)
	err := c.cc.Invoke(ctx, "/calendar.v1.SchedulerModifierService/UpdateScheduler", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schedulerModifierServiceClient) CreateManySchedulers(ctx context.Context, in *CreateManySchedulersRequest, opts ...grpc.CallOption) (*CreateManySchedulersResponse, error) {
	out := new(CreateManySchedulersResponse)
	err := c.cc.Invoke(ctx, "/calendar.v1.SchedulerModifierService/CreateManySchedulers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SchedulerModifierServiceServer is the server API for SchedulerModifierService service.
// All implementations should embed UnimplementedSchedulerModifierServiceServer
// for forward compatibility
type SchedulerModifierServiceServer interface {
	CreateScheduler(context.Context, *CreateSchedulerRequest) (*CreateSchedulerResponse, error)
	UpdateScheduler(context.Context, *UpdateSchedulerRequest) (*UpdateSchedulerResponse, error)
	CreateManySchedulers(context.Context, *CreateManySchedulersRequest) (*CreateManySchedulersResponse, error)
}

// UnimplementedSchedulerModifierServiceServer should be embedded to have forward compatible implementations.
type UnimplementedSchedulerModifierServiceServer struct {
}

func (UnimplementedSchedulerModifierServiceServer) CreateScheduler(context.Context, *CreateSchedulerRequest) (*CreateSchedulerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateScheduler not implemented")
}
func (UnimplementedSchedulerModifierServiceServer) UpdateScheduler(context.Context, *UpdateSchedulerRequest) (*UpdateSchedulerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateScheduler not implemented")
}
func (UnimplementedSchedulerModifierServiceServer) CreateManySchedulers(context.Context, *CreateManySchedulersRequest) (*CreateManySchedulersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateManySchedulers not implemented")
}

// UnsafeSchedulerModifierServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SchedulerModifierServiceServer will
// result in compilation errors.
type UnsafeSchedulerModifierServiceServer interface {
	mustEmbedUnimplementedSchedulerModifierServiceServer()
}

func RegisterSchedulerModifierServiceServer(s grpc.ServiceRegistrar, srv SchedulerModifierServiceServer) {
	s.RegisterService(&_SchedulerModifierService_serviceDesc, srv)
}

func _SchedulerModifierService_CreateScheduler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSchedulerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerModifierServiceServer).CreateScheduler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/calendar.v1.SchedulerModifierService/CreateScheduler",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerModifierServiceServer).CreateScheduler(ctx, req.(*CreateSchedulerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchedulerModifierService_UpdateScheduler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSchedulerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerModifierServiceServer).UpdateScheduler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/calendar.v1.SchedulerModifierService/UpdateScheduler",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerModifierServiceServer).UpdateScheduler(ctx, req.(*UpdateSchedulerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchedulerModifierService_CreateManySchedulers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateManySchedulersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchedulerModifierServiceServer).CreateManySchedulers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/calendar.v1.SchedulerModifierService/CreateManySchedulers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchedulerModifierServiceServer).CreateManySchedulers(ctx, req.(*CreateManySchedulersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SchedulerModifierService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "calendar.v1.SchedulerModifierService",
	HandlerType: (*SchedulerModifierServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateScheduler",
			Handler:    _SchedulerModifierService_CreateScheduler_Handler,
		},
		{
			MethodName: "UpdateScheduler",
			Handler:    _SchedulerModifierService_UpdateScheduler_Handler,
		},
		{
			MethodName: "CreateManySchedulers",
			Handler:    _SchedulerModifierService_CreateManySchedulers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "calendar/v1/scheduler.proto",
}
