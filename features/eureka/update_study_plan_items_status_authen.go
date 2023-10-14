package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userUpdateStatusWithValidRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	ctx, studyPlanItems, err := s.getStudyPlanItemByStudyPlanID(ctx, database.TextArray([]string{stepState.StudyPlanID}))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error: %w", err)
	}

	ids := make([]string, 0, len(studyPlanItems))
	for _, item := range studyPlanItems {
		ids = append(ids, item.ID.String)
	}

	stepState.Response, stepState.ResponseErr = epb.NewStudyPlanModifierServiceClient(s.Conn).UpdateStudyPlanItemsStatus(ctx, &epb.UpdateStudyPlanItemsStatusRequest{
		StudentId:           stepState.StudentID,
		StudyPlanItemIds:    ids,
		StudyPlanItemStatus: epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ARCHIVED,
	})
	return StepStateToContext(ctx, stepState), nil
}
