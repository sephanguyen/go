// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package vpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// VirtualLessonReaderServiceClient is the client API for VirtualLessonReaderService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VirtualLessonReaderServiceClient interface {
	GetLiveLessonsByLocations(ctx context.Context, in *GetLiveLessonsByLocationsRequest, opts ...grpc.CallOption) (*GetLiveLessonsByLocationsResponse, error)
	GetLearnersByLessonID(ctx context.Context, in *GetLearnersByLessonIDRequest, opts ...grpc.CallOption) (*GetLearnersByLessonIDResponse, error)
	GetLearnersByLessonIDs(ctx context.Context, in *GetLearnersByLessonIDsRequest, opts ...grpc.CallOption) (*GetLearnersByLessonIDsResponse, error)
	GetLessons(ctx context.Context, in *GetLessonsRequest, opts ...grpc.CallOption) (*GetLessonsResponse, error)
	GetClassDoURL(ctx context.Context, in *GetClassDoURLRequest, opts ...grpc.CallOption) (*GetClassDoURLResponse, error)
}

type virtualLessonReaderServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewVirtualLessonReaderServiceClient(cc grpc.ClientConnInterface) VirtualLessonReaderServiceClient {
	return &virtualLessonReaderServiceClient{cc}
}

func (c *virtualLessonReaderServiceClient) GetLiveLessonsByLocations(ctx context.Context, in *GetLiveLessonsByLocationsRequest, opts ...grpc.CallOption) (*GetLiveLessonsByLocationsResponse, error) {
	out := new(GetLiveLessonsByLocationsResponse)
	err := c.cc.Invoke(ctx, "/virtualclassroom.v1.VirtualLessonReaderService/GetLiveLessonsByLocations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *virtualLessonReaderServiceClient) GetLearnersByLessonID(ctx context.Context, in *GetLearnersByLessonIDRequest, opts ...grpc.CallOption) (*GetLearnersByLessonIDResponse, error) {
	out := new(GetLearnersByLessonIDResponse)
	err := c.cc.Invoke(ctx, "/virtualclassroom.v1.VirtualLessonReaderService/GetLearnersByLessonID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *virtualLessonReaderServiceClient) GetLearnersByLessonIDs(ctx context.Context, in *GetLearnersByLessonIDsRequest, opts ...grpc.CallOption) (*GetLearnersByLessonIDsResponse, error) {
	out := new(GetLearnersByLessonIDsResponse)
	err := c.cc.Invoke(ctx, "/virtualclassroom.v1.VirtualLessonReaderService/GetLearnersByLessonIDs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *virtualLessonReaderServiceClient) GetLessons(ctx context.Context, in *GetLessonsRequest, opts ...grpc.CallOption) (*GetLessonsResponse, error) {
	out := new(GetLessonsResponse)
	err := c.cc.Invoke(ctx, "/virtualclassroom.v1.VirtualLessonReaderService/GetLessons", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *virtualLessonReaderServiceClient) GetClassDoURL(ctx context.Context, in *GetClassDoURLRequest, opts ...grpc.CallOption) (*GetClassDoURLResponse, error) {
	out := new(GetClassDoURLResponse)
	err := c.cc.Invoke(ctx, "/virtualclassroom.v1.VirtualLessonReaderService/GetClassDoURL", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VirtualLessonReaderServiceServer is the server API for VirtualLessonReaderService service.
// All implementations should embed UnimplementedVirtualLessonReaderServiceServer
// for forward compatibility
type VirtualLessonReaderServiceServer interface {
	GetLiveLessonsByLocations(context.Context, *GetLiveLessonsByLocationsRequest) (*GetLiveLessonsByLocationsResponse, error)
	GetLearnersByLessonID(context.Context, *GetLearnersByLessonIDRequest) (*GetLearnersByLessonIDResponse, error)
	GetLearnersByLessonIDs(context.Context, *GetLearnersByLessonIDsRequest) (*GetLearnersByLessonIDsResponse, error)
	GetLessons(context.Context, *GetLessonsRequest) (*GetLessonsResponse, error)
	GetClassDoURL(context.Context, *GetClassDoURLRequest) (*GetClassDoURLResponse, error)
}

// UnimplementedVirtualLessonReaderServiceServer should be embedded to have forward compatible implementations.
type UnimplementedVirtualLessonReaderServiceServer struct {
}

func (UnimplementedVirtualLessonReaderServiceServer) GetLiveLessonsByLocations(context.Context, *GetLiveLessonsByLocationsRequest) (*GetLiveLessonsByLocationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLiveLessonsByLocations not implemented")
}
func (UnimplementedVirtualLessonReaderServiceServer) GetLearnersByLessonID(context.Context, *GetLearnersByLessonIDRequest) (*GetLearnersByLessonIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLearnersByLessonID not implemented")
}
func (UnimplementedVirtualLessonReaderServiceServer) GetLearnersByLessonIDs(context.Context, *GetLearnersByLessonIDsRequest) (*GetLearnersByLessonIDsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLearnersByLessonIDs not implemented")
}
func (UnimplementedVirtualLessonReaderServiceServer) GetLessons(context.Context, *GetLessonsRequest) (*GetLessonsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLessons not implemented")
}
func (UnimplementedVirtualLessonReaderServiceServer) GetClassDoURL(context.Context, *GetClassDoURLRequest) (*GetClassDoURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClassDoURL not implemented")
}

// UnsafeVirtualLessonReaderServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VirtualLessonReaderServiceServer will
// result in compilation errors.
type UnsafeVirtualLessonReaderServiceServer interface {
	mustEmbedUnimplementedVirtualLessonReaderServiceServer()
}

func RegisterVirtualLessonReaderServiceServer(s grpc.ServiceRegistrar, srv VirtualLessonReaderServiceServer) {
	s.RegisterService(&_VirtualLessonReaderService_serviceDesc, srv)
}

func _VirtualLessonReaderService_GetLiveLessonsByLocations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLiveLessonsByLocationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VirtualLessonReaderServiceServer).GetLiveLessonsByLocations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/virtualclassroom.v1.VirtualLessonReaderService/GetLiveLessonsByLocations",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VirtualLessonReaderServiceServer).GetLiveLessonsByLocations(ctx, req.(*GetLiveLessonsByLocationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VirtualLessonReaderService_GetLearnersByLessonID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLearnersByLessonIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VirtualLessonReaderServiceServer).GetLearnersByLessonID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/virtualclassroom.v1.VirtualLessonReaderService/GetLearnersByLessonID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VirtualLessonReaderServiceServer).GetLearnersByLessonID(ctx, req.(*GetLearnersByLessonIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VirtualLessonReaderService_GetLearnersByLessonIDs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLearnersByLessonIDsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VirtualLessonReaderServiceServer).GetLearnersByLessonIDs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/virtualclassroom.v1.VirtualLessonReaderService/GetLearnersByLessonIDs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VirtualLessonReaderServiceServer).GetLearnersByLessonIDs(ctx, req.(*GetLearnersByLessonIDsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VirtualLessonReaderService_GetLessons_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLessonsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VirtualLessonReaderServiceServer).GetLessons(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/virtualclassroom.v1.VirtualLessonReaderService/GetLessons",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VirtualLessonReaderServiceServer).GetLessons(ctx, req.(*GetLessonsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VirtualLessonReaderService_GetClassDoURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetClassDoURLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VirtualLessonReaderServiceServer).GetClassDoURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/virtualclassroom.v1.VirtualLessonReaderService/GetClassDoURL",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VirtualLessonReaderServiceServer).GetClassDoURL(ctx, req.(*GetClassDoURLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _VirtualLessonReaderService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "virtualclassroom.v1.VirtualLessonReaderService",
	HandlerType: (*VirtualLessonReaderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetLiveLessonsByLocations",
			Handler:    _VirtualLessonReaderService_GetLiveLessonsByLocations_Handler,
		},
		{
			MethodName: "GetLearnersByLessonID",
			Handler:    _VirtualLessonReaderService_GetLearnersByLessonID_Handler,
		},
		{
			MethodName: "GetLearnersByLessonIDs",
			Handler:    _VirtualLessonReaderService_GetLearnersByLessonIDs_Handler,
		},
		{
			MethodName: "GetLessons",
			Handler:    _VirtualLessonReaderService_GetLessons_Handler,
		},
		{
			MethodName: "GetClassDoURL",
			Handler:    _VirtualLessonReaderService_GetClassDoURL_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "virtualclassroom/v1/lesson.proto",
}
