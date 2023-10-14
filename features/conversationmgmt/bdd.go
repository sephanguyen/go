package conversationmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	conv_common "github.com/manabie-com/backend/features/conversationmgmt/common"
	"github.com/manabie-com/backend/features/conversationmgmt/common/helpers"
	"github.com/manabie-com/backend/internal/bob/entities"
	conv_config "github.com/manabie-com/backend/internal/conversationmgmt/configurations"
	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"

	firebase "firebase.google.com/go"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	common.RegisterTest("conversationmgmt", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	zapLogger              *zap.Logger
	firebaseAddr           string
	applicantID            string
	searchClient           *elastic.SearchFactoryImpl
	connections            *common.Connections
	conversationmgmtConfig conv_config.Config
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
	state := ctx.Value(common.StepStateKey{})
	if state == nil {
		return &common.StepState{}
	}
	return state.(*common.StepState)
}

func StepStateToContext(ctx context.Context, state *common.StepState) context.Context {
	return context.WithValue(ctx, common.StepStateKey{}, state)
}

func ScenarioInitializer(conf *common.Config, _ common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		mapFeaturesToStepFuncs(ctx, conf)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
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
			return StepStateToContext(ctx, stepState), nil
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
		common.WithNotificationMgmtSvcAddress(),
		common.WithMasterMgmtSvcAddress(),
		common.WithSpikeSvcAddress(),
		common.WithConversationMgmtSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
		common.WithFatimaDBConfig(c.PostgresV2.Databases["fatima"]),
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
		common.WithTomDBConfig(c.PostgresV2.Databases["tom"]),
		common.WithEurekaDBConfig(c.PostgresV2.Databases["eureka"]),
		common.WithZeusDBConfig(c.PostgresV2.Databases["zeus"]),
		common.WithTomPostgresDBConfig(c.PostgresV2.Databases["tom"], c.PostgresMigrate.Database.Password),
		common.WithNotificationmgmtDBConfig(c.PostgresV2.Databases["notificationmgmt"]),
		common.WithNotificationmgmtPostgresDBConfig(c.PostgresV2.Databases["notificationmgmt"], c.PostgresMigrate.Database.Password),
		common.WithAuthPostgresDBConfig(c.PostgresV2.Databases["auth"], c.PostgresMigrate.Database.Password),
	)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connectDB: %v", err))
	}

	err = updateResourcePath(connections.BobPostgresDB)
	if err != nil {
		zapLogger.Fatal("failed to update database", zap.Error(err))
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

	connections.Kafka, err = kafka.NewKafkaManagement(c.KafkaCluster.Address, c.KafkaCluster.IsLocal, c.KafkaCluster.ObjectNamePrefix, zapLogger)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to create kafka management: %v", err))
	}
	connections.Kafka.ConnectToKafka()

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

	_, err = connections.BobPostgresDB.Exec(ctx, stmt)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}

	conversationmgmtConfig = conv_config.Config{
		Common:     c.Common,
		PostgresV2: c.PostgresV2,
	}
}

func updateResourcePath(db *pgxpool.Pool) error {
	ctx := context.Background()
	query := `UPDATE school_configs SET resource_path = '1';
	UPDATE schools SET resource_path = '1';
	UPDATE configs SET resource_path = '1';
	UPDATE cities SET resource_path = '1';
	UPDATE districts SET resource_path = '1';`
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "1",
			DefaultRole:  entities.UserGroupAdmin,
			UserGroup:    entities.UserGroupAdmin,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
	_, err := db.Exec(ctx, query)
	return err
}

func initConversationMgmtCommonState(ctx context.Context) context.Context {
	return conv_common.StepStateToContext(ctx, &conv_common.StepState{
		MapCourseIDAndStudentIDs: make(map[string][]string),
		MapStudentIDAndParentIDs: make(map[string][]string),
	})
}

func newConversationMgmtCommonSuite(cfg *common.Config) *conv_common.ConversationMgmtSuite {
	csuite := &conv_common.ConversationMgmtSuite{}
	csuite.StepState = &conv_common.StepState{
		MapCourseIDAndStudentIDs: make(map[string][]string),
		MapStudentIDAndParentIDs: make(map[string][]string),
	}
	csuite.ConversationMgmtHelper = helpers.NewConversationMgmtHelper(firebaseAddr, applicantID, connections, cfg)

	return csuite
}
