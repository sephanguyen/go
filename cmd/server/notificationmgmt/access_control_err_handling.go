package notificationmgmt

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	errMsgViolateRLS = "ERROR: new row violates row-level security policy"
)

func UnaryAccessControlErrorHandlingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	actionType := info.FullMethod

	statusLog := "OK"
	if err != nil {
		statusLog = status.Code(err).String()
	}

	if statusLog == "Internal" {
		errMsg := err.Error()
		if strings.Contains(errMsg, errMsgViolateRLS) {
			if strings.Contains(actionType, "UpsertNotification") ||
				strings.Contains(actionType, "SendNotification") ||
				strings.Contains(actionType, "NotifyUnreadUser") {
				return resp, status.Errorf(codes.InvalidArgument, "PermissionDenied: Unauthorized to edit notification. Trace more detail: %v", err.Error())
			}
			if strings.Contains(actionType, "DiscardNotification") {
				return resp, status.Errorf(codes.InvalidArgument, "PermissionDenied: Unauthorized to discard notification. Trace more detail: %v", err.Error())
			}
			if strings.Contains(actionType, "DeleteNotification") {
				return resp, status.Errorf(codes.InvalidArgument, "PermissionDenied: Unauthorized to delete notification. Trace more detail: %v", err.Error())
			}
			return resp, status.Errorf(codes.InvalidArgument, "PermissionDenied: Unauthorized. Trace more detail: %v", err.Error())
		}
	}

	return resp, err
}
