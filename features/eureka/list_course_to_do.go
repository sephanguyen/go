package eureka

import (
	"context"
	"fmt"
	"reflect"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) listCourseStudyPlansToDoItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	teacherID := idutil.ULIDNow()

	if _, err := s.aValidUser(ctx, teacherID, constant.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(teacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = token

	ctx, courseStudyPlanIDs, err := s.getStudyPlanIDOfCourse(ctx, stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response, stepState.ResponseErr = pb.NewAssignmentReaderServiceClient(s.Conn).ListCourseTodo(contextWithToken(s, ctx), &pb.ListCourseTodoRequest{
		StudyPlanId: courseStudyPlanIDs[0],
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsListOfToDoItemsWithCorrectStatisticInfor(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.returnsStatusCode(ctx, "OK")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp, ok := stepState.Response.(*pb.ListCourseTodoResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect response type pb.ListCourseTodoResponse")
	}

	ctx, courseStudyPlanIDs, err := s.getStudyPlanIDOfCourse(ctx, stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	const studyPlanItemOrderedQuery = `
SELECT spi.study_plan_item_id
FROM study_plan_items spi
JOIN books b ON b.book_id = content_structure ->> 'book_id'
JOIN chapters c ON c.chapter_id = content_structure ->> 'chapter_id'
JOIN topics t ON t.topic_id = content_structure ->> 'topic_id'
LEFT JOIN topics_assignments ta ON ta.topic_id = content_structure ->> 'topic_id'
AND ta.assignment_id = content_structure ->> 'assignment_id'
LEFT JOIN topics_learning_objectives tlo ON tlo.topic_id = content_structure ->> 'topic_id'
AND tlo.lo_id = content_structure ->> 'lo_id'
WHERE study_plan_id = $1::TEXT
AND b.deleted_at IS NULL
AND c.deleted_at IS NULL
AND t.deleted_at IS NULL
AND spi.deleted_at IS NULL
ORDER BY b.book_id,
         c.display_order,
         t.display_order,
         coalesce(ta.display_order, tlo.display_order)
  `

	var spiOrdered []string
	rows, err := s.DB.Query(ctx, studyPlanItemOrderedQuery, database.Text(courseStudyPlanIDs[0]))

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when get study_plan_items from DB %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ID string
		if err := rows.Scan(&ID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when get study_plan_items from DB %w", err)
		}
		spiOrdered = append(spiOrdered, ID)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when get study_plan_items from DB %w", err)
	}

	if len(spiOrdered) != len(resp.StatisticItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v study plan items for study plan but got %v ", len(spiOrdered), len(resp.StatisticItems))
	}

	var respOrdered []string
	for _, item := range resp.StatisticItems {
		respOrdered = append(respOrdered, item.GetItem().GetStudyPlanItem().GetStudyPlanItemId())
	}

	if !reflect.DeepEqual(spiOrdered, respOrdered) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected order study plan item \nwant = %v\ngot = %v\n", spiOrdered, respOrdered)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudyPlanIDOfCourse(ctx context.Context, courseID string) (context.Context, []string, error) {
	stepState := StepStateFromContext(ctx)
	query := `SELECT study_plan_id FROM course_study_plans WHERE course_id = $1`
	rows, err := s.DB.Query(ctx, query, courseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	studyPlanID := make([]string, 0)

	for rows.Next() {
		var id pgtype.Text
		err := rows.Scan(&id)
		if err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		studyPlanID = append(studyPlanID, id.String)
	}

	return StepStateToContext(ctx, stepState), studyPlanID, nil
}

func (s *suite) listCourseStudyPlansToDoItemsWith(ctx context.Context, failedCase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	teacherID := idutil.ULIDNow()

	if _, err := s.aValidUser(ctx, teacherID, constant.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(teacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = token

	var req *pb.ListCourseTodoRequest
	switch failedCase {
	case "empty study plan":
		req = &pb.ListCourseTodoRequest{
			StudyPlanId: "",
		}
	case "not existed study plan":
		nonExistedStudyPlanID := idutil.ULIDNow()
		req = &pb.ListCourseTodoRequest{
			StudyPlanId: nonExistedStudyPlanID,
		}
	}
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentReaderServiceClient(s.Conn).ListCourseTodo(contextWithToken(s, ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsEmptyListOfToDoItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.returnsStatusCode(ctx, "OK")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp, ok := stepState.Response.(*pb.ListCourseTodoResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect response type pb.ListCourseTodoResponse")
	}

	if len(resp.StatisticItems) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect response have empty result")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) deleteStudyPlanItemsByStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, courseStudyPlanIDs, err := s.getStudyPlanIDOfCourse(ctx, stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	const softDeleteByStudyPlanIDs = `
UPDATE study_plan_items SET deleted_at = NOW() 
WHERE study_plan_id = ANY($1)
`
	studyPlanIDs := database.TextArray([]string{courseStudyPlanIDs[0]})
	if _, err := db.Exec(ctx, softDeleteByStudyPlanIDs, &studyPlanIDs); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
