package eureka

import (
	"context"
	"fmt"
	"time"

	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	common "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) ourSystemMustDeleteStudentSubmissionCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentSubmissionID := stepState.Request.(*pb.DeleteStudentSubmissionRequest).StudentSubmissionId

	var (
		studyPlanItemId pgtype.Text
		deletedAt       pgtype.Timestamptz
		completedAt     pgtype.Timestamptz
		deletedBy       pgtype.Text
	)

	if err := database.Select(ctx, s.DB, `
      SELECT study_plan_item_id
			FROM student_submissions
			WHERE student_submission_id = $1
			LIMIT 1
  `, database.Text(studentSubmissionID)).ScanFields(&studyPlanItemId); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can not get study_plan_item_id from student_submission_id = %v: %w", studentSubmissionID, err)
	}

	selectStudentLatestSubmissionStmt := `
		SELECT deleted_at, deleted_by
		FROM student_latest_submissions 
		WHERE study_plan_item_id = $1;`

	if err := database.Select(ctx, s.DB, selectStudentLatestSubmissionStmt, &studyPlanItemId).
		ScanFields(&deletedAt, &deletedBy); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if deletedAt.Status != pgtype.Present {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected deleted_at not null, got null, %v", deletedAt)
	}

	if deletedAt.Time.IsZero() || deletedBy.String == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected deleted_at and deleted_by not null, got null, %v %v", deletedAt, deletedBy.String)
	}

	if err := database.Select(ctx, s.DB, `
  SELECT completed_at
  FROM study_plan_items
  WHERE study_plan_item_id = $1
  LIMIT 1`, &studyPlanItemId).
		ScanFields(&completedAt); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if completedAt.Status != pgtype.Null {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected completed_at null, got %v", completedAt)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherRemoveStudentSubmission(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// get student_submission to delete
	selectStmt := "SELECT student_submission_id FROM student_submissions WHERE deleted_at IS NULL LIMIT 1"
	var studentSubmissionID string
	if err := database.Select(ctx, s.DB, selectStmt).ScanFields(&studentSubmissionID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	deleteStudentSubmissionReq := &pb.DeleteStudentSubmissionRequest{
		StudentSubmissionId: studentSubmissionID,
	}

	stepState.Response, stepState.ResponseErr = pb.NewStudentSubmissionModifierServiceClient(s.Conn).DeleteStudentSubmission(contextWithToken(s, ctx), deleteStudentSubmissionReq)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	stepState.Request = deleteStudentSubmissionReq

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherRemoveStudentSubmissionAfterListSubmissions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, _ := stepState.Response.(*pb.ListSubmissionsResponse)

	deleteStudentSubmissionReq := &pb.DeleteStudentSubmissionRequest{
		StudentSubmissionId: resp.Items[0].SubmissionId,
	}
	stepState.Request = deleteStudentSubmissionReq

	stepState.Response, stepState.ResponseErr = pb.NewStudentSubmissionModifierServiceClient(s.Conn).DeleteStudentSubmission(contextWithToken(s, ctx), deleteStudentSubmissionReq)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), nil
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) responseSubmissionsDontContainSubmissionsWereDeleted(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.DeleteStudentSubmissionRequest)
	studentSubmissionID := req.StudentSubmissionId
	teacherID := idutil.ULIDNow()

	if _, err := s.aValidUser(ctx, teacherID, consta.RoleTeacher); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.generateExchangeToken(teacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = token
	ctx = contextWithToken(s, ctx)
	resp, err := pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		ListSubmissions(ctx, &pb.ListSubmissionsRequest{
			Paging: &common.Paging{
				Limit: 100,
			},
			CourseId: wrapperspb.String(stepState.CourseID),
			Start:    timestamppb.New(time.Now().Add(-10 * time.Second)),
			End:      timestamppb.New(time.Now().Add(10 * time.Second)),
		})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, item := range resp.Items {
		if item.SubmissionId == studentSubmissionID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected student submission %v will ignore because it was deleted", studentSubmissionID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
