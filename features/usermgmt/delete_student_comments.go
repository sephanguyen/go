package usermgmt

import (
	"context"
	"fmt"
	"math/rand"

	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
)

func (s *suite) aStudentWithSomeComments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := newID()

	ctx, err := s.aValidStudentInDB(ctx, id)
	if err != nil {
		return ctx, err
	}

	var studentID pgtype.Text
	err = studentID.Set(id)
	if err != nil {
		return ctx, err
	}
	stepState.AssignedStudentIDs = append(stepState.AssignedStudentIDs, id)

	for i := 0; i < rand.Intn(5)+5; i++ {
		commentContent := fmt.Sprintf("random comment %s %d", studentID.String, i)
		cmtID := newID()
		stepState.CommentIDs = append(stepState.CommentIDs, cmtID)
		upsertStudentCommentRequest := &pb.UpsertStudentCommentRequest{
			StudentComment: &pb.StudentComment{
				StudentId:      studentID.String,
				CommentContent: commentContent,
				CommentId:      cmtID,
			},
		}
		stepState.Request = upsertStudentCommentRequest
		_, err := s.upsertCommentForStudent(ctx)
		if err != nil {
			return ctx, err
		}
		stepState = StepStateFromContext(ctx)
		if stepState.ResponseErr != nil {
			return ctx, stepState.ResponseErr
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theSignedInUserDeleteStudentComments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewStudentServiceClient(s.UserMgmtConn).DeleteStudentComments(contextWithToken(ctx), &pb.DeleteStudentCommentsRequest{
		CommentIds: []string{stepState.CommentIDs[0]},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theSignedInUserDeleteStudentCommentsWithNilCmtIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = pb.NewStudentServiceClient(s.UserMgmtConn).DeleteStudentComments(contextWithToken(ctx), &pb.DeleteStudentCommentsRequest{
		CommentIds: nil,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theSignedInUserDeleteStudentCommentsButCommentNotExist(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randomCmtId := newID()
	stepState.Response, stepState.ResponseErr = pb.NewStudentServiceClient(s.UserMgmtConn).DeleteStudentComments(contextWithToken(ctx), &pb.DeleteStudentCommentsRequest{
		CommentIds: []string{randomCmtId},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToStoreCommentsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	retrieveStudentCommentsResponse, err := pb.NewStudentServiceClient(s.UserMgmtConn).RetrieveStudentComment(contextWithToken(ctx), &pb.RetrieveStudentCommentRequest{
		StudentId: stepState.AssignedStudentIDs[0],
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve student comment")
	}
	if len(retrieveStudentCommentsResponse.GetComment()) != len(stepState.CommentIDs)-1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of comment in db, want %d, actual: %d", len(stepState.CommentIDs)-1, len(retrieveStudentCommentsResponse.GetComment()))
	}
	return StepStateToContext(ctx, stepState), nil
}
