package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	ec "github.com/manabie-com/backend/features/invoicemgmt/entities_creator"
	"github.com/manabie-com/backend/features/unleash"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"

	"github.com/cucumber/godog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	common.RegisterTest("invoicemgmt", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	connections  *common.Connections
	zapLogger    *zap.Logger
	firebaseAddr string
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

		resourcePath := fmt.Sprintf("%d", int32(-2147483634))

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: resourcePath,
					DefaultRole:  entities.UserGroupSchoolAdmin,
					UserGroup:    entities.UserGroupSchoolAdmin,
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
			s.StepState.CurrentSchoolID = -2147483634
			s.StepState.ResourcePath = resourcePath

			initLocation(s)

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

func initLocation(s *suite) {
	// Use the existing default location
	s.StepState.LocationID = "01FR4M51XJY9E77GSN4QZ1Q8N5"
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
		common.WithPaymentSvcAddress(),
		common.WithInvoiceMgmtSvcAddress(),
		common.WithMasterMgmtSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}
	bobPostgresDB := c.PostgresV2.Databases["bob"]
	bobPostgresDB.User = "postgres"
	bobPostgresDB.Password = c.PostgresMigrate.Database.Password

	invoicemgmtDB := c.PostgresV2.Databases["invoicemgmt"]
	invoicemgmtDB.User = "invoicemgmt"
	invoicemgmtDB.Password = c.PostgresMigrate.Database.Password

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(bobPostgresDB),
		common.WithInvoiceMgmtDBConfig(invoicemgmtDB),
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
		common.WithInvoiceMgmtPostgresDBConfig(c.PostgresV2.Databases["invoicemgmt"], c.PostgresMigrate.Database.Password),
		common.WithFatimaDBConfig(c.PostgresV2.Databases["fatima"]),
		common.WithMastermgmtDBConfig(c.PostgresV2.Databases["mastermgmt"]),
		common.WithAuthPostgresDBConfig(c.PostgresV2.Databases["auth"], c.PostgresMigrate.Database.Password),
		// common.WithAuthDBConfig(c.PostgresV2.Databases["auth"]),
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
	ZapLogger       *zap.Logger
	Cfg             *common.Config
	CommonSuite     *common.Suite
	ApplicantID     string
	EntitiesCreator *ec.EntitiesCreator
	UnleashSuite    *unleash.Suite
	MinIOClient     *filestorage.MinIOStorageService
}

func newSuite(c *common.Config) *suite {
	s := &suite{
		Connections:     connections,
		Cfg:             c,
		ZapLogger:       zapLogger,
		CommonSuite:     &common.Suite{},
		EntitiesCreator: ec.NewEntitiesCreator(),
		UnleashSuite:    &unleash.Suite{},
	}

	s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = s.CommonSuite.StepState

	s.CommonSuite.StepState.FirebaseAddress = firebaseAddr
	s.CommonSuite.StudentBillItemMap = make(map[string][]int32)
	s.CommonSuite.InvoiceStudentMap = make(map[string]string)
	s.CommonSuite.StudentInvoiceTotalMap = make(map[string]int64)
	s.CommonSuite.InvoiceIDInvoiceReferenceMap = make(map[string]string)
	s.CommonSuite.InvoiceIDInvoiceReference2Map = make(map[string]string)
	s.CommonSuite.InvoiceIDInvoiceTotalMap = make(map[string]float64)
	s.CommonSuite.StudentInvoiceReferenceMap = make(map[string]string)
	s.CommonSuite.StudentInvoiceReference2Map = make(map[string]string)
	s.CommonSuite.StudentBillItemTotalPrice = make(map[string]float64)
	s.CommonSuite.StudentPaymentMethodMap = make(map[string]string)
	s.CommonSuite.PaymentStatusIDsMap = make(map[string][]string)

	// Init org maps
	s.StepState.OrganizationInvoiceHistoryMap = make(map[string]string)
	s.StepState.OrganizationStudentListMap = make(map[string][]string)
	s.StepState.OrganizationStudentNumberMap = make(map[string]int)

	// Unleash
	s.UnleashSuite.Connections = s.Connections
	s.UnleashSuite.StepState = &common.StepState{}
	s.UnleashSuite.UnleashSrvAddr = c.UnleashSrvAddr
	s.UnleashSuite.UnleashAPIKey = c.UnleashAPIKey
	s.UnleashSuite.UnleashLocalAdminAPIKey = c.UnleashLocalAdminAPIKey

	s.MinIOClient, _ = filestorage.NewMinIOStorageServiceService(&s.Cfg.Storage)

	return s
}
