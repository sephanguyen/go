package yasuo

import (
	"context"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/repositories"
	"github.com/manabie-com/backend/internal/yasuo/utils"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerActivityLogRequestInterceptor(activityLogRepo *repositories.ActivityLogRepo, db *database.DBTrace, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userID := interceptors.UserIDFromContext(ctx)
		if userID == "" {
			if utils.IsContain(ignoreAuthEndpoint, info.FullMethod) {
				return handler(ctx, req)
			}
			return "", status.Error(codes.Unauthenticated, "missing user id")
		}

		payload := map[string]interface{}{
			"req": req,
		}

		resp, err := handler(ctx, req)
		actionType := info.FullMethod + "_OK"

		if err != nil {
			payload["err"] = err.Error()
			actionType = info.FullMethod + "_" + status.Code(err).String()
		}

		activityLog := &bob_entities.ActivityLog{}
		database.AllNullEntity(activityLog)

		cerr := multierr.Combine(
			activityLog.UserID.Set(userID),
			activityLog.ActionType.Set(actionType),
			activityLog.Payload.Set(payload),
		)
		if cerr != nil {
			logger.Warn("multierr.Combine", zap.Error(cerr))
		}
		cerr = activityLogRepo.CreateV2(ctx, db, activityLog)
		if cerr != nil {
			logger.Warn("activityLogRepo.Create", zap.Error(cerr))
		}

		return resp, err
	}
}
