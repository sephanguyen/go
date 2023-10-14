package notificationmgmt

import (
	"log"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/notification/config"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func setupInterceptors(c *config.Config, db database.QueryExecer, zapLogger *zap.Logger, jsm nats.JetStreamManagement) ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	// list options for lesson service
	opts := []grpc_zap.Option{
		grpc_zap.WithLevels(grpc_zap.DefaultCodeToLevel),
	}
	interceptors.GrpcServerViews = append(
		interceptors.GrpcServerViews,
		nats.JetstreamProcessedMessagesView,
		nats.JetstreamProcessedMessagesLatencyView,
	)
	if err := view.Register(interceptors.GrpcServerViews...); err != nil {
		log.Panicf("Failed to register ocgrpc server views: %v", err)
	}

	authInterceptor := authInterceptor(c, zapLogger, db)

	fakeSchoolAdminInterceptor := fakeSchoolAdminJwtInterceptor()
	grpcUnary := []grpc.UnaryServerInterceptor{
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(zapLogger, opts...),
		authInterceptor.UnaryServerInterceptor,
		fakeSchoolAdminInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(jsm, zapLogger, "notificationmgmt"),
	}

	grpcStream := []grpc.StreamServerInterceptor{
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.StreamServerInterceptor(zapLogger, opts...),
		authInterceptor.StreamServerInterceptor,
	}

	return grpcUnary, grpcStream
}
