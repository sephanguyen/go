package eureka

import (
	"context"
	"math/rand"

	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) assignAssignmentToTopic(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.AssignAssignmentsToTopicRequest{
		TopicId: "topic-1",
		Assignment: []*pb.AssignAssignmentsToTopicRequest_Assignment{
			{
				AssignmentId: stepState.Assignments[0].AssignmentId,
				DisplayOrder: int32(rand.Intn(5)),
			},
		},
	}

	for _, assignment := range stepState.Assignments {
		req.Assignment = append(req.Assignment, &pb.AssignAssignmentsToTopicRequest_Assignment{
			AssignmentId: assignment.AssignmentId,
			DisplayOrder: int32(rand.Intn(5)),
		})
	}
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).
		AssignAssignmentsToTopic(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
