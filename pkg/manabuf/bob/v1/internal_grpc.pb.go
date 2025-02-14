// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package bpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// InternalModifierServiceClient is the client API for InternalModifierService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type InternalModifierServiceClient interface {
	SubmitQuizAnswers(ctx context.Context, in *SubmitQuizAnswersRequest, opts ...grpc.CallOption) (*SubmitQuizAnswersResponse, error)
}

type internalModifierServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewInternalModifierServiceClient(cc grpc.ClientConnInterface) InternalModifierServiceClient {
	return &internalModifierServiceClient{cc}
}

func (c *internalModifierServiceClient) SubmitQuizAnswers(ctx context.Context, in *SubmitQuizAnswersRequest, opts ...grpc.CallOption) (*SubmitQuizAnswersResponse, error) {
	out := new(SubmitQuizAnswersResponse)
	err := c.cc.Invoke(ctx, "/bob.v1.InternalModifierService/SubmitQuizAnswers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InternalModifierServiceServer is the server API for InternalModifierService service.
// All implementations should embed UnimplementedInternalModifierServiceServer
// for forward compatibility
type InternalModifierServiceServer interface {
	SubmitQuizAnswers(context.Context, *SubmitQuizAnswersRequest) (*SubmitQuizAnswersResponse, error)
}

// UnimplementedInternalModifierServiceServer should be embedded to have forward compatible implementations.
type UnimplementedInternalModifierServiceServer struct {
}

func (UnimplementedInternalModifierServiceServer) SubmitQuizAnswers(context.Context, *SubmitQuizAnswersRequest) (*SubmitQuizAnswersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitQuizAnswers not implemented")
}

// UnsafeInternalModifierServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to InternalModifierServiceServer will
// result in compilation errors.
type UnsafeInternalModifierServiceServer interface {
	mustEmbedUnimplementedInternalModifierServiceServer()
}

func RegisterInternalModifierServiceServer(s grpc.ServiceRegistrar, srv InternalModifierServiceServer) {
	s.RegisterService(&_InternalModifierService_serviceDesc, srv)
}

func _InternalModifierService_SubmitQuizAnswers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitQuizAnswersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InternalModifierServiceServer).SubmitQuizAnswers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bob.v1.InternalModifierService/SubmitQuizAnswers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InternalModifierServiceServer).SubmitQuizAnswers(ctx, req.(*SubmitQuizAnswersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _InternalModifierService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bob.v1.InternalModifierService",
	HandlerType: (*InternalModifierServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitQuizAnswers",
			Handler:    _InternalModifierService_SubmitQuizAnswers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "bob/v1/internal.proto",
}

// InternalReaderServiceClient is the client API for InternalReaderService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type InternalReaderServiceClient interface {
	RetrieveTopics(ctx context.Context, in *RetrieveTopicsRequest, opts ...grpc.CallOption) (*RetrieveTopicsResponse, error)
	VerifyAppVersion(ctx context.Context, in *VerifyAppVersionRequest, opts ...grpc.CallOption) (*VerifyAppVersionResponse, error)
}

type internalReaderServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewInternalReaderServiceClient(cc grpc.ClientConnInterface) InternalReaderServiceClient {
	return &internalReaderServiceClient{cc}
}

func (c *internalReaderServiceClient) RetrieveTopics(ctx context.Context, in *RetrieveTopicsRequest, opts ...grpc.CallOption) (*RetrieveTopicsResponse, error) {
	out := new(RetrieveTopicsResponse)
	err := c.cc.Invoke(ctx, "/bob.v1.InternalReaderService/RetrieveTopics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *internalReaderServiceClient) VerifyAppVersion(ctx context.Context, in *VerifyAppVersionRequest, opts ...grpc.CallOption) (*VerifyAppVersionResponse, error) {
	out := new(VerifyAppVersionResponse)
	err := c.cc.Invoke(ctx, "/bob.v1.InternalReaderService/VerifyAppVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InternalReaderServiceServer is the server API for InternalReaderService service.
// All implementations should embed UnimplementedInternalReaderServiceServer
// for forward compatibility
type InternalReaderServiceServer interface {
	RetrieveTopics(context.Context, *RetrieveTopicsRequest) (*RetrieveTopicsResponse, error)
	VerifyAppVersion(context.Context, *VerifyAppVersionRequest) (*VerifyAppVersionResponse, error)
}

// UnimplementedInternalReaderServiceServer should be embedded to have forward compatible implementations.
type UnimplementedInternalReaderServiceServer struct {
}

func (UnimplementedInternalReaderServiceServer) RetrieveTopics(context.Context, *RetrieveTopicsRequest) (*RetrieveTopicsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RetrieveTopics not implemented")
}
func (UnimplementedInternalReaderServiceServer) VerifyAppVersion(context.Context, *VerifyAppVersionRequest) (*VerifyAppVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyAppVersion not implemented")
}

// UnsafeInternalReaderServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to InternalReaderServiceServer will
// result in compilation errors.
type UnsafeInternalReaderServiceServer interface {
	mustEmbedUnimplementedInternalReaderServiceServer()
}

func RegisterInternalReaderServiceServer(s grpc.ServiceRegistrar, srv InternalReaderServiceServer) {
	s.RegisterService(&_InternalReaderService_serviceDesc, srv)
}

func _InternalReaderService_RetrieveTopics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RetrieveTopicsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InternalReaderServiceServer).RetrieveTopics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bob.v1.InternalReaderService/RetrieveTopics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InternalReaderServiceServer).RetrieveTopics(ctx, req.(*RetrieveTopicsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _InternalReaderService_VerifyAppVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyAppVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InternalReaderServiceServer).VerifyAppVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bob.v1.InternalReaderService/VerifyAppVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InternalReaderServiceServer).VerifyAppVersion(ctx, req.(*VerifyAppVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _InternalReaderService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bob.v1.InternalReaderService",
	HandlerType: (*InternalReaderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RetrieveTopics",
			Handler:    _InternalReaderService_RetrieveTopics_Handler,
		},
		{
			MethodName: "VerifyAppVersion",
			Handler:    _InternalReaderService_VerifyAppVersion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "bob/v1/internal.proto",
}
