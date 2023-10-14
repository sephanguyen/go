package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	common "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
)

func (s *suite) someStudentHasSubmissionWithStatus(ctx context.Context, stringStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err1 := s.someStudentHasTheirSubmissionGraded(ctx)
	ctx, err2 := s.teacherChangeStudentsSubmissionStatusTo(ctx, "SUBMISSION_STATUS_RETURNED")
	err := multierr.Combine(err1, err2)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) validateEqualGrade(ctx context.Context, a *pb.SubmissionGrade, b *pb.SubmissionGrade) (context.Context, bool) {
	stepState := StepStateFromContext(ctx)
	flag := isGradeEqual(a.Grade, b.Grade)
	flag = flag && (a.Note == b.Note)
	flag = flag && (a.SubmissionId == b.SubmissionId)
	return StepStateToContext(ctx, stepState), flag
}

func (s *suite) eurekaMustReturnCorrectGradesForEachSubmission(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveSubmissionGradesRespose)
	for _, grade := range rsp.Grades {
		submissionGradeID := grade.SubmissionGradeId
		rspGrade := grade.Grade
		storedGrade := stepState.LatestGrade[submissionGradeID]
		ctx, v := s.validateEqualGrade(ctx, rspGrade, storedGrade)
		if !v {
			return StepStateToContext(ctx, stepState), fmt.Errorf("response grade %s does not match with stored grade %s", rspGrade.SubmissionId, storedGrade.SubmissionId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) teacherRetrieveStudentGradeBaseOnSubmissionGradeId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := generateValidAuthenticationToken(idutil.ULIDNow(), common.UserGroup_USER_GROUP_TEACHER.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	gradeIDs := make([]string, 0, len(stepState.Submissions))
	for _, submission := range stepState.Submissions {
		gradeIDs = append(gradeIDs, submission.Submission.SubmissionGradeId.GetValue())
	}

	client := pb.NewStudentAssignmentReaderServiceClient(s.Conn)
	stepState.Response, stepState.ResponseErr = client.RetrieveSubmissionGrades(contextWithToken(s, ctx), &pb.RetrieveSubmissionGradesRequest{
		SubmissionGradeIds: gradeIDs,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentRetrieveTheirGrade(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveSubmissionsResponse)
	gradeIDs := make([]string, 0, len(rsp.Items))
	for _, item := range rsp.Items {
		gradeIDs = append(gradeIDs, item.SubmissionGradeId.GetValue())
	}

	client := pb.NewStudentAssignmentReaderServiceClient(s.Conn)
	stepState.Response, stepState.ResponseErr = client.RetrieveSubmissionGrades(contextWithToken(s, ctx), &pb.RetrieveSubmissionGradesRequest{
		SubmissionGradeIds: gradeIDs,
	})

	return StepStateToContext(ctx, stepState), nil
}
