// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package tpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// StaffTransportationExpenseServiceClient is the client API for StaffTransportationExpenseService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StaffTransportationExpenseServiceClient interface {
	UpsertStaffTransportationExpense(ctx context.Context, in *UpsertStaffTransportationExpenseRequest, opts ...grpc.CallOption) (*UpsertStaffTransportationExpenseResponse, error)
}

type staffTransportationExpenseServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewStaffTransportationExpenseServiceClient(cc grpc.ClientConnInterface) StaffTransportationExpenseServiceClient {
	return &staffTransportationExpenseServiceClient{cc}
}

func (c *staffTransportationExpenseServiceClient) UpsertStaffTransportationExpense(ctx context.Context, in *UpsertStaffTransportationExpenseRequest, opts ...grpc.CallOption) (*UpsertStaffTransportationExpenseResponse, error) {
	out := new(UpsertStaffTransportationExpenseResponse)
	err := c.cc.Invoke(ctx, "/timesheet.v1.StaffTransportationExpenseService/UpsertStaffTransportationExpense", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StaffTransportationExpenseServiceServer is the server API for StaffTransportationExpenseService service.
// All implementations should embed UnimplementedStaffTransportationExpenseServiceServer
// for forward compatibility
type StaffTransportationExpenseServiceServer interface {
	UpsertStaffTransportationExpense(context.Context, *UpsertStaffTransportationExpenseRequest) (*UpsertStaffTransportationExpenseResponse, error)
}

// UnimplementedStaffTransportationExpenseServiceServer should be embedded to have forward compatible implementations.
type UnimplementedStaffTransportationExpenseServiceServer struct {
}

func (UnimplementedStaffTransportationExpenseServiceServer) UpsertStaffTransportationExpense(context.Context, *UpsertStaffTransportationExpenseRequest) (*UpsertStaffTransportationExpenseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertStaffTransportationExpense not implemented")
}

// UnsafeStaffTransportationExpenseServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StaffTransportationExpenseServiceServer will
// result in compilation errors.
type UnsafeStaffTransportationExpenseServiceServer interface {
	mustEmbedUnimplementedStaffTransportationExpenseServiceServer()
}

func RegisterStaffTransportationExpenseServiceServer(s grpc.ServiceRegistrar, srv StaffTransportationExpenseServiceServer) {
	s.RegisterService(&_StaffTransportationExpenseService_serviceDesc, srv)
}

func _StaffTransportationExpenseService_UpsertStaffTransportationExpense_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpsertStaffTransportationExpenseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StaffTransportationExpenseServiceServer).UpsertStaffTransportationExpense(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/timesheet.v1.StaffTransportationExpenseService/UpsertStaffTransportationExpense",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StaffTransportationExpenseServiceServer).UpsertStaffTransportationExpense(ctx, req.(*UpsertStaffTransportationExpenseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _StaffTransportationExpenseService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "timesheet.v1.StaffTransportationExpenseService",
	HandlerType: (*StaffTransportationExpenseServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpsertStaffTransportationExpense",
			Handler:    _StaffTransportationExpenseService_UpsertStaffTransportationExpense_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "timesheet/v1/staff_transportation_expense.proto",
}
