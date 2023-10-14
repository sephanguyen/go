package yasuo

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	cconstants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	"github.com/manabie-com/backend/internal/yasuo/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	natsJS "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func init() {
	common.RegisterTest("yasuo", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

// fields
const (
	Name                = "name"
	Book                = "book"
	Chapter             = "chapter"
	DisplayOrder        = "display_order"
	Country             = "country"
	SchoolID            = "schoolID"
	Subject             = "subject"
	Grade               = "grade"
	CountryAndGrade     = "country and grade"
	All                 = "all"
	None                = "none"
	DefaultResourcePath = "1"
)

var (
	firebaseAddr string
	zapLogger    *zap.Logger

	otelFlushFunc func() = func() {} // noop by default
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Suite struct {
	suite
}

type suite struct {
	DB                 *pgxpool.Pool
	EurekaDB           database.Ext
	DBTrace            *database.DBTrace
	BobConn            *grpc.ClientConn
	EurekaConn         *grpc.ClientConn
	FatimaConn         *grpc.ClientConn
	tomConn            *grpc.ClientConn
	Conn               *grpc.ClientConn
	userManagementConn *grpc.ClientConn
	ShamirConn         *grpc.ClientConn
	Cfg                *common.Config
	JSM                nats.JetStreamManagement
	ZapLogger          *zap.Logger
	ApplicantID        string

	StepState
	FirebaseClient *auth.Client
	TenantManager  multitenant.TenantManager
}

func (s *suite) signedCtx(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", stepState.AuthToken)
}

type StepState struct {
	AuthToken     string
	Request       interface{}
	Response      interface{}
	ResponseErr   error
	RequestSentAt time.Time
	Random        string
	Examples      interface{}

	CurrentUserID               string
	CurrentUserGroup            string
	CurrentNotificationID       string
	CurrentNotificationTargetID string
	CurrentLessonNames          []string
	CurrentClassID              int32
	CurrentClassIDs             []int32

	ExistingStudents          []*entities.Student
	ExistingParents           entities.Parents
	ExistingStudentPackageIds []string
	ChapterID                 string
	TopicID                   string

	LoID                     string
	CurrentTeacherID         string
	CourseIDs                []string
	QuizID                   string
	LessonGroupID            string
	CurrentCourseID          string
	StudentIDs               []string
	CurrentBookIDs           []string
	ParentIDs                []string
	MaterialIds              []string
	CurrentSchoolID          int32
	DefaultSchoolLocation    string
	CurrentLessonIDs         []string
	Grades                   []int
	CourseStudentIDs         map[string][]string
	NotificationID           string
	MapSchoolIDAndTeacherIDs map[int][]string

	BookId      string
	CourseIds   []string
	LoIDMap     map[string]string
	Quizzes     entities.Quizzes
	QuizSet     entities.QuizSet
	DeletedQuiz *entities.Quiz
	PackageIDs  []int32

	AcademicID string

	ExistedLessons         []*npb.EventMasterRegistration_Lesson
	ExpectedUserBeforeSend []string
	ExpectedUserAfterSend  []string

	Notification               *cpb.Notification
	NotificationNeedToSent     *cpb.Notification
	NotificationDontNeedToSent *cpb.Notification
	StudentParent              map[string][]string

	RemovedStudentLessons []string

	ChapterIDs            []string
	FoundChanForJetStream chan interface{}

	PartnerSyncDataLogId      string
	PartnerSyncDataLogSplitId string

	Subs []*natsJS.Subscription
}

type stateKeyForYasuo struct{}

func StepStateFromContext(ctx context.Context) *StepState {
	state := ctx.Value(stateKeyForYasuo{})
	if state == nil {
		return &StepState{}
	}
	return state.(*StepState)
}

func StepStateToContext(ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, stateKeyForYasuo{}, state)
}

func InitYasuoState(ctx context.Context) (context.Context, error) {
	return StepStateToContext(ctx, &StepState{
		StudentParent:            make(map[string][]string),
		CourseStudentIDs:         make(map[string][]string),
		MapSchoolIDAndTeacherIDs: make(map[int][]string),
	}), nil
}

func (s *suite) ReturnsStatusCode(ctx context.Context, code string) (context.Context, error) {
	return s.returnsStatusCode(ctx, code)
}

func (s *suite) returnsStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}

	if stt.Code().String() != arg1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting %s, got %s status code, message: %s", arg1, stt.Code().String(), stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) storeActivityLogs(ctx context.Context, methodName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userID := stepState.CurrentUserID

	activityLogRepo := repositories.ActivityLogRepo{}
	log, err := activityLogRepo.FindByUserIDAndType(ctx, s.DBTrace, userID, methodName)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	iPayload := log.Payload.Get()
	payload, ok := iPayload.(map[string]interface{})
	if !ok {
		return StepStateToContext(ctx, stepState), errors.New("cannot parse payload")
	}
	_, ok = payload["err"].(map[string]interface{})
	if !ok && strings.Contains(methodName, "_FAIL") {
		return StepStateToContext(ctx, stepState), errors.New("do not have payload error of fail logs")
	}

	if ok && strings.Contains(methodName, "_OK") {
		return StepStateToContext(ctx, stepState), errors.New("have payload error of success logs")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) SignedAsAccount(ctx context.Context, arg1 string) (context.Context, error) {
	return s.signedAsAccount(ctx, arg1)
}

func (s *suite) signedAsAccount(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if group == "unauthenticated" {
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	}

	id := idutil.ULIDNow()
	var userGroup string

	switch group {
	case "teacher":
		userGroup = constant.UserGroupTeacher
	case "school admin":
		userGroup = constant.UserGroupSchoolAdmin
	case "parent":
		userGroup = constant.UserGroupParent
	}

	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = userGroup

	return s.aValidUserInDB(StepStateToContext(ctx, stepState), withID(id), withRole(userGroup))
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(s *suite, ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)

	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", stepState.AuthToken)
}

func generateAuthenticationToken(sub string, template string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(b), nil
}

func generateValidAuthenticationToken(sub, userGroup string) (string, error) {
	return generateAuthenticationToken(sub, "templates/"+userGroup+".template")
}

func (s *suite) aRandomNumber() {
	s.Random = strconv.Itoa(rand.Int())
}

//nolint:gocyclo
func (s *suite) bobMustPushMsgSubjectToNats(ctx context.Context, eName, qName string) (context.Context, error) {
	time.Sleep(500 * time.Millisecond)
	stepState := StepStateFromContext(ctx)

	foundChn := make(chan struct{}, 1)

	switch qName {
	case cconstants.SubjectLessonCreated:
		if eName != "EvtLesson from Jprep" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("eName != EvtLesson from Jprep")
		}
		timer := time.NewTimer(time.Minute)
		defer timer.Stop()
		select {
		case <-stepState.FoundChanForJetStream:
			return StepStateToContext(ctx, stepState), nil
		case <-timer.C:
			return StepStateToContext(ctx, stepState), errors.New("time out")
		}
	case cconstants.SubjectClassUpserted:
		timer := time.NewTimer(time.Minute)
		defer timer.Stop()
		for {
			select {
			case message := <-stepState.FoundChanForJetStream:
				switch message := message.(type) {
				case *pb.EvtClassRoom_CreateClass_:
					if eName == "CreateClass" {
						if message.CreateClass.ClassId == stepState.CurrentClassID && message.CreateClass.ClassName != "" {
							return StepStateToContext(ctx, stepState), nil
						}
					}
				case *pb.EvtClassRoom_EditClass_:
					if eName == "EditClass" {
						if message.EditClass.ClassId == stepState.CurrentClassID && message.EditClass.ClassName != "" {
							return StepStateToContext(ctx, stepState), nil
						}
					}
				case *pb.EvtClassRoom_JoinClass_:
					if eName == "JoinClass" {
						if message.JoinClass.ClassId == stepState.CurrentClassID {
							return StepStateToContext(ctx, stepState), nil
						}
						for _, classID := range stepState.CurrentClassIDs {
							if message.JoinClass.ClassId == classID {
								return StepStateToContext(ctx, stepState), nil
							}
						}
					}
				case *pb.EvtClassRoom_LeaveClass_:
					if eName == "LeaveClass" {
						return StepStateToContext(ctx, stepState), nil
					}
					if strings.Contains(eName, "LeaveClass") {
						if message.LeaveClass.ClassId == stepState.CurrentClassID && len(message.LeaveClass.UserIds) != 0 {
							if strings.Contains(eName, fmt.Sprintf("-is_kicked=%v", message.LeaveClass.IsKicked)) {
								return StepStateToContext(ctx, stepState), nil
							}
						}
					}
				case *pb.EvtClassRoom_ActiveConversation_:
					active := eName == "ActiveConversation"
					if message.ActiveConversation.ClassId == stepState.CurrentClassID && message.ActiveConversation.Active == active {
						return StepStateToContext(ctx, stepState), nil
					}
				}
			case <-timer.C:
				return StepStateToContext(ctx, stepState), errors.New("time out")
			}
		}
	}

	timer := time.NewTimer(time.Minute)
	defer timer.Stop()

	select {
	case <-stepState.FoundChanForJetStream:
		switch qName {
		// Jprep -> yasuo -> tom
		case cconstants.SubjectSyncStudentLessons:
			if eName != "EventSyncUserCourse" {
				return StepStateToContext(ctx, stepState), fmt.Errorf("eName not matched, expected EventSyncUserCourse, got %s", eName)
			}
			return StepStateToContext(ctx, stepState), nil
		default:
			return StepStateToContext(ctx, stepState), errors.New("not matched any qname")
		}
	case <-foundChn:
		return StepStateToContext(ctx, stepState), nil
	case <-timer.C:
		return StepStateToContext(ctx, stepState), errors.New("time out")
	}
}

// func initOtel(c *common.Config) trace.TracerProvider {
//         _, tp, flush := interceptors.InitTelemetry(&c.Common, "yasuo-gandalf", 1)
//         otelFlushFunc = flush
//         return tp
// }

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(parentContext *godog.ScenarioContext) {
		// if c.TraceEnabled {
		//         setupTraceForStepFuncs(parentContext)
		// }

		parentContext.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			s := newSuite(c)
			initSteps(parentContext, s)
			ctx, err := InitYasuoState(ctx)
			if err != nil {
				return ctx, err
			}
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

		parentContext.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
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

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^a signed in "([^"]*)"$`:         s.aSignedIn,
		`^returns "([^"]*)" status code$`: s.returnsStatusCode,
		`^signed as "([^"]*)" account$`:   s.signedAsAccount,
		`^a random number$`:               s.aRandomNumber,

		// get user profile
		`^user get profile$`:                            s.userGetProfile,
		`^yasuo must return user profile$`:              s.yasuoMustReturnUserProfile,
		`^signed as "([^"]*)" account have user group$`: s.signedAsAccountHaveUserGroup,

		// upsert course,
		`^a UpsertCourseRequest with "([^"]*)" data$`:                 s.aUpsertCourseRequestWithData,
		`^user upsert courses$`:                                       s.userUpsertCourses,
		`^courses upserted in DB match with request "([^"]*)"$`:       s.coursesUpsertedInDBMatchWithRequest,
		`^course book upserted in DB match with request$`:             s.courseBookUpsertInDBMatchWithRequest,
		`^activity logs of "([^"]*)" event is "([^"]*)"$`:             s.activityLogsOfEventIs,
		`^user sent duplicate of existed course$`:                     s.userSendDuplicateOfExistedCourse,
		`^admin upsert course$`:                                       s.adminUpsertCourses,
		`^yasuo must store course$`:                                   s.yasuoMustStoreCourse,
		`^the admin upsert courses$`:                                  s.theAdminUpsertCourses,
		`^our system has to store upsert course with book correctly$`: s.ourSystemHasToStoreUpsertCourseWithBookCorrectly,

		// delete cours,
		`^a DeleteCourseRequest with id "([^"]*)"$`:       s.aDeleteCourseRequestWithID,
		`^user delete courses$`:                           s.userDeleteCourses,
		`^courses are "([^"]*)" in DB$`:                   s.coursesIsInDB,
		`^a list of "([^"]*)" courses are existed in DB$`: s.aListOfCoursesAreExistedInDB,

		`^user create brightcove upload url for video "([^"]*)"$`: s.userCreateBrightcoveUploadUrlForVideo,
		`^yasuo must return a video upload url$`:                  s.yasuoMustReturnAVideoUploadUrl,

		`^user finish brightcove upload url for video "([^"]*)"$`: s.userFinishBrightcoveUploadUrlForVideo,

		`^api v2 user create brightcove upload url for video "([^"]*)"$`: s.userCreateBrightcoveUploadUrlForVideoV2,
		`^api v2 yasuo must return a video upload url$`:                  s.yasuoMustReturnAVideoUploadUrlV2,
		`^api v2 user finish brightcove upload url for video "([^"]*)"$`: s.userFinishBrightcoveUploadUrlForVideoV2,

		`^api v2 get brightcove profile data$`: s.getBrightcoveProfileData,

		`^yasuo must store activity logs "([^"]*)"$`: s.storeActivityLogs,

		// upsert live course,
		`^a UpsertLiveCourseRequest with missing "([^"]*)"$`: s.aUpsertLiveCourseRequestWithMissing,
		`^user upsert live courses$`:                         s.userUpsertLiveCourses,
		`^a class$`:                                          s.aClass,
		`^yasuo must store live course$`:                     s.yasuoMustStoreLiveCourse,

		// upsert courses book,
		`^an existed course with id "([^"]*)"$`: s.anExistedCourseWithID,
		`^a list of books in our DB$`:           s.listOfBooksInOurDB,

		// event JPREP sync course,
		`^jpref sync "([^"]*)" courses with action "([^"]*)" and "([^"]*)" course with action "([^"]*)" to our system$`: s.jprepSyncCoursesWithActionAndCourseWithActionToOurSystem,
		`^these courses must be store in our system$`:                                                                   s.theseCoursesMustBeStoreInOurSystem,

		`^jpref sync "([^"]*)" students with action "([^"]*)" and "([^"]*)" students with action "([^"]*)"$`: s.jprepSyncStudentsWithActionAndStudentsWithAction,
		`^these students must be store in our system$`:                                                       s.theseStudentsMustBeStoreInOurSystem,

		`^jprep sync "([^"]*)" lesson members with action "([^"]*)" and "([^"]*)" lesson members with action "([^"]*)" at (\d+) ago$`: s.jprepSyncLessonMembersWithActionAndLessonMembersWithAction,
		`^jprep sync "([^"]*)" lesson with action "([^"]*)" and "([^"]*)" lesson with action "([^"]*)"$`:                              s.jprefSyncLessonWithActionAndLessonWithAction,
		`^these lesson members must be store in our system$`:                                                                          s.theseLessonMembersMustBeStoreInOurSystem,
		`^these no lesson members store in our system$`:                                                                               s.theseNoLessonMembersStoreInOurSystem,
		`^these lessons must be store in our system correctly$`:                                                                       s.theseLessonsMustBeStoreInOurSystem,
		`^jprep resync lesson members but excluding a lesson$`:                                                                        s.jprepResyncLessonMembersButExcludingALesson,
		`^jprep sync some lessons to student$`:                                                                                        s.jprepSyncSomeLessonsToStudent,
		`^yasuo must push event removing lesson members to "([^"]*)" for excluded lesson$`:                                            s.yasuoMustPushEventRemovingLessonMembersToForExcludedLesson,

		`^JPREP sync "([^"]*)" class with action "([^"]*)" and "([^"]*)" class with action "([^"]*)"$`: s.jprepSyncClassWithActionAndClassWithAction,
		`^these classes must be store in our system$`:                                                  s.theseClassesMustBeStoreInOutSystem,
		`^these new classes must be store in our system$`:                                              s.theseNewClassesMustBeStoreInOutSystem,

		`^JPREP sync "([^"]*)" class members with action "([^"]*)" and "([^"]*)" class members with action "([^"]*)"$`: s.jprepSyncClassMembersWithActionAndClassMembersWithAction,
		`^these class members must be store in out system$`:                                                            s.theseClassMembersMustBeStoreInOutSystem,
		`^these new class members must be stored in out system$`:                                                       s.theseClassMembersNewMustBeStoreInOutSystem,

		// attach materials to lesson group,
		`^a valid course$`:                                           s.aValidCourse,
		`^admin attach materials into lesson group$`:                 s.attachMaterialsToLessonGroup,
		`^system must attach materials into lesson group correctly$`: s.bobMustAttachMaterialToLessonGroup,

		`^JPREP sync "([^"]*)" staffs with action "([^"]*)" and "([^"]*)" staffs with action "([^"]*)"$`:    s.jprepSyncStaffsWithActionAndStaffsWithAction,
		`^these staffs must be store in our system$`:                                                        s.theseStaffsMustBeStoreInOurSystem,
		`^some courses existed in db$`:                                                                      s.someCoursesExistedInDB,
		`^some courses must have icon$`:                                                                     s.someCoursesMustHaveIcon,
		`^these courses have to save correctly$`:                                                            s.theseCoursesHaveToSaveCorrectly,
		`^jpref sync arbitrary number new courses and existed courses with action "([^"]*)" to our system$`: s.jprefSyncArbitraryNumberNewCoursesAndExistedCoursesWithActionToOurSystem,
		`^jprep sync some new lesson with action "([^"]*)" and some existed lesson with action "([^"]*)"$`:  s.jrefSyncSomeNewLessonWithActionAndSomeExistedLessonWithAction,
		`^some existed lesson in database$`:                                                                 s.someExistedLessonInDatabase,
		`^these lesson updated type "([^"]*)"$`:                                                             s.theseLessonUpdatedType,
		`^after the deleted staff were "([^"]*)"$`:                                                          s.jprepSyncSyncDeletedStaffWithAction,
		`^they login our system and "([^"]*)" get self-profile info$`:                                       s.checkAfterSignedInGetSelfProfile,

		`^jprep sync academic year to our system$`:           s.jprepSyncAcademicYearToOurSystem,
		`^some academic year message$`:                       s.someAcademicYearMessage,
		`^these academic years must be store in our system$`: s.theseAcademicYearsMustBeStoreInOurSystem,

		`^these lesson have to deleted$`: s.theseLessonHaveToDeleted,

		`^some courses must have book$`: s.someCoursesMustHaveBook,

		`^user gets info of a "([^"]*)" video$`:       s.userGetsInfoOfAVideo,
		`^the correct info of the video is returned$`: s.theCorrectInfoOfTheVideoIsReturned,

		`^user upsert courses with some "([^"]*)" are missing$`: s.userUpsertCoursesWithSomeAreMissing,

		`^data log split store correct "([^"]*)"`:                               s.storeLogDataSplitWithCorrectStatus,
		`^last time received message store correctly with config key "([^"]*)"`: s.storeLastTimeRecievedMessageCorrect,
	}
	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMap(steps)
	})

	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

var (
	conn, bobConn, eurekaConn, fatimaConn, tomConn, userManagementConn, shamirConn *grpc.ClientConn
	db                                                                             *pgxpool.Pool
	eurekaDB                                                                       *pgxpool.Pool
	dbTrace                                                                        *database.DBTrace
	firebaseClient                                                                 *auth.Client
	tenantManager                                                                  multitenant.TenantManager
	jsm                                                                            nats.JetStreamManagement
	applicantID                                                                    string
)

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

func (s *suite) generateExchangeTokenWithSchool(userID, userGroup string, school int64) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", fmt.Errorf("error when create generateValidAuthenticationToken: %v", err)
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.ApplicantID, school, s.ShamirConn)
	if err != nil {
		return "", fmt.Errorf("error when create exchange token: %v", err)
	}
	return token, nil
}

func (s *suite) generateExchangeToken(userID, userGroup string) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", fmt.Errorf("error when create generateValidAuthenticationToken: %v", err)
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.ApplicantID, 1, s.ShamirConn)
	if err != nil {
		return "", fmt.Errorf("error when create exchange token: %v", err)
	}
	return token, nil
}

func setup(c *common.Config, fakeFirebaseAddr string, otelEnabled bool) {
	rsc := bootstrap.NewResources().WithLoggerC(&c.Common)
	firebaseAddr = fakeFirebaseAddr
	var err error

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	opts := []grpc.DialOption{grpc.WithInsecure()}

	if otelEnabled {
		// trprovider := initOtel(c)
		// opts = append(opts, grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor(otelgrpc.WithTracerProvider(trprovider))))
	}

	conn, err = grpc.Dial(rsc.GetAddress("yasuo"), opts...)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("can not connect to service: %v", err))
	}

	bobConn, err = grpc.Dial(rsc.GetAddress("bob"), opts...)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("grpc.Dial connect to bob: %+v", zap.Error(err)))
	}

	eurekaConn = rsc.GRPCDial("eureka")
	fatimaConn = rsc.GRPCDial("fatima")
	tomConn = rsc.GRPCDial("tom")
	userManagementConn = rsc.GRPCDial("usermgmt")
	shamirConn = rsc.GRPCDial("shamir")

	applicantID = c.JWTApplicant

	ctx := context.Background()

	bobDBConfig := c.PostgresV2.Databases["bob"]
	db, _, err = database.NewPool(ctx, zapLogger, configs.PostgresDatabaseConfig{
		User:              "yasuo",
		Password:          bobDBConfig.Password,
		Host:              bobDBConfig.Host,
		Port:              bobDBConfig.Port,
		DBName:            bobDBConfig.DBName,
		MaxConns:          bobDBConfig.MaxConns,
		RetryAttempts:     bobDBConfig.RetryAttempts,
		RetryWaitInterval: bobDBConfig.RetryWaitInterval,
		MaxConnIdleTime:   bobDBConfig.MaxConnIdleTime,
	})
	if err != nil {
		log.Fatalf("failed to connect to bob database: %s", err)
	}

	err = updateResourcePath(db)
	if err != nil {
		log.Fatal("failed to update database: %w", err)
	}

	eurekaDB, _, _ = database.NewPool(ctx, zapLogger, c.PostgresV2.Databases["eureka"])

	dbTrace = &database.DBTrace{
		DB: db,
	}

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	firebaseClient, err = app.Auth(ctx)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create firebase client: %v", err))
	}

	firebaseProject := c.Common.FirebaseProject
	if firebaseProject == "" {
		firebaseProject = c.Common.GoogleCloudProject
	}
	gcpApp, err := gcp.NewApp(ctx, "", firebaseProject)
	if err != nil {
		zapLogger.Fatal("failed to initialize gcp app", zap.Error(err))
	}
	tenantManager, err = multitenant.NewTenantManagerFromGCP(ctx, gcpApp)
	if err != nil {
		zapLogger.Fatal("failed to initialize identity platform tenant manager", zap.Error(err))
	}

	jsm, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to create jetstream management: %v", err))
	}
	jsm.ConnectToJS()

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
	_, err = db.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
}

func newSuite(c *common.Config) *suite {
	s := &suite{}

	s.Cfg = c
	s.Conn = conn
	s.BobConn = bobConn
	s.ShamirConn = shamirConn
	s.EurekaConn = eurekaConn
	s.FatimaConn = fatimaConn
	s.tomConn = tomConn
	s.DB = db
	s.EurekaDB = eurekaDB
	s.DBTrace = dbTrace
	s.FirebaseClient = firebaseClient
	s.TenantManager = tenantManager
	s.StepState = StepState{}
	s.JSM = jsm
	s.ZapLogger = zapLogger
	s.userManagementConn = userManagementConn
	s.ApplicantID = applicantID

	return s
}

func TestSuiteInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c, f.FirebaseAddr, f.OtelEnabled)
		})

		ctx.AfterSuite(func() {
			otelFlushFunc()
			db.Close()
			conn.Close()
			bobConn.Close()
			eurekaDB.Close()
			eurekaConn.Close()
			jsm.Close()
			tomConn.Close()
			userManagementConn.Close()
		})
	}
}

func SetFirebaseAddr(fireBaseAddr string) {
	firebaseAddr = fireBaseAddr
}

func setupTraceForStepFuncs(ctx *godog.ScenarioContext) {
	ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
		// Get parent scenario span, so that next span created is a children of this span
		// instead of the current span in the context, which should be its cousin
		sometype := ctx.Value(traceScenarioSpanKey{})
		if sometype != nil {
			parSpan, ok := sometype.(interceptors.TimedSpan)
			if ok {
				ctx = trace.ContextWithSpan(ctx, parSpan.Span())
			}
		}
		ctx, stepspan := interceptors.StartSpan(ctx, st.Text)
		ctx = context.WithValue(ctx, traceStepSpanKey(st.Id), stepspan) //nolint:revive,staticcheck
		return ctx, nil
	})

	ctx.StepContext().After(func(ctx context.Context, st *godog.Step, stat godog.StepResultStatus, err error) (context.Context, error) {
		someval := ctx.Value(traceStepSpanKey(st.Id))
		if someval != nil {
			stepspan, ok := someval.(interceptors.TimedSpan)
			if ok {
				if err != nil {
					stepspan.RecordError(err)
				}
				stepspan.End()
			}
		}

		return ctx, err
	})
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		ctx, span := interceptors.StartSpan(ctx, fmt.Sprintf("Starting: %s", sc.Name))
		// children steps need this parent span
		ctx = context.WithValue(ctx, traceScenarioSpanKey{}, span)
		return ctx, nil
	})
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		sometype := ctx.Value(traceScenarioSpanKey{})
		traceID := "undefined"
		if sometype != nil {
			span, ok := sometype.(interceptors.TimedSpan)
			if ok {
				defer span.End()
				traceID = span.Span().SpanContext().TraceID().String()
				if err != nil {
					span.SetAttributes(attribute.KeyValue{
						Key:   "x-has-error",
						Value: attribute.StringValue("true"),
					})
					span.RecordError(err)
				}
			}
		}
		if err != nil {
			return ctx, fmt.Errorf("traceID %s: %s", traceID, err)
		}

		return ctx, err
	})
}

func traceStepSpanKey(id string) string {
	return "x-trace-step-key" + id
}

type traceScenarioSpanKey struct{}
