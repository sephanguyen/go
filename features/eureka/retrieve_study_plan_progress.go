package eureka

import (
	"context"
	"fmt"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) teacherRetrievesStudyPlanProgressOfStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = contextWithToken(s, ctx)
	studentID := stepState.StudentIDs[0]

	stepState.Response, stepState.ResponseErr = pb.NewAssignmentReaderServiceClient(s.Conn).RetrieveStudyPlanProgress(ctx, &pb.RetrieveStudyPlanProgressRequest{
		StudentId:   studentID,
		StudyPlanId: stepState.StudentsSubmittedAssignments[studentID][0].StudyPlanItem.StudyPlanId,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsStudyPlanProgressOfStudentsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	resp := stepState.Response.(*pb.RetrieveStudyPlanProgressResponse)

	if resp.CompletedAssignments != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected completed assignments greater than 0")
	}
	if want := len(stepState.StudyPlanItemIDs); resp.TotalAssignments != int32(want) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total assignments, got: %d, want: %d", resp.TotalAssignments, want)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudentsAreAssignedSomeStudyPlanWithAvailableFrom(ctx context.Context, availableTime string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.someStudentsAreAssignedSomeValidStudyPlans(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var availableFrom pgtype.Timestamptz
	_ = availableFrom.Set(nil)
	if availableTime != "empty" {
		time, err := time.Parse("2006-01-02", availableTime)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		_ = availableFrom.Set(time)
	}
	ctx, err = s.studentSubmitTheirAssignment(ctx, "existed", "single")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	findStudyPlanItemIDsByCourse := `SELECT study_plan_item_id
	FROM public.study_plans sp JOIN public.study_plan_items spi2 ON sp.study_plan_id = spi2.study_plan_id
	WHERE sp.course_id = $1`

	rows, err := s.DB.Query(ctx, findStudyPlanItemIDsByCourse, &stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		ids = append(ids, id)
	}

	query := `UPDATE public.study_plan_items spi SET available_from = $1 WHERE spi.study_plan_item_id = ANY($2)`
	_, err = s.DB.Exec(ctx, query, &availableFrom, &ids)
	// query := `UPDATE public.study_plan_items spi SET available_from = $1 WHERE spi.study_plan_item_id = ANY(SELECT study_plan_item_id
	// 	FROM public.study_plans sp JOIN public.study_plan_items spi2 ON sp.study_plan_id = spi2.study_plan_id
	// 	WHERE sp.course_id = $2)`
	// _, err = s.DB.Exec(ctx, query, &availableFrom, &stepState.CourseID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsStudyPlanProgressOfStudentsAre(ctx context.Context, progress int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveStudyPlanProgressResponse)
	if rsp.CompletedAssignments != int32(progress) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d completed assignment got %d", progress, rsp.CompletedAssignments)
	}
	if rsp.TotalAssignments != int32(progress) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d total assignment got %d", progress, rsp.TotalAssignments)
	}
	return StepStateToContext(ctx, stepState), nil
}
