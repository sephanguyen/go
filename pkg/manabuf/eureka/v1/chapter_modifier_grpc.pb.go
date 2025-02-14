// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package epb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ChapterModifierServiceClient is the client API for ChapterModifierService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChapterModifierServiceClient interface {
	UpsertChapters(ctx context.Context, in *UpsertChaptersRequest, opts ...grpc.CallOption) (*UpsertChaptersResponse, error)
	DeleteChapters(ctx context.Context, in *DeleteChaptersRequest, opts ...grpc.CallOption) (*DeleteChaptersResponse, error)
}

type chapterModifierServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewChapterModifierServiceClient(cc grpc.ClientConnInterface) ChapterModifierServiceClient {
	return &chapterModifierServiceClient{cc}
}

func (c *chapterModifierServiceClient) UpsertChapters(ctx context.Context, in *UpsertChaptersRequest, opts ...grpc.CallOption) (*UpsertChaptersResponse, error) {
	out := new(UpsertChaptersResponse)
	err := c.cc.Invoke(ctx, "/eureka.v1.ChapterModifierService/UpsertChapters", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chapterModifierServiceClient) DeleteChapters(ctx context.Context, in *DeleteChaptersRequest, opts ...grpc.CallOption) (*DeleteChaptersResponse, error) {
	out := new(DeleteChaptersResponse)
	err := c.cc.Invoke(ctx, "/eureka.v1.ChapterModifierService/DeleteChapters", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChapterModifierServiceServer is the server API for ChapterModifierService service.
// All implementations should embed UnimplementedChapterModifierServiceServer
// for forward compatibility
type ChapterModifierServiceServer interface {
	UpsertChapters(context.Context, *UpsertChaptersRequest) (*UpsertChaptersResponse, error)
	DeleteChapters(context.Context, *DeleteChaptersRequest) (*DeleteChaptersResponse, error)
}

// UnimplementedChapterModifierServiceServer should be embedded to have forward compatible implementations.
type UnimplementedChapterModifierServiceServer struct {
}

func (UnimplementedChapterModifierServiceServer) UpsertChapters(context.Context, *UpsertChaptersRequest) (*UpsertChaptersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertChapters not implemented")
}
func (UnimplementedChapterModifierServiceServer) DeleteChapters(context.Context, *DeleteChaptersRequest) (*DeleteChaptersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteChapters not implemented")
}

// UnsafeChapterModifierServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChapterModifierServiceServer will
// result in compilation errors.
type UnsafeChapterModifierServiceServer interface {
	mustEmbedUnimplementedChapterModifierServiceServer()
}

func RegisterChapterModifierServiceServer(s grpc.ServiceRegistrar, srv ChapterModifierServiceServer) {
	s.RegisterService(&_ChapterModifierService_serviceDesc, srv)
}

func _ChapterModifierService_UpsertChapters_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpsertChaptersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChapterModifierServiceServer).UpsertChapters(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/eureka.v1.ChapterModifierService/UpsertChapters",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChapterModifierServiceServer).UpsertChapters(ctx, req.(*UpsertChaptersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChapterModifierService_DeleteChapters_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteChaptersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChapterModifierServiceServer).DeleteChapters(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/eureka.v1.ChapterModifierService/DeleteChapters",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChapterModifierServiceServer).DeleteChapters(ctx, req.(*DeleteChaptersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ChapterModifierService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "eureka.v1.ChapterModifierService",
	HandlerType: (*ChapterModifierServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpsertChapters",
			Handler:    _ChapterModifierService_UpsertChapters_Handler,
		},
		{
			MethodName: "DeleteChapters",
			Handler:    _ChapterModifierService_DeleteChapters_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "eureka/v1/chapter_modifier.proto",
}
