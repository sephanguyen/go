package eureka

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

const (
	archived = "archived"
	active   = "active"
)

func (s *suite) returnsTodoItemsTotalCorrectlyWithStatus(ctx context.Context, statusArg string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.ListStudentToDoItemsResponse)

	if stepState.ListStudentTodoItemStatus == archived {
		if len(resp.Items) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("total of to do items is wrong, expect 0 but got %v", len(resp.Items))
		}
	}
	if stepState.ListStudentTodoItemStatus == active {
		switch statusArg {
		case "active":
			if len(resp.Items) != len(stepState.StudyPlanItemIDsActive) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("total of to do items is wrong, expect %v but got %v", len(stepState.StudyPlanItemIDsActive), len(resp.Items))
			}
		case "overdue":
			if len(resp.Items) != len(stepState.StudyPlanItemIDsOverDue)-len(stepState.StudyPlanItemIDsOverDueDeleted) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("total of to do items is wrong, expect %v but got %v", len(stepState.StudyPlanItemIDsOverDue)-len(stepState.StudyPlanItemIDsOverDueDeleted), len(resp.Items))
			}
		case "completed":
			if len(resp.Items) != len(stepState.StudyPlanItemIDsCompleted) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("total of to do items is wrong, expect %v but got %v", len(stepState.StudyPlanItemIDsCompleted), len(resp.Items))
			}
		}
	}

	if ctx, err := s.todoItemsMustnotReturnDeletedTopicChaptes(ctx, resp); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) todoItemsMustnotReturnDeletedTopicChaptes(ctx context.Context, req *epb.ListStudentToDoItemsResponse) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	const query = `
SELECT count(spi.study_plan_item_id) FROM study_plan_items spi 
JOIN books b ON b.book_id = spi.content_structure ->> 'book_id'
JOIN chapters c ON c.chapter_id = spi.content_structure ->> 'chapter_id'
JOIN topics t ON t.topic_id = spi.content_structure ->> 'topic_id'
WHERE
  spi.study_plan_item_id = ANY($1::TEXT[])
  AND spi.deleted_at IS NULL
  AND b.deleted_at IS NULL
  AND c.deleted_at IS NULL
  AND t.deleted_at IS NULL
`
	studyPlanItemIDs := make([]string, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		studyPlanItemID := item.GetStudyPlanItem().GetStudyPlanItemId()
		studyPlanItemIDs = append(studyPlanItemIDs, studyPlanItemID)
	}

	var count int
	err := s.DB.QueryRow(ctx, query, database.TextArray(studyPlanItemIDs)).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error query data: %w", err)
	}

	if len(studyPlanItemIDs) != count {
		return StepStateToContext(ctx, stepState), fmt.Errorf("return item of deleted chapters/topics expected %v, got = %v", len(studyPlanItemIDs), count)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateSomeValidStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	n := rand.Intn(5) + 5
	studyPlanIDs := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		if ctx, err := s.userCreateAValidStudyPlan(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studyPlanIDs = append(studyPlanIDs, stepState.StudyPlanID)
	}
	stepState.StudyPlanIDs = studyPlanIDs

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRetrieveListStudentTodoItemsWithStatus(ctx context.Context, statusArg string) (context.Context, error) {
	var status epb.ToDoStatus
	switch statusArg {
	case "active":
		status = epb.ToDoStatus_TO_DO_STATUS_ACTIVE
	case "completed":
		status = epb.ToDoStatus_TO_DO_STATUS_COMPLETED
	case "overdue":
		status = epb.ToDoStatus_TO_DO_STATUS_OVERDUE
	case "upcoming":
		status = epb.ToDoStatus_TO_DO_STATUS_UPCOMING
	default:
		status = epb.ToDoStatus_TO_DO_STATUS_NONE
	}
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = epb.NewAssignmentReaderServiceClient(s.Conn).ListStudentToDoItems(s.signedCtx(ctx), &epb.ListStudentToDoItemsRequest{
		Paging: &cpb.Paging{
			Limit: 100,
		},
		StudentId: stepState.StudentID,
		Status:    status,
		CourseIds: []string{stepState.CourseID},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateStudyPlansStatusTo(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.StudyPlan{}
	var stt string
	switch status {
	case archived:
		stt = epb.StudyPlanStatus_STUDY_PLAN_STATUS_ARCHIVED.String()
	case active:
		stt = epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE.String()
	}
	sspe := &entities.StudentStudyPlan{}
	query := fmt.Sprintf(`
		WITH tmp AS (
			SELECT study_plan_id FROM %s
			WHERE student_id = $1
		)
		UPDATE %s
		SET status = '%s'
		WHERE study_plan_id IN(SELECT * FROM tmp)
		AND master_study_plan_id = ANY($2)
	`, sspe.TableName(), e.TableName(), stt)

	cmd, err := s.DB.Exec(ctx, query, &stepState.UserId, &stepState.StudyPlanIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update study plan status: %w", err)
	}
	if cmd.RowsAffected() != int64(len(stepState.StudyPlanIDs)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("update status is wrong, expect %v but got %v rows", len(stepState.StudyPlanIDs), cmd.RowsAffected())
	}

	stepState.ListStudentTodoItemStatus = status
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateStudyPlanItemsStatusTo(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.StudyPlanItem{}
	var stt string
	switch status {
	case archived:
		stt = epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED.String()
	case active:
		stt = epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String()
	}
	sspe := &entities.StudentStudyPlan{}
	spe := &entities.StudyPlan{}
	query := fmt.Sprintf(`
		WITH tmp AS (
			SELECT study_plan_id FROM %s
			JOIN %s as sp
			USING (study_plan_id)
			WHERE student_id = $1
			AND sp.master_study_plan_id = ANY($2)
		)
		UPDATE %s
		SET status = '%s'
		WHERE study_plan_id IN(SELECT * FROM tmp)
	`, sspe.TableName(), spe.TableName(), e.TableName(), stt)

	cmd, err := s.DB.Exec(ctx, query, &stepState.UserId, &stepState.StudyPlanIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update study plan item status: %w", err)
	}
	if cmd.RowsAffected() != int64(len(stepState.StudyPlanItemInfos)*len(stepState.StudyPlanIDs)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("update status is wrong, expect %v but got %v rows", len(stepState.StudyPlanItemInfos)*len(stepState.StudyPlanIDs), cmd.RowsAffected())
	}

	stepState.ListStudentTodoItemStatus = status
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidCourseStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseID := idutil.ULIDNow()
	stepState.CourseID = courseID
	if ctx, err := s.insertUserIntoBob(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentID = stepState.UserId
	courseStudent, err := generateCourseByStudentId(stepState.UserId, courseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	cmd, err := database.Insert(ctx, courseStudent, s.DB.Exec)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if cmd.RowsAffected() != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error insert course student")
	}
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) updateDatesOfStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.StudyPlanItem{}
	sspe := &entities.StudentStudyPlan{}
	spe := &entities.StudyPlan{}

	query := fmt.Sprintf(`
	WITH tmp AS (
		SELECT study_plan_id FROM %s
		JOIN %s as sp
		USING (study_plan_id)
		WHERE student_id = $1
		AND sp.master_study_plan_id = ANY($2)
	)
	SELECT study_plan_item_id
	FROM study_plan_items
	WHERE study_plan_id IN(SELECT * FROM tmp)
`, sspe.TableName(), spe.TableName())

	rows, err := s.DB.Query(ctx, query, &stepState.UserId, &stepState.StudyPlanIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	var studyPlanItemIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studyPlanItemIDs = append(studyPlanItemIDs, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudyPlanItemIDs = studyPlanItemIDs
	oneThird := int32(len(studyPlanItemIDs) / 3)
	stepState.StudyPlanItemIDsOverDue = studyPlanItemIDs[:oneThird]
	stepState.StudyPlanItemIDsCompleted = studyPlanItemIDs[oneThird : oneThird*2]
	stepState.StudyPlanItemIDsActive = studyPlanItemIDs[oneThird*2:]

	// update startDate, availableFrom, availableTo
	query = fmt.Sprintf(`
	WITH tmp AS (
		SELECT study_plan_id FROM %s
		JOIN %s as sp
		USING (study_plan_id)
		WHERE student_id = $1
		AND sp.master_study_plan_id = ANY($2)
	)
	UPDATE %s
	SET start_date = NOW() - INTERVAL '30 DAYS', available_from = NOW() - INTERVAL '30 DAYS'
	WHERE study_plan_id IN(SELECT * FROM tmp)
`, sspe.TableName(), spe.TableName(), e.TableName())

	cmd, err := s.DB.Exec(ctx, query, &stepState.UserId, &stepState.StudyPlanIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update start_date/available_from/available_to study plan item: %w", err)
	}
	if cmd.RowsAffected() != int64(len(stepState.StudyPlanItemInfos)*len(stepState.StudyPlanIDs)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("update start_date/available_from/available_to is wrong, expect %v but got %v rows", len(stepState.StudyPlanItemInfos)*len(stepState.StudyPlanIDs), cmd.RowsAffected())
	}

	// update dueDate
	// 1. active ignore
	// 2. overdue
	query = fmt.Sprintf(`UPDATE %s
	SET start_date = NOW() - INTERVAL '30 DAYS', available_from = NOW() - INTERVAL '30 DAYS', end_date = NOW() - INTERVAL '3 DAYS' 
	WHERE study_plan_item_id = ANY($1)
`, e.TableName())

	cmd, err = s.DB.Exec(ctx, query, &stepState.StudyPlanItemIDsOverDue)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update start_date/available_from/available_to/end_date study plan item (overdue): %w", err)
	}
	if cmd.RowsAffected() != int64(len(stepState.StudyPlanItemIDsOverDue)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("update start_date/available_from/available_to/end_date (overdue) is wrong, expect %v but got %v rows", len(stepState.StudyPlanItemIDsOverDue), cmd.RowsAffected())
	}

	// --  some studyPlanItems (overdue) will be deleted
	query = fmt.Sprintf(`UPDATE %s
	SET deleted_at = NOW()
	WHERE study_plan_item_id = ANY($1)
`, e.TableName())

	stepState.StudyPlanItemIDsOverDueDeleted = stepState.StudyPlanItemIDsOverDue[:2]
	cmd, err = s.DB.Exec(ctx, query, &stepState.StudyPlanItemIDsOverDueDeleted)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to deleted_at study plan item (overdue): %w", err)
	}
	if cmd.RowsAffected() != int64(len(stepState.StudyPlanItemIDsOverDueDeleted)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("update deleted_at (overdue) is wrong, expect %v but got %v rows", len(stepState.StudyPlanItemIDsOverDueDeleted), cmd.RowsAffected())
	}

	// 3.completed
	query = fmt.Sprintf(`UPDATE %s
	SET start_date = NOW() - INTERVAL '30 DAYS', available_from = NOW() - INTERVAL '30 DAYS', completed_at = Now()
	WHERE study_plan_item_id = ANY($1)
`, e.TableName())

	cmd, err = s.DB.Exec(ctx, query, &stepState.StudyPlanItemIDsCompleted)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to update start_date/available_from/available_to/completed_at study plan item (completed): %w", err)
	}
	if cmd.RowsAffected() != int64(len(stepState.StudyPlanItemIDsCompleted)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("update start_date/available_from/available_to/completed_at (completed) is wrong, expect %v but got %v rows", len(stepState.StudyPlanItemIDsCompleted), cmd.RowsAffected())
	}

	return StepStateToContext(ctx, stepState), nil
}
