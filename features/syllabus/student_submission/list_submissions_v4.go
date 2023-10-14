package student_submission

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) userUsingListSubmissionsV4(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Response, stepState.ResponseErr = sspb.NewStudentSubmissionServiceClient(s.EurekaConn).ListSubmissionsV4(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListSubmissionsV4Request{
		LocationIds: stepState.LocationIDs,
		CourseId:    wrapperspb.String(stepState.CourseID),
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
		Start: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		End:   timestamppb.New(time.Now().Add(24 * time.Hour)),
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUsingListSubmissionsVWithStudentName(ctx context.Context, args string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var studentName string
	if args == "true" {
		studentName = "valid-user"
	} else {
		studentName = "wrong-user"
	}

	stepState.Response, stepState.ResponseErr = sspb.NewStudentSubmissionServiceClient(s.EurekaConn).ListSubmissionsV4(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListSubmissionsV4Request{
		StudentName: wrapperspb.String(studentName),
		LocationIds: stepState.LocationIDs,
		CourseId:    wrapperspb.String(stepState.CourseID),
		Paging: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
		Start: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		End:   timestamppb.New(time.Now().Add(24 * time.Hour)),
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsListStudentSubmissionsCorrectlyWith(ctx context.Context, args string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp := stepState.Response.(*sspb.ListSubmissionsV4Response)
	if args == "false" {
		if len(resp.Items) != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("search name return wrong, expected: 0 items, get: %d", len(resp.Items))
		}
	}

	for _, item := range resp.Items {
		studyPlanItem := item.StudyPlanItemIdentity
		if studyPlanItem.LearningMaterialId == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong LearningMaterialId of %v", item.SubmissionId)
		}
		if studyPlanItem.StudyPlanId == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong StudyPlanId of %v", item.SubmissionId)
		}
		if studyPlanItem.StudentId == nil || studyPlanItem.StudentId.Value == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong StudentId of %v", item.SubmissionId)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
