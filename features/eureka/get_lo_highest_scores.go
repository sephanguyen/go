package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) getLoHighestScores(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studyPlanItemRepo := repositories.StudyPlanItemRepo{}
	studyPlanItems, err := studyPlanItemRepo.FetchByStudyProgressRequest(ctx, s.DB, database.Text(stepState.CourseID), database.Text(stepState.BookID), database.Text(stepState.StudentID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch study plan items: %w", err)
	}

	for _, each := range studyPlanItems {
		cs := new(entities.ContentStructure)
		each.ContentStructure.AssignTo(cs)

		if cs.LoID != "" {
			stepState.StudyPlanItemIDs = append(stepState.StudyPlanItemIDs, each.ID.String)
		}
	}

	stepState.AuthToken = stepState.StudentToken
	stepState.Response, err = epb.NewStudyPlanReaderServiceClient(s.Conn).GetLOHighestScoresByStudyPlanItemIDs(s.signedCtx(ctx), &epb.GetLOHighestScoresByStudyPlanItemIDsRequest{
		StudyPlanItemIds: stepState.StudyPlanItemIDs,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to fetch lo highest scores: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnCorrectHighestScoresBelongToStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.GetLOHighestScoresByStudyPlanItemIDsResponse)

	for _, each := range resp.LoHighestScores {
		if each.Percentage == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study plan item %v has 0 percentage", each.StudyPlanItemId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
