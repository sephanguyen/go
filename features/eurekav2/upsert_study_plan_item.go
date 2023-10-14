package eurekav2

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) thereAreStudyPlanCreatedInCourses(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studyPlanID := idutil.ULIDNow()

	// TODO: replace with insert study plan grpc call
	_, err := s.EurekaDBTrace.Exec(ctx, `
		INSERT INTO lms_study_plans (study_plan_id, name, course_id, academic_year, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6, $7)`,
		studyPlanID, "study-plan-direct-insert", stepState.CourseID, "random-academic-year", "STUDY_PLAN_STATUS_ACTIVE", time.Now(), time.Now())
	if err != nil {
		return ctx, fmt.Errorf("insert study plan: %w", err)
	}
	s.StudyPlanID = studyPlanID
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreatesNewStudyPlanItem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	stepState.Response, stepState.ResponseErr = epb.NewStudyPlanItemServiceClient(s.Connections.EurekaConn).UpsertStudyPlanItem(ctx, &epb.UpsertStudyPlanItemRequest{
		StudyPlanId:  stepState.StudyPlanID,
		Name:         "study-plan-item",
		LmIds:        []string{stepState.LearningMaterialID},
		StartDate:    timestamppb.Now(),
		EndDate:      timestamppb.Now(),
		DisplayOrder: 1,
	})

	return StepStateToContext(ctx, stepState), nil
}
