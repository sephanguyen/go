package usermgmt

import (
	"context"
	"fmt"
	"reflect"

	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *suite) theSignedInUserRetrieveStudentComments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	stepState.Response, stepState.ResponseErr = pb.NewStudentServiceClient(s.UserMgmtConn).RetrieveStudentComment(contextWithToken(ctx), &pb.RetrieveStudentCommentRequest{
		StudentId: stepState.AssignedStudentIDs[0],
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getCommentsBelongToUserCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	commentsFromResp := stepState.Response.(*pb.RetrieveStudentCommentResponse).GetComment()
	if len(commentsFromResp) != len(stepState.CommentIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of comment in db, want %d, actual: %d", len(stepState.CommentIDs)-1, len(commentsFromResp))
	}
	commendIdsFromResp := make([]string, 0)
	for idx := range commentsFromResp {
		commendIdsFromResp = append(commendIdsFromResp, commentsFromResp[idx].GetStudentComment().CommentId)
	}
	if !reflect.DeepEqual(stepState.CommentIDs, commendIdsFromResp) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected comments in response not equal comments in database")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theSignedInUserRetrieveStudentCommentsWithNilStudentId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	stepState.Response, stepState.ResponseErr = pb.NewStudentServiceClient(s.UserMgmtConn).RetrieveStudentComment(contextWithToken(ctx), &pb.RetrieveStudentCommentRequest{
		StudentId: "",
	})
	return StepStateToContext(ctx, stepState), nil
}
