package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt/withus"
	enigmaRepo "github.com/manabie-com/backend/internal/enigma/repositories"
	enigmaService "github.com/manabie-com/backend/internal/enigma/services"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/caching"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	fClient "github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service/schoolmaster"
	staff_service "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service/staff"
	usergroup_service "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service/user_group"
	grpc_port "github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	http_port "github.com/manabie-com/backend/internal/usermgmt/modules/user/port/http"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/http/middleware"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/subscriber"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/features"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
	sppb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	health "google.golang.org/grpc/health/grpc_health_v1"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithHTTP(s).
		WithMonitorServicer(s).
		WithNatsServicer(s).
		Register(s)
}

// func getDB(c configurations.Config, rsc *bootstrap.Resources) *database.DBTrace {
// 	if c.Postgres.Disabled {
// 		return rsc.DBWith("bob")
// 	}
// 	return rsc.DB()
// }

type server struct {
	bootstrap.DefaultMonitorService[configurations.Config]

	authInterceptor    *interceptors.Auth
	tenantManager      multitenant.TenantManager
	firebaseAuthClient internal_auth_tenant.TenantClient

	shamirConn *grpc.ClientConn
	fatimaConn *grpc.ClientConn
	spikeConn  *grpc.ClientConn
	mainQueue  taskq.Queue

	userModifierSvc        *service.UserModifierService
	staffSvc               *staff_service.StaffService
	userGroupSvc           *usergroup_service.UserGroupService
	authSvc                *grpc_port.AuthService
	schoolInfoSvc          *schoolmaster.SchoolInfoService
	userReaderSvc          *service.UserReaderService
	studentSvc             *service.StudentService
	withusStudentSvc       *withus.StudentPortService
	userRegistrationSvc    *service.UserRegistrationService
	studentRegistrationSvc *service.StudentRegistrationService
	configurationClient    *clients.ConfigurationClient
}

func (*server) ServerName() string {
	return "usermgmt"
}

func (s *server) WithUnaryServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
	}

	grpcUnary := bootstrap.DefaultUnaryServerInterceptor(rsc)
	grpcUnary = append(grpcUnary, customs...)
	return grpcUnary
}

func (s *server) WithStreamServerInterceptors(c configurations.Config, rsc *bootstrap.Resources) []grpc.StreamServerInterceptor {
	grpcStream := bootstrap.DefaultStreamServerInterceptor(rsc)
	grpcStream = append(grpcStream, s.authInterceptor.StreamServerInterceptor)

	return grpcStream
}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (*server) WithOpencensusViews() []*view.View {
	return []*view.View{
		caching.CacheCounterView,
	}
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var err error
	db := rsc.DBWith("bob")
	jsm := rsc.NATS()
	unleashClientInstance := rsc.Unleash()

	s.authInterceptor = authInterceptor(&c, rsc.Logger(), db)

	s.shamirConn = rsc.GRPCDial("shamir")
	s.fatimaConn = rsc.GRPCDial("fatima")
	s.spikeConn = rsc.GRPCDial("spike")

	queueFactory := memqueue.NewFactory()
	s.mainQueue = queueFactory.RegisterQueue(&taskq.QueueOptions{
		Name: constants.UserMgmtTask,
	})

	firebaseProject := c.Common.FirebaseProject
	if firebaseProject == "" {
		firebaseProject = c.Common.GoogleCloudProject
	}
	firebaseApp, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: firebaseProject,
	})
	if err != nil {
		return fmt.Errorf("error initializing app: %w", err)
	}

	firebaseClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		return fmt.Errorf("error getting Auth client: %w", err)
	}

	singleTenantGCPApp, err := gcp.NewApp(ctx, "", firebaseProject)
	if err != nil {
		return fmt.Errorf("error init GCP app for single tenant env: %w", err)
	}

	s.firebaseAuthClient, err = internal_auth_tenant.NewFirebaseAuthClientFromGCP(ctx, singleTenantGCPApp)
	if err != nil {
		return fmt.Errorf("error init firebaseAuthClient for single tenant env: %w", err)
	}

	s.configurationClient = clients.InitConfigurationClient(rsc.GRPCDialContext(ctx, "mastermgmt", configs.RetryOptions{}))

	identityPlatformProject := c.Common.IdentityPlatformProject
	if identityPlatformProject == "" {
		identityPlatformProject = c.Common.GoogleCloudProject
	}
	multiTenantGCPApp, err := gcp.NewApp(ctx, "", identityPlatformProject)
	if err != nil {
		return fmt.Errorf("error init GCP app for multi tenant env: %w", err)
	}

	// For token signer
	multiTenantGCPAppForTokenSigner, err := gcp.NewGCPApp(ctx, "", c.MultiTenantTokenSignerConfig())
	if err != nil {
		return fmt.Errorf("error init GCP app for multi tenant token signer: %w, projectID: %s, serviceAccountID: %s",
			err, c.MultiTenantTokenSignerConfig().GetGCPProjectID(), c.MultiTenantTokenSignerConfig().GetGCPServiceAccountID())
	}

	secondaryTenantConfigProvider := &repository.TenantConfigRepo{
		QueryExecer:      db,
		ConfigAESKey:     c.IdentityPlatform.ConfigAESKey,
		ConfigAESIv:      c.IdentityPlatform.ConfigAESIv,
		OrganizationRepo: &repository.OrganizationRepo{},
	}

	s.tenantManager, err = internal_auth_tenant.NewTenantManagerFromGCP(ctx, multiTenantGCPApp, internal_auth_tenant.WithSecondaryTenantConfigProvider(secondaryTenantConfigProvider))
	if err != nil {
		return fmt.Errorf("error init tenantManager for multi tenant env: %w", err)
	}

	// For token signer
	tenantManagerForTokenSigner, err := internal_auth_tenant.NewTenantManagerFromGCP(ctx, multiTenantGCPAppForTokenSigner, internal_auth_tenant.WithSecondaryTenantConfigProvider(secondaryTenantConfigProvider))
	if err != nil {
		tenantManagerForTokenSigner = s.tenantManager
		rsc.Logger().Error("failed to init tenantManagerForTokenSigner", zap.Error(err))
	}

	subscriptionModifierServiceClient := fpb.NewSubscriptionModifierServiceClient(s.fatimaConn)

	OrganizationRepoWithDefaultValue := (&repository.OrganizationRepo{}).WithDefaultValue(c.Common.Environment)

	featureManager := &features.FeatureManager{
		UnleashClient:             unleashClientInstance,
		Env:                       c.Common.Environment,
		DB:                        db,
		InternalConfigurationRepo: &repository.DomainInternalConfigurationRepo{},
	}

	s.userModifierSvc = &service.UserModifierService{
		DB:                        db,
		OrganizationRepo:          OrganizationRepoWithDefaultValue,
		UsrEmailRepo:              &repository.UsrEmailRepo{},
		DomainUserRepo:            &repository.DomainUserRepo{},
		DomainStudentRepo:         &repository.DomainStudentRepo{},
		DomainUsrEmailRepo:        &repository.DomainUsrEmailRepo{},
		UserRepo:                  &repository.UserRepo{},
		TeacherRepo:               &repository.TeacherRepo{},
		SchoolAdminRepo:           &repository.SchoolAdminRepo{},
		UserGroupRepo:             &repository.UserGroupRepo{},
		UserGroupV2Repo:           &repository.UserGroupV2Repo{},
		UserGroupsMemberRepo:      &repository.UserGroupsMemberRepo{},
		StudentRepo:               &repository.StudentRepo{},
		ParentRepo:                &repository.ParentRepo{},
		StudentParentRepo:         &repository.StudentParentRepo{},
		FirebaseAuthClient:        s.firebaseAuthClient,
		TenantManager:             s.tenantManager,
		UserAccessPathRepo:        &repository.UserAccessPathRepo{},
		ImportUserEventRepo:       &repository.ImportUserEventRepo{},
		LocationRepo:              &location_repo.LocationRepo{},
		FirebaseClient:            firebaseClient,
		UnleashClient:             unleashClientInstance,
		FatimaClient:              subscriptionModifierServiceClient,
		JSM:                       jsm,
		TaskQueue:                 s.mainQueue,
		Env:                       c.Common.Environment,
		SchoolHistoryRepo:         &repository.SchoolHistoryRepo{},
		SchoolInfoRepo:            &repository.SchoolInfoRepo{},
		SchoolCourseRepo:          &repository.SchoolCourseRepo{},
		UserAddressRepo:           &repository.UserAddressRepo{},
		PrefectureRepo:            &repository.PrefectureRepo{},
		UserPhoneNumberRepo:       &repository.UserPhoneNumberRepo{},
		DomainGradeRepo:           &repository.DomainGradeRepo{},
		GradeOrganizationRepo:     &repository.GradeOrganizationRepo{},
		DomainTagRepo:             &repository.DomainTagRepo{},
		DomainTaggedUserRepo:      &repository.DomainTaggedUserRepo{},
		InternalConfigurationRepo: &repository.DomainInternalConfigurationRepo{},
	}

	s.userGroupSvc = &usergroup_service.UserGroupService{
		DB:                        db,
		UnleashClient:             unleashClientInstance,
		Env:                       c.Common.Environment,
		RoleRepo:                  &repository.RoleRepo{},
		UserGroupRepo:             &repository.UserGroupRepo{},
		UserGroupV2Repo:           &repository.UserGroupV2Repo{},
		GrantedRoleRepo:           &repository.GrantedRoleRepo{},
		UserGroupsMemberRepo:      &repository.UserGroupsMemberRepo{},
		GrantedRoleAccessPathRepo: &repository.GrantedRoleAccessPathRepo{},
		LocationRepo:              &location_repo.LocationRepo{},
		UserRepo:                  &repository.UserRepo{},
		UserModifierService:       s.userModifierSvc,
		JSM:                       jsm,
	}

	s.staffSvc = &staff_service.StaffService{
		DB:                 db,
		FirebaseAuthClient: s.firebaseAuthClient,
		TenantManager:      s.tenantManager,
		FirebaseClient:     firebaseClient,
		FirebaseUtils:      fClient.NewAuthUtils(),
		FatimaClient:       subscriptionModifierServiceClient,
		JSM:                jsm,
		UnleashClient:      unleashClientInstance,
		Env:                c.Common.Environment,

		UserModifierService: s.userModifierSvc,
		UserGroupV2Service:  s.userGroupSvc,

		SchoolAdminRepo:     &repository.SchoolAdminRepo{},
		TeacherRepo:         &repository.TeacherRepo{},
		StaffRepo:           &repository.StaffRepo{},
		UserGroupRepo:       &repository.UserGroupRepo{},
		UserAccessPathRepo:  &repository.UserAccessPathRepo{},
		UserPhoneNumberRepo: &repository.UserPhoneNumberRepo{},
		UserRepo:            &repository.DomainUserRepo{},
		RoleRepo:            &repository.DomainRoleRepo{},
		DomainUser: &service.DomainUser{
			DB:       db,
			UserRepo: &repository.DomainUserRepo{},
		},
	}
	s.authSvc = &grpc_port.AuthService{
		DomainAuthService: &service.DomainAuthService{
			TenantManager:             tenantManagerForTokenSigner,
			FirebaseClient:            firebaseClient,
			ShamirClient:              spb.NewTokenReaderServiceClient(s.shamirConn),
			EmailServiceClient:        sppb.NewEmailModifierServiceClient(s.spikeConn),
			DB:                        db,
			ExternalConfigurationRepo: &repository.DomainExternalConfigurationRepo{},
		},
	}
	s.schoolInfoSvc = schoolmaster.NewSchoolInfoService(db, &repository.SchoolInfoRepo{}, jsm)

	s.userReaderSvc = &service.UserReaderService{
		DB:               db,
		UserRepo:         &repository.UserRepo{},
		StudentRepo:      &repository.StudentRepo{},
		UserGroupV2Repo:  &repository.UserGroupV2Repo{},
		OrganizationRepo: &repository.OrganizationRepo{},
	}

	s.studentSvc = &service.StudentService{
		DB:                          db,
		FirebaseAuthClient:          s.firebaseAuthClient,
		StudentCommentRepo:          &repository.StudentCommentRepo{},
		UserRepo:                    &repository.UserRepo{},
		StudentRepo:                 &repository.StudentRepo{},
		UserGroupRepo:               &repository.UserGroupRepo{},
		UserGroupV2Repo:             &repository.UserGroupV2Repo{},
		UserGroupsMemberRepo:        &repository.UserGroupsMemberRepo{},
		UsrEmailRepo:                &repository.UsrEmailRepo{},
		UserAccessPathRepo:          &repository.UserAccessPathRepo{},
		OrganizationRepo:            OrganizationRepoWithDefaultValue,
		UserModifierService:         s.userModifierSvc,
		ImportUserEventRepo:         &repository.ImportUserEventRepo{},
		JSM:                         jsm,
		TaskQueue:                   s.mainQueue,
		UnleashClient:               unleashClientInstance,
		Env:                         c.Common.Environment,
		GradeOrganizationRepo:       &repository.GradeOrganizationRepo{},
		UserAddressRepo:             &repository.UserAddressRepo{},
		PrefectureRepo:              &repository.PrefectureRepo{},
		UserPhoneNumberRepo:         &repository.UserPhoneNumberRepo{},
		SchoolHistoryRepo:           &repository.SchoolHistoryRepo{},
		SchoolInfoRepo:              &repository.SchoolInfoRepo{},
		SchoolCourseRepo:            &repository.SchoolCourseRepo{},
		DomainTagRepo:               &repository.DomainTagRepo{},
		EnrollmentStatusHistoryRepo: &repository.DomainEnrollmentStatusHistoryRepo{},
		DomainUserAccessPathRepo:    &repository.DomainUserAccessPathRepo{},
		DomainLocationRepo:          &repository.DomainLocationRepo{},
		ConfigurationClient:         s.configurationClient,
		DomainStudentService: &service.DomainStudent{
			DB:                  db,
			JSM:                 jsm,
			FirebaseAuthClient:  s.firebaseAuthClient,
			TenantManager:       s.tenantManager,
			FatimaClient:        subscriptionModifierServiceClient,
			ConfigurationClient: s.configurationClient,
			StudentRepo: &repository.DomainStudentRepo{
				UserRepo:            &repository.DomainUserRepo{},
				LegacyUserGroupRepo: &repository.LegacyUserGroupRepo{},
				UserAccessPathRepo:  &repository.DomainUserAccessPathRepo{},
				UserGroupMemberRepo: &repository.DomainUserGroupMemberRepo{},
			},
			UserRepo:                    &repository.DomainUserRepo{},
			UserGroupRepo:               &repository.DomainUserGroupRepo{},
			UserAddressRepo:             &repository.DomainUserAddressRepo{},
			UserPhoneNumberRepo:         &repository.DomainUserPhoneNumberRepo{},
			SchoolHistoryRepo:           &repository.DomainSchoolHistoryRepo{},
			SchoolRepo:                  &repository.DomainSchoolRepo{},
			SchoolCourseRepo:            &repository.DomainSchoolCourseRepo{},
			LocationRepo:                &repository.DomainLocationRepo{},
			GradeRepo:                   &repository.DomainGradeRepo{},
			PrefectureRepo:              &repository.DomainPrefectureRepo{},
			UsrEmailRepo:                &repository.DomainUsrEmailRepo{},
			OrganizationRepo:            OrganizationRepoWithDefaultValue,
			TagRepo:                     &repository.DomainTagRepo{},
			TaggedUserRepo:              &repository.DomainTaggedUserRepo{},
			EnrollmentStatusHistoryRepo: &repository.DomainEnrollmentStatusHistoryRepo{},
			UserAccessPathRepo:          &repository.DomainUserAccessPathRepo{},
			StudentPackage:              &repository.DomainStudentPackageRepo{},
			AuthUserUpserter:            service.NewLegacyAuthUserUpserter(&repository.DomainUserRepo{}, OrganizationRepoWithDefaultValue, s.firebaseAuthClient, s.tenantManager),
			UnleashClient:               unleashClientInstance,
			Env:                         c.Common.Environment,
			InternalConfigurationRepo:   &repository.DomainInternalConfigurationRepo{},
			StudentValidationManager: &service.StudentValidationManager{
				UserRepo:                    &repository.DomainUserRepo{},
				UserGroupRepo:               &repository.DomainUserGroupRepo{},
				LocationRepo:                &repository.DomainLocationRepo{},
				GradeRepo:                   &repository.DomainGradeRepo{},
				SchoolRepo:                  &repository.DomainSchoolRepo{},
				SchoolCourseRepo:            &repository.DomainSchoolCourseRepo{},
				PrefectureRepo:              &repository.DomainPrefectureRepo{},
				TagRepo:                     &repository.DomainTagRepo{},
				InternalConfigurationRepo:   &repository.DomainInternalConfigurationRepo{},
				EnrollmentStatusHistoryRepo: &repository.DomainEnrollmentStatusHistoryRepo{},
				StudentRepo:                 &repository.DomainStudentRepo{},
			},
			// NOTE: example of how to use SF repo
			// StudentSFRepo: &objects.DomainStudentRepo{},
			FeatureManager: featureManager,
		},
		FeatureManager: featureManager,
	}

	withusStudentService, err := withus.NewStudentService(ctx, &c, rsc)
	if err != nil {
		return errors.Wrap(err, "withus.NewStudentService")
	}
	s.withusStudentSvc = &withus.StudentPortService{
		StudentService: withusStudentService,
	}

	partnerSyncDataLogService := &enigmaService.PartnerSyncDataLogService{
		DB:                          db,
		PartnerSyncDataLogSplitRepo: &enigmaRepo.PartnerSyncDataLogSplitRepo{},
		PartnerSyncDataLogRepo:      &enigmaRepo.PartnerSyncDataLogRepo{},
	}
	s.userRegistrationSvc = &service.UserRegistrationService{
		DB:                                 db,
		Logger:                             rsc.Logger(),
		UserRepo:                           &repository.UserRepo{},
		StudentRepo:                        &repository.StudentRepo{},
		TeacherRepo:                        &repository.TeacherRepo{},
		StaffRepo:                          &repository.StaffRepo{},
		PartnerSyncDataLogService:          partnerSyncDataLogService,
		UserGroupV2Repo:                    &repository.UserGroupV2Repo{},
		UserGroupMemberRepo:                &repository.UserGroupsMemberRepo{},
		LocationRepo:                       &location_repo.LocationRepo{},
		UserAccessPathRepo:                 &repository.UserAccessPathRepo{},
		StudentEnrollmentStatusHistoryRepo: &repository.StudentEnrollmentStatusHistoryRepo{},
	}
	s.studentRegistrationSvc = &service.StudentRegistrationService{
		DB:                                       db,
		Logger:                                   rsc.Logger(),
		DomainEnrollmentStatusHistoryRepo:        &repository.DomainEnrollmentStatusHistoryRepo{},
		DomainUserAccessPathRepo:                 &repository.DomainUserAccessPathRepo{},
		UnleashClient:                            unleashClientInstance,
		Env:                                      c.Common.Environment,
		StudentRepo:                              &repository.DomainStudentRepo{},
		LocationRepo:                             &repository.DomainLocationRepo{},
		EnrollmentStatusHistoryStartDateModifier: service.EnrollmentStatusHistoryStartDateModifier,
	}
	s.studentRegistrationSvc.OrderFlowEnrollmentStatusManager = &service.HandelOrderFlowEnrollmentStatus{
		Logger:                            rsc.Logger(),
		DomainEnrollmentStatusHistoryRepo: &repository.DomainEnrollmentStatusHistoryRepo{},
		DomainUserAccessPathRepo:          &repository.DomainUserAccessPathRepo{},
		SyncEnrollmentStatusHistory:       s.studentRegistrationSvc.SyncEnrollmentStatusHistory,
		DeactivateAndReactivateStudents:   s.studentRegistrationSvc.DeactivateAndReactivateStudents,
	}

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcserver *grpc.Server, c configurations.Config, rsc *bootstrap.Resources) error {
	pb.RegisterUserReaderServiceServer(grpcserver, s.userReaderSvc)
	pb.RegisterUserModifierServiceServer(grpcserver, s.userModifierSvc)
	pb.RegisterStaffServiceServer(grpcserver, s.staffSvc)
	pb.RegisterUserGroupMgmtServiceServer(grpcserver, s.userGroupSvc)
	pb.RegisterAuthServiceServer(grpcserver, s.authSvc)

	pb.RegisterSchoolInfoServiceServer(grpcserver, s.schoolInfoSvc)
	pb.RegisterStudentServiceServer(grpcserver, s.studentSvc)
	pb.RegisterWithusStudentServiceServer(grpcserver, s.withusStudentSvc)

	health.RegisterHealthServer(grpcserver, &healthcheck.Service{DB: rsc.DBWith("bob").DB.(*pgxpool.Pool)})
	return nil
}

func (s *server) SetupHTTP(c configurations.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	s.fatimaConn = rsc.GRPCDial("fatima")

	bobDB := rsc.DBWith("bob")

	userRepo := &repository.DomainUserRepo{}
	userGroupRepo := &repository.DomainUserGroupRepo{}
	userAddressRepo := &repository.DomainUserAddressRepo{}
	userPhoneNumberRepo := &repository.DomainUserPhoneNumberRepo{}
	schoolHistoryRepo := &repository.DomainSchoolHistoryRepo{}
	legacyUserGroup := &repository.LegacyUserGroupRepo{}
	userAccessPathRepo := &repository.DomainUserAccessPathRepo{}
	userGroupMemberRepo := &repository.DomainUserGroupMemberRepo{}
	locationRepo := &repository.DomainLocationRepo{}
	gradeRepo := &repository.DomainGradeRepo{}
	schoolRepo := &repository.DomainSchoolRepo{}
	schoolCourseRepo := &repository.DomainSchoolCourseRepo{}
	prefectureRepo := &repository.DomainPrefectureRepo{}
	usrEmailRepo := &repository.DomainUsrEmailRepo{}
	organizationRepo := (&repository.OrganizationRepo{}).WithDefaultValue(c.Common.Environment)
	enrollmentStatusHistoryRepo := &repository.DomainEnrollmentStatusHistoryRepo{}
	tagRepo := &repository.DomainTagRepo{}
	taggedUserRepo := &repository.DomainTaggedUserRepo{}
	studentPackage := &repository.DomainStudentPackageRepo{}
	studentParentRepo := &repository.DomainStudentParentRelationshipRepo{}
	internalConfigurationRepo := &repository.DomainInternalConfigurationRepo{}
	subscriptionModifierServiceClient := fpb.NewSubscriptionModifierServiceClient(s.fatimaConn)
	unleashClientInstance := rsc.Unleash()

	healthCheckHttp := http_port.HealthCheckService{}
	domainStudentHttp := http_port.DomainStudentService{
		DomainStudent: &service.DomainStudent{
			DB:                  bobDB,
			JSM:                 rsc.NATS(),
			FirebaseAuthClient:  s.firebaseAuthClient,
			TenantManager:       s.tenantManager,
			ConfigurationClient: s.configurationClient,
			StudentRepo: &repository.DomainStudentRepo{
				UserRepo:            userRepo,
				LegacyUserGroupRepo: legacyUserGroup,
				UserAccessPathRepo:  userAccessPathRepo,
				UserGroupMemberRepo: userGroupMemberRepo,
			},
			UserRepo:                         userRepo,
			UserGroupRepo:                    userGroupRepo,
			UserAddressRepo:                  userAddressRepo,
			UserPhoneNumberRepo:              userPhoneNumberRepo,
			SchoolHistoryRepo:                schoolHistoryRepo,
			SchoolRepo:                       schoolRepo,
			SchoolCourseRepo:                 schoolCourseRepo,
			LocationRepo:                     locationRepo,
			GradeRepo:                        gradeRepo,
			PrefectureRepo:                   prefectureRepo,
			UsrEmailRepo:                     usrEmailRepo,
			OrganizationRepo:                 organizationRepo,
			TagRepo:                          tagRepo,
			TaggedUserRepo:                   taggedUserRepo,
			EnrollmentStatusHistoryRepo:      enrollmentStatusHistoryRepo,
			UserAccessPathRepo:               userAccessPathRepo,
			InternalConfigurationRepo:        internalConfigurationRepo,
			StudentPackage:                   studentPackage,
			FatimaClient:                     subscriptionModifierServiceClient,
			StudentParentRelationshipManager: service.NewStudentParentRelationshipManager(&repository.DomainStudentParentRelationshipRepo{}),
			AuthUserUpserter:                 service.NewLegacyAuthUserUpserter(userRepo, organizationRepo, s.firebaseAuthClient, s.tenantManager),
			UnleashClient:                    unleashClientInstance,
			Env:                              c.Common.Environment,
			StudentValidationManager: &service.StudentValidationManager{
				UserRepo:                    &repository.DomainUserRepo{},
				UserGroupRepo:               &repository.DomainUserGroupRepo{},
				LocationRepo:                &repository.DomainLocationRepo{},
				GradeRepo:                   &repository.DomainGradeRepo{},
				SchoolRepo:                  &repository.DomainSchoolRepo{},
				SchoolCourseRepo:            &repository.DomainSchoolCourseRepo{},
				PrefectureRepo:              &repository.DomainPrefectureRepo{},
				TagRepo:                     &repository.DomainTagRepo{},
				InternalConfigurationRepo:   &repository.DomainInternalConfigurationRepo{},
				EnrollmentStatusHistoryRepo: &repository.DomainEnrollmentStatusHistoryRepo{},
				StudentRepo:                 &repository.DomainStudentRepo{},
			},
		},
		FeatureManager: &features.FeatureManager{
			UnleashClient:             unleashClientInstance,
			Env:                       c.Common.Environment,
			DB:                        bobDB,
			InternalConfigurationRepo: &repository.DomainInternalConfigurationRepo{},
		},
	}

	domainParent := http_port.DomainParentService{
		DomainParent: &service.DomainParent{
			DB:                 bobDB,
			JSM:                rsc.NATS(),
			FirebaseAuthClient: s.firebaseAuthClient,
			TenantManager:      s.tenantManager,
			UnleashClient:      unleashClientInstance,
			Env:                c.Common.Environment,
			UserRepo:           userRepo,
			UserGroupRepo:      userGroupRepo,
			ParentRepo: &repository.DomainParentRepo{
				UserRepo:            userRepo,
				LegacyUserGroupRepo: legacyUserGroup,
				UserAccessPathRepo:  userAccessPathRepo,
				UserGroupMemberRepo: userGroupMemberRepo,
			},
			UserPhoneNumberRepo:           userPhoneNumberRepo,
			UserAccessPathRepo:            userAccessPathRepo,
			TagRepo:                       tagRepo,
			UsrEmailRepo:                  usrEmailRepo,
			OrganizationRepo:              organizationRepo,
			TaggedUserRepo:                taggedUserRepo,
			StudentParentRepo:             studentParentRepo,
			InternalConfigurationRepo:     internalConfigurationRepo,
			AssignParentToStudentsManager: service.NewAssignParentToStudentsManager(&repository.DomainStudentParentRelationshipRepo{}),
			AuthUserUpserter:              service.NewLegacyAuthUserUpserter(userRepo, organizationRepo, s.firebaseAuthClient, s.tenantManager),
		},
		FeatureManager: &features.FeatureManager{
			UnleashClient:             unleashClientInstance,
			Env:                       c.Common.Environment,
			DB:                        bobDB,
			InternalConfigurationRepo: &repository.DomainInternalConfigurationRepo{},
		},
		UnleashClient: unleashClientInstance,
		Env:           c.Common.Environment,
	}

	groupDecider := middleware.NewGroupDecider(rsc.DBWith("bob"))

	r.Use(middleware.VerifySignature(zapLogger, groupDecider, spb.NewTokenReaderServiceClient(s.shamirConn)))

	r.GET(constant.HealthCheckStatusEndpoint, healthCheckHttp.Status)
	r.PUT(constant.DomainStudentEndpoint, domainStudentHttp.UpsertStudents)
	r.PUT(constant.DomainParentEndpoint, domainParent.UpsertParents)

	return nil
}

func (s *server) GracefulShutdown(ctx context.Context) {
	s.mainQueue.Close()
	s.fatimaConn.Close()
	s.shamirConn.Close()
}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	userRegistrationSubscriber := &subscriber.UserRegistrationSubscriber{
		JSM:                     rsc.NATS(),
		UserRegistrationService: s.userRegistrationSvc,
	}

	studentRegistrationSubscriber := &subscriber.StudentRegistrationSubscriber{
		JSM:                        rsc.NATS(),
		StudentRegistrationService: s.studentRegistrationSvc,
	}

	reallocateStudentEnrollmentStatusSubscriber := &subscriber.ReallocateStudentEnrollmentStatusSubscriber{
		JSM:                        rsc.NATS(),
		StudentRegistrationService: s.studentRegistrationSvc,
	}

	if err := userRegistrationSubscriber.Subscribe(); err != nil {
		return fmt.Errorf("userRegistrationSubscriber: failed to subscribe user registration %w", err)
	}

	if err := studentRegistrationSubscriber.Subscribe(); err != nil {
		return fmt.Errorf("StudentRegistrationSubscriber: failed to subscribe student registration %w", err)
	}

	if err := reallocateStudentEnrollmentStatusSubscriber.Subscribe(); err != nil {
		return fmt.Errorf("reallocateStudentEnrollmentStatusSubscriber: failed to subscribe student registration %w", err)
	}

	return nil
}
