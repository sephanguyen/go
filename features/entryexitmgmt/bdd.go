package entryexitmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/unleash"
	"github.com/manabie-com/backend/internal/entryexitmgmt/repositories"
	"github.com/manabie-com/backend/internal/entryexitmgmt/services"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/yasuo/constant"

	"github.com/cucumber/godog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	DefaultResourcePath   = "1"
	DownloadURLPrefix     = "http://minio-infras.emulator.svc.cluster.local:9000/manabie/entryexitmgmt-upload/"
	CurrentUpdatedVersion = "v2"
)

func init() {
	common.RegisterTest("entryexitmgmt", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	connections    *common.Connections
	zapLogger      *zap.Logger
	firebaseAddr   string
	useEntryExitDB bool
)

func TestSuiteInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c, f.FirebaseAddr)
		})

		ctx.AfterSuite(func() {
			connections.CloseAllConnections()
		})
	}
}

func StepStateFromContext(ctx context.Context) *common.StepState {
	return ctx.Value(common.StepStateKey{}).(*common.StepState)
}

func StepStateToContext(ctx context.Context, state *common.StepState) context.Context {
	return context.WithValue(ctx, common.StepStateKey{}, state)
}

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c)
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			schoolID := int32(-2147483634) // Currently this is the school ID that contains all user group
			resourcePath := fmt.Sprintf("%d", schoolID)

			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: resourcePath,
					DefaultRole:  entities.UserGroupSchoolAdmin,
					UserGroup:    entities.UserGroupSchoolAdmin,
				},
			}
			s.CommonSuite.StepState.ResourcePath = resourcePath
			s.CommonSuite.StepState.CurrentSchoolID = schoolID
			ctx = interceptors.ContextWithJWTClaims(ctx, &claim)

			// Create School Admin to be used as app.user_id for pre-condition data
			id := idutil.ULIDNow()
			ctx, err := s.aValidUser(
				StepStateToContext(ctx, s.CommonSuite.StepState),
				s.BobPostgresDBTrace,
				withID(id),
				withUserGroup(constant.UserGroupSchoolAdmin),
				withResourcePath(s.CommonSuite.StepState.ResourcePath),
				withRole(userConstant.RoleSchoolAdmin),
			)
			if err != nil {
				return StepStateToContext(ctx, s.StepState), fmt.Errorf("Error on creating init school admin %v", err)
			}

			// Assign the school admin ID to claims
			claim.Manabie.UserID = id

			return StepStateToContext(ctx, s.StepState), nil
		})

		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			stepState := StepStateFromContext(ctx)
			for _, v := range stepState.Subs {
				if v.IsValid() {
					err := v.Drain()
					if err != nil {
						return nil, err
					}
				}
			}
			return ctx, nil
		})
	}
}

func setup(c *common.Config, fakeFirebaseAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	connections = &common.Connections{}

	firebaseAddr = fakeFirebaseAddr

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	err := connections.ConnectGRPC(ctx,
		common.WithCredentials(grpc.WithInsecure()),
		common.WithEntryExitMgmtSvcAddress(),
		common.WithYasuoSvcAddress(),
		common.WithNotificationMgmtSvcAddress(),
		common.WithShamirSvcAddress(),
		common.WithMasterMgmtSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		common.WithEntryExitMgmtDBConfig(c.PostgresV2.Databases["entryexitmgmt"]),
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
		common.WithMastermgmtDBConfig(c.PostgresV2.Databases["mastermgmt"]),
	)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connectDB: %v", err))
	}

	connections.JSM, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to create jetstream management: %v", err))
	}
	connections.JSM.ConnectToJS()

	// Init auth info
	defaultValues := (&repository.OrganizationRepo{}).DefaultOrganizationAuthValues(c.Common.Environment)
	stmt := fmt.Sprintf(`
		INSERT INTO organization_auths
			(organization_id, auth_project_id, auth_tenant_id)
		SELECT
			school_id, 'fake_aud', ''
		FROM
			schools
		UNION 
		SELECT
			school_id, 'dev-manabie-online', ''
		FROM
			schools
		UNION %s
		ON CONFLICT 
			DO NOTHING
		;
		`, defaultValues)
	_, err = connections.BobDBTrace.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
}

type suite struct {
	*common.Connections
	*common.StepState
	ZapLogger                   *zap.Logger
	Cfg                         *common.Config
	CommonSuite                 *common.Suite
	UnleashSuite                *unleash.Suite
	StudentQrRepo               services.IStudentQRRepo
	StudentEntryExitRecordsRepo services.IStudentEntryExitRecordsRepo
	ApplicantID                 string
}

func newSuite(c *common.Config) *suite {
	s := &suite{
		Connections:  connections,
		Cfg:          c,
		ZapLogger:    zapLogger,
		CommonSuite:  &common.Suite{},
		UnleashSuite: &unleash.Suite{},
	}

	s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = s.CommonSuite.StepState

	s.CommonSuite.StepState.FirebaseAddress = firebaseAddr

	// Unleash
	s.UnleashSuite.Connections = s.Connections
	s.UnleashSuite.StepState = &common.StepState{}
	s.UnleashSuite.UnleashSrvAddr = c.UnleashSrvAddr
	s.UnleashSuite.UnleashAPIKey = c.UnleashAPIKey
	s.UnleashSuite.UnleashLocalAdminAPIKey = c.UnleashLocalAdminAPIKey

	s.StudentQrRepo = &repositories.StudentQRRepo{}
	s.StudentEntryExitRecordsRepo = &repositories.StudentEntryExitRecordsRepo{}

	return s
}
