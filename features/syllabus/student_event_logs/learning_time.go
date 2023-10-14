package student_event_logs

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GenericPayload struct {
	Event           string `json:"event,omitempty"`
	LoID            string `json:"lo_id,omitempty"`
	SessionID       string `json:"session_id,omitempty"`
	StudyPlanItemID string `json:"study_plan_item_id,omitempty"`
	TimeSpent       int    `json:"time_spent,omitempty"`
}

// nolint
func genStudentEvenLog(id, timeSpent int, studentID, loID, sessionID, evtType, event, studyPlanItemID, createdAt, studyPlanID string) (*entities.StudentEventLog, error) {
	layout := "2006-01-02 15:04:05.000000 +00:00"
	layout2 := "2006-01-02 15:04:05.000000"
	studentEvtLog := &entities.StudentEventLog{}
	database.AllNullEntity(studentEvtLog)
	studentEvtLog.EventID = database.Varchar(idutil.ULIDNow())
	payload := &GenericPayload{
		Event:           event,
		LoID:            loID,
		SessionID:       sessionID,
		StudyPlanItemID: studyPlanItemID,
		TimeSpent:       timeSpent,
	}
	studentEvtLog.Payload.Set(payload)

	err := multierr.Combine(studentEvtLog.CreatedAt.Set(timeutil.Now()),
		studentEvtLog.EventType.Set(evtType),
		studentEvtLog.StudentID.Set(studentID),
		studentEvtLog.ID.Set(id),
		studentEvtLog.CreatedAt.Set(timeutil.Now()),
		studentEvtLog.StudyPlanID.Set(studyPlanID),
		studentEvtLog.LearningMaterialID.Set(loID),
	)

	if err != nil {
		return nil, fmt.Errorf("unable to set value student event: %w", err)
	}
	if createdAt != "" {
		parsedTime, err := time.Parse(layout, createdAt)
		if err != nil {
			parsedTime, err = time.Parse(layout2, createdAt)
			if err != nil {
				return nil, fmt.Errorf("unable to parse time: %w", err)
			}
		}
		studentEvtLog.CreatedAt.Set(parsedTime)
	}
	return studentEvtLog, nil
}
func (s *Suite) genStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	_, authToken, err := s.AuthHelper.AUserSignedInAsRole(ctx, "admin")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	bookID, _, topicIDs, err := utils.AValidBookContent(s.AuthHelper.SignedCtx(ctx, authToken), s.EurekaConn, s.EurekaDB, constants.ManabieSchool)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidBookContent: %w", err)
	}

	numOfLos := int(utils.CryptRand(5) + 1)
	los := make([]*cpb.LearningObjective, 0, numOfLos)
	for i := 0; i < numOfLos; i++ {
		lo := utils.GenerateLearningObjective(topicIDs[0])
		stepState.LmIDs = append(stepState.LmIDs, lo.Info.Id)
		los = append(los, lo)
	}

	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.EurekaConn).UpsertLOs(s.AuthHelper.SignedCtx(ctx, authToken),
		&epb.UpsertLOsRequest{
			LearningObjectives: los,
		}); err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.PermissionDenied {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("NewLearningObjectiveModifierServiceClient: %w", err)
		}
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable create los: %w", err)
	}

	courseID, err := utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, authToken), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateCourse: %w", err)
	}
	if err := utils.GenerateCourseBooks(s.AuthHelper.SignedCtx(ctx, authToken), courseID, []string{bookID}, s.EurekaConn); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateCourseBooks: %w", err)
	}
	studyPlanResult, err := utils.GenerateStudyPlanV2(s.AuthHelper.SignedCtx(ctx, authToken), s.EurekaConn, courseID, bookID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateStudyPlanV2: %w", err)
	}
	stepState.StudyPlanID = studyPlanResult.StudyPlanID

	_, err = epb.NewAssignmentModifierServiceClient(s.EurekaConn).AssignStudyPlan(s.AuthHelper.SignedCtx(ctx, authToken), &epb.AssignStudyPlanRequest{
		StudyPlanId: stepState.StudyPlanID,
		Data: &epb.AssignStudyPlanRequest_StudentId{
			StudentId: stepState.UserID,
		},
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("epb.NewAssignmentModifierServiceClient(s.EurekaConn).AssignStudyPlan: %w", err)
	}
	studyPlanItems, err := (&repositories.StudyPlanItemRepo{}).FindByStudyPlanID(ctx, s.EurekaDB, database.Text(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve master study plan items: %w", err)
	}

	studyPlanItemIDs := make([]string, len(studyPlanItems))
	for i, studyPlanItem := range studyPlanItems {
		studyPlanItemIDs[i] = studyPlanItem.ID.String
	}
	stepState.StudyPlanItemIDs = studyPlanItemIDs
	return utils.StepStateToContext(ctx, stepState), nil
}
func (s *Suite) createMockIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	individualStudyPlanRepo := &repositories.IndividualStudyPlan{}
	individualStudyPlan := &entities.IndividualStudyPlan{}
	database.AllNullEntity(individualStudyPlan)
	studyPlanID := stepState.StudyPlanID
	learningMaterialID := stepState.LmIDs[0]
	err := multierr.Combine(
		individualStudyPlan.ID.Set(studyPlanID),
		individualStudyPlan.StudentID.Set(stepState.UserID),
		individualStudyPlan.LearningMaterialID.Set(learningMaterialID),
		individualStudyPlan.StartDate.Set(timeutil.Now()),
		individualStudyPlan.EndDate.Set(timeutil.Now().AddDate(0, 0, 1)),
		individualStudyPlan.AvailableFrom.Set(timeutil.Now()),
		individualStudyPlan.AvailableTo.Set(timeutil.Now().AddDate(0, 0, 1)),
		individualStudyPlan.CreatedAt.Set(timeutil.Now()),
		individualStudyPlan.UpdatedAt.Set(timeutil.Now()),
		individualStudyPlan.Status.Set("active"),
	)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value individual study plan: %w", err)
	}
	_, err = individualStudyPlanRepo.BulkSync(ctx, s.EurekaDB, []*entities.IndividualStudyPlan{individualStudyPlan})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create individual study plan: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func reproduce2Start1CompletedV1(ctx context.Context) (context.Context, error) {
	studentEvtLogs := make([]*entities.StudentEventLog, 0)
	stepState := utils.StepStateFromContext[StepState](ctx)

	studyPlanID := stepState.StudyPlanID
	lmID := stepState.LmIDs[0]
	sessionID := idutil.ULIDNow()
	studyPlanItemID := stepState.StudyPlanItemIDs[0]
	e, err := genStudentEvenLog(153428, 1, stepState.UserID, lmID, sessionID, "learning_objective", "started", studyPlanItemID, "2021-12-20 06:00:00.000000 +00:00", studyPlanID)
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153429, 10, stepState.UserID, lmID, sessionID, "learning_objective", "paused", studyPlanItemID, "2021-12-20 06:10:00.000000 +00:00", studyPlanID)
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	e, err = genStudentEvenLog(153430, 32, stepState.UserID, lmID, sessionID, "learning_objective", "resumed", studyPlanItemID, "2021-12-20 06:20:00.000000 +00:00", studyPlanID)
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, stepState.UserID, lmID, sessionID, "learning_objective", "completed", studyPlanItemID, "2021-12-20 06:30:00.000000 +00:00", studyPlanID)
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, stepState.UserID, lmID, sessionID, "learning_objective", "started", studyPlanItemID, "2021-12-20 06:40:00.000000 +00:00", studyPlanID)
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)

	e, err = genStudentEvenLog(153443, 0, stepState.UserID, lmID, sessionID, "learning_objective", "paused", studyPlanItemID, "2021-12-20 06:55:00.000000 +00:00", studyPlanID)
	if err != nil {
		return nil, err
	}
	studentEvtLogs = append(studentEvtLogs, e)
	stepState.EStudentEventLogs_1 = studentEvtLogs
	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint: gosec
func (s *Suite) studentCreateEventLogv2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx = s.AuthHelper.SignedCtx((ctx), stepState.Token)
	ctx, err := s.genStudyPlan(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.createMockIndividualStudyPlan(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	ctx, err = reproduce2Start1CompletedV1(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	repo := &repositories.StudentEventLogRepo{}
	if err := repo.Create(ctx, s.EurekaDB, stepState.EStudentEventLogs_1); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) calculateLearningTime(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	expected := 20

	query1 := `SELECT learning_time_by_minutes FROM calculate_learning_time($1)`
	var count1 int
	if err := s.EurekaDB.QueryRow(ctx, query1, stepState.UserID).Scan(&count1); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	if count1 != expected {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected learning_time sql1 %d, got %d\nid: %s", expected, count1, stepState.UserID)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) maxScoreMustBeStored(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	query := `SELECT count(*) FROM max_score_submission where study_plan_id = $1 AND learning_material_id = $2 AND student_id = $3`
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, stepState.StudyPlanID, stepState.LmIDs[0], stepState.UserID).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("max score submission is not store: %w", err)
	}
	if count != 1 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected max score submission is store")
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentsEventLogIsStored(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	query := `SELECT count(*) FROM student_event_logs where study_plan_id = $1 AND learning_material_id = $2 AND student_id = $3`
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, stepState.StudyPlanID, stepState.LmIDs[0], stepState.UserID).Scan(&count); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot scan student event logs: %w", err)
	}
	if count == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected student event logs is store")
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
