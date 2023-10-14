package hephaestus

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"github.com/cucumber/godog"
	"github.com/go-kafka/connect"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	common.RegisterTest("hephaestus", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

type suite struct {
	*common.Connections
	*common.StepState
	ConnectClient                        *connect.Client
	ZapLogger                            *zap.Logger
	Cfg                                  *common.Config
	CommonSuite                          *common.Suite
	SourceConnectorDir, SinkConnectorDir string
	TableMetaData                        ITable
	SourceConnectorFileName              string
}

var (
	connections *common.Connections
	zapLogger   *zap.Logger
)

func TestSuiteInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c)
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
			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: fmt.Sprint(constants.ManabieSchool),
					DefaultRole:  entities.UserGroupAdmin,
					UserGroup:    entities.UserGroupAdmin,
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
			// clean up
			return ctx, nil
		})
	}
}

func setup(c *common.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	connections = &common.Connections{}

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
		common.WithInvoiceMgmtSvcAddress(),
		common.WithLessonMgmtSvcAddress(),
		common.WithMasterMgmtSvcAddress(),
		common.WithTimesheetSvcAddress(),
		common.WithCalendarSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
		common.WithInvoiceMgmtDBConfig(c.PostgresV2.Databases["invoicemgmt"]),
		common.WithFatimaDBConfig(c.PostgresV2.Databases["fatima"]),
		common.WithTomDBConfig(c.PostgresV2.Databases["tom"]),
		common.WithEurekaDBConfig(c.PostgresV2.Databases["eureka"]),
		common.WithFatimaDBConfig(c.PostgresV2.Databases["fatima"]),
		common.WithZeusDBConfig(c.PostgresV2.Databases["zeus"]),
		common.WithTimesheetPostgresDBConfig(c.PostgresV2.Databases["timesheet"], c.PostgresMigrate.Database.Password),
		common.WithCalendarDBConfig(c.PostgresV2.Databases["calendar"]),
		common.WithLessonmgmtDBConfig(c.PostgresV2.Databases["lessonmgmt"]),
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
}

func newSuite(c *common.Config) *suite {
	s := &suite{
		Connections:   connections,
		Cfg:           c,
		ZapLogger:     zapLogger,
		CommonSuite:   &common.Suite{},
		ConnectClient: connect.NewClient(c.KafkaConnectConfig.Addr),
	}

	s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = s.CommonSuite.StepState

	s.SourceConnectorDir = "/connectors/source"
	s.SinkConnectorDir = "/connectors/sink"

	_ = os.MkdirAll(s.SourceConnectorDir, 0o775)
	_ = os.MkdirAll(s.SinkConnectorDir, 0o775)

	// HARD CODE:
	s.CommonSuite.CurrentSchoolID = constants.ManabieSchool

	return s
}
