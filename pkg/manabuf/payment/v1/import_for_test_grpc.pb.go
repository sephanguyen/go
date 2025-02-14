// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pmpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ImportMasterDataForTestServiceClient is the client API for ImportMasterDataForTestService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ImportMasterDataForTestServiceClient interface {
	ImportAllForTest(ctx context.Context, in *ImportAllForTestRequest, opts ...grpc.CallOption) (*ImportAllForTestResponse, error)
}

type importMasterDataForTestServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewImportMasterDataForTestServiceClient(cc grpc.ClientConnInterface) ImportMasterDataForTestServiceClient {
	return &importMasterDataForTestServiceClient{cc}
}

func (c *importMasterDataForTestServiceClient) ImportAllForTest(ctx context.Context, in *ImportAllForTestRequest, opts ...grpc.CallOption) (*ImportAllForTestResponse, error) {
	out := new(ImportAllForTestResponse)
	err := c.cc.Invoke(ctx, "/payment.v1.ImportMasterDataForTestService/ImportAllForTest", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ImportMasterDataForTestServiceServer is the server API for ImportMasterDataForTestService service.
// All implementations should embed UnimplementedImportMasterDataForTestServiceServer
// for forward compatibility
type ImportMasterDataForTestServiceServer interface {
	ImportAllForTest(context.Context, *ImportAllForTestRequest) (*ImportAllForTestResponse, error)
}

// UnimplementedImportMasterDataForTestServiceServer should be embedded to have forward compatible implementations.
type UnimplementedImportMasterDataForTestServiceServer struct {
}

func (UnimplementedImportMasterDataForTestServiceServer) ImportAllForTest(context.Context, *ImportAllForTestRequest) (*ImportAllForTestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportAllForTest not implemented")
}

// UnsafeImportMasterDataForTestServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ImportMasterDataForTestServiceServer will
// result in compilation errors.
type UnsafeImportMasterDataForTestServiceServer interface {
	mustEmbedUnimplementedImportMasterDataForTestServiceServer()
}

func RegisterImportMasterDataForTestServiceServer(s grpc.ServiceRegistrar, srv ImportMasterDataForTestServiceServer) {
	s.RegisterService(&_ImportMasterDataForTestService_serviceDesc, srv)
}

func _ImportMasterDataForTestService_ImportAllForTest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImportAllForTestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ImportMasterDataForTestServiceServer).ImportAllForTest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.v1.ImportMasterDataForTestService/ImportAllForTest",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ImportMasterDataForTestServiceServer).ImportAllForTest(ctx, req.(*ImportAllForTestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ImportMasterDataForTestService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "payment.v1.ImportMasterDataForTestService",
	HandlerType: (*ImportMasterDataForTestServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ImportAllForTest",
			Handler:    _ImportMasterDataForTestService_ImportAllForTest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment/v1/import_for_test.proto",
}
