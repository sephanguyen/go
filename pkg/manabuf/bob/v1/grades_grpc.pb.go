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

// GradeManagementServiceClient is the client API for GradeManagementService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GradeManagementServiceClient interface {
	ImportGrade(ctx context.Context, in *ImportGradeRequest, opts ...grpc.CallOption) (*ImportGradeResponse, error)
}

type gradeManagementServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGradeManagementServiceClient(cc grpc.ClientConnInterface) GradeManagementServiceClient {
	return &gradeManagementServiceClient{cc}
}

func (c *gradeManagementServiceClient) ImportGrade(ctx context.Context, in *ImportGradeRequest, opts ...grpc.CallOption) (*ImportGradeResponse, error) {
	out := new(ImportGradeResponse)
	err := c.cc.Invoke(ctx, "/bob.v1.GradeManagementService/ImportGrade", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GradeManagementServiceServer is the server API for GradeManagementService service.
// All implementations should embed UnimplementedGradeManagementServiceServer
// for forward compatibility
type GradeManagementServiceServer interface {
	ImportGrade(context.Context, *ImportGradeRequest) (*ImportGradeResponse, error)
}

// UnimplementedGradeManagementServiceServer should be embedded to have forward compatible implementations.
type UnimplementedGradeManagementServiceServer struct {
}

func (UnimplementedGradeManagementServiceServer) ImportGrade(context.Context, *ImportGradeRequest) (*ImportGradeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportGrade not implemented")
}

// UnsafeGradeManagementServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GradeManagementServiceServer will
// result in compilation errors.
type UnsafeGradeManagementServiceServer interface {
	mustEmbedUnimplementedGradeManagementServiceServer()
}

func RegisterGradeManagementServiceServer(s grpc.ServiceRegistrar, srv GradeManagementServiceServer) {
	s.RegisterService(&_GradeManagementService_serviceDesc, srv)
}

func _GradeManagementService_ImportGrade_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImportGradeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GradeManagementServiceServer).ImportGrade(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bob.v1.GradeManagementService/ImportGrade",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GradeManagementServiceServer).ImportGrade(ctx, req.(*ImportGradeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _GradeManagementService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "bob.v1.GradeManagementService",
	HandlerType: (*GradeManagementServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ImportGrade",
			Handler:    _GradeManagementService_ImportGrade_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "bob/v1/grades.proto",
}
