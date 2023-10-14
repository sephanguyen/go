package student_progress

import (
	"context"

	"github.com/manabie-com/backend/features/syllabus/utils"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) schoolAdminParentTeacherAndStudentLogin(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err := s.aSignedIn(ctx, "parent")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.aSignedIn(ctx, "school admin")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aSignedIn(ctx, "teacher")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aSignedIn(ctx, "student")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) parentCalculateStudentProgress(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Token = stepState.Parent.Token

	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("student-id", stepState.Student.ID, "token", stepState.Token, "pkg", "com.manabie.liz", "version", "1.0.0"))
	stepState.Response, stepState.ResponseErr = sspb.NewStatisticsClient(s.EurekaConn).GetStudentProgress(ctx, &sspb.GetStudentProgressRequest{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId: stepState.StudyPlanID,
			StudentId:   wrapperspb.String(stepState.Student.ID),
		},
		CourseId: stepState.CourseID,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}
