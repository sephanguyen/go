package interceptor

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	glInterceptors "github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UpdateUserIDForParent(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := glInterceptors.StartSpan(ctx, "Auth.UpdateUserIDForParent")
	defer span.End()

	roleName := interceptors.UserRolesFromContext(ctx)
	if len(roleName) == 0 {
		return handler(ctx, req)
	}
	if roleName[0] != constant.RoleParent {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)

	if ok {
		studentIDs := md.Get("student-id")
		if len(studentIDs) > 0 && len(studentIDs[0]) > 0 {
			ctx = glInterceptors.ContextWithUserID(ctx, studentIDs[0])
			ctxzap.AddFields(ctx, zap.String("userID", studentIDs[0]))
		}
	}

	return handler(ctx, req)
}
