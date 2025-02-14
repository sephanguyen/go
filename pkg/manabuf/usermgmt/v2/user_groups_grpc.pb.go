// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package upb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// UserGroupMgmtServiceClient is the client API for UserGroupMgmtService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserGroupMgmtServiceClient interface {
	CreateUserGroup(ctx context.Context, in *CreateUserGroupRequest, opts ...grpc.CallOption) (*CreateUserGroupResponse, error)
	UpdateUserGroup(ctx context.Context, in *UpdateUserGroupRequest, opts ...grpc.CallOption) (*UpdateUserGroupResponse, error)
	ValidateUserLogin(ctx context.Context, in *ValidateUserLoginRequest, opts ...grpc.CallOption) (*ValidateUserLoginResponse, error)
}

type userGroupMgmtServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserGroupMgmtServiceClient(cc grpc.ClientConnInterface) UserGroupMgmtServiceClient {
	return &userGroupMgmtServiceClient{cc}
}

func (c *userGroupMgmtServiceClient) CreateUserGroup(ctx context.Context, in *CreateUserGroupRequest, opts ...grpc.CallOption) (*CreateUserGroupResponse, error) {
	out := new(CreateUserGroupResponse)
	err := c.cc.Invoke(ctx, "/usermgmt.v2.UserGroupMgmtService/CreateUserGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userGroupMgmtServiceClient) UpdateUserGroup(ctx context.Context, in *UpdateUserGroupRequest, opts ...grpc.CallOption) (*UpdateUserGroupResponse, error) {
	out := new(UpdateUserGroupResponse)
	err := c.cc.Invoke(ctx, "/usermgmt.v2.UserGroupMgmtService/UpdateUserGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userGroupMgmtServiceClient) ValidateUserLogin(ctx context.Context, in *ValidateUserLoginRequest, opts ...grpc.CallOption) (*ValidateUserLoginResponse, error) {
	out := new(ValidateUserLoginResponse)
	err := c.cc.Invoke(ctx, "/usermgmt.v2.UserGroupMgmtService/ValidateUserLogin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserGroupMgmtServiceServer is the server API for UserGroupMgmtService service.
// All implementations should embed UnimplementedUserGroupMgmtServiceServer
// for forward compatibility
type UserGroupMgmtServiceServer interface {
	CreateUserGroup(context.Context, *CreateUserGroupRequest) (*CreateUserGroupResponse, error)
	UpdateUserGroup(context.Context, *UpdateUserGroupRequest) (*UpdateUserGroupResponse, error)
	ValidateUserLogin(context.Context, *ValidateUserLoginRequest) (*ValidateUserLoginResponse, error)
}

// UnimplementedUserGroupMgmtServiceServer should be embedded to have forward compatible implementations.
type UnimplementedUserGroupMgmtServiceServer struct {
}

func (UnimplementedUserGroupMgmtServiceServer) CreateUserGroup(context.Context, *CreateUserGroupRequest) (*CreateUserGroupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUserGroup not implemented")
}
func (UnimplementedUserGroupMgmtServiceServer) UpdateUserGroup(context.Context, *UpdateUserGroupRequest) (*UpdateUserGroupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUserGroup not implemented")
}
func (UnimplementedUserGroupMgmtServiceServer) ValidateUserLogin(context.Context, *ValidateUserLoginRequest) (*ValidateUserLoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateUserLogin not implemented")
}

// UnsafeUserGroupMgmtServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserGroupMgmtServiceServer will
// result in compilation errors.
type UnsafeUserGroupMgmtServiceServer interface {
	mustEmbedUnimplementedUserGroupMgmtServiceServer()
}

func RegisterUserGroupMgmtServiceServer(s grpc.ServiceRegistrar, srv UserGroupMgmtServiceServer) {
	s.RegisterService(&_UserGroupMgmtService_serviceDesc, srv)
}

func _UserGroupMgmtService_CreateUserGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateUserGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserGroupMgmtServiceServer).CreateUserGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/usermgmt.v2.UserGroupMgmtService/CreateUserGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserGroupMgmtServiceServer).CreateUserGroup(ctx, req.(*CreateUserGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserGroupMgmtService_UpdateUserGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateUserGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserGroupMgmtServiceServer).UpdateUserGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/usermgmt.v2.UserGroupMgmtService/UpdateUserGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserGroupMgmtServiceServer).UpdateUserGroup(ctx, req.(*UpdateUserGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserGroupMgmtService_ValidateUserLogin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateUserLoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserGroupMgmtServiceServer).ValidateUserLogin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/usermgmt.v2.UserGroupMgmtService/ValidateUserLogin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserGroupMgmtServiceServer).ValidateUserLogin(ctx, req.(*ValidateUserLoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _UserGroupMgmtService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "usermgmt.v2.UserGroupMgmtService",
	HandlerType: (*UserGroupMgmtServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateUserGroup",
			Handler:    _UserGroupMgmtService_CreateUserGroup_Handler,
		},
		{
			MethodName: "UpdateUserGroup",
			Handler:    _UserGroupMgmtService_UpdateUserGroup_Handler,
		},
		{
			MethodName: "ValidateUserLogin",
			Handler:    _UserGroupMgmtService_ValidateUserLogin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "usermgmt/v2/user_groups.proto",
}
