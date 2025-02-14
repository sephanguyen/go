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

// SchoolServiceClient is the client API for SchoolService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SchoolServiceClient interface {
	MergeSchools(ctx context.Context, in *MergeSchoolsRequest, opts ...grpc.CallOption) (*MergeSchoolsResponse, error)
	UpdateSchool(ctx context.Context, in *UpdateSchoolRequest, opts ...grpc.CallOption) (*UpdateSchoolResponse, error)
	RemoveTeacherFromSchool(ctx context.Context, in *RemoveTeacherFromSchoolRequest, opts ...grpc.CallOption) (*RemoveTeacherFromSchoolResponse, error)
	AddTeacher(ctx context.Context, in *AddTeacherRequest, opts ...grpc.CallOption) (*AddTeacherResponse, error)
	CreateSchoolConfig(ctx context.Context, in *CreateSchoolConfigRequest, opts ...grpc.CallOption) (*CreateSchoolConfigResponse, error)
	UpdateSchoolConfig(ctx context.Context, in *UpdateSchoolConfigRequest, opts ...grpc.CallOption) (*UpdateSchoolConfigResponse, error)
}

type schoolServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSchoolServiceClient(cc grpc.ClientConnInterface) SchoolServiceClient {
	return &schoolServiceClient{cc}
}

func (c *schoolServiceClient) MergeSchools(ctx context.Context, in *MergeSchoolsRequest, opts ...grpc.CallOption) (*MergeSchoolsResponse, error) {
	out := new(MergeSchoolsResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.SchoolService/MergeSchools", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schoolServiceClient) UpdateSchool(ctx context.Context, in *UpdateSchoolRequest, opts ...grpc.CallOption) (*UpdateSchoolResponse, error) {
	out := new(UpdateSchoolResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.SchoolService/UpdateSchool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schoolServiceClient) RemoveTeacherFromSchool(ctx context.Context, in *RemoveTeacherFromSchoolRequest, opts ...grpc.CallOption) (*RemoveTeacherFromSchoolResponse, error) {
	out := new(RemoveTeacherFromSchoolResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.SchoolService/RemoveTeacherFromSchool", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schoolServiceClient) AddTeacher(ctx context.Context, in *AddTeacherRequest, opts ...grpc.CallOption) (*AddTeacherResponse, error) {
	out := new(AddTeacherResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.SchoolService/AddTeacher", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schoolServiceClient) CreateSchoolConfig(ctx context.Context, in *CreateSchoolConfigRequest, opts ...grpc.CallOption) (*CreateSchoolConfigResponse, error) {
	out := new(CreateSchoolConfigResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.SchoolService/CreateSchoolConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schoolServiceClient) UpdateSchoolConfig(ctx context.Context, in *UpdateSchoolConfigRequest, opts ...grpc.CallOption) (*UpdateSchoolConfigResponse, error) {
	out := new(UpdateSchoolConfigResponse)
	err := c.cc.Invoke(ctx, "/yasuo.v1.SchoolService/UpdateSchoolConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SchoolServiceServer is the server API for SchoolService service.
// All implementations should embed UnimplementedSchoolServiceServer
// for forward compatibility
type SchoolServiceServer interface {
	MergeSchools(context.Context, *MergeSchoolsRequest) (*MergeSchoolsResponse, error)
	UpdateSchool(context.Context, *UpdateSchoolRequest) (*UpdateSchoolResponse, error)
	RemoveTeacherFromSchool(context.Context, *RemoveTeacherFromSchoolRequest) (*RemoveTeacherFromSchoolResponse, error)
	AddTeacher(context.Context, *AddTeacherRequest) (*AddTeacherResponse, error)
	CreateSchoolConfig(context.Context, *CreateSchoolConfigRequest) (*CreateSchoolConfigResponse, error)
	UpdateSchoolConfig(context.Context, *UpdateSchoolConfigRequest) (*UpdateSchoolConfigResponse, error)
}

// UnimplementedSchoolServiceServer should be embedded to have forward compatible implementations.
type UnimplementedSchoolServiceServer struct {
}

func (UnimplementedSchoolServiceServer) MergeSchools(context.Context, *MergeSchoolsRequest) (*MergeSchoolsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MergeSchools not implemented")
}
func (UnimplementedSchoolServiceServer) UpdateSchool(context.Context, *UpdateSchoolRequest) (*UpdateSchoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSchool not implemented")
}
func (UnimplementedSchoolServiceServer) RemoveTeacherFromSchool(context.Context, *RemoveTeacherFromSchoolRequest) (*RemoveTeacherFromSchoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveTeacherFromSchool not implemented")
}
func (UnimplementedSchoolServiceServer) AddTeacher(context.Context, *AddTeacherRequest) (*AddTeacherResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddTeacher not implemented")
}
func (UnimplementedSchoolServiceServer) CreateSchoolConfig(context.Context, *CreateSchoolConfigRequest) (*CreateSchoolConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSchoolConfig not implemented")
}
func (UnimplementedSchoolServiceServer) UpdateSchoolConfig(context.Context, *UpdateSchoolConfigRequest) (*UpdateSchoolConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSchoolConfig not implemented")
}

// UnsafeSchoolServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SchoolServiceServer will
// result in compilation errors.
type UnsafeSchoolServiceServer interface {
	mustEmbedUnimplementedSchoolServiceServer()
}

func RegisterSchoolServiceServer(s grpc.ServiceRegistrar, srv SchoolServiceServer) {
	s.RegisterService(&_SchoolService_serviceDesc, srv)
}

func _SchoolService_MergeSchools_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MergeSchoolsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchoolServiceServer).MergeSchools(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.SchoolService/MergeSchools",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchoolServiceServer).MergeSchools(ctx, req.(*MergeSchoolsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchoolService_UpdateSchool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSchoolRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchoolServiceServer).UpdateSchool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.SchoolService/UpdateSchool",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchoolServiceServer).UpdateSchool(ctx, req.(*UpdateSchoolRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchoolService_RemoveTeacherFromSchool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveTeacherFromSchoolRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchoolServiceServer).RemoveTeacherFromSchool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.SchoolService/RemoveTeacherFromSchool",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchoolServiceServer).RemoveTeacherFromSchool(ctx, req.(*RemoveTeacherFromSchoolRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchoolService_AddTeacher_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddTeacherRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchoolServiceServer).AddTeacher(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.SchoolService/AddTeacher",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchoolServiceServer).AddTeacher(ctx, req.(*AddTeacherRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchoolService_CreateSchoolConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSchoolConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchoolServiceServer).CreateSchoolConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.SchoolService/CreateSchoolConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchoolServiceServer).CreateSchoolConfig(ctx, req.(*CreateSchoolConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchoolService_UpdateSchoolConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSchoolConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchoolServiceServer).UpdateSchoolConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/yasuo.v1.SchoolService/UpdateSchoolConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchoolServiceServer).UpdateSchoolConfig(ctx, req.(*UpdateSchoolConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SchoolService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "yasuo.v1.SchoolService",
	HandlerType: (*SchoolServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MergeSchools",
			Handler:    _SchoolService_MergeSchools_Handler,
		},
		{
			MethodName: "UpdateSchool",
			Handler:    _SchoolService_UpdateSchool_Handler,
		},
		{
			MethodName: "RemoveTeacherFromSchool",
			Handler:    _SchoolService_RemoveTeacherFromSchool_Handler,
		},
		{
			MethodName: "AddTeacher",
			Handler:    _SchoolService_AddTeacher_Handler,
		},
		{
			MethodName: "CreateSchoolConfig",
			Handler:    _SchoolService_CreateSchoolConfig_Handler,
		},
		{
			MethodName: "UpdateSchoolConfig",
			Handler:    _SchoolService_UpdateSchoolConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "yasuo/v1/schools.proto",
}
