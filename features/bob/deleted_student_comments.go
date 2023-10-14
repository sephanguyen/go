package bob

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/multierr"
)

func (s *suite) aStudentWithSomeComments(ctx context.Context) (context.Context, error) {
	ctx, err := s.anAssignedStudent(ctx)
	if err != nil {
		return ctx, err
	}
	for i := 0; i < rand.Intn(5)+5; i++ {
		// change created_at
		time.Sleep(time.Second)
		ctx, err1 := s.aUserCommentForHisStudent(ctx)
		ctx, err2 := s.userUpsertCommentForHisStudent(ctx)
		stepState := StepStateFromContext(ctx)
		if stepState.ResponseErr != nil {
			return ctx, stepState.ResponseErr
		}
		if err := multierr.Combine(err1, err2); err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) theTeacherDeleteStudentsComment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = bpb.NewStudentModifierServiceClient(s.Conn).DeleteStudentComments(s.signedCtx(ctx), &bpb.DeleteStudentCommentsRequest{
		CommentIds: []string{stepState.CommentIDs[0]},
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemsHaveToStoreCommentCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}

	resp, err := pb.NewStudentClient(s.Conn).RetrieveStudentComment(s.signedCtx(ctx), &pb.RetrieveStudentCommentRequest{
		StudentId: stepState.AssignedStudentIDs[0],
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve student comment")
	}
	if len(resp.GetComment()) != len(stepState.CommentIDs)-1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of comment in db, want %d, actual: %d", len(stepState.CommentIDs)-1, len(resp.GetComment()))
	}

	return StepStateToContext(ctx, stepState), nil
}
