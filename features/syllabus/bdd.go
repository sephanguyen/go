package syllabus

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/features/syllabus/allocate_marker"
	"github.com/manabie-com/backend/features/syllabus/assignment"
	"github.com/manabie-com/backend/features/syllabus/course_statistical"
	"github.com/manabie-com/backend/features/syllabus/entity"
	"github.com/manabie-com/backend/features/syllabus/exam_lo"
	"github.com/manabie-com/backend/features/syllabus/flashcard"
	"github.com/manabie-com/backend/features/syllabus/individual_study_plan"
	"github.com/manabie-com/backend/features/syllabus/learning_history_data_sync"
	learning_material "github.com/manabie-com/backend/features/syllabus/learning_material"
	learning_objective "github.com/manabie-com/backend/features/syllabus/learning_objective"
	"github.com/manabie-com/backend/features/syllabus/nat_sync"
	"github.com/manabie-com/backend/features/syllabus/question_tag"
	"github.com/manabie-com/backend/features/syllabus/question_tag_type"
	"github.com/manabie-com/backend/features/syllabus/quiz"
	"github.com/manabie-com/backend/features/syllabus/shuffled_quiz_set"
	student_event_logs "github.com/manabie-com/backend/features/syllabus/student_event_logs"
	"github.com/manabie-com/backend/features/syllabus/student_progress"
	"github.com/manabie-com/backend/features/syllabus/student_submission"
	"github.com/manabie-com/backend/features/syllabus/study_plan"
	"github.com/manabie-com/backend/features/syllabus/study_plan_item"
	"github.com/manabie-com/backend/features/syllabus/task_assignment"
	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/constants"
	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/yasuo/constant"

	firebase "firebase.google.com/go"
	"github.com/cucumber/godog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	connections  *common.Connections
	zapLogger    *zap.Logger
	firebaseAddr string
	applicantID  string

	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
	authHelper         *utils.AuthHelper
)

const ManabieSchool = constant.ManabieSchool

func init() {
	rand.Seed(time.Now().UnixNano())
	common.RegisterTest("syllabus", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
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

func InitSyllabusState(ctx context.Context, s *Suite) context.Context {
	ctx = utils.StepStateToContext(ctx, s.StepState)
	return ctx
}

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c)
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			ctx = InitSyllabusState(ctx, s)
			uriSplit := strings.Split(sc.Uri, ":")
			entityName := strings.Split(uriSplit[0], "/")[1]
			switch entityName {
			case "learning_objective":
				ctx = utils.StepStateToContext(ctx, s.LOStepState)
			case "flashcard":
				ctx = utils.StepStateToContext(ctx, s.FlashcardStepState)
			case "assignment":
				ctx = utils.StepStateToContext(ctx, s.AssignmentStepState)
			case "learning_material":
				ctx = utils.StepStateToContext(ctx, s.LearningMaterialStepState)
			case "exam_lo":
				ctx = utils.StepStateToContext(ctx, s.ExamLOStepState)
			case "individual_study_plan":
				ctx = utils.StepStateToContext(ctx, s.IndividualStudyPlanStepState)
			case "task_assignment":
				ctx = utils.StepStateToContext(ctx, s.TaskAssignmentStepState)
			case "course_statistical":
				ctx = utils.StepStateToContext(ctx, s.TopicStatisticalState)
			case "student_event_logs":
				ctx = utils.StepStateToContext(ctx, s.StudentEventLogsStepState)
			case "study_plan":
				ctx = utils.StepStateToContext(ctx, s.StudyPlanStepState)
			case "study_plan_item":
				ctx = utils.StepStateToContext(ctx, s.StudyPlanItemStepState)
			case "shuffled_quiz_set":
				ctx = utils.StepStateToContext(ctx, s.ShuffledQuizSetStepState)
			case "student_progress":
				ctx = utils.StepStateToContext(ctx, s.StudentProgressStepState)
			case "student_submission":
				ctx = utils.StepStateToContext(ctx, s.StudentSubmissionStepState)
			case "nat_sync":
				ctx = utils.StepStateToContext(ctx, s.JprepSync)
			case "question_tag":
				ctx = utils.StepStateToContext(ctx, s.QuestionTagStepState)
			case "question_tag_type":
				ctx = utils.StepStateToContext(ctx, s.QuestionTagTypeStepState)
			case "allocate_marker":
				ctx = utils.StepStateToContext(ctx, s.AllocateMarkerStepState)
			case "quiz":
				ctx = utils.StepStateToContext(ctx, s.QuizStepState)
			case "learning_history_data_sync":
				ctx = utils.StepStateToContext(ctx, s.LearningHistoryDataSyncStepState)
			default:
				ctx = utils.StepStateToContext(ctx, s.StepState)
			}

			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: strconv.Itoa(ManabieSchool),
					DefaultRole:  constants.RoleSchoolAdmin,
					UserGroup:    entities.UserGroupSchoolAdmin,
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

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	err := connections.ConnectGRPC(ctx,
		common.WithCredentials(grpc.WithTransportCredentials(insecure.NewCredentials())),
		common.WithBobSvcAddress(),
		common.WithTomSvcAddress(),
		common.WithEurekaSvcAddress(),
		common.WithFatimaSvcAddress(),
		common.WithShamirSvcAddress(),
		common.WithYasuoSvcAddress(),
		common.WithUserMgmtSvcAddress(),
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
	connections.TenantManager, err = internal_auth_tenant.NewTenantManagerFromGCP(ctx, connections.GCPApp)
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

	err = common.UpdateResourcePath(connections.BobDB)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to update resource_path: %v", err))
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
	authHelper = utils.NewAuthHelper(connections.BobDBTrace, connections.EurekaDBTrace, connections.FatimaDBTrace, applicantID, fakeFirebaseAddr, connections.ShamirConn)
}

func initSteps(ctx *godog.ScenarioContext, s *Suite) {
	learningMaterialSteps := initLearningMaterialStep(s)
	studyPlanSteps := initStudyPlanStep(s)
	steps := map[string]interface{}{}
	utils.AppendSteps(steps, learningMaterialSteps)
	utils.AppendSteps(steps, studyPlanSteps)
	buildRegexpMapOnce.Do(func() { regexpMap = helper.BuildRegexpMapV2(steps) })
	for pattern, stepFunc := range steps {
		ctx.Step(regexpMap[pattern], stepFunc)
	}
}

type Suite struct {
	*common.Connections
	*StepState
	ZapLogger  *zap.Logger
	Cfg        *common.Config
	AuthHelper *utils.AuthHelper
}

type StepState struct {
	LOStepState                      *learning_objective.StepState
	FlashcardStepState               *flashcard.StepState
	AssignmentStepState              *assignment.StepState
	LearningMaterialStepState        *learning_material.StepState
	ExamLOStepState                  *exam_lo.StepState
	IndividualStudyPlanStepState     *individual_study_plan.StepState
	TaskAssignmentStepState          *task_assignment.StepState
	TopicStatisticalState            *course_statistical.StepState
	StudentEventLogsStepState        *student_event_logs.StepState
	StudyPlanStepState               *study_plan.StepState
	StudyPlanItemStepState           *study_plan_item.StepState
	ShuffledQuizSetStepState         *shuffled_quiz_set.StepState
	StudentProgressStepState         *student_progress.StepState
	StudentSubmissionStepState       *student_submission.StepState
	JprepSync                        *nat_sync.StepState
	QuestionTagStepState             *question_tag.StepState
	QuestionTagTypeStepState         *question_tag_type.StepState
	AllocateMarkerStepState          *allocate_marker.StepState
	QuizStepState                    *quiz.StepState
	LearningHistoryDataSyncStepState *learning_history_data_sync.StepState
	Response                         interface{}
	Request                          interface{}
	ResponseErr                      error
	BookID                           string
	TopicIDs                         []string
	ChapterIDs                       []string
	Token                            string
	SchoolAdmin                      entity.SchoolAdmin
	Student                          entity.Student
	OfflineLearningID                string
}

func newSuite(c *common.Config) *Suite {
	s := &Suite{
		Connections: connections,
		Cfg:         c,
		ZapLogger:   zapLogger,
		AuthHelper:  authHelper,
	}

	s.StepState = &StepState{
		LOStepState:                      &learning_objective.StepState{},
		FlashcardStepState:               &flashcard.StepState{},
		AssignmentStepState:              &assignment.StepState{},
		LearningMaterialStepState:        &learning_material.StepState{},
		ExamLOStepState:                  &exam_lo.StepState{},
		IndividualStudyPlanStepState:     &individual_study_plan.StepState{},
		TaskAssignmentStepState:          &task_assignment.StepState{},
		TopicStatisticalState:            &course_statistical.StepState{},
		StudentEventLogsStepState:        &student_event_logs.StepState{},
		StudyPlanStepState:               &study_plan.StepState{},
		StudyPlanItemStepState:           &study_plan_item.StepState{},
		ShuffledQuizSetStepState:         &shuffled_quiz_set.StepState{},
		StudentProgressStepState:         &student_progress.StepState{},
		StudentSubmissionStepState:       &student_submission.StepState{},
		JprepSync:                        &nat_sync.StepState{},
		QuestionTagStepState:             &question_tag.StepState{},
		QuestionTagTypeStepState:         &question_tag_type.StepState{},
		AllocateMarkerStepState:          &allocate_marker.StepState{},
		QuizStepState:                    &quiz.StepState{},
		LearningHistoryDataSyncStepState: &learning_history_data_sync.StepState{},
	}
	return s
}
