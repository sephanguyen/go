package jprep

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/manabie-com/backend/features/gandalf"
	"github.com/manabie-com/backend/features/yasuo"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	bobConn               *grpc.ClientConn
	tomConn               *grpc.ClientConn
	yasuoConn             *grpc.ClientConn
	eurekaConn            *grpc.ClientConn
	fatimaConn            *grpc.ClientConn
	shamirConn            *grpc.ClientConn
	usermgmtConn          *grpc.ClientConn
	entryExitMgmtConn     *grpc.ClientConn
	bobDB                 *pgxpool.Pool
	tomDB                 *pgxpool.Pool
	eurekaDB              *pgxpool.Pool
	fatimaDB              *pgxpool.Pool
	zeusDB                *pgxpool.Pool
	enigmaSrvURL          string
	jprepKey              string
	jprepSignature        string
	requestAt             *timestamppb.Timestamp
	hasClassJPREPInActive bool
	firebaseAddr          string
	bobDBTrace            *database.DBTrace
	zapLogger             *zap.Logger
	jsm                   nats.JetStreamManagement
	applicantID           string
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TestSuiteInitializer ...
func TestSuiteInitializer(c *gandalf.Config, fakeFirebaseAddr string) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c, fakeFirebaseAddr)
		})

		ctx.AfterSuite(func() {
			bobDB.Close()
			tomDB.Close()
			eurekaDB.Close()
			fatimaDB.Close()
			zeusDB.Close()
			bobConn.Close()
			tomConn.Close()
			yasuoConn.Close()
			eurekaConn.Close()
			fatimaConn.Close()
			shamirConn.Close()
			entryExitMgmtConn.Close()
			jsm.Close()
		})
	}
}

// ScenarioInitializer ...
func ScenarioInitializer(c *gandalf.Config) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		ctx.BeforeScenario(func(p *messages.Pickle) {
			s := newSuite(p.Id)
			initSteps(ctx, s)
		})
	}
}

func setup(c *gandalf.Config, fakeFirebaseAddr string) {
	firebaseAddr = fakeFirebaseAddr
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	err := c.ConnectGRPCInsecure(ctx, &bobConn, &tomConn, &yasuoConn, &eurekaConn, &fatimaConn, &shamirConn, &usermgmtConn, &entryExitMgmtConn)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to run BDD setup: %s", err))
	}
	c.ConnectDB(ctx, &bobDB, &tomDB, &eurekaDB, &fatimaDB, &zeusDB)

	jsm, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to create jetstream management: %v", err))
	}
	jsm.ConnectToJS()

	enigmaSrvURL = "http://" + c.EnigmaSrvAddr
	jprepKey = c.JPREPSignatureSecret
	db, _, _ := database.NewPool(context.Background(), zap.NewNop(), c.PostgresV2.Databases["bob"])
	bobDBTrace = &database.DBTrace{
		DB: db,
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
	_, err = bobDBTrace.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
}

type suite struct {
	connections
	stepState
	YasuoStepState

	yasuoSuite  *yasuo.Suite
	ZapLogger   *zap.Logger
	ApplicantID string
}

type connections struct {
	bobConn               *grpc.ClientConn
	tomConn               *grpc.ClientConn
	yasuoConn             *grpc.ClientConn
	shamirConn            *grpc.ClientConn
	bobDB                 *pgxpool.Pool
	tomDB                 *pgxpool.Pool
	eurekaDB              *pgxpool.Pool
	fatimaDB              *pgxpool.Pool
	zeusDB                *pgxpool.Pool
	enigmaSrvURL          string
	JPREPKey              string
	JPREPSignature        string
	RequestAt             *timestamppb.Timestamp
	HasClassJPREPInActive bool
	bobDBTrace            *database.DBTrace
	jsm                   nats.JetStreamManagement
}

type stepState struct {
	ID       string
	User     *entities.User
	Class    *entities.Class
	Request  interface{}
	Response interface{}
	Payload  interface{}

	CurrentUserID string
}

func newSuite(id string) *suite {
	s := &suite{
		connections: connections{
			bobConn,
			tomConn,
			yasuoConn,
			shamirConn,
			bobDB,
			tomDB,
			eurekaDB,
			fatimaDB,
			zeusDB,
			enigmaSrvURL,
			jprepKey,
			jprepSignature,
			requestAt,
			hasClassJPREPInActive,
			bobDBTrace,
			jsm,
		},
		stepState:   stepState{ID: id},
		ZapLogger:   zapLogger,
		ApplicantID: applicantID,
	}

	s.newYasuoSuite(firebaseAddr)
	return s
}

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^a request with m_course_name payload$`:                              s.stepARequestWithCourseNamePayload,
		`^a request with m_course_name payload missing "([^"]*)"$`:            s.stepARequestWithCourseNamePayloadMissing,
		`^a request with current m_course_name payload and action "([^"]*)"$`: s.stepRequestWithCurrentCourseNamePayloadAndAction,
		`^a valid JPREP signature in its header$`:                             s.stepAValidJPREPSignatureInItsHeader,
		`^the request user registration is performed$`:                        s.stepPerformUserRegistrationRequest,
		`^the request master registration is performed$`:                      s.stepPerformMasterRegistrationRequest,
		`^a "([^"]*)" status is returned$`:                                    s.stepACorrectStatusIsReturned,
		`^the course must be registered in the system$`:                       s.stepYasuoMustCreateCourse,
		`^the course must not be registered in the system$`:                   s.stepYasuoMustNotCreateCourse,
		`^the course must be registered in the system with action "([^"]*)"$`: s.stepYasuoMustCreateCourseWithAction,

		`^the class must be store in our system$`:                               s.theClassesMustBeStoreInOutSystem,
		`^the class must not be store in our system$`:                           s.theClassesMustNotBeStoreInOutSystem,
		`^the course class must not be store in our system$`:                    s.theCoursesClassesMustNotBeStoreInOutSystem,
		`^the course academic year must not be store in our system$`:            s.theCoursesAcademicYearsMustNotBeStoreInOutSystem,
		`^jprep Tom must store conversation with status "([^"]*)"$$`:            s.tomMustStoreConversationWithStatus,
		`^jprep Tom must record messages create class`:                          s.tomMustRecordMessageCreated,
		`^a request new class with m_regular_course payload$`:                   s.stepARequestNewClassWithRegularCoursePayload,
		`^a request new class with m_regular_course payload missing "([^"]*)"$`: s.stepARequestNewClassWithRegularCoursePayloadMissing,
		`^a request new class member with m_student payload$`:                   s.stepARequestNewClassMemberWithStudentPayload,
		`^a request new class member with m_student payload missing "([^"]*)"$`: s.stepARequestNewClassMemberWithStudentPayloadMissing,
		`^the students must be store in our system$`:                            s.theStudentsMustBeStoreInOutSystem,
		`^the students must not be store in our system$`:                        s.theStudentsMustNotBeStoreInOutSystem,
		`^the teachers must be store in our system$`:                            s.theTeachersMustBeStoreInOutSystem,
		`^the teachers must not be store in our system$`:                        s.theTeachersMustNotBeStoreInOutSystem,
		`^the class members must be store in our system$`:                       s.theClassMembersMustBeStoreInOutSystem,
		`^jprep Eureka must store class member with action "([^"]*)"$`:          s.eurekaMustStoreClassMemberWithAction,
		`^jprep Eureka must store course students with action "([^"]*)"$`:       s.eurekaMustStoreCourseStudentsWithAction,

		`^a request exist class member with m_student payload and action "([^"]*)"$`:                    s.stepARequestExistClassMemberWithStudentPayload,
		`^a request new teacher with m_staff payload$`:                                                  s.stepARequestNewClassMemberWithStaffPayload,
		`^a request new teacher with m_staff payload missing "([^"]*)"$`:                                s.stepARequestNewClassMemberWithStaffPayloadMissing,
		`^a request exist class with m_regular_course payload with action "([^"]*)"$`:                   s.stepARequestExistClassWithRegularCoursePayload,
		`^jprep Tom must record message join class of current user with message "([^"]*)"$`:             s.tomMustRecordMessageJoinClassOfCurrentUser,

		`^a request new live lesson with m_lesson payload$`:                   s.stepRequestNewLiveLessonWithLessonPayload,
		`^a request new live lesson with m_lesson payload missing "([^"]*)"$`: s.stepRequestNewLiveLessonWithLessonPayloadMissing,
		`^the lesson must be store in our system with action "([^"]*)"$`:      s.theLessonMustBeStoreInOurSystemWithAction,
		`^the lesson group must be store in our system$`:                      s.theLessonGroupMustBeStoreInOurSystem,
		`^the preset study plan weekly must be store in our system$`:          s.thePresetStudyPlanWeeklyMustBeStoreInOurSystem,
		`^the topics must be store in our system with action "([^"]*)"$`:      s.theTopicMustBeStoreInOurSystemWithAction,
		`^the preset study plan must be store in our system$`:                 s.thePresetStudyPlanMustBeStoreInOurSystem,
		`^jprep Tom must store conversation lesson$`:                          s.tomMustStoreConversationLesson,
		`^the preset study plan weekly must be delete in our system$`:         s.thePresetStudyPlanWeeklyMustBeDeleteInOurSystem,
		`^the course must be update in our system$`:                           s.theCourseMustBeUpdateInOurSystem,
		`^the lesson must not be store in our system$`:                        s.theLessonMustNotBeStoreInOurSystem,
		`^the lesson group must not be store in our system$`:                  s.theLessonGroupMustNotBeStoreInOurSystem,
		`^the preset study plan must not be store in our system$`:             s.thePresetStudyPlanMustNotBeStoreInOurSystem,
		`^the preset study plan weekly must not be store in our system$`:      s.thePresetStudyPlanWeeklyMustNotBeStoreInOurSystem,
		`^the topic must not be store in our system$`:                         s.theTopicMustNotBeStoreInOurSystem,

		`^a request exist live lesson with m_lesson payload and action "([^"]*)"$`: s.stepRequestExistLiveLessonWithLessonPayload,
	}

	yasuoSteps := initStepForYasuoServiceFeature(s)
	appendSteps(yasuoSteps, steps)

	for pattern, stepFunc := range steps {
		ctx.Step(pattern, stepFunc)
	}
}

func appendSteps(src, dest map[string]interface{}) {
	for k, v := range src {
		dest[k] = v
	}
}

func (s *suite) stepAValidJPREPSignatureInItsHeader() error {
	data, err := json.Marshal(s.Request)
	if err != nil {
		return err
	}
	sig, err := s.generateSignature(s.JPREPKey, string(data))
	if err != nil {
		return nil
	}
	s.JPREPSignature = sig
	return nil
}

func (s *suite) generateSignature(key, message string) (string, error) {
	sig := hmac.New(sha256.New, []byte(key))
	if _, err := sig.Write([]byte(message)); err != nil {
		return "", err
	}
	return hex.EncodeToString(sig.Sum(nil)), nil
}

func (s *suite) stepPerformUserRegistrationRequest() error {
	url := fmt.Sprintf("%s/jprep/user-registration", s.enigmaSrvURL)
	bodyBytes, err := s.makeJPREPHTTPRequest(http.MethodPut, url)
	if err != nil {
		return err
	}

	if bodyBytes == nil {
		return fmt.Errorf("body is nil")
	}

	return nil
}

func (s *suite) stepPerformMasterRegistrationRequest() error {
	url := fmt.Sprintf("%s/jprep/master-registration", s.enigmaSrvURL)
	bodyBytes, err := s.makeJPREPHTTPRequest(http.MethodPut, url)
	if err != nil {
		return err
	}

	if bodyBytes == nil {
		return fmt.Errorf("body is nil")
	}

	return nil
}

func (s *suite) makeJPREPHTTPRequest(method, url string) ([]byte, error) {
	bodyRequest, err := json.Marshal(s.Request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyRequest))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("JPREP-Signature", s.JPREPSignature)
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: time.Duration(30) * time.Second}
	s.RequestAt = &timestamppb.Timestamp{Seconds: time.Now().Unix()}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}
	s.Response = resp
	return body, nil
}

func (s *suite) stepACorrectStatusIsReturned(statusCode int) error {
	actualStatusCode := s.Response.(*http.Response).StatusCode
	if actualStatusCode != statusCode {
		return fmt.Errorf(
			"expected status code %d, got %d in response %v",
			statusCode, actualStatusCode, s.Response.(*http.Response),
		)
	}
	return nil
}
