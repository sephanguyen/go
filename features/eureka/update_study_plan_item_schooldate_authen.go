package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) userUpdateSchoolDateWithValidRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)
	ctx, studyPlanItems, err := s.getStudyPlanItemByStudyPlanID(ctx, database.TextArray([]string{stepState.StudyPlanID}))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error: %w", err)
	}

	ids := make([]string, 0)
	for _, item := range studyPlanItems {
		ids = append(ids, item.ID.String)
	}

	stepState.Response, stepState.ResponseErr = epb.NewStudyPlanModifierServiceClient(s.Conn).UpdateStudyPlanItemsSchoolDate(ctx, &epb.UpdateStudyPlanItemsSchoolDateRequest{
		StudentId:        stepState.StudentID,
		StudyPlanItemIds: ids,
		SchoolDate:       timestamppb.New(time.Now()),
	})
	return StepStateToContext(ctx, stepState), nil
}
