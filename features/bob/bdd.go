package bob

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/features/common"
	test_usermgmt "github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	notiEntities "github.com/manabie-com/backend/internal/notification/entities"
	usermgmt_repo "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/cucumber/godog"
	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	natsJS "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	common.RegisterTest("bob", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

// fields
const (
	Name            = "name"
	Country         = "country"
	SchoolID        = "schoolID"
	Subject         = "subject"
	Grade           = "grade"
	CountryAndGrade = "country and grade"
	All             = "all"
	None            = "none"
)

var firebaseAddr string

var (
	flagCheckRetrieveLesson  int32
	scenarioOfRetrieveLesson = []string{
		"school admin retrieve live lesson without filter",
	}
)

var (
	flagCheckImportMasterData  int32
	scenarioOfImportMasterData = []string{
		"Import csv data for table preset_study_plan",
		"Import csv data for table preset_study_plan",
		"Unauthenticated admin try to import",
		"Import csv master data for table learning objective",
		"Import csv vn data for table learning objective",
		"Import csv invalid data for table learning objective",
		"Unauthenticate admin try to import topic",
		"Import csv master data for table topics",
		"Import csv master data for table topics",
		"Import csv master data for table topics has chapter id",
	}
)

var (
	flagCheckUploadTopicCSV int32
	zapLogger               *zap.Logger
	rootAccount             map[int]common.AuthInfo
)

// suite must not contain any global variables
type suite struct {
	Cfg                  *common.Config
	DB                   database.Ext
	DBPostgres           database.Ext
	AuthDB               *pgxpool.Pool
	EurekaDB             database.Ext
	Conn                 *grpc.ClientConn
	YsConn               *grpc.ClientConn
	UsermgmtConn         *grpc.ClientConn
	FatimaConn           *grpc.ClientConn
	EurekaConn           *grpc.ClientConn
	VirtualClassroomConn *grpc.ClientConn
	FirebaseClient       *auth.Client
	ZapLogger            *zap.Logger
	ApplicantID          string
	ShamirConn           *grpc.ClientConn
	GCPApp               *gcp.App
	FirebaseAuthClient   internal_auth_tenant.TenantClient
	TenantManager        internal_auth_tenant.TenantManager
	KeycloakClient       *internal_auth.IdentityServiceImpl
	JSM                  nats.JetStreamManagement
	// Unleash
	UnleashClient           unleashclient.ClientInstance
	UnleashLocalAdminAPIKey string
	UnleashSrvAddr          string

	// *common.StepState

	RootAccount     map[int]common.AuthInfo
	FirebaseAddress string
	CommonSuite     *common.Suite
	*common.StepState
	LessonmgmtConn *grpc.ClientConn
}

func (s *suite) newID() string {
	return idutil.ULIDNow()
}

type Suite struct {
	suite
}

type StepState struct {
	AuthToken     string
	Request       interface{}
	Response      interface{}
	ResponseErr   error
	RequestSentAt time.Time

	AssignedStudentIDs    []string
	UnAssignedStudentIDs  []string
	OtherStudentIDs       []string
	StudentIds            []string
	LearingObjectiveIDs   []string
	CurrentStudentID      string
	CurrentUserID         string
	CurrentOrderID        int32
	CurrentTeacherID      string
	CurrentClassID        int32
	CurrentClassCode      string
	CurrentSchoolID       int32
	CurrentLessonID       string
	StudentInCurrentClass []string
	ValidNotificationIds  []string
	CurrentChapterIDs     []string
	CurrentBookID         string

	ExistedLoID string

	// learning progress filter
	From      *types.Timestamp
	To        *types.Timestamp
	SessionID string

	// used for assign preset study plans test
	Random                  string
	PresetStudyPlanCSVFiles map[string]string
	PresetStudyPlans        map[*pb.PresetStudyPlan][]*entities.PresetStudyPlanWeekly

	// used for registration with school test
	Schools []*entities.School

	// payment_confirm recent order
	LastOrderIDs []int

	CreateOrderErrors []error

	// use for check_profile test
	expectingUserID    string
	CurrentPromotionID int32
	// check allocate student question
	CurrentQuestionID               string
	CurrentTopicID                  string
	CurrentTopicIDs                 []string
	CurrentPresetStudyPlanWeeklyIDs []string

	// custom assignment
	CurrentCustomAssignmentTopicID string
	CurrentCustomAssignmentID      string
	CurrentSubmissionIDs           []string
	SubmissionScores               []*pb.SubmissionScore
	CurrentLOIds                   []string
	AllStudentSubmissions          []string
	CurrentAssignmentIDs           []string

	Topics             []*epb.Topic
	LearningObjectives []*pb.LearningObjective
	QuizSets           []*pb.QuizSets

	LearningObjectivesv1 []*cpb.LearningObjective

	CurrentUserGroup string
	// create quiz test
	Quizzes         entities.Quizzes
	QuizSet         entities.QuizSet
	AllQuizzesRes   []*cpb.Quiz
	LoID            string
	Offset          int
	Limit           int
	NextPage        *cpb.Paging
	SetID           string
	StudyPlanItemID string
	QuizOptions     map[string]map[string][]*cpb.QuizOption

	// create flashcard study
	FlashcardStudies entities.FlashcardProgression
	StudySetID       string

	// check quiz correctness
	ShuffledQuizSetID string
	QuizItems         []*cpb.Quiz
	SelectedQuiz      []int
	SelectedIndex     map[string]map[string][]*bpb.Answer
	FilledText        map[string]map[string][]*bpb.Answer
	QuizAnswers       []*bpb.QuizAnswer

	// retrieve student submission history
	SetIDs                 []string
	AnswerLogs             []*cpb.AnswerLog
	NumShuffledQuizSetLogs int
	PaginatedBooks         [][]*cpb.Book
	PaginatedChapters      [][]*cpb.Chapter
	PaginatedCourses       [][]*cpb.Course

	// retrieve total quiz of los
	LOIDs      []string
	LOIDsInReq []string

	// list lesson medias
	MediaIDs             []string
	CurrentCourseID      string
	CurrentLessonGroupID string
	MediaItems           []*bpb.Media
	// check first quiz correctness
	FirstQuizCorrectness float32
	HighestQuizScore     float32
	// search basic profile
	NumberOfId        int
	SearchText        string
	ExpectedStudentID string
	studentID         string
	// prepare publish
	lessonID          string
	numberOfStream    int
	firstLearner      string
	secondLearner     string
	firstResponse     interface{}
	secondResponse    interface{}
	firstResponseErr  error
	secondResponseErr error

	StudentEventLogs []*entities.StudentEventLog

	numberOfLearners int

	totalNumberOfStudents           int
	numberOfStudentsArePublishing   int
	numberOfStudentsWantToUnpublish int
	numberOfStudentsWantToPublish   int

	courseIds            []string
	numberCourseHaveIcon int

	studentRemovedIds []string

	StudentDoingQuizExamLogs map[string][]*cpb.AnswerLog

	ClassIDs     []int32
	TeacherIDs   []string
	ClassMembers []*bpb.RetrieveClassMembersResponse_Member

	CommentIDs []string

	UserFillInTheBlankOld bool

	Notification             *notiEntities.InfoNotification
	UserNotification         *notiEntities.UserInfoNotification
	TimeRandom               time.Time
	NotificationList         []*notiEntities.InfoNotification
	NotificationMsgMap       map[string]*notiEntities.InfoNotificationMsg
	NotificationInfoListResp []*bpb.RetrieveNotificationsResponse_NotificationInfo

	ReadNotiCount int

	// For update_live_lesson.feature
	RemovedMediaIDs  []string
	RemovedCourseIDs []string
	CourseByID       map[pgtype.Text]*entities.Course

	// For retrieve lesson with filter & search
	FilterCourseIDs []string
	FilterFromTime  time.Time
	FilterToTime    time.Time
	SchoolIDs       []int32
	GradeIDs        []string
	// For retrieve lesson management with filter & search
	FilterTeacherIDs []string
	FilterStudentIDs []string
	FilterFromDate   time.Time
	FilterToDate     time.Time
	RandSchoolID     int

	MapPackageNameAndPackageID   map[string]int32
	MapPackageIDAndPackageItemID map[int32]string
	CurrentPackageID             int32

	StartTimeString string
	EndTimeString   string

	TopicID         string
	Courses         [][]interface{}
	CurrentParentID string
	ChapterID       string
	ChapterIDs      []string

	IsParentGroup bool
	UserID        string

	AudioOptions []*bpb.AudioOptionRequest
	NumText      int

	RetryShuffledQuizSetID   string
	LessonReportID           string
	DynamicFieldValuesByUser map[string][]*entities.PartnerDynamicFormFieldValue

	// modify live room state feature
	SubmitPollingAnswer []string
	BookID              string
	BookIDs             []string
	CourseID            string

	TopicIDs              []string
	FoundChanForJetStream chan interface{}
	FoundChanForLessonES  chan interface{}
	Subs                  []*natsJS.Subscription

	// Import grade with invalid rows
	InvalidCsvRows []string

	// Import grade with valid rows
	ValidCsvRows        []string
	CenterIDs           []string
	CreateLessonRequest *bpb.CreateLessonRequest

	// location
	LocationTypesID   []string
	LocationTypeOrgID string
	LocationIDs       []string

	// tenant
	TenantID     string
	UserPassword string

	CourseIDs        []string
	StudentPackageID string
	RoomID           string

	LocationID string

	StudentIDWithCourseID []string
	StartDate             time.Time
	EndDate               time.Time
}

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

var (
	conn, ysConn, eurekaConn, usermgmtConn, fatimaConn, shamirConn, virtualConn, lessonmgmtConn *grpc.ClientConn
	dbPostgres, db, eurekaDB, authDB                                                            *pgxpool.Pool
	applicantID                                                                                 string
	firebaseClient                                                                              *auth.Client // changing to firebaseAuthClient
	gcpApp                                                                                      *gcp.App
	firebaseAuthClient                                                                          internal_auth_tenant.TenantClient
	tenantManager                                                                               internal_auth_tenant.TenantManager
	keycloakClient                                                                              *internal_auth.IdentityServiceImpl
	jsm                                                                                         nats.JetStreamManagement
	unleashClient                                                                               unleashclient.ClientInstance
)

func updateResourcePath(db *pgxpool.Pool) error {
	ctx := context.Background()
	query := fmt.Sprintf(`UPDATE school_configs SET resource_path = %s;
		UPDATE schools SET resource_path = %s;
		UPDATE configs SET resource_path = %s;
		UPDATE cities SET resource_path = %s;
		UPDATE districts SET resource_path = %s;`,
		fmt.Sprint(constants.ManabieSchool),
		fmt.Sprint(constants.ManabieSchool),
		fmt.Sprint(constants.ManabieSchool),
		fmt.Sprint(constants.ManabieSchool),
		fmt.Sprint(constants.ManabieSchool))
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: fmt.Sprint(constants.ManabieSchool),
			DefaultRole:  entities.UserGroupAdmin,
			UserGroup:    entities.UserGroupAdmin,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
	_, err := db.Exec(ctx, query)
	return err
}

func setup(c *common.Config, fakeFirebaseAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	firebaseAddr = fakeFirebaseAddr

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	var err error

	rsc := bootstrap.NewResources().WithLoggerC(&c.Common)
	conn = rsc.GRPCDial("bob")
	ysConn = rsc.GRPCDial("yasuo")
	eurekaConn = rsc.GRPCDial("eureka")
	shamirConn = rsc.GRPCDial("shamir")
	virtualConn = rsc.GRPCDial("virtualclassroom")
	applicantID = c.JWTApplicant

	usermgmtConn = rsc.GRPCDial("usermgmt")
	fatimaConn = rsc.GRPCDial("fatima")
	lessonmgmtConn = rsc.GRPCDial("lessonmgmt")
	ctx = context.Background()

	db, _, _ = database.NewPool(ctx, zap.NewNop(), c.PostgresV2.Databases["bob"])
	// nolint
	bobPostgres := c.PostgresV2.Databases["bob"]
	bobPostgres.User = "postgres"
	bobPostgres.Password = c.PostgresMigrate.Database.Password
	dbPostgres, _, _ = database.NewPool(ctx, zap.NewNop(), bobPostgres)

	authPostgres := c.PostgresV2.Databases["auth"]
	authPostgres.User = "postgres"
	authPostgres.Password = c.PostgresMigrate.Database.Password
	authDB, _, _ = database.NewPool(ctx, zap.NewNop(), authPostgres)

	eurekaDB, _, _ = database.NewPool(ctx,
		zap.NewNop(),
		c.PostgresV2.Databases["eureka"],
	)

	firebaseApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	firebaseClient, err = firebaseApp.Auth(ctx)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create firebase client: %v", err))
	}

	gcpApp, err = gcp.NewApp(ctx, "", c.Common.IdentityPlatformProject)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	firebaseAuthClient, err = internal_auth_tenant.NewFirebaseAuthClientFromGCP(ctx, gcpApp)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	tenantManager, err = internal_auth_tenant.NewTenantManagerFromGCP(ctx, gcpApp)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create tenant manager: %v", err))
	}

	manabieTenantClient, err := tenantManager.TenantClient(ctx, internal_auth.LocalTenants[constants.ManabieSchool])
	if err != nil {
		zapLogger.Fatal("TenantClient")
	}
	mockScryptHashForManabie, err := mockScryptHash("mAaX5DSYQLUj3XD60McZ3n6m/AdZxEpfiLYqIFtYf2jlNIVaJ6Esu1sWe5HrsyLO1sTD/pygrtoFsQaFhfuRDg==", "Bw==", 8, 14)
	if err != nil {
		zapLogger.Fatal("mockScryptHash")
	}
	manabieTenantClient.SetHashConfig(mockScryptHashForManabie)

	srcIntegrationTestTenantClient, err := tenantManager.TenantClient(ctx, "integration-test-1-909wx")
	if err != nil {
		zapLogger.Fatal("TenantClient")
	}
	mockScryptHashForIntegrationTest, err := mockScryptHash("QRZsFCHcWnFf9+aMB0ajUo419AEtyrJf1YIv2S/kruwb8Zn3GJX3ZQj4bc5Mp8npCXVn8admB5iw5dg5lLtmNg==", "Bw==", 8, 14)
	if err != nil {
		zapLogger.Fatal("mockScryptHash")
	}
	srcIntegrationTestTenantClient.SetHashConfig(mockScryptHashForIntegrationTest)

	srcMigrationTenantClient, err := tenantManager.TenantClient(ctx, usermgmt.LocalTestMigrationTenant)
	if err != nil {
		zapLogger.Fatal("TenantClient")
	}
	mockScryptHashForSrcMigration, err := mockScryptHash("mAaX5DSYQLUj3XD60McZ3n6m/AdZxEpfiLYqIFtYf2jlNIVaJ6Esu1sWe5HrsyLO1sTD/pygrtoFsQaFhfuRDg==", "Bw==", 8, 14)
	if err != nil {
		zapLogger.Fatal("mockScryptHash")
	}
	srcMigrationTenantClient.SetHashConfig(mockScryptHashForSrcMigration)

	keycloakOpts := internal_auth.KeyCloakOpts{
		Path:     "https://d2020-ji-sso.jprep.jp",
		Realm:    "manabie-test",
		ClientID: "manabie-app",
	}

	keycloakClient, err = internal_auth.NewKeyCloakClient(keycloakOpts)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create keycloak client: %v", err))
	}

	jsm, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to create jetstream management: %v", err))
	}
	// unleash
	unleashClientIns, err := unleashclient.NewUnleashClientInstance(c.UnleashClientConfig.URL, c.UnleashClientConfig.AppName, c.UnleashClientConfig.APIToken, zapLogger)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to create unleash client: %v", err))
	}

	err = updateResourcePath(db)
	if err != nil {
		log.Fatal("failed to update database: %w", err)
	}
	jsm.ConnectToJS()
	unleashClientIns.ConnectToUnleashClient()
	defaultValues := (&usermgmt_repo.OrganizationRepo{}).DefaultOrganizationAuthValues(c.Common.Environment)

	// Init auth info
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
	_, err = db.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}

	rootAccount, err = test_usermgmt.InitRootAccount(ctx, shamirConn, firebaseAddr, c.JWTApplicant)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init rootAccount: %v", err))
	}
}

func TestSuiteInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c, f.FirebaseAddr)
		})
		ctx.AfterSuite(func() {
			db.Close()
			dbPostgres.Close()
			authDB.Close()
			eurekaDB.Close()
			conn.Close()
			ysConn.Close()
			eurekaConn.Close()
			usermgmtConn.Close()
			jsm.Close()
			virtualConn.Close()
			lessonmgmtConn.Close()
		})
	}
}

type stateKeyForBob struct{}

func StepStateFromContext(ctx context.Context) *StepState {
	state := ctx.Value(stateKeyForBob{})
	if state == nil {
		return &StepState{}
	}
	return state.(*StepState)
}

func StepStateToContext(ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, stateKeyForBob{}, state)
}

func InitBobState(ctx context.Context) context.Context {
	ctx = StepStateToContext(ctx, &StepState{
		PresetStudyPlanCSVFiles:      make(map[string]string),
		PresetStudyPlans:             make(map[*pb.PresetStudyPlan][]*entities.PresetStudyPlanWeekly),
		NotificationMsgMap:           make(map[string]*notiEntities.InfoNotificationMsg),
		MapPackageIDAndPackageItemID: make(map[int32]string),
		MapPackageNameAndPackageID:   make(map[string]int32),
	})
	return ctx
}

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c)
		initSteps(ctx, s)
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			ctx = InitBobState(ctx)
			//nolint:lostcancel
			ctx, _ = context.WithTimeout(ctx, time.Second*20)
			if golibs.InArrayString(sc.Name, scenarioOfRetrieveLesson) {
				for {
					if atomic.LoadInt32(&flagCheckRetrieveLesson) != 0 {
						time.Sleep(time.Second)
					} else {
						atomic.AddInt32(&flagCheckRetrieveLesson, 1)
						break
					}
				}
			} else if golibs.InArrayString(sc.Name, scenarioOfImportMasterData) {
				for {
					if atomic.LoadInt32(&flagCheckImportMasterData) != 0 {
						time.Sleep(time.Second)
					} else {
						atomic.AddInt32(&flagCheckImportMasterData, 1)
						break
					}
				}
			}
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
		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			if golibs.InArrayString(sc.Name, scenarioOfRetrieveLesson) {
				atomic.AddInt32(&flagCheckRetrieveLesson, -1)
			} else if golibs.InArrayString(sc.Name, scenarioOfImportMasterData) {
				atomic.AddInt32(&flagCheckImportMasterData, -1)
			}
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

func newSuite(c *common.Config) *suite {
	s := &suite{
		Cfg:                  c,
		Conn:                 conn,
		YsConn:               ysConn,
		EurekaConn:           eurekaConn,
		UsermgmtConn:         usermgmtConn,
		FatimaConn:           fatimaConn,
		VirtualClassroomConn: virtualConn,
		DB:                   db,
		DBPostgres:           dbPostgres,
		AuthDB:               authDB,
		EurekaDB:             eurekaDB,
		FirebaseClient:       firebaseClient,
		JSM:                  jsm,
		ZapLogger:            zapLogger,
		ApplicantID:          applicantID,
		ShamirConn:           shamirConn,
		GCPApp:               gcpApp,
		FirebaseAuthClient:   firebaseAuthClient,
		TenantManager:        tenantManager,
		KeycloakClient:       keycloakClient,
		UnleashClient:        unleashClient,
		// Unleash
		UnleashLocalAdminAPIKey: c.UnleashLocalAdminAPIKey,
		UnleashSrvAddr:          c.UnleashSrvAddr,

		RootAccount:     rootAccount,
		FirebaseAddress: firebaseAddr,

		CommonSuite:    &common.Suite{},
		LessonmgmtConn: lessonmgmtConn,
	}
	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = s.CommonSuite.StepState

	return s
}

func (s *suite) waitingFor(ctx context.Context, arg1 string) (context.Context, error) {
	d, err := time.ParseDuration(arg1)
	if err != nil {
		return ctx, err
	}
	time.Sleep(d)
	return ctx, nil
}

func (s *suite) inArray(array []string, search string) bool {
	for _, r := range array {
		if r == search {
			return true
		}
	}
	return false
}

func SetFirebaseAddr(fireBaseAddr string) {
	firebaseAddr = fireBaseAddr
}

func mockScryptHash(signerKey string, saltSeparator string, rounds int, memoryCost int) (*gcp.HashConfig, error) {
	decodedHashSignerKeyBytes, err := base64.StdEncoding.DecodeString(signerKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode hashSignerKey")
	}

	decodedHashSaltSeparator, err := base64.StdEncoding.DecodeString(saltSeparator)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode hashSignerKey")
	}

	hashConfig := &gcp.HashConfig{
		HashAlgorithm: "SCRYPT",
		HashSignerKey: gcp.Base64EncodedStr{
			Value:        signerKey,
			DecodedBytes: decodedHashSignerKeyBytes,
		},
		HashSaltSeparator: gcp.Base64EncodedStr{
			Value:        saltSeparator,
			DecodedBytes: decodedHashSaltSeparator,
		},
		HashRounds:     rounds,
		HashMemoryCost: memoryCost,
	}

	return hashConfig, nil
}
