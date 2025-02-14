// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package invoice_pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ExportMasterDataServiceClient is the client API for ExportMasterDataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExportMasterDataServiceClient interface {
	ExportInvoiceSchedule(ctx context.Context, in *ExportInvoiceScheduleRequest, opts ...grpc.CallOption) (*ExportInvoiceScheduleResponse, error)
	ExportBank(ctx context.Context, in *ExportBankRequest, opts ...grpc.CallOption) (*ExportBankResponse, error)
	ExportBankBranch(ctx context.Context, in *ExportBankBranchRequest, opts ...grpc.CallOption) (*ExportBankBranchResponse, error)
	ExportBankMapping(ctx context.Context, in *ExportBankMappingRequest, opts ...grpc.CallOption) (*ExportBankMappingResponse, error)
}

type exportMasterDataServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewExportMasterDataServiceClient(cc grpc.ClientConnInterface) ExportMasterDataServiceClient {
	return &exportMasterDataServiceClient{cc}
}

func (c *exportMasterDataServiceClient) ExportInvoiceSchedule(ctx context.Context, in *ExportInvoiceScheduleRequest, opts ...grpc.CallOption) (*ExportInvoiceScheduleResponse, error) {
	out := new(ExportInvoiceScheduleResponse)
	err := c.cc.Invoke(ctx, "/invoicemgmt.v1.ExportMasterDataService/ExportInvoiceSchedule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *exportMasterDataServiceClient) ExportBank(ctx context.Context, in *ExportBankRequest, opts ...grpc.CallOption) (*ExportBankResponse, error) {
	out := new(ExportBankResponse)
	err := c.cc.Invoke(ctx, "/invoicemgmt.v1.ExportMasterDataService/ExportBank", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *exportMasterDataServiceClient) ExportBankBranch(ctx context.Context, in *ExportBankBranchRequest, opts ...grpc.CallOption) (*ExportBankBranchResponse, error) {
	out := new(ExportBankBranchResponse)
	err := c.cc.Invoke(ctx, "/invoicemgmt.v1.ExportMasterDataService/ExportBankBranch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *exportMasterDataServiceClient) ExportBankMapping(ctx context.Context, in *ExportBankMappingRequest, opts ...grpc.CallOption) (*ExportBankMappingResponse, error) {
	out := new(ExportBankMappingResponse)
	err := c.cc.Invoke(ctx, "/invoicemgmt.v1.ExportMasterDataService/ExportBankMapping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExportMasterDataServiceServer is the server API for ExportMasterDataService service.
// All implementations should embed UnimplementedExportMasterDataServiceServer
// for forward compatibility
type ExportMasterDataServiceServer interface {
	ExportInvoiceSchedule(context.Context, *ExportInvoiceScheduleRequest) (*ExportInvoiceScheduleResponse, error)
	ExportBank(context.Context, *ExportBankRequest) (*ExportBankResponse, error)
	ExportBankBranch(context.Context, *ExportBankBranchRequest) (*ExportBankBranchResponse, error)
	ExportBankMapping(context.Context, *ExportBankMappingRequest) (*ExportBankMappingResponse, error)
}

// UnimplementedExportMasterDataServiceServer should be embedded to have forward compatible implementations.
type UnimplementedExportMasterDataServiceServer struct {
}

func (UnimplementedExportMasterDataServiceServer) ExportInvoiceSchedule(context.Context, *ExportInvoiceScheduleRequest) (*ExportInvoiceScheduleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExportInvoiceSchedule not implemented")
}
func (UnimplementedExportMasterDataServiceServer) ExportBank(context.Context, *ExportBankRequest) (*ExportBankResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExportBank not implemented")
}
func (UnimplementedExportMasterDataServiceServer) ExportBankBranch(context.Context, *ExportBankBranchRequest) (*ExportBankBranchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExportBankBranch not implemented")
}
func (UnimplementedExportMasterDataServiceServer) ExportBankMapping(context.Context, *ExportBankMappingRequest) (*ExportBankMappingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExportBankMapping not implemented")
}

// UnsafeExportMasterDataServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExportMasterDataServiceServer will
// result in compilation errors.
type UnsafeExportMasterDataServiceServer interface {
	mustEmbedUnimplementedExportMasterDataServiceServer()
}

func RegisterExportMasterDataServiceServer(s grpc.ServiceRegistrar, srv ExportMasterDataServiceServer) {
	s.RegisterService(&_ExportMasterDataService_serviceDesc, srv)
}

func _ExportMasterDataService_ExportInvoiceSchedule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExportInvoiceScheduleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExportMasterDataServiceServer).ExportInvoiceSchedule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/invoicemgmt.v1.ExportMasterDataService/ExportInvoiceSchedule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExportMasterDataServiceServer).ExportInvoiceSchedule(ctx, req.(*ExportInvoiceScheduleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExportMasterDataService_ExportBank_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExportBankRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExportMasterDataServiceServer).ExportBank(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/invoicemgmt.v1.ExportMasterDataService/ExportBank",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExportMasterDataServiceServer).ExportBank(ctx, req.(*ExportBankRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExportMasterDataService_ExportBankBranch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExportBankBranchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExportMasterDataServiceServer).ExportBankBranch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/invoicemgmt.v1.ExportMasterDataService/ExportBankBranch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExportMasterDataServiceServer).ExportBankBranch(ctx, req.(*ExportBankBranchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExportMasterDataService_ExportBankMapping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExportBankMappingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExportMasterDataServiceServer).ExportBankMapping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/invoicemgmt.v1.ExportMasterDataService/ExportBankMapping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExportMasterDataServiceServer).ExportBankMapping(ctx, req.(*ExportBankMappingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ExportMasterDataService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "invoicemgmt.v1.ExportMasterDataService",
	HandlerType: (*ExportMasterDataServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ExportInvoiceSchedule",
			Handler:    _ExportMasterDataService_ExportInvoiceSchedule_Handler,
		},
		{
			MethodName: "ExportBank",
			Handler:    _ExportMasterDataService_ExportBank_Handler,
		},
		{
			MethodName: "ExportBankBranch",
			Handler:    _ExportMasterDataService_ExportBankBranch_Handler,
		},
		{
			MethodName: "ExportBankMapping",
			Handler:    _ExportMasterDataService_ExportBankMapping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "invoicemgmt/v1/export.proto",
}
