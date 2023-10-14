package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) RetrieveStudyPlanItemEventLogs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, ctxCancel := context.WithTimeout(ctx, 5*time.Second)
	defer ctxCancel()

	mStudyPlanItemID := make(map[string]bool)
	for _, studentEventLog := range stepState.StudentEventLogs {
		payload := make(map[string]interface{})
		if err := studentEventLog.Payload.AssignTo(&payload); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		studyPlanItemID, ok := payload["study_plan_item_id"].(string)
		if !ok {
			continue
		}
		mStudyPlanItemID[studyPlanItemID] = true
	}

	studyPlanItemIDs := make([]string, 0, len(mStudyPlanItemID))
	for studyPlanItemID := range mStudyPlanItemID {
		studyPlanItemIDs = append(studyPlanItemIDs, studyPlanItemID)
	}
	stepState.Response, stepState.ResponseErr = epb.NewStudyPlanReaderServiceClient(s.Conn).
		RetrieveStudyPlanItemEventLogs(s.signedCtx(ctx), &epb.RetrieveStudyPlanItemEventLogsRequest{
			StudyPlanItemId: studyPlanItemIDs,
		})
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), nil
	}

	resp := stepState.Response.(*epb.RetrieveStudyPlanItemEventLogsResponse)

	for _, item := range resp.Items {
		for _, log := range item.Logs {
			if log.CreatedAt.AsTime().IsZero() {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected created_at mustn't be zero")
			}
			if log.LearningTime == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected learning_time mustn't be zero")
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someLearning_objectiveStudentEventLogsAreExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentID := stepState.AssignedStudentIDs[0]

	var logs []*entities.StudentEventLog

	for i := 0; i < 5; i++ {
		sessionID := strconv.Itoa(rand.Int())
		studyPlanItemID := s.newID()

		logStarted := new(entities.StudentEventLog)
		database.AllNullEntity(logStarted)
		logStarted.ID.Set(rand.Int31())
		logStarted.StudentID.Set(studentID)
		logStarted.EventID.Set(strconv.Itoa(rand.Int()))
		logStarted.EventType.Set("learning_objective")
		logStarted.Payload.Set(map[string]interface{}{
			"session_id":         sessionID,
			"event":              "started",
			"study_plan_item_id": studyPlanItemID,
		})
		logStarted.CreatedAt.Set(time.Now().UTC().Add(-26 * time.Hour))

		logMissingEvent := new(entities.StudentEventLog)
		database.AllNullEntity(logMissingEvent)
		logMissingEvent.ID.Set(rand.Int31())
		logMissingEvent.StudentID.Set(studentID)
		logMissingEvent.EventID.Set(strconv.Itoa(rand.Int()))
		logMissingEvent.EventType.Set("learning_objective")
		logMissingEvent.Payload.Set(map[string]interface{}{
			"session_id":         sessionID,
			"study_plan_item_id": studyPlanItemID,
		})
		logMissingEvent.CreatedAt.Set(time.Now().UTC().Add(-26 * time.Hour))

		logEmptyPayload := new(entities.StudentEventLog)
		database.AllNullEntity(logEmptyPayload)
		logEmptyPayload.ID.Set(rand.Int31())
		logEmptyPayload.StudentID.Set(studentID)
		logEmptyPayload.EventID.Set(strconv.Itoa(rand.Int()))
		logEmptyPayload.EventType.Set("learning_objective")
		logEmptyPayload.Payload.Set(map[string]interface{}{})
		logEmptyPayload.CreatedAt.Set(time.Now().UTC().Add(-26 * time.Hour))

		logCompleted := new(entities.StudentEventLog)
		database.AllNullEntity(logCompleted)
		logCompleted.ID.Set(rand.Int31())
		logCompleted.StudentID.Set(studentID)
		logCompleted.EventID.Set(strconv.Itoa(rand.Int()))
		logCompleted.EventType.Set("learning_objective")
		logCompleted.Payload.Set(map[string]interface{}{
			"session_id":         sessionID,
			"event":              "completed",
			"study_plan_item_id": studyPlanItemID,
		})
		logCompleted.CreatedAt.Set(time.Now().UTC().Add(-25 * time.Hour))

		logs = append(logs, logStarted, logMissingEvent, logEmptyPayload, logCompleted)
	}
	for _, log := range logs {
		if _, err := database.Insert(ctx, log, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	stepState.StudentEventLogs = logs
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anAssignedStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()

	s.aValidStudentInDB(ctx, id)

	var studentID pgtype.Text
	studentID.Set(id)

	stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, id)
	return StepStateToContext(ctx, stepState), nil
}
