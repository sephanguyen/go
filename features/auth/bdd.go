package auth

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/unleash"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/logger"

	firebase "firebase.google.com/go"
	"github.com/cucumber/godog"
	"github.com/vmihailenco/taskq/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	common.RegisterTest("auth", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	connections  *common.Connections
	zapLogger    *zap.Logger
	applicantID  string
	firebaseAddr string
	// mainQueue    taskq.Queue
	mapOrgUser map[int]common.MapRoleAndAuthInfo
	// unleashManager unleash_manager.Manager
	rootAccount map[int]common.AuthInfo
)

type suite struct {
	*common.Connections
	*common.StepState
	ZapLogger    *zap.Logger
	Cfg          *common.Config
	CommonSuite  *common.Suite
	TaskQueue    taskq.Queue
	UnleashSuite *unleash.Suite
	// UnleashManager unleash_manager.Manager
}

func newSuite(c *common.Config) *suite {
	s := &suite{
		Connections: connections,
		Cfg:         c,
		ZapLogger:   zapLogger,
		// ApplicantID:    applicantID,
		CommonSuite: &common.Suite{},
		// TaskQueue:      mainQueue,
		// UnleashSuite:   &unleash.Suite{},
		// UnleashManager: unleashManager,
	}

	s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = s.CommonSuite.StepState

	// s.UnleashSuite.Connections = s.Connections
	// s.UnleashSuite.StepState = &common.StepState{}
	// s.UnleashSuite.UnleashSrvAddr = c.UnleashSrvAddr
	// s.UnleashSuite.UnleashAPIKey = c.UnleashAPIKey
	// s.UnleashSuite.UnleashLocalAdminAPIKey = c.UnleashLocalAdminAPIKey

	s.CommonSuite.StepState.FirebaseAddress = firebaseAddr
	s.CommonSuite.StepState.ApplicantID = applicantID
	// s.CommonSuite.StepState.ExistingLocations = existingLocations
	// s.CommonSuite.StepState.LocationIDs = brandAndCenterLocationIDs
	// s.CommonSuite.StepState.LocationTypesID = brandAndCenterLocationTypeIDs
	s.CommonSuite.StepState.MapOrgStaff = mapOrgUser

	s.RootAccount = rootAccount
	return s
}

func setup(c *common.Config, fakeFirebaseAddress string) {
	ctx := context.Background()
	connections = &common.Connections{}
	firebaseAddr = fakeFirebaseAddress

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	err := connections.ConnectGRPC(ctx,
		common.WithCredentials(grpc.WithTransportCredentials(insecure.NewCredentials())),
		// common.WithBobSvcAddress(),
		common.WithShamirSvcAddress(),
		common.WithUserMgmtSvcAddress(),
		// common.WithMasterMgmtSvcAddress(),
		common.WithAuthSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		// common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
		common.WithMastermgmtDBConfig(c.PostgresV2.Databases["mastermgmt"]),
		// common.WithAuthDBConfig(c.PostgresV2.Databases["auth"]),
		common.WithAuthPostgresDBConfig(c.PostgresV2.Databases["auth"], c.PostgresMigrate.Database.Password),
	)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connectDB: %v", err))
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

	// if err := InitOrganizationTenantConfig(ctx, connections.BobPostgresDBTrace); err != nil {
	// 	zapLogger.Fatal(fmt.Sprintf("InitOrganizationTenantConfig: %v", err))
	// }

	// queueFactory := memqueue.NewFactory()
	// mainQueue = queueFactory.RegisterQueue(&taskq.QueueOptions{
	// 	Name: constants.UserMgmtTask,
	// })

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	connections.FirebaseClient, err = app.Auth(ctx)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create firebase client: %v", err))
	}

	// connections.JSM, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
	// 	c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	// if err != nil {
	// 	zapLogger.Panic(fmt.Sprintf("failed to create jetstream management: %v", err))
	// }
	// connections.JSM.ConnectToJS()

	connections.GCPApp, err = gcp.NewApp(ctx, "", c.Common.IdentityPlatformProject)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	// connections.FirebaseAuthClient, err = internal_auth_tenant.NewFirebaseAuthClientFromGCP(ctx, connections.GCPApp)
	// if err != nil {
	// 	zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	// }

	// secondaryTenantConfigProvider := &repository.TenantConfigRepo{
	// 	QueryExecer:      connections.BobPostgresDBTrace,
	// 	ConfigAESKey:     c.IdentityPlatform.ConfigAESKey,
	// 	ConfigAESIv:      c.IdentityPlatform.ConfigAESIv,
	// 	OrganizationRepo: &repository.OrganizationRepo{},
	// }

	// connections.TenantManager, err = internal_auth_tenant.NewTenantManagerFromGCP(ctx, connections.GCPApp, internal_auth_tenant.WithSecondaryTenantConfigProvider(secondaryTenantConfigProvider))
	// if err != nil {
	// 	zapLogger.Fatal(fmt.Sprintf("cannot create tenant manager: %v", err))
	// }

	// keycloakOpts := internal_auth.KeyCloakOpts{
	// 	Path:     "https://d2020-ji-sso.jprep.jp",
	// 	Realm:    "manabie-test",
	// 	ClientID: "manabie-app",
	// }

	// connections.KeycloakClient, err = internal_auth.NewKeyCloakClient(keycloakOpts)
	// if err != nil {
	// 	zapLogger.Fatal(fmt.Sprintf("cannot create keycloak client: %v", err))
	// }

	applicantID = c.JWTApplicant

	rootAccount, err = InitRootAccount(ctx, connections.ShamirConn, fakeFirebaseAddress, applicantID)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init root account: %v", err))
	}

	// existingLocations, err = PrepareLocations(connections.BobPostgresDBTrace.DB)
	// if err != nil {
	// 	zapLogger.Fatal(fmt.Sprintf("cannot seed locations: %v", err))
	// }

	// locationTypeIDs, locationIDs, err := prepairManabieBrandAndCenterLocations(ctx, connections.BobPostgresDB)
	// if err != nil {
	// 	zapLogger.Fatal(err.Error())
	// }
	// brandAndCenterLocationIDs = locationIDs
	// brandAndCenterLocationTypeIDs = locationTypeIDs

	mapOrgUser, err = InitUser(ctx, connections.AuthPostgresDB, applicantID)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init default user: %v", err))
	}
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

func ScenarioInitializer(c *common.Config, _ common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c)
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			// tagNames := make([]string, 0, len(sc.Tags))
			// for _, tag := range sc.Tags {
			// 	tagNames = append(tagNames, tag.Name)
			// }

			// featureFlagTags, err := usermgmt.ParseTags(tagNames...)
			// if err != nil {
			// 	return ctx, err
			// }

			// for _, featureFlagTag := range featureFlagTags {
			// 	if err := s.UnleashManager.Toggle(ctx, featureFlagTag.Name, featureFlagTag.ToggleChoice); err != nil {
			// 		return ctx, errors.Wrap(err, "s.UnleashManager.Toggle")
			// 	}
			// }

			return StepStateToContext(ctx, s.StepState), nil
		})

		ctx.After(func(ctx context.Context, sc *godog.Scenario, hookErr error) (context.Context, error) {
			// tagNames := make([]string, 0, len(sc.Tags))
			// for _, tag := range sc.Tags {
			// 	tagNames = append(tagNames, tag.Name)
			// }

			// featureFlagTags, err := usermgmt.ParseTags(tagNames...)
			// if err != nil {
			// 	return ctx, err
			// }
			// for _, featureFlagTag := range featureFlagTags {
			// 	if err := unleashManager.Unlock(ctx, featureFlagTag.Name); err != nil {
			// 		return ctx, err
			// 	}
			// }

			// if s.Cfg.Common.IdentityPlatformProject != "dev-manabie-online" {
			// 	return ctx, nil
			// }

			// if s.SrcTenant != nil {
			// 	_ = s.TenantManager.DeleteTenant(ctx, s.SrcTenant.GetID())
			// }
			// if s.DestTenant != nil {
			// 	_ = s.TenantManager.DeleteTenant(ctx, s.DestTenant.GetID())
			// }

			// stepState := StepStateFromContext(ctx)

			// for _, v := range stepState.Subs {
			// 	if v.IsValid() {
			// 		err := v.Drain()
			// 		if err != nil {
			// 			return ctx, fmt.Errorf("failed to drain subscription: %w", err)
			// 		}
			// 	}
			// }

			return ctx, nil
		})
	}
}
