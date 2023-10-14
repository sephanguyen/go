package student_submission // nolint

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userUsingListSubmissionsV2WithStudentName(ctx context.Context, args string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var studentName string
	if args == "true" {
		studentName = "valid-user"
	} else {
		studentName = "wrong-user"
	}

	stepState.Response, stepState.ResponseErr = epb.NewStudentAssignmentReaderServiceClient(s.EurekaConn).ListSubmissionsV2(
		s.AuthHelper.SignedCtx(ctx, stepState.Token),
		&epb.ListSubmissionsV2Request{
			CourseId:    wrapperspb.String(stepState.CourseID),
			LocationIds: stepState.LocationIDs,
			Start:       timestamppb.New(time.Now().Add(-24 * time.Hour)),
			End:         timestamppb.New(time.Now().Add(24 * time.Hour)),
			Paging: &cpb.Paging{
				Limit: 10,
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: 1,
				},
			},
			StudentName: wrapperspb.String(studentName),
		})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsListStudentSubmissionsV2CorrectlyWith(ctx context.Context, args string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp := stepState.Response.(*epb.ListSubmissionsV2Response)
	if args == "false" {
		if len(resp.Items) != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("search name return wrong, expected: 0 items, get: %d", len(resp.Items))
		}
	}

	for _, item := range resp.Items {
		if item.StudentId == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong StudentId of %v", item.SubmissionId)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
