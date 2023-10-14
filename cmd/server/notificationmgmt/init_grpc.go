package notificationmgmt

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/notification/config"
	newNotiInfra "github.com/manabie-com/backend/internal/notification/infra"
	metrics "github.com/manabie-com/backend/internal/notification/infra/metrics"
	mediaController "github.com/manabie-com/backend/internal/notification/modules/media/controller"
	systemNotificationController "github.com/manabie-com/backend/internal/notification/modules/system_notification/controller"
	tagServices "github.com/manabie-com/backend/internal/notification/modules/tagmgmt/services"
	"github.com/manabie-com/backend/internal/notification/services"
	grpctrans "github.com/manabie-com/backend/internal/notification/transports/grpc"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	npbv2 "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v2"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/aws/aws-sdk-go/aws/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func initNewNotificationServer(
	server grpc.ServiceRegistrar,
	notiModifierSvc *services.NotificationModifierService,
	notiReaderSvc *services.NotificationReaderService,
) {
	modifierGRPCSvc := grpctrans.NewNotificationModifierService(notiModifierSvc)
	readerGRPCSvc := grpctrans.NewNotificationReaderService(notiReaderSvc)

	npb.RegisterNotificationModifierServiceServer(server, modifierGRPCSvc)
	npb.RegisterNotificationReaderServiceServer(server, readerGRPCSvc)
}

func initServersFromOldBob(config *config.Config, server grpc.ServiceRegistrar, db database.Ext) {
	bobNotiReaderSvc := grpctrans.NewBobLegacyNotificationReaderService(db, config.Common.Environment)
	bpb.RegisterNotificationReaderServiceServer(server, bobNotiReaderSvc)

	// nolint
	bobNotiModifierSvc := grpctrans.NewSimpleBobLegacyNotificationModifierService(db)
	bpb.RegisterNotificationModifierServiceServer(server, bobNotiModifierSvc)
}

func initServersFromOldYasuo(
	config *config.Config,
	s grpc.ServiceRegistrar,
	db database.Ext,
	s3Sess client.ConfigProvider,
	notificationPusher firebase.NotificationPusher,
	notiMetrics metrics.NotificationMetrics,
	logger *zap.Logger,
	jsm nats.JetStreamManagement,
) {
	pushNotificationService := newNotiInfra.NewPushNotificationService(notificationPusher, notiMetrics)

	yasuoNotiModifierSvc := grpctrans.NewYasuoLegacyNotificationModifierService(db, config.Storage, s3Sess, pushNotificationService, notiMetrics, jsm, config.Common.Environment)

	ypb.RegisterNotificationModifierServiceServer(s, yasuoNotiModifierSvc)
}

func initNewTagmgmtServer(
	server grpc.ServiceRegistrar,
	tagReaderSvc *tagServices.TagMgmtReaderService,
	tagModifierSvc *tagServices.TagMgmtModifierService,
) {
	npb.RegisterTagMgmtModifierServiceServer(server, tagModifierSvc)
	npb.RegisterTagMgmtReaderServiceServer(server, tagReaderSvc)
}

func initBobUserServerService(server *grpc.Server, notiModifierSvc *services.NotificationModifierService) {
	bobUserSvc := grpctrans.NewBobLegacyUserService(notiModifierSvc)
	pb.RegisterUserServiceServer(server, bobUserSvc)
}

func initInternalServerService(server *grpc.Server, internalSvc *services.InternalService) {
	npb.RegisterInternalServiceServer(server, internalSvc)
}

func initNewMediaServer(
	server grpc.ServiceRegistrar,
	mediaModifierSvc *mediaController.MediaModifierService,
) {
	npb.RegisterMediaModifierServiceServer(server, mediaModifierSvc)
}

func initNewNotificationV2Service(
	server grpc.ServiceRegistrar,
	notiReaderSvc *services.NotificationReaderService,
) {
	readerGRPCSvc := grpctrans.NewNotificationReaderV2Service(notiReaderSvc)

	npbv2.RegisterNotificationReaderServiceServer(server, readerGRPCSvc)
}

func initNewSystemNotificationService(
	server grpc.ServiceRegistrar,
	systemNotificationReaderSvc *systemNotificationController.SystemNotificationReaderService,
	systemNotificationModifierSvc *systemNotificationController.SystemNotificationModifierService,
) {
	npb.RegisterSystemNotificationReaderServiceServer(server, systemNotificationReaderSvc)
	npb.RegisterSystemNotificationModifierServiceServer(server, systemNotificationModifierSvc)
}
