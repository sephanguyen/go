package tracer

import (
	"context"
	"encoding/json"
	"path"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// now is replaced in unit tests to return a fixed time instead.
var now = timestamppb.Now

func UnaryActivityLogRequestInterceptor(jsm nats.JetStreamManagement, logger *zap.Logger, _ string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		actionType := info.FullMethod
		userID := interceptors.UserIDFromContext(ctx)
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		appVersion := getAppVersionFromContext(ctx)

		// See https://github.com/manabie-com/backend/blob/45858b2f14b71222be237e6d133fd488ba32d913/internal/golibs/interceptors/app_version.go#L218
		// more more information on what is required in VerifyAppVersion API.
		var authorityHeader string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			authorityHeader = md.Get(":authority")[0]
		}

		if strings.Contains(authorityHeader, ":31") {
			service := path.Dir(actionType)[1:]
			method := path.Base(actionType)

			// Use the provided zap.Logger for logging but use the fields from context.
			logEntry := logger.With(append([]zapcore.Field{
				grpc_zap.SystemField,
				grpc_zap.ServerField,
				zap.String("grpc.service", service),
				zap.String("grpc.method", method),
			}, ctxzap.TagsToFields(ctx)...)...)

			logEntry.Check(zapcore.DebugLevel, "VerifyAppVersion request").
				Write(
					zap.String("grpc.request.method", info.FullMethod),
					zap.String("grpc.request.header.authority", authorityHeader),
					zap.String("grpc.request.header.userId", userID),
					zap.String("grpc.request.header.app_version", appVersion),
				)
		}

		requestAt := now()
		resp, err := handler(ctx, req)
		if strings.Contains(actionType, "Health/Check") || strings.Contains(actionType, "/PingSubscribeV2") ||
			strings.Contains(actionType, "/VerifyAppVersion") || jsm == nil {
			return resp, err
		}

		payload := map[string]interface{}{
			"req":         req,
			"app_version": appVersion,
		}

		statusLog := "OK"
		if err != nil {
			payload["err"] = err.Error()
			statusLog = status.Code(err).String()
		}
		payloadJSON, cerr := json.Marshal(payload)
		if cerr != nil {
			logger.Warn("json.Marshal", zap.Error(cerr))
		}
		msg := npb.ActivityLogEvtCreated{
			UserId:       userID,
			ActionType:   actionType,
			ResourcePath: resourcePath,
			RequestAt:    requestAt,
			FinishedAt:   now(),
			Payload:      payloadJSON,
			Status:       statusLog,
		}

		data, cerr := proto.Marshal(&msg)
		if cerr != nil {
			logger.Warn("proto.Marshal", zap.Error(cerr))
		}

		_, cerr = jsm.PublishAsyncContext(ctx, constants.SubjectActivityLogCreated, data)
		if cerr != nil {
			logger.Warn("jetstream.PublishAsync", zap.Error(cerr))
		}

		return resp, err
	}
}

func getAppVersionFromContext(ctx context.Context) string {
	headers, ok := metadata.FromIncomingContext(ctx)
	if ok {
		vers := headers.Get("version")
		if len(vers) > 0 {
			return vers[0]
		}
	}
	return ""
}
