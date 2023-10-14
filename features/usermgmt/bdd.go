package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/features/unleash"
	unleash_manager "github.com/manabie-com/backend/features/usermgmt/unleash"
	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	firebase "firebase.google.com/go"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/vmihailenco/taskq/v3"
	"github.com/vmihailenco/taskq/v3/memqueue"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	common.RegisterTest("usermgmt", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	connections       *common.Connections
	existingLocations []*location_repo.Location
	zapLogger         *zap.Logger
	firebaseAddr      string
	applicantID       string
	mainQueue         taskq.Queue
	mapOrgUser        map[int]common.MapRoleAndAuthInfo
	unleashManager    unleash_manager.Manager
	rootAccount       map[int]common.AuthInfo

	// brandAndCenterLocationIDs [0]: brand location id, [1...]: center location id
	brandAndCenterLocationIDs []string
	// brandAndCenterLocationTypeIDs [0]: brand location type id, [1...]: center location type id
	brandAndCenterLocationTypeIDs []string
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

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c)
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			tagNames := make([]string, 0, len(sc.Tags))
			for _, tag := range sc.Tags {
				tagNames = append(tagNames, tag.Name)
			}

			featureFlagTags, err := ParseTags(tagNames...)
			if err != nil {
				return ctx, err
			}

			for _, featureFlagTag := range featureFlagTags {
				if err := s.UnleashManager.Toggle(ctx, featureFlagTag.Name, featureFlagTag.ToggleChoice); err != nil {
					return ctx, errors.Wrap(err, "s.UnleashManager.Toggle")
				}
			}

			return StepStateToContext(ctx, s.StepState), nil
		})

		ctx.After(func(ctx context.Context, sc *godog.Scenario, hookErr error) (context.Context, error) {
			tagNames := make([]string, 0, len(sc.Tags))
			for _, tag := range sc.Tags {
				tagNames = append(tagNames, tag.Name)
			}

			featureFlagTags, err := ParseTags(tagNames...)
			if err != nil {
				return ctx, err
			}
			for _, featureFlagTag := range featureFlagTags {
				if err := unleashManager.Unlock(ctx, featureFlagTag.Name); err != nil {
					return ctx, err
				}
			}

			if s.Cfg.Common.IdentityPlatformProject != "dev-manabie-online" {
				return ctx, nil
			}

			if s.SrcTenant != nil {
				_ = s.TenantManager.DeleteTenant(ctx, s.SrcTenant.GetID())
			}
			if s.DestTenant != nil {
				_ = s.TenantManager.DeleteTenant(ctx, s.DestTenant.GetID())
			}

			stepState := StepStateFromContext(ctx)

			if len(stepState.FirebaseResourceIDs) > 0 {
				_ = removeUserInFireBase(s.FirebaseClient, stepState.FirebaseResourceIDs)
			}

			for _, v := range stepState.Subs {
				if v.IsValid() {
					err := v.Drain()
					if err != nil {
						return ctx, fmt.Errorf("failed to drain subscription: %w", err)
					}
				}
			}

			return ctx, nil
		})
	}
}

func setup(c *common.Config, fakeFirebaseAddr string) {
	ctx := context.Background()

	unleashManager = unleash_manager.NewManager(c.UnleashSrvAddr, c.UnleashAPIKey, c.UnleashLocalAdminAPIKey)

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
		common.WithMasterMgmtSvcAddress(),
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
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
		common.WithMastermgmtDBConfig(c.PostgresV2.Databases["mastermgmt"]),
		// common.WithAuthDBConfig(c.PostgresV2.Databases["auth"]),
		common.WithAuthPostgresDBConfig(c.PostgresV2.Databases["auth"], c.PostgresMigrate.Database.Password),
		common.WithNotificationmgmtPostgresDBConfig(c.PostgresV2.Databases["notificationmgmt"], c.PostgresMigrate.Database.Password),
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
	_, err = connections.BobPostgresDBTrace.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}

	if err := InitOrganizationTenantConfig(ctx, connections.BobPostgresDBTrace); err != nil {
		zapLogger.Fatal(fmt.Sprintf("InitOrganizationTenantConfig: %v", err))
	}

	queueFactory := memqueue.NewFactory()
	mainQueue = queueFactory.RegisterQueue(&taskq.QueueOptions{
		Name: constants.UserMgmtTask,
	})

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

	secondaryTenantConfigProvider := &repository.TenantConfigRepo{
		QueryExecer:      connections.BobPostgresDBTrace,
		ConfigAESKey:     c.IdentityPlatform.ConfigAESKey,
		ConfigAESIv:      c.IdentityPlatform.ConfigAESIv,
		OrganizationRepo: &repository.OrganizationRepo{},
	}

	connections.TenantManager, err = internal_auth_tenant.NewTenantManagerFromGCP(ctx, connections.GCPApp, internal_auth_tenant.WithSecondaryTenantConfigProvider(secondaryTenantConfigProvider))
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

	rootAccount, err = InitRootAccount(ctx, connections.ShamirConn, fakeFirebaseAddr, applicantID)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init root account: %v", err))
	}

	existingLocations, err = PrepareLocations(connections.BobPostgresDBTrace.DB)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot seed locations: %v", err))
	}

	locationTypeIDs, locationIDs, err := prepairManabieBrandAndCenterLocations(ctx, connections.BobPostgresDB)
	if err != nil {
		zapLogger.Fatal(err.Error())
	}
	brandAndCenterLocationIDs = locationIDs
	brandAndCenterLocationTypeIDs = locationTypeIDs

	mapOrgUser, err = InitUser(ctx, connections.BobPostgresDB, connections.UserMgmtConn, connections.TenantManager, applicantID, c.FirebaseAPIKey)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init default user: %v", err))
	}
}

type suite struct {
	*common.Connections
	*common.StepState
	ZapLogger      *zap.Logger
	Cfg            *common.Config
	CommonSuite    *common.Suite
	ApplicantID    string
	TaskQueue      taskq.Queue
	UnleashSuite   *unleash.Suite
	UnleashManager unleash_manager.Manager
}

func newSuite(c *common.Config) *suite {
	s := &suite{
		Connections:    connections,
		Cfg:            c,
		ZapLogger:      zapLogger,
		ApplicantID:    applicantID,
		CommonSuite:    &common.Suite{},
		TaskQueue:      mainQueue,
		UnleashSuite:   &unleash.Suite{},
		UnleashManager: unleashManager,
	}

	s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = s.CommonSuite.StepState

	s.UnleashSuite.Connections = s.Connections
	s.UnleashSuite.StepState = &common.StepState{}
	s.UnleashSuite.UnleashSrvAddr = c.UnleashSrvAddr
	s.UnleashSuite.UnleashAPIKey = c.UnleashAPIKey
	s.UnleashSuite.UnleashLocalAdminAPIKey = c.UnleashLocalAdminAPIKey

	s.CommonSuite.StepState.FirebaseAddress = firebaseAddr
	s.CommonSuite.StepState.ApplicantID = applicantID
	s.CommonSuite.StepState.ExistingLocations = existingLocations
	s.CommonSuite.StepState.LocationIDs = brandAndCenterLocationIDs
	s.CommonSuite.StepState.LocationTypesID = brandAndCenterLocationTypeIDs
	s.CommonSuite.StepState.MapOrgStaff = mapOrgUser

	s.RootAccount = rootAccount
	return s
}

func (s *suite) generateExchangeToken(userID, userGroup string) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", fmt.Errorf("error when create generateValidAuthenticationToken: %v", err)
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.ApplicantID, constants.ManabieSchool, s.ShamirConn, helper.NewAuthUserListener(context.Background(), s.AuthPostgresDB))
	if err != nil {
		return "", fmt.Errorf("error when create exchange token: %v", err)
	}
	return token, nil
}

func (s *suite) aValidStudentWithSchoolID(ctx context.Context, id string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	sql := "UPDATE students SET school_id = $1 WHERE student_id = $2"
	_, err := s.BobDBTrace.Exec(ctx, sql, &schoolID, &id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidTeacherProfileWithID(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	teacher := entity.Teacher{}
	database.AllNullEntity(&teacher.LegacyUser)
	database.AllNullEntity(&teacher)
	teacher.ID.Set(id)
	var schoolIDs []int32
	if len(stepState.Schools) > 0 {
		schoolIDs = []int32{stepState.Schools[0].ID.Int}
	}
	if schoolID != 0 {
		schoolIDs = append(schoolIDs, schoolID)
	}
	teacher.SchoolIDs.Set(schoolIDs)
	now := time.Now()
	if err := teacher.UpdatedAt.Set(now); err != nil {
		return nil, err
	}
	if err := teacher.CreatedAt.Set(now); err != nil {
		return nil, err
	}
	user, err := newUserEntity()
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("newUserEntity: %w", err).Error())
	}
	err = multierr.Combine(
		user.ID.Set(teacher.ID),
		user.Group.Set(entity.UserGroupTeacher),
	)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
	}
	userGroup := entity.UserGroup{}
	database.AllNullEntity(&userGroup)
	err = multierr.Combine(
		userGroup.UserID.Set(teacher.ID),
		userGroup.GroupID.Set(database.Text(cpb.UserGroup_USER_GROUP_TEACHER.String())),
		userGroup.IsOrigin.Set(database.Bool(true)),
		userGroup.Status.Set(entity.UserGroupStatusActive),
		userGroup.CreatedAt.Set(user.CreatedAt),
		userGroup.UpdatedAt.Set(user.UpdatedAt),
	)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
	}

	_, err = database.InsertExcept(internal_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), user, []string{"resource_path"}, s.BobDBTrace.Exec)
	if err != nil {
		return ctx, err
	}

	_, err = database.InsertExcept(internal_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), &teacher, []string{"resource_path"}, s.BobDBTrace.Exec)
	if err != nil {
		return ctx, err
	}

	cmdTag, err := database.InsertExcept(internal_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), &userGroup, []string{"resource_path"}, s.BobDBTrace.Exec)
	if err != nil {
		return ctx, err
	}

	if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert teacher for testing")
	}
	return ctx, nil
}

func (s *suite) aSignedInAdminWithProfileId(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	user, err := newUserEntity()
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("newUserEntity: %w", err).Error())
	}
	err = multierr.Combine(
		user.ID.Set(id),
		user.Group.Set(entity.UserGroupAdmin),
	)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
	}

	userRepo := repository.UserRepo{}
	err = userRepo.Create(internal_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidSchoolAdminProfileWithId(ctx context.Context, id, userGroup string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	if userGroup == "" {
		userGroup = entity.UserGroupSchoolAdmin
	}

	schoolAdmin := entity.SchoolAdmin{}
	database.AllNullEntity(&schoolAdmin)

	err := multierr.Combine(
		schoolAdmin.SchoolAdminID.Set(id),
		schoolAdmin.SchoolID.Set(schoolID),
		schoolAdmin.ResourcePath.Set(fmt.Sprint(schoolID)),
		schoolAdmin.UpdatedAt.Set(now),
		schoolAdmin.CreatedAt.Set(now),
	)
	if err != nil {
		return nil, err
	}

	user, err := newUserEntity()
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("newUserEntity: %w", err).Error())
	}
	if err := multierr.Combine(
		user.ID.Set(schoolAdmin.SchoolAdminID),
		user.Group.Set(userGroup),
		user.ResourcePath.Set(fmt.Sprint(schoolID)),
	); err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
	}

	userRepo := repository.UserRepo{}
	err = userRepo.Create(internal_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	schoolAdminRepo := repository.SchoolAdminRepo{}
	err = schoolAdminRepo.CreateMultiple(internal_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, []*entity.SchoolAdmin{&schoolAdmin})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userGroupEnt := entity.UserGroup{}
	database.AllNullEntity(&userGroupEnt)
	err = multierr.Combine(
		userGroupEnt.UserID.Set(id),
		userGroupEnt.GroupID.Set(userGroup),
		userGroupEnt.UpdatedAt.Set(now),
		userGroupEnt.CreatedAt.Set(now),
		userGroupEnt.IsOrigin.Set(true),
		userGroupEnt.Status.Set(entity.UserGroupStatusActive),
		userGroupEnt.ResourcePath.Set(fmt.Sprint(schoolID)),
	)
	if err != nil {
		return ctx, status.Errorf(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
	}

	userGroupRepo := repository.UserGroupRepo{}
	err = userGroupRepo.Upsert(internal_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, &userGroupEnt)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInSchoolAdminWithSchoolID(ctx context.Context, group string, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := newID()
	ctx, err := s.aValidSchoolAdminProfileWithId(ctx, id, group, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, group)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = id
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aUserSignedInAsAParentWithSchoolID(ctx context.Context, _ int) (context.Context, error) {
	return s.aUserSignedInAsAParent(ctx)
}

func (s *suite) aUserSignedInAsAParent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, parentType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInStudentWithSchool(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentStudentID = ""
	id := newID()
	ctx, err := s.aValidStudentInDB(ctx, id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aValidStudentWithSchoolID(ctx, id, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, entity.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInTeacherWithSchoolID(ctx context.Context, schoolID int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := newID()
	ctx, err := s.aValidTeacherProfileWithID(ctx, id, int32(schoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentTeacherID = id
	stepState.CurrentUserID = id

	token, err := s.generateExchangeToken(id, entity.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ToggleUnleashFeatureWithName(ctx context.Context, toggleChoice string, featureName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := s.UnleashManager.Toggle(ctx, featureName, unleash_manager.ToggleChoice(toggleChoice))
	return StepStateToContext(ctx, stepState), err
	// stepState := StepStateFromContext(ctx)
	// /*fmt.Printf("%p \n", s.UnleashManager)

	// if locked := stepState.FeatureFlagLockStatus[featureName]; !locked {
	// 	return StepStateToContext(ctx, stepState), s.UnleashManager.Toggle(ctx, featureName, unleash_manager.ToggleChoice(toggleChoice))
	// }
	// return StepStateToContext(ctx, stepState), nil*/

	// return StepStateToContext(ctx, stepState), s.UnleashManager.Toggle(ctx, featureName, unleash_manager.ToggleChoice(toggleChoice))
}
