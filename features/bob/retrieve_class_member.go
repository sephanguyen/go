package bob

import (
	"context"

	"github.com/pkg/errors"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) aRetrieveClassMemberRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = &pb.RetrieveClassMemberRequest{}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userRetrieveClassMember(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).RetrieveClassMember(contextWithToken(s, ctx), stepState.Request.(*pb.RetrieveClassMemberRequest))

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aValidClassIDInRetrieveClassMemberRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request.(*pb.RetrieveClassMemberRequest).ClassId = stepState.CurrentClassID
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsStudentsAndTeachersRetrieveClassMemberResponse(ctx context.Context, expectedTotalStudent, expectedTotalTeacher int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.RetrieveClassMemberResponse)

	totalStudent := 0
	totalTeacher := 0
	for _, u := range resp.Members {
		switch u.UserGroup {
		case pb.USER_GROUP_STUDENT:
			totalStudent++
		case pb.USER_GROUP_TEACHER:
			totalTeacher++
		}
	}

	if totalStudent != expectedTotalStudent {
		return StepStateToContext(ctx, stepState), errors.Errorf("total student does not match, expected: %d, got: %d", expectedTotalStudent, totalStudent)
	}

	if totalTeacher != expectedTotalTeacher {
		return StepStateToContext(ctx, stepState), errors.Errorf("total student does not match, expected: %d, got: %d", expectedTotalTeacher, totalTeacher)
	}

	return StepStateToContext(ctx, stepState), nil
}
