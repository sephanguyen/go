package study_plan // nolint

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) aValidAllocateMarker(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	allocMarkerRepo := repositories.AllocateMarkerRepo{}

	uid := idutil.ULIDNow()

	allocMarker := &entities.AllocateMarker{
		AllocateMarkerID:   database.Text(uid),
		TeacherID:          database.Text(fmt.Sprintf("TeacherID_%s", uid)),
		StudentID:          database.Text(fmt.Sprintf("StudentID_%s", uid)),
		StudyPlanID:        database.Text(fmt.Sprintf("StudyPlanID_%s", uid)),
		LearningMaterialID: database.Text(fmt.Sprintf("LearningMaterialID_%s", uid)),
		CreatedBy:          database.Text(fmt.Sprintf("CreatedBy_%s", uid)),
	}
	allocMarker.BaseEntity.Now()

	stepState.AllocateMarker = allocMarker

	err := allocMarkerRepo.BulkUpsert(ctx, s.EurekaDB, []*entities.AllocateMarker{
		allocMarker,
	})

	if err != nil {
		return ctx, fmt.Errorf("AllocateMarker.BulkUpsert failed, %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userRetrieveAllocateMarker(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	allocMarker := stepState.AllocateMarker

	stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).RetrieveAllocateMarker(s.AuthHelper.SignedCtx((ctx), stepState.Token), &sspb.RetrieveAllocateMarkerRequest{
		StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
			StudyPlanId:        allocMarker.StudyPlanID.String,
			LearningMaterialId: allocMarker.LearningMaterialID.String,
			StudentId:          wrapperspb.String(allocMarker.StudentID.String),
		},
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnAllocateMarkerCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	response := stepState.Response.(*sspb.RetrieveAllocateMarkerResponse)
	allocMarker := stepState.AllocateMarker

	if response.MarkerId != allocMarker.TeacherID.String {
		return ctx, fmt.Errorf("RetrieveAllocateMarkerResponse expect MarkerId is %s but got %s", allocMarker.TeacherID.String, response.MarkerId)
	}

	return ctx, nil
}
