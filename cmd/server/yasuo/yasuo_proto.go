package yasuo

import (
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	newNotiInfra "github.com/manabie-com/backend/internal/notification/infra"
	newNotiMetrics "github.com/manabie-com/backend/internal/notification/infra/metrics"
	notigrpctrans "github.com/manabie-com/backend/internal/notification/transports/grpc"
	"github.com/manabie-com/backend/internal/yasuo/configurations"
	"github.com/manabie-com/backend/internal/yasuo/services"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	firebaseAuthV4 "firebase.google.com/go/v4/auth"
	"github.com/aws/aws-sdk-go/aws/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// nolint
func initV1Yasuo(
	c *configurations.Config,
	s grpc.ServiceRegistrar,
	db database.Ext,
	lessonDB database.Ext,
	jsm nats.JetStreamManagement,
	unleashClientIns unleashclient.ClientInstance,
	firebaseClient *firebaseAuthV4.Client,
	firebaseAuthClient multitenant.TenantClient,
	tenantManager multitenant.TenantManager,
	fatimaConn grpc.ClientConnInterface,
	storageConfig configs.StorageConfig,
	s3Sess client.ConfigProvider,
	notificationPusher firebase.NotificationPusher,
	oldCourseService *services.CourseService,
	notiMetrics newNotiMetrics.NotificationMetrics,
	logger *zap.Logger,
) {
	subscriptionModifierServiceClient := fpb.NewSubscriptionModifierServiceClient(fatimaConn)

	ypb.RegisterCourseModifierServiceServer(s, services.NewCourseModifierService(oldCourseService.EurekaDBTrace, lessonDB, db, oldCourseService, unleashClientIns, c.Common.Environment))
	ypb.RegisterUserModifierServiceServer(s, services.NewUserModifierService(c, db, jsm, firebaseClient, firebaseAuthClient, tenantManager, subscriptionModifierServiceClient))
	pushNotificationService := newNotiInfra.NewPushNotificationService(notificationPusher, notiMetrics)

	yasuoNotiModifierSvc := notigrpctrans.NewYasuoLegacyNotificationModifierService(db, storageConfig, s3Sess, pushNotificationService, notiMetrics, jsm, c.Common.Environment)

	ypb.RegisterNotificationModifierServiceServer(s, yasuoNotiModifierSvc)

	// This one support bdd test for scheduled notification
	internalSvc := services.NewInternalService()
	ypb.RegisterInternalServiceServer(s, internalSvc)
}
