package eurekav2

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/features/unleash"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	course_domain "github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"

	firebase "firebase.google.com/go"
	"github.com/cucumber/godog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	common.RegisterTest("eurekav2", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	connections  *common.Connections
	zapLogger    *zap.Logger
	rootAccount  map[int]common.AuthInfo
	firebaseAddr string
	applicantID  string
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

func StepStateFromContext(ctx context.Context) *StepState {
	return ctx.Value(common.StepStateKey{}).(*StepState)
}

func StepStateToContext(ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, common.StepStateKey{}, state)
}

func ScenarioInitializer(c *common.Config, _ common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c)
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			ctx = StepStateToContext(ctx, s.StepState)

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
		common.WithShamirSvcAddress(),
		common.WithUserMgmtSvcAddress(),
		common.WithEurekaSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		common.WithEurekaDBConfig(c.PostgresV2.Databases["eureka"]),
		common.WithFatimaDBConfig(c.PostgresV2.Databases["fatima"]),
		common.WithAuthPostgresDBConfig(c.PostgresV2.Databases["auth"], c.PostgresMigrate.Database.Password),
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

	applicantID = c.JWTApplicant

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
	_, err = connections.BobDBTrace.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}

	rootAccount, err = usermgmt.InitRootAccount(ctx, connections.ShamirConn, firebaseAddr, c.JWTApplicant)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init rootAccount: %v", err))
	}
}

type suite struct {
	*common.Connections
	*StepState

	ZapLogger    *zap.Logger
	Cfg          *common.Config
	CommonSuite  *common.Suite
	UnleashSuite *unleash.Suite
	ApplicantID  string
}

func newSuite(c *common.Config) *suite {
	s := &suite{
		Connections:  connections,
		Cfg:          c,
		ZapLogger:    zapLogger,
		ApplicantID:  applicantID,
		CommonSuite:  &common.Suite{},
		UnleashSuite: &unleash.Suite{},
	}

	s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = &StepState{}

	s.CommonSuite.StepState.FirebaseAddress = firebaseAddr
	s.CommonSuite.StepState.ApplicantID = applicantID

	s.RootAccount = rootAccount

	// Unleash
	s.UnleashSuite.Connections = s.Connections
	s.UnleashSuite.StepState = &common.StepState{}
	s.UnleashSuite.UnleashSrvAddr = c.UnleashSrvAddr
	s.UnleashSuite.UnleashAPIKey = c.UnleashAPIKey
	s.UnleashSuite.UnleashLocalAdminAPIKey = c.UnleashLocalAdminAPIKey
	return s
}

func (s *suite) newID() string {
	return idutil.ULIDNow()
}

type StepState struct {
	Request     interface{}
	Response    interface{}
	ResponseErr error

	RootAccount map[int]common.AuthInfo
	AuthToken   string

	StudentID        string
	UserID           string
	LocationID       string
	TeacherID        string
	SchoolAdminToken string
	StudentToken     string
	TeacherToken     string

	// for BDD
	BookID         string
	BookIDs        []string
	UpdatedBookIDs []string
	Books          []domain.Book
	TopicIDs       []string
	ChapterIDs     []string
	BookContent    domain.Book

	LearningMaterialID  string
	LearningMaterialIDs []string

	LearningObjectiveIDs []string

	LearningMaterialIsPublished bool

	// for Courses
	CourseID         string
	CourseIDs        []string
	UpdatedCourseIDs []string
	Courses          []course_domain.Course

	StudyPlanID string
}

func contextWithToken(s *suite, ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return helper.GRPCContext(ctx, "token", stepState.AuthToken)
}
