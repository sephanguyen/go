// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package sspb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// LearningObjectiveClient is the client API for LearningObjective service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LearningObjectiveClient interface {
	// InsertLearningObjective insert a Learning Objective
	InsertLearningObjective(ctx context.Context, in *InsertLearningObjectiveRequest, opts ...grpc.CallOption) (*InsertLearningObjectiveResponse, error)
	// UpdateLearningObjective update a Learning Objective
	UpdateLearningObjective(ctx context.Context, in *UpdateLearningObjectiveRequest, opts ...grpc.CallOption) (*UpdateLearningObjectiveResponse, error)
	ListLearningObjective(ctx context.Context, in *ListLearningObjectiveRequest, opts ...grpc.CallOption) (*ListLearningObjectiveResponse, error)
	UpsertLOProgression(ctx context.Context, in *UpsertLOProgressionRequest, opts ...grpc.CallOption) (*UpsertLOProgressionResponse, error)
	RetrieveLOProgression(ctx context.Context, in *RetrieveLOProgressionRequest, opts ...grpc.CallOption) (*RetrieveLOProgressionResponse, error)
}

type learningObjectiveClient struct {
	cc grpc.ClientConnInterface
}

func NewLearningObjectiveClient(cc grpc.ClientConnInterface) LearningObjectiveClient {
	return &learningObjectiveClient{cc}
}

func (c *learningObjectiveClient) InsertLearningObjective(ctx context.Context, in *InsertLearningObjectiveRequest, opts ...grpc.CallOption) (*InsertLearningObjectiveResponse, error) {
	out := new(InsertLearningObjectiveResponse)
	err := c.cc.Invoke(ctx, "/syllabus.v1.LearningObjective/InsertLearningObjective", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *learningObjectiveClient) UpdateLearningObjective(ctx context.Context, in *UpdateLearningObjectiveRequest, opts ...grpc.CallOption) (*UpdateLearningObjectiveResponse, error) {
	out := new(UpdateLearningObjectiveResponse)
	err := c.cc.Invoke(ctx, "/syllabus.v1.LearningObjective/UpdateLearningObjective", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *learningObjectiveClient) ListLearningObjective(ctx context.Context, in *ListLearningObjectiveRequest, opts ...grpc.CallOption) (*ListLearningObjectiveResponse, error) {
	out := new(ListLearningObjectiveResponse)
	err := c.cc.Invoke(ctx, "/syllabus.v1.LearningObjective/ListLearningObjective", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *learningObjectiveClient) UpsertLOProgression(ctx context.Context, in *UpsertLOProgressionRequest, opts ...grpc.CallOption) (*UpsertLOProgressionResponse, error) {
	out := new(UpsertLOProgressionResponse)
	err := c.cc.Invoke(ctx, "/syllabus.v1.LearningObjective/UpsertLOProgression", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *learningObjectiveClient) RetrieveLOProgression(ctx context.Context, in *RetrieveLOProgressionRequest, opts ...grpc.CallOption) (*RetrieveLOProgressionResponse, error) {
	out := new(RetrieveLOProgressionResponse)
	err := c.cc.Invoke(ctx, "/syllabus.v1.LearningObjective/RetrieveLOProgression", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LearningObjectiveServer is the server API for LearningObjective service.
// All implementations should embed UnimplementedLearningObjectiveServer
// for forward compatibility
type LearningObjectiveServer interface {
	// InsertLearningObjective insert a Learning Objective
	InsertLearningObjective(context.Context, *InsertLearningObjectiveRequest) (*InsertLearningObjectiveResponse, error)
	// UpdateLearningObjective update a Learning Objective
	UpdateLearningObjective(context.Context, *UpdateLearningObjectiveRequest) (*UpdateLearningObjectiveResponse, error)
	ListLearningObjective(context.Context, *ListLearningObjectiveRequest) (*ListLearningObjectiveResponse, error)
	UpsertLOProgression(context.Context, *UpsertLOProgressionRequest) (*UpsertLOProgressionResponse, error)
	RetrieveLOProgression(context.Context, *RetrieveLOProgressionRequest) (*RetrieveLOProgressionResponse, error)
}

// UnimplementedLearningObjectiveServer should be embedded to have forward compatible implementations.
type UnimplementedLearningObjectiveServer struct {
}

func (UnimplementedLearningObjectiveServer) InsertLearningObjective(context.Context, *InsertLearningObjectiveRequest) (*InsertLearningObjectiveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InsertLearningObjective not implemented")
}
func (UnimplementedLearningObjectiveServer) UpdateLearningObjective(context.Context, *UpdateLearningObjectiveRequest) (*UpdateLearningObjectiveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateLearningObjective not implemented")
}
func (UnimplementedLearningObjectiveServer) ListLearningObjective(context.Context, *ListLearningObjectiveRequest) (*ListLearningObjectiveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListLearningObjective not implemented")
}
func (UnimplementedLearningObjectiveServer) UpsertLOProgression(context.Context, *UpsertLOProgressionRequest) (*UpsertLOProgressionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertLOProgression not implemented")
}
func (UnimplementedLearningObjectiveServer) RetrieveLOProgression(context.Context, *RetrieveLOProgressionRequest) (*RetrieveLOProgressionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RetrieveLOProgression not implemented")
}

// UnsafeLearningObjectiveServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LearningObjectiveServer will
// result in compilation errors.
type UnsafeLearningObjectiveServer interface {
	mustEmbedUnimplementedLearningObjectiveServer()
}

func RegisterLearningObjectiveServer(s grpc.ServiceRegistrar, srv LearningObjectiveServer) {
	s.RegisterService(&_LearningObjective_serviceDesc, srv)
}

func _LearningObjective_InsertLearningObjective_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertLearningObjectiveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LearningObjectiveServer).InsertLearningObjective(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/syllabus.v1.LearningObjective/InsertLearningObjective",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LearningObjectiveServer).InsertLearningObjective(ctx, req.(*InsertLearningObjectiveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LearningObjective_UpdateLearningObjective_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateLearningObjectiveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LearningObjectiveServer).UpdateLearningObjective(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/syllabus.v1.LearningObjective/UpdateLearningObjective",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LearningObjectiveServer).UpdateLearningObjective(ctx, req.(*UpdateLearningObjectiveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LearningObjective_ListLearningObjective_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListLearningObjectiveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LearningObjectiveServer).ListLearningObjective(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/syllabus.v1.LearningObjective/ListLearningObjective",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LearningObjectiveServer).ListLearningObjective(ctx, req.(*ListLearningObjectiveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LearningObjective_UpsertLOProgression_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpsertLOProgressionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LearningObjectiveServer).UpsertLOProgression(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/syllabus.v1.LearningObjective/UpsertLOProgression",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LearningObjectiveServer).UpsertLOProgression(ctx, req.(*UpsertLOProgressionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LearningObjective_RetrieveLOProgression_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RetrieveLOProgressionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LearningObjectiveServer).RetrieveLOProgression(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/syllabus.v1.LearningObjective/RetrieveLOProgression",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LearningObjectiveServer).RetrieveLOProgression(ctx, req.(*RetrieveLOProgressionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _LearningObjective_serviceDesc = grpc.ServiceDesc{
	ServiceName: "syllabus.v1.LearningObjective",
	HandlerType: (*LearningObjectiveServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InsertLearningObjective",
			Handler:    _LearningObjective_InsertLearningObjective_Handler,
		},
		{
			MethodName: "UpdateLearningObjective",
			Handler:    _LearningObjective_UpdateLearningObjective_Handler,
		},
		{
			MethodName: "ListLearningObjective",
			Handler:    _LearningObjective_ListLearningObjective_Handler,
		},
		{
			MethodName: "UpsertLOProgression",
			Handler:    _LearningObjective_UpsertLOProgression_Handler,
		},
		{
			MethodName: "RetrieveLOProgression",
			Handler:    _LearningObjective_RetrieveLOProgression_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "syllabus/v1/learning_objective_service.proto",
}
