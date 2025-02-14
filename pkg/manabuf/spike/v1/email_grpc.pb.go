// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package spb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// EmailModifierServiceClient is the client API for EmailModifierService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EmailModifierServiceClient interface {
	SendEmail(ctx context.Context, in *SendEmailRequest, opts ...grpc.CallOption) (*SendEmailResponse, error)
}

type emailModifierServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewEmailModifierServiceClient(cc grpc.ClientConnInterface) EmailModifierServiceClient {
	return &emailModifierServiceClient{cc}
}

func (c *emailModifierServiceClient) SendEmail(ctx context.Context, in *SendEmailRequest, opts ...grpc.CallOption) (*SendEmailResponse, error) {
	out := new(SendEmailResponse)
	err := c.cc.Invoke(ctx, "/spike.v1.EmailModifierService/SendEmail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EmailModifierServiceServer is the server API for EmailModifierService service.
// All implementations should embed UnimplementedEmailModifierServiceServer
// for forward compatibility
type EmailModifierServiceServer interface {
	SendEmail(context.Context, *SendEmailRequest) (*SendEmailResponse, error)
}

// UnimplementedEmailModifierServiceServer should be embedded to have forward compatible implementations.
type UnimplementedEmailModifierServiceServer struct {
}

func (UnimplementedEmailModifierServiceServer) SendEmail(context.Context, *SendEmailRequest) (*SendEmailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendEmail not implemented")
}

// UnsafeEmailModifierServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EmailModifierServiceServer will
// result in compilation errors.
type UnsafeEmailModifierServiceServer interface {
	mustEmbedUnimplementedEmailModifierServiceServer()
}

func RegisterEmailModifierServiceServer(s grpc.ServiceRegistrar, srv EmailModifierServiceServer) {
	s.RegisterService(&_EmailModifierService_serviceDesc, srv)
}

func _EmailModifierService_SendEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendEmailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmailModifierServiceServer).SendEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/spike.v1.EmailModifierService/SendEmail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmailModifierServiceServer).SendEmail(ctx, req.(*SendEmailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _EmailModifierService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "spike.v1.EmailModifierService",
	HandlerType: (*EmailModifierServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendEmail",
			Handler:    _EmailModifierService_SendEmail_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "spike/v1/email.proto",
}
