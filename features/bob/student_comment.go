package bob

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) aUserCommentForHisStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.AssignedStudentIDs) == 0 {
		//for invalid test case
		stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, stepState.studentID)
	}

	cmtID := s.newID()
	stepState.CommentIDs = append(stepState.CommentIDs, cmtID)
	stepState.Request = &pb.UpsertStudentCommentRequest{
		StudentComment: &pb.StudentComment{
			CommentId:      cmtID,
			StudentId:      stepState.AssignedStudentIDs[0],
			CoachId:        stepState.CurrentTeacherID,
			CommentContent: "Some comment content" + stepState.AssignedStudentIDs[0],
		},
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userUpsertCommentForHisStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewStudentClient(s.Conn).UpsertStudentComment(s.signedCtx(ctx), stepState.Request.(*pb.UpsertStudentCommentRequest))
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustStoreCommentForStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	rsp := stepState.Response.(*pb.UpsertStudentCommentResponse)
	if !rsp.Successful {
		return StepStateToContext(ctx, stepState), errors.New("cannot insert into student comment table")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userRetrieveCommentForStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewStudentClient(s.Conn).RetrieveStudentComment(s.signedCtx(ctx), &pb.RetrieveStudentCommentRequest{
		StudentId: stepState.AssignedStudentIDs[0],
	})
	return StepStateToContext(ctx, stepState), nil

}
func (s *suite) validCommentForStudentInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token := stepState.AuthToken
	s.anAssignedStudent(ctx)
	s.aUserCommentForHisStudent(ctx)
	s.userUpsertCommentForHisStudent(ctx)
	s.aUserCommentForHisStudent(ctx)

	s.userUpsertCommentForHisStudent(ctx)
	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnAllCommentForStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	res := stepState.Response.(*pb.RetrieveStudentCommentResponse)
	req := stepState.Request.(*pb.UpsertStudentCommentRequest)
	flag := false
	for _, comment := range res.Comment {
		studentComment := comment.StudentComment
		if studentComment.StudentId == req.StudentComment.StudentId {
			flag = true
		}
	}
	for i := 0; i < len(res.GetComment()); i++ {
		for j := i + 1; j < len(res.GetComment())-1; j++ {
			if res.GetComment()[i].StudentComment.CreatedAt.Compare(res.GetComment()[j].StudentComment.CreatedAt) != -1 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("wrong order created_at comment: expected: %s > %s, actual: %s > %s", res.GetComment()[i].StudentComment.CreatedAt.String(), res.GetComment()[j].StudentComment.CreatedAt, res.GetComment()[j].StudentComment.CreatedAt, res.GetComment()[i].StudentComment.CreatedAt)
			}
		}
	}

	if !flag {
		return StepStateToContext(ctx, stepState), errors.New("student Id return invalid")
	}

	if len(res.Comment) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("our system did not return all comment of student")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aTeacherGivesSomeCommentsForStudent(ctx context.Context) (context.Context, error) {
	return s.aStudentWithSomeComments(ctx)
}
func (s *suite) theTeacherRetrievesCommentForStudent(ctx context.Context) (context.Context, error) {
	return s.userRetrieveCommentForStudent(ctx)
}
func (s *suite) ourSystemHaveToResponseRetrieveCommentCorrectly(ctx context.Context) (context.Context, error) {
	return s.bobMustReturnAllCommentForStudent(ctx)
}
