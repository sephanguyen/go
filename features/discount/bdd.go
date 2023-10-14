package discount

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/usermgmt"
	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	firebase "firebase.google.com/go"
	"github.com/cucumber/godog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	common.RegisterTest("discount", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	connections  *common.Connections
	zapLogger    *zap.Logger
	firebaseAddr string
	applicantID  string
	rootAccount  map[int]common.AuthInfo // map org_id with auth info of admin
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

func ScenarioInitializer(c *common.Config, _ common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c)
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
					DefaultRole:  constant.UserGroupAdmin,
					UserGroup:    constant.UserGroupAdmin,
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
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
		common.WithCredentials(grpc.WithTransportCredentials(insecure.NewCredentials())),
		common.WithBobSvcAddress(),
		common.WithTomSvcAddress(),
		common.WithEurekaSvcAddress(),
		common.WithFatimaSvcAddress(),
		common.WithShamirSvcAddress(),
		common.WithYasuoSvcAddress(),
		common.WithUserMgmtSvcAddress(),
		common.WithPaymentSvcAddress(),
		common.WithDiscountSvcAddress(),
		common.WithNotificationMgmtSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		common.WithTomDBConfig(c.PostgresV2.Databases["tom"]),
		common.WithEurekaDBConfig(c.PostgresV2.Databases["eureka"]),
		common.WithFatimaDBConfig(c.PostgresV2.Databases["fatima"]),
		common.WithZeusDBConfig(c.PostgresV2.Databases["zeus"]),
		common.WithAuthPostgresDBConfig(c.PostgresV2.Databases["auth"], c.PostgresMigrate.Database.Password),
		common.WithNotificationmgmtDBConfig(c.PostgresV2.Databases["notificationmgmt"]),
		common.WithNotificationmgmtPostgresDBConfig(c.PostgresV2.Databases["notificationmgmt"], c.PostgresMigrate.Database.Password),
	)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connectDB: %v", err))
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

	connections.Kafka, err = kafka.NewKafkaManagement(c.KafkaCluster.Address, c.KafkaCluster.IsLocal, c.KafkaCluster.ObjectNamePrefix, zapLogger)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to create kafka management: %v", err))
	}
	connections.Kafka.ConnectToKafka()

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
		zapLogger.Fatal(fmt.Sprintf("cannot init root account: %v", err))
	}
}

type suite struct {
	*common.Connections
	*common.StepState
	ZapLogger   *zap.Logger
	Cfg         *common.Config
	CommonSuite *common.Suite
	ApplicantID string
}

func newSuite(c *common.Config) *suite {
	s := &suite{
		Connections: connections,
		Cfg:         c,
		ZapLogger:   zapLogger,
		ApplicantID: applicantID,
		CommonSuite: &common.Suite{},
	}

	s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = s.CommonSuite.StepState

	s.CommonSuite.StepState.FirebaseAddress = firebaseAddr
	s.CommonSuite.StepState.ApplicantID = applicantID
	s.RootAccount = rootAccount

	// initialize map step state variables
	s.CommonSuite.DiscountTagTypeAndIDMap = make(map[string]string)

	return s
}

func ContextWithJWTClaims(ctx context.Context) context.Context {
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.ManabieSchool),
			DefaultRole:  UserGroupAdmin,
			UserGroup:    UserGroupAdmin,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
	return ctx
}
func checkCSVHeaderForExport(expected []string, actual []string) (err error) {
	if len(expected) != len(actual) {
		err = fmt.Errorf("expected header length to be %d got %d", len(expected), len(actual))
		return
	}

	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			err = fmt.Errorf("expected header name to be %s got %s", expected[i], actual[i])
			return
		}
	}

	return nil
}
