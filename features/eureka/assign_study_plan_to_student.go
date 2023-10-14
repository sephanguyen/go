package eureka

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) eurekaMustAssignStudyPlanToStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var studyPlanID pgtype.Text
	_ = studyPlanID.Set(nil)
	studentID := stepState.Request.(*pb.AssignStudyPlanRequest).Data.(*pb.AssignStudyPlanRequest_StudentId).StudentId
	ctx, err := s.eurekaMustStoreStudentStudyPlan(ctx, []string{studentID}, studyPlanID)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) userAssignStudyPlanToAStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.AssignStudyPlanRequest{
		StudyPlanId: stepState.StudyPlanID,
		Data: &pb.AssignStudyPlanRequest_StudentId{
			StudentId: stepState.StudentIDs[0],
		},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentModifierServiceClient(s.Conn).AssignStudyPlan(contextWithToken(s, ctx), req)

	return StepStateToContext(ctx, stepState), nil
}
