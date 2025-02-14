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

// ImportMasterDataServiceClient is the client API for ImportMasterDataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ImportMasterDataServiceClient interface {
	ImportInvoiceSchedule(ctx context.Context, in *ImportInvoiceScheduleRequest, opts ...grpc.CallOption) (*ImportInvoiceScheduleResponse, error)
	ImportPartnerBank(ctx context.Context, in *ImportPartnerBankRequest, opts ...grpc.CallOption) (*ImportPartnerBankResponse, error)
}

type importMasterDataServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewImportMasterDataServiceClient(cc grpc.ClientConnInterface) ImportMasterDataServiceClient {
	return &importMasterDataServiceClient{cc}
}

func (c *importMasterDataServiceClient) ImportInvoiceSchedule(ctx context.Context, in *ImportInvoiceScheduleRequest, opts ...grpc.CallOption) (*ImportInvoiceScheduleResponse, error) {
	out := new(ImportInvoiceScheduleResponse)
	err := c.cc.Invoke(ctx, "/invoicemgmt.v1.ImportMasterDataService/ImportInvoiceSchedule", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *importMasterDataServiceClient) ImportPartnerBank(ctx context.Context, in *ImportPartnerBankRequest, opts ...grpc.CallOption) (*ImportPartnerBankResponse, error) {
	out := new(ImportPartnerBankResponse)
	err := c.cc.Invoke(ctx, "/invoicemgmt.v1.ImportMasterDataService/ImportPartnerBank", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ImportMasterDataServiceServer is the server API for ImportMasterDataService service.
// All implementations should embed UnimplementedImportMasterDataServiceServer
// for forward compatibility
type ImportMasterDataServiceServer interface {
	ImportInvoiceSchedule(context.Context, *ImportInvoiceScheduleRequest) (*ImportInvoiceScheduleResponse, error)
	ImportPartnerBank(context.Context, *ImportPartnerBankRequest) (*ImportPartnerBankResponse, error)
}

// UnimplementedImportMasterDataServiceServer should be embedded to have forward compatible implementations.
type UnimplementedImportMasterDataServiceServer struct {
}

func (UnimplementedImportMasterDataServiceServer) ImportInvoiceSchedule(context.Context, *ImportInvoiceScheduleRequest) (*ImportInvoiceScheduleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportInvoiceSchedule not implemented")
}
func (UnimplementedImportMasterDataServiceServer) ImportPartnerBank(context.Context, *ImportPartnerBankRequest) (*ImportPartnerBankResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportPartnerBank not implemented")
}

// UnsafeImportMasterDataServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ImportMasterDataServiceServer will
// result in compilation errors.
type UnsafeImportMasterDataServiceServer interface {
	mustEmbedUnimplementedImportMasterDataServiceServer()
}

func RegisterImportMasterDataServiceServer(s grpc.ServiceRegistrar, srv ImportMasterDataServiceServer) {
	s.RegisterService(&_ImportMasterDataService_serviceDesc, srv)
}

func _ImportMasterDataService_ImportInvoiceSchedule_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImportInvoiceScheduleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ImportMasterDataServiceServer).ImportInvoiceSchedule(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/invoicemgmt.v1.ImportMasterDataService/ImportInvoiceSchedule",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ImportMasterDataServiceServer).ImportInvoiceSchedule(ctx, req.(*ImportInvoiceScheduleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ImportMasterDataService_ImportPartnerBank_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImportPartnerBankRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ImportMasterDataServiceServer).ImportPartnerBank(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/invoicemgmt.v1.ImportMasterDataService/ImportPartnerBank",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ImportMasterDataServiceServer).ImportPartnerBank(ctx, req.(*ImportPartnerBankRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ImportMasterDataService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "invoicemgmt.v1.ImportMasterDataService",
	HandlerType: (*ImportMasterDataServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ImportInvoiceSchedule",
			Handler:    _ImportMasterDataService_ImportInvoiceSchedule_Handler,
		},
		{
			MethodName: "ImportPartnerBank",
			Handler:    _ImportMasterDataService_ImportPartnerBank_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "invoicemgmt/v1/import.proto",
}
