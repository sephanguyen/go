package eureka

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/eureka/entities"
	entities_mnt "github.com/manabie-com/backend/internal/eureka/entities/monitors"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	common.RegisterTest("eureka", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	conn          *grpc.ClientConn
	bobConn       *grpc.ClientConn
	yasuoConn     *grpc.ClientConn
	usermgmtConn  *grpc.ClientConn
	shamirConn    *grpc.ClientConn
	db            *pgxpool.Pool
	dbTrace       *database.DBTrace
	bobDB         *pgxpool.Pool
	bobPgDB       *pgxpool.Pool
	bobDBTrace    *database.DBTrace
	fatimaDB      *pgxpool.Pool
	fatimaDBTrace *database.DBTrace
	firebaseAddr  string
	jsm           nats.JetStreamManagement
	zapLogger     *zap.Logger
	applicantID   string
)

const (
	teacherRawText       = "teacher"
	schoolAdminRawText   = "school admin"
	adminRawText         = "admin"
	studentRawText       = "student"
	parentRawText        = "parent"
	hqStaffRawText       = "hq staff"
	centerLeadRawText    = "center lead"
	centerManagerRawText = "center manager"
	centerStaffRawText   = "center staff"
)

type stateKeyForEureka struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func StepStateFromContext(ctx context.Context) *StepState {
	state := ctx.Value(stateKeyForEureka{})
	if state == nil {
		return &StepState{}
	}
	return state.(*StepState)
}

func StepStateToContext(ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, stateKeyForEureka{}, state)
}

func InitEurekaState(ctx context.Context) context.Context {
	ctx = StepStateToContext(ctx, &StepState{
		StudentsSubmittedAssignments: make(map[string][]*pb.Content),
		ExistingQuestionHierarchy:    make(entities.QuestionHierarchy, 0),
	})
	return ctx
}

func TestSuiteInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c, f.FirebaseAddr)
		})
		ctx.AfterSuite(func() {
			db.Close()
			conn.Close()
			bobDB.Close()
			usermgmtConn.Close()
			shamirConn.Close()
			bobConn.Close()
			jsm.Close()
		})
	}
}

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite()
		initSteps(ctx, s)
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			ctx = InitEurekaState(ctx)
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
	}
}

type Suite struct {
	suite
}

type suite struct {
	DB           database.Ext
	DBTrace      *database.DBTrace
	Conn         *grpc.ClientConn
	BobConn      *grpc.ClientConn
	UsermgmtConn *grpc.ClientConn
	YasuoConn    *grpc.ClientConn
	*StepState
	BobDB         database.Ext
	BobPgDB       database.Ext
	BobDBTrace    *database.DBTrace
	FatimaDB      database.Ext
	FatimaDBTrace *database.DBTrace
	JSM           nats.JetStreamManagement
	ZapLogger     *zap.Logger

	assignments []*pb.Assignment
	ApplicantID string
	ShamirConn  *grpc.ClientConn
}

type releaseOpenTimeSchedulerContext struct {
	brandID             string
	centerIDs           []string
	selectedCenterIDs   []string
	unSelectedCenterIDs []string
	hasError            bool
}

// nolint:revive
type StepState struct {
	QuestionID                            string
	QuizItems                             []*cpb.Quiz
	ID                                    string
	AuthToken                             string
	Request                               interface{}
	Response                              interface{}
	BobResponse                           interface{}
	ResponseErr                           error
	RequestSentAt                         time.Time
	AssignmentIDs                         []string
	StudyPlanID                           string
	CourseID                              string
	StudyPlanItemIDs                      []string
	LoIDs                                 []string
	DeletedLoIDs                          []string
	UserId                                string //nolint:stylecheck
	CurrentUserID                         string
	StartAt                               *timestamppb.Timestamp
	EndAt                                 *timestamppb.Timestamp
	CourseIDs                             []string
	Event                                 interface{}
	ClassIDs                              []int32
	ClassIDsString                        []string
	StudentIDs                            []string
	StudentToken                          string
	StudentStudyPlanID                    string
	ShuffledQuizSetIDs                    []string
	QuizSetIDs                            []string
	QuizExternalIDs                       []string
	Grade                                 int32
	NumQuizzes                            int
	WrongQuizExternalIDs                  []string
	BookID                                string
	BookIDs                               []string
	DeletedBookIDs                        []string
	ChapterID                             string
	ChapterIDs                            []string
	CurrentChapterIDs                     []string
	UpdatedChapterIDs                     []string
	Chapters                              []*entities.Chapter
	TopicID                               string
	TopicList                             []*pb.Topic
	CurrentTopicID                        string
	ListStatus                            []pb.SubmissionStatus
	RotsCtx                               releaseOpenTimeSchedulerContext
	StudyPlanItems                        []*pb.StudyPlanItem
	TopicEntities                         []*entities.Topic
	PaginatedBooks                        [][]*cpb.Book
	QuizOptions                           map[string]map[string][]*cpb.QuizOption
	ShuffledQuizSetID                     string
	CheckQuizCorrectnessResponses         []*pb.CheckQuizCorrectnessResponse
	SelectedQuiz                          []int
	SelectedIndex                         map[string]map[string][]*pb.Answer
	FilledText                            map[string]map[string][]*pb.Answer
	SubmittedKeys                         map[string]map[string][]*pb.Answer          // key: set id, value: (key: quiz external id, value: answer)
	CheckQuizCorrectnessResponsesByQuizID map[string]*pb.CheckQuizCorrectnessResponse // key: quiz external id, value: response which after submit answer
	expectedCorrectnessByQuizID           map[string][]bool                           // key: quiz external id, value: expected list correctness
	expectedCorrectKeysByQuizID           map[string][]string                         // key: quiz external id, value: list correct keys
	expectedIsCorrectAllByQuizID          map[string]bool                             // key: quiz external id, value: IsCorrectAll

	UpdatedBookIDs []string
	Books          []*entities.Book

	SchoolID                       string
	SchoolIDInt                    int32
	ArchivedStudyPlanIDs           []string
	ArchivedStudyPlanItemIDs       []string
	StudentIDsWithMissingStudyPlan []string
	QuizAnswers                    []*pb.QuizAnswer

	SchoolAdminID    string
	SchoolAdminToken string
	School           *bob_entities.School
	Schools          []*bob_entities.School
	TaskID           string
	TeacherID        string
	TeacherToken     string

	HqStaffID    string
	HqStaffToken string

	AvailableStudyPlanIDs  []string
	Random                 string
	OldStudyPlanItemStatus pb.StudyPlanItemStatus
	CourseStudyPlanID      string

	// retrieve student submission history
	SetIDs                   []string
	AnswerLogs               []*cpb.AnswerLog
	NumShuffledQuizSetLogs   int
	PaginatedCourses         [][]*cpb.Course
	StudentDoingQuizExamLogs map[string][]*cpb.AnswerLog

	// store available contents of StudentIDs
	Contents [][]*pb.Content
	// student submission test
	Submissions []*pb.SubmitAssignmentRequest
	LatestGrade map[string]*pb.SubmissionGrade

	// store assigned study plans of StudentIDs
	PaginatedStudyPlans [][][]*pb.StudyPlan

	PaginatedToDoItems [][][]*pb.ToDoItem

	// list student by course
	NumberOfId       int
	StatusSubmission string

	// student event
	SessionID        string
	StudyPlanItemID  string
	LoID             string
	DeletedQuizID    string
	AssignmentID     string
	CurrentStudentID string
	SubmissionID     string
	SubmissionIDs    []string
	LOs              []*entities.ContentStructure
	BookCourseMap    map[string][]string
	Students         []*Student

	LoItemMissing              []string
	AssignmentItemMissing      []string
	StudyPlanItemMonitors      entities_mnt.StudyPlanMonitors
	CurrentChapterDisplayOrder int32
	CurrentTopicDisplayOrder   int32

	// for concurrency two school admin
	AnotherSchoolAdminID    string
	AnotherSchoolAdminToken string
	AnotherChapterIDs       []string

	// for new and old chapters (created before and now)
	OldChapterIDs []string
	NewChapterIDs []string
	OldTopicIDs   []string
	NewTopicIDs   []string

	NumberOfUpdatedOldTopics int

	StudentsSubmittedAssignments map[string][]*pb.Content

	NumberOfSubmissionGraded int
	StudentsCourseMap        map[string]string

	StudentPackageStatus bool
	CurrentClassID       int32

	StudyPlanIDs                   []string
	StudyPlanItemIDsOverDue        []string
	StudyPlanItemIDsOverDueDeleted []string
	StudyPlanItemIDsCompleted      []string
	StudyPlanItemIDsActive         []string
	ToDoItems                      []*pb.ToDoItem
	LengthStudyPlanItems           int

	StudyPlans []*pb.StudyPlan

	NumOfCompletedStudent int
	ClassID               string

	CurrentOrgID              string
	CurrentBrandID            string
	IsDeleteStudyPlan         bool
	TeacherStudyPlans         []*pb.StudyPlan
	AssignmentStudyPlanItemID string

	StudyPlanType string
	TopicIDs      []string

	StudyPlanItemInfos []*StudyPlanItemInfo
	ToDoItemList       []*ToDoItem

	StudyPlanStatus           string
	StudyPlanItemStatus       string
	ListStudentTodoItemStatus string
	StudentID                 string

	StudentIDExpired             string
	CurrentSchoolID              int32
	MapExistingPackageAndCourses map[string]string

	CourseStudents   []*entities.CourseStudent
	Assignments      []*pb.Assignment
	LocationIDs      []string
	ClassLocationIDs []string

	UserFillInTheBlankOld bool

	Offset     int
	Limit      int
	NextPage   *cpb.Paging
	SetID      string
	StudySetID string

	NewStartDate time.Time
	NewEndDate   time.Time
	OldEndDate   time.Time
	OldStartDate time.Time

	Topics          []*pb.Topic
	QuizSet         entities.QuizSet
	Quizzes         entities.Quizzes
	QuizIDs         []string
	QuizLOs         []*ypb.QuizLO
	QuestionTagIds  []string
	LoIDMap         map[string]string
	AnotherTopicIDs []string

	NumTopics         int
	NumChapter        int
	SkippedTopics     []string
	PaginatedChapters [][]*cpb.Chapter

	StudentEventLogs []*entities.StudentEventLog

	QuizLOList []*pb.QuizLO
	QuizID     string
	// retrieve total quiz of los
	LOIDs      []string
	LOIDsInReq []string

	AllQuizzesRes        []*cpb.Quiz
	OtherStudentIDs      []string
	UnAssignedStudentIDs []string
	AssignedStudentIDs   []string
	UserID               string

	OldToken    string
	ExistedLoID string

	// retrieve learning progress
	From *timestamppb.Timestamp
	To   *timestamppb.Timestamp

	DeletedQuiz *entities.Quiz

	Examples interface{}

	ImgInfos           []*ImgInfo
	FormulasImgInfos   []*FormulasImgInfo
	LearningObjectives []*cpb.LearningObjective

	// exam lo updated fields
	UpdatedGradeToPass    int32
	UpdatedTimeLimit      int32
	UpdatedManualGrading  bool
	UpdatedMaximumAttempt int32
	UpdatedApproveGrading bool
	UpdatedGradeCapping   bool
	UpdatedReviewOption   string

	ExistingQuestionHierarchy entities.QuestionHierarchy
	QuestionGroupID           string
	QuestionTagID             string
	QuestionTagTypeID         string
	QuestionGroupIDs          []string
	GroupedQuizzes            []string

	// Update learning objective name
	LOName string
}

func (s *StepState) initMap() {
	if s.SubmittedKeys == nil {
		s.SubmittedKeys = make(map[string]map[string][]*pb.Answer)
	}
	if s.CheckQuizCorrectnessResponsesByQuizID == nil {
		s.CheckQuizCorrectnessResponsesByQuizID = make(map[string]*pb.CheckQuizCorrectnessResponse)
	}
	if s.expectedCorrectnessByQuizID == nil {
		s.expectedCorrectnessByQuizID = make(map[string][]bool)
	}
	if s.expectedCorrectKeysByQuizID == nil {
		s.expectedCorrectKeysByQuizID = make(map[string][]string)
	}
	if s.expectedIsCorrectAllByQuizID == nil {
		s.expectedIsCorrectAllByQuizID = make(map[string]bool)
	}
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

func generateValidAuthenticationToken(sub string, userGroup string) (string, error) {
	url := ""
	switch userGroup {
	case "USER_GROUP_TEACHER":
		url = "http://" + firebaseAddr + "/token?template=templates/USER_GROUP_TEACHER.template&UserID="
	case "USER_GROUP_SCHOOL_ADMIN":
		url = "http://" + firebaseAddr + "/token?template=templates/USER_GROUP_SCHOOL_ADMIN.template&UserID="
	default:
		url = "http://" + firebaseAddr + "/token?template=templates/phone.template&UserID="
	}

	resp, err := http.Get(url + sub)
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

func (s *suite) aValidAuthenticationToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token, err := generateValidAuthenticationToken(stepState.UserId, "")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anInvalidAuthenticationToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = "invalid-token"
	return StepStateToContext(ctx, stepState), nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(s *suite, ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return helper.GRPCContext(ctx, "token", stepState.AuthToken)
}

func (s *suite) SetFirebaseAddr(firebase string) {
	firebaseAddr = firebase
}

func (s *suite) newID() string {
	return idutil.ULIDNow()
}

func setup(c *common.Config, fakeFirebaseAddr string) {
	firebaseAddr = fakeFirebaseAddr

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	var err error
	rsc := bootstrap.NewResources().WithLoggerC(&c.Common)
	bobConn = rsc.GRPCDial("bob")
	yasuoConn = rsc.GRPCDial("yasuo")
	usermgmtConn = rsc.GRPCDial("usermgmt")
	conn = rsc.GRPCDial("eureka")

	db, _, _ = database.NewPool(context.Background(), zapLogger, c.PostgresV2.Databases["eureka"])
	dbTrace = &database.DBTrace{
		DB: db,
	}
	bobDB, _, _ = database.NewPool(context.Background(), zapLogger, c.PostgresV2.Databases["bob"])
	bobPostgres := c.PostgresV2.Databases["bob"]
	bobPostgres.User = "postgres"
	bobPostgres.Password = c.PostgresMigrate.Database.Password
	bobPgDB, _, _ = database.NewPool(context.Background(), zapLogger, bobPostgres)
	fatimaDB, _, _ = database.NewPool(context.Background(), zapLogger, c.PostgresV2.Databases["fatima"])
	fatimaDBTrace = &database.DBTrace{
		DB: fatimaDB,
	}
	bobDBTrace = &database.DBTrace{
		DB: bobDB,
	}
	shamirConn = rsc.GRPCDial("shamir")

	applicantID = c.JWTApplicant

	jsm, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to create jetstream management: %v", err))
	}

	jsm.ConnectToJS()
	stmt := `
	INSERT INTO organization_auths
		(organization_id, auth_project_id, auth_tenant_id)
	SELECT
		school_id, 'fake_aud', ''
	FROM
		schools
	ON CONFLICT 
		DO NOTHING
	;
	`
	ctx := context.Background()
	_, err = bobDB.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
}

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^a valid authentication token$`:          s.aValidAuthenticationToken,
		`^an invalid authentication token$`:       s.anInvalidAuthenticationToken,
		`^a valid "([^"]*)" token$`:               s.AValidToken,
		`^eureka must store correct study plan$`:  s.eurekaMustStoreCorrectStudyPlan,
		`^user create study plan$`:                s.userCreateStudyPlan,
		`^returns "([^"]*)" status code$`:         s.returnsStatusCode,
		`^eureka must store correct assignments$`: s.eurekaMustStoreCorrectAssignments,
		`^eureka must store correct assignment when create assignment with empty assignment_id$`: s.eurekaMustStoreCorrectAssignmentWhenCreateAssignment,
		`^user create new assignments$`:                      s.userCreateNewAssignments,
		`^a study plan name "([^"]*)" in db$`:                s.aStudyPlanNameInDb,
		`^eureka must store correct study plan item$`:        s.eurekaMustStoreCorrectStudyPlanItem,
		`^eureka must store correct study plan item v(\d+)$`: s.eurekaMustStoreCorrectStudyPlanItemV2,
		`^user upsert a list of study plan item$`:            s.userUpsertAListOfStudyPlanItem,
		`^user upsert a list of study plan item v(\d+)$`:     s.userUpsertAListOfStudyPlanItemV2,
		`^assign assignment to topic$`:                       s.assignAssignmentToTopic,
		`^user try to upsert assignments$`:                   s.userTryToUpsertAssignments,

		`^return a list of chapters$`:       s.returnAListOfChapters,
		`^some chapters are existed in DB$`: s.someChaptersAreExistedInDB,
		`^student list chapters by ids$`:    s.studentListChaptersByIds,

		`^a valid course with class in database$`:                                                         s.aValidCourseWithClassInDatabase,
		`^eureka must assign study plan to course$`:                                                       s.eurekaMustAssignStudyPlanToCourse,
		`^user assign course with study plan$`:                                                            s.userAssignCourseWithStudyPlan,
		`^a valid course and study plan background$`:                                                      s.aValidCourseAndStudyPlanBackground,
		`^user assign study plan to a student$`:                                                           s.userAssignStudyPlanToAStudent,
		`^eureka must assign study plan to student$`:                                                      s.eurekaMustAssignStudyPlanToStudent,
		`^user import a "([^"]*)" study plan to course$`:                                                  s.userImportAStudyPlanToCourse,
		`^eureka must store and assign study plan correctly$`:                                             s.eurekaMustStoreAndAssignStudyPlanCorrectly,
		`^returns a list of study plan items content$`:                                                    s.returnsAListOfStudyPlanItemsContent,
		`^some students are assigned some valid study plans$`:                                             s.someStudentsAreAssignedSomeValidStudyPlans,
		`^students list available contents$`:                                                              s.studentsListAvailableContents,
		`^all study plan items were created with status when upsert new los$`:                             s.allStudyPlanItemsWereCreatedWithStatusWhenUpsertNewLOs,
		`^all study plan items were created with status when upsert new assignments$`:                     s.allStudyPlanItemsWereCreatedWithStatusWhenUpsertNewAssignments,
		`^all study plan items were created with status active after import study plan with type create$`: s.allStudyPlanItemsWereCreatedWithStatusActiveAfterImportStudyPlanWithTypeCreate,

		`^our system must records all the submissions from student "([^"]*)" times$`:                    s.ourSystemMustRecordsAllTheSubmissionsFromStudent,
		`^student submit their "([^"]*)" content assignment "([^"]*)" times$`:                           s.studentSubmitTheirAssignment,
		`^our system must update daily learning time correctly$`:                                        s.ourSystemMustUpdateDailyLearningTimeCorrectly,
		`^teacher submit "([^"]*)" content assignment "([^"]*)" times$`:                                 s.teacherSubmitContentAssignmentTimes,
		`^our system must reject that$`:                                                                 s.ourSystemMustRejectThat,
		`^some students$`:                                                                               s.someStudents,
		`^student submit random assignment$`:                                                            s.studentSubmitRandomAssignment,
		`^unrelated assignments$`:                                                                       s.unrelatedAssignments,
		`^"([^"]*)" list the submissions$`:                                                              s.listTheSubmissions,
		`^teacher list the submissions with status filter$`:                                             s.listTheSubmissionsWithMultiStatusFilter,
		`^our system must returns only with specific status$`:                                           s.ourSystemMustReturnsOnlyLatestSubmissionForSpecificStatus,
		`^our system must returns only latest submission for each assignment$`:                          s.ourSystemMustReturnsOnlyLatestSubmissionForEachAssignment,
		`^"([^"]*)" retrieve some else submissions$`:                                                    s.retrieveSomeElseSubmissions,
		`^"([^"]*)" retrieve their own submissions$`:                                                    s.retrieveTheirOwnSubmissions,
		`^student submit their "([^"]*)" content assignment "([^"]*)" times for different assignments$`: s.studentSubmitTheirAssignmentTimesForDifferentAssignments,
		`^all related study plan items mark as completed$`:                                              s.allRelatedStudyPlanItemsMarkAsCompleted,
		`^our system must update the submissions with latest result$`:                                   s.ourSystemMustUpdateTheSubmissionsWithLatestResult,
		`^our system must update created_at for each latest submission$`:                                s.ourSystemMustUpdateCreatedAtForEachLatestSubmission,
		`^teacher grade the submissions multiple times$`:                                                s.teacherGradeTheSubmissionsMultipleTimes,
		`^teacher grade the submissions multiple times with "([^"]*)" content`:                          s.teacherGradeTheSubmissionsMultipleTimes,

		`^modify not marked status to multi status$`: s.modifyNotMarkedStatusToMultiStatus,

		`^returns a list of assigned study plans of each student$`:     s.returnsAListOfAssignedStudyPlansOfEachStudent,
		`^teacher list study plans for each student$`:                  s.teacherListStudyPlansForEachStudent,
		`^eureka must delete these assignments$`:                       s.eurekaMustDeleteTheseAssignments,
		`^some assignments in db$`:                                     s.someAssignmentsInDb,
		`^user delete assignments$`:                                    s.userDeleteAssignments,
		`^user update assignments$`:                                    s.userUpdateAssignments,
		`^user create assignment with empty assignment_id$`:            s.userCreateAssignmentWithEmptyAssignmentID,
		`^student list active study plan items$`:                       s.studentListActiveStudyPlanItems,
		`^returns a list of completed study plan items$`:               s.returnsAListOfCompletedStudyPlanItems,
		`^student list completed study plan items$`:                    s.studentListCompletedStudyPlanItems,
		`^returns a list of overdue study plan items$`:                 s.returnsAListOfOverdueStudyPlanItems,
		`^student haven\'t completed any study plan items$`:            s.studentHaventCompletedAnyStudyPlanItems,
		`^student list overdue study plan items$`:                      s.studentListOverdueStudyPlanItems,
		`^students list available contents with incorrect filters$`:    s.studentsListAvailableContentsWithIncorrectFilters,
		`^returns a list of "([^"]*)" active study plan items$`:        s.returnsAListOfActiveStudyPlanItems,
		`^returns a list of "([^"]*)" upcoming study plan items$`:      s.returnsAListOfUpcomingStudyPlanItems,
		`^eureka must return assignments correctly$`:                   s.eurekaMustReturnAssignmentsCorrectly,
		`^user list assignments by ids$`:                               s.userListAssignmentsByIds,
		`^some student has their submission graded$`:                   s.someStudentHasTheirSubmissionGraded,
		`^teacher change student\'s submission status to "([^"]*)"$`:   s.teacherChangeStudentsSubmissionStatusTo,
		`^our system must update the submissions status to "([^"]*)"$`: s.ourSystemMustUpdateTheSubmissionsStatusTo,
		`^eureka must return correct grades for each submission$`:      s.eurekaMustReturnCorrectGradesForEachSubmission,
		`^some student has submission with status "([^"]*)"$`:          s.someStudentHasSubmissionWithStatus,
		`^teacher retrieve student grade base on submission grade id$`: s.teacherRetrieveStudentGradeBaseOnSubmissionGradeId,
		`^student retrieve their grade$`:                               s.studentRetrieveTheirGrade,

		`^an valid SyncStudentPackageEvent with ActionKind_ACTION_KIND_UPSERTED$`: s.aValidEvent_Upsert,
		`^an valid SyncStudentPackageEvent with ActionKind_ACTION_KIND_DELETED$`:  s.aValidEvent_Delete,
		`^send "([^"]*)" topic "([^"]*)" to nats js$`:                             s.sendEventToNatsJS,
		`^our system must upsert CourseStudent data correctly$`:                   s.eurekaMustCreateCourseStudent,
		`^our system must update CourseStudent data correctly$`:                   s.eurekaMustUpdateCoursestudent,

		`^an valid JprefMasterRegistration with ActionKind_ACTION_KIND_UPSERTED$`: s.avalideventMasterRegistrationUpsert,
		`^an valid JprefMasterRegistration with ActionKind_ACTION_KIND_DELETED$`:  s.avalideventMasterRegistrationDelete,
		`^our system must upsert CourseClass data correctly$`:                     s.eurekaMustCreateCourseClass,
		`^our system must update CourseClass data correctly$`:                     s.eurekaMustUpdateCourseClass,

		`^an valid JoinClass event$`:                                                 s.aValidEvent_JoinClass,
		`^an valid LeaveClass event$`:                                                s.aValidEvent_LeaveClass,
		`^a valid JoinMasterMgmtClass event$`:                                        s.aValidEvent_JoinMasterMgmtClass,
		`^a valid LeaveMasterMgmtClass event$`:                                       s.aValidEvent_LeaveMasterMgmtClass,
		`^a valid CreateCourseMasterMgmtClass event$`:                                s.aValidEvent_CreateCourseMasterMgmtClass,
		`^our system must upsert ClassMember data correctly$`:                        s.eurekaMustUpsertClassMember,
		`^our system must update ClassMember data correctly$`:                        s.eurekaMustUpdateClassMember,
		`^our system must update MasterMgmtClass data correctly$`:                    s.eurekaMustUpdateMasterMgmtClassMember,
		`^a valid course background$`:                                                s.aValidCourseBackground,
		`^classes belong to course in bob$`:                                          s.createClassInBob,
		`^user list class by course$`:                                                s.userListClassByCourse,
		`^user list class by course and locations$`:                                  s.userListClassByCourseAndLocations,
		`^user list class by course and not exist locations$`:                        s.userListClassByCourseAndNotExistLocations,
		`^a signed in "([^"]*)"$`:                                                    s.aSignedIn,
		`^eureka must return correct list of class ids$`:                             s.eurekaMustReturnCorrectListOfClassIds,
		`^eureka must return nil list of class ids$`:                                 s.eurekaReturnNilListOfClassIds,
		`^valid book in bob$`:                                                        s.validBookInBob,
		`^user duplicate book$`:                                                      s.userDuplicateBook,
		`^valid assignment in current book$`:                                         s.validAssignmentInCurrentBook,
		`^eureka must duplicate all assignments$`:                                    s.eurekaMustDuplicateAllAssignments,
		`^"([^"]*)" retrieve their submission grade$`:                                s.retrieveTheirSubmissionGrade,
		`^user import a individual study plan to a student$`:                         s.userImportAIndividualStudyPlanToAStudent,
		`^eureka must store and assign study plan for individual student correctly$`: s.eurekaMustStoreAndAssignStudyPlanForIndividualStudentCorrectly,
		`^delete all classes belong to course$`:                                      s.deleteAllClassesBelongToCourse,

		`^user update course study plan$`:                                                              s.userUpdateCourseStudyPlan,
		`^user update course study plan with times$`:                                                   s.userUpdateCourseStudyPlanWithTimes,
		`^user update course study plan with times with study_plan_items don\'t belong to study_plan$`: s.userUpdateCourseStudyPlanWithStudyPlanItemsDonotBelongToStudyPlan,
		`^user import a generated study plan to course$`:                                               s.userImportAGeneratedStudyPlanToCourse,
		`^make study plan item completed$`:                                                             s.makeStudyPlanItemCompleted,
		`^study plan item still completed$`:                                                            s.studyPlanItemStillCompleted,
		`^user individual study plan must be update$`:                                                  s.userIndividualStudyPlanMustBeUpdate,
		`^valid assignment in db$`:                                                                     s.validAssignmentInDb,

		`^"([^"]*)" list the submissions with assignment name$`:           s.listTheSubmissionsWithAssignmentName,
		`^"([^"]*)" list the submissions with invalid assignment name$`:   s.listTheSubmissionsWithInvalidAssignmentName,
		`^our system must returns empty submission$`:                      s.ourSystemMustReturnsEmptySubmission,
		`^our system must returns submission with valid assignment name$`: s.ourSystemMustReturnsSubmissionWithValidAssignmentName,
		`^a book without content in bob$`:                                 s.aBookWithoutContentInBob,
		// list student course
		`^a valid course student background$`:                            s.aValidCourseStudentBackground,
		`^user list student by course$`:                                  s.CallListStudentByCourse,
		`^eureka must return correct list of basic profile of students$`: s.eurekaMustReturnCorrectListOfBasicProfile,
		`^user list student by course two times with paging$`:            s.callMultiListStudentByCoursePaging,
		`^user list student by course with search_text and paging`:       s.CallListStudentByCourseWithSearchText,
		`^a Japanese student`:                                            s.aJapaneseStudent,
		//
		`^some student has their submission haven\'t graded$`: s.someStudentHasTheirSubmissionHaventGraded,
		`^grade infomations have to included to submissions`:  s.gradeInfomationHaveToIncludedToSubmissions,

		// check quiz correctness
		`^user create a study plan of exam lo to database$`:                                                                         s.userCreateAStudyPlanOfExamLoToDatabase,
		`^a quiz test "([^"]*)" fill in the blank quizzes with "([^"]*)" quizzes per page and do quiz test$`:                        s.aQuizTestFillInTheBlankQuizzesWithQuizzesPerPageAndDoQuizTest,
		`^a quiz test include "([^"]*)" multiple choice quizzes with "([^"]*)" quizzes per page and do quiz test$`:                  s.aQuizTestIncludeMultipleChoiceQuizzesWithQuizzesPerPageAndDoQuizTest,
		`^a quiz test include "([^"]*)" pair of word quizzes with "([^"]*)" quizzes per page and do quiz test$`:                     s.aQuizTestIncludePairOfWordQuizzesWithQuizzesPerPageAndDoQuizTest,
		`^returns expected result pair of word quizzes$`:                                                                            s.returnsExpectedResultPairOfWordQuizzes,
		`^returns expected result pair of word quizzes for submit quiz answers$`:                                                    s.returnsExpectedResultPairOfWordQuizzesForSubmitQuizAnswers,
		`^student answer pair of word quizzes$`:                                                                                     s.studentAnswerPairOfWordQuizzes,
		`^student answer pair of word quizzes for submit quiz answers$`:                                                             s.studentAnswerPairOfWordQuizzesForSubmitQuizAnswers,
		`^a quiz test include "([^"]*)" term and definition quizzes with "([^"]*)" quizzes per page and do quiz test$`:              s.aQuizTestIncludeTermAndDefinitionQuizzesWithQuizzesPerPageAndDoQuizTest,
		`^returns expected result term and definition quizzes$`:                                                                     s.returnsExpectedResultTermAndDefinitionQuizzes,
		`^returns expected result term and definition quizzes for submit quiz answers$`:                                             s.returnsExpectedResultTermAndDefinitionQuizzesForSubmitQuizAnswers,
		`^student answer term and definition quizzes$`:                                                                              s.studentAnswerTermAndDefinitionQuizzes,
		`^student answer term and definition quizzes for submit quiz answers$`:                                                      s.studentAnswerTermAndDefinitionQuizzesForSubmiteQuizAnswers,
		`^a quiz test include "([^"]*)" with "([^"]*)" quizzes with "([^"]*)" quizzes per page and do quiz test$`:                   s.aQuizTestIncludeWithQuizzesWithQuizzesPerPageAndDoQuizTest,
		`^student answer "([^"]*)" quizzes$`:                                                                                        s.studentAnswerQuizzes,
		`^returns expected result "([^"]*)" quizzes$`:                                                                               s.returnsExpectedResultQuizzes,
		`^a quiz test of learning objective belong to "([^"]*)" topic include "([^"]*)" quizzes with "([^"]*)" quizzes every page$`: s.aQuizTestOfLearningObjectiveBelongToTopicIncludeQuizzesWithQuizzesEveryPage,
		`^student choose option "([^"]*)" of the quiz "([^"]*)"$`:                                                                   s.studentChooseOptionOfTheQuiz,
		`^student missing quiz id in request$`:                                                                                      s.studentMissingQuizIdInRequest,
		`^student fill text "([^"]*)" of the quiz "([^"]*)"$`:                                                                       s.studentFillTextOfTheQuiz,
		`^student fill text "([^"]*)" of the quiz "([^"]*)" for submit quiz answers$`:                                               s.studentFillTextOfTheQuizForSubmitQuizAnswers,
		`^returns expected result multiple choice type$`:                                                                            s.returnsExpectedResultMultipleChoiceType,
		`^returns expected result fill in the blank type$`:                                                                          s.returnsExpectedResultFillInTheBlankType,
		`^returns expected result fill in the blank type for submit quiz answers$`:                                                  s.returnsExpectedResultFillInTheBlankTypeForSubmitQuizAnswers,
		`^student choose option "([^"]*)"$`:                                                                                         s.studentChooseOption,
		`^returns isCorrectAll: "([^"]*)"$`:                                                                                         s.returnsIsCorrectAll,
		`^a FIB quiz test with case sensitive config and correct answers "([^"]*)"$`:                                                s.aFIBQuizTestWithCaseSensitiveConfigAndCorrectAnswers,
		`^a quiz test with partial config on and test case\'s data with "([^"]*)" correct, "([^"]*)" not correct$`:                  s.aQuizTestWithPartialConfigOnAndTestCasesDataWithCorrectNotCorrect,
		`^a quiz test with partial config off and test case\'s data with "([^"]*)" correct, "([^"]*)" not correct$`:                 s.aQuizTestWithPartialConfigOffAndTestCasesDataWithCorrectNotCorrect,
		`^a FIB quiz test with partial config on and correct answers "([^"]*)"$`:                                                    s.aFIBQuizTestWithPartialConfigOnAndCorrectAnwsers,
		`^student fill in text "([^"]*)"$`:                                                                                          s.studentFillInText,
		`^this is absolutely an "([^"]*)" answer with isCorrectAll: "([^"]*)"$`:                                                     s.thisIsAbsolutelyAnAnswerWithIsCorrectAll,
		`^a FIB quiz test with no config and correct answers "([^"]*)"$`:                                                            s.aFIBQuizTestWithNoConfigAndCorrectAnswers,
		`^a FIB quiz test with case sensitive and partial config and correct answers "([^"]*)"$`:                                    s.aFIBQuizTestWithCaseSensitiveAndPartialConfigAndCorrectAnswers,
		`^user create a learning objective$`:                                                                                        s.userCreateALearningObjective,
		`^user upsert a topic$`:                                                                                                     s.useUpsertATopic,
		`^returns expected result fill in the blank quiz$`:                                                                          s.returnsExpectedResultFillInTheBlankQuiz,
		`^student answer fill in the blank quiz with ocr$`:                                                                          s.studentAnswerFillInTheBlankQuizWithOcr,
		`^a quiz test "([^"]*)" "([^"]*)" quizzes with "([^"]*)" quizzes per page and do quiz test$`:                                s.aQuizTestQuizzesWithQuizzesPerPageAndDoQuizTest,
		`^returns result all correct in submit quiz answers for ordering question$`:                                                 s.returnsResultAllCorrectInSubmitQuizAnswersForOrderingQuestion,
		`^student answer correct order options for all quizzes$`:                                                                    s.studentAnswerCorrectOrderingOptionOfTheQuizForSubmitQuizAnswers,
		`^student finish essay questions$`:                                                                                          s.studentFinishEssay,
		`^returns essay quiz answers$`:                                                                                              s.returnsEssaySubmitQuizAnswer,

		// student event finish an lo
		`^a valid student account$`: s.aValidStudentAccount,

		`^returns study plan progress of students correctly$`:           s.returnsStudyPlanProgressOfStudentsCorrectly,
		`^teacher retrieves study plan progress of students$`:           s.teacherRetrievesStudyPlanProgressOfStudents,
		`^our system must update null content for each submission`:      s.ourSystemMustUpdateNullContentForEachSubmission,
		`^our system must update the submissions with null content$`:    s.ourSystemMustReturnsNullGradeContent,
		`^"([^"]*)" list the submissions with course id$`:               s.listTheSubmissionsWithCourseId,
		`^our system must returns submission with valid courses$`:       s.ourSystemMustReturnsSubmissionWithValidCourses,
		`^our system must stores correctly$`:                            s.ourSystemMustStoresCorrectly,
		`^some student has their submission have graded and commented$`: s.someStudentHasTheirSubmissionGraded,

		`^course study plan and individual student study plan must be update$`: s.courseStudyPlanAndIndividualStudentStudyPlanMustBeUpdate,
		`^user remove one row from study plan$`:                                s.userRemoveOneRowFromStudyPlan,
		`^user insert one row to study plan$`:                                  s.userInsertOneRowToStudyPlan,

		`^individual study plan should be update$`:                                                  s.individualStudyPlanShouldBeUpdate,
		`^user update individual study plan$`:                                                       s.userUpdateIndividualStudyPlan,
		`^user download a student\'s study plan$`:                                                   s.userDownloadAStudentsStudyPlan,
		`^our system must remove all course student study plan$`:                                    s.ourSystemMustRemoveAllCourseStudentStudyPlan,
		`^our system must create new study plan for each course student$`:                           s.ourSystemMustCreateNewStudyPlanForEachCourseStudent,
		`^some students of different courses are assigned some valid study plans$`:                  s.someStudentsOfDifferentCoursesAreAssignedSomeValidStudyPlans,
		`^wait for assign study plan task to completed$`:                                            s.waitForAssignStudyPlanTaskToCompleted,
		`^student is remove from a class after they submit their submission$`:                       s.studentIsRemoveFromAClassAfterTheySubmitTheirSubmission,
		`^the response submissions don\'t contain submission of student who is removed from class$`: s.theResponseSubmissionsDontContainSubmissionOfStudentWhoIsRemovedFromClass,

		`^an student package with "([^"]*)"$`:                                        s.anStudentPackageWith,
		`^our system have to handle correctly$`:                                      s.ourSystemHaveToHandleCorrectly,
		`^the admin add a new student package with a package or courses$`:            s.theAdminAddANewStudentPackageWithAn,
		`^the admin add a student package and update location_id$`:                   s.theAdminAddAStudentPackageAndUpdateLocationID,
		`^our system have to updated course student access paths correctly$`:         s.ourSystemHaveToUpdatedCourseStudentAccessPathsCorrectly,
		`^the admin toggle student package status$`:                                  s.theAdminToggleStudentPackageStatus,
		`^courseStudentAccessPaths were created$`:                                    s.courseStudentAccessPathsWereCreated,
		`^user change study plan item order$`:                                        s.userChangeStudyPlanItemOrder,
		`^returns a list of empty study plan items content$`:                         s.returnsAListOfEmptyStudyPlanItemsContent,
		`^students list available contents with "([^"]*)" course id$`:                s.studentsListAvailableContentsWithCourseId,
		`^some students are assigned some study plan with available from "([^"]*)"$`: s.someStudentsAreAssignedSomeStudyPlanWithAvailableFrom,
		`^returns study plan progress of students are (\d+)$`:                        s.returnsStudyPlanProgressOfStudentsAre,

		`^a "([^"]*)" status code is returned$`: s.returnsStatusCode,

		`^list course-study plan\'s to do items$`:                    s.listCourseStudyPlansToDoItems,
		`^returns list of to do items with correct statistic infor$`: s.returnsListOfToDoItemsWithCorrectStatisticInfor,
		`^returns empty list of to do items$`:                        s.returnsEmptyListOfToDoItems,
		`^list course-study plan\'s to do items with "([^"]*)"$`:     s.listCourseStudyPlansToDoItemsWith,
		`^delete study plan items by study plans$`:                   s.deleteStudyPlanItemsByStudyPlans,

		`^a course and some study plans$`:                                 s.avalidCourseAndSomeStudyPlanBackground,
		`^the teacher list study plan by course$`:                         s.listStudyPlanByCourse,
		`^teacher archives some study plans$`:                             s.teacherArchivesSomeStudyPlans,
		`^our system have to return list study plan by course correctly$`: s.ourSystemHaveToReturnListStudyPlanByCourseCorrectly,

		`^a course and assigned this course to some students$`:         s.aCourseAndAssignedThisCourseToSomeStudents,
		`^our system have to return child study plan items correctly$`: s.ourSystemHaveToReturnChildStudyPlanItemsCorrectly,
		`^teacher get child study plan items$`:                         s.teacherGetChildStudyPlanItems,

		`^some study plans to individual$`:                 s.someStudyPlanToIndividual,
		`^our system have to update study plan correctly$`: s.ourSystemHaveToUpdateStudyPlanCorrectly,
		`^the user update study plans$`:                    s.theUserUpdateTheStudyPlan,

		`^the teacher retrieve statistic assignment class$`:                s.theTeacherRetrieveStatisticAssignmentClass,
		`^our system have to return statistic assignment class correctly$`: s.ourSystemHaveToReturnStatisticAssignmentClassCorrectly,
		`^some students join in a class$`:                                  s.someStudentsJoinInAClass,
		`^some students submit their assignments$`:                         s.someStudentsSubmitTheirAssignments,
		`^our system have to handle error study plan correctly$`:           s.ourSystemHaveToHandleErrorStudyPlanCorrectly,

		`^an "([^"]*)" status code is returned after release$`: s.returnsStatusCode,

		`^teacher remove student submission$`: s.teacherRemoveStudentSubmission,

		`^our system must delete student submission correctly$`:              s.ourSystemMustDeleteStudentSubmissionCorrectly,
		`^teacher remove student submission after list submissions$`:         s.teacherRemoveStudentSubmissionAfterListSubmissions,
		`^the response submissions don\'t contain submissions were deleted$`: s.responseSubmissionsDontContainSubmissionsWereDeleted,

		`^an learning objectives created event$`:                             s.anLearningObjectivesCreatedEvent,
		`^our system must update study plan items correctly$`:                s.ourSystemMustUpdateStudyPlanItemsCorrectly,
		`^our system receives learning objectives created event$`:            s.ourSystemReceivesLearningObjectivesCreatedEvent,
		`^user try to upsert "([^"]*)" learning objectives using APIv(\d+)$`: s.userTryToUpsertLearningObjectivesUsingAPIV1,
		`^Add new study plans$`:                                              s.addNewStudyPlans,
		`^All study plans were inserted$`:                                    s.allStudyPlansWereInserted,
		`^user get list study plans and filter with new order collation$`:    s.userGetListStudyPlansAndFilterWithNewOrderCollation,

		`^an assignments created event$`:                                 s.anAssignmentsCreatedEvent,
		`^our system receives assignments created event$`:                s.ourSystemReceivesAssignmentsCreatedEvent,
		`^assign assignments to topic$`:                                  s.assignAssignmentsToTopic,
		`^our system must update assignment study plan items correctly$`: s.ourSystemMustUpdateAssignmentStudyPlanItemsCorrectly,

		`^all study plan item has empty start date$`: s.allStudyPlanItemHasEmptyStartDate,

		`^retrieve assignments with that topic$`:                   s.retrieveAssignmentsWithThatTopic,
		`^returns a assignment list with different display order$`: s.returnsAAssignmentListWithDifferentDisplayOrder,
		`^user create some assignments same topic and time$`:       s.userCreateSomeAssignmentsSameTopicAndTime,

		`^retrieve assignments$`:                                 s.retrieveAssignments,
		`^returns assignment list with display order correctly$`: s.returnsAssignmentListWithDisplayOrderCorrectly,
		`^user create assignments with display order$`:           s.userCreateAssignmentsWithDisplayOrder,
		`^user create assignments without display order$`:        s.userCreateAssignmentsWithoutDisplayOrder,

		`^a validated book with chapters and topics$`:         s.aValidatedBookWithChaptersAndTopics,
		`^user list topics by study plan$`:                    s.userListTopicsByStudyPlan,
		`^verify topic data after list topics by study plan$`: s.verifyTopicDataAfterListTopicsByStudyPlan,
		`^user create a valid study plan with "([^"]*)"$`:     s.userCreateAValidStudyPlanWith,

		`^study plans and related items have been stored$`:              s.studyPlansAndRelatedItemsHaveBeenStored,
		`^study plans have been updated$`:                               s.studyPlansHaveBeenUpdated,
		`^user create a valid study plan$`:                              s.userCreateAValidStudyPlan,
		`^user update a study plan with invalid study_plan_id$`:         s.userUpdateAStudyPlanWithInvalidStudy_plan_id,
		`^user update study plan$`:                                      s.userUpdateStudyPlan,
		`^user add a book to course does not have any students$`:        s.userAddABookToCourseDoesNotHaveAnyStudents,
		`^add a valid book with some learning objectives to course$`:    s.addAValidBookWithSomeLearningObjectivesToCourse,
		`^user add a valid book does not have any "([^"]*)" to course$`: s.userAddAValidBookDoesNotHaveAnyToCourse,

		`^returns todo items have order correctly$`:       s.returnsTodoItemsHaveOrderCorrectly,
		`^user add leaning objectives "([^"]*)" to book$`: s.userAddLeaningObjectivesToBook,
		`^user get list todo items by topics$`:            s.userGetListTodoItemsByTopics,

		`^user get list todo items by topics with invalid study plan id$`: s.userGetListTodoItemsByTopicsWithInvalidStudyPlanId,
		`^user get list todo items by topics with null study plan id$`:    s.userGetListTodoItemsByTopicsWithNullStudyPlanId,

		`^returns todo items total correctly with status "([^"]*)"$`:    s.returnsTodoItemsTotalCorrectlyWithStatus,
		`^user create some valid study plan$`:                           s.userCreateSomeValidStudyPlan,
		`^user retrieve list student todo items with status "([^"]*)"$`: s.userRetrieveListStudentTodoItemsWithStatus,
		`^user update study plans status to "([^"]*)"$`:                 s.userUpdateStudyPlansStatusTo,
		`^user update study plan items status to "([^"]*)"$`:            s.userUpdateStudyPlanItemsStatusTo,
		`^a valid course student$`:                                      s.aValidCourseStudent,

		`^a study plan does not have any study plan items$`: s.aStudyPlanDoesNotHaveAnyStudyPlanItems,
		`^book of study plan has stored correctly$`:         s.bookOfStudyPlanHasStoredCorrectly,

		`^user add a leaning objective and an assignment with same topic to book$`: s.userAddALeaningObjectiveAndAnAssignmentWithSameTopicToBook,
		`^user delete a "([^"]*)" in book$`:                                        s.userDeleteAInBook,
		`^update dates of study plan items$`:                                       s.updateDatesOfStudyPlanItems,

		`^user get list todo items by topics with available dates$`: s.userGetListTodoItemsByTopicsWithAvailableDates,
		`^update available dates for study plan items$`:             s.updateAvailableDatesForStudyPlanItems,

		`^"([^"]*)" list submissions using v(\d+) with valid locations$`:                                 s.listSubmissionsUsingVWithValidLocations,
		`^"([^"]*)" list submissions using v(\d+) with invalid locations$`:                               s.listSubmissionsUsingVWithInvalidLocations,
		`^"([^"]*)" list submissions using v(\d+) with some valid locations and some invalid locations$`: s.listSubmissionsUsingVWithSomeValidLocationsAndSomeInvalidLocations,
		`^our system must returns list submissions correctly$`:                                           s.ourSystemMustReturnsListSubmissionsCorrectly,
		`^some students added to course in some valid locations$`:                                        s.someStudentsAddedToCourseInSomeValidLocations,
		`^students are assigned assignments in study plan$`:                                              s.studentsAreAssignedAssignmentsInStudyPlan,
		`^students submit their assignments$`:                                                            s.studentsSubmitTheirAssignments,
		`^our system must returns list submissions is empty$`:                                            s.ourSystemMustReturnsListSubmissionsIsEmpty,
		`^"([^"]*)" list submissions using v(\d+) with valid locations and course is null$`:              s.listSubmissionsUsingVWithValidLocationsAndCourseIsNull,
		`^a list submissions of students with random locations$`:                                         s.aListSubmissionsOfStudentsWithRandomLocations,
		`^a student expired in course$`:                                                                  s.aStudentExpiredInCourse,

		`^a learning objective belonged to a "([^"]*)" topic has no quizset$`:                 s.aLearningObjectiveBelongedToATopicHasNoQuizset,
		`^a learning objective belonged to a "([^"]*)" topic has quizset with (\d+) quizzes$`: s.aLearningObjectiveBelongedToATopicHasQuizsetWithQuizzes,
		`^total quiz set is "([^"]*)"$`:                                                       s.totalQuizSetIs,
		`^user get total quiz of lo "([^"]*)"$`:                                               s.userGetTotalQuizOfLo,
		`^user get total quiz of lo "([^"]*)" with role$`:                                     s.userGetTotalQuizOfLoWithRole,
		`^user get total quiz of lo without lo ids$`:                                          s.userGetTotalQuizOfLoWithoutLoIds,
		`^a study plan item id$`:                                                              s.aStudyPlanItemId,
		`^returns "([^"]*)" study_set_id$`:                                                    s.returnsStudySetID,
		`^retrieve last flashcard study progress with "([^"]*)" arguments$`:                   s.retrieveLastFlashcardStudyProgressWithArguments,

		`^user create flashcard study with "([^"]*)" and "([^"]*)" flashcard study quizzes per page$`: s.userCreateFlashcardStudyWithAndFlashcardStudyQuizzesPerPage,
		`^retrieve flashcard study progress with "([^"]*)" arguments$`:                                s.retrieveFlashcardStudyProgressWithArguments,
		`^returns expected flashcard study progress$`:                                                 s.returnsExpectedFlashcardStudyProgress,
		`^flashcard study progress response match with the response from bob service$`:                s.flashcardStudyProgressResponseMatch,
		`^last flashcard study progress response match with the response from bob service$`:           s.lastFlashcardStudyProgressResponseMatch,

		`^"([^"]*)" has created an empty book$`:                                             s.hasCreatedAnEmptyBook,
		`^"([^"]*)" has created a content book$`:                                            s.hasCreatedAContentBook,
		`^"([^"]*)" has created a studyplan exact match with the book content for student$`: s.hasCreatedAStudyplanExactMatchWithTheBookContentForStudent,
		`^"([^"]*)" has created a studyplan exact match with the book empty for student$`:   s.hasCreatedAStudyplanExactMatchWithTheBookEmptyForStudent,
		`^another school admin logins$`:                                                     s.anotherSchoolAdminLogins,
		`^"([^"]*)" create study plan from the book$`:                                       s.createStudyPlanFromTheBook,
		`^"([^"]*)" delete lo study plan item$`:                                             s.deleteLoStudyPlanItem,
		`^"([^"]*)" add student to the course$`:                                             s.enrollToTheCourse,
		`^our system has to delete lo study plan items correctly$`:                          s.ourSystemHasToDeleteLoStudyPlanItemsCorrectly,
		`^"([^"]*)" logins "([^"]*)"$`:                                                      s.logins,
		`^school admin has created a "([^"]*)" book$`:                                       s.hasCreatedABook,
		// `^school admin downloads the book from course$`:                                     s.downloadsTheBookFromCourse,
		`^school admin adds the newly created book for course$`:                            s.addsTheNewlyCreatedBookForCourse,
		`^school admin has added the new course for student$`:                              s.schoolAdminhasAddedTheNewCourseForStudent,
		`^school admin has created a new course$`:                                          s.schoolAdminHasCreatedANewCourse,
		`^our system returns "([^"]*)" status code$`:                                       s.returnsStatusCode,
		`^"([^"]*)" logins$`:                                                               s.logins,
		`^a random number$`:                                                                s.aRandomNumber,
		`^a school name "([^"]*)", country "([^"]*)", city "([^"]*)", district "([^"]*)"$`: s.aSchoolNameCountryCityDistrict,
		`^admin inserts schools$`:                                                          s.adminInsertsSchools,

		// add books to course
		`^an "([^"]*)" add books request$`:                 s.anAddBooksRequest,
		`^user try to add books to course$`:                s.userTryToAddBooksToCourse,
		`^our system must adds books to course correctly$`: s.ourSystemMustAddsBooksToCourseCorrectly,

		// add assignment after remove items from book
		`^remove all items from book$`:                                               s.removeAllItemsFromBook,
		`^upsert assignment into book$`:                                              s.upsertAssignmentsIntoBook,
		`^upsert learning objective into book$`:                                      s.upsertLearningObjectiveIntoBook,
		`^study plan items belong to assignments were successfully created$`:         s.studyPlanItemsBelongsToAssignmentsWereSuccesfullyCreated,
		`^study plan items belong to learning objectives were successfully created$`: s.studyPlanItemsBelongsToLOsWereSuccesfullyCreated,

		// delete assignment update
		`^assignment was successfully deleted in system$`:                   s.assignmentWasSuccessfullyDeletedInSystem,
		`^school admin delete assignment$`:                                  s.schoolAdminDeleteAssignment,
		`^study plan items belong to assignment were successfully deleted$`: s.studyPlanItemsBelongToAssignmentWereSuccessfullyDeleted,

		// update assignment time
		`^admin re-update assignment time with "([^"]*)"$`:                                 s.adminReupdateAssignmentTimeWith,
		`^admin update assignment time with "([^"]*)"$`:                                    s.adminUpdateAssignmentTimeWith,
		`^admin update assignment time with null data and "([^"]*)"$`:                      s.adminUpdateAssignmentTimeWithNullDataAnd,
		`^assignment time was updated with according update_type "([^"]*)"$`:               s.assignmentTimeWasUpdatedWithAccordingUpdate_type,
		`^assignment time was updated with new data and according update_type "([^"]*)"$`:  s.assignmentTimeWasUpdatedWithNewDataAndAccordingUpdate_type,
		`^assignment time was updated with null data and according update_type "([^"]*)"$`: s.assignmentTimeWasUpdatedWithNullDataAndAccordingUpdate_type,

		// calculate student progress
		`^"([^"]*)" has created a book with each (\d+) los, (\d+) assignments, (\d+) topics, (\d+) chapters, (\d+) quizzes$`:                           s.hasCreatedABookWithEachLosAssignmentsTopicsChaptersQuizzes,
		`^"([^"]*)" do test and done "([^"]*)" los with "([^"]*)" correctly and "([^"]*)" assignments with "([^"]*)" point and skip "([^"]*)" topics$`: s.doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics,
		`^topic score is "([^"]*)" and chapter score is "([^"]*)"$`:                                                                                    s.topicScoreIsAndChapterScoreIs,
		`^first pair topic score is "([^"]*)" and second pair topic score is "([^"]*)" and chapter score is "([^"]*)"$`:                                s.firstPairTopicScoreIsAndSecondPairTopicScoreIsAndChapterScoreIs,
		`^calculate student progress$`:                        s.calculateStudentProgress,
		`^calculate student progress with missing "([^"]*)"$`: s.calculateStudentProgressWithMissing,
		`^correct lo completed with "([^"]*)" and "([^"]*)"$`: s.correctLoCompletedWithAnd,
		`^school admin delete "([^"]*)" topics$`:              s.schoolAdminDeleteTopics,
		`^"([^"]*)" do test and done "([^"]*)" los with "([^"]*)" correctly and "([^"]*)" assignments with "([^"]*)" point in the first two topics and done "([^"]*)" los with "([^"]*)" correctly and "([^"]*)" assignments with "([^"]*)" point in the other$`: s.doTestAndDoneLosWithCorrectlyAndAssignmentsWithPointInFourTopics,
		`^some of created assignments are task assignment$`:       s.someOfCreatedAssignmentAreTaskAssignment,
		`^student retry do wrong quizzes with "([^"]*)" correct$`: s.studentRetryDoQuizzes,
		`^remove last quiz from lo$`:                              s.removeLastQuizFromLo,
		// Re calculate student progress
		`^school admin delete a los$`:                                          s.schoolAdminDeleteALos,
		`^teacher retrieve student progress$`:                                  s.teacherRetrieveStudentProgress,
		`^our system have to return student progress correctly$`:               s.ourSystemHaveToReturnStudentProgressCorrectly,
		`^topic learning objectives of deleted los were successfully deleted$`: s.topicLearningObjectivesOfDeletedLosWereSuccessfullyDeleted,

		// db admin add student study plans
		`^a course study plan created$`: s.aCourseStudyPlansCreated,
		`^the database admin create some study plans for a which have "([^"]*)" master study plan$`: s.theDatabaseAdminCreateSomeStudentStudyPlans,
		`^our database have to handle <"([^"]*)"> correctly$`:                                       s.ourDatabaseHaveToRaiseAnErrorViolateUniqueIndex,

		// delete study plan
		`^user chooses a study plan in course for deleting$`:      s.userChoosesAStudyPlanInCourseForDeleting,
		`^user deletes selected study plan$`:                      s.userDeletesSelectedStudyPlan,
		`^user fetchs new study plans$`:                           s.fetchsNewStudyPlans,
		`^user fetchs old study plans$`:                           s.fetchsOldStudyPlans,
		`^Check selected study plan has been absolutely deleted$`: s.checkSelectedStudyPlanHasBeenAbsolutelyDeleted,

		// upsert books
		`^user upsert valid books$`:                    s.userUpsertBooks,
		`^user creates new "([^"]*)" books$`:           s.userCreateNewBooks,
		`^our system must stores correct books$`:       s.ourSystemMustStoresCorrectBooks,
		`^there are books existed$`:                    s.thereAreBooksExisted,
		`^user updates "([^"]*)" books$`:               s.userUpdatesBooks,
		`^our system must update the books correctly$`: s.ourSystemMustUpdateTheBooksCorrectly,
		`^user has created an empty book$`:             s.userHasCreatedAnEmptyBook,

		`^a valid book in db$`:               s.aValidBookInDB,
		`^eureka must return copied topics$`: s.eurekaMustReturnCopiedTopics,
		`^user send duplicate book request$`: s.userSendDuplicateBookRequest,

		// update status study plan item
		`^user update status with valid request`: s.userUpdateStatusWithValidRequest,
		`^update status with "([^"]*)"$`:         s.updateStatusWith,
		`^return with error "([^"]*)"$`:          s.returnWithError,

		// update school date study plan items
		`^return error "([^"]*)"$`: s.returnError,
		`^return successful with updated record with school date "([^"]*)"$`: s.returnSuccessfulWithUpdatedRecordWithSchoolDate,
		`^update school date with missing school date$`:                      s.updateSchoolDateWithMissingSchoolDate,
		`^update school date with missing student id$`:                       s.updateSchoolDateWithMissingStudentId,
		`^update school date with missing study plan item ids$`:              s.updateSchoolDateWithMissingStudyPlanItemIds,
		`^update school date with valid request$`:                            s.updateSchoolDateWithValidRequest,
		`^user update school date with valid request$`:                       s.userUpdateSchoolDateWithValidRequest,
		// upsert lo assignments
		`^user create new los and assignments$`:     s.userCreateNewLosAndAssignments,
		`^a school name "([^"]*)"$`:                 s.aSchoolName,
		`^new los and assignments must be created$`: s.newLOsAndAssignmentsMustBeCreated,

		// upsert flash-card v2
		`^user create flashcard study with valid request and limit "([^"]*)" the first time$`: s.userCreateFlashcardStudyTestWithValidRequestAndLimitTheFirstTime,
		`^user finish flashcard study without restart$`:                                       s.userFinishFlashcardStudyWithoutRestart,
		`^user finish flashcard study without restart and remembered questions$`:              s.userFinishFlashcardStudyWithoutRestartAndRememberedQuestions,
		`^user finish flashcard study with restart$`:                                          s.userFinishFlashcardStudyWithRestart,
		`^verify data after finish flashcard without restart$`:                                s.verifyDataAfterFinishFlashcardWithoutRestart,
		`^verify data after finish flashcard with restart$`:                                   s.verifyDataAfterFinishFlashcardWithRestart,

		`^a quizset with "([^"]*)" quizzes in Learning Objective belonged to a "([^"]*)" topic$`:  s.aQuizsetWithQuizzesInLearningObjectiveBelongedToATopic,
		`^a quiz set with "([^"]*)" quizzes in Learning Objective belonged to a "([^"]*)" topic$`: s.aQuizsetWithQuizzesInLearningObjectiveBelongedToATopic,

		`^return a list of books$`:       s.returnAListOfBooks,
		`^some books are existed in DB$`: s.someBooksAreExistedInDB,
		`^user list books by ids$`:       s.userListBooksByIds,
		`^student list books by ids$`:    s.studentListBooksByIds,

		`^a signed in admin$`:               s.aSignedInAdmin,
		`^user create a quiz using v(\d+)$`: s.userCreateAQuizV2UsingV2,
		`^study plan items for student of los and assignments must be created$`: s.studyPlanItemsForStudentOfLOsAndAssignmentsMustBeCreated,
		`^user update display orders for los and assignments$`:                  s.userUpdateDisplayOrdersForLOsAndAssignments,
		`^display order of los and assignments must be updated$`:                s.displayOrderOfLOsAndAssignmentsMustBeUpdated,
		`^user create los and assignments$`:                                     s.userCreateLosAndAssignments,

		`^data for list student available contents with "([^"]*)" book\(s\)$`:                   s.dataForListStudentAvailableContentsWithBooks,
		`^list student available contents$`:                                                     s.listStudentAvailableContents,
		`^verify list contents after list student available contents with "([^"]*)" book\(s\)$`: s.verifyListContentsAfterListStudentAvailableContentsWithBooks,

		// get lo highest scores
		`^get lo highest scores$`:                                    s.getLoHighestScores,
		`^return correct highest scores belong to study plan items$`: s.returnCorrectHighestScoresBelongToStudyPlanItems,

		`^study plan items wrong book_id$`: s.studyPlanItemsWrongBookID,

		// upsert topic
		`^our system have to save the topics correctly$`:                      s.ourSystemHaveToSaveTheTopicsCorrectly,
		`^school admin has create some "([^"]*)" topics$`:                     s.schoolAdminHasCreateSomeTopics,
		`^school admin has created some topics before$`:                       s.schoolAdminHasCreatedSomeTopicsBefore,
		`^our system have to store the topics in concurrency correctly$`:      s.ourSystemHaveToStoreTheTopicsInConcurrencyCorrectly,
		`^two school admin create some topics$`:                               s.twoSchoolAdminCreateSomeTopics,
		`^our system have to save topics on both old and new flow correctly$`: s.ourSystemHaveToSaveTopicsOnBothOldAndNewFlowCorrectly,
		`^school admin has created some topics before by old flow$`:           s.schoolAdminHasCreatedSomeTopicsBeforeByOldFlow,
		`^user has created some valid topics$`:                                s.userHasCreatedSomeValidTopics,
		`^user has created some "([^"]*)" topics$`:                            s.userHasCreateSomeTopics,

		// publish topics
		`^user public some missing topics$`: s.userPublicSomeMissingTopics,
		`^user public some topics$`:         s.userSomePublicTopics,

		// delete topics
		`^some missing topic ids$`:                      s.someMissingTopicIds,
		`^user delete some topics$`:                     s.userDeleteSomeTopics,
		`^user delete some topics with role$`:           s.userDeleteSomeTopicsWithRole,
		`^our system must delete the topics correctly$`: s.ourSystemMustDeleteTheTopicsCorrectly,

		// upsert chapter
		`^user creates new "([^"]*)" chapters$`:           s.userCreatesNewChapters,
		`^our system must stores correct chapters$`:       s.ourSystemMustStoresCorrectChapters,
		`^our system must update the chapters correctly$`: s.ourSystemMustUpdateTheChaptersCorrectly,
		`^there are chapters existed$`:                    s.thereAreChaptersExisted,
		`^user upsert valid chapters$`:                    s.userUpsertValidChapters,
		`^user updates "([^"]*)" chapters$`:               s.userUpdatesChapters,
		`^user create a valid chapter$`:                   s.userCreateAValidChapter,

		// delete chapters
		`^some missing chapter ids$`:                      s.someMissingChapterIds,
		`^user delete some chapters$`:                     s.userDeleteSomeChapters,
		`^our system must delete the chapters correctly$`: s.ourSystemMustDeleteTheChaptersCorrectly,

		// upsert quiz
		`^user upsert "([^"]*)" quiz$`:                          s.userUpsertQuiz,
		`^user upsert a "([^"]*)" quiz$`:                        s.userUpsertAQuiz,
		`^user upsert a "([^"]*)" quiz with role$`:              s.userUpsertQuizWithRole,
		`^user upsert "([^"]*)" single quiz$`:                   s.userUpsertSingleQuiz,
		`^user upsert a valid "([^"]*)" single quiz$`:           s.userUpsertAValidSingleQuiz,
		`^user upsert a valid "([^"]*)" single quiz with role$`: s.userUpsertAValidSingleQuizWithRole,
		`^quiz created successfully with new version$`:          s.quizCreatedSuccessfullyWithNewVersion,
		`^quiz_set also updated with new version$`:              s.quizSetAlsoUpdatedWithNewVersion,
		`^learning objective belonged to a topic$`:              s.learningObjectiveBelongedToTopic,
		`^a question tag existed in database$`:                  s.aQuestionTagExistedInDatabase,
		`^a question tag type existed in database$`:             s.aQuestionTagTypeExistedInDatabase,

		`^"([^"]*)" get book ids belong to student study plan items$`:                   s.getBookIdsBelongToStudentStudyPlanItems,
		`^our system has to get book ids belong to student study plan items correctly$`: s.ourSystemHasToGetBookIdsBelongToStudentStudyPlanItemsCorrectly,

		`^user create a course with a study plan$`:      s.userCreateACourseWithAStudyPlan,
		`^study plan of student have stored correctly$`: s.studyPlanOfStudentHaveStoredCorrectly,
		`^user add course to student$`:                  s.userAddCourseToStudent,

		`^"([^"]*)" do assignment$`:                                                               s.doAssignment,
		`^"([^"]*)" grade submission with status returned$`:                                       s.gradeSubmissionWithStatusReturned,
		`^notification has been stored correctly$`:                                                s.notificationHasBeenStoredCorrectly,
		`^"([^"]*)" add student to course$`:                                                       s.addStudentToCourse,
		`^"([^"]*)" create a study plan with book have an assignment$`:                            s.createAStudyPlanWithBookHaveAnAssignment,
		`^"([^"]*)" has created some studyplans exact match with some books content for student$`: s.hasCreatedSomeStudyplansExactMatchWithSomeBooksContentForStudent,
		`^"([^"]*)" has created some studyplans exact match with the book content for student$`:   s.hasCreatedSomeStudyplansExactMatchWithTheBookContentForStudent,
		`^study plan items have created correctly$`:                                               s.studyPlanItemsHaveCreatedCorrectly,
		`^study plan items have created on assignments created correctly$`:                        s.studyPlanItemsHaveCreatedOnAssignmentsCreatedCorrectly,
		`^user creates some los in book$`:                                                         s.userCreatesSomeLosInBook,
		`^user creates some assignments in book$`:                                                 s.userCreatesSomeAssignmentsInBook,
		`^user creates some los in books$`:                                                        s.userCreatesSomeLosInBooks,
		`^user creates some assignments in books$`:                                                s.userCreateSomeAssignmentsInBooks,

		`^a list students logins "([^"]*)"$`:                        s.aListStudentsLogins,
		`^"([^"]*)" add students to course$`:                        s.addStudentsToCourse,
		`^"([^"]*)" grade submissions with status in progress$`:     s.gradeSubmissionsWithStatusInProgress,
		`^"([^"]*)" update student submissions status to returned$`: s.updateStudentSubmissionsStatusToReturned,
		`^notifications has been stored correctly$`:                 s.notificationsHasBeenStoredCorrectly,
		`^students do assignment$`:                                  s.studentsDoAssignment,

		`^"([^"]*)" add some students to the course$`:  s.schoolAdminAddSomeStudent,
		`^our monitor save missing student correctly$`: s.ourMonitorSaveMissingStudentCorrectly,
		`^some student\'s study plans not created$`:    s.someStudentsStudyPlansNotCreated,
		`^run monitor upsert course student$`:          s.runMonitorUpsertCourseStudent,

		`^our monitor save missing learning item correctly$`:        s.ourMonitorSaveMissingLearningItemCorrectly,
		`^run monitor upsert learning item$`:                        s.runMonitorUpsertLearningItem,
		`^some study plan items not created$`:                       s.someStudyPlanItemsNotCreated,
		`^our monitor auto upsert missing learning item correctly$`: s.ourMonitorAutoUpsertMissingLearningItemCorrectly,

		`^notification has not been stored$`: s.notificationHasNotBeenStored,
		`^update student school to null$`:    s.updateStudentSchoolToNull,

		`^retrieve study plan item event logs$`:                          s.RetrieveStudyPlanItemEventLogs,
		`^some learning_objective student event logs are existed in DB$`: s.someLearning_objectiveStudentEventLogsAreExistedInDB,
		`^an assigned student$`:                                          s.anAssignedStudent,
		`^user create quiz test$`:                                        s.userCreateQuizTest,
		`^school admin add student to a course have a study plan$`:       s.schoolAdminAddStudentToACourseHaveAStudyPlan,
		`^student create quiz test$`:                                     s.studentCreateQuizTest,
		`^our system must returns quizzes correctly$`:                    s.ourSystemMustReturnsQuizzesCorrectly,
		`^quiz test have question hierarchy$`:                            s.quizTestHaveQuestionHierarchy,
		// retrieve quiz tests
		`^(\d+) quiz tests infor$`:                           s.quizTestsInfor,
		`^"([^"]*)" students do test of a study plan item$`:  s.studentsDoQuizExamOfAStudyPlanItem,
		`^teacher get quiz test without study plan item id$`: s.teacherGetQuizTestWithoutStudyPlanItemId,
		`^teacher get quiz test of a study plan item$`:       s.teacherGetQuizTestOfAStudyPlanItem,
		`^compare quiz tests list with bob service$`:         s.compareQuizTestsListWithBobService,

		`^return list of "([^"]*)" flashcard study items$`:                                                                                 s.returnListOfFlashcardStudyItems,
		`^returns empty flashcard study items$`:                                                                                            s.returnsEmptyFlashcardStudyItems,
		`^returns expected list of flashcard study quizzes with "([^"]*)"$`:                                                                s.returnsExpectedListOfFlashcardStudyQuizzesWith,
		`^student doing a long exam with "([^"]*)" flashcard study quizzes per page$`:                                                      s.studentDoingALongExamWithFlashcardStudyQuizzesPerPage,
		`^that student can fetch the list of flashcard study quizzes page by page using limit "([^"]*)" flashcard study quizzes per page$`: s.thatStudentCanFetchTheListOfFlashcardStudyQuizzesPageByPageUsingLimitFlashcardStudyQuizzesPerPage,
		`^user create flashcard study with valid request and offset "([^"]*)" and limit "([^"]*)"$`:                                        s.userCreateFlashcardStudyWithValidRequestAndOffsetAndLimit,
		`^user create flashcard study without loID$`:                                                                                       s.userCreateFlashcardStudyWithoutLoID,
		`^user create flashcard study without paging$`:                                                                                     s.userCreateFlashcardStudyWithoutPaging,
		`^a list of learning_objective event logs$`:                                                                                        s.aListOfLearningObjectiveEventLogs,
		`^a student inserts a list of event logs$`:                                                                                         s.aStudentInsertsAListOfEventLogs,
		`^achievement crown "([^"]*)" must be (\d+)$`:                                                                                      s.achievementCrownMustBe,
		`^an other student profile in DB$`:                                                                                                 s.anOtherStudentProfileInDB,
		`^his owned student UUID$`:                                                                                                         s.hisOwnedStudentUUID,
		// `^some preset study plans is existed in DB$`:                  s.somePresetStudyPlansIsExistedInDB,
		`^student finishes "([^"]*)" unassigned learning objectives$`:                            s.studentFinishesUnassignedLearningObjectives,
		`^student retrieves preset study plans$`:                                                 s.studentRetrievesPresetStudyPlans,
		`^total_learning_time must be "([^"]*)"$`:                                                s.totalLearningTimeMustBe,
		`^total_learning_time v2 must be "([^"]*)"$`:                                             s.totalLearningTimeMustBeV2,
		`^total_lo_finished must be "([^"]*)"$`:                                                  s.totalLoFinishedMustBe,
		`^user retrieves student stats$`:                                                         s.userRetrievesStudentStats,
		`^user create flashcard study with valid request and limit "([^"]*)" in the first time$`: s.userCreateFlashcardStudyTestWithValidRequestAndLimitInTheFirstTime,

		`^a learning objective is existed in DB$`:                                                                   s.aLearningObjectiveIsExistedInDB,
		`^a list of quiz_finished event logs$`:                                                                      s.aListOfQuiz_finishedEventLogs,
		`^a list of quiz_finished event logs with correctness is (\d+)$`:                                            s.aListOfQuiz_finishedEventLogsWithCorrectnessIs,
		`^a list of study_guide_finished event logs$`:                                                               s.aListOfStudy_guide_finishedEventLogs,
		`^a list of video_finished event logs$`:                                                                     s.aListOfVideo_finishedEventLogs,
		`^a student retries the last finished learning objective$`:                                                  s.aStudentRetriesTheLastFinishedLearningObjective,
		`^Eureka must record all student\'s event logs$`:                                                            s.eurekaMustRecordAllStudentsEventLogs,
		`^"([^"]*)" completeness must be "([^"]*)"$`:                                                                s.completenessMustBe,
		`^student finishes "([^"]*)" assigned learning objectives of the "([^"]*)" week$`:                           s.studentFinishesAssignedLearningObjectivesOfTheWeek,
		`^student finishes tutorial lo$`:                                                                            s.studentFinishesTutorialLo,
		`^student inserts a list of learning_objective event logs then sleeping "([^"]*)"$`:                         s.studentInsertsAListOfLearning_objectiveEventLogsThenSleeping,
		`^student inserts a list of learning_objective event logs with session id empty then sleeping "([^"]*)"$`:   s.studentInsertsAListOfLearning_objectiveEventLogsWithSessionIdEmptyThenSleeping,
		`^student inserts a list of learning_objective event logs without completed event then sleeping "([^"]*)"$`: s.studentInsertsAListOfLearning_objectiveEventLogsWithoutCompletedEventThenSleeping,
		`^total_learning_time must not be existed$`:                                                                 s.total_learning_timeMustNotBeExisted,
		`^total_lo_finished must not be updated$`:                                                                   s.total_lo_finishedMustNotBeUpdated,
		`^waiting for "([^"]*)"$`:                                                                                   s.waitingFor,
		`^a valid book content$`:                                                                                    s.aValidBookContent,

		`^learning objectives must be created$`:                                          s.learningObjectivesMustBeCreated,
		`^learning objectives must be updated$`:                                          s.learningObjectivesMustBeUpdated,
		`^user create learning objectives$`:                                              s.userCreateLearningObjectives,
		`^user update learning objectives$`:                                              s.userUpdateLearningObjectives,
		`^there is no quizset that contains deleted quiz$`:                               s.thereIsNoQuizsetThatContainsDeletedQuiz,
		`^user delete a quiz "([^"]*)"$`:                                                 s.userDeleteAQuiz,
		`^user delete a quiz without quiz id$`:                                           s.userDeleteAQuizWithoutQuizId,
		`^user create (\d+) learning objectives with type "([^"]*)"$`:                    s.userCreateNLearningObjectivesWithType,
		`^user update "([^"]*)" of learning objectives$`:                                 s.userUpdateLearningObjectivesFields,
		`^"([^"]*)" of learning objectives must be updated$`:                             s.fieldOfLearningObjectsMustBeUpdated,
		`^user create (\d+) learning objectives with default values and type "([^"]*)"$`: s.userCreateNLearningObjectivesWithDefaultValuesAndType,
		`^learning objectives must be created with "([^"]*)" as default value$`:          s.learningObjectivesMustBeCreatedWithFieldsAsDefaultValue,
		// Retrieve student submission history by lo_ids
		`^student retrieve submission history by lo_ids$`: s.studentRetrieveSubmissionHistoryByLo_ids,

		`^flashcard study progress must be updated$`:                 s.flashcardStudyProgressMustBeUpdated,
		`^update flashcard study progress with "([^"]*)" arguments$`: s.updateFlashcardStudyWithArguments,

		// Create retry quiz test
		`^our system have to return the retry quizzes correctly$`: s.ourSystemHaveToReturnTheRetryQuizzesCorrectly,
		`^student does the quiz set and wrong some quizzes$`:      s.studentDoesTheQuizSetAndWrongSomeQuizzes,
		`^the student choose option retry quiz$`:                  s.theStudentChooseOptionRetryQuiz,

		`^show correct logs info$`:                                                                                         s.showCorrectLogsInfo,
		`^the ordered of logs must be correct$`:                                                                            s.theOrderedOfLogsMustBeCorrect,
		`^teacher retrieve all student\'s submission history in that study plan item$`:                                     s.teacherRetrieveAllStudentsSubmissionHistoryInThatStudyPlanItem,
		`^student retrieve all student\'s submission history in that study plan item$`:                                     s.studentRetrieveAllStudentsSubmissionHistoryInThatStudyPlanItem,
		`^a study plan item is learning objective belonged to a "([^"]*)" topic which has quizset with "([^"]*)" quizzes$`: s.aStudyPlanItemIsLearningObjectiveBelongedToATopicWhichHasQuizsetWithQuizzes,
		`^get "([^"]*)" student\'s submission history$`:                                                                    s.getStudentsSubmissionHistory,
		`^each item have returned addition fields for flashcard$`:                                                          s.eachItemHaveReturnedAdditionFieldsForFlashcard,
		`^"([^"]*)" students didn\'t finish the test$`:                                                                     s.studentsDidntFinishTheTest,
		`^LO does not contain deleted quiz$`:                                                                               s.lODoesNotContainDeletedQuiz,
		`^user remove a quiz "([^"]*)" from lo$`:                                                                           s.userRemoveAQuizFromLo,
		`^user remove a quiz without lo id$`:                                                                               s.userRemoveAQuizWithoutLoId,
		`^user remove a quiz without quiz id$`:                                                                             s.userRemoveAQuizWithoutQuizId,
		`^a list of los created$`:                                                                                          s.aListOfLosCreated,
		`^los have been deleted correctly$`:                                                                                s.losHaveBeenDeletedCorrectly,
		`^user delete los$`:                                                                                                s.userDeleteLos,
		`^user delete los again$`:                                                                                          s.userDeleteLosAgain,

		// Retrieve learning progress
		`^a filter range with "([^"]*)" is "([^"]*)"$`:                                   s.aFilterRangeWithIs,
		`^a "([^"]*)" learning objective event log with session "([^"]*)" at "([^"]*)"$`: s.aLearningObjectiveEventLogWithSessionAt,

		`^a valid filter range$`:                                                            s.aValidFilterRange,
		`^an invalid filter range$`:                                                         s.anInvalidFilterRange,
		`^previous request data is reset$`:                                                  s.previousRequestDataIsReset,
		`^returns LP with all total_time_spent_in_day equal to zero$`:                       s.returnsLPWithAllTotal_time_spent_in_dayEqualToZero,
		`^returns LP with some total_time_spent_in_day larger than zero$`:                   s.returnsLPWithSomeTotal_time_spent_in_dayLargerThanZero,
		`^student hasn\'t learned any LO$`:                                                  s.studentHasNotLearnedAnyLO,
		`^student retrieves LP$`:                                                            s.studentRetrievesLP,
		`^total_learning_time at "([^"]*)" must be "([^"]*)"$`:                              s.total_learning_timeAtMustBe,
		`^a signed in "student" with filter range is "([^"]*)"$`:                            s.aSignedInStudentWithFilterRangeIs,
		`^a signed in "student" with filter range is "([^"]*)" use his owned student UUID$`: s.aSignedInStudentWithFilterRangeIsUseHisOwnedStudentUUID,
		`^a signed in "student" with use his owned student UUID$`:                           s.aSignedInStudentWithUseHisOwnedStudentUUID,
		`^returns expected "([^"]*)" list of quizzes$`:                                      s.returnsExpectedListOfQuizzes,
		`^returns expected totalQuestion when retrieveLO$`:                                  s.returnsExpectedTotalQuestionWhenRetrieveLO,
		`^returns expected totalQuestion when retrieveLOV(\d+)$`:                            s.returnsExpectedTotalQuestionWhenRetrieveLOV1,
		`^teacher get quizzes of "([^"]*)" lo$`:                                             s.teacherGetQuizzesOfLo,
		`^user remove a quiz without "([^"]*)"$`:                                            s.userRemoveAQuizWithout,
		`^a list of valid learning objectives$`:                                             s.aListOfValidLearningObjectives,
		`^user try to assign topic items$`:                                                  s.userTryToAssignTopicItems,
		`^user try to assign topic items with role$`:                                        s.userTryToAssignTopicItemsWithRole,
		`^user try to assign topic items with invalid request$`:                             s.userTryToAssignTopicItemsWithInvalidRequest,
		`^a list of valid topics$`:                                                          s.aListOfValidTopics,
		`^admin inserts a list of valid topics$`:                                            s.adminInsertsAListOfValidTopics,

		// Add adhoc assignment
		`^add adhoc assignment "([^"]*)"$`:       s.addAdHocAssignment,
		`^add adhoc assignment with "([^"]*)"$`:  s.addAdHocAssignmentWith,
		`^our system must add adhoc assignment$`: s.ourSystemMustAddAdhocAssignment,

		// Update display orders of quizset
		`^user change order with "([^"]*)" times in quiz set$`: s.userChangeOrderWithTimesInQuizSet,
		`^user move one quiz "([^"]*)" in quiz set$`:           s.userMoveOneQuizInQuizSet,
		`^update the order quizzes in quiz set as expected$`:   s.updateTheOrderQuizzesInQuizSetAsExpected,

		// Retrieve course statistic
		`^teacher retrieve course statistic$`:                                                         s.retrieveCourseStatistic,
		`^<course_statistical_v2>teacher retrieve course statistic$`:                                  s.retrieveCourseStatisticV2,
		`^(\d+) students logins "([^"]*)"$`:                                                           s.studentsLogins,
		`^"([^"]*)" has created a studyplan exact match with the book content for all login student$`: s.hasCreatedAStudyplanExactMatchWithTheBookContentForAllLoginStudent,
		`^(\d+) students do test and done "([^"]*)" los with "([^"]*)" correctly and "([^"]*)" assignments with "([^"]*)" point and skip "([^"]*)" topics$`: s.someStudentDoTestAndDoneLosWithCorrectlyAndAssignmentsWithPointAndSkipTopics,
		`^some of created study plan item are archived$`:                                                                      s.someOfCreatedStudyPlanItemAreArchived,
		`^some students are members of some classes$`:                                                                         s.someStudentsAreMembersOfSomeClasses,
		`^our system returns correct course statistic$`:                                                                       s.returnsCorrectCourseStatisticItems,
		`^<course_statistical_v2>our system returns correct course statistic$`:                                                s.returnsCorrectCourseStatisticItemsV2,
		`^topic total assigned student is (\d+), completed students is (\d+), average score is (\d+)$`:                        s.topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIs,
		`^<course_statistical_v2>topic total assigned student is (\d+), completed students is (\d+), average score is (\d+)$`: s.topicTotalAssignedStudentIsCompletedStudentsIsAverageScoreIsV2,

		`^a list of document images$`:              s.aListOfDocumentImages,
		`^detect text from document images$`:       s.detectTextFromDocumentImages,
		`^our system must return texts correctly$`: s.ourSystemMustReturnTextsCorrectly,

		`^a list of formula images$`:                  s.aListOfFormulaImages,
		`^detect formula from images$`:                s.detectFormulaFromImages,
		`^our system must return formulas correctly$`: s.ourSystemMustReturnFormulasCorrectly,

		`^some lo completenesses existed in db$`:                 s.someLoCompletenessesExistedInDB,
		`^retrieve learning objectives with "([^"]*)"$`:          s.retrieveLearningObjectivesWith,
		`^our system must return learning objectives correctly$`: s.ourSystemMustReturnLearningObjectivesCorrectly,
		// student package event v2
		`^a student package v(\d+) with "([^"]*)"$`:                                       s.aStudentPackageVWith,
		`^the admin add a new student package v(\d+) with a package or courses$`:          s.theAdminAddANewStudentPackageVWithAPackageOrCourses,
		`^the admin add a student package v(\d+) and update location_id$`:                 s.theAdminAddAStudentPackageVAndUpdateLocation_id,
		`^the admin toggle student package v(\d+) status$`:                                s.theAdminToggleStudentPackageVStatus,
		`^courseStudentAccessPaths were created for v(\d+)$`:                              s.courseStudentAccessPathsWereCreatedForV,
		`^our system have to handle student package v(\d+) correctly$`:                    s.ourSystemHaveToHandleStudentPackageVCorrectly,
		`^our system have to updated course student access paths correctly for v(\d+)$`:   s.ourSystemHaveToUpdatedCourseStudentAccessPathsCorrectlyForV,
		`^returns expected result multiple choice type for submit quiz answers$`:          s.returnsExpectedResultMultipleChoiceTypeForSubmitQuizAnswers,
		`^student choose option "([^"]*)" of the quiz "([^"]*)" for submit quiz answers$`: s.studentChooseOptionOfTheQuizForSubmitQuizAnswers,
		`^student submit quiz answers$`:                                                   s.studentSubmitQuizAnswers,
		`^teacher retrieves LP$`:                                                          s.teacherRetrievesLP,

		// question group
		`^existing question group$`:                               s.existingQuestionGroup,
		`^<(\d+)> existing question group$`:                       s.existingQuestionGroups,
		`^<(\d+)> quiz belong to question group$`:                 s.quizBelongToQuestionGroup,
		`^user got quiz test response$`:                           s.userGetQuizTestResponse,
		`^question group existed in response$`:                    s.questionGroupReturnedInResp,
		`^our system must add adhoc assignment correctly$`:        s.ourSystemMustAddAdhocAssignmentCorrectly,
		`^insert question group with "([^"]*)" rich description$`: s.insertQuestionGroupWithRichDescription,

		// Update learning objective name
		`^user update learning objective name$`:                      s.userUpdateLearningObjectiveName,
		`^our system must update learning objective name correctly$`: s.ourSystemMustUpdateLearningObjectiveNameCorrectly,

		// Retrieve map lm id and study plan item id by study plan id
		`^user retrieve map lm id and study plan item id$`:                    s.userRetrieveLMStudyPlanItemID,
		`^our system must return map lm id and study plan item id correctly$`: s.ourSystemMustReturnLMAndStudyPlanItemsCorrectly,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})

	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func newSuite() *suite {
	return &suite{
		Conn:    conn,
		DB:      db,
		DBTrace: dbTrace,
		StepState: &StepState{
			StudentsSubmittedAssignments: make(map[string][]*pb.Content),
		},
		BobConn:       bobConn,
		YasuoConn:     yasuoConn,
		UsermgmtConn:  usermgmtConn,
		BobDB:         bobDB,
		BobPgDB:       bobPgDB,
		BobDBTrace:    bobDBTrace,
		FatimaDB:      fatimaDB,
		FatimaDBTrace: fatimaDBTrace,
		JSM:           jsm,
		ZapLogger:     zapLogger,
		ApplicantID:   applicantID,
		ShamirConn:    shamirConn,
	}
}

func (s *suite) signedCtx(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return StepStateToContext(helper.GRPCContext(ctx, "token", stepState.AuthToken), stepState)
}

func (s *suite) aSignedInStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := idutil.ULIDNow()
	var err error
	ctx, err = s.aValidUser(ctx, id, consta.RoleStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUserInDB error: %v", err)
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateExchangeToken error: %v", err)
	}
	randomID := idutil.ULIDNow()
	password := fmt.Sprintf("password-%v", randomID)
	email := fmt.Sprintf("%v@example.com", randomID)
	name := randomID
	req := &upb.CreateStudentRequest{
		SchoolId: constant.ManabieSchool,
		StudentProfile: &upb.CreateStudentRequest_StudentProfile{
			Email:            email,
			Password:         password,
			Name:             name,
			CountryCode:      cpb.Country_COUNTRY_VN,
			PhoneNumber:      fmt.Sprintf("phone-number-%v", randomID),
			Grade:            5,
			EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
		},
	}
	resp, err := upb.NewUserModifierServiceClient(s.UsermgmtConn).CreateStudent(s.signedCtx(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudentID = resp.StudentProfile.Student.UserProfile.UserId
	stepState.CurrentStudentID = stepState.StudentID
	stepState.AuthToken, err = s.generateExchangeToken(stepState.StudentID, entities.UserGroupStudent)
	stepState.StudentToken = stepState.AuthToken
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateExchangeToken(userID, userGroup string) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", err
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.ApplicantID, constants.ManabieSchool, s.ShamirConn)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *suite) aSignedInAdmin(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	var err error
	ctx, err = s.aValidUser(ctx, id, "USER_GROUP_ADMIN")
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUserInDB error: %v", err)
	}
	stepState.AuthToken, err = s.generateExchangeToken(id, "USER_GROUP_ADMIN")
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateExchangeToken error: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) waitingFor(ctx context.Context, arg1 string) (context.Context, error) {
	d, err := time.ParseDuration(arg1)
	if err != nil {
		return ctx, err
	}
	time.Sleep(d)
	return ctx, nil
}

func (s *suite) setFakeClaimToContext(ctx context.Context, resourcePath string, userGroup string) context.Context {
	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
			UserGroup:    userGroup,
		},
	}
	return interceptors.ContextWithJWTClaims(ctx, claims)
}

func (s *suite) aValidUser(ctx context.Context, id string, group string) (context.Context, error) {
	var oldGroup string
	switch group {
	case consta.RoleTeacher:
		oldGroup = "USER_GROUP_TEACHER"
	case consta.RoleStudent:
		oldGroup = "USER_GROUP_STUDENT"
	case consta.RoleSchoolAdmin:
		oldGroup = "USER_GROUP_SCHOOL_ADMIN"
	case consta.RoleParent:
		oldGroup = "USER_GROUP_PARENT"
	default:
		oldGroup = "USER_GROUP_STUDENT"
	}
	ctx, err := s.aValidUserInDB(ctx, s.BobDBTrace, id, group, oldGroup)
	if err != nil {
		return ctx, err
	}
	ctx, err = s.aValidUserInDB(ctx, s.DBTrace, id, group, group)
	if err != nil {
		return ctx, err
	}
	ctx, err = s.aValidUserInDB(ctx, s.FatimaDBTrace, id, group, group)
	if err != nil {
		return ctx, err
	}
	return ctx, err
}

func (s *suite) aValidUserInDB(ctx context.Context, dbConn *database.DBTrace, id, newgroup, oldGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	num := rand.Int()
	var now pgtype.Timestamptz
	now.Set(time.Now())
	u := entities_bob.User{}
	database.AllNullEntity(&u)
	u.ID = database.Text(id)
	u.LastName.Set(fmt.Sprintf("valid-user-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num))
	u.Country.Set(bpb.COUNTRY_VN.String())
	u.Group.Set(oldGroup)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = now
	u.UpdatedAt = now
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)
	u.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	gr := &entities_bob.Group{}
	database.AllNullEntity(gr)
	gr.ID.Set(oldGroup)
	gr.Name.Set(oldGroup)
	gr.UpdatedAt.Set(time.Now())
	gr.CreatedAt.Set(time.Now())
	fieldNames, _ := gr.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	stmt := fmt.Sprintf("INSERT INTO groups (%s) VALUES(%s) ON CONFLICT DO NOTHING", strings.Join(fieldNames, ","), placeHolders)
	if _, err := dbConn.Exec(ctx, stmt, database.GetScanFields(gr, fieldNames)...); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert group error: %v", err)
	}
	ctx = s.setFakeClaimToContext(context.Background(), u.ResourcePath.String, oldGroup)

	ugroup := &entity.UserGroupV2{}
	database.AllNullEntity(ugroup)
	ugroup.UserGroupID.Set(idutil.ULIDNow())
	ugroup.UserGroupName.Set("name")
	ugroup.UpdatedAt.Set(time.Now())
	ugroup.CreatedAt.Set(time.Now())
	ugroup.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	ugMember := &entity.UserGroupMember{}
	database.AllNullEntity(ugMember)
	ugMember.UserID.Set(u.ID)
	ugMember.UserGroupID.Set(ugroup.UserGroupID.String)
	ugMember.CreatedAt.Set(time.Now())
	ugMember.UpdatedAt.Set(time.Now())
	ugMember.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	uG := entities_bob.UserGroup{
		UserID:   u.ID,
		GroupID:  database.Text(oldGroup),
		IsOrigin: database.Bool(true),
	}
	uG.Status.Set("USER_GROUP_STATUS_ACTIVE")
	uG.CreatedAt = u.CreatedAt
	uG.UpdatedAt = u.UpdatedAt

	role := &entity.Role{}
	database.AllNullEntity(role)
	role.RoleID.Set(idutil.ULIDNow())
	role.RoleName.Set(newgroup)
	role.CreatedAt.Set(time.Now())
	role.UpdatedAt.Set(time.Now())
	role.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	grantedRole.RoleID.Set(role.RoleID.String)
	grantedRole.UserGroupID.Set(ugroup.UserGroupID.String)
	grantedRole.GrantedRoleID.Set(idutil.ULIDNow())
	grantedRole.CreatedAt.Set(time.Now())
	grantedRole.UpdatedAt.Set(time.Now())
	grantedRole.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool))

	if _, err := database.InsertOnConflictDoNothing(ctx, &u, dbConn.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user error: %v", err)
	}

	if _, err := database.InsertOnConflictDoNothing(ctx, &uG, dbConn.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group error: %v", err)
	}
	if _, err := database.InsertOnConflictDoNothing(ctx, ugroup, dbConn.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.InsertOnConflictDoNothing(ctx, ugMember, dbConn.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.InsertOnConflictDoNothing(ctx, role, dbConn.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}
	if _, err := database.InsertOnConflictDoNothing(ctx, grantedRole, dbConn.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}
	if u.Group.String == constant.UserGroupTeacher {
		teacher := &entities_bob.Teacher{}
		database.AllNullEntity(teacher)

		err := multierr.Combine(
			teacher.ID.Set(u.ID.String),
			teacher.SchoolIDs.Set([]int64{constant.ManabieSchool}),
			teacher.UpdatedAt.Set(time.Now()),
			teacher.CreatedAt.Set(time.Now()),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.InsertOnConflictDoNothing(ctx, teacher, dbConn.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("insert teacher error: %v", err)
		}
	}

	if u.Group.String == constant.UserGroupStudent {
		stepState.StudentID = u.ID.String
		stepState.CurrentStudentID = u.ID.String

		student := &entities_bob.Student{}
		database.AllNullEntity(student)
		err := multierr.Combine(
			student.ID.Set(u.ID.String),
			student.SchoolID.Set(constant.ManabieSchool),
			student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
			student.StudentNote.Set("example-student-note"),
			student.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
			student.CurrentGrade.Set(12),
			student.OnTrial.Set(true),
			student.TotalQuestionLimit.Set(10),
			student.BillingDate.Set(now),
			student.UpdatedAt.Set(time.Now()),
			student.CreatedAt.Set(time.Now()),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.InsertOnConflictDoNothing(ctx, student, dbConn.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("insert student error: %v", err)
		}
	}

	if u.Group.String == constant.UserGroupSchoolAdmin {
		schoolAdminAccount := &entities_bob.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		err := multierr.Combine(
			schoolAdminAccount.SchoolAdminID.Set(u.ID.String),
			schoolAdminAccount.SchoolID.Set(u.ResourcePath.String),
			schoolAdminAccount.UpdatedAt.Set(time.Now()),
			schoolAdminAccount.CreatedAt.Set(time.Now()),
			schoolAdminAccount.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.InsertOnConflictDoNothing(ctx, schoolAdminAccount, dbConn.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser insert school error: %w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidStudentAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	id := idutil.ULIDNow()
	s.aValidUser(ctx, id, consta.RoleStudent)
	stepState.AuthToken, err = s.generateExchangeToken(id, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentStudentID = id

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidStudentInDB(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	num := idutil.ULIDNow()

	student := &bob_entities.Student{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.User)
	database.AllNullEntity(&student.User.AppleUser)

	err := multierr.Combine(
		student.ID.Set(id),
		student.LastName.Set(fmt.Sprintf("valid-student-%s", num)),
		student.Country.Set(bpb.COUNTRY_VN.String()),
		student.PhoneNumber.Set(fmt.Sprintf("phone-number+%s", id)),
		student.Email.Set(fmt.Sprintf("email+%s", id)),
		student.CurrentGrade.Set(rand.Intn(12)),
		student.TargetUniversity.Set("TG11DT"),
		student.TotalQuestionLimit.Set(5),
		student.SchoolID.Set(constant.ManabieSchool),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = database.ExecInTx(ctx, s.BobDB, func(ctx context.Context, tx pgx.Tx) error {
		if err := (&repositories.StudentRepo{}).Create(ctx, tx, student); err != nil {
			return errors.Wrap(err, "s.StudentRepo.CreateTx")
		}

		if student.AppleUser.ID.String != "" {
			if err := (&repositories.AppleUserRepo{}).Create(ctx, tx, &student.AppleUser); err != nil {
				return errors.Wrap(err, "s.AppleUserRepo.Create")
			}
		}
		return nil
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentStudentID = id
	stepState.StudentID = id

	return StepStateToContext(ctx, stepState), nil
}

func SetFirebaseAddr(fireBaseAddr string) {
	firebaseAddr = fireBaseAddr
}
