package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/unleash"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/bob/entities"
	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/cucumber/godog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	common.RegisterTest("lessonmgmt", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	connections    *common.Connections
	searchClient   *elastic.SearchFactoryImpl
	firebaseClient *auth.Client // changing to firebaseAuthClient
	zapLogger      *zap.Logger
	firebaseAddr   string
	applicantID    string

	rootAccount map[int]common.AuthInfo
)

func LoadLocalLocation() *time.Location {
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	return loc
}

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
			ctx = StepStateToContext(ctx, s.CommonSuite.StepState)
			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
					DefaultRole:  entities.UserGroupAdmin,
					UserGroup:    entities.UserGroupAdmin,
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, &claim)

			return ctx, nil
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
		common.WithBobSvcAddress(),
		common.WithTomSvcAddress(),
		common.WithEurekaSvcAddress(),
		common.WithFatimaSvcAddress(),
		common.WithShamirSvcAddress(),
		common.WithYasuoSvcAddress(),
		common.WithUserMgmtSvcAddress(),
		common.WithLessonMgmtSvcAddress(),
		common.WithMasterMgmtSvcAddress(),
		common.WithTimesheetSvcAddress(),
		common.WithCalendarSvcAddress(),
		common.WithPaymentSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
		common.WithTomDBConfig(c.PostgresV2.Databases["tom"]),
		common.WithEurekaDBConfig(c.PostgresV2.Databases["eureka"]),
		common.WithFatimaDBConfig(c.PostgresV2.Databases["fatima"]),
		common.WithZeusDBConfig(c.PostgresV2.Databases["zeus"]),
		common.WithTimesheetPostgresDBConfig(c.PostgresV2.Databases["timesheet"], c.PostgresMigrate.Database.Password),
		common.WithCalendarDBConfig(c.PostgresV2.Databases["calendar"]),
		common.WithLessonmgmtDBConfig(c.PostgresV2.Databases["lessonmgmt"]),
		common.WithAuthPostgresDBConfig(c.PostgresV2.Databases["auth"], c.PostgresMigrate.Database.Password),
	)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connectDB: %v", err))
	}

	searchClient, err = elastic.NewSearchFactory(zapLogger, c.ElasticSearch.Addresses, c.ElasticSearch.Username, c.ElasticSearch.Password, "", "")
	if err != nil {
		zapLogger.Fatal("unable to connect elasticsearch", zap.Error(err))
	}

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	connections.FirebaseClient, err = app.Auth(ctx)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create firebase client: %v", err))
	}

	connections.JSM, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to create jetstream management: %v", err))
	}
	connections.JSM.ConnectToJS()

	connections.GCPApp, err = gcp.NewApp(ctx, "", c.Common.IdentityPlatformProject)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	connections.FirebaseAuthClient, err = internal_auth_tenant.NewFirebaseAuthClientFromGCP(ctx, connections.GCPApp)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	connections.TenantManager, err = internal_auth_tenant.NewTenantManagerFromGCP(ctx, connections.GCPApp)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create tenant manager: %v", err))
	}

	keycloakOpts := internal_auth.KeyCloakOpts{
		Path:     "https://d2020-ji-sso.jprep.jp",
		Realm:    "manabie-test",
		ClientID: "manabie-app",
	}

	connections.KeycloakClient, err = internal_auth.NewKeyCloakClient(keycloakOpts)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create keycloak client: %v", err))
	}

	applicantID = c.JWTApplicant

	err = common.UpdateResourcePath(connections.BobDB)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to update resource_path: %v", err))
	}

	// Init auth info
	stmt := `
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
		ON CONFLICT 
			DO NOTHING
		;
		`
	_, err = connections.BobDB.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}

	rootAccount, err = usermgmt.InitRootAccount(ctx, connections.ShamirConn, firebaseAddr, c.JWTApplicant)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init rootAccount: %v", err))
	}
}

type Suite struct {
	*common.Connections
	ZapLogger    *zap.Logger
	Cfg          *common.Config
	CommonSuite  *common.Suite
	UnleashSuite *unleash.Suite
	ApplicantID  string
	RootAccount  map[int]common.AuthInfo
}

func newSuite(c *common.Config) *Suite {
	s := &Suite{
		Connections:  connections,
		Cfg:          c,
		ZapLogger:    zapLogger,
		ApplicantID:  applicantID,
		CommonSuite:  &common.Suite{},
		UnleashSuite: &unleash.Suite{},
	}

	s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{}

	s.CommonSuite.StepState.FirebaseAddress = firebaseAddr
	s.CommonSuite.StepState.ApplicantID = applicantID

	// HARD CODE:
	s.CommonSuite.CurrentSchoolID = constants.ManabieSchool

	// Unleash
	s.UnleashSuite.Connections = s.Connections
	s.UnleashSuite.StepState = &common.StepState{}
	s.UnleashSuite.UnleashSrvAddr = c.UnleashSrvAddr
	s.UnleashSuite.UnleashAPIKey = c.UnleashAPIKey
	s.UnleashSuite.UnleashLocalAdminAPIKey = c.UnleashLocalAdminAPIKey

	s.RootAccount = rootAccount
	return s
}

func (s *Suite) enterASchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = constants.ManabieSchool
	ctx, err := s.signedAsAccountV2(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) someCenters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aListOfLocationTypesInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aListOfLocationsInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
