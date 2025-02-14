// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: notificationmgmt/v1/system_notifications.proto

/*
Package npb is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package npb

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join

func request_SystemNotificationReaderService_RetrieveSystemNotifications_0(ctx context.Context, marshaler runtime.Marshaler, client SystemNotificationReaderServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq RetrieveSystemNotificationsRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.RetrieveSystemNotifications(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_SystemNotificationReaderService_RetrieveSystemNotifications_0(ctx context.Context, marshaler runtime.Marshaler, server SystemNotificationReaderServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq RetrieveSystemNotificationsRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.RetrieveSystemNotifications(ctx, &protoReq)
	return msg, metadata, err

}

func request_SystemNotificationModifierService_SetSystemNotificationStatus_0(ctx context.Context, marshaler runtime.Marshaler, client SystemNotificationModifierServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq SetSystemNotificationStatusRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.SetSystemNotificationStatus(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_SystemNotificationModifierService_SetSystemNotificationStatus_0(ctx context.Context, marshaler runtime.Marshaler, server SystemNotificationModifierServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq SetSystemNotificationStatusRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.SetSystemNotificationStatus(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterSystemNotificationReaderServiceHandlerServer registers the http handlers for service SystemNotificationReaderService to "mux".
// UnaryRPC     :call SystemNotificationReaderServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterSystemNotificationReaderServiceHandlerFromEndpoint instead.
func RegisterSystemNotificationReaderServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server SystemNotificationReaderServiceServer) error {

	mux.Handle("POST", pattern_SystemNotificationReaderService_RetrieveSystemNotifications_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateIncomingContext(ctx, mux, req, "/notificationmgmt.v1.SystemNotificationReaderService/RetrieveSystemNotifications", runtime.WithHTTPPathPattern("/notificationmgmt/api/v1/proxy/notificationmgmts/retrieve_system_notifications"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_SystemNotificationReaderService_RetrieveSystemNotifications_0(annotatedContext, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_SystemNotificationReaderService_RetrieveSystemNotifications_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterSystemNotificationModifierServiceHandlerServer registers the http handlers for service SystemNotificationModifierService to "mux".
// UnaryRPC     :call SystemNotificationModifierServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterSystemNotificationModifierServiceHandlerFromEndpoint instead.
func RegisterSystemNotificationModifierServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server SystemNotificationModifierServiceServer) error {

	mux.Handle("POST", pattern_SystemNotificationModifierService_SetSystemNotificationStatus_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateIncomingContext(ctx, mux, req, "/notificationmgmt.v1.SystemNotificationModifierService/SetSystemNotificationStatus", runtime.WithHTTPPathPattern("/notificationmgmt/api/v1/proxy/notificationmgmts/set_system_notifications_status"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_SystemNotificationModifierService_SetSystemNotificationStatus_0(annotatedContext, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_SystemNotificationModifierService_SetSystemNotificationStatus_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterSystemNotificationReaderServiceHandlerFromEndpoint is same as RegisterSystemNotificationReaderServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterSystemNotificationReaderServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterSystemNotificationReaderServiceHandler(ctx, mux, conn)
}

// RegisterSystemNotificationReaderServiceHandler registers the http handlers for service SystemNotificationReaderService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterSystemNotificationReaderServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterSystemNotificationReaderServiceHandlerClient(ctx, mux, NewSystemNotificationReaderServiceClient(conn))
}

// RegisterSystemNotificationReaderServiceHandlerClient registers the http handlers for service SystemNotificationReaderService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "SystemNotificationReaderServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "SystemNotificationReaderServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "SystemNotificationReaderServiceClient" to call the correct interceptors.
func RegisterSystemNotificationReaderServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client SystemNotificationReaderServiceClient) error {

	mux.Handle("POST", pattern_SystemNotificationReaderService_RetrieveSystemNotifications_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/notificationmgmt.v1.SystemNotificationReaderService/RetrieveSystemNotifications", runtime.WithHTTPPathPattern("/notificationmgmt/api/v1/proxy/notificationmgmts/retrieve_system_notifications"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_SystemNotificationReaderService_RetrieveSystemNotifications_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_SystemNotificationReaderService_RetrieveSystemNotifications_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_SystemNotificationReaderService_RetrieveSystemNotifications_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4, 2, 5}, []string{"notificationmgmt", "api", "v1", "proxy", "notificationmgmts", "retrieve_system_notifications"}, ""))
)

var (
	forward_SystemNotificationReaderService_RetrieveSystemNotifications_0 = runtime.ForwardResponseMessage
)

// RegisterSystemNotificationModifierServiceHandlerFromEndpoint is same as RegisterSystemNotificationModifierServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterSystemNotificationModifierServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterSystemNotificationModifierServiceHandler(ctx, mux, conn)
}

// RegisterSystemNotificationModifierServiceHandler registers the http handlers for service SystemNotificationModifierService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterSystemNotificationModifierServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterSystemNotificationModifierServiceHandlerClient(ctx, mux, NewSystemNotificationModifierServiceClient(conn))
}

// RegisterSystemNotificationModifierServiceHandlerClient registers the http handlers for service SystemNotificationModifierService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "SystemNotificationModifierServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "SystemNotificationModifierServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "SystemNotificationModifierServiceClient" to call the correct interceptors.
func RegisterSystemNotificationModifierServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client SystemNotificationModifierServiceClient) error {

	mux.Handle("POST", pattern_SystemNotificationModifierService_SetSystemNotificationStatus_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/notificationmgmt.v1.SystemNotificationModifierService/SetSystemNotificationStatus", runtime.WithHTTPPathPattern("/notificationmgmt/api/v1/proxy/notificationmgmts/set_system_notifications_status"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_SystemNotificationModifierService_SetSystemNotificationStatus_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_SystemNotificationModifierService_SetSystemNotificationStatus_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_SystemNotificationModifierService_SetSystemNotificationStatus_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4, 2, 5}, []string{"notificationmgmt", "api", "v1", "proxy", "notificationmgmts", "set_system_notifications_status"}, ""))
)

var (
	forward_SystemNotificationModifierService_SetSystemNotificationStatus_0 = runtime.ForwardResponseMessage
)
