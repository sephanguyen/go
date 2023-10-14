package syllabus

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/gandalf"
	"github.com/manabie-com/backend/features/repository/syllabus/assignment"
	"github.com/manabie-com/backend/features/repository/syllabus/book"
	"github.com/manabie-com/backend/features/repository/syllabus/course_student"
	csp "github.com/manabie-com/backend/features/repository/syllabus/course_study_plan"
	"github.com/manabie-com/backend/features/repository/syllabus/exam_lo_submission"
	"github.com/manabie-com/backend/features/repository/syllabus/learning_objectives"
	"github.com/manabie-com/backend/features/repository/syllabus/shuffled_quiz_set"
	"github.com/manabie-com/backend/features/repository/syllabus/student_event_log"
	"github.com/manabie-com/backend/features/repository/syllabus/student_latest_submissions"
	studentstudyplan "github.com/manabie-com/backend/features/repository/syllabus/student_study_plan"
	"github.com/manabie-com/backend/features/repository/syllabus/student_submissions"
	"github.com/manabie-com/backend/features/repository/syllabus/study_plan"
	"github.com/manabie-com/backend/features/repository/syllabus/user"
	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

//nolint:revive
const (
	EnableDefaultRLS = false
	HasuraPassword   = "M@nabie123"
	DefaultSchoolID  = 1
)

var (
	eurekaDB       *pgxpool.Pool
	bobDB          *pgxpool.Pool
	zapLogger      *zap.Logger
	hasuraAdminURL string
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TestSuiteInitializer ...
func TestSuiteInitializer(c *gandalf.Config) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c)
		})
		ctx.AfterSuite(func() {
			eurekaDB.Close()
		})
	}
}

func setup(c *gandalf.Config) {
	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)
	eurekaDB, _, _ = database.NewPool(context.Background(), zapLogger, c.PostgresV2.Databases["eureka"])

	bobDB, _, _ = database.NewPool(context.Background(), zapLogger, c.PostgresV2.Databases["bob"])

	hasuraAdminURL = c.EurekaHasuraAdminURL
}

type Suite struct {
	*StepState
	DB         database.Ext
	BobDBTrace *database.DBTrace
	ZapLogger  *zap.Logger

	HasuraAdminURL string
	HasuraPassword string
}

type StepState struct {
	DefaultCourseID          string
	DefaultStudyPlanID       string
	CourseStudyPlanStepState *csp.StepState

	StudentStudyPlanStepState *studentstudyplan.StepState

	DefaultSchoolID        int32
	BookStepState          *book.StepState
	CourseStudentStepState *course_student.StepState
	StudyPlanStepState     *study_plan.StepState
	AssignmentStepState    *assignment.StepState
	UserStepState          *user.StepState

	ShuffledQuizSetStepState         *shuffled_quiz_set.StepState
	StudentSubmissionStepState       *student_submissions.StepState
	StudentLatestSubmissionStepState *student_latest_submissions.StepState
	StudentEventLogStepState         *student_event_log.StepState

	ExamLOSubmissionStepState *exam_lo_submission.StepState

	LearningObjectivesStepState *learning_objectives.StepState
}

func newSuite() *Suite {
	return &Suite{
		StepState: &StepState{
			BookStepState: &book.StepState{
				DefaultSchoolID: DefaultSchoolID,
			},
			CourseStudyPlanStepState: &csp.StepState{
				DefaultSchoolID: DefaultSchoolID,
			},

			CourseStudentStepState: &course_student.StepState{},
			StudyPlanStepState:     &study_plan.StepState{},

			StudentStudyPlanStepState: &studentstudyplan.StepState{
				DefaultSchoolID: DefaultSchoolID,
			},

			StudentSubmissionStepState:       &student_submissions.StepState{},
			StudentLatestSubmissionStepState: &student_latest_submissions.StepState{},

			AssignmentStepState: &assignment.StepState{
				DefaultSchoolID: DefaultSchoolID,
			},
			ShuffledQuizSetStepState:    &shuffled_quiz_set.StepState{},
			UserStepState:               &user.StepState{},
			StudentEventLogStepState:    &student_event_log.StepState{},
			ExamLOSubmissionStepState:   &exam_lo_submission.StepState{},
			LearningObjectivesStepState: &learning_objectives.StepState{},
		},
		DB:             eurekaDB,
		ZapLogger:      zapLogger,
		HasuraAdminURL: hasuraAdminURL,
		HasuraPassword: HasuraPassword,
		BobDBTrace:     &database.DBTrace{DB: bobDB},
	}
}

func initSteps(ctx *godog.ScenarioContext, s *Suite) {
	learningMaterialSteps := initLearningMaterialStep(s)
	studyPlanSteps := initStudyPlanStep(s)
	steps := map[string]interface{}{}
	utils.AppendSteps(steps, learningMaterialSteps)
	utils.AppendSteps(steps, studyPlanSteps)

	for pattern, stepFunc := range steps {
		ctx.Step(pattern, stepFunc)
	}
}

func ScenarioInitializer(c *gandalf.Config) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite()
		initSteps(ctx, s)
		s.DefaultSchoolID = DefaultSchoolID
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			uriSplit := strings.Split(sc.Uri, ":")
			entityName := strings.Split(uriSplit[0], "/")[2]
			switch entityName {
			case "course_student":
				ctx = utils.StepStateToContext(ctx, s.CourseStudentStepState)
			case "book":
				ctx = utils.StepStateToContext(ctx, s.BookStepState)

			case "study_plan":
				ctx = utils.StepStateToContext(ctx, s.StudyPlanStepState)

			case "course_study_plan":
				ctx = utils.StepStateToContext(ctx, s.CourseStudyPlanStepState)
			case "student_study_plan":
				ctx = utils.StepStateToContext(ctx, s.StudentStudyPlanStepState)
			case "assignment":
				ctx = utils.StepStateToContext(ctx, s.AssignmentStepState)
			case "user":
				ctx = utils.StepStateToContext(ctx, s.UserStepState)
			case "shuffled_quiz_set":
				ctx = utils.StepStateToContext(ctx, s.ShuffledQuizSetStepState)
			case "student_submissions":
				ctx = utils.StepStateToContext(ctx, s.StudentSubmissionStepState)
			case "student_latest_submissions":
				ctx = utils.StepStateToContext(ctx, s.StudentLatestSubmissionStepState)
			case "student_event_log":
				ctx = utils.StepStateToContext(ctx, s.StudentEventLogStepState)
			case "exam_lo_submission":
				ctx = utils.StepStateToContext(ctx, s.ExamLOSubmissionStepState)
			case "learning_objectives":
				ctx = utils.StepStateToContext(ctx, s.LearningObjectivesStepState)
			default:
				ctx = utils.StepStateToContext(ctx, s.StepState)
			}
			if EnableDefaultRLS {
				ctx = addResourcePathToCtx(ctx, s.DefaultSchoolID)
			}
			return ctx, nil
		})
	}
}

func addResourcePathToCtx(ctx context.Context, schoolID int32) context.Context {
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: strconv.Itoa(int(schoolID)),
			DefaultRole:  entities.UserGroupSchoolAdmin,
			UserGroup:    entities.UserGroupSchoolAdmin,
		},
	}

	return interceptors.ContextWithJWTClaims(ctx, &claim)
}
