package eureka

import (
	"context"
	"fmt"
	"reflect"

	eureka_repo "github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) retrieveLearningObjectivesWith(ctx context.Context, params string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var req *epb.RetrieveLOsRequest
	switch params {
	case "TopicIds":
		req = &epb.RetrieveLOsRequest{
			StudentId: stepState.CurrentStudentID,
			TopicIds:  []string{stepState.TopicID},
		}
	case "LoIds":
		req = &epb.RetrieveLOsRequest{
			StudentId: stepState.CurrentStudentID,
			LoIds:     stepState.LoIDs,
		}
	case "WithCompleteness":
		req = &epb.RetrieveLOsRequest{
			StudentId:        stepState.CurrentStudentID,
			LoIds:            stepState.LoIDs,
			WithCompleteness: true,
		}
	case "WithAchievementCrown":
		req = &epb.RetrieveLOsRequest{
			StudentId:            stepState.CurrentStudentID,
			TopicIds:             []string{stepState.TopicID},
			WithAchievementCrown: true,
		}
	default:
		req = &epb.RetrieveLOsRequest{}
	}
	stepState.Response, stepState.ResponseErr = epb.NewCourseReaderServiceClient(s.Conn).RetrieveLOs(s.signedCtx(ctx), req)
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustReturnLearningObjectivesCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := stepState.Response.(*epb.RetrieveLOsResponse)
	req := stepState.Request.(*epb.RetrieveLOsRequest)

	if len(resp.LearningObjectives) != len(stepState.LearningObjectives) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of learning objectives: want: %d got: %d", len(stepState.LearningObjectives), len(resp.LearningObjectives))
	}
	for i, lo := range resp.LearningObjectives {
		if stepState.LearningObjectives[i].Info.Id != lo.Info.Id {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective id: want: %s got: %s", stepState.LearningObjectives[i].Info.Id, lo.Info.Id)
		}
		if stepState.LearningObjectives[i].Info.Name != lo.Info.Name {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective name: want: %s got: %s", stepState.LearningObjectives[i].Info.Name, lo.Info.Name)
		}
		if !reflect.DeepEqual(stepState.LearningObjectives[i].Video, lo.Video) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective video: want: %v got: %v", stepState.LearningObjectives[i].Video, lo.Video)
		}
		if !reflect.DeepEqual(stepState.LearningObjectives[i].Prerequisites, lo.Prerequisites) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective Prerequisites: want: %v got: %v", stepState.LearningObjectives[i].Prerequisites, lo.Prerequisites)
		}
		if stepState.LearningObjectives[i].GradeToPass.String() != lo.GradeToPass.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective grade to pass: want: %s got: %s", stepState.LearningObjectives[i].GradeToPass.String(), lo.GradeToPass.String())
		}
		if stepState.LearningObjectives[i].ManualGrading != lo.ManualGrading {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective manualGrading: want: %t got: %t", stepState.LearningObjectives[i].ManualGrading, lo.ManualGrading)
		}
		if stepState.LearningObjectives[i].TimeLimit.String() != lo.TimeLimit.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective time limit: want: %s got: %s", stepState.LearningObjectives[i].TimeLimit.String(), lo.TimeLimit.String())
		}
		if safeInt32Value(stepState.LearningObjectives[i].MaximumAttempt).Value != safeInt32Value(lo.MaximumAttempt).Value {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective maximum_attempt: want: %d got: %d", safeInt32Value(stepState.LearningObjectives[i].MaximumAttempt).Value, safeInt32Value(lo.MaximumAttempt).Value)
		}
		if stepState.LearningObjectives[i].ApproveGrading != lo.ApproveGrading {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective approve_grading: want: %t got: %t", stepState.LearningObjectives[i].ApproveGrading, lo.ApproveGrading)
		}
		if stepState.LearningObjectives[i].GradeCapping != lo.GradeCapping {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective grade_capping: want: %t got: %t", stepState.LearningObjectives[i].GradeCapping, lo.GradeCapping)
		}
		if stepState.LearningObjectives[i].ReviewOption != lo.ReviewOption {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective review_option: want: %s got: %s", stepState.LearningObjectives[i].ReviewOption.String(), lo.ReviewOption.String())
		}
		if stepState.LearningObjectives[i].VendorType != lo.VendorType {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning objective vendor_type: want: %s got: %s", stepState.LearningObjectives[i].VendorType.String(), lo.VendorType.String())
		}
	}
	if req.WithAchievementCrown {
		if len(resp.Crowns) != len(resp.LearningObjectives) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected length of Crowns: want: %d got: %d", len(resp.LearningObjectives), len(resp.Crowns))
		}
	}
	if req.WithCompleteness {
		if len(resp.Completenesses) != len(resp.LearningObjectives) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected length of Completenesses: want: %d got: %d", len(resp.LearningObjectives), len(resp.Completenesses))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func safeInt32Value(v *wrapperspb.Int32Value) wrapperspb.Int32Value {
	if v != nil {
		return *v
	}
	return wrapperspb.Int32Value{}
}

func (s *suite) someLoCompletenessesExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentStudentID == "" {
		stepState.CurrentStudentID = idutil.ULIDNow()
	}
	repo := eureka_repo.StudentsLearningObjectivesCompletenessRepo{}
	studentID := database.Text(stepState.CurrentStudentID)

	for _, loID := range stepState.LoIDs {
		loID := database.Text(loID)
		quizzScore := database.Float4(5.0)
		err := repo.UpsertFirstQuizCompleteness(ctx, s.DB, loID, studentID, quizzScore)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("can't set up some LO completenesses in db")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
