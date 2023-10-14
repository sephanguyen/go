package accesscontrol

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	bob_repository "github.com/manabie-com/backend/internal/bob/repositories"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	firebase "firebase.google.com/go"
	"github.com/cucumber/godog"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	common.RegisterTest("accesscontrol", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	connections  *common.Connections
	zapLogger    *zap.Logger
	firebaseAddr string
	applicantID  string
)

const (
	insertCommand string = "insert"
	updateCommand string = "update"
	deleteCommand string = "delete"
)

func setup(c *common.Config, fakeFirebaseAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	connections = &common.Connections{}

	firebaseAddr = fakeFirebaseAddr

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	err := connections.ConnectGRPC(ctx,
		common.WithCredentials(grpc.WithInsecure()),
		common.WithMasterMgmtSvcAddress(),
		common.WithShamirSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}
	err = connections.ConnectDB(
		ctx,
		common.WithMastermgmtPostgresDBConfig(c.PostgresV2.Databases["mastermgmt"], c.PostgresMigrate.Database.Password),
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
	)

	const cantConnectFirebase = "cannot connect to firebase: %v"

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connectDB: %v", err))
	}

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf(cantConnectFirebase, err))
	}
	connections.FirebaseClient, err = app.Auth(ctx)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create firebase client: %v", err))
	}

	connections.GCPApp, err = gcp.NewApp(ctx, "", c.Common.IdentityPlatformProject)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf(cantConnectFirebase, err))
	}
	connections.FirebaseAuthClient, err = internal_auth_tenant.NewFirebaseAuthClientFromGCP(ctx, connections.GCPApp)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf(cantConnectFirebase, err))
	}
	connections.TenantManager, err = internal_auth_tenant.NewTenantManagerFromGCP(ctx, connections.GCPApp)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create tenant manager: %v", err))
	}

	applicantID = c.JWTApplicant

	// Init auth info
	stmt := `
		INSERT INTO organization_auths
			(organization_id, auth_project_id, auth_tenant_id)
			values
			($1, 'fake_aud', '')
		ON CONFLICT
			DO NOTHING
		;
		`
	_, err = connections.BobDB.Exec(ctx, stmt, testResourcePath)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
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
			ctx = StepStateToContext(ctx, s.StepState)
			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: testResourcePath,
					DefaultRole:  entities.UserGroupAdmin,
					UserGroup:    entities.UserGroupAdmin,
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, &claim)

			return ctx, nil
		})

		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			stepState := common.StepStateFromContext(ctx)
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
		Cfg:         c,
		ZapLogger:   zapLogger,
		ApplicantID: applicantID,
		CommonSuite: &common.Suite{},
		Connections: connections,
	}

	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = s.CommonSuite.StepState
	s.CommonSuite.StepState.FirebaseAddress = firebaseAddr
	s.CommonSuite.StepState.ApplicantID = applicantID
	s.CommonSuite.Connections = s.Connections

	return s
}

func (s *suite) newID() string {
	return idutil.ULIDNow()
}

type userOption func(u *entities_bob.User)

func withID(id string) userOption {
	return func(u *entities_bob.User) {
		_ = u.ID.Set(id)
	}
}

func withRole(group string) userOption {
	return func(u *entities_bob.User) {
		_ = u.Group.Set(group)
	}
}

func (s *suite) aValidUser(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	num := s.newID()

	userRepo := bob_repository.UserRepo{}
	u := &entities.User{}
	database.AllNullEntity(u)

	err := multierr.Combine(
		u.ID.Set(s.newID()),
		u.LastName.Set(fmt.Sprintf("valid-user-%s", num)),
		u.PhoneNumber.Set(fmt.Sprintf("+848%s", num)),
		u.Email.Set(fmt.Sprintf("valid-user-%s@email.com", num)),
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%s", num)),
		u.Country.Set(pb.COUNTRY_VN.String()),
		u.Group.Set(entities.UserGroupStudent),
		u.DeviceToken.Set(nil),
		u.AllowNotification.Set(true),
		u.CreatedAt.Set(time.Now()),
		u.UpdatedAt.Set(time.Now()),
		u.IsTester.Set(nil),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, opt := range opts {
		opt(u)
	}

	err = userRepo.Create(ctx, s.Connections.BobDB, u)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser insert user err: %w", err)
	}

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = 1
	}
	if u.Group.String == entities.UserGroupTeacher {
		teacher := &entities.Teacher{}
		database.AllNullEntity(teacher)

		err = multierr.Combine(
			teacher.ID.Set(u.ID.String),
			teacher.SchoolIDs.Set([]int64{schoolID}),
			teacher.UpdatedAt.Set(time.Now()),
			teacher.CreatedAt.Set(time.Now()),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(ctx, teacher, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser insert teacher error: %w", err)
		}
	}

	if u.Group.String == constant.UserGroupSchoolAdmin {
		schoolAdminAccount := &entities.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		err := multierr.Combine(
			schoolAdminAccount.SchoolAdminID.Set(u.ID.String),
			schoolAdminAccount.SchoolID.Set(schoolID),
			schoolAdminAccount.UpdatedAt.Set(time.Now()),
			schoolAdminAccount.CreatedAt.Set(time.Now()),
			schoolAdminAccount.ResourcePath.Set("1"),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(ctx, schoolAdminAccount, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser insert school error: %w", err)
		}
	}

	ug := entities.UserGroup{}
	database.AllNullEntity(&ug)

	now := time.Now()
	ug.UserID.Set(u.ID.String)
	ug.GroupID.Set(u.Group.String)
	ug.UpdatedAt.Set(now)
	ug.CreatedAt.Set(now)
	ug.IsOrigin.Set(true)
	ug.Status.Set(entities.UserGroupStatusActive)

	userGroupRepo := bob_repository.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.BobDB, &ug)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
