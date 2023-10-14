package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/yasuo/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) eurekaMustReturnAssignmentsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveAssignmentsResponse)
	req := stepState.Request.(*pb.RetrieveAssignmentsRequest)
	assignmentIDs := req.Ids
	if len(assignmentIDs) != len(req.Ids) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %d is not in array: %d", len(assignmentIDs), len(req.Ids))
	}

	for _, item := range rsp.Items {
		if !utils.IsContain(assignmentIDs, item.AssignmentId) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("assignment is not requested: %s", item.AssignmentId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userListAssignmentsByIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	assignments := stepState.Request.(*pb.UpsertAssignmentsRequest).Assignments
	assignmentIDs := make([]string, len(assignments))
	for i, assignment := range assignments {
		assignmentIDs[i] = assignment.AssignmentId
	}

	req := &pb.RetrieveAssignmentsRequest{
		Ids: assignmentIDs,
	}

	stepState.Response, stepState.ResponseErr = pb.NewAssignmentReaderServiceClient(s.Conn).RetrieveAssignments(contextWithToken(s, ctx), req)

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}
