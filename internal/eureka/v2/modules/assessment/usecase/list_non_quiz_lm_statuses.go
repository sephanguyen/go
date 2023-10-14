package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
)

func (a *AssessmentUsecaseImpl) ListNonQuizLearningMaterialStatuses(ctx context.Context, courseID, userID string, learningMaterialIDs []string) (map[string]bool, error) {
	eventTypes := []string{"study_guide_finished", "video_finished"}
	events, err := a.StudentEventLogRepo.GetManyByEventTypesAndLMs(ctx, a.DB, courseID, userID, eventTypes, learningMaterialIDs)
	if err != nil {
		return nil, errors.New("AssessmentUsecase.ListNonQuizLearningMaterialStatuses", err)
	}

	type distinctEvent struct {
		StudyGuideFinished bool
		VideoFinished      bool
	}
	eventGroup := make(map[string]distinctEvent)
	for _, v := range events {
		ev := eventGroup[v.LearningMaterialID]
		switch v.EventType {
		case "study_guide_finished":
			ev.StudyGuideFinished = true
		case "video_finished":
			ev.VideoFinished = true
		default:
			continue
		}
		eventGroup[v.LearningMaterialID] = ev
	}

	statusMap := make(map[string]bool)
	for k, v := range eventGroup {
		statusMap[k] = v.VideoFinished && v.StudyGuideFinished
	}

	return statusMap, nil
}
