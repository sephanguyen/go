package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
)

func (s *suite) eurekaMustDeleteTheseAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	assignmentIDs := stepState.Request.(*pb.DeleteAssignmentsRequest).AssignmentIds

	res, err := pb.NewAssignmentReaderServiceClient(s.Conn).RetrieveAssignments(contextWithToken(s, ctx), &pb.RetrieveAssignmentsRequest{
		Ids: assignmentIDs,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to get assignments: %v", err)
	}

	if len(res.Items) != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("assignments were not deleted")
	}

	return StepStateToContext(ctx, stepState), nil
}

//nolint
func (s *suite) someAssignmentsInDb(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TopicID = idutil.ULIDNow()
	ctx, err1 := s.aValidToken(ctx, constants.RoleSchoolAdmin)
	ctx, err2 := s.userCreateNewAssignments(ctx)
	err := multierr.Combine(err1, err2)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) userDeleteAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	assignments := stepState.Request.(*pb.UpsertAssignmentsRequest).Assignments
	assignmentIDs := make([]string, len(assignments))
	for i, assignment := range assignments {
		assignmentIDs[i] = assignment.AssignmentId
	}

	req := &pb.DeleteAssignmentsRequest{
		AssignmentIds: assignmentIDs,
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).DeleteAssignments(contextWithToken(s, ctx), req)

	return StepStateToContext(ctx, stepState), nil
}
