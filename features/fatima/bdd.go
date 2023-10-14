package fatima

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	cfg_bob "github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/fatima/configurations"
	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/yasuo/constant"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v4/pgxpool"
	natsJS "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	common.RegisterTest("fatima", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	conn         *grpc.ClientConn
	db           *pgxpool.Pool
	bobConn      *grpc.ClientConn
	shamirConn   *grpc.ClientConn
	bobDB        *pgxpool.Pool
	applicantID  string
	firebaseAddr string
	zapLogger    *zap.Logger
	jsm          nats.JetStreamManagement
	bobConfig    cfg_bob.Config
	fatimaConfig configurations.Config
	usermgmtConn *grpc.ClientConn
)

func TestSuiteInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c, f.FirebaseAddr)
		})

		ctx.AfterSuite(func() {
			db.Close()
			bobDB.Close()
			conn.Close()
			bobConn.Close()
			jsm.Close()
			usermgmtConn.Close()
		})
	}
}

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite()

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
			for _, v := range s.Subs {
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

type Suite struct {
	suite
}

func (s *suite) SetFirebaseAddr(fakefirebaseAddr string) {
	firebaseAddr = fakefirebaseAddr
}

type suite struct {
	DB      database.Ext
	BobConn *grpc.ClientConn
	Conn    *grpc.ClientConn
	BobDB   database.Ext
	*StepState

	EurekaDB  database.Ext
	ZapLogger *zap.Logger
	JSM       nats.JetStreamManagement

	ApplicantID  string
	ShamirConn   *grpc.ClientConn
	UsermgmtConn *grpc.ClientConn
}

type StepState struct {
	AuthToken        string
	Request          interface{}
	Response         interface{}
	ResponseErr      error
	RequestSentAt    time.Time
	Packages         map[string]*entities.Package
	Event            interface{}
	User             interface{}
	UserID           string
	StartAt          *timestamppb.Timestamp
	EndAt            *timestamppb.Timestamp
	CourseIDs        []string
	StudentID        string
	StudentPackageID string

	LocationIDs []string
	ClassIDs    []string

	MasterStudyPlanIDs []string

	// course reader
	NumberOfId int
	CourseID   string
	StudentIDs []string

	// migrate student subscriptions
	CurrentSchoolID int32

	// migrate student_packages to student_package_access_path
	StudentPackages []*entities.StudentPackage

	Subs                  []*natsJS.Subscription
	FoundChanForJetStream chan interface{}
}

func (s *suite) signedCtx(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", s.AuthToken)
}

func (s *suite) ReturnsStatusCode(arg1 string) error {
	return s.returnsStatusCode(arg1)
}

func (s *suite) returnsStatusCode(arg1 string) error {
	stt, ok := status.FromError(s.ResponseErr)

	if !ok {
		return fmt.Errorf("returned error is not status.Status, err: %s", s.ResponseErr.Error())
	}
	if stt.Code().String() != arg1 {
		return fmt.Errorf("expecting %s, got %s status code, message: %s", arg1, stt.Code().String(), stt.Message())
	}

	return nil
}

func (s *suite) GenerateValidAuthenticationToken(sub string) (string, error) {
	return s.generateValidAuthenticationToken(sub, "")
}

func (s *suite) generateValidAuthenticationToken(sub string, userGroup string) (string, error) {
	url := ""
	switch userGroup {
	case "USER_GROUP_TEACHER":
		url = "http://" + firebaseAddr + "/token?template=templates/USER_GROUP_TEACHER.template&UserID="
	case "USER_GROUP_SCHOOL_ADMIN":
		url = "http://" + firebaseAddr + "/token?template=templates/USER_GROUP_SCHOOL_ADMIN.template&UserID="
	case "USER_GROUP_ADMIN":
		url = "http://" + firebaseAddr + "/token?template=templates/USER_GROUP_ADMIN.template&UserID="
	default:
		url = "http://" + firebaseAddr + "/token?template=templates/phone.template&UserID="
	}

	resp, err := http.Get(url + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()
	return string(b), nil
}

func (s *suite) anInvalidAuthenticationToken(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = "invalid-token"
	return StepStateToContext(ctx, stepState)
}

func setup(c *common.Config, fakeFirebaseAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	rsc := bootstrap.NewResources().WithLoggerC(&c.Common)
	firebaseAddr = fakeFirebaseAddr

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	var err error
	conn = rsc.GRPCDial("fatima")

	db, _, _ = database.NewPool(context.Background(), zapLogger, c.PostgresV2.Databases["fatima"])

	bobConn = rsc.GRPCDial("bob")
	jsm, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to create jetstream management: %v", err))
	}

	jsm.ConnectToJS()

	bobDB, _, _ = database.NewPool(context.Background(), zapLogger, c.PostgresV2.Databases["bob"])
	shamirConn = rsc.GRPCDial("shamir")

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
	_, err = bobDB.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
	bobConfig = cfg_bob.Config{
		Common:     c.Common,
		PostgresV2: c.PostgresV2,
		NatsJS:     c.NatsJS,
	}
	fatimaConfig = configurations.Config{
		Common:       c.Common,
		PostgresV2:   c.PostgresV2,
		NatsJS:       c.NatsJS,
		Issuers:      c.Issuers,
		JWTApplicant: c.JWTApplicant,
	}
	usermgmtConn = rsc.GRPCDial("usermgmt")
}

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^an invalid authentication token$`:                   s.anInvalidAuthenticationToken,
		`^returns all CourseAccessibleResponse of this user$`: s.returnsAllCourseAccessibleResponseOfThisUser,
		`^returns "([^"]*)" status code$`:                     s.returnsStatusCode,
		`^some package data in db$`:                           s.somePackageDataInDb,
		`^this user has package "([^"]*)" is "([^"]*)"$`:      s.thisUserHasPackageIs,
		`^user retrieve accessible course$`:                   s.userRetrieveAccessibleCourse,
		`^a signed in "([^"]*)"$`:                             s.aSignedIn,

		`^server must store this package for this student$`: s.serverMustStoreThisPackageForThisStudent,
		`^user add a "([^"]*)" package for a student$`:      s.userAddAPackageForAStudent,

		`^an valid SyncStudentPackageEvent with ActionKind_ACTION_KIND_UPSERTED$`: s.aValidEvent_Upsert,
		`^an valid SyncStudentPackageEvent with ActionKind_ACTION_KIND_DELETED$`:  s.aValidEvent_Delete,
		`^bob send SyncStudentPackageEvent to nats$`:                              s.sendSyncStudentPackageEvent,
		`^our system must create StudentPackage data correctly$`:                  s.fatimaMustCreateStudentPackage,
		`^our system must update StudentPackage data correctly$`:                  s.fatimaMustUpdateStudentPackage,
		`^our system must createStudentPackage access path$`:                      s.fatimaSaveStudentPackageAccessPath,

		`^user add a package by courses for a student$`:                                                                s.userAddACourseForAStudent,
		`^user add a package by courses with student package extra for a student$`:                                     s.userAddCourseWithStudentPackageExtraForAStudent,
		`^server must store these courses for this student$`:                                                           s.serverMustStoreTheseCoursesForThisStudent,
		`^server must store these courses and class for this student$`:                                                 s.serverMustStoreTheseCoursesAndClassForThisStudent,
		`^user edit time a "([^"]*)" student package$`:                                                                 s.userEditTimeAStudentPackage,
		`^user edit time a "([^"]*)" student package with time from "([^"]*)" to "([^"]*)"$`:                           s.userEditTimeAStudentPackageWithTime,
		`^user edit time a "([^"]*)" student package with time from "([^"]*)" to "([^"]*)" and student package extra$`: s.userEditTimeAStudentPackageWithTimeAndStudentPackageExtra,
		`^server must store this student package with time from "([^"]*)" to "([^"]*)"$`:                               s.serverMustStoreThisStudentPackageWithTime,
		`^server must store this student package with time from "([^"]*)" to "([^"]*)" and class`:                      s.serverMustStoreThisStudentPackageWithTimeAndClass,

		`^a signed as <"([^"]*)">$`:                              s.aSignedAs,
		`^a signed as "([^"]*)"$`:                                s.aSignedAs,
		`^the user retrieve student accessible course$`:          s.theUserRetrieveStudentAccessibleCourse,
		`^a student has package "([^"]*)" is "([^"]*)"$`:         s.aStudentHasPackageIs,
		`^returns all CourseAccessibleResponse of this student$`: s.returnsAllCourseAccessibleResponseOfThisStudent,

		// `^some student packages data in db$`:                              s.someStudentPackagesDataInDB,
		`^user list student by course$`:                                   s.callListStudentByCourse,
		`^a list student by course valid request payload with "([^"]*)"$`: s.aListStudentByCourseValidRequestPayloadWith,
		`^fatima must return correct list of basic profile of students$`:  s.fatimaMustReturnCorrectListOfBasicProfile,

		// migrate student_packages to student_package_access_path
		`^a number of existing student packages$`:                                     s.aNumberOfExistingStudentPackages,
		`^system run job to migrate student_packages to student_package_access_path$`: s.systemRunJobToMigrateStudentPackagesToStudentPackageAccessPath,
		`^student_packages and student_package_access_path are correspondent$`:        s.studentPackagesAndStudentPackageAccessPathAreCorrespondent,
	}

	buildRegexpMapOnce.Do(func() {
		// nolint
		regexpMap = helper.BuildRegexpMap(steps)
	})

	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

func newSuite() *suite {
	return &suite{
		Conn:         conn,
		DB:           db,
		StepState:    &StepState{},
		BobConn:      bobConn,
		ZapLogger:    zapLogger,
		JSM:          jsm,
		ApplicantID:  applicantID,
		ShamirConn:   shamirConn,
		BobDB:        bobDB,
		UsermgmtConn: usermgmtConn,
	}
}

type stateKeyForFatima struct{}

func StepStateFromContext(ctx context.Context) *StepState {
	state := ctx.Value(stateKeyForFatima{})
	if state == nil {
		return &StepState{}
	}
	return state.(*StepState)
}

func StepStateToContext(ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, stateKeyForFatima{}, state)
}
