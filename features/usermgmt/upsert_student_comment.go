package usermgmt

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *suite) aValidUpsertStudentCommentRequestWith(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudentID = newID()
	if _, err := s.aValidStudentInDB(ctx, stepState.StudentID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	commentContent := fmt.Sprintf("This is just for test %s", newID())

	stepState.Request = &pb.UpsertStudentCommentRequest{
		StudentComment: &pb.StudentComment{
			StudentId:      stepState.StudentID,
			CommentContent: commentContent,
		},
	}

	switch condition {
	case "new comment":
		stepState.Request.(*pb.UpsertStudentCommentRequest).StudentComment.CommentId = ""
	case "existing comment":
		_ctx, _ := s.upsertCommentForStudent(StepStateToContext(ctx, stepState))
		_stepState := StepStateFromContext(_ctx)
		if _stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), _stepState.ResponseErr
		}

		var commentID string
		stmt := `SELECT comment_id FROM student_comments where coach_id = $1 AND student_id = $2 AND comment_content = $3`
		err := s.BobDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID, stepState.StudentID, commentContent).Scan(&commentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Request.(*pb.UpsertStudentCommentRequest).StudentComment.CommentId = commentID
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertCommentForStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewStudentServiceClient(s.UserMgmtConn).UpsertStudentComment(contextWithToken(ctx), stepState.Request.(*pb.UpsertStudentCommentRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) bobDBMustStoreCommentForStudent(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var stmt string
	switch action {
	case "store":
		stmt = `SELECT count(*) FROM student_comments where coach_id = $1 AND student_id = $2`
	case "update":
		stmt = `SELECT count(*) FROM student_comments where coach_id = $1 AND student_id = $2 AND updated_at > created_at`
	}
	var count int
	err := s.BobDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID, stepState.StudentID).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot find any comment with student_id is %v, coach_id is %v", stepState.StudentID, stepState.CurrentUserID)
	}

	return StepStateToContext(ctx, stepState), nil
}
