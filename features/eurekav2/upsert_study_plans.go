package eurekav2

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"
)

func (s *suite) generateStudyPlan(template *epb.UpsertStudyPlanRequest) *epb.UpsertStudyPlanRequest {
	if template == nil {
		// A valid create course req template
		template = &epb.UpsertStudyPlanRequest{
			Name:         "name",
			CourseId:     "course_id_%s" + idutil.ULIDNow(),
			AcademicYear: "academic_year_id",
			Status:       epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		}
	}

	return template
}

func (s *suite) createNewStudyPlan(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	var req *epb.UpsertStudyPlanRequest
	switch validity {
	case "valid":
		req = s.generateStudyPlan(nil)
	case "invalid":
		req = s.generateStudyPlan(&epb.UpsertStudyPlanRequest{
			Name: "",
		})
	}

	stepState.Response, stepState.ResponseErr = epb.NewStudyPlanServiceClient(s.EurekaConn).UpsertStudyPlan(ctx, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkUpsertedStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := stepState.Response.(*epb.UpsertStudyPlanResponse).StudyPlanId
	query := "SELECT count(*) FROM lms_study_plans WHERE study_plan_id = $1"
	var count int
	if err := s.EurekaDB.QueryRow(ctx, query, id).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("eureka doesn't store correct study plan")
	}

	return StepStateToContext(ctx, stepState), nil
}
